# Project context (short)

- **What**: Realmforge, an engine for persistent multiplayer economic-military strategy games, plus the original game *Ashen Crowns* and the demo pack *Void Frontier*.
- **Origin**: game systems reconstructed from public descriptions of *My Lands: Black Gem Hunting* (research in `docs/research/`, evidence via search excerpts only). No original assets, names or texts are used.
- **Stack**: Go 1.27 + PostgreSQL 16 (single binary: API + event worker); React 19 / TypeScript 7 / Vite 8 client; Capacitor 8 for Android; HTML5 ZIP for Yandex Games.
- **Core ideas**:
  - Server-authoritative timestamps + scheduled events (no per-entity timers).
  - Exactly-once processing under player row locks.
  - Idempotent commands.
  - Declarative, versioned Game Packs pinned per world.
- **Where things are**:
  - `server/internal/rules`: formulas.
  - `server/internal/game`: use cases and transactions.
  - `server/internal/api`: HTTP.
  - `packs/`: games.
  - `client/src`: UI.
  - `docs/spec`: the normative documents.
- **Status**: see `docs/IMPLEMENTATION_STATUS.md`. Limits: `docs/KNOWN_LIMITATIONS.md`. Next: `docs/ROADMAP.md`. Rules for contributors: `AGENT.md`.
