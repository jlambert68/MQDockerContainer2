package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqcore"
)

type PutRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Payload to put.
	Message string `json:"message"`
}

type PutResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type GetRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}

type GetResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Empty   bool   `json:"empty"`
	Error   string `json:"error,omitempty"`
}

type BrowseFirstRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}

type BrowseNextRequest struct {
	// Browse session token returned from /browse/first.
	BrowseID string `json:"browse_id"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}

type BrowseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Empty   bool   `json:"empty"`
	// BrowseID is only set for BrowseFirst responses.
	BrowseID string `json:"browse_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type InquireQueueRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
}

type InquireQueueResponse struct {
	Status string `json:"status"`
	// Queue is the resolved queue name (may be normalized by MQ).
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

type Handler struct {
	// GW provides access to MQ operations.
	GW mqcore.Gateway
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	// Decode and validate the request.
	var req PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Queue == "" {
		http.Error(w, "queue required", http.StatusBadRequest)
		return
	}

	err := h.GW.Put(req.Queue, req.Message)
	resp := PutResponse{Status: "ok"}
	if err != nil {
		slog.Error("[REST] Put error",
			"error", err,
			"id", "73c893e6-e2f2-4e1b-a85f-e1422649436d")

		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Decode and validate the request.
	var req GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Queue == "" {
		http.Error(w, "queue required", http.StatusBadRequest)
		return
	}

	msg, empty, err := h.GW.Get(req.Queue, req.WaitMs, req.MaxMsgBytes)
	resp := GetResponse{Status: "ok", Message: msg, Empty: empty}
	if err != nil {
		slog.Error("[REST] Get error",
			"error", err,
			"id", "793094b5-ddcf-497c-9772-dbf9d1df9867")

		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BrowseFirst(w http.ResponseWriter, r *http.Request) {
	// Decode and validate the request.
	var req BrowseFirstRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Queue == "" {
		http.Error(w, "queue required", http.StatusBadRequest)
		return
	}

	msg, empty, browseID, err := h.GW.BrowseFirst(req.Queue, req.WaitMs, req.MaxMsgBytes)
	resp := BrowseResponse{Status: "ok", Message: msg, Empty: empty, BrowseID: browseID}
	if err != nil {
		slog.Error("[REST] BrowseFirst error",
			"error", err,
			"id", "3a0a4b6d-292b-4db3-8a83-a2d9b804db9e")
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BrowseNext(w http.ResponseWriter, r *http.Request) {
	// Decode and validate the request.
	var req BrowseNextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.BrowseID == "" {
		http.Error(w, "browse_id required", http.StatusBadRequest)
		return
	}

	msg, empty, err := h.GW.BrowseNext(req.BrowseID, req.WaitMs, req.MaxMsgBytes)
	resp := BrowseResponse{Status: "ok", Message: msg, Empty: empty, BrowseID: req.BrowseID}
	if err != nil {
		slog.Error("[REST] BrowseNext error",
			"error", err,
			"id", "b52be2e8-30e4-43f2-aa5d-b1f7f117d7a3")
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) InquireQueue(w http.ResponseWriter, r *http.Request) {
	// Decode and validate the request.
	var req InquireQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Queue == "" {
		http.Error(w, "queue required", http.StatusBadRequest)
		return
	}

	info, err := h.GW.InquireQueue(req.Queue)
	resp := InquireQueueResponse{Status: "ok"}
	if err != nil {
		slog.Error("[REST] InquireQueue error",
			"error", err,
			"id", "dd21d12e-b129-4244-bb8b-08a6bb6eea3c")
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	} else {
		resp.Queue = info.Name
		resp.QueueDesc = info.Description
		resp.QueueType = info.Type
		resp.QueueUsage = info.Usage
		resp.DefPersistence = info.DefPersistence
		resp.InhibitGet = info.InhibitGet
		resp.InhibitPut = info.InhibitPut
		resp.CurrentQDepth = info.CurrentDepth
		resp.MaxQDepth = info.MaxDepth
		resp.OpenInputCount = info.OpenInputCount
		resp.OpenOutputCount = info.OpenOutputCount
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Routes() http.Handler {
	// Register REST endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/put", h.Put)
	mux.HandleFunc("/get", h.Get)
	mux.HandleFunc("/browse/first", h.BrowseFirst)
	mux.HandleFunc("/browse/next", h.BrowseNext)
	mux.HandleFunc("/inquire/queue", h.InquireQueue)
	return mux
}
