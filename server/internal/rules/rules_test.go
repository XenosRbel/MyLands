package rules

import (
	"math"
	"testing"
	"time"

	"realmforge/server/internal/pack"
)

func testPack(t testing.TB) *pack.Loaded {
	t.Helper()
	l, err := pack.LoadFile("../pack/testdata/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestCostAndDuration(t *testing.T) {
	c := pack.Cost{Base: pack.Amounts{"wood": 100, "food": 33}, Growth: 1.5}
	got := CostAt(c, 3) // 100*2.25=225, 33*2.25=74.25→74
	if got["wood"] != 225 || got["food"] != 74 {
		t.Fatalf("CostAt = %v", got)
	}
	if CostAt(c, 1)["wood"] != 100 {
		t.Fatal("level 1 must equal base")
	}
	d := pack.Duration{BaseSeconds: 60, Growth: 2}
	if s := DurationAt(d, 3, 1, 1); s != 240 {
		t.Fatalf("DurationAt = %d", s)
	}
	if s := DurationAt(d, 1, 0.5, 2); s != 15 { // 60*0.5/2
		t.Fatalf("time factor/world speed: %d", s)
	}
	if s := DurationAt(pack.Duration{BaseSeconds: 1, Growth: 1}, 1, 0.2, 100); s != 1 {
		t.Fatalf("min duration must be 1, got %d", s)
	}
}

func TestModsOrder(t *testing.T) {
	m := NewMods()
	m.Add(pack.Modifier{Target: "attack.*", Op: "pct", Value: 10}, 1)
	m.Add(pack.Modifier{Target: "attack.infantry", Op: "pct", Value: 5}, 2)
	m.Add(pack.Modifier{Target: "attack.*", Op: "flat", Value: 2}, 1)
	// (10 + 2) * (1 + 20/100) = 14.4
	if v := m.Apply(10, "attack.spear", "attack.infantry", "attack.*"); math.Abs(v-14.4) > 1e-9 {
		t.Fatalf("Apply = %v", v)
	}
	m.Add(pack.Modifier{Target: "build_time", Op: "pct", Value: -50}, 3)
	if f := m.TimeFactor("build_time"); f != MinTimeFactor {
		t.Fatalf("time factor floor: %v", f)
	}
}

func TestRatesEmploymentAndStorage(t *testing.T) {
	p := testPack(t)
	in := EconInput{Pack: p, Buildings: map[string]int{"hall": 1, "lumber": 2, "farm": 1}, Mods: NewMods(), Population: 15}
	r := ComputeRates(in)
	// workers: lumber 2*10 + farm 1*10 = 30; pop 15 → employment 0.5, idle 0
	if r.WorkersNeeded != 30 || r.Employment != 0.5 {
		t.Fatalf("workers=%d employment=%v", r.WorkersNeeded, r.Employment)
	}
	if r.Gross["wood"] != 100 { // 200 * 0.5
		t.Fatalf("gross wood = %v", r.Gross["wood"])
	}
	// food consumption = 15 * 0.1 = 1.5 ; gross food = 100*0.5 = 50
	if math.Abs(r.Net["food"]-48.5) > 1e-9 {
		t.Fatalf("net food = %v", r.Net["food"])
	}
	if r.Storage["wood"] != 1000 || r.Housing != 100 {
		t.Fatalf("storage=%v housing=%v", r.Storage, r.Housing)
	}
	if _, ok := r.Storage["gold"]; ok {
		t.Fatal("uncapped resource must not have storage")
	}
	in.Mods.Add(pack.Modifier{Target: "production.wood", Op: "pct", Value: 10}, 1)
	if r := ComputeRates(in); math.Abs(r.Gross["wood"]-110) > 1e-9 {
		t.Fatalf("modified wood = %v", r.Gross["wood"])
	}
	// flat is added before the percentage (R-MOD-1): (100 + 20) × 1.1 = 132
	in.Mods.Add(pack.Modifier{Target: "production.wood", Op: "flat", Value: 20}, 1)
	if r := ComputeRates(in); math.Abs(r.Gross["wood"]-132) > 1e-9 {
		t.Fatalf("flat-then-pct wood = %v", r.Gross["wood"])
	}
}

func TestAccrueCapsAndFloors(t *testing.T) {
	p := testPack(t)
	r := Rates{Net: map[string]float64{"wood": 3600, "food": -3600, "gold": 36}, Storage: map[string]int64{"wood": 1000, "food": 1000}}
	stock := map[string]int64{"wood": 999_500, "food": 500, "gold": 0}
	Accrue(stock, r, p, 1000) // 1 second
	if stock["wood"] != 1_000_000 {
		t.Fatalf("wood should cap at 1000 units, got %d", stock["wood"])
	}
	if stock["food"] != 0 {
		t.Fatalf("food must floor at 0, got %d", stock["food"])
	}
	if stock["gold"] != 10 { // 36/h for 1s = 0.01 units = 10 milli
		t.Fatalf("gold = %d", stock["gold"])
	}
	over := map[string]int64{"wood": 2_000_000}
	Accrue(over, r, p, 1000)
	if over["wood"] != 2_000_000 {
		t.Fatal("stock above cap must not be reduced by positive production")
	}
	Accrue(over, r, p, -5)
	if over["wood"] != 2_000_000 {
		t.Fatal("negative elapsed must be a no-op")
	}
}

func tick(pop int64, housing int64, food int64, netFood float64, p *pack.Loaded) int64 {
	st := &TickState{Population: pop, Stock: map[string]int64{"food": food}, Carry: map[string]float64{}}
	TickStep(st, Rates{Housing: housing, Net: map[string]float64{"food": netFood}, Idle: pop / 2}, p, NewMods())
	return st.Population
}

func TestTickPopulation(t *testing.T) {
	p := testPack(t)
	// free_housing mode: 50 free * 10%/h * 1/6 h = 0.83 → carry; after 2 ticks ≥ 1
	if got := tick(50, 100, 1000, 10, p); got != 50 {
		t.Fatalf("fractional growth should be carried, got %d", got)
	}
	st := &TickState{Population: 50, Stock: map[string]int64{"food": 1000}, Carry: map[string]float64{}}
	r := Rates{Housing: 100, Net: map[string]float64{"food": 10}}
	TickStep(st, r, p, NewMods())
	TickStep(st, r, p, NewMods())
	if st.Population != 51 {
		t.Fatalf("carried growth: %d", st.Population)
	}
	if got := tick(100, 100, 1000, 10, p); got != 100 {
		t.Fatalf("at cap: %d", got)
	}
	if got := tick(100, 100, 0, -5, p); got != 99 { // 1% starvation
		t.Fatalf("starvation: %d", got)
	}
	if got := tick(150, 100, 1000, 10, p); got != 100 {
		t.Fatalf("over housing must clamp: %d", got)
	}
}

func TestDesertionAndIdleMode(t *testing.T) {
	l := testPack(t)
	p := *l.Pack
	p.Params.UpkeepResource = "gold"
	p.Params.DesertionPctPerHour = 6 // 1% per 10-min tick
	p.Params.PopGrowthMode = "idle"
	l2, err := pack.Build(&p)
	if err != nil {
		t.Fatal(err)
	}
	st := &TickState{Population: 100, Garrison: map[string]int64{"spear": 1000}, Stock: map[string]int64{"food": 1, "gold": 0}, Carry: map[string]float64{}}
	r := Rates{Housing: 200, Idle: 60, Net: map[string]float64{"food": 1, "gold": -10}}
	TickStep(st, r, l2, NewMods())
	if st.Garrison["spear"] != 990 {
		t.Fatalf("desertion: %d", st.Garrison["spear"])
	}
	// idle mode: 60 idle * 10%/h /6 = 1 per tick
	if st.Population != 101 {
		t.Fatalf("idle growth: %d", st.Population)
	}
}

func TestIdleTaxAndUpkeep(t *testing.T) {
	l := testPack(t)
	p := *l.Pack
	p.Buildings = append([]pack.Building(nil), p.Buildings...)
	p.Buildings[0].Effects = append(p.Buildings[0].Effects, pack.Effect{Type: "idle_tax", Resource: "gold", PerLevel: 0.1})
	p.Params.UpkeepResource = "gold"
	l2, err := pack.Build(&p)
	if err != nil {
		t.Fatal(err)
	}
	r := ComputeRates(EconInput{Pack: l2, Buildings: map[string]int{"hall": 2, "lumber": 1}, Population: 50,
		Garrison: map[string]int64{"spear": 10}, AwayUpkeep: 4})
	// idle = 50-10 = 40; tax = 0.1*2 (level 2) * 40 = 8 gold/h; upkeep = 10*1 + 4 = 14
	if r.Idle != 40 || math.Abs(r.Gross["gold"]-8) > 1e-9 || r.Upkeep != 14 || math.Abs(r.Net["gold"]+6) > 1e-9 {
		t.Fatalf("idle=%d gross gold=%v upkeep=%v net=%v", r.Idle, r.Gross["gold"], r.Upkeep, r.Net["gold"])
	}
}

func TestOverflowDecay(t *testing.T) {
	l := testPack(t)
	p := *l.Pack
	p.Params.OverflowDecayPctPerHour = 6 // 1% per tick
	l2, _ := pack.Build(&p)
	r := Rates{Net: map[string]float64{"wood": 3600}, Storage: map[string]int64{"wood": 1000}}
	stock := map[string]int64{"wood": 999_000}
	Accrue(stock, r, l2, 2000) // +2 units, not capped in decay mode
	if stock["wood"] != 1_001_000 {
		t.Fatalf("decay mode must not hard cap: %d", stock["wood"])
	}
	st := &TickState{Population: 0, Stock: map[string]int64{"wood": 1_100_000, "food": 10}, Carry: map[string]float64{}}
	TickStep(st, Rates{Storage: map[string]int64{"wood": 1000}, Net: map[string]float64{}}, l2, NewMods())
	if st.Stock["wood"] != 1_099_000 { // surplus 100 units → -1 unit
		t.Fatalf("decay: %d", st.Stock["wood"])
	}
}

// Settling once vs. many times must give the same result (path independence).
func TestAdvancePathIndependence(t *testing.T) {
	p := testPack(t)
	in := EconInput{Pack: p, Buildings: map[string]int{"hall": 3, "lumber": 3, "farm": 2}, Mods: NewMods(), AwayUpkeep: 5}
	start := time.Date(2026, 1, 1, 10, 3, 17, 0, time.UTC)
	end := start.Add(7*time.Hour + 13*time.Minute + 9*time.Second)
	mk := func() *EconState {
		return &EconState{Stock: map[string]int64{"wood": 100_000, "food": 200_000}, Population: 20, UpdatedAt: start}
	}
	a := mk()
	AdvanceCity(a, end, in)
	b := mk()
	for t2 := start; t2.Before(end); t2 = t2.Add(137 * time.Second) {
		AdvanceCity(b, t2, in)
	}
	AdvanceCity(b, end, in)
	if a.Population != b.Population {
		t.Fatalf("population diverged: %d vs %d", a.Population, b.Population)
	}
	for r := range a.Stock {
		d := a.Stock[r] - b.Stock[r]
		if d < -300 || d > 300 { // ≤1 milli per settlement (~190 settlements)
			t.Fatalf("%s diverged: %d vs %d", r, a.Stock[r], b.Stock[r])
		}
	}
	if a.Population <= 20 {
		t.Fatal("population should have grown")
	}
}

func TestStarvationOverLongAbsence(t *testing.T) {
	p := testPack(t)
	// Huge army upkeep, no farms: food runs out, population declines.
	in := EconInput{Pack: p, Buildings: map[string]int{"hall": 5}, Mods: NewMods(), AwayUpkeep: 500}
	s := &EconState{Stock: map[string]int64{"food": 1000 * Milli}, Population: 200, UpdatedAt: time.Unix(0, 0).UTC()}
	AdvanceCity(s, s.UpdatedAt.Add(48*time.Hour), in)
	if s.Stock["food"] != 0 {
		t.Fatalf("food should be exhausted: %d", s.Stock["food"])
	}
	if s.Population >= 200 {
		t.Fatalf("population should starve: %d", s.Population)
	}
}

func TestSpendGrant(t *testing.T) {
	stock := map[string]int64{"wood": 500 * Milli}
	cost := pack.Amounts{"wood": 200, "food": 1}
	if CanAfford(stock, cost) {
		t.Fatal("food missing, must not afford")
	}
	if m := Missing(stock, cost); m["food"] != 1 || m["wood"] != 0 {
		t.Fatalf("missing = %v", m)
	}
	Spend(stock, pack.Amounts{"wood": 200})
	if stock["wood"] != 300*Milli {
		t.Fatal("spend")
	}
	lost := Grant(stock, pack.Amounts{"wood": 900}, map[string]int64{"wood": 1000})
	if stock["wood"] != 1000*Milli || lost["wood"] != 200 {
		t.Fatalf("grant cap: %d lost %v", stock["wood"], lost)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("overspend must panic")
		}
	}()
	Spend(stock, pack.Amounts{"wood": 5000})
}

