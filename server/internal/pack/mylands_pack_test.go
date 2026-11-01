package pack

import (
	"strings"
	"testing"
)

const myLandsPackDir = "../../../packs/mylands"

func loadMyLands(t *testing.T) *Loaded {
	t.Helper()
	l, err := LoadFile(myLandsPackDir)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func tableOf(t *testing.T, l *Loaded, id string) BuildingTable {
	t.Helper()
	for _, tb := range l.BuildingTables {
		if tb.Building == id {
			return tb
		}
	}
	t.Fatalf("no building table %q", id)
	return BuildingTable{}
}

// ML-D02: eight races, two factions, no invented percentages.
func TestMyLandsPackRaces(t *testing.T) {
	l := loadMyLands(t)
	if len(l.Races) != 8 {
		t.Fatalf("races = %d, want 8", len(l.Races))
	}
	want := map[string]string{"knights": "Рыцари", "elves": "Эльфы", "wizards": "Волшебники", "elementals": "Элементали",
		"demons": "Демоны", "dark_elves": "Тёмные Эльфы", "undead": "Нежить", "chimeras": "Химеры"}
	light := 0
	for _, r := range l.Races {
		if want[r.ID] != r.Name["ru"] {
			t.Errorf("race %s name %q", r.ID, r.Name["ru"])
		}
		if len(r.Modifiers) != 0 {
			t.Errorf("race %s has invented modifiers", r.ID)
		}
		if r.Faction == "light" {
			light++
		}
	}
	if light != 4 {
		t.Errorf("light faction races = %d, want 4", light)
	}
}

func TestMyLandsPackResourcesAndCrystals(t *testing.T) {
	l := loadMyLands(t)
	for _, id := range []string{"gold", "wood", "stone", "iron", "grain", "pearl"} {
		if l.resources[id] == nil {
			t.Errorf("missing resource %s", id)
		}
	}
	if l.Params.PremiumResource != "pearl" || l.resources["pearl"].Kind != "premium" {
		t.Error("pearl must be the premium resource")
	}
	if len(l.Crystals) != 10 {
		t.Fatalf("crystals = %d, want 10", len(l.Crystals))
	}
	tiers := map[int]int{}
	for _, c := range l.Crystals {
		tiers[c.Tier]++
	}
	if tiers[1] != 5 || tiers[2] != 5 {
		t.Errorf("crystal tiers %v", tiers)
	}
	if g := l.Geometry; g == nil || g.EmpireWidth != 90 || g.EmpireHeight != 30 || g.FieldsPerKingdom() != 81 {
		t.Errorf("geometry %+v", g)
	}
}

// ML-D03: every table-backed building has the documented level count.
func TestMyLandsPackBuildings(t *testing.T) {
	l := loadMyLands(t)
	levels := map[string]int{"tavern": 3, "terraformer": 2, "recovery": 2, "colony": 1, "city_hall": 5}
	std9 := []string{"houses", "warehouse", "market", "shop", "windmill", "samwill", "mine_gold", "mine_iron", "mine_stone",
		"alchemist", "prison", "artefact_store", "temple", "resurrector", "exped", "fortification", "tower", "magic_tower",
		"watch_tower", "antiwatch_tower", "spy_guild", "antispy_guild", "merch_guild", "clan_warehouse", "clan_reservation",
		"scavenger_build", "melee_build", "cavalry_build", "flyer_build", "range_build", "healer_build", "miner_build", "mage_build"}
	for _, id := range std9 {
		levels[id] = 9
	}
	for _, id := range []string{"warehouse_kz", "houses_kz", "fortification_kz", "tower_kz", "magic_tower_kz", "resurrector_kz",
		"antiwatch_tower_kz", "antispy_guild_kz", "merch_guild_kz", "scavenger_build_kz", "melee_build_kz", "cavalry_build_kz",
		"flyer_build_kz", "range_build_kz", "healer_build_kz", "miner_build_kz", "mage_build_kz"} {
		levels[id] = 7
	}
	for id, n := range levels {
		if l.buildings[id] == nil {
			t.Errorf("building %s missing", id)
			continue
		}
		if got := len(tableOf(t, l, id).Levels); got != n {
			t.Errorf("%s levels = %d, want %d", id, got, n)
		}
		if l.buildings[id].MaxLevel != n {
			t.Errorf("%s maxLevel = %d, want %d", id, l.buildings[id].MaxLevel, n)
		}
	}
	if len(l.Buildings) != len(levels) || len(l.BuildingTables) != len(levels) {
		t.Errorf("buildings %d tables %d, want %d", len(l.Buildings), len(l.BuildingTables), len(levels))
	}
}

func TestMyLandsPackBuildingValues(t *testing.T) {
	l := loadMyLands(t)
	fort := tableOf(t, l, "fortification").Levels[8]
	if fort.Effects["kills"].Base != 1000 {
		t.Errorf("fortification L9 kills = %v", fort.Effects["kills"])
	}
	if fort.Cost["gold"] != 5000000 || fort.Cost["iron"] != 2500000 {
		t.Errorf("fortification L9 cost %v", fort.Cost)
	}
	// INFERRED crystals: one unit per icon (Alchemy page): emerald, white, amethyst x2.
	if fort.Crystals["emerald"] != 1 || fort.Crystals["white"] != 1 || fort.Crystals["amethyst"] != 2 {
		t.Errorf("fortification L9 crystals %v", fort.Crystals)
	}
	if got := tableOf(t, l, "temple").Levels[8].Effects["mercyLimit"].Base; got != 750 {
		t.Errorf("temple L9 mercyLimit = %v", got)
	}
	if got := tableOf(t, l, "exped").Levels[8].Effects["explorerCount"].Base; got != 15 {
		t.Errorf("exped L9 explorerCount = %v", got)
	}
	// blank source cell stays absent (Houses L1 needs only wood).
	h1 := tableOf(t, l, "houses").Levels[0]
	if _, ok := h1.Cost["gold"]; ok || h1.Cost["wood"] != 500 {
		t.Errorf("houses L1 cost %v", h1.Cost)
	}
	// race override, plus per-profile override with race on top (Windmill L2).
	w := tableOf(t, l, "windmill").Levels[1]
	if e := w.Effects["grainPerHour"]; e.Base != 12 || e.For("knights") != 18 || e.For("wizards") != 18 || e.For("elementals") != 15 {
		t.Errorf("windmill L2 grainPerHour %+v", e)
	}
	if e := w.Overrides["economic"]["grainPerHour"]; e.Base != 14 || e.For("knights") != 21 {
		t.Errorf("windmill L2 economic %+v", e)
	}
	if w.Overrides["tactical"]["grainPerHour"].Base != 14 {
		t.Error("tactical profile must mirror economic")
	}
	if e := tableOf(t, l, "warehouse").Levels[0].Effects["resourceCapacity"]; e.Base != 5000 || e.For("knights") != 10000 {
		t.Errorf("warehouse L1 capacity %+v", e)
	}
}

// ML-D04: units by category, group and level; names from the roster.
func TestMyLandsPackUnits(t *testing.T) {
	l := loadMyLands(t)
	if len(l.UnitDefs) != 88 {
		t.Fatalf("unitDefs = %d, want 88 (8 categories x (4 player + 3 monster + 4 castle))", len(l.UnitDefs))
	}
	byID := map[string]UnitDef{}
	names := 0
	for _, u := range l.UnitDefs {
		byID[u.ID] = u
		names += len(u.Names)
	}
	if names != 280 {
		t.Errorf("roster names = %d, want 280", names)
	}
	if n := byID["scavenger_1"].Names["knights"]["ru"]; n != "Прислужник" {
		t.Errorf("scavenger_1 knights name %q", n)
	}
	c := byID["scavenger_1"].Params["capacity"]
	if c.Base != 50 || c.For("knights") != 100 {
		t.Errorf("carrier capacity %+v", c)
	}
	if r := byID["healer_1"].Params["resurrectCount"]; r.Base != 1 || r.For("elves") != 2 {
		t.Errorf("healer resurrect %+v", r)
	}
	if a := byID["miner_1"].Params["attackMin"]; a.Base != 3 || a.For("demons") != 5 {
		t.Errorf("mercenary attackMin %+v", a)
	}
	if m := byID["mage_1"].Params["towerReduction"]; m.Base != 10 || m.For("dark_elves") != 20 {
		t.Errorf("mage tower reduction %+v", m)
	}
	if o := byID["range_1"].Overrides["economic"].Params["attackMin"]; o.Base != 15 {
		t.Errorf("archer economic attackMin %+v", o)
	}
	if _, ok := byID["melee_1"].Params["capacity"]; ok {
		t.Error("blank cells must stay absent")
	}
	if _, ok := byID["melee_1"].Cost["iron"]; ok {
		t.Error("blank iron cost must stay absent")
	}
	if len(byID["scavenger_castle_4"].Params) == 0 || len(byID["scavenger_monster_3"].Params) == 0 {
		t.Error("monster and castle groups missing")
	}
	if _, ok := byID["scavenger_monster_4"]; ok {
		t.Error("monsters have levels 1..3 only")
	}
}

func TestMyLandsPackNotes(t *testing.T) {
	l := loadMyLands(t)
	found := false
	for k := range l.Strings {
		if strings.HasPrefix(k, "unknown.") {
			found = true
		}
	}
	if !found {
		t.Error("UNKNOWN items must be listed in strings under unknown.*")
	}
}

// ML-D08: eight profiles; no invented multipliers; default combat is PROPOSED (TARGET_PROFILE A-01).
func TestMyLandsPackProfiles(t *testing.T) {
	l := loadMyLands(t)
	want := []string{"miner", "combat", "economic", "tactical", "heroic", "strategic", "operational", "faction_war"}
	if len(l.ServerProfiles) != len(want) {
		t.Fatalf("profiles = %d, want %d", len(l.ServerProfiles), len(want))
	}
	for i, id := range want {
		s := l.ServerProfiles[i]
		if s.ID != id {
			t.Errorf("profile[%d] = %s, want %s", i, s.ID, id)
		}
		if len(s.Multipliers) != 0 {
			t.Errorf("profile %s has invented multipliers", id)
		}
		if s.Name["ru"] == "" || s.Desc["ru"] == "" {
			t.Errorf("profile %s needs name and desc", id)
		}
		if s.Default != (id == "combat") {
			t.Errorf("profile %s default = %v", id, s.Default)
		}
	}
	if !l.ServerProfiles[1].Proposed {
		t.Error("default profile combat must be PROPOSED")
	}
	if len(l.ParamDefs) == 0 {
		t.Fatal("paramDefs missing")
	}
	for _, d := range l.ParamDefs {
		if !d.Proposed || d.Default != 1 {
			t.Errorf("paramDef %s must be PROPOSED with default 1", d.ID)
		}
	}
	if _, err := l.ResolveProfile("", nil); err != nil {
		t.Fatal(err)
	}
}

// ML-D09 / CS-T23: paid service catalog; prices UNKNOWN; vacation levels and renewal rule.
func TestMyLandsPackServices(t *testing.T) {
	l := loadMyLands(t)
	groups := map[string]int{}
	byID := map[string]Service{}
	for _, s := range l.Services {
		groups[s.Group]++
		byID[s.ID] = s
		if !s.Proposed || s.Renewal != RenewalUpgradeReplaces {
			t.Errorf("service %s must be proposed with renewal rule", s.ID)
		}
		for _, lv := range s.Levels {
			if lv.Price != 0 || !lv.PriceUnknown {
				t.Errorf("service %s: price must stay UNKNOWN", s.ID)
			}
		}
	}
	for _, g := range []string{"premium", "global", "local", "hero_automation", "vacation"} {
		if groups[g] == 0 {
			t.Errorf("no services in group %s", g)
		}
	}
	v := byID["vacation"]
	days, months := []int64{15, 30, 60}, []int{1, 6, 12}
	if len(v.Levels) != 3 {
		t.Fatalf("vacation levels = %d", len(v.Levels))
	}
	for i, lv := range v.Levels {
		if lv.DurationSeconds != days[i]*86400 || lv.CooldownMonths != months[i] {
			t.Errorf("vacation level %d = %+v", i, lv)
		}
	}
	if v.Params["attackLockoutSeconds"].Base != 21600 {
		t.Error("vacation is forbidden 6 hours before an attack")
	}
	if byID["premium"].Levels[0].DurationSeconds != 0 || !byID["premium"].Levels[0].DurationUnknown {
		t.Error("premium duration is a CONFLICT and must stay unknown")
	}
}

func TestServiceValidation(t *testing.T) {
	base := func() *Pack {
		p := withMyLands(t)
		p.Services = []Service{{ID: "vac", Group: "vacation", Renewal: RenewalUpgradeReplaces, Levels: []ServiceLevel{
			{PriceUnknown: true, DurationSeconds: 100}, {PriceUnknown: true, DurationSeconds: 200}}}}
		return p
	}
	if _, err := Build(base()); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(p *Pack)
		want   string
	}{
		{"unknown price with value", func(p *Pack) { p.Services[0].Levels[0].Price = 5 }, "priceUnknown"},
		{"duration unknown with value", func(p *Pack) { p.Services[0].Levels[0].DurationUnknown = true }, "durationUnknown"},
		{"missing price", func(p *Pack) { p.Services[0].Levels[0].PriceUnknown = false }, "price must be >= 1"},
		{"bad renewal", func(p *Pack) { p.Services[0].Renewal = "forever" }, "renewal"},
		{"levels not stronger", func(p *Pack) { p.Services[0].Levels[1].DurationSeconds = 100 }, "stronger"},
		{"negative cooldown", func(p *Pack) { p.Services[0].Levels[0].CooldownMonths = -1 }, "cooldownMonths"},
		{"bad param name", func(p *Pack) { p.Services[0].Params = map[string]Num{"Bad": {Base: 1}} }, "parameter name"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := base()
			c.mutate(p)
			expectErr(t, p, c.want)
		})
	}
}
