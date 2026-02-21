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
	s.registry.UpdateLoad(req.WorkerId, req.Load)
	return &workerpb.HeartbeatResponse{Status: "alive"}, nil
}

func (s *GRPCServer) ListWorkers(ctx context.Context, _ *workerpb.Empty) (*workerpb.WorkerList, error) {
	snapshot := s.registry.List()

	resp := &workerpb.WorkerList{}

	for id, info := range snapshot {
		resp.Workers = append(resp.Workers, &workerpb.Worker{
			WorkerId:     id,
			LastSeenUnix: info.LastSeen.Unix(),
			Load:         info.Load,
		})
	}

	return resp, nil
}