func stacks(t *testing.T, p *pack.Loaded, m map[string]int64, wall float64) []Stack {
	var out []Stack
	for id, n := range m {
		out = append(out, BuildStack(p.Unit(id), n, NewMods(), wall))
	}
	return out
}

var opt4 = BattleOptions{Rounds: 4, DefenseCap: 90}

func TestCombatDeterministicAndConservative(t *testing.T) {
	p := testPack(t)
	att := stacks(t, p, map[string]int64{"spear": 100, "rider": 20}, 0)
	def := stacks(t, p, map[string]int64{"spear": 60}, 0)
	r1 := Resolve(att, def, opt4)
	r2 := Resolve(att, def, opt4)
	if r1.AttackerWon != r2.AttackerWon || r1.AttackerLosses["spear"] != r2.AttackerLosses["spear"] {
		t.Fatal("combat must be deterministic")
	}
	for u, n := range r1.AttackerStart {
		if r1.AttackerLosses[u]+r1.AttackerLeft[u] != n {
			t.Fatalf("attacker unit conservation broken for %s", u)
		}
	}
	for u, n := range r1.DefenderStart {
		if r1.DefenderLosses[u]+r1.DefenderLeft[u] != n {
			t.Fatalf("defender unit conservation broken for %s", u)
		}
	}
	if !r1.AttackerWon {
		t.Fatalf("overwhelming attacker should win: %+v", r1)
	}
	// Inputs must not be mutated.
	if att[0].Count == 0 && att[1].Count == 0 {
		t.Fatal("Resolve mutated input")
	}
}

