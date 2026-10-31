package game

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

// ---------- Alliances (GAME_RULES R-SOC-1..8) ----------

type Alliance struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Tag         string           `json:"tag"`
	Description string           `json:"description"`
	CreatedAt   time.Time        `json:"createdAt"`
	Members     []AllianceMember `json:"members,omitempty"`
	Relations   []Relation       `json:"relations,omitempty"`
	MemberCount int              `json:"memberCount"`
	Score       int64            `json:"score"`
}

type AllianceMember struct {
	PlayerID int64  `json:"playerId"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Score    int64  `json:"score"`
}

type Relation struct {
	OtherID  int64  `json:"otherId"`
	OtherTag string `json:"otherTag"`
	Status   string `json:"status"`
}

type Invite struct {
	ID          int64     `json:"id"`
	AllianceID  int64     `json:"allianceId"`
	AllianceTag string    `json:"allianceTag"`
	PlayerID    int64     `json:"playerId"`
	Player      string    `json:"player"`
	CreatedAt   time.Time `json:"createdAt"`
}

var reTag = regexp.MustCompile(`^[A-Za-z0-9]{2,5}$`)

// CreateAlliance founds an alliance; the creation cost is paid from the capital.
func (c *pctx) CreateAlliance(name, tag, desc string) (*Alliance, error) {
	if c.pl.AllianceID != nil {
		return nil, rule("already_member", "leave your current alliance first")
	}
	if !validName(name) || !reTag.MatchString(tag) || utf8.RuneCountInString(desc) > 1000 {
		return nil, invalid("name 3-24 chars, tag 2-5 letters/digits, description ≤ 1000 chars")
	}
	if cost := c.p.Params.AllianceCreateCost; len(cost) > 0 {
		if err := c.spend(c.capital(), cost, "alliance_create", tag); err != nil {
			return nil, err
		}
	}
	a := &Alliance{Name: name, Tag: tag, Description: desc, CreatedAt: c.now}
	err := c.tx.QueryRow(c.ctx, `INSERT INTO alliances(world_id, name, tag, description, created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		c.pl.WorldID, name, tag, desc, c.now).Scan(&a.ID)
	if err != nil {
		if isUnique(err) {
			return nil, conflict("alliance_exists", "alliance name or tag already taken")
		}
		return nil, err
	}
	role := "leader"
	c.pl.AllianceID, c.pl.AllianceRole = &a.ID, &role
	_, err = c.tx.Exec(c.ctx, `UPDATE players SET alliance_id=$2, alliance_role='leader' WHERE id=$1`, c.pl.ID, a.ID)
	return a, err
}

func (c *pctx) requireRole(roles ...string) error {
	if c.pl.AllianceID == nil {
		return rule("not_member", "you are not in an alliance")
	}
	for _, r := range roles {
		if *c.pl.AllianceRole == r {
			return nil
		}
	}
	return forbidden("requires alliance role %s", strings.Join(roles, " or "))
}

func (c *pctx) lockAlliance() error {
	var dis *time.Time
	err := c.tx.QueryRow(c.ctx, `SELECT disbanded_at FROM alliances WHERE id=$1 FOR UPDATE`, *c.pl.AllianceID).Scan(&dis)
	if err != nil {
		return err
	}
	if dis != nil {
		return rule("alliance_disbanded", "alliance was disbanded")
	}
	return nil
}

func (c *pctx) Invite(playerName string) (*Invite, error) {
	if err := c.requireRole("leader", "officer"); err != nil {
		return nil, err
	}
	var target int64
	var tAlliance *int64
	err := c.tx.QueryRow(c.ctx, `SELECT id, alliance_id FROM players WHERE world_id=$1 AND lower(name)=lower($2)`, c.pl.WorldID, playerName).Scan(&target, &tAlliance)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("player")
	}
	if err != nil {
		return nil, err
	}
	if tAlliance != nil {
		return nil, rule("already_member", "player is already in an alliance")
	}
	inv := &Invite{AllianceID: *c.pl.AllianceID, PlayerID: target, CreatedAt: c.now}
	err = c.tx.QueryRow(c.ctx, `INSERT INTO alliance_invites(alliance_id, player_id, invited_by, created_at) VALUES ($1,$2,$3,$4) RETURNING id`,
		inv.AllianceID, target, c.pl.ID, c.now).Scan(&inv.ID)
	if err != nil {
		if isUnique(err) {
			return nil, conflict("already_invited", "player already has a pending invitation")
		}
		return nil, err
	}
	_ = notify(c.ctx, c.tx, target, "alliance_invite", map[string]any{"inviteId": inv.ID, "from": c.pl.Name})
	return inv, nil
}

