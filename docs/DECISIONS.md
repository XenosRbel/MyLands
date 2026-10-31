# Architecture & design decisions (ADR log)

Format: ID — decision — context/alternatives — consequences. Newest last.

**D-001 — Modular monolith in Go + PostgreSQL as the single source of truth.** Alternatives: microservices, actor framework, Redis-first state. The game's consistency needs (spend once, apply events once) map directly onto relational transactions and row locks. One binary runs API + worker. Consequence: horizontal scaling via stateless instances; no distributed coordination layer.

**D-002 — Latest stable toolchain (user request 2026-10-31).** Go 1.27.3 (via `toolchain` directive), pgx v5.11.0 (v5.12.0 was <1 day old; skipped), Node 24/22, React 19.3.1, Vite 8.3.4, TypeScript 7.0.2, Vitest 5.0.5, Capacitor 8.5.3 (8.6.0 was 3 days old). Rule: pin exact versions, prefer releases ≥ ~2 weeks old.

**D-003 — Time model = stored state + timestamps + scheduled events, lazy settlement + background worker.** No per-entity timers. Events processed in (due_at, id) order inside the owning player's lock; economy integrated piecewise between events with absolute-aligned ticks. Worker only improves latency; correctness holds without it. See ARCHITECTURE §4.

**D-004 — `game` as one package with per-subsystem files** sharing a locked player context, rather than one Go package per subsystem. Subsystems all start from the same lock+settle step; separate packages would need an interface layer with no isolation benefit at this size. Revisit if a subsystem becomes a separately deployed service.

**D-005 — City aggregate stored as JSONB maps (resources, buildings, units).** Pack-driven IDs; whole-row read/write under lock; pack validator + migration tool provide integrity instead of FKs.

**D-006 — Integer milli-units for stocks.** Avoids float drift; ≤0.001 unit loss per settlement (bounded, tested).

**D-007 — No Redis in v0.1.** PostgreSQL provides the job queue (`game_events` + `FOR UPDATE SKIP LOCKED`), locking and idempotency. Rate limiting is in-memory per instance. Redis becomes justified only for multi-instance rate limiting or pub/sub push — then added behind an interface.

**D-008 — REST + polling, no WebSocket in v0.1.** Authoritative state is always fetched via REST; 15 s polling + timer-triggered refresh is enough for a strategy game with minute-scale timers. Push (SSE) is a planned optimisation for incoming-attack alerts.

**D-009 — Lock order players → cities → missions → heroes → tiles → orders/items; never two player locks in one transaction.** Cross-player effects are routed as events of the affected player (market proceeds = `delivery` event; returning loot belongs to the attacker's `mission_return`). Mission rows may be written by the defender's transaction.

**D-010 — Upkeep of units away is charged at departure strength until return (simplification).** Keeps D-009 (losses abroad would otherwise change another player's economy mid-interval). Research: missions cost double upkeep (ECO-015) → `missionUpkeepMult = 2` in Ashen Crowns.

**D-011 — Deterministic combat with simultaneous rounds**, count-proportional damage spread, defense as % reduction with cap, wounds carried within a battle, retreat thresholds, attacker wins only if defenders are destroyed or retreat. Derived from research CMB-001..006 (round cap 20, defense reduces attack by %, spread by group size, retreat thresholds, both-retreat → defender wins). The original's min–max damage ranges and averaged simulations (CMB-010) are replaced by determinism (testable, replay-safe). Randomness only where the design needs probability (scouting detection, item drops) via RNG seeded by mission id.

**D-012 — Buildings are level-based per type, not placed instances on city cells.** Research shows the original uses cell-based placement (capital 51 cells, BLD-003) with multiple instances; that requires a different data model and UI. v0.1 uses per-type levels (one row per type) — documented divergence; cell layout is a candidate for format v2.

**D-013 — Territories give yield accrued on the tile (collect action), not production modifiers.** Keeps D-009 when territories change owner in another player's transaction, and creates a "capture the uncollected yield" mechanic. PROPOSED design (the original's outer domains are full mini-cities, ECO-011).

**D-014 — Pack snapshot pinned per world + explicit audited migration with dry run.** Changing pack files never mutates live worlds.

**D-015 — World speed multiplier stored on the world** (multiplies pack `worldSpeed`): speed/test worlds without forking packs. Flows (production, consumption, upkeep), durations, travel (incl. minimum travel), respawns and territory yield scale; stocks and capacities do not.

**D-016 — Idempotency key inserted inside the command transaction with a savepoint**, exact response bytes stored (BYTEA) for byte-identical replay; domain errors stored, infrastructure errors not.

**D-017 — Ashen Crowns uses `free_housing` population growth** although research verified idle-based reproduction (POP-002), because idle-only growth deadlocks a city whose population is fully employed (growth 0) when cottage/keep base growth values are unknown. The engine implements both (`popGrowthMode`); Void Frontier demonstrates `idle`.

**D-018 — Pre-release migration edits.** `0001_init.sql` was edited in place during v0.x development (econ_carry, world speed, BYTEA response). From the first deployed release on, migrations are append-only.

**D-019 — Map rendering with Canvas 2D, not PixiJS.** A 15×15 tile viewport with vector glyphs does not need a WebGL engine; Canvas keeps bundle size at ~93 KB gzip. Revisit when art assets/animation arrive.

**D-020 — Platform identity sign-in is pluggable; only `dev` exists.** Yandex/Google identity must be verified server-side (signatures/tokens) — not implemented; endpoint returns 501 for unconfigured platforms. Guest play covers Yandex's "no forced login" rule meanwhile.

**D-021 — No real-money features.** Premium currency (`starglass`) is earned in-game only, non-tradable, non-lootable. Payments/ads require separate product, legal and platform review (brief §14).
