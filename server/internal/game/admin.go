package game

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"realmforge/server/internal/pack"
)

func (s *Service) audit(ctx context.Context, q querier, adminID int64, action, target string, payload any) error {
	_, err := q.Exec(ctx, `INSERT INTO admin_audit(admin_id, action, target, payload) VALUES ($1,$2,$3,$4)`, adminID, action, target, payload)
	return err
}

// AdminGrant adds (or with negative values removes) resources to a city with
// a mandatory reason; it goes through the normal settlement path and is audited.
func (s *Service) AdminGrant(ctx context.Context, adminID, cityID int64, amounts pack.Amounts, reason string) error {
	if len(reason) < 5 {
		return invalid("a reason of at least 5 characters is required")
	}
	return s.InTx(ctx, func(tx pgx.Tx) error {
		var pid int64
		if err := tx.QueryRow(ctx, `SELECT player_id FROM cities WHERE id=$1`, cityID).Scan(&pid); err != nil {
			return notFound("city")
		}
		c, err := s.lockPlayer(ctx, tx, pid, s.Clock.Now())
		if err != nil {
			return err
		}
		ci := c.city(cityID)
		pos, neg := pack.Amounts{}, pack.Amounts{}
		for r, n := range amounts {
			if c.p.Resource(r) == nil {
				return invalid("unknown resource %q", r)
			}
			if n > 0 {
				pos[r] = n
			} else if n < 0 {
				neg[r] = -n
			}
		}
		if len(neg) > 0 {
			if err := c.spend(ci, neg, "admin", reason); err != nil {
				return err
			}
		}
		if len(pos) > 0 {
			if _, err := c.grant(ci, pos, "admin", reason); err != nil {
				return err
			}
		}
		if err := s.audit(ctx, tx, adminID, "grant", fmt.Sprintf("city:%d", cityID), map[string]any{"amounts": amounts, "reason": reason}); err != nil {
			return err
		}
		return c.save()
	})
}

type AdminPlayer struct {
	Player *Player     `json:"player"`
	Cities []*CityView `json:"cities"`
	Events []EventRow  `json:"pendingEvents"`
	Ledger []LedgerRow `json:"recentLedger"`
}

type EventRow struct {
	ID        int64     `json:"id"`
	PlayerID  int64     `json:"playerId"`
	Kind      string    `json:"kind"`
	RefID     int64     `json:"refId"`
	DueAt     time.Time `json:"dueAt"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
	LastError *string   `json:"lastError"`
}

type LedgerRow struct {
	ID     int64          `json:"id"`
	CityID int64          `json:"cityId"`
	Reason string         `json:"reason"`
	Ref    string         `json:"ref"`
	Delta  map[string]any `json:"delta"`
	At     time.Time      `json:"at"`
}

// AdminInspect settles and returns a player's full state (read-mostly; audited).
func (s *Service) AdminInspect(ctx context.Context, adminID, playerID int64) (*AdminPlayer, error) {
	var out *AdminPlayer
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		c, err := s.lockPlayer(ctx, tx, playerID, s.Clock.Now())
		if err != nil {
			return err
		}
		out = &AdminPlayer{Player: c.pl}
		for _, ci := range c.cities {
			v, err := c.View(ci)
			if err != nil {
				return err
			}
			out.Cities = append(out.Cities, v)
		}
		out.Events, err = s.events(ctx, tx, `player_id=$1 AND status='pending'`, playerID)
		if err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `SELECT id, city_id, reason, ref, delta, at FROM resource_ledger WHERE player_id=$1 ORDER BY id DESC LIMIT 100`, playerID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var l LedgerRow
			if err := rows.Scan(&l.ID, &l.CityID, &l.Reason, &l.Ref, &l.Delta, &l.At); err != nil {
				rows.Close()
				return err
			}
			out.Ledger = append(out.Ledger, l)
		}
		rows.Close()
		if err := s.audit(ctx, tx, adminID, "inspect", fmt.Sprintf("player:%d", playerID), map[string]any{}); err != nil {
			return err
		}
		return c.save()
	})
	return out, err
}

func (s *Service) events(ctx context.Context, q querier, where string, args ...any) ([]EventRow, error) {
	rows, err := q.Query(ctx, `SELECT id, player_id, kind, ref_id, due_at, status, attempts, last_error FROM game_events WHERE `+where+` ORDER BY due_at, id LIMIT 500`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EventRow{}
	for rows.Next() {
		var e EventRow
		if err := rows.Scan(&e.ID, &e.PlayerID, &e.Kind, &e.RefID, &e.DueAt, &e.Status, &e.Attempts, &e.LastError); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Service) FailedEvents(ctx context.Context) ([]EventRow, error) {
	return s.events(ctx, s.DB, `status='failed'`)
}

// RetryEvent re-queues a parked event after the underlying bug was fixed.
func (s *Service) RetryEvent(ctx context.Context, adminID, eventID int64) error {
	return s.InTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE game_events SET status='pending', attempts=0, last_error=NULL WHERE id=$1 AND status='failed'`, eventID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return notFound("failed event")
		}
		return s.audit(ctx, tx, adminID, "retry_event", fmt.Sprintf("event:%d", eventID), map[string]any{})
	})
}

