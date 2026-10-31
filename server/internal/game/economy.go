package game

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

// ---------- Quests (GAME_RULES R-QST-1..4) ----------

type QuestView struct {
	ID       string       `json:"id"`
	Status   string       `json:"status"` // locked | active | complete | claimed
	Progress int64        `json:"progress"`
	Goal     int64        `json:"goal"`
	Rewards  pack.Amounts `json:"rewards"`
	Tutorial bool         `json:"tutorial"`
	Screen   string       `json:"screen,omitempty"`
}

func (c *pctx) questProgress(q *pack.Quest) (int64, error) {
	g := q.Goal
	switch g.Type {
	case "building_level":
		var best int
		for _, ci := range c.cities {
			best = max(best, ci.Buildings[g.ID])
		}
		return int64(best), nil
	case "tech_level":
		return int64(c.pl.Techs[g.ID]), nil
	case "units_trained":
		if g.ID != "" {
			return c.pl.Stats["units_trained."+g.ID], nil
		}
		return c.pl.Stats["units_trained"], nil
	case "missions_won", "neutrals_cleared", "market_trade", "hero_hired", "scouts_sent":
		return c.pl.Stats[g.Type], nil
	case "join_alliance":
		if c.pl.AllianceID != nil {
			return 1, nil
		}
		return 0, nil
	case "cities":
		return int64(len(c.cities)), nil
	case "territories":
		n, err := c.territoryCount()
		return int64(n), err
	}
	return 0, nil
}

func (c *pctx) claimed() (map[string]bool, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT quest_id FROM quest_progress WHERE player_id=$1`, c.pl.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (c *pctx) Quests() ([]QuestView, error) {
	done, err := c.claimed()
	if err != nil {
		return nil, err
	}
	out := []QuestView{}
	for i := range c.p.Quests {
		q := &c.p.Quests[i]
		v := QuestView{ID: q.ID, Goal: q.Goal.Amount, Rewards: q.Rewards, Tutorial: q.Tutorial, Screen: q.Screen}
		switch {
		case done[q.ID]:
			v.Status, v.Progress = "claimed", q.Goal.Amount
		default:
			locked := false
			for _, r := range q.Requires {
				if !done[r] {
					locked = true
				}
			}
			if locked {
				v.Status = "locked"
				break
			}
			p, err := c.questProgress(q)
			if err != nil {
				return nil, err
			}
			v.Progress = min(p, q.Goal.Amount)
			v.Status = "active"
			if p >= q.Goal.Amount {
				v.Status = "complete"
			}
		}
		out = append(out, v)
	}
	return out, nil
}

// ClaimQuest grants rewards once; the quest_progress primary key makes a
// second claim impossible even under concurrent requests.
func (c *pctx) ClaimQuest(id string, cityID int64) (pack.Amounts, error) {
	q := c.p.Quest(id)
	if q == nil {
		return nil, notFound("quest")
	}
	qs, err := c.Quests()
	if err != nil {
		return nil, err
	}
	for _, v := range qs {
		if v.ID != id {
			continue
		}
		switch v.Status {
		case "claimed":
			return nil, conflict("already_claimed", "quest reward already claimed")
		case "locked", "active":
			return nil, rule("quest_incomplete", "quest is not complete")
		}
	}
	tag, err := c.tx.Exec(c.ctx, `INSERT INTO quest_progress(player_id, quest_id, claimed_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, c.pl.ID, id, c.now)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, conflict("already_claimed", "quest reward already claimed")
	}
	ci := c.city(cityID)
	if ci == nil {
		ci = c.capital()
	}
	_, err = c.grant(ci, q.Rewards, "quest", id)
	return q.Rewards, err
}

// ---------- Market (GAME_RULES R-MKT-1..7) ----------

