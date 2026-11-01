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
	// PROPOSED crystals: emerald, white, amethyst x2 at tier B = 10 each per icon.
	if fort.Crystals["emerald"] != 10 || fort.Crystals["white"] != 10 || fort.Crystals["amethyst"] != 20 {
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
