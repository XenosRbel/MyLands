package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

// Player is the in-memory aggregate root of one player in one world.
type Player struct {
	ID             int64            `json:"id"`
	WorldID        int64            `json:"worldId"`
	AccountID      int64            `json:"-"`
	Name           string           `json:"name"`
	Race           string           `json:"race"`
	Techs          map[string]int   `json:"techs"`
	Stats          map[string]int64 `json:"stats"`
	Score          int64            `json:"score"`
	ProtectedUntil *time.Time       `json:"protectedUntil"`
	AllianceID     *int64           `json:"allianceId"`
	AllianceRole   *string          `json:"allianceRole"`
	CreatedAt      time.Time        `json:"createdAt"`
}

// City holds a city's persisted state.
type City struct {
	ID        int64              `json:"id"`
	PlayerID  int64              `json:"playerId"`
	Name      string             `json:"name"`
	X         int                `json:"x"`
	Y         int                `json:"y"`
	IsCapital bool               `json:"isCapital"`
	Res       map[string]int64   `json:"-"` // milli
	Pop       int64              `json:"population"`
	UpdatedAt time.Time          `json:"-"`
	Buildings map[string]int     `json:"buildings"`
	Units     map[string]int64   `json:"units"`
	Carry     map[string]float64 `json:"-"`
	CreatedAt time.Time          `json:"createdAt"`
	away      float64            // upkeep of units away (cached)
}

// pctx is a locked, settled player context inside a transaction.
type pctx struct {
	s      *Service
	tx     pgx.Tx
	ctx    context.Context
	p      *pack.Loaded
	pl     *Player
	cities []*City
	now    time.Time
	mods   *rules.Mods // race + techs (player-wide)
	failed int64       // event id that failed (for worker bookkeeping)
}

func (c *pctx) city(id int64) *City {
	for _, x := range c.cities {
		if x.ID == id {
			return x
		}
	}
	return nil
}

func (c *pctx) capital() *City {
	for _, x := range c.cities {
		if x.IsCapital {
			return x
		}
	}
	if len(c.cities) > 0 {
		return c.cities[0]
	}
	return nil
}

const playerCols = `id, world_id, account_id, name, race_id, techs, stats, score, protected_until, alliance_id, alliance_role, created_at`

func scanPlayer(row pgx.Row) (*Player, error) {
	p := &Player{}
	err := row.Scan(&p.ID, &p.WorldID, &p.AccountID, &p.Name, &p.Race, &p.Techs, &p.Stats, &p.Score,
		&p.ProtectedUntil, &p.AllianceID, &p.AllianceRole, &p.CreatedAt)
	if p.Techs == nil {
		p.Techs = map[string]int{}
	}
	if p.Stats == nil {
		p.Stats = map[string]int64{}
	}
	return p, err
}

const cityCols = `id, player_id, name, x, y, is_capital, resources, population, econ_updated_at, buildings, units, econ_carry, created_at`

func scanCity(row pgx.Row) (*City, error) {
	c := &City{}
	err := row.Scan(&c.ID, &c.PlayerID, &c.Name, &c.X, &c.Y, &c.IsCapital, &c.Res, &c.Pop, &c.UpdatedAt,
		&c.Buildings, &c.Units, &c.Carry, &c.CreatedAt)
	if c.Res == nil {
		c.Res = map[string]int64{}
	}
	if c.Buildings == nil {
		c.Buildings = map[string]int{}
	}
	if c.Units == nil {
		c.Units = map[string]int64{}
	}
	if c.Carry == nil {
		c.Carry = map[string]float64{}
	}
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, err
}

// lockPlayer locks the player and all their cities, processes every pending
// event due up to `upTo` in (due_at, id) order — advancing the economy of all
// cities to each event's instant before applying it — and finally advances all
// cities to `upTo`. Event effects and status changes commit atomically with
// the caller's transaction, so an event can never be applied twice.
func (s *Service) lockPlayer(ctx context.Context, tx pgx.Tx, playerID int64, upTo time.Time) (*pctx, error) {
	pl, err := scanPlayer(tx.QueryRow(ctx, `SELECT `+playerCols+` FROM players WHERE id=$1 FOR UPDATE`, playerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notFound("player")
		}
		return nil, err
	}
	return s.settleLocked(ctx, tx, pl, upTo)
}

