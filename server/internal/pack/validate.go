package pack

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Errors aggregates all validation problems found in a pack.
type Errors []string

func (e Errors) Error() string {
	return fmt.Sprintf("pack validation failed (%d problems):\n  - %s", len(e), strings.Join(e, "\n  - "))
}

var (
	reEntityID  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	rePackID    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,47}$`)
	reVersion   = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	reLocale    = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	reColor     = regexp.MustCompile(`^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)
	reAssetPath = regexp.MustCompile(`^[a-z0-9_-]+(/[a-z0-9_-]+)*\.(png|webp|svg|jpg)$`)
)

var (
	ResourceKinds   = set("basic", "food", "rare", "premium")
	BuildingCats    = set("core", "economy", "storage", "military", "science", "special")
	EffectTypes     = set("produce", "storage", "protect", "housing", "modifier", "idle_tax")
	UnitClasses     = set("infantry", "ranged", "cavalry", "siege", "scout", "transport", "settler", "magic", "beast")
	ReqTypes        = set("building", "tech", "race")
	NeutralKinds    = set("lair", "ruin", "node")
	ItemRarities    = set("common", "rare", "epic", "legendary")
	QuestGoalTypes  = set("building_level", "tech_level", "units_trained", "missions_won", "neutrals_cleared", "join_alliance", "market_trade", "hero_hired", "cities", "territories", "scouts_sent")
	ModifierOps     = set("pct", "flat")
	KnownScreens    = set("city", "economy", "research", "army", "heroes", "map", "missions", "reports", "market", "alliance", "quests", "mail", "profile", "settings")
	KnownModes      = set("standard", "peaceful", "speed")
	requiredPalette = []string{"bg", "surface", "text", "muted", "accent", "danger", "ok"}
)

// ModifierTargetValid validates the closed modifier target grammar.
func ModifierTargetValid(target string, unitOrClass func(string) bool, resource func(string) bool) bool {
	head, tail, hasTail := strings.Cut(target, ".")
	switch head {
	case "production", "storage", "protect":
		return hasTail && (tail == "*" || resource(tail))
	case "attack", "defense", "hp":
		return hasTail && (tail == "*" || unitOrClass(tail))
	case "housing", "pop_growth", "build_time", "research_time", "train_time", "upkeep",
		"speed", "carry", "loot", "scout", "wall", "hero_xp", "market_fee":
		return !hasTail
	}
	return false
}

