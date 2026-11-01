package game

import (
	"regexp"
	"sort"

	"realmforge/server/internal/rules"
)

// Cell and building mirror tables (migration 0004). cities.buildings stays the source the engine
// reads; this file dual-writes city_cells/city_buildings in the same transaction (GAME_RULES R-CITY-4).

var reBuildingID = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,63}$`)

const maxStoredLevel = 9 // city_buildings.level CHECK; higher levels are not mirrored (as in migration 0004)

// cellCount returns the number of cells of a city, or -1 when the pack sets no cell rules.
func (c *pctx) cellCount(ci *City) (int, error) {
	l := c.p.Params.CityLayout
	if l == nil {
		return -1, nil
	}
	n := rules.CellsOnFounding(l, ci.IsCapital)
	var have int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM city_cells WHERE city_id=$1`, ci.ID).Scan(&have); err != nil {
		return 0, err
	}
	return max(n, have), nil // expanded cells exist as rows
}

// checkFreeCell rejects a new building type when every cell is used (one building per cell).
func (c *pctx) checkFreeCell(ci *City, id string, jobs []Job) error {
	cells, err := c.cellCount(ci)
	if err != nil || cells < 0 || ci.Buildings[id] > 0 {
		return err
	}
	used := map[string]bool{}
	for b, l := range ci.Buildings {
		if l > 0 {
			used[b] = true
		}
	}
	for _, j := range jobs {
		used[j.Item] = true
	}
	if used[id] {
		return nil
	}
	if len(used) >= cells {
		return rule("no_free_cell", "no free cell in this city")
	}
	return nil
}

// syncCityBuildings makes city_buildings match ci.Buildings and creates missing city_cells.
// Rows of buildings under construction (not yet in ci.Buildings) are left alone.
func (c *pctx) syncCityBuildings(ci *City) error {
	cells, err := c.cellCount(ci)
	if err != nil {
		return err
	}
	type row struct {
		level, cell int
	}
	have := map[string]row{}
	occupied := map[int]bool{}
	rows, err := c.tx.Query(c.ctx, `SELECT b.building_id, b.level, k.cell_index FROM city_buildings b JOIN city_cells k ON k.id=b.cell_id WHERE b.city_id=$1`, ci.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var r row
		if err := rows.Scan(&id, &r.level, &r.cell); err != nil {
			rows.Close()
			return err
		}
		have[id] = r
		occupied[r.cell] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if cells > 0 {
		if _, err := c.tx.Exec(c.ctx, `INSERT INTO city_cells(city_id, cell_index) SELECT $1, g FROM generate_series(0, $2::int - 1) g ON CONFLICT (city_id, cell_index) DO NOTHING`, ci.ID, cells); err != nil {
			return err
		}
	}
	for _, id := range sortedBuildingIDs(ci.Buildings) {
		lvl := ci.Buildings[id]
		if lvl < 1 || lvl > maxStoredLevel || !reBuildingID.MatchString(id) {
			continue
		}
		if r, ok := have[id]; ok {
			if r.level != lvl {
				if _, err := c.tx.Exec(c.ctx, `UPDATE city_buildings SET level=$3, status='idle', finished_at=NULL, updated_at=now() WHERE city_id=$1 AND building_id=$2`, ci.ID, id, lvl); err != nil {
					return err
				}
			}
			continue
		}
		limit := cells
		if limit < 0 {
			limit = len(occupied) + 1 // no cell rules: grow the plan on demand
		}
		idx, err := rules.FreeCell(occupied, limit)
		if err != nil {
			return rule("no_free_cell", "no free cell in this city")
		}
		var cellID int64
		if err := c.tx.QueryRow(c.ctx, `INSERT INTO city_cells(city_id, cell_index) VALUES ($1,$2)
			ON CONFLICT (city_id, cell_index) DO UPDATE SET cell_index=EXCLUDED.cell_index RETURNING id`, ci.ID, idx).Scan(&cellID); err != nil {
			return err
		}
		if _, err := c.tx.Exec(c.ctx, `INSERT INTO city_buildings(city_id, cell_id, building_id, level) VALUES ($1,$2,$3,$4)`, ci.ID, cellID, id, lvl); err != nil {
			return err
		}
		occupied[idx] = true
	}
	return nil
}

func sortedBuildingIDs(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
