package rules

import (
	"math"
	"sort"

	"realmforge/server/internal/pack"
)

// Stack is one unit type on one side of a battle, with modifiers already
// applied to per-unit stats (see BuildStack).
type Stack struct {
	UnitID  string            `json:"unit"`
	Class   string            `json:"class"`
	Count   int64             `json:"count"`
	Attack  float64           `json:"attack"`
	Defense float64           `json:"defense"`
	HP      float64           `json:"hp"`
	Bonuses []pack.ClassBonus `json:"-"`
	Score   int               `json:"-"`
	Group   int               `json:"-"` // owner group (garrison, each reinforcement) for loss attribution
	wounds  float64           // damage carried over between rounds
}

// BuildStack applies modifiers in the fixed order (GAME_RULES R-CMB-1):
// stat = (base + flat) × (1 + Σpct(stat.unitId, stat.class, stat.*)/100).
// Defense is a damage-reduction percentage; wallPts adds percentage points
// for city defenders (capped later by BattleOptions.DefenseCap).
func BuildStack(u *pack.Unit, count int64, m *Mods, wallPts float64) Stack {
	if m == nil {
		m = NewMods()
	}
	tg := func(stat string) []string { return []string{stat + "." + u.ID, stat + "." + u.Class, stat + ".*"} }
	def := m.Apply(u.Defense, tg("defense")...) + wallPts
	return Stack{
		UnitID: u.ID, Class: u.Class, Count: count,
		Attack: m.Apply(u.Attack, tg("attack")...), Defense: def, HP: math.Max(1, m.Apply(u.HP, tg("hp")...)),
		Bonuses: u.Bonuses, Score: u.Score,
	}
}

type RoundLog struct {
	Round        int              `json:"round"`
	AttackerDmg  float64          `json:"attackerDamage"`
	DefenderDmg  float64          `json:"defenderDamage"`
	AttackerLost map[string]int64 `json:"attackerLost"`
	DefenderLost map[string]int64 `json:"defenderLost"`
}

type BattleResult struct {
	Rounds            []RoundLog       `json:"rounds"`
	AttackerWon       bool             `json:"attackerWon"`
	AttackerStart     map[string]int64 `json:"attackerStart"`
	DefenderStart     map[string]int64 `json:"defenderStart"`
	AttackerLosses    map[string]int64 `json:"attackerLosses"`
	DefenderLosses    map[string]int64 `json:"defenderLosses"`
	AttackerLeft      map[string]int64 `json:"attackerLeft"`
	DefenderLeft      map[string]int64 `json:"defenderLeft"`
	AttackerScoreLost int64            `json:"attackerScoreLost"`
	DefenderScoreLost int64            `json:"defenderScoreLost"`
	AttackerRetreated bool             `json:"attackerRetreated"`
	DefenderRetreated bool             `json:"defenderRetreated"`
	AttackerFinal     []Stack          `json:"-"`
	DefenderFinal     []Stack          `json:"-"`
}

// GroupCounts returns surviving counts per unit for one owner group.
func GroupCounts(st []Stack, group int) map[string]int64 {
	m := map[string]int64{}
	for _, s := range st {
		if s.Group == group {
			m[s.UnitID] += s.Count
		}
	}
	return m
}

// BattleOptions are per-battle rule parameters.
type BattleOptions struct {
	Rounds        int     // maximum rounds
	DefenseCap    float64 // max damage reduction in percent
	AttRetreatPct int     // attacker retreats once this % of its units is lost (100 = never)
	DefRetreatPct int     // defender retreats likewise
}

func alive(s []Stack) bool {
	for _, x := range s {
		if x.Count > 0 {
			return true
		}
	}
	return false
}

func bonusVs(s Stack, class string) float64 {
	for _, b := range s.Bonuses {
		if b.VsClass == class {
			return b.Pct
		}
	}
	return 0
}

