# Mechanics Catalog: reference game vs Realmforge engine

Owner: Agent B (Game Designer). Date: **2026-10-31**. Code checked at commit `dfa660e` and re-checked at `ed6acc7` (no rule changes in between) (`server/`), packs `packs/ashen-crowns/pack.json` v1.0.0 and `packs/void-frontier/pack.json`.

This is the single table of **every mechanic ID** found in the research notes (`RACES_AND_FACTIONS.md`, `ECONOMY_AND_RESOURCES.md`, `BUILDINGS_AND_TECHNOLOGIES.md`, `ARMY_AND_COMBAT.md`, `HEROES_AND_PROGRESSION.md`, `SOCIAL_AND_MULTIPLAYER.md`, `WORLD_AND_EXPLORATION.md`), with what our engine does about it.

## How to read this table

- **Status / Conf.** are copied from the research note. They are never raised here. All evidence came from search-engine excerpts; no page was opened directly (see `SOURCE_REGISTER.md`). "VERIFIED" therefore means "verified against excerpts", not "read on the live page" or "observed in play".
- **Our impl.**
  - **Implemented**: the engine has the rule with the same shape. Numbers may still be our own (PROPOSED) values.
  - **Impl-diff** (implemented differently): the engine covers the same design purpose with a different rule.
  - **Not impl.**: nothing in the engine does this today.
- **Engine location** uses paths under `server/internal/` (for example `rules/combat.go Resolve`) and pack parameter names (`params.x`) or pack entity IDs. "AC" means the Ashen Crowns pack; "VF" means the Void Frontier demo pack.
- Code comments cite normative rule IDs (e.g. R-CMB-3) defined in `docs/spec/GAME_RULES.md`. This catalog cites code locations directly; see `GAME_RULES.md` for the normative formulas.

## Summary counts

| Area | IDs | Implemented | Impl-diff | Not impl. |
|---|---|---|---|---|
| Races & factions (RACE, FAC) | 13 | 3 | 3 | 7 |
| Economy (ECO) | 24 | 5 | 14 | 5 |
| Buildings (BLD) | 24 | 8 | 8 | 8 |
| Sciences & specialisation (TECH) | 21 | 1 | 8 | 12 |
| Population (POP) | 9 | 0 | 4 | 5 |
| Quests (QST) | 6 | 2 | 3 | 1 |
| Army (ARMY) | 14 | 2 | 6 | 6 |
| Missions (MSN) | 12 | 2 | 7 | 3 |
| Combat (CMB) | 12 | 4 | 5 | 3 |
| Heroes & progression (HERO) | 12 | 2 | 6 | 4 |
| Social & multiplayer (SOC) | 24 | 3 | 9 | 12 |
| World (WLD) | 15 | 0 | 7 | 8 |
| **Total** | **186** | **32** | **80** | **74** |

Explicitly **not implemented** (frequently asked): healers and resurrection, hero capture and ransom, spells and mana, auction of premium currency, deposit depletion, cell-based city layout, villages, slave capture, building destruction, joint attack, clan castle, tournaments, gates, exploration/fog of war, renegades and faction loyalty, paid boosts.

## Catalog

### Races and factions

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| RACE-001 | One race per account/server, fixes roster, economy, faction | VERIFIED | high | Implemented | `game/world.go Join` (race chosen at join); `pack.Race`; one player per account per world (unique insert) | AC has 4 races, VF 2. No race-change endpoint exists (race is immutable). |
| RACE-002 | Elves: best wood; healer bonus | VERIFIED | high (wood) / low (healer) | Impl-diff | AC race `thornkin`: `production.timber` +15%, `attack.ranged` +10%, `speed` +5% | Original race identities are not reused; AC races are new. No healers. |
| RACE-003 | Dark Elves: wood + iron; faster/better mages | VERIFIED | med | Impl-diff | AC races carry `modifiers` + race-only units (`Unit.Races`) | Race asymmetry exists, but no AC race copies this profile. |
| RACE-004 | Knights: gold, carrier capacity, warehouses (crops disputed) | VERIFIED | high / med / disputed | Impl-diff | Modifier grammar supports `production.gold`, `carry`, `storage.*` (`pack/validate.go ModifierTargetValid`) | No AC race uses these exact bonuses. `ironvow` = defense +8%, cavalry attack +10%, stone +10%. |
| RACE-005 | Demons: stone; strongest mercenary | VERIFIED | med | Not impl. | — | No mercenary category; `ashborn` = attack +8%, iron +15%, loot +10%. |
| RACE-006 | Wizards: crops + iron; half mana cost; spell resistance | VERIFIED | med | Not impl. | — | No mana/spells. |
| RACE-007 | Undead: wood + stone | VERIFIED | low | Not impl. | — | `gravewarden` = upkeep −20%, HP +8%, lore +15% (own design). |
| RACE-008 | Chimeras: underground specialist | VERIFIED | low | Not impl. | — | No underground domains. |
| RACE-009 | Native-land +25 Defence, max Defence 50, not for Renegades | VERIFIED | med | Not impl. | — | Terrain has no combat effect. Our defense cap is `params.defenseCapPct` = 80 (AC), a different concept (max damage reduction %). |
| FAC-001 | Two factions (Light / Dark) derived from race | VERIFIED | high | Implemented | `pack.Race.Faction`; AC `dawnpact` (thornkin, ironvow), `duskbound` (ashborn, gravewarden) | Faction IDs have no localized display names in the pack yet. |
| FAC-002 | Attack −50% vs own race, −25% vs own faction | VERIFIED | high (values) / stacking INFERRED | Implemented | `game/missions.go factionMods`; `params.sameRaceAttackPct` = −50, `params.sameFactionAttackPct` = −25 (AC; 0 in VF) | Applied as an `attack.*` pct modifier in `battleCity` and `battleTerritory` (not vs neutrals). Same race gets −50 only (not −75). It is **added** to other attack pct modifiers, not multiplied (see `rules/modifiers.go Apply`). No dedicated test. |
| FAC-003 | Renegade status 7 days after attacking own race | VERIFIED | med | Not impl. | — | |
| FAC-004 | Faction Loyalty rating (gold value of enemy-faction kills) | VERIFIED | med | Not impl. | — | Only a single power `score` ranking exists (`game/economy.go Rankings`). |

