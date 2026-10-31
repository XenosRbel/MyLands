package game

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"
)

type Account struct {
	ID          int64      `json:"id"`
	Login       *string    `json:"login"`
	DisplayName string     `json:"displayName"`
	IsGuest     bool       `json:"isGuest"`
	IsAdmin     bool       `json:"isAdmin"`
	Locale      string     `json:"locale"`
	CreatedAt   time.Time  `json:"createdAt"`
	BannedUntil *time.Time `json:"bannedUntil,omitempty"`
}

// Argon2id parameters (OWASP 2024 baseline: m=19 MiB, t=2, p=1).
const (
	argonTime    = 2
	argonMemory  = 19 * 1024
	argonThreads = 1
	argonKeyLen  = 32
)

func HashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(pw), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func CheckPassword(hash, pw string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var m uint32
	var t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

var reLogin = regexp.MustCompile(`^[a-z0-9_.@-]{3,64}$`)

func validPassword(pw string) bool { return len(pw) >= 8 && len(pw) <= 128 }

func newToken() (string, []byte) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	tok := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(tok))
	return tok, h[:]
}

func tokenHash(tok string) []byte {
	h := sha256.Sum256([]byte(tok))
	return h[:]
}

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	Account   *Account  `json:"account"`
}

func (s *Service) newSession(ctx context.Context, q querier, accountID int64, ttl time.Duration) (*Session, error) {
	tok, h := newToken()
	exp := s.Clock.Now().Add(ttl)
	if _, err := q.Exec(ctx, `INSERT INTO sessions(token_hash, account_id, expires_at) VALUES ($1,$2,$3)`, h, accountID, exp); err != nil {
		return nil, err
	}
	a, err := s.accountByID(ctx, q, accountID)
	if err != nil {
		return nil, err
	}
	return &Session{Token: tok, ExpiresAt: exp, Account: a}, nil
}

func (s *Service) accountByID(ctx context.Context, q querier, id int64) (*Account, error) {
	a := &Account{}
	err := q.QueryRow(ctx, `SELECT id, login, display_name, is_guest, is_admin, locale, created_at, banned_until FROM accounts WHERE id=$1 AND deleted_at IS NULL`, id).
		Scan(&a.ID, &a.Login, &a.DisplayName, &a.IsGuest, &a.IsAdmin, &a.Locale, &a.CreatedAt, &a.BannedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notFound("account")
	}
	return a, err
}

func normLogin(l string) string { return strings.ToLower(strings.TrimSpace(l)) }

func (s *Service) Register(ctx context.Context, login, password, display string, ttl time.Duration) (*Session, error) {
	login = normLogin(login)
	if !reLogin.MatchString(login) || !validPassword(password) {
		return nil, invalid("login 3-64 chars [a-z0-9_.@-], password 8-128 chars")
	}
	display = strings.TrimSpace(display)
	if display == "" {
		display = login
	}
	if len(display) > 40 {
		return nil, invalid("display name too long")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	var sess *Session
	err = s.InTx(ctx, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO accounts(login, password_hash, display_name) VALUES ($1,$2,$3) RETURNING id`, login, hash, display).Scan(&id)
		if err != nil {
			if isUnique(err) {
				return conflict("login_taken", "login already registered")
			}
			return err
		}
		sess, err = s.newSession(ctx, tx, id, ttl)
		return err
	})
	return sess, err
}

// dummyHash equalises timing for unknown logins.
var dummyHash, _ = HashPassword("timing-equaliser")

func (s *Service) Login(ctx context.Context, login, password string, ttl time.Duration) (*Session, error) {
	login = normLogin(login)
	var id int64
	var hash *string
	var banned *time.Time
	err := s.DB.QueryRow(ctx, `SELECT id, password_hash, banned_until FROM accounts WHERE login=$1 AND deleted_at IS NULL`, login).Scan(&id, &hash, &banned)
	if errors.Is(err, pgx.ErrNoRows) {
		CheckPassword(dummyHash, password)
		return nil, errf(401, "invalid_credentials", "invalid login or password")
	}
	if err != nil {
		return nil, err
	}
	if hash == nil || !CheckPassword(*hash, password) {
		return nil, errf(401, "invalid_credentials", "invalid login or password")
	}
	if banned != nil && banned.After(s.Clock.Now()) {
		return nil, errf(403, "banned", "account suspended until %s", banned.Format(time.RFC3339))
	}
	return s.newSession(ctx, s.DB, id, ttl)
}

// Guest creates an anonymous account (required by web game catalogues:
// play without forced registration). It can later be upgraded.
func (s *Service) Guest(ctx context.Context, ttl time.Duration) (*Session, error) {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	name := "Guest-" + strings.ToUpper(hex.EncodeToString(b))
	var sess *Session
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO accounts(display_name, is_guest) VALUES ($1, true) RETURNING id`, name).Scan(&id); err != nil {
			return err
		}
		var err error
		sess, err = s.newSession(ctx, tx, id, ttl)
		return err
	})
	return sess, err
}

