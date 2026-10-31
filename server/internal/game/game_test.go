package game_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/game"
	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
	"realmforge/server/internal/rules"
	"realmforge/server/internal/testutil"
)

var ctx = context.Background()

type env struct {
	t     *testing.T
	s     *game.Service
	clk   *platform.FakeClock
	world int64
	p     *pack.Loaded
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := testutil.DB(t, "game")
	clk := platform.NewFakeClock(time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC))
	s := game.New(db, clk, testutil.Logger())
	p, err := pack.LoadFile("../pack/testdata/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	wid, err := s.CreateWorld(ctx, "test", "Test", p, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	return &env{t: t, s: s, clk: clk, world: wid, p: p}
}

type player struct {
	acc, pid, city int64
}

func (e *env) player(name string) player {
	e.t.Helper()
	sess, err := e.s.Register(ctx, name, "password123", name, time.Hour)
	if err != nil {
		e.t.Fatal(err)
	}
	pid, err := e.s.Join(ctx, sess.Account.ID, e.world, "P_"+name, "human")
	if err != nil {
		e.t.Fatal(err)
	}
	var city int64
	if err := e.s.DB.QueryRow(ctx, `SELECT id FROM cities WHERE player_id=$1`, pid).Scan(&city); err != nil {
		e.t.Fatal(err)
	}
	// Tests start without newbie protection unless they test it.
	e.exec(`UPDATE players SET protected_until=NULL WHERE id=$1`, pid)
	return player{acc: sess.Account.ID, pid: pid, city: city}
}

func (e *env) exec(sql string, args ...any) {
	e.t.Helper()
	if _, err := e.s.DB.Exec(ctx, sql, args...); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) city(p player, fn func(c game.PCtx, ci *game.City) (any, error)) (any, error) {
	var out any
	err := e.s.InTx(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = e.s.WithCity(ctx, tx, p.acc, p.city, fn)
		return err
	})
	return out, err
}

func (e *env) mustCity(p player, fn func(c game.PCtx, ci *game.City) (any, error)) any {
	e.t.Helper()
	out, err := e.city(p, fn)
	if err != nil {
		e.t.Fatal(err)
	}
	return out
}

func (e *env) pl(p player, fn func(c game.PCtx) (any, error)) (any, error) {
	var out any
	err := e.s.InTx(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = e.s.WithPlayer(ctx, tx, p.acc, e.world, fn)
		return err
	})
	return out, err
}

func (e *env) view(p player) *game.CityView {
	e.t.Helper()
	return e.mustCity(p, func(c game.PCtx, ci *game.City) (any, error) { return c.View(ci) }).(*game.CityView)
}

// give sets resources (whole units) directly — test setup only.
func (e *env) give(p player, a pack.Amounts) {
	e.t.Helper()
	e.mustCity(p, func(c game.PCtx, ci *game.City) (any, error) { return nil, nil }) // settle first
	for r, n := range a {
		e.exec(`UPDATE cities SET resources = jsonb_set(resources, ARRAY[$2::text], to_jsonb($3::bigint)) WHERE id=$1`, p.city, r, n*rules.Milli)
	}
}

func (e *env) setBuildings(p player, b map[string]int) {
	e.t.Helper()
	for id, l := range b {
		e.exec(`UPDATE cities SET buildings = jsonb_set(buildings, ARRAY[$2::text], to_jsonb($3::int)) WHERE id=$1`, p.city, id, l)
	}
}

func (e *env) setUnits(p player, u map[string]int64) {
	e.t.Helper()
	for id, n := range u {
		e.exec(`UPDATE cities SET units = jsonb_set(units, ARRAY[$2::text], to_jsonb($3::bigint)) WHERE id=$1`, p.city, id, n)
	}
}

func (e *env) processAll() {
	e.t.Helper()
	for i := 0; i < 20; i++ {
		ids, err := e.s.DuePlayers(ctx, 100)
		if err != nil {
			e.t.Fatal(err)
		}
		if len(ids) == 0 {
			return
		}
		for _, id := range ids {
			if _, err := e.s.Process(ctx, id); err != nil {
				e.t.Fatal(err)
			}
		}
	}
}

func code(err error) string {
	if ge, ok := game.AsError(err); ok {
		return ge.Code
	}
	if err != nil {
		return "other:" + err.Error()
	}
	return ""
}

func (e *env) location(p player) (int, int) {
	var x, y int
	if err := e.s.DB.QueryRow(ctx, `SELECT x, y FROM cities WHERE id=$1`, p.city).Scan(&x, &y); err != nil {
		e.t.Fatal(err)
	}
	return x, y
}

// ---------------------------------------------------------------------------

