package platform_test

import (
	"strings"
	"testing"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/platform"
)

// ML-D08: one server instance selects one profile and optional bonuses from the environment.
func TestLoadConfigServerProfile(t *testing.T) {
	t.Setenv("RF_SERVER_PROFILE", "economic")
	t.Setenv("RF_PROFILE_BONUSES", "production_multiplier=1.5, upkeep_multiplier=0.8")
	c, err := platform.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.ServerProfile != "economic" || c.ProfileBonuses["production_multiplier"] != 1.5 || c.ProfileBonuses["upkeep_multiplier"] != 0.8 {
		t.Fatalf("%+v", c)
	}
}

func TestLoadConfigProfileDefaultsEmpty(t *testing.T) {
	t.Setenv("RF_SERVER_PROFILE", "")
	t.Setenv("RF_PROFILE_BONUSES", "")
	c, err := platform.LoadConfig()
	if err != nil || c.ServerProfile != "" || len(c.ProfileBonuses) != 0 {
		t.Fatalf("%+v %v", c, err)
	}
}

func TestLoadConfigBadBonuses(t *testing.T) {
	for _, v := range []string{"production_multiplier", "production_multiplier=abc", "=1", "a=1,a=2", "a=NaN", "a=Inf"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("RF_PROFILE_BONUSES", v)
			if _, err := platform.LoadConfig(); err == nil || !strings.Contains(err.Error(), "RF_PROFILE_BONUSES") {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestConfigProfileAgainstMyLandsPack(t *testing.T) {
	l, err := pack.LoadFile("../../../packs/mylands")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("RF_SERVER_PROFILE", "")
	t.Setenv("RF_PROFILE_BONUSES", "production_multiplier=1.5")
	c, err := platform.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	r, err := l.ResolveProfile(c.ServerProfile, c.ProfileBonuses)
	if err != nil || r.ID != "combat" || r.Params["production_multiplier"] != 1.5 {
		t.Fatalf("%+v %v", r, err)
	}
	for _, bad := range []struct{ id, bonus string }{
		{"no_such_profile", ""}, {"combat", "production_multiplier=99"}, {"combat", "mysteryParam=1"},
	} {
		t.Setenv("RF_SERVER_PROFILE", bad.id)
		t.Setenv("RF_PROFILE_BONUSES", bad.bonus)
		c, err := platform.LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := l.ResolveProfile(c.ServerProfile, c.ProfileBonuses); err == nil {
			t.Errorf("%+v must be rejected", bad)
		}
	}
}
