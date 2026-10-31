# Security Specification and Threat Model

Owner: Agent F (QA / Security Engineer)
Status: v1. Describes the security model **as implemented** at commit `dfa660e`, plus findings and required hardening. This is not a penetration-test report.
Date: 2026-10-31

Code reviewed:
- `server/internal/api/{http.go,routes.go}`
- `server/internal/game/{accounts,admin,settle,service,world,economy,social,missions,heroes}.go`
- `server/internal/platform/platform.go`
- `server/cmd/server/main.go`
- `server/migrations/0001_init.sql`
- `client/src/{api/client.ts,App.tsx,screens/*.tsx,game/pack.ts}`
- `Dockerfile`, `docker-compose.yml`, `.env.example`
- `docs/spec/PLATFORM_REQUIREMENTS.md`

---

## 1. System and trust boundaries

```
 [Browser / Yandex iframe / Android WebView]  — untrusted: any value it sends may be forged
          │  HTTPS (terminated by a reverse proxy, not by the app)
          ▼
 [Reverse proxy / LB]  — TLS, (future) WAF, real client IP
          │  HTTP
          ▼
 [Go server: api → game → rules]  — authoritative; one process = API + event worker
          │  pgx (TLS required in production unless RF_ALLOW_INSECURE_DB=true)
          ▼
 [PostgreSQL 16]  — single source of truth (state, sessions, idempotency keys, audit, ledger)
```

Assets to protect:
- Account credentials and sessions.
- The integrity of game state: resources, units, ownership, scores.
- Fair play: abuse and multi-accounting.
- Private messages (UGC).
- Admin powers.
- Availability during store moderation (PLATFORM_REQUIREMENTS R-7).

---

## 2. Threat model (STRIDE)

| # | STRIDE | Threat | Current controls | Residual risk / finding |
|---|---|---|---|---|
| T1 | **S**poofing | Stolen or guessed session token | 256-bit random opaque token; only its SHA-256 is stored; TTL (default 720 h); revocation on logout, ban and delete | Token sits in `localStorage`, so any XSS would steal it (S-9). TTL is absolute with no idle timeout. **S-1:** world-screen logout does not revoke the token. |
| T2 | Spoofing | Credential stuffing or brute force on `/auth/login` | Argon2id; constant-ish timing for unknown logins (`dummyHash`); per-IP token bucket (20 rps, burst 60) | No per-account lockout or back-off. The IP bucket is keyed on `RemoteAddr` (S-4), which is too generous per IP and collapses into one bucket behind a proxy. |
| T3 | Spoofing | Forged platform identity (Yandex, Android) | Only the `dev` platform is accepted, and `LoadConfig` refuses `RF_DEV_PLATFORM_AUTH=true` in production. Other platforms return **501**. | Yandex signature verification is not implemented (S-13), so Yandex users play as plain guests. |
| T4 | Spoofing | Admin impersonation | `is_admin` flag; `/admin/*` returns 403 for others | **S-2:** `RF_ADMIN_LOGIN` promotes *any* account registered under that login at every start. No second factor for admins. |
| T5 | **T**ampering | Client sends forged resources, timers or outcomes | Server-authoritative: the client only sends intents (ids and quantities). All costs, times and combat are computed by the server from the pinned pack. | — |
| T6 | Tampering | Double spend, double claim, double fill through concurrency or retries | Row locks in a fixed order; idempotency key in the same transaction; quest PK; order row lock plus status check; events processed exactly once | Covered by concurrency tests (TEST_STRATEGY C4, C5). **S-7:** one lock-order deviation, recovered by retry. |
| T7 | Tampering | SQL injection | Parameterised SQL everywhere. The only dynamic SQL fragments (`events(where)`, `Messages` box `where`) come from constant strings chosen in code. | Tested (`TestCrossAccountAccessDenied`) |
| T8 | Tampering | Malicious Game Pack (path traversal, invalid references) | Strict pack decoding (unknown fields rejected); 22 validation classes; icon paths must be relative; admin `packDir` must be a plain name | Pack authors are trusted operators |
| T9 | **R**epudiation | An admin denies a grant, ban or migration | `admin_audit` rows written in the same transaction for `grant`, `inspect`, `retry_event`, `ban` and `migrate_pack`; `resource_ledger` for every non-production resource change; `world_pack_history` | Reads of reported messages, failed events and the audit log are **not** audited. Request logs carry no account id or IP, which limits forensics. |
| T10 | **I**nformation disclosure | Access to another player's city, mission, hero, item, order, report or invite by id (IDOR) | Ownership checks return **404** for foreign ids (`cityOwner`, `ownHero`, `Recall`, `CancelOrder`, `RespondInvite`, `Report`, `Equip`) | Tested for city, job and mission |
| T11 | Information disclosure | Operational data | `/metrics` | **S-3:** unauthenticated, and it runs a `count(*)` query on every call. |
| T12 | Information disclosure | Map and intel | Incoming mission composition is hidden; scouting is needed for garrison and resources | No fog of war: owners, cities and lair levels are public (design choice) |
| T13 | **D**enial of service | Request floods, large bodies, slowloris | Token bucket per account (authenticated) and per IP (auth endpoints); 64 KiB body limit (413); `ReadHeaderTimeout` 5 s, `ReadTimeout` 15 s, `WriteTimeout` 30 s, `IdleTimeout` 120 s; panic recovery | **S-4:** argon2id uses 19 MiB per hash with no concurrency cap. `/metrics` is an unthrottled database query. The limiter is per instance and in memory. |
| T14 | Denial of service | A poison event blocks a player | Parked as `failed` after 5 attempts and no longer blocks the player (tested) | Until parked, that player's requests fail with 500 |
| T15 | DoS / abuse | Spam, harassment | 60 messages per hour per player; block; report | **S-8:** no moderation actions (delete, mute) and no profanity filter |
| T16 | **E**levation of privilege | Mass assignment, such as `{"admin":true}` | `DisallowUnknownFields` on every JSON body (tested); admin flag never writable through the API | — |
| T17 | Elevation of privilege | Alliance role abuse | Role checks under the alliance row lock; officers may kick only members; a leader cannot change their own role | — |
| T18 | Abuse | Multi-account farming (gift orders, transports, attack farming) | Market price band 1:10..10:1; premium resource not tradable or lootable; attack caps per target per 24 h; score floor; newbie protection | **S-12:** unlimited guest creation, transport missions push resources freely, and the band applies only when both resources have exchange rates. |