type Order struct {
	ID         int64     `json:"id"`
	PlayerID   int64     `json:"playerId"`
	Player     string    `json:"player"`
	CityID     int64     `json:"cityId"`
	SellRes    string    `json:"sellRes"`
	SellAmount int64     `json:"sellAmount"`
	BuyRes     string    `json:"buyRes"`
	BuyAmount  int64     `json:"buyAmount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (c *pctx) tradable(r string) bool {
	x := c.p.Resource(r)
	return x != nil && x.Tradable && x.Kind != "premium"
}

// PlaceOrder escrows the offered resources.
func (c *pctx) PlaceOrder(ci *City, sell string, sellAmt int64, buy string, buyAmt int64) (*Order, error) {
	if !c.tradable(sell) || !c.tradable(buy) || sell == buy {
		return nil, invalid("both resources must be distinct and tradable")
	}
	if sellAmt < 1 || buyAmt < 1 || sellAmt > 1e9 || buyAmt > 1e9 {
		return nil, invalid("amounts must be in [1, 1e9]")
	}
	// Price sanity band prevents trivially abusive "gift" orders used to push
	// resources between accounts (R-MKT-6): ratio must stay within 1:10..10:1
	// of NPC exchange value when rates are known.
	if r1, r2 := c.p.Params.ExchangeRates[sell], c.p.Params.ExchangeRates[buy]; r1 > 0 && r2 > 0 {
		ratio := (float64(buyAmt) * r2) / (float64(sellAmt) * r1)
		if ratio < 0.1 || ratio > 10 {
			return nil, rule("price_out_of_band", "price must be within 10x of the reference value")
		}
	}
	var open int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM market_orders WHERE player_id=$1 AND status='open'`, c.pl.ID).Scan(&open); err != nil {
		return nil, err
	}
	if open >= c.p.Params.MarketMaxOpenOrders {
		return nil, rule("too_many_orders", "at most %d open orders", c.p.Params.MarketMaxOpenOrders)
	}
	if err := c.spend(ci, pack.Amounts{sell: sellAmt}, "market_escrow", "new"); err != nil {
		return nil, err
	}
	o := &Order{PlayerID: c.pl.ID, Player: c.pl.Name, CityID: ci.ID, SellRes: sell, SellAmount: sellAmt, BuyRes: buy, BuyAmount: buyAmt, Status: "open", CreatedAt: c.now}
	err := c.tx.QueryRow(c.ctx, `INSERT INTO market_orders(world_id, player_id, city_id, sell_res, sell_amount, buy_res, buy_amount, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, c.pl.WorldID, c.pl.ID, ci.ID, sell, sellAmt, buy, buyAmt, c.now).Scan(&o.ID)
	return o, err
}

func (c *pctx) lockOrder(id int64) (*Order, error) {
	o := &Order{}
	err := c.tx.QueryRow(c.ctx, `SELECT id, player_id, city_id, sell_res, sell_amount, buy_res, buy_amount, status, created_at FROM market_orders WHERE id=$1 AND world_id=$2 FOR UPDATE`, id, c.pl.WorldID).
		Scan(&o.ID, &o.PlayerID, &o.CityID, &o.SellRes, &o.SellAmount, &o.BuyRes, &o.BuyAmount, &o.Status, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("order")
	}
	return o, err
}

// AcceptOrder: the buyer pays immediately and receives the goods; the
// seller's proceeds (minus fee) are delivered through a 'delivery' event to
// avoid locking two players at once. The order row lock makes double-fills impossible.
func (c *pctx) AcceptOrder(ci *City, id int64) (*Order, error) {
	o, err := c.lockOrder(id)
	if err != nil {
		return nil, err
	}
	if o.Status != "open" {
		return nil, conflict("order_closed", "order is no longer open")
	}
	if o.PlayerID == c.pl.ID {
		return nil, rule("own_order", "you cannot accept your own order")
	}
	if err := c.spend(ci, pack.Amounts{o.BuyRes: o.BuyAmount}, "market_buy", fmt.Sprintf("order:%d", o.ID)); err != nil {
		return nil, err
	}
	if _, err := c.grant(ci, pack.Amounts{o.SellRes: o.SellAmount}, "market_receive", fmt.Sprintf("order:%d", o.ID)); err != nil {
		return nil, err
	}
	fee := c.mods.Apply(float64(c.p.Params.MarketFeePct), "market_fee")
	proceeds := o.BuyAmount - int64(math.Ceil(float64(o.BuyAmount)*fee/100))
	if _, err := c.tx.Exec(c.ctx, `UPDATE market_orders SET status='filled', buyer_player_id=$2, buyer_city_id=$3, closed_at=$4 WHERE id=$1`, o.ID, c.pl.ID, ci.ID, c.now); err != nil {
		return nil, err
	}
	if err := c.deliverTo(o.PlayerID, o.CityID, pack.Amounts{o.BuyRes: proceeds}, "market_sale", fmt.Sprintf("order:%d", o.ID)); err != nil {
		return nil, err
	}
	c.stat("market_trade", 1)
	_ = notify(c.ctx, c.tx, o.PlayerID, "order_filled", map[string]any{"orderId": o.ID, "buyer": c.pl.Name, "proceeds": proceeds, "res": o.BuyRes})
	o.Status = "filled"
	return o, nil
}

// deliverTo schedules resources for another player's city (processed in their lock).
func (c *pctx) deliverTo(playerID, cityID int64, a pack.Amounts, reason, ref string) error {
	var id int64
	if err := c.tx.QueryRow(c.ctx, `INSERT INTO deliveries(world_id, player_id, city_id, amounts, reason, ref) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		c.pl.WorldID, playerID, cityID, a, reason, ref).Scan(&id); err != nil {
		return err
	}
	return schedule(c.ctx, c.tx, c.pl.WorldID, playerID, "delivery", id, c.now)
}

