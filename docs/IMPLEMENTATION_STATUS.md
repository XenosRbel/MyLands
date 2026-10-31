# Implementation status

Last updated: 2026-11-01. Legend: ✅ implemented & tested · 🟡 partial · ⛔ not implemented.

| Subsystem | Status | Where | Tests |
|---|---|---|---|
| Shared Codex / Claude Code harness | 🟡 implementation verified; baseline lint and live activation pending (T-032) | scripts/harness/, .codex/, .claude/settings.json | 30 harness tests; real hook commands and red Stop |
| Principal Architect role | 🟡 shared instructions and Codex / Claude / GitHub Copilot profiles added; activation pending (T-033) | .agents/roles/, .codex/agents/, .claude/agents/, .github/agents/ | TOML/YAML parsing, local links, harness tests |
| Accounts: register/login (argon2id), guest, upgrade guest, logout, sessions TTL/revocation, ban, delete (anonymise) | ✅ | game/accounts.go | TestAccountsAuth, TestGuestAndDevPlatform, TestEndToEnd |
| Platform identity sign-in | 🟡 `dev` only; Yandex/Google verification ⛔ | accounts.go, platform/adapter.ts | TestGuestAndDevPlatform |
| Worlds: pinned pack, deterministic map, speed multiplier, join with race, spawn search, newbie protection | ✅ | world.go, rules/world.go | TestJoinStartingState, TestMapDeterministic |
| Time model: lazy settlement, ordered events, exactly-once, worker, parked failures, restart recovery, long absence | ✅ | settle.go, rules/simulate.go | TestBuildCompletesOfflineAndPersists, TestConcurrentEventProcessingAppliesOnce, TestFailedEventIsParkedNotBlocking, TestAdvancePathIndependence, TestStarvationAndDesertionOverLongAbsence |
| Resources: production, consumption, storage caps / overflow decay, protection, ledger | ✅ | rules/economy.go | TestProductionAccruesAndCaps, TestAccrueCapsAndFloors, TestOverflowDecay |
| Population: housing, employment, idle tax, growth modes, starvation | ✅ | rules/economy.go | TestRatesEmploymentAndStorage, TestTickPopulation, TestIdleTaxAndUpkeep |
| Buildings: costs/times, requirements, race limits, sequential queue, cancel last with refund | ✅ | city.go | TestBuildValidation, TestCancelOnlyLastWithRefund, TestConcurrentSpendingCannotDoubleSpend |
| Research: tree, requirements, player-wide modifiers, exclusive specialisations | ✅ | city.go | TestResearchAndSpecialisation |
| Army: training, upkeep in configurable resource, mission upkeep ×, desertion | ✅ | city.go, rules | TestDesertionAndIdleMode, TestPvEMissionFullCycle |
| Missions: attack, scout, transport, reinforce (incl. relocation), settle, occupy; recall; travel time | ✅ | missions.go | TestPvE…, TestPvP…, TestScoutAndAllianceRules, TestSettleSecondCity, TestTerritory… |
| Combat: deterministic rounds, class bonuses, defense cap, walls, wounds, retreat, faction penalties, loot, reports | ✅ | rules/combat.go, missions.go | TestCombat*, TestRetreat, TestLoot, TestPvPAttackLootProtectionAndRecall |
| Scouting with seeded detection probability | ✅ | rules/combat.go | TestScout, TestScoutAndAllianceRules |
| PvE: neutrals by level, respawn, rewards, item drops | ✅ | missions.go | TestPvEMissionFullCycle, TestEndToEnd |
| PvP protections: newbie protection (ends on attack), friendly-fire block, daily attack cap, score floor | ✅ | missions.go | TestPvP…, TestScoutAndAllianceRules |
| Territories: occupy, cap, yield accrual & collect, capture with garrison battle, abandon | ✅ | missions.go, economy.go | TestTerritoryOccupyCollectAndContest |
| Second and further cities (settlers, city cap) | ✅ | missions.go | TestSettleSecondCity |
| Heroes: hire, cap, XP/levels, skill points, items/equipment, wounds, battle modifiers | ✅ | heroes.go | TestHeroesHireSkillEquip |
| Heroes: spells/mana, capture/ransom | ⛔ | — | — |
| Market: escrow orders, accept (row lock), cancel refund, fee, price band, delivery events; NPC exchange | ✅ | economy.go | TestMarketConcurrentAcceptAndDelivery |
| Premium auction | ⛔ | — | — |
| Alliances: create (cost), invite/accept/decline, roles, kick, leave/disband, diplomacy, reinforcement of allies | ✅ | social.go | TestScoutAndAllianceRules |
| Joint attack | ⛔ | — | — |
| Messages: private, alliance broadcast, report, block; notifications | ✅ | social.go | TestEndToEnd (notifications) |
| Quests & tutorial chain, exactly-once claim | ✅ | economy.go | TestQuestClaimExactlyOnce, TestEndToEnd |
| Rankings | ✅ | economy.go | (manual/E2E) |
| Admin: inspect, grant (audited), failed events + retry, ban, audit log, reported messages, pack migration | ✅ | admin.go | TestAdminGrantIsAudited, TestPackMigrationRenames |
| API: auth, idempotency, rate limit, CORS, strict JSON, OpenAPI coverage | ✅ | api/ | TestIdempotentCommands, TestCORS, TestHealthAndAuthRequired, TestCrossAccountAccessDenied, TestOpenAPICoversAllRoutes |
| Game Pack engine: format, validator, CLI, 2 packs | ✅ | pack/, packs/ | pack_test.go (22 negative cases + shipped packs) |
| Web client: all 14 screens, responsive, i18n EN/RU, server-time timers, idempotent commands | ✅ | client/ | vitest (7), browser E2E e2e/play.mjs |
| Pack art: isometric city scene, building sprites, unit/hero portraits, item & resource icons, textured map (Resource Pack v0.1) | ✅ (UI frames, neutral sites, 10 unit portraits missing) | client/src/screens/City.tsx, Map.tsx; packs/ashen-crowns/assets; scripts/import_art.py | TestShippedPacks (asset files exist), browser E2E |
| Yandex HTML5 build + adapter | 🟡 build/zip ✅; SDK identity/moderation ⛔ | client/scripts/yandex-zip.mjs | manual |
| Android (Capacitor) | 🟡 config ✅; native project/AAB ⛔ (no SDK) | client/capacitor.config.ts | — |
| Backup/restore | ✅ (logical) | scripts/*.sh | verify-backup executed |
| Docker | 🟡 files written, not executed | Dockerfile, docker-compose.yml | — |
| Load testing | 🟡 baseline: 200 players, 648 req/s, 0 errors (docs/reports/LOAD_TEST_2026-10-31.md) | server/cmd/loadtest | manual run |

## Verification: agent harness, 2026-11-01

`make harness-test` passed (30 tests). `make test` passed with PostgreSQL 16, required DB tests and Go race detector.
Client typecheck, 10 client tests, Go vet and both Game Pack validations passed.
`make harness-check` remains blocked by existing Go formatting in three files; see T-032.
Live client activation and Codex hook trust remain unverified. No gameplay or API behavior changed.

## Previous verification run (2026-10-31)

```
go vet ./... && RF_REQUIRE_DB=1 go test -race -count=1 ./...   → api ok, game ok (22 tests), pack ok, rules ok
npm run typecheck && npm test                                   → 7/7
npm run build                                                   → dist 306 KB JS (93 KB gzip)
node e2e/play.mjs http://127.0.0.1:8080 <dir>                   → OK (guest → join → build → reload persisted → all screens → quest claim → mobile no overflow)
scripts/verify-backup.sh <demo db>                              → backup verified
```