func (s *Service) Ban(ctx context.Context, adminID, accountID int64, until time.Time, reason string) error {
	return s.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE accounts SET banned_until=$2 WHERE id=$1`, accountID, until); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sessions SET revoked_at=now() WHERE account_id=$1 AND revoked_at IS NULL`, accountID); err != nil {
			return err
		}
		return s.audit(ctx, tx, adminID, "ban", fmt.Sprintf("account:%d", accountID), map[string]any{"until": until, "reason": reason})
	})
}

type AuditRow struct {
	ID      int64          `json:"id"`
	AdminID int64          `json:"adminId"`
	Action  string         `json:"action"`
	Target  string         `json:"target"`
	Payload map[string]any `json:"payload"`
	At      time.Time      `json:"at"`
}

func (s *Service) Audit(ctx context.Context, limit int) ([]AuditRow, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, admin_id, action, target, payload, at FROM admin_audit ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditRow{}
	for rows.Next() {
		var a AuditRow
		if err := rows.Scan(&a.ID, &a.AdminID, &a.Action, &a.Target, &a.Payload, &a.At); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Service) ReportedMessages(ctx context.Context) ([]Message, error) {
	rows, err := s.DB.Query(ctx, `SELECT m.id, f.name, m.from_player, t.name, m.alliance_id, m.subject, m.body, m.created_at, m.read_at
		FROM messages m JOIN players f ON f.id=m.from_player LEFT JOIN players t ON t.id=m.to_player WHERE m.reported_at IS NOT NULL ORDER BY m.reported_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.From, &m.FromID, &m.To, &m.AllianceID, &m.Subject, &m.Body, &m.CreatedAt, &m.ReadAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---------- Game Pack migration of a running world ----------

// PackMigrationReport describes a (dry-run or applied) pack migration.
type PackMigrationReport struct {
	FromVersion string            `json:"fromVersion"`
	ToVersion   string            `json:"toVersion"`
	Renames     map[string]string `json:"renames"`
	Removed     map[string]string `json:"removed"`
	Missing     []string          `json:"missing"` // ids in use that the new pack lacks and no rule maps
	Rows        map[string]int64  `json:"rowsChanged"`
	Applied     bool              `json:"applied"`
}

type idUse struct{ kind, id string }

// MigrateWorldPack moves a world to a new pack version. IDs used by the
// world's state must exist in the new pack or be mapped by the new pack's
// `migrations` entry whose fromVersion equals the world's current version.
// With dryRun the report is computed without changes. Worlds never change
// pack implicitly (GAME_PACK_SPEC §6).
func (s *Service) MigrateWorldPack(ctx context.Context, adminID, worldID int64, next *pack.Loaded, dryRun bool) (*PackMigrationReport, error) {
	var rep *PackMigrationReport
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var curID, curVer, curHash string
		if err := tx.QueryRow(ctx, `SELECT pack_id, pack_version, pack_hash FROM worlds WHERE id=$1 FOR UPDATE`, worldID).Scan(&curID, &curVer, &curHash); err != nil {
			return notFound("world")
		}
		if curID != next.ID {
			return rule("pack_mismatch", "world uses pack %q, not %q", curID, next.ID)
		}
		if curHash == next.Hash {
			return rule("same_pack", "world already uses this exact pack")
		}
		rep = &PackMigrationReport{FromVersion: curVer, ToVersion: next.Version, Renames: map[string]string{}, Removed: map[string]string{}, Rows: map[string]int64{}}
		for _, m := range next.Migrations {
			if m.FromVersion == curVer {
				for k, v := range m.Renames {
					rep.Renames[k] = v
				}
				for k, v := range m.Removed {
					rep.Removed[k] = v
				}
			}
		}
		used, err := usedIDs(ctx, tx, worldID)
		if err != nil {
			return err
		}
		exists := func(u idUse) bool {
			switch u.kind {
			case "building":
				return next.Building(u.id) != nil
			case "unit":
				return next.Unit(u.id) != nil
			case "tech":
				return next.Tech(u.id) != nil
			case "race":
				return next.Race(u.id) != nil
			case "terrain":
				return next.Terrain(u.id) != nil
			case "neutral":
				return next.Neutral(u.id) != nil
			case "item":
				return next.HeroItem(u.id) != nil
			case "heroClass":
				return next.HeroClass(u.id) != nil
			case "heroSkill":
				return next.HeroSkill(u.id) != nil
			case "resource":
				return next.Resource(u.id) != nil
			}
			return false
		}
		for _, u := range used {
			if exists(u) {
				continue
			}
			if nw, ok := rep.Renames[u.id]; ok && exists(idUse{u.kind, nw}) {
				continue
			}
			if repl, ok := rep.Removed[u.id]; ok && (repl == "" || exists(idUse{u.kind, repl})) {
				if repl == "" && (u.kind == "race" || u.kind == "terrain" || u.kind == "heroClass") {
					rep.Missing = append(rep.Missing, u.kind+":"+u.id+" (cannot be dropped)")
				}
				continue
			}
			rep.Missing = append(rep.Missing, u.kind+":"+u.id)
		}
		sort.Strings(rep.Missing)
		if len(rep.Missing) > 0 || dryRun {
			return nil
		}
		mapping := map[string]string{}
		for k, v := range rep.Removed {
			mapping[k] = v
		}
		for k, v := range rep.Renames {
			mapping[k] = v
		}
		if len(mapping) > 0 {
			if err := applyIDMapping(ctx, tx, worldID, mapping, rep.Rows); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE worlds SET pack_version=$2, pack_hash=$3, pack_json=$4 WHERE id=$1`, worldID, next.Version, next.Hash, next.JSON()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO world_pack_history(world_id, from_version, to_version, from_hash, to_hash, admin_id, report) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			worldID, curVer, next.Version, curHash, next.Hash, adminID, rep); err != nil {
			return err
		}
		rep.Applied = true
		return s.audit(ctx, tx, adminID, "migrate_pack", fmt.Sprintf("world:%d", worldID), rep)
	})
	if err == nil && rep != nil && rep.Applied {
		s.packs.Delete(worldID)
	}
	return rep, err
}