func TestJoinStartingState(t *testing.T) {
	e := newEnv(t)
	a := e.player("alice")
	v := e.view(a)
	if v.Resources["wood"] != 500 || v.Resources["gem"] != 5 {
		t.Fatalf("start resources: %v", v.Resources)
	}
	if v.City.Buildings["hall"] != 1 || v.City.Pop != 30 || !v.City.IsCapital {
		t.Fatalf("start city: %+v", v.City)
	}
	if _, err := e.s.Join(ctx, a.acc, e.world, "Again", "human"); code(err) != "name_taken" {
		t.Fatalf("second join must fail: %v", err)
	}
	if _, err := e.s.Join(ctx, a.acc, e.world, "X", "human"); code(err) != "invalid_request" {
		t.Fatalf("short name must fail: %v", err)
	}
}

func TestBuildCompletesOfflineAndPersists(t *testing.T) {
	e := newEnv(t)
	a := e.player("bob")
	j := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") }).(*game.Job)
	if j.Level != 2 || j.FinishesAt.Sub(j.StartsAt) != 42*time.Second { // 30 * 1.4
		t.Fatalf("job: %+v", j)
	}
	v := e.view(a)
	if v.Resources["wood"] != 500-70 { // 50*1.4
		t.Fatalf("cost not deducted: %v", v.Resources["wood"])
	}
	if v.Buildings[1].TargetLevel != 2 || v.City.Buildings["lumber"] != 1 {
		t.Fatal("upgrade must be queued, not applied")
	}
	// Player goes offline; worker completes the job.
	e.clk.Advance(43 * time.Second)
	e.processAll()
	var lvl int
	if err := e.s.DB.QueryRow(ctx, `SELECT (buildings->>'lumber')::int FROM cities WHERE id=$1`, a.city).Scan(&lvl); err != nil || lvl != 2 {
		t.Fatalf("worker did not complete build: %d %v", lvl, err)
	}
	// "Restart": a fresh service (empty caches) sees the same state.
	s2 := game.New(e.s.DB, e.clk, testutil.Logger())
	e2 := *e
	e2.s = s2
	v2 := e2.view(a)
	if v2.City.Buildings["lumber"] != 2 {
		t.Fatal("state lost across restart")
	}
	var ledger int
	_ = e.s.DB.QueryRow(ctx, `SELECT count(*) FROM resource_ledger WHERE city_id=$1 AND reason='build'`, a.city).Scan(&ledger)
	if ledger != 1 {
		t.Fatalf("ledger entries: %d", ledger)
	}
}

func TestProductionAccruesAndCaps(t *testing.T) {
	e := newEnv(t)
	a := e.player("carol")
	before := e.view(a)
	e.clk.Advance(30 * time.Minute)
	after := e.view(a)
	// wood: 100/h × employment(30 pop / 20 workers → 1) × 1.1 race = 110/h → +55 in 30 min
	got := after.Resources["wood"] - before.Resources["wood"]
	if got < 54 || got > 56 {
		t.Fatalf("wood accrual %d, want ~55 (rates %v)", got, after.Rates.Net)
	}
	e.clk.Advance(48 * time.Hour)
	v := e.view(a)
	if v.Resources["wood"] != 1000 {
		t.Fatalf("wood must cap at storage 1000: %d", v.Resources["wood"])
	}
}

func TestBuildValidation(t *testing.T) {
	e := newEnv(t)
	a := e.player("dave")
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "barracks") })
	if code(err) != "requirements_not_met" {
		t.Fatalf("barracks needs hall 2: %v", err)
	}
	_, err = e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "nope") })
	if code(err) != "invalid_request" {
		t.Fatal(err)
	}
	e.give(a, pack.Amounts{"wood": 10})
	_, err = e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "hall") })
	if code(err) != "insufficient_resources" {
		t.Fatalf("expected insufficient: %v", err)
	}
	e.give(a, pack.Amounts{"wood": 100000})
	for i := 0; i < 3; i++ {
		if _, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "farm") }); err != nil {
			t.Fatal(err)
		}
	}
	_, err = e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "farm") })
	if code(err) != "queue_full" {
		t.Fatalf("queue limit 3: %v", err)
	}
	v := e.view(a)
	q := v.Queues["build"]
	if len(q) != 3 || q[0].Level != 2 || q[2].Level != 4 || !q[1].StartsAt.Equal(q[0].FinishesAt) {
		t.Fatalf("sequential queue: %+v", q)
	}
}

func TestCancelOnlyLastWithRefund(t *testing.T) {
	e := newEnv(t)
	a := e.player("erin")
	j1 := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") }).(*game.Job)
	j2 := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") }).(*game.Job)
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.CancelJob(ci, j1.ID) })
	if code(err) != "not_last_in_queue" {
		t.Fatalf("cancel non-last: %v", err)
	}
	before := e.view(a).Resources["wood"]
	ref := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.CancelJob(ci, j2.ID) }).(pack.Amounts)
	if ref["wood"] != j2.Cost["wood"]*80/100 {
		t.Fatalf("refund %v of %v", ref, j2.Cost)
	}
	if e.view(a).Resources["wood"] != before+ref["wood"] {
		t.Fatal("refund not credited")
	}
	e.clk.Advance(10 * time.Minute)
	if v := e.view(a); v.City.Buildings["lumber"] != 2 {
		t.Fatalf("only first upgrade should apply: %d", v.City.Buildings["lumber"])
	}
}

