package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerGroup struct {
	wsServer     *WebSocketServer
	healthServer *HealthServer
	db           *pgxpool.Pool
	httpServer   *http.Server
	llmClient    *LLMClient // Add this field

}

func NewServerGroup(pool *pgxpool.Pool, llmServiceAddr string, kafkaBrokers []string) (*ServerGroup, error) {
	baseServer := NewBaseServer(pool)

	// Create LLM client
	llmClient, err := NewLLMClient(llmServiceAddr, kafkaBrokers)
	if err != nil {
		return nil, err
	}

	// Create WebSocket server
	wsServer := NewWebSocketServer(baseServer, llmClient, kafkaBrokers)

	// Create HTTP server
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    ":8080", // You might want to make this configurable
		Handler: mux,
	}

	sg := &ServerGroup{
		db:           pool,
		wsServer:     wsServer,
		healthServer: newHealthServer(pool),
		httpServer:   httpServer,
		llmClient:    llmClient, // Store the client

	}

	// Set up WebSocket route
	mux.HandleFunc("/ws", wsServer.HandleWebSocket)

	return sg, nil
}

func (s *ServerGroup) Register(grpcServer *grpc.Server) {
	healthpb.RegisterHealthServer(grpcServer, s.healthServer)
	reflection.Register(grpcServer)
}

func (s *ServerGroup) Start(ctx context.Context) error {
	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return s.httpServer.Shutdown(context.Background())
}

func (sg *ServerGroup) Shutdown(ctx context.Context) error {
	var errs []error

	// Close WebSocket server first (this will close Kafka consumer)
	if err := sg.wsServer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("websocket server close error: %v", err))
	}

	// Close LLM client (this will close Kafka producer)
	if err := sg.llmClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("llm client close error: %v", err))
	}

	// Close HTTP server
	if err := sg.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http server shutdown error: %v", err))
	}

	// Close DB connection
	sg.db.Close()

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
