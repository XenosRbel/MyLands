package pack

import (
	"bytes"
	"encoding/json"
)

// Num is a numeric parameter with an optional per-race override.
// JSON form: a plain number, or {"base": N, "races": {"<raceId>": M}}.
type Num struct {
	Base  float64
	Races map[string]float64
}

// For returns the value for race, or Base when the race has no override.
func (n Num) For(race string) float64 {
	if v, ok := n.Races[race]; ok {
		return v
	}
	return n.Base
}

func (n Num) MarshalJSON() ([]byte, error) {
	if len(n.Races) == 0 {
		return json.Marshal(n.Base)
	}
	return json.Marshal(struct {
		Base  float64            `json:"base"`
		Races map[string]float64 `json:"races"`
	}{n.Base, n.Races})
}

func (n *Num) UnmarshalJSON(b []byte) error {
	if t := bytes.TrimSpace(b); len(t) > 0 && t[0] != '{' {
		*n = Num{}
		return json.Unmarshal(b, &n.Base)
	}
	var o struct {
		Base  float64            `json:"base"`
		Races map[string]float64 `json:"races"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields() // custom unmarshalers do not inherit the parent's strictness
	if err := dec.Decode(&o); err != nil {
		return err
	}
	*n = Num{Base: o.Base, Races: o.Races}
	return nil
}

// CrystalType is a pack-defined crystal class. Tier orders crystals by rarity.
type CrystalType struct {
	ID   string `json:"id"`
	Tier int    `json:"tier"`
}

// BuildingTable holds explicit per-level data for one building (levels 1..9).
type BuildingTable struct {
	Building string          `json:"building"`
	Levels   []BuildingLevel `json:"levels"`
}

// BuildingLevel is one row of a level table. An absent cost key means "not required".
// Crystal quantities are PROPOSED design values chosen by the pack author.
type BuildingLevel struct {
	Level       int            `json:"level"`
	Cost        Amounts        `json:"cost,omitempty"`
	Crystals    Amounts        `json:"crystals,omitempty"`
	TimeSeconds int64          `json:"timeSeconds,omitempty"`
	Effects     map[string]Num `json:"effects,omitempty"` // building-specific named numbers
}

// ParamDef declares a server profile parameter and its allowed range.
type ParamDef struct {
	ID      string  `json:"id"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Default float64 `json:"default"`
}

// ServerProfile sets named multipliers (production, upkeep, speeds) by ParamDef id.
type ServerProfile struct {
	ID          string             `json:"id"`
	Multipliers map[string]float64 `json:"multipliers"`
}

// UnitDef is a unit by category and level 1..4 with a parameter table.
type UnitDef struct {
	ID       string         `json:"id"`
	Category string         `json:"category"`
	Level    int            `json:"level"`
	Params   map[string]Num `json:"params"` // attackMin, attackMax, health, upkeep, population, rating, ...
	Cost     Amounts        `json:"cost,omitempty"`
	// Overrides replaces parameters or cost for one server profile id.
	Overrides map[string]UnitOverride `json:"overrides,omitempty"`
}

type UnitOverride struct {
	Params map[string]Num `json:"params,omitempty"`
	Cost   Amounts        `json:"cost,omitempty"`
}

// Service is a paid service priced in Params.PremiumResource.
// Proposed marks values that are not confirmed by the source.
type Service struct {
	ID       string         `json:"id"`
	Group    string         `json:"group"`
	Proposed bool           `json:"proposed,omitempty"`
	Levels   []ServiceLevel `json:"levels"`
}

type ServiceLevel struct {
	Price           int64 `json:"price"`
	DurationSeconds int64 `json:"durationSeconds"`
}

// MapGeometry describes the empire grid. Kingdom address is x-y, field address x-y-n.
type MapGeometry struct {
	EmpireWidth   int  `json:"empireWidth"`
	EmpireHeight  int  `json:"empireHeight"`
	KingdomWidth  int  `json:"kingdomWidth"`
	KingdomHeight int  `json:"kingdomHeight"`
	Proposed      bool `json:"proposed,omitempty"`
}

func (g MapGeometry) FieldsPerKingdom() int { return g.KingdomWidth * g.KingdomHeight }
