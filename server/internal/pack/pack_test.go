package pack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadMinimal(t *testing.T) *Pack {
	t.Helper()
	data, err := os.ReadFile("testdata/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	p := &Pack{Params: Defaults()}
	if err := json.Unmarshal(data, p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMinimalPackValid(t *testing.T) {
	data, _ := os.ReadFile("testdata/minimal.json")
	l, err := Parse(data)
	if err != nil {
		t.Fatalf("minimal pack invalid: %v", err)
	}
	if l.Building("hall") == nil || l.Unit("spear") == nil || l.Quest("q_army") == nil {
		t.Fatal("index lookups failed")
	}
	if len(l.Hash) != 64 {
		t.Fatal("hash missing")
	}
	// Defaults applied for params not in JSON.
	if l.Params.CombatRounds != Defaults().CombatRounds {
		t.Fatalf("default combatRounds not applied: %d", l.Params.CombatRounds)
	}
}

func TestHashStable(t *testing.T) {
	data, _ := os.ReadFile("testdata/minimal.json")
	a, _ := Parse(data)
	b, _ := Parse(a.JSON())
	if a.Hash != b.Hash {
		t.Fatal("canonical re-encoding changed hash")
	}
}

func TestUnknownFieldRejected(t *testing.T) {
	_, err := Parse([]byte(`{"formatVersion":1,"evil":"code()"}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func expectErr(t *testing.T, p *Pack, substr string) {
	t.Helper()
	errs := Validate(p)
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Fatalf("expected error containing %q, got %v", substr, errs)
}

func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(p *Pack)
		want   string
	}{
		{"dup id", func(p *Pack) { p.Units[0].ID = "hall" }, "duplicate id"},
		{"bad cost resource", func(p *Pack) { p.Buildings[0].Cost.Base["stone"] = 5 }, `unknown resource "stone"`},
		{"bad requirement", func(p *Pack) {
			p.Buildings[1].Requires = []Requirement{{Type: "tech", ID: "nope", Level: 1}}
		}, `unknown tech "nope"`},
		{"req level too high", func(p *Pack) {
			p.Buildings[1].Requires = []Requirement{{Type: "building", ID: "hall", Level: 99}}
		}, "out of range"},
		{"cycle", func(p *Pack) {
			p.Buildings[0].Requires = []Requirement{{Type: "building", ID: "barracks", Level: 1}}
		}, "dependency cycle"},
		{"bad modifier target", func(p *Pack) { p.Races[0].Modifiers[0].Target = "production.mana" }, "invalid modifier target"},
		{"bad modifier op", func(p *Pack) { p.Races[0].Modifiers[0].Op = "mul" }, "op"},
		{"premium tradable", func(p *Pack) { p.Resources[3].Tradable = true }, "premium"},
		{"growth range", func(p *Pack) { p.Buildings[0].Cost.Growth = 0.5 }, "cost.growth"},
		{"unit trainedIn", func(p *Pack) { p.Units[0].TrainedIn = "castle" }, "trainedIn"},
		{"neutral guard", func(p *Pack) { p.Neutrals[0].Levels[0].Guards["dragon"] = 1 }, "unknown guard unit"},
		{"level sequence", func(p *Pack) { p.Neutrals[0].Levels[1].Level = 5 }, "consecutive"},
		{"quest cycle", func(p *Pack) { p.Quests[0].Requires = []string{"q_army"} }, "prerequisite cycle"},
		{"start building level", func(p *Pack) { p.Start.Buildings["hall"] = 50 }, "start.buildings"},
		{"missing locale", func(p *Pack) { p.Locales = []string{"en", "ru"} }, `missing text for locale "ru"`},
		{"palette", func(p *Pack) { p.Theme.Palette["bg"] = "red" }, "invalid colour"},
		{"icon path traversal", func(p *Pack) { p.Theme.Icons = map[string]string{"wood": "../../etc/passwd"} }, "relative"},
		{"food resource", func(p *Pack) { p.Params.FoodResource = "bread" }, "foodResource"},
		{"tick alignment", func(p *Pack) { p.Params.EconomyTickSeconds = 700 }, "divide 86400"},
		{"screen", func(p *Pack) { p.UI.Screens = append(p.UI.Screens, "casino") }, "unknown screen"},
		{"format version", func(p *Pack) { p.FormatVersion = 99 }, "formatVersion"},
		{"migration target", func(p *Pack) {
			p.Migrations = []Migration{{FromVersion: "0.9.0", Renames: map[string]string{"old": "missing"}}}
		}, "targets unknown id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := loadMinimal(t)
			c.mutate(p)
			expectErr(t, p, c.want)
		})
	}
}

// TestShippedPacks validates every pack in the repository's packs/ directory.
func TestShippedPacks(t *testing.T) {
	dirs, _ := filepath.Glob("../../../packs/*/pack.json")
	for _, d := range dirs {
		t.Run(filepath.Base(filepath.Dir(d)), func(t *testing.T) {
			if _, err := LoadFile(d); err != nil {
				t.Fatal(err)
			}
		})
	}
}