func TestCombatWallHelpsDefender(t *testing.T) {
	p := testPack(t)
	att := stacks(t, p, map[string]int64{"spear": 50}, 0)
	plain := Resolve(att, stacks(t, p, map[string]int64{"spear": 50}, 0), opt4)
	walled := Resolve(att, stacks(t, p, map[string]int64{"spear": 50}, 30), opt4)
	if walled.DefenderLosses["spear"] >= plain.DefenderLosses["spear"] {
		t.Fatalf("wall should reduce defender losses: %d vs %d", walled.DefenderLosses["spear"], plain.DefenderLosses["spear"])
	}
	if plain.AttackerWon {
		t.Fatal("equal armies: defence must hold (ties go to defender)")
	}
}

func TestCombatEmptyDefender(t *testing.T) {
	p := testPack(t)
	r := Resolve(stacks(t, p, map[string]int64{"spear": 1}, 0), nil, opt4)
	if !r.AttackerWon || len(r.Rounds) != 0 {
		t.Fatalf("empty defender: %+v", r)
	}
	r = Resolve(nil, stacks(t, p, map[string]int64{"spear": 1}, 0), opt4)
	if r.AttackerWon {
		t.Fatal("no attackers cannot win")
	}
}

func TestCombatWoundsCarryOver(t *testing.T) {
	p := testPack(t)
	// 1 spear deals 10*(1-5%)=9.5 per round to an 80hp rider: kill needs wound carry-over.
	r := Resolve(stacks(t, p, map[string]int64{"spear": 1}, 0), stacks(t, p, map[string]int64{"rider": 1}, 0), BattleOptions{Rounds: 20, DefenseCap: 90})
	if r.DefenderLosses["rider"] != 1 && r.AttackerLosses["spear"] != 1 {
		t.Fatalf("someone must die eventually via wound carry-over: %+v", r)
	}
}

