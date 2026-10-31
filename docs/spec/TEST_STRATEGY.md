# Test Strategy

Owner: Agent F (QA / Security Engineer)
Status: v1. Describes the tests **as they exist in the repository** on 2026-10-31, how to run them, what they prove, and what is missing.
Verified locally on 2026-10-31 (PostgreSQL 16 from `scripts/dev-postgres.sh`):

| Command | Result |
|---|---|
| `RF_REQUIRE_DB=1 go test -race -count=1 ./...` | All 5 packages with tests pass: `api`, `game`, `pack`, `rules`, plus the OpenAPI coverage test. |
| `npm test` (vitest) | 7 / 7 pass |
| `node e2e/play.mjs` | **Not** re-run in this pass. It needs a running server, the built client and Playwright. |

---

## 1. Test pyramid as implemented

```
                       ┌───────────────────────────┐
  Browser E2E (1)      │ client/e2e/play.mjs        │  Playwright, desktop + mobile, real server + DB
                       ├───────────────────────────┤
  API / HTTP (7)       │ internal/api/*_test.go     │  httptest server + real PostgreSQL + fake clock
                       ├───────────────────────────┤
  Game integration(22) │ internal/game/game_test.go │  real PostgreSQL, fake clock, concurrency
                       ├───────────────────────────┤
  Pack validator (5)   │ internal/pack/pack_test.go │  pure; 22 table cases + every shipped pack
  Rules unit (24)      │ internal/rules/rules_test.go│ pure, deterministic formulas
  Client unit (7)      │ client/src/__tests__/       │  vitest + jsdom
                       └───────────────────────────┘
```

Counts are top-level `Test*` functions or vitest `it` blocks. No benchmarks or fuzz tests exist yet.

### 1.1 Test infrastructure

- **`server/internal/testutil/testutil.go`** gives each test package its own database, `rf_test_<pkg>`. On every `DB()` call it drops and recreates the `public` schema and applies the embedded migrations, so separate packages can run in parallel.
  - The admin URL comes from `RF_TEST_ADMIN_URL`, default `postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable`.
  - If PostgreSQL is unreachable, the tests **skip**, unless `RF_REQUIRE_DB=1` is set, in which case they **fail**. CI and `make test-go` set it.
  - Logs are discarded unless `RF_TEST_LOG=1`.
- **`platform.FakeClock`** (`NewFakeClock`, `Advance`, `Set`) drives all time-dependent tests: offline progression, travel, respawn, expiry. No test sleeps on wall-clock time; only the browser E2E does.
- **Packs:**
  - `server/internal/pack/testdata/minimal.json` holds small, hand-checkable numbers (for example wood 100/h, storage 1000, queue limit 3, refund 80 %, market fee 2 %).
  - The game and API tests also use the shipped `packs/ashen-crowns` pack, in `TestEndToEnd`.

---

## 2. Inventory: every test and what it proves

### 2.1 Rules unit tests: `server/internal/rules/rules_test.go` (pure, no DB)