---

## 3. Authentication

| Item | Implementation (`game/accounts.go`) |
|---|---|
| Password hashing | **Argon2id**, m = 19 MiB (`19*1024` KiB), t = 2, p = 1, 32-byte key, 16-byte random salt. This is the OWASP baseline. Stored as a PHC string `$argon2id$v=19$m=…,t=…,p=…$salt$key`. `CheckPassword` parses the parameters from the stored hash, so the parameters can be raised later; there is **no automatic rehash on login** yet. Comparison uses `subtle.ConstantTimeCompare`. |
| Password policy | 8–128 bytes (`validPassword`). No breached-password list and no complexity rules. NIST-style length-only checking is acceptable, but a breached-password check is recommended. |
| Login identifier | Normalized with `TrimSpace` and `ToLower`. It must match `^[a-z0-9_.@-]{3,64}$`, which is also a database `CHECK` constraint. Unique. |
| User enumeration | Login: unknown users run a dummy argon2 verification to equalise timing, and every failure returns the same `401 invalid_credentials`. Register: `login_taken` (409) reveals that a login exists. This is accepted, with the IP rate limit as mitigation. |
| Session token | `newToken()` takes 32 bytes from `crypto/rand` and encodes them as base64url (43 characters). **Only `SHA-256(token)` is stored** (`sessions.token_hash`, the primary key). Lookup is by hash, so a database leak does not reveal usable tokens. Tokens longer than 200 characters are rejected before lookup. |
| Transport | `Authorization: Bearer <token>`. No cookies, so CSRF does not apply (§10). |
| TTL | `RF_SESSION_TTL`, default 720 h (30 days). Absolute expiry (`expires_at`); no sliding renewal and no idle timeout. |
| Revocation | `POST /auth/logout` sets `revoked_at` on that token. `DELETE /me` and an admin ban revoke **all** of the account's sessions. `Authenticate` checks `revoked_at IS NULL AND expires_at > now AND deleted_at IS NULL`. |
| Ban | `accounts.banned_until`. `Login` returns `403 banned` with the end date. `Authenticate` returns `403 banned` for still-valid tokens, and `Ban` also revokes the sessions. |
| Guest accounts | `POST /auth/guest` creates an account with `is_guest=true`, no login and no password, and returns a session. `POST /auth/upgrade` attaches a login and password to the same account id, which keeps all progress. Guests are full accounts: they can be deleted and banned. |
| Platform sign-in | `POST /auth/platform`. Only `platform=dev` (trusts the client, for development) is accepted. It is **refused at startup in production**. Any other platform returns 501 until its server-side signature verification exists. |
| Gaps | No password change or reset (there is no email). No session listing or "log out everywhere". No MFA for admins. No rehash-on-login. Expired session rows are never pruned (S-11). |

