package pack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loaded is a validated pack plus lookup indexes. It is immutable after
// construction and safe for concurrent use.
type Loaded struct {
	*Pack
	Hash      string
	resources map[string]*Resource
	races     map[string]*Race
	buildings map[string]*Building
	techs     map[string]*Tech
	units     map[string]*Unit
	classes   map[string]*HeroClass
	skills    map[string]*HeroSkill
	items     map[string]*HeroItem
	terrains  map[string]*Terrain
	neutrals  map[string]*Neutral
	quests    map[string]*Quest
}

// Parse decodes pack JSON strictly (unknown fields are errors), applies
// parameter defaults and validates the result.
func Parse(data []byte) (*Loaded, error) {
	p := &Pack{Params: Defaults()}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(p); err != nil {
		return nil, fmt.Errorf("pack: decode: %w", err)
	}
	return Build(p)
}

// Build validates an in-memory pack and builds its indexes.
func Build(p *Pack) (*Loaded, error) {
	if errs := Validate(p); len(errs) > 0 {
		return nil, errs
	}
	canonical, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(canonical)
	l := &Loaded{Pack: p, Hash: hex.EncodeToString(sum[:])}
	l.index()
	return l, nil
}

// LoadFile reads a pack from pack.json inside dir (or the file itself).
func LoadFile(path string) (*Loaded, error) {
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		path = filepath.Join(path, "pack.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// JSON returns the canonical JSON encoding (used to pin a pack snapshot to a world).
func (l *Loaded) JSON() []byte {
	b, _ := json.Marshal(l.Pack)
	return b
}

func (l *Loaded) index() {
	p := l.Pack
	l.resources = map[string]*Resource{}
	for i := range p.Resources {
		l.resources[p.Resources[i].ID] = &p.Resources[i]
	}
	l.races = map[string]*Race{}
	for i := range p.Races {
		l.races[p.Races[i].ID] = &p.Races[i]
	}
	l.buildings = map[string]*Building{}
	for i := range p.Buildings {
		l.buildings[p.Buildings[i].ID] = &p.Buildings[i]
	}
	l.techs = map[string]*Tech{}
	for i := range p.Techs {
		l.techs[p.Techs[i].ID] = &p.Techs[i]
	}
	l.units = map[string]*Unit{}
	for i := range p.Units {
		l.units[p.Units[i].ID] = &p.Units[i]
	}
	l.classes = map[string]*HeroClass{}
	for i := range p.Heroes.Classes {
		l.classes[p.Heroes.Classes[i].ID] = &p.Heroes.Classes[i]
	}
	l.skills = map[string]*HeroSkill{}
	for i := range p.Heroes.Skills {
		l.skills[p.Heroes.Skills[i].ID] = &p.Heroes.Skills[i]
	}
	l.items = map[string]*HeroItem{}
	for i := range p.Heroes.Items {
		l.items[p.Heroes.Items[i].ID] = &p.Heroes.Items[i]
	}
	l.terrains = map[string]*Terrain{}
	for i := range p.Terrains {
		l.terrains[p.Terrains[i].ID] = &p.Terrains[i]
	}
	l.neutrals = map[string]*Neutral{}
	for i := range p.Neutrals {
		l.neutrals[p.Neutrals[i].ID] = &p.Neutrals[i]
	}
	l.quests = map[string]*Quest{}
	for i := range p.Quests {
		l.quests[p.Quests[i].ID] = &p.Quests[i]
	}
}

func (l *Loaded) Resource(id string) *Resource   { return l.resources[id] }
func (l *Loaded) Race(id string) *Race           { return l.races[id] }
func (l *Loaded) Building(id string) *Building   { return l.buildings[id] }
func (l *Loaded) Tech(id string) *Tech           { return l.techs[id] }
func (l *Loaded) Unit(id string) *Unit           { return l.units[id] }
func (l *Loaded) HeroClass(id string) *HeroClass { return l.classes[id] }
func (l *Loaded) HeroSkill(id string) *HeroSkill { return l.skills[id] }
func (l *Loaded) HeroItem(id string) *HeroItem   { return l.items[id] }
func (l *Loaded) Terrain(id string) *Terrain     { return l.terrains[id] }
func (l *Loaded) Neutral(id string) *Neutral     { return l.neutrals[id] }
func (l *Loaded) Quest(id string) *Quest         { return l.quests[id] }

// Allowed reports whether an entity restricted to races is available to race.
func Allowed(races []string, race string) bool {
	if len(races) == 0 {
		return true
	}
	for _, r := range races {
		if r == race {
			return true
		}
	}
	return false
}
