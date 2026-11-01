package pack

import "fmt"

// ResolvedProfile is the effective parameter set of one server instance.
type ResolvedProfile struct {
	ID     string
	Params map[string]float64 // every paramDef id -> effective multiplier
}

// ResolveProfile picks a server profile and applies bonus multipliers (ML-D08).
// An empty id selects the pack's default profile. Each bonus multiplies the profile
// value (paramDef default when the profile has none). Both the bonus and the result
// must lie within the paramDef range. No profile name is known to the engine.
func (l *Loaded) ResolveProfile(id string, bonuses map[string]float64) (ResolvedProfile, error) {
	var prof *ServerProfile
	for i := range l.ServerProfiles {
		s := &l.ServerProfiles[i]
		if (id == "" && s.Default) || (id != "" && s.ID == id) {
			prof = s
		}
	}
	if prof == nil {
		if id == "" {
			return ResolvedProfile{}, fmt.Errorf("pack %s: no default server profile; set the profile id", l.ID)
		}
		return ResolvedProfile{}, fmt.Errorf("pack %s: unknown server profile %q", l.ID, id)
	}
	r := ResolvedProfile{ID: prof.ID, Params: map[string]float64{}}
	defs := map[string]ParamDef{}
	for _, d := range l.ParamDefs {
		defs[d.ID] = d
		r.Params[d.ID] = d.Default
		if x, ok := prof.Multipliers[d.ID]; ok {
			r.Params[d.ID] = x
		}
	}
	for _, k := range sortedKeys(bonuses) {
		d, ok := defs[k]
		if !ok {
			return ResolvedProfile{}, fmt.Errorf("bonus: unknown parameter %q", k)
		}
		b := bonuses[k]
		if b < d.Min || b > d.Max {
			return ResolvedProfile{}, fmt.Errorf("bonus %s=%v out of range [%v, %v]", k, b, d.Min, d.Max)
		}
		x := r.Params[k] * b
		if x < d.Min || x > d.Max {
			return ResolvedProfile{}, fmt.Errorf("bonus %s=%v gives %v, out of range [%v, %v]", k, b, x, d.Min, d.Max)
		}
		r.Params[k] = x
	}
	return r, nil
}
