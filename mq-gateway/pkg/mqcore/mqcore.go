package mqcore

import internalmqcore "github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/mqcore"

// Gateway defines the public MQ operations used by callers outside this repo.
type Gateway interface {
	Put(queueName, message string) error
	Get(queueName string, waitMs int, maxBytes int) (string, bool, error)
	BrowseFirst(queueName string, waitMs int, maxBytes int) (string, bool, string, error)
	BrowseNext(browseID string, waitMs int, maxBytes int) (string, bool, error)
	InquireQueue(queueName string) (*QueueInfo, error)
	Close()
}

// QueueInfo exposes a stable subset of queue attributes for external callers.
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

type gateway struct {
	inner *internalmqcore.Gateway
}

// NewGateway creates a new MQ gateway using the internal implementation.
func NewGateway() (Gateway, error) {
	g, err := internalmqcore.NewGateway()
	if err != nil {
		return nil, err
	}
	return &gateway{inner: g}, nil
}

func (g *gateway) Put(queueName, message string) error {
	return g.inner.Put(queueName, message)
}

func (g *gateway) Get(queueName string, waitMs int, maxBytes int) (string, bool, error) {
	return g.inner.Get(queueName, waitMs, maxBytes)
}

func (g *gateway) BrowseFirst(queueName string, waitMs int, maxBytes int) (string, bool, string, error) {
	return g.inner.BrowseFirst(queueName, waitMs, maxBytes)
}

func (g *gateway) BrowseNext(browseID string, waitMs int, maxBytes int) (string, bool, error) {
	return g.inner.BrowseNext(browseID, waitMs, maxBytes)
}

func (g *gateway) InquireQueue(queueName string) (*QueueInfo, error) {
	info, err := g.inner.InquireQueue(queueName)
	if err != nil {
		return nil, err
	}
	return &QueueInfo{
		Name:            info.Name,
		Description:     info.Description,
		Type:            info.Type,
		Usage:           info.Usage,
		DefPersistence:  info.DefPersistence,
		InhibitGet:      info.InhibitGet,
		InhibitPut:      info.InhibitPut,
		CurrentDepth:    info.CurrentDepth,
		MaxDepth:        info.MaxDepth,
		OpenInputCount:  info.OpenInputCount,
		OpenOutputCount: info.OpenOutputCount,
	}, nil
}

func (g *gateway) Close() {
	g.inner.Close()
}