| Test | Proves |
|---|---|
| `TestCostAndDuration` | `CostAt` growth and rounding (100×1.5² = 225; 74.25 → 74); level 1 = base cost; `DurationAt` applies the time factor and world speed; minimum duration is 1 s. |
| `TestModsOrder` | Modifier evaluation order: flat additions first, then percent: (10+2)×1.2 = 14.4. The time-factor floor holds. |
| `TestRatesEmploymentAndStorage` | Employment = pop / workers needed, which scales gross output. Food consumption per inhabitant, storage and housing totals. Uncapped resources have no storage cap. Modifiers apply to production. |
| `TestAccrueCapsAndFloors` | Accrual is capped at storage and floored at 0. Milli-unit precision (36/h for 1 s = 10 milli). Stock above the cap is not reduced by positive production. Negative elapsed time is a no-op. |
| `TestTickPopulation` | Fractional growth is carried between ticks. Growth stops at the housing cap. 1 % starvation decline. Population above housing is clamped. |
| `TestDesertionAndIdleMode` | Unpaid upkeep causes desertion at the configured percent per hour. Idle-based population growth mode. |
| `TestIdleTaxAndUpkeep` | Idle-inhabitant tax per building level; unit upkeep plus away upkeep in the net rate. |
| `TestOverflowDecay` | In overflow-decay mode there is no hard cap, and the surplus decays by a percentage each tick. |
| `TestAdvancePathIndependence` | Settling once over a long period equals settling many times (within ≤ 1 milli per settlement). This is what makes lazy settlement correct. |
| `TestStarvationOverLongAbsence` | Huge upkeep and no food: food reaches 0 and population declines over a long absence. |
| `TestSpendGrant` | `CanAfford` and `Missing`; `Grant` respects caps and reports the lost amount; **overspend panics**, the last line of defence against negative stock. |
| `TestCombatDeterministicAndConservative` | Same inputs give the same outcome. Units are conserved (start = losses + left) for both sides. An overwhelming attacker wins. Inputs are not mutated. |
| `TestCombatWallHelpsDefender` | The wall bonus reduces defender losses. Equal armies tie, and a tie goes to the defender. |
| `TestCombatEmptyDefender` | An empty defender gives the attacker an automatic win; no attackers cannot win. |
| `TestCombatWoundsCarryOver` | Partial damage carries over between rounds, so small forces eventually kill. |
| `TestRetreat` | Attacker retreat threshold; defender retreat gives the attacker the win; simultaneous retreat means the defender holds. |
| `TestDefenseCap` | Defence is capped at 90 %. |
| `TestClassBonus` | Lookup of unit-class bonuses vs other classes. |
| `TestLoot` | Loot without a capacity limit; loot proportional to carry capacity (10 % case); zero capacity loots nothing. Protected amounts are excluded. |
| `TestScout` | Equal scouting power fails and is detected; unopposed scouting succeeds; detection roll; partial intel when the ratio is below 2. |
| `TestTravel` | Distance; travel time; minimum travel time scaled by world speed; speed modifiers. |
| `TestMapDeterministic` | Map generation is deterministic per seed, differs across seeds, has the right size and puts no neutrals on impassable terrain. |
| `TestXPLevels` | Hero XP curve and level-for-XP inverse. |
| `TestRNG` | Seeded RNG is reproducible, in the range [0,1), and `SeedFrom(a,b)` depends on argument order. |

### 2.2 Pack validator tests: `server/internal/pack/pack_test.go` (pure)

| Test | Proves |
|---|---|
| `TestMinimalPackValid` | The minimal pack loads; index lookups work; a hash is computed; defaults are applied (e.g. `combatRounds`). |
| `TestHashStable` | A canonical re-encoding gives an identical hash. Worlds pin a pack by this hash. |
| `TestUnknownFieldRejected` | Strict decoding: an unknown JSON field in a pack is an error. |
| `TestValidationErrors` | 22 sub-cases. Each must be rejected with a specific message:<br>• Identifiers and references: duplicate id; unknown cost resource; unknown requirement target; requirement level out of range; building dependency cycle; quest prerequisite cycle; unit `trainedIn` unknown; neutral guard unknown; migration targets an unknown id.<br>• Values and parameters: invalid modifier target; invalid modifier op; cost growth < 1; non-consecutive neutral levels; start building level too high; `foodResource` unknown; economy tick that does not divide 86400; `formatVersion` unsupported.<br>• Economy rules: a tradable premium resource.<br>• Presentation and localization: missing locale text; invalid palette colour; **icon path traversal** (`../../etc/passwd`); unknown UI screen. |
| `TestShippedPacks` | Every `packs/*/pack.json` in the repo validates (currently `ashen-crowns` and `void-frontier`). |

### 2.3 Game integration tests: `server/internal/game/game_test.go` (real PostgreSQL, fake clock)

Every test calls `newEnv`, which creates a fresh schema, a world on the minimal pack (seed 1, speed 1) and players registered through the real `Register` and `Join`. Newbie protection is removed unless the test is about protection.

