package mqgrpc

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client defines a small, stable gRPC client surface for MQ operations.
type Client interface {
	Put(ctx context.Context, req *mq_grpc_api.PutRequest) (*mq_grpc_api.PutResponse, error)
	Get(ctx context.Context, req *mq_grpc_api.GetRequest) (*mq_grpc_api.GetResponse, error)
	BrowseFirst(ctx context.Context, req *mq_grpc_api.BrowseFirstRequest) (*mq_grpc_api.BrowseResponse, error)
	BrowseNext(ctx context.Context, req *mq_grpc_api.BrowseNextRequest) (*mq_grpc_api.BrowseResponse, error)
	InquireQueue(ctx context.Context, req *mq_grpc_api.InquireQueueRequest) (*mq_grpc_api.InquireQueueResponse, error)
	Close() error
}

// Config controls how the client connects to the gRPC server.
type Config struct {
	Address string
	Timeout time.Duration
	// TLSConfig enables mTLS/TLS when non-nil.
	TLSConfig *tls.Config
	// TransportCredentials can be provided for full control. Takes precedence over TLSConfig.
	TransportCredentials *credentials.TransportCredentials
	// DialOptions can be used for interceptors, retries, etc.
	DialOptions []grpc.DialOption
}

type client struct {
	conn *grpc.ClientConn
	api  mq_grpc_api.MqGrpcServicesClient
}

// NewClient dials the MQ gRPC service and returns a wrapper client.
func NewClient(cfg Config) (Client, error) {
	if cfg.Address == "" {
		cfg.Address = "localhost:9090"
	}

	opts := make([]grpc.DialOption, 0, 2+len(cfg.DialOptions))
	if cfg.TransportCredentials != nil {
		opts = append(opts, grpc.WithTransportCredentials(*cfg.TransportCredentials))
	} else if cfg.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(cfg.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, cfg.DialOptions...)

	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return nil, err
	}

	return &client{
		conn: conn,
		api:  mq_grpc_api.NewMqGrpcServicesClient(conn),
	}, nil
}

func (c *client) Put(ctx context.Context, req *mq_grpc_api.PutRequest) (*mq_grpc_api.PutResponse, error) {
	return c.api.Put(ctx, req)
}

func (c *client) Get(ctx context.Context, req *mq_grpc_api.GetRequest) (*mq_grpc_api.GetResponse, error) {
	return c.api.Get(ctx, req)
}

func (c *client) BrowseFirst(ctx context.Context, req *mq_grpc_api.BrowseFirstRequest) (*mq_grpc_api.BrowseResponse, error) {
	return c.api.BrowseFirst(ctx, req)
}

func (c *client) BrowseNext(ctx context.Context, req *mq_grpc_api.BrowseNextRequest) (*mq_grpc_api.BrowseResponse, error) {
	return c.api.BrowseNext(ctx, req)
}

func (c *client) InquireQueue(ctx context.Context, req *mq_grpc_api.InquireQueueRequest) (*mq_grpc_api.InquireQueueResponse, error) {
	return c.api.InquireQueue(ctx, req)
}

func (c *client) Close() error {
	return c.conn.Close()
}
