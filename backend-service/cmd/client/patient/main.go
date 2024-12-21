package main

import (
	"bufio"
	"context"
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

	// Generate UUIDs
	chatID := uuid.New()
	patientID := uuid.New()

	// Start chat
	err = stream.Send(&pb.ChatRequest{
		ChatId:   &pb.UUID{Value: chatID[:]},
		SenderId: &pb.UUID{Value: patientID[:]},
		Type:     pb.RequestType_START_CHAT,
		Role:     pb.Role_ROLE_PATIENT,
	})
	if err != nil {
		log.Fatalf("Failed to start chat: %v", err)
	}

	fmt.Printf("Chat started with ID: %s\n", chatID.String())
	fmt.Printf("Your ID: %s\n", patientID.String())

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

	// Create a buffered reader for stdin
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type your messages (press Ctrl+C to quit):")

	for {
		// Read until newline (including spaces)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		// Trim the trailing newline and any carriage return
		input = strings.TrimSpace(input)

		// Skip empty messages
		if input == "" {
			continue
		}

		err = stream.Send(&pb.ChatRequest{
			ChatId:   &pb.UUID{Value: chatID[:]},
			SenderId: &pb.UUID{Value: patientID[:]},
			Content:  input,
			Type:     pb.RequestType_SEND_MESSAGE,
			Role:     pb.Role_ROLE_PATIENT,
		})
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}
