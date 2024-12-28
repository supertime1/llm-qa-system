package server

import (
	pb "llm-qa-system/backend-service/src/proto"

	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LLMClient struct {
	client pb.MedicalQAServiceClient
	conn   *grpc.ClientConn
	kafka  *kafka.Writer
}

func NewLLMClient(addr string, kafkaBrokers []string) (*LLMClient, error) {
	log.Printf("Attempting to connect to LLM service at: %s", addr)

	// Add connection timeout and retry
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LLM service at %s: %v", addr, err)
	}

	// Create Kafka writer just for llm-responses
	w := NewKafkaWriter(kafkaBrokers, TopicLLMResponses)

	log.Printf("Successfully connected to LLM service")
	client := pb.NewMedicalQAServiceClient(conn)
	return &LLMClient{
		client: client,
		conn:   conn,
		kafka:  w,
	}, nil
}

func (c *LLMClient) RequestDraft(sessionID string, message string) error {
	// Create request with proper protobuf structures
	req := &pb.QuestionRequest{
		QuestionId: &pb.UUID{
			Value: []byte(sessionID),
		},
		QuestionText: message,
		UserContext: &pb.UserContext{
			UserInfo: &pb.UserInfo{
				Age:    "35", // Changed to string as per proto definition
				Gender: pb.Gender_GENDER_MALE,
				MedicalHistory: []string{
					"Type 2 Diabetes",
					"Hypertension",
				},
			},
			BiometricData: []*pb.BiometricData{
				{
					Type:      pb.BiometricType_BIOMETRIC_BLOOD_PRESSURE,
					Value:     "120/80",
					Timestamp: timestamppb.Now(),
				},
				{
					Type:      pb.BiometricType_BIOMETRIC_HEART_RATE,
					Value:     "75",
					Timestamp: timestamppb.Now(),
				},
			},
			ChatHistory: []*pb.ChatMessage{
				{
					Role:      pb.Role_ROLE_PATIENT,
					Content:   message,
					Timestamp: timestamppb.Now(),
				},
			},
		},
	}

	// Make gRPC call to LLM service
	resp, err := c.client.GenerateDraftAnswer(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to generate answer: %v", err)
	}

	// Create draft message using AIDraftReady protobuf
	draftMsg := &pb.AIDraftReady{
		MessageId:       fmt.Sprintf("msg_%d", time.Now().Unix()),
		OriginalMessage: message,
		Draft:           resp.DraftAnswer, // Changed from Answer to DraftAnswer as per proto
		Timestamp:       timestamppb.Now(),
	}

	// Send to Kafka
	msgBytes, err := proto.Marshal(draftMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal draft: %v", err)
	}

	err = c.kafka.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(sessionID),
			Value: msgBytes,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to write to kafka: %v", err)
	}

	return nil
}

func (c *LLMClient) Close() error {
	var errs []error
	if c.kafka != nil {
		if err := c.kafka.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close kafka writer: %v", err))
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close grpc conn: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}
