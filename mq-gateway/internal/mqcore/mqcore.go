package mqcore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

type Gateway struct {
	QMgr ibmmq.MQQueueManager
	// browseMu protects browseSessions and browse cursor state.
	browseMu sync.Mutex
	// browseSessions holds active browse cursors keyed by browse_id.
	browseSessions map[string]*browseSession
	// browseSessionTTL limits how long an idle browse cursor can stay open.
	browseSessionTTL time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getbool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func NewGateway() (*Gateway, error) {
	// Read connection settings from environment variables.
	tlsEnabled := getbool("MQ_TLS_ENABLED", false)
	qMgrName := getenv("MQ_QMGR", "QM1")
	channel := getenv("MQ_CHANNEL", "DEV.TLS.SVRCONN") //"DEV.APP.SVRCONN")
	host := getenv("MQ_HOST", "mq-local_host")
	port := getenv("MQ_PORT", "1414")
	user := getenv("MQ_USER", "app")
	password := getenv("MQ_PASSWORD", "passw0rd")
	sslCipherSpec := getenv("MQ_SSLCIPH", "")
	sslKeyRepo := getenv("MQ_KEY_REPOSITORY", "")

	connName := fmt.Sprintf("%s(%s)", host, port)

	cd := ibmmq.NewMQCD()
	cd.ChannelName = channel
	cd.ConnectionName = connName

	cno := ibmmq.NewMQCNO()
	cno.Options = ibmmq.MQCNO_CLIENT_BINDING
	cno.ClientConn = cd

	if tlsEnabled {
		// Fail fast if TLS is requested but configuration is incomplete.
		// Fail fast if misconfigured
		if sslCipherSpec == "" || sslKeyRepo == "" {
			return nil, fmt.Errorf("TLS enabled but MQ_SSLCIPH or MQ_KEY_REPOSITORY is missing")
		}

		cd.SSLCipherSpec = sslCipherSpec

		sco := ibmmq.NewMQSCO()
		sco.KeyRepository = sslKeyRepo
		cno.SSLConfig = sco
	}

	if user != "" {
		// Enable user/password authentication when provided.
		csp := ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = user
		csp.Password = password
		cno.SecurityParms = csp
	}

	slog.Info(fmt.Sprintf("[mqcore] Connecting to MQ qmgr=%s at %s over channel=%s", qMgrName, connName, channel),
		"csp.AuthenticationType ", ibmmq.MQCSP_AUTH_USER_ID_AND_PWD,
		"csp.UserId", user,
		"csp.Password", password,
		"cno.SecurityParms", cno.SecurityParms,
		"id", "6d63fb38-b7b3-44ae-96de-81787257d3aa")

	qMgr, err := ibmmq.Connx(qMgrName, cno)

	if err != nil {
		return nil, err
	}
	slog.Info("[mqcore] Connected to queue manager",
		"QueueManager", qMgrName,
		"id", "bbbbe2e7-43b8-4163-8bd4-68ff6a8aba06")

	if tlsEnabled {
		// Best-effort TLS status logging via PCF.
		logTLSStatus(qMgr, channel)
	}

	return &Gateway{
		QMgr:             qMgr,
		browseSessions:   make(map[string]*browseSession),
		browseSessionTTL: 5 * time.Minute,
	}, nil
}

func (g *Gateway) Close() {
	// Close any browse cursors before disconnecting the QMgr.
	g.browseMu.Lock()
	for _, sess := range g.browseSessions {
		_ = sess.qObj.Close(0)
	}
	g.browseSessions = make(map[string]*browseSession)
	g.browseMu.Unlock()
	_ = g.QMgr.Disc()
}

type browseSession struct {
	// qObj is the open queue handle used for browsing.
	qObj ibmmq.MQObject
	// lastUsed tracks idle time for cleanup.
	lastUsed time.Time
}

// QueueInfo represents a stable subset of queue attributes we expose.
type QueueInfo struct {
	Name            string
	Description     string
	Type            int32
	Usage           int32
	DefPersistence  int32
	InhibitGet      int32
	InhibitPut      int32
	CurrentDepth    int32
	MaxDepth        int32
	OpenInputCount  int32
	OpenOutputCount int32
}

func logTLSStatus(qMgr ibmmq.MQQueueManager, channel string) {
	const (
		qCommandName = "SYSTEM.ADMIN.COMMAND.QUEUE"
		qReplyName   = "SYSTEM.DEFAULT.MODEL.QUEUE"
	)

	// Open command queue for PCF requests.
	odCmd := ibmmq.NewMQOD()
	odCmd.ObjectType = ibmmq.MQOT_Q
	odCmd.ObjectName = qCommandName
	cmdQ, err := qMgr.Open(odCmd, ibmmq.MQOO_OUTPUT)
	if err != nil {
		slog.Warn("[mqcore] TLS status unavailable (open command queue failed)", "error", err)
		return
	}
	defer cmdQ.Close(0)

	// Open reply queue (model) for PCF responses.
	odReply := ibmmq.NewMQOD()
	odReply.ObjectType = ibmmq.MQOT_Q
	odReply.ObjectName = qReplyName
	replyQ, err := qMgr.Open(odReply, ibmmq.MQOO_INPUT_EXCLUSIVE)
	if err != nil {
		slog.Warn("[mqcore] TLS status unavailable (open reply queue failed)", "error", err)
		return
	}
	defer replyQ.Close(0)

	putMQMD := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT | ibmmq.MQPMO_NEW_MSG_ID | ibmmq.MQPMO_NEW_CORREL_ID | ibmmq.MQPMO_FAIL_IF_QUIESCING
	putMQMD.Format = "MQADMIN"
	putMQMD.ReplyToQ = replyQ.Name
	putMQMD.MsgType = ibmmq.MQMT_REQUEST
	putMQMD.Report = ibmmq.MQRO_PASS_DISCARD_AND_EXPIRY

	cfh := ibmmq.NewMQCFH()
	cfh.Version = ibmmq.MQCFH_VERSION_3
	cfh.Type = ibmmq.MQCFT_COMMAND_XR
	cfh.Command = ibmmq.MQCMD_INQUIRE_CHANNEL_STATUS

	buf := make([]byte, 0)
	pcfparm := new(ibmmq.PCFParameter)
	pcfparm.Type = ibmmq.MQCFT_STRING
	pcfparm.Parameter = ibmmq.MQCACH_CHANNEL_NAME
	pcfparm.String = []string{channel}
	cfh.ParameterCount++
	buf = append(buf, pcfparm.Bytes()...)
	buf = append(cfh.Bytes(), buf...)

	if err := cmdQ.Put(putMQMD, pmo, buf); err != nil {
		slog.Warn("[mqcore] TLS status unavailable (PCF put failed)", "error", err)
		return
	}

	var sslCipherSpec string
	var sslCipherSuite string
	var sslPeer string
	var gotResponse bool

	getMQMD := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT | ibmmq.MQGMO_CONVERT | ibmmq.MQGMO_WAIT
	gmo.WaitInterval = int32((3 * time.Second) / time.Millisecond)

	for {
		buffer := make([]byte, 0, 10*1024)
		var datalen int
		buffer, datalen, err = replyQ.GetSlice(getMQMD, gmo, buffer)
		if err != nil {
			if mqret, ok := err.(*ibmmq.MQReturn); ok && mqret.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
				break
			}
			slog.Warn("[mqcore] TLS status unavailable (PCF get failed)", "error", err)
			return
		}

		// Decode PCF header to check response reason/control.
		cfh, offset := ibmmq.ReadPCFHeader(buffer)
		if cfh.Reason != ibmmq.MQRC_NONE {
			slog.Warn("[mqcore] TLS status unavailable (PCF response error)", "reason", cfh.Reason)
			return
		}
		gotResponse = true

		for offset < datalen {
			pcfParm, bytesRead := ibmmq.ReadPCFParameter(buffer[offset:])
			switch pcfParm.Parameter {
			case ibmmq.MQCACH_SSL_CIPHER_SPEC:
				if pcfParm.Type == ibmmq.MQCFT_STRING && len(pcfParm.String) > 0 {
					sslCipherSpec = strings.TrimSpace(pcfParm.String[0])
				}
			case ibmmq.MQCACH_SSL_CIPHER_SUITE:
				if pcfParm.Type == ibmmq.MQCFT_STRING && len(pcfParm.String) > 0 {
					sslCipherSuite = strings.TrimSpace(pcfParm.String[0])
				}
			case ibmmq.MQCACH_SSL_PEER_NAME:
				if pcfParm.Type == ibmmq.MQCFT_STRING && len(pcfParm.String) > 0 {
					sslPeer = strings.TrimSpace(pcfParm.String[0])
				}
			}
			offset += bytesRead
		}

		if cfh.Control == ibmmq.MQCFC_LAST {
			break
		}
	}

	if !gotResponse {
		slog.Warn("[mqcore] TLS status unavailable (no PCF response)")
		return
	}

	slog.Info("[mqcore] TLS negotiated",
		"Channel", channel,
		"SSLCIPH", sslCipherSpec,
		"SSLCIPH_SUITE", sslCipherSuite,
		"SSLPEER", sslPeer)
}