// damage computes damage dealt by `from` to each stack of `to`, distributed
// proportionally to each target stack's unit count (GAME_RULES R-CMB-3).
func damage(from, to []Stack, defCap float64) []float64 {
	out := make([]float64, len(to))
	var total int64
	for _, t := range to {
		total += max(0, t.Count)
	}
	if total <= 0 {
		return out
	}
	for _, s := range from {
		if s.Count <= 0 || s.Attack <= 0 {
			continue
		}
		for j, t := range to {
			if t.Count <= 0 {
				continue
			}
			share := float64(t.Count) / float64(total)
			raw := float64(s.Count) * s.Attack * share * (1 + bonusVs(s, t.Class)/100)
			out[j] += raw * (1 - math.Min(defCap, math.Max(0, t.Defense))/100)
		}
	}
	return out
}

func applyDamage(st []Stack, dmg []float64, lost map[string]int64) {
	for i := range st {
		if st[i].Count <= 0 || dmg[i] <= 0 {
			continue
		}
		total := dmg[i] + st[i].wounds
		kills := int64(math.Floor(total / st[i].HP))
		if kills >= st[i].Count {
			lost[st[i].UnitID] += st[i].Count
			st[i].Count = 0
			st[i].wounds = 0
			continue
		}
		st[i].Count -= kills
		st[i].wounds = total - float64(kills)*st[i].HP
		lost[st[i].UnitID] += kills
	}
}

func counts(s []Stack) map[string]int64 {
	m := map[string]int64{}
	for _, x := range s {
		m[x.UnitID] += x.Count
	}
	return m
}

func total(m map[string]int64) int64 {
	var n int64
	for _, v := range m {
		n += v
	}
	return n
}

// Resolve runs a deterministic simultaneous-fire battle (GAME_RULES R-CMB-2..6):
//
//   - up to opt.Rounds rounds; both sides deal damage computed from the counts
//     at the start of the round;
//   - damage = count · attack · share(target count) · (1 + class bonus) · (1 − min(cap, defense%)/100);
//   - partial damage carries over as wounds within the battle;
//   - after each round a side whose cumulative unit losses reach its retreat
//     threshold retreats; if both retreat in the same round the defender holds;
//   - the attacker wins iff the defender is destroyed or retreats while the
//     attacker still has units and did not retreat.
func Resolve(att, def []Stack, opt BattleOptions) BattleResult {
	if opt.Rounds <= 0 {
		opt.Rounds = 1
	}
	if opt.AttRetreatPct <= 0 {
		opt.AttRetreatPct = 100
	}
	if opt.DefRetreatPct <= 0 {
		opt.DefRetreatPct = 100
	}
	att = append([]Stack(nil), att...)
	def = append([]Stack(nil), def...)
	byKey := func(st []Stack) func(i, j int) bool {
		return func(i, j int) bool {
			if st[i].Group != st[j].Group {
				return st[i].Group < st[j].Group
			}
			return st[i].UnitID < st[j].UnitID
		}
	}
	sort.SliceStable(att, byKey(att))
	sort.SliceStable(def, byKey(def))
	res := BattleResult{AttackerStart: counts(att), DefenderStart: counts(def),
		AttackerLosses: map[string]int64{}, DefenderLosses: map[string]int64{}}
	attTotal, defTotal := total(res.AttackerStart), total(res.DefenderStart)
	for r := 1; r <= opt.Rounds && alive(att) && alive(def); r++ {
		da := damage(att, def, opt.DefenseCap)
		dd := damage(def, att, opt.DefenseCap)
		log := RoundLog{Round: r, AttackerLost: map[string]int64{}, DefenderLost: map[string]int64{}}
		for _, x := range da {
			log.AttackerDmg += x
		}
		for _, x := range dd {
			log.DefenderDmg += x
		}
		applyDamage(def, da, log.DefenderLost)
		applyDamage(att, dd, log.AttackerLost)
		for k, v := range log.AttackerLost {
			res.AttackerLosses[k] += v
		}
		for k, v := range log.DefenderLost {
			res.DefenderLosses[k] += v
		}
		log.AttackerDmg = math.Round(log.AttackerDmg*100) / 100
		log.DefenderDmg = math.Round(log.DefenderDmg*100) / 100
		res.Rounds = append(res.Rounds, log)
		attOut := alive(att) && opt.AttRetreatPct < 100 && total(res.AttackerLosses)*100 >= int64(opt.AttRetreatPct)*attTotal
		defOut := alive(def) && opt.DefRetreatPct < 100 && total(res.DefenderLosses)*100 >= int64(opt.DefRetreatPct)*defTotal
		if attOut || defOut {
			res.AttackerRetreated = attOut
			res.DefenderRetreated = defOut && !attOut
			break
		}
	}
	res.AttackerLeft = counts(att)
	res.DefenderLeft = counts(def)
	res.AttackerFinal, res.DefenderFinal = att, def
	res.AttackerWon = alive(att) && !res.AttackerRetreated && (!alive(def) || res.DefenderRetreated)
	scoreOf := map[string]int{}
	for _, s := range att {
		scoreOf[s.UnitID] = s.Score
	}
	for _, s := range def {
		scoreOf[s.UnitID] = s.Score
	}
	for u, n := range res.AttackerLosses {
		res.AttackerScoreLost += n * int64(scoreOf[u])
	}
	for u, n := range res.DefenderLosses {
		res.DefenderScoreLost += n * int64(scoreOf[u])
	}
	return res
}