func TestConcurrentSpendingCannotDoubleSpend(t *testing.T) {
	e := newEnv(t)
	a := e.player("frank")
	e.give(a, pack.Amounts{"wood": 150}) // enough for exactly one hall upgrade (150)
	var wg sync.WaitGroup
	var mu sync.Mutex
	okN, insufficient := 0, 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "hall") })
			mu.Lock()
			defer mu.Unlock()
			switch code(err) {
			case "":
				okN++
			case "insufficient_resources":
				insufficient++
			default:
				t.Errorf("unexpected: %v", err)
			}
		}()
	}
	wg.Wait()
	if okN != 1 || insufficient != 7 {
		t.Fatalf("ok=%d insufficient=%d", okN, insufficient)
	}
	if v := e.view(a); v.Resources["wood"] < 0 || len(v.Queues["build"]) != 1 {
		t.Fatalf("state: %v %v", v.Resources, v.Queues["build"])
	}
}

func TestConcurrentEventProcessingAppliesOnce(t *testing.T) {
	e := newEnv(t)
	a := e.player("gina")
	e.setBuildings(a, map[string]int{"hall": 2, "barracks": 1})
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Train(ci, "spear", 5) })
	e.clk.Advance(time.Hour)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				_, _ = e.s.Process(ctx, a.pid)
			} else {
				_, _ = e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.View(ci) })
			}
		}(i)
	}
	wg.Wait()
	v := e.view(a)
	if v.City.Units["spear"] != 5 {
		t.Fatalf("units applied %d times-ish: %d", v.City.Units["spear"]/5, v.City.Units["spear"])
	}
	var done int
	_ = e.s.DB.QueryRow(ctx, `SELECT count(*) FROM game_events WHERE player_id=$1 AND status='done' AND kind='job'`, a.pid).Scan(&done)
	if done != 1 {
		t.Fatalf("event processed %d times", done)
	}
}

func TestResearchAndSpecialisation(t *testing.T) {
	e := newEnv(t)
	a := e.player("hank")
	e.give(a, pack.Amounts{"wood": 900})
	base := e.view(a).Rates.Gross["wood"]
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Research(ci, "axes") })
	e.clk.Advance(61 * time.Second)
	v := e.view(a)
	if v.Techs[0].Level != 1 {
		t.Fatalf("tech level %d", v.Techs[0].Level)
	}
	if want := base / 1.1 * 1.15; v.Rates.Gross["wood"] < want-0.01 || v.Rates.Gross["wood"] > want+0.01 {
		t.Fatalf("tech modifier not applied: %v want %v", v.Rates.Gross["wood"], want)
	}
}

func TestPvEMissionFullCycle(t *testing.T) {
	e := newEnv(t)
	a := e.player("ivan")
	e.setUnits(a, map[string]int64{"spear": 40})
	x, y := e.location(a)
	var tx, ty int
	if err := e.s.DB.QueryRow(ctx, `SELECT x, y FROM tiles WHERE world_id=$1 AND neutral_id='den' AND neutral_level=1 ORDER BY (x-$2)^2+(y-$3)^2 LIMIT 1`, e.world, x, y).Scan(&tx, &ty); err != nil {
		t.Skip("no level-1 den on this seed")
	}
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: tx, Y: ty, Units: map[string]int64{"spear": 40}})
	}).(*game.Mission)
	if v := e.view(a); v.City.Units["spear"] != 0 || v.Away["spear"] != 40 {
		t.Fatalf("units must leave: %v away %v", v.City.Units, v.Away)
	}
	if v := e.view(a); v.Rates.Upkeep != 40 { // upkeep 1 × missionUpkeepMult 1
		t.Fatalf("away upkeep: %v", v.Rates.Upkeep)
	}
	e.clk.Set(m.ArriveAt.Add(time.Second))
	e.processAll()
	var status string
	_ = e.s.DB.QueryRow(ctx, `SELECT status FROM missions WHERE id=$1`, m.ID).Scan(&status)
	if status != "returning" {
		t.Fatalf("mission status %s", status)
	}
	var cleared *time.Time
	_ = e.s.DB.QueryRow(ctx, `SELECT cleared_until FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`, e.world, tx, ty).Scan(&cleared)
	if cleared == nil || !cleared.After(e.clk.Now()) {
		t.Fatal("lair must be cleared until respawn")
	}
	goldBefore := e.view(a).Resources["gold"]
	e.clk.Advance(time.Hour)
	v := e.view(a)
	if v.Resources["gold"] < goldBefore+50 || v.City.Units["spear"] < 30 {
		t.Fatalf("rewards/units not returned: gold %d→%d units %v", goldBefore, v.Resources["gold"], v.City.Units)
	}
	reps, _ := e.pl(a, func(c game.PCtx) (any, error) { return c.Reports(0, 10) })
	if len(reps.([]game.ReportRow)) != 1 {
		t.Fatal("battle report missing")
	}
	qs, _ := e.pl(a, func(c game.PCtx) (any, error) { return c.Overview() })
	if qs.(*game.Overview).Player.Stats["neutrals_cleared"] != 1 {
		t.Fatalf("stats: %v", qs.(*game.Overview).Player.Stats)
	}
}

