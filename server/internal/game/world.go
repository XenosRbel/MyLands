package game

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
	"realmforge/server/internal/rules"
)

type World struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	PackID      string    `json:"packId"`
	PackVersion string    `json:"packVersion"`
	Speed       float64   `json:"speed"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	Players     int64     `json:"players"`
	Joined      bool      `json:"joined"`
	PlayerID    *int64    `json:"playerId,omitempty"`
}

var reSlug = regexp.MustCompile(`^[a-z0-9-]{2,40}$`)

// CreateWorld pins a pack snapshot and generates the map deterministically.
func (s *Service) CreateWorld(ctx context.Context, slug, name string, p *pack.Loaded, seed int64, speed float64) (int64, error) {
	if !reSlug.MatchString(slug) {
		return 0, invalid("slug must match %s", reSlug)
	}
	if speed <= 0 || speed > 1000 {
		return 0, invalid("speed must be in (0, 1000]")
	}
	var id int64
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `INSERT INTO worlds(slug, name, pack_id, pack_version, pack_hash, pack_json, seed, speed)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, slug, name, p.ID, p.Version, p.Hash, p.JSON(), seed, speed).Scan(&id)
		if err != nil {
			if isUnique(err) {
				return conflict("world_exists", "world %q already exists", slug)
			}
			return err
		}
		tiles := rules.GenerateMap(uint64(seed), p)
		rows := make([][]any, 0, len(tiles))
		for _, t := range tiles {
			var nid any
			var nl any
			if t.Neutral != "" {
				nid, nl = t.Neutral, t.NeutralLevel
			}
			rows = append(rows, []any{id, t.X, t.Y, t.Terrain, nid, nl})
		}
		_, err = tx.CopyFrom(ctx, pgx.Identifier{"tiles"}, []string{"world_id", "x", "y", "terrain", "neutral_id", "neutral_level"}, pgx.CopyFromRows(rows))
		return err
	})
	return id, err
}

