package game_test

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"realmforge/server/internal/game"
	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
	"realmforge/server/internal/testutil"
)

// newLayoutEnv is newEnv with params.cityLayout set (ML-M01).
func newLayoutEnv(t *testing.T, l pack.CityLayout) *env {
	t.Helper()
	raw, err := os.ReadFile("../pack/testdata/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m["params"].(map[string]any)["cityLayout"] = l
	b, _ := json.Marshal(m)
	p, err := pack.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	db := testutil.DB(t, "game")
	clk := platform.NewFakeClock(time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC))
	s := game.New(db, clk, testutil.Logger())
	wid, err := s.CreateWorld(ctx, "test", "Test", p, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	return &env{t: t, s: s, clk: clk, world: wid, p: p}
}

// assertCityMirror checks cities.buildings equals city_buildings and that no cell holds two buildings.
func (e *env) assertCityMirror(cityID int64, wantCells int) {
	e.t.Helper()
	var legacy map[string]int
	if err := e.s.DB.QueryRow(ctx, `SELECT buildings FROM cities WHERE id=$1`, cityID).Scan(&legacy); err != nil {
		e.t.Fatal(err)
	}
	rows, err := e.s.DB.Query(ctx, `SELECT building_id, level FROM city_buildings WHERE city_id=$1`, cityID)
	if err != nil {
		e.t.Fatal(err)
	}
	got := map[string]int{}
	for rows.Next() {
		var id string
		var l int
		if err := rows.Scan(&id, &l); err != nil {
			e.t.Fatal(err)
		}
		got[id] = l
	}
	rows.Close()
	if len(got) != len(legacy) {
		e.t.Fatalf("mirror differs: legacy %v rows %v", legacy, got)
	}
	for id, l := range legacy {
		if got[id] != l {
			e.t.Fatalf("mirror differs: legacy %v rows %v", legacy, got)
		}
	}
	var cells, dup int
	e.s.DB.QueryRow(ctx, `SELECT count(*) FROM city_cells WHERE city_id=$1`, cityID).Scan(&cells)
	e.s.DB.QueryRow(ctx, `SELECT count(*) FROM (SELECT cell_id FROM city_buildings WHERE city_id=$1 GROUP BY cell_id HAVING count(*)>1) x`, cityID).Scan(&dup)
	if cells != wantCells || dup != 0 {
		e.t.Fatalf("cells %d (want %d), cells with two buildings %d", cells, wantCells, dup)
	}
}

func TestCityCellsDualWriteAndOnePerCell(t *testing.T) { // R-CITY-3, R-CITY-4
	e := newLayoutEnv(t, pack.CityLayout{StartCells: 22, CapitalCells: 51, MaxByLevel: []int{22, 27, 33, 38, 43}, ExtraCities: 7})
	a := e.player("cells")
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return nil, nil })
	e.assertCityMirror(a.city, 51) // capital: 51 cells, start buildings mirrored
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") })
	e.clk.Advance(time.Minute)
	e.processAll()
	e.assertCityMirror(a.city, 51)
	var lvl int
	e.s.DB.QueryRow(ctx, `SELECT level FROM city_buildings WHERE city_id=$1 AND building_id='lumber'`, a.city).Scan(&lvl)
	if lvl != 2 {
		t.Fatalf("mirror level %d", lvl)
	}
	// The database refuses a second building on a used cell.
	_, err := e.s.DB.Exec(ctx, `INSERT INTO city_buildings(city_id, cell_id, building_id, level)
		SELECT city_id, cell_id, 'other', 1 FROM city_buildings WHERE city_id=$1 LIMIT 1`, a.city)
	if err == nil {
		t.Fatal("two buildings on one cell must fail")
	}
}

