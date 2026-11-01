package sim

import (
	"fmt"
	"math"
	"sort"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

// PvEOptions tune the neutral-difficulty analysis.
type PvEOptions struct {
	MaxLossPct    float64 // largest acceptable share of the army lost (default 30)
	DistanceTiles float64 // one-way distance to the site (default 6)
}

func (o *PvEOptions) defaults() {
	if o.MaxLossPct <= 0 {
		o.MaxLossPct = 30
	}
	if o.DistanceTiles <= 0 {
		o.DistanceTiles = 6
	}
}

// PvERow is the cheapest single-unit-type army that clears one neutral level.
// Unit is empty when no unit of the race can do it within the loss limit.
type PvERow struct {
	Neutral     string       `json:"neutral"`
	Level       int          `json:"level"`
	GuardPower  float64      `json:"guardPower"`
	Unit        string       `json:"unit"`
	MinCount    int64        `json:"minCount"`
	LossPct     float64      `json:"lossPct"`
	TrainCost   pack.Amounts `json:"trainCost"`
	TrainValue  float64      `json:"trainValue"`
	RewardValue float64      `json:"rewardValue"`
	LossValue   float64      `json:"lossValue"`
	WageValue   float64      `json:"wageValue"`   // mission upkeep for the round trip
	NetValue    float64      `json:"netValue"`    // reward − losses − wages, per mission
	PaybackRuns float64      `json:"paybackRuns"` // missions to earn back the army (0 = never)
	TripHours   float64      `json:"tripHours"`
	RunsPerDay  float64      `json:"runsPerDay"` // limited by the site's respawn time
	XP          int64        `json:"xp"`
	Unpriced    []string     `json:"unpricedRewards,omitempty"` // reward resources without an exchange rate (valued at 1)
}

func valueOf(p *pack.Loaded, a pack.Amounts) float64 {
	v := 0.0
	for _, r := range sortedKeys(a) {
		x := 1.0
		if e, ok := p.Params.ExchangeRates[r]; ok {
			x = e
		}
		v += float64(a[r]) * x
	}
	return v
}

func raceMods(p *pack.Loaded, race string) *rules.Mods {
	m := rules.NewMods()
	if r := p.Race(race); r != nil {
		m.AddAll(r.Modifiers, 1)
	}
	return m
}

func guardStacks(p *pack.Loaded, guards map[string]int64) []rules.Stack {
	var out []rules.Stack
	for _, id := range sortedKeys(guards) {
		if u := p.Unit(id); u != nil && guards[id] > 0 {
			out = append(out, rules.BuildStack(u, guards[id], rules.NewMods(), 0))
		}
	}
	return out
}

func battleOpts(p *pack.Loaded) rules.BattleOptions {
	return rules.BattleOptions{Rounds: p.Params.CombatRounds, DefenseCap: p.Params.DefenseCapPct,
		AttRetreatPct: 100, DefRetreatPct: p.Params.DefenderRetreatPct}
}

func level(n *pack.Neutral, lvl int) *pack.NeutralLevel {
	for i := range n.Levels {
		if n.Levels[i].Level == lvl {
			return &n.Levels[i]
		}
	}
	return nil
}

// lossPct is the share of the attacker's units that die.
func lossPct(res rules.BattleResult) float64 {
	var start, lost int64
	for _, v := range res.AttackerStart {
		start += v
	}
	for _, v := range res.AttackerLosses {
		lost += v
	}
	if start == 0 {
		return 100
	}
	return 100 * float64(lost) / float64(start)
}

func fight(p *pack.Loaded, race, unit string, n int64, nl *pack.NeutralLevel) rules.BattleResult {
	u := p.Unit(unit)
	att := []rules.Stack{rules.BuildStack(u, n, raceMods(p, race), 0)}
	return rules.Resolve(att, guardStacks(p, nl.Guards), battleOpts(p))
}

// winsWithin reports whether n units of one type clear the level with at most
// maxLoss percent casualties.
func winsWithin(p *pack.Loaded, race, unit string, n int64, neutral string, lvl int, maxLoss float64) bool {
	nd := p.Neutral(neutral)
	if nd == nil || p.Unit(unit) == nil {
		return false
	}
	nl := level(nd, lvl)
	if nl == nil {
		return false
	}
	res := fight(p, race, unit, n, nl)
	return res.AttackerWon && lossPct(res) <= maxLoss
}

const maxArmy = 200000

// minArmy finds the smallest army of one unit that satisfies the limit.
func minArmy(p *pack.Loaded, race, unit string, nl *pack.NeutralLevel, maxLoss float64) (int64, rules.BattleResult, bool) {
	ok := func(n int64) (rules.BattleResult, bool) {
		res := fight(p, race, unit, n, nl)
		return res, res.AttackerWon && lossPct(res) <= maxLoss
	}
	hi := int64(1)
	var res rules.BattleResult
	for {
		var good bool
		if res, good = ok(hi); good {
			break
		}
		if hi >= maxArmy {
			return 0, res, false
		}
		hi *= 2
	}
	lo := hi / 2 // lo fails (or is 0)
	for hi-lo > 1 {
		mid := (lo + hi) / 2
		if r, good := ok(mid); good {
			hi, res = mid, r
		} else {
			lo = mid
		}
	}
	// A larger army may lose a smaller share, so the search is monotone enough
	// for a minimum; the test pins n−1 as failing.
	return hi, res, true
}

// PvE returns one row per neutral level for the given race. Only the race's
// own modifiers apply (no techs, heroes or items), which is the day-one state.
func PvE(p *pack.Loaded, race string, opt PvEOptions) []PvERow {
	opt.defaults()
	upkeepRes := rules.UpkeepResource(p.Params)
	upVal := valueOf(p, pack.Amounts{upkeepRes: 1})
	var rows []PvERow
	for ni := range p.Neutrals {
		n := &p.Neutrals[ni]
		for li := range n.Levels {
			nl := &n.Levels[li]
			row := PvERow{Neutral: n.ID, Level: nl.Level, XP: nl.XP, RewardValue: valueOf(p, nl.Rewards),
				GuardPower: rules.Power(guardStacks(p, nl.Guards))}
			for _, r := range sortedKeys(nl.Rewards) {
				if _, ok := p.Params.ExchangeRates[r]; !ok {
					row.Unpriced = append(row.Unpriced, r)
				}
			}
			var best *PvERow
			for ui := range p.Units {
				u := &p.Units[ui]
				if u.TrainedIn == "" || u.Attack <= 0 || !pack.Allowed(u.Races, race) {
					continue
				}
				cnt, res, ok := minArmy(p, race, u.ID, nl, opt.MaxLossPct)
				if !ok {
					continue
				}
				cost := rules.ScaleAmounts(u.Cost, cnt)
				c := row
				c.Unit, c.MinCount, c.LossPct = u.ID, cnt, math.Round(lossPct(res)*10)/10
				c.TrainCost, c.TrainValue = cost, valueOf(p, cost)
				if best == nil || c.TrainValue < best.TrainValue || (c.TrainValue == best.TrainValue && c.Unit < best.Unit) {
					cc := c
					best = &cc
				}
			}
			if best != nil {
				row = *best
				u := p.Unit(row.Unit)
				secs := rules.TravelSeconds(opt.DistanceTiles, u.Speed, raceMods(p, race).Pct("speed"), p.Params.WorldSpeed, p.Params.MinTravelSeconds)
				row.TripHours = math.Round(2*float64(secs)/36) / 100
				row.LossValue = row.TrainValue * row.LossPct / 100
				row.WageValue = float64(row.MinCount) * u.Upkeep * p.Params.MissionUpkeepMult * row.TripHours * p.Params.WorldSpeed * upVal
				row.NetValue = row.RewardValue - row.LossValue - row.WageValue
				if row.NetValue > 0 {
					row.PaybackRuns = math.Round(row.TrainValue/row.NetValue*10) / 10
				}
				cycle := math.Max(row.TripHours, float64(n.RespawnSeconds)/3600/p.Params.WorldSpeed)
				row.RunsPerDay = math.Round(24/cycle*10) / 10
			}
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].GuardPower < rows[j].GuardPower })
	return rows
}

// beatable names the strongest neutral level the current garrison clears with
// acceptable losses, or "-" when it clears none.
func (b *bot) beatable() string {
	var units map[string]int64 = b.st.Garrison
	mods := raceMods(b.p, b.race.ID)
	var att []rules.Stack
	for _, id := range sortedKeys(units) {
		if u := b.p.Unit(id); u != nil && units[id] > 0 && u.Attack > 0 {
			att = append(att, rules.BuildStack(u, units[id], mods, 0))
		}
	}
	if len(att) == 0 {
		return "-"
	}
	best, bestPow := "-", -1.0
	for ni := range b.p.Neutrals {
		n := &b.p.Neutrals[ni]
		for li := range n.Levels {
			nl := &n.Levels[li]
			res := rules.Resolve(att, guardStacks(b.p, nl.Guards), battleOpts(b.p))
			if res.AttackerWon && lossPct(res) <= 30 {
				if pw := rules.Power(guardStacks(b.p, nl.Guards)); pw > bestPow {
					best, bestPow = fmt.Sprintf("%s L%d", n.ID, nl.Level), pw
				}
			}
		}
	}
	return best
}
