package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"envsync/internal/db"
	"envsync/internal/poller"
)

func main() {
	var (
		dsn      string
		pidFile  string
		filename string
	)

	flag.StringVar(
		&dsn,
		"dsn",
		"postgres://postgres:postgres@localhost:5432/envsync_db",
		"Postgres connection string",
	)
	flag.StringVar(&filename, "env-file", ".env", "path to env file")
	flag.StringVar(&pidFile, "pid-file", "app.pid", "path to app's pid file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pdb := db.PConn(ctx, dsn)

	poller := &poller.Poller{
		DB:       pdb,
		Filename: filename,
		PidFile:  pidFile,
		Interval: 10 * time.Second,
	}
	poller.Start(ctx)
}