// Upgrade attaches credentials to a guest account, preserving progress.
func (s *Service) Upgrade(ctx context.Context, accountID int64, login, password string) (*Account, error) {
	login = normLogin(login)
	if !reLogin.MatchString(login) || !validPassword(password) {
		return nil, invalid("login 3-64 chars [a-z0-9_.@-], password 8-128 chars")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	tag, err := s.DB.Exec(ctx, `UPDATE accounts SET login=$2, password_hash=$3, is_guest=false WHERE id=$1 AND is_guest AND deleted_at IS NULL`, accountID, login, hash)
	if err != nil {
		if isUnique(err) {
			return nil, conflict("login_taken", "login already registered")
		}
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, rule("not_guest", "account already has credentials")
	}
	return s.accountByID(ctx, s.DB, accountID)
}

// PlatformLogin signs in with a platform identity (e.g. a catalogue player id).
// Only the "dev" platform is implemented; it must be disabled in production
// because it trusts the client. Real platforms need server-side signature
// verification (see docs/spec/PLATFORM_REQUIREMENTS.md) and return 501 until configured.
func (s *Service) PlatformLogin(ctx context.Context, platform, externalID string, devAllowed bool, ttl time.Duration) (*Session, error) {
	if platform != "dev" {
		return nil, errf(501, "platform_not_configured", "platform %q sign-in is not configured on this server", platform)
	}
	if !devAllowed {
		return nil, forbidden("dev platform sign-in is disabled")
	}
	if len(externalID) < 3 || len(externalID) > 128 {
		return nil, invalid("invalid external id")
	}
	var sess *Session
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, `SELECT account_id FROM platform_links WHERE platform=$1 AND external_id=$2`, platform, externalID).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.QueryRow(ctx, `INSERT INTO accounts(display_name, is_guest) VALUES ($1, true) RETURNING id`, "Player-"+externalID[:min(8, len(externalID))]).Scan(&id); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO platform_links(platform, external_id, account_id) VALUES ($1,$2,$3)`, platform, externalID, id); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		sess, err = s.newSession(ctx, tx, id, ttl)
		return err
	})
	return sess, err
}

// Authenticate resolves a bearer token.
func (s *Service) Authenticate(ctx context.Context, token string) (*Account, error) {
	if token == "" || len(token) > 200 {
		return nil, errf(401, "unauthenticated", "missing or invalid token")
	}
	a := &Account{}
	err := s.DB.QueryRow(ctx, `SELECT a.id, a.login, a.display_name, a.is_guest, a.is_admin, a.locale, a.created_at, a.banned_until
		FROM sessions s JOIN accounts a ON a.id=s.account_id
		WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at > $2 AND a.deleted_at IS NULL`, tokenHash(token), s.Clock.Now()).
		Scan(&a.ID, &a.Login, &a.DisplayName, &a.IsGuest, &a.IsAdmin, &a.Locale, &a.CreatedAt, &a.BannedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errf(401, "unauthenticated", "missing or invalid token")
	}
	if err != nil {
		return nil, err
	}
	if a.BannedUntil != nil && a.BannedUntil.After(s.Clock.Now()) {
		return nil, errf(403, "banned", "account suspended")
	}
	return a, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	_, err := s.DB.Exec(ctx, `UPDATE sessions SET revoked_at=$2 WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash(token), s.Clock.Now())
	return err
}

func (s *Service) SetLocale(ctx context.Context, accountID int64, locale string) error {
	if !regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`).MatchString(locale) {
		return invalid("invalid locale")
	}
	_, err := s.DB.Exec(ctx, `UPDATE accounts SET locale=$2 WHERE id=$1`, accountID, locale)
	return err
}

// DeleteAccount anonymises the account (store policies require in-app deletion):
// credentials, platform links and sessions are removed; in-world players are
// renamed so other players' history (reports, messages) stays consistent.
func (s *Service) DeleteAccount(ctx context.Context, accountID int64) error {
	return s.InTx(ctx, func(tx pgx.Tx) error {
		now := s.Clock.Now()
		if _, err := tx.Exec(ctx, `UPDATE accounts SET login=NULL, password_hash=NULL, display_name='deleted', deleted_at=$2 WHERE id=$1`, accountID, now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM platform_links WHERE account_id=$1`, accountID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE sessions SET revoked_at=$2 WHERE account_id=$1 AND revoked_at IS NULL`, accountID, now); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE players SET name='deleted-' || id WHERE account_id=$1`, accountID)
		return err
	})
}
