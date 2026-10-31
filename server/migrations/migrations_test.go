package migrations_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/platform"
	"realmforge/server/internal/testutil"
	"realmforge/server/migrations"
)

var ctx = context.Background()

func mustExec(t *testing.T, db *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := db.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func mustID(t *testing.T, db *pgxpool.Pool, sql string, args ...any) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return id
}

// wantCode asserts that the statement fails with the given SQLSTATE.
func wantCode(t *testing.T, code string, db *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	_, err := db.Exec(ctx, sql, args...)
	var pe *pgconn.PgError
	if !errors.As(err, &pe) {
		t.Fatalf("want SQLSTATE %s, got %v for %s", code, err, sql)
	}
	if pe.Code != code {
		t.Fatalf("want SQLSTATE %s, got %s (%s) for %s", code, pe.Code, pe.Message, sql)
	}
}

type seed struct{ account, world, player int64 }

func seedPlayer(t *testing.T, db *pgxpool.Pool, tag string) seed {
	t.Helper()
	var s seed
	s.account = mustID(t, db, `INSERT INTO accounts(display_name) VALUES ($1) RETURNING id`, tag)
	s.world = mustID(t, db, `INSERT INTO worlds(slug,name,pack_id,pack_version,pack_hash,pack_json,seed) VALUES ($1,$1,'p','1','h','{}',1) RETURNING id`, tag)
	s.player = mustID(t, db, `INSERT INTO players(world_id,account_id,name,race_id) VALUES ($1,$2,$3,'r') RETURNING id`, s.world, s.account, tag)
	return s
}

func newWallet(t *testing.T, db *pgxpool.Pool, s seed) int64 {
	t.Helper()
	return mustID(t, db, `INSERT INTO premium_wallets(account_id,world_id) VALUES ($1,$2) RETURNING id`, s.account, s.world)
}

func TestMigrationAppliesOnEmptyDatabase(t *testing.T) {
	db := testutil.DB(t, "migrations")
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM schema_migrations WHERE version IN ('0001_init.sql','0002_mylands_foundations.sql')`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("migrations recorded: %d %v", n, err)
	}
	for _, tbl := range []string{"server_profiles", "server_param_defs", "server_profile_params", "premium_wallets", "premium_ledger", "payment_orders", "service_defs", "service_instances"} {
		var ok bool
		if err := db.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, tbl).Scan(&ok); err != nil || !ok {
			t.Errorf("table %s missing", tbl)
		}
	}
	// Re-running the migrator is a no-op.
	if err := platform.Migrate(ctx, db, migrations.FS, testutil.Logger()); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationAppliesOnTopOf0001WithData(t *testing.T) {
	db := testutil.DB(t, "migrations")
	mustExec(t, db, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	first, err := migrations.FS.ReadFile("0001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	only0001 := fstest.MapFS{"0001_init.sql": {Data: first}}
	if err := platform.Migrate(ctx, db, only0001, testutil.Logger()); err != nil {
		t.Fatal(err)
	}
	s := seedPlayer(t, db, "legacy")
	city := mustID(t, db, `INSERT INTO cities(world_id,player_id,name,x,y,resources,population,econ_updated_at) VALUES ($1,$2,'c',1,1,'{}',10,now()) RETURNING id`, s.world, s.player)
	mustExec(t, db, `INSERT INTO resource_ledger(world_id,player_id,city_id,reason,ref,delta,at) VALUES ($1,$2,$3,'grant','x','{}',now())`, s.world, s.player, city)

	if err := platform.Migrate(ctx, db, migrations.FS, testutil.Logger()); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM resource_ledger`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("legacy rows kept: %d %v", n, err)
	}
	var prof *int64
	if err := db.QueryRow(ctx, `SELECT profile_id FROM worlds WHERE id=$1`, s.world).Scan(&prof); err != nil || prof != nil {
		t.Fatalf("legacy world profile must be NULL: %v %v", prof, err)
	}
	// New FK on legacy ledger applies to new writes at once.
	wantCode(t, "23503", db, `INSERT INTO resource_ledger(world_id,player_id,city_id,reason,ref,delta,at) VALUES ($1,$2,999999,'grant','y','{}',now())`, s.world, s.player)
}

