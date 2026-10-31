# AGENT.md — permanent rules for anyone (human or AI agent) changing this repository

`AGENTS.md` is a symlink to this file.

Shared Codex and Claude Code workflow: read `.agents/HARNESS.md`.
Setup, hook trust and limitations: `docs/AGENT_HARNESS.md`.

## Before you start
1. Read `PROJECT_CONTEXT.md`, `docs/IMPLEMENTATION_STATUS.md`, `docs/KNOWN_LIMITATIONS.md`, and the relevant spec in `docs/spec/`.
2. Pick or create a task card in `tasks/` (ID, goal, dependencies, modules, acceptance criteria, required tests).
3. Reconcile docs with code before starting a new stage; fix drift first.

## Non-negotiable engineering rules
- **Server is authoritative.** Never trust client values for resources, timers, combat, ownership.
- **Game rules live in `server/internal/rules` (pure, deterministic) and are configured by Game Packs.** Never hard-code names or numbers of races, buildings, units, resources, techs or quests in engine or client code.
- **Every state change runs in one transaction** through `WithPlayer`/`WithCity` (lock + settle first). Respect the lock order `players → cities → missions → heroes → tiles → orders/items`, and **never lock two players in one transaction**. Route cross-player effects as events of the affected player.
- **Every spend/grant writes `resource_ledger`.** Stocks must never become negative (`rules.Spend` panics on overspend).
- **Scheduled effects go through `game_events`**; handlers must be idempotent with respect to the referenced row's status.
- **New game commands** must be registered through `cityCmd`/`playerCmd` (idempotency key required), added to `server/api/gen_openapi.py`, and regenerate `api/openapi.yaml` (coverage test enforces it).
- **Ownership checks**: foreign IDs return 404. Admin actions must be audited.
- **Validation**: strict JSON (unknown fields rejected), bounded numbers, parameterised SQL only.
- **Migrations are append-only** once released (`server/migrations/000N_*.sql`).
- **Packs**: run `make packs`; a pack change that renames/removes IDs must add a `migrations` entry.
- **Research honesty**: never raise a mechanic's evidence status without a new source; mark own designs as PROPOSED.
- **No third-party game assets.** All art/text must be original or properly licensed and recorded in `theme.assetsLicense`.
- **No real-money features** without a product/legal/platform review.

## Definition of done for a change
- `make lint` and `make test` pass (Go tests with `RF_REQUIRE_DB=1` and `-race`; client typecheck and vitest).
- For user-facing flows, the browser E2E (`make e2e`) passes against a running server.
- New behaviour has tests: unit (rules), integration (DB state), concurrency where resources/ownership are involved.
- Docs updated: GAME_RULES (rule IDs referenced in code comments), IMPLEMENTATION_STATUS, KNOWN_LIMITATIONS, DECISIONS (if a design choice was made), task card result.
- Never disable or weaken tests to get green; never present partial work as complete.

## Commands
```bash
make db          # native PostgreSQL for dev/tests
make build       # client + server binaries
make server      # run API + worker with demo worlds (x50 speed)
make test        # all automated tests
make e2e         # browser E2E against :8080
make packs       # validate Game Packs
```

## Style
Go: `gofmt`, `go vet`, small functions, comments state the rule ID (`GAME_RULES R-…`). TypeScript: strict mode, no `any` in new APIs where avoidable, screens get data via `useGame()` and commands via `command()` (idempotency key per action).
