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
