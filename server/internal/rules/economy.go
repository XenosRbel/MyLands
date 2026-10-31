package rules

import (
	"math"
	"sort"

	"realmforge/server/internal/pack"
)

// Milli is the fixed-point scale for stored resource amounts: 1 unit = 1000 milli.
// Integer storage avoids float drift; production remainders below 1 milli are
// floored per settlement (max loss 0.001 unit per resource per settlement).
const Milli = 1000

// CostAt returns the cost of reaching level (level >= 1):
// cost_r(L) = round(base_r × growth^(L-1)).
func CostAt(c pack.Cost, level int) pack.Amounts {
	out := pack.Amounts{}
	f := math.Pow(c.Growth, float64(level-1))
	for r, b := range c.Base {
		out[r] = int64(math.Round(float64(b) * f))
	}
	return out
}

// ScaleAmounts multiplies amounts by n (training batches), saturating at 1e15.
func ScaleAmounts(a pack.Amounts, n int64) pack.Amounts {
	out := pack.Amounts{}
	for r, v := range a {
		x := float64(v) * float64(n)
		if x > 1e15 {
			x = 1e15
		}
		out[r] = int64(x)
	}
	return out
}

// DurationAt returns seconds for reaching level:
// t(L) = ceil(base × growth^(L-1) × timeFactor / worldSpeed), min 1 s.
func DurationAt(d pack.Duration, level int, timeFactor, worldSpeed float64) int64 {
	t := float64(d.BaseSeconds) * math.Pow(d.Growth, float64(level-1)) * timeFactor / worldSpeed
	return max(1, int64(math.Ceil(t-1e-9)))
}

// EffectValue returns the cumulative value of an effect at level L:
// v(L) = perLevel × L × growth^(L-1).
func EffectValue(e pack.Effect, level int) float64 {
	if level <= 0 {
		return 0
	}
	g := e.Growth
	if g == 0 {
		g = 1
	}
	return e.PerLevel * float64(level) * math.Pow(g, float64(level-1))
}

// EconInput is everything that determines a city's economic rates.
type EconInput struct {
	Pack       *pack.Loaded
	Buildings  map[string]int   // building -> level
	Mods       *Mods            // race + techs (+ hero/territory) modifiers
	Population int64            // current population
	Garrison   map[string]int64 // units at home (pay upkeep, may desert)
	AwayUpkeep float64          // upkeep/hour of units away, already × MissionUpkeepMult, before modifiers
}

// Rates are derived (never stored) economy figures for a city.
type Rates struct {
	Gross         map[string]float64 `json:"gross"`       // production per hour
	Consumption   map[string]float64 `json:"consumption"` // per hour
	Net           map[string]float64 `json:"net"`
	Storage       map[string]int64   `json:"storage"` // capacity in whole units (capped resources only)
	Protected     map[string]int64   `json:"protected"`
	Housing       int64              `json:"housing"`
	WorkersNeeded int64              `json:"workersNeeded"`
	Employment    float64            `json:"employment"` // 0..1
	Idle          int64              `json:"idle"`
	Upkeep        float64            `json:"upkeep"` // per hour, in params.UpkeepResource
}

// UpkeepResource returns the resource that pays army upkeep.
func UpkeepResource(p pack.Params) string {
	if p.UpkeepResource != "" {
		return p.UpkeepResource
	}
	return p.FoodResource
}