| Test | Category | Proves |
|---|---|---|
| `TestJoinStartingState` | Domain | Starting resources, buildings, population and capital flag come from the pack. A second join to the same world gives `name_taken`. A short name gives `invalid_request`. |
| `TestBuildCompletesOfflineAndPersists` | Offline progression, restart, ledger | Build cost and duration (×1.4 race and time factors). The cost is deducted at enqueue, and the level is **not** applied until completion. While the player is offline, the **worker** (`DuePlayers` / `Process`) completes the job. A **fresh `Service`** (empty caches, simulating a process restart) sees the same state. Exactly one `resource_ledger` row has reason `build`. |
| `TestProductionAccruesAndCaps` | Economy over time | +55 wood in 30 min (rate × employment × race bonus), capped at storage 1000 after 48 h. |
| `TestBuildValidation` | Validation | `requirements_not_met`, unknown building gives `invalid_request`, `insufficient_resources`, `queue_full` at 3, and queued jobs chain sequentially (each starts when the previous finishes). |
| `TestCancelOnlyLastWithRefund` | Domain rule | Cancelling a non-last job gives `not_last_in_queue`. Cancelling the last job refunds 80 % and credits it. Only the first upgrade then applies. |
| `TestConcurrentSpendingCannotDoubleSpend` | **Concurrency** | 8 goroutines try to spend the same 150 wood. Exactly 1 succeeds and 7 get `insufficient_resources`. Stock never goes negative, and there is one queued job. |
| `TestConcurrentEventProcessingAppliesOnce` | **Concurrency, exactly-once events** | 3 worker `Process` calls race 3 player reads that settle the same due training event. The units are added once (5, not 10 or 15), and exactly one event row is `done`. |
| `TestResearchAndSpecialisation` | Domain | Research completes after its duration, and the tech modifier changes gross production (×1.15/1.1). **Note:** despite the name, it does not assert exclusive-group (specialisation) blocking. |
| `TestPvEMissionFullCycle` | Missions, time | Units leave home; away upkeep is charged; on arrival the lair battle resolves and the mission returns; the lair is `cleared_until` respawn; rewards and survivors come home; a battle report exists; the `neutrals_cleared` stat is incremented. Skips if the seed has no level-1 den. |
| `TestPvPAttackLootProtectionAndRecall` | PvP rules | A protected target gives `target_protected`. Recalling an outbound attack returns the troops after the elapsed time. The defender sees the incoming attack **without its composition**. The defender is plundered. The **premium resource is not lootable**. Loot ≤ carry capacity. The attacker receives the loot on return. Both sides get the report. |
| `TestScoutAndAllianceRules` | Scouting, alliance | Non-scout units cannot scout. A successful scout report contains resources. Create alliance, invite, accept. Attacking an ally gives `friendly_target`. Reinforce an ally city, troops are stationed, recall them, they return home. A member leaves. The leader alone may leave, which disbands the alliance. |
| `TestMarketConcurrentAcceptAndDelivery` | **Concurrency**, market | 3 buyers race to accept one order: exactly one wins and the others get `order_closed`. Sell escrow is deducted at placement. The seller's proceeds arrive via a delivery event, minus the 2 % fee (rounded up). Cancel refunds the escrow. A premium resource is not tradable. A "gift" order is rejected. NPC exchange arithmetic: 40 × 4/1 × 0.75 = 120. **Weakness:** the gift-order assertion accepts `price_out_of_band` **or** `insufficient_resources`, so the price band is not strictly proven. |
| `TestQuestClaimExactlyOnce` | **Concurrency**, exactly-once | 5 concurrent claims give exactly 1 success and 4 `already_claimed`. The reward is credited once. An incomplete quest gives `quest_incomplete`. |
| `TestHeroesHireSkillEquip` | Heroes | Hero limit (1 per hall level); skill points; `no_skill_points`; equipping an item stacks the class, skill and item modifiers; `hero_busy` blocks unequip during a mission; the hero is idle with XP after returning. |
| `TestTerritoryOccupyCollectAndContest` | Territories | Occupy a free tile; yield of 20/h × 2 h = 40 on collect; another player captures it with `capture=true` and garrisons it. |
| `TestSettleSecondCity` | Settling | A settler founds a second city; the overview lists 2 cities. |
| `TestIsolationBetweenPlayers` | **AuthZ (IDOR)** | Reading a foreign city by id gives `not_found`, not forbidden, so existence is not revealed. Cancelling a foreign job gives `not_found`. Recalling an unknown mission gives `not_found`. |
| `TestStarvationAndDesertionOverLongAbsence` | Offline progression | After 7 days with upkeep far above production: food is 0, population declines and units desert. |
| `TestPackMigrationRenames` | Live pack migration | A dry run reports no missing ids and applies nothing. The real run renames `spear` → `lancer` in city units. A pack that drops an in-use id without a mapping is **refused** (`Missing` is non-empty, `Applied` is false). |
| `TestAccountsAuth` | AuthN | Weak password rejected; duplicate login gives `login_taken`, case-insensitively; a wrong password gives `invalid_credentials`; logins are case-insensitive; logout revokes the token; the token expires after its TTL; guest upgrade; a deleted account cannot log in; argon2id `HashPassword` / `CheckPassword` round-trip. |
| `TestAdminGrantIsAudited` | Admin, audit | A reason shorter than 5 characters is rejected. The grant is credited. Exactly one `admin_audit` row has action `grant`. |
| `TestFailedEventIsParkedNotBlocking` | Resilience | An event of unknown kind fails 5 times, is then **parked** as `failed`, is listed by `FailedEvents`, and **no longer blocks** the player's requests. |

