package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "llm-qa-system/backend-service/src/proto"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddr := "localhost:50052"
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMedicalChatServiceClient(conn)
	stream, err := client.ChatStream(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	fmt.Print("Enter chat ID to join: ")
	var chatIDStr string
	fmt.Scanln(&chatIDStr)
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		log.Fatalf("Invalid chat ID: %v", err)
	}

	doctorID := uuid.New()

	// Join chat
	err = stream.Send(&pb.ChatRequest{
		ChatId:   &pb.UUID{Value: chatID[:]},
		SenderId: &pb.UUID{Value: doctorID[:]},
		Type:     pb.RequestType_JOIN_CHAT,
		Role:     pb.Role_ROLE_DOCTOR,
	})
	if err != nil {
		log.Fatalf("Failed to join chat: %v", err)
	}

	fmt.Printf("Joined chat as doctor with ID: %s\n", doctorID.String())

	// Start goroutine to receive messages
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Printf("Stream closed: %v", err)
				os.Exit(1)
			}
			log.Printf("Received: %+v\n", resp)
		}
	}()

	// Read commands from stdin
	fmt.Println("Commands:")
	fmt.Println("  review <approve|reject|modify> [content] - Review AI draft")
	fmt.Println("  send <message> - Send a message")
	fmt.Println("Type your commands (press Ctrl+C to quit):")

	for {
		var input string
		fmt.Scanln(&input)
		parts := strings.SplitN(input, " ", 3)

		switch parts[0] {
		case "review":
			if len(parts) < 2 {
				fmt.Println("Usage: review <approve|reject|modify> [content]")
				continue
			}

			var status pb.ReviewStatus
			var content string

			switch parts[1] {
			case "approve":
				status = pb.ReviewStatus_APPROVED
				content = "This is a simulated AI response to your question. The doctor will review this shortly."
			case "reject":
				status = pb.ReviewStatus_REJECTED
			case "modify":
				if len(parts) < 3 {
					fmt.Println("Usage: review modify <content>")
					continue
				}
				status = pb.ReviewStatus_MODIFIED
				content = parts[2]
			default:
				fmt.Println("Invalid review status")
				continue
			}

			review := map[string]interface{}{
				"status":  status,
				"content": content,
			}
			reviewContent, _ := json.Marshal(review)

			err = stream.Send(&pb.ChatRequest{
				ChatId:   &pb.UUID{Value: chatID[:]},
				SenderId: &pb.UUID{Value: doctorID[:]},
				Content:  string(reviewContent),
				Type:     pb.RequestType_SUBMIT_REVIEW,
				Role:     pb.Role_ROLE_DOCTOR,
			})

		case "send":
			if len(parts) < 2 {
				fmt.Println("Usage: send <message>")
				continue
			}
			err = stream.Send(&pb.ChatRequest{
				ChatId:   &pb.UUID{Value: chatID[:]},
				SenderId: &pb.UUID{Value: doctorID[:]},
				Content:  parts[1],
				Type:     pb.RequestType_SEND_MESSAGE,
				Role:     pb.Role_ROLE_DOCTOR,
			})
		}

		if err != nil {
			log.Printf("Failed to send: %v", err)
		}
	}
}
