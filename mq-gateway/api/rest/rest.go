package rest

import (
	"encoding/json"
	"log"
	"mq-gateway/internal/mqcore"
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
		log.Printf("[REST] Put error: %v", err)
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
		log.Printf("[REST] Get error: %v", err)
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
	return mux
}
