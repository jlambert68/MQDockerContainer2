package mqrest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Client defines a small, stable REST client surface for MQ operations.
type Client interface {
	Put(ctx context.Context, req PutRequest) (PutResponse, error)
	Get(ctx context.Context, req GetRequest) (GetResponse, error)
	BrowseFirst(ctx context.Context, req BrowseFirstRequest) (BrowseResponse, error)
	BrowseNext(ctx context.Context, req BrowseNextRequest) (BrowseResponse, error)
	InquireQueue(ctx context.Context, req InquireQueueRequest) (InquireQueueResponse, error)
}

// Config controls how the REST client connects to the gateway.
type Config struct {
	BaseURL string
	Timeout time.Duration
	// TLSConfig enables mTLS/TLS when non-nil.
	TLSConfig *tls.Config
}

type client struct {
	baseURL string
	httpc   *http.Client
}

// NewClient creates a REST client with the given config.
func NewClient(cfg Config) Client {
	base := cfg.BaseURL
	if base == "" {
		base = "http://localhost:8080"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	var transport *http.Transport
	if cfg.TLSConfig != nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = cfg.TLSConfig
	}
	httpc := &http.Client{Timeout: timeout}
	if transport != nil {
		httpc.Transport = transport
	}
	return &client{
		baseURL: base,
		httpc:   httpc,
	}
}

type PutRequest struct {
	Queue   string `json:"queue"`
	Message string `json:"message"`
}

type PutResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type GetRequest struct {
	Queue       string `json:"queue"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}

type GetResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Empty   bool   `json:"empty"`
	Error   string `json:"error,omitempty"`
}

type BrowseFirstRequest struct {
	Queue       string `json:"queue"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}

type BrowseNextRequest struct {
	BrowseID    string `json:"browse_id"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}

type BrowseResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Empty    bool   `json:"empty"`
	BrowseID string `json:"browse_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type InquireQueueRequest struct {
	Queue string `json:"queue"`
}

type InquireQueueResponse struct {
	Status          string `json:"status"`
	Queue           string `json:"queue"`
	QueueDesc       string `json:"queue_desc"`
	QueueType       int32  `json:"queue_type"`
	QueueUsage      int32  `json:"queue_usage"`
	DefPersistence  int32  `json:"def_persistence"`
	InhibitGet      int32  `json:"inhibit_get"`
	InhibitPut      int32  `json:"inhibit_put"`
	CurrentQDepth   int32  `json:"current_q_depth"`
	MaxQDepth       int32  `json:"max_q_depth"`
	OpenInputCount  int32  `json:"open_input_count"`
	OpenOutputCount int32  `json:"open_output_count"`
	Error           string `json:"error,omitempty"`
}

func (c *client) Put(ctx context.Context, req PutRequest) (PutResponse, error) {
	var resp PutResponse
	err := c.post(ctx, "/put", req, &resp)
	return resp, err
}

func (c *client) Get(ctx context.Context, req GetRequest) (GetResponse, error) {
	var resp GetResponse
	err := c.post(ctx, "/get", req, &resp)
	return resp, err
}

func (c *client) BrowseFirst(ctx context.Context, req BrowseFirstRequest) (BrowseResponse, error) {
	var resp BrowseResponse
	err := c.post(ctx, "/browse/first", req, &resp)
	return resp, err
}

func (c *client) BrowseNext(ctx context.Context, req BrowseNextRequest) (BrowseResponse, error) {
	var resp BrowseResponse
	err := c.post(ctx, "/browse/next", req, &resp)
	return resp, err
}

func (c *client) InquireQueue(ctx context.Context, req InquireQueueRequest) (InquireQueueResponse, error) {
	var resp InquireQueueResponse
	err := c.post(ctx, "/inquire/queue", req, &resp)
	return resp, err
}

func (c *client) post(ctx context.Context, path string, req any, out any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}
	return nil
}