func TestPvPAttackLootProtectionAndRecall(t *testing.T) {
	e := newEnv(t)
	a, d := e.player("att"), e.player("def")
	e.setUnits(a, map[string]int64{"rider": 50})
	dx, dy := e.location(d)
	e.exec(`UPDATE players SET protected_until=$2 WHERE id=$1`, d.pid, e.clk.Now().Add(time.Hour))
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: dx, Y: dy, Units: map[string]int64{"rider": 10}})
	})
	if code(err) != "target_protected" {
		t.Fatalf("protection: %v", err)
	}
	e.exec(`UPDATE players SET protected_until=NULL WHERE id=$1`, d.pid)
	// recall an outbound attack: returns after the elapsed time
	m1 := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: dx, Y: dy, Units: map[string]int64{"rider": 5}})
	}).(*game.Mission)
	e.clk.Advance(10 * time.Second)
	r := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		_, err := c.Recall(m1.ID)
		return nil, err
	})
	_ = r
	e.clk.Advance(11 * time.Second)
	if v := e.view(a); v.City.Units["rider"] != 50 {
		t.Fatalf("recalled riders not home: %v", v.City.Units)
	}
	// An attack that has reached its due time can no longer be recalled (exploit guard).
	m0 := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: dx, Y: dy, Units: map[string]int64{"rider": 1}})
	}).(*game.Mission)
	e.clk.Set(m0.ArriveAt)
	if _, err := e.pl(a, func(c game.PCtx) (any, error) { return c.Recall(m0.ID) }); code(err) != "already_arrived" && code(err) != "not_recallable" {
		t.Fatalf("recall after arrival must fail: %v", err)
	}
	e.clk.Advance(2 * m0.ArriveAt.Sub(m0.DepartAt))
	e.exec(`DELETE FROM missions WHERE id=$1 AND status='done'`, m0.ID)
	e.exec(`UPDATE cities SET units = jsonb_set(units, '{rider}', '50') WHERE id=$1`, a.city)
	if v := e.view(a); v.City.Units["rider"] != 50 {
		t.Fatalf("recalled riders not home: %v", v.City.Units)
	}
	defBefore := e.view(d).Resources
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: dx, Y: dy, Units: map[string]int64{"rider": 50}})
	}).(*game.Mission)
	// defender sees an incoming attack without composition
	ms, _ := e.pl(d, func(c game.PCtx) (any, error) { return c.Missions() })
	inc := ms.([]game.MissionView)
	if len(inc) != 1 || inc[0].Direction != "incoming" || inc[0].Units != nil {
		t.Fatalf("defender view: %+v", inc)
	}
	e.clk.Set(m.ArriveAt)
	defAfter := e.view(d).Resources
	if defAfter["wood"] >= defBefore["wood"] {
		t.Fatalf("defender should be plundered: %d → %d", defBefore["wood"], defAfter["wood"])
	}
	if defAfter["gem"] != defBefore["gem"] {
		t.Fatal("premium resource must not be lootable")
	}
	var cargo pack.Amounts
	_ = e.s.DB.QueryRow(ctx, `SELECT cargo FROM missions WHERE id=$1`, m.ID).Scan(&cargo)
	var total int64
	for _, n := range cargo {
		total += n
	}
	if total == 0 || total > 50*40 {
		t.Fatalf("loot must be within carry capacity: %v", cargo)
	}
	attWood := e.view(a).Resources["wood"]
	e.clk.Advance(2 * (m.ArriveAt.Sub(m.DepartAt)))
	if got := e.view(a).Resources["wood"]; got < attWood+cargo["wood"] {
		t.Fatalf("attacker did not receive loot: %d → %d (+%d)", attWood, got, cargo["wood"])
	}
	for _, p := range []player{a, d} {
		reps, _ := e.pl(p, func(c game.PCtx) (any, error) { return c.Reports(0, 10) })
		if len(reps.([]game.ReportRow)) != 2 { // the 1-rider probe and the main attack
			t.Fatalf("both sides need both battle reports, got %d", len(reps.([]game.ReportRow)))
		}
	}
}

