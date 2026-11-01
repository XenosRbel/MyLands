package game

import (
	"fmt"
	"time"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

type Job struct {
	ID         int64        `json:"id"`
	Kind       string       `json:"kind"`
	Item       string       `json:"item"`
	Level      int          `json:"level"`
	Quantity   int64        `json:"quantity"`
	Cost       pack.Amounts `json:"cost"`
	StartsAt   time.Time    `json:"startsAt"`
	FinishesAt time.Time    `json:"finishesAt"`
	CityID     int64        `json:"cityId"`
}

type NextStep struct {
	Level      int                `json:"level"`
	Cost       pack.Amounts       `json:"cost"`
	Seconds    int64              `json:"seconds"`
	Unmet      []pack.Requirement `json:"unmet"`
	Affordable bool               `json:"affordable"`
	Blocked    string             `json:"blocked,omitempty"` // machine-readable reason the action is unavailable
}

type BuildingView struct {
	ID          string    `json:"id"`
	Level       int       `json:"level"`
	TargetLevel int       `json:"targetLevel"` // after queued upgrades
	MaxLevel    int       `json:"maxLevel"`
	Next        *NextStep `json:"next,omitempty"`
}

type TechView struct {
	ID          string    `json:"id"`
	Level       int       `json:"level"`
	TargetLevel int       `json:"targetLevel"`
	MaxLevel    int       `json:"maxLevel"`
	Next        *NextStep `json:"next,omitempty"`
}

type UnitView struct {
	ID        string             `json:"id"`
	Count     int64              `json:"count"`
	Trainable bool               `json:"trainable"`
	Unmet     []pack.Requirement `json:"unmet"`
	Seconds   int64              `json:"secondsEach"`
	Blocked   string             `json:"blocked,omitempty"`
}

type CityView struct {
	City       *City              `json:"city"`
	Resources  pack.Amounts       `json:"resources"`
	Rates      rules.Rates        `json:"rates"`
	ServerTime time.Time          `json:"serverTime"`
	Buildings  []BuildingView     `json:"buildings"`
	Techs      []TechView         `json:"techs"`
	Units      []UnitView         `json:"units"`
	Queues     map[string][]Job   `json:"queues"`
	Away       map[string]int64   `json:"unitsAway"`
	Foreign    []StationedGroup   `json:"foreignTroops"`
	Mods       map[string]float64 `json:"modifiers"`
}

type StationedGroup struct {
	MissionID int64            `json:"missionId"`
	PlayerID  int64            `json:"playerId"`
	Player    string           `json:"player"`
	Units     map[string]int64 `json:"units"`
}

func (c *pctx) unmet(ci *City, reqs []pack.Requirement) []pack.Requirement {
	out := []pack.Requirement{}
	for _, r := range reqs {
		switch r.Type {
		case "building":
			if ci.Buildings[r.ID] < r.Level {
				out = append(out, r)
			}
		case "tech":
			if c.pl.Techs[r.ID] < r.Level {
				out = append(out, r)
			}
		case "race":
			if c.pl.Race != r.ID {
				out = append(out, r)
			}
		}
	}
	return out
}

func (c *pctx) openJobs(cityID int64, kind string) ([]Job, error) {
	q := `SELECT id, kind, item_id, level, quantity, cost, starts_at, finishes_at, city_id FROM jobs WHERE status='queued' AND kind=$2 AND `
	arg := cityID
	if kind == "research" {
		q += `player_id=$1`
		arg = c.pl.ID
	} else {
		q += `city_id=$1`
	}
	rows, err := c.tx.Query(c.ctx, q+` ORDER BY finishes_at, id`, arg, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Job{}
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Kind, &j.Item, &j.Level, &j.Quantity, &j.Cost, &j.StartsAt, &j.FinishesAt, &j.CityID); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func targetLevel(cur int, jobs []Job, item string) int {
	t := cur
	for _, j := range jobs {
		if j.Item == item && j.Level > t {
			t = j.Level
		}
	}
	return t
}

func queueEnd(now time.Time, jobs []Job) time.Time {
	end := now
	for _, j := range jobs {
		if j.FinishesAt.After(end) {
			end = j.FinishesAt
		}
	}
	return end
}

func (c *pctx) worldSpeed() float64 { return c.p.Params.WorldSpeed }

func (c *pctx) buildStep(ci *City, b *pack.Building, jobs []Job) *NextStep {
	tl := targetLevel(ci.Buildings[b.ID], jobs, b.ID)
	if tl >= b.MaxLevel {
		return nil
	}
	lvl := tl + 1
	cost := rules.CostAt(b.Cost, lvl)
	n := &NextStep{Level: lvl, Cost: cost, Unmet: c.unmet(ci, b.Requires),
		Seconds:    rules.DurationAt(b.Time, lvl, c.cityMods(ci).TimeFactor("build_time"), c.worldSpeed()),
		Affordable: rules.CanAfford(ci.Res, cost)}
	switch {
	case !pack.Allowed(b.Races, c.pl.Race):
		n.Blocked = "race"
	case len(n.Unmet) > 0:
		n.Blocked = "requirements"
	case len(jobs) >= c.p.Params.BuildQueueMax:
		n.Blocked = "queue_full"
	case !n.Affordable:
		n.Blocked = "resources"
	}
	return n
}

func (c *pctx) techStep(ci *City, t *pack.Tech, jobs []Job) *NextStep {
	tl := targetLevel(c.pl.Techs[t.ID], jobs, t.ID)
	if tl >= t.MaxLevel {
		return nil
	}
	lvl := tl + 1
	cost := rules.CostAt(t.Cost, lvl)
	n := &NextStep{Level: lvl, Cost: cost, Unmet: c.unmet(ci, t.Requires),
		Seconds:    rules.DurationAt(t.Time, lvl, c.cityMods(ci).TimeFactor("research_time"), c.worldSpeed()),
		Affordable: rules.CanAfford(ci.Res, cost)}
	switch {
	case !pack.Allowed(t.Races, c.pl.Race):
		n.Blocked = "race"
	case c.exclusiveTaken(t, jobs) != "":
		n.Blocked = "specialisation_taken"
	case len(n.Unmet) > 0:
		n.Blocked = "requirements"
	case len(jobs) >= c.p.Params.ResearchQueueMax:
		n.Blocked = "queue_full"
	case !n.Affordable:
		n.Blocked = "resources"
	}
	return n
}

// clearOtherExclusive removes earlier completed paths from the same group.
// This preserves one active specialisation per player.
func (c *pctx) clearOtherExclusive(t *pack.Tech) bool {
	if t == nil || t.ExclusiveGroup == "" {
		return false
	}
	changed := false
	for _, other := range c.p.Techs {
		if other.ID != t.ID && other.ExclusiveGroup == t.ExclusiveGroup && c.pl.Techs[other.ID] > 0 {
			delete(c.pl.Techs, other.ID)
			changed = true
		}
	}
	if changed {
		c.recomputeMods()
	}
	return changed
}

// exclusiveTaken returns the tech of the same exclusive group already
// researched or queued (specialisations are mutually exclusive), or "".
func (c *pctx) exclusiveTaken(t *pack.Tech, jobs []Job) string {
	if t.ExclusiveGroup == "" {
		return ""
	}
	for _, o := range c.p.Techs {
		if o.ID == t.ID || o.ExclusiveGroup != t.ExclusiveGroup {
			continue
		}
		// A completed calling can be replaced by studying a new calling.
		// Prevent concurrent paths in the research queue.
		for _, j := range jobs {
			if j.Item == o.ID {
				return o.ID
			}
		}
	}
	return ""
}

// View builds the full city screen model.
func (c *pctx) View(ci *City) (*CityView, error) {
	v := &CityView{City: ci, Resources: rules.Whole(ci.Res), Rates: c.rates(ci), ServerTime: c.now,
		Queues: map[string][]Job{}, Away: map[string]int64{}, Mods: c.cityMods(ci).Snapshot()}
	for _, k := range []string{"build", "research", "train"} {
		j, err := c.openJobs(ci.ID, k)
		if err != nil {
			return nil, err
		}
		v.Queues[k] = j
	}
	for i := range c.p.Buildings {
		b := &c.p.Buildings[i]
		if !pack.Allowed(b.Races, c.pl.Race) {
			continue
		}
		v.Buildings = append(v.Buildings, BuildingView{ID: b.ID, Level: ci.Buildings[b.ID], MaxLevel: b.MaxLevel,
			TargetLevel: targetLevel(ci.Buildings[b.ID], v.Queues["build"], b.ID), Next: c.buildStep(ci, b, v.Queues["build"])})
	}
	for i := range c.p.Techs {
		t := &c.p.Techs[i]
		if !pack.Allowed(t.Races, c.pl.Race) {
			continue
		}
		v.Techs = append(v.Techs, TechView{ID: t.ID, Level: c.pl.Techs[t.ID], MaxLevel: t.MaxLevel,
			TargetLevel: targetLevel(c.pl.Techs[t.ID], v.Queues["research"], t.ID), Next: c.techStep(ci, t, v.Queues["research"])})
	}
	for i := range c.p.Units {
		u := &c.p.Units[i]
		if u.TrainedIn == "" || !pack.Allowed(u.Races, c.pl.Race) {
			continue
		}
		uv := UnitView{ID: u.ID, Count: ci.Units[u.ID], Unmet: c.unmet(ci, u.Requires),
			Seconds: rules.DurationAt(pack.Duration{BaseSeconds: u.TrainSeconds, Growth: 1}, 1, c.cityMods(ci).TimeFactor("train_time"), c.worldSpeed())}
		switch {
		case ci.Buildings[u.TrainedIn] < 1:
			uv.Blocked = "building"
			uv.Unmet = append(uv.Unmet, pack.Requirement{Type: "building", ID: u.TrainedIn, Level: 1})
		case len(uv.Unmet) > 0:
			uv.Blocked = "requirements"
		case len(v.Queues["train"]) >= c.p.Params.TrainQueueMax:
			uv.Blocked = "queue_full"
		}
		uv.Trainable = uv.Blocked == ""
		v.Units = append(v.Units, uv)
	}
	rows, err := c.tx.Query(c.ctx, `SELECT units FROM missions WHERE from_city_id=$1 AND status <> 'done'`, ci.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var u map[string]int64
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return nil, err
		}
		for k, n := range u {
			v.Away[k] += n
		}
	}
	rows.Close()
	v.Foreign, err = stationedAtCity(c, ci.ID)
	return v, err
}

func stationedAtCity(c *pctx, cityID int64) ([]StationedGroup, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT m.id, m.player_id, p.name, m.units FROM missions m JOIN players p ON p.id=m.player_id
		WHERE m.target_city_id=$1 AND m.status='stationed' ORDER BY m.id`, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StationedGroup{}
	for rows.Next() {
		var g StationedGroup
		if err := rows.Scan(&g.MissionID, &g.PlayerID, &g.Player, &g.Units); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (c *pctx) insertJob(ci *City, kind, item string, level int, qty int64, cost pack.Amounts, start time.Time, seconds int64) (*Job, error) {
	j := &Job{Kind: kind, Item: item, Level: level, Quantity: qty, Cost: cost, StartsAt: start,
		FinishesAt: start.Add(time.Duration(seconds) * time.Second), CityID: ci.ID}
	err := c.tx.QueryRow(c.ctx, `INSERT INTO jobs(world_id, player_id, city_id, kind, item_id, level, quantity, cost, starts_at, finishes_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		c.pl.WorldID, c.pl.ID, ci.ID, kind, item, level, qty, cost, j.StartsAt, j.FinishesAt).Scan(&j.ID)
	if err != nil {
		return nil, err
	}
	return j, schedule(c.ctx, c.tx, c.pl.WorldID, c.pl.ID, "job", j.ID, j.FinishesAt)
}

func blockedErr(n *NextStep) error {
	switch n.Blocked {
	case "":
		return nil
	case "resources":
		e := rule("insufficient_resources", "not enough resources")
		return e
	case "requirements":
		e := rule("requirements_not_met", "requirements not met")
		e.Details = map[string]any{"unmet": n.Unmet}
		return e
	case "queue_full":
		return rule("queue_full", "queue is full")
	case "race":
		return rule("race_forbidden", "not available for your race")
	case "specialisation_taken":
		return rule("specialisation_taken", "another specialisation of this group is already chosen")
	}
	return rule(n.Blocked, "action unavailable")
}

// Build queues the next level of a building (GAME_RULES R-BLD-1..6).
func (c *pctx) Build(ci *City, buildingID string) (*Job, error) {
	b := c.p.Building(buildingID)
	if b == nil {
		return nil, invalid("unknown building %q", buildingID)
	}
	jobs, err := c.openJobs(ci.ID, "build")
	if err != nil {
		return nil, err
	}
	n := c.buildStep(ci, b, jobs)
	if n == nil {
		return nil, rule("max_level", "building is at maximum level")
	}
	if err := c.checkFreeCell(ci, b.ID, jobs); err != nil {
		return nil, err
	}
	if err := blockedErr(n); err != nil {
		if e, ok := AsError(err); ok && e.Code == "insufficient_resources" {
			e.Details = map[string]any{"missing": rules.Missing(ci.Res, n.Cost)}
		}
		return nil, err
	}
	if err := c.spend(ci, n.Cost, "build", fmt.Sprintf("%s:%d", b.ID, n.Level)); err != nil {
		return nil, err
	}
	return c.insertJob(ci, "build", b.ID, n.Level, 1, n.Cost, queueEnd(c.now, jobs), n.Seconds)
}

// Research queues the next level of a technology (player-wide effect).
func (c *pctx) Research(ci *City, techID string) (*Job, error) {
	t := c.p.Tech(techID)
	if t == nil {
		return nil, invalid("unknown technology %q", techID)
	}
	jobs, err := c.openJobs(ci.ID, "research")
	if err != nil {
		return nil, err
	}
	n := c.techStep(ci, t, jobs)
	if n == nil {
		return nil, rule("max_level", "technology is at maximum level")
	}
	if err := blockedErr(n); err != nil {
		if e, ok := AsError(err); ok && e.Code == "insufficient_resources" {
			e.Details = map[string]any{"missing": rules.Missing(ci.Res, n.Cost)}
		}
		return nil, err
	}
	if err := c.spend(ci, n.Cost, "research", fmt.Sprintf("%s:%d", t.ID, n.Level)); err != nil {
		return nil, err
	}
	// Changing a specialisation cancels the previous one when research begins.
	// This is atomic with the new research order and resource spend.
	c.clearOtherExclusive(t)
	return c.insertJob(ci, "research", t.ID, n.Level, 1, n.Cost, queueEnd(c.now, jobs), n.Seconds)
}

// MaxTrainBatch bounds a single training order.
const MaxTrainBatch = 100000

// Train queues a batch of units; the batch completes at once (GAME_RULES R-ARMY-2).
func (c *pctx) Train(ci *City, unitID string, qty int64) (*Job, error) {
	u := c.p.Unit(unitID)
	if u == nil || u.TrainedIn == "" {
		return nil, invalid("unknown or untrainable unit %q", unitID)
	}
	if qty < 1 || qty > MaxTrainBatch {
		return nil, invalid("quantity must be in [1, %d]", MaxTrainBatch)
	}
	if !pack.Allowed(u.Races, c.pl.Race) {
		return nil, rule("race_forbidden", "not available for your race")
	}
	if ci.Buildings[u.TrainedIn] < 1 {
		e := rule("requirements_not_met", "requires %s", u.TrainedIn)
		e.Details = map[string]any{"unmet": []pack.Requirement{{Type: "building", ID: u.TrainedIn, Level: 1}}}
		return nil, e
	}
	if un := c.unmet(ci, u.Requires); len(un) > 0 {
		e := rule("requirements_not_met", "requirements not met")
		e.Details = map[string]any{"unmet": un}
		return nil, e
	}
	jobs, err := c.openJobs(ci.ID, "train")
	if err != nil {
		return nil, err
	}
	if len(jobs) >= c.p.Params.TrainQueueMax {
		return nil, rule("queue_full", "training queue is full")
	}
	cost := rules.ScaleAmounts(u.Cost, qty)
	each := rules.DurationAt(pack.Duration{BaseSeconds: u.TrainSeconds, Growth: 1}, 1, c.cityMods(ci).TimeFactor("train_time"), c.worldSpeed())
	if err := c.spend(ci, cost, "train", fmt.Sprintf("%s:%d", u.ID, qty)); err != nil {
		return nil, err
	}
	return c.insertJob(ci, "train", u.ID, 0, qty, cost, queueEnd(c.now, jobs), each*qty)
}

// CancelJob cancels the LAST job of a queue (only the tail can be cancelled so
// other jobs' times stay valid) and refunds cancelRefundPct of its cost.
func (c *pctx) CancelJob(ci *City, jobID int64) (pack.Amounts, error) {
	var kind string
	var cityID, playerID int64
	var status string
	if err := c.tx.QueryRow(c.ctx, `SELECT kind, city_id, player_id, status FROM jobs WHERE id=$1 FOR UPDATE`, jobID).Scan(&kind, &cityID, &playerID, &status); err != nil {
		return nil, notFound("job")
	}
	if playerID != c.pl.ID || (kind != "research" && cityID != ci.ID) {
		return nil, notFound("job")
	}
	if status != "queued" {
		return nil, conflict("job_finished", "job is no longer queued")
	}
	jobs, err := c.openJobs(ci.ID, kind)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 || jobs[len(jobs)-1].ID != jobID {
		return nil, rule("not_last_in_queue", "only the last job of a queue can be cancelled")
	}
	j := jobs[len(jobs)-1]
	refund := pack.Amounts{}
	for r, n := range j.Cost {
		refund[r] = n * int64(c.p.Params.CancelRefundPct) / 100
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE jobs SET status='cancelled' WHERE id=$1`, jobID); err != nil {
		return nil, err
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE game_events SET status='cancelled' WHERE kind='job' AND ref_id=$1 AND status='pending'`, jobID); err != nil {
		return nil, err
	}
	target := c.city(j.CityID)
	if target == nil {
		target = ci
	}
	_, err = c.grant(target, refund, "cancel_refund", fmt.Sprintf("job:%d", jobID))
	return refund, err
}

// completeJob applies a finished job (event handler).
func (c *pctx) completeJob(jobID int64) error {
	var j Job
	var status string
	err := c.tx.QueryRow(c.ctx, `SELECT id, kind, item_id, level, quantity, city_id, status FROM jobs WHERE id=$1 FOR UPDATE`, jobID).
		Scan(&j.ID, &j.Kind, &j.Item, &j.Level, &j.Quantity, &j.CityID, &status)
	if err != nil {
		return err
	}
	if status != "queued" {
		return nil // cancelled or already applied
	}
	ci := c.city(j.CityID)
	if ci == nil {
		return fmt.Errorf("job %d city %d not owned by player %d", j.ID, j.CityID, c.pl.ID)
	}
	switch j.Kind {
	case "build":
		if ci.Buildings[j.Item] < j.Level {
			ci.Buildings[j.Item] = j.Level
		}
	case "research":
		if t := c.p.Tech(j.Item); t != nil {
			c.clearOtherExclusive(t)
		}
		if c.pl.Techs[j.Item] < j.Level {
			c.pl.Techs[j.Item] = j.Level
		}
		c.recomputeMods()
	case "train":
		ci.Units[j.Item] += j.Quantity
		c.stat("units_trained", j.Quantity)
		c.stat("units_trained."+j.Item, j.Quantity)
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE jobs SET status='done' WHERE id=$1`, j.ID); err != nil {
		return err
	}
	return notify(c.ctx, c.tx, c.pl.ID, j.Kind+"_done", map[string]any{"item": j.Item, "level": j.Level, "quantity": j.Quantity, "cityId": ci.ID})
}

// Rename renames a city.
func (c *pctx) Rename(ci *City, name string) error {
	if !validName(name) {
		return invalid("name must be 3-24 letters, digits, spaces, _ . ' -")
	}
	ci.Name = name
	return nil
}
