package server

import (
	"log"
	"net/http"
	"sync"
	"time"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/gorilla/websocket"
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
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("ReadJSON error from %s: %v", role, err)
			break
		}

		log.Printf("Received raw message from %s: %+v", role, msg)

		msgType, ok := msg["type"].(float64)
		if !ok {
			log.Printf("Invalid message type from %s: %+v", role, msg["type"])
			continue
		}

		switch int32(msgType) {
		case int32(pb.MessageType_PATIENT_MESSAGE):
			log.Printf("Broadcasting patient message to doctors")
			s.broadcastToRole("doctor", msg)

		case int32(pb.MessageType_DOCTOR_MESSAGE):
			log.Printf("Broadcasting doctor message to patients")
			s.broadcastToRole("patient", msg)
		}
	}
}

func (s *WebSocketServer) broadcastToRole(targetRole string, msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Printf("Broadcasting message to role %s: %+v", targetRole, msg)

	recipientCount := 0
	for conn, info := range s.connections {
		if info.role == targetRole {
			recipientCount++
			if err := conn.WriteJSON(msg); err != nil {
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
