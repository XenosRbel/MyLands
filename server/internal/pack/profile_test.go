package pack

import (
	"strings"
	"testing"
)

// ML-D08: profile resolution is generic; no profile name is known to the engine.
func profilePack(t *testing.T) *Loaded {
	t.Helper()
	p := withMyLands(t)
	p.ParamDefs = []ParamDef{{ID: "production", Min: 0.5, Max: 4, Default: 1, Proposed: true}}
	p.ServerProfiles = []ServerProfile{
		{ID: "economic", Default: true, Proposed: true, Multipliers: map[string]float64{"production": 2}},
		{ID: "tactical"},
	}
	l, err := Build(p)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestResolveProfileDefaultAndBonus(t *testing.T) {
	l := profilePack(t)
	r, err := l.ResolveProfile("", nil)
	if err != nil || r.ID != "economic" || r.Params["production"] != 2 {
		t.Fatalf("default: %+v %v", r, err)
	}
	r, err = l.ResolveProfile("tactical", map[string]float64{"production": 3})
	if err != nil || r.Params["production"] != 3 {
		t.Fatalf("bonus on default value: %+v %v", r, err)
	}
	r, err = l.ResolveProfile("economic", map[string]float64{"production": 1.5})
	if err != nil || r.Params["production"] != 3 {
		t.Fatalf("bonus multiplies profile value: %+v %v", r, err)
	}
}

func TestResolveProfileRejects(t *testing.T) {
	l := profilePack(t)
	cases := []struct {
		name, id string
		bonus    map[string]float64
		want     string
	}{
		{"unknown profile", "ghost", nil, `unknown server profile "ghost"`},
		{"unknown bonus", "tactical", map[string]float64{"luck": 1}, `unknown parameter "luck"`},
		{"bonus below min", "tactical", map[string]float64{"production": 0.1}, "out of range"},
		{"bonus above max", "tactical", map[string]float64{"production": 5}, "out of range"},
		{"effective above max", "economic", map[string]float64{"production": 3}, "out of range"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := l.ResolveProfile(c.id, c.bonus)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want %q", err, c.want)
			}
		})
	}
}

func TestResolveProfileNoDefault(t *testing.T) {
	p := withMyLands(t)
	l, err := Build(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.ResolveProfile("", nil); err == nil {
		t.Fatal("empty id without default profile must fail")
	}
}

func TestProfileValidation(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(p *Pack)
		want   string
	}{
		{"default must be proposed", func(p *Pack) { p.ServerProfiles[0].Proposed = false }, "default profile must be proposed"},
		{"two defaults", func(p *Pack) {
			p.ServerProfiles = append(p.ServerProfiles, ServerProfile{ID: "b", Default: true, Proposed: true})
		}, "at most one default"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := withMyLands(t)
			p.ServerProfiles[0].Default = true
			p.ServerProfiles[0].Proposed = true
			c.mutate(p)
			expectErr(t, p, c.want)
		})
	}
}
