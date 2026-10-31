# Realmforge

A server-authoritative engine for persistent, multiplayer economic-military strategy games, plus its first game, **Ashen Crowns**, an original fantasy strategy (EN/RU), and a second demo pack, **Void Frontier**, which shows that new games need no engine code changes.

> Game systems were studied from public descriptions of the genre classic *My Lands: Black Gem Hunting*. No names, texts, art or data of that game are used. See `docs/research/` for evidence and `docs/KNOWN_LIMITATIONS.md` for what is not implemented.

| Part | Tech |
|---|---|
| Server | Go 1.27 · net/http · pgx v5 · PostgreSQL 16 · slog |
| Client | React 19 · TypeScript 7 · Vite 8 · Canvas 2D map |
| Packaging | Yandex Games HTML5 ZIP · Capacitor 8 (Android) · Docker |

## Quick start (native, no Docker)

Requirements: Go ≥ 1.24 (the 1.27.3 toolchain downloads automatically), Node ≥ 22, PostgreSQL 16 binaries.

```bash
make db
```

```bash
make build
```

```bash
make server
```

Open <http://127.0.0.1:8080>, press **Play as guest**, and join **Ember Vale (x50)**, a ×50 speed world. A **Void Frontier Demo** world is also created.

For client development with hot reload, run `make client` in a second shell, then open <http://localhost:5173>. It proxies `/api` to `:8080`.

## Quick start (Docker)

```bash
docker compose up --build
```

(The Dockerfile and compose file are provided but were not executed in the authoring environment; see `docs/KNOWN_LIMITATIONS.md`.)

## Tests

```bash
make test
```

This runs Go unit tests, integration tests against real PostgreSQL (`RF_REQUIRE_DB=1`, race detector on) and the client unit tests.

```bash
make e2e
```

This drives a real browser through a full session against a running server.

```bash
make packs
```

This validates all Game Packs.

## Configuration

All configuration is through environment variables; see `.env.example`. Secrets are never stored in the repository. Production (`RF_ENV=production`) refuses `RF_DEV_PLATFORM_AUTH=true` and non-TLS database URLs unless explicitly allowed.

`RF_WORLDS="slug:packDir:Display Name:speed,..."` creates missing worlds at startup. A world pins a snapshot of its pack.

## Operations

| Task | Command |
|---|---|
| Backup | `make backup` (pg_dump, custom format) |
| Restore into an empty DB | `make restore FILE=backups/x.dump` |
| Backup drill | `make verify-backup` (restore into a scratch DB and compare row counts) |
| Health | `GET /healthz`, `GET /readyz` |
| Metrics | `GET /metrics` (Prometheus text) |
| Admin | promote a login with `RF_ADMIN_LOGIN`; endpoints under `/api/v1/admin` (audited) |
| Move a world to a new pack version | `POST /api/v1/admin/worlds/{id}/migrate-pack {"packDir":"…","dryRun":true}` |
| Yandex build | `cd client && VITE_API_BASE=https://api.example npm run build:yandex && npm run yandex:zip` |
| Android | `cd client && npm run android:add && npm run android:sync` (needs Android SDK) |

## Repository map

```
server/            Go module: cmd/server, cmd/packtool, internal/{api,game,rules,pack,platform}, migrations, api/openapi.yaml
client/            React client, platform adapters, e2e/play.mjs, capacitor.config.ts
packs/             ashen-crowns (game 1), void-frontier (demo)
docs/research/     public-source research, mechanics catalogue, unknowns
docs/spec/         PRD, GDD, rules, balance, architecture, data model, API, security, UI/UX, pack spec, platforms, tests, roadmap, risks
docs/              DECISIONS, IMPLEMENTATION_STATUS, KNOWN_LIMITATIONS, ROADMAP
tasks/             atomic task cards with acceptance criteria
scripts/           dev-postgres, backup/restore/verify
```

Start with [`PRODUCT.md`](PRODUCT.md) for the product, [`AGENT.md`](AGENT.md) for contributor rules and [`docs/spec/ARCHITECTURE.md`](docs/spec/ARCHITECTURE.md) for internals.