func TestServerProfileParamsAreTypedAndRanged(t *testing.T) {
	db := testutil.DB(t, "migrations")
	p := mustID(t, db, `INSERT INTO server_profiles(code,name) VALUES ('combat','Combat') RETURNING id`)
	mustExec(t, db, `INSERT INTO server_param_defs(code,kind,min_value,max_value,default_value) VALUES ('mine.mult','multiplier',0.1,10,1),('queue.slots','integer',1,10,1)`)
	mustExec(t, db, `INSERT INTO server_profile_params VALUES ($1,'mine.mult',1.5)`, p)
	wantCode(t, "23514", db, `INSERT INTO server_profile_params VALUES ($1,'mine.mult',11)`, p)
	wantCode(t, "23514", db, `INSERT INTO server_profile_params VALUES ($1,'queue.slots',2.5)`, p)
	wantCode(t, "23503", db, `INSERT INTO server_profile_params VALUES ($1,'unknown.param',1)`, p)
	wantCode(t, "23505", db, `INSERT INTO server_profile_params VALUES ($1,'mine.mult',2)`, p)
	wantCode(t, "23514", db, `INSERT INTO server_profiles(code,name) VALUES ('Bad Code','x')`)
	s := seedPlayer(t, db, "prof")
	mustExec(t, db, `UPDATE worlds SET profile_id=$1 WHERE id=$2`, p, s.world)
	wantCode(t, "23503", db, `UPDATE worlds SET profile_id=999999 WHERE id=$1`, s.world)
}

