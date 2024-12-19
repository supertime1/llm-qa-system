package server

import (
	"context"

	db "llm-qa-system/backend-service/src/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BaseServer struct {
	db  *pgxpool.Pool
	dbq *db.Queries
}

func NewBaseServer(pool *pgxpool.Pool) *BaseServer {
	return &BaseServer{
		db:  pool,
		dbq: db.New(pool),
	}
}

func (s *BaseServer) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