---

## 4. Authorization

1. **Authenticated routes** go through `Server.auth`, which resolves the token, rejects banned accounts and applies the per-account rate limit. Admin routes go through `Server.admin`, which adds an `IsAdmin` check and returns 403.
2. **World membership:** `playerRead`, `playerCmd` and `worldRead` resolve `(account, world)` to a player. If the account has not joined the world, the request gets `403 forbidden` ("you have not joined this world").
3. **Object ownership returns 404, not 403**, so callers cannot probe for other players' ids:

| Object | Check |
|---|---|
| City | `cityOwner` returns `not_found` when `players.account_id ≠ caller` |
| Mission recall | `m.PlayerID ≠ c.pl.ID` gives `not_found` |
| Hero | `ownHero` |
| Item | owner ≠ caller gives `not_found` |
| Market order cancel | owner ≠ caller gives `not_found` |
| Alliance invite response | `pid ≠ caller` gives `not_found` |
| Report | Only rows in `report_recipients` for the caller |
| Message | Read and report only as recipient (or current alliance member for alliance mail) |
| Alliance member role | Target must be in the same alliance, otherwise `not_found` |

4. **Admin role.** `accounts.is_admin` can only be set in the database or by the `RF_ADMIN_LOGIN` bootstrap at startup (S-2). Admin endpoints:
   - `GET /admin/players/{pid}` (audited `inspect`)
   - `POST /admin/cities/{cid}/grant` (audited, reason ≥ 5 characters, goes through the normal spend/grant path and the ledger)
   - `POST /admin/accounts/{aid}/ban` (audited)
   - `GET /admin/events/failed`
   - `POST /admin/events/{eid}/retry` (audited)
   - `GET /admin/audit`
   - `GET /admin/messages/reported`
   - `POST /admin/worlds/{wid}/migrate-pack` (audited and recorded in `world_pack_history`)

   The audit row is written **in the same transaction** as the action, so a committed action always has an audit row.

---

## 5. Integrity of game state

