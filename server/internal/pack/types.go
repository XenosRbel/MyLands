// Package pack defines the declarative, versioned Game Pack format.
//
// A Game Pack fully describes one game product built on the engine: names,
// locales, theme, resources, races, buildings, technologies, units, heroes,
// neutral objects, quests, balance parameters and starting conditions.
// Packs are pure data (JSON). They never contain executable code; every
// behaviour they configure is implemented by the engine and selected through
// closed enumerations (effect types, modifier targets, quest condition types).
package pack

// FormatVersion is the newest pack format version this engine understands.
const FormatVersion = 1

// LocText maps a locale code ("en", "ru", ...) to a localized string.
type LocText map[string]string

// Get returns the text for locale, falling back to fallback locale.
func (t LocText) Get(locale, fallback string) string {
	if s, ok := t[locale]; ok && s != "" {
		return s
	}
	return t[fallback]
}

// Amounts maps resource IDs to whole-unit quantities.
type Amounts map[string]int64

type Pack struct {
	FormatVersion int        `json:"formatVersion"`
	ID            string     `json:"id"`
	Version       string     `json:"version"`
	Meta          Meta       `json:"meta"`
	Locales       []string   `json:"locales"`
	DefaultLocale string     `json:"defaultLocale"`
	Theme         Theme      `json:"theme"`
	Params        Params     `json:"params"`
	Resources     []Resource `json:"resources"`
	Races         []Race     `json:"races"`
	Buildings     []Building `json:"buildings"`
	Techs         []Tech     `json:"techs"`
	Units         []Unit     `json:"units"`
	Heroes        HeroConfig `json:"heroes"`
	Terrains      []Terrain  `json:"terrains"`
	Neutrals      []Neutral  `json:"neutrals"`
	Quests        []Quest    `json:"quests"`
	Start         Start      `json:"start"`
	Modes         []string   `json:"modes"`
	UI            UIConfig   `json:"ui"`
	// Migrations describe how worlds pinned to older versions of this pack
	// can be moved to this version (ID renames, removed IDs replacement).
	Migrations []Migration `json:"migrations,omitempty"`
	// My Lands extensions. All optional; see docs/modernization/PACK_SCHEMA.md.
	Crystals       []CrystalType   `json:"crystals,omitempty"`
	BuildingTables []BuildingTable `json:"buildingTables,omitempty"`
	ParamDefs      []ParamDef      `json:"paramDefs,omitempty"`
	ServerProfiles []ServerProfile `json:"serverProfiles,omitempty"`
	UnitDefs       []UnitDef       `json:"unitDefs,omitempty"`
	Services       []Service       `json:"services,omitempty"`
	Geometry       *MapGeometry    `json:"geometry,omitempty"`
	// Strings holds free UI texts per locale (tutorial lines, flavour text).
	Strings map[string]LocText `json:"strings,omitempty"`
}

type Meta struct {
	Title       LocText `json:"title"`
	Subtitle    LocText `json:"subtitle,omitempty"`
	Description LocText `json:"description"`
	Studio      string  `json:"studio,omitempty"`
	License     string  `json:"license,omitempty"`
}

type Theme struct {
	// Palette are CSS colour tokens consumed by the client (e.g. "bg", "accent").
	Palette map[string]string `json:"palette"`
	Fonts   map[string]string `json:"fonts,omitempty"`
	// Icons maps entity IDs to asset paths relative to the pack's assets dir,
	// or to a built-in glyph name prefixed with "glyph:".
	Icons map[string]string `json:"icons,omitempty"`
	// Art maps entity IDs to large illustrations (building sprites, unit and
	// hero portraits, terrain textures, neutral sites), relative to the pack
	// assets dir. Optional; the client falls back to icons/glyphs.
	Art map[string]string `json:"art,omitempty"`
	// AssetsLicense documents the provenance of all bundled art.
	AssetsLicense string `json:"assetsLicense"`
}

