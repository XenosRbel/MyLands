package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"realmforge/server/internal/api"
	"realmforge/server/internal/game"
	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
	"realmforge/server/internal/testutil"
)

type harness struct {
	t     *testing.T
	srv   *httptest.Server
	g     *game.Service
	clk   *platform.FakeClock
	world int64
	n     int
	mu    sync.Mutex
}

func newHarness(t *testing.T, packPath string) *harness {
	db := testutil.DB(t, "api")
	clk := platform.NewFakeClock(time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC))
	g := game.New(db, clk, testutil.Logger())
	p, err := pack.LoadFile(packPath)
	if err != nil {
		t.Fatal(err)
	}
	wid, err := g.CreateWorld(context.Background(), "w1", "World One", p, 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	cfg := platform.Config{SessionTTL: 10000 * time.Hour, RateLimitRPS: 1000, RateLimitBurst: 1000, DevPlatformAuth: true,
		AllowedOrigins: []string{"http://localhost:5173"}, PacksDir: "../../../packs"}
	srv := httptest.NewServer(api.New(g, cfg, testutil.Logger()))
	t.Cleanup(srv.Close)
	return &harness{t: t, srv: srv, g: g, clk: clk, world: wid}
}

type resp struct {
	Status int
	Body   map[string]any
	Raw    []byte
	Header http.Header
}

func (h *harness) do(method, path, token string, body any, idem string) resp {
	h.t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if idem != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		h.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := resp{Status: res.StatusCode, Raw: raw, Header: res.Header}
	_ = json.Unmarshal(raw, &out.Body)
	return out
}

func (h *harness) key() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.n++
	return fmt.Sprintf("key-%08d", h.n)
}

func data(r resp) map[string]any {
	d, _ := r.Body["data"].(map[string]any)
	return d
}

func errCode(r resp) string {
	e, _ := r.Body["error"].(map[string]any)
	c, _ := e["code"].(string)
	return c
}

func (h *harness) register(login string) (string, int64) {
	h.t.Helper()
	r := h.do("POST", "/api/v1/auth/register", "", map[string]string{"login": login, "password": "password123"}, "")
	if r.Status != 200 {
		h.t.Fatalf("register: %d %s", r.Status, r.Raw)
	}
	tok := data(r)["token"].(string)
	acc := data(r)["account"].(map[string]any)
	return tok, int64(acc["id"].(float64))
}

func (h *harness) join(tok, name, race string) int64 {
	h.t.Helper()
	r := h.do("POST", fmt.Sprintf("/api/v1/worlds/%d/join", h.world), tok, map[string]string{"name": name, "race": race}, "")
	if r.Status != 200 {
		h.t.Fatalf("join: %d %s", r.Status, r.Raw)
	}
	ov := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/overview", h.world), tok, nil, "")
	cities := data(ov)["cities"].([]any)
	return int64(cities[0].(map[string]any)["id"].(float64))
}

func TestHealthAndAuthRequired(t *testing.T) {
	h := newHarness(t, "../pack/testdata/minimal.json")
	if r := h.do("GET", "/healthz", "", nil, ""); r.Status != 200 {
		t.Fatal("healthz")
	}
	if r := h.do("GET", "/readyz", "", nil, ""); r.Status != 200 {
		t.Fatal("readyz")
	}
	if r := h.do("GET", "/api/v1/worlds", "", nil, ""); r.Status != 401 || errCode(r) != "unauthenticated" {
		t.Fatalf("auth required: %d %s", r.Status, r.Raw)
	}
	if r := h.do("GET", "/api/v1/worlds", "garbage", nil, ""); r.Status != 401 {
		t.Fatal("bad token")
	}
	tok, _ := h.register("user1")
	if r := h.do("GET", "/api/v1/admin/audit", tok, nil, ""); r.Status != 403 {
		t.Fatalf("admin endpoints must be forbidden: %d", r.Status)
	}
	if r := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/overview", h.world), tok, nil, ""); r.Status != 403 {
		t.Fatalf("not joined: %d", r.Status)
	}
	// unknown JSON fields are rejected
	if r := h.do("POST", "/api/v1/auth/login", "", map[string]any{"login": "user1", "password": "password123", "admin": true}, ""); r.Status != 400 {
		t.Fatalf("unknown field must be rejected: %d", r.Status)
	}
	// metrics endpoint renders
	if r := h.do("GET", "/metrics", "", nil, ""); r.Status != 200 || !bytes.Contains(r.Raw, []byte("rf_events_processed_total")) {
		t.Fatal("metrics")
	}
}