| Mechanism | Detail |
|---|---|
| Server-authoritative state | The client sends intents only. Economy, timers, combat, loot and travel are computed in `internal/rules` (pure, deterministic, unit-tested) from the world's **pinned** pack snapshot (`worlds.pack_json` and `pack_hash`). Disk changes never affect a running world. |
| Transactions | `Service.InTx` runs a READ COMMITTED transaction and retries up to 4 attempts with back-off on `40001` (serialization failure) or `40P01` (deadlock). Domain errors roll back. |
| Lock order | `players` (by ascending id), then `cities` (by ascending id), then `missions`, then `tiles`, then `market_orders`. It is documented in `0001_init.sql` and `game/service.go`. `lockPlayer` locks the player row and **all** of their cities `FOR UPDATE`, processes due events in `(due_at, id)` order and advances every city to each event's instant before applying it. **A transaction locks one player only:** cross-player effects are written as `game_events` (`mission_arrive` is owned by the target city's owner) or as `deliveries` (market proceeds) for the affected player. |
| Event exactly-once | Events are selected `FOR UPDATE` inside the owning player's lock, and the effect and the `status='done'` update commit together. `UNIQUE(kind, ref_id, due_at)` prevents double scheduling. Handlers are no-ops when the referenced row's status has changed (a recalled mission or a cancelled job). The worker uses `FOR UPDATE SKIP LOCKED`, so several workers or instances are safe. A failing event is parked after 5 attempts (`failed`), and an admin can retry it. Tested. |
| Idempotency keys | Every game command requires `Idempotency-Key` (`^[A-Za-z0-9_-]{8,80}$`). The key row (`account_id, key, request_hash = SHA-256(method + path + body)`) is inserted **in the same transaction** as the effects. A savepoint allows a domain error to be stored as the outcome. A concurrent duplicate blocks on the primary key until the first commits, then replays the stored bytes (`Idempotent-Replay: true`). The same key with a different request returns 422. An infrastructure error rolls back the key too, so retrying is safe. Keys are pruned after 24 h by the worker. Tested. |
| Quest claims | `quest_progress PRIMARY KEY (player_id, quest_id)` makes a second claim impossible even under concurrency. Tested. |
| Market | The seller's resources are held in escrow at placement. Accept locks the order row (`FOR UPDATE`) and checks `status='open'`, so only one buyer wins; the others get `order_closed`. Tested. Proceeds arrive through a delivery event in the seller's own lock. |
| Ledger | `resource_ledger` records every non-production change (build, research, train, refund, loot, rewards, market, exchange, admin) with reason and reference. `rules.Spend` **panics** on overspend, and stocks are stored as integer milli-units. |
| Known deviation (S-7) | `SetRole` (`social.go`) locks a **second player row** (`SELECT … FROM players WHERE id=$1 FOR UPDATE`) while holding the caller's lock. This contradicts the "one player per transaction" rule. A deadlock is detected by PostgreSQL and retried by `InTx`, so integrity holds, but it should use the documented order or an event. |

---

## 6. Input validation

- **Strict JSON.** `decode()` and `cmd()` use `DisallowUnknownFields`. An empty body becomes `{}` for commands. Malformed JSON returns `400 invalid_json`.
- **Body size.** `http.MaxBytesReader` caps the body at **64 KiB**; larger bodies get `413 body_too_large`.
- **Path ids** must be positive `int64` (`pathID`), otherwise 400.
- **Regexes:**

  | Field | Pattern |
  |---|---|
  | Login | `^[a-z0-9_.@-]{3,64}$` |
  | Ruler or hero name | `^[\p{L}\p{N}_ .'-]{3,24}$`, with no leading or trailing spaces |
  | Locale | `^[a-z]{2}(-[A-Z]{2})?$` |
  | Idempotency key | `^[A-Za-z0-9_-]{8,80}$` |

- **Bounds:**
  - Password: 8–128 characters.
  - Display name: ≤ 40 characters.
  - Message subject: ≤ 120 runes. Body: 1–4000 runes.
  - Market and exchange amounts: [1, 1e9].
  - Mission unit counts: [0, 1e9]. Cargo must be ≥ 0 and non-premium.
  - Train quantity: [1, `MaxTrainBatch`].
  - `retreatPct`: [0, 100].
  - Map viewport: ≤ 64×64.
  - List `limit`s are clamped: reports 1–100, notifications and messages 1–100, rankings 1–100, audit 1–500.
