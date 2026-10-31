# Game Pack specification (format version 1)

A Game Pack is a directory `packs/<pack-id>/` containing `pack.json` (and optionally `assets/`). It defines one game product on the Realmforge engine. Packs are **data only**: no scripts, no expressions, no code. Every behaviour is selected from closed enumerations implemented by the engine.

Reference implementation: `server/internal/pack/types.go` (types), `validate.go` (rules), `load.go` (strict parsing, hashing, indexes). CLI: `go run ./cmd/packtool validate|summary <dir>`.

Shipped packs:

| Pack | Status | Purpose |
|---|---|---|
| `ashen-crowns` 1.0.0 | Playable, original content, EN/RU | First product |
| `void-frontier` 0.1.0 | Demo, validates, playable world (×10) | Proves a second game runs on the same core with no code changes (different resources, races, upkeep resource, growth mode, storage model, map size, screens). Not a finished game. |

## 1. Top-level fields

| Field | Type | Required | Notes |
|---|---|---|---|
| `formatVersion` | int | ✔ | 1 ≤ v ≤ engine `FormatVersion` |
| `id` | slug `[a-z0-9-]{2,48}` | ✔ | stable product id; worlds can only migrate within the same id |
| `version` | semver | ✔ | bump on any change; worlds pin a version |
| `meta` | `{title, subtitle?, description, studio?, license?}` | ✔ | localized texts |
| `locales`, `defaultLocale` | string[] / string | ✔ | every `name` must exist for every locale |
| `theme` | `{palette, fonts?, icons?, art?, assetsLicense}` | ✔ | palette needs `bg surface text muted accent danger ok` (hex); `icons` maps entity id → `glyph:<name>` or relative asset path (small icons: resources, items); `art` maps entity id → relative asset path of a large illustration (building sprite, unit/hero portrait, terrain texture); paths match `[a-z0-9_-/]+.(png|webp|svg|jpg)`; `assetsLicense` mandatory provenance statement |
| `params` | object | – | balance/rule parameters; engine defaults apply (see §3) |
| `resources`, `races`, `buildings`, `techs`, `units` | arrays | ✔ (resources, races) | see §2 |
| `heroes` | `{classes, skills, items, slots, maxLevel, xpBase, xpGrowth, skillPointsPerLevel, xpPerScore}` | – | requires `params.heroBuilding` if classes exist |
| `terrains`, `neutrals` | arrays | ✔ terrains | at least one passable terrain |
| `quests` | array | – | closed goal types |
| `start` | `{resources, buildings, units?, population}` | ✔ | new player's capital |
| `modes` | `standard` / `peaceful` / `speed` | ✔ | declared modes (only standard rules + world speed are implemented) |
| `ui` | `{screens[], cityLayout?}` | ✔ | which client screens exist, in order |
| `migrations` | `[{fromVersion, renames{old:new}, removed{old:replacement or ""}}]` | – | how worlds on older versions move to this one |
| `strings` | `{key: LocText}` | – | free texts (tutorial, lore) |

IDs (`[a-z][a-z0-9_]{0,47}`) are **unique across all entity kinds** so renames and references are unambiguous.

## 2. Entities

* **Resource**: `kind` basic|food|rare|premium, `capped`, `tradable`, `lootable`, `baseStorage`. Premium resources may be neither tradable nor lootable (validator).
* **Race**: modifiers, colour, optional `faction` (same-faction/same-race attack penalties).
* **Building**: `category` core|economy|storage|military|science|special, `maxLevel` 1–100, `cost {base, growth 1–5}`, `time {baseSeconds, growth 1–5}`, `requires[]` (building/tech/race + level), `effects[]`, `races[]`, `score`.
* **Effect types** (closed): `produce` (resource, perLevel/h, workers per level, growth), `storage` (resource or `*`), `protect` (hidden from loot), `housing`, `idle_tax` (resource per idle inhabitant per level per hour), `modifier` (Modifier × level). Value at level L = `perLevel × L × growth^(L−1)`.
* **Modifier**: `{target, op: pct|flat, value}`. Targets (closed grammar): `production.<res|*>`, `storage.<res|*>`, `protect.<res|*>`, `attack|defense|hp.<unitId|class|*>`, `housing`, `pop_growth`, `build_time`, `research_time`, `train_time`, `upkeep`, `speed`, `carry`, `loot`, `scout`, `wall`, `hero_xp`, `market_fee`.
* **Tech**: like buildings plus `branch` and optional `exclusiveGroup` (specialisations: one tech per group).
* **Unit**: `class` infantry|ranged|cavalry|siege|scout|transport|settler|magic|beast, attack, defense (damage reduction %), hp, speed (tiles/h), carry, upkeep (per hour, in `params.upkeepResource`), scout, cost, trainSeconds, `trainedIn` (omit for neutral-only guards), requires, class `bonuses[]`, races, score.
* **Heroes**: classes (innate + per-level modifiers), skills (per-level modifiers, class restriction, maxLevel ≤ 20), items (slot, rarity, modifiers).
* **Terrain**: weight (map generation), passable, colour, resource (territory yield).
* **Neutral**: kind lair|ruin|node, `levels[]` (consecutive from 1: guards, rewards, xp, item drops with %), respawnSeconds, capturable.
* **Quest**: goal type (closed): `building_level, tech_level, units_trained, missions_won, neutrals_cleared, join_alliance, market_trade, hero_hired, cities, territories, scouts_sent`; `requires[]` quests; rewards; tutorial flag; target `screen`; hint.