// Put sends a message to the given queue.
func (g *Gateway) Put(queueName, message string) error {
	// Put writes a single message to the queue (non-transactional).
	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = queueName

	qObj, err := g.QMgr.Open(od, ibmmq.MQOO_OUTPUT)
	if err != nil {
		return fmt.Errorf("MQOPEN: %w", err)
	}
	defer qObj.Close(0)

	md := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT

	if err := qObj.Put(md, pmo, []byte(message)); err != nil {
		return fmt.Errorf("MQPUT: %w", err)
	}
	return nil
}

// Get receives a message from the given queue.
func (g *Gateway) Get(queueName string, waitMs int, maxBytes int) (string, bool, error) {
	// Get consumes one message from the queue.
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}

	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = queueName

	qObj, err := g.QMgr.Open(od, ibmmq.MQOO_INPUT_AS_Q_DEF)
	if err != nil {
		return "", false, fmt.Errorf("MQOPEN: %w", err)
	}
	defer qObj.Close(0)

	md := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_FAIL_IF_QUIESCING | ibmmq.MQGMO_CONVERT

	if waitMs > 0 {
		// Wait for up to waitMs.
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.WaitInterval = int32(waitMs)
	} else {
		// Return immediately if no message is available.
		gmo.Options |= ibmmq.MQGMO_NO_WAIT
	}

	buf := make([]byte, maxBytes)
	msgLen, err := qObj.Get(md, gmo, buf)
	if err != nil {
		if mqret, ok := err.(*ibmmq.MQReturn); ok && mqret.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
			return "", true, nil
		}
		return "", false, fmt.Errorf("MQGET: %w", err)
	}
	return string(buf[:msgLen]), false, nil
}

