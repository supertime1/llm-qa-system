package server

import (
	"context"
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
}

func NewServerGroup(pool *pgxpool.Pool, llmServiceAddr string, redisAddr string) (*ServerGroup, error) {
	baseServer := NewBaseServer(pool)

	// Create LLM client
	llmClient, err := NewLLMClient(llmServiceAddr)
	if err != nil {
		return nil, err
	}

	// Create WebSocket server
	wsServer := NewWebSocketServer(baseServer, llmClient)

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

func (s *ServerGroup) Stop() {
	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Cleanup database connection
	if s.db != nil {
		s.db.Close()
	}
}
