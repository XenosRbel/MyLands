// Package api exposes the game over a versioned REST/JSON API (/api/v1).
// See docs/spec/API.md and api/openapi.yaml.
package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/game"
	"realmforge/server/internal/platform"
)

type Server struct {
	G       *game.Service
	Cfg     platform.Config
	Log     *slog.Logger
	limiter *limiter
	mux     *http.ServeMux
	packsFS http.Handler
}

func New(g *game.Service, cfg platform.Config, log *slog.Logger) *Server {
	s := &Server{G: g, Cfg: cfg, Log: log, limiter: newLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst), mux: http.NewServeMux()}
	s.routes()
	return s
}

const maxBody = 64 << 10

type ctxKey int

const (
	keyAccount ctxKey = iota
	keyReqID
)

func account(r *http.Request) *game.Account {
	a, _ := r.Context().Value(keyAccount).(*game.Account)
	return a
}

// ServeHTTP applies global middleware.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rid := r.Header.Get("X-Request-ID")
	if rid == "" || len(rid) > 64 {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		rid = hex.EncodeToString(b)
	}
	w.Header().Set("X-Request-ID", rid)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	s.cors(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	rec := &statusRecorder{ResponseWriter: w, status: 200}
	defer func() {
		if p := recover(); p != nil {
			s.Log.Error("panic", "err", p, "path", r.URL.Path, "rid", rid)
			writeErr(rec, &game.Error{Status: 500, Code: "internal", Message: "internal error"})
		}
		if r.URL.Path != "/healthz" {
			s.Log.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", time.Since(start).Milliseconds(), "rid", rid)
		}
	}()
	r.Body = http.MaxBytesReader(rec, r.Body, maxBody)
	r = r.WithContext(context.WithValue(r.Context(), keyReqID, rid))
	s.mux.ServeHTTP(rec, r)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(c int) { s.status = c; s.ResponseWriter.WriteHeader(c) }

func (s *Server) cors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}
	for _, o := range s.Cfg.AllowedOrigins {
		if o == "*" || o == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, e *game.Error) {
	writeJSON(w, e.Status, map[string]any{"error": e})
}

func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := game.AsError(err); ok {
		writeErr(w, e)
		return
	}
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		writeErr(w, &game.Error{Status: 413, Code: "body_too_large", Message: "request body too large"})
		return
	}
	s.Log.Error("request failed", "err", err, "path", r.URL.Path, "rid", r.Context().Value(keyReqID))
	writeErr(w, &game.Error{Status: 500, Code: "internal", Message: "internal error"})
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return err
		}
		if errors.Is(err, io.EOF) {
			return &game.Error{Status: 400, Code: "invalid_request", Message: "request body required"}
		}
		return &game.Error{Status: 400, Code: "invalid_json", Message: err.Error()}
	}
	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || v <= 0 {
		return 0, &game.Error{Status: 400, Code: "invalid_request", Message: "invalid " + name}
	}
	return v, nil
}

func queryInt(r *http.Request, name string, def int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		return def
	}
	return v
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if t, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return ""
}

// auth wraps a handler requiring a valid session; also rate limits per account.
func (s *Server) auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := s.G.Authenticate(r.Context(), bearer(r))
		if err != nil {
			s.fail(w, r, err)
			return
		}
		if !s.limiter.allow("a:" + strconv.FormatInt(a.ID, 10)) {
			w.Header().Set("Retry-After", "1")
			writeErr(w, &game.Error{Status: 429, Code: "rate_limited", Message: "too many requests"})
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), keyAccount, a)))
	}
}

func (s *Server) admin(h http.HandlerFunc) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request) {
		if !account(r).IsAdmin {
			writeErr(w, &game.Error{Status: 403, Code: "forbidden", Message: "admin only"})
			return
		}
		h(w, r)
	})
}

// ipLimited rate limits unauthenticated endpoints per client IP.
func (s *Server) ipLimited(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if i := strings.LastIndex(ip, ":"); i > 0 {
			ip = ip[:i]
		}
		if s.Cfg.TrustProxy {
			if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
				ip = strings.TrimSpace(strings.Split(xf, ",")[0])
			}
		}
		if !s.limiter.allow("ip:" + ip) {
			w.Header().Set("Retry-After", "1")
			writeErr(w, &game.Error{Status: 429, Code: "rate_limited", Message: "too many requests"})
			return
		}
		h(w, r)
	}
}

var reIdemKey = regexp.MustCompile(`^[A-Za-z0-9_-]{8,80}$`)