func (g *Gateway) InquireQueue(queueName string) (*QueueInfo, error) {
	// InquireQueue returns attributes for the specified queue.
	if queueName == "" {
		return nil, fmt.Errorf("queue required")
	}

	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = queueName

	qObj, err := g.QMgr.Open(od, ibmmq.MQOO_INQUIRE)
	if err != nil {
		return nil, fmt.Errorf("MQOPEN: %w", err)
	}
	defer qObj.Close(0)

	// Selectors define which attributes to return.
	selectors := []int32{
		ibmmq.MQCA_Q_NAME,
		ibmmq.MQCA_Q_DESC,
		ibmmq.MQIA_Q_TYPE,
		ibmmq.MQIA_USAGE,
		ibmmq.MQIA_DEF_PERSISTENCE,
		ibmmq.MQIA_INHIBIT_GET,
		ibmmq.MQIA_INHIBIT_PUT,
		ibmmq.MQIA_CURRENT_Q_DEPTH,
		ibmmq.MQIA_MAX_Q_DEPTH,
		ibmmq.MQIA_OPEN_INPUT_COUNT,
		ibmmq.MQIA_OPEN_OUTPUT_COUNT,
	}

	attrs, err := qObj.Inq(selectors)
	if err != nil {
		return nil, fmt.Errorf("MQINQ: %w", err)
	}

	// Map raw selector results into a typed struct.
	info := &QueueInfo{
		Name:            stringAttr(attrs, ibmmq.MQCA_Q_NAME),
		Description:     stringAttr(attrs, ibmmq.MQCA_Q_DESC),
		Type:            intAttr(attrs, ibmmq.MQIA_Q_TYPE),
		Usage:           intAttr(attrs, ibmmq.MQIA_USAGE),
		DefPersistence:  intAttr(attrs, ibmmq.MQIA_DEF_PERSISTENCE),
		InhibitGet:      intAttr(attrs, ibmmq.MQIA_INHIBIT_GET),
		InhibitPut:      intAttr(attrs, ibmmq.MQIA_INHIBIT_PUT),
		CurrentDepth:    intAttr(attrs, ibmmq.MQIA_CURRENT_Q_DEPTH),
		MaxDepth:        intAttr(attrs, ibmmq.MQIA_MAX_Q_DEPTH),
		OpenInputCount:  intAttr(attrs, ibmmq.MQIA_OPEN_INPUT_COUNT),
		OpenOutputCount: intAttr(attrs, ibmmq.MQIA_OPEN_OUTPUT_COUNT),
	}

	return info, nil
}

