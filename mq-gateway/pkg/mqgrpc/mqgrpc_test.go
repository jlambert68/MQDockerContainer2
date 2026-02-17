package mqgrpc

import (
	"context"
	"net"
	"testing"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type testServer struct {
	mq_grpc_api.UnimplementedMqGrpcServicesServer
}

func (testServer) Put(context.Context, *mq_grpc_api.PutRequest) (*mq_grpc_api.PutResponse, error) {
	return &mq_grpc_api.PutResponse{Status: "ok"}, nil
}
func (testServer) Get(context.Context, *mq_grpc_api.GetRequest) (*mq_grpc_api.GetResponse, error) {
	return &mq_grpc_api.GetResponse{Status: "ok"}, nil
}
func (testServer) BrowseFirst(context.Context, *mq_grpc_api.BrowseFirstRequest) (*mq_grpc_api.BrowseResponse, error) {
	return &mq_grpc_api.BrowseResponse{Status: "ok"}, nil
}
func (testServer) BrowseNext(context.Context, *mq_grpc_api.BrowseNextRequest) (*mq_grpc_api.BrowseResponse, error) {
	return &mq_grpc_api.BrowseResponse{Status: "ok"}, nil
}
func (testServer) InquireQueue(context.Context, *mq_grpc_api.InquireQueueRequest) (*mq_grpc_api.InquireQueueResponse, error) {
	return &mq_grpc_api.InquireQueueResponse{Status: "ok"}, nil
}

func TestClientPutViaBufconn(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	mq_grpc_api.RegisterMqGrpcServicesServer(srv, testServer{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	c, err := NewClient(Config{
		Address: "passthrough:///bufnet",
		DialOptions: []grpc.DialOption{
			grpc.WithContextDialer(dialer),
		},
		TransportCredentials: func() *credentials.TransportCredentials {
			var cred credentials.TransportCredentials = insecure.NewCredentials()
			return &cred
		}(),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer c.Close()

	resp, err := c.Put(context.Background(), &mq_grpc_api.PutRequest{Queue: "Q1", Message: "m"})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if resp.GetStatus() != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