### Economy and resources

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| ECO-001 | Wood from Main Building and Woodcutter | VERIFIED | high | Impl-diff | AC `timber` from `lumberyard` (`produce` effect) | AC Keep does not produce timber. |
| ECO-002 | Crops from Windmill only; price band 8–20 | VERIFIED | high / med | Impl-diff | AC `grain` from `farmstead`; `params.foodResource` = grain | Grain is **eaten** by population (`params.foodPerPopPerHour` 0.12) — PROPOSED; research found no evidence that population eats. No price band; market uses a 10x band around `params.exchangeRates`. |
| ECO-003 | Stone from quarry / mines / pits | VERIFIED | high | Impl-diff | AC `stone` from `quarry`; also territory yield on `hills` | No mines/pits. |
| ECO-004 | Iron from plant / mines / pits | VERIFIED | high | Impl-diff | AC `iron` from `ironworks` (needs keep 2); territory yield on `ironridge` | |
| ECO-005 | Gold from Trading Post (idle tax), plant, mines | VERIFIED | high / TP formula INFERRED low | Implemented | `rules/economy.go ComputeRates` (`idle_tax`); AC `tradepost` 0.06 gold/h per idle per level; `mint` 14/h per level (keep 4, ironworks 2) | The INFERRED idle × tax formula is what we built (PROPOSED value). Covered by `rules TestIdleTaxAndUpkeep`. |
| ECO-006 | Premium Black Gems: from lakes, grail, ruins, auction, real money; not stealable | VERIFIED | high / med (grail rate) | Impl-diff | AC `starglass` (`kind: premium`, not tradable/lootable); `params.premiumResource`; validator forbids tradable/lootable premium (`pack/validate.go`); transport/settle cargo refuse premium (`game/missions.go SendMission`) | Earned only in-game (start 20, quests, `barrow`/`crater` neutrals). **No real-money sales and no spending sink implemented yet.** |
| ECO-007 | Science Points from labs, pooled across domains, surplus lost | VERIFIED | high | Impl-diff | AC `lore` from `scriptorium`; tech cost `lore` + resources paid by the city that queues research (`game/city.go Research`) | Lore is a **per-city** stock (uncapped), not a pooled flow; no surplus loss. Tech levels are player-wide. |
| ECO-008 | Hero mana | VERIFIED (exists) | med | Not impl. | — | |
| ECO-009 | Finite deposits, drop to 1/h, recovery building | VERIFIED | high | Not impl. | — | Production is unlimited. |
| ECO-010 | Production modifier stack (terrain −50..+200%, village +30%, pit +100%, paid +25%) | VERIFIED | high (values) / INFERRED (stacking) | Impl-diff | `rules/modifiers.go Apply`: value = (base + Σflat) × (1 + Σpct/100) | Stacking is **additive pct** (our decision). Terrain/village/pit/paid modifiers do not exist. `params.territoryBonusPct` is declared but not read by game code. |
| ECO-011 | Outer domains (mines, pits, lakes, grail) as contested nodes | VERIFIED | high | Impl-diff | Territories: `game/missions.go occupy`, `battleTerritory` (capture); `game/economy.go territoryYield`, `Collect` | One tile = one territory yielding its terrain resource (`params.territoryYieldPerHour` 25, cap `territoryYieldCapHours` 24 h). Cap = keep level × `territoriesPerLevel` (1). No lifespan. |
| ECO-012 | Per-domain storage limit from warehouses + main building; production halts at cap | VERIFIED | high (rule) | Impl-diff | `rules/economy.go ComputeRates` (`storage` effects, `Resource.BaseStorage`), `Accrue` | Per-city caps exist. In AC (`overflowDecayPctPerHour` > 0) production does **not** halt at the cap; surplus decays instead. No allocation chart. VF uses hard cap. |
| ECO-013 | Over-storage surplus lost 1%/h; hard cap 999,999,999 | VERIFIED | high / scope med | Implemented | `rules/economy.go TickStep` step 3; `params.overflowDecayPctPerHour` = 1 (AC) | Decay at each economy tick (10 min, ×world speed). 999,999,999 cap not implemented. Test `rules TestOverflowDecay`. |
| ECO-014 | Cache: theft-protected amount | VERIFIED | med | Impl-diff | AC `vault` (Hidden Vault) `protect` effect; `rules/combat.go Loot` subtracts protected | A building, not a science. |
| ECO-015 | Gold upkeep; 0.5%/h desertion when unpaid; double on missions; Exchequer up to 96% | VERIFIED | med | Implemented | `params.upkeepResource` = gold, `desertionPctPerHour` = 0.5, `missionUpkeepMult` = 2 (AC); `rules/economy.go TickStep` step 2; `game/settle.go refreshAway` | Desertion only when gold stock is 0 **and** net gold < 0. Mission ×2 applies to units away **and** stationed. AC `exchequer` tech −4%/level (max −40%). Test `game TestStarvationAndDesertionOverLongAbsence`. |
| ECO-016 | Costs paid from selected domain; units cost resources and population | VERIFIED | med / numbers UNKNOWN | Impl-diff | `game/settle.go spend`; `rules/economy.go CostAt` (base × growth^(L−1)) | Paid from the acting city. Units cost resources only (no population). |
| ECO-017 | Player market: resources ↔ gold; merchants needed to buy; tax | VERIFIED | high | Impl-diff | `game/economy.go PlaceOrder`, `AcceptOrder`, `CancelOrder`; `params.marketFeePct` 5, `marketMaxOpenOrders` 10 | Order book for **any pair** of tradable resources, escrowed. No merchants or Market building needed. 10x price band vs `exchangeRates`. |
| ECO-018 | Exchange tab: resource swap at worse rates | VERIFIED | low | Implemented | `game/economy.go Exchange`; `params.exchangeRates`, `exchangeFeePct` 30 (AC) | NPC sink: received = floor(amount × rate_from / rate_to × 0.7). |
| ECO-019 | Black Gem auction for gold (rating ≥ 10M, +5,000 step) | VERIFIED | med | Not impl. | — | Out of MVP scope. |
| ECO-020 | Transport units (merchants/carriers), capacity | VERIFIED | high | Impl-diff | AC `wagon` (class `transport`, carry 300); `game/missions.go carryOf`, `transport` mission | Every unit has `carry`; wagons are just the biggest. Not needed for market. |
| ECO-021 | Ancient Magic: paid boosts and instant completion | VERIFIED | med / prices low | Not impl. | — | Deliberately absent (no pay-to-win). |
| ECO-022 | Market-average gold valuation for ratings | VERIFIED | med | Not impl. | — | Score = Σ building score × level + tech score × level + home unit score (`game/settle.go score`). |
| ECO-023 | Loot/items from monster ruins | VERIFIED | med | Impl-diff | `game/missions.go battleNeutral` (rewards + seeded item drops) | Rewards are fixed per neutral level; items drop by `chance`. |
| ECO-024 | Domain-local economy; only SP global | VERIFIED | high | Implemented | Per-city `resources` (`game/settle.go City.Res`); techs player-wide | Moving goods needs a `transport` mission. |

