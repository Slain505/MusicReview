package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"MusicReview/internal/httpapi"
	"MusicReview/internal/sse"
	"MusicReview/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN env is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	st := store.New(pool)
	hub := sse.NewHub()
	api := httpapi.New(st, hub)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, api.Router()); err != nil {
		log.Fatal(err)
	}
}
