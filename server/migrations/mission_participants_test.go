package migrations_test

import (
	"testing"
	"time"

	"realmforge/server/internal/testutil"
)

func TestMissionParticipantsAndReports(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "mp")
	city := newCity(t, db, s, 1, `{}`)
	ally := extraPlayer(t, db, s, "mp-ally")
	allyCity := mustID(t, db, `INSERT INTO cities(world_id,player_id,name,x,y,resources,population,econ_updated_at) VALUES ($1,$2,'c',9,9,'{}',10,now()) RETURNING id`, s.world, ally)
	foe := seedPlayer(t, db, "mp-foe")
	foeCity := newCity(t, db, foe, 5, `{}`)

	now := time.Now()
	ins := `INSERT INTO missions(world_id,player_id,from_city_id,kind,target_x,target_y,target_city_id,depart_at,arrive_at,status) VALUES ($1,$2,$3,'attack',5,5,$4,$5,$6,'outbound') RETURNING id`
	m := mustID(t, db, ins, s.world, s.player, city, foeCity, now, now.Add(time.Hour))

	// Released code inserts only the mission: the initiator row appears by trigger.
	var role string
	var pid, cid int64
	if err := db.QueryRow(ctx, `SELECT role, player_id, city_id FROM mission_participants WHERE mission_id=$1`, m).Scan(&role, &pid, &cid); err != nil {
		t.Fatal(err)
	}
	if role != "initiator" || pid != s.player || cid != city {
		t.Fatalf("initiator row = %s %d %d", role, pid, cid)
	}

	// Sender and its city must match; the sender's city cannot belong to another player.
	wantCode(t, "23503", db, ins, s.world, s.player, foeCity, foeCity, now, now.Add(time.Hour))
	wantCode(t, "23514", db, ins, s.world, s.player, city, foeCity, now, now.Add(-time.Minute)) // arrival before departure on the mission

	p := `INSERT INTO mission_participants(mission_id,player_id,world_id,role,city_id,depart_at,arrive_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	mustExec(t, db, p, m, ally, s.world, "joined", allyCity, now, now.Add(2*time.Hour))
	wantCode(t, "23505", db, p, m, ally, s.world, "support", allyCity, now, now.Add(time.Hour))      // a player takes part once
	wantCode(t, "23505", db, p, m, s.player, s.world, "joined", city, now, now.Add(time.Hour))       // the sender is already in
	wantCode(t, "23503", db, p, m, foe.player, s.world, "support", foeCity, now, now.Add(time.Hour)) // player of another world
	p3 := extraPlayer(t, db, s, "mp-p3")
	c3 := mustID(t, db, `INSERT INTO cities(world_id,player_id,name,x,y,resources,population,econ_updated_at) VALUES ($1,$2,'c',8,8,'{}',10,now()) RETURNING id`, s.world, p3)
	wantCode(t, "23514", db, p, m, p3, s.world, "support", c3, now, now.Add(-time.Second)) // arrival before departure
	wantCode(t, "23514", db, p, m, p3, s.world, "bogus", c3, now, now)
	wantCode(t, "23503", db, p, m, p3, s.world, "support", allyCity, now, now) // city of another player
	wantCode(t, "23514", db, p, m, p3, s.world, "initiator", c3, now, now)     // only the sender can be the initiator
	mustExec(t, db, p, m, p3, s.world, "support", c3, now, now)                // equal times are allowed
	wantCode(t, "23001", db, `UPDATE mission_participants SET role='joined' WHERE mission_id=$1 AND player_id=$2`, m, s.player)
	wantCode(t, "23503", db, p, 999999, p3, s.world, "joined", c3, now, now)

	// Reports link to missions many-to-many.
	rep := mustID(t, db, `INSERT INTO reports(world_id,kind,occurred_at,data) VALUES ($1,'battle',now(),'{}') RETURNING id`, s.world)
	m2 := mustID(t, db, ins, s.world, s.player, city, foeCity, now, now.Add(time.Hour))
	mustExec(t, db, `INSERT INTO mission_reports(report_id,mission_id) VALUES ($1,$2),($1,$3)`, rep, m, m2)
	wantCode(t, "23505", db, `INSERT INTO mission_reports(report_id,mission_id) VALUES ($1,$2)`, rep, m)
	wantCode(t, "23503", db, `INSERT INTO mission_reports(report_id,mission_id) VALUES ($1,999999)`, rep)
	wantCode(t, "23503", db, `INSERT INTO mission_reports(report_id,mission_id) VALUES (999999,$1)`, m)

	// Deleting a mission removes participants and report links, not the report.
	mustExec(t, db, `DELETE FROM missions WHERE id=$1`, m)
	var parts, links, reps int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM mission_participants WHERE mission_id=$1),(SELECT count(*) FROM mission_reports WHERE mission_id=$1),(SELECT count(*) FROM reports WHERE id=$2)`, m, rep).Scan(&parts, &links, &reps); err != nil {
		t.Fatal(err)
	}
	if parts != 0 || links != 0 || reps != 1 {
		t.Fatalf("participants=%d links=%d reports=%d, want 0 0 1", parts, links, reps)
	}
}

func TestMissionInitiatorBackfill(t *testing.T) {
	db := testutil.DB(t, "migrations")
	migrateOnly(t, db, "0001_init.sql", "0002_mylands_foundations.sql", "0003_resource_ledger_entries.sql", "0004_city_buildings.sql", "0005_army.sql", "0006_crystals_runes.sql", "0007_clans_alliances_castles.sql")
	s := seedPlayer(t, db, "mpbf")
	city := newCity(t, db, s, 1, `{}`)
	now := time.Now()
	for i := 0; i < 2; i++ {
		mustExec(t, db, `INSERT INTO missions(world_id,player_id,from_city_id,kind,target_x,target_y,depart_at,arrive_at,status) VALUES ($1,$2,$3,'scout',2,2,$4,$5,'done')`, s.world, s.player, city, now, now.Add(time.Minute))
	}
	migrateAll(t, db)
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM mission_participants WHERE role='initiator' AND player_id=$1 AND city_id=$2 AND arrive_at = depart_at + interval '1 minute'`, s.player, city).Scan(&n); err != nil || n != 2 {
		t.Fatalf("initiator rows=%d %v, want 2", n, err)
	}
}
