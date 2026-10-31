package game

import (
	"testing"

	"realmforge/server/internal/pack"
)

func TestClearOtherExclusiveKeepsUnrelatedResearch(t *testing.T) {
	p := &pack.Pack{Techs: []pack.Tech{
		{ID: "path_architect", ExclusiveGroup: "calling"},
		{ID: "path_steward", ExclusiveGroup: "calling"},
		{ID: "forestry"},
	}}
	c := &pctx{
		p: &pack.Loaded{Pack: p},
		pl: &Player{Techs: map[string]int{
			"path_architect": 2,
			"forestry":       4,
		}},
	}
	if !c.clearOtherExclusive(&p.Techs[1]) {
		t.Fatal("changing paths must remove the previous specialization")
	}
	if c.pl.Techs["path_architect"] != 0 {
		t.Fatal("the previous specialization remained active")
	}
	if c.pl.Techs["forestry"] != 4 {
		t.Fatal("unrelated research changed")
	}
	if c.clearOtherExclusive(&p.Techs[1]) {
		t.Fatal("repeating an idempotent path switch must not change anything")
	}
}

func TestClearOtherExclusiveLeavesStandardResearchUnchanged(t *testing.T) {
	p := &pack.Pack{Techs: []pack.Tech{
		{ID: "path_architect", ExclusiveGroup: "calling"},
		{ID: "forestry"},
	}}
	c := &pctx{p: &pack.Loaded{Pack: p},
		pl: &Player{Techs: map[string]int{"path_architect": 1, "forestry": 2}}}
	if c.clearOtherExclusive(&p.Techs[1]) {
		t.Fatal("ordinary technology must not clear a specialization")
	}
	if c.pl.Techs["path_architect"] != 1 {
		t.Fatal("ordinary technology cleared a specialization")
	}
}
