package game

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

var MissionKinds = map[string]bool{"attack": true, "scout": true, "transport": true, "reinforce": true, "settle": true, "occupy": true}

type MissionRequest struct {
	Kind       string           `json:"kind"`
	X          int              `json:"x"`
	Y          int              `json:"y"`
	Units      map[string]int64 `json:"units"`
	HeroID     *int64           `json:"heroId,omitempty"`
	Cargo      pack.Amounts     `json:"cargo,omitempty"`
	Capture    bool             `json:"capture,omitempty"`    // attack on territory: take it on victory
	RetreatPct int              `json:"retreatPct,omitempty"` // attacker retreat threshold (1..100)
}

type Mission struct {
	ID           int64            `json:"id"`
	PlayerID     int64            `json:"playerId"`
	FromCityID   int64            `json:"fromCityId"`
	Kind         string           `json:"kind"`
	TargetX      int              `json:"targetX"`
	TargetY      int              `json:"targetY"`
	TargetCityID *int64           `json:"targetCityId"`
	Units        map[string]int64 `json:"units"`
	UpkeepUnits  map[string]int64 `json:"-"`
	HeroID       *int64           `json:"heroId"`
	Cargo        pack.Amounts     `json:"cargo"`
	Options      map[string]any   `json:"options"`
	DepartAt     time.Time        `json:"departAt"`
	ArriveAt     time.Time        `json:"arriveAt"`
	ReturnAt     *time.Time       `json:"returnAt"`
	Status       string           `json:"status"`
	Result       map[string]any   `json:"result"`
}

const missionCols = `id, player_id, from_city_id, kind, target_x, target_y, target_city_id, units, upkeep_units, hero_id, cargo, options, depart_at, arrive_at, return_at, status, result`

func scanMission(row pgx.Row) (*Mission, error) {
	m := &Mission{}
	err := row.Scan(&m.ID, &m.PlayerID, &m.FromCityID, &m.Kind, &m.TargetX, &m.TargetY, &m.TargetCityID, &m.Units, &m.UpkeepUnits,
		&m.HeroID, &m.Cargo, &m.Options, &m.DepartAt, &m.ArriveAt, &m.ReturnAt, &m.Status, &m.Result)
	if m.Units == nil {
		m.Units = map[string]int64{}
	}
	if m.Cargo == nil {
		m.Cargo = pack.Amounts{}
	}
	if m.Result == nil {
		m.Result = map[string]any{}
	}
	if m.Options == nil {
		m.Options = map[string]any{}
	}
	return m, err
}

type tileInfo struct {
	X, Y         int
	Terrain      string
	Neutral      *string
	NeutralLevel *int
	ClearedUntil *time.Time
	Owner        *int64
	OwnerSince   *time.Time
	YieldAt      *time.Time
	CityID       *int64
}

// neutralActive: a neutral guards its tile unless cleared (respawn pending) or
// the tile is an owned territory (territories are pacified, R-TER-2).
func (t *tileInfo) neutralActive(now time.Time) bool {
	return t.Neutral != nil && t.Owner == nil && (t.ClearedUntil == nil || !t.ClearedUntil.After(now))
}

func (c *pctx) tile(x, y int, forUpdate bool) (*tileInfo, error) {
	q := `SELECT x, y, terrain, neutral_id, neutral_level, cleared_until, owner_player_id, owner_since, yield_collected_at, city_id FROM tiles WHERE world_id=$1 AND x=$2 AND y=$3`
	if forUpdate {
		q += ` FOR UPDATE`
	}
	t := &tileInfo{}
	err := c.tx.QueryRow(c.ctx, q, c.pl.WorldID, x, y).Scan(&t.X, &t.Y, &t.Terrain, &t.Neutral, &t.NeutralLevel, &t.ClearedUntil, &t.Owner, &t.OwnerSince, &t.YieldAt, &t.CityID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, invalid("tile (%d,%d) is outside the map", x, y)
	}
	return t, err
}

// otherPlayer is a read-only snapshot of another player used for combat modifiers.
type otherPlayer struct {
	ID         int64
	Name       string
	Race       string
	AllianceID *int64
	Score      int64
	Protected  *time.Time
	Mods       *rules.Mods
}

func (c *pctx) other(id int64) (*otherPlayer, error) {
	if id == c.pl.ID {
		return &otherPlayer{ID: id, Name: c.pl.Name, Race: c.pl.Race, AllianceID: c.pl.AllianceID, Score: c.pl.Score, Protected: c.pl.ProtectedUntil, Mods: c.mods}, nil
	}
	o := &otherPlayer{ID: id}
	var techs map[string]int
	err := c.tx.QueryRow(c.ctx, `SELECT name, race_id, alliance_id, score, protected_until, techs FROM players WHERE id=$1`, id).
		Scan(&o.Name, &o.Race, &o.AllianceID, &o.Score, &o.Protected, &techs)
	if err != nil {
		return nil, err
	}
	m := rules.NewMods()
	if r := c.p.Race(o.Race); r != nil {
		m.AddAll(r.Modifiers, 1)
	}
	for tid, lvl := range techs {
		if t := c.p.Tech(tid); t != nil {
			for _, e := range t.Effects {
				if e.Type == "modifier" && e.Modifier != nil {
					m.Add(*e.Modifier, float64(lvl))
				}
			}
		}
	}
	o.Mods = m
	return o, nil
}

// friendly reports whether two players must not fight: same player, same
// alliance, or alliances in an 'ally' or 'nap' relation (GAME_RULES R-SOC-3).
func (c *pctx) friendly(a, b *otherPlayer) (bool, error) {
	if a.ID == b.ID {
		return true, nil
	}
	if a.AllianceID == nil || b.AllianceID == nil {
		return false, nil
	}
	if *a.AllianceID == *b.AllianceID {
		return true, nil
	}
	var n int
	err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM alliance_relations WHERE alliance_id=$1 AND other_id=$2 AND status IN ('ally','nap')`,
		*a.AllianceID, *b.AllianceID).Scan(&n)
	return n > 0, err
}

func (c *pctx) allied(a, b *otherPlayer) (bool, error) {
	if a.ID == b.ID {
		return true, nil
	}
	if a.AllianceID == nil || b.AllianceID == nil {
		return false, nil
	}
	if *a.AllianceID == *b.AllianceID {
		return true, nil
	}
	var n int
	err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM alliance_relations WHERE alliance_id=$1 AND other_id=$2 AND status='ally'`,
		*a.AllianceID, *b.AllianceID).Scan(&n)
	return n > 0, err
}