func TestBuildRejectedWhenNoFreeCell(t *testing.T) {
	e := newLayoutEnv(t, pack.CityLayout{StartCells: 3, CapitalCells: 3, MaxByLevel: []int{3}, ExtraCities: 1})
	a := e.player("full")
	e.setBuildings(a, map[string]int{"hall": 2})
	e.give(a, pack.Amounts{"wood": 100000})
	if _, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "barracks") }); code(err) != "no_free_cell" {
		t.Fatalf("want no_free_cell, got %v", err)
	}
	// Upgrading an existing building needs no new cell.
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "lumber") })
	e.assertCityMirror(a.city, 3)
}

func TestConcurrentBuildOnLastCell(t *testing.T) {
	e := newLayoutEnv(t, pack.CityLayout{StartCells: 4, CapitalCells: 4, MaxByLevel: []int{4}, ExtraCities: 1})
	a := e.player("race")
	e.setBuildings(a, map[string]int{"hall": 2})
	e.give(a, pack.Amounts{"wood": 100000})
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) { return c.Build(ci, "barracks") })
			mu.Lock()
			defer mu.Unlock()
			switch code(err) {
			case "":
				ok++
			case "queue_full":
			default:
				t.Errorf("unexpected: %v", err)
			}
		}()
	}
	wg.Wait()
	e.clk.Advance(time.Hour)
	e.processAll()
	var n int
	e.s.DB.QueryRow(ctx, `SELECT count(*) FROM city_buildings WHERE city_id=$1 AND building_id='barracks'`, a.city).Scan(&n)
	if ok < 1 || n != 1 {
		t.Fatalf("ok=%d barracks rows=%d", ok, n)
	}
	e.assertCityMirror(a.city, 4)
}

func TestSettledCityGetsStartCellsAndCityLimit(t *testing.T) {
	e := newLayoutEnv(t, pack.CityLayout{StartCells: 22, CapitalCells: 51, MaxByLevel: []int{22, 27}, ExtraCities: 1})
	a := e.player("founder")
	e.setBuildings(a, map[string]int{"hall": 5})
	e.setUnits(a, map[string]int64{"settler": 2})
	ax, ay := e.location(a)
	found := func(minD int) (int, int) {
		var x, y int
		if err := e.s.DB.QueryRow(ctx, `SELECT x, y FROM tiles WHERE world_id=$1 AND neutral_id IS NULL AND city_id IS NULL AND terrain IN ('plains','forest')
			AND abs(x-$2) + abs(y-$3) BETWEEN $4 AND $5 ORDER BY x, y LIMIT 1`, e.world, ax, ay, minD, minD+3).Scan(&x, &y); err != nil {
			t.Fatal(err)
		}
		return x, y
	}
	send := func(x, y int) *game.Mission {
		return e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) {
			return c.SendMission(ci, game.MissionRequest{Kind: "settle", X: x, Y: y, Units: map[string]int64{"settler": 1}, Cargo: pack.Amounts{}})
		}).(*game.Mission)
	}
	x, y := found(3)
	m := send(x, y)
	e.clk.Set(m.ArriveAt)
	e.mustCity(a, func(c game.PCtx, ci *game.City) (any, error) { return nil, nil }) // settle the arrival
	var second int64
	if err := e.s.DB.QueryRow(ctx, `SELECT id FROM cities WHERE player_id=$1 AND NOT is_capital`, a.pid).Scan(&second); err != nil {
		t.Fatalf("second city not founded: %v", err)
	}
	e.assertCityMirror(second, 22) // new city: 22 cells, not the capital's 51
	// Limit is capital + 1 extra: dispatching a third settle is refused.
	x, y = found(7)
	_, err := e.city(a, func(c game.PCtx, ci *game.City) (any, error) {
		return c.SendMission(ci, game.MissionRequest{Kind: "settle", X: x, Y: y, Units: map[string]int64{"settler": 1}, Cargo: pack.Amounts{}})
	})
	if code(err) != "city_limit" {
		t.Fatalf("want city_limit, got %v", err)
	}
}
