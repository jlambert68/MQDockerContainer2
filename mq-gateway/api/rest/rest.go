package rest

import (
	"encoding/json"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/mqcore"
	"log/slog"
	"net/http"
)

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

type Handler struct {
	GW *mqcore.Gateway
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("[REST] Put error: %v", err,
			"id", "73c893e6-e2f2-4e1b-a85f-e1422649436d")

		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("[REST] Get error: %v", err,
			"id", "793094b5-ddcf-497c-9772-dbf9d1df9867")

		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BrowseFirst(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("[REST] BrowseFirst error: %v", err,
			"id", "3a0a4b6d-292b-4db3-8a83-a2d9b804db9e")
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BrowseNext(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("[REST] BrowseNext error: %v", err,
			"id", "b52be2e8-30e4-43f2-aa5d-b1f7f117d7a3")
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/put", h.Put)
	mux.HandleFunc("/get", h.Get)
	mux.HandleFunc("/browse/first", h.BrowseFirst)
	mux.HandleFunc("/browse/next", h.BrowseNext)
	return mux
}
