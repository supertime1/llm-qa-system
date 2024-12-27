package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	pb "llm-qa-system/backend-service/src/proto"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DoctorClient struct {
	conn        *websocket.Conn
	latestDraft struct {
		originalMessage string
		draft           string
	}
}

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "role=doctor"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	client := &DoctorClient{conn: c}

	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		log.Println("\nReceived interrupt signal, closing connection...")
		err := c.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error during closing websocket: %v", err)
		}
		close(done)
		os.Exit(0)
	}()

	// Set up ping handler
	c.SetPingHandler(func(string) error {
		err := c.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second*10))
		if err != nil {
			log.Printf("Error sending pong: %v", err)
		}
		return nil
	})

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
					client.latestDraft.originalMessage = draft.MessageId
					client.latestDraft.draft = draft.Draft
					fmt.Printf("\nAI Draft ready:\n%s\n", draft.Draft)
					fmt.Println("Use 'review <accept|modify|reject> [content]' to review")
					fmt.Print("> ")
				}
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Commands:")
	fmt.Println("  review <accept|modify|reject> [content] - Review AI draft")
	fmt.Println("  send <message> - Send a message to patient")
	fmt.Println("Press Ctrl+C to quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

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

			var wsMsg pb.WebSocketMessage
			wsMsg.Type = pb.MessageType_DRAFT_REVIEW

			review := &pb.DraftReview{
				MessageId: client.latestDraft.originalMessage,
				Timestamp: timestamppb.Now(),
			}

			switch parts[1] {
			case "accept":
				review.Action = pb.ReviewAction_ACCEPT
				review.Content = client.latestDraft.draft
			case "modify":
				review.Action = pb.ReviewAction_MODIFY
				review.Content = strings.Join(parts[2:], " ")
			case "reject":
				review.Action = pb.ReviewAction_REJECT
			default:
				fmt.Println("Invalid action. Use accept, modify, or reject")
				continue
			}

			wsMsg.Payload = &pb.WebSocketMessage_Review{Review: review}

			jsonBytes, _ := marshaler.Marshal(&wsMsg)
			if err := c.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
				log.Printf("write error: %v", err)
			}
		}
	}
}
