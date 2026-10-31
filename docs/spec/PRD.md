# Product Requirements Document: Ashen Crowns (Realmforge engine)

| Field | Value |
|---|---|
| Owner | Agent B (Game Designer) |
| Status | DRAFT v1 |
| Date | 2026-10-31 |
| Code baseline | commit `dfa660e`, re-checked at `ed6acc7` (no game-rule changes) (Go server `server/`, React client `client/`) |
| Content baseline | `packs/ashen-crowns/pack.json` v1.0.0; demo `packs/void-frontier/pack.json` |
| Related | `docs/spec/GAME_RULES.md` (normative formulas), `docs/spec/GAME_DESIGN_DOCUMENT.md`, `docs/spec/PLATFORM_REQUIREMENTS.md`, `docs/research/MECHANICS_CATALOG.md`, `docs/research/UNKNOWN_AND_ASSUMPTIONS.md`, `docs/research/SOURCE_REGISTER.md` |

**Status words used in this document.** *Implemented* = present in code at the baseline commit; the test that exercises it is named when one exists. *Implemented (untested)* = in code, no dedicated automated test. *Partial* = some of the requirement exists. *Planned* = not in code. Statements about the reference game carry research IDs; all such evidence came from search excerpts only (no page opened directly), so it is never stated as fact.

---

## 1. Product summary

**Ashen Crowns** is a persistent, browser-and-mobile, massively multiplayer strategy game about rebuilding a fallen realm "beneath a shattered sky". Players grow cities, research, train armies, clear monster lairs, hunt *starglass* in fallen craters, hold territory and forge alliances against rival crowns (pack `meta.description`).

It runs on **Realmforge**, our own engine: a server-authoritative Go backend whose rules are pure functions (`server/internal/rules`) driven by a declarative **game pack** (JSON). The same engine runs a second, unrelated demo pack (*Void Frontier*), which proves that setting, rules and balance are data, not code.

## 2. Positioning and IP rules

1. **Original IP.** Ashen Crowns is inspired by a *genre*: long-session, timer-based "city-builder plus army" browser MMOs. Our research studied the public systems of one such game, *My Lands: Black Gem Hunting*, only to understand genre mechanics (see `MECHANICS_CATALOG.md`).
2. **Must not be marketed as My Lands** or as a sequel, clone, remake, "spiritual successor", private server or replacement of it. Store pages, ads, trailers, metadata, keywords and press material must not mention that game or its trademarks (titles, race names, "Black Gem", unit names, building names).
3. **No reuse of assets or texts.** All names, texts and data in our packs are original (`meta.license`: "All texts and data original to this project (CC BY 4.0)"). Icons are drawn in code; fonts are SIL OFL (`theme.assetsLicense`).
4. **Genre mechanics are fair to use; expression is not.** Where a rule shape came from research (for example gold upkeep with desertion, ECO-015), our numbers, names and presentation are our own.
5. **Differentiators we can claim:** no pay-to-win (no real-money sales at all in MVP), fair offline progression, short readable timers at the start, cross-platform play (web, Yandex Games, Android), two languages (EN, RU).

## 3. Audience

| Segment | Description | What they need from us | Label |
|---|---|---|---|
| Primary | Adults 25–45 who like slow, persistent strategy and check in several times a day on desktop or phone | Short sessions, meaningful offline progress, social standing | PROPOSED |
| Secondary | Former players of the genre put off by pay-to-win and slow starts (genre criticism of slow builds and pay-to-win is recorded in research, SRC-A1-39) | Fair economy, fast first hour | PROPOSED (motivation from GAME_OVERVIEW "Reception") |
| Tertiary | Yandex Games portal players (RU-speaking) | Guest play, instant start, RU locale | PROPOSED; platform rules in `PLATFORM_REQUIREMENTS.md` |

## 4. Core loop

1. **Produce**: staffed buildings yield timber, grain, stone, iron, gold and lore into the city's own stock.
2. **Grow**: housing fills with people; people staff producers; idle people pay taxes at the Trading Post.
3. **Build and research**: spend resources (and lore) on building levels and technologies within queue limits.
4. **Train**: armies cost resources and a gold wage every hour (double while away).
5. **Act on the map**: scout, clear neutral lairs for resources/items/starglass, occupy territories for yield, raid other players, reinforce allies, found new cities.
6. **Cooperate and compete**: alliances, diplomacy, market trade, mail, rankings.
7. **Return**: claim quests, collect territory yield (capped at 24 h), queue the next jobs.

