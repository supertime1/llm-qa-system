package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Handle incoming messages
	go func() {
		for {
			var msg Message
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Println("read:", err)
				return
			}

			switch msg.Type {
			case "NEW_QUESTION":
				var payload struct {
					Content string `json:"content"`
				}
				json.Unmarshal(msg.Payload, &payload)
				fmt.Printf("\nNew patient question: %s\n", payload.Content)

			case "AI_DRAFT_READY":
				var payload struct {
					Question string `json:"question"`
					Draft    string `json:"draft"`
				}
				json.Unmarshal(msg.Payload, &payload)
				fmt.Printf("\nAI Draft ready for question: %s\nDraft: %s\n", payload.Question, payload.Draft)
				fmt.Println("Use 'review <accept|modify|reject> [modified_content]' to review")
			}
		}
	}()

	// Handle user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Commands:")
	fmt.Println("  review <accept|modify|reject> [content] - Review AI draft")
	fmt.Println("  send <message> - Send a regular message")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.SplitN(input, " ", 3)

		switch parts[0] {
		case "review":
			if len(parts) < 2 {
				fmt.Println("Usage: review <accept|modify|reject> [content]")
				continue
			}

			msg := Message{
				Type: "DOCTOR_REVIEW",
			}

			review := struct {
				Action          string `json:"action"`
				ModifiedContent string `json:"modified_content,omitempty"`
			}{
				Action: parts[1],
			}

			if parts[1] == "modify" || parts[1] == "reject" {
				if len(parts) < 3 {
					fmt.Println("Content required for modify/reject")
					continue
				}
				review.ModifiedContent = parts[2]
			}

			payload, _ := json.Marshal(review)
			msg.Payload = payload
			c.WriteJSON(msg)

		case "send":
			if len(parts) < 2 {
				fmt.Println("Usage: send <message>")
				continue
			}
			msg := Message{
				Type: "DOCTOR_MESSAGE",
				Payload: json.RawMessage(fmt.Sprintf(`{"content":%q}`,
					strings.Join(parts[1:], " "))),
			}
			c.WriteJSON(msg)
		}
	}
}