// ComputeRates derives production, consumption, storage, protection and housing
// (GAME_RULES §3):
//
//	employed   = min(pop, Σ workers(b)·L_b);  idle = pop − employed
//	employment = employed / Σ workers   (1 if no workers are needed)
//	gross_r    = (Σ_b v_b(L) · employment + Σ idle_tax_r · idle) · (1 + pct(production.r|*)/100) + flat
//	food       = pop · foodPerPop
//	upkeep     = (Σ garrison·unit.upkeep + awayUpkeep) · (1 + pct(upkeep)/100)
//	storage_r  = (baseStorage_r + Σ storage effects) · (1 + pct(storage.r|*)/100)
func ComputeRates(in EconInput) Rates {
	p := in.Pack
	mods := in.Mods
	if mods == nil {
		mods = NewMods()
	}
	r := Rates{Gross: map[string]float64{}, Consumption: map[string]float64{}, Net: map[string]float64{},
		Storage: map[string]int64{}, Protected: map[string]int64{}}
	rawProd := map[string]float64{}
	idleTax := map[string]float64{}
	storageAll, protectAll := 0.0, 0.0
	storage := map[string]float64{}
	protect := map[string]float64{}
	housing := 0.0
	for _, id := range sortedKeys(in.Buildings) {
		lvl := in.Buildings[id]
		b := p.Building(id)
		if b == nil || lvl <= 0 {
			continue
		}
		for _, e := range b.Effects {
			v := EffectValue(e, lvl)
			switch e.Type {
			case "produce":
				rawProd[e.Resource] += v
				r.WorkersNeeded += int64(e.Workers) * int64(lvl)
			case "idle_tax":
				idleTax[e.Resource] += v
			case "storage":
				if e.Resource == "*" {
					storageAll += v
				} else {
					storage[e.Resource] += v
				}
			case "protect":
				if e.Resource == "*" {
					protectAll += v
				} else {
					protect[e.Resource] += v
				}
			case "housing":
				housing += v
			}
		}
	}
	r.Housing = int64(math.Floor(mods.Apply(housing, "housing")))
	employed := min(in.Population, r.WorkersNeeded)
	r.Idle = max(0, in.Population-employed)
	r.Employment = 1
	if r.WorkersNeeded > 0 {
		r.Employment = float64(employed) / float64(r.WorkersNeeded)
	}
	for _, res := range p.Resources {
		g := rawProd[res.ID]*r.Employment + idleTax[res.ID]*float64(r.Idle)
		// Order per R-MOD-1: (base + Σflat) × (1 + Σpct/100).
		g = (g + mods.Flat("production."+res.ID, "production.*")) * math.Max(0, 1+mods.Pct("production."+res.ID, "production.*")/100)
		r.Gross[res.ID] = math.Max(0, g)
		if res.Capped {
			s := float64(res.BaseStorage) + storageAll + storage[res.ID]
			r.Storage[res.ID] = int64(math.Floor(mods.Apply(s, "storage."+res.ID, "storage.*")))
		}
		pr := protectAll + protect[res.ID]
		if pr > 0 || mods.Flat("protect."+res.ID, "protect.*") > 0 {
			r.Protected[res.ID] = int64(math.Floor(mods.Apply(pr, "protect."+res.ID, "protect.*")))
		}
	}
	upkeep := in.AwayUpkeep
	for _, id := range sortedKeys(in.Garrison) {
		if u := p.Unit(id); u != nil {
			upkeep += float64(in.Garrison[id]) * u.Upkeep
		}
	}
	r.Upkeep = mods.Apply(upkeep, "upkeep")
	r.Consumption[p.Params.FoodResource] += float64(in.Population) * p.Params.FoodPerPopPerHour
	r.Consumption[UpkeepResource(p.Params)] += r.Upkeep
	// World speed scales all flows (per real hour) but not stocks or capacities.
	if ws := p.Params.WorldSpeed; ws != 1 {
		for k := range r.Gross {
			r.Gross[k] *= ws
		}
		for k := range r.Consumption {
			r.Consumption[k] *= ws
		}
		r.Upkeep *= ws
	}
	for _, res := range p.Resources {
		r.Net[res.ID] = r.Gross[res.ID] - r.Consumption[res.ID]
	}
	return r
}

// Accrue advances stock (milli units) by elapsedMs at the given net rates.
// Positive accrual of capped resources stops at the cap but never reduces a
// stock that is already above it (e.g. after loot). Stocks never go negative.
//
// When params.overflowDecayPctPerHour > 0 the cap does not stop production;
// instead the surplus decays at economy ticks (see TickStep).
func Accrue(stock map[string]int64, r Rates, p *pack.Loaded, elapsedMs int64) {
	if elapsedMs <= 0 {
		return
	}
	hardCap := p.Params.OverflowDecayPctPerHour == 0
	for _, res := range p.Resources {
		net := r.Net[res.ID]
		if net == 0 {
			continue
		}
		delta := int64(math.Floor(net * Milli * float64(elapsedMs) / 3_600_000))
		cur := stock[res.ID]
		next := cur + delta
		if delta > 0 && res.Capped && hardCap {
			capM := r.Storage[res.ID] * Milli
			if cur >= capM {
				next = cur
			} else if next > capM {
				next = capM
			}
		}
		if next < 0 {
			next = 0
		}
		stock[res.ID] = next
	}
}

// TickState is the state mutated at an economy tick.
type TickState struct {
	Population int64
	Garrison   map[string]int64
	Stock      map[string]int64   // milli
	Carry      map[string]float64 // fractional remainders (growth, desertion) kept between ticks
}

