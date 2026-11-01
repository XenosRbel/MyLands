package pack

import (
	"encoding/json"
	"strings"
	"testing"
)

// withMyLands extends the minimal pack with a valid My Lands section.
func withMyLands(t *testing.T) *Pack {
	t.Helper()
	p := loadMinimal(t)
	p.Races = append(p.Races, Race{ID: "knights", Name: LocText{"en": "Knights"}})
	p.Crystals = []CrystalType{{ID: "red", Tier: 1}, {ID: "green", Tier: 2}}
	p.BuildingTables = []BuildingTable{{
		Building: "hall",
		Levels: []BuildingLevel{
			{Level: 1, Cost: Amounts{"wood": 500}, TimeSeconds: 60, Effects: map[string]Num{"kills": {Base: 30}}},
			{Level: 2, Cost: Amounts{"wood": 1000, "gold": 5}, Crystals: Amounts{"red": 2}, Effects: map[string]Num{"capacity": {Base: 50, Races: map[string]float64{"knights": 100}}}},
		},
	}}
	p.ParamDefs = []ParamDef{{ID: "production", Min: 0.1, Max: 10, Default: 1}}
	p.ServerProfiles = []ServerProfile{{ID: "economic", Multipliers: map[string]float64{"production": 2}}}
	p.UnitDefs = []UnitDef{{
		ID: "carrier1", Category: "carriers", Level: 1,
		Params:    map[string]Num{"health": {Base: 10}, "upkeep": {Base: 1}},
		Cost:      Amounts{"gold": 5},
		Overrides: map[string]UnitOverride{"economic": {Params: map[string]Num{"upkeep": {Base: 0.5}}}},
	}}
	p.Services = []Service{{ID: "builders", Group: "speed", Proposed: true,
		Levels: []ServiceLevel{{Price: 10, DurationSeconds: 86400}}}}
	p.Params.PremiumResource = "gem"
	p.Geometry = &MapGeometry{EmpireWidth: 90, EmpireHeight: 30, KingdomWidth: 9, KingdomHeight: 9, Proposed: true}
	return p
}

func TestMyLandsValid(t *testing.T) {
	l, err := Build(withMyLands(t))
	if err != nil {
		t.Fatal(err)
	}
	if l.Geometry.FieldsPerKingdom() != 81 {
		t.Fatal("fields per kingdom")
	}
}

func TestMyLandsErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(p *Pack)
		want   string
	}{
		{"table unknown building", func(p *Pack) { p.BuildingTables[0].Building = "nope" }, `unknown building "nope"`},
		{"level gap", func(p *Pack) { p.BuildingTables[0].Levels[1].Level = 3 }, "consecutive"},
		{"level >9", func(p *Pack) {
			p.BuildingTables[0].Levels = make([]BuildingLevel, 10)
			for i := range p.BuildingTables[0].Levels {
				p.BuildingTables[0].Levels[i].Level = i + 1
			}
		}, "at most 9"},
		{"cost unknown resource", func(p *Pack) { p.BuildingTables[0].Levels[0].Cost["mana"] = 1 }, `unknown resource "mana"`},
		{"negative cost", func(p *Pack) { p.BuildingTables[0].Levels[0].Cost["wood"] = -1 }, "negative"},
		{"unknown crystal", func(p *Pack) { p.BuildingTables[0].Levels[1].Crystals["blue"] = 1 }, `unknown crystal "blue"`},
		{"crystal tier", func(p *Pack) { p.Crystals[0].Tier = 0 }, "tier"},
		{"override unknown race", func(p *Pack) {
			p.BuildingTables[0].Levels[1].Effects["capacity"] = Num{Base: 50, Races: map[string]float64{"orcs": 1}}
		}, `unknown race "orcs"`},
		{"param range", func(p *Pack) { p.ServerProfiles[0].Multipliers["production"] = 99 }, "out of range"},
		{"unknown multiplier", func(p *Pack) { p.ServerProfiles[0].Multipliers["luck"] = 1 }, `unknown param "luck"`},
		{"paramdef bounds", func(p *Pack) { p.ParamDefs[0].Min = 20 }, "min"},
		{"unit level", func(p *Pack) { p.UnitDefs[0].Level = 5 }, "level"},
		{"unit override profile", func(p *Pack) {
			p.UnitDefs[0].Overrides = map[string]UnitOverride{"ghost": {}}
		}, `unknown server profile "ghost"`},
		{"service no levels", func(p *Pack) { p.Services[0].Levels = nil }, "levels"},
		{"service price", func(p *Pack) { p.Services[0].Levels[0].Price = 0 }, "price"},
		{"service needs premium", func(p *Pack) { p.Params.PremiumResource = "" }, "premiumResource"},
		{"geometry", func(p *Pack) { p.Geometry.KingdomWidth = 0 }, "kingdomWidth"},
		{"dup id", func(p *Pack) { p.Services[0].ID = "red" }, "duplicate id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := withMyLands(t)
			c.mutate(p)
			expectErr(t, p, c.want)
		})
	}
}

func TestNumJSON(t *testing.T) {
	var m map[string]Num
	if err := json.Unmarshal([]byte(`{"a":5,"b":{"base":50,"races":{"knights":100}}}`), &m); err != nil {
		t.Fatal(err)
	}
	if m["a"].Base != 5 || m["b"].For("knights") != 100 || m["b"].For("other") != 50 {
		t.Fatalf("bad decode %+v", m)
	}
	out, _ := json.Marshal(m)
	if !strings.Contains(string(out), `"a":5`) {
		t.Fatalf("plain number must stay a number: %s", out)
	}
	if err := json.Unmarshal([]byte(`{"b":{"base":1,"evil":2}}`), &m); err == nil {
		t.Fatal("unknown field in Num must be rejected")
	}
}

func TestMyLandsUnknownFieldRejected(t *testing.T) {
	_, err := Parse([]byte(`{"formatVersion":1,"geometry":{"empireWidth":1,"bogus":2}}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("got %v", err)
	}
}

// Removing a crystal id used by an older version needs a migrations entry.
func TestMyLandsMigrationTargets(t *testing.T) {
	p := withMyLands(t)
	p.Migrations = []Migration{{FromVersion: "0.9.0", Renames: map[string]string{"old_service": "builders"}}}
	if errs := Validate(p); errs != nil {
		t.Fatalf("rename to new id must validate: %v", errs)
	}
	p.Migrations[0].Renames["old_crystal"] = "gone"
	expectErr(t, p, "targets unknown id")
}

func TestSamplePackTables(t *testing.T) {
	l, err := LoadFile("../../../packs/mylands-sample")
	if err != nil {
		t.Fatal(err)
	}
	var fort, hall BuildingTable
	for _, tb := range l.BuildingTables {
		switch tb.Building {
		case "fortification":
			fort = tb
		case "hall":
			hall = tb
		}
	}
	if len(fort.Levels) != 9 || fort.Levels[8].Effects["kills"].Base != 1000 || fort.Levels[8].Cost["gold"] != 5000000 {
		t.Fatalf("fortification level 9 mismatch: %+v", fort.Levels)
	}
	if _, ok := fort.Levels[0].Cost["pearl"]; ok {
		t.Fatal("zero source cost must stay absent")
	}
	if c := hall.Levels[0].Effects["carrierCapacity"]; c.For("human") != 50 || c.For("knights") != 100 {
		t.Fatalf("race override: %+v", c)
	}
}
