# Architecture

Status: implemented (v0.1). Verified by the automated suites listed in `TEST_STRATEGY.md`.

## 1. Goals that drive the design

1. **Server-authoritative persistent world.** The client never decides resources, timers, combat or ownership.
2. **Correct offline progression without per-entity timers.** Time is modelled with timestamps and scheduled events.
3. **Exactly-once effects** under retries, concurrency, worker restarts and crashes.
4. **One engine, many games.** Everything game-specific lives in a declarative, versioned *Game Pack*.
5. **Simple to run.** One Go binary + PostgreSQL. No microservices, no mandatory Redis.

## 2. Context diagram

```mermaid
flowchart LR
  P[Player browser / Android WebView / Yandex Games iframe] -- HTTPS JSON REST --> S[Realmforge server (Go)]
  A[Admin / ops] -- REST /api/v1/admin --> S
  S -- SQL (pgx) --> DB[(PostgreSQL 16)]
  S -- reads at boot / migration --> GP[Game Packs (JSON on disk)]
  P -. platform SDK .-> PL[Yandex SDK / Capacitor]
```

## 3. Components

```mermaid
flowchart TB
  subgraph server [server/ — single Go module]
    API[internal/api\nHTTP, auth, idempotency,\nrate limit, CORS, routes]
    GAME[internal/game\naccounts · worlds · settlement/events ·\ncities · research · army · missions ·\ncombat resolution · heroes · market ·\nterritories · alliances · messages ·\nquests · notifications · admin]
    RULES[internal/rules\npure formulas: economy, ticks,\ncombat, loot, scouting, travel,\nmap generation, XP, RNG]
    PACK[internal/pack\nGame Pack types, strict parser,\nvalidator, indexes, hash]
    PLAT[internal/platform\nconfig, logging, DB pool,\nmigrations, clock]
    W[Worker loop\n(same binary, RF_WORKER)]
  end
  API --> GAME --> RULES
  GAME --> PACK
  RULES --> PACK
  API --> PLAT
  GAME --> PLAT
  W --> GAME
  subgraph client [client/ — React 19 + TS 7 + Vite 8]
    UI[screens/*] --> CTX[game/context] --> APIC[api/client]
    UI --> KIT[ui/kit]
    CTX --> PIDX[game/pack: PackIndex, theme]
    MAIN[main.tsx] --> ADP[platform/adapter: Web / Yandex / Android]
  end
```

Module boundaries:

| Module | Responsibility | Depends on | Must not |
|---|---|---|---|
| `pack` | Format, validation, lookups | — | touch DB/clock |
| `rules` | Deterministic formulas | `pack` | touch DB/clock/rand without seed |
| `game` | Transactions, locking, events, use cases | `rules`, `pack`, `platform` | parse HTTP |
| `api` | Transport, authN, idempotency, validation of envelopes | `game` | contain game rules |
| `platform` | Infra | — | know about games |

`game` is a single package with one file per subsystem (`city.go`, `missions.go`, `heroes.go`, `economy.go` (quests, market, territories, map, rankings), `social.go` (alliances, messages, notifications, reports), `accounts.go`, `admin.go`, `world.go`, `settle.go`). They share one transactional player context (`pctx`) because nearly every use case must lock and settle the same aggregate first; splitting them into Go packages would add interface plumbing without new isolation (decision D-004).

## 4. Game time model (critical)

* **Authoritative clock**: server UTC, millisecond precision (`platform.Clock`; tests inject `FakeClock`). Client time is used only to *display* countdowns, corrected by the offset from `serverTime` fields.
* **Stored state + timestamp**: each city stores resources (integer milli-units), population, garrison and `econ_updated_at`.
* **Lazy settlement**: before any read or command, the player's aggregate is *settled* to `now`: `lockPlayer` → process due events in `(due_at, id)` order, advancing the economy of all the player's cities to each event's instant before applying it → advance to `now`.
* **Economy integration** (`rules.AdvanceCity`): constant rates between absolute tick boundaries (multiples of `economyTickSeconds` since the epoch); population growth/starvation, desertion and overflow decay apply only at boundaries. Settling once or a thousand times gives the same result (path independence, tested).
* **Scheduled events** (`game_events`): job completions, mission arrivals/returns, deliveries. Each event belongs to the player whose state it changes. A background worker (`RunWorker`) settles players with due events so offline players progress, notifications appear and attacks resolve on time. If the worker is down, the next request settles lazily anyway — the worker affects latency, never correctness.
* **Exactly-once**: an event row is locked `FOR UPDATE`, applied, and marked `done` in the same transaction as its effects. Handlers are additionally idempotent with respect to the referenced row's status (e.g. a recalled mission makes its pending arrival a no-op).
* **Failures**: an event that throws is retried by the worker; after 5 attempts it is parked as `failed` (visible in admin, retryable) so it never blocks the player.
* **Long absences**: handled by the same integration (tested over 7 days: starvation and desertion apply; storage caps or overflow decay apply).

What is lazy vs. background:

| Process | Lazy on access | Background worker | Why |
|---|---|---|---|
| Resource production / consumption | ✔ | (implicitly via settle) | pure function of time |
| Population, desertion, overflow decay | ✔ | | tick-aligned, deterministic |
| Building/research/training completion | ✔ | ✔ | notifications + fresh rankings for offline players |
| Mission arrival (PvP/PvE), return | ✔ | ✔ | defender may be offline; attacker needs report |
| Market proceeds delivery | ✔ | ✔ | seller may be offline |
| Neutral respawn, hero wound expiry, territory yield | ✔ (computed from timestamps) | | no event needed |

## 5. Concurrency and consistency

* PostgreSQL READ COMMITTED + explicit row locks. Bounded automatic retry on deadlock/serialization errors (`InTx`).
* **Lock order** (must be respected everywhere): `players` (ascending id) → `cities` (ascending id) → `missions` → `heroes` → `tiles` → `market_orders` / `items`.
* **Invariant (R-TIME-4)**: a player's economic inputs (buildings, techs, garrison, away-upkeep) change only while that player's row is locked, at an event's due time or at `now`. Cross-player effects are routed as events of the affected player (e.g. market proceeds → `delivery` event) or confined to non-economic rows (mission rows, reports). This is why no transaction ever needs two player locks.
* Upkeep of units away is charged at departure strength until they return (keeps the invariant; documented in GAME_RULES).
* Resources never go negative: `rules.Spend` panics on overspend; every spend is preceded by `CanAfford` inside the lock.
* Double-spend, double-claim, double-fill: prevented by the player lock, `quest_progress` primary key, and order row lock respectively (concurrency tests exist for each).

## 6. Idempotency of commands

Every state-changing game command requires `Idempotency-Key`. The key row is inserted **inside the command's transaction**:
first request inserts → executes in a savepoint → stores status + exact response bytes → commits. A concurrent duplicate blocks on the unique index until the first commits, then replays the stored bytes (`Idempotent-Replay: true`). Same key + different body → 422. Domain errors are stored and replayed; infrastructure errors roll back the key so the client can retry. Keys expire after 24 h.

## 7. Game Pack versioning

* Packs are strict JSON (unknown fields rejected), validated (schema, ranges, references, cycles, locales, theme, paths), and hashed (SHA-256 of canonical encoding).
* A world **pins a snapshot** (`worlds.pack_json`, `pack_version`, `pack_hash`). Editing files on disk never changes a running world.
* Moving a world to a new version is an explicit, audited admin operation (`MigrateWorldPack`) with dry-run, compatibility check against all IDs in use, and ID renames/removals declared in the new pack's `migrations`. History is kept in `world_pack_history`.
* Worlds may set a **speed multiplier** applied on top of the pack's `worldSpeed` (speed worlds without forking a pack).

## 8. Persistence, migrations, backup

* Schema migrations: ordered SQL files embedded in the binary, applied at startup under an advisory lock, one transaction each (`platform.Migrate`).
* Policy: released migrations are immutable; changes go into new files (`0002_...sql`). Pre-release (v0.x) `0001_init.sql` was edited in place twice during initial development (noted in DECISIONS).
* Backup: `scripts/backup.sh` (pg_dump custom format, single snapshot), `scripts/restore.sh`, `scripts/verify-backup.sh` (restore into scratch DB + row-count comparison; executed successfully). Production should add WAL archiving / PITR (KNOWN_LIMITATIONS).

## 9. Scaling strategy

* Stateless API instances behind a load balancer; all coordination is in PostgreSQL row locks, so N instances are safe. Workers on several instances skip players locked by others (`FOR UPDATE SKIP LOCKED`).
* Hot paths are per-player; the largest table growth is `resource_ledger`, `notifications`, `reports`, `game_events` → partition by time / archive (planned).
* The in-memory rate limiter is per instance; for multiple instances introduce Redis or a gateway limiter (only then Redis becomes justified — decision D-007).
* Realtime push (WebSocket/SSE) is optional; REST + 15 s polling is sufficient for authoritative state.

## 10. Observability

* Structured logs (`log/slog`, JSON in production) with request IDs (`X-Request-ID`).
* `/healthz` (process), `/readyz` (DB ping), `/metrics` (Prometheus text: events processed/failed/overdue/parked, tx retries, commands, idempotent replays, pool stats).
* Admin endpoints: inspect player (settled state, pending events, ledger), failed events, audit log, reported messages.

## 11. Client architecture

* Renders state, sends commands; never computes authoritative outcomes. Shows server-derived previews only (costs/times/blocked reasons come from the server's `CityView`).
* `api/client.ts`: bearer token, idempotency key per user action reused across retries, retry with backoff on network/5xx/429, server clock offset.
* `game/context.tsx`: overview + current city, 15 s polling paused when hidden, persisted UI preferences (screen, city, locale, reduced motion).
* `game/pack.ts`: names, descriptions, icons and palette from the world's pack (no game names in the client code).
* `platform/adapter.ts`: Web / Yandex / Android adapters behind one interface.