func (c *pctx) cityCap() int {
	if l := c.p.Params.CityLayout; l != nil && c.p.Params.CityCapBuilding == "" {
		return rules.MaxCities(l) // GAME_RULES R-CITY-3
	}
	b := c.p.Params.CityCapBuilding
	if b == "" {
		return 1
	}
	lvl := 0
	for _, ci := range c.cities {
		lvl = max(lvl, ci.Buildings[b])
	}
	n := min(c.p.Params.MaxCities, 1+lvl/c.p.Params.CitiesPerCapLevels)
	if l := c.p.Params.CityLayout; l != nil {
		n = min(n, rules.MaxCities(l))
	}
	return n
}

func (c *pctx) territoryCap() int {
	b := c.p.Params.TerritoryCapBuilding
	if b == "" {
		return 0
	}
	lvl := 0
	for _, ci := range c.cities {
		lvl = max(lvl, ci.Buildings[b])
	}
	return lvl * c.p.Params.TerritoriesPerLevel
}

func (c *pctx) territoryCount() (int, error) {
	var n int
	err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM tiles WHERE world_id=$1 AND owner_player_id=$2`, c.pl.WorldID, c.pl.ID).Scan(&n)
	return n, err
}

func (c *pctx) speedOf(units map[string]int64, mods *rules.Mods) float64 {
	slow := math.Inf(1)
	for id, n := range units {
		if n > 0 {
			if u := c.p.Unit(id); u != nil && u.Speed < slow {
				slow = u.Speed
			}
		}
	}
	return slow
}

func (c *pctx) carryOf(units map[string]int64, mods *rules.Mods) int64 {
	var total float64
	for id, n := range units {
		if u := c.p.Unit(id); u != nil {
			total += float64(n * u.Carry)
		}
	}
	return int64(math.Floor(mods.Apply(total, "carry")))
}

func amountsTotal(a pack.Amounts) int64 {
	var t int64
	for _, v := range a {
		t += v
	}
	return t
}

// SendMission validates and dispatches a mission (GAME_RULES R-MSN-1..12).
func (c *pctx) SendMission(ci *City, req MissionRequest) (*Mission, error) {
	if !MissionKinds[req.Kind] {
		return nil, invalid("unknown mission kind %q", req.Kind)
	}
	if req.RetreatPct < 0 || req.RetreatPct > 100 {
		return nil, invalid("retreatPct must be in [0, 100]")
	}
	units := map[string]int64{}
	for id, n := range req.Units {
		if n == 0 {
			continue
		}
		if n < 0 || n > 1e9 {
			return nil, invalid("invalid unit count for %q", id)
		}
		if c.p.Unit(id) == nil {
			return nil, invalid("unknown unit %q", id)
		}
		if ci.Units[id] < n {
			return nil, rule("insufficient_units", "not enough %s in the city (have %d)", id, ci.Units[id])
		}
		units[id] = n
	}
	if len(units) == 0 {
		return nil, invalid("a mission needs at least one unit")
	}
	if req.X == ci.X && req.Y == ci.Y {
		return nil, invalid("target is the departure city")
	}
	t, err := c.tile(req.X, req.Y, false)
	if err != nil {
		return nil, err
	}
	if ter := c.p.Terrain(t.Terrain); ter == nil || !ter.Passable {
		return nil, rule("impassable", "target terrain is impassable")
	}
	mods := c.cityMods(ci)
	var hero *Hero
	if req.HeroID != nil {
		h, hm, err := c.heroByID(*req.HeroID, true)
		if err != nil || h.PlayerID != c.pl.ID {
			return nil, notFound("hero")
		}
		if h.Status != "idle" || (h.WoundedUntil != nil && h.WoundedUntil.After(c.now)) {
			return nil, rule("hero_busy", "hero is busy or wounded")
		}
		if h.HomeCityID != ci.ID {
			return nil, rule("hero_elsewhere", "hero is based in another city")
		}
		hero = h
		mods.Merge(hm)
	}
	var targetCity *int64
	var targetOwner *otherPlayer
	if t.CityID != nil {
		targetCity = t.CityID
		var ownerID int64
		if err := c.tx.QueryRow(c.ctx, `SELECT player_id FROM cities WHERE id=$1`, *t.CityID).Scan(&ownerID); err != nil {
			return nil, err
		}
		if targetOwner, err = c.other(ownerID); err != nil {
			return nil, err
		}
	} else if t.Owner != nil {
		if targetOwner, err = c.other(*t.Owner); err != nil {
			return nil, err
		}
	}
	me, _ := c.other(c.pl.ID)
	cargo := pack.Amounts{}
	switch req.Kind {
	case "scout":
		for id := range units {
			if c.p.Unit(id).Class != "scout" {
				return nil, invalid("only scout units can be sent on a scouting mission")
			}
		}
		if targetOwner != nil && targetOwner.ID != c.pl.ID && targetOwner.Protected != nil && targetOwner.Protected.After(c.now) {
			return nil, rule("target_protected", "target is under newbie protection")
		}
		c.stat("scouts_sent", 1)
	case "attack":
		switch {
		case targetCity != nil:
			if err := c.checkPvP(me, targetOwner, req.X, req.Y); err != nil {
				return nil, err
			}
		case t.neutralActive(c.now):
			// PvE
		case targetOwner != nil:
			if err := c.checkPvP(me, targetOwner, req.X, req.Y); err != nil {
				return nil, err
			}
		default:
			return nil, rule("no_target", "nothing to attack at this location")
		}
	case "transport":
		if targetCity == nil {
			return nil, rule("no_target", "transport target must be a city")
		}
		// Do not allow transfers to unrelated accounts (anti-farming).
		// Cross-account transfers require a current alliance relationship.
		ok, err := c.allied(me, targetOwner)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, rule("not_allied", "resources can only be sent to your own or allied cities")
		}
		cap := c.carryOf(units, mods)
		for r, n := range req.Cargo {
			res := c.p.Resource(r)
			if res == nil || n < 0 {
				return nil, invalid("invalid cargo %q", r)
			}
			if res.Kind == "premium" {
				return nil, rule("not_transportable", "%s cannot be transported", r)
			}
			if n > 0 {
				cargo[r] = n
			}
		}
		if amountsTotal(cargo) == 0 {
			return nil, invalid("transport needs cargo")
		}
		if amountsTotal(cargo) > cap {
			return nil, rule("over_capacity", "cargo %d exceeds carrying capacity %d", amountsTotal(cargo), cap)
		}
	case "reinforce":
		if targetCity == nil {
			return nil, rule("no_target", "reinforcements must target a city")
		}
		ok, err := c.allied(me, targetOwner)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, rule("not_allied", "you can only reinforce your own or allied cities")
		}
	case "settle":
		su := c.p.Params.SettlerUnit
		if su == "" || units[su] < 1 {
			return nil, rule("settler_required", "a settler unit is required")
		}
		if targetCity != nil || t.neutralActive(c.now) || (t.Owner != nil && *t.Owner != c.pl.ID) {
			return nil, rule("tile_occupied", "location is not free")
		}
		if len(c.cities) >= c.cityCap() {
			return nil, rule("city_limit", "city limit reached (%d)", c.cityCap())
		}
		for r, n := range req.Cargo {
			if res := c.p.Resource(r); res == nil || res.Kind == "premium" || n < 0 {
				return nil, invalid("invalid cargo %q", r)
			}
			if n > 0 {
				cargo[r] = n
			}
		}
		if amountsTotal(cargo) > c.carryOf(units, mods) {
			return nil, rule("over_capacity", "cargo exceeds carrying capacity")
		}
	case "occupy":
		if targetCity != nil || t.neutralActive(c.now) {
			return nil, rule("tile_occupied", "location is not free")
		}
		if t.Neutral != nil && !c.p.Neutral(*t.Neutral).Capturable {
			return nil, rule("not_capturable", "this site cannot be held as a territory")
		}
		if targetOwner != nil && targetOwner.ID != c.pl.ID {
			return nil, rule("tile_owned", "location is owned by another player; attack it with capture enabled")
		}
		if targetOwner == nil {
			n, err := c.territoryCount()
			if err != nil {
				return nil, err
			}
			if n >= c.territoryCap() {
				return nil, rule("territory_limit", "territory limit reached (%d)", c.territoryCap())
			}
		}
	}
	if len(cargo) > 0 {
		if err := c.spend(ci, cargo, "mission_cargo", req.Kind); err != nil {
			return nil, err
		}
	}
	dist := rules.Distance(ci.X, ci.Y, req.X, req.Y)
	secs := rules.TravelSeconds(dist, c.speedOf(units, mods), mods.Pct("speed"), c.worldSpeed(), c.p.Params.MinTravelSeconds)
	m := &Mission{PlayerID: c.pl.ID, FromCityID: ci.ID, Kind: req.Kind, TargetX: req.X, TargetY: req.Y, TargetCityID: targetCity,
		Units: units, UpkeepUnits: units, HeroID: req.HeroID, Cargo: cargo, Status: "outbound",
		Options: map[string]any{"capture": req.Capture, "retreatPct": req.RetreatPct}, DepartAt: c.now,
		ArriveAt: c.now.Add(time.Duration(secs) * time.Second), Result: map[string]any{}}
	for id, n := range units {
		ci.Units[id] -= n
	}
	err = c.tx.QueryRow(c.ctx, `INSERT INTO missions(world_id, player_id, from_city_id, kind, target_x, target_y, target_city_id, units, upkeep_units, hero_id, cargo, options, depart_at, arrive_at, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8,$9,$10,$11,$12,$13,'outbound') RETURNING id`,
		c.pl.WorldID, c.pl.ID, ci.ID, m.Kind, m.TargetX, m.TargetY, m.TargetCityID, m.Units, m.HeroID, m.Cargo, m.Options, m.DepartAt, m.ArriveAt).Scan(&m.ID)
	if err != nil {
		return nil, err
	}
	if hero != nil {
		if _, err := c.tx.Exec(c.ctx, `UPDATE heroes SET status='mission' WHERE id=$1`, hero.ID); err != nil {
			return nil, err
		}
	}
	// The arrival event belongs to the player whose state it changes: the
	// owner of the target city, otherwise the mission owner.
	evPlayer := c.pl.ID
	if targetCity != nil {
		evPlayer = targetOwner.ID
	}
	if err := schedule(c.ctx, c.tx, c.pl.WorldID, evPlayer, "mission_arrive", m.ID, m.ArriveAt); err != nil {
		return nil, err
	}
	if targetOwner != nil && targetOwner.ID != c.pl.ID && (req.Kind == "attack") {
		_ = notify(c.ctx, c.tx, targetOwner.ID, "incoming_attack", map[string]any{"x": req.X, "y": req.Y, "arriveAt": m.ArriveAt, "from": c.pl.Name})
	}
	return m, c.refreshAway(ci)
}

// checkPvP enforces PvP restrictions at dispatch.
func (c *pctx) checkPvP(me, target *otherPlayer, x, y int) error {
	if target.ID == me.ID {
		return rule("own_target", "you cannot attack yourself")
	}
	fr, err := c.friendly(me, target)
	if err != nil {
		return err
	}
	if fr {
		return rule("friendly_target", "you cannot attack allied or non-aggression partners")
	}
	if target.Protected != nil && target.Protected.After(c.now) {
		return rule("target_protected", "target is under newbie protection")
	}
	if f := c.p.Params.AttackScoreFloorPct; f > 0 && target.Score*100 < me.Score*int64(f) {
		return rule("target_too_weak", "target is too weak compared to you")
	}
	if lim := c.p.Params.MaxAttacksPerTargetPerDay; lim > 0 {
		var n int
		if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM missions WHERE player_id=$1 AND kind='attack' AND target_x=$2 AND target_y=$3 AND world_id=$4 AND depart_at > $5`,
			me.ID, x, y, c.pl.WorldID, c.now.Add(-24*time.Hour)).Scan(&n); err != nil {
			return err
		}
		if n >= lim {
			return rule("attack_limit", "at most %d attacks per target per 24 hours", lim)
		}
	}
	// Attacking another player ends your own newbie protection (GAME_RULES R-PVP-2).
	c.pl.ProtectedUntil = nil
	return nil
}