- **Gap (S-6):** the map's `x` and `y` query values are not bounded. A value above the int4 range causes a SQL encode error, which surfaces as a 500 rather than a 400. Low severity.
- **Parameterised SQL only** (pgx placeholders). Identifiers in `testutil` are sanitized with `pgx.Identifier`.
- **Files:**
  - `GET /packs/{pack}/assets/{file...}` rejects `..` and serves `filepath.Clean("/"+file)` under `RF_PACKS_DIR/<pack>/assets`.
  - The SPA fallback cleans the path the same way.
  - Admin `packDir` must not contain `/`, `\` or `.`.
- **Output encoding.** The React client renders all user content (names, subjects, bodies) as text. There is no `dangerouslySetInnerHTML` in `client/src`. JSON responses carry `X-Content-Type-Options: nosniff`.

---

## 7. Abuse prevention

| Control | Where | Value / notes |
|---|---|---|
| Rate limit: authenticated | `Server.auth` → `limiter.allow("a:<accountId>")` | Token bucket, `RF_RATE_RPS`=20, `RF_RATE_BURST`=60, returns 429 with `Retry-After: 1`. **In memory, per instance:** with N instances the effective limit is N×. Buckets idle for 10 min are swept. |
| Rate limit: unauthenticated | `ipLimited` on register, login, guest, platform | Keyed on `RemoteAddr` with the port stripped. **Behind a reverse proxy every client shares one bucket.** The real IP must come from a trusted proxy header, and the limit for auth endpoints should be stricter: e.g. 5 per minute for login, and a small number per hour for guest creation per IP. |
| Message rate | `SendMessage` | **60 per player per hour**, counted under the player lock; returns `429 rate_limited` |
| Market price band | `PlaceOrder` | Price ratio must be within **1:10..10:1** of the NPC exchange reference (`price_out_of_band`) when both resources have rates. Premium resources are not tradable (pack validator plus runtime check). Open orders are capped by `MarketMaxOpenOrders`. |
| Attack caps | `checkPvP` | `MaxAttacksPerTargetPerDay` (`attack_limit`); `AttackScoreFloorPct` (`target_too_weak`); allied or NAP targets are blocked (`friendly_target`); no self-attack |
| Newbie protection | `Join` sets `protected_until = now + NewbieProtectionHours`. Attacks on a protected target are refused (`target_protected`, tested). **Attacking another player ends your own protection.** |
| Loot protection | Premium resources are not lootable (tested); the warehouse protected amount is excluded from loot |
| Resource caps | Storage caps and city and territory caps (`city_limit`, `territory_limit`), hero limit, queue limit |
| Gaps (S-12) | Guest creation is limited only by the IP bucket, and guest accounts never expire. Transport missions can move unlimited cargo between accounts. There is no device or account linkage for multi-account detection. The client derives some limits for display only; the server enforces them all. |

---

## 8. User-generated content and moderation

Google Play requires this for apps with chat (GP-16); it is also good practice on Yandex.

| Requirement | Status |
|---|---|
| Report a message | `POST /worlds/{wid}/messages/{id}/report` sets `messages.reported_at` (recipient or alliance member only). The UI has "Report abuse" on inbox and alliance messages. |
| Block a player | `POST /worlds/{wid}/blocks {playerId, block}` (`player_blocks`). A blocked sender gets `403 forbidden` on send, and their messages are hidden from the blocker's **inbox**. The UI has "Block" with confirmation. |
| Moderation queue | `GET /admin/messages/reported`: the latest 200 reported messages with their bodies |
| **Gaps (S-8)** | <ul><li>No reporter id or reason is stored, and repeated reports collapse into one timestamp.</li><li>No admin action on a report: delete message, mute or warn. Only a full account ban exists.</li><li>No "resolved" state.</li><li>Blocked players' messages still show in **alliance chat**.</li><li>No unblock UI and no list of blocked players.</li><li>No profanity or link filter.</li><li>Admin reads of reported content are not audited.</li><li>Terms of use and a community policy (GP-16) are not written.</li><li>The send error reveals block status (403 vs 404).</li></ul> |

---

## 9. Privacy and data deletion

- **Data collected:**
  - Login (it may be an email-like string, since `@` is allowed), password hash, display name, locale.
  - In-game state and private messages.
  - Ledger and audit rows.
  - Idempotency responses, kept for 24 h.
- Request logs record method, path, status, latency and request id. They do **not** record IP, user agent, token or account id.
- **Third parties:**
  - The web client loads pack fonts from `fonts.googleapis.com` at runtime, which sends the player's IP to Google. This must be declared, or the fonts self-hosted (recommended for the Yandex CSP as well).
  - Platform SDKs (Yandex) are loaded only in the Yandex build.
- **Account deletion.** `DELETE /me` (in-app, GP-10) runs in one transaction:
  1. `login` and `password_hash` are set to NULL, `display_name` to `'deleted'` and `deleted_at` to now.
  2. `platform_links` rows are deleted.
  3. All sessions are revoked.
  4. In-world `players.name` is set to `'deleted-<id>'`.

  This is **anonymization**, chosen so that other players' reports and messages stay consistent.
- **Retained after deletion (S-8, disclose it in the privacy policy or purge it):**
  - Message bodies the user sent.
  - In-world players, cities and armies, which still exist on the map and in rankings.
  - Ledger and audit rows.
  - Idempotency rows until pruning.
- **Missing:**
  - The **web deletion link** that GP-10 requires.
  - A grace period.
  - A data export.
  - Purging the in-world player (or converting it to an abandoned NPC).
  - A written retention policy.

---

## 10. Secrets, configuration, transport and headers

**Configuration**
- All configuration comes from environment variables (`platform.LoadConfig`, `.env.example`). Nothing secret lives in the code.
- `RF_ENV=production` **refuses to start** when:
  - `RF_DEV_PLATFORM_AUTH=true`, or
  - the database URL has `sslmode=disable`, unless `RF_ALLOW_INSECURE_DB=true` explicitly marks a private network.
- `docker-compose.yml` defaults `POSTGRES_PASSWORD` to `realmforge-dev`. This is for local use only, and production must override it.

**Container**
- The image is `gcr.io/distroless/static-debian12:nonroot` and runs as `USER nonroot`, with static `CGO_ENABLED=0` binaries and `-trimpath`.

**Secrets (future)**
- Yandex secret key, Google Play service account, Android upload keystore. Keep them in the CI or deployment secret store and never in the repository (PLATFORM_REQUIREMENTS §3.3).

**TLS**
- The app does **not** terminate TLS. A reverse proxy or load balancer must provide HTTPS (HTTP/2) and add HSTS.
- Yandex accepts only https and wss hosts (YG-15). Android cleartext is disabled from API 28 onwards.

**CORS**
- Exact-match allow-list from `RF_ALLOWED_ORIGINS`. The default is `http://localhost:5173,https://localhost,capacitor://localhost`.
- The server reflects the origin only when it is listed; `"*"` is supported but must not be used in production.
- Preflight answers allow headers `Authorization, Content-Type, Idempotency-Key, X-Request-ID` and methods `GET, POST, PUT, DELETE, OPTIONS`, with `Max-Age: 600`.
- No `Allow-Credentials`, which is correct for bearer tokens.
- **Production must add the Yandex game origins**, which are still UNVERIFIED (capture them in draft mode, PLATFORM_REQUIREMENTS R-6). Tested (`TestCORS`).