### 2.4 API tests: `server/internal/api/api_test.go`, `openapi_test.go` (HTTP via `httptest`, real PostgreSQL)

| Test | Proves |
|---|---|
| `TestHealthAndAuthRequired` | `/healthz` and `/readyz` return 200. Without a token, or with a garbage token, a protected route returns 401 `unauthenticated`. Non-admins get **403** on `/admin/*`. A player not joined to the world gets 403. **Unknown JSON fields are rejected with 400** (`{"admin":true}` on login, which guards against mass assignment). `/metrics` renders. |
| `TestCORS` | The preflight from an allowed origin returns 204 and echoes the origin. A foreign origin gets **no** `Access-Control-Allow-Origin`. |
| `TestIdempotentCommands` | A missing key gives 400 `idempotency_key_required`. The same key and body give a **byte-identical replay** with `Idempotent-Replay: true`. The same key with a different body gives 422 `idempotency_key_reused`. Only one job exists. **6 concurrent duplicates** with one key all receive the same body and create exactly one job. **Domain errors are stored and replayed** (422, identical bytes). |
| `TestCrossAccountAccessDenied` | A foreign city's read and command both return **404**. A non-numeric id returns 400. A SQL-injection string in a query parameter (`sell=wood' OR 1=1--`) is inert: 200 and an empty list. |
| `TestEndToEnd` | A full HTTP session on the real `ashen-crowns` pack: register → list worlds → pack download with an ETag → join → build → time advance → level 2 → claim the tutorial quest (a second claim gives **409**) → admin-funded keep, warehouse, scriptorium and barracks → research `forestry` (persisted in `players.techs`) → train 30 militia (40 in total) → attack the nearest level-1 wolf den → one battle report, mission `done`, lair cleared if won → at least 3 notifications, mark read, unread = 0 → `DELETE /me` → the old token gets **401**. Skips if the seed has no level-1 den. |
| `TestGuestAndDevPlatform` | Guest sign-in returns `isGuest`. The same `dev` platform identity maps to the same account. An unconfigured platform (`yandex`) returns **501**. |
| `TestOpenAPICoversAllRoutes` (`openapi_test.go`) | Every `"METHOD /path"` registered in `routes.go` has a path and method entry in `server/api/openapi.yaml`, so the contract cannot silently drift. |

### 2.5 Client unit tests: `client/src/__tests__/client.test.ts` (vitest + jsdom)

| Test | Proves |
|---|---|
| reuses the same Idempotency-Key when retrying after a network failure | The first fetch throws and the retry sends the **same** key. |
| does not retry domain errors and surfaces the error code | A 422 is called once and rejects with `{code, status}`. |
| clears the session token on 401 | `api.token` becomes `null`. |
| generates unique, header-safe idempotency keys | The key matches the server regex `^[A-Za-z0-9_-]{8,80}$`. |
| localises with fallback to the default locale | `PackIndex` falls back from ru to en, and an unknown id falls back to the id itself. |
| formats numbers and durations | `fmt` (12.3k, 2.5M) and `duration` (59s, 1h 02m, 1d 1h). |
| translates UI strings and error codes | `makeT`; `errorMessage` uses the localized text or the fallback. |

### 2.6 Browser E2E: `client/e2e/play.mjs` (Playwright Chromium)

It runs against a real server serving the built client (`RF_CLIENT_DIR`), with the `ember-1` world on Ashen Crowns at speed ×50:

1. Desktop at 1360×860: **Play as guest** → world list → join "Ember Vale" with a ruler name and the Ashborn race.
2. City screen: open Lumberyard → **Upgrade** inside the `role=dialog` → ok toast → Escape → wait 3.5 s → **reload** → asserts "Level 2" (server persistence and session restore).
3. Visits 10 screens through the sidebar and screenshots each one.
4. Claims a tutorial quest through the UI, if one is claimable.
5. Map: clicks the centre tile and screenshots it.
6. Mobile at 390×844 (touch, DPR 2), reusing the storage state: restores the last screen, uses the bottom-bar City and **More** sheet, and asserts **no horizontal overflow** (≤ 1 px).
7. **Fails on any `pageerror` or console error.**

Output: screenshots `01-worlds.png` … `16-mobile-more.png` in `<dir>`.

---

## 3. How to run

```bash
# PostgreSQL 16 for tests (no Docker needed); creates realmforge and realmforge_test
scripts/dev-postgres.sh start            # or: make db

