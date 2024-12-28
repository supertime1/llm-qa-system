package main

import (
	"context"
	"llm-qa-system/backend-service/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx := context.Background()

	// Get configuration from environment variables
	dbURL := getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/medical_chat")
	llmServiceAddr := getEnvOrDefault("LLM_SERVICE_ADDR", "localhost:50051")
	kafkaBrokers := []string{getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")}

	// Initialize Kafka first with all topics
	kafkaConfig := server.KafkaConfig{
		Brokers: kafkaBrokers,
		Topics:  server.GetDefaultTopics(),
	}

	if err := server.InitializeKafka(kafkaConfig); err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}

	// Connect to database using pgx
	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Create server group
	serverGroup, err := server.NewServerGroup(dbpool, llmServiceAddr, kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create server group: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Start server
	log.Printf("Server starting on :8080")
	if err := serverGroup.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
