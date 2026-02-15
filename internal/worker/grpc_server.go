package worker

import (
	"context"
	"log"

	"jobqueue/proto/workerpb"
)

type GRPCServer struct {
	registry *Registry
	workerpb.UnimplementedWorkerServiceServer
}

func NewGRPCServer(reg *Registry) *GRPCServer {
	return &GRPCServer{registry: reg}
}

func (s *GRPCServer) Register(ctx context.Context, req *workerpb.RegisterRequest) (*workerpb.RegisterResponse, error) {
	log.Printf("worker registered: %s", req.WorkerId)
	s.registry.Register(req.WorkerId)

	return &workerpb.RegisterResponse{Status: "ok"}, nil
}

func (s *GRPCServer) Heartbeat(ctx context.Context, req *workerpb.HeartbeatRequest) (*workerpb.HeartbeatResponse, error) {
	s.registry.Heartbeat(req.WorkerId)
	return &workerpb.HeartbeatResponse{Status: "alive"}, nil
}
