// Command packtool validates Game Packs and prints a summary.
//
//	packtool validate <packDir|pack.json>...
//	packtool summary  <packDir>
package main

import (
	"fmt"
	"os"
	"sort"

	"realmforge/server/internal/pack"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: packtool validate|summary <pack>...")
		os.Exit(2)
	}
	failed := false
	for _, path := range os.Args[2:] {
		p, err := pack.LoadFile(path)
		if err != nil {
			fmt.Printf("FAIL %s\n%v\n", path, err)
			failed = true
			continue
		}
		fmt.Printf("OK   %s  id=%s version=%s hash=%s…\n", path, p.ID, p.Version, p.Hash[:12])
		if os.Args[1] == "summary" {
			fmt.Printf("     title=%q locales=%v\n", p.Meta.Title.Get(p.DefaultLocale, p.DefaultLocale), p.Locales)
			counts := map[string]int{"resources": len(p.Resources), "races": len(p.Races), "buildings": len(p.Buildings), "techs": len(p.Techs),
				"units": len(p.Units), "heroClasses": len(p.Heroes.Classes), "heroSkills": len(p.Heroes.Skills), "items": len(p.Heroes.Items),
				"terrains": len(p.Terrains), "neutrals": len(p.Neutrals), "quests": len(p.Quests)}
			keys := make([]string, 0, len(counts))
			for k := range counts {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("     %-12s %d\n", k, counts[k])
			}
		}
	}
	if failed {
		os.Exit(1)
	}
}
