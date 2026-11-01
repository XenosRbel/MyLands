package migrations_test

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/testutil"
)

func newClan(t *testing.T, db *pgxpool.Pool, world int64, tag string) int64 {
	t.Helper()
	return mustID(t, db, `INSERT INTO clans(world_id,name,tag) VALUES ($1,$2,$2) RETURNING id`, world, tag)
}

// extraPlayer adds one more player to the world of s.
func extraPlayer(t *testing.T, db *pgxpool.Pool, s seed, tag string) int64 {
	t.Helper()
	acc := mustID(t, db, `INSERT INTO accounts(display_name) VALUES ($1) RETURNING id`, tag)
	return mustID(t, db, `INSERT INTO players(world_id,account_id,name,race_id) VALUES ($1,$2,$3,'r') RETURNING id`, s.world, acc, tag)
}

func TestClanMembershipRolesAndCapacity(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "cl1")
	other := seedPlayer(t, db, "cl1other")
	clan := newClan(t, db, s.world, "alpha")
	clan2 := newClan(t, db, s.world, "beta")
	wantCode(t, "23505", db, `INSERT INTO clans(world_id,name,tag) VALUES ($1,'ALPHA','x')`, s.world)
	wantCode(t, "23514", db, `INSERT INTO clans(world_id,name,tag,level) VALUES ($1,'l0','l0',0)`, s.world)
	wantCode(t, "23514", db, `INSERT INTO clans(world_id,name,tag,capacity) VALUES ($1,'c0','c0',0)`, s.world)

	ins := `INSERT INTO clan_members(player_id,clan_id,world_id,role_id) VALUES ($1,$2,$3,$4)`
	mustExec(t, db, ins, s.player, clan, s.world, "leader")
	wantCode(t, "23505", db, ins, s.player, clan2, s.world, "member") // one clan per player
	wantCode(t, "23503", db, ins, extraPlayer(t, db, s, "p-bad"), clan, s.world, "chancellor")
	wantCode(t, "23503", db, ins, other.player, clan, other.world, "member") // player and clan in different worlds

	// Singleton roles: a second leader fails (checked at commit, here per statement).
	p2 := extraPlayer(t, db, s, "p2")
	wantCode(t, "23505", db, ins, p2, clan, s.world, "leader")
	mustExec(t, db, ins, p2, clan, s.world, "treasurer")
	p3 := extraPlayer(t, db, s, "p3")
	wantCode(t, "23505", db, ins, p3, clan, s.world, "treasurer")
	mustExec(t, db, ins, p3, clan, s.world, "officer")
	mustExec(t, db, `INSERT INTO clan_members(player_id,clan_id,world_id) VALUES ($1,$2,$3)`, extraPlayer(t, db, s, "p4"), clan, s.world) // default role member

	// A leader hand-over in one transaction passes through two leaders and is accepted at commit.
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE clan_members SET role_id='leader' WHERE player_id=$1`, p3); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE clan_members SET role_id='officer' WHERE player_id=$1`, s.player); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("hand-over: %v", err)
	}

	// Capacity: 4 members now. Lowering below that fails; lowering to 4 and adding a 5th fails.
	wantCode(t, "23514", db, `UPDATE clans SET capacity=3 WHERE id=$1`, clan)
	mustExec(t, db, `UPDATE clans SET capacity=4 WHERE id=$1`, clan)
	wantCode(t, "23514", db, ins, extraPlayer(t, db, s, "p5"), clan, s.world, "member")
	mustExec(t, db, `UPDATE clans SET capacity=NULL WHERE id=$1`, clan)
	mustExec(t, db, ins, extraPlayer(t, db, s, "p6"), clan, s.world, "member")
	// Moving a member into a full clan fails.
	mustExec(t, db, `UPDATE clans SET capacity=1 WHERE id=$1`, clan2)
	mustExec(t, db, ins, extraPlayer(t, db, s, "p7"), clan2, s.world, "leader")
	wantCode(t, "23514", db, `UPDATE clan_members SET clan_id=$1 WHERE player_id=$2`, clan2, p2)
	// Deleting a player removes the membership.
	mustExec(t, db, `DELETE FROM clan_members WHERE player_id=$1`, p2)
}