// EnsureWorld creates the world if the slug does not exist yet.
func (s *Service) EnsureWorld(ctx context.Context, slug, name string, p *pack.Loaded, seed int64, speed float64) (int64, error) {
	var id int64
	err := s.DB.QueryRow(ctx, `SELECT id FROM worlds WHERE slug=$1`, slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	return s.CreateWorld(ctx, slug, name, p, seed, speed)
}

func (s *Service) ListWorlds(ctx context.Context, accountID int64) ([]World, error) {
	rows, err := s.DB.Query(ctx, `SELECT w.id, w.slug, w.name, w.pack_id, w.pack_version, w.speed, w.status, w.created_at,
		(SELECT count(*) FROM players p WHERE p.world_id=w.id),
		(SELECT p.id FROM players p WHERE p.world_id=w.id AND p.account_id=$1)
		FROM worlds w WHERE w.status <> 'archived' ORDER BY w.id`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []World{}
	for rows.Next() {
		var w World
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name, &w.PackID, &w.PackVersion, &w.Speed, &w.Status, &w.CreatedAt, &w.Players, &w.PlayerID); err != nil {
			return nil, err
		}
		w.Joined = w.PlayerID != nil
		out = append(out, w)
	}
	return out, rows.Err()
}

var reName = regexp.MustCompile(`^[\p{L}\p{N}_ .'-]{3,24}$`)

func validName(n string) bool {
	return reName.MatchString(n) && strings.TrimSpace(n) == n
}

// Join creates the player's account-in-world and capital city.
func (s *Service) Join(ctx context.Context, accountID, worldID int64, name, race string) (int64, error) {
	if !validName(name) {
		return 0, invalid("name must be 3-24 letters, digits, spaces, _ . ' -")
	}
	var playerID int64
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		p, err := s.Pack(ctx, tx, worldID)
		if err != nil {
			return err
		}
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM worlds WHERE id=$1`, worldID).Scan(&status); err != nil {
			return err
		}
		if status != "open" {
			return rule("world_closed", "world is not open for new players")
		}
		if p.Race(race) == nil {
			return invalid("unknown race %q", race)
		}
		now := s.Clock.Now()
		prot := now.Add(time.Duration(p.Params.NewbieProtectionHours) * time.Hour)
		err = tx.QueryRow(ctx, `INSERT INTO players(world_id, account_id, name, race_id, protected_until, created_at)
			VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, worldID, accountID, name, race, prot, now).Scan(&playerID)
		if err != nil {
			if isUnique(err) {
				return conflict("name_taken", "you already play in this world or the name is taken")
			}
			return err
		}
		x, y, err := s.findSpawn(ctx, tx, worldID, p)
		if err != nil {
			return err
		}
		res := map[string]int64{}
		for r, n := range p.Start.Resources {
			res[r] = n * rules.Milli
		}
		units := map[string]int64{}
		for u, n := range p.Start.Units {
			units[u] = n
		}
		buildings := map[string]int{}
		for b, l := range p.Start.Buildings {
			buildings[b] = l
		}
		var cityID int64
		err = tx.QueryRow(ctx, `INSERT INTO cities(world_id, player_id, name, x, y, is_capital, resources, population, econ_updated_at, buildings, units)
			VALUES ($1,$2,$3,$4,$5,true,$6,$7,$8,$9,$10) RETURNING id`,
			worldID, playerID, name, x, y, res, p.Start.Population, now, buildings, units).Scan(&cityID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE tiles SET city_id=$1, owner_player_id=NULL WHERE world_id=$2 AND x=$3 AND y=$4`, cityID, worldID, x, y); err != nil {
			return err
		}
		c, err := s.lockPlayer(ctx, tx, playerID, now)
		if err != nil {
			return err
		}
		if err := notify(ctx, tx, playerID, "welcome", map[string]any{"cityId": cityID}); err != nil {
			return err
		}
		return c.save()
	})
	return playerID, err
}

// findSpawn picks a free passable tile without a neutral, at least 3 tiles
// (Chebyshev) from any city, closest to the centre with jitter (PROPOSED).
func (s *Service) findSpawn(ctx context.Context, tx pgx.Tx, worldID int64, p *pack.Loaded) (int, int, error) {
	var passable []string
	for _, t := range p.Terrains {
		if t.Passable {
			passable = append(passable, t.ID)
		}
	}
	cx, cy := float64(p.Params.MapWidth-1)/2, float64(p.Params.MapHeight-1)/2
	var x, y int
	err := tx.QueryRow(ctx, `SELECT x, y FROM tiles t
		WHERE world_id=$1 AND city_id IS NULL AND neutral_id IS NULL AND owner_player_id IS NULL AND terrain = ANY($2)
		  AND NOT EXISTS (SELECT 1 FROM cities c WHERE c.world_id=t.world_id AND abs(c.x-t.x) <= 2 AND abs(c.y-t.y) <= 2)
		ORDER BY (x-$3)*(x-$3) + (y-$4)*(y-$4) + random()*20 LIMIT 1 FOR UPDATE SKIP LOCKED`,
		worldID, passable, cx, cy).Scan(&x, &y)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, rule("world_full", "no free location left in this world")
	}
	return x, y, err
}

// playerFor returns the player id of an account in a world.
func (s *Service) playerFor(ctx context.Context, q querier, accountID, worldID int64) (int64, error) {
	var id int64
	err := q.QueryRow(ctx, `SELECT id FROM players WHERE account_id=$1 AND world_id=$2`, accountID, worldID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, forbidden("you have not joined this world")
	}
	return id, err
}

// cityOwner returns (playerID, worldID) of a city, verifying the account owns it.
func (s *Service) cityOwner(ctx context.Context, q querier, accountID, cityID int64) (int64, int64, error) {
	var pid, wid, acc int64
	err := q.QueryRow(ctx, `SELECT p.id, p.world_id, p.account_id FROM cities c JOIN players p ON p.id=c.player_id WHERE c.id=$1`, cityID).Scan(&pid, &wid, &acc)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, notFound("city")
	}
	if err != nil {
		return 0, 0, err
	}
	if acc != accountID {
		return 0, 0, notFound("city") // do not reveal existence of other players' cities by id
	}
	return pid, wid, nil
}

// WithPlayer locks and settles the account's player in a world and runs fn;
// state is saved afterwards. This is the standard path for player commands.
func (s *Service) WithPlayer(ctx context.Context, tx pgx.Tx, accountID, worldID int64, fn func(c *pctx) (any, error)) (any, error) {
	pid, err := s.playerFor(ctx, tx, accountID, worldID)
	if err != nil {
		return nil, err
	}
	c, err := s.lockPlayer(ctx, tx, pid, s.Clock.Now())
	if err != nil {
		return nil, err
	}
	out, err := fn(c)
	if err != nil {
		return nil, err
	}
	return out, c.save()
}

// WithCity is WithPlayer addressed by a city id owned by the account.
func (s *Service) WithCity(ctx context.Context, tx pgx.Tx, accountID, cityID int64, fn func(c *pctx, ci *City) (any, error)) (any, error) {
	pid, _, err := s.cityOwner(ctx, tx, accountID, cityID)
	if err != nil {
		return nil, err
	}
	c, err := s.lockPlayer(ctx, tx, pid, s.Clock.Now())
	if err != nil {
		return nil, err
	}
	ci := c.city(cityID)
	if ci == nil {
		return nil, notFound("city")
	}
	out, err := fn(c, ci)
	if err != nil {
		return nil, err
	}
	return out, c.save()
}

func isUnique(err error) bool {
	return pgCode(err) == "23505"
}
