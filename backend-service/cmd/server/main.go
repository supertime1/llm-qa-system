package main

import (
	"context"
	"llm-qa-system/backend-service/server"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
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
	serverPort := getEnvOrDefault("SERVER_PORT", "50052")

	// Connect to database using pgx
	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Create server group
	serverGroup, err := server.NewServerGroup(dbpool, "", "")
	if err != nil {
		log.Fatalf("Failed to create server group: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	serverGroup.Register(grpcServer)

	// Start listening
	lis, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gracefully...")
		serverGroup.Stop()
		grpcServer.GracefulStop()
	}()

	// Start server
	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
