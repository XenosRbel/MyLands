package game

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

type Hero struct {
	ID           int64              `json:"id"`
	PlayerID     int64              `json:"playerId"`
	HomeCityID   int64              `json:"homeCityId"`
	Name         string             `json:"name"`
	Class        string             `json:"class"`
	Level        int                `json:"level"`
	XP           int64              `json:"xp"`
	Skills       map[string]int     `json:"skills"`
	Status       string             `json:"status"`
	WoundedUntil *time.Time         `json:"woundedUntil"`
	Items        []Item             `json:"items"`
	SkillPoints  int                `json:"skillPoints"`
	NextLevelXP  int64              `json:"nextLevelXp"`
	Modifiers    map[string]float64 `json:"modifiers"`
}

type Item struct {
	ID     int64   `json:"id"`
	Def    string  `json:"def"`
	HeroID *int64  `json:"heroId"`
	Slot   *string `json:"slot"`
	Source string  `json:"source"`
}

const heroCols = `id, player_id, home_city_id, name, class_id, level, xp, skills, status, wounded_until`

func scanHero(row pgx.Row) (*Hero, error) {
	h := &Hero{}
	err := row.Scan(&h.ID, &h.PlayerID, &h.HomeCityID, &h.Name, &h.Class, &h.Level, &h.XP, &h.Skills, &h.Status, &h.WoundedUntil)
	if h.Skills == nil {
		h.Skills = map[string]int{}
	}
	return h, err
}

// heroMods computes a hero's modifiers (GAME_RULES R-HERO-3):
// class innate + perLevel × (level−1) + Σ skill × level + equipped item modifiers.
func heroMods(p *pack.Loaded, h *Hero, items []Item) *rules.Mods {
	m := rules.NewMods()
	if cl := p.HeroClass(h.Class); cl != nil {
		m.AddAll(cl.Modifiers, 1)
		m.AddAll(cl.PerLevel, float64(h.Level-1))
	}
	for id, lvl := range h.Skills {
		if sk := p.HeroSkill(id); sk != nil {
			m.AddAll(sk.Modifiers, float64(lvl))
		}
	}
	for _, it := range items {
		if it.HeroID != nil && *it.HeroID == h.ID {
			if def := p.HeroItem(it.Def); def != nil {
				m.AddAll(def.Modifiers, 1)
			}
		}
	}
	return m
}

func (c *pctx) heroByID(id int64, forUpdate bool) (*Hero, *rules.Mods, error) {
	q := `SELECT ` + heroCols + ` FROM heroes WHERE id=$1`
	if forUpdate {
		q += ` FOR UPDATE`
	}
	h, err := scanHero(c.tx.QueryRow(c.ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, notFound("hero")
	}
	if err != nil {
		return nil, nil, err
	}
	items, err := c.itemsOf(h.PlayerID)
	if err != nil {
		return nil, nil, err
	}
	h.Items = filterItems(items, h.ID)
	return h, heroMods(c.p, h, items), nil
}

func filterItems(items []Item, heroID int64) []Item {
	out := []Item{}
	for _, it := range items {
		if it.HeroID != nil && *it.HeroID == heroID {
			out = append(out, it)
		}
	}
	return out
}

func (c *pctx) itemsOf(playerID int64) ([]Item, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT id, item_def, hero_id, slot, source FROM items WHERE player_id=$1 ORDER BY id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Def, &it.HeroID, &it.Slot, &it.Source); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// homeHero returns an idle, unwounded hero stationed in the city (defends it).
func (c *pctx) homeHero(cityID int64) (*Hero, *rules.Mods, error) {
	var id int64
	err := c.tx.QueryRow(c.ctx, `SELECT id FROM heroes WHERE home_city_id=$1 AND status='idle' AND (wounded_until IS NULL OR wounded_until <= $2)
		ORDER BY level DESC, id LIMIT 1`, cityID, c.now).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return c.heroByID(id, true)
}

func (c *pctx) enrichHero(h *Hero, items []Item) {
	hc := c.p.Heroes
	spent := 0
	for _, l := range h.Skills {
		spent += l
	}
	h.SkillPoints = 1 + (h.Level-1)*hc.SkillPointsPerLevel - spent
	if h.Level < hc.MaxLevel {
		h.NextLevelXP = rules.XPForLevel(hc.XPBase, hc.XPGrowth, h.Level+1)
	}
	h.Items = filterItems(items, h.ID)
	h.Modifiers = heroMods(c.p, h, items).Snapshot()
}

type HeroesView struct {
	Heroes    []*Hero      `json:"heroes"`
	Items     []Item       `json:"inventory"`
	MaxHeroes int          `json:"maxHeroes"`
	HireCost  pack.Amounts `json:"hireCost"`
}

func (c *pctx) maxHeroes() int {
	b := c.p.Params.HeroBuilding
	lvl := 0
	for _, ci := range c.cities {
		lvl = max(lvl, ci.Buildings[b])
	}
	return min(c.p.Params.MaxHeroes, lvl*c.p.Params.HeroesPerBuildingLevel)
}

func (c *pctx) hireCost(n int) pack.Amounts {
	return rules.ScaleAmounts(c.p.Params.HeroHireCost, int64(n+1))
}

func (c *pctx) Heroes() (*HeroesView, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT `+heroCols+` FROM heroes WHERE player_id=$1 ORDER BY id`, c.pl.ID)
	if err != nil {
		return nil, err
	}
	var hs []*Hero
	for rows.Next() {
		h, err := scanHero(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		hs = append(hs, h)
	}
	rows.Close()
	items, err := c.itemsOf(c.pl.ID)
	if err != nil {
		return nil, err
	}
	for _, h := range hs {
		c.enrichHero(h, items)
	}
	if hs == nil {
		hs = []*Hero{}
	}
	return &HeroesView{Heroes: hs, Items: items, MaxHeroes: c.maxHeroes(), HireCost: c.hireCost(len(hs))}, nil
}

// HireHero recruits a hero of a class in a city with the hero building.
func (c *pctx) HireHero(ci *City, class, name string) (*Hero, error) {
	if c.p.HeroClass(class) == nil {
		return nil, invalid("unknown hero class %q", class)
	}
	if !validName(name) {
		return nil, invalid("invalid hero name")
	}
	if ci.Buildings[c.p.Params.HeroBuilding] < 1 {
		return nil, rule("requirements_not_met", "requires %s in this city", c.p.Params.HeroBuilding)
	}
	var n int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM heroes WHERE player_id=$1`, c.pl.ID).Scan(&n); err != nil {
		return nil, err
	}
	if n >= c.maxHeroes() {
		return nil, rule("hero_limit", "hero limit reached (%d)", c.maxHeroes())
	}
	if err := c.spend(ci, c.hireCost(n), "hero_hire", class); err != nil {
		return nil, err
	}
	h, err := scanHero(c.tx.QueryRow(c.ctx, `INSERT INTO heroes(world_id, player_id, home_city_id, name, class_id) VALUES ($1,$2,$3,$4,$5) RETURNING `+heroCols,
		c.pl.WorldID, c.pl.ID, ci.ID, name, class))
	if err != nil {
		return nil, err
	}
	c.stat("hero_hired", 1)
	c.enrichHero(h, nil)
	return h, nil
}

