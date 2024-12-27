package server

import (
	"fmt"
	pb "llm-qa-system/backend-service/src/proto"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Connection struct {
	conn *websocket.Conn
	role string
}

type WebSocketServer struct {
	*BaseServer
	llmClient   *LLMClient
	connections map[*websocket.Conn]*Connection
	mu          sync.RWMutex
}

func NewWebSocketServer(base *BaseServer, llmClient *LLMClient) *WebSocketServer {
	return &WebSocketServer{
		BaseServer:  base,
		llmClient:   llmClient,
		connections: make(map[*websocket.Conn]*Connection),
		mu:          sync.RWMutex{},
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Configure protojson

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	role := r.URL.Query().Get("role")
	log.Printf("New %s connected", role)

	s.mu.Lock()
	s.connections[conn] = &Connection{conn: conn, role: role}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
		conn.Close()
		log.Printf("%s disconnected", role)
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage error from %s: %v", role, err)
			break
		}

		var wsMsg pb.WebSocketMessage
		if err := unmarshaler.Unmarshal(rawMsg, &wsMsg); err != nil {
			log.Printf("Unmarshal error from %s: %v", role, err)
			break
		}

		switch wsMsg.Type {
		case pb.MessageType_PATIENT_MESSAGE:
			if msg := wsMsg.GetMessage(); msg != nil {
				// Forward to doctors
				s.broadcastToRole("doctor", &wsMsg)

				// Generate AI draft
				draftMsg := &pb.WebSocketMessage{
					Type: pb.MessageType_AI_DRAFT_READY,
					Payload: &pb.WebSocketMessage_AiDraft{
						AiDraft: &pb.AIDraftReady{
							MessageId:       fmt.Sprintf("msg_%d", time.Now().Unix()),
							OriginalMessage: msg.Content,
							Draft:           "This is a hardcoded AI draft answer",
							Timestamp:       timestamppb.Now(),
						},
					},
				}
				s.broadcastToRole("doctor", draftMsg)
			}

		case pb.MessageType_DOCTOR_MESSAGE:
			s.broadcastToRole("patient", &wsMsg)

		case pb.MessageType_DRAFT_REVIEW:
			if review := wsMsg.GetReview(); review != nil {
				switch review.Action {
				case pb.ReviewAction_ACCEPT, pb.ReviewAction_MODIFY:
					responseMsg := &pb.WebSocketMessage{
						Type: pb.MessageType_DOCTOR_MESSAGE,
						Payload: &pb.WebSocketMessage_Message{
							Message: &pb.Message{
								Content:   review.Content,
								Timestamp: timestamppb.Now(),
							},
						},
					}
					s.broadcastToRole("patient", responseMsg)
				}
			}
		}
	}
}

func (s *WebSocketServer) broadcastToRole(targetRole string, msg *pb.WebSocketMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Marshal message once
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	jsonBytes, err := marshaler.Marshal(msg)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}

	recipientCount := 0
	for conn, info := range s.connections {
		if info.role == targetRole {
			recipientCount++
			if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				log.Printf("Error broadcasting to %s: %v", targetRole, err)
			} else {
				log.Printf("Successfully sent message to a %s", targetRole)
			}
		}
	}

	log.Printf("Found %d recipients for role %s", recipientCount, targetRole)
}

func handlePatientMessage(s *WebSocketServer, msg *pb.Message) {
	// Forward patient's question to all doctors
	s.broadcastToRole("doctor", &pb.WebSocketMessage{
		Type: pb.MessageType_PATIENT_MESSAGE,
		Payload: &pb.WebSocketMessage_Message{
			Message: msg,
		},
	})

	// Generate AI draft and send to doctors
	draftMsg := &pb.WebSocketMessage{
		Type: pb.MessageType_AI_DRAFT_READY,
		Payload: &pb.WebSocketMessage_AiDraft{
			AiDraft: &pb.AIDraftReady{
				MessageId:       "msg_" + time.Now().String(), // TODO: Generate proper ID
				OriginalMessage: msg.Content,
				Draft:           "This is a hardcoded AI draft answer for: " + msg.Content,
				Timestamp:       timestamppb.Now(),
			},
		},
	}
	s.broadcastToRole("doctor", draftMsg)
}

func handleDoctorMessage(s *WebSocketServer, msg *pb.Message) {
	// Forward doctor's message to all patients
	s.broadcastToRole("patient", &pb.WebSocketMessage{
		Type: pb.MessageType_DOCTOR_MESSAGE,
		Payload: &pb.WebSocketMessage_Message{
			Message: msg,
		},
	})
}

func handleDraftReview(s *WebSocketServer, review *pb.DraftReview) {
	switch review.Action {
	case pb.ReviewAction_ACCEPT, pb.ReviewAction_MODIFY:
		// Send accepted or modified draft as a doctor message to patients
		s.broadcastToRole("patient", &pb.WebSocketMessage{
			Type: pb.MessageType_DOCTOR_MESSAGE,
			Payload: &pb.WebSocketMessage_Message{
				Message: &pb.Message{
					Content:   review.Content,
					Timestamp: timestamppb.Now(),
				},
			},
		})
	case pb.ReviewAction_REJECT:
		// For rejected drafts, doctor should send a separate message
		// No automatic message is sent here
	}

	// TODO: Save review action and content to database
	// This could include:
	// - Original question
	// - AI draft
	// - Review action
	// - Final content
	// - Timestamps
}