func (c *pctx) Invites() ([]Invite, error) {
	rows, err := c.tx.Query(c.ctx, `SELECT i.id, i.alliance_id, a.tag, i.player_id, p.name, i.created_at FROM alliance_invites i
		JOIN alliances a ON a.id=i.alliance_id JOIN players p ON p.id=i.player_id
		WHERE i.status='pending' AND a.disbanded_at IS NULL AND (i.player_id=$1 OR i.alliance_id=$2) ORDER BY i.id`, c.pl.ID, c.pl.AllianceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Invite{}
	for rows.Next() {
		var i Invite
		if err := rows.Scan(&i.ID, &i.AllianceID, &i.AllianceTag, &i.PlayerID, &i.Player, &i.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// RespondInvite accepts or declines. Member limit is enforced under the alliance row lock.
func (c *pctx) RespondInvite(id int64, accept bool) error {
	var aid, pid int64
	var status string
	err := c.tx.QueryRow(c.ctx, `SELECT alliance_id, player_id, status FROM alliance_invites WHERE id=$1 FOR UPDATE`, id).Scan(&aid, &pid, &status)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && pid != c.pl.ID) {
		return notFound("invite")
	}
	if err != nil {
		return err
	}
	if status != "pending" {
		return conflict("invite_closed", "invitation is no longer pending")
	}
	if !accept {
		_, err = c.tx.Exec(c.ctx, `UPDATE alliance_invites SET status='declined' WHERE id=$1`, id)
		return err
	}
	if c.pl.AllianceID != nil {
		return rule("already_member", "leave your current alliance first")
	}
	var dis *time.Time
	if err := c.tx.QueryRow(c.ctx, `SELECT disbanded_at FROM alliances WHERE id=$1 FOR UPDATE`, aid).Scan(&dis); err != nil {
		return err
	}
	if dis != nil {
		return rule("alliance_disbanded", "alliance was disbanded")
	}
	var n int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM players WHERE alliance_id=$1`, aid).Scan(&n); err != nil {
		return err
	}
	if n >= c.p.Params.AllianceMaxMembers {
		return rule("alliance_full", "alliance is full (%d members)", c.p.Params.AllianceMaxMembers)
	}
	role := "member"
	c.pl.AllianceID, c.pl.AllianceRole = &aid, &role
	if _, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_id=$2, alliance_role='member' WHERE id=$1`, c.pl.ID, aid); err != nil {
		return err
	}
	_, err = c.tx.Exec(c.ctx, `UPDATE alliance_invites SET status='accepted' WHERE id=$1`, id)
	return err
}

// Leave leaves the alliance. A leader must hand over leadership first unless alone.
func (c *pctx) Leave() error {
	if err := c.requireRole("leader", "officer", "member"); err != nil {
		return err
	}
	if err := c.lockAlliance(); err != nil {
		return err
	}
	var n int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM players WHERE alliance_id=$1`, *c.pl.AllianceID).Scan(&n); err != nil {
		return err
	}
	if *c.pl.AllianceRole == "leader" && n > 1 {
		return rule("leader_must_transfer", "transfer leadership before leaving")
	}
	if n == 1 {
		if _, err := c.tx.Exec(c.ctx, `UPDATE alliances SET disbanded_at=$2 WHERE id=$1`, *c.pl.AllianceID, c.now); err != nil {
			return err
		}
		if _, err := c.tx.Exec(c.ctx, `DELETE FROM alliance_relations WHERE alliance_id=$1 OR other_id=$1`, *c.pl.AllianceID); err != nil {
			return err
		}
	}
	c.pl.AllianceID, c.pl.AllianceRole = nil, nil
	_, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_id=NULL, alliance_role=NULL WHERE id=$1`, c.pl.ID)
	return err
}

// SetRole changes a member's role (leader only). Making someone leader demotes the caller to officer.
// Kick removes a member (leader, or officer for plain members).
func (c *pctx) SetRole(playerID int64, role string) error {
	if role != "leader" && role != "officer" && role != "member" && role != "kick" {
		return invalid("role must be leader, officer, member or kick")
	}
	if role == "kick" {
		if err := c.requireRole("leader", "officer"); err != nil {
			return err
		}
	} else if err := c.requireRole("leader"); err != nil {
		return err
	}
	if playerID == c.pl.ID {
		return rule("self", "you cannot change your own role")
	}
	if err := c.lockAlliance(); err != nil {
		return err
	}
	var tAlliance *int64
	var tRole *string
	if err := c.tx.QueryRow(c.ctx, `SELECT alliance_id, alliance_role FROM players WHERE id=$1 FOR UPDATE`, playerID).Scan(&tAlliance, &tRole); err != nil {
		return notFound("member")
	}
	if tAlliance == nil || *tAlliance != *c.pl.AllianceID {
		return notFound("member")
	}
	switch role {
	case "kick":
		if *c.pl.AllianceRole == "officer" && *tRole != "member" {
			return forbidden("officers can only remove members")
		}
		if _, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_id=NULL, alliance_role=NULL WHERE id=$1`, playerID); err != nil {
			return err
		}
		return notify(c.ctx, c.tx, playerID, "alliance_kicked", map[string]any{"by": c.pl.Name})
	case "leader":
		if _, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_role='leader' WHERE id=$1`, playerID); err != nil {
			return err
		}
		r := "officer"
		c.pl.AllianceRole = &r
		_, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_role='officer' WHERE id=$1`, c.pl.ID)
		return err
	}
	_, err := c.tx.Exec(c.ctx, `UPDATE players SET alliance_role=$2 WHERE id=$1`, playerID, role)
	return err
}

// SetRelation sets diplomacy towards another alliance (leader only). Relations
// are directional records; 'ally' and 'nap' block attacks from the declaring side.
func (c *pctx) SetRelation(otherID int64, status string) error {
	if err := c.requireRole("leader"); err != nil {
		return err
	}
	if otherID == *c.pl.AllianceID {
		return invalid("cannot set a relation with your own alliance")
	}
	var dis *time.Time
	var world int64
	if err := c.tx.QueryRow(c.ctx, `SELECT disbanded_at, world_id FROM alliances WHERE id=$1`, otherID).Scan(&dis, &world); err != nil || dis != nil || world != c.pl.WorldID {
		return notFound("alliance")
	}
	if status == "neutral" {
		_, err := c.tx.Exec(c.ctx, `DELETE FROM alliance_relations WHERE alliance_id=$1 AND other_id=$2`, *c.pl.AllianceID, otherID)
		return err
	}
	if status != "ally" && status != "nap" && status != "war" {
		return invalid("status must be ally, nap, war or neutral")
	}
	_, err := c.tx.Exec(c.ctx, `INSERT INTO alliance_relations(alliance_id, other_id, status, updated_at) VALUES ($1,$2,$3,$4)
		ON CONFLICT (alliance_id, other_id) DO UPDATE SET status=EXCLUDED.status, updated_at=EXCLUDED.updated_at`, *c.pl.AllianceID, otherID, status, c.now)
	return err
}

func (s *Service) Alliance(ctx contextT, q querier, worldID, id int64) (*Alliance, error) {
	a := &Alliance{ID: id}
	err := q.QueryRow(ctx, `SELECT name, tag, description, created_at FROM alliances WHERE id=$1 AND world_id=$2 AND disbanded_at IS NULL`, id, worldID).
		Scan(&a.Name, &a.Tag, &a.Description, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("alliance")
	}
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT id, name, alliance_role, score FROM players WHERE alliance_id=$1 ORDER BY CASE alliance_role WHEN 'leader' THEN 0 WHEN 'officer' THEN 1 ELSE 2 END, score DESC`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var m AllianceMember
		if err := rows.Scan(&m.PlayerID, &m.Name, &m.Role, &m.Score); err != nil {
			rows.Close()
			return nil, err
		}
		a.Score += m.Score
		a.Members = append(a.Members, m)
	}
	rows.Close()
	a.MemberCount = len(a.Members)
	rows, err = q.Query(ctx, `SELECT r.other_id, a.tag, r.status FROM alliance_relations r JOIN alliances a ON a.id=r.other_id WHERE r.alliance_id=$1 ORDER BY a.tag`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	a.Relations = []Relation{}
	for rows.Next() {
		var r Relation
		if err := rows.Scan(&r.OtherID, &r.OtherTag, &r.Status); err != nil {
			return nil, err
		}
		a.Relations = append(a.Relations, r)
	}
	return a, rows.Err()
}

func (s *Service) Alliances(ctx contextT, q querier, worldID int64) ([]Alliance, error) {
	rows, err := q.Query(ctx, `SELECT a.id, a.name, a.tag, a.description, a.created_at, count(p.id), COALESCE(sum(p.score),0)
		FROM alliances a LEFT JOIN players p ON p.alliance_id=a.id WHERE a.world_id=$1 AND a.disbanded_at IS NULL
		GROUP BY a.id ORDER BY 7 DESC, a.id LIMIT 200`, worldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Alliance{}
	for rows.Next() {
		var a Alliance
		if err := rows.Scan(&a.ID, &a.Name, &a.Tag, &a.Description, &a.CreatedAt, &a.MemberCount, &a.Score); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------- Messages, blocks, reports, notifications ----------

type Message struct {
	ID         int64      `json:"id"`
	From       string     `json:"from"`
	FromID     int64      `json:"fromId"`
	To         *string    `json:"to"`
	AllianceID *int64     `json:"allianceId"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	ReadAt     *time.Time `json:"readAt"`
}

// SendMessage sends a private message (by player name) or an alliance broadcast.
func (c *pctx) SendMessage(to string, alliance bool, subject, body string) (int64, error) {
	subject, body = strings.TrimSpace(subject), strings.TrimSpace(body)
	if utf8.RuneCountInString(subject) > 120 || utf8.RuneCountInString(body) > 4000 || body == "" {
		return 0, invalid("subject ≤ 120, body 1..4000 characters")
	}
	var countLastHour int
	if err := c.tx.QueryRow(c.ctx, `SELECT count(*) FROM messages WHERE from_player=$1 AND created_at > $2`, c.pl.ID, c.now.Add(-time.Hour)).Scan(&countLastHour); err != nil {
		return 0, err
	}
	if countLastHour >= 60 {
		return 0, errf(429, "rate_limited", "too many messages, try later")
	}
	var id int64
	if alliance {
		if c.pl.AllianceID == nil {
			return 0, rule("not_member", "you are not in an alliance")
		}
		err := c.tx.QueryRow(c.ctx, `INSERT INTO messages(world_id, from_player, alliance_id, subject, body, created_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
			c.pl.WorldID, c.pl.ID, *c.pl.AllianceID, subject, body, c.now).Scan(&id)
		return id, err
	}
	var target int64
	err := c.tx.QueryRow(c.ctx, `SELECT id FROM players WHERE world_id=$1 AND lower(name)=lower($2)`, c.pl.WorldID, to).Scan(&target)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, notFound("player")
	}
	if err != nil {
		return 0, err
	}
	var blocked bool
	if err := c.tx.QueryRow(c.ctx, `SELECT EXISTS(SELECT 1 FROM player_blocks WHERE player_id=$1 AND blocked_id=$2)`, target, c.pl.ID).Scan(&blocked); err != nil {
		return 0, err
	}
	if blocked {
		return 0, forbidden("this player does not accept your messages")
	}
	err = c.tx.QueryRow(c.ctx, `INSERT INTO messages(world_id, from_player, to_player, subject, body, created_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		c.pl.WorldID, c.pl.ID, target, subject, body, c.now).Scan(&id)
	if err == nil {
		_ = notify(c.ctx, c.tx, target, "message", map[string]any{"messageId": id, "from": c.pl.Name})
	}
	return id, err
}

func (c *pctx) Messages(box string, before int64, limit int) ([]Message, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if before <= 0 {
		before = 1 << 62
	}
	var where string
	args := []any{c.pl.ID, before, limit}
	switch box {
	case "sent":
		where = `m.from_player=$1`
	case "alliance":
		if c.pl.AllianceID == nil {
			return []Message{}, nil
		}
		args[0] = *c.pl.AllianceID
		where = `m.alliance_id=$1`
	default:
		where = `m.to_player=$1 AND NOT EXISTS (SELECT 1 FROM player_blocks b WHERE b.player_id=$1 AND b.blocked_id=m.from_player)`
	}
	rows, err := c.tx.Query(c.ctx, `SELECT m.id, f.name, m.from_player, t.name, m.alliance_id, m.subject, m.body, m.created_at, m.read_at
		FROM messages m JOIN players f ON f.id=m.from_player LEFT JOIN players t ON t.id=m.to_player
		WHERE `+where+` AND m.id < $2 ORDER BY m.id DESC LIMIT $3`, args...)
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

func (c *pctx) MarkMessageRead(id int64) error {
	_, err := c.tx.Exec(c.ctx, `UPDATE messages SET read_at=$3 WHERE id=$1 AND to_player=$2 AND read_at IS NULL`, id, c.pl.ID, c.now)
	return err
}

// ReportMessage flags a received message for moderation (store requirement for UGC).
func (c *pctx) ReportMessage(id int64) error {
	tag, err := c.tx.Exec(c.ctx, `UPDATE messages SET reported_at=$3 WHERE id=$1 AND (to_player=$2 OR alliance_id=$4)`, id, c.pl.ID, c.now, c.pl.AllianceID)
	if err == nil && tag.RowsAffected() == 0 {
		return notFound("message")
	}
	return err
}

func (c *pctx) Block(playerID int64, block bool) error {
	if playerID == c.pl.ID {
		return invalid("cannot block yourself")
	}
	if !block {
		_, err := c.tx.Exec(c.ctx, `DELETE FROM player_blocks WHERE player_id=$1 AND blocked_id=$2`, c.pl.ID, playerID)
		return err
	}
	var w int64
	if err := c.tx.QueryRow(c.ctx, `SELECT world_id FROM players WHERE id=$1`, playerID).Scan(&w); err != nil || w != c.pl.WorldID {
		return notFound("player")
	}
	_, err := c.tx.Exec(c.ctx, `INSERT INTO player_blocks(player_id, blocked_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, c.pl.ID, playerID)
	return err
}

type Notification struct {
	ID        int64          `json:"id"`
	Kind      string         `json:"kind"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"createdAt"`
	ReadAt    *time.Time     `json:"readAt"`
}

func (c *pctx) Notifications(after int64, limit int) ([]Notification, int, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := c.tx.Query(c.ctx, `SELECT id, kind, data, created_at, read_at FROM notifications WHERE player_id=$1 AND id > $2 ORDER BY id DESC LIMIT $3`, c.pl.ID, after, limit)
	if err != nil {
		return nil, 0, err
	}
	out := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Kind, &n.Data, &n.CreatedAt, &n.ReadAt); err != nil {
			rows.Close()
			return nil, 0, err
		}
		out = append(out, n)
	}
	rows.Close()
	var unread int
	err = c.tx.QueryRow(c.ctx, `SELECT count(*) FROM notifications WHERE player_id=$1 AND read_at IS NULL`, c.pl.ID).Scan(&unread)
	return out, unread, err
}

func (c *pctx) MarkNotificationsRead(upTo int64) error {
	_, err := c.tx.Exec(c.ctx, `UPDATE notifications SET read_at=$3 WHERE player_id=$1 AND id <= $2 AND read_at IS NULL`, c.pl.ID, upTo, c.now)
	return err
}

type ReportRow struct {
	ID         int64          `json:"id"`
	Kind       string         `json:"kind"`
	Side       string         `json:"side"`
	OccurredAt time.Time      `json:"occurredAt"`
	ReadAt     *time.Time     `json:"readAt"`
	Data       map[string]any `json:"data,omitempty"`
}

func (c *pctx) Reports(before int64, limit int) ([]ReportRow, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	if before <= 0 {
		before = 1 << 62
	}
	rows, err := c.tx.Query(c.ctx, `SELECT r.id, r.kind, rr.side, r.occurred_at, rr.read_at, r.data FROM report_recipients rr JOIN reports r ON r.id=rr.report_id
		WHERE rr.player_id=$1 AND r.id < $2 ORDER BY r.id DESC LIMIT $3`, c.pl.ID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ReportRow{}
	for rows.Next() {
		var r ReportRow
		if err := rows.Scan(&r.ID, &r.Kind, &r.Side, &r.OccurredAt, &r.ReadAt, &r.Data); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (c *pctx) Report(id int64) (*ReportRow, error) {
	r := &ReportRow{ID: id}
	err := c.tx.QueryRow(c.ctx, `SELECT r.kind, rr.side, r.occurred_at, rr.read_at, r.data FROM report_recipients rr JOIN reports r ON r.id=rr.report_id
		WHERE rr.player_id=$1 AND r.id=$2`, c.pl.ID, id).Scan(&r.Kind, &r.Side, &r.OccurredAt, &r.ReadAt, &r.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("report")
	}
	if err != nil {
		return nil, err
	}
	_, err = c.tx.Exec(c.ctx, `UPDATE report_recipients SET read_at=$3 WHERE report_id=$1 AND player_id=$2 AND read_at IS NULL`, id, c.pl.ID, c.now)
	return r, err
}