## 3. Parameters

All parameters, defaults and allowed ranges are defined in `pack.Defaults()` and `validate.go → params()`. The complete table with Ashen Crowns values and gameplay impact is in `BALANCE_MODEL.md`. Key switches that change rules (not just numbers): `upkeepResource`, `popGrowthMode` (idle | free_housing), `overflowDecayPctPerHour` (0 = hard storage cap), `sameRaceAttackPct`/`sameFactionAttackPct`, `maxAttacksPerTargetPerDay`, `attackScoreFloorPct`, `missionUpkeepMult`, `defenseCapPct`, `defenderRetreatPct`, `cityCapBuilding`, `territoryCapBuilding`, `settlerUnit`, `heroBuilding`, `exchangeRates`.

## 4. Validation (enforced at load; a world cannot be created from an invalid pack)

Schema & strictness (unknown fields rejected) · ID format & global uniqueness · localization coverage · numeric ranges · referential integrity (costs, requirements, effects, modifiers, units in neutrals, item drops, quest goals, start conditions, params references) · requirement level within target's max · dependency cycles (buildings/techs, quests) · palette colours · icon paths relative (no `..`, absolute paths or URLs) · tick alignment (`economyTickSeconds` divides 86400) · migration targets exist. Tested in `pack_test.go` (22 negative cases + shipped packs).

## 5. Theming & assets

The client applies `theme.palette` as CSS variables and loads `theme.fonts` from Google Fonts; icons use built-in original SVG glyphs (`glyph:*`) or files from `packs/<id>/assets/` served at `/packs/<id>/assets/…` (the pack directory name must equal the pack id). `theme.art` drives the isometric city scene (building sprites placed on plots; positions from `ui.cityLayout` or an automatic ring layout), unit and hero portraits and map terrain textures. Import new art with `scripts/import_art.py <src assets> packs/<id>` (trims, resizes, converts to WebP). The shipped-packs test fails if any referenced file is missing. Replacing art = replacing files + icon map, no code change. Every pack must state asset provenance in `assetsLicense`.

## 6. Versioning & world migration

1. Worlds store the full pack snapshot + hash; disk changes never alter running worlds.
2. To upgrade: publish pack vN+1 with a `migrations` entry `fromVersion: vN` listing renames/removals.
3. Admin runs `POST /api/v1/admin/worlds/{wid}/migrate-pack {packDir, dryRun:true}` → report of IDs in use and anything missing.
4. If `missing` is empty, run with `dryRun:false`: IDs are rewritten in cities, players, missions, jobs, tiles, heroes, items in one transaction; history + audit recorded.
5. Constraints: same pack `id`; races/terrains/hero classes cannot be dropped without replacement; numbers (costs, stats) take effect from the migration instant (already-paid jobs keep their times).

## 7. Creating a new game (checklist)

1. Copy a pack directory, change `id`, `meta`, `theme`, entities and params.
2. `go run ./cmd/packtool summary ../packs/<id>` until it validates.
3. Start a world: `RF_WORLDS="slug:<id>:Name:speed"`.
4. Run `make test` (shipped packs are validated by tests) and play through the tutorial quests.
5. Provide translations for all declared locales; document asset licenses.
