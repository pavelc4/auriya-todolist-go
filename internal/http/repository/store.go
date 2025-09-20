// internal/repository/store.go
package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
)

type Store struct {
	DB      *pgxpool.Pool
	Queries *db.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		DB:      pool,
		Queries: db.New(pool),
	}
}
