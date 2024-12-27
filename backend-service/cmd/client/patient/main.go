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

	// Handle incoming messages
	go func() {
		for {
			var msg map[string]interface{}
			if err := c.ReadJSON(&msg); err != nil {
				log.Printf("read error: %v", err)
				return
			}

			msgType, _ := msg["type"].(float64)

			switch int32(msgType) {
			case int32(pb.MessageType_DOCTOR_MESSAGE):
				if message, ok := msg["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						fmt.Printf("\nDoctor: %s\n", content)
						fmt.Print("> ")
					}
				}
			}
		}
	}()

	// Handle user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type your question (press Ctrl+C to quit):")

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		msg := map[string]interface{}{
			"type": pb.MessageType_PATIENT_MESSAGE.Number(),
			"message": map[string]interface{}{
				"content":   input,
				"timestamp": timestamppb.Now(),
			},
		}

		log.Printf("Sending message structure: %+v", msg)

		if err := c.WriteJSON(msg); err != nil {
			log.Printf("Error sending message: %v", err)
			continue
		}
	}
}
