package game_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type cityPos struct{ x, y int }

func (e *env) cityPositions() []cityPos {
	e.t.Helper()
	rows, err := e.s.DB.Query(ctx, `SELECT x, y FROM cities WHERE world_id=$1 ORDER BY id`, e.world)
	if err != nil {
		e.t.Fatal(err)
	}
	defer rows.Close()
	var out []cityPos
	for rows.Next() {
		var c cityPos
		if err := rows.Scan(&c.x, &c.y); err != nil {
			e.t.Fatal(err)
		}
		out = append(out, c)
	}
	return out
}

// assertSpawnRules checks R-CITY-3: every capital is at Chebyshev distance >= 3
// from every other city and stands on a passable tile without a neutral.
func (e *env) assertSpawnRules() {
	e.t.Helper()
	cs := e.cityPositions()
	for i := range cs {
		for j := i + 1; j < len(cs); j++ {
			dx, dy := cs[i].x-cs[j].x, cs[i].y-cs[j].y
			if max(abs(dx), abs(dy)) < 3 {
				e.t.Fatalf("cities %v and %v are closer than 3 tiles", cs[i], cs[j])
			}
		}
		var terrain string
		var neutral *string
		if err := e.s.DB.QueryRow(ctx, `SELECT terrain, neutral_id FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`, e.world, cs[i].x, cs[i].y).Scan(&terrain, &neutral); err != nil {
			e.t.Fatal(err)
		}
		if neutral != nil || terrain == "lake" {
			e.t.Fatalf("city %v stands on %s / neutral %v", cs[i], terrain, neutral)
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func TestSpawnKeepsSpacingForSequentialJoins(t *testing.T) {
	e := newEnv(t)
	for i := 0; i < 10; i++ {
		e.player(fmt.Sprintf("seq%d", i))
	}
	e.assertSpawnRules()
}

// A burst of simultaneous joins must still respect the spacing rule. Without a
// per-world lock two transactions can pick neighbouring tiles because neither
// sees the other's uncommitted city.
func TestSpawnKeepsSpacingForConcurrentJoins(t *testing.T) {
	e := newEnv(t)
	const n = 12
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("burst%02d", i)
			sess, err := e.s.Register(ctx, name, "password123", name, time.Hour)
			if err != nil {
				errs <- err
				return
			}
			if _, err := e.s.Join(ctx, sess.Account.ID, e.world, "P_"+name, "human"); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		// world_full is acceptable on the small test map; spacing is not negotiable.
		if code(err) != "world_full" {
			t.Fatal(err)
		}
	}
	e.assertSpawnRules()
}