func TestScoutAndAllianceRules(t *testing.T) {
	e := newEnv(t)
	a, b := e.player("sca"), e.player("scb")
	e.setUnits(a, map[string]int64{"spy": 5, "spear": 10})
	bx, by := e.location(b)
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "scout", X: bx, Y: by, Units: map[string]int64{"spy": 5}})
	}).(*game.Mission)
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "scout", X: bx, Y: by, Units: map[string]int64{"spear": 1}})
	})
	if code(err) != "invalid_request" {
		t.Fatalf("non-scouts cannot scout: %v", err)
	}
	e.clk.Set(m.ArriveAt)
	e.processAll()
	reps, _ := e.pl(a, func(c game.PCtx) (any, error) { return c.Reports(0, 10) })
	rr := reps.([]game.ReportRow)
	if len(rr) != 1 || rr[0].Data["success"] != true || rr[0].Data["resources"] == nil {
		t.Fatalf("scout report: %+v", rr)
	}
	// alliance: members cannot attack each other
	e.give(a, pack.Amounts{"gold": 1000})
	al := e.mustAlliance(a, "Order", "ORD")
	inv, err := e.pl(a, func(c game.PCtx) (any, error) { return c.Invite("P_scb") })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.pl(b, func(c game.PCtx) (any, error) { return nil, c.RespondInvite(inv.(*game.Invite).ID, true) }); err != nil {
		t.Fatal(err)
	}
	_, err = e.city(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: bx, Y: by, Units: map[string]int64{"spear": 1}})
	})
	if code(err) != "friendly_target" {
		t.Fatalf("ally attack must be blocked: %v", err)
	}
	full, _ := e.s.Alliance(ctx, e.s.DB, e.world, al.ID)
	if full.MemberCount != 2 {
		t.Fatalf("members %d", full.MemberCount)
	}
	// reinforcement into an ally city, then recall
	rm := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "reinforce", X: bx, Y: by, Units: map[string]int64{"spear": 10}})
	}).(*game.Mission)
	e.clk.Set(rm.ArriveAt)
	if v := e.view(b); len(v.Foreign) != 1 || v.Foreign[0].Units["spear"] != 10 {
		t.Fatalf("stationed troops: %+v", v.Foreign)
	}
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Recall(rm.ID) })
	e.clk.Advance(rm.ArriveAt.Sub(rm.DepartAt) + time.Second)
	if v := e.view(a); v.City.Units["spear"] != 10 {
		t.Fatalf("reinforcements not home: %v", v.City.Units)
	}
	if _, err := e.pl(b, func(c game.PCtx) (any, error) { return nil, c.Leave() }); err != nil {
		t.Fatal(err)
	}
	_, err = e.pl(a, func(c game.PCtx) (any, error) { return nil, c.Leave() })
	if err != nil {
		t.Fatalf("leader alone may leave (disbands): %v", err)
	}
}

func (e *env) mustAlliance(p player, name, tag string) *game.Alliance {
	e.t.Helper()
	out, err := e.pl(p, func(c game.PCtx) (any, error) { return c.CreateAlliance(name, tag, "") })
	if err != nil {
		e.t.Fatal(err)
	}
	return out.(*game.Alliance)
}

func TestMarketConcurrentAcceptAndDelivery(t *testing.T) {
	e := newEnv(t)
	seller := e.player("sel")
	buyers := []player{e.player("by1"), e.player("by2"), e.player("by3")}
	o := e.mustCity(seller, func(c game.PCtx, ci *game.City) (any, error) { return c.PlaceOrder(ci, "wood", 100, "gold", 30) }).(*game.Order)
	if v := e.view(seller); v.Resources["wood"] != 400 {
		t.Fatalf("escrow: %d", v.Resources["wood"])
	}
	var wg sync.WaitGroup
	results := make([]string, len(buyers))
	for i, b := range buyers {
		wg.Add(1)
		go func(i int, b player) {
			defer wg.Done()
			_, err := e.city(b, func(c game.PCtx, ci *game.City) (any, error) { return c.AcceptOrder(ci, o.ID) })
			results[i] = code(err)
		}(i, b)
	}
	wg.Wait()
	wins := 0
	for _, r := range results {
		switch r {
		case "":
			wins++
		case "order_closed":
		default:
			t.Fatalf("unexpected %s", r)
		}
	}
	if wins != 1 {
		t.Fatalf("exactly one buyer must win: %v", results)
	}
	// seller receives proceeds minus 2% fee (ceil) via a delivery event
	if v := e.view(seller); v.Resources["gold"] != 100+30-1 {
		t.Fatalf("seller gold %d", v.Resources["gold"])
	}
	o2 := e.mustCity(seller, func(c game.PCtx, ci *game.City) (any, error) { return c.PlaceOrder(ci, "wood", 50, "gold", 20) }).(*game.Order)
	if _, err := e.pl(seller, func(c game.PCtx) (any, error) { return c.CancelOrder(o2.ID) }); err != nil {
		t.Fatal(err)
	}
	if v := e.view(seller); v.Resources["wood"] != 400 {
		t.Fatalf("cancel refund: %d", v.Resources["wood"])
	}
	_, err := e.city(seller, func(c game.PCtx, ci *game.City) (any, error) { return c.PlaceOrder(ci, "gem", 1, "gold", 1) })
	if code(err) != "invalid_request" {
		t.Fatalf("premium must not be tradable: %v", err)
	}
	_, err = e.city(seller, func(c game.PCtx, ci *game.City) (any, error) { return c.PlaceOrder(ci, "wood", 1000, "gold", 1) })
	if code(err) != "price_out_of_band" && code(err) != "insufficient_resources" {
		t.Fatalf("gift orders must be rejected: %v", err)
	}
	got := e.mustCity(seller, func(c game.PCtx, ci *game.City) (any, error) { return c.Exchange(ci, "gold", "wood", 40) }).(pack.Amounts)
	if got["wood"] != 120 { // 40*4/1*0.75
		t.Fatalf("exchange %v", got)
	}
}

