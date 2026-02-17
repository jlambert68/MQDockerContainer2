package mq_grpc_api

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestPutRequestGettersAndReset(t *testing.T) {
	m := &PutRequest{Queue: "Q1", Message: "msg"}
	if m.GetQueue() != "Q1" || m.GetMessage() != "msg" {
		t.Fatalf("unexpected getters: %+v", m)
	}
	m.Reset()
	if m.GetQueue() != "" || m.GetMessage() != "" {
		t.Fatalf("expected reset zero values: %+v", m)
	}
}

func TestBrowseResponseGetters(t *testing.T) {
	m := &BrowseResponse{Status: "ok", Message: "m", Empty: false, BrowseId: "id"}
	if m.GetStatus() != "ok" || m.GetBrowseId() != "id" {
		t.Fatalf("unexpected browse response getters: %+v", m)
	}
}

func TestUnimplementedServerReturnsError(t *testing.T) {
	var s UnimplementedMqGrpcServicesServer
	_, err := s.Put(context.Background(), &PutRequest{})
	if err == nil {
		t.Fatal("expected unimplemented error")
	}
}

type testProtoServer struct {
	UnimplementedMqGrpcServicesServer
}

func (testProtoServer) Put(context.Context, *PutRequest) (*PutResponse, error) {
	return &PutResponse{Status: "ok"}, nil
}
func (testProtoServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return &GetResponse{Status: "ok"}, nil
}
func (testProtoServer) BrowseFirst(context.Context, *BrowseFirstRequest) (*BrowseResponse, error) {
	return &BrowseResponse{Status: "ok"}, nil
}
func (testProtoServer) BrowseNext(context.Context, *BrowseNextRequest) (*BrowseResponse, error) {
	return &BrowseResponse{Status: "ok"}, nil
}
func (testProtoServer) InquireQueue(context.Context, *InquireQueueRequest) (*InquireQueueResponse, error) {
	return &InquireQueueResponse{Status: "ok"}, nil
}

func TestGeneratedClientCanCallServer(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	RegisterMqGrpcServicesServer(s, testProtoServer{})
	go func() { _ = s.Serve(lis) }()
	defer s.Stop()

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.DialContext(
		context.Background(),
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	cli := NewMqGrpcServicesClient(conn)
	resp, err := cli.Put(context.Background(), &PutRequest{Queue: "Q1"})
	if err != nil {
		t.Fatalf("put rpc: %v", err)
	}
	if resp.GetStatus() != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestServiceDescHasMethods(t *testing.T) {
	if len(MqGrpcServices_ServiceDesc.Methods) != 5 {
		t.Fatalf("expected 5 methods, got %d", len(MqGrpcServices_ServiceDesc.Methods))
	}
	if MqGrpcServices_ServiceDesc.Streams != nil && len(MqGrpcServices_ServiceDesc.Streams) != 0 {
		t.Fatal(errors.New("expected no stream methods"))
	}
}