// Power is a rough army strength indicator for UI and matchmaking:
// Σ count × (attack + defense) × hp / 10.
func Power(st []Stack) float64 {
	p := 0.0
	for _, s := range st {
		p += float64(s.Count) * (s.Attack + s.Defense) * s.HP / 10
	}
	return math.Round(p)
}

// Loot computes plunder (GAME_RULES R-CMB-7):
//
//	available_r = max(0, stock_r − protected_r) × lootPct/100  (lootable resources only)
//	if Σavailable ≤ capacity: take all; else take floor(available_r × capacity / Σavailable).
func Loot(stockUnits pack.Amounts, protected map[string]int64, lootable func(string) bool, lootPct float64, capacity int64) pack.Amounts {
	avail := map[string]int64{}
	var total int64
	for _, r := range sortedKeys(stockUnits) {
		if !lootable(r) {
			continue
		}
		a := int64(math.Floor(float64(max(0, stockUnits[r]-protected[r])) * lootPct / 100))
		if a > 0 {
			avail[r] = a
			total += a
		}
	}
	out := pack.Amounts{}
	if total == 0 || capacity <= 0 {
		return out
	}
	for r, a := range avail {
		if total <= capacity {
			out[r] = a
		} else {
			out[r] = int64(math.Floor(float64(a) * float64(capacity) / float64(total)))
		}
	}
	return out
}

// ScoutResult describes an espionage outcome (GAME_RULES R-SCT-1..3).
type ScoutResult struct {
	Success  bool `json:"success"`
	Detected bool `json:"detected"`
	Intel    int  `json:"intel"` // 0 none, 1 resources+units, 2 + buildings
}

// Scout resolves espionage. attPower/defPower are Σ count × scout stat.
// roll is uniform in [0,1) from the mission-seeded RNG.
//
//	success  ⇔ attPower > defPower
//	intel    = 2 if attPower ≥ 2×defPower, else 1 (on success)
//	P(detect)= base + (100−base) × defPower/(attPower+defPower)   (percent)
//	failure  ⇒ always detected
func Scout(attPower, defPower float64, basePct int, roll float64) ScoutResult {
	if attPower <= 0 {
		return ScoutResult{Detected: true}
	}
	if attPower <= defPower {
		return ScoutResult{Detected: true}
	}
	r := ScoutResult{Success: true, Intel: 1}
	if attPower >= 2*defPower {
		r.Intel = 2
	}
	p := (float64(basePct) + (100-float64(basePct))*defPower/(attPower+defPower)) / 100
	r.Detected = roll < p
	return r
}