func TestClanSyncFromReleasedTables(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "sync")
	p2 := extraPlayer(t, db, s, "sync2")
	al := mustID(t, db, `INSERT INTO alliances(world_id,name,tag) VALUES ($1,'Old','OLD') RETURNING id`, s.world)
	mustExec(t, db, `UPDATE players SET alliance_id=$2, alliance_role='leader' WHERE id=$1`, s.player, al)
	mustExec(t, db, `UPDATE players SET alliance_id=$2, alliance_role='member' WHERE id=$1`, p2, al)
	// Released leader transfer: new leader first, then the old one demotes. Must not fail.
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []struct {
		id   int64
		role string
	}{{p2, "leader"}, {s.player, "officer"}} {
		if _, err := tx.Exec(ctx, `UPDATE players SET alliance_role=$2 WHERE id=$1`, q.id, q.role); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("released leader transfer: %v", err)
	}
	var role string
	var n int
	if err := db.QueryRow(ctx, `SELECT role_id FROM clan_members WHERE player_id=$1`, p2).Scan(&role); err != nil || role != "leader" {
		t.Fatalf("role=%q %v", role, err)
	}
	mustExec(t, db, `UPDATE alliances SET disbanded_at=now() WHERE id=$1`, al)
	mustExec(t, db, `UPDATE players SET alliance_id=NULL, alliance_role=NULL WHERE alliance_id=$1`, al)
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM clan_members) + (SELECT count(*) FROM clans WHERE disbanded_at IS NOT NULL)`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("members+disbanded=%d %v, want 1", n, err)
	}
}

func TestClanBackfillFromReleasedTables(t *testing.T) {
	db := testutil.DB(t, "migrations")
	migrateOnly(t, db, "0001_init.sql", "0002_mylands_foundations.sql", "0003_resource_ledger_entries.sql", "0004_city_buildings.sql", "0005_army.sql", "0006_crystals_runes.sql")
	s := seedPlayer(t, db, "bf")
	p2 := extraPlayer(t, db, s, "bf2")
	a := mustID(t, db, `INSERT INTO alliances(world_id,name,tag) VALUES ($1,'A','A') RETURNING id`, s.world)
	var others []int64
	for i := 0; i < 5; i++ {
		others = append(others, mustID(t, db, `INSERT INTO alliances(world_id,name,tag) VALUES ($1,$2,$2) RETURNING id`, s.world, fmt.Sprintf("o%d", i)))
	}
	mustExec(t, db, `UPDATE players SET alliance_id=$2, alliance_role='leader' WHERE id=$1`, s.player, a)
	mustExec(t, db, `UPDATE players SET alliance_id=$2 WHERE id=$1`, p2, a)
	for i, o := range others { // 5 'war' relations from A: only 3 fit the limit
		mustExec(t, db, `INSERT INTO alliance_relations(alliance_id,other_id,status) VALUES ($1,$2,'war')`, a, o)
		_ = i
	}
	mustExec(t, db, `INSERT INTO alliance_relations(alliance_id,other_id,status) VALUES ($1,$2,'ally')`, others[0], others[1])
	migrateAll(t, db)
	var clans, members, pacts int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM clans),(SELECT count(*) FROM clan_members),(SELECT count(*) FROM clan_pacts)`).Scan(&clans, &members, &pacts); err != nil {
		t.Fatal(err)
	}
	if clans != 6 || members != 2 || pacts != 3 {
		t.Fatalf("clans=%d members=%d pacts=%d, want 6 2 3", clans, members, pacts)
	}
	var role string
	if err := db.QueryRow(ctx, `SELECT role_id FROM clan_members WHERE player_id=$1`, p2).Scan(&role); err != nil || role != "member" {
		t.Fatalf("null alliance_role becomes member: %q %v", role, err)
	}
	// Identity continues after the copied ids.
	if id := newClan(t, db, s.world, "fresh"); id <= others[4] {
		t.Fatalf("new clan id %d collides with copied ids", id)
	}
}

func TestPlayerAllianceLimitAndBreaking(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "al")
	al := mustID(t, db, `INSERT INTO player_alliances(world_id) VALUES ($1) RETURNING id`, s.world)
	ins := `INSERT INTO player_alliance_members(alliance_id,player_id,world_id) VALUES ($1,$2,$3)`
	mustExec(t, db, ins, al, s.player, s.world)
	wantCode(t, "23505", db, ins, al, s.player, s.world)
	other := seedPlayer(t, db, "alx")
	wantCode(t, "23503", db, ins, al, other.player, s.world) // player of another world
	var ps []int64
	for i := 0; i < 7; i++ {
		p := extraPlayer(t, db, s, fmt.Sprintf("al%d", i))
		ps = append(ps, p)
		if i < 6 {
			mustExec(t, db, ins, al, p, s.world)
		}
	}
	wantCode(t, "23514", db, ins, al, ps[6], s.world) // the 8th member

	// Breaking: both timestamps are required and the deadline is exactly 24 hours.
	wantCode(t, "23514", db, `UPDATE player_alliances SET status='breaking' WHERE id=$1`, al)
	wantCode(t, "23514", db, `UPDATE player_alliances SET status='breaking', breaking_started_at=now(), break_deadline=now()+interval '23 hours' WHERE id=$1`, al)
	wantCode(t, "23514", db, `UPDATE player_alliances SET status='active', breaking_started_at=now(), break_deadline=now()+interval '24 hours' WHERE id=$1`, al)
	wantCode(t, "23001", db, `UPDATE player_alliances SET status='dissolved', dissolved_at=now(), breaking_started_at=now()-interval '2 days', break_deadline=now()-interval '1 day' WHERE id=$1`, al) // must pass breaking
	mustExec(t, db, `UPDATE player_alliances SET status='breaking', breaking_started_at=now(), break_deadline=now()+interval '24 hours' WHERE id=$1`, al)
	wantCode(t, "23514", db, ins, al, ps[6], s.world)                                                                                      // no new members while breaking
	wantCode(t, "23001", db, `UPDATE player_alliances SET status='active', breaking_started_at=NULL, break_deadline=NULL WHERE id=$1`, al) // cannot cancel
	wantCode(t, "23514", db, `UPDATE player_alliances SET status='dissolved', dissolved_at=now() WHERE id=$1`, al)                         // before the deadline
	mustExec(t, db, `UPDATE player_alliances SET status='dissolved', dissolved_at=now()+interval '25 hours' WHERE id=$1`, al)
	wantCode(t, "23001", db, `UPDATE player_alliances SET status='breaking' WHERE id=$1`, al)
	var d string
	if err := db.QueryRow(ctx, `SELECT extract(epoch FROM break_deadline - breaking_started_at)::bigint::text FROM player_alliances WHERE id=$1`, al).Scan(&d); err != nil || d != "86400" {
		t.Fatalf("stored term %q %v", d, err)
	}
}

