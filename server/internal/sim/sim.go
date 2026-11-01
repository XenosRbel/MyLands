// Package sim runs a deterministic bot player against the pure rules engine.
//
// It uses the same functions as the server (rules.AdvanceCity, CostAt,
// DurationAt, Resolve) but no database, so a whole first day plays in
// milliseconds. The bot is a measuring instrument, not a model of human play:
// it shows what a simple greedy player experiences under a pack's numbers.
// Nothing here names a resource, building or unit; everything comes from the pack.
package sim

import (
	"fmt"
	"math"
	"sort"
	"time"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

// Config selects a pack, a race and the bot's behaviour.
type Config struct {
	Pack        *pack.Loaded
	Race        string
	Horizon     time.Duration // real time to simulate (default 24h)
	Step        time.Duration // decision granularity (default 5m)
	SampleEvery time.Duration // curve sampling interval (default 1h)
	// MilitaryPct is the share of current stock the bot spends on the army per
	// decision once the training building exists (default 20).
	MilitaryPct int
	// MilitaryStart delays the military building chain (default 4h).
	MilitaryStart time.Duration
	// WageAware stops training when the upkeep resource would go net-negative.
	WageAware bool
	// WorldSpeed overrides params.worldSpeed when > 0.
	WorldSpeed float64
}

func (c *Config) defaults() {
	if c.Horizon <= 0 {
		c.Horizon = 24 * time.Hour
	}
	if c.Step <= 0 {
		c.Step = 5 * time.Minute
	}
	if c.SampleEvery <= 0 {
		c.SampleEvery = time.Hour
	}
	if c.MilitaryPct <= 0 {
		c.MilitaryPct = 20
	}
	if c.MilitaryStart <= 0 {
		c.MilitaryStart = 4 * time.Hour
	}
}

// Sample is one point of the curve.
type Sample struct {
	Hour        float64          `json:"hour"`
	Stock       map[string]int64 `json:"stock"`
	Net         map[string]float64
	Population  int64   `json:"population"`
	Housing     int64   `json:"housing"`
	Employment  float64 `json:"employment"`
	BuildLevels int     `json:"buildLevels"`
	Garrison    int64   `json:"garrison"`
	Upkeep      float64 `json:"upkeepPerHour"`
	WagePct     float64 `json:"wagePct"` // upkeep as % of gross production of the upkeep resource
	IncomeValue float64 `json:"incomeValuePerHour"`
	Score       int64   `json:"score"`
	Beatable    string  `json:"beatable"` // strongest neutral level the garrison beats
}

// BuildEvent records a finished building level.
type BuildEvent struct {
	Hour     float64 `json:"hour"`
	Building string  `json:"building"`
	Level    int     `json:"level"`
}

// Stats are first-day health indicators, weighted by simulated time.
type Stats struct {
	MaxQueueLen      int     `json:"maxQueueLen"`
	QueueIdlePct     float64 `json:"buildQueueIdlePct"` // time with an empty build queue
	StarvedHours     float64 `json:"starvedHours"`      // food stock empty with a deficit
	PopAtHousingPct  float64 `json:"popAtHousingPct"`   // time population is capped by housing
	AtStorageCapPct  float64 `json:"atStorageCapPct"`   // time at least one capped stock is >= 99% full
	WageDeficitHours float64 `json:"wageDeficitHours"`  // upkeep resource empty with a deficit
	Trained          int64   `json:"unitsTrained"`
	Deserted         int64   `json:"unitsDeserted"`
	TargetUnit       string  `json:"targetUnit"`
	FirstArmyHour    float64 `json:"firstArmyHour"` // first time the garrison exceeds the starting one
}

// Result is one finished run.
type Result struct {
	Pack    string         `json:"pack"`
	Race    string         `json:"race"`
	Config  ConfigView     `json:"config"`
	Samples []Sample       `json:"samples"`
	Builds  []BuildEvent   `json:"builds"`
	Final   map[string]int `json:"finalBuildings"`
	Stats   Stats          `json:"stats"`
}

// ConfigView is the serialisable part of Config.
type ConfigView struct {
	HorizonHours  float64 `json:"horizonHours"`
	StepMinutes   float64 `json:"stepMinutes"`
	MilitaryPct   int     `json:"militaryPct"`
	MilitaryStart float64 `json:"militaryStartHours"`
	WageAware     bool    `json:"wageAware"`
	WorldSpeed    float64 `json:"worldSpeed"`
}

type job struct {
	kind   string // build | train
	item   string
	level  int
	qty    int64
	finish time.Time
}

type bot struct {
	cfg    Config
	p      *pack.Loaded
	race   *pack.Race
	st     rules.EconState
	bld    map[string]int
	queue  []job
	mods   *rules.Mods
	value  map[string]float64 // exchange value per unit of resource
	target *pack.Unit         // unit the army is built from
	chain  map[string]int     // buildings the target unit needs
	start  time.Time
	res    *Result
	rates  rules.Rates
}

// Run plays cfg.Horizon of game time and returns the curve and statistics.
func Run(cfg Config) (*Result, error) {
	cfg.defaults()
	p := cfg.Pack
	if p == nil {
		return nil, fmt.Errorf("sim: no pack")
	}
	race := p.Race(cfg.Race)
	if race == nil {
		return nil, fmt.Errorf("sim: unknown race %q", cfg.Race)
	}
	if cfg.WorldSpeed > 0 {
		p.Params.WorldSpeed = cfg.WorldSpeed
	}
	// A fixed instant aligned to the economy tick keeps runs reproducible.
	tick := time.Duration(p.Params.EconomyTickSeconds) * time.Second
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(tick)
	b := &bot{cfg: cfg, p: p, race: race, start: start, bld: map[string]int{}, value: map[string]float64{}}
	for k, v := range p.Start.Buildings {
		b.bld[k] = v
	}
	stock := map[string]int64{}
	for r, n := range p.Start.Resources {
		stock[r] = n * rules.Milli
	}
	garrison := map[string]int64{}
	var startGarrison int64
	for u, n := range p.Start.Units {
		garrison[u] = n
		startGarrison += n
	}
	b.st = rules.EconState{Stock: stock, Population: p.Start.Population, Garrison: garrison, Carry: map[string]float64{}, UpdatedAt: start}
	for _, r := range p.Resources {
		v := 1.0
		if x, ok := p.Params.ExchangeRates[r.ID]; ok {
			v = x
		}
		b.value[r.ID] = v
	}
	b.res = &Result{Pack: p.ID, Race: race.ID, Final: map[string]int{}, Config: ConfigView{
		HorizonHours: cfg.Horizon.Hours(), StepMinutes: cfg.Step.Minutes(), MilitaryPct: cfg.MilitaryPct,
		MilitaryStart: cfg.MilitaryStart.Hours(), WageAware: cfg.WageAware, WorldSpeed: p.Params.WorldSpeed}}
	b.pickTarget()
	b.res.Stats.TargetUnit = ""
	if b.target != nil {
		b.res.Stats.TargetUnit = b.target.ID
	}
	b.res.Stats.FirstArmyHour = -1
	b.loop(startGarrison)
	for k, v := range b.bld {
		b.res.Final[k] = v
	}
	return b.res, nil
}

func (b *bot) rebuildMods() {
	m := rules.NewMods()
	m.AddAll(b.race.Modifiers, 1)
	for _, id := range sortedKeys(b.bld) {
		if bd := b.p.Building(id); bd != nil {
			for _, e := range bd.Effects {
				if e.Type == "modifier" && e.Modifier != nil {
					m.Add(*e.Modifier, float64(b.bld[id]))
				}
			}
		}
	}
	b.mods = m
}

func (b *bot) input(bld map[string]int, pop int64) rules.EconInput {
	return rules.EconInput{Pack: b.p, Buildings: bld, Mods: b.mods, Population: pop, Garrison: b.st.Garrison}
}

func (b *bot) hours(t time.Time) float64 { return t.Sub(b.start).Hours() }

func (b *bot) loop(startGarrison int64) {
	end := b.start.Add(b.cfg.Horizon)
	t := b.start
	nextSample := b.start
	b.rebuildMods()
	b.rates = rules.ComputeRates(b.input(b.bld, b.st.Population))
	b.decide(t)
	var dtTotal, qIdle, popCap, stoCap, starve, wage float64
	for t.Before(end) {
		next := t.Add(b.cfg.Step)
		for _, j := range b.queue {
			if j.finish.After(t) && j.finish.Before(next) {
				next = j.finish
			}
		}
		if next.After(end) {
			next = end
		}
		popBefore := b.st.Population
		garBefore := total(b.st.Garrison)
		b.rates = rules.AdvanceCity(&b.st, next, b.input(b.bld, b.st.Population))
		dt := next.Sub(t).Hours()
		dtTotal += dt
		if !b.hasJob("build") {
			qIdle += dt
		}
		if b.rates.Housing > 0 && b.st.Population >= b.rates.Housing {
			popCap += dt
		}
		if b.nearCap() {
			stoCap += dt
		}
		if b.emptyAndShort(b.p.Params.FoodResource) {
			starve += dt
		}
		if b.emptyAndShort(rules.UpkeepResource(b.p.Params)) {
			wage += dt
		}
		if d := garBefore - total(b.st.Garrison); d > 0 {
			b.res.Stats.Deserted += d
		}
		_ = popBefore
		t = next
		b.complete(t)
		b.decide(t)
		if g := total(b.st.Garrison); b.res.Stats.FirstArmyHour < 0 && g > startGarrison {
			b.res.Stats.FirstArmyHour = b.hours(t)
		}
		if !t.Before(nextSample) {
			b.sample(t)
			nextSample = nextSample.Add(b.cfg.SampleEvery)
		}
	}
	if dtTotal > 0 {
		s := &b.res.Stats
		s.QueueIdlePct = round1(100 * qIdle / dtTotal)
		s.PopAtHousingPct = round1(100 * popCap / dtTotal)
		s.AtStorageCapPct = round1(100 * stoCap / dtTotal)
	}
	b.res.Stats.StarvedHours = round2(starve)
	b.res.Stats.WageDeficitHours = round2(wage)
}

func (b *bot) hasJob(kind string) bool {
	for _, j := range b.queue {
		if j.kind == kind {
			return true
		}
	}
	return false
}

func (b *bot) count(kind string) int {
	n := 0
	for _, j := range b.queue {
		if j.kind == kind {
			n++
		}
	}
	return n
}

// emptyAndShort: the stock is empty while the net rate is negative.
func (b *bot) emptyAndShort(res string) bool {
	return b.st.Stock[res] <= 0 && b.rates.Net[res] < 0
}

func (b *bot) nearCap() bool {
	for r, c := range b.rates.Storage {
		if c > 0 && b.st.Stock[r] >= c*rules.Milli*99/100 {
			return true
		}
	}
	return false
}

// complete applies every job that finished at or before t.
func (b *bot) complete(t time.Time) {
	kept := b.queue[:0]
	changed := false
	for _, j := range b.queue {
		if j.finish.After(t) {
			kept = append(kept, j)
			continue
		}
		switch j.kind {
		case "build":
			b.bld[j.item] = j.level
			b.res.Builds = append(b.res.Builds, BuildEvent{Hour: round2(b.hours(j.finish)), Building: j.item, Level: j.level})
			changed = true
		case "train":
			b.st.Garrison[j.item] += j.qty
			b.res.Stats.Trained += j.qty
			changed = true
		}
	}
	b.queue = kept
	if changed {
		b.rebuildMods()
		b.rates = rules.ComputeRates(b.input(b.bld, b.st.Population))
	}
}

func (b *bot) queueEnd(kind string, now time.Time) time.Time {
	end := now
	for _, j := range b.queue {
		if j.kind == kind && j.finish.After(end) {
			end = j.finish
		}
	}
	return end
}

func (b *bot) pendingLevel(id string) int {
	l := b.bld[id]
	for _, j := range b.queue {
		if j.kind == "build" && j.item == id && j.level > l {
			l = j.level
		}
	}
	return l
}

func (b *bot) unmet(reqs []pack.Requirement) bool {
	for _, r := range reqs {
		switch r.Type {
		case "building":
			if b.bld[r.ID] < r.Level {
				return true
			}
		case "tech":
			return true // the bot does not research
		case "race":
			if b.race.ID != r.ID {
				return true
			}
		}
	}
	return false
}

func (b *bot) costValue(c pack.Amounts) float64 {
	v := 0.0
	for _, r := range sortedKeys(c) {
		v += float64(c[r]) * b.value[r]
	}
	return v
}

// incomeValue is Σ net(r) × value(r) per hour.
func (b *bot) incomeValue(r rules.Rates) float64 {
	v := 0.0
	for _, id := range sortedKeys(r.Net) {
		v += r.Net[id] * b.value[id]
	}
	return v
}

// decide queues buildings and training for the current instant.
func (b *bot) decide(now time.Time) {
	for b.count("build") < b.p.Params.BuildQueueMax {
		c := b.pickAffordable(now)
		if c == nil {
			break
		}
		cost := rules.CostAt(c.b.Cost, c.level)
		rules.Spend(b.st.Stock, cost)
		secs := rules.DurationAt(c.b.Time, c.level, b.mods.TimeFactor("build_time"), b.p.Params.WorldSpeed)
		b.queue = append(b.queue, job{kind: "build", item: c.b.ID, level: c.level, finish: b.queueEnd("build", now).Add(time.Duration(secs) * time.Second)})
		if n := b.count("build"); n > b.res.Stats.MaxQueueLen {
			b.res.Stats.MaxQueueLen = n
		}
	}
	b.train(now)
}

type candidate struct {
	b     *pack.Building
	level int
	score float64
}

// maxWait is how long the bot will save for its favourite option before it
// buys something cheaper that it can afford now.
const maxWait = 15 * time.Minute

// pickAffordable returns the best option the bot can pay for now. It waits for
// the favourite only when the favourite is about to become affordable.
func (b *bot) pickAffordable(now time.Time) *candidate {
	for i, c := range b.candidates(now) {
		cost := rules.CostAt(c.b.Cost, c.level)
		if rules.CanAfford(b.st.Stock, cost) {
			return &c
		}
		if i == 0 && b.waitFor(cost) <= maxWait {
			return nil
		}
	}
	return nil
}

// waitFor estimates the time until cost is affordable at current net rates.
func (b *bot) waitFor(cost pack.Amounts) time.Duration {
	worst := 0.0
	for r, n := range rules.Missing(b.st.Stock, cost) {
		if n <= 0 {
			continue
		}
		rate := b.rates.Net[r]
		if rate <= 0 {
			return 1000 * time.Hour
		}
		worst = math.Max(worst, float64(n)/rate)
	}
	return time.Duration(worst * float64(time.Hour))
}

// candidates lists every legal next building level, best score first.
func (b *bot) candidates(now time.Time) []candidate {
	var list []candidate
	baseVal := b.incomeValue(b.rates)
	perCap := 0.0
	if b.st.Population > 0 {
		perCap = baseVal / float64(b.st.Population)
	}
	housingFull := b.rates.Housing > 0 && b.st.Population*10 >= b.rates.Housing*9
	military := now.Sub(b.start) >= b.cfg.MilitaryStart
	for i := range b.p.Buildings {
		bd := &b.p.Buildings[i]
		cur := b.pendingLevel(bd.ID)
		if cur >= bd.MaxLevel || !pack.Allowed(bd.Races, b.race.ID) || b.unmet(bd.Requires) {
			continue
		}
		lvl := cur + 1
		cost := rules.CostAt(bd.Cost, lvl)
		cv := b.costValue(cost)
		if cv <= 0 {
			cv = 1
		}
		if !b.canEverAfford(cost) {
			continue
		}
		trial := map[string]int{}
		for k, v := range b.bld {
			trial[k] = v
		}
		trial[bd.ID] = lvl
		tr := rules.ComputeRates(b.input(trial, b.st.Population))
		gain := b.incomeValue(tr) - baseVal
		food := b.p.Params.FoodResource
		if b.rates.Net[food] < 0 { // a food deficit is worth more than its price tag
			gain += (tr.Net[food] - b.rates.Net[food]) * b.value[food] * 4
		}
		if housingFull {
			gain += float64(tr.Housing-b.rates.Housing) * perCap
		}
		for r, c := range b.rates.Storage {
			if c > 0 && b.st.Stock[r] >= c*rules.Milli*85/100 {
				gain += float64(tr.Storage[r]-c) * b.value[r] * 0.05
			}
		}
		if need, ok := b.chain[bd.ID]; ok && military && cur < need {
			gain = math.Max(gain, 0) + baseVal*0.5
			if gain <= 0 {
				gain = 1
			}
		}
		if gain <= 0 {
			continue
		}
		list = append(list, candidate{b: bd, level: lvl, score: gain / cv})
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		return list[i].b.ID < list[j].b.ID
	})
	return list
}

