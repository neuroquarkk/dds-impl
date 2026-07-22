package poller

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"envsync/internal/cfg"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Poller struct {
	DB          *pgxpool.Pool
	Filename    string
	PidFile     string
	Interval    time.Duration
	CurrentHash string
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	log.Println("envsync started")
	p.executeTick(ctx)

	for {
		select {
		case <-ticker.C:
			p.executeTick(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *Poller) executeTick(ctx context.Context) {
	var hash string

	err := p.DB.
		QueryRow(ctx, "SELECT hash FROM config_metadata LIMIT 1").
		Scan(&hash)
	if err != nil {
		log.Printf("failed to query config metadata: %v\n", err)
		return
	}

	if hash == p.CurrentHash {
		return
	}

	rows, err := p.DB.Query(ctx, "SELECT key, value FROM configs")
	if err != nil {
		log.Printf("failed to query configs: %v\n", err)
		return
	}
	defer rows.Close()

	rawCfg := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			log.Printf("failed to scan row: %v\n", err)
			return
		}
		rawCfg[k] = v
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v\n", err)
		return
	}

	cfg, err := cfg.Parse(rawCfg)
	if err != nil {
		log.Printf("failed to parse config: %v\n", err)
		return
	}

	if err := saveFile(p.Filename, cfg); err != nil {
		log.Printf("failed to save file: %v\n", err)
		return
	}

	if err := p.signalReload(); err != nil {
		log.Printf("failed to signal app: %v\n", err)
	}

	p.CurrentHash = hash
	log.Println("config updated successfully")
}

func (p *Poller) signalReload() error {
	data, err := os.ReadFile(p.PidFile)
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("parse pid: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := proc.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("send signal: %w", err)
	}

	return nil
}

func saveFile(filename string, data cfg.Cfg) error {
	v := reflect.ValueOf(data)
	t := v.Type()

	var b strings.Builder
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("env")
		if key == "" {
			continue
		}
		fmt.Fprintf(&b, "%s=%v\n", key, v.Field(i).Interface())
	}

	temp := filename + ".tmp"
	if err := os.WriteFile(temp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(temp, filename)
}