### Buildings

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| BLD-000 | Cells, builders, time bonus cap 80%, building deletion | VERIFIED | high | Impl-diff | `game/city.go Build`; `rules/modifiers.go TimeFactor` (`MinTimeFactor` 0.2 = max 80% reduction); `params.buildQueueMax` 3 | 80% cap implemented. No cells, no builders, no demolition. Queue is sequential. |
| BLD-001 | Main Building (max level 5): wood, storage, housing, cells | VERIFIED | med | Impl-diff | AC `keep` (max 25): housing, build time −2%/level, gates buildings, cities (`game/missions.go cityCap`) and territories (`territoryCap`) | |
| BLD-002 | City Gate unlocks outward actions | VERIFIED | med / timing UNKNOWN | Not impl. | — | Missions are available from the start. |
| BLD-003 | Cell grid: Capital 51, new city 22→43, City Expansion | VERIFIED | high | Not impl. | — | Cell-based layout not implemented: one instance per building type, levels only (`City.Buildings` map). |
| BLD-004 | 4 villages per city, +30% production | VERIFIED | high | Not impl. | — | |
| BLD-005 | Deposit-cell restriction until land-use science | VERIFIED | low | Not impl. | — | |
| BLD-006 | Terraformer clears obstacle cells | VERIFIED | med | Not impl. | — | |
| BLD-010 | Woodcutter | VERIFIED | high | Implemented | AC `lumberyard` (32/h per level, 6 workers/level, growth 1.06) | |
| BLD-011 | Windmill | VERIFIED | high | Implemented | AC `farmstead` (36/h, 6 workers) | |
| BLD-012 | Stone Quarry | VERIFIED | high | Implemented | AC `quarry` (26/h, 6 workers) | No deposit. |
| BLD-013 | Iron Plant | VERIFIED | high | Implemented | AC `ironworks` (20/h, 8 workers) | |
| BLD-014 | Gold Plant | VERIFIED | high | Implemented | AC `mint` (14/h, 10 workers) | |
| BLD-015 | Trading Post (idle tax) | VERIFIED | med / formula INFERRED | Implemented | AC `tradepost` `idle_tax` 0.06 gold per idle per level per hour | Value PROPOSED. |
| BLD-016 | Chalice of Abundance (deposit recovery) | VERIFIED | high | Not impl. | — | |
| BLD-017 | Alchemist Lab (SP) | VERIFIED | high | Impl-diff | AC `scriptorium` (`lore` 10/h, 5 workers, growth 1.08); `academy` −3% research time/level | Research does not require a lab, only lore and tech requirements (most techs require `scriptorium`). |
| BLD-018 | Warehouse | VERIFIED | high (rule) | Implemented | AC `warehouse` storage `*` 1200 × L × 1.12^(L−1); `masonry` +3%/level | |
| BLD-019 | Cottage: living places and growth | VERIFIED | high | Impl-diff | AC `cottages` housing 90 × L × 1.06^(L−1); growth by `params.popGrowthMode` | Growth is per city, not per cottage. |
| BLD-020 | Market building | VERIFIED | high | Impl-diff | AC `market` (Market Hall, keep 3): `market_fee` −4%/level, `carry` +3%/level | Not required to trade. Its `market_fee` modifier is a building effect, but `AcceptOrder` uses player-wide mods only, so it has no effect today (see UNKNOWN_AND_ASSUMPTIONS I-03). |
| BLD-021 | Tavern (Capital only) for heroes | VERIFIED | med | Impl-diff | AC `warhall` (War Hall, keep 3, max 7); `params.heroBuilding`; `game/heroes.go HireHero` | Any city may build it; hero is hired in that city. |
| BLD-022 | Military training buildings | VERIFIED | med / mapping UNKNOWN | Implemented | AC `barracks`, `range`, `stables`, `aerie`, `workshop`, `sanctum`; `Unit.TrainedIn`; `game/city.go Train` | |
| BLD-023 | Special-unit buildings (merchants, spies, explorers, settlers, visionaries, mirrors) | VERIFIED | med | Impl-diff | AC `scout`, `wagon` (barracks), `settler` (keep 5) | Spy and counter-spy are one `scout` unit. No explorers, visionaries, mirrors. |
| BLD-024 | Defensive buildings: Fortification, Tower, Magic Tower | VERIFIED | med | Impl-diff | AC `walls` (`wall` flat +2 defense points/level to city defenders); `watchtower` (`scout` +8%/level) | No damage-dealing towers. Walls add to defense % before the cap. |
| BLD-025 | Outer-domain main buildings | VERIFIED | high | Not impl. | — | Territories have no buildings. |
| BLD-026 | Improved / perfect / monumental buildings | VERIFIED | low | Not impl. | — | |

