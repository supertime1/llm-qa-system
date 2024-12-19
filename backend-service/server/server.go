package server

import (
	"context"
	"database/sql"

	db "llm-qa-system/backend-service/src/db"
	pb "llm-qa-system/backend-service/src/proto"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMedicalServiceServer
	db        *db.Queries
	llmClient pb.MedicalQAServiceClient
}

func NewServer(database *sql.DB, llmServiceAddr string) (*Server, error) {
	// Connect to the database
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, llmServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Server{
		db:        db.New(database),
		llmClient: pb.NewMedicalQAServiceClient(conn),
	}, nil
}
