package migrations_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"realmforge/server/internal/testutil"
)

func newPendingOrder(t *testing.T, db *pgxpool.Pool, wallet int64, key string, crystals int64) int64 {
	t.Helper()
	o := mustID(t, db, `INSERT INTO payment_orders(wallet_id,idempotency_key,package_code,crystals,price_minor) VALUES ($1,$2,'pack_s',$3,500) RETURNING id`, wallet, key, crystals)
	mustExec(t, db, `UPDATE payment_orders SET status='pending' WHERE id=$1`, o)
	return o
}

// race runs fn on n separate connections, released together by a barrier.
func race(t *testing.T, db *pgxpool.Pool, n int, fn func(i int, tx pgx.Tx) error) []error {
	t.Helper()
	conns := make([]*pgxpool.Conn, n)
	for i := range conns {
		c, err := db.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		conns[i] = c
		defer c.Release()
	}
	errs := make([]error, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			tx, err := conns[i].Begin(ctx)
			if err != nil {
				errs[i] = err
				return
			}
			if err := fn(i, tx); err != nil {
				_ = tx.Rollback(ctx)
				errs[i] = err
				return
			}
			errs[i] = tx.Commit(ctx)
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}

func walletState(t *testing.T, db *pgxpool.Pool, w int64) (balance int64, ledgerRows int, consistent bool) {
	t.Helper()
	if err := db.QueryRow(ctx, `SELECT a.balance, (SELECT count(*) FROM premium_ledger WHERE wallet_id=$1), a.consistent FROM premium_wallet_audit a WHERE a.wallet_id=$1`, w).Scan(&balance, &ledgerRows, &consistent); err != nil {
		t.Fatal(err)
	}
	return
}

// ML-B07: two connections confirm the same mock payment order. The wallet is credited once.
func TestConcurrentConfirmCreditsWalletOnce(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "conf")
	w := newWallet(t, db, s)

	// Guarded flow: the status update is the gate. The loser sees 0 rows after the winner commits.
	guarded := func(order int64) func(int, pgx.Tx) error {
		return func(i int, tx pgx.Tx) error {
			tag, err := tx.Exec(ctx, `UPDATE payment_orders SET status='confirmed' WHERE id=$1 AND status='pending'`, order)
			if err != nil || tag.RowsAffected() == 0 {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,payment_order_id) VALUES ($1,100,'purchase',$2,$3)`, w, fmt.Sprintf("order:%d", order), order)
			return err
		}
	}
	// Naive flow: confirm and credit without checking the previous status. The database must still refuse the second credit.
	naive := func(order int64) func(int, pgx.Tx) error {
		return func(i int, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `UPDATE payment_orders SET status='confirmed' WHERE id=$1`, order); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,payment_order_id) VALUES ($1,100,'purchase',$2,$3)`, w, fmt.Sprintf("n-%d-%d", order, i), order)
			return err
		}
	}

	const rounds = 20
	for r := 0; r < rounds; r++ {
		flow := guarded
		if r%2 == 1 {
			flow = naive
		}
		o := newPendingOrder(t, db, w, fmt.Sprintf("conc-key-%04d", r), 100)
		errs := race(t, db, 2, flow(o))
		failed := 0
		for _, e := range errs {
			if e != nil {
				failed++
				if flowName := r % 2; flowName == 0 {
					t.Fatalf("round %d: guarded flow must not fail, got %v", r, e)
				}
			}
		}
		if r%2 == 1 && failed != 1 {
			t.Fatalf("round %d: naive flow: %d failures, want exactly 1 (unique purchase per order)", r, failed)
		}
		var credits int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM premium_ledger WHERE payment_order_id=$1 AND reason='purchase'`, o).Scan(&credits); err != nil || credits != 1 {
			t.Fatalf("round %d: purchase credits for order = %d %v, want 1", r, credits, err)
		}
	}
	bal, rows, ok := walletState(t, db, w)
	if bal != 100*rounds || rows != rounds || !ok {
		t.Fatalf("balance=%d rows=%d consistent=%v, want %d %d true", bal, rows, ok, 100*rounds, rounds)
	}
}

// A purchase credit must match its order: same wallet, confirmed status, same amount.
func TestPurchaseCreditMustMatchOrder(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s1, s2 := seedPlayer(t, db, "pm1"), seedPlayer(t, db, "pm2")
	w1, w2 := newWallet(t, db, s1), newWallet(t, db, s2)
	pending := newPendingOrder(t, db, w1, "match-key-0001", 100)
	ins := `INSERT INTO premium_ledger(wallet_id,delta,reason,ref,payment_order_id) VALUES ($1,$2,'purchase',$3,$4)`
	wantCode(t, "23514", db, ins, w1, 100, "p-unconfirmed", pending) // order not confirmed yet
	mustExec(t, db, `UPDATE payment_orders SET status='confirmed' WHERE id=$1`, pending)
	wantCode(t, "23514", db, ins, w2, 100, "p-foreign", pending) // another wallet
	wantCode(t, "23514", db, ins, w1, 999, "p-amount", pending)  // amount differs from the order
	mustExec(t, db, ins, w1, 100, "p-ok", pending)
	if bal, rows, ok := walletState(t, db, w1); bal != 100 || rows != 1 || !ok {
		t.Fatalf("w1 balance=%d rows=%d consistent=%v", bal, rows, ok)
	}
	if bal, rows, _ := walletState(t, db, w2); bal != 0 || rows != 0 {
		t.Fatalf("w2 must stay empty: balance=%d rows=%d", bal, rows)
	}
}

// ML-B07: concurrent spends of one wallet never make the balance negative.
func TestConcurrentSpendsNeverGoNegative(t *testing.T) {
	db := testutil.DB(t, "migrations")
	s := seedPlayer(t, db, "spend")
	w := newWallet(t, db, s)
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,100,'admin_grant','seed')`, w)

	spend := func(amount int64, prefix string) func(int, pgx.Tx) error {
		return func(i int, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,$2,'service_spend',$3)`, w, -amount, fmt.Sprintf("%s-%d", prefix, i))
			return err
		}
	}
	count := func(errs []error) (ok int) {
		for _, e := range errs {
			if e == nil {
				ok++
			}
		}
		return
	}

	// Two spends of 80 from 100: exactly one succeeds.
	if ok := count(race(t, db, 2, spend(80, "a"))); ok != 1 {
		t.Fatalf("two spends of 80: %d succeeded, want 1", ok)
	}
	if bal, _, good := walletState(t, db, w); bal != 20 || !good {
		t.Fatalf("balance=%d consistent=%v, want 20 true", bal, good)
	}
	// Ten spends of 5 from 20: exactly four succeed and the balance ends at 0.
	if ok := count(race(t, db, 10, spend(5, "b"))); ok != 4 {
		t.Fatalf("ten spends of 5: %d succeeded, want 4", ok)
	}
	bal, rows, good := walletState(t, db, w)
	if bal != 0 || !good || rows != 1+1+4 {
		t.Fatalf("balance=%d rows=%d consistent=%v, want 0 6 true", bal, rows, good)
	}
	var minAfter int64
	if err := db.QueryRow(ctx, `SELECT min(balance_after) FROM premium_ledger WHERE wallet_id=$1`, w).Scan(&minAfter); err != nil || minAfter < 0 {
		t.Fatalf("min balance_after=%d %v", minAfter, err)
	}
	// The same spend reference sent twice at once is charged once.
	mustExec(t, db, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,50,'admin_grant','seed2')`, w)
	same := func(i int, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO premium_ledger(wallet_id,delta,reason,ref) VALUES ($1,-10,'service_spend','same-ref')`, w)
		return err
	}
	if ok := count(race(t, db, 2, same)); ok != 1 {
		t.Fatalf("same ref twice: %d succeeded, want 1", ok)
	}
	if bal, _, good := walletState(t, db, w); bal != 40 || !good {
		t.Fatalf("balance=%d consistent=%v, want 40 true", bal, good)
	}
}