// canEverAfford skips options that need a resource nothing produces.
func (b *bot) canEverAfford(cost pack.Amounts) bool {
	for r, n := range cost {
		if n > 0 && b.st.Stock[r] < n*rules.Milli && b.rates.Gross[r] <= 0 {
			return false
		}
	}
	return true
}

// pickTarget chooses the unit with the best power per cost among the race's
// trainable combat units and records the building chain it needs.
func (b *bot) pickTarget() {
	var best *pack.Unit
	bestScore := 0.0
	for i := range b.p.Units {
		u := &b.p.Units[i]
		if u.TrainedIn == "" || u.Attack <= 0 || !pack.Allowed(u.Races, b.race.ID) || hasTechReq(u.Requires) {
			continue
		}
		if !b.chainPossible(u) {
			continue
		}
		stack := rules.BuildStack(u, 1, b.mods0(), 0)
		pw := rules.Power([]rules.Stack{{Count: 1000, Attack: stack.Attack, Defense: stack.Defense, HP: stack.HP}})
		sc := pw / math.Max(1, b.costValue(u.Cost))
		if sc > bestScore || (sc == bestScore && best != nil && u.ID < best.ID) {
			best, bestScore = u, sc
		}
	}
	b.target = best
	b.chain = map[string]int{}
	if best != nil {
		b.addChain(best.TrainedIn, 1)
		for _, r := range best.Requires {
			if r.Type == "building" {
				b.addChain(r.ID, max(1, r.Level))
			}
		}
		b.addProducers(best)
	}
}