Loop details and numbers: `GAME_DESIGN_DOCUMENT.md` §3.

## 5. Retention principles

| # | Principle | How the product realises it | Status |
|---|---|---|---|
| R1 | Progress while away, without loss from simply being offline | Economy is integrated on access; result is independent of how often a city is read (`rules/simulate.go AdvanceCity`, test `TestAdvancePathIndependence`). Overflow decays slowly (1%/h) instead of being cut. | Implemented |
| R2 | Something to do in every session | Three queues (build 3, research 2, train 5), quests to claim, territory yield to collect | Implemented |
| R3 | Reasons to come back on a schedule | Territory yield stops at 24 h; neutrals respawn in 4–24 h; jobs finish | Implemented |
| R4 | Fast, readable first hour | Tutorial chain t01–t10 with sub-3-minute early builds (GDD §15) | Implemented (content) |
| R5 | Safety for newcomers | 72 h shield; ends early only if the player attacks another player | Implemented |
| R6 | Social glue | Alliances up to 100, alliance mail, reinforcement, diplomacy | Implemented |
| R7 | Long-term goals | Keep to level 25, up to 6 cities (I-01), 7 heroes to level 30, specialisation Paths, rankings | Implemented (content depth PROPOSED) |
| R8 | Fairness | No real-money advantage; premium currency earned in-game only and not tradable or lootable | Implemented (no sink yet, I-09) |
| R9 | Live events | — | Planned (not in MVP) |

## 6. Game modes

| Mode | Definition | How it is configured | Status |
|---|---|---|---|
| Standard | World speed 1 | `RF_WORLDS="slug:packDir[:Name]"` or speed `1`; pack `params.worldSpeed` 1 | Implemented |
| Speed world | Same pack, all timers and flows scaled by a multiplier | 4th field of `RF_WORLDS` (`slug:packDir:Name:speed`), stored in `worlds.speed`; effective speed = pack `worldSpeed` × world speed (`game/service.go Pack`) | Implemented (scaling unit-tested in `rules TestCostAndDuration`, `TestTravel`) |
| Peaceful | Listed as a known mode label | `pack/validate.go KnownModes` accepts `peaceful`; no rule changes attached | Planned |

What the world speed multiplier scales: build/research/train durations, travel time and its 900 s minimum, all production/consumption/upkeep rates, economy-tick effects (growth, starvation, desertion, overflow decay), neutral respawn, territory yield accrual. What it does **not** scale today: newbie protection (72 real hours), the 24 h attack window, hero wound hours (see `UNKNOWN_AND_ASSUMPTIONS.md` I-08). AC declares `modes: ["standard", "speed"]`.

## 7. Functional requirements

