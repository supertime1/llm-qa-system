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
			case "NEW_MESSAGE", "ACCEPTED_DRAFT", "MODIFIED_DRAFT", "DOCTOR_ANSWER":
				var payload struct {
					Content string `json:"content"`
				}
				json.Unmarshal(msg.Payload, &payload)
				fmt.Printf("\nDoctor: %s\n", payload.Content)
			}
		}
	}()

	// Handle user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type your question (press Ctrl+C to quit):")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input != "" {
			msg := Message{
				Type:    "PATIENT_QUESTION",
				Payload: json.RawMessage(fmt.Sprintf(`{"content":%q}`, input)),
			}
			c.WriteJSON(msg)
		}
	}
}
