// Command sim plays the first day of a Game Pack with a bot and writes a balance report.
//
//	sim -pack ../packs/ashen-crowns -out ../docs/reports/BALANCE_SIM.md [-json out.json] [-hours 24]
//
// It needs no database. Output is deterministic for a given pack and flags.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/sim"
)

func main() {
	packs := flag.String("pack", "", "comma-separated pack directories")
	out := flag.String("out", "", "Markdown report path (default: stdout)")
	jsonOut := flag.String("json", "", "write full curves as JSON")
	hours := flag.Float64("hours", 24, "game hours to simulate")
	milPct := flag.Int("military-pct", 40, "percent of stock the bot spends on the army per decision")
	flag.Parse()
	if *packs == "" {
		fmt.Fprintln(os.Stderr, "usage: sim -pack <dir>[,<dir>...] [-out report.md] [-json curves.json] [-hours 24]")
		os.Exit(2)
	}
	var md strings.Builder
	var all []sim.PackReport
	for _, dir := range strings.Split(*packs, ",") {
		rep, text, err := runPack(dir, time.Duration(*hours*float64(time.Hour)), *milPct)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sim: %s: %v\n", dir, err)
			os.Exit(1)
		}
		all = append(all, rep)
		md.WriteString(text)
	}
	if *jsonOut != "" {
		data, _ := json.MarshalIndent(all, "", " ")
		if err := os.WriteFile(*jsonOut, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "sim:", err)
			os.Exit(1)
		}
	}
	if *out == "" {
		fmt.Print(md.String())
		return
	}
	if err := os.WriteFile(*out, []byte(md.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "sim:", err)
		os.Exit(1)
	}
}

func runPack(dir string, horizon time.Duration, milPct int) (sim.PackReport, string, error) {
	rep := sim.PackReport{}
	// Each scenario reloads the pack because the simulator may override params.
	p, err := pack.LoadFile(dir)
	if err != nil {
		return rep, "", err
	}
	rep = sim.PackReport{Pack: p.ID, Version: p.Version, Hash: p.Hash}
	for _, race := range p.Races {
		rr := sim.RaceReport{Race: race.ID}
		for _, aware := range []bool{true, false} {
			pp, err := pack.LoadFile(dir)
			if err != nil {
				return rep, "", err
			}
			label := "Wage-aware bot (trains only while upkeep income covers the new units)"
			if !aware {
				label = "Careless bot (ignores wages)"
			}
			res, err := sim.Run(sim.Config{Pack: pp, Race: race.ID, Horizon: horizon, MilitaryPct: milPct, WageAware: aware})
			if err != nil {
				return rep, "", err
			}
			rr.Scenarios = append(rr.Scenarios, sim.Scenario{Label: label, Result: res})
		}
		rr.PvE = sim.PvE(p, race.ID, sim.PvEOptions{})
		rep.Races = append(rep.Races, rr)
	}
	return rep, sim.Markdown(p, rep, ""), nil
}