| ID | Requirement | Status | Code | Test evidence |
|---|---|---|---|---|
| FR-01 | Accounts: register, log in/out, guest play, upgrade guest to full account, platform login, change locale, delete account | Implemented | `game/accounts.go`; `api/routes.go` `/auth/*`, `/me` | `TestAccountsAuth`, `TestGuestAndDevPlatform` |
| FR-02 | World list; join a world with a name and a race; one player per account per world; capital placed near the centre, ≥ 3 tiles from other cities | Implemented | `game/world.go Join`, `findSpawn` | `TestJoinStartingState` |
| FR-03 | Pack download with ETag for the client | Implemented | `GET /worlds/{wid}/pack` | `TestEndToEnd` |
| FR-04 | City economy: production with employment, idle tax, storage, protection, housing, food, upkeep; offline integration | Implemented | `rules/economy.go`, `rules/simulate.go` | `TestRatesEmploymentAndStorage`, `TestProductionAccruesAndCaps`, `TestAdvancePathIndependence` |
| FR-05 | Population growth (`free_housing` or `idle` mode), starvation, desertion, overflow decay | Implemented | `rules/economy.go TickStep` | `TestTickPopulation`, `TestDesertionAndIdleMode`, `TestOverflowDecay`, `TestStarvationAndDesertionOverLongAbsence` |
| FR-06 | Construction queue (max 3/city), requirements, cancel last job with 80% refund | Implemented | `game/city.go Build`, `CancelJob` | `TestBuildValidation`, `TestCancelOnlyLastWithRefund`, `TestBuildCompletesOfflineAndPersists` |
| FR-07 | Research queue (max 2/player), player-wide tech effects | Implemented | `game/city.go Research` | `TestResearchAndSpecialisation` |
| FR-08 | Mutually exclusive specialisations (`exclusiveGroup`) | Implemented (untested) | `game/city.go exclusiveTaken` | — |
| FR-09 | Unit training queue (max 5/city), batch completion | Implemented | `game/city.go Train` | `TestEndToEnd` (API), game tests |
| FR-10 | Missions: attack, scout, transport, reinforce, settle, occupy; recall/withdraw | Implemented | `game/missions.go` | `TestPvEMissionFullCycle`, `TestPvPAttackLootProtectionAndRecall`, `TestScoutAndAllianceRules`, `TestSettleSecondCity`, `TestTerritoryOccupyCollectAndContest` |
| FR-11 | Deterministic combat (20 rounds in AC), class counters, walls, defense cap, retreat thresholds, loot with protection | Implemented | `rules/combat.go` | `TestCombatDeterministicAndConservative`, `TestCombatWallHelpsDefender`, `TestRetreat`, `TestDefenseCap`, `TestClassBonus`, `TestLoot` |
| FR-12 | PvP protections: newbie shield, alliance/NAP block, same-race −50% / same-faction −25% attack, max 2 attacks per target per 24 h, score floor (off in AC), max 3 reinforcing players | Partial tests | `game/missions.go checkPvP`, `factionMods`, `missionArrive` | Shield and friendly-block tested; faction penalty, attack cap and reinforcing cap untested (I-14) |
| FR-13 | Espionage with intel levels and detection | Implemented | `rules/combat.go Scout`; `game/missions.go scoutCity` | `TestScout`, `TestScoutAndAllianceRules` |
| FR-14 | PvE neutrals with levels, guards, rewards, item drops, respawn | Implemented | `game/missions.go battleNeutral`; `rules/world.go GenerateMap` | `TestPvEMissionFullCycle`, `TestMapDeterministic` |
| FR-15 | Territories: occupy, garrison, yield and collect, abandon, capture by attack | Implemented (abandon untested) | `game/economy.go Collect`, `Abandon`; `game/missions.go occupy`, `battleTerritory` | `TestTerritoryOccupyCollectAndContest` |
| FR-16 | Additional cities via settlers, capped by keep level | Implemented | `game/missions.go settle`, `cityCap` | `TestSettleSecondCity` |
| FR-17 | Heroes: hire (War Hall), classes, skills, items/equipment, XP and levels, wounds | Implemented | `game/heroes.go` | `TestHeroesHireSkillEquip`, `TestXPLevels` |
| FR-18 | Alliances: create (cost), invite, accept/decline, leave, roles, kick, diplomacy (ally/nap/war/neutral) | Implemented (relations partly tested) | `game/social.go` | `TestScoutAndAllianceRules` |
| FR-19 | Mail: private and alliance; block; report a message; rate limit 60/h | Implemented (untested) | `game/social.go SendMessage`, `Block`, `ReportMessage` | — |
| FR-20 | Market: escrowed orders, accept, cancel, fee, price band, open-order limit; NPC exchange | Implemented | `game/economy.go` | `TestMarketConcurrentAcceptAndDelivery` |
| FR-21 | Battle/scout reports, notifications, unread counters, overview | Implemented | `game/social.go Reports`, `Notifications`; `game/admin.go Overview` | `TestEndToEnd` (notifications) |
| FR-22 | Rankings by score; alliance list with summed score | Implemented (untested) | `game/economy.go Rankings`; `game/social.go Alliances` | — |
| FR-23 | Quests with chains and manual one-time claim; tutorial flag; client tutorial bar | Implemented | `game/economy.go Quests`, `ClaimQuest`; `client/src/screens/Shell.tsx TutorialBar` | `TestQuestClaimExactlyOnce` |
| FR-24 | Game packs: validated, hashed, versioned, pinned per world; migrations for live worlds | Implemented | `pack/`; `game/admin.go MigrateWorldPack` | `pack` tests, `TestPackMigrationRenames` |
| FR-25 | Localisation: EN and RU for all pack texts and UI | Implemented | pack `locales`; `client/src/i18n` | `TestShippedPacks` (validation) |
| FR-26 | Admin: inspect player, grant resources (audited), failed-event list/retry, ban, audit log, reported messages, pack migration | Implemented | `game/admin.go`; `/api/v1/admin/*` | `TestAdminGrantIsAudited`, `TestFailedEventIsParkedNotBlocking` |
| FR-27 | Clients: responsive web client; Yandex Games build; Android build via Capacitor | Partial | `client/` (`build:yandex`, `android:sync` scripts) | client vitest + `client/e2e/play.mjs`; store submission items open (`PLATFORM_REQUIREMENTS.md` §5) |
| FR-28 | Starglass sinks (non-pay-to-win services) | Planned | — | — |
| FR-29 | Live events (timed world events, tournaments) | Planned | — | — |
| FR-30 | Real-time chat | Planned | — | — |

