package rules

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"realmforge/server/internal/pack"
)

// tablePack extends the minimal test pack with level tables in the MyLands style.
func tablePack(t *testing.T) *pack.Loaded {
	t.Helper()
	raw, err := os.ReadFile("../pack/testdata/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m["paramDefs"] = []any{map[string]any{"id": "production_multiplier", "min": 0.1, "max": 10, "default": 1, "proposed": true}}
	m["serverProfiles"] = []any{
		map[string]any{"id": "base", "multipliers": map[string]any{}, "default": true, "proposed": true},
		map[string]any{"id": "rich", "multipliers": map[string]any{"production_multiplier": 2}},
	}
	m["buildingTables"] = []any{
		map[string]any{"building": "lumber", "levels": []any{map[string]any{"level": 1, "effects": map[string]any{
			"woodPerHour": map[string]any{"base": 15, "races": map[string]any{"human": 30}}},
			"overrides": map[string]any{"rich": map[string]any{"woodPerHour": 50}}}}},
		map[string]any{"building": "hall", "levels": []any{
			map[string]any{"level": 1, "effects": map[string]any{"housingCapacity": 400, "populationGrowth": 5, "resourceCapacity": 15000}},
			map[string]any{"level": 2, "effects": map[string]any{"housingCapacity": 500, "populationGrowth": 7, "resourceCapacity": 25000}}}},
		map[string]any{"building": "barracks", "levels": []any{map[string]any{"level": 1, "effects": map[string]any{
			"resourceCapacity": map[string]any{"base": 5000, "races": map[string]any{"human": 10000}}}}}},
	}
	b, _ := json.Marshal(m)
	l, err := pack.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestTableEffectRaceAndProfileOverride(t *testing.T) { // GAME_RULES R-ECON-TBL-1
	p := tablePack(t)
	if v, ok := TableEffect(p, "lumber", 1, "woodPerHour", "human", nil); !ok || v != 30 {
		t.Fatalf("race value: %v %v", v, ok)
	}
	if v, _ := TableEffect(p, "lumber", 1, "woodPerHour", "elf", nil); v != 15 {
		t.Fatalf("base value: %v", v)
	}
	prof, err := p.ResolveProfile("rich", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := TableEffect(p, "lumber", 1, "woodPerHour", "human", &prof); v != 50 {
		t.Fatalf("profile override: %v", v)
	}
	if _, ok := TableEffect(p, "lumber", 2, "woodPerHour", "human", nil); ok {
		t.Fatal("level beyond the table must have no value")
	}
}

func TestRatesFromTables(t *testing.T) {
	p := tablePack(t)
	b := map[string]int{"lumber": 1, "hall": 2, "barracks": 1}
	base := ComputeRates(EconInput{Pack: p, Buildings: b, Race: "human", Population: 0})
	// housing: table 500 + legacy hall effect 100/level x 2
	if base.Housing != 700 {
		t.Fatalf("housing %d", base.Housing)
	}
	// capacity: baseStorage 1000 + hall 25000 + barracks (human) 10000
	if base.Storage["wood"] != 1000+25000+10000 {
		t.Fatalf("storage %d", base.Storage["wood"])
	}
	// wood: table 30 (human), legacy produce 0 (nobody employed), no mods in this input
	if math.Abs(base.Gross["wood"]-30) > 1e-9 {
		t.Fatalf("table production missing: %v", base.Gross["wood"])
	}
	prof, _ := p.ResolveProfile("rich", nil)
	rich := ComputeRates(EconInput{Pack: p, Buildings: b, Race: "human", Profile: &prof})
	// override 50 x production_multiplier 2 = 100 from the table
	if d := rich.Gross["wood"] - ComputeRates(EconInput{Pack: p, Buildings: b, Race: "human", Profile: &pack.ResolvedProfile{ID: "rich", Params: map[string]float64{}}}).Gross["wood"]; math.Abs(d-50) > 1e-9 {
		t.Fatalf("multiplier delta %v", d)
	}
}

func TestPopulationGrowthFromTables(t *testing.T) {
	p := tablePack(t)
	if g := SumTables(p, map[string]int{"hall": 2}, "human", nil).Growth; g != 7 {
		t.Fatalf("growth %v", g)
	}
}

// R-ARMY-UPKEEP-2: 0.5% of unpaid units desert per hour (source default).
func TestDesertionHalfPercentPerHour(t *testing.T) {
	l := testPack(t)
	p := *l.Pack
	p.Params.UpkeepResource = "gold"
	p.Params.DesertionPctPerHour = 0.5
	p.Params.WorldSpeed = 1
	p.Params.EconomyTickSeconds = 3600
	l2, err := pack.Build(&p)
	if err != nil {
		t.Fatal(err)
	}
	st := &TickState{Population: 10, Garrison: map[string]int64{"spear": 1000}, Stock: map[string]int64{"food": 1, "gold": 0}, Carry: map[string]float64{}}
	r := Rates{Housing: 10, Net: map[string]float64{"food": 1, "gold": -1}}
	TickStep(st, r, l2, NewMods())
	if st.Garrison["spear"] != 995 {
		t.Fatalf("0.5%% of 1000 per hour: %d", st.Garrison["spear"])
	}
	// Paid units (upkeep stock > 0) do not desert.
	st2 := &TickState{Population: 10, Garrison: map[string]int64{"spear": 1000}, Stock: map[string]int64{"food": 1, "gold": 5000}, Carry: map[string]float64{}}
	TickStep(st2, r, l2, NewMods())
	if st2.Garrison["spear"] != 1000 {
		t.Fatalf("paid units deserted: %d", st2.Garrison["spear"])
	}
}
