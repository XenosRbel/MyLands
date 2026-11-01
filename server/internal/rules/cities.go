package rules

import (
	"errors"

	"realmforge/server/internal/pack"
)

// City cell rules (ML-M01, WORLD W-02). All numbers come from pack.CityLayout.
// A nil layout means the pack defines no cell rules; callers apply no limit.

var (
	ErrCapitalNotDeletable = errors.New("the capital cannot be deleted")
	ErrCityNotCapturable   = errors.New("a city cannot be captured")
	ErrExpansionLocked     = errors.New("cell expansion is not unlocked")
	ErrCellsAtMaximum      = errors.New("city is at its maximum cell count")
	ErrNoFreeCell          = errors.New("no free cell")
)

// CellsOnFounding is the cell count of a new city; the capital has its own count.
func CellsOnFounding(l *pack.CityLayout, capital bool) int {
	if capital {
		return l.CapitalCells
	}
	return l.StartCells
}

// MaxCellsByLevel returns the cell maximum for a city level (1-based); ok is false outside the table.
func MaxCellsByLevel(l *pack.CityLayout, level int) (int, bool) {
	if level < 1 || level > len(l.MaxByLevel) {
		return 0, false
	}
	return l.MaxByLevel[level-1], true
}

// MaxExpandedCells is the highest cell count that science expansion can reach.
func MaxExpandedCells(l *pack.CityLayout) int { return l.MaxByLevel[len(l.MaxByLevel)-1] }

// ExpandCell adds one cell. unlocked is the science unlock (its level is UNKNOWN in the source,
// so it is a boolean input). The capital is not expandable: its count already exceeds the maximum.
func ExpandCell(l *pack.CityLayout, cells int, unlocked bool) (int, error) {
	if !unlocked {
		return cells, ErrExpansionLocked
	}
	if cells >= MaxExpandedCells(l) {
		return cells, ErrCellsAtMaximum
	}
	return cells + 1, nil
}

// MaxCities is the city limit: the capital plus the extra cities.
func MaxCities(l *pack.CityLayout) int { return 1 + l.ExtraCities }

// CanBeCaptured is always false: attackers can plunder a city but never take it.
func CanBeCaptured() error { return ErrCityNotCapturable }

// PlanCityDeletion checks that a city may be deleted. Its troops leave with an escape mission
// (all units, none are lost). The engine has no delete command yet, so this is not wired.
func PlanCityDeletion(capital bool, troops map[string]int64) (escape map[string]int64, err error) {
	if capital {
		return nil, ErrCapitalNotDeletable
	}
	escape = map[string]int64{}
	for u, n := range troops {
		if n > 0 {
			escape[u] = n
		}
	}
	return escape, nil
}

// FreeCell returns the lowest cell index below cells that is not occupied.
func FreeCell(occupied map[int]bool, cells int) (int, error) {
	for i := 0; i < cells; i++ {
		if !occupied[i] {
			return i, nil
		}
	}
	return 0, ErrNoFreeCell
}
