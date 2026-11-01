package sim

import (
	"fmt"
	"sort"
	"strings"

	"realmforge/server/internal/pack"
)

// Scenario is one labelled run.
type Scenario struct {
	Label  string  `json:"label"`
	Result *Result `json:"result"`
}

// RaceReport groups what is known about one race.
type RaceReport struct {
	Race      string     `json:"race"`
	Scenarios []Scenario `json:"scenarios"`
	PvE       []PvERow   `json:"pve"`
}

// PackReport is everything the simulator says about one pack.
type PackReport struct {
	Pack    string       `json:"pack"`
	Version string       `json:"version"`
	Hash    string       `json:"hash"`
	Races   []RaceReport `json:"races"`
}

// Flags turns a run into plain-language warnings. Thresholds are PROPOSED
// review prompts, not rules: they tell a designer where to look.
func Flags(p *pack.Loaded, r *Result) []string {
	var out []string
	s := r.Stats
	if s.QueueIdlePct >= 50 {
		out = append(out, fmt.Sprintf("resource-bound: the build queue is empty %.0f%% of the time, so saving up, not timers, sets the pace", s.QueueIdlePct))
	}
	last := r.Samples[len(r.Samples)-1]
	for _, res := range sortedKeys(last.Net) {
		if rs := p.Resource(res); rs != nil && rs.Kind == "premium" {
			continue // premium currency is not produced by buildings
		}
		if last.Net[res] <= 2 && costsSomething(p, res) {
			out = append(out, fmt.Sprintf("bottleneck: `%s` nets %.1f/h at the end, yet buildings or units cost it", res, last.Net[res]))
		}
	}
	if stalled, from := populationStalled(r); stalled {
		last := r.Samples[len(r.Samples)-1]
		out = append(out, fmt.Sprintf("population stalled from hour %.0f at %d of %d housing (growth mode %q: growth needs idle people, and every one has a job)",
			from, last.Population, last.Housing, p.Params.PopGrowthMode))
	}
	if s.PopAtHousingPct >= 50 {
		out = append(out, fmt.Sprintf("population capped by housing %.0f%% of the time", s.PopAtHousingPct))
	}
	if s.AtStorageCapPct >= 20 {
		out = append(out, fmt.Sprintf("a stock sits at its storage cap %.0f%% of the time", s.AtStorageCapPct))
	}
	if s.StarvedHours > 0 {
		out = append(out, fmt.Sprintf("food ran out for %.1f h (starvation)", s.StarvedHours))
	}
	if s.WageDeficitHours > 0 {
		out = append(out, fmt.Sprintf("upkeep resource ran out for %.1f h; %d units deserted", s.WageDeficitHours, s.Deserted))
	}
	if s.TargetUnit == "" {
		out = append(out, "no trainable combat unit without a technology requirement")
	} else if s.FirstArmyHour < 0 {
		out = append(out, "no army was trained within the horizon")
	}
	if last.WagePct >= 60 {
		out = append(out, fmt.Sprintf("upkeep consumes %.0f%% of the upkeep resource's production at the end", last.WagePct))
	}
	return out
}

// populationStalled reports a population that grew less than 5%% over the last
// third of the run while at least 20%% of its housing stayed free.
func populationStalled(r *Result) (bool, float64) {
	n := len(r.Samples)
	if n < 6 {
		return false, 0
	}
	from := r.Samples[n-1-n/3]
	last := r.Samples[n-1]
	room := last.Housing - last.Population
	if room*5 >= last.Housing && last.Housing > 0 && float64(last.Population) < float64(from.Population)*1.05 {
		return true, from.Hour
	}
	return false, 0
}

func costsSomething(p *pack.Loaded, res string) bool {
	for i := range p.Buildings {
		if p.Buildings[i].Cost.Base[res] > 0 {
			return true
		}
	}
	for i := range p.Units {
		if p.Units[i].Cost[res] > 0 {
			return true
		}
	}
	return false
}

func milestoneHours() []float64 { return []float64{0, 1, 2, 4, 8, 12, 16, 24} }