// Params are global balance and rule parameters. All have engine defaults
// (see Defaults) so packs only override what they need.
type Params struct {
	WorldSpeed                float64            `json:"worldSpeed"`         // global time multiplier (>=1 speeds up timers)
	EconomyTickSeconds        int                `json:"economyTickSeconds"` // population / starvation step
	BuildQueueMax             int                `json:"buildQueueMax"`      // queued building upgrades per city
	TrainQueueMax             int                `json:"trainQueueMax"`      // queued training batches per city
	ResearchQueueMax          int                `json:"researchQueueMax"`   // queued researches per player
	CancelRefundPct           int                `json:"cancelRefundPct"`    // refund when cancelling a queued job
	MapWidth                  int                `json:"mapWidth"`
	MapHeight                 int                `json:"mapHeight"`
	NeutralDensityPct         int                `json:"neutralDensityPct"` // share of tiles with a neutral object
	NewbieProtectionHours     int                `json:"newbieProtectionHours"`
	MinTravelSeconds          int                `json:"minTravelSeconds"`
	CombatRounds              int                `json:"combatRounds"`
	LootPct                   int                `json:"lootPct"` // max share of unprotected stock that can be looted
	ScoutDetectBasePct        int                `json:"scoutDetectBasePct"`
	PopGrowthPctPerHour       float64            `json:"popGrowthPctPerHour"` // of free housing per hour
	FoodPerPopPerHour         float64            `json:"foodPerPopPerHour"`
	StarvationPctPerTick      float64            `json:"starvationPctPerTick"`      // population lost per tick while food is 0 and deficit
	FoodResource              string             `json:"foodResource"`              // resource consumed by population and armies
	PremiumResource           string             `json:"premiumResource,omitempty"` // non-tradable hard/soft premium currency (no real money in MVP)
	MarketFeePct              int                `json:"marketFeePct"`
	MarketMaxOpenOrders       int                `json:"marketMaxOpenOrders"`
	ExchangeRates             map[string]float64 `json:"exchangeRates,omitempty"` // NPC merchant value of 1 unit, in a common value scale
	ExchangeFeePct            int                `json:"exchangeFeePct"`
	AllianceCreateCost        Amounts            `json:"allianceCreateCost,omitempty"`
	AllianceMaxMembers        int                `json:"allianceMaxMembers"`
	CityCapBuilding           string             `json:"cityCapBuilding,omitempty"` // building whose level grants extra cities
	CitiesPerCapLevels        int                `json:"citiesPerCapLevels"`        // one extra city per N levels of CityCapBuilding
	SettlerUnit               string             `json:"settlerUnit,omitempty"`     // unit consumed to found a new city
	TerritoryCapBuilding      string             `json:"territoryCapBuilding,omitempty"`
	TerritoriesPerLevel       int                `json:"territoriesPerLevel"`
	TerritoryBonusPct         int                `json:"territoryBonusPct"` // production bonus per owned territory for its resource
	HeroHireCost              Amounts            `json:"heroHireCost,omitempty"`
	HeroBuilding              string             `json:"heroBuilding,omitempty"`
	HeroesPerBuildingLevel    int                `json:"heroesPerBuildingLevel"`
	HeroWoundHours            int                `json:"heroWoundHours"`
	AttackScoreFloorPct       int                `json:"attackScoreFloorPct"`       // cannot attack players weaker than this % of own score (anti-bullying); 0 disables
	UpkeepResource            string             `json:"upkeepResource,omitempty"`  // resource paying army upkeep (defaults to foodResource)
	DesertionPctPerHour       float64            `json:"desertionPctPerHour"`       // garrison lost per hour while upkeep is unpaid
	MissionUpkeepMult         float64            `json:"missionUpkeepMult"`         // upkeep multiplier for units away on missions
	PopGrowthMode             string             `json:"popGrowthMode"`             // "idle" (growth ∝ idle people) | "free_housing"
	OverflowDecayPctPerHour   float64            `json:"overflowDecayPctPerHour"`   // >0: stock may exceed storage and the surplus decays; 0: hard cap
	DefenseCapPct             float64            `json:"defenseCapPct"`             // max damage reduction from defense (%)
	DefenderRetreatPct        int                `json:"defenderRetreatPct"`        // defender retreats when losses reach this % (100 = never)
	SameRaceAttackPct         float64            `json:"sameRaceAttackPct"`         // attack modifier when attacking own race (e.g. -50)
	SameFactionAttackPct      float64            `json:"sameFactionAttackPct"`      // attack modifier when attacking own faction (e.g. -25)
	MaxAttacksPerTargetPerDay int                `json:"maxAttacksPerTargetPerDay"` // 0 = unlimited
	MaxReinforcingPlayers     int                `json:"maxReinforcingPlayers"`     // distinct foreign players that may station troops in a city
	MaxHeroes                 int                `json:"maxHeroes"`
	MaxCities                 int                `json:"maxCities"`
	TerritoryYieldPerHour     float64            `json:"territoryYieldPerHour"`  // units/hour of the terrain resource accrued on an owned tile
	TerritoryYieldCapHours    int                `json:"territoryYieldCapHours"` // uncollected yield stops accruing after this many hours
}

