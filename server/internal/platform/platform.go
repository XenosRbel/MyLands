// Package platform provides infrastructure: configuration, logging, database
// pool, schema migrations and the injectable clock.
package platform

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config is read from environment variables only (no secrets in code).
type Config struct {
	Env             string // development | test | production
	HTTPAddr        string
	DatabaseURL     string
	PacksDir        string
	ClientDir       string // optional: serve built client
	WorkerEnabled   bool
	WorkerInterval  time.Duration
	AllowedOrigins  []string
	AdminBootstrap  string // login granted admin at startup (development only unless explicitly set)
	SessionTTL      time.Duration
	RateLimitRPS    float64
	RateLimitBurst  int
	DevPlatformAuth bool   // accept unsigned "dev" platform identities (never in production)
	MetricsToken    string // if set, /metrics requires "Authorization: Bearer <token>"; in production /metrics is disabled without it
	TrustProxy      bool   // use the left-most X-Forwarded-For address for per-IP limits (only behind a trusted proxy)
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// LoadConfig reads configuration; production refuses unsafe settings.
func LoadConfig() (Config, error) {
	c := Config{
		Env:            env("RF_ENV", "development"),
		HTTPAddr:       env("RF_HTTP_ADDR", ":8080"),
		DatabaseURL:    env("RF_DATABASE_URL", "postgres://postgres@127.0.0.1:5432/realmforge?sslmode=disable"),
		PacksDir:       env("RF_PACKS_DIR", "../packs"),
		ClientDir:      env("RF_CLIENT_DIR", ""),
		AdminBootstrap: env("RF_ADMIN_LOGIN", ""),
		MetricsToken:   env("RF_METRICS_TOKEN", ""),
		TrustProxy:     env("RF_TRUST_PROXY", "") == "true",
	}
	c.WorkerEnabled = env("RF_WORKER", "true") == "true"
	iv, err := time.ParseDuration(env("RF_WORKER_INTERVAL", "1s"))
	if err != nil {
		return c, fmt.Errorf("RF_WORKER_INTERVAL: %w", err)
	}
	c.WorkerInterval = iv
	ttl, err := time.ParseDuration(env("RF_SESSION_TTL", "720h"))
	if err != nil {
		return c, fmt.Errorf("RF_SESSION_TTL: %w", err)
	}
	c.SessionTTL = ttl
	c.RateLimitRPS, _ = strconv.ParseFloat(env("RF_RATE_RPS", "20"), 64)
	c.RateLimitBurst, _ = strconv.Atoi(env("RF_RATE_BURST", "60"))
	for _, o := range strings.Split(env("RF_ALLOWED_ORIGINS", "http://localhost:5173,https://localhost,capacitor://localhost"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, o)
		}
	}
	c.DevPlatformAuth = env("RF_DEV_PLATFORM_AUTH", map[bool]string{true: "true", false: "false"}[c.Env != "production"]) == "true"
	if c.Env == "production" {
		if c.DevPlatformAuth {
			return c, fmt.Errorf("RF_DEV_PLATFORM_AUTH must be false in production")
		}
		if strings.Contains(c.DatabaseURL, "sslmode=disable") && env("RF_ALLOW_INSECURE_DB", "") != "true" {
			return c, fmt.Errorf("production database URL must use TLS (or set RF_ALLOW_INSECURE_DB=true for a private network)")
		}
	}
	return c, nil
}

// NewLogger returns a structured JSON logger (text in development).
func NewLogger(env string) *slog.Logger {
	if env == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// OpenDB creates a connection pool and verifies connectivity.
func OpenDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// Migrate applies pending *.sql migrations from fsys in lexical order. Each
// migration runs in its own transaction; a session advisory lock prevents two
// processes from migrating concurrently.
func Migrate(ctx context.Context, db *pgxpool.Pool, fsys fs.FS, log *slog.Logger) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(727001)`); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock(727001)`)
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sql, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, name); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		log.Info("migration applied", "version", name)
	}
	return nil
}

// Clock abstracts time so tests can control it. Server time is authoritative.
type Clock interface{ Now() time.Time }

type SystemClock struct{}

// Now returns UTC time truncated to milliseconds (the persisted precision).
func (SystemClock) Now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

// FakeClock is a manually advanced clock for tests.
type FakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func NewFakeClock(t time.Time) *FakeClock { return &FakeClock{t: t.UTC().Truncate(time.Millisecond)} }
func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d).Truncate(time.Millisecond)
	c.mu.Unlock()
}
func (c *FakeClock) Set(t time.Time) {
	c.mu.Lock()
	c.t = t.UTC().Truncate(time.Millisecond)
	c.mu.Unlock()
}