Mechanics from the reference genre that are explicitly **out of MVP scope** (healers/resurrection, hero capture/ransom, spells/mana, premium-currency auction, deposit depletion, cell-based city layout, villages, slave capture, building destruction, joint attack, clan castle, tournaments, gates, exploration/fog of war): see `MECHANICS_CATALOG.md`.

## 8. Non-functional requirements

| ID | Requirement | Status | Evidence |
|---|---|---|---|
| NFR-01 | **Server authority.** The client never decides outcomes; every command is validated server-side. | Implemented | All mutations go through `game.Service` under a player row lock (`game/settle.go lockPlayer`). |
| NFR-02 | **Determinism and reproducibility.** Map, scouting rolls and drops come from explicit seeds; combat is deterministic. | Implemented | `rules/world.go RNG`, `SeedFrom`; `TestMapDeterministic`, `TestRNG`, `TestCombatDeterministicAndConservative` |
| NFR-03 | **Exactly-once effects.** Scheduled events, quest claims and market fills apply once even under concurrency. | Implemented | `TestConcurrentEventProcessingAppliesOnce`, `TestConcurrentSpendingCannotDoubleSpend`, `TestMarketConcurrentAcceptAndDelivery`, `TestQuestClaimExactlyOnce` |
| NFR-04 | **Idempotent commands** via `Idempotency-Key` (24 h retention). | Implemented | `api/http.go`; `TestIdempotentCommands`; cleanup in `RunWorker` |
| NFR-05 | **Fault isolation.** A failing event is retried and parked after 5 attempts without blocking the player. | Implemented | `game/settle.go Process`; `TestFailedEventIsParkedNotBlocking` |
| NFR-06 | **Isolation between players and worlds.** No cross-account access; foreign city IDs are not revealed. | Implemented | `TestCrossAccountAccessDenied`, `TestIsolationBetweenPlayers` |
| NFR-07 | **Security.** Hashed passwords, token sessions (default TTL 720 h), per-IP rate limit (20 rps, burst 60), CORS allow-list, dev platform auth forbidden in production, TLS DB in production. | Implemented | `game/accounts.go`; `platform/platform.go LoadConfig`; `TestCORS`, `TestHealthAndAuthRequired` |
| NFR-08 | **Observability.** `/healthz`, `/readyz`, `/metrics`, request IDs, structured logs. | Implemented | `api/routes.go`, `api/http.go` |
| NFR-09 | **Data durability.** PostgreSQL with migrations; backup, restore and backup-verification scripts. | Implemented (scripts) | `server/migrations`, `scripts/backup.sh`, `restore.sh`, `verify-backup.sh` |
| NFR-10 | **Content safety.** Message report and player block (store requirement for user-generated content); admin review queue. | Implemented | `ReportMessage`, `Block`, `ReportedMessages` |
| NFR-11 | **Privacy.** In-app account deletion. | Implemented | `DELETE /api/v1/me`; store web-link deletion still open (`PLATFORM_REQUIREMENTS.md` §3.5) |
| NFR-12 | **Localisation** EN/RU with fallback to the default locale. | Implemented | `pack.LocText.Get` |
| NFR-13 | **Mobile-friendly UI** at phone width. | Implemented (per client commit message "mobile layout fixes") | `client/src/styles.css`; E2E script |
| NFR-14 | **Performance targets**: p95 command latency < 300 ms at 1,000 concurrent players per world; settle of a 6-city player < 50 ms. | Planned (targets PROPOSED; not measured) | — |
| NFR-15 | **Platform compliance** (Yandex Games, Google Play): pause on focus loss, guest play, SDK-only payments/ads, target API 36, Data safety. | Partial | `client/src/platform/adapter.ts`; checklist in `PLATFORM_REQUIREMENTS.md` |
| NFR-16 | **Content-as-data.** New settings and balance changes ship as packs without code changes; validator rejects unknown fields and broken references. | Implemented | `pack/validate.go`; `TestUnknownFieldRejected`, `TestValidationErrors`, `TestShippedPacks` |