// Defaults returns engine default parameters.
func Defaults() Params {
	return Params{
		WorldSpeed: 1, EconomyTickSeconds: 600, BuildQueueMax: 3, TrainQueueMax: 5, ResearchQueueMax: 2,
		CancelRefundPct: 80, MapWidth: 64, MapHeight: 64, NeutralDensityPct: 12, NewbieProtectionHours: 72,
		MinTravelSeconds: 60, CombatRounds: 4, LootPct: 50, ScoutDetectBasePct: 30,
		PopGrowthPctPerHour: 10, FoodPerPopPerHour: 0.1, StarvationPctPerTick: 1,
		MarketFeePct: 2, MarketMaxOpenOrders: 10, ExchangeFeePct: 25, AllianceMaxMembers: 30,
		CitiesPerCapLevels: 5, TerritoriesPerLevel: 1, TerritoryBonusPct: 5,
		HeroesPerBuildingLevel: 1, HeroWoundHours: 6, AttackScoreFloorPct: 0,
		DesertionPctPerHour: 0.5, MissionUpkeepMult: 1, PopGrowthMode: "free_housing", DefenseCapPct: 90,
		DefenderRetreatPct: 100, MaxReinforcingPlayers: 3, MaxHeroes: 7, MaxCities: 9,
		TerritoryYieldPerHour: 20, TerritoryYieldCapHours: 24,
	}
}

type Resource struct {
	ID          string  `json:"id"`
	Name        LocText `json:"name"`
	Desc        LocText `json:"desc,omitempty"`
	Kind        string  `json:"kind"`                  // basic | food | rare | premium
	Capped      bool    `json:"capped"`                // limited by storage
	Tradable    bool    `json:"tradable"`              // can be listed on the market
	Lootable    bool    `json:"lootable"`              // can be plundered
	BaseStorage int64   `json:"baseStorage,omitempty"` // storage before buildings
}

// Modifier is a declarative additive-percentage or flat bonus.
// Target grammar (closed set, validated):
//
//	production.<resource|*>, storage.<resource|*>, housing, pop_growth,
//	build_time, research_time, train_time, upkeep,
//	attack.<unitclass|unitId|*>, defense.<unitclass|unitId|*>, hp.<...>,
//	speed, carry, loot, scout, wall, hero_xp, market_fee
type Modifier struct {
	Target string  `json:"target"`
	Op     string  `json:"op"` // pct | flat
	Value  float64 `json:"value"`
}

type Race struct {
	ID        string     `json:"id"`
	Name      LocText    `json:"name"`
	Desc      LocText    `json:"desc"`
	Modifiers []Modifier `json:"modifiers"`
	Color     string     `json:"color,omitempty"`
	Faction   string     `json:"faction,omitempty"` // races sharing a faction get SameFactionAttackPct
}

// Requirement gates buildings, techs, units and quests.
type Requirement struct {
	Type  string `json:"type"` // building | tech | race
	ID    string `json:"id"`
	Level int    `json:"level,omitempty"`
}

