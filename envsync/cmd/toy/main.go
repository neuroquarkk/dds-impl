package main

import (
	"context"
	"envsync/internal/cfg"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	var (
		filename string
		pidFile  string
	)

	flag.StringVar(&filename, "env-file", ".env", "path to env file")
	flag.StringVar(&pidFile, "pid-file", "app.pid", "path to write this process's pid")
	flag.Parse()

	// write this application's PID to a file
	// envsync uses this to determine which PID to send the reload signal to
	if err := os.WriteFile(
		pidFile,
		[]byte(strconv.Itoa(os.Getpid())),
		0644,
	); err != nil {
		log.Fatalf("failed to write pid file: %v\n", err)
	}

	// envsync starts after this container so .env won't exists on a cold start
	// wait for sidecar's first fetch before trying to load config
	if err := waitForFile(filename, 40*time.Second); err != nil {
		log.Fatalf("startup failed: %v\n", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGHUP)

	loadEnv(filename)

	log.Println("toyapp started")
	for {
		select {
		case <-reload:
			log.Println("received SIGHUP, reloading config")
			loadEnv(filename)

		case <-ctx.Done():
			return
		}
	}
}

func waitForFile(filename string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(filename); err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s to appear", filename)
}

func loadEnv(filename string) {
	raw, err := godotenv.Read(filename)
	if err != nil {
		log.Printf("failed to read env file: %v\n", err)
		return
	}

	parsed, err := cfg.Parse(raw)
	if err != nil {
		log.Printf("failed to parse config: %v\n", err)
		return
	}

	log.Printf("config loaded: %+v\n", parsed)
}