**CSRF**
- Not applicable: authentication is a bearer header that browsers never attach automatically, and no cookies are used.
- If cookies are ever introduced (for example for a web deletion page), add SameSite=strict cookies and CSRF tokens.

**Response headers set today**
- `X-Request-ID` (echoed if ≤ 64 characters, otherwise generated)
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: no-referrer`
- `Cache-Control: no-store` on API JSON
- `ETag` plus `private, max-age=0, must-revalidate` on the pack
- `public, max-age=3600` on pack assets

**Response headers missing (S-9)**
- `Content-Security-Policy` for the served web client. Suggested: `default-src 'self'; connect-src 'self' https://api.<domain>; img-src 'self' data:; style-src 'self' 'unsafe-inline'; font-src 'self'; frame-ancestors` set to the Yandex game domains.
- `Strict-Transport-Security` (at the proxy).
- `Permissions-Policy`.
- The CSP is especially important because the session token lives in `localStorage`.

---

## 11. Findings (report only, not fixed)

| ID | Severity | Finding | Location | Recommendation |
|---|---|---|---|---|
| S-1 | Medium | **World-screen "Log out" does not revoke the session.** It only clears the token locally, so the server token stays valid for up to 30 days. The in-game Profile logout does call `POST /auth/logout`. | `client/src/App.tsx` (`onLogout={() => api.setToken(null)}`) | Call `POST /auth/logout` before clearing, as `ProfileScreen` does. |
| S-2 | Medium | **Admin bootstrap by login name.** On every start, `UPDATE accounts SET is_admin=true WHERE login=$RF_ADMIN_LOGIN` runs in any environment. If the operator sets the variable before registering that login, anyone who registers it first becomes admin at the next restart. | `server/cmd/server/main.go` | Bootstrap only an existing account id, only once (or only in development), log it, and audit it. Better: a CLI command (`packtool`-style) for granting admin. |
| S-3 | Medium | **`/metrics` is public and queries the database** with `count(*) … FILTER` over `game_events` on every hit, with no rate limit. It is an information leak and a cheap DoS amplifier. | `api/http.go` `metrics`, `routes.go` | Bind it to an internal port or require an admin token or network ACL; cache the event counts. |
| S-4 | Medium | **Argon2id memory DoS and weak per-IP limit.** Each login or register allocates 19 MiB, with no concurrency cap; 100 parallel logins means about 1.9 GiB. The IP limiter is keyed on `RemoteAddr`: one shared bucket behind a proxy, and a generous 20 rps / burst 60 per IP for authentication. No per-account lockout. | `game/accounts.go`, `api/http.go` `ipLimited` | Add a semaphore around hashing (e.g. 2×CPU), stricter auth-specific buckets, per-login back-off and trusted-proxy real-IP handling. |
| S-5 | Low | **Admin mutations are not idempotent and loosely validated.** `grant` and `ban` do not use `Idempotency-Key`, so a retried grant is applied twice. `ban` with a missing `until` stores the zero time, which effectively means "not banned" while still revoking sessions and writing a misleading audit row. Ban `reason` is optional. An admin can ban themselves or other admins. | `routes.go` admin handlers, `game/admin.go` `Ban` | Route admin mutations through `command()`; require `until > now` and a reason; forbid self-ban. |
| S-6 | Low | Map `x` and `y` are unbounded: out-of-int4 values give 500 instead of 400. | `routes.go` `/map`, `economy.go` `Map` | Clamp to `[0, mapWidth)` and `[0, mapHeight)`. |
| S-7 | Low | `SetRole` locks another player's row inside the caller's transaction, against the documented lock rule. Possible deadlocks are retried by `InTx`. | `game/social.go` `SetRole` | Lock in ascending id order, or apply role changes through the alliance row only. |
| S-8 | Low (compliance) | Deletion leaves message bodies and in-world entities. Reports carry no reporter or reason and allow no moderation action. Alliance chat ignores blocks. | `game/accounts.go` `DeleteAccount`, `game/social.go` | See §8 and §9. Required before Google Play submission. |
| S-9 | Low | No CSP, HSTS or `frame-ancestors`. The token is in `localStorage`. Fonts come from a third-party origin. | `api/http.go` `ServeHTTP`, `client/src/game/pack.ts` `applyTheme` | Add the headers; self-host the fonts. |
| S-10 | Low | Block status is disclosed (403 "does not accept your messages" vs 404). A new alliance member can read the entire alliance history. | `game/social.go` | Return the same response either way; optionally limit history to messages after the join date. |
| S-11 | Info | Expired or revoked sessions are never deleted; no session list or rotation; absolute 30-day TTL. | `sessions` table | Prune them in the worker; consider a 7-day sliding TTL. |
| S-12 | Info | Multi-account farming vectors: unlimited guests, transport cargo between players, band only when rates exist. | `world.go`, `missions.go`, `economy.go` | Rate-limit guest creation; cap or tax transports to non-alliance players; add an abuse dashboard. |
| S-13 | Info | Yandex `getPlayer({signed:true})` signature verification and Google Play purchase verification are not implemented. Correctly, they return 501. | `game/accounts.go` `PlatformLogin` | Implement before using platform identity or any purchase (PLATFORM_REQUIREMENTS YG-10, YG-25, GP-11). |