func TestCORS(t *testing.T) {
	h := newHarness(t, "../pack/testdata/minimal.json")
	req, _ := http.NewRequest("OPTIONS", h.srv.URL+"/api/v1/worlds", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" || res.StatusCode != 204 {
		t.Fatalf("cors preflight: %d %v", res.StatusCode, res.Header)
	}
	req.Header.Set("Origin", "https://evil.example")
	res, _ = http.DefaultClient.Do(req)
	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("foreign origin must not be allowed")
	}
}

func TestIdempotentCommands(t *testing.T) {
	h := newHarness(t, "../pack/testdata/minimal.json")
	tok, _ := h.register("idem")
	city := h.join(tok, "Idem", "human")
	path := fmt.Sprintf("/api/v1/cities/%d/build", city)
	if r := h.do("POST", path, tok, map[string]string{"building": "lumber"}, ""); r.Status != 400 || errCode(r) != "idempotency_key_required" {
		t.Fatalf("missing key: %d %s", r.Status, r.Raw)
	}
	r1 := h.do("POST", path, tok, map[string]string{"building": "lumber"}, "build-lumber-1")
	if r1.Status != 200 {
		t.Fatalf("build: %s", r1.Raw)
	}
	r2 := h.do("POST", path, tok, map[string]string{"building": "lumber"}, "build-lumber-1")
	if r2.Status != 200 || !bytes.Equal(r1.Raw, r2.Raw) || r2.Header.Get("Idempotent-Replay") != "true" {
		t.Fatalf("replay must return the identical response: %s vs %s", r1.Raw, r2.Raw)
	}
	r3 := h.do("POST", path, tok, map[string]string{"building": "farm"}, "build-lumber-1")
	if r3.Status != 422 || errCode(r3) != "idempotency_key_reused" {
		t.Fatalf("key reuse with other body: %d %s", r3.Status, r3.Raw)
	}
	v := h.do("GET", fmt.Sprintf("/api/v1/cities/%d", city), tok, nil, "")
	q := data(v)["queues"].(map[string]any)["build"].([]any)
	if len(q) != 1 {
		t.Fatalf("exactly one job must exist, got %d", len(q))
	}
	// concurrent duplicates with the same key → one effect
	var wg sync.WaitGroup
	bodies := make([][]byte, 6)
	for i := range bodies {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			bodies[i] = h.do("POST", path, tok, map[string]string{"building": "farm"}, "farm-concurrent").Raw
		}(i)
	}
	wg.Wait()
	for _, b := range bodies[1:] {
		if !bytes.Equal(b, bodies[0]) {
			t.Fatalf("concurrent duplicates must see the same response:\n%s\n%s", bodies[0], b)
		}
	}
	v = h.do("GET", fmt.Sprintf("/api/v1/cities/%d", city), tok, nil, "")
	if q := data(v)["queues"].(map[string]any)["build"].([]any); len(q) != 2 {
		t.Fatalf("concurrent duplicate created %d jobs", len(q)-1)
	}
	// domain errors are stored and replayed as well
	e1 := h.do("POST", path, tok, map[string]string{"building": "barracks"}, "err-key-1")
	e2 := h.do("POST", path, tok, map[string]string{"building": "barracks"}, "err-key-1")
	if e1.Status != 422 || !bytes.Equal(e1.Raw, e2.Raw) {
		t.Fatalf("error replay: %d %s / %s", e1.Status, e1.Raw, e2.Raw)
	}
}

func TestCrossAccountAccessDenied(t *testing.T) {
	h := newHarness(t, "../pack/testdata/minimal.json")
	ta, _ := h.register("owner")
	tb, _ := h.register("intruder")
	ca := h.join(ta, "Owner", "human")
	h.join(tb, "Intruder", "human")
	if r := h.do("GET", fmt.Sprintf("/api/v1/cities/%d", ca), tb, nil, ""); r.Status != 404 {
		t.Fatalf("foreign city read: %d", r.Status)
	}
	if r := h.do("POST", fmt.Sprintf("/api/v1/cities/%d/build", ca), tb, map[string]string{"building": "lumber"}, h.key()); r.Status != 404 {
		t.Fatalf("foreign city command: %d", r.Status)
	}
	if r := h.do("GET", "/api/v1/cities/abc", tb, nil, ""); r.Status != 400 {
		t.Fatalf("bad id: %d", r.Status)
	}
	// SQL injection attempts are inert (parameterised queries + strict ids)
	if r := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/market?sell=wood'%%20OR%%201=1--", h.world), tb, nil, ""); r.Status != 200 || len(r.Body["data"].([]any)) != 0 {
		t.Fatalf("injection: %d %s", r.Status, r.Raw)
	}
}