// command executes a state-changing game command exactly once per
// Idempotency-Key: the key row is inserted inside the same transaction as the
// command's effects. A concurrent duplicate blocks on the key's unique index
// until the first commits, then replays the stored response. A reused key
// with a different request body is rejected (422). Domain errors are stored
// too, so a retry observes the same outcome.
func (s *Server) command(w http.ResponseWriter, r *http.Request, body []byte, fn func(tx pgx.Tx) (any, error)) {
	key := r.Header.Get("Idempotency-Key")
	if !reIdemKey.MatchString(key) {
		writeErr(w, &game.Error{Status: 400, Code: "idempotency_key_required", Message: "Idempotency-Key header (8-80 chars [A-Za-z0-9_-]) is required"})
		return
	}
	acc := account(r)
	sum := sha256.Sum256(append([]byte(r.Method+" "+r.URL.Path+"\n"), body...))
	hash := hex.EncodeToString(sum[:])
	s.G.Metrics.Commands.Add(1)
	var status int
	var resp []byte
	err := s.G.InTx(r.Context(), func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(), `INSERT INTO idempotency_keys(account_id, key, request_hash) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, acc.ID, key, hash)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			var prevHash string
			var code int
			var stored []byte
			if err := tx.QueryRow(r.Context(), `SELECT request_hash, status_code, response FROM idempotency_keys WHERE account_id=$1 AND key=$2`, acc.ID, key).Scan(&prevHash, &code, &stored); err != nil {
				return err
			}
			if prevHash != hash {
				return &game.Error{Status: 422, Code: "idempotency_key_reused", Message: "Idempotency-Key was used for a different request"}
			}
			s.G.Metrics.IdemReplays.Add(1)
			status, resp = code, stored
			w.Header().Set("Idempotent-Replay", "true")
			return nil
		}
		// Savepoint so a domain error rolls back the command's effects but
		// keeps the key row, letting us store the error outcome.
		if _, err := tx.Exec(r.Context(), `SAVEPOINT cmd`); err != nil {
			return err
		}
		out, cerr := fn(tx)
		if cerr != nil {
			ge, ok := game.AsError(cerr)
			if !ok {
				return cerr // infrastructure error: roll back everything, key is released for retry
			}
			if _, err := tx.Exec(r.Context(), `ROLLBACK TO SAVEPOINT cmd`); err != nil {
				return err
			}
			status = ge.Status
			resp, _ = json.Marshal(map[string]any{"error": ge})
		} else {
			status = 200
			resp, _ = json.Marshal(map[string]any{"data": out})
		}
		_, err = tx.Exec(r.Context(), `UPDATE idempotency_keys SET status_code=$3, response=$4 WHERE account_id=$1 AND key=$2`, acc.ID, key, status, resp)
		return err
	})
	if err != nil {
		s.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(resp)
}

// cmd decodes a JSON body into T and runs a command.
func cmd[T any](s *Server, w http.ResponseWriter, r *http.Request, fn func(tx pgx.Tx, req T) (any, error)) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var req T
	if len(strings.TrimSpace(string(body))) == 0 {
		body = []byte("{}")
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeErr(w, &game.Error{Status: 400, Code: "invalid_json", Message: err.Error()})
		return
	}
	s.command(w, r, body, func(tx pgx.Tx) (any, error) { return fn(tx, req) })
}

// read runs a read transaction (which may settle due events) and writes data.
func (s *Server) read(w http.ResponseWriter, r *http.Request, fn func(tx pgx.Tx) (any, error)) {
	var out any
	err := s.G.InTx(r.Context(), func(tx pgx.Tx) error {
		var err error
		out, err = fn(tx)
		return err
	})
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"data": out})
}

// ---------- rate limiter (in-memory token bucket; per instance) ----------

type bucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu      sync.Mutex
	rps     float64
	burst   float64
	buckets map[string]*bucket
	sweep   time.Time
}

func newLimiter(rps float64, burst int) *limiter {
	return &limiter{rps: rps, burst: float64(burst), buckets: map[string]*bucket{}}
}

func (l *limiter) allow(key string) bool {
	if l.rps <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if now.Sub(l.sweep) > time.Minute {
		for k, b := range l.buckets {
			if now.Sub(b.last) > 10*time.Minute {
				delete(l.buckets, k)
			}
		}
		l.sweep = now
	}
	b := l.buckets[key]
	if b == nil {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens = math.Min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rps)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.MetricsToken != "" {
		if bearer(r) != s.Cfg.MetricsToken {
			writeErr(w, &game.Error{Status: 401, Code: "unauthenticated", Message: "metrics token required"})
			return
		}
	} else if s.Cfg.Env == "production" {
		http.NotFound(w, r)
		return
	}
	m := &s.G.Metrics
	st := s.G.DB.Stat()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "rf_events_processed_total %d\nrf_events_failed_total %d\nrf_tx_retries_total %d\nrf_commands_total %d\nrf_idempotent_replays_total %d\n",
		m.EventsProcessed.Load(), m.EventsFailed.Load(), m.TxRetries.Load(), m.Commands.Load(), m.IdemReplays.Load())
	fmt.Fprintf(w, "rf_db_conns_acquired %d\nrf_db_conns_idle %d\nrf_db_conns_total %d\n", st.AcquiredConns(), st.IdleConns(), st.TotalConns())
	var pending, failed int64
	var lag float64
	_ = s.G.DB.QueryRow(r.Context(), `SELECT count(*) FILTER (WHERE status='pending' AND due_at <= now()), count(*) FILTER (WHERE status='failed'),
		COALESCE(EXTRACT(EPOCH FROM now() - min(due_at) FILTER (WHERE status='pending' AND due_at <= now())), 0)::float8
		FROM game_events`).Scan(&pending, &failed, &lag)
	fmt.Fprintf(w, "rf_events_overdue %d\nrf_events_parked %d\nrf_events_max_lag_seconds %.3f\n", pending, failed, lag)
}