func intAttr(attrs map[int32]interface{}, key int32) int32 {
	// intAttr safely reads integer selector values.
	v, ok := attrs[key]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case int32:
		return t
	case int:
		return int32(t)
	case int64:
		return int32(t)
	default:
		return 0
	}
}

func stringAttr(attrs map[int32]interface{}, key int32) string {
	// stringAttr safely reads string selector values.
	v, ok := attrs[key]
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func (g *Gateway) BrowseFirst(queueName string, waitMs int, maxBytes int) (string, bool, string, error) {
	// BrowseFirst opens a browse cursor and returns the first message.
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}

	// Evict idle browse cursors before creating a new one.
	g.cleanupBrowseSessions()

	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = queueName

	qObj, err := g.QMgr.Open(od, ibmmq.MQOO_BROWSE)
	if err != nil {
		return "", false, "", fmt.Errorf("MQOPEN: %w", err)
	}

	md := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_FAIL_IF_QUIESCING | ibmmq.MQGMO_BROWSE_FIRST | ibmmq.MQGMO_CONVERT

	if waitMs > 0 {
		// Wait for up to waitMs.
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.WaitInterval = int32(waitMs)
	} else {
		// Return immediately if no message is available.
		gmo.Options |= ibmmq.MQGMO_NO_WAIT
	}

	buf := make([]byte, maxBytes)
	msgLen, err := qObj.Get(md, gmo, buf)
	if err != nil {
		_ = qObj.Close(0)
		if mqret, ok := err.(*ibmmq.MQReturn); ok && mqret.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
			return "", true, "", nil
		}
		return "", false, "", fmt.Errorf("MQGET(BROWSE_FIRST): %w", err)
	}

	browseID, err := newBrowseID()
	if err != nil {
		_ = qObj.Close(0)
		return "", false, "", fmt.Errorf("browse id: %w", err)
	}

	// Store the browse cursor for subsequent BrowseNext calls.
	g.browseMu.Lock()
	g.browseSessions[browseID] = &browseSession{
		qObj:     qObj,
		lastUsed: time.Now(),
	}
	g.browseMu.Unlock()

	return string(buf[:msgLen]), false, browseID, nil
}

func (g *Gateway) BrowseNext(browseID string, waitMs int, maxBytes int) (string, bool, error) {
	// BrowseNext continues an existing browse cursor.
	if browseID == "" {
		return "", false, fmt.Errorf("browse_id required")
	}
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}

	sess, err := g.getBrowseSession(browseID)
	if err != nil {
		return "", false, err
	}

	md := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_FAIL_IF_QUIESCING | ibmmq.MQGMO_BROWSE_NEXT | ibmmq.MQGMO_CONVERT

	if waitMs > 0 {
		// Wait for up to waitMs.
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.WaitInterval = int32(waitMs)
	} else {
		// Return immediately if no message is available.
		gmo.Options |= ibmmq.MQGMO_NO_WAIT
	}

	buf := make([]byte, maxBytes)
	msgLen, err := sess.qObj.Get(md, gmo, buf)
	if err != nil {
		if mqret, ok := err.(*ibmmq.MQReturn); ok && mqret.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
			return "", true, nil
		}
		return "", false, fmt.Errorf("MQGET(BROWSE_NEXT): %w", err)
	}

	// Refresh idle timer after successful browse.
	g.touchBrowseSession(browseID)

	return string(buf[:msgLen]), false, nil
}

func newBrowseID() (string, error) {
	// Generate a random token for the browse session.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (g *Gateway) getBrowseSession(browseID string) (*browseSession, error) {
	// Drop expired sessions before lookup.
	g.cleanupBrowseSessions()

	g.browseMu.Lock()
	defer g.browseMu.Unlock()
	sess := g.browseSessions[browseID]
	if sess == nil {
		return nil, fmt.Errorf("browse_id not found or expired")
	}
	// Touch on read to extend the session lifetime.
	sess.lastUsed = time.Now()
	return sess, nil
}

func (g *Gateway) touchBrowseSession(browseID string) {
	// Touch without returning the session.
	g.browseMu.Lock()
	if sess := g.browseSessions[browseID]; sess != nil {
		sess.lastUsed = time.Now()
	}
	g.browseMu.Unlock()
}

func (g *Gateway) cleanupBrowseSessions() {
	// Close and remove sessions idle beyond browseSessionTTL.
	g.browseMu.Lock()
	defer g.browseMu.Unlock()
	now := time.Now()
	for id, sess := range g.browseSessions {
		if now.Sub(sess.lastUsed) > g.browseSessionTTL {
			_ = sess.qObj.Close(0)
			delete(g.browseSessions, id)
		}
	}
}
