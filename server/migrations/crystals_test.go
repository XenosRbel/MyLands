package migrations_test

import (
	"testing"

	"realmforge/server/internal/testutil"
)

func TestCrystalLedgerBalanceAndGuards(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "cry")
	city := newCity(t, db, s, 1, `{}`)
	stock := mustID(t, db, `INSERT INTO crystal_stocks(player_id,crystal_type) VALUES ($1,'red') RETURNING id`, s.player)
	wantCode(t, "23505", db, `INSERT INTO crystal_stocks(player_id,crystal_type) VALUES ($1,'red')`, s.player)
	wantCode(t, "23514", db, `INSERT INTO crystal_stocks(player_id,crystal_type,quantity) VALUES ($1,'blue',-1)`, s.player)
	wantCode(t, "23514", db, `INSERT INTO crystal_stocks(player_id,crystal_type,quantity) VALUES ($1,'blue',5)`, s.player)
	wantCode(t, "23001", db, `UPDATE crystal_stocks SET quantity=-1 WHERE id=$1`, stock) // guard fires before CHECK
	wantCode(t, "23001", db, `UPDATE crystal_stocks SET quantity=50 WHERE id=$1`, stock) // direct edit

	// mirror writes a resource_ledger header plus one entry and returns the header id.
	mirror := func(ref, typ string, amount int64) int64 {
		id := mustID(t, db, `INSERT INTO resource_ledger(world_id,player_id,city_id,reason,ref,delta,at) VALUES ($1,$2,$3,'grant',$4,'{}',now()) RETURNING id`, s.world, s.player, city, ref)
		mustExec(t, db, `INSERT INTO resource_ledger_entries VALUES ($1,$2,$3)`, id, typ, amount)
		return id
	}
	ins := `INSERT INTO crystal_ledger(stock_id,resource_ledger_id,delta,reason,ref) VALUES ($1,$2,$3,$4,$5)`
	mustExec(t, db, ins, stock, mirror("q1", "red", 100), 100, "quest_reward", "q1")
	mustExec(t, db, ins, stock, mirror("b1", "red", -30), -30, "building_spend", "b1")

	// Rejected: overdraw, wrong sign, missing or mismatching mirror, repeated ref.
	wantCode(t, "23514", db, ins, stock, mirror("b2", "red", -71), -71, "building_spend", "b2")
	wantCode(t, "23514", db, ins, stock, mirror("b3", "red", -5), -5, "quest_reward", "b3")
	wantCode(t, "23503", db, ins, stock, mirror("g1", "blue", 5), 5, "admin_grant", "g1")
	wantCode(t, "23503", db, ins, stock, mirror("g2", "red", 7), 8, "admin_grant", "g2")
	wantCode(t, "23503", db, ins, stock, 999999, 5, "admin_grant", "g3")
	wantCode(t, "23505", db, ins, stock, mirror("q1b", "red", 100), 100, "quest_reward", "q1")

	var qty, sum int64
	var ok bool
	if err := db.QueryRow(ctx, `SELECT s.quantity, a.ledger_sum, a.consistent FROM crystal_stocks s JOIN crystal_stock_audit a ON a.stock_id=s.id WHERE s.id=$1`, stock).Scan(&qty, &sum, &ok); err != nil {
		t.Fatal(err)
	}
	if qty != 70 || sum != 70 || !ok {
		t.Fatalf("quantity=%d sum=%d consistent=%v, want 70 70 true", qty, sum, ok)
	}
	var last int64
	if err := db.QueryRow(ctx, `SELECT balance_after FROM crystal_ledger WHERE stock_id=$1 ORDER BY id DESC LIMIT 1`, stock).Scan(&last); err != nil || last != 70 {
		t.Fatalf("balance_after=%d %v", last, err)
	}
	wantCode(t, "23001", db, `UPDATE crystal_ledger SET ref='z' WHERE stock_id=$1`, stock)
	wantCode(t, "23001", db, `DELETE FROM crystal_ledger WHERE stock_id=$1`, stock)
	wantCode(t, "23001", db, `TRUNCATE crystal_ledger`)

	// A foreign player's ledger header cannot back this stock.
	s2 := seedPlayer(t, db, "cry2")
	c2 := newCity(t, db, s2, 1, `{}`)
	foreign := mustID(t, db, `INSERT INTO resource_ledger(world_id,player_id,city_id,reason,ref,delta,at) VALUES ($1,$2,$3,'grant','f','{}',now()) RETURNING id`, s2.world, s2.player, c2)
	mustExec(t, db, `INSERT INTO resource_ledger_entries VALUES ($1,'red',5)`, foreign)
	wantCode(t, "23503", db, ins, stock, foreign, 5, "admin_grant", "f")
}

