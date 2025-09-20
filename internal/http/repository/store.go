package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
)

type Store struct {
	DB      *pgxpool.Pool
	Queries *sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		DB:      pool,
		Queries: sqlc.New(pool),
	}
}