func (c *pctx) CancelOrder(id int64) (*Order, error) {
	o, err := c.lockOrder(id)
	if err != nil {
		return nil, err
	}
	if o.PlayerID != c.pl.ID {
		return nil, notFound("order")
	}
	if o.Status != "open" {
		return nil, conflict("order_closed", "order is no longer open")
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE market_orders SET status='cancelled', closed_at=$2 WHERE id=$1`, o.ID, c.now); err != nil {
		return nil, err
	}
	ci := c.city(o.CityID)
	if _, err := c.grant(ci, pack.Amounts{o.SellRes: o.SellAmount}, "market_cancel", fmt.Sprintf("order:%d", o.ID)); err != nil {
		return nil, err
	}
	o.Status = "cancelled"
	return o, nil
}

// Exchange trades with the NPC merchant at reference rates minus a fee (a resource sink).
//
//	received = floor(amount × rate_from / rate_to × (1 − fee/100))
func (c *pctx) Exchange(ci *City, from, to string, amount int64) (pack.Amounts, error) {
	rf, rt := c.p.Params.ExchangeRates[from], c.p.Params.ExchangeRates[to]
	if rf <= 0 || rt <= 0 || from == to || !c.tradable(from) || !c.tradable(to) {
		return nil, invalid("exchange between %q and %q is not offered", from, to)
	}
	if amount < 1 || amount > 1e9 {
		return nil, invalid("amount must be in [1, 1e9]")
	}
	got := int64(math.Floor(float64(amount) * rf / rt * (1 - float64(c.p.Params.ExchangeFeePct)/100)))
	if got < 1 {
		return nil, rule("amount_too_small", "amount too small to exchange")
	}
	if err := c.spend(ci, pack.Amounts{from: amount}, "exchange", to); err != nil {
		return nil, err
	}
	out := pack.Amounts{to: got}
	_, err := c.grant(ci, out, "exchange", from)
	return out, err
}

func (s *Service) Orders(ctx contextT, q querier, worldID int64, sell, buy string, mine int64) ([]Order, error) {
	sql := `SELECT o.id, o.player_id, p.name, o.city_id, o.sell_res, o.sell_amount, o.buy_res, o.buy_amount, o.status, o.created_at
		FROM market_orders o JOIN players p ON p.id=o.player_id WHERE o.world_id=$1 AND o.status='open'`
	args := []any{worldID}
	if sell != "" {
		args = append(args, sell)
		sql += fmt.Sprintf(` AND o.sell_res=$%d`, len(args))
	}
	if buy != "" {
		args = append(args, buy)
		sql += fmt.Sprintf(` AND o.buy_res=$%d`, len(args))
	}
	if mine != 0 {
		args = append(args, mine)
		sql += fmt.Sprintf(` AND o.player_id=$%d`, len(args))
	}
	sql += ` ORDER BY (o.buy_amount::float / o.sell_amount), o.id LIMIT 200`
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.PlayerID, &o.Player, &o.CityID, &o.SellRes, &o.SellAmount, &o.BuyRes, &o.BuyAmount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ---------- Territories (GAME_RULES R-TER-1..5) ----------

type Territory struct {
	X        int              `json:"x"`
	Y        int              `json:"y"`
	Terrain  string           `json:"terrain"`
	Resource string           `json:"resource"`
	Pending  pack.Amounts     `json:"pending"`
	Since    *time.Time       `json:"since"`
	Garrison map[string]int64 `json:"garrison"`
}

// territoryYield: yield = floor(min(elapsed, capHours) × yieldPerHour × worldSpeed) of the terrain resource.
func (c *pctx) territoryYield(terrain string, since *time.Time) pack.Amounts {
	t := c.p.Terrain(terrain)
	out := pack.Amounts{}
	if t == nil || t.Resource == "" || since == nil {
		return out
	}
	h := math.Min(c.now.Sub(*since).Hours()*c.worldSpeed(), float64(c.p.Params.TerritoryYieldCapHours))
	if n := int64(math.Floor(math.Max(0, h) * c.p.Params.TerritoryYieldPerHour)); n > 0 {
		out[t.Resource] = n
	}
	return out
}

func (c *pctx) Territories() ([]Territory, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT x, y, terrain, owner_since, yield_collected_at FROM tiles WHERE world_id=$1 AND owner_player_id=$2 ORDER BY x, y`, c.pl.WorldID, c.pl.ID)
	if err != nil {
		return nil, err
	}
	var out []Territory
	for rows.Next() {
		var t Territory
		var yat *time.Time
		if err := rows.Scan(&t.X, &t.Y, &t.Terrain, &t.Since, &yat); err != nil {
			rows.Close()
			return nil, err
		}
		t.Pending = c.territoryYield(t.Terrain, yat)
		if ter := c.p.Terrain(t.Terrain); ter != nil {
			t.Resource = ter.Resource
		}
		out = append(out, t)
	}
	rows.Close()
	for i := range out {
		st, err := c.stationedAtTile(out[i].X, out[i].Y, false)
		if err != nil {
			return nil, err
		}
		out[i].Garrison = map[string]int64{}
		for _, m := range st {
			for u, n := range m.Units {
				out[i].Garrison[u] += n
			}
		}
	}
	if out == nil {
		out = []Territory{}
	}
	return out, nil
}