// addProducers adds the cheapest producer of every resource the unit costs
// that the starting city does not yet produce.
func (b *bot) addProducers(u *pack.Unit) {
	start := rules.ComputeRates(rules.EconInput{Pack: b.p, Buildings: b.p.Start.Buildings, Mods: b.mods0(), Population: b.p.Start.Population})
	for _, r := range sortedKeys(u.Cost) {
		if u.Cost[r] <= 0 || start.Gross[r] > 0 {
			continue
		}
		var pick *pack.Building
		pickCost := 0.0
		for i := range b.p.Buildings {
			bd := &b.p.Buildings[i]
			if !pack.Allowed(bd.Races, b.race.ID) || !producesResource(bd, r) {
				continue
			}
			if c := b.costValue(rules.CostAt(bd.Cost, 1)); pick == nil || c < pickCost {
				pick, pickCost = bd, c
			}
		}
		if pick != nil {
			b.addChain(pick.ID, 1)
		}
	}
}

func producesResource(bd *pack.Building, res string) bool {
	for _, e := range bd.Effects {
		if e.Type == "produce" && e.Resource == res {
			return true
		}
	}
	return false
}

func (b *bot) mods0() *rules.Mods {
	m := rules.NewMods()
	m.AddAll(b.race.Modifiers, 1)
	return m
}