func usedIDs(ctx context.Context, tx pgx.Tx, worldID int64) ([]idUse, error) {
	seen := map[idUse]bool{}
	add := func(kind string, sql string) error {
		rows, err := tx.Query(ctx, sql, worldID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			seen[idUse{kind, id}] = true
		}
		return rows.Err()
	}
	qs := []struct{ kind, sql string }{
		{"building", `SELECT DISTINCT jsonb_object_keys(buildings) FROM cities WHERE world_id=$1`},
		{"unit", `SELECT DISTINCT jsonb_object_keys(units) FROM cities WHERE world_id=$1`},
		{"unit", `SELECT DISTINCT jsonb_object_keys(units) FROM missions WHERE world_id=$1 AND status <> 'done'`},
		{"resource", `SELECT DISTINCT jsonb_object_keys(resources) FROM cities WHERE world_id=$1`},
		{"tech", `SELECT DISTINCT jsonb_object_keys(techs) FROM players WHERE world_id=$1`},
		{"race", `SELECT DISTINCT race_id FROM players WHERE world_id=$1`},
		{"terrain", `SELECT DISTINCT terrain FROM tiles WHERE world_id=$1`},
		{"neutral", `SELECT DISTINCT neutral_id FROM tiles WHERE world_id=$1 AND neutral_id IS NOT NULL`},
		{"item", `SELECT DISTINCT item_def FROM items WHERE world_id=$1`},
		{"heroClass", `SELECT DISTINCT class_id FROM heroes WHERE world_id=$1`},
		{"heroSkill", `SELECT DISTINCT jsonb_object_keys(skills) FROM heroes WHERE world_id=$1`},
		{"building", `SELECT DISTINCT item_id FROM jobs WHERE world_id=$1 AND status='queued' AND kind='build'`},
		{"tech", `SELECT DISTINCT item_id FROM jobs WHERE world_id=$1 AND status='queued' AND kind='research'`},
		{"unit", `SELECT DISTINCT item_id FROM jobs WHERE world_id=$1 AND status='queued' AND kind='train'`},
	}
	for _, q := range qs {
		if err := add(q.kind, q.sql); err != nil {
			return nil, err
		}
	}
	out := make([]idUse, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].kind+out[i].id < out[j].kind+out[j].id })
	return out, nil
}

