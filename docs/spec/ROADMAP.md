# Implementation roadmap (stages, dependencies, exit criteria)

Current status per stage is tracked in [`../IMPLEMENTATION_STATUS.md`](../IMPLEMENTATION_STATUS.md); the short-term plan in [`../ROADMAP.md`](../ROADMAP.md).

| Stage | Scope | Depends on | Exit criteria | Status |
|---|---|---|---|---|
| 0 Research & spec | Public-source research, mechanics catalogue, unknowns, full spec set | — | All research/spec docs exist; unknowns separated | **Done with caveat**: evidence only via search excerpts (original sites blocked by network policy) |
| 1 Foundation | Go server, migrations, config, logs, health, React client, Docker/compose, Makefile | 0 | Server + client build reproducibly; health checks; tests run | **Done** (Docker files not executed: no daemon in authoring env) |
| 2 Core loop | Accounts, world join, city, resources, production, buildings, queue, time model, persistence, tutorial quests | 1 | New player can build, wait, reload and see persisted result (browser E2E) | **Done** |
| 3 Empire | Population/housing/employment, research tree, specialisations, many buildings, second city | 2 | Tests for each; settle second city | **Done** |
| 4 Map & PvE | Map generation, viewport, neutrals, missions, travel, rewards, reports | 3 | PvE mission full cycle tested | **Done** |
| 5 Army & PvP | Units, training, upkeep/desertion, scouting, battles, loot, reports, territory capture, protections, caps | 4 | PvP & scouting tests; concurrency tests | **Done** |
| 6 Heroes | Hire, XP/levels, skills, items/equipment, wounds | 5 | Hero test incl. mission XP | **Done** (no spells/mana, no capture/ransom) |
| 7 Social economy | Market (escrow, fees), NPC exchange, alliances, roles, diplomacy, reinforcements, messages, report/block, notifications | 5 | Concurrent market & quest tests; alliance rules tests | **Done** (no joint attack, no auction, no chat channels beyond alliance mail) |
| 8 Game Pack engine | Versioned format, validator, theming, second pack, world pinning, migration tool | 2 | Two packs validate; migration test | **Done** |
| 9 Platforms | Yandex HTML5 build & adapter, Android (Capacitor) packaging, mobile UI | 8 | Builds produced; adapter wired; requirements doc | **Partial**: Yandex ZIP builds; Capacitor configured; Android project not generated (no Android SDK); platform sign-in & Yandex moderation not done |
| 10 Hardening | Load tests, security review, restore drill, responsive QA, deployment docs | all | Load test report; restore verified; checklists | **Partial**: restore verified, race-detector + concurrency tests; no load test yet |

Next increments are listed in `docs/ROADMAP.md`.
