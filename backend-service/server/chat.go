package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatServer struct {
	pb.UnimplementedMedicalChatServiceServer
	*BaseServer
	llmClient     *LLMClient
	redis         *redis.Client
	activeStreams sync.Map
}

func NewChatServer(base *BaseServer, llmAddr string, redisClient *redis.Client) (*ChatServer, error) {
	llmClient, err := NewLLMClient(llmAddr)
	if err != nil {
		return nil, err
	}

	server := &ChatServer{
		BaseServer:    base,
		llmClient:     llmClient,
		redis:         redisClient,
		activeStreams: sync.Map{},
	}

	// Start Redis subscription handler
	go server.handleRedisSubscriptions(context.Background())

	return server, nil
}

func (s *ChatServer) handleRedisSubscriptions(ctx context.Context) {
	pubsub := s.redis.Subscribe(ctx, "chat:*")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			fmt.Printf("Redis subscription error: %v\n", err)
			continue
		}

		// Parse message and broadcast to relevant streams
		var redisMsg struct {
			ChatID   string         `json:"chat_id"`
			Type     string         `json:"type"`
			Content  string         `json:"content"`
			Metadata map[string]any `json:"metadata"`
		}

		if err := json.Unmarshal([]byte(msg.Payload), &redisMsg); err != nil {
			fmt.Printf("Failed to parse Redis message: %v\n", err)
			continue
		}

		// Convert chat ID from string to UUID bytes
		chatIDBytes, err := uuid.Parse(redisMsg.ChatID)
		if err != nil {
			fmt.Printf("Invalid chat ID in Redis message: %v\n", err)
			continue
		}

		// Broadcast to active streams
		s.broadcastResponse(&pb.ChatResponse{
			ChatId: &pb.UUID{Value: chatIDBytes[:]},
			Type:   pb.ResponseType_NEW_MESSAGE,
			Payload: &pb.ChatResponse_Message{
				Message: &pb.Message{
					SenderId: &pb.UUID{Value: uuid.Nil[:]},
					Content:  redisMsg.Content,
					Role:     pb.Role_ROLE_SYSTEM,
				},
			},
			Timestamp: timestamppb.Now(),
		})
	}
}

func (s *ChatServer) ChatStream(stream pb.MedicalChatService_ChatStreamServer) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			req, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("error receiving message: %v", err)
			}

			switch req.Type {
			case pb.RequestType_START_CHAT:
				err = s.handleStartChat(ctx, req, stream)
			case pb.RequestType_SEND_MESSAGE:
				err = s.handleMessage(ctx, req, stream)
			case pb.RequestType_JOIN_CHAT:
				err = s.handleJoinChat(ctx, req, stream)
			case pb.RequestType_SUBMIT_REVIEW:
				err = s.handleReview(ctx, req, stream)
			}

			if err != nil {
				return fmt.Errorf("error handling request: %v", err)
			}
		}
	}
}