func TestHeroArtefactSlotsAndRunes(t *testing.T) {
	db := testutil.DB(t, "migrations")
	migrateOnly(t, db, "0001_init.sql", "0002_mylands_foundations.sql", "0003_resource_ledger_entries.sql", "0004_city_buildings.sql", "0005_army.sql")
	s := seedPlayer(t, db, "gear")
	city := newCity(t, db, s, 1, `{}`)
	hero := mustID(t, db, `INSERT INTO heroes(world_id,player_id,home_city_id,name,class_id) VALUES ($1,$2,$3,'h','k') RETURNING id`, s.world, s.player, city)
	legacy := mustID(t, db, `INSERT INTO items(world_id,player_id,item_def,hero_id,slot,source) VALUES ($1,$2,'sword',$3,'weapon','t') RETURNING id`, s.world, s.player, hero)
	spare := mustID(t, db, `INSERT INTO items(world_id,player_id,item_def,source) VALUES ($1,$2,'ring','t') RETURNING id`, s.world, s.player)
	migrateAll(t, db)

	count := func() (n int) {
		if err := db.QueryRow(ctx, `SELECT count(*) FROM hero_artefact_slots WHERE hero_id=$1`, hero).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return
	}
	if count() != 1 {
		t.Fatalf("backfill: %d slots, want 1", count())
	}
	// The trigger keeps the table in step with released code writing items.
	mustExec(t, db, `UPDATE items SET hero_id=$2, slot='ring' WHERE id=$1`, spare, hero)
	if count() != 2 {
		t.Fatalf("equip: %d slots, want 2", count())
	}
	mustExec(t, db, `UPDATE items SET hero_id=NULL, slot=NULL WHERE id=$1`, legacy)
	if count() != 1 {
		t.Fatalf("unequip: %d slots, want 1", count())
	}
	third := mustID(t, db, `INSERT INTO items(world_id,player_id,item_def,source) VALUES ($1,$2,'helm','t') RETURNING id`, s.world, s.player)
	wantCode(t, "23505", db, `INSERT INTO hero_artefact_slots(hero_id,slot,item_id) VALUES ($1,'ring',$2)`, hero, third) // slot unique per hero
	wantCode(t, "23505", db, `UPDATE items SET hero_id=$2, slot='ring' WHERE id=$1`, third, hero)

	// Runes: five sockets per artefact.
	for i := 1; i <= 5; i++ {
		mustExec(t, db, `INSERT INTO item_runes(item_id,socket,rune_id) VALUES ($1,$2,'fire')`, spare, i)
	}
	wantCode(t, "23514", db, `INSERT INTO item_runes(item_id,socket,rune_id) VALUES ($1,6,'fire')`, spare)
	wantCode(t, "23505", db, `INSERT INTO item_runes(item_id,socket,rune_id) VALUES ($1,5,'ice')`, spare)
	wantCode(t, "23514", db, `INSERT INTO item_runes(item_id,socket,rune_id) VALUES ($1,0,'fire')`, third)
	wantCode(t, "23503", db, `INSERT INTO item_runes(item_id,socket,rune_id) VALUES (999999,1,'fire')`)
	mustExec(t, db, `DELETE FROM items WHERE id=$1`, spare) // runes and slot rows follow the item
	var n int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM item_runes WHERE item_id=$1)+(SELECT count(*) FROM hero_artefact_slots WHERE item_id=$1)`, spare).Scan(&n); err != nil || n != 0 {
		t.Fatalf("orphans %d %v", n, err)
	}
}
