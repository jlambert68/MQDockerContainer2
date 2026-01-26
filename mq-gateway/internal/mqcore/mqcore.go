package mqcore

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

type Gateway struct {
	QMgr ibmmq.MQQueueManager
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

	return &Gateway{QMgr: qMgr}, nil
}

func (g *Gateway) Close() {
	_ = g.QMgr.Disc()
}

// Put sends a message to the given queue.
func (g *Gateway) Put(queueName, message string) error {
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
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.WaitInterval = int32(waitMs)
	} else {
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