## 9. MVP constraints

1. **No monetisation.** No real-money purchases, no ads, no subscriptions. The premium currency *starglass* is earned in-game only (start 20; quests; Ancient Barrow and Starfall Crater neutrals), is not tradable, lootable or transportable. No sink exists yet (I-09).
2. **One content pack** for players (Ashen Crowns v1.0.0); Void Frontier is a demo of engine flexibility only.
3. **Two locales**: EN (default) and RU.
4. **Single server process** with an in-process event worker and one PostgreSQL database. Horizontal scaling is out of scope.
5. **Small worlds**: 48 × 48 map; capacity is limited by free spawn tiles (`world_full`).
6. **No real-time push**: the client polls; there is no chat.
7. **No live-ops tooling** beyond admin grant, ban, event retry and pack migration.
8. Balance values are **PROPOSED** and unvalidated by playtests (`UNKNOWN_AND_ASSUMPTIONS.md` §3).

## 10. Definition of Done

### 10.1 For any feature

- Rules live in `server/internal/rules` (pure) or `server/internal/game` (stateful) and are configurable through pack data where they are balance-relevant.
- Unit test for every formula; integration test for every rule code a player can hit (`rule("…")`).
- API endpoint documented and covered by at least one API test or the E2E script.
- Pack validator updated for any new field; both shipped packs validate (`TestShippedPacks`).
- EN and RU texts present.
- `MECHANICS_CATALOG.md` row updated when the feature corresponds to a research ID.

### 10.2 For the MVP release

| # | Criterion | Current state |
|---|---|---|
| D1 | All Go tests pass with `RF_REQUIRE_DB=1` | Passing locally on 2026-10-31 (`rules`, `pack`, `game`, `api`) |
| D2 | Tests added for untested rules: `attack_limit`, `too_many_reinforcing_players`, `specialisation_taken`, faction penalties, mail, rankings, abandon | Open (I-14) |
| D3 | Code follow-ups decided: city cap (I-01), market fee modifier (I-03), transport targets (I-07), speed-scaling of limits (I-08) | Open |
| D4 | A new player completes tutorial quests t01–t10 in the first hour at world speed 1 without help (playtest, 5 testers) | Open |
| D5 | 72-hour soak on a test world with simulated players: no failed events parked, no negative stocks | Open |
| D6 | Platform checklists in `PLATFORM_REQUIREMENTS.md` §5 completed for the target store(s) | Open |
| D7 | Store listing reviewed against §2 (no reference-game names or trademarks) | Open |
| D8 | Privacy policy and in-app/web account deletion live | Partial (in-app only) |

## 11. Risks

| Risk | Mitigation |
|---|---|
| IP confusion with the reference game | §2 rules; original names; legal review of store text |
| Balance untested (all values PROPOSED) | Telemetry from resource ledger (`resource_ledger` table); pack hot-fix via migration |
| Defence bias (defenders never retreat, I-12) frustrates attackers | Playtest; tune `defenderRetreatPct` or wall values |
| Premium currency with no sink (I-09) | Design sinks (GDD §17) before players hoard large balances |
| Yandex/Google moderation | Follow `PLATFORM_REQUIREMENTS.md`; guest play already exists |

## 12. Open questions

1. Should starglass sinks be cosmetic only, or also convenience (for example an extra queue slot), given the fairness promise?
2. Is a world ever "finished"? (No end condition exists.) Seasonal worlds would need an end event and rewards.
3. Should speed worlds shorten the newbie shield (I-08)?