### Sciences and specialisations

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| TECH-001 | Science system: SP, levels, confirmation + payment | VERIFIED | high | Impl-diff | `game/city.go Research`, `techStep`; `params.researchQueueMax` 2; `rules/economy.go CostAt`, `DurationAt` | Pay lore + resources up front; no confirmation step. AC has 18 techs. |
| TECH-002 | "Charged" repeatable sciences (City Expansion, Capital Transfer) | VERIFIED | high | Not impl. | — | |
| TECH-003 | Gate science | VERIFIED | high | Not impl. | — | |
| TECH-004 | Villages science | VERIFIED | high | Not impl. | — | |
| TECH-005 | Colonisation: more cities | VERIFIED | high | Impl-diff | `params.settlerUnit` (AC `settler`, keep 5); `cityCap` = min(`maxCities`, 1 + max keep level / `citiesPerCapLevels`) | Building-gated, not science-gated. AC keep max 25 ⇒ effective cap 6 cities (see I-01). |
| TECH-006 | Market science (range by level) | VERIFIED | high | Not impl. | — | Market is world-wide from the start. |
| TECH-007 | Merchantry (merchants) | VERIFIED | med | Not impl. | — | Wagon needs only barracks. |
| TECH-008 | Exchequer: salary reduction up to 96% | VERIFIED | med | Implemented | AC `exchequer` (−4% upkeep, +3% gold per level, max 10) | Max −40% in AC (PROPOSED). |
| TECH-009 | Cache science | VERIFIED | med | Impl-diff | AC `vault` building | |
| TECH-010 | Construction, Repair, Towers, Fortification, Magic Towers | VERIFIED | med (effects UNKNOWN) | Impl-diff | Build speed via `keep` (−2%/level) and `path_architect`; walls need keep 3 | No repair; no tower sciences. |
| TECH-011 | Heroes Managing / Hero Skills (cap 12) | VERIFIED | med | Impl-diff | Hero slots from `warhall` level; `heroics` (+10% hero XP/level); `heroes.maxLevel` 30 | |
| TECH-012 | Vision, Reflection, Espionage, Diplomacy | VERIFIED | med | Impl-diff | AC `espionage` (+10% scout/level); diplomacy = alliance relations (no science) | No vision/reflection. |
| TECH-013 | Mission Control (simultaneous missions) | VERIFIED | med | Not impl. | — | No limit on concurrent missions. |
| TECH-014 | Underground domains science | VERIFIED | med | Not impl. | — | |
| TECH-015 | Clan Castles science | VERIFIED | med | Not impl. | — | |
| TECH-016 | Alchemy (levels 1–6/7) | VERIFIED | low | Not impl. | — | |
| TECH-017 | Land-use science | VERIFIED | low | Not impl. | — | |
| TECH-018 | Resurrection science | INFERRED | low | Not impl. | — | |
| TECH-019 | Capital Transfer, City Expansion | VERIFIED | high | Not impl. | — | `is_capital` is fixed on the first city. |
| TECH-020 | One specialisation (3 levels); a second cancels the first; Architect −50/65/80%, Trader, Benefactor, Defender | VERIFIED | high (Architect, Trader) / med | Impl-diff | `pack.Tech.ExclusiveGroup`; `game/city.go exclusiveTaken`; AC group `calling`: `path_architect` (−15% build time/level), `path_steward` (+6% production/level), `path_warlord` (+6% attack, −8% train time/level), keep 5 | Choosing one **permanently blocks** the others (blocked as `specialisation_taken`); no switching. No dedicated test. |
| TECH-021 | Community early science order | VERIFIED (community advice) | low | Impl-diff | AC tutorial chain `t01`..`t10` | Used only as pacing reference. |

