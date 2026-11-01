package pack

import (
	"fmt"
	"regexp"
	"sort"
)

const (
	maxBuildingLevel = 9
	maxUnitLevel     = 4
)

// Parameter names inside level tables and unit tables are camelCase.
var reParamName = regexp.MustCompile(`^[a-z][A-Za-z0-9]{0,47}$`)

func (v *validator) collectMyLandsIDs() {
	for _, x := range v.p.Crystals {
		v.addID("crystal", x.ID)
	}
	for _, x := range v.p.ParamDefs {
		v.addID("paramDef", x.ID)
	}
	for _, x := range v.p.ServerProfiles {
		v.addID("serverProfile", x.ID)
	}
	for _, x := range v.p.UnitDefs {
		v.addID("unitDef", x.ID)
	}
	for _, x := range v.p.Services {
		v.addID("service", x.ID)
	}
}

func sortedKeys[T any](m map[string]T) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func (v *validator) kindAmounts(where string, a Amounts, kind string) {
	for _, k := range sortedKeys(a) {
		if !v.is(kind, k) {
			v.errf("%s: unknown %s %q", where, kind, k)
		}
		if a[k] < 0 {
			v.errf("%s: %q is negative", where, k)
		}
	}
}

func (v *validator) nums(where string, m map[string]Num) {
	for _, k := range sortedKeys(m) {
		if !reParamName.MatchString(k) {
			v.errf("%s: parameter name %q invalid", where, k)
		}
		for _, r := range sortedKeys(m[k].Races) {
			if !v.is("race", r) {
				v.errf("%s.%s: unknown race %q", where, k, r)
			}
		}
	}
}

func (v *validator) cityLayout() {
	l := v.p.Params.CityLayout
	if l == nil {
		return
	}
	if l.StartCells < 1 || l.CapitalCells < l.StartCells || l.ExtraCities < 0 || len(l.MaxByLevel) == 0 {
		v.errf("params.cityLayout: need startCells >= 1, capitalCells >= startCells, extraCities >= 0 and maxByLevel")
		return
	}
	for i, m := range l.MaxByLevel {
		if m < l.StartCells || (i > 0 && m < l.MaxByLevel[i-1]) {
			v.errf("params.cityLayout.maxByLevel[%d]=%d must be ascending and >= startCells", i, m)
		}
	}
}

