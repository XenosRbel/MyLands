// Package game implements the authoritative game modules (accounts, worlds,
// cities, construction, research, army, missions, combat, heroes, market,
// alliances, quests, messaging, notifications, administration) on top of
// PostgreSQL. All state changes happen inside transactions that follow the
// lock order documented in docs/spec/ARCHITECTURE.md §5:
//
//	players (ascending id) → cities (ascending id) → missions → tiles → market orders
//
// Every change to a player's economic inputs happens while that player's row
// is locked, at the due time of a scheduled event or at "now".
package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
)

// Service is the game application layer.
type Service struct {
	DB    *pgxpool.Pool
	Clock platform.Clock
	Log   *slog.Logger
	packs sync.Map // worldID -> *pack.Loaded
	// spawnRing remembers per world the ring where the last capital was placed
	// (worldID -> int). It is only a hint for the search order, never for correctness.
	spawnRing sync.Map
	Metrics   Metrics
}

// Metrics are simple counters exported in Prometheus text format.
type Metrics struct {
	EventsProcessed atomic.Int64
	EventsFailed    atomic.Int64
	TxRetries       atomic.Int64
	Commands        atomic.Int64
	IdemReplays     atomic.Int64
}

func New(db *pgxpool.Pool, clock platform.Clock, log *slog.Logger) *Service {
	return &Service{DB: db, Clock: clock, Log: log}
}

// Error is a domain error with a stable machine-readable code.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func errf(status int, code, format string, a ...any) *Error {
	return &Error{Status: status, Code: code, Message: fmt.Sprintf(format, a...)}
}

func invalid(format string, a ...any) *Error {
	return errf(http.StatusBadRequest, "invalid_request", format, a...)
}
func notFound(what string) *Error {
	return errf(http.StatusNotFound, "not_found", "%s not found", what)
}
func forbidden(format string, a ...any) *Error {
	return errf(http.StatusForbidden, "forbidden", format, a...)
}
func conflict(code, format string, a ...any) *Error {
	return errf(http.StatusConflict, code, format, a...)
}
func rule(code, format string, a ...any) *Error {
	return errf(http.StatusUnprocessableEntity, code, format, a...)
}

// AsError extracts a domain error.
func AsError(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

func retryable(err error) bool {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code == "40001" || pg.Code == "40P01" // serialization failure, deadlock
	}
	return false
}

// InTx runs fn in a READ COMMITTED transaction, retrying on deadlock or
// serialization failure (bounded). Domain errors roll back and are returned.
func (s *Service) InTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	var err error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			s.Metrics.TxRetries.Add(1)
			time.Sleep(time.Duration(attempt*attempt*15) * time.Millisecond)
		}
		err = s.runTx(ctx, fn)
		if err == nil || !retryable(err) {
			return err
		}
	}
	return err
}

func (s *Service) runTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Pack returns the pinned pack of a world (cached; immutable per world+hash).
func (s *Service) Pack(ctx context.Context, q querier, worldID int64) (*pack.Loaded, error) {
	var hash string
	var speed float64
	// Read only the hash and speed first. pack_json is tens of kilobytes and
	// this function runs several times per request, so fetch it on a cache miss only.
	if err := q.QueryRow(ctx, `SELECT pack_hash, speed FROM worlds WHERE id=$1`, worldID).Scan(&hash, &speed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFound("world")
		}
		return nil, err
	}
	if v, ok := s.packs.Load(worldID); ok {
		if c := v.(cachedPack); c.hash == hash && c.speed == speed {
			return c.p, nil
		}
	}
	var raw []byte
	if err := q.QueryRow(ctx, `SELECT pack_json FROM worlds WHERE id=$1 AND pack_hash=$2`, worldID, hash).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFound("world") // the pack changed between the two reads; the caller retries
		}
		return nil, err
	}
	l, err := pack.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("world %d pack corrupt: %w", worldID, err)
	}
	// The world speed multiplies the pack's worldSpeed. The hash stays the
	// pack's own hash; the world row records the multiplier.
	l.Params.WorldSpeed *= speed
	s.packs.Store(worldID, cachedPack{p: l, hash: hash, speed: speed})
	return l, nil
}

type cachedPack struct {
	p     *pack.Loaded
	hash  string
	speed float64
}

// querier is satisfied by pgx.Tx and *pgxpool.Pool.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func pgCode(err error) string {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code
	}
	return ""
}

type contextT = context.Context

// PCtx is the exported handle of a locked, settled player context used by the API layer.
type PCtx = *pctx

// PlayerID returns the player of an account in a world (forbidden if not joined).
func (s *Service) PlayerID(ctx context.Context, q querier, accountID, worldID int64) (int64, error) {
	return s.playerFor(ctx, q, accountID, worldID)
}
