package grpcsrv

import (
	"context"
	"log/slog"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/mqcore"
)

type Server struct {
	mq_grpc_api.UnimplementedMqGrpcServicesServer
	GW *mqcore.Gateway
}

func (s *Server) Put(ctx context.Context, req *mq_grpc_api.PutRequest) (*mq_grpc_api.PutResponse, error) {
	// Validate request early to keep MQ errors clean.
	if req.GetQueue() == "" {
		return &mq_grpc_api.PutResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	err := s.GW.Put(req.GetQueue(), req.GetMessage())
	if err != nil {
		slog.Error("[gRPC] Put error: %v", err,
			"id", "52985a11-d814-403e-a00b-5cdeb2784025")
		return &mq_grpc_api.PutResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mq_grpc_api.PutResponse{Status: "ok"}, nil
}

func (s *Server) Get(ctx context.Context, req *mq_grpc_api.GetRequest) (*mq_grpc_api.GetResponse, error) {
	// Validate request early to keep MQ errors clean.
	if req.GetQueue() == "" {
		return &mq_grpc_api.GetResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	msg, empty, err := s.GW.Get(req.GetQueue(), int(req.GetWaitMs()), int(req.GetMaxMsgBytes()))
	if err != nil {
		slog.Error("[gRPC] Get error: %v", err,
			"id", "1b7707de-cf17-4080-922c-450362159d29")
		return &mq_grpc_api.GetResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mq_grpc_api.GetResponse{
		Status:  "ok",
		Message: msg,
		Empty:   empty,
	}, nil
}

func (s *Server) BrowseFirst(ctx context.Context, req *mq_grpc_api.BrowseFirstRequest) (*mq_grpc_api.BrowseResponse, error) {
	// BrowseFirst opens a server-side browse cursor.
	if req.GetQueue() == "" {
		return &mq_grpc_api.BrowseResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	msg, empty, browseID, err := s.GW.BrowseFirst(req.GetQueue(), int(req.GetWaitMs()), int(req.GetMaxMsgBytes()))
	if err != nil {
		slog.Error("[gRPC] BrowseFirst error: %v", err,
			"id", "0f17a466-8c6f-4f50-b7fb-7de77c8943b0")
		return &mq_grpc_api.BrowseResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mq_grpc_api.BrowseResponse{
		Status:   "ok",
		Message:  msg,
		Empty:    empty,
		BrowseId: browseID,
	}, nil
}

func (s *Server) BrowseNext(ctx context.Context, req *mq_grpc_api.BrowseNextRequest) (*mq_grpc_api.BrowseResponse, error) {
	// BrowseNext continues an existing browse cursor.
	if req.GetBrowseId() == "" {
		return &mq_grpc_api.BrowseResponse{
			Status: "error",
			Error:  "browse_id required",
		}, nil
	}

	msg, empty, err := s.GW.BrowseNext(req.GetBrowseId(), int(req.GetWaitMs()), int(req.GetMaxMsgBytes()))
	if err != nil {
		slog.Error("[gRPC] BrowseNext error: %v", err,
			"id", "6d0aa16a-5bb5-4ec9-9a1b-d4099af02bb5")
		return &mq_grpc_api.BrowseResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mq_grpc_api.BrowseResponse{
		Status:   "ok",
		Message:  msg,
		Empty:    empty,
		BrowseId: req.GetBrowseId(),
	}, nil
}

func (s *Server) InquireQueue(ctx context.Context, req *mq_grpc_api.InquireQueueRequest) (*mq_grpc_api.InquireQueueResponse, error) {
	// InquireQueue returns a subset of queue attributes.
	if req.GetQueue() == "" {
		return &mq_grpc_api.InquireQueueResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	info, err := s.GW.InquireQueue(req.GetQueue())
	if err != nil {
		slog.Error("[gRPC] InquireQueue error: %v", err,
			"id", "f5d27a47-3a67-46ca-9cfe-f5a7c12ee2f5")
		return &mq_grpc_api.InquireQueueResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mq_grpc_api.InquireQueueResponse{
		Status:          "ok",
		Queue:           info.Name,
		QueueDesc:       info.Description,
		QueueType:       info.Type,
		QueueUsage:      info.Usage,
		DefPersistence:  info.DefPersistence,
		InhibitGet:      info.InhibitGet,
		InhibitPut:      info.InhibitPut,
		CurrentQDepth:   info.CurrentDepth,
		MaxQDepth:       info.MaxDepth,
		OpenInputCount:  info.OpenInputCount,
		OpenOutputCount: info.OpenOutputCount,
	}, nil
}
