package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthServer struct {
	healthpb.UnimplementedHealthServer
	db *pgxpool.Pool
}

func newHealthServer(pool *pgxpool.Pool) *HealthServer {
	return &HealthServer{
		db: pool,
	}
}

func (h *HealthServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if err := h.db.Ping(ctx); err != nil {
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_NOT_SERVING,
		}, status.Error(codes.Unavailable, "database connection failed")
	}

	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthServer) Watch(req *healthpb.HealthCheckRequest, stream healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented")
}