func hasTechReq(reqs []pack.Requirement) bool {
	for _, r := range reqs {
		if r.Type == "tech" {
			return true
		}
	}
	return false
}

func (b *bot) chainPossible(u *pack.Unit) bool {
	bd := b.p.Building(u.TrainedIn)
	return bd != nil && pack.Allowed(bd.Races, b.race.ID)
}

func (b *bot) addChain(id string, lvl int) {
	if b.chain[id] >= lvl {
		return
	}
	b.chain[id] = lvl
	if bd := b.p.Building(id); bd != nil {
		for _, r := range bd.Requires {
			if r.Type == "building" {
				b.addChain(r.ID, max(1, r.Level))
			}
		}
	}
}

// train spends MilitaryPct of the stock on the target unit.
func (b *bot) train(now time.Time) {
	u := b.target
	if u == nil || b.bld[u.TrainedIn] < 1 || b.unmet(u.Requires) || now.Sub(b.start) < b.cfg.MilitaryStart {
		return
	}
	if b.count("train") >= b.p.Params.TrainQueueMax {
		return
	}
	qty := int64(math.MaxInt64)
	for r, c := range u.Cost {
		if c <= 0 {
			continue
		}
		budget := b.st.Stock[r] / rules.Milli * int64(b.cfg.MilitaryPct) / 100
		qty = min(qty, budget/c)
	}
	if qty == math.MaxInt64 || qty < 1 {
		return
	}
	if b.cfg.WageAware {
		up := rules.UpkeepResource(b.p.Params)
		per := b.mods.Apply(u.Upkeep, "upkeep") * b.p.Params.WorldSpeed
		if per > 0 {
			room := b.rates.Net[up]
			qty = min(qty, int64(math.Floor(math.Max(0, room)/per)))
		}
	}
	qty = min(qty, 5000)
	if qty < 1 {
		return
	}
	rules.Spend(b.st.Stock, rules.ScaleAmounts(u.Cost, qty))
	each := rules.DurationAt(pack.Duration{BaseSeconds: u.TrainSeconds, Growth: 1}, 1, b.mods.TimeFactor("train_time"), b.p.Params.WorldSpeed)
	b.queue = append(b.queue, job{kind: "train", item: u.ID, qty: qty, finish: b.queueEnd("train", now).Add(time.Duration(each*qty) * time.Second)})
}