// renameKeys rewrites map keys according to mapping ("" drops the key; merged values add up).
func renameKeys[V int | int64](m map[string]V, mapping map[string]string) (map[string]V, bool) {
	out := map[string]V{}
	changed := false
	for k, v := range m {
		nk, ok := mapping[k]
		if !ok {
			out[k] += v
			continue
		}
		changed = true
		if nk == "" {
			continue
		}
		if _, isInt := any(v).(int); isInt {
			// levels: keep the max when merging
			if any(out[nk]).(int) < any(v).(int) {
				out[nk] = v
			}
			continue
		}
		out[nk] += v
	}
	return out, changed
}

func applyIDMapping(ctx context.Context, tx pgx.Tx, worldID int64, mapping map[string]string, counts map[string]int64) error {
	type cityRow struct {
		id        int64
		buildings map[string]int
		units     map[string]int64
		res       map[string]int64
	}
	rows, err := tx.Query(ctx, `SELECT id, buildings, units, resources FROM cities WHERE world_id=$1 ORDER BY id FOR UPDATE`, worldID)
	if err != nil {
		return err
	}
	var cities []cityRow
	for rows.Next() {
		var c cityRow
		if err := rows.Scan(&c.id, &c.buildings, &c.units, &c.res); err != nil {
			rows.Close()
			return err
		}
		cities = append(cities, c)
	}
	rows.Close()
	for _, c := range cities {
		b, ch1 := renameKeys(c.buildings, mapping)
		u, ch2 := renameKeys(c.units, mapping)
		r, ch3 := renameKeys(c.res, mapping)
		if ch1 || ch2 || ch3 {
			if _, err := tx.Exec(ctx, `UPDATE cities SET buildings=$2, units=$3, resources=$4 WHERE id=$1`, c.id, b, u, r); err != nil {
				return err
			}
			counts["cities"]++
		}
	}
	rows, err = tx.Query(ctx, `SELECT id, techs FROM players WHERE world_id=$1 ORDER BY id FOR UPDATE`, worldID)
	if err != nil {
		return err
	}
	type prow struct {
		id    int64
		techs map[string]int
	}
	var ps []prow
	for rows.Next() {
		var p prow
		if err := rows.Scan(&p.id, &p.techs); err != nil {
			rows.Close()
			return err
		}
		ps = append(ps, p)
	}
	rows.Close()
	for _, p := range ps {
		if t, ch := renameKeys(p.techs, mapping); ch {
			if _, err := tx.Exec(ctx, `UPDATE players SET techs=$2 WHERE id=$1`, p.id, t); err != nil {
				return err
			}
			counts["players"]++
		}
	}
	rows, err = tx.Query(ctx, `SELECT id, units, upkeep_units FROM missions WHERE world_id=$1 AND status <> 'done' ORDER BY id FOR UPDATE`, worldID)
	if err != nil {
		return err
	}
	type mrow struct {
		id    int64
		u, uk map[string]int64
	}
	var ms []mrow
	for rows.Next() {
		var m mrow
		if err := rows.Scan(&m.id, &m.u, &m.uk); err != nil {
			rows.Close()
			return err
		}
		ms = append(ms, m)
	}
	rows.Close()
	for _, m := range ms {
		u, ch1 := renameKeys(m.u, mapping)
		uk, ch2 := renameKeys(m.uk, mapping)
		if ch1 || ch2 {
			if _, err := tx.Exec(ctx, `UPDATE missions SET units=$2, upkeep_units=$3 WHERE id=$1`, m.id, u, uk); err != nil {
				return err
			}
			counts["missions"]++
		}
	}
	for old, nw := range mapping {
		var simple = []struct{ name, sql string }{
			{"players.race", `UPDATE players SET race_id=$3 WHERE world_id=$1 AND race_id=$2`},
			{"tiles.terrain", `UPDATE tiles SET terrain=$3 WHERE world_id=$1 AND terrain=$2`},
			{"heroes.class", `UPDATE heroes SET class_id=$3 WHERE world_id=$1 AND class_id=$2`},
			{"jobs", `UPDATE jobs SET item_id=$3 WHERE world_id=$1 AND item_id=$2 AND status='queued'`},
		}
		if nw == "" {
			simple = []struct{ name, sql string }{
				{"tiles.neutral", `UPDATE tiles SET neutral_id=NULL, neutral_level=NULL WHERE world_id=$1 AND neutral_id=$2 AND $3=''`},
				{"items", `DELETE FROM items WHERE world_id=$1 AND item_def=$2 AND $3=''`},
			}
		} else {
			simple = append(simple,
				struct{ name, sql string }{"tiles.neutral", `UPDATE tiles SET neutral_id=$3 WHERE world_id=$1 AND neutral_id=$2`},
				struct{ name, sql string }{"items", `UPDATE items SET item_def=$3 WHERE world_id=$1 AND item_def=$2`})
		}
		for _, q := range simple {
			tag, err := tx.Exec(ctx, q.sql, worldID, old, nw)
			if err != nil {
				return err
			}
			counts[q.name] += tag.RowsAffected()
		}
	}
	return nil
}

