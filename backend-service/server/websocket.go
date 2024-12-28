package server

import (
	"context"
	"fmt"
	pb "llm-qa-system/backend-service/src/proto"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Connection struct {
	conn      *websocket.Conn
	role      string
	sessionID string
}

type ChatSession struct {
	patientConn *Connection
	doctorConn  *Connection
	sessionID   string
	created     time.Time
}

type WebSocketServer struct {
	*BaseServer
	llmClient  *LLMClient
	sessions   map[string]*ChatSession
	mu         sync.RWMutex
	reader     *kafka.Reader
	writer     *kafka.Writer
	cancelFunc context.CancelFunc
}

func NewWebSocketServer(base *BaseServer, llmClient *LLMClient, kafkaBrokers []string) *WebSocketServer {
	ctx, cancel := context.WithCancel(context.Background())

	ws := &WebSocketServer{
		BaseServer: base,
		llmClient:  llmClient,
		sessions:   make(map[string]*ChatSession),
		mu:         sync.RWMutex{},
		reader:     NewKafkaReader(kafkaBrokers, TopicLLMResponses, GroupIDWebSocket),
		writer:     NewKafkaWriter(kafkaBrokers, TopicPatientMessages),
		cancelFunc: cancel,
	}

	// Start Kafka consumer
	go ws.consumeKafkaMessages(ctx)

	return ws
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // You might want to restrict this in production
	},
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Configure protojson
	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	role := r.URL.Query().Get("role")
	sessionID := r.URL.Query().Get("session")
	doctorToken := r.URL.Query().Get("token")

	connection := &Connection{
		conn:      conn,
		role:      role,
		sessionID: sessionID,
	}

	// Handle session management
	if err := s.handleSession(connection, role, sessionID, doctorToken); err != nil {
		log.Printf("Session error: %v", err)
		conn.Close()
		return
	}

	defer s.handleDisconnect(connection)

	log.Printf("New %s connected to session %s", role, sessionID)

	// Message handling loop
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
				// 1. Forward original message to doctor
				s.broadcastToRole(connection.sessionID, "doctor", &wsMsg)

				// 2. Write to Kafka for LLM processing
				patientMsg := &pb.Message{
					Content:   msg.Content,
					Timestamp: timestamppb.Now(),
				}

				msgBytes, err := proto.Marshal(patientMsg)
				if err != nil {
					log.Printf("Error marshaling message: %v", err)
					continue
				}

				err = s.writer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(connection.sessionID),
					Value: msgBytes,
				})

				if err != nil {
					log.Printf("Error writing to Kafka: %v", err)
					// Send error message to patient
					s.broadcastToRole(connection.sessionID, "doctor", &pb.WebSocketMessage{
						Type: pb.MessageType_ERROR,
						Payload: &pb.WebSocketMessage_Error{
							Error: &pb.Error{Message: "Failed to send message to LLM"},
						},
					})
				}
			}

		case pb.MessageType_DOCTOR_MESSAGE:
			if msg := wsMsg.GetMessage(); msg != nil {
				s.broadcastToRole(connection.sessionID, "patient", &wsMsg)
			}

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
					s.broadcastToRole(connection.sessionID, "patient", responseMsg)
				}
			}
		}
	}
}

func (s *WebSocketServer) handleSession(conn *Connection, role, sessionID, doctorToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch role {
	case "patient":
		if sessionID == "" {
			// Create new session for patient
			sessionID = generateSessionID()
			s.sessions[sessionID] = &ChatSession{
				patientConn: conn,
				sessionID:   sessionID,
				created:     time.Now(),
			}
			// Inform patient of their session ID
			conn.conn.WriteJSON(map[string]string{"session_id": sessionID})
			conn.sessionID = sessionID
		} else {
			return fmt.Errorf("patients cannot join existing sessions")
		}

	case "doctor":
		if !isValidDoctorToken(doctorToken) {
			return fmt.Errorf("invalid doctor token")
		}
		if session, exists := s.sessions[sessionID]; exists {
			if session.doctorConn != nil {
				return fmt.Errorf("session already has a doctor")
			}
			session.doctorConn = conn
		} else {
			return fmt.Errorf("session not found")
		}

	default:
		return fmt.Errorf("invalid role: %s", role)
	}

	return nil
}

func (s *WebSocketServer) handleDisconnect(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[conn.sessionID]; exists {
		switch conn.role {
		case "patient":
			if session.patientConn == conn {
				delete(s.sessions, conn.sessionID)
			}
		case "doctor":
			if session.doctorConn == conn {
				session.doctorConn = nil
			}
		}
	}

	conn.conn.Close()
	log.Printf("%s disconnected from session %s", conn.role, conn.sessionID)
}

func (s *WebSocketServer) broadcastToRole(sessionID, targetRole string, msg *pb.WebSocketMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		log.Printf("Session %s not found", sessionID)
		return
	}

	var targetConn *Connection
	switch targetRole {
	case "patient":
		targetConn = session.patientConn
	case "doctor":
		targetConn = session.doctorConn
	}

	if targetConn == nil {
		log.Printf("No %s connected to session %s", targetRole, sessionID)
		return
	}

	// Marshal message
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	jsonBytes, err := marshaler.Marshal(msg)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}

	if err := targetConn.conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
		log.Printf("Error broadcasting to %s: %v", targetRole, err)
	} else {
		log.Printf("Successfully sent message to %s in session %s", targetRole, sessionID)
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func isValidDoctorToken(token string) bool {
	// TODO: Implement proper token validation
	return token != ""
}

// This consumer handles the LLM draft from Kafka and broadcasts to doctor
func (s *WebSocketServer) consumeKafkaMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := s.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading kafka message: %v", err)
				continue
			}

			// Unmarshal the message
			var draftReady pb.AIDraftReady
			if err := proto.Unmarshal(msg.Value, &draftReady); err != nil {
				log.Printf("Error unmarshaling draft: %v", err)
				continue
			}

			// Create WebSocket message with the draft
			wsMsg := &pb.WebSocketMessage{
				Type: pb.MessageType_AI_DRAFT_READY,
				Payload: &pb.WebSocketMessage_AiDraft{
					AiDraft: &draftReady,
				},
			}

			// Get session ID from Kafka message key and broadcast to doctor
			sessionID := string(msg.Key)
			s.broadcastToRole(sessionID, "doctor", wsMsg)
		}
	}
}

func (s *WebSocketServer) Close() error {
	// Cancel the context to stop the Kafka consumer
	s.cancelFunc()

	// Close the Kafka reader
	if err := s.reader.Close(); err != nil {
		return fmt.Errorf("failed to close kafka reader: %v", err)
	}
	return nil
}
