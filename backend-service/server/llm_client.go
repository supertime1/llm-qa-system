package server

import (
	pb "llm-qa-system/backend-service/src/proto"

	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LLMClient struct {
	client pb.MedicalQAServiceClient
	conn   *grpc.ClientConn
}

func NewLLMClient(addr string) (*LLMClient, error) {
	log.Printf("Attempting to connect to LLM service at: %s", addr)

	// Add connection timeout and retry
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),                 // Make connection blocking
		grpc.WithReturnConnectionError(), // Return detailed connection errors
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LLM service at %s: %v", addr, err)
	}

	log.Printf("Successfully connected to LLM service")
	client := pb.NewMedicalQAServiceClient(conn)
	return &LLMClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *LLMClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