// Cost grows geometrically with level: cost(L) = round(base * growth^(L-1)).
type Cost struct {
	Base   Amounts `json:"base"`
	Growth float64 `json:"growth"`
}

// Duration grows geometrically: t(L) = base * growth^(L-1) seconds.
type Duration struct {
	BaseSeconds int     `json:"baseSeconds"`
	Growth      float64 `json:"growth"`
}

// Effect is what a building or technology does per level. Type is closed:
//
//	produce   (Resource, PerLevel units/hour, Workers per level)
//	storage   (Resource or "*", PerLevel)
//	protect   (Resource or "*", PerLevel units hidden from loot)
//	housing   (PerLevel population capacity)
//	idle_tax  (Resource, PerLevel units/hour per idle inhabitant per level)
//	modifier  (Modifier with value per level)
//	queue     (extra build queue slots per level) — reserved
type Effect struct {
	Type     string  `json:"type"`
	Resource string  `json:"resource,omitempty"`
	PerLevel float64 `json:"perLevel,omitempty"`
	// Growth makes per-level output geometric: v(L) = perLevel*L*growth^(L-1). 0 or 1 = linear.
	Growth   float64   `json:"growth,omitempty"`
	Workers  int       `json:"workers,omitempty"`
	Modifier *Modifier `json:"modifier,omitempty"`
}

type Building struct {
	ID       string        `json:"id"`
	Name     LocText       `json:"name"`
	Desc     LocText       `json:"desc"`
	Category string        `json:"category"` // core | economy | storage | military | science | special
	MaxLevel int           `json:"maxLevel"`
	Cost     Cost          `json:"cost"`
	Time     Duration      `json:"time"`
	Requires []Requirement `json:"requires,omitempty"`
	Effects  []Effect      `json:"effects,omitempty"`
	Races    []string      `json:"races,omitempty"` // empty = all
	// Score points per level (power ranking).
	Score int `json:"score"`
}

type Tech struct {
	ID       string        `json:"id"`
	Name     LocText       `json:"name"`
	Desc     LocText       `json:"desc"`
	Branch   string        `json:"branch"` // economy | military | civic | magic ...
	MaxLevel int           `json:"maxLevel"`
	Cost     Cost          `json:"cost"`
	Time     Duration      `json:"time"`
	Requires []Requirement `json:"requires,omitempty"`
	Effects  []Effect      `json:"effects"`
	Races    []string      `json:"races,omitempty"`
	// Exclusive groups implement specialisations: researching one tech of a
	// group locks the other techs of that group.
	ExclusiveGroup string `json:"exclusiveGroup,omitempty"`
	Score          int    `json:"score"`
}

type Unit struct {
	ID           string        `json:"id"`
	Name         LocText       `json:"name"`
	Desc         LocText       `json:"desc"`
	Class        string        `json:"class"` // infantry | ranged | cavalry | siege | scout | transport | settler | magic
	Races        []string      `json:"races,omitempty"`
	Attack       float64       `json:"attack"`
	Defense      float64       `json:"defense"`
	HP           float64       `json:"hp"`
	Speed        float64       `json:"speed"` // tiles per hour
	Carry        int64         `json:"carry"`
	Upkeep       float64       `json:"upkeep"` // food per hour
	Scout        float64       `json:"scout,omitempty"`
	Cost         Amounts       `json:"cost"`
	TrainSeconds int           `json:"trainSeconds"`
	TrainedIn    string        `json:"trainedIn"` // building ID
	Requires     []Requirement `json:"requires,omitempty"`
	Bonuses      []ClassBonus  `json:"bonuses,omitempty"`
	Score        int           `json:"score"`
}

type ClassBonus struct {
	VsClass string  `json:"vsClass"`
	Pct     float64 `json:"pct"`
}

type HeroConfig struct {
	Classes             []HeroClass `json:"classes"`
	Skills              []HeroSkill `json:"skills"`
	Items               []HeroItem  `json:"items"`
	Slots               []string    `json:"slots"`
	MaxLevel            int         `json:"maxLevel"`
	XPBase              int64       `json:"xpBase"`   // xp to reach level 2
	XPGrowth            float64     `json:"xpGrowth"` // geometric growth per level
	SkillPointsPerLevel int         `json:"skillPointsPerLevel"`
	// XP gained per point of enemy score destroyed.
	XPPerScore float64 `json:"xpPerScore"`
}