// TickStep applies one economy tick (GAME_RULES R-POP-1..4, R-ARMY-UPKEEP-2, R-RES-5):
//
//  1. Starvation: food stock 0 and net food < 0 → pop −= ceil(pop · starvationPct/100).
//     Otherwise growth: base = idle ("idle" mode) or free housing ("free_housing"),
//     Δ = base · growthPct/100 · tickHours (fraction carried), pop ≤ housing.
//  2. Desertion: upkeep stock 0 and net upkeep < 0 → each garrison stack loses
//     count · desertionPct/100 · tickHours (fraction carried per unit type).
//  3. Overflow decay (when enabled): surplus above storage loses decayPct/100 · tickHours.
//
// It reports whether population or garrison changed (rates must be recomputed).
func TickStep(st *TickState, r Rates, p *pack.Loaded, growthMods *Mods) bool {
	pr := p.Params
	tickHours := float64(pr.EconomyTickSeconds) / 3600 * pr.WorldSpeed
	changed := false
	food := pr.FoodResource
	pop := st.Population
	if st.Stock[food] <= 0 && r.Net[food] < 0 && pop > 0 {
		pop -= int64(math.Ceil(float64(pop) * pr.StarvationPctPerTick / 100))
		st.Carry["pop"] = 0
	} else if pop < r.Housing {
		base := float64(r.Housing - pop)
		if pr.PopGrowthMode == "idle" {
			base = float64(r.Idle)
		}
		g := base*pr.PopGrowthPctPerHour*math.Max(0, 1+growthMods.Pct("pop_growth")/100)/100*tickHours + st.Carry["pop"]
		whole := math.Floor(g)
		st.Carry["pop"] = g - whole
		pop += int64(whole)
	}
	pop = max(0, min(pop, r.Housing))
	if pop != st.Population {
		st.Population = pop
		changed = true
	}
	up := UpkeepResource(pr)
	if st.Stock[up] <= 0 && r.Net[up] < 0 && pr.DesertionPctPerHour > 0 {
		for _, id := range sortedKeys(st.Garrison) {
			n := st.Garrison[id]
			if n <= 0 {
				continue
			}
			key := "desert." + id
			d := float64(n)*pr.DesertionPctPerHour/100*tickHours + st.Carry[key]
			whole := math.Floor(d)
			st.Carry[key] = d - whole
			if whole > 0 {
				st.Garrison[id] = max(0, n-int64(whole))
				changed = true
			}
		}
	}
	if pr.OverflowDecayPctPerHour > 0 {
		for _, res := range p.Resources {
			c, ok := r.Storage[res.ID]
			if !ok {
				continue
			}
			if surplus := st.Stock[res.ID] - c*Milli; surplus > 0 {
				st.Stock[res.ID] -= int64(math.Ceil(float64(surplus) * pr.OverflowDecayPctPerHour / 100 * tickHours))
			}
		}
	}
	return changed
}

// CanAfford reports whether stock (milli) covers cost (units).
func CanAfford(stock map[string]int64, cost pack.Amounts) bool {
	for r, n := range cost {
		if stock[r] < n*Milli {
			return false
		}
	}
	return true
}

// Missing returns resources lacking (in whole units, rounded up).
func Missing(stock map[string]int64, cost pack.Amounts) pack.Amounts {
	out := pack.Amounts{}
	for r, n := range cost {
		if d := n*Milli - stock[r]; d > 0 {
			out[r] = (d + Milli - 1) / Milli
		}
	}
	return out
}

// Spend deducts cost; caller must check CanAfford first. Panics on overspend
// to make accounting bugs loud in tests.
func Spend(stock map[string]int64, cost pack.Amounts) {
	for r, n := range cost {
		stock[r] -= n * Milli
		if stock[r] < 0 {
			panic("rules.Spend: negative stock for " + r)
		}
	}
}

// Grant adds amounts (units). When caps is non-nil, capped resources are
// clamped to max(current, cap) — excess is lost (GAME_RULES R-RES-4).
// Pass nil caps in overflow-decay mode.
func Grant(stock map[string]int64, a pack.Amounts, caps map[string]int64) (lost pack.Amounts) {
	lost = pack.Amounts{}
	for r, n := range a {
		next := stock[r] + n*Milli
		if caps != nil {
			if c, ok := caps[r]; ok {
				capM := c * Milli
				limit := max(capM, stock[r])
				if next > limit {
					lost[r] = (next - limit) / Milli
					next = limit
				}
			}
		}
		stock[r] = next
	}
	return lost
}

// Whole converts milli stock to whole units (floor).
func Whole(stock map[string]int64) pack.Amounts {
	out := pack.Amounts{}
	for r, v := range stock {
		out[r] = v / Milli
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