func (s *ChatServer) handleStartChat(ctx context.Context, req *pb.ChatRequest, stream pb.MedicalChatService_ChatStreamServer) error {
	chatID := req.ChatId.Value
	s.registerStream(chatID, stream)

	// Subscribe to Redis channel for this chat
	go s.subscribeToChat(ctx, fmt.Sprintf("%x", chatID))

	return stream.Send(&pb.ChatResponse{
		ChatId: req.ChatId,
		Type:   pb.ResponseType_NEW_MESSAGE,
		Payload: &pb.ChatResponse_Message{
			Message: &pb.Message{
				SenderId: &pb.UUID{Value: []byte(uuid.Nil[:])},
				Content:  "Chat session started",
				Role:     pb.Role_ROLE_SYSTEM,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) handleMessage(ctx context.Context, req *pb.ChatRequest, stream pb.MedicalChatService_ChatStreamServer) error {
	// Broadcast message to all participants
	if err := s.broadcastMessage(req); err != nil {
		return err
	}

	// If message is from patient, generate AI draft
	if req.Role == pb.Role_ROLE_PATIENT {
		go s.generateAndBroadcastDraft(ctx, req)
	}

	return nil
}

func (s *ChatServer) generateAndBroadcastDraft(ctx context.Context, req *pb.ChatRequest) {
	// Call LLM service using the client field
	llmResp, err := s.llmClient.client.GenerateDraftAnswer(ctx, &pb.QuestionRequest{
		QuestionId:   &pb.UUID{Value: req.ChatId.Value},
		QuestionText: req.Content,
		UserContext:  &pb.UserContext{}, // Add user context if needed
	})
	if err != nil {
		s.broadcastError(req.ChatId, fmt.Sprintf("Failed to generate AI draft: %v", err))
		return
	}

	// Publish draft to Redis for doctor notification
	notification := map[string]interface{}{
		"chat_id":          fmt.Sprintf("%x", req.ChatId.Value),
		"draft_answer":     llmResp.DraftAnswer,
		"confidence_score": llmResp.ConfidenceScore,
		"timestamp":        time.Now().Unix(),
	}

	if payload, err := json.Marshal(notification); err == nil {
		s.redis.Publish(ctx, fmt.Sprintf("chat:%x:draft", req.ChatId.Value), payload)
	}

	// Broadcast AI draft to all participants
	s.broadcastResponse(&pb.ChatResponse{
		ChatId: req.ChatId,
		Type:   pb.ResponseType_AI_DRAFT_READY,
		Payload: &pb.ChatResponse_AiDraft{
			AiDraft: &pb.AIDraft{
				Content:         llmResp.DraftAnswer,
				ConfidenceScore: llmResp.ConfidenceScore,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

// Helper methods for managing streams and broadcasting
func (s *ChatServer) registerStream(chatID []byte, stream pb.MedicalChatService_ChatStreamServer) {
	key := fmt.Sprintf("%x", chatID)
	if streams, ok := s.activeStreams.Load(key); ok {
		streams = append(streams.([]pb.MedicalChatService_ChatStreamServer), stream)
		s.activeStreams.Store(key, streams)
	} else {
		s.activeStreams.Store(key, []pb.MedicalChatService_ChatStreamServer{stream})
	}
}

func (s *ChatServer) broadcastMessage(req *pb.ChatRequest) error {
	return s.broadcastResponse(&pb.ChatResponse{
		ChatId: req.ChatId,
		Type:   pb.ResponseType_NEW_MESSAGE,
		Payload: &pb.ChatResponse_Message{
			Message: &pb.Message{
				SenderId: req.SenderId,
				Content:  req.Content,
				Role:     req.Role,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) broadcastResponse(resp *pb.ChatResponse) error {
	key := fmt.Sprintf("%x", resp.ChatId.Value)
	if streams, ok := s.activeStreams.Load(key); ok {
		for _, stream := range streams.([]pb.MedicalChatService_ChatStreamServer) {
			if err := stream.Send(resp); err != nil {
				// Log error but continue broadcasting to others
				fmt.Printf("Error broadcasting to stream: %v\n", err)
			}
		}
	}
	return nil
}

func (s *ChatServer) broadcastError(chatID *pb.UUID, errMsg string) {
	s.broadcastResponse(&pb.ChatResponse{
		ChatId: chatID,
		Type:   pb.ResponseType_ERROR,
		Payload: &pb.ChatResponse_Message{
			Message: &pb.Message{
				SenderId: &pb.UUID{Value: []byte(uuid.Nil[:])},
				Content:  errMsg,
				Role:     pb.Role_ROLE_SYSTEM,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) handleJoinChat(ctx context.Context, req *pb.ChatRequest, stream pb.MedicalChatService_ChatStreamServer) error {
	// Register the doctor's stream
	chatID := req.ChatId.Value
	s.registerStream(chatID, stream)

	// Subscribe to Redis channel for this chat
	go s.subscribeToChat(ctx, fmt.Sprintf("%x", chatID))

	// Notify all participants that a doctor has joined
	return s.broadcastResponse(&pb.ChatResponse{
		ChatId: req.ChatId,
		Type:   pb.ResponseType_NEW_MESSAGE,
		Payload: &pb.ChatResponse_Message{
			Message: &pb.Message{
				SenderId: req.SenderId,
				Content:  "Doctor has joined the chat",
				Role:     pb.Role_ROLE_SYSTEM,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) handleReview(ctx context.Context, req *pb.ChatRequest, stream pb.MedicalChatService_ChatStreamServer) error {
	// Parse review status from message content
	var review struct {
		Status  pb.ReviewStatus `json:"status"`
		Content string          `json:"content"`
	}
	if err := json.Unmarshal([]byte(req.Content), &review); err != nil {
		return fmt.Errorf("invalid review format: %v", err)
	}

	// Publish review to Redis
	notification := map[string]interface{}{
		"chat_id":          fmt.Sprintf("%x", req.ChatId.Value),
		"review_status":    review.Status.String(),
		"modified_content": review.Content,
		"reviewer_id":      fmt.Sprintf("%x", req.SenderId.Value),
		"timestamp":        time.Now().Unix(),
	}

	if payload, err := json.Marshal(notification); err == nil {
		s.redis.Publish(ctx, fmt.Sprintf("chat:%x:review", req.ChatId.Value), payload)
	}

	// Broadcast review status to all participants
	return s.broadcastResponse(&pb.ChatResponse{
		ChatId: req.ChatId,
		Type:   pb.ResponseType_REVIEW_DONE,
		Payload: &pb.ChatResponse_Review{
			Review: &pb.ReviewUpdate{
				Status:          review.Status,
				ModifiedContent: review.Content,
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) handleDraftNotification(data map[string]interface{}) {
	chatIDStr, ok := data["chat_id"].(string)
	if !ok {
		fmt.Printf("Invalid chat ID in draft notification\n")
		return
	}

	chatIDBytes, err := uuid.Parse(chatIDStr)
	if err != nil {
		fmt.Printf("Failed to parse chat ID: %v\n", err)
		return
	}

	s.broadcastResponse(&pb.ChatResponse{
		ChatId: &pb.UUID{Value: chatIDBytes[:]},
		Type:   pb.ResponseType_AI_DRAFT_READY,
		Payload: &pb.ChatResponse_AiDraft{
			AiDraft: &pb.AIDraft{
				Content:         data["draft_answer"].(string),
				ConfidenceScore: float32(data["confidence_score"].(float64)),
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) handleReviewNotification(data map[string]interface{}) {
	chatIDStr, ok := data["chat_id"].(string)
	if !ok {
		fmt.Printf("Invalid chat ID in review notification\n")
		return
	}

	chatIDBytes, err := uuid.Parse(chatIDStr)
	if err != nil {
		fmt.Printf("Failed to parse chat ID: %v\n", err)
		return
	}

	status := pb.ReviewStatus_REVIEW_UNKNOWN
	switch data["review_status"].(string) {
	case "APPROVED":
		status = pb.ReviewStatus_APPROVED
	case "MODIFIED":
		status = pb.ReviewStatus_MODIFIED
	case "REJECTED":
		status = pb.ReviewStatus_REJECTED
	}

	s.broadcastResponse(&pb.ChatResponse{
		ChatId: &pb.UUID{Value: chatIDBytes[:]},
		Type:   pb.ResponseType_REVIEW_DONE,
		Payload: &pb.ChatResponse_Review{
			Review: &pb.ReviewUpdate{
				Status:          status,
				ModifiedContent: data["modified_content"].(string),
			},
		},
		Timestamp: timestamppb.Now(),
	})
}

func (s *ChatServer) subscribeToChat(ctx context.Context, chatID string) {
	pubsub := s.redis.Subscribe(ctx, fmt.Sprintf("chat:%s:*", chatID))
	defer pubsub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				fmt.Printf("Error receiving Redis message: %v\n", err)
				continue
			}
			// Handle message based on channel
			s.handleRedisMessage(msg)
		}
	}
}

func (s *ChatServer) handleRedisMessage(msg *redis.Message) {
	// Parse message and update relevant streams
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
		fmt.Printf("Error parsing Redis message: %v\n", err)
		return
	}

	// Handle different message types based on channel pattern
	switch {
	case msg.Channel == "chat:*:draft":
		s.handleDraftNotification(data)
	case msg.Channel == "chat:*:review":
		s.handleReviewNotification(data)
	}
}