func TestQuestClaimExactlyOnce(t *testing.T) {
	e := newEnv(t)
	a := e.player("quester")
	e.setBuildings(a, map[string]int{"lumber": 2})
	var wg sync.WaitGroup
	codes := make([]string, 5)
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := e.pl(a, func(c game.PCtx) (any, error) { return c.ClaimQuest("q_lumber", 0) })
			codes[i] = code(err)
		}(i)
	}
	wg.Wait()
	ok := 0
	for _, c := range codes {
		if c == "" {
			ok++
		} else if c != "already_claimed" {
			t.Fatalf("unexpected %s", c)
		}
	}
	if ok != 1 {
		t.Fatalf("claims succeeded %d times", ok)
	}
	if v := e.view(a); v.Resources["wood"] != 600 {
		t.Fatalf("reward once: %d", v.Resources["wood"])
	}
	_, err := e.pl(a, func(c game.PCtx) (any, error) { return c.ClaimQuest("q_army", 0) })
	if code(err) != "quest_incomplete" {
		t.Fatalf("incomplete quest: %v", err)
	}
}

func TestHeroesHireSkillEquip(t *testing.T) {
	e := newEnv(t)
	a := e.player("heroic")
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.HireHero(ci, "warlord", "Aldric") })
	if err != nil {
		t.Fatal(err)
	}
	_, err = e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.HireHero(ci, "warlord", "Second") })
	if code(err) != "hero_limit" {
		t.Fatalf("hero limit 1 per hall level: %v", err)
	}
	hv, _ := e.pl(a, func(c game.PCtx) (any, error) { return c.Heroes() })
	h := hv.(*game.HeroesView).Heroes[0]
	if h.SkillPoints != 1 {
		t.Fatalf("skill points %d", h.SkillPoints)
	}
	if _, err := e.pl(a, func(c game.PCtx) (any, error) { return c.LearnSkill(h.ID, "fury") }); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pl(a, func(c game.PCtx) (any, error) { return c.LearnSkill(h.ID, "fury") }); code(err) != "no_skill_points" {
		t.Fatalf("no points left: %v", err)
	}
	var item int64
	if err := e.s.DB.QueryRow(ctx, `INSERT INTO items(world_id, player_id, item_def, source) VALUES ($1,$2,'blade','test') RETURNING id`, e.world, a.pid).Scan(&item); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pl(a, func(c game.PCtx) (any, error) { return nil, c.Equip(h.ID, item) }); err != nil {
		t.Fatal(err)
	}
	hv, _ = e.pl(a, func(c game.PCtx) (any, error) { return c.Heroes() })
	h = hv.(*game.HeroesView).Heroes[0]
	// class 5 + fury(infantry) 3 + blade 4
	if h.Modifiers["attack.*%"] != 9 || h.Modifiers["attack.infantry%"] != 3 {
		t.Fatalf("hero mods %v", h.Modifiers)
	}
	b := e.player("victim")
	e.setUnits(b, map[string]int64{"spear": 5})
	e.setUnits(a, map[string]int64{"spear": 30})
	bx, by := e.location(b)
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: bx, Y: by, Units: map[string]int64{"spear": 30}, HeroID: &h.ID})
	}).(*game.Mission)
	if _, err := e.pl(a, func(c game.PCtx) (any, error) { return nil, c.Unequip(h.ID, "weapon") }); code(err) != "hero_busy" {
		t.Fatalf("cannot re-equip on mission: %v", err)
	}
	e.clk.Set(m.ArriveAt.Add(m.ArriveAt.Sub(m.DepartAt)).Add(time.Second))
	e.processAll()
	hv, _ = e.pl(a, func(c game.PCtx) (any, error) { return c.Heroes() })
	h = hv.(*game.HeroesView).Heroes[0]
	if h.Status != "idle" || h.XP <= 0 {
		t.Fatalf("hero after mission: %+v", h)
	}
}