func set(xs ...string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

type validator struct {
	p    *Pack
	errs Errors
	ids  map[string]string // id -> kind
}

func (v *validator) errf(format string, a ...any) { v.errs = append(v.errs, fmt.Sprintf(format, a...)) }

// Validate checks schema, ranges, uniqueness and referential integrity.
// It returns nil when the pack is valid.
func Validate(p *Pack) Errors {
	v := &validator{p: p, ids: map[string]string{}}
	v.header()
	v.collectIDs()
	v.resources()
	v.races()
	v.buildings()
	v.techs()
	v.units()
	v.heroes()
	v.world()
	v.quests()
	v.start()
	v.params()
	v.theme()
	v.ui()
	v.myLands()
	v.migrations()
	v.requirementCycles()
	if len(v.errs) == 0 {
		return nil
	}
	return v.errs
}

func (v *validator) header() {
	p := v.p
	if p.FormatVersion < 1 || p.FormatVersion > FormatVersion {
		v.errf("formatVersion %d unsupported (engine supports 1..%d)", p.FormatVersion, FormatVersion)
	}
	if !rePackID.MatchString(p.ID) {
		v.errf("id %q must match %s", p.ID, rePackID)
	}
	if !reVersion.MatchString(p.Version) {
		v.errf("version %q must be semantic (MAJOR.MINOR.PATCH)", p.Version)
	}
	if len(p.Locales) == 0 {
		v.errf("locales must not be empty")
	}
	seen := map[string]bool{}
	for _, l := range p.Locales {
		if !reLocale.MatchString(l) {
			v.errf("locale %q invalid", l)
		}
		if seen[l] {
			v.errf("locale %q duplicated", l)
		}
		seen[l] = true
	}
	if !seen[p.DefaultLocale] {
		v.errf("defaultLocale %q not in locales", p.DefaultLocale)
	}
	v.text("meta.title", p.Meta.Title, true)
	v.text("meta.description", p.Meta.Description, true)
	if len(p.Modes) == 0 {
		v.errf("modes must not be empty")
	}
	for _, m := range p.Modes {
		if !KnownModes[m] {
			v.errf("mode %q unknown", m)
		}
	}
}

// text requires the default locale; when all is set every declared locale is required.
func (v *validator) text(where string, t LocText, all bool) {
	if t[v.p.DefaultLocale] == "" {
		v.errf("%s: missing text for default locale %q", where, v.p.DefaultLocale)
	}
	if all {
		for _, l := range v.p.Locales {
			if t[l] == "" {
				v.errf("%s: missing text for locale %q", where, l)
			}
		}
	}
	for l := range t {
		if !contains(v.p.Locales, l) {
			v.errf("%s: text for undeclared locale %q", where, l)
		}
	}
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func (v *validator) addID(kind, id string) {
	if !reEntityID.MatchString(id) {
		v.errf("%s id %q must match %s", kind, id, reEntityID)
		return
	}
	if prev, ok := v.ids[id]; ok {
		v.errf("duplicate id %q (%s and %s); ids are unique across all entity kinds", id, prev, kind)
		return
	}
	v.ids[id] = kind
}

func (v *validator) collectIDs() {
	p := v.p
	for _, x := range p.Resources {
		v.addID("resource", x.ID)
	}
	for _, x := range p.Races {
		v.addID("race", x.ID)
	}
	for _, x := range p.Buildings {
		v.addID("building", x.ID)
	}
	for _, x := range p.Techs {
		v.addID("tech", x.ID)
	}
	for _, x := range p.Units {
		v.addID("unit", x.ID)
	}
	for _, x := range p.Heroes.Classes {
		v.addID("heroClass", x.ID)
	}
	for _, x := range p.Heroes.Skills {
		v.addID("heroSkill", x.ID)
	}
	for _, x := range p.Heroes.Items {
		v.addID("heroItem", x.ID)
	}
	for _, x := range p.Terrains {
		v.addID("terrain", x.ID)
	}
	for _, x := range p.Neutrals {
		v.addID("neutral", x.ID)
	}
	for _, x := range p.Quests {
		v.addID("quest", x.ID)
	}
	v.collectMyLandsIDs()
}

func (v *validator) is(kind, id string) bool { return v.ids[id] == kind }

func (v *validator) isUnitOrClass(s string) bool { return v.is("unit", s) || UnitClasses[s] }
func (v *validator) isResource(s string) bool    { return v.is("resource", s) }

func (v *validator) modifiers(where string, ms []Modifier) {
	for i, m := range ms {
		v.modifier(fmt.Sprintf("%s.modifiers[%d]", where, i), m)
	}
}

func (v *validator) modifier(where string, m Modifier) {
	if !ModifierOps[m.Op] {
		v.errf("%s: op %q must be pct|flat", where, m.Op)
	}
	if !ModifierTargetValid(m.Target, v.isUnitOrClass, v.isResource) {
		v.errf("%s: invalid modifier target %q", where, m.Target)
	}
	if m.Op == "pct" && (m.Value < -95 || m.Value > 1000) {
		v.errf("%s: pct value %v out of range [-95, 1000]", where, m.Value)
	}
}

func (v *validator) amounts(where string, a Amounts, allowEmpty bool) {
	if len(a) == 0 && !allowEmpty {
		v.errf("%s: must not be empty", where)
	}
	for r, n := range a {
		if !v.is("resource", r) {
			v.errf("%s: unknown resource %q", where, r)
		}
		if n < 0 || n > 1e12 {
			v.errf("%s: amount %d for %q out of range [0, 1e12]", where, n, r)
		}
	}
}

func (v *validator) requirements(where string, rs []Requirement) {
	for i, r := range rs {
		w := fmt.Sprintf("%s.requires[%d]", where, i)
		if !ReqTypes[r.Type] {
			v.errf("%s: type %q invalid", w, r.Type)
			continue
		}
		if !v.is(r.Type, r.ID) {
			v.errf("%s: unknown %s %q", w, r.Type, r.ID)
			continue
		}
		var maxL int
		switch r.Type {
		case "building":
			maxL = v.building(r.ID).MaxLevel
		case "tech":
			maxL = v.tech(r.ID).MaxLevel
		case "race":
			continue
		}
		if r.Level < 1 || r.Level > maxL {
			v.errf("%s: level %d out of range [1, %d] for %q", w, r.Level, maxL, r.ID)
		}
	}
}

func (v *validator) building(id string) *Building {
	for i := range v.p.Buildings {
		if v.p.Buildings[i].ID == id {
			return &v.p.Buildings[i]
		}
	}
	return nil
}

func (v *validator) tech(id string) *Tech {
	for i := range v.p.Techs {
		if v.p.Techs[i].ID == id {
			return &v.p.Techs[i]
		}
	}
	return nil
}

func (v *validator) races_(where string, rs []string) {
	for _, r := range rs {
		if !v.is("race", r) {
			v.errf("%s: unknown race %q", where, r)
		}
	}
}

func (v *validator) costTime(where string, c Cost, d Duration) {
	v.amounts(where+".cost.base", c.Base, false)
	if c.Growth < 1 || c.Growth > 5 {
		v.errf("%s.cost.growth %v out of range [1, 5]", where, c.Growth)
	}
	if d.BaseSeconds < 1 || d.BaseSeconds > 30*86400 {
		v.errf("%s.time.baseSeconds %d out of range [1, 2592000]", where, d.BaseSeconds)
	}
	if d.Growth < 1 || d.Growth > 5 {
		v.errf("%s.time.growth %v out of range [1, 5]", where, d.Growth)
	}
}

func (v *validator) effects(where string, es []Effect) {
	for i, e := range es {
		w := fmt.Sprintf("%s.effects[%d]", where, i)
		if !EffectTypes[e.Type] {
			v.errf("%s: type %q invalid", w, e.Type)
			continue
		}
		switch e.Type {
		case "produce":
			if !v.is("resource", e.Resource) {
				v.errf("%s: unknown resource %q", w, e.Resource)
			}
			if e.PerLevel <= 0 {
				v.errf("%s: perLevel must be > 0", w)
			}
			if e.Workers < 0 {
				v.errf("%s: workers must be >= 0", w)
			}
		case "storage", "protect":
			if e.Resource != "*" && !v.is("resource", e.Resource) {
				v.errf("%s: unknown resource %q", w, e.Resource)
			}
			if e.PerLevel <= 0 {
				v.errf("%s: perLevel must be > 0", w)
			}
		case "idle_tax":
			if !v.is("resource", e.Resource) {
				v.errf("%s: unknown resource %q", w, e.Resource)
			}
			if e.PerLevel <= 0 || e.PerLevel > 10 {
				v.errf("%s: perLevel must be in (0, 10]", w)
			}
		case "housing":
			if e.PerLevel <= 0 {
				v.errf("%s: perLevel must be > 0", w)
			}
		case "modifier":
			if e.Modifier == nil {
				v.errf("%s: modifier required", w)
			} else {
				v.modifier(w, *e.Modifier)
			}
		}
	}
}

func (v *validator) resources() {
	p := v.p
	if len(p.Resources) == 0 {
		v.errf("resources must not be empty")
	}
	for i, r := range p.Resources {
		w := fmt.Sprintf("resources[%d](%s)", i, r.ID)
		v.text(w+".name", r.Name, true)
		if !ResourceKinds[r.Kind] {
			v.errf("%s: kind %q invalid", w, r.Kind)
		}
		if r.Kind == "premium" && (r.Tradable || r.Lootable) {
			v.errf("%s: premium resources must be neither tradable nor lootable", w)
		}
		if r.BaseStorage < 0 {
			v.errf("%s: baseStorage must be >= 0", w)
		}
	}
}

func (v *validator) races() {
	if len(v.p.Races) == 0 {
		v.errf("races must not be empty")
	}
	for i, r := range v.p.Races {
		w := fmt.Sprintf("races[%d](%s)", i, r.ID)
		v.text(w+".name", r.Name, true)
		v.modifiers(w, r.Modifiers)
		if r.Color != "" && !reColor.MatchString(r.Color) {
			v.errf("%s: color %q invalid", w, r.Color)
		}
	}
}

func (v *validator) buildings() {
	tabled := map[string]bool{}
	for _, t := range v.p.BuildingTables {
		tabled[t.Building] = true
	}
	for i, b := range v.p.Buildings {
		w := fmt.Sprintf("buildings[%d](%s)", i, b.ID)
		v.text(w+".name", b.Name, true)
		if !BuildingCats[b.Category] {
			v.errf("%s: category %q invalid", w, b.Category)
		}
		if b.MaxLevel < 1 || b.MaxLevel > 100 {
			v.errf("%s: maxLevel %d out of range [1, 100]", w, b.MaxLevel)
		}
		if !tabled[b.ID] { // an explicit level table replaces the geometric legacy cost
			v.costTime(w, b.Cost, b.Time)
		}
		v.requirements(w, b.Requires)
		v.effects(w, b.Effects)
		v.races_(w, b.Races)
		if b.Score < 0 {
			v.errf("%s: score must be >= 0", w)
		}
	}
}

func (v *validator) techs() {
	for i, t := range v.p.Techs {
		w := fmt.Sprintf("techs[%d](%s)", i, t.ID)
		v.text(w+".name", t.Name, true)
		if t.MaxLevel < 1 || t.MaxLevel > 100 {
			v.errf("%s: maxLevel %d out of range [1, 100]", w, t.MaxLevel)
		}
		if t.Branch == "" {
			v.errf("%s: branch required", w)
		}
		v.costTime(w, t.Cost, t.Time)
		v.requirements(w, t.Requires)
		v.effects(w, t.Effects)
		v.races_(w, t.Races)
	}
}

func (v *validator) units() {
	for i, u := range v.p.Units {
		w := fmt.Sprintf("units[%d](%s)", i, u.ID)
		v.text(w+".name", u.Name, true)
		if !UnitClasses[u.Class] {
			v.errf("%s: class %q invalid", w, u.Class)
		}
		v.races_(w, u.Races)
		if u.Attack < 0 || u.Defense < 0 || u.Scout < 0 {
			v.errf("%s: attack/defense/scout must be >= 0", w)
		}
		if u.HP <= 0 || u.HP > 1e6 {
			v.errf("%s: hp must be in (0, 1e6]", w)
		}
		if u.Speed <= 0 || u.Speed > 1000 {
			v.errf("%s: speed must be in (0, 1000]", w)
		}
		if u.Carry < 0 || u.Upkeep < 0 {
			v.errf("%s: carry/upkeep must be >= 0", w)
		}
		if u.TrainedIn != "" { // trainable unit
			if !v.is("building", u.TrainedIn) {
				v.errf("%s: trainedIn unknown building %q", w, u.TrainedIn)
			}
			v.amounts(w+".cost", u.Cost, false)
			if u.TrainSeconds < 1 {
				v.errf("%s: trainSeconds must be >= 1", w)
			}
		}
		v.requirements(w, u.Requires)
		for _, b := range u.Bonuses {
			if !UnitClasses[b.VsClass] {
				v.errf("%s: bonus vsClass %q invalid", w, b.VsClass)
			}
			if b.Pct < -90 || b.Pct > 500 {
				v.errf("%s: bonus pct %v out of range [-90, 500]", w, b.Pct)
			}
		}
	}
}

func (v *validator) heroes() {
	h := v.p.Heroes
	slots := set(h.Slots...)
	if len(h.Classes) > 0 {
		if h.MaxLevel < 1 || h.MaxLevel > 200 {
			v.errf("heroes.maxLevel %d out of range [1, 200]", h.MaxLevel)
		}
		if h.XPBase < 1 || h.XPGrowth < 1 || h.XPGrowth > 5 {
			v.errf("heroes.xpBase must be >= 1 and xpGrowth in [1, 5]")
		}
	}
	for i, c := range h.Classes {
		w := fmt.Sprintf("heroes.classes[%d](%s)", i, c.ID)
		v.text(w+".name", c.Name, true)
		v.modifiers(w, c.Modifiers)
		v.modifiers(w+".perLevel", c.PerLevel)
	}
	for i, s := range h.Skills {
		w := fmt.Sprintf("heroes.skills[%d](%s)", i, s.ID)
		v.text(w+".name", s.Name, true)
		if s.MaxLevel < 1 || s.MaxLevel > 20 {
			v.errf("%s: maxLevel out of range [1, 20]", w)
		}
		for _, c := range s.Classes {
			if !v.is("heroClass", c) {
				v.errf("%s: unknown class %q", w, c)
			}
		}
		v.modifiers(w, s.Modifiers)
	}
	for i, it := range h.Items {
		w := fmt.Sprintf("heroes.items[%d](%s)", i, it.ID)
		v.text(w+".name", it.Name, true)
		if !slots[it.Slot] {
			v.errf("%s: slot %q not declared in heroes.slots", w, it.Slot)
		}
		if !ItemRarities[it.Rarity] {
			v.errf("%s: rarity %q invalid", w, it.Rarity)
		}
		v.modifiers(w, it.Modifiers)
	}
}

func (v *validator) world() {
	passable := false
	for i, t := range v.p.Terrains {
		w := fmt.Sprintf("terrains[%d](%s)", i, t.ID)
		v.text(w+".name", t.Name, true)
		if t.Weight < 1 {
			v.errf("%s: weight must be >= 1", w)
		}
		if t.Resource != "" && !v.is("resource", t.Resource) {
			v.errf("%s: unknown resource %q", w, t.Resource)
		}
		if !reColor.MatchString(t.Color) {
			v.errf("%s: color %q invalid", w, t.Color)
		}
		passable = passable || t.Passable
	}
	if !passable {
		v.errf("terrains: at least one passable terrain required")
	}
	for i, n := range v.p.Neutrals {
		w := fmt.Sprintf("neutrals[%d](%s)", i, n.ID)
		v.text(w+".name", n.Name, true)
		if !NeutralKinds[n.Kind] {
			v.errf("%s: kind %q invalid", w, n.Kind)
		}
		if len(n.Levels) == 0 {
			v.errf("%s: levels must not be empty", w)
		}
		if n.RespawnSeconds < 0 {
			v.errf("%s: respawnSeconds must be >= 0", w)
		}
		for j, l := range n.Levels {
			lw := fmt.Sprintf("%s.levels[%d]", w, j)
			if l.Level != j+1 {
				v.errf("%s: level must be %d (levels are consecutive from 1)", lw, j+1)
			}
			for u, c := range l.Guards {
				if !v.is("unit", u) {
					v.errf("%s: unknown guard unit %q", lw, u)
				}
				if c < 0 {
					v.errf("%s: guard count must be >= 0", lw)
				}
			}
			v.amounts(lw+".rewards", l.Rewards, true)
			for _, d := range l.Drops {
				if !v.is("heroItem", d.Item) {
					v.errf("%s: unknown drop item %q", lw, d.Item)
				}
				if d.Chance < 0 || d.Chance > 100 {
					v.errf("%s: drop chance %d out of range [0, 100]", lw, d.Chance)
				}
			}
		}
	}
}

func (v *validator) quests() {
	for i, q := range v.p.Quests {
		w := fmt.Sprintf("quests[%d](%s)", i, q.ID)
		v.text(w+".name", q.Name, true)
		v.text(w+".desc", q.Desc, true)
		for _, r := range q.Requires {
			if !v.is("quest", r) {
				v.errf("%s: requires unknown quest %q", w, r)
			}
		}
		g := q.Goal
		if !QuestGoalTypes[g.Type] {
			v.errf("%s: goal type %q invalid", w, g.Type)
		}
		switch g.Type {
		case "building_level":
			if !v.is("building", g.ID) {
				v.errf("%s: goal building %q unknown", w, g.ID)
			}
		case "tech_level":
			if !v.is("tech", g.ID) {
				v.errf("%s: goal tech %q unknown", w, g.ID)
			}
		case "units_trained":
			if g.ID != "" && !v.is("unit", g.ID) {
				v.errf("%s: goal unit %q unknown", w, g.ID)
			}
		}
		if g.Amount < 1 {
			v.errf("%s: goal amount must be >= 1", w)
		}
		v.amounts(w+".rewards", q.Rewards, true)
		if q.Screen != "" && !KnownScreens[q.Screen] {
			v.errf("%s: screen %q unknown", w, q.Screen)
		}
	}
	// quest prerequisite cycles
	graph := map[string][]string{}
	for _, q := range v.p.Quests {
		graph[q.ID] = q.Requires
	}
	if c := findCycle(graph); c != nil {
		v.errf("quests: prerequisite cycle %s", strings.Join(c, " -> "))
	}
}

func (v *validator) start() {
	s := v.p.Start
	v.amounts("start.resources", s.Resources, true)
	for id, lvl := range s.Buildings {
		b := v.building(id)
		if b == nil {
			v.errf("start.buildings: unknown building %q", id)
			continue
		}
		if lvl < 1 || lvl > b.MaxLevel {
			v.errf("start.buildings: level %d for %q out of range", lvl, id)
		}
	}
	for id, n := range s.Units {
		if !v.is("unit", id) || n < 0 {
			v.errf("start.units: invalid entry %q=%d", id, n)
		}
	}
	if s.Population < 0 {
		v.errf("start.population must be >= 0")
	}
}

func (v *validator) params() {
	pr := v.p.Params
	rng := func(name string, x, lo, hi float64) {
		if x < lo || x > hi {
			v.errf("params.%s=%v out of range [%v, %v]", name, x, lo, hi)
		}
	}
	rng("worldSpeed", pr.WorldSpeed, 0.1, 1000)
	rng("economyTickSeconds", float64(pr.EconomyTickSeconds), 60, 3600)
	if pr.EconomyTickSeconds > 0 && 86400%pr.EconomyTickSeconds != 0 {
		v.errf("params.economyTickSeconds=%d must divide 86400 (ticks are aligned to absolute time)", pr.EconomyTickSeconds)
	}
	rng("buildQueueMax", float64(pr.BuildQueueMax), 1, 20)
	rng("trainQueueMax", float64(pr.TrainQueueMax), 1, 50)
	rng("researchQueueMax", float64(pr.ResearchQueueMax), 1, 20)
	rng("cancelRefundPct", float64(pr.CancelRefundPct), 0, 100)
	rng("mapWidth", float64(pr.MapWidth), 8, 512)
	rng("mapHeight", float64(pr.MapHeight), 8, 512)
	rng("neutralDensityPct", float64(pr.NeutralDensityPct), 0, 60)
	rng("newbieProtectionHours", float64(pr.NewbieProtectionHours), 0, 24*30)
	rng("minTravelSeconds", float64(pr.MinTravelSeconds), 1, 3600)
	rng("combatRounds", float64(pr.CombatRounds), 1, 20)
	rng("lootPct", float64(pr.LootPct), 0, 100)
	rng("scoutDetectBasePct", float64(pr.ScoutDetectBasePct), 0, 100)
	rng("popGrowthPctPerHour", pr.PopGrowthPctPerHour, 0, 100)
	rng("foodPerPopPerHour", pr.FoodPerPopPerHour, 0, 10)
	rng("starvationPctPerTick", pr.StarvationPctPerTick, 0, 50)
	rng("marketFeePct", float64(pr.MarketFeePct), 0, 50)
	rng("marketMaxOpenOrders", float64(pr.MarketMaxOpenOrders), 1, 100)
	rng("exchangeFeePct", float64(pr.ExchangeFeePct), 0, 90)
	rng("allianceMaxMembers", float64(pr.AllianceMaxMembers), 2, 500)
	rng("citiesPerCapLevels", float64(pr.CitiesPerCapLevels), 1, 100)
	rng("territoriesPerLevel", float64(pr.TerritoriesPerLevel), 0, 20)
	rng("territoryBonusPct", float64(pr.TerritoryBonusPct), 0, 100)
	rng("heroesPerBuildingLevel", float64(pr.HeroesPerBuildingLevel), 0, 10)
	rng("heroWoundHours", float64(pr.HeroWoundHours), 0, 168)
	rng("attackScoreFloorPct", float64(pr.AttackScoreFloorPct), 0, 100)
	rng("desertionPctPerHour", pr.DesertionPctPerHour, 0, 100)
	rng("missionUpkeepMult", pr.MissionUpkeepMult, 0, 10)
	rng("overflowDecayPctPerHour", pr.OverflowDecayPctPerHour, 0, 100)
	rng("defenseCapPct", pr.DefenseCapPct, 0, 99)
	rng("defenderRetreatPct", float64(pr.DefenderRetreatPct), 1, 100)
	rng("sameRaceAttackPct", pr.SameRaceAttackPct, -95, 0)
	rng("sameFactionAttackPct", pr.SameFactionAttackPct, -95, 0)
	rng("maxAttacksPerTargetPerDay", float64(pr.MaxAttacksPerTargetPerDay), 0, 1000)
	rng("maxReinforcingPlayers", float64(pr.MaxReinforcingPlayers), 0, 100)
	rng("maxHeroes", float64(pr.MaxHeroes), 0, 50)
	rng("maxCities", float64(pr.MaxCities), 1, 100)
	rng("territoryYieldPerHour", pr.TerritoryYieldPerHour, 0, 1e6)
	rng("territoryYieldCapHours", float64(pr.TerritoryYieldCapHours), 1, 720)
	if pr.PopGrowthMode != "idle" && pr.PopGrowthMode != "free_housing" {
		v.errf("params.popGrowthMode %q must be idle|free_housing", pr.PopGrowthMode)
	}
	if pr.UpkeepResource != "" && !v.is("resource", pr.UpkeepResource) {
		v.errf("params.upkeepResource %q unknown", pr.UpkeepResource)
	}
	food := v.p.Params.FoodResource
	if food == "" || !v.is("resource", food) {
		v.errf("params.foodResource %q must reference a resource", food)
	}
	if pr.PremiumResource != "" && !v.is("resource", pr.PremiumResource) {
		v.errf("params.premiumResource %q unknown", pr.PremiumResource)
	}
	for r, x := range pr.ExchangeRates {
		if !v.is("resource", r) || x <= 0 {
			v.errf("params.exchangeRates: invalid entry %q=%v", r, x)
		}
	}
	v.amounts("params.allianceCreateCost", pr.AllianceCreateCost, true)
	v.amounts("params.heroHireCost", pr.HeroHireCost, true)
	ref := func(name, id, kind string) {
		if id != "" && !v.is(kind, id) {
			v.errf("params.%s: unknown %s %q", name, kind, id)
		}
	}
	ref("cityCapBuilding", pr.CityCapBuilding, "building")
	ref("territoryCapBuilding", pr.TerritoryCapBuilding, "building")
	ref("heroBuilding", pr.HeroBuilding, "building")
	ref("settlerUnit", pr.SettlerUnit, "unit")
	if len(v.p.Heroes.Classes) > 0 && pr.HeroBuilding == "" {
		v.errf("params.heroBuilding required when hero classes are defined")
	}
}

func (v *validator) theme() {
	t := v.p.Theme
	for _, k := range requiredPalette {
		c, ok := t.Palette[k]
		if !ok {
			v.errf("theme.palette.%s required", k)
		} else if !reColor.MatchString(c) {
			v.errf("theme.palette.%s=%q invalid colour", k, c)
		}
	}
	for field, m := range map[string]map[string]string{"icons": t.Icons, "art": t.Art} {
		for id, path := range m {
			if _, ok := v.ids[id]; !ok && id != "app" {
				v.errf("theme.%s: unknown entity %q", field, id)
			}
			if !strings.HasPrefix(path, "glyph:") && !reAssetPath.MatchString(path) {
				v.errf("theme.%s.%s: path %q must be a relative asset path (a-z0-9_-/ . png|webp|svg|jpg) or glyph:<name>", field, id, path)
			}
			if strings.Contains(path, "..") || strings.HasPrefix(path, "/") || strings.Contains(path, "://") {
				v.errf("theme.%s.%s: path %q must be relative to the pack assets dir", field, id, path)
			}
		}
	}
	if strings.TrimSpace(t.AssetsLicense) == "" {
		v.errf("theme.assetsLicense required (document provenance of all art)")
	}
}

func (v *validator) ui() {
	if len(v.p.UI.Screens) == 0 {
		v.errf("ui.screens must not be empty")
	}
	for _, s := range v.p.UI.Screens {
		if !KnownScreens[s] {
			v.errf("ui.screens: unknown screen %q", s)
		}
	}
	for id, pos := range v.p.UI.CityLayout {
		if !v.is("building", id) {
			v.errf("ui.cityLayout: unknown building %q", id)
		}
		if pos[0] < 0 || pos[0] > 100 || pos[1] < 0 || pos[1] > 100 {
			v.errf("ui.cityLayout.%s: position out of range", id)
		}
	}
}

func (v *validator) migrations() {
	for i, m := range v.p.Migrations {
		w := fmt.Sprintf("migrations[%d]", i)
		if !reVersion.MatchString(m.FromVersion) {
			v.errf("%s: fromVersion %q invalid", w, m.FromVersion)
		}
		for old, nw := range m.Renames {
			if _, ok := v.ids[nw]; !ok {
				v.errf("%s: rename %q -> %q targets unknown id", w, old, nw)
			}
		}
		for old, repl := range m.Removed {
			if repl != "" {
				if _, ok := v.ids[repl]; !ok {
					v.errf("%s: removed %q replacement %q unknown", w, old, repl)
				}
			}
		}
	}
}

func (v *validator) requirementCycles() {
	graph := map[string][]string{}
	add := func(id string, rs []Requirement) {
		for _, r := range rs {
			if r.Type == "building" || r.Type == "tech" {
				graph[id] = append(graph[id], r.ID)
			}
		}
	}
	for _, b := range v.p.Buildings {
		add(b.ID, b.Requires)
	}
	for _, t := range v.p.Techs {
		add(t.ID, t.Requires)
	}
	if c := findCycle(graph); c != nil {
		// A building requiring itself at a lower level is legitimate only via
		// the implicit previous-level rule, so explicit self-loops are errors.
		v.errf("requirements: dependency cycle %s", strings.Join(c, " -> "))
	}
}

// findCycle returns one cycle in a directed graph or nil (deterministic order).
func findCycle(g map[string][]string) []string {
	const (
		white = iota
		grey
		black
	)
	color := map[string]int{}
	var stack []string
	var cycle []string
	var visit func(n string) bool
	visit = func(n string) bool {
		color[n] = grey
		stack = append(stack, n)
		for _, m := range g[n] {
			switch color[m] {
			case grey:
				for i, s := range stack {
					if s == m {
						cycle = append(append([]string{}, stack[i:]...), m)
					}
				}
				return true
			case white:
				if visit(m) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[n] = black
		return false
	}
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if color[k] == white && visit(k) {
			return cycle
		}
	}
	return nil
}
