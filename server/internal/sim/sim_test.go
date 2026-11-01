package sim

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"realmforge/server/internal/pack"
)

func loadPack(t *testing.T, name string) *pack.Loaded {
	t.Helper()
	p, err := pack.LoadFile(filepath.Join("..", "..", "..", "packs", name))
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return p
}

func firstRace(p *pack.Loaded) string { return p.Races[0].ID }

func TestRunIsDeterministic(t *testing.T) {
	cfg := func() Config {
		return Config{Pack: loadPack(t, "ashen-crowns"), Race: "thornkin", Horizon: 6 * time.Hour}
	}
	a, err := Run(cfg())
	if err != nil {
		t.Fatal(err)
	}
	b, err := Run(cfg())
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatal("two runs with the same config differ")
	}
}

func TestRunRespectsGameInvariants(t *testing.T) {
	for _, name := range []string{"ashen-crowns", "void-frontier"} {
		p := loadPack(t, name)
		for _, race := range p.Races {
			res, err := Run(Config{Pack: p, Race: race.ID, Horizon: 24 * time.Hour})
			if err != nil {
				t.Fatalf("%s/%s: %v", name, race.ID, err)
			}
			if len(res.Samples) < 24 {
				t.Fatalf("%s/%s: want hourly samples, got %d", name, race.ID, len(res.Samples))
			}
			for _, s := range res.Samples {
				for r, v := range s.Stock {
					if v < 0 {
						t.Fatalf("%s/%s: negative stock %s at h%.1f", name, race.ID, r, s.Hour)
					}
				}
				if s.Population < 0 || s.Population > s.Housing && s.Housing > 0 {
					t.Fatalf("%s/%s: population %d outside housing %d at h%.1f", name, race.ID, s.Population, s.Housing, s.Hour)
				}
			}
			for _, e := range res.Builds {
				b := p.Building(e.Building)
				if b == nil || e.Level < 1 || e.Level > b.MaxLevel {
					t.Fatalf("%s/%s: invalid build event %+v", name, race.ID, e)
				}
			}
			if res.Stats.MaxQueueLen > p.Params.BuildQueueMax {
				t.Fatalf("%s/%s: build queue %d exceeds max %d", name, race.ID, res.Stats.MaxQueueLen, p.Params.BuildQueueMax)
			}
		}
	}
}

func TestEconomyGrowsOnFirstDay(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	res, err := Run(Config{Pack: p, Race: firstRace(p), Horizon: 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	first, last := res.Samples[0], res.Samples[len(res.Samples)-1]
	if last.IncomeValue <= first.IncomeValue {
		t.Fatalf("income did not grow: %.1f -> %.1f", first.IncomeValue, last.IncomeValue)
	}
	if len(res.Builds) == 0 {
		t.Fatal("bot never built anything")
	}
}

func TestWageAwareBotCarriesLessUpkeepThanCarelessBot(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	race := firstRace(p)
	careless, err := Run(Config{Pack: p, Race: race, Horizon: 24 * time.Hour, MilitaryPct: 60})
	if err != nil {
		t.Fatal(err)
	}
	careful, err := Run(Config{Pack: p, Race: race, Horizon: 24 * time.Hour, MilitaryPct: 60, WageAware: true})
	if err != nil {
		t.Fatal(err)
	}
	cl := careless.Samples[len(careless.Samples)-1]
	cf := careful.Samples[len(careful.Samples)-1]
	if cl.WagePct <= 100 {
		t.Fatalf("careless bot should out-spend its wage income, got %.1f%%", cl.WagePct)
	}
	if cf.WagePct >= cl.WagePct || cf.Garrison >= cl.Garrison {
		t.Fatalf("wage-aware bot should carry less: wage %.1f%% vs %.1f%%, garrison %d vs %d", cf.WagePct, cl.WagePct, cf.Garrison, cl.Garrison)
	}
}

func TestUnknownRaceIsRejected(t *testing.T) {
	if _, err := Run(Config{Pack: loadPack(t, "ashen-crowns"), Race: "nope"}); err == nil {
		t.Fatal("expected an error for an unknown race")
	}
}

func TestPvETableCoversEveryNeutralLevel(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	rows := PvE(p, firstRace(p), PvEOptions{})
	want := 0
	for _, n := range p.Neutrals {
		want += len(n.Levels)
	}
	if len(rows) != want {
		t.Fatalf("rows %d, want %d", len(rows), want)
	}
	for _, r := range rows {
		if r.Unit == "" {
			continue // no unit of this race can win; reported as unwinnable
		}
		if r.MinCount < 1 || r.LossPct > 30.0001 {
			t.Fatalf("row %+v violates the loss limit", r)
		}
	}
}

func TestPvEMinimalArmyIsMinimal(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	rows := PvE(p, firstRace(p), PvEOptions{})
	for _, r := range rows {
		if r.Unit == "" || r.MinCount < 2 {
			continue
		}
		if winsWithin(p, firstRace(p), r.Unit, r.MinCount-1, r.Neutral, r.Level, 30) {
			t.Fatalf("%s L%d: %d x %s is not minimal", r.Neutral, r.Level, r.MinCount, r.Unit)
		}
	}
}

func TestFlagsReportBottleneckAndWages(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	r := &Result{Samples: []Sample{{Hour: 24, Net: map[string]float64{"gold": 0.5, "grain": 200}, WagePct: 70}},
		Stats: Stats{TargetUnit: "x", FirstArmyHour: 1, WageDeficitHours: 2, Deserted: 5}}
	joined := strings.Join(Flags(p, r), "\n")
	for _, want := range []string{"bottleneck: `gold`", "ran out for 2.0 h", "70%"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("flags missing %q:\n%s", want, joined)
		}
	}
}

func TestMarkdownMentionsEveryRaceAndNeutral(t *testing.T) {
	p := loadPack(t, "ashen-crowns")
	rep := PackReport{Pack: p.ID, Version: p.Version, Hash: p.Hash}
	for _, race := range p.Races {
		res, err := Run(Config{Pack: p, Race: race.ID, Horizon: 3 * time.Hour})
		if err != nil {
			t.Fatal(err)
		}
		rep.Races = append(rep.Races, RaceReport{Race: race.ID, Scenarios: []Scenario{{Label: "a", Result: res}}, PvE: PvE(p, race.ID, PvEOptions{})})
	}
	md := Markdown(p, rep, "")
	for _, race := range p.Races {
		if !strings.Contains(md, "### Race `"+race.ID+"`") {
			t.Fatalf("report lacks race %s", race.ID)
		}
	}
	for _, n := range p.Neutrals {
		if !strings.Contains(md, n.ID+" L1") {
			t.Fatalf("report lacks neutral %s", n.ID)
		}
	}
}

func TestPopulationStallIsDetected(t *testing.T) {
	var s []Sample
	for h := 0; h <= 24; h += 2 {
		s = append(s, Sample{Hour: float64(h), Population: 61, Housing: 220})
	}
	if ok, _ := populationStalled(&Result{Samples: s}); !ok {
		t.Fatal("flat population with free housing should be a stall")
	}
	for i := range s {
		s[i].Population = int64(60 + i*10)
		s[i].Housing = 400
	}
	if ok, _ := populationStalled(&Result{Samples: s}); ok {
		t.Fatal("growing population is not a stall")
	}
}