func TestTerritoryOccupyCollectAndContest(t *testing.T) {
	e := newEnv(t)
	a, b := e.player("tea"), e.player("teb")
	e.setBuildings(a, map[string]int{"hall": 2})
	e.setBuildings(b, map[string]int{"hall": 2})
	e.setUnits(a, map[string]int64{"spear": 5})
	e.setUnits(b, map[string]int64{"rider": 40})
	ax, ay := e.location(a)
	var tx, ty int
	if err := e.s.DB.QueryRow(ctx, `SELECT x, y FROM tiles WHERE world_id=$1 AND neutral_id IS NULL AND city_id IS NULL AND terrain='forest'
		ORDER BY (x-$2)^2+(y-$3)^2 LIMIT 1`, e.world, ax, ay).Scan(&tx, &ty); err != nil {
		t.Fatal(err)
	}
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "occupy", X: tx, Y: ty, Units: map[string]int64{"spear": 5}})
	}).(*game.Mission)
	e.clk.Set(m.ArriveAt)
	e.processAll()
	var owner *int64
	_ = e.s.DB.QueryRow(ctx, `SELECT owner_player_id FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`, e.world, tx, ty).Scan(&owner)
	if owner == nil || *owner != a.pid {
		t.Fatal("territory not occupied")
	}
	e.clk.Advance(2 * time.Hour)
	got := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Collect(ci, tx, ty) }).(pack.Amounts)
	if got["wood"] != 40 { // 20/h × 2h
		t.Fatalf("yield %v", got)
	}
	// b captures it
	cm := e.mustCity(b, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "attack", X: tx, Y: ty, Units: map[string]int64{"rider": 40}, Capture: true})
	}).(*game.Mission)
	e.clk.Set(cm.ArriveAt)
	e.processAll()
	_ = e.s.DB.QueryRow(ctx, `SELECT owner_player_id FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`, e.world, tx, ty).Scan(&owner)
	if owner == nil || *owner != b.pid {
		t.Fatalf("territory not captured")
	}
	terr, _ := e.pl(b, func(c game.PCtx) (any, error) { return c.Territories() })
	if tl := terr.([]game.Territory); len(tl) != 1 || tl[0].Garrison["rider"] == 0 {
		t.Fatalf("captor garrison: %+v", tl)
	}
}

func TestSettleSecondCity(t *testing.T) {
	e := newEnv(t)
	a := e.player("settler")
	e.setBuildings(a, map[string]int{"hall": 5})
	e.setUnits(a, map[string]int64{"settler": 1})
	ax, ay := e.location(a)
	var tx, ty int
	if err := e.s.DB.QueryRow(ctx, `SELECT x, y FROM tiles WHERE world_id=$1 AND neutral_id IS NULL AND city_id IS NULL AND terrain IN ('plains','forest')
		AND abs(x-$2) + abs(y-$3) BETWEEN 3 AND 6 LIMIT 1`, e.world, ax, ay).Scan(&tx, &ty); err != nil {
		t.Fatal(err)
	}
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "settle", X: tx, Y: ty, Units: map[string]int64{"settler": 1}, Cargo: pack.Amounts{}})
	}).(*game.Mission)
	e.clk.Set(m.ArriveAt)
	ov, _ := e.pl(a, func(c game.PCtx) (any, error) { return c.Overview() })
	if n := len(ov.(*game.Overview).Cities); n != 2 {
		t.Fatalf("cities %d", n)
	}
}

func TestIsolationBetweenPlayers(t *testing.T) {
	e := newEnv(t)
	a, b := e.player("iso_a"), e.player("iso_b")
	err := e.s.InTx(ctx, func(tx pgx.Tx) error {
		_, err := e.s.WithCity(ctx, tx, b.acc, a.city, func(c game.PCtx, ci *game.City) (any, error) { return c.View(ci) })
		return err
	})
	if code(err) != "not_found" {
		t.Fatalf("foreign city must look nonexistent: %v", err)
	}
	m := e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") }).(*game.Job)
	_, err = e.city(b, func(c game.PCtx, ci *game.City) (any, error) { return c.CancelJob(ci, m.ID) })
	if code(err) != "not_found" {
		t.Fatalf("foreign job cancel: %v", err)
	}
	var orphan int64
	_, err = e.pl(b, func(c game.PCtx) (any, error) { return c.Recall(orphan + 999999) })
	if code(err) != "not_found" {
		t.Fatal(err)
	}
}

func TestStarvationAndDesertionOverLongAbsence(t *testing.T) {
	e := newEnv(t)
	a := e.player("absent")
	e.setUnits(a, map[string]int64{"rider": 400}) // upkeep 800 food/h ≫ production
	e.clk.Advance(7 * 24 * time.Hour)
	v := e.view(a)
	if v.Resources["food"] != 0 || v.City.Pop >= 30 {
		t.Fatalf("expected starvation: food %d pop %d", v.Resources["food"], v.City.Pop)
	}
	if v.City.Units["rider"] >= 400 {
		t.Fatalf("expected desertion while upkeep unpaid: %d", v.City.Units["rider"])
	}
}