func pickSamples(r *Result) []Sample {
	var out []Sample
	for _, h := range milestoneHours() {
		for _, s := range r.Samples {
			if s.Hour >= h-0.001 {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

// Markdown renders a pack report. It lists facts and flags; it does not judge.
func Markdown(p *pack.Loaded, rep PackReport, date string) string {
	var b strings.Builder
	w := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	w("## Pack `%s` %s", rep.Pack, rep.Version)
	w("")
	w("Pack hash `%s…`. World speed ×%v. Income value is Σ net/h × exchange rate.", rep.Hash[:12], p.Params.WorldSpeed)
	w("")
	for _, rr := range rep.Races {
		w("### Race `%s`", rr.Race)
		w("")
		for _, sc := range rr.Scenarios {
			r := sc.Result
			w("**%s** — target unit `%s`; first army at %s; trained %d, deserted %d.", sc.Label, orDash(r.Stats.TargetUnit),
				hourOrDash(r.Stats.FirstArmyHour), r.Stats.Trained, r.Stats.Deserted)
			w("")
			w("| Hour | Pop/Housing | Employment | Income value/h | Wage %% | Garrison | Build levels | Score | Strongest neutral cleared |")
			w("|---:|---:|---:|---:|---:|---:|---:|---:|---|")
			for _, s := range pickSamples(r) {
				w("| %.0f | %d/%d | %.0f%% | %.1f | %.1f | %d | %d | %d | %s |", s.Hour, s.Population, s.Housing, s.Employment*100,
					s.IncomeValue, s.WagePct, s.Garrison, s.BuildLevels, s.Score, s.Beatable)
			}
			w("")
			w("Queue idle %.1f%%, population at housing cap %.1f%%, storage cap %.1f%%, food-starved %.2f h, upkeep-starved %.2f h.",
				r.Stats.QueueIdlePct, r.Stats.PopAtHousingPct, r.Stats.AtStorageCapPct, r.Stats.StarvedHours, r.Stats.WageDeficitHours)
			w("")
			if fl := Flags(p, r); len(fl) > 0 {
				for _, f := range fl {
					w("- ⚠ %s", f)
				}
			} else {
				w("- No warnings.")
			}
			w("")
		}
		w("**PvE difficulty** — cheapest single-unit army with ≤30%% losses, race modifiers only, 6 tiles one way.")
		w("")
		w("| Neutral | Guard power | Unit | Army | Loss %% | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |")
		w("|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|")
		for _, r := range rr.PvE {
			mark := ""
			if len(r.Unpriced) > 0 {
				mark = " †"
			}
			if r.Unit == "" {
				w("| %s L%d%s | %.0f | — | no single unit wins | | | %.0f | | | | |", r.Neutral, r.Level, mark, r.GuardPower, r.RewardValue)
				continue
			}
			w("| %s L%d%s | %.0f | %s | %d | %.1f | %.0f | %.0f | %.0f | %.0f | %s | %.1f |", r.Neutral, r.Level, mark, r.GuardPower, r.Unit,
				r.MinCount, r.LossPct, r.TrainValue, r.RewardValue, r.WageValue, r.NetValue, payback(r.PaybackRuns), r.RunsPerDay)
		}
		w("")
		w("† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.")
		w("")
	}
	return b.String()
}

func payback(x float64) string {
	if x <= 0 {
		return "never"
	}
	return fmt.Sprintf("%.1f", x)
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func hourOrDash(h float64) string {
	if h < 0 {
		return "never"
	}
	return fmt.Sprintf("h%.1f", h)
}

// Summary lists, per race, the headline numbers for a cross-race comparison.
func Summary(rep PackReport) string {
	var b strings.Builder
	w := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	w("| Race | Income value/h at 24 h | Pop at 24 h | First army | Strongest neutral cleared | Wage %% (careless) |")
	w("|---|---:|---:|---:|---|---:|")
	races := append([]RaceReport(nil), rep.Races...)
	sort.SliceStable(races, func(i, j int) bool { return races[i].Race < races[j].Race })
	for _, rr := range races {
		if len(rr.Scenarios) == 0 {
			continue
		}
		careful := rr.Scenarios[0].Result
		careless := rr.Scenarios[len(rr.Scenarios)-1].Result
		l := careful.Samples[len(careful.Samples)-1]
		w("| %s | %.1f | %d | %s | %s | %.1f |", rr.Race, l.IncomeValue, l.Population, hourOrDash(careful.Stats.FirstArmyHour),
			l.Beatable, careless.Samples[len(careless.Samples)-1].WagePct)
	}
	return b.String()
}