// ---------- Player overview ----------

type Overview struct {
	Player        *Player       `json:"player"`
	Cities        []CitySummary `json:"cities"`
	Unread        int           `json:"unreadNotifications"`
	UnreadReports int           `json:"unreadReports"`
	QuestsReady   int           `json:"questsReady"`
	ServerTime    time.Time     `json:"serverTime"`
	AllianceTag   *string       `json:"allianceTag"`
	CityCap       int           `json:"cityCap"`
	TerritoryCap  int           `json:"territoryCap"`
	Territories   int           `json:"territories"`
	PackVersion   string        `json:"packVersion"`
}

type CitySummary struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	X         int          `json:"x"`
	Y         int          `json:"y"`
	IsCapital bool         `json:"isCapital"`
	Resources pack.Amounts `json:"resources"`
}

func (c *pctx) Overview() (*Overview, error) {
	o := &Overview{Player: c.pl, ServerTime: c.now, CityCap: c.cityCap(), TerritoryCap: c.territoryCap(), PackVersion: c.p.Version}
	for _, ci := range c.cities {
		o.Cities = append(o.Cities, CitySummary{ID: ci.ID, Name: ci.Name, X: ci.X, Y: ci.Y, IsCapital: ci.IsCapital, Resources: wholeMilli(ci.Res)})
	}
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM notifications WHERE player_id=$1 AND read_at IS NULL`, c.pl.ID).Scan(&o.Unread); err != nil {
		return nil, err
	}
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM report_recipients WHERE player_id=$1 AND read_at IS NULL`, c.pl.ID).Scan(&o.UnreadReports); err != nil {
		return nil, err
	}
	qs, err := c.Quests()
	if err != nil {
		return nil, err
	}
	for _, q := range qs {
		if q.Status == "complete" {
			o.QuestsReady++
		}
	}
	if c.pl.AllianceID != nil {
		var tag string
		if err := c.tx.QueryRow(c.ctx, `SELECT tag FROM alliances WHERE id=$1`, *c.pl.AllianceID).Scan(&tag); err == nil {
			o.AllianceTag = &tag
		}
	}
	o.Territories, err = c.territoryCount()
	return o, err
}
