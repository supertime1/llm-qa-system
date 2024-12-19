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

func main() {
	ctx := context.Background()

	// Connect to database using pgx
	dbpool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Create server group
	serverGroup, err := server.NewServerGroup(dbpool, "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create server group: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	serverGroup.Register(grpcServer)

	// Start listening
	lis, err := net.Listen("tcp", ":50052")
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