# Go: all packages, race detector, DB mandatory (same as `make test-go`)
cd server && RF_REQUIRE_DB=1 go test -race -count=1 ./...
#   pure tests only (no DB):      go test ./internal/rules ./internal/pack
#   one test:                     RF_REQUIRE_DB=1 go test -run TestIdempotentCommands ./internal/api
#   other DB host:                RF_TEST_ADMIN_URL=postgres://user:pw@host:5432/postgres?sslmode=disable
#   show server logs:             RF_TEST_LOG=1

# Client unit tests (same as `make test-client`)
cd client && npm test                     # vitest run (jsdom)
cd client && npm run typecheck            # tsc --noEmit (part of `make lint`)

# Browser E2E: start the stack, then run against it
make dev                                  # db + build client + server on :8080 with ember-1 at x50
cd client && node e2e/play.mjs http://127.0.0.1:8080 /tmp/shots   # or: make e2e SHOTS=/tmp/shots

# Packs and lint
make packs                                # packtool summary of every pack
make lint                                 # go vet, gofmt, tsc

# Backup drill
make verify-backup                        # dump → restore into scratch DB → compare row counts
```

Without `RF_REQUIRE_DB=1`, missing PostgreSQL makes the integration and API tests **skip silently**. A green run without the variable proves only the pure tests. CI must always set it.

---

## 4. Required test categories mapped to existing tests

The original brief text is not stored in the repository. The categories below are the ones the project rules require (AGENT.md "Definition of done") plus the standard categories for a server-authoritative MMO.

| # | Required category | Covered by | Status |
|---|---|---|---|
| C1 | Unit tests of game formulas (economy, combat, travel, loot, XP, map gen) | `rules_test.go` (24) | **Covered** |
| C2 | Content / pack validation, including path traversal | `pack_test.go` (5, plus 22 sub-cases, plus shipped packs) | **Covered** |
| C3 | Integration against a real database, not mocks | `game_test.go`, `api_test.go` via `testutil.DB` | **Covered** |
| C4 | Concurrency: double spend, double claim, double fill, duplicate events | `TestConcurrentSpendingCannotDoubleSpend`, `TestConcurrentEventProcessingAppliesOnce`, `TestMarketConcurrentAcceptAndDelivery`, `TestQuestClaimExactlyOnce`, the concurrent part of `TestIdempotentCommands`, and `-race` | **Covered** for spend, events, market and quests. **Gap:** a concurrent recall racing an arrival; concurrent territory capture by two attackers; a concurrent alliance join beyond the member limit. |
| C5 | Idempotency of commands (replay, conflicting reuse, concurrent duplicates, stored errors) | `TestIdempotentCommands`; client retry test | **Covered** |
| C6 | Time and offline progression (fake clock, worker) | `TestBuildCompletesOfflineAndPersists`, `TestProductionAccruesAndCaps`, `TestStarvation*`, `TestAdvancePathIndependence` | **Covered** |
| C7 | Persistence and restart | Fresh `Service` in `TestBuildCompletesOfflineAndPersists`; reload in `play.mjs` | **Covered**. Process-kill mid-transaction is not tested; atomicity relies on PostgreSQL transactions. |
| C8 | AuthN lifecycle (register, login, logout, expiry, guest, upgrade, delete, ban) | `TestAccountsAuth`, `TestGuestAndDevPlatform`, the end of `TestEndToEnd` | **Mostly covered**. **Gap:** ban (login 403 and existing tokens revoked); upgrade `login_taken`; HTTP logout. |
| C9 | AuthZ and isolation (IDOR returns 404, admin 403, world membership 403) | `TestIsolationBetweenPlayers`, `TestCrossAccountAccessDenied`, `TestHealthAndAuthRequired` | **Covered** for cities, jobs and missions. **Gap:** foreign hero, item, order, report, message and invite ids over HTTP. |
| C10 | Input validation (unknown fields, bad ids, injection, bounds) | `TestHealthAndAuthRequired`, `TestCrossAccountAccessDenied`, `TestBuildValidation`, `TestJoinStartingState` | **Partial**. **Gap:** 413 body limit; out-of-range map viewport (`w > 64`); huge `x/y` values; message length; train quantity bounds. |
| C11 | Abuse controls (rate limits, message cap, newbie protection, attack caps, market band) | Protection in `TestPvPAttackLootProtectionAndRecall`; band only weakly asserted | **Gap:** HTTP 429 from the token bucket; message cap of 60/h; `attack_limit`; `target_too_weak`; strict `price_out_of_band`; `too_many_orders`. |
| C12 | UGC safety (report, block) | none | **Gap:** report then admin `/admin/messages/reported`; block hides the inbox and refuses sending. |
| C13 | Exactly-once event processing and failure parking | `TestConcurrentEventProcessingAppliesOnce`, `TestFailedEventIsParkedNotBlocking` | **Covered**. **Gap:** admin retry of a parked event over HTTP. |
| C14 | Data and pack migration | `TestPackMigrationRenames` | **Covered** for pack migration. SQL migrations are exercised on every test DB build; a forward-only upgrade test from a populated v1 DB does not exist yet. |
| C15 | Admin actions are audited | `TestAdminGrantIsAudited` | **Partial**. **Gap:** ban, retry and migrate audit rows; inspect audit. |
| C16 | API contract coverage | `TestOpenAPICoversAllRoutes` | **Covered** (paths and methods only; schemas are not validated). |
| C17 | End-to-end over HTTP | `TestEndToEnd` | **Covered** |
| C18 | Browser E2E, desktop and mobile | `play.mjs` | **Covered** for the main loop. Not in CI yet. |
| C19 | Client unit tests | `client.test.ts` (7) | **Partial**: api client, i18n, formatting. No component tests. |
| C20 | Accessibility checks | none automated | **Gap** (UI_UX_SPEC §9) |
| C21 | Load / performance / soak | none | **Gap**: planned in §6 |
| C22 | Platform and device (Yandex draft, Android device or emulator) | none | **Gap** |
| C23 | Backup and restore | `scripts/verify-backup.sh` (manual / `make verify-backup`) | **Script exists**. Not run in CI and not run in this pass. |
| C24 | Security headers and CORS | `TestCORS` | **Partial**. **Gap:** asserting `X-Content-Type-Options` and `Referrer-Policy`; `/metrics` exposure policy. |

---

## 5. Known gaps and weaknesses (prioritised)

**High priority**
1. **No load test.** Throughput, lock contention and worker lag are unknown. See §6.
2. **Abuse controls are untested** over HTTP: the 429 token bucket, the 60/h message cap, the attack caps, strict assertion of the market price band, `too_many_orders`.
3. **UGC report and block flow is untested**, yet it is mandatory for Google Play (GP-16).
4. **No Android device or emulator test** and no Yandex draft-mode test (CSP, `LoadingAPI.ready`, pause on `game_api_pause`). The Android project has not been generated yet.
5. **The browser E2E is not in CI.** It needs a server, the built client and Playwright.

**Medium priority**
6. **Two tests can silently skip on a map seed without a level-1 lair:** `TestPvEMissionFullCycle` and the mission part of `TestEndToEnd`. Fix: insert a lair tile in setup instead of `t.Skip`.
7. **No component or a11y tests.** Vitest covers only pure modules. There is no keyboard, focus or axe test.
8. **No fuzzing** of the JSON decoders, the pack loader or the rules functions. Recommended: `FuzzPackLoad`, `FuzzAdvanceCity`, `FuzzResolveBattle` with invariants for unit conservation and no negative stock.
9. **Ban, upgrade conflict, admin retry, migrate and inspect audit, and HTTP logout** are untested (C8, C15).

**Low priority**
10. `TestResearchAndSpecialisation` does not test specialisation (the exclusive group).

---

## 6. Planned load test

### 6.1 Goals

Measure p50/p95/p99 latency, throughput, PostgreSQL lock waits, transaction retries (`rf_tx_retries_total`) and worker lag (`rf_events_overdue`) under a realistic MMO mix. Find the player count at which one app instance plus one PostgreSQL instance still meets the targets in §7.

### 6.2 Tooling

- **Primary: k6** (`loadtest/k6/*.js`, to be created). It has scenarios, thresholds and an HTTP-level view; the setup creates accounts via `POST /auth/register` or `/auth/guest`.
- **Secondary: a Go load generator** (`server/cmd/loadgen`, to be created) for invariant checks k6 cannot do. It reuses `game.Service` to verify the database state after the run.

### 6.3 Environment

- Production-like build (`RF_ENV=production`, `RF_DEV_PLATFORM_AUTH=false`, TLS-terminating proxy), on a dedicated PostgreSQL 16 with the same `MaxConns=20` pool.
- `RF_RATE_RPS` and `RF_RATE_BURST` stay at production values (20 / 60). k6 spreads accounts across source IPs (or the test bypasses the proxy) so that the per-IP auth bucket does not dominate.
- Worlds:
  - World A at speed ×1, with 5 000 pre-seeded players (seed script via the API).
  - World B at speed ×50, to stress event density.

### 6.4 Scenarios

| ID | Scenario | Shape |
|---|---|---|
| L1 | **Steady polling.** Each virtual player runs `GET /worlds/{w}/overview` and `GET /cities/{c}` every 15 s. Each of these takes the player row lock and settles. | Ramp 0 → 2 000 VUs over 10 min, hold 30 min (≈ 267 req/s) |
| L2 | **Command mix.** Per VU every 20 s, weighted: build 30 %, train 25 %, research 10 %, mission 15 %, market place/accept 10 %, message 5 %, quest claim 5 %. Every command carries a fresh `Idempotency-Key`; 2 % are resent with the same key (replays). | Runs on top of L1 |
| L3 | **Hot spots.** 200 VUs accept the same market order; 100 attackers target one city, all arriving in the same second; 50 alliance members message the alliance. | Spike, 2 min |
| L4 | **Event backlog.** At speed ×50, 50 000 jobs fall due within 1 min; measure worker drain time and `rf_events_overdue`. | One-shot |
| L5 | **Auth storm.** 300 logins per second for 60 s. Argon2id uses 19 MiB per hash, so measure RSS and latency. | Spike |
| L6 | **Soak.** L1 + L2 at 60 % load for 8 h: check memory and the size of `idempotency_keys` (pruned after 24 h) and `sessions` (never pruned). | Long |

### 6.5 Post-run invariant checks (Go loadgen or SQL)

- No negative amount in any `cities.resources` value.
- `game_events`: no `status='pending' AND due_at < now() - 10s` after the drain, and no `failed` rows.
- Each `(kind, ref_id, due_at)` was processed once (guaranteed by the unique key; verify `done` counts against the jobs that completed).
- Each market order is filled at most once: no order has more than one `market_buy` ledger row.
- Quest claims: `quest_progress` has at most one row per `(player, quest)`.
- Replayed commands produced no extra effects: the number of jobs equals the number of distinct keys sent for build and train.

### 6.6 Monitoring during the run

- `/metrics`: `rf_commands_total`, `rf_idempotent_replays_total`, `rf_tx_retries_total`, `rf_events_processed_total`, `rf_events_failed_total`, `rf_events_overdue`, `rf_db_conns_*`.
- PostgreSQL: `pg_stat_activity` wait events (`Lock:tuple`, `Lock:transactionid`), `pg_locks`, `pg_stat_statements` top queries.
- Host: CPU and RSS of the app (watch argon2), PostgreSQL I/O.

---

## 7. Acceptance criteria

### 7.1 Every merge (CI gate)

1. `RF_REQUIRE_DB=1 go test -race -count=1 ./...` passes with **zero skips** in `api` and `game`. A skip counts as a failure in CI.
2. `npm run typecheck` and `npm test` pass; `go vet` and `gofmt` are clean (`make lint`).
3. `TestOpenAPICoversAllRoutes` passes: every new route is documented.
4. A new command route uses `cityCmd` or `playerCmd` and has a test for success and for at least one domain error. A change that touches resources or ownership adds a concurrency test.

### 7.2 Every release candidate

5. `play.mjs` passes against the RC build, desktop and mobile, with no console errors.
6. `make verify-backup` passes against a staging copy.
7. The tests listed under "high priority" in §5 exist and pass. Before store submission, UGC report and block and the abuse limits are mandatory.

### 7.3 Load targets (release gate for an open beta, single instance)

| Metric | Target |
|---|---|
| L1 + L2 at 2 000 concurrent players: p95 for GET overview/city | **< 150 ms** |
| L1 + L2 at 2 000 concurrent players: p95 for a command | **< 300 ms** |
| L1 + L2 at 2 000 concurrent players: p99 for any request | **< 1 s** |
| L1 + L2 at 2 000 concurrent players: 5xx rate | **< 0.1 %** |
| L1 + L2 at 2 000 concurrent players: `rf_tx_retries_total` | **< 1 %** of commands |
| L3 | Exactly one market winner; no deadlock-induced 500s; every attack resolves |
| L4 | Overdue events drained within **60 s**; `rf_events_failed_total` = 0 |
| L5 | App RSS < 2 GiB; login p95 < 500 ms. If not met, add an argon2 concurrency semaphore (SECURITY.md S-4). |
| L6 | No unbounded growth except documented tables; RSS stable within ±10 % |
| All runs | All §6.5 invariants hold |

### 7.4 Platform gate (before store submission)

8. **Yandex draft:**
   - The game loads with CSP enforced, with no external hosts beyond the declared API.
   - `LoadingAPI.ready` fires only when the game is interactive.
   - Polling stops on `game_api_pause`.
   - Guest play works with no login prompt.
9. **Android:**
   - A closed-track build on an API 36 emulator and on one physical device.
   - Back button behaviour, safe areas and offline/reconnect.
   - In-app account deletion.
   - Report and block are visible in mail.