func (c *pctx) ownHero(id int64) (*Hero, error) {
	h, _, err := c.heroByID(id, true)
	if err != nil {
		return nil, err
	}
	if h.PlayerID != c.pl.ID {
		return nil, notFound("hero")
	}
	return h, nil
}

// LearnSkill spends one skill point.
func (c *pctx) LearnSkill(heroID int64, skill string) (*Hero, error) {
	h, err := c.ownHero(heroID)
	if err != nil {
		return nil, err
	}
	sk := c.p.HeroSkill(skill)
	if sk == nil {
		return nil, invalid("unknown skill %q", skill)
	}
	if len(sk.Classes) > 0 && !pack.Allowed(sk.Classes, h.Class) {
		return nil, rule("class_forbidden", "skill not available for this class")
	}
	if h.Skills[skill] >= sk.MaxLevel {
		return nil, rule("max_level", "skill at maximum level")
	}
	c.enrichHero(h, nil)
	if h.SkillPoints < 1 {
		return nil, rule("no_skill_points", "no skill points available")
	}
	h.Skills[skill]++
	if _, err := c.tx.Exec(c.ctx, `UPDATE heroes SET skills=$2 WHERE id=$1`, h.ID, h.Skills); err != nil {
		return nil, err
	}
	items, err := c.itemsOf(c.pl.ID)
	if err != nil {
		return nil, err
	}
	c.enrichHero(h, items)
	return h, nil
}

// Equip puts an inventory item on an idle hero, replacing the slot occupant.
func (c *pctx) Equip(heroID, itemID int64) error {
	h, err := c.ownHero(heroID)
	if err != nil {
		return err
	}
	if h.Status != "idle" {
		return rule("hero_busy", "hero is on a mission")
	}
	var owner int64
	var def string
	var cur *int64
	if err := c.tx.QueryRow(c.ctx, `SELECT player_id, item_def, hero_id FROM items WHERE id=$1 FOR UPDATE`, itemID).Scan(&owner, &def, &cur); err != nil || owner != c.pl.ID {
		return notFound("item")
	}
	idef := c.p.HeroItem(def)
	if idef == nil {
		return rule("item_obsolete", "item no longer exists in this world's rules")
	}
	if _, err := c.tx.Exec(c.ctx, `UPDATE items SET hero_id=NULL, slot=NULL WHERE hero_id=$1 AND slot=$2`, h.ID, idef.Slot); err != nil {
		return err
	}
	_, err = c.tx.Exec(c.ctx, `UPDATE items SET hero_id=$2, slot=$3 WHERE id=$1`, itemID, h.ID, idef.Slot)
	return err
}

// Unequip removes the item from a slot of an idle hero.
func (c *pctx) Unequip(heroID int64, slot string) error {
	h, err := c.ownHero(heroID)
	if err != nil {
		return err
	}
	if h.Status != "idle" {
		return rule("hero_busy", "hero is on a mission")
	}
	_, err = c.tx.Exec(c.ctx, `UPDATE items SET hero_id=NULL, slot=NULL WHERE hero_id=$1 AND slot=$2`, h.ID, slot)
	return err
}

// gainXP adds experience and recomputes level (capped).
func (c *pctx) gainXP(h *Hero, xp int64) {
	hc := c.p.Heroes
	h.XP += max(0, xp)
	h.Level = rules.LevelForXP(hc.XPBase, hc.XPGrowth, hc.MaxLevel, h.XP)
}

func (c *pctx) saveHero(h *Hero) error {
	_, err := c.tx.Exec(c.ctx, `UPDATE heroes SET level=$2, xp=$3, status=$4, wounded_until=$5, skills=$6 WHERE id=$1`,
		h.ID, h.Level, h.XP, h.Status, h.WoundedUntil, h.Skills)
	return err
}

func (c *pctx) grantItem(playerID int64, def, source string) error {
	if c.p.HeroItem(def) == nil {
		return fmt.Errorf("unknown item %q", def)
	}
	_, err := c.tx.Exec(c.ctx, `INSERT INTO items(world_id, player_id, item_def, source) VALUES ($1,$2,$3,$4)`, c.pl.WorldID, playerID, def, source)
	return err
}