func TestRetreat(t *testing.T) {
	p := testPack(t)
	att := stacks(t, p, map[string]int64{"spear": 100}, 0)
	def := stacks(t, p, map[string]int64{"spear": 100}, 0)
	r := Resolve(att, def, BattleOptions{Rounds: 20, DefenseCap: 90, AttRetreatPct: 10})
	if !r.AttackerRetreated || r.AttackerWon {
		t.Fatalf("attacker should retreat at 10%% losses: %+v", r)
	}
	if lost := r.AttackerLosses["spear"]; lost < 10 || lost > 40 {
		t.Fatalf("retreat should stop the battle early, lost %d", lost)
	}
	r = Resolve(stacks(t, p, map[string]int64{"spear": 200}, 0), def, BattleOptions{Rounds: 20, DefenseCap: 90, DefRetreatPct: 20})
	if !r.DefenderRetreated || !r.AttackerWon {
		t.Fatalf("defender retreat gives attacker the win: %+v", r)
	}
	r = Resolve(att, def, BattleOptions{Rounds: 20, DefenseCap: 90, AttRetreatPct: 5, DefRetreatPct: 5})
	if r.AttackerWon || r.DefenderRetreated {
		t.Fatalf("simultaneous retreat: defender holds: %+v", r)
	}
}

