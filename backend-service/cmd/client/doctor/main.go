package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DoctorClient struct {
	conn        *websocket.Conn
	sessionID   string
	latestDraft *pb.AIDraftReady
}

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	sessionID := flag.String("session", "", "session ID to join")
	token := flag.String("token", "doctor123", "doctor authentication token")
	flag.Parse()

	if *sessionID == "" {
		log.Fatal("session ID is required")
	}

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	q := u.Query()
	q.Set("role", "doctor")
	q.Set("session", *sessionID)
	q.Set("token", *token)
	u.RawQuery = q.Encode()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	client := &DoctorClient{
		conn:      c,
		sessionID: *sessionID,
	}

	fmt.Printf("Connected to session: %s\n", client.sessionID)

	// Configure protojson
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}

	// Handle incoming messages
	go func() {
		for {
			_, rawMsg, err := c.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			var wsMsg pb.WebSocketMessage
			if err := unmarshaler.Unmarshal(rawMsg, &wsMsg); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}

			switch wsMsg.Type {
			case pb.MessageType_PATIENT_MESSAGE:
				if msg := wsMsg.GetMessage(); msg != nil {
					fmt.Printf("\nPatient: %s\n", msg.Content)
					fmt.Print("> ")
				}
			case pb.MessageType_AI_DRAFT_READY:
				if draft := wsMsg.GetAiDraft(); draft != nil {
					client.latestDraft = draft
					fmt.Printf("\nAI Draft ready:\n%s\n", draft.Draft)
					fmt.Println("Use 'review <accept|modify|reject> [content]' to review")
					fmt.Print("> ")
				}
			}
		}
	}()

	// Handle commands
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Commands: review <accept|modify|reject> [content], send <message>, quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			return
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "review":
			if len(parts) < 2 {
				fmt.Println("Usage: review <accept|modify|reject> [content]")
				continue
			}

			if client.latestDraft == nil {
				fmt.Println("No draft available to review")
				continue
			}

			var wsMsg pb.WebSocketMessage
			wsMsg.Type = pb.MessageType_DRAFT_REVIEW

			review := &pb.DraftReview{
				MessageId: client.latestDraft.MessageId,
				Timestamp: timestamppb.Now(),
			}

			switch parts[1] {
			case "accept":
				review.Action = pb.ReviewAction_ACCEPT
				review.Content = client.latestDraft.Draft
			case "modify":
				if len(parts) < 3 {
					fmt.Println("Content required for modify")
					continue
				}
				review.Action = pb.ReviewAction_MODIFY
				review.Content = strings.Join(parts[2:], " ")
			case "reject":
				review.Action = pb.ReviewAction_REJECT
			default:
				fmt.Println("Invalid action. Use accept, modify, or reject")
				continue
			}

			wsMsg.Payload = &pb.WebSocketMessage_Review{Review: review}

			jsonBytes, err := marshaler.Marshal(&wsMsg)
			if err != nil {
				log.Printf("marshal error: %v", err)
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				log.Printf("write error: %v", err)
				continue
			}

		case "send":
			if len(parts) < 2 {
				fmt.Println("Usage: send <message>")
				continue
			}

			wsMsg := &pb.WebSocketMessage{
				Type: pb.MessageType_DOCTOR_MESSAGE,
				Payload: &pb.WebSocketMessage_Message{
					Message: &pb.Message{
						Content:   strings.Join(parts[1:], " "),
						Timestamp: timestamppb.Now(),
					},
				},
			}

			jsonBytes, err := marshaler.Marshal(wsMsg)
			if err != nil {
				log.Printf("marshal error: %v", err)
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				log.Printf("write error: %v", err)
				continue
			}
		}
	}
}
