package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	pb "llm-qa-system/backend-service/src/proto"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Helper function to format UUID bytes to string
func formatUUID(bytes []byte) string {
	if len(bytes) != 16 {
		return "invalid-uuid"
	}
	uuid := [16]byte{}
	copy(uuid[:], bytes)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16])
}

// Helper function to format chat response
func formatResponse(resp *pb.ChatResponse) string {
	var msgContent string
	switch payload := resp.Payload.(type) {
	case *pb.ChatResponse_Message:
		senderID := formatUUID(payload.Message.SenderId.Value)
		role := payload.Message.Role.String()
		msgContent = fmt.Sprintf("[%s][%s]: %s", role, senderID, payload.Message.Content)
	case *pb.ChatResponse_AiDraft:
		msgContent = fmt.Sprintf("[AI DRAFT] Confidence: %.2f\nContent: %s",
			payload.AiDraft.ConfidenceScore,
			payload.AiDraft.Content)
	case *pb.ChatResponse_Review:
		msgContent = fmt.Sprintf("[REVIEW] Status: %s\nContent: %s",
			payload.Review.Status.String(),
			payload.Review.ModifiedContent)
	}

	timestamp := time.Unix(resp.Timestamp.Seconds, int64(resp.Timestamp.Nanos))
	return fmt.Sprintf("\n=== Message at %s ===\n%s\n",
		timestamp.Format("2006-01-02 15:04:05"),
		msgContent)
}

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

	// Create a buffered reader for chat ID input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter chat ID to join: ")
	chatIDStr, _ := reader.ReadString('\n')
	chatIDStr = strings.TrimSpace(chatIDStr)

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
			fmt.Print(formatResponse(resp))
		}
	}()

	fmt.Println("Commands:")
	fmt.Println("  review <approve|reject|modify> [content] - Review AI draft")
	fmt.Println("  send <message> - Send a message")
	fmt.Println("Type your commands (press Ctrl+C to quit):")

	for {
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
				fmt.Println("Usage: review <approve|reject|modify> [content]")
				continue
			}

			var status pb.ReviewStatus
			var content string

			switch parts[1] {
			case "approve":
				status = pb.ReviewStatus_APPROVED
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
			message := strings.Join(parts[1:], " ") // Join all parts after "send" as the message
			err = stream.Send(&pb.ChatRequest{
				ChatId:   &pb.UUID{Value: chatID[:]},
				SenderId: &pb.UUID{Value: doctorID[:]},
				Content:  message,
				Type:     pb.RequestType_SEND_MESSAGE,
				Role:     pb.Role_ROLE_DOCTOR,
			})
		}

		if err != nil {
			log.Printf("Failed to send: %v", err)
		}
	}
}
