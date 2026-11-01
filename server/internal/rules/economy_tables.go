package rules

import (
	"sort"

	"realmforge/server/internal/pack"
)

// Level-table economy (ML-M05). Values come from pack.BuildingTables; race values use
// pack.Num and server profile overrides replace a named effect for that profile id.
// Effect names are looked up by pack convention: "<resourceId>PerHour", "resourceCapacity",
// "housingCapacity", "populationGrowth". Units of populationGrowth are UNKNOWN in the source.

// Effect names read from tables.
const (
	EffectResourceCapacity = "resourceCapacity"
	EffectHousingCapacity  = "housingCapacity"
	EffectPopulationGrowth = "populationGrowth"
	ParamProduction        = "production_multiplier"
	ParamUpkeep            = "upkeep_multiplier"
)

// TableEffect returns the named effect of a building at a level for a race and profile.
func TableEffect(p *pack.Loaded, building string, level int, name, race string, prof *pack.ResolvedProfile) (float64, bool) {
	for i := range p.BuildingTables {
		t := &p.BuildingTables[i]
		if t.Building != building {
			continue
		}
		if level < 1 || level > len(t.Levels) {
			return 0, false
		}
		row := &t.Levels[level-1]
		if prof != nil {
			if n, ok := row.Overrides[prof.ID][name]; ok {
				return n.For(race), true
			}
		}
		if n, ok := row.Effects[name]; ok {
			return n.For(race), true
		}
		return 0, false
	}
	return 0, false
}

// TableTotals sums the table effects of all buildings of one city.
type TableTotals struct {
	Production map[string]float64 // per hour, profile production multiplier applied
	Capacity   float64            // resourceCapacity per capped resource (warehouse, city hall)
	Housing    float64
	Growth     float64
}

// SumTables adds table effects over buildings (building -> level).
func SumTables(p *pack.Loaded, buildings map[string]int, race string, prof *pack.ResolvedProfile) TableTotals {
	t := TableTotals{Production: map[string]float64{}}
	ids := make([]string, 0, len(buildings))
	for id := range buildings {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	mult := 1.0
	if prof != nil {
		if m, ok := prof.Params[ParamProduction]; ok {
			mult = m
		}
	}
	for _, id := range ids {
		lvl := buildings[id]
		if lvl <= 0 {
			continue
		}
		for _, res := range p.Resources {
			if v, ok := TableEffect(p, id, lvl, res.ID+"PerHour", race, prof); ok {
				t.Production[res.ID] += v * mult
			}
		}
		if v, ok := TableEffect(p, id, lvl, EffectResourceCapacity, race, prof); ok {
			t.Capacity += v
		}
		if v, ok := TableEffect(p, id, lvl, EffectHousingCapacity, race, prof); ok {
			t.Housing += v
		}
		if v, ok := TableEffect(p, id, lvl, EffectPopulationGrowth, race, prof); ok {
			t.Growth += v
		}
	}
	return t
}
