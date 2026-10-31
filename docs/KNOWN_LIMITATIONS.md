# Known limitations (honest list)

## Research
- **No original page was opened directly.** The network policy of the authoring environment blocked the official encyclopedia, site, Steam and related hosts; all evidence came from search-engine excerpts. Confidence is capped accordingly. Several numbers (unit stats, building cost tables, travel formula, map size) are UNKNOWN and our values are PROPOSED.
- Platform requirements rows marked `[S]` come from search excerpts of Yandex/Google pages and must be re-checked before submission.

## Game rules vs. original (documented divergences)
- Buildings are per-type levels, not placed instances on city cells (D-012). No villages, deposits depletion, terraforming.
- Not implemented: healers/resurrection, fortification pre-round, hero spells/mana, hero capture/ransom, slave capture, building destruction, domain seizure of cities, joint attack, premium auction, clan castle/warehouse, mentor system, tournaments/events, renegade status, rank system beyond score.
- Combat is deterministic (no damage ranges). Training batches complete at once (no per-unit trickle). Upkeep of units abroad is charged at departure strength until return (D-010).
- No fog of war: every map tile is visible to every player in the world.

## Engineering
- Docker image and compose file were written but **not executed** (no Docker daemon in the authoring environment); local runs used native PostgreSQL 16 (`scripts/dev-postgres.sh`).
- Android project not generated and no APK/AAB built (no Android SDK in the environment). Capacitor config and scripts exist.
- Yandex/Google platform identity verification not implemented (`/auth/platform` returns 501 except the `dev` platform, which is disabled in production).
- Rate limiter is in-memory per instance (not shared across instances).
- No load/performance test yet. No WebSocket/SSE push (polling).
- No PITR/WAL archiving configured; logical backups + verified restore script only.
- Client UI strings exist in English and Russian only; Void Frontier is a demo, not a complete game.
- No monetisation, ads or payments (intentional).
