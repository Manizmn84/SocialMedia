package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sikozonpc/social/internal/db"
	"github.com/sikozonpc/social/internal/env"
	"github.com/sikozonpc/social/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		return fmt.Errorf("connect seed database: %w", err)
	}
	defer conn.Close()

	storage := store.NewStorage(conn)
	return db.Seed(ctx, storage)
}
