package grpcsrv

import (
	"context"
	"testing"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqcore"
)

type fakeGateway struct {
	putFn          func(queueName, message string) error
	getFn          func(queueName string, waitMs int, maxBytes int) (string, bool, error)
	browseFirstFn  func(queueName string, waitMs int, maxBytes int) (string, bool, string, error)
	browseNextFn   func(browseID string, waitMs int, maxBytes int) (string, bool, error)
	inquireQueueFn func(queueName string) (*mqcore.QueueInfo, error)
}

func (f *fakeGateway) Put(queueName, message string) error {
	if f.putFn != nil {
		return f.putFn(queueName, message)
	}
	return nil
}
func (f *fakeGateway) Get(queueName string, waitMs int, maxBytes int) (string, bool, error) {
	if f.getFn != nil {
		return f.getFn(queueName, waitMs, maxBytes)
	}
	return "", true, nil
}
func (f *fakeGateway) BrowseFirst(queueName string, waitMs int, maxBytes int) (string, bool, string, error) {
	if f.browseFirstFn != nil {
		return f.browseFirstFn(queueName, waitMs, maxBytes)
	}
	return "", true, "", nil
}
func (f *fakeGateway) BrowseNext(browseID string, waitMs int, maxBytes int) (string, bool, error) {
	if f.browseNextFn != nil {
		return f.browseNextFn(browseID, waitMs, maxBytes)
	}
	return "", true, nil
}
func (f *fakeGateway) InquireQueue(queueName string) (*mqcore.QueueInfo, error) {
	if f.inquireQueueFn != nil {
		return f.inquireQueueFn(queueName)
	}
	return &mqcore.QueueInfo{Name: queueName}, nil
}
func (f *fakeGateway) Close() {}

func TestPutValidationError(t *testing.T) {
	s := &Server{GW: &fakeGateway{}}
	resp, err := s.Put(context.Background(), &mq_grpc_api.PutRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.GetStatus() != "error" || resp.GetError() == "" {
		t.Fatalf("expected validation error response, got %+v", resp)
	}
}

func TestPutGatewayError(t *testing.T) {
	s := &Server{GW: &fakeGateway{putFn: func(_, _ string) error { return testErr("put failed") }}}
	resp, err := s.Put(context.Background(), &mq_grpc_api.PutRequest{Queue: "DEV.QUEUE.1", Message: "x"})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.GetStatus() != "error" {
		t.Fatalf("expected status error, got %q", resp.GetStatus())
	}
}

func TestInquireQueueSuccess(t *testing.T) {
	s := &Server{GW: &fakeGateway{inquireQueueFn: func(queueName string) (*mqcore.QueueInfo, error) {
		return &mqcore.QueueInfo{Name: queueName, CurrentDepth: 3, MaxDepth: 5000}, nil
	}}}
	resp, err := s.InquireQueue(context.Background(), &mq_grpc_api.InquireQueueRequest{Queue: "DEV.QUEUE.1"})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.GetStatus() != "ok" || resp.GetQueue() != "DEV.QUEUE.1" || resp.GetCurrentQDepth() != 3 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestBrowseNextValidationError(t *testing.T) {
	s := &Server{GW: &fakeGateway{}}
	resp, err := s.BrowseNext(context.Background(), &mq_grpc_api.BrowseNextRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.GetStatus() != "error" {
		t.Fatalf("expected error status, got %q", resp.GetStatus())
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