type HeroClass struct {
	ID        string     `json:"id"`
	Name      LocText    `json:"name"`
	Desc      LocText    `json:"desc"`
	Modifiers []Modifier `json:"modifiers"`          // innate
	PerLevel  []Modifier `json:"perLevel,omitempty"` // added per hero level
}

type HeroSkill struct {
	ID        string     `json:"id"`
	Name      LocText    `json:"name"`
	Desc      LocText    `json:"desc"`
	MaxLevel  int        `json:"maxLevel"`
	Classes   []string   `json:"classes,omitempty"`
	Modifiers []Modifier `json:"modifiers"` // per skill level
}

type HeroItem struct {
	ID        string     `json:"id"`
	Name      LocText    `json:"name"`
	Desc      LocText    `json:"desc"`
	Slot      string     `json:"slot"`
	Rarity    string     `json:"rarity"` // common | rare | epic | legendary
	Modifiers []Modifier `json:"modifiers"`
}

type Terrain struct {
	ID       string  `json:"id"`
	Name     LocText `json:"name"`
	Weight   int     `json:"weight"`             // relative frequency in map generation
	Resource string  `json:"resource,omitempty"` // resource boosted when owned as territory
	Passable bool    `json:"passable"`
	Color    string  `json:"color"`
}

type Neutral struct {
	ID             string         `json:"id"`
	Name           LocText        `json:"name"`
	Desc           LocText        `json:"desc"`
	Kind           string         `json:"kind"` // lair | ruin | node
	Levels         []NeutralLevel `json:"levels"`
	RespawnSeconds int            `json:"respawnSeconds"`
	Capturable     bool           `json:"capturable"` // becomes a territory once cleared
}

type NeutralLevel struct {
	Level   int              `json:"level"`
	Guards  map[string]int64 `json:"guards"` // unitId -> count (pack units, may be neutral-only)
	Rewards Amounts          `json:"rewards"`
	XP      int64            `json:"xp"`
	// Item drops: chance in percent per listed item.
	Drops []Drop `json:"drops,omitempty"`
}

type Drop struct {
	Item   string `json:"item"`
	Chance int    `json:"chance"`
}

// Quest conditions use a closed set of types evaluated against server state.
type Quest struct {
	ID       string    `json:"id"`
	Name     LocText   `json:"name"`
	Desc     LocText   `json:"desc"`
	Tutorial bool      `json:"tutorial,omitempty"`
	Requires []string  `json:"requires,omitempty"` // previous quest IDs (claimed)
	Goal     QuestGoal `json:"goal"`
	Rewards  Amounts   `json:"rewards"`
	Hint     LocText   `json:"hint,omitempty"`
	// UI target the tutorial highlights (screen id).
	Screen string `json:"screen,omitempty"`
}

type QuestGoal struct {
	Type   string `json:"type"` // building_level | tech_level | units_trained | missions_won | neutrals_cleared | join_alliance | market_trade | hero_hired | cities | territories
	ID     string `json:"id,omitempty"`
	Amount int64  `json:"amount"`
}

type Start struct {
	Resources  Amounts          `json:"resources"`
	Buildings  map[string]int   `json:"buildings"`
	Units      map[string]int64 `json:"units,omitempty"`
	Population int64            `json:"population"`
}

type UIConfig struct {
	// Screens enabled in the client, in navigation order.
	Screens []string `json:"screens"`
	// Optional per-building plot positions for the city view (0..100 %).
	CityLayout map[string][2]int `json:"cityLayout,omitempty"`
}

type Migration struct {
	FromVersion string            `json:"fromVersion"`
	Renames     map[string]string `json:"renames,omitempty"` // oldId -> newId (any entity kind)
	Removed     map[string]string `json:"removed,omitempty"` // removedId -> replacementId or "" to drop
}
