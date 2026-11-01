package migrations_test

import (
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/platform"
	"realmforge/server/internal/testutil"
	"realmforge/server/migrations"
)

// migrateOnly resets the schema and applies only the named migration files.
func migrateOnly(t *testing.T, db *pgxpool.Pool, names ...string) {
	t.Helper()
	mustExec(t, db, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	fs := fstest.MapFS{}
	for _, n := range names {
		b, err := migrations.FS.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		fs[n] = &fstest.MapFile{Data: b}
	}
	if err := platform.Migrate(ctx, db, fs, testutil.Logger()); err != nil {
		t.Fatal(err)
	}
}

func migrateAll(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	if err := platform.Migrate(ctx, db, migrations.FS, testutil.Logger()); err != nil {
		t.Fatal(err)
	}
}

func newCity(t *testing.T, db *pgxpool.Pool, s seed, x int, buildings string) int64 {
	t.Helper()
	return mustID(t, db, `INSERT INTO cities(world_id,player_id,name,x,y,resources,population,econ_updated_at,buildings) VALUES ($1,$2,'c',$3,$3,'{}',10,now(),$4::jsonb) RETURNING id`, s.world, s.player, x, buildings)
}

func newCell(t *testing.T, db *pgxpool.Pool, city int64, idx int) int64 {
	t.Helper()
	return mustID(t, db, `INSERT INTO city_cells(city_id,cell_index) VALUES ($1,$2) RETURNING id`, city, idx)
}

func TestCityBuildingConstraints(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "b02")
	city := newCity(t, db, s, 1, `{}`)
	c0, c1 := newCell(t, db, city, 0), newCell(t, db, city, 1)
	wantCode(t, "23505", db, `INSERT INTO city_cells(city_id,cell_index) VALUES ($1,0)`, city)
	wantCode(t, "23514", db, `INSERT INTO city_cells(city_id,cell_index) VALUES ($1,-1)`, city)
	wantCode(t, "23514", db, `INSERT INTO city_cells(city_id,cell_index,obstacle_kind) VALUES ($1,5,'Bad Kind')`, city)
	mustExec(t, db, `INSERT INTO city_cells(city_id,cell_index,obstacle_kind) VALUES ($1,6,'rock')`, city)

	ins := `INSERT INTO city_buildings(city_id,cell_id,building_id,level) VALUES ($1,$2,$3,$4)`
	mustExec(t, db, ins, city, c0, "house", 9)
	wantCode(t, "23505", db, ins, city, c0, "tower", 1) // two buildings in one cell
	wantCode(t, "23514", db, ins, city, c1, "tower", 10)
	wantCode(t, "23514", db, ins, city, c1, "tower", 0)
	wantCode(t, "23514", db, `INSERT INTO city_buildings(city_id,cell_id,building_id,level,status) VALUES ($1,$2,'tower',1,'constructing')`, city, c1) // needs finished_at
	wantCode(t, "23514", db, `INSERT INTO city_buildings(city_id,cell_id,building_id,level,status) VALUES ($1,$2,'tower',1,'broken')`, city, c1)
	other := newCity(t, db, s, 2, `{}`)
	wantCode(t, "23503", db, ins, other, c1, "tower", 1) // cell belongs to another city
	mustExec(t, db, ins, city, c1, "house", 1)           // the same building type may repeat on other cells
}

func TestConstructionQueuePositions(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "q")
	city := newCity(t, db, s, 1, `{}`)
	var b [6]int64
	for i := range b {
		c := newCell(t, db, city, i)
		b[i] = mustID(t, db, `INSERT INTO city_buildings(city_id,cell_id,building_id,level) VALUES ($1,$2,'house',1) RETURNING id`, city, c)
	}
	ins := `INSERT INTO construction_queue(player_id,city_id,building_id,target_level,position,started_at,finishes_at) VALUES ($1,$2,$3,2,$4,now(),now()+interval '1 hour')`
	for p := 1; p <= 4; p++ {
		mustExec(t, db, ins, s.player, city, b[p-1], p)
	}
	wantCode(t, "23514", db, ins, s.player, city, b[4], 5)
	wantCode(t, "23514", db, ins, s.player, city, b[4], 0)
	wantCode(t, "23505", db, ins, s.player, city, b[4], 3) // position taken
	wantCode(t, "23514", db, `INSERT INTO construction_queue(player_id,city_id,building_id,target_level,position,started_at,finishes_at) VALUES ($1,$2,$3,10,1,now(),now()+interval '1 hour')`, s.player, city, b[4])
	// A closed row frees its position.
	mustExec(t, db, `UPDATE construction_queue SET status='done' WHERE position=3`)
	mustExec(t, db, ins, s.player, city, b[4], 3)
	// Another player has own positions and cannot queue on a foreign city.
	s2 := seedPlayer(t, db, "q2")
	wantCode(t, "23503", db, ins, s2.player, city, b[5], 1)
}

func TestCityBuildingsBackfillFromLegacy(t *testing.T) {
	db := testutil.DB(t, "migrations")
	migrateOnly(t, db, "0001_init.sql", "0002_mylands_foundations.sql", "0003_resource_ledger_entries.sql")
	s := seedPlayer(t, db, "legacy")
	city := newCity(t, db, s, 1, `{"farm":3,"house":2,"unbuilt":0,"huge":12}`)
	mustExec(t, db, `INSERT INTO jobs(world_id,player_id,city_id,kind,item_id,level,cost,starts_at,finishes_at) VALUES
		($1,$2,$3,'build','house',3,'{}',now(),now()+interval '1 hour'),
		($1,$2,$3,'build','mill',1,'{}',now(),now()+interval '2 hour'),
		($1,$2,$3,'research','tech',1,'{}',now(),now()+interval '1 hour'),
		($1,$2,$3,'build','done_one',1,'{}',now(),now()+interval '1 hour')`, s.world, s.player, city)
	mustExec(t, db, `UPDATE jobs SET status='done' WHERE item_id='done_one'`)
	migrateAll(t, db)

	rows, err := db.Query(ctx, `SELECT b.building_id, b.level, b.status, k.cell_index FROM city_buildings b JOIN city_cells k ON k.id=b.cell_id WHERE b.city_id=$1 ORDER BY k.cell_index`, city)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id, st string
		var lvl, idx int
		if err := rows.Scan(&id, &lvl, &st, &idx); err != nil {
			t.Fatal(err)
		}
		got = append(got, id+":"+st+":"+string(rune('0'+lvl))+":"+string(rune('0'+idx)))
	}
	want := []string{"farm:idle:3:0", "house:upgrading:2:1", "mill:constructing:1:2"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM construction_queue WHERE player_id=$1 AND status='queued' AND job_id IS NOT NULL`, s.player).Scan(&n); err != nil || n != 2 {
		t.Fatalf("queue rows %d %v, want 2", n, err)
	}
	var pos int
	if err := db.QueryRow(ctx, `SELECT position FROM construction_queue q JOIN city_buildings b ON b.id=q.building_id WHERE b.building_id='mill'`).Scan(&pos); err != nil || pos != 2 {
		t.Fatalf("mill position %d %v, want 2", pos, err)
	}
	// Legacy data stays in place.
	var legacy int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM jobs WHERE kind='build'`).Scan(&legacy); err != nil || legacy != 3 {
		t.Fatalf("legacy jobs %d %v", legacy, err)
	}
}
