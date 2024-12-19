package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	pb "llm-qa-system/backend-service/src/proto"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerGroup struct {
	medicalServer *MedicalServer
	healthServer  *HealthServer
	db            *pgxpool.Pool
}

func NewServerGroup(pool *pgxpool.Pool, llmServiceAddr string) (*ServerGroup, error) {
	medicalServer, err := NewMedicalServer(pool, llmServiceAddr)
	if err != nil {
		return nil, err
	}

	return &ServerGroup{
		db:            pool,
		medicalServer: medicalServer,
		healthServer:  newHealthServer(pool),
	}, nil
}

func (s *ServerGroup) Register(grpcServer *grpc.Server) {
	pb.RegisterMedicalServiceServer(grpcServer, s.medicalServer)
	healthpb.RegisterHealthServer(grpcServer, s.healthServer)
	reflection.Register(grpcServer)
}

func (s *ServerGroup) Start(ctx context.Context) error {
	return nil
}

func (s *ServerGroup) Stop() {
	// Cleanup resources
	if s.db != nil {
		s.db.Close()
	}
}
