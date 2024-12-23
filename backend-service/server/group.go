package server

import (
	"context"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerGroup struct {
	chatServer   *ChatServer
	healthServer *HealthServer
	db           *pgxpool.Pool
}

func NewServerGroup(pool *pgxpool.Pool, llmServiceAddr string, redisAddr string) (*ServerGroup, error) {

	baseServer := NewBaseServer(pool)

	// Create LLM client
	llmClient, err := NewLLMClient(llmServiceAddr)
	if err != nil {
		return nil, err
	}

	// Create chat server
	chatServer, err := NewChatServer(baseServer, llmClient, redisAddr)
	if err != nil {
		return nil, err
	}

	return &ServerGroup{
		db:           pool,
		chatServer:   chatServer,
		healthServer: newHealthServer(pool),
	}, nil
}

func (s *ServerGroup) Register(grpcServer *grpc.Server) {
	pb.RegisterMedicalChatServiceServer(grpcServer, s.chatServer) // Add this
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