// TestEndToEnd drives a complete session through HTTP using the original
// Ashen Crowns pack and verifies resulting database state.
func TestEndToEnd(t *testing.T) {
	h := newHarness(t, "../../../packs/ashen-crowns")
	tok, accID := h.register("e2e_player")
	if r := h.do("GET", "/api/v1/worlds", tok, nil, ""); r.Status != 200 || len(r.Body["data"].([]any)) != 1 {
		t.Fatalf("worlds: %s", r.Raw)
	}
	if r := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/pack", h.world), tok, nil, ""); r.Status != 200 || r.Header.Get("ETag") == "" {
		t.Fatal("pack download")
	}
	city := h.join(tok, "Lady Ember", "thornkin")
	cp := fmt.Sprintf("/api/v1/cities/%d", city)
	// 1. build: lumberyard to level 2
	if r := h.do("POST", cp+"/build", tok, map[string]string{"building": "lumberyard"}, h.key()); r.Status != 200 {
		t.Fatalf("build: %s", r.Raw)
	}
	h.clk.Advance(5 * time.Minute)
	v := data(h.do("GET", cp, tok, nil, ""))
	if lvl := v["city"].(map[string]any)["buildings"].(map[string]any)["lumberyard"].(float64); lvl != 2 {
		t.Fatalf("lumberyard level %v", lvl)
	}
	// 2. tutorial quest claim (exactly once)
	qp := fmt.Sprintf("/api/v1/worlds/%d/quests/t01_lumber/claim", h.world)
	if r := h.do("POST", qp, tok, map[string]any{"cityId": city}, h.key()); r.Status != 200 {
		t.Fatalf("claim: %s", r.Raw)
	}
	if r := h.do("POST", qp, tok, map[string]any{"cityId": city}, h.key()); r.Status != 409 {
		t.Fatalf("second claim must conflict: %d", r.Status)
	}
	// 3. keep 2 → scriptorium → research forestry
	if err := h.g.AdminGrant(context.Background(), mustAdmin(t, h), city, pack.Amounts{"timber": 5000, "stone": 5000, "gold": 3000, "grain": 5000, "iron": 2000}, "e2e setup funds"); err != nil {
		t.Fatal(err)
	}
	for _, b := range []string{"keep", "warehouse"} {
		if r := h.do("POST", cp+"/build", tok, map[string]string{"building": b}, h.key()); r.Status != 200 {
			t.Fatalf("build %s: %s", b, r.Raw)
		}
	}
	h.clk.Advance(30 * time.Minute)
	for _, b := range []string{"scriptorium", "barracks"} {
		if r := h.do("POST", cp+"/build", tok, map[string]string{"building": b}, h.key()); r.Status != 200 {
			t.Fatalf("build %s: %s", b, r.Raw)
		}
	}
	h.clk.Advance(30 * time.Minute)
	if err := h.g.AdminGrant(context.Background(), mustAdmin(t, h), city, pack.Amounts{"lore": 500}, "e2e setup lore"); err != nil {
		t.Fatal(err)
	}
	if r := h.do("POST", cp+"/research", tok, map[string]string{"tech": "forestry"}, h.key()); r.Status != 200 {
		t.Fatalf("research: %s", r.Raw)
	}
	// 4. train troops
	if r := h.do("POST", cp+"/train", tok, map[string]any{"unit": "militia", "quantity": 30}, h.key()); r.Status != 200 {
		t.Fatalf("train: %s", r.Raw)
	}
	h.clk.Advance(2 * time.Hour)
	v = data(h.do("GET", cp, tok, nil, ""))
	units := v["city"].(map[string]any)["units"].(map[string]any)
	if units["militia"].(float64) != 40 {
		t.Fatalf("militia %v", units["militia"])
	}
	var forestry int
	if err := h.g.DB.QueryRow(context.Background(), `SELECT (techs->>'forestry')::int FROM players WHERE account_id=$1`, accID).Scan(&forestry); err != nil || forestry != 1 {
		t.Fatalf("research not persisted: %d %v", forestry, err)
	}
	// 5. send a mission against the nearest level-1 neutral lair and get the result
	cx, cy := int(v["city"].(map[string]any)["x"].(float64)), int(v["city"].(map[string]any)["y"].(float64))
	var tx, ty int
	err := h.g.DB.QueryRow(context.Background(), `SELECT x, y FROM tiles WHERE world_id=$1 AND neutral_id='wolf_den' AND neutral_level=1
		ORDER BY (x-$2)^2+(y-$3)^2 LIMIT 1`, h.world, cx, cy).Scan(&tx, &ty)
	if err != nil {
		t.Skip("no level-1 wolf den on this seed")
	}
	mr := h.do("POST", cp+"/missions", tok, map[string]any{"kind": "attack", "x": tx, "y": ty, "units": map[string]int{"militia": 40}}, h.key())
	if mr.Status != 200 {
		t.Fatalf("mission: %s", mr.Raw)
	}
	h.clk.Advance(24 * time.Hour)
	rep := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/reports", h.world), tok, nil, "")
	reports := rep.Body["data"].([]any)
	if len(reports) != 1 {
		t.Fatalf("expected one battle report, got %d", len(reports))
	}
	won := reports[0].(map[string]any)["data"].(map[string]any)["battle"].(map[string]any)["attackerWon"].(bool)
	var status string
	if err := h.g.DB.QueryRow(context.Background(), `SELECT status FROM missions WHERE player_id=(SELECT id FROM players WHERE account_id=$1)`, accID).Scan(&status); err != nil || status != "done" {
		t.Fatalf("mission must be done: %s %v", status, err)
	}
	if won {
		var cleared bool
		_ = h.g.DB.QueryRow(context.Background(), `SELECT cleared_until > $4 FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`, h.world, tx, ty, h.clk.Now().Add(-24*time.Hour)).Scan(&cleared)
		if !cleared {
			t.Fatal("won battle must clear the lair")
		}
	}
	// 6. notifications exist and can be marked read
	n := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/notifications", h.world), tok, nil, "")
	items := data(n)["items"].([]any)
	if len(items) < 3 {
		t.Fatalf("notifications: %d", len(items))
	}
	top := int64(items[0].(map[string]any)["id"].(float64))
	if r := h.do("POST", fmt.Sprintf("/api/v1/worlds/%d/notifications/read", h.world), tok, map[string]any{"upTo": top}, h.key()); r.Status != 200 {
		t.Fatal(string(r.Raw))
	}
	if n := h.do("GET", fmt.Sprintf("/api/v1/worlds/%d/notifications", h.world), tok, nil, ""); data(n)["unread"].(float64) != 0 {
		t.Fatal("notifications should be read")
	}
	// 7. account deletion endpoint
	if r := h.do("DELETE", "/api/v1/me", tok, nil, ""); r.Status != 200 {
		t.Fatal("delete")
	}
	if r := h.do("GET", "/api/v1/me", tok, nil, ""); r.Status != 401 {
		t.Fatal("deleted account token must stop working")
	}
}