func TestClanPactsLimitThreePerKind(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "pact")
	a := newClan(t, db, s.world, "pa")
	var o []int64
	for i := 0; i < 5; i++ {
		o = append(o, newClan(t, db, s.world, fmt.Sprintf("po%d", i)))
	}
	ins := `INSERT INTO clan_pacts(world_id,kind,clan_id,other_clan_id) VALUES ($1,$2,$3,$4)`
	wantCode(t, "23514", db, ins, s.world, "war", a, a)
	wantCode(t, "23514", db, ins, s.world, "truce", a, o[0])
	for i := 0; i < 3; i++ {
		mustExec(t, db, ins, s.world, "peace", a, o[i])
	}
	wantCode(t, "23514", db, ins, s.world, "peace", a, o[3]) // a is at 3 peace pacts
	wantCode(t, "23514", db, ins, s.world, "peace", o[3], a) // the limit counts both sides
	wantCode(t, "23505", db, ins, s.world, "war", o[0], a)   // one active pact per pair, either kind
	mustExec(t, db, ins, s.world, "war", a, o[3])            // war has its own limit of 3
	mustExec(t, db, ins, s.world, "war", o[4], a)
	mustExec(t, db, `UPDATE clan_pacts SET ended_at=now() WHERE clan_id=$1 AND other_clan_id=$2`, a, o[0])
	mustExec(t, db, ins, s.world, "peace", a, o[0]) // a slot is free again after a treaty ends
	otherWorld := seedPlayer(t, db, "pactx")
	foreign := newClan(t, db, otherWorld.world, "pf")
	wantCode(t, "23503", db, ins, s.world, "war", a, foreign)
}

func TestClanCastlesLimitAndStandards(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "cast")
	clan := newClan(t, db, s.world, "ca")
	clan2 := newClan(t, db, s.world, "cb")
	ins := `INSERT INTO clan_castles(world_id,castle_key,clan_id,captured_at) VALUES ($1,$2,$3,now())`
	var first int64
	for i := 0; i < 6; i++ {
		id := mustID(t, db, ins+` RETURNING id`, s.world, fmt.Sprintf("castle-%d", i), clan)
		if i == 0 {
			first = id
		}
	}
	wantCode(t, "23514", db, ins, s.world, "castle-6", clan)
	wantCode(t, "23505", db, ins, s.world, "castle-0", clan2)
	// Monster-held castle: no owner, no capture time. Owner and capture time go together.
	free := mustID(t, db, `INSERT INTO clan_castles(world_id,castle_key) VALUES ($1,'free-1') RETURNING id`, s.world)
	wantCode(t, "23514", db, `INSERT INTO clan_castles(world_id,castle_key,clan_id) VALUES ($1,'bad-1',$2)`, s.world, clan2)
	wantCode(t, "23514", db, `UPDATE clan_castles SET clan_id=$2, captured_at=now() WHERE id=$1`, free, clan) // clan is at 6
	mustExec(t, db, `UPDATE clan_castles SET clan_id=$2, captured_at=now() WHERE id=$1`, free, clan2)
	mustExec(t, db, `UPDATE clan_castles SET clan_id=$2 WHERE id=$1`, first, clan2) // transfer frees a slot in clan
	mustExec(t, db, ins, s.world, "castle-7", clan)
	wantCode(t, "23503", db, `DELETE FROM clans WHERE id=$1`, clan) // castles must be released first

	// Standards are bought per castle and kind.
	st := `INSERT INTO castle_standards(castle_id,kind,quantity) VALUES ($1,$2,$3)`
	mustExec(t, db, st, first, "peace", 4)
	mustExec(t, db, st, first, "war", 0)
	wantCode(t, "23505", db, st, first, "peace", 1)
	wantCode(t, "23514", db, st, free, "peace", -1)
	wantCode(t, "23514", db, st, free, "flag", 1)
	wantCode(t, "23503", db, st, 999999, "peace", 1)
	mustExec(t, db, `DELETE FROM clan_castles WHERE id=$1`, first)
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM castle_standards WHERE castle_id=$1`, first).Scan(&n); err != nil || n != 0 {
		t.Fatalf("standards survive castle delete: %d %v", n, err)
	}
}