func TestWalletRejectsNegativeBalanceAndDirectEdits(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "wal")
	w := newWallet(t, db, s)
	wantCode(t, "23505", db, `INSERT INTO premium_wallets(account_id,world_id) VALUES ($1,$2)`, s.account, s.world)
	wantCode(t, "23514", db, `INSERT INTO premium_wallets(account_id,world_id,balance) VALUES ($1,$2,-1)`, s.account, s.world+1)
	wantCode(t, "23001", db, `UPDATE premium_wallets SET balance=-5 WHERE id=$1`, w)
	wantCode(t, "23001", db, `UPDATE premium_wallets SET balance=50 WHERE id=$1`, w)
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,100,'admin_grant','g1')`, w)
	wantCode(t, "23514", db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,-101,'admin_revoke','r1')`, w)
	wantCode(t, "23514", db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,-5,'admin_grant','bad-sign')`, w)
}

func TestLedgerSumEqualsBalanceAndIsAppendOnly(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "led")
	w := newWallet(t, db, s)
	for i, d := range []int64{100, -30, 50, -20} {
		reason := "admin_grant"
		if d < 0 {
			reason = "admin_revoke"
		}
		mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,$2,$3,$4)`, w, d, reason, string(rune('a'+i)))
	}
	var bal, sum int64
	var ok bool
	if err := db.QueryRow(ctx, `SELECT w.balance, a.ledger_sum, a.consistent FROM premium_wallets w JOIN premium_wallet_audit a ON a.wallet_id=w.id WHERE w.id=$1`, w).Scan(&bal, &sum, &ok); err != nil {
		t.Fatal(err)
	}
	if bal != 100 || sum != 100 || !ok {
		t.Fatalf("balance=%d sum=%d consistent=%v", bal, sum, ok)
	}
	var last int64
	if err := db.QueryRow(ctx, `SELECT balance_after FROM premium_ledger WHERE wallet_id=$1 ORDER BY id DESC LIMIT 1`, w).Scan(&last); err != nil || last != 100 {
		t.Fatalf("balance_after=%d %v", last, err)
	}
	wantCode(t, "23001", db, `UPDATE premium_ledger SET ref='z' WHERE wallet_id=$1`, w)
	wantCode(t, "23001", db, `DELETE FROM premium_ledger WHERE wallet_id=$1`, w)
	wantCode(t, "23001", db, `TRUNCATE premium_ledger`)
	wantCode(t, "23505", db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,1,'admin_grant','a')`, w)
}

func TestPaymentOrderIdempotencyAndStateMachine(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "pay")
	w := newWallet(t, db, s)
	ins := `INSERT INTO payment_orders(wallet_id,idempotency_key,package_code,crystals,price_minor) VALUES ($1,$2,'pack_s',100,500) RETURNING id`
	o := mustID(t, db, ins, w, "key-00000001")
	wantCode(t, "23505", db, `INSERT INTO payment_orders(wallet_id,idempotency_key,package_code,crystals,price_minor) VALUES ($1,'key-00000001','pack_s',100,500)`, w)
	wantCode(t, "23514", db, `INSERT INTO payment_orders(wallet_id,idempotency_key,provider,package_code,crystals,price_minor) VALUES ($1,'key-00000002','stripe','p',1,1)`, w)

	wantCode(t, "23514", db, `UPDATE payment_orders SET status='confirmed' WHERE id=$1`, o) // created -> confirmed
	mustExec(t, db, `UPDATE payment_orders SET status='pending' WHERE id=$1`, o)
	wantCode(t, "23001", db, `UPDATE payment_orders SET crystals=1000000 WHERE id=$1`, o)
	mustExec(t, db, `UPDATE payment_orders SET status='confirmed' WHERE id=$1`, o)
	var has bool
	if err := db.QueryRow(ctx, `SELECT confirmed_at IS NOT NULL FROM payment_orders WHERE id=$1`, o).Scan(&has); err != nil || !has {
		t.Fatalf("confirmed_at must be set: %v %v", has, err)
	}
	wantCode(t, "23514", db, `UPDATE payment_orders SET status='pending' WHERE id=$1`, o) // terminal
	wantCode(t, "23514", db, `UPDATE payment_orders SET status='failed' WHERE id=$1`, o)  // terminal

	// Purchase credit: one per order, so a repeated confirm cannot double the balance.
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,payment_order_id) VALUES ($1,100,'purchase','order:1',$2)`, w, o)
	wantCode(t, "23505", db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,payment_order_id) VALUES ($1,100,'purchase','order:1-again',$2)`, w, o)
	wantCode(t, "23514", db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,100,'purchase','no-order')`, w)
}

func TestOneActiveServicePerGroup(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "svc")
	mustExec(t, db, `INSERT INTO service_defs(code,service_group,pack_id) VALUES ('premium_1','premium','p'),('premium_2','premium','p'),('vacation','global','p')`)
	ins := `INSERT INTO service_instances(player_id,service_code,service_group,level,price_paid,started_at,expires_at) VALUES ($1,$2,$3,$4,10,now(),now()+interval '1 day') RETURNING id`
	a := mustID(t, db, ins, s.player, "premium_1", "premium", 1)
	wantCode(t, "23505", db, ins, s.player, "premium_2", "premium", 2)
	wantCode(t, "23503", db, ins, s.player, "premium_1", "global", 1) // group must match the catalog
	mustExec(t, db, ins, s.player, "vacation", "global", 1)           // other group is independent
	wantCode(t, "23514", db, `INSERT INTO service_instances(player_id,service_code,service_group,level,price_paid,started_at,expires_at) VALUES ($1,'vacation','global',1,0,now(),now())`, s.player)

	mustExec(t, db, `UPDATE service_instances SET status='replaced', ended_at=now() WHERE id=$1`, a)
	mustExec(t, db, ins, s.player, "premium_2", "premium", 2) // slot is free after close
	wantCode(t, "23001", db, `UPDATE service_instances SET level=5 WHERE id=$1`, a)

	w := newWallet(t, db, s)
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,50,'admin_grant','g')`, w)
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,service_instance_id) VALUES ($1,-10,'service_spend','svc:1',$2)`, w, a)
}
