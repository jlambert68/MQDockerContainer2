package mqrest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientTransportSelection(t *testing.T) {
	c1 := NewClient(Config{}).(*client)
	if c1.httpc.Transport != nil {
		t.Fatal("expected nil transport when TLS config is not provided")
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	c2 := NewClient(Config{TLSConfig: tlsCfg}).(*client)
	if c2.httpc.Transport == nil {
		t.Fatal("expected non-nil transport when TLS config is provided")
	}
}

func TestPutSuccess(t *testing.T) {
	var gotPath string
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewEncoder(w).Encode(PutResponse{Status: "ok"})
	}))
	defer ts.Close()

	c := NewClient(Config{BaseURL: ts.URL, Timeout: 2 * time.Second})
	resp, err := c.Put(context.Background(), PutRequest{Queue: "Q1", Message: "hello"})
	if err != nil {
		t.Fatalf("put error: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if gotPath != "/put" || gotMethod != http.MethodPost {
		t.Fatalf("unexpected call path=%s method=%s", gotPath, gotMethod)
	}
}

func TestPostReturnsErrorForHTTPFailureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(GetResponse{Status: "error", Error: "backend failure"})
	}))
	defer ts.Close()

	c := NewClient(Config{BaseURL: ts.URL, Timeout: 2 * time.Second})
	_, err := c.Get(context.Background(), GetRequest{Queue: "Q1"})
	if err == nil {
		t.Fatal("expected error for HTTP >= 400")
	}
}
