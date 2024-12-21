package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatParticipant represents a participant in the chat
type ChatParticipant struct {
	stream pb.MedicalChatService_ChatStreamServer
	role   pb.Role
}

type ChatServer struct {
	pb.UnimplementedMedicalChatServiceServer
	*BaseServer
	activeStreams sync.Map
}

func NewChatServer(base *BaseServer) (*ChatServer, error) {

	server := &ChatServer{
		BaseServer:    base,
		activeStreams: sync.Map{},
	}

	return server, nil
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
	s.registerStream(chatID, req.SenderId.Value, stream, req.Role)

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
	// First broadcast the original message
	if err := s.broadcastMessage(req); err != nil {
		return err
	}

	// If message is from patient, simulate an AI response for doctor to review
	if req.Role == pb.Role_ROLE_PATIENT {

		s.broadcastResponse(&pb.ChatResponse{
			ChatId: req.ChatId,
			Type:   pb.ResponseType_DOCTOR_REVIEWING,
			Payload: &pb.ChatResponse_Message{
				Message: &pb.Message{
					SenderId: &pb.UUID{Value: []byte(uuid.Nil[:])},
					Content:  "Doctor is reviewing your question...",
					Role:     pb.Role_ROLE_SYSTEM,
				},
			},
			Timestamp: timestamppb.Now(),
		})

		// Hardcoded AI response for now
		aiResponse := "This is a simulated AI response to patient's question.."

		// Send AI draft only to doctor
		chatKey := fmt.Sprintf("%x", req.ChatId.Value)
		if participants, ok := s.activeStreams.Load(chatKey); ok {
			for _, participant := range participants.(map[string]*ChatParticipant) {
				if participant.role == pb.Role_ROLE_DOCTOR {
					participant.stream.Send(&pb.ChatResponse{
						ChatId: req.ChatId,
						Type:   pb.ResponseType_AI_DRAFT_READY,
						Payload: &pb.ChatResponse_AiDraft{
							AiDraft: &pb.AIDraft{
								Content:         aiResponse,
								ConfidenceScore: 0.85,
							},
						},
						Timestamp: timestamppb.Now(),
					})
				}
			}
		}
	}

	return nil
}

// Helper method to register a new participant
func (s *ChatServer) registerStream(chatID []byte, senderID []byte, stream pb.MedicalChatService_ChatStreamServer, role pb.Role) {
	chatKey := fmt.Sprintf("%x", chatID)
	participantKey := fmt.Sprintf("%x", senderID)

	if participants, ok := s.activeStreams.Load(chatKey); ok {
		participantsMap := participants.(map[string]*ChatParticipant)
		participantsMap[participantKey] = &ChatParticipant{
			stream: stream,
			role:   role,
		}
		s.activeStreams.Store(chatKey, participantsMap)
	} else {
		newParticipants := make(map[string]*ChatParticipant)
		newParticipants[participantKey] = &ChatParticipant{
			stream: stream,
			role:   role,
		}
		s.activeStreams.Store(chatKey, newParticipants)
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
	chatKey := fmt.Sprintf("%x", resp.ChatId.Value)
	if participants, ok := s.activeStreams.Load(chatKey); ok {
		for _, participant := range participants.(map[string]*ChatParticipant) {
			if err := participant.stream.Send(resp); err != nil {
				fmt.Printf("Error broadcasting to stream: %v\n", err)
			}
		}
	}
	return nil
}

func (s *ChatServer) handleJoinChat(ctx context.Context, req *pb.ChatRequest, stream pb.MedicalChatService_ChatStreamServer) error {
	// Register the doctor's stream
	chatID := req.ChatId.Value
	s.registerStream(chatID, req.SenderId.Value, stream, req.Role)

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
	// Parse review action from message content
	var review struct {
		Status  pb.ReviewStatus `json:"status"`
		Content string          `json:"content,omitempty"` // Optional for REJECTED status
	}

	if err := json.Unmarshal([]byte(req.Content), &review); err != nil {
		return fmt.Errorf("invalid review format: %v", err)
	}

	switch review.Status {
	case pb.ReviewStatus_APPROVED:
		// Doctor approved AI response - broadcast it to all
		return s.broadcastResponse(&pb.ChatResponse{
			ChatId: req.ChatId,
			Type:   pb.ResponseType_REVIEW_DONE,
			Payload: &pb.ChatResponse_Message{
				Message: &pb.Message{
					SenderId: req.SenderId,
					Content:  review.Content,
					Role:     pb.Role_ROLE_DOCTOR,
				},
			},
			Timestamp: timestamppb.Now(),
		})

	case pb.ReviewStatus_REJECTED:
		// Just acknowledge the rejection - doctor will send new response separately
		return stream.Send(&pb.ChatResponse{
			ChatId: req.ChatId,
			Type:   pb.ResponseType_NEW_MESSAGE,
			Payload: &pb.ChatResponse_Message{
				Message: &pb.Message{
					SenderId: &pb.UUID{Value: []byte(uuid.Nil[:])},
					Content:  "AI response rejected. Please compose new response.",
					Role:     pb.Role_ROLE_SYSTEM,
				},
			},
			Timestamp: timestamppb.Now(),
		})

	case pb.ReviewStatus_MODIFIED:
		// Doctor modified AI response - broadcast modified version
		return s.broadcastResponse(&pb.ChatResponse{
			ChatId: req.ChatId,
			Type:   pb.ResponseType_REVIEW_DONE,
			Payload: &pb.ChatResponse_Message{
				Message: &pb.Message{
					SenderId: req.SenderId,
					Content:  review.Content,
					Role:     pb.Role_ROLE_DOCTOR,
				},
			},
			Timestamp: timestamppb.Now(),
		})
	}

	return nil
}