---

## 12. Known gaps (summary)

- **No TLS termination in the app**: a proxy is required.
- **No WAF** and no bot detection.
- **The rate limiter is not distributed** (in memory, per instance). Scaling to N instances needs Redis or PostgreSQL-backed buckets, or limits at the proxy.
- **Yandex signature verification is not implemented**, and neither is the Google Play Billing server verification.
- **CSRF protection is not needed** today, because tokens travel in a bearer header and no cookies are used. This must be revisited if cookies appear.
- No security headers beyond `nosniff` and `Referrer-Policy`. No admin MFA. No audit of admin reads.
- No dependency or vulnerability scanning in CI (`govulncheck`, `npm audit`), and no SAST.

---

## 13. Backup and restore

| Item | Current state |
|---|---|
| Backup | `scripts/backup.sh <db-url> [outdir]` runs `pg_dump --format=custom` (a single consistent snapshot) and verifies the archive with `pg_restore --list`. Also available as `make backup`. |
| Restore | `scripts/restore.sh <db-url> <file>` runs `pg_restore --single-transaction --exit-on-error` into an **empty** database (`make restore FILE=…`). Stop the app servers first. |
| Verification drill | `scripts/verify-backup.sh` (`make verify-backup`): backup, restore into a scratch database, then compare row counts for `accounts`, `players`, `cities`, `game_events`, `missions` and `resource_ledger`. |
| Game-time behaviour after a restore | State is stored as "value at `econ_updated_at`" plus scheduled events. After a restore, the first settle of each player accrues production and processes every event due during the outage, applying offline progression for the gap. Combat that was due during the outage resolves with the restored armies. Announce rollbacks to players. |
| Gaps | No PITR or WAL archiving, so up to one backup interval of data can be lost. No automated schedule or off-site encrypted storage. Backups contain password hashes and private messages, so encrypt them at rest and restrict access. No restore drill in CI. Idempotency keys restored from an older backup may replay stale responses for up to 24 h; this is harmless but should be noted. |
| Target | Daily logical dump plus continuous WAL archiving (PITR), RPO ≤ 5 min and RTO ≤ 1 h. Quarterly restore drill using `verify-backup.sh` plus a smoke `play.mjs`. Retain backups 30 days. Deleting an account must also be honoured in backups: document in the privacy policy that backups expire within the retention window. |