// Collect transfers the accrued yield of an owned territory to a city.
func (c *pctx) Collect(ci *City, x, y int) (pack.Amounts, error) {
	t, err := c.tile(x, y, true)
	if err != nil {
		return nil, err
	}
	if t.Owner == nil || *t.Owner != c.pl.ID {
		return nil, rule("not_owner", "you do not own this territory")
	}
	y0 := c.territoryYield(t.Terrain, t.YieldAt)
	if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET yield_collected_at=$4 WHERE world_id=$1 AND x=$2 AND y=$3`, c.pl.WorldID, x, y, c.now); err != nil {
		return nil, err
	}
	_, err = c.grant(ci, y0, "territory_yield", fmt.Sprintf("%d:%d", x, y))
	return y0, err
}

// Abandon releases an owned territory; stationed own troops are sent home.
func (c *pctx) Abandon(x, y int) error {
	t, err := c.tile(x, y, true)
	if err != nil {
		return err
	}
	if t.Owner == nil || *t.Owner != c.pl.ID {
		return rule("not_owner", "you do not own this territory")
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET owner_player_id=NULL, owner_since=NULL, yield_collected_at=NULL WHERE world_id=$1 AND x=$2 AND y=$3`, c.pl.WorldID, x, y); err != nil {
		return err
	}
	st, err := c.stationedAtTile(x, y, true)
	if err != nil {
		return err
	}
	for _, m := range st {
		if m.PlayerID == c.pl.ID {
			if err := c.sendBack(m, m.ArriveAt.Sub(m.DepartAt)); err != nil {
				return err
			}
		}
	}
	return nil
}

