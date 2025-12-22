package grpcsrv

import (
	"context"
	"log"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/mqcore"
)

type Server struct {
	//mqpb.UnimplementedMQServer
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
		log.Printf("[gRPC] Put error: %v", err)
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
		log.Printf("[gRPC] Get error: %v", err)
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