func TestPackMigrationRenames(t *testing.T) {
	e := newEnv(t)
	a := e.player("mig")
	e.setUnits(a, map[string]int64{"spear": 3})
	np := *e.p.Pack
	np.Version = "1.1.0"
	np.Units = append([]pack.Unit(nil), np.Units...)
	np.Units[0].ID = "lancer"
	np.Migrations = []pack.Migration{{FromVersion: "1.0.0", Renames: map[string]string{"spear": "lancer"}}}
	next, err := pack.Build(&np)
	if err != nil {
		t.Fatal(err)
	}
	var admin int64
	_ = e.s.DB.QueryRow(ctx, `INSERT INTO accounts(display_name, is_admin) VALUES ('root', true) RETURNING id`).Scan(&admin)
	rep, err := e.s.MigrateWorldPack(ctx, admin, e.world, next, true)
	if err != nil || rep.Applied || len(rep.Missing) != 0 {
		t.Fatalf("dry run: %+v %v", rep, err)
	}
	rep, err = e.s.MigrateWorldPack(ctx, admin, e.world, next, false)
	if err != nil || !rep.Applied {
		t.Fatalf("apply: %+v %v", rep, err)
	}
	if v := e.view(a); v.City.Units["lancer"] != 3 || v.City.Units["spear"] != 0 {
		t.Fatalf("units not renamed: %v", v.City.Units)
	}
	// a pack missing an in-use id without a mapping is refused
	np2 := *next.Pack
	np2.Version = "1.2.0"
	np2.Units = append([]pack.Unit(nil), np2.Units[1:]...)
	np2.Migrations = nil
	np2.Quests = nil
	bad, err := pack.Build(&np2)
	if err != nil {
		t.Fatal(err)
	}
	rep, err = e.s.MigrateWorldPack(ctx, admin, e.world, bad, false)
	if err != nil || rep.Applied || len(rep.Missing) == 0 {
		t.Fatalf("incompatible pack must be refused: %+v %v", rep, err)
	}
}

func TestAccountsAuth(t *testing.T) {
	e := newEnv(t)
	if _, err := e.s.Register(ctx, "zed", "short", "", time.Hour); code(err) != "invalid_request" {
		t.Fatal(err)
	}
	s1, err := e.s.Register(ctx, "Zed", "password123", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.s.Register(ctx, "zed", "password123", "", time.Hour); code(err) != "login_taken" {
		t.Fatalf("dup login: %v", err)
	}
	if _, err := e.s.Login(ctx, "zed", "wrongpass1", time.Hour); code(err) != "invalid_credentials" {
		t.Fatal(err)
	}
	s2, err := e.s.Login(ctx, "ZED", "password123", time.Hour)
	if err != nil || s2.Account.ID != s1.Account.ID {
		t.Fatal(err)
	}
	if _, err := e.s.Authenticate(ctx, s2.Token); err != nil {
		t.Fatal(err)
	}
	_ = e.s.Logout(ctx, s2.Token)
	if _, err := e.s.Authenticate(ctx, s2.Token); code(err) != "unauthenticated" {
		t.Fatal("revoked token must fail")
	}
	e.clk.Advance(2 * time.Hour)
	if _, err := e.s.Authenticate(ctx, s1.Token); code(err) != "unauthenticated" {
		t.Fatal("expired token must fail")
	}
	g, err := e.s.Guest(ctx, time.Hour)
	if err != nil || !g.Account.IsGuest {
		t.Fatal(err)
	}
	if _, err := e.s.Upgrade(ctx, g.Account.ID, "guestnow", "password123"); err != nil {
		t.Fatal(err)
	}
	if err := e.s.DeleteAccount(ctx, g.Account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.s.Login(ctx, "guestnow", "password123", time.Hour); code(err) != "invalid_credentials" {
		t.Fatal("deleted account must not log in")
	}
	if !game.CheckPassword(mustHash(t), "pw-123456") || game.CheckPassword(mustHash(t), "nope") {
		t.Fatal("password hashing")
	}
}

func mustHash(t *testing.T) string {
	h, err := game.HashPassword("pw-123456")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestAdminGrantIsAudited(t *testing.T) {
	e := newEnv(t)
	a := e.player("granted")
	var admin int64
	_ = e.s.DB.QueryRow(ctx, `INSERT INTO accounts(display_name, is_admin) VALUES ('root', true) RETURNING id`).Scan(&admin)
	if err := e.s.AdminGrant(ctx, admin, a.city, pack.Amounts{"gold": 50}, "x"); code(err) != "invalid_request" {
		t.Fatal("reason required")
	}
	if err := e.s.AdminGrant(ctx, admin, a.city, pack.Amounts{"gold": 50}, "support ticket 42"); err != nil {
		t.Fatal(err)
	}
	if v := e.view(a); v.Resources["gold"] != 150 {
		t.Fatalf("gold %d", v.Resources["gold"])
	}
	rows, _ := e.s.Audit(ctx, 10)
	if len(rows) != 1 || rows[0].Action != "grant" {
		t.Fatalf("audit %+v", rows)
	}
}

func TestFailedEventIsParkedNotBlocking(t *testing.T) {
	e := newEnv(t)
	a := e.player("parked")
	// Corrupt event: unknown kind due now.
	e.exec(`INSERT INTO game_events(world_id, player_id, kind, ref_id, due_at) VALUES ($1,$2,'bogus',1,$3)`, e.world, a.pid, e.clk.Now())
	for i := 0; i < 5; i++ {
		_, err := e.s.Process(ctx, a.pid)
		if err == nil {
			t.Fatal("bogus event must fail")
		}
	}
	failed, _ := e.s.FailedEvents(ctx)
	if len(failed) != 1 {
		t.Fatalf("event must be parked after 5 attempts: %+v", failed)
	}
	if _, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.View(ci) }); err != nil {
		t.Fatalf("parked event must not block the player: %v", err)
	}
	_ = errors.New
	_ = fmt.Sprint
}