### Population

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| POP-001 | Domain population cap 150,000 | VERIFIED | med | Not impl. | — | Only the housing cap exists. |
| POP-002 | Growth: Main Building + Cottages + idle reproduction 0.04 × idle/h | VERIFIED | high | Impl-diff | `rules/economy.go TickStep`; `params.popGrowthMode`, `popGrowthPctPerHour` | AC uses `free_housing` (12%/h of free housing). The engine also has `idle` mode (growth ∝ idle people; VF uses it at 8%/h). Test `rules TestDesertionAndIdleMode`. |
| POP-003 | Idle population: flees first, raises gold | VERIFIED | med | Impl-diff | `Rates.Idle` = pop − employed; `idle_tax` | Idle feeds gold. No flight order. |
| POP-004 | Workers staff producers | VERIFIED | high / scaling INFERRED | Impl-diff | `rules/economy.go ComputeRates`: employment = min(pop, Σworkers) / Σworkers, applied to every producer | Automatic, proportional. No manual assignment. Test `rules TestRatesEmploymentAndStorage`. |
| POP-005 | Builders | VERIFIED | high | Not impl. | — | Build time does not depend on population. |
| POP-006 | Overcrowding flight formula | VERIFIED | med | Impl-diff | `TickStep`: pop clamped to housing at each tick | Excess removed at once, not by formula. |
| POP-007 | "Distribute Builders/Workers" buttons | VERIFIED | med | Not impl. | — | Not needed (automatic). |
| POP-008 | Enemy destruction kills workers | VERIFIED | med | Not impl. | — | No building destruction. |
| POP-009 | Units consume population | VERIFIED | med | Not impl. | — | Units cost resources only (`game/city.go Train`). |

### Quests and onboarding

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| QST-001 | Optional advisor quests, manual reward claim | VERIFIED | high | Implemented | `game/economy.go Quests`, `ClaimQuest`; AC 17 quests | Exactly-once claim (test `game TestQuestClaimExactlyOnce`). |
| QST-002 | Rewards to Capital; blocked if a Capital battle is < 2 h away | VERIFIED | med | Impl-diff | `ClaimQuest(cityID)`: chosen city, else capital | No battle-proximity guard. |
| QST-003 | Protection rewards (from monster attacks) | VERIFIED | med | Not impl. | — | No monster attacks to protect from. |
| QST-004 | Quest chains | VERIFIED | low | Implemented | `pack.Quest.Requires` (previous quests claimed) | |
| QST-005 | Typical early flow (community) | INFERRED | low | Impl-diff | AC tutorial chain + client `TutorialBar` (`client/src/screens/Shell.tsx`) | |
| QST-006 | Formal tutorial / newbie shield | UNKNOWN | — | Impl-diff | `params.newbieProtectionHours` 72; tutorial quests flagged `tutorial: true` | Our own design; nothing to compare with. |

### Army

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| ARMY-001 | 8 unit categories per race; one race change | VERIFIED | high ("8 types") / med (list) | Impl-diff | `pack/validate.go UnitClasses`: infantry, ranged, cavalry, beast, siege, magic, scout, transport, settler | AC: 12 shared trainable units + 1 race unit per race + 6 neutral-only units. No race change. |
| ARMY-002 | 4 levels per category via war science; upgrade cost difference | VERIFIED | med | Impl-diff | Separate unit types gated by `Unit.Requires` (e.g. `pikeman` needs barracks 3 + `armour` 1) | No in-place upgrade of owned units. |
| ARMY-003 | Unit stats: attack range, health, defence, salary, rating | VERIFIED | med / numbers INFERRED low | Impl-diff | `pack.Unit` (attack, defense %, hp, speed, carry, upkeep, cost, trainSeconds, score) | Attack is a single value (deterministic), not a range. |
| ARMY-004 | Carriers: zero attack, capacity | VERIFIED | med | Implemented | AC `wagon` (attack 0, carry 300); `rules/combat.go Loot` | |
| ARMY-005 | Healers resurrect after battle | VERIFIED | med | Not impl. | — | |
| ARMY-006 | Mercenaries ignore fortification round | VERIFIED | high | Not impl. | — | No fortification round. |
| ARMY-007 | Mages reduce tower damage | VERIFIED | med | Not impl. | — | Mages (`battlemage`) are a damage class only. |
| ARMY-008 | Training pipeline: science → building → queue | VERIFIED | med | Implemented | `game/city.go Train`; `params.trainQueueMax` 5; train-time cap 80% (`MinTimeFactor`) | Batch completes at once. Reference cap is 90%. |
| ARMY-009 | Unit levelling cost on science upgrade | VERIFIED | med | Not impl. | — | |
| ARMY-010 | Gold salary; 0.5%/h desertion; double on missions | VERIFIED | high (gold) / med (0.5%) | Impl-diff | See ECO-015 | Same numbers in AC; differs in that stationed units also pay ×2 and desertion needs net gold < 0. |
| ARMY-011 | Sanctuary (hide troops) | VERIFIED | med | Not impl. | — | |
| ARMY-012 | Captured underground Barracks give training discounts | VERIFIED | med | Not impl. | — | |
| ARMY-013 | Race specialisations affecting the army | VERIFIED | med | Impl-diff | Race modifiers (`attack.ranged`, `attack.cavalry`, `defense.*`, `hp.*`, `upkeep`, `loot`) + race units `thornweaver`, `bannerguard`, `cinderhound`, `bonelord` | |
| ARMY-014 | Monsters as NPC race | VERIFIED | med | Impl-diff | `pack.Neutral` with guard units (`wolf`, `brigand`, `troll`, `wraith`, `golem`, `drake`) | Static defenders only. |

