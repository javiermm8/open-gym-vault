// internal/persistence/store.go
package persistence

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/javiermm8/open-gym-vault/internal/db"
)

type Store struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}

	cfg.MaxConns = envInt32("DB_POOL_MAX_CONNS", 10)
	cfg.MinConns = envInt32("DB_POOL_MIN_CONNS", 1)
	cfg.MaxConnLifetime = envDuration("DB_POOL_CONN_MAX_LIFETIME", time.Hour)
	cfg.MaxConnIdleTime = envDuration("DB_POOL_CONN_IDLE_TIME", 30*time.Minute)
	cfg.HealthCheckPeriod = envDuration("DB_POOL_HEALTH_CHECK_PERIOD", time.Minute)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// ping db
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &Store{
		pool:    pool,
		Queries: db.New(pool),
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func envInt32(name string, def int32) int32 {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			return int32(n)
		}
		log.Printf("ignoring invalid %s=%q, using default %d", name, v, def)
	}
	return def
}

func envDuration(name string, def time.Duration) time.Duration {
	if v := os.Getenv(name); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		log.Printf("ignoring invalid %s=%q, using default %s", name, v, def)
	}
	return def
}