func mustAdmin(t *testing.T, h *harness) int64 {
	var id int64
	err := h.g.DB.QueryRow(context.Background(), `INSERT INTO accounts(display_name, is_admin) VALUES ('ops', true) RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestGuestAndDevPlatform(t *testing.T) {
	h := newHarness(t, "../pack/testdata/minimal.json")
	g := h.do("POST", "/api/v1/auth/guest", "", nil, "")
	if g.Status != 200 || !data(g)["account"].(map[string]any)["isGuest"].(bool) {
		t.Fatalf("guest: %s", g.Raw)
	}
	p1 := h.do("POST", "/api/v1/auth/platform", "", map[string]string{"platform": "dev", "externalId": "device-123"}, "")
	p2 := h.do("POST", "/api/v1/auth/platform", "", map[string]string{"platform": "dev", "externalId": "device-123"}, "")
	if p1.Status != 200 || data(p1)["account"].(map[string]any)["id"] != data(p2)["account"].(map[string]any)["id"] {
		t.Fatal("same platform identity must map to the same account")
	}
	if r := h.do("POST", "/api/v1/auth/platform", "", map[string]string{"platform": "yandex", "externalId": "x123"}, ""); r.Status != 501 {
		t.Fatalf("unconfigured platform: %d", r.Status)
	}
}