### Missions

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| MSN-001 | Mission taxonomy (combat / hero / non-combat) | VERIFIED | med | Impl-diff | `game/missions.go MissionKinds`: attack, scout, transport, reinforce, settle, occupy | Robbery ≈ attack with loot; Seizure ≈ attack with `capture` on a territory; City Foundation ≈ settle; Relocation/Assistance ≈ reinforce. |
| MSN-002 | Speed tiers 1 / 1.25 / 1.5 / 2 | VERIFIED | high (coeffs) | Not impl. | — | One speed: slowest unit. |
| MSN-003 | Squad award; 15-min minimum duration | VERIFIED | med | Impl-diff | `rules/world.go TravelSeconds`; `params.minTravelSeconds` 900 (AC) ÷ world speed | 900 s minimum one way implemented. Pay = upkeep × `missionUpkeepMult` for the whole time away (no up-front award). |
| MSN-004 | Max 2 attacks on the same domain per 24 h | VERIFIED | med | Implemented | `game/missions.go checkPvP`; `params.maxAttacksPerTargetPerDay` 2 (AC) | Counted per attacker and target coordinates, rolling 24 h (not scaled by world speed). PvP only. No dedicated test. |
| MSN-005 | Recall until "too close" | VERIFIED | med | Impl-diff | `game/missions.go Recall` | Recall allowed at any time while outbound; return takes the time already travelled. |
| MSN-006 | Robbery with 50% cap on outer domains; hero needed for artefacts | VERIFIED | med | Impl-diff | `rules/combat.go Loot`; `params.lootPct` 50; `loot` modifiers; vault protection | 50% applies to cities. No artefact theft. |
| MSN-007 | Slaves capture | VERIFIED | med | Not impl. | — | |
| MSN-008 | Buildings destruction (outside city wall) | VERIFIED | med | Not impl. | — | |
| MSN-009 | Domain seizure (no defensive buildings left) | VERIFIED | med | Impl-diff | `battleTerritory` with `capture: true` | Territories only; cities cannot be captured. |
| MSN-010 | Espionage with access levels 1–4 | VERIFIED | med | Impl-diff | `rules/combat.go Scout`; `game/missions.go scoutCity`, `scoutTile`; `params.scoutDetectBasePct` 30 | Intel 1 (resources, units) or 2 (+buildings, population); detection roll. Test `rules TestScout`. |
| MSN-011 | Resource transport / troop relocation | VERIFIED | med (existence) | Implemented | `transport` and own-city `reinforce` missions | Transport may target **any** city, own or foreign (no ally check). |
| MSN-012 | Assistance at 2× speed | VERIFIED | med | Impl-diff | `reinforce` to own/allied city | Same speed as other missions. |

### Combat

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| CMB-001 | Max 20 rounds; attacker loses on timeout | VERIFIED | high | Implemented | `rules/combat.go Resolve`; `params.combatRounds` 20 (AC; 8 in VF) | |
| CMB-002 | Fortifications pre-round kills attackers | VERIFIED | high | Not impl. | — | Walls add defense instead. |
| CMB-003 | Damage: sum attack, reduce by defence %, proportional spread, overflow redistributed; stochastic | VERIFIED | med / randomness INFERRED | Impl-diff | `rules/combat.go damage`, `applyDamage`, `BuildStack` | Deterministic, simultaneous fire; damage = count × attack × share × (1 + class bonus) × (1 − min(cap, defense)/100); wounds carry over; excess on a destroyed stack is lost (no redistribution). |
| CMB-004 | Pursuer/victim loop: Warriors→Archers→Airborne→Cavalry→Warriors | VERIFIED | high | Impl-diff | `pack.Unit.Bonuses` (`vsClass`, `pct`); `rules/combat.go bonusVs` | AC loop: infantry +40% vs ranged, ranged +40% vs beast, beast +40% vs cavalry, cavalry +40% vs infantry. A damage multiplier, not an extra strike. Test `rules TestClassBonus`. |
| CMB-005 | Retreat thresholds per army | VERIFIED | high | Implemented | `BattleOptions.AttRetreatPct` (per mission `retreatPct`), `DefRetreatPct` (`params.defenderRetreatPct` 100 = never in AC) | Defender threshold is world-wide, not per city. Test `rules TestRetreat`. |
| CMB-006 | Victory rule; defender wins ties | VERIFIED | med | Implemented | `Resolve`: both retreat ⇒ defender holds; attacker wins only if defender destroyed or retreated | |
| CMB-007 | Multiple armies per side | VERIFIED | high | Impl-diff | `battleCity`: garrison + stationed reinforcements, each a `Group` with its owner's modifiers | Attacker side is one army (no joint attack). |
| CMB-008 | Defensive buildings (towers, magic towers) | VERIFIED | med (existence) | Impl-diff | `walls` → defense points; `params.defenseCapPct` 80 | No towers. Test `rules TestCombatWallHelpsDefender`. |
| CMB-009 | Resurrection/necromancy rounds; kill counting after them | VERIFIED | med | Not impl. | — | |
| CMB-010 | Battle calculator (30-run average) | VERIFIED | high | Not impl. | — | `rules/combat.go Power` is only a strength indicator. |
| CMB-011 | Losing hero stunned / imprisoned | VERIFIED | med | Impl-diff | `params.heroWoundHours` 6; `game/missions.go missionReturn`, `battleCity` | Losing attacker hero, or defending home hero when the attacker wins, is wounded for 6 h. |
| CMB-012 | Battle and spy reports | INFERRED | low | Implemented | `game/missions.go report`; `game/social.go Reports`, `Report` | Report shows rounds, losses, loot, heroes, wall. |

