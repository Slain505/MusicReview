package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{DB: db}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}
