package migrations_test

import (
	"testing"

	"realmforge/server/internal/testutil"
)

func TestUnitGroupConstraints(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "army")
	city := newCity(t, db, s, 1, `{}`)
	city2 := newCity(t, db, s, 2, `{}`)
	ins := `INSERT INTO unit_groups(player_id,city_id,slot_kind,slot_index,category,unit_id,level,count) VALUES ($1,$2,$3,$4,'archer','archer_1',$5,$6)`
	mustExec(t, db, ins, s.player, city, "garrison", 0, 1, 10)
	mustExec(t, db, ins, s.player, city, "garrison", 1, 1, 0)          // count 0 is allowed
	wantCode(t, "23505", db, ins, s.player, city, "garrison", 0, 2, 5) // duplicate garrison slot
	mustExec(t, db, ins, s.player, city, "shelter", 0, 1, 5)           // shelter slot is independent
	wantCode(t, "23505", db, ins, s.player, city, "shelter", 0, 1, 5)
	mustExec(t, db, ins, s.player, city2, "garrison", 0, 1, 5) // other city
	wantCode(t, "23514", db, ins, s.player, city, "garrison", 2, 1, -1)
	wantCode(t, "23514", db, ins, s.player, city, "garrison", 2, 0, 1)
	wantCode(t, "23514", db, ins, s.player, city, "camp", 2, 1, 1)
	wantCode(t, "23514", db, ins, s.player, city, "mission", 0, 1, 1) // mission slot needs a mission
	s2 := seedPlayer(t, db, "army2")
	wantCode(t, "23503", db, ins, s2.player, city, "garrison", 5, 1, 1) // foreign city

	m := mustID(t, db, `INSERT INTO missions(world_id,player_id,from_city_id,kind,target_x,target_y,depart_at,arrive_at,status) VALUES ($1,$2,$3,'attack',5,5,now(),now()+interval '1 hour','outbound') RETURNING id`, s.world, s.player, city)
	msIns := `INSERT INTO unit_groups(player_id,city_id,mission_id,slot_kind,slot_index,category,unit_id,level,count) VALUES ($1,$2,$3,$4,$5,'archer','archer_1',1,3)`
	mustExec(t, db, msIns, s.player, city, m, "mission", 0)
	wantCode(t, "23505", db, msIns, s.player, city, m, "mission", 0)
	wantCode(t, "23514", db, msIns, s.player, city, m, "garrison", 9) // garrison must not carry a mission
}

func TestHeroPlacementConstraintsAndBackfill(t *testing.T) {
	db := testutil.DB(t, "migrations")
	migrateOnly(t, db, "0001_init.sql", "0002_mylands_foundations.sql", "0003_resource_ledger_entries.sql", "0004_city_buildings.sql")
	s := seedPlayer(t, db, "hero")
	city := newCity(t, db, s, 1, `{}`)
	hero := func(name, status string) int64 {
		return mustID(t, db, `INSERT INTO heroes(world_id,player_id,home_city_id,name,class_id,status) VALUES ($1,$2,$3,$4,'k',$5) RETURNING id`, s.world, s.player, city, name, status)
	}
	idle, away, lost := hero("a", "idle"), hero("b", "mission"), hero("c", "mission")
	m := mustID(t, db, `INSERT INTO missions(world_id,player_id,from_city_id,kind,target_x,target_y,hero_id,depart_at,arrive_at,status) VALUES ($1,$2,$3,'attack',5,5,$4,now(),now()+interval '1 hour','outbound') RETURNING id`, s.world, s.player, city, away)
	migrateAll(t, db)

	var kind string
	var cid, mid *int64
	check := func(h int64, wantKind string, wantCity, wantMission bool) {
		t.Helper()
		if err := db.QueryRow(ctx, `SELECT placement, city_id, mission_id FROM hero_placements WHERE hero_id=$1`, h).Scan(&kind, &cid, &mid); err != nil {
			t.Fatal(err)
		}
		if kind != wantKind || (cid != nil) != wantCity || (mid != nil) != wantMission {
			t.Fatalf("hero %d: %s city=%v mission=%v", h, kind, cid, mid)
		}
	}
	check(idle, "garrison", true, false)
	check(away, "mission", false, true)
	check(lost, "garrison", true, false) // status mission without an open mission row

	ins := `INSERT INTO hero_placements(hero_id,player_id,placement,city_id,mission_id) VALUES ($1,$2,$3,$4,$5)`
	h4 := hero("d", "idle")
	wantCode(t, "23505", db, ins, idle, s.player, "garrison", city, nil) // one place per hero
	wantCode(t, "23514", db, ins, h4, s.player, "shelter", nil, nil)
	wantCode(t, "23514", db, ins, h4, s.player, "mission", city, nil)
	wantCode(t, "23514", db, ins, h4, s.player, "garrison", city, m)
	wantCode(t, "23503", db, ins, h4, s.player+99, "shelter", city, nil)
	mustExec(t, db, ins, h4, s.player, "shelter", city, nil)
}
