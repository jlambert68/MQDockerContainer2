package grpcsrv

import (
	"context"
	"log"

	"mq-gateway/api/proto/mqpb"
	"mq-gateway/internal/mqcore"
)

type Server struct {
	//mqpb.UnimplementedMQServer
	GW *mqcore.Gateway
}

func (s *Server) Put(ctx context.Context, req *mqpb.PutRequest) (*mqpb.PutResponse, error) {
	if req.GetQueue() == "" {
		return &mqpb.PutResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	err := s.GW.Put(req.GetQueue(), req.GetMessage())
	if err != nil {
		log.Printf("[gRPC] Put error: %v", err)
		return &mqpb.PutResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mqpb.PutResponse{Status: "ok"}, nil
}

func (s *Server) Get(ctx context.Context, req *mqpb.GetRequest) (*mqpb.GetResponse, error) {
	if req.GetQueue() == "" {
		return &mqpb.GetResponse{
			Status: "error",
			Error:  "queue required",
		}, nil
	}

	msg, empty, err := s.GW.Get(req.GetQueue(), int(req.GetWaitMs()), int(req.GetMaxMsgBytes()))
	if err != nil {
		log.Printf("[gRPC] Get error: %v", err)
		return &mqpb.GetResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &mqpb.GetResponse{
		Status:  "ok",
		Message: msg,
		Empty:   empty,
	}, nil
}
