package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestPutSuccess(t *testing.T) {
	called := false
	h := &Handler{GW: &fakeGateway{putFn: func(queueName, message string) error {
		called = true
		if queueName != "DEV.QUEUE.1" || message != "hello" {
			t.Fatalf("unexpected input queue=%q message=%q", queueName, message)
		}
		return nil
	}}}

	req := httptest.NewRequest(http.MethodPost, "/put", bytes.NewBufferString(`{"queue":"DEV.QUEUE.1","message":"hello"}`))
	rec := httptest.NewRecorder()
	h.Put(rec, req)

	if !called {
		t.Fatal("expected gateway Put to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp PutResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q", resp.Status)
	}
}

func TestPutInvalidJSON(t *testing.T) {
	h := &Handler{GW: &fakeGateway{}}
	req := httptest.NewRequest(http.MethodPost, "/put", bytes.NewBufferString(`{"queue":`))
	rec := httptest.NewRecorder()
	h.Put(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestBrowseNextMissingID(t *testing.T) {
	h := &Handler{GW: &fakeGateway{}}
	req := httptest.NewRequest(http.MethodPost, "/browse/next", bytes.NewBufferString(`{"wait_ms":1000}`))
	rec := httptest.NewRecorder()
	h.BrowseNext(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestInquireQueueErrorReturnsBadGateway(t *testing.T) {
	h := &Handler{GW: &fakeGateway{inquireQueueFn: func(queueName string) (*mqcore.QueueInfo, error) {
		return nil, assertErr("boom")
	}}}

	req := httptest.NewRequest(http.MethodPost, "/inquire/queue", bytes.NewBufferString(`{"queue":"DEV.QUEUE.1"}`))
	rec := httptest.NewRecorder()
	h.InquireQueue(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	var resp InquireQueueResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected status error, got %q", resp.Status)
	}
}

func TestRoutesRegistersEndpoints(t *testing.T) {
	h := &Handler{GW: &fakeGateway{}}
	ts := httptest.NewServer(h.Routes())
	defer ts.Close()

	for _, path := range []string{"/put", "/get", "/browse/first", "/browse/next", "/inquire/queue"} {
		resp, err := http.Post(ts.URL+path, "application/json", bytes.NewBufferString(`{}`))
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		if resp.StatusCode == http.StatusNotFound {
			t.Fatalf("expected route %s to exist", path)
		}
		_ = resp.Body.Close()
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
