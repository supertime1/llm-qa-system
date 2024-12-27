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

type PatientClient struct {
	conn      *websocket.Conn
	sessionID string
}

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	flag.Parse()

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	q := u.Query()
	q.Set("role", "patient")
	u.RawQuery = q.Encode()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	client := &PatientClient{conn: c}

	// Get session ID from server
	var sessionResp map[string]string
	if err := c.ReadJSON(&sessionResp); err != nil {
		log.Fatal("read session:", err)
	}
	client.sessionID = sessionResp["session_id"]
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
			case pb.MessageType_DOCTOR_MESSAGE:
				if msg := wsMsg.GetMessage(); msg != nil {
					fmt.Printf("\nDoctor: %s\n", msg.Content)
					fmt.Print("> ")
				}
			}
		}
	}()

	// Handle user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type your message (or 'quit' to exit):")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			return
		}

		if input == "" {
			continue
		}

		wsMsg := &pb.WebSocketMessage{
			Type: pb.MessageType_PATIENT_MESSAGE,
			Payload: &pb.WebSocketMessage_Message{
				Message: &pb.Message{
					Content:   input,
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
			return
		}
	}
}