// ---------- Map & rankings ----------

type MapTile struct {
	X            int     `json:"x"`
	Y            int     `json:"y"`
	Terrain      string  `json:"terrain"`
	Neutral      *string `json:"neutral,omitempty"`
	NeutralLevel *int    `json:"neutralLevel,omitempty"`
	Cleared      bool    `json:"cleared,omitempty"`
	OwnerID      *int64  `json:"ownerId,omitempty"`
	Owner        *string `json:"owner,omitempty"`
	Alliance     *string `json:"alliance,omitempty"`
	CityID       *int64  `json:"cityId,omitempty"`
	CityName     *string `json:"cityName,omitempty"`
	Race         *string `json:"race,omitempty"`
}

// MaxViewport bounds map queries.
const MaxViewport = 64

func (s *Service) Map(ctx contextT, q querier, worldID int64, x, y, w, h int, now time.Time) ([]MapTile, error) {
	if w < 1 || h < 1 || w > MaxViewport || h > MaxViewport {
		return nil, invalid("viewport must be between 1x1 and %dx%d", MaxViewport, MaxViewport)
	}
	rows, err := q.Query(ctx, `SELECT t.x, t.y, t.terrain, t.neutral_id, t.neutral_level, t.cleared_until,
			COALESCE(cp.id, op.id), COALESCE(cp.name, op.name), a.tag, t.city_id, c.name, COALESCE(cp.race_id, op.race_id)
		FROM tiles t
		LEFT JOIN cities c ON c.id = t.city_id
		LEFT JOIN players cp ON cp.id = c.player_id
		LEFT JOIN players op ON op.id = t.owner_player_id
		LEFT JOIN alliances a ON a.id = COALESCE(cp.alliance_id, op.alliance_id)
		WHERE t.world_id=$1 AND t.x >= $2 AND t.x < $2+$4 AND t.y >= $3 AND t.y < $3+$5
		ORDER BY t.y, t.x`, worldID, x, y, w, h)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MapTile{}
	for rows.Next() {
		var t MapTile
		var cleared *time.Time
		if err := rows.Scan(&t.X, &t.Y, &t.Terrain, &t.Neutral, &t.NeutralLevel, &cleared, &t.OwnerID, &t.Owner, &t.Alliance, &t.CityID, &t.CityName, &t.Race); err != nil {
			return nil, err
		}
		t.Cleared = cleared != nil && cleared.After(now)
		out = append(out, t)
	}
	return out, rows.Err()
}

type RankRow struct {
	Rank     int     `json:"rank"`
	PlayerID int64   `json:"playerId"`
	Name     string  `json:"name"`
	Race     string  `json:"race"`
	Score    int64   `json:"score"`
	Alliance *string `json:"alliance"`
	Cities   int     `json:"cities"`
}

func (s *Service) Rankings(ctx contextT, q querier, worldID int64, limit, offset int) ([]RankRow, error) {
	rows, err := q.Query(ctx, `SELECT p.id, p.name, p.race_id, p.score, a.tag, (SELECT count(*) FROM cities c WHERE c.player_id=p.id)
		FROM players p LEFT JOIN alliances a ON a.id=p.alliance_id WHERE p.world_id=$1 ORDER BY p.score DESC, p.id LIMIT $2 OFFSET $3`, worldID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RankRow{}
	i := offset
	for rows.Next() {
		i++
		r := RankRow{Rank: i}
		if err := rows.Scan(&r.PlayerID, &r.Name, &r.Race, &r.Score, &r.Alliance, &r.Cities); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// wholeMilli is a helper for tests and views.
func wholeMilli(m map[string]int64) pack.Amounts { return rules.Whole(m) }
