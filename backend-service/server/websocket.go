package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	*BaseServer
	llmClient *LLMClient
}

func NewWebSocketServer(base *BaseServer, llmClient *LLMClient) *WebSocketServer {
	return &WebSocketServer{
		BaseServer: base,
		llmClient:  llmClient,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type PatientMessage struct {
	Content string `json:"content"`
}

type DraftReview struct {
	Action          string `json:"action"` // "accept", "modify", "reject"
	ModifiedContent string `json:"modified_content,omitempty"`
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		var wsMsg WebSocketMessage
		if err := conn.ReadJSON(&wsMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch wsMsg.Type {
		// case "PATIENT_QUESTION":
		// 	var msg PatientMessage
		// 	if err := json.Unmarshal(wsMsg.Payload, &msg); err != nil {
		// 		continue
		// 	}

		// 	// Forward question to doctor immediately
		// 	s.sendResponse(conn, "NEW_QUESTION", map[string]string{
		// 		"content": msg.Content,
		// 	})

		// 	// Generate AI draft using LLM service
		// 	draftResp, err := s.llmClient.GenerateDraftAnswer(r.Context(), &pb.QuestionRequest{
		// 		QuestionText: msg.Content,
		// 	})
		// 	if err != nil {
		// 		s.sendErrorResponse(conn, "Failed to generate draft")
		// 		continue
		// 	}

		// 	// Send draft to doctor for review
		// 	s.sendResponse(conn, "AI_DRAFT_READY", map[string]interface{}{
		// 		"question": msg.Content,
		// 		"draft":    draftResp.DraftAnswer,
		// 	})
		case "PATIENT_QUESTION":
			var msg PatientMessage
			if err := json.Unmarshal(wsMsg.Payload, &msg); err != nil {
				continue
			}

			// Forward question to doctor immediately
			s.sendResponse(conn, "NEW_QUESTION", map[string]string{
				"content": msg.Content,
			})

			// Simulate LLM response for now
			// TODO: Replace with actual async LLM call + message queue
			draftAnswer := "This is a hardcoded AI draft answer for: " + msg.Content

			// Send draft to doctor for review
			s.sendResponse(conn, "AI_DRAFT_READY", map[string]interface{}{
				"question": msg.Content,
				"draft":    draftAnswer,
			})

		case "DOCTOR_REVIEW":
			var review DraftReview
			if err := json.Unmarshal(wsMsg.Payload, &review); err != nil {
				continue
			}

			var responseType string
			var content string

			switch review.Action {
			case "accept":
				responseType = "ACCEPTED_DRAFT"
				content = review.ModifiedContent // Original draft content
			case "modify":
				responseType = "MODIFIED_DRAFT"
				content = review.ModifiedContent // Modified draft content
			case "reject":
				responseType = "DOCTOR_ANSWER"
				content = review.ModifiedContent // Doctor's new answer
			}

			// Send final answer to patient
			s.sendResponse(conn, responseType, map[string]string{
				"content": content,
			})

		case "DOCTOR_MESSAGE":
			var msg PatientMessage
			if err := json.Unmarshal(wsMsg.Payload, &msg); err != nil {
				continue
			}

			// Forward regular doctor message to patient
			s.sendResponse(conn, "NEW_MESSAGE", map[string]string{
				"content": msg.Content,
			})
		}
	}
}

func (s *WebSocketServer) sendResponse(conn *websocket.Conn, msgType string, payload interface{}) {
	response := WebSocketMessage{
		Type:    msgType,
		Payload: marshal(payload),
	}
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func (s *WebSocketServer) sendErrorResponse(conn *websocket.Conn, message string) {
	s.sendResponse(conn, "ERROR", map[string]string{"message": message})
}

func marshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return []byte("{}")
	}
	return data
}