### Heroes and progression

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| HERO-001 | Hiring: science + Tavern + rank; daily random offers | VERIFIED | med | Impl-diff | `game/heroes.go HireHero`; `params.heroHireCost` (300 gold + 300 grain) × (heroes owned + 1) | Player picks class and name. No rank gate. |
| HERO-002 | Max 7 heroes | VERIFIED | high | Implemented | `params.maxHeroes` 7; `maxHeroes()` = min(7, War Hall level × `heroesPerBuildingLevel`) | Test `game TestHeroesHireSkillEquip` (`hero_limit`). |
| HERO-003 | Hero classes gate skills | INFERRED (list) / VERIFIED (exists) | low / med | Impl-diff | AC classes `marshal`, `warden`, `pathfinder`, `sage`; `HeroSkill.Classes` | |
| HERO-004 | Hero specialty with 3 levels (economic governor) | VERIFIED | med | Not impl. | — | Heroes affect only armies they lead or the city they defend. |
| HERO-005 | 3 basic skills, +2 per level; active vs passive | VERIFIED | high (frames) / med (counts) | Impl-diff | `game/heroes.go enrichHero`: points = 1 + (level − 1) × `skillPointsPerLevel` (2) | All skills passive. |
| HERO-006 | Spells and mana | VERIFIED | med | Not impl. | — | |
| HERO-007 | Stun, capture, dungeon, ransom | VERIFIED | med | Not impl. | — | Hero capture/ransom not implemented; wounding instead (CMB-011). |
| HERO-008 | XP from killed troops' value; level cap 12 | VERIFIED | med | Impl-diff | `game/missions.go xp`: floor(enemy score lost × `xpPerScore` 0.5 × (1 + hero_xp%)); + neutral level XP; `rules/world.go XPForLevel` (base 200, growth 1.35); max level 30 | |
| HERO-009 | Artefacts, sets (9 of 12), vault | VERIFIED | med | Impl-diff | `pack.HeroItem` (slots weapon, armor, helm, trinket; rarity); `game/heroes.go Equip` | No sets, no vault. Items drop from neutrals. |
| HERO-010 | Alchemy: crystals, perfection, runes | VERIFIED | med | Not impl. | — | |
| HERO-011 | Player rank from rating; rank gates | VERIFIED | med | Impl-diff | `game/settle.go score`; `game/economy.go Rankings` | Ranking only; no named ranks and no gates. |
| HERO-012 | Sciences as progression tree | VERIFIED | med | Implemented | AC 18 techs in branches economy, civic, military, arcane, calling | |

### Social and multiplayer

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| SOC-001 | Rank rule: no attacks on lower-rank cities | VERIFIED | med | Impl-diff | `params.attackScoreFloorPct` in `checkPvP` | Engine supports a score floor; **0 (off) in AC**. |
| SOC-002 | One-time revenge right for 24 h | VERIFIED | med | Not impl. | — | |
| SOC-003 | Newbie protection layers (compassion, Ancients 60 days) | VERIFIED | med (existence) | Impl-diff | `params.newbieProtectionHours` 72; `game/world.go Join`; `checkPvP` | Hard shield on attack and scouting; ends early when the player attacks another player. Test `game TestPvPAttackLootProtectionAndRecall` (`target_protected`). |
| SOC-004 | Max 2 attacks per domain per 24 h | VERIFIED | med | Implemented | See MSN-004 | |
| SOC-005 | Faction loyalty rating | VERIFIED | med | Not impl. | — | |
| SOC-006 | Bounty brotherhood | VERIFIED | med | Not impl. | — | |
| SOC-007 | Temple protection (≤ 5 players) | VERIFIED | low | Not impl. | — | |
| SOC-010 | Clans 5–100 members, roles, two governance types | VERIFIED | med | Impl-diff | `game/social.go` alliances; `params.allianceMaxMembers` 100, `allianceCreateCost` 500 gold; roles leader / officer / member | No science or building needed; one governance type. |
| SOC-011 | Clan warehouse | VERIFIED | med | Not impl. | — | |
| SOC-012 | Clan Castle, War & Peace points, standards | VERIFIED | med | Not impl. | — | Clan castle not implemented. |
| SOC-013 | Alliances via Diplomacy (≤ 7) | VERIFIED | med / cap low | Impl-diff | `SetRelation`: ally / nap / war / neutral between alliances | Relations are directional; `ally`/`nap` block attacks (`friendly`); `ally` allows reinforcing. |
| SOC-014 | Joint attack (one joiner, 50/50 loot) | VERIFIED | med | Not impl. | — | |
| SOC-015 | Joint defence: up to 3 players + owner | VERIFIED | med | Implemented | `params.maxReinforcingPlayers` 3; `missionArrive` case `reinforce` | Extra reinforcement is sent back (`too_many_reinforcing_players`). No dedicated test. |
| SOC-016 | Mentor / apprentice | VERIFIED | med | Not impl. | — | |
| SOC-020 | Player market | VERIFIED | med | Impl-diff | See ECO-017 | Test `game TestMarketConcurrentAcceptAndDelivery`. |
| SOC-021 | Premium-currency auction | VERIFIED | med | Not impl. | — | Auction of premium currency not implemented (by design). |
| SOC-022 | Transfers between own domains; 3% merchant death; gem gifting | VERIFIED | med | Impl-diff | `transport` mission | No losses; premium cannot be transported or gifted. |
| SOC-030 | Mail, contacts, filters, reports | VERIFIED (mail) / UNKNOWN (chat) | med | Impl-diff | `game/social.go SendMessage`, `Messages`, `Block`, `ReportMessage`, notifications | Private + alliance mail, block, report, 60 messages/h limit. No contacts or chat. |
| SOC-031 | Conduct rules, gag, bans | VERIFIED | med | Impl-diff | `game/admin.go Ban`, `ReportedMessages`, `Audit` | Ban only; no gag. |
| SOC-040 | Advisor quests | VERIFIED | med | Implemented | See QST-001 | |
| SOC-041 | Tournaments | VERIFIED | med | Not impl. | — | |
| SOC-042 | Random events at 19:00 | VERIFIED | low | Not impl. | — | |
| SOC-043 | Server types and Faction War mode | VERIFIED | med | Impl-diff | World speed (`worlds.speed` × `params.worldSpeed`, `game/service.go Pack`); pack `modes` | Rule sets differ by pack, not by server type. `modes` (standard, peaceful, speed) are labels; only speed has an engine effect via the world speed multiplier. |
| SOC-044 | Premium touchpoints | VERIFIED | med | Not impl. | — | No premium services in MVP. |

