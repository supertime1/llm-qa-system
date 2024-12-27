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

func main() {
	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "role=patient"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Create interrupt and done channels
	interrupt := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(interrupt, os.Interrupt)

	// Handle graceful shutdown
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
	fmt.Println("Type your question (press Ctrl+C to quit):")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
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
			continue
		}
	}
}
