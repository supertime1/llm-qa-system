package server

import (
	pb "llm-qa-system/backend-service/src/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LLMClient struct {
	client pb.MedicalQAServiceClient
	conn   *grpc.ClientConn
}

func NewLLMClient(addr string) (*LLMClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

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
