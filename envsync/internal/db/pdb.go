package db

import (
	"context"
	_ "embed"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

func PConn(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create new pool: %v\n", err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		log.Fatalf("failed to apply schema: %v\n", err)
	}

	log.Println("postgres connected successfully")
	return pool
}
