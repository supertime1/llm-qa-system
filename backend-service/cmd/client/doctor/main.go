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
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "role=doctor"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
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
			case int32(pb.MessageType_PATIENT_MESSAGE):
				if message, ok := msg["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						fmt.Printf("\nPatient: %s\n", content)
						fmt.Print("> ")
					}
				}
			case int32(pb.MessageType_AI_DRAFT_READY):
				if draft, ok := msg["ai_draft"].(map[string]interface{}); ok {
					fmt.Printf("\nAI Draft for question '%s':\n%s\n",
						draft["original_message"], draft["draft"])
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
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 3)
		switch parts[0] {
		case "review":
			if len(parts) < 2 {
				fmt.Println("Usage: review <accept|modify|reject> [content]")
				continue
			}

			msg := map[string]interface{}{
				"type": pb.MessageType_DRAFT_REVIEW.Number(),
				"review": map[string]interface{}{
					"action":    parts[1],
					"content":   strings.Join(parts[2:], " "),
					"timestamp": timestamppb.Now(),
				},
			}

			if err := c.WriteJSON(msg); err != nil {
				log.Printf("Error sending review: %v", err)
				continue
			}

		case "send":
			if len(parts) < 2 {
				fmt.Println("Usage: send <message>")
				continue
			}

			msg := map[string]interface{}{
				"type": pb.MessageType_DOCTOR_MESSAGE.Number(),
				"message": map[string]interface{}{
					"content":   strings.Join(parts[1:], " "),
					"timestamp": timestamppb.Now(),
				},
			}

			if err := c.WriteJSON(msg); err != nil {
				log.Printf("Error sending message: %v", err)
				continue
			}
		}
	}
}