func (c *pctx) lockMission(id int64) (*Mission, error) {
	m, err := scanMission(c.tx.QueryRow(c.ctx, `SELECT `+missionCols+` FROM missions WHERE id=$1 FOR UPDATE`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("mission")
	}
	return m, err
}

func (c *pctx) saveMission(m *Mission) error {
	_, err := c.tx.Exec(c.ctx, `UPDATE missions SET units=$2, cargo=$3, status=$4, return_at=$5, result=$6 WHERE id=$1`,
		m.ID, m.Units, m.Cargo, m.Status, m.ReturnAt, m.Result)
	return err
}

// sendBack turns a mission home; the return event belongs to its owner.
func (c *pctx) sendBack(m *Mission, travel time.Duration) error {
	if travel < 0 {
		travel = 0
	}
	if amountsAlive(m.Units) == 0 {
		travel = 0
	}
	rt := c.now.Add(travel)
	m.Status = "returning"
	m.ReturnAt = &rt
	if err := c.saveMission(m); err != nil {
		return err
	}
	return schedule(c.ctx, c.tx, c.pl.WorldID, m.PlayerID, "mission_return", m.ID, rt)
}

func amountsAlive(u map[string]int64) int64 {
	var n int64
	for _, v := range u {
		n += v
	}
	return n
}

func (c *pctx) report(kind string, data map[string]any, recipients map[int64]string) (int64, error) {
	var id int64
	if err := c.tx.QueryRow(c.ctx, `INSERT INTO reports(world_id, kind, occurred_at, data) VALUES ($1,$2,$3,$4) RETURNING id`, c.pl.WorldID, kind, c.now, data).Scan(&id); err != nil {
		return 0, err
	}
	ids := make([]int64, 0, len(recipients))
	for pid := range recipients {
		ids = append(ids, pid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, pid := range ids {
		if _, err := c.tx.Exec(c.ctx, `INSERT INTO report_recipients(report_id, player_id, side) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, id, pid, recipients[pid]); err != nil {
			return 0, err
		}
		if err := notify(c.ctx, c.tx, pid, "report", map[string]any{"reportId": id, "kind": kind, "side": recipients[pid]}); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func addResult(m *Mission, key string, v any) { m.Result[key] = v }

func resultStats(m *Mission, key string, n int64) {
	st, _ := m.Result["stats"].(map[string]any)
	if st == nil {
		st = map[string]any{}
	}
	prev, _ := st[key].(float64)
	st[key] = prev + float64(n)
	m.Result["stats"] = st
}

// missionArrive resolves a mission at its target (event handler).
func (c *pctx) missionArrive(id int64) error {
	m, err := c.lockMission(id)
	if err != nil {
		return err
	}
	if m.Status != "outbound" {
		return nil // recalled or already resolved
	}
	outbound := m.ArriveAt.Sub(m.DepartAt)
	owner, err := c.other(m.PlayerID)
	if err != nil {
		return err
	}
	attMods := owner.Mods.Clone()
	// Building modifiers of the departure city (e.g. carry, loot) apply to the
	// army just as they did when it was sent (read-only, no lock needed).
	var homeB map[string]int
	if err := c.tx.QueryRow(c.ctx, `SELECT buildings FROM cities WHERE id=$1`, m.FromCityID).Scan(&homeB); err != nil {
		return err
	}
	attMods.Merge(c.buildingMods(homeB))
	var hero *Hero
	if m.HeroID != nil {
		h, hm, err := c.heroByID(*m.HeroID, true)
		if err != nil {
			return err
		}
		hero = h
		attMods.Merge(hm)
	}
	if m.TargetCityID != nil {
		ci := c.city(*m.TargetCityID)
		if ci == nil {
			return fmt.Errorf("mission %d arrival processed by non-owner %d", m.ID, c.pl.ID)
		}
		switch m.Kind {
		case "transport":
			lost, err := c.grant(ci, m.Cargo, "transport_in", fmt.Sprintf("mission:%d", m.ID))
			if err != nil {
				return err
			}
			addResult(m, "delivered", m.Cargo)
			addResult(m, "lostToStorage", lost)
			_ = notify(c.ctx, c.tx, ci.PlayerID, "transport_arrived", map[string]any{"from": owner.Name, "cargo": m.Cargo, "cityId": ci.ID})
			if m.PlayerID != c.pl.ID {
				_ = notify(c.ctx, c.tx, m.PlayerID, "transport_delivered", map[string]any{"to": ci.Name, "cargo": m.Cargo})
			}
			m.Cargo = pack.Amounts{}
			return c.sendBack(m, outbound)
		case "reinforce":
			if m.PlayerID == c.pl.ID { // relocation between own cities
				for u, n := range m.Units {
					ci.Units[u] += n
				}
				m.Units = map[string]int64{}
				m.Status = "done"
				if err := c.saveMission(m); err != nil {
					return err
				}
				if hero != nil {
					hero.Status = "idle"
					if _, err := c.tx.Exec(c.ctx, `UPDATE heroes SET status='idle', home_city_id=$2 WHERE id=$1`, hero.ID, ci.ID); err != nil {
						return err
					}
				}
				from := c.city(m.FromCityID)
				if from != nil {
					return c.refreshAway(from)
				}
				return nil
			}
			var players int
			if err := c.tx.QueryRow(c.ctx, `SELECT count(DISTINCT player_id) FROM missions WHERE target_city_id=$1 AND status='stationed' AND player_id <> $2`, ci.ID, m.PlayerID).Scan(&players); err != nil {
				return err
			}
			if players >= c.p.Params.MaxReinforcingPlayers {
				addResult(m, "rejected", "too_many_reinforcing_players")
				return c.sendBack(m, outbound)
			}
			m.Status = "stationed"
			_ = notify(c.ctx, c.tx, ci.PlayerID, "reinforcement_arrived", map[string]any{"from": owner.Name, "units": m.Units, "cityId": ci.ID})
			return c.saveMission(m)
		case "scout":
			return c.scoutCity(m, ci, owner, attMods, outbound)
		case "attack":
			return c.battleCity(m, ci, owner, attMods, hero, outbound)
		}
		return c.sendBack(m, outbound)
	}
	t, err := c.tile(m.TargetX, m.TargetY, true)
	if err != nil {
		return err
	}
	switch {
	case t.CityID != nil:
		// A city was founded here after departure: missions to tiles cannot hit it.
		addResult(m, "outcome", "target_changed")
		return c.sendBack(m, outbound)
	case m.Kind == "scout":
		return c.scoutTile(m, t, owner, outbound)
	case m.Kind == "attack" && t.neutralActive(c.now):
		return c.battleNeutral(m, t, attMods, hero, outbound)
	case m.Kind == "attack" && t.Owner != nil && *t.Owner != m.PlayerID:
		return c.battleTerritory(m, t, owner, attMods, hero, outbound)
	case m.Kind == "occupy" || (m.Kind == "reinforce" && t.Owner != nil && *t.Owner == m.PlayerID):
		return c.occupy(m, t, outbound)
	case m.Kind == "settle":
		return c.settle(m, t, outbound)
	}
	addResult(m, "outcome", "nothing_found")
	return c.sendBack(m, outbound)
}

func (c *pctx) scoutPower(units map[string]int64, mods *rules.Mods) float64 {
	p := 0.0
	for id, n := range units {
		if u := c.p.Unit(id); u != nil {
			p += float64(n) * u.Scout
		}
	}
	return mods.Apply(p, "scout")
}

func (c *pctx) scoutCity(m *Mission, ci *City, owner *otherPlayer, attMods *rules.Mods, outbound time.Duration) error {
	att := c.scoutPower(m.Units, attMods)
	def := c.scoutPower(ci.Units, c.cityMods(ci))
	roll := rules.NewRNG(rules.SeedFrom(m.ID, 7)).Float()
	r := rules.Scout(att, def, c.p.Params.ScoutDetectBasePct, roll)
	data := map[string]any{"mission": m.ID, "attacker": owner.Name, "defender": c.pl.Name, "target": map[string]any{"x": ci.X, "y": ci.Y, "city": ci.Name},
		"success": r.Success, "detected": r.Detected, "intel": r.Intel}
	if r.Success {
		data["resources"] = rules.Whole(ci.Res)
		foreign, err := stationedAtCity(c, ci.ID)
		if err != nil {
			return err
		}
		data["units"] = ci.Units
		data["foreignTroops"] = foreign
		if r.Intel >= 2 {
			data["buildings"] = ci.Buildings
			data["population"] = ci.Pop
		}
	} else {
		data["scoutsLost"] = m.Units
		m.Units = map[string]int64{}
	}
	rec := map[int64]string{m.PlayerID: "attacker"}
	if r.Detected {
		rec[c.pl.ID] = "defender"
	}
	rid, err := c.report("scout", data, rec)
	if err != nil {
		return err
	}
	addResult(m, "reportId", rid)
	addResult(m, "success", r.Success)
	return c.sendBack(m, outbound)
}

func (c *pctx) scoutTile(m *Mission, t *tileInfo, owner *otherPlayer, outbound time.Duration) error {
	data := map[string]any{"mission": m.ID, "attacker": owner.Name, "target": map[string]any{"x": t.X, "y": t.Y, "terrain": t.Terrain}, "success": true, "intel": 2}
	if t.neutralActive(c.now) {
		n := c.p.Neutral(*t.Neutral)
		lvl := n.Levels[min(*t.NeutralLevel, len(n.Levels))-1]
		data["neutral"] = map[string]any{"id": n.ID, "level": lvl.Level, "guards": lvl.Guards, "rewards": lvl.Rewards}
	}
	if t.Owner != nil {
		groups, err := c.stationedAtTile(t.X, t.Y, false)
		if err != nil {
			return err
		}
		g := map[string]int64{}
		for _, s := range groups {
			for u, n := range s.Units {
				g[u] += n
			}
		}
		data["garrison"] = g
		data["owner"] = *t.Owner
	}
	rid, err := c.report("scout", data, map[int64]string{m.PlayerID: "attacker"})
	if err != nil {
		return err
	}
	addResult(m, "reportId", rid)
	return c.sendBack(m, outbound)
}

func (c *pctx) stationedAtTile(x, y int, lock bool) ([]*Mission, error) {
	q := `SELECT ` + missionCols + ` FROM missions WHERE world_id=$1 AND target_x=$2 AND target_y=$3 AND target_city_id IS NULL AND status='stationed' ORDER BY id`
	if lock {
		q += ` FOR UPDATE`
	}
	rows, err := c.tx.Query(c.ctx, q, c.pl.WorldID, x, y)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (c *pctx) stationedAtCityLocked(cityID int64) ([]*Mission, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT `+missionCols+` FROM missions WHERE target_city_id=$1 AND status='stationed' ORDER BY id FOR UPDATE`, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (c *pctx) stacks(units map[string]int64, mods *rules.Mods, wall float64, group int) []rules.Stack {
	var out []rules.Stack
	for _, id := range sortedIDs(units) {
		if n := units[id]; n > 0 {
			if u := c.p.Unit(id); u != nil {
				s := rules.BuildStack(u, n, mods, wall)
				s.Group = group
				out = append(out, s)
			}
		}
	}
	return out
}

func sortedIDs[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// factionMods applies same-race / same-faction attack penalties (VERIFIED in
// research FAC-002; values are pack parameters).
func (c *pctx) factionMods(att *rules.Mods, attRace, defRace string) {
	if attRace == defRace {
		if v := c.p.Params.SameRaceAttackPct; v != 0 {
			att.Add(pack.Modifier{Target: "attack.*", Op: "pct", Value: v}, 1)
		}
		return
	}
	ra, rd := c.p.Race(attRace), c.p.Race(defRace)
	if ra != nil && rd != nil && ra.Faction != "" && ra.Faction == rd.Faction {
		if v := c.p.Params.SameFactionAttackPct; v != 0 {
			att.Add(pack.Modifier{Target: "attack.*", Op: "pct", Value: v}, 1)
		}
	}
}

func (c *pctx) battleOpts(m *Mission) rules.BattleOptions {
	rp := 100
	if v, ok := m.Options["retreatPct"].(float64); ok && v >= 1 && v <= 100 {
		rp = int(v)
	}
	return rules.BattleOptions{Rounds: c.p.Params.CombatRounds, DefenseCap: c.p.Params.DefenseCapPct, AttRetreatPct: rp, DefRetreatPct: c.p.Params.DefenderRetreatPct}
}

func (c *pctx) xp(score int64, mods *rules.Mods) int64 {
	return int64(math.Floor(mods.Apply(float64(score)*c.p.Heroes.XPPerScore, "hero_xp")))
}

func (c *pctx) battleCity(m *Mission, ci *City, owner *otherPlayer, attMods *rules.Mods, hero *Hero, outbound time.Duration) error {
	me, err := c.other(c.pl.ID)
	if err != nil {
		return err
	}
	if fr, err := c.friendly(owner, me); err != nil {
		return err
	} else if fr {
		addResult(m, "outcome", "target_became_friendly")
		return c.sendBack(m, outbound)
	}
	c.factionMods(attMods, owner.Race, c.pl.Race)
	defMods := c.cityMods(ci)
	dHero, dHeroMods, err := c.homeHero(ci.ID)
	if err != nil {
		return err
	}
	if dHero != nil {
		defMods.Merge(dHeroMods)
	}
	wall := defMods.Apply(0, "wall") // defense percentage points from fortifications (flat), scaled by wall pct
	def := c.stacks(ci.Units, defMods, wall, 0)
	support, err := c.stationedAtCityLocked(ci.ID)
	if err != nil {
		return err
	}
	for i, s := range support {
		o, err := c.other(s.PlayerID)
		if err != nil {
			return err
		}
		def = append(def, c.stacks(s.Units, o.Mods, wall, i+1)...)
	}
	att := c.stacks(m.Units, attMods, 0, 0)
	res := rules.Resolve(att, def, c.battleOpts(m))
	// apply losses
	ci.Units = rules.GroupCounts(res.DefenderFinal, 0)
	recipients := map[int64]string{m.PlayerID: "attacker", c.pl.ID: "defender"}
	for i, s := range support {
		s.Units = rules.GroupCounts(res.DefenderFinal, i+1)
		recipients[s.PlayerID] = "support"
		if amountsAlive(s.Units) == 0 {
			if err := c.sendBack(s, 0); err != nil {
				return err
			}
		} else if err := c.saveMission(s); err != nil {
			return err
		}
	}
	m.Units = rules.GroupCounts(res.AttackerFinal, 0)
	loot := pack.Amounts{}
	if res.AttackerWon {
		// Expose escrowed resources to the same raid as city stock.
		if err := c.releaseRaidEscrow(ci); err != nil {
			return err
		}
		capacity := c.carryOf(m.Units, attMods)
		lootPct := attMods.Apply(float64(c.p.Params.LootPct), "loot")
		loot = rules.Loot(rules.Whole(ci.Res), c.rates(ci).Protected, func(r string) bool {
			x := c.p.Resource(r)
			return x != nil && x.Lootable
		}, math.Min(100, lootPct), capacity)
		if len(loot) > 0 {
			if err := c.spend(ci, loot, "plundered", fmt.Sprintf("mission:%d", m.ID)); err != nil {
				return err
			}
		}
		for r, n := range loot {
			m.Cargo[r] += n
		}
		resultStats(m, "missions_won", 1)
	} else {
		c.stat("defenses_won", 1)
	}
	attXP := c.xp(res.DefenderScoreLost, attMods)
	data := map[string]any{"mission": m.ID, "attacker": owner.Name, "defender": c.pl.Name,
		"target": map[string]any{"x": ci.X, "y": ci.Y, "city": ci.Name}, "battle": res, "loot": loot,
		"attackerHero": heroBrief(hero), "defenderHero": heroBrief(dHero), "wall": wall}
	if hero != nil {
		addResult(m, "heroXp", attXP)
		addResult(m, "heroWounded", !res.AttackerWon)
	}
	if dHero != nil {
		c.gainXP(dHero, c.xp(res.AttackerScoreLost, dHeroMods))
		if res.AttackerWon {
			wu := c.now.Add(time.Duration(c.p.Params.HeroWoundHours) * time.Hour)
			dHero.WoundedUntil = &wu
		}
		if err := c.saveHero(dHero); err != nil {
			return err
		}
	}
	rid, err := c.report("battle", data, recipients)
	if err != nil {
		return err
	}
	addResult(m, "reportId", rid)
	addResult(m, "won", res.AttackerWon)
	return c.sendBack(m, outbound)
}

func heroBrief(h *Hero) any {
	if h == nil {
		return nil
	}
	return map[string]any{"name": h.Name, "class": h.Class, "level": h.Level}
}

func (c *pctx) battleNeutral(m *Mission, t *tileInfo, attMods *rules.Mods, hero *Hero, outbound time.Duration) error {
	n := c.p.Neutral(*t.Neutral)
	lvl := n.Levels[min(*t.NeutralLevel, len(n.Levels))-1]
	def := c.stacks(lvl.Guards, rules.NewMods(), 0, 0)
	att := c.stacks(m.Units, attMods, 0, 0)
	res := rules.Resolve(att, def, c.battleOpts(m))
	m.Units = rules.GroupCounts(res.AttackerFinal, 0)
	data := map[string]any{"mission": m.ID, "attacker": c.pl.Name, "neutral": map[string]any{"id": n.ID, "level": lvl.Level},
		"target": map[string]any{"x": t.X, "y": t.Y}, "battle": res, "attackerHero": heroBrief(hero)}
	if res.AttackerWon {
		until := c.now.Add(time.Duration(float64(n.RespawnSeconds)/c.worldSpeed()) * time.Second)
		if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET cleared_until=$4 WHERE world_id=$1 AND x=$2 AND y=$3`, c.pl.WorldID, t.X, t.Y, until); err != nil {
			return err
		}
		for r, a := range lvl.Rewards {
			m.Cargo[r] += a
		}
		rng := rules.NewRNG(rules.SeedFrom(m.ID, 11))
		var drops []string
		for _, d := range lvl.Drops {
			if rng.Intn(100) < d.Chance {
				drops = append(drops, d.Item)
			}
		}
		addResult(m, "items", drops)
		resultStats(m, "missions_won", 1)
		resultStats(m, "neutrals_cleared", 1)
		data["rewards"] = lvl.Rewards
		data["items"] = drops
		data["respawnAt"] = until
	}
	if hero != nil {
		xp := c.xp(res.DefenderScoreLost, attMods)
		if res.AttackerWon {
			xp += lvl.XP
		}
		addResult(m, "heroXp", xp)
		addResult(m, "heroWounded", !res.AttackerWon)
	}
	rid, err := c.report("battle", data, map[int64]string{m.PlayerID: "attacker"})
	if err != nil {
		return err
	}
	addResult(m, "reportId", rid)
	addResult(m, "won", res.AttackerWon)
	return c.sendBack(m, outbound)
}

func (c *pctx) battleTerritory(m *Mission, t *tileInfo, owner *otherPlayer, attMods *rules.Mods, hero *Hero, outbound time.Duration) error {
	defOwner, err := c.other(*t.Owner)
	if err != nil {
		return err
	}
	fr, err := c.friendly(owner, defOwner)
	if err != nil {
		return err
	}
	if fr {
		addResult(m, "outcome", "target_became_friendly")
		return c.sendBack(m, outbound)
	}
	c.factionMods(attMods, owner.Race, defOwner.Race)
	support, err := c.stationedAtTile(t.X, t.Y, true)
	if err != nil {
		return err
	}
	var def []rules.Stack
	recipients := map[int64]string{m.PlayerID: "attacker", defOwner.ID: "defender"}
	for i, s := range support {
		o, err := c.other(s.PlayerID)
		if err != nil {
			return err
		}
		def = append(def, c.stacks(s.Units, o.Mods, 0, i)...)
		if s.PlayerID != defOwner.ID {
			recipients[s.PlayerID] = "support"
		}
	}
	res := rules.Resolve(c.stacks(m.Units, attMods, 0, 0), def, c.battleOpts(m))
	m.Units = rules.GroupCounts(res.AttackerFinal, 0)
	for i, s := range support {
		s.Units = rules.GroupCounts(res.DefenderFinal, i)
		if amountsAlive(s.Units) == 0 || res.AttackerWon {
			// defeated garrisons that survive retreat home
			if err := c.sendBack(s, s.ArriveAt.Sub(s.DepartAt)); err != nil {
				return err
			}
		} else if err := c.saveMission(s); err != nil {
			return err
		}
	}
	captured := false
	if res.AttackerWon {
		resultStats(m, "missions_won", 1)
		if capture, _ := m.Options["capture"].(bool); capture {
			n, err := c.territoryCount()
			if err != nil {
				return err
			}
			if n < c.territoryCap() {
				if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET owner_player_id=$4, owner_since=$5 WHERE world_id=$1 AND x=$2 AND y=$3`,
					c.pl.WorldID, t.X, t.Y, m.PlayerID, c.now); err != nil {
					return err
				}
				captured = true
			} else {
				addResult(m, "captureFailed", "territory_limit")
			}
		}
	}
	if hero != nil {
		addResult(m, "heroXp", c.xp(res.DefenderScoreLost, attMods))
		addResult(m, "heroWounded", !res.AttackerWon)
	}
	data := map[string]any{"mission": m.ID, "attacker": owner.Name, "defender": defOwner.Name, "target": map[string]any{"x": t.X, "y": t.Y, "territory": true},
		"battle": res, "captured": captured, "attackerHero": heroBrief(hero)}
	rid, err := c.report("battle", data, recipients)
	if err != nil {
		return err
	}
	addResult(m, "reportId", rid)
	addResult(m, "won", res.AttackerWon)
	addResult(m, "captured", captured)
	if captured && amountsAlive(m.Units) > 0 {
		m.Status = "stationed"
		return c.saveMission(m)
	}
	return c.sendBack(m, outbound)
}

func (c *pctx) occupy(m *Mission, t *tileInfo, outbound time.Duration) error {
	if t.neutralActive(c.now) || (t.Owner != nil && *t.Owner != m.PlayerID) || (t.Neutral != nil && !c.p.Neutral(*t.Neutral).Capturable) {
		addResult(m, "outcome", "tile_unavailable")
		return c.sendBack(m, outbound)
	}
	if t.Owner == nil {
		n, err := c.territoryCount()
		if err != nil {
			return err
		}
		if n >= c.territoryCap() {
			addResult(m, "outcome", "territory_limit")
			return c.sendBack(m, outbound)
		}
		if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET owner_player_id=$4, owner_since=$5, yield_collected_at=$5 WHERE world_id=$1 AND x=$2 AND y=$3`,
			c.pl.WorldID, t.X, t.Y, m.PlayerID, c.now); err != nil {
			return err
		}
		c.stat("territories_taken", 1)
		_ = notify(c.ctx, c.tx, m.PlayerID, "territory_taken", map[string]any{"x": t.X, "y": t.Y})
	}
	m.Status = "stationed"
	addResult(m, "outcome", "occupied")
	return c.saveMission(m)
}

func (c *pctx) settle(m *Mission, t *tileInfo, outbound time.Duration) error {
	su := c.p.Params.SettlerUnit
	if t.neutralActive(c.now) || (t.Owner != nil && *t.Owner != m.PlayerID) || len(c.cities) >= c.cityCap() || m.Units[su] < 1 {
		addResult(m, "outcome", "settle_failed")
		return c.sendBack(m, outbound)
	}
	m.Units[su]--
	res := map[string]int64{}
	for r, n := range m.Cargo {
		res[r] = n * rules.Milli
	}
	buildings := map[string]int{}
	for b, l := range c.p.Start.Buildings {
		buildings[b] = 1
		_ = l
	}
	name := fmt.Sprintf("%s %d", c.pl.Name, len(c.cities)+1)
	if len(name) > 24 {
		name = name[:24]
	}
	ci := &City{PlayerID: c.pl.ID, Name: name, X: t.X, Y: t.Y, Res: res, Pop: c.p.Start.Population / 2, UpdatedAt: c.now,
		Buildings: buildings, Units: map[string]int64{}, Carry: map[string]float64{}}
	for u, n := range m.Units {
		ci.Units[u] = n
	}
	err := c.tx.QueryRow(c.ctx, `INSERT INTO cities(world_id, player_id, name, x, y, resources, population, econ_updated_at, buildings, units)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at`, c.pl.WorldID, c.pl.ID, ci.Name, ci.X, ci.Y, ci.Res, ci.Pop, c.now, ci.Buildings, ci.Units).
		Scan(&ci.ID, &ci.CreatedAt)
	if err != nil {
		if isUnique(err) {
			addResult(m, "outcome", "settle_failed")
			return c.sendBack(m, outbound)
		}
		return err
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE tiles SET city_id=$4, owner_player_id=NULL WHERE world_id=$1 AND x=$2 AND y=$3`, c.pl.WorldID, t.X, t.Y, ci.ID); err != nil {
		return err
	}
	c.cities = append(c.cities, ci)
	m.Units = map[string]int64{}
	m.Cargo = pack.Amounts{}
	m.Status = "done"
	addResult(m, "outcome", "city_founded")
	addResult(m, "cityId", ci.ID)
	if err := c.saveMission(m); err != nil {
		return err
	}
	if m.HeroID != nil {
		if _, err := c.tx.Exec(c.ctx, `UPDATE heroes SET status='idle', home_city_id=$2 WHERE id=$1`, *m.HeroID, ci.ID); err != nil {
			return err
		}
	}
	if from := c.city(m.FromCityID); from != nil {
		if err := c.refreshAway(from); err != nil {
			return err
		}
	}
	return notify(c.ctx, c.tx, c.pl.ID, "city_founded", map[string]any{"cityId": ci.ID, "x": t.X, "y": t.Y})
}

// missionReturn brings units, cargo, hero XP and items home (event handler).
func (c *pctx) missionReturn(id int64) error {
	m, err := c.lockMission(id)
	if err != nil {
		return err
	}
	if m.Status != "returning" {
		return nil
	}
	ci := c.city(m.FromCityID)
	if ci == nil {
		return fmt.Errorf("mission %d returns to foreign city", m.ID)
	}
	for u, n := range m.Units {
		ci.Units[u] += n
	}
	if amountsTotal(m.Cargo) > 0 {
		lost, err := c.grant(ci, m.Cargo, "mission_return", fmt.Sprintf("mission:%d", m.ID))
		if err != nil {
			return err
		}
		addResult(m, "lostToStorage", lost)
	}
	if st, ok := m.Result["stats"].(map[string]any); ok {
		for k, v := range st {
			if f, ok := v.(float64); ok {
				c.stat(k, int64(f))
			}
		}
	}
	if items, ok := m.Result["items"].([]any); ok {
		for _, it := range items {
			if s, ok := it.(string); ok {
				if err := c.grantItem(c.pl.ID, s, fmt.Sprintf("mission:%d", m.ID)); err != nil {
					return err
				}
			}
		}
	}
	if m.HeroID != nil {
		h, _, err := c.heroByID(*m.HeroID, true)
		if err != nil {
			return err
		}
		h.Status = "idle"
		if xp, ok := m.Result["heroXp"].(float64); ok {
			c.gainXP(h, int64(xp))
		}
		if w, ok := m.Result["heroWounded"].(bool); ok && w {
			wu := c.now.Add(time.Duration(c.p.Params.HeroWoundHours) * time.Hour)
			h.WoundedUntil = &wu
		}
		if err := c.saveHero(h); err != nil {
			return err
		}
	}
	m.Status = "done"
	if err := c.saveMission(m); err != nil {
		return err
	}
	if err := c.refreshAway(ci); err != nil {
		return err
	}
	return notify(c.ctx, c.tx, c.pl.ID, "mission_returned", map[string]any{"missionId": m.ID, "kind": m.Kind, "units": m.Units, "cargo": m.Cargo})
}

// Recall orders an outbound or stationed mission home.
func (c *pctx) Recall(id int64) (*Mission, error) {
	m, err := c.lockMission(id)
	if err != nil {
		return nil, err
	}
	if m.PlayerID != c.pl.ID {
		return nil, notFound("mission")
	}
	switch m.Status {
	case "outbound":
		// A mission that has reached its due time is resolved by its arrival
		// event (possibly in another player's transaction); it can no longer be
		// recalled, otherwise an attack could be cancelled after arriving (R-MSN-12).
		if !c.now.Before(m.ArriveAt) {
			return nil, conflict("already_arrived", "the mission has already reached its target")
		}
		// the pending arrival event becomes a no-op because status changes
		addResult(m, "outcome", "recalled")
		return m, c.sendBack(m, c.now.Sub(m.DepartAt))
	case "stationed":
		addResult(m, "outcome", "withdrawn")
		return m, c.sendBack(m, m.ArriveAt.Sub(m.DepartAt))
	}
	return nil, rule("not_recallable", "mission cannot be recalled in status %s", m.Status)
}

type MissionView struct {
	*Mission
	Direction string `json:"direction"` // own | incoming | stationed_here
	Owner     string `json:"owner"`
}

// Missions lists own missions plus foreign missions targeting own cities.
func (c *pctx) Missions() ([]MissionView, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT `+prefixed("m", missionCols)+`, p.name FROM missions m JOIN players p ON p.id=m.player_id
		WHERE m.status <> 'done' AND (m.player_id=$1 OR m.target_city_id IN (SELECT id FROM cities WHERE player_id=$1))
		ORDER BY m.arrive_at, m.id LIMIT 500`, c.pl.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MissionView{}
	for rows.Next() {
		m := &Mission{}
		var owner string
		if err := rows.Scan(&m.ID, &m.PlayerID, &m.FromCityID, &m.Kind, &m.TargetX, &m.TargetY, &m.TargetCityID, &m.Units, &m.UpkeepUnits,
			&m.HeroID, &m.Cargo, &m.Options, &m.DepartAt, &m.ArriveAt, &m.ReturnAt, &m.Status, &m.Result, &owner); err != nil {
			return nil, err
		}
		v := MissionView{Mission: m, Owner: owner, Direction: "own"}
		if m.PlayerID != c.pl.ID {
			v.Direction = "incoming"
			if m.Status == "stationed" {
				v.Direction = "stationed_here"
			}
			if m.Kind == "attack" || m.Kind == "scout" {
				// Defenders see that something is coming, not its composition (GAME_RULES R-SCT-4).
				v.Units = nil
				v.Cargo = nil
				v.Result = nil
				v.HeroID = nil
			}
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func prefixed(alias, cols string) string {
	out := ""
	for i, col := range splitCols(cols) {
		if i > 0 {
			out += ", "
		}
		out += alias + "." + col
	}
	return out
}

func splitCols(cols string) []string {
	var out []string
	cur := ""
	for _, ch := range cols {
		switch ch {
		case ',':
			out = append(out, cur)
			cur = ""
		case ' ', '\n', '\t':
		default:
			cur += string(ch)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
