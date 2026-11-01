package rules

import (
	"errors"
	"testing"

	"realmforge/server/internal/pack"
)

// Layout values are the source numbers for WORLD W-02; the pack supplies them in production.
var testLayout = &pack.CityLayout{StartCells: 22, CapitalCells: 51, MaxByLevel: []int{22, 27, 33, 38, 43}, ExtraCities: 7}

func TestCityCells(t *testing.T) { // GAME_RULES R-CITY-3
	if CellsOnFounding(testLayout, false) != 22 || CellsOnFounding(testLayout, true) != 51 {
		t.Fatal("founding cells")
	}
	for lvl, want := range []int{22, 27, 33, 38, 43} {
		if got, ok := MaxCellsByLevel(testLayout, lvl+1); !ok || got != want {
			t.Fatalf("level %d: %d %v", lvl+1, got, ok)
		}
	}
	if _, ok := MaxCellsByLevel(testLayout, 6); ok {
		t.Fatal("level 6 must be outside the table")
	}
	if MaxCities(testLayout) != 8 {
		t.Fatal("capital + 7")
	}
}

func TestExpandCell(t *testing.T) {
	if _, err := ExpandCell(testLayout, 22, false); !errors.Is(err, ErrExpansionLocked) {
		t.Fatalf("locked: %v", err)
	}
	n := 22
	for n < 43 {
		var err error
		if n, err = ExpandCell(testLayout, n, true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ExpandCell(testLayout, 43, true); !errors.Is(err, ErrCellsAtMaximum) {
		t.Fatalf("max: %v", err)
	}
}

func TestCityDeletionAndCapture(t *testing.T) {
	if _, err := PlanCityDeletion(true, nil); !errors.Is(err, ErrCapitalNotDeletable) {
		t.Fatal("capital must not be deletable")
	}
	esc, err := PlanCityDeletion(false, map[string]int64{"spear": 5, "wolf": 0})
	if err != nil || len(esc) != 1 || esc["spear"] != 5 {
		t.Fatalf("escape %v %v", esc, err)
	}
	if !errors.Is(CanBeCaptured(), ErrCityNotCapturable) {
		t.Fatal("cities cannot be captured")
	}
}

func TestFreeCell(t *testing.T) {
	if i, err := FreeCell(map[int]bool{0: true, 1: true}, 3); err != nil || i != 2 {
		t.Fatalf("%d %v", i, err)
	}
	if _, err := FreeCell(map[int]bool{0: true}, 1); !errors.Is(err, ErrNoFreeCell) {
		t.Fatal("full")
	}
}