---

## 14. Security test checklist (to automate)

Existing tests: TEST_STRATEGY C5, C8, C9, C10 and C24. Missing tests to add:

1. `429` from the token bucket, both per IP and per account.
2. `413` on a 65 KiB body.
3. Banned account: login returns 403 and existing tokens return 403.
4. Admin endpoints return 403 for players. The `ban` and `retry` audit rows exist.
5. Foreign hero, item, order, report, message and invite ids return 404 over HTTP.
6. Message cap: the 61st message within an hour returns 429.
7. Block: sending returns 403 and the inbox hides the sender. Report: the admin list shows the message.
8. `DELETE /me`: all tokens revoked and the player renamed.
9. Response headers present (after S-9 is fixed); `/metrics` not reachable from outside (after S-3).
10. In CI: `govulncheck ./...`, `npm audit --omit=dev`, `gofmt` / `go vet`.

## Remediation status (2026-10-31, after review)

| Finding | Status | Change |
|---|---|---|
| S-1 | **Fixed** | World-screen logout calls `POST /auth/logout` before clearing the token (`client/src/App.tsx`). |
| S-2 | **Mitigated** | `RF_ADMIN_LOGIN` is ignored when `RF_ENV=production` (warning logged); production admins are promoted explicitly in SQL. A dedicated audited admin CLI is still open. |
| S-3 | **Fixed** | `/metrics` requires `Authorization: Bearer $RF_METRICS_TOKEN` when set and is disabled (404) in production without a token. Count caching is still open. |
| S-4 | **Partially fixed** | At most 4 concurrent argon2id computations (`hashSlots`); stored hash parameters are bounded; `RF_TRUST_PROXY=true` uses the X-Forwarded-For client IP. Per-login back-off and auth-specific buckets are still open. |
| S-5 | **Fixed** (grant, ban validation) | Admin grant goes through the idempotent command path (Idempotency-Key required). Ban requires a future `until` and a reason. Self-ban protection is still open. |
| S-6 | **Fixed** | Map viewport origin is bounded (400 instead of 500). |
| S-7 | **Accepted exception** | Alliance role and kick updates touch another player's row; deadlocks are detected by PostgreSQL and retried by `InTx`. Recorded in DECISIONS D-009. |
| S-8 | **Partially fixed** | Alliance chat now hides messages from blocked players. Reporter reason and moderation actions are still open. |
| S-9 to S-13 | Open | Tracked in RISK_REGISTER and ROADMAP. |