func (b *bot) sample(t time.Time) {
	r := b.rates
	s := Sample{Hour: round2(b.hours(t)), Stock: rules.Whole(b.st.Stock), Net: map[string]float64{},
		Population: b.st.Population, Housing: r.Housing, Employment: round2(r.Employment), Upkeep: round2(r.Upkeep),
		IncomeValue: round1(b.incomeValue(r))}
	for id, v := range r.Net {
		s.Net[id] = round1(v)
	}
	for _, l := range b.bld {
		s.BuildLevels += l
	}
	s.Garrison = total(b.st.Garrison)
	if g := r.Gross[rules.UpkeepResource(b.p.Params)]; g > 0 {
		s.WagePct = round1(100 * r.Upkeep / g)
	}
	for id, l := range b.bld {
		if bd := b.p.Building(id); bd != nil {
			s.Score += int64(bd.Score) * int64(l)
		}
	}
	for id, n := range b.st.Garrison {
		if u := b.p.Unit(id); u != nil {
			s.Score += int64(u.Score) * n
		}
	}
	s.Beatable = b.beatable()
	b.res.Samples = append(b.res.Samples, s)
}

func total(m map[string]int64) int64 {
	var n int64
	for _, v := range m {
		n += v
	}
	return n
}

func round1(x float64) float64 { return math.Round(x*10) / 10 }
func round2(x float64) float64 { return math.Round(x*100) / 100 }

// sortedKeys returns map keys in order (stable reports).
func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