### World and exploration

| ID | Mechanic (reference game) | Status | Conf. | Our impl. | Engine location | Note |
|---|---|---|---|---|---|---|
| WLD-001 | Empire → Kingdom → Square hierarchy | VERIFIED | med | Impl-diff | Flat grid; AC `params.mapWidth` × `mapHeight` = 48 × 48; `rules/world.go Distance` (Euclidean) | |
| WLD-002 | Domain types | VERIFIED | med | Impl-diff | City, territory tile, neutral (`lair`, `ruin`, `node`) | |
| WLD-003 | Capital 51 cells; up to 8 more cities; 4 villages | VERIFIED | high / med | Impl-diff | `params.maxCities` 9 (1 + 8) and `cityCap`; `settle` mission | With keep max 25 and 1 city per 5 keep levels, AC reaches 6 cities, not 9. No cells, no villages. |
| WLD-004 | Outer domains & seizure | VERIFIED | med | Impl-diff | Territories; see ECO-011, MSN-009 | Test `game TestTerritoryOccupyCollectAndContest`. |
| WLD-005 | Ruins: 1–7 per kingdom, respawn, rank-filtered discovery, gems | VERIFIED | med | Impl-diff | `rules/world.go GenerateMap` (`neutralDensityPct` 12, level grows with distance from centre); `Neutral.RespawnSeconds` | No discovery filter; all neutrals visible. |
| WLD-006 | Gem hot-spots (lakes, grail 200/h, relocation) | VERIFIED | med / teleport low | Impl-diff | AC `crater` neutral (starglass 5–50 per clear, respawn 24 h) | Clear-and-respawn, not continuous production. |
| WLD-007 | Exploration mission (2–10% death) | VERIFIED | med | Not impl. | — | No fog of war; scouting a tile reveals guards. |
| WLD-008 | Pits (temporary, +100%) | VERIFIED | med | Not impl. | — | |
| WLD-009 | Mines | VERIFIED | med | Impl-diff | Territory yield on `hills`, `ironridge`, `ashland` | |
| WLD-010 | Gates and tunnels (open 5 min / close 6 h) | VERIFIED | med | Not impl. | — | |
| WLD-011 | Camps (mobile bases) | VERIFIED | med | Not impl. | — | |
| WLD-012 | Barracks / underground estates | VERIFIED / UNKNOWN | med | Not impl. | — | |
| WLD-013 | Clan Castle siege | VERIFIED | med | Not impl. | — | |
| WLD-014 | Dungeon (cross-ref HERO-007) | (see HERO-007) | — | Not impl. | — | |
| WLD-015 | Monster attacks on players | VERIFIED | med | Not impl. | — | Neutrals never attack. |

## Engine features with no research ID (our own design)

| Feature | Location | Label |
|---|---|---|
| Food consumption and starvation (1% population per 10-min tick at zero grain and deficit) | `rules/economy.go TickStep`; `params.foodPerPopPerHour`, `starvationPctPerTick` | PROPOSED |
| Territory yield collected manually, capped at 24 h | `game/economy.go territoryYield`, `Collect` | PROPOSED |
| NPC exchange as a resource sink | `game/economy.go Exchange` | PROPOSED (shape from ECO-018) |
| Cancel last queued job for 80% refund | `game/city.go CancelJob`; `params.cancelRefundPct` | PROPOSED |
| Deterministic seeded RNG for map, scouting and drops | `rules/world.go RNG`, `SeedFrom` | PROPOSED (engine principle) |
| Path-independent offline economy (absolute 10-min ticks) | `rules/simulate.go AdvanceCity` | PROPOSED (engine principle); test `rules TestAdvancePathIndependence` |
| World speed multiplier for speed worlds | `game/service.go Pack`; `cmd/server/main.go` (`RF_WORLDS` `slug:pack:name:speed`) | PROPOSED |
| Pack migrations for live worlds | `game/admin.go MigrateWorldPack` | PROPOSED |