func (v *validator) myLands() {
	p := v.p
	v.cityLayout()
	for i, c := range p.Crystals {
		if c.Tier < 1 {
			v.errf("crystals[%d](%s): tier must be >= 1", i, c.ID)
		}
	}
	seenTable := map[string]bool{}
	for i, t := range p.BuildingTables {
		w := fmt.Sprintf("buildingTables[%d](%s)", i, t.Building)
		if !v.is("building", t.Building) {
			v.errf("%s: unknown building %q", w, t.Building)
		}
		if seenTable[t.Building] {
			v.errf("%s: duplicate table for building", w)
		}
		seenTable[t.Building] = true
		if len(t.Levels) > maxBuildingLevel {
			v.errf("%s: at most %d levels", w, maxBuildingLevel)
		}
		for j, l := range t.Levels {
			lw := fmt.Sprintf("%s.levels[%d]", w, j)
			if l.Level != j+1 {
				v.errf("%s: levels must be consecutive from 1 (got %d)", lw, l.Level)
			}
			v.kindAmounts(lw+".cost", l.Cost, "resource")
			v.kindAmounts(lw+".crystals", l.Crystals, "crystal")
			if l.TimeSeconds < 0 {
				v.errf("%s: timeSeconds is negative", lw)
			}
			v.nums(lw+".effects", l.Effects)
			for _, prof := range sortedKeys(l.Overrides) {
				if !v.is("serverProfile", prof) {
					v.errf("%s.overrides: unknown server profile %q", lw, prof)
				}
				v.nums(lw+".overrides."+prof, l.Overrides[prof])
			}
		}
	}
	defs := map[string]ParamDef{}
	for i, d := range p.ParamDefs {
		w := fmt.Sprintf("paramDefs[%d](%s)", i, d.ID)
		if d.Min > d.Max {
			v.errf("%s: min %v > max %v", w, d.Min, d.Max)
		}
		if d.Default < d.Min || d.Default > d.Max {
			v.errf("%s: default %v out of range [%v, %v]", w, d.Default, d.Min, d.Max)
		}
		defs[d.ID] = d
	}
	defaults := 0
	for i, s := range p.ServerProfiles {
		w := fmt.Sprintf("serverProfiles[%d](%s)", i, s.ID)
		if len(s.Name) > 0 {
			v.text(w+".name", s.Name, false)
		}
		if len(s.Desc) > 0 {
			v.text(w+".desc", s.Desc, false)
		}
		if s.Default {
			defaults++
			if !s.Proposed {
				v.errf("%s: default profile must be proposed", w)
			}
		}
		for _, k := range sortedKeys(s.Multipliers) {
			d, ok := defs[k]
			if !ok {
				v.errf("%s: unknown param %q", w, k)
				continue
			}
			if x := s.Multipliers[k]; x < d.Min || x > d.Max {
				v.errf("%s.%s=%v out of range [%v, %v]", w, k, x, d.Min, d.Max)
			}
		}
	}
	if defaults > 1 {
		v.errf("serverProfiles: at most one default profile, got %d", defaults)
	}
	for i, u := range p.UnitDefs {
		w := fmt.Sprintf("unitDefs[%d](%s)", i, u.ID)
		if !reEntityID.MatchString(u.Category) {
			v.errf("%s: category %q invalid", w, u.Category)
		}
		if u.Level < 1 || u.Level > maxUnitLevel {
			v.errf("%s: level %d out of range [1, %d]", w, u.Level, maxUnitLevel)
		}
		v.nums(w+".params", u.Params)
		for _, k := range sortedKeys(u.Names) {
			if k != NeutralNames && !v.is("race", k) {
				v.errf("%s.names: unknown race %q", w, k)
			}
			v.text(w+".names."+k, u.Names[k], false)
		}
		v.kindAmounts(w+".cost", u.Cost, "resource")
		for _, prof := range sortedKeys(u.Overrides) {
			ow := w + ".overrides." + prof
			if !v.is("serverProfile", prof) {
				v.errf("%s: unknown server profile %q", ow, prof)
			}
			v.nums(ow+".params", u.Overrides[prof].Params)
			v.kindAmounts(ow+".cost", u.Overrides[prof].Cost, "resource")
		}
	}
	for i, s := range p.Services {
		w := fmt.Sprintf("services[%d](%s)", i, s.ID)
		if !reEntityID.MatchString(s.Group) {
			v.errf("%s: group %q invalid", w, s.Group)
		}
		if len(s.Levels) == 0 {
			v.errf("%s: levels must not be empty", w)
		}
		if len(s.Name) > 0 {
			v.text(w+".name", s.Name, false)
		}
		v.nums(w+".params", s.Params)
		if s.Renewal != "" && s.Renewal != RenewalUpgradeReplaces {
			v.errf("%s: renewal %q invalid", w, s.Renewal)
		}
		var prev int64
		for j, l := range s.Levels {
			lw := fmt.Sprintf("%s.levels[%d]", w, j)
			if l.PriceUnknown && l.Price != 0 {
				v.errf("%s: priceUnknown excludes price", lw)
			} else if !l.PriceUnknown && l.Price < 1 {
				v.errf("%s: price must be >= 1", lw)
			}
			if l.DurationUnknown && l.DurationSeconds != 0 {
				v.errf("%s: durationUnknown excludes durationSeconds", lw)
			} else if !l.DurationUnknown && l.DurationSeconds < 1 {
				v.errf("%s: durationSeconds must be >= 1", lw)
			}
			if l.CooldownMonths < 0 {
				v.errf("%s: cooldownMonths is negative", lw)
			}
			if s.Renewal != "" && l.DurationSeconds > 0 {
				if l.DurationSeconds <= prev {
					v.errf("%s: levels must run from weaker to stronger (durationSeconds must grow)", lw)
				}
				prev = l.DurationSeconds
			}
		}
		if !v.is("resource", p.Params.PremiumResource) {
			v.errf("%s: params.premiumResource must name a resource", w)
		}
	}
	if g := p.Geometry; g != nil {
		for _, f := range []struct {
			n string
			x int
		}{{"empireWidth", g.EmpireWidth}, {"empireHeight", g.EmpireHeight}, {"kingdomWidth", g.KingdomWidth}, {"kingdomHeight", g.KingdomHeight}} {
			if f.x < 1 {
				v.errf("geometry.%s must be a positive integer", f.n)
			}
		}
	}
}
