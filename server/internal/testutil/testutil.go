// Package testutil provides a real PostgreSQL database for integration tests.
//
// Each test package gets its own database (rf_test_<name>) recreated from the
// migrations, so packages can run in parallel. If PostgreSQL is unreachable
// the tests are skipped, unless RF_REQUIRE_DB=1 (CI / make test) in which
// case they fail.
package testutil

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/platform"
	"realmforge/server/migrations"
)

func adminURL() string {
	if u := os.Getenv("RF_TEST_ADMIN_URL"); u != "" {
		return u
	}
	return "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable"
}

// DB returns a freshly migrated, empty database pool for this test package.
func DB(t testing.TB, name string) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, adminURL())
	if err != nil {
		if os.Getenv("RF_REQUIRE_DB") == "1" {
			t.Fatalf("postgres required but unavailable: %v", err)
		}
		t.Skipf("postgres unavailable (%v); set RF_REQUIRE_DB=1 to fail instead", err)
	}
	defer admin.Close(ctx)
	db := "rf_test_" + name
	var exists bool
	if err := admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, db).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{db}.Sanitize()); err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Fatal(err)
		}
	}
	url := strings.Replace(adminURL(), "/postgres?", "/"+db+"?", 1)
	pool, err := platform.OpenDB(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatal(err)
	}
	if err := platform.Migrate(ctx, pool, migrations.FS, Logger()); err != nil {
		t.Fatal(fmt.Errorf("migrate: %w", err))
	}
	t.Cleanup(pool.Close)
	return pool
}

// Logger discards logs unless RF_TEST_LOG=1.
func Logger() *slog.Logger {
	if os.Getenv("RF_TEST_LOG") == "1" {
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