func (s *Service) settleLocked(ctx context.Context, tx pgx.Tx, pl *Player, upTo time.Time) (*pctx, error) {
	p, err := s.Pack(ctx, tx, pl.WorldID)
	if err != nil {
		return nil, err
	}
	c := &pctx{s: s, tx: tx, ctx: ctx, p: p, pl: pl, now: upTo}
	rows, err := tx.Query(ctx, `SELECT `+cityCols+` FROM cities WHERE player_id=$1 ORDER BY id FOR UPDATE`, pl.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		ci, err := scanCity(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		c.cities = append(c.cities, ci)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	c.recomputeMods()
	for _, ci := range c.cities {
		if err := c.refreshAway(ci); err != nil {
			return nil, err
		}
	}
	for {
		type ev struct {
			id, ref int64
			kind    string
			due     time.Time
		}
		var batch []ev
		rows, err := tx.Query(ctx, `SELECT id, kind, ref_id, due_at FROM game_events
			WHERE player_id=$1 AND status='pending' AND due_at <= $2 ORDER BY due_at, id LIMIT 100 FOR UPDATE`, pl.ID, upTo)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var e ev
			if err := rows.Scan(&e.id, &e.kind, &e.ref, &e.due); err != nil {
				rows.Close()
				return nil, err
			}
			batch = append(batch, e)
		}
		rows.Close()
		if len(batch) == 0 {
			break
		}
		for _, e := range batch {
			c.advanceAll(e.due.UTC())
			c.now = e.due.UTC()
			if err := c.handleEvent(e.kind, e.ref); err != nil {
				c.failed = e.id
				return c, fmt.Errorf("event %d (%s ref %d): %w", e.id, e.kind, e.ref, err)
			}
			if _, err := tx.Exec(ctx, `UPDATE game_events SET status='done', processed_at=$2, attempts=attempts+1 WHERE id=$1`, e.id, upTo); err != nil {
				return nil, err
			}
			s.Metrics.EventsProcessed.Add(1)
		}
	}
	c.now = upTo
	c.advanceAll(upTo)
	return c, nil
}

// recomputeMods rebuilds player-wide modifiers (race + techs).
func (c *pctx) recomputeMods() {
	m := rules.NewMods()
	if r := c.p.Race(c.pl.Race); r != nil {
		m.AddAll(r.Modifiers, 1)
	}
	for id, lvl := range c.pl.Techs {
		t := c.p.Tech(id)
		if t == nil {
			continue
		}
		for _, e := range t.Effects {
			if e.Type == "modifier" && e.Modifier != nil {
				m.Add(*e.Modifier, float64(lvl))
			}
		}
	}
	c.mods = m
}

// cityMods = player mods + modifier effects of the city's buildings.
func (c *pctx) cityMods(ci *City) *rules.Mods {
	m := c.mods.Clone()
	m.Merge(c.buildingMods(ci.Buildings))
	return m
}

// buildingMods returns the modifier effects of a set of building levels.
func (c *pctx) buildingMods(buildings map[string]int) *rules.Mods {
	m := rules.NewMods()
	for id, lvl := range buildings {
		b := c.p.Building(id)
		if b == nil {
			continue
		}
		for _, e := range b.Effects {
			if e.Type == "modifier" && e.Modifier != nil {
				m.Add(*e.Modifier, float64(lvl))
			}
		}
	}
	return m
}

// refreshAway recomputes upkeep of the city's units that are away.
func (c *pctx) refreshAway(ci *City) error {
	rows, err := c.tx.Query(c.ctx, `SELECT upkeep_units FROM missions WHERE from_city_id=$1 AND status <> 'done'`, ci.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	total := 0.0
	for rows.Next() {
		var u map[string]int64
		if err := rows.Scan(&u); err != nil {
			return err
		}
		for id, n := range u {
			if def := c.p.Unit(id); def != nil {
				total += float64(n) * def.Upkeep
			}
		}
	}
	ci.away = total * c.p.Params.MissionUpkeepMult
	return rows.Err()
}

func (c *pctx) econInput(ci *City) rules.EconInput {
	return rules.EconInput{Pack: c.p, Buildings: ci.Buildings, Mods: c.cityMods(ci), Population: ci.Pop,
		Garrison: ci.Units, AwayUpkeep: ci.away}
}

func (c *pctx) advance(ci *City, to time.Time) rules.Rates {
	st := &rules.EconState{Stock: ci.Res, Population: ci.Pop, Garrison: ci.Units, Carry: ci.Carry, UpdatedAt: ci.UpdatedAt}
	r := rules.AdvanceCity(st, to, c.econInput(ci))
	ci.Pop, ci.UpdatedAt = st.Population, st.UpdatedAt
	return r
}

func (c *pctx) advanceAll(to time.Time) {
	for _, ci := range c.cities {
		c.advance(ci, to)
	}
}

// rates returns current rates (city must already be advanced to now).
func (c *pctx) rates(ci *City) rules.Rates { return rules.ComputeRates(c.econInput(ci)) }

// caps returns storage caps for Grant (nil in overflow-decay mode).
func (c *pctx) caps(ci *City) map[string]int64 {
	if c.p.Params.OverflowDecayPctPerHour > 0 {
		return nil
	}
	return c.rates(ci).Storage
}

// spend deducts cost after verifying affordability, and records the ledger.
func (c *pctx) spend(ci *City, cost pack.Amounts, reason, ref string) error {
	if !rules.CanAfford(ci.Res, cost) {
		e := rule("insufficient_resources", "not enough resources")
		e.Details = map[string]any{"missing": rules.Missing(ci.Res, cost)}
		return e
	}
	rules.Spend(ci.Res, cost)
	neg := pack.Amounts{}
	for r, n := range cost {
		if n != 0 {
			neg[r] = -n
		}
	}
	return c.ledger(ci, reason, ref, neg)
}

// grant adds resources (respecting caps) and records the ledger.
func (c *pctx) grant(ci *City, a pack.Amounts, reason, ref string) (pack.Amounts, error) {
	lost := rules.Grant(ci.Res, a, c.caps(ci))
	got := pack.Amounts{}
	for r, n := range a {
		if v := n - lost[r]; v != 0 {
			got[r] = v
		}
	}
	return lost, c.ledger(ci, reason, ref, got)
}

func (c *pctx) ledger(ci *City, reason, ref string, delta pack.Amounts) error {
	if len(delta) == 0 {
		return nil
	}
	var id int64
	err := c.tx.QueryRow(c.ctx, `INSERT INTO resource_ledger(world_id, player_id, city_id, reason, ref, at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		c.pl.WorldID, c.pl.ID, ci.ID, reason, ref, c.now).Scan(&id)
	if err != nil {
		return err
	}
	for r, n := range delta {
		if n == 0 {
			continue
		}
		if _, err := c.tx.Exec(c.ctx, `INSERT INTO resource_ledger_entries(ledger_id, resource_id, amount) VALUES ($1,$2,$3)`, id, r, n); err != nil {
			return err
		}
	}
	return nil
}

func (c *pctx) stat(key string, n int64) { c.pl.Stats[key] += n }

// notify inserts a notification for any player (no lock needed: insert-only).
func notify(ctx context.Context, q querier, playerID int64, kind string, data map[string]any) error {
	_, err := q.Exec(ctx, `INSERT INTO notifications(player_id, kind, data) VALUES ($1,$2,$3)`, playerID, kind, data)
	return err
}

// schedule inserts a scheduled event. (kind, ref, due) is unique so the same
// event cannot be scheduled twice.
func schedule(ctx context.Context, q querier, worldID, playerID int64, kind string, ref int64, due time.Time) error {
	_, err := q.Exec(ctx, `INSERT INTO game_events(world_id, player_id, kind, ref_id, due_at) VALUES ($1,$2,$3,$4,$5)`,
		worldID, playerID, kind, ref, due)
	return err
}

// score recomputes the player's power score from buildings, techs and home units.
func (c *pctx) score() int64 {
	var sc int64
	for _, ci := range c.cities {
		for id, lvl := range ci.Buildings {
			if b := c.p.Building(id); b != nil {
				sc += int64(b.Score * lvl)
			}
		}
		for id, n := range ci.Units {
			if u := c.p.Unit(id); u != nil {
				sc += int64(u.Score) * n
			}
		}
	}
	for id, lvl := range c.pl.Techs {
		if t := c.p.Tech(id); t != nil {
			sc += int64(t.Score * lvl)
		}
	}
	return sc
}

// save persists all cities and the player row.
func (c *pctx) save() error {
	for _, ci := range c.cities {
		for k, v := range ci.Units {
			if v <= 0 {
				delete(ci.Units, k)
			}
		}
		if _, err := c.tx.Exec(c.ctx, `UPDATE cities SET resources=$2, population=$3, econ_updated_at=$4, buildings=$5, units=$6, econ_carry=$7, name=$8 WHERE id=$1`,
			ci.ID, ci.Res, ci.Pop, ci.UpdatedAt, ci.Buildings, ci.Units, ci.Carry, ci.Name); err != nil {
			return err
		}
		if err := c.syncCityBuildings(ci); err != nil {
			return err
		}
	}
	c.pl.Score = c.score()
	_, err := c.tx.Exec(c.ctx, `UPDATE players SET techs=$2, stats=$3, score=$4, protected_until=$5, last_seen_at=now() WHERE id=$1`,
		c.pl.ID, c.pl.Techs, c.pl.Stats, c.pl.Score, c.pl.ProtectedUntil)
	return err
}

// handleEvent dispatches a due event. Handlers must be idempotent with respect
// to the referenced row's status (a cancelled job or recalled mission is a no-op).
func (c *pctx) handleEvent(kind string, ref int64) error {
	switch kind {
	case "job":
		return c.completeJob(ref)
	case "mission_arrive":
		return c.missionArrive(ref)
	case "mission_return":
		return c.missionReturn(ref)
	case "delivery":
		return c.deliver(ref)
	}
	return fmt.Errorf("unknown event kind %q", kind)
}

func (c *pctx) deliver(id int64) error {
	var cityID int64
	var amounts pack.Amounts
	var reason, ref string
	err := c.tx.QueryRow(c.ctx, `SELECT city_id, amounts, reason, ref FROM deliveries WHERE id=$1`, id).Scan(&cityID, &amounts, &reason, &ref)
	if err != nil {
		return err
	}
	ci := c.city(cityID)
	if ci == nil {
		ci = c.capital()
	}
	_, err = c.grant(ci, amounts, reason, ref)
	return err
}

// Process settles one player up to now (used by the worker). It skips the
// player when another transaction currently holds the lock.
func (s *Service) Process(ctx context.Context, playerID int64) (bool, error) {
	var failedEvent int64
	processed := false
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		pl, err := scanPlayer(tx.QueryRow(ctx, `SELECT `+playerCols+` FROM players WHERE id=$1 FOR UPDATE SKIP LOCKED`, playerID))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		c, err := s.settleLocked(ctx, tx, pl, s.Clock.Now())
		if err != nil {
			if c != nil {
				failedEvent = c.failed
			}
			return err
		}
		processed = true
		return c.save()
	})
	if err != nil && failedEvent != 0 {
		s.Metrics.EventsFailed.Add(1)
		// Record the failure outside the rolled-back transaction; after 5
		// attempts the event is parked as 'failed' for admin inspection so it
		// cannot block the player forever.
		_, _ = s.DB.Exec(ctx, `UPDATE game_events SET attempts=attempts+1, last_error=$2,
			status = CASE WHEN attempts+1 >= 5 THEN 'failed' ELSE status END WHERE id=$1`, failedEvent, err.Error())
	}
	return processed, err
}

// DuePlayers returns players having due events (worker scan).
func (s *Service) DuePlayers(ctx context.Context, limit int) ([]int64, error) {
	rows, err := s.DB.Query(ctx, `SELECT DISTINCT player_id FROM game_events WHERE status='pending' AND due_at <= $1 LIMIT $2`, s.Clock.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, rows.Err()
}

// RunWorker processes due events until ctx is cancelled.
func (s *Service) RunWorker(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	lastCleanup := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		ids, err := s.DuePlayers(ctx, 200)
		if err != nil {
			s.Log.Error("worker scan failed", "err", err)
			continue
		}
		for _, id := range ids {
			if _, err := s.Process(ctx, id); err != nil && ctx.Err() == nil {
				s.Log.Error("event processing failed", "player", id, "err", err)
			}
		}
		if time.Since(lastCleanup) > time.Hour {
			lastCleanup = time.Now()
			if _, err := s.DB.Exec(ctx, `DELETE FROM idempotency_keys WHERE created_at < now() - interval '24 hours'`); err != nil {
				s.Log.Warn("idempotency cleanup failed", "err", err)
			}
		}
	}
}

func jsonb(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
