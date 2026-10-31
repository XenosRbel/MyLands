// Package rules contains the pure, deterministic game formulas: economy,
// construction, research, training, travel, combat, scouting, loot, heroes
// and map generation. Nothing here touches the database or the clock; every
// function takes explicit inputs so it can be unit tested exhaustively.
package rules

import (
	"math"
	"sort"

	"realmforge/server/internal/pack"
)

// Mods accumulates modifiers. Order of application is fixed and documented in
// GAME_RULES.md §2: value = (base + Σflat) × (1 + Σpct/100), clamped by caller.
type Mods struct {
	pct  map[string]float64
	flat map[string]float64
}

func NewMods() *Mods { return &Mods{pct: map[string]float64{}, flat: map[string]float64{}} }

// Add adds a modifier scaled by times (e.g. building/tech level).
func (m *Mods) Add(mod pack.Modifier, times float64) {
	if times == 0 {
		return
	}
	if mod.Op == "flat" {
		m.flat[mod.Target] += mod.Value * times
	} else {
		m.pct[mod.Target] += mod.Value * times
	}
}

func (m *Mods) AddAll(mods []pack.Modifier, times float64) {
	for _, x := range mods {
		m.Add(x, times)
	}
}

// Merge adds all modifiers of o into m.
func (m *Mods) Merge(o *Mods) {
	if o == nil {
		return
	}
	for k, v := range o.pct {
		m.pct[k] += v
	}
	for k, v := range o.flat {
		m.flat[k] += v
	}
}

// Clone returns an independent copy.
func (m *Mods) Clone() *Mods {
	c := NewMods()
	c.Merge(m)
	return c
}

// Pct returns the summed percentage across the given targets.
func (m *Mods) Pct(targets ...string) float64 {
	s := 0.0
	for _, t := range targets {
		s += m.pct[t]
	}
	return s
}

// Flat returns the summed flat bonus across the given targets.
func (m *Mods) Flat(targets ...string) float64 {
	s := 0.0
	for _, t := range targets {
		s += m.flat[t]
	}
	return s
}

// Apply returns (base + flat) × max(0, 1 + pct/100).
func (m *Mods) Apply(base float64, targets ...string) float64 {
	return math.Max(0, (base+m.Flat(targets...))*math.Max(0, 1+m.Pct(targets...)/100))
}

// TimeFactor converts a (usually negative) time pct into a duration multiplier.
// Durations can be reduced to at most MinTimeFactor of their base value.
func (m *Mods) TimeFactor(target string) float64 {
	return math.Max(MinTimeFactor, 1+m.Pct(target)/100)
}

// MinTimeFactor is the floor on any duration multiplier (GAME_RULES R-TIME-3).
const MinTimeFactor = 0.2

// Snapshot returns a sorted, serialisable view (for debugging / API).
func (m *Mods) Snapshot() map[string]float64 {
	out := map[string]float64{}
	keys := make([]string, 0, len(m.pct))
	for k := range m.pct {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if m.pct[k] != 0 {
			out[k+"%"] = m.pct[k]
		}
	}
	for k, v := range m.flat {
		if v != 0 {
			out[k] = v
		}
	}
	return out
}