func TestDefenseCap(t *testing.T) {
	d := damage([]Stack{{UnitID: "a", Count: 10, Attack: 10, HP: 1}}, []Stack{{UnitID: "b", Count: 1, Defense: 500, HP: 1}}, 90)
	if math.Abs(d[0]-10) > 1e-9 {
		t.Fatalf("defense must be capped at 90%%: %v", d[0])
	}
}

func TestClassBonus(t *testing.T) {
	p := testPack(t)
	u := *p.Unit("rider")
	s := BuildStack(&u, 10, NewMods(), 0)
	if bonusVs(s, "ranged") != 50 || bonusVs(s, "infantry") != 0 {
		t.Fatal("class bonus lookup")
	}
}

func TestLoot(t *testing.T) {
	stock := pack.Amounts{"wood": 1000, "food": 500, "gem": 100}
	lootable := func(r string) bool { return r != "gem" }
	got := Loot(stock, map[string]int64{"wood": 200}, lootable, 50, 10_000)
	if got["wood"] != 400 || got["food"] != 250 || got["gem"] != 0 {
		t.Fatalf("uncapped loot = %v", got)
	}
	got = Loot(stock, map[string]int64{"wood": 200}, lootable, 50, 65) // total avail 650 → 10%
	if got["wood"] != 40 || got["food"] != 25 {
		t.Fatalf("capacity-limited loot = %v", got)
	}
	if len(Loot(stock, nil, lootable, 50, 0)) != 0 {
		t.Fatal("zero capacity loots nothing")
	}
}

func TestScout(t *testing.T) {
	if r := Scout(10, 10, 30, 0.99); r.Success || !r.Detected {
		t.Fatal("equal scouting power fails and is detected")
	}
	r := Scout(20, 0, 30, 0.5)
	if !r.Success || r.Intel != 2 || r.Detected { // p = 30%
		t.Fatalf("unopposed scout: %+v", r)
	}
	if r := Scout(20, 0, 30, 0.1); !r.Detected {
		t.Fatal("roll below base detects")
	}
	if r := Scout(15, 10, 0, 0.99); r.Intel != 1 {
		t.Fatal("ratio <2 gives partial intel")
	}
}

func TestTravel(t *testing.T) {
	if d := Distance(0, 0, 3, 4); d != 5 {
		t.Fatal("distance")
	}
	if s := TravelSeconds(5, 10, 0, 1, 60); s != 1800 {
		t.Fatalf("travel = %d", s)
	}
	if s := TravelSeconds(0.1, 10, 0, 1, 60); s != 60 {
		t.Fatal("min travel")
	}
	if s := TravelSeconds(0.1, 10, 0, 10, 60); s != 6 {
		t.Fatalf("min travel scales with world speed: %d", s)
	}
	if s := TravelSeconds(5, 10, 100, 2, 1); s != 450 {
		t.Fatalf("speed mods: %d", s)
	}
}

func TestMapDeterministic(t *testing.T) {
	p := testPack(t)
	a := GenerateMap(42, p)
	b := GenerateMap(42, p)
	c := GenerateMap(43, p)
	if len(a) != 16*16 {
		t.Fatal("size")
	}
	same, diff := true, false
	for i := range a {
		if a[i] != b[i] {
			same = false
		}
		if a[i] != c[i] {
			diff = true
		}
	}
	if !same || !diff {
		t.Fatal("map must be deterministic per seed and vary across seeds")
	}
	for _, tl := range a {
		if tl.Neutral != "" && !p.Terrain(tl.Terrain).Passable {
			t.Fatal("neutral on impassable terrain")
		}
	}
}

func TestXPLevels(t *testing.T) {
	if XPForLevel(100, 1.5, 1) != 0 || XPForLevel(100, 1.5, 2) != 100 || XPForLevel(100, 1.5, 3) != 250 {
		t.Fatal("xp curve")
	}
	if LevelForXP(100, 1.5, 30, 249) != 2 || LevelForXP(100, 1.5, 30, 250) != 3 || LevelForXP(100, 1.5, 3, 1e9) != 3 {
		t.Fatal("level for xp")
	}
}

func TestRNG(t *testing.T) {
	a, b := NewRNG(SeedFrom(1, 2)), NewRNG(SeedFrom(1, 2))
	for i := 0; i < 100; i++ {
		x := a.Float()
		if x != b.Float() || x < 0 || x >= 1 {
			t.Fatal("rng")
		}
	}
	if SeedFrom(1, 2) == SeedFrom(2, 1) {
		t.Fatal("seed order should matter")
	}
}
