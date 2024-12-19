package server

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	pb "llm-qa-system/backend-service/src/proto"

	"google.golang.org/grpc"
)

type MedicalServer struct {
	pb.UnimplementedMedicalServiceServer
	*BaseServer
	llmClient pb.MedicalQAServiceClient
}

func NewMedicalServer(pool *pgxpool.Pool, llmServiceAddr string) (*MedicalServer, error) {
	conn, err := grpc.Dial(llmServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LLM service: %v", err)
	}

	return &MedicalServer{
		BaseServer: NewBaseServer(pool),
		llmClient:  pb.NewMedicalQAServiceClient(conn),
	}, nil
}

// Implement your RPC methods here...
