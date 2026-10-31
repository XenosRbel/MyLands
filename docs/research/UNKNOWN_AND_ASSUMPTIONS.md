# Unknowns, Source Conflicts and Assumptions

Owner: Agent B (Game Designer). Date: **2026-10-31**. Code checked at commit `dfa660e`; pack `packs/ashen-crowns/pack.json` v1.0.0 ("AC").

This document consolidates every open question from the research notes, the places where sources disagree, the values we chose ourselves (PROPOSED) for the Ashen Crowns pack, and implementation notes found while comparing code with research.

**Evidence caveat (applies to every row).** All research evidence came from search-engine excerpts. No page of the reference game was opened directly and no gameplay was observed (see `SOURCE_REGISTER.md`). Statuses below are copied from the research notes and are never raised here.

**ID prefixes in this file:** `U-` unknown about the reference game, `C-` source conflict, `A-` our assumption (PROPOSED value), `I-` implementation note about our own engine/pack.

---

## 1. Unknowns about the reference game

"Impact" describes what changes in **our** product if the true answer differs from what we assumed. *Arch* = architecture/engine impact; *Rule* = game-rule or balance impact. Many unknowns have little impact because our game is an original design that does not need to match the reference.

| ID | Unknown | Research IDs | Verification method | Impact if the answer differs from our assumption |
|---|---|---|---|---|
| U-01 | Current race roster (4 vs up to 7) and which races live servers offer | RACE-001..008, FAC-001 | Open the Race page on `steam.mlgame.org` and the registration screen; compare mirror revisions | Arch: none (races are pack data). Rule: none; AC uses 4 original races. |
| U-02 | Exact race bonus percentages (baked into building base rates) | RACE-002..008, ECO-010 | Per-race building info pages or in-game building windows | Arch: if bonuses are per-building base rates, our pack would need per-race building variants (`Building.Races`) instead of race modifiers; both exist. Rule: pack data only. |
| U-03 | All per-level tables: costs, times, outputs, storage, housing, job counts | BLD-000..026, ECO-012, POP-004 | Fetch individual building pages (Warehouse, Cottage, Main Building …) | Arch: our formulas are geometric (`base × growth^(L−1)`, `rules/economy.go CostAt`, `DurationAt`, `EffectValue`). If the reference uses irregular tables and we want that shape, the pack format needs a table form for `Cost`, `Duration` and `Effect` (format version bump). Rule: pacing only. |
| U-04 | Unit stat tables and the server coefficient | ARMY-003 | UnitsParameters (EN) / UnitParametrs (RU) pages | Arch: none; a server coefficient maps onto pack params or the world speed multiplier. Rule: pack data. |
| U-05 | Travel-time formula, distance metric and base unit speeds | MSN-002, WLD-001 | Combat/Units pages; gameplay video | Arch: we use Euclidean tiles / slowest unit (`rules/world.go TravelSeconds`). A kingdom/square hierarchy would need a new coordinate model and map API. Rule: tempo of raids. |
| U-06 | Combat internals: attack ranges and randomness, overflow redistribution, size of pursuer extra strike, fortification kill formula, mage reduction units | CMB-002, CMB-003, CMB-004, ARMY-007 | RU Combat page in full; Calculator page; 30-run calculator behaviour | Arch: `rules/combat.go Resolve` is deterministic. Adding randomness means a seeded RNG per battle (pattern exists in `rules/world.go RNG`) and changes to reports; a calculator would need Monte-Carlo runs. Rule: variance of outcomes. |
| U-07 | How production modifiers combine (add vs multiply) | ECO-010 | Resources page tables; any worked example | Arch: `rules/modifiers.go Apply` sums percentages. Multiplicative stacking changes this one function and every balance number. Rule: large late-game difference. |
| U-08 | Whether population or armies eat crops | ECO-002, ECO-015 | Resources/Population pages; gameplay | Arch: none (food per pop is a param; can be set to 0). Rule: AC food upkeep and starvation are PROPOSED (A-12). |
| U-09 | Trading Post formula (idle × tax?) | ECO-005, BLD-015 | Trading Post page | Arch: `idle_tax` effect already models "idle × rate". Rule: gold income curve. |
| U-10 | Does production stop at the storage cap in addition to surplus decay? Scope of decay (domain vs city) | ECO-012, ECO-013 | Resources page EN vs ES/RU | Arch: `rules/economy.go Accrue` either hard-caps (decay = 0) or lets stock exceed the cap (decay > 0). Doing both (halt **and** decay) needs one extra branch. Rule: whether offline overflow is punished. |
| U-11 | Base of the overcrowding flight percentage (excess vs total) | POP-006 | Population page | Arch: would replace the clamp in `TickStep` with a formula. Rule: housing loss speed. |
| U-12 | Gate timing semantics; what a closed gate blocks for its owner | BLD-002, WLD-010 | Gate page; video | Arch: a gate would be a new city state machine plus mission validation. Rule: defensive play option (not in MVP). |
| U-13 | Main Building max level (5, RU only); city cell counts | BLD-001, BLD-003 | EN Main Building and City pages | Arch: cell layout is not implemented; adopting it is a large change (city grid, placement API, client layout). Rule: none today. |
| U-14 | Hero details: hiring rank, class list, stun duration, mana pool and regen, XP curve, XP multipliers other than "attacker wins", equipment slots | HERO-001, -003, -006, -007, -008, -009 | Heroes, HeroClasses, Skills, Spells, Stunned, Dungeon pages; video | Arch: spells/mana/capture would add hero state machines and new mission kinds. Rule: hero power curve. |
| U-15 | Rank-rule wording on Mining/Tactical servers ("required rank is Suzerain") | SOC-001 | Rank page in full | Arch: none; `params.attackScoreFloorPct` covers a score floor. Rule: anti-bullying strength. |
| U-16 | Compassion spell effect; whether a formal tutorial or newbie shield exists | SOC-003, QST-006 | Rank, Quests pages; new-account playthrough | Arch: none. Rule: AC uses a 72 h shield (A-07). |
| U-17 | Real-time chat existence; message retention | SOC-030 | FAQ; video | Arch: chat would need a push channel (WebSocket) and moderation tooling. Rule: social retention. |
| U-18 | Alliance size cap (7, RU only) | SOC-013 | EN Alliance/Diplomacy pages | Arch: none. Rule: AC uses 100-member alliances (A-17). |
| U-19 | Market tax, price bounds, lot limits; auction lot supply | ECO-017, ECO-019, SOC-020, SOC-021 | Market and Auction pages | Arch: none for fees. Rule: market liquidity. |
| U-20 | Tournament prize tables and event calendar | SOC-041, SOC-042 | Tournament pages (EN/RU) | Arch: events need a scheduler for world-wide timed rules and leaderboards. Rule: none in MVP. |
| U-21 | Full list of server types and their rule differences | SOC-043 | Serv page | Arch: our equivalent is "one pack + world speed per world". Rule: mode design. |
| U-22 | Outer-domain limits, pit lifespan, ruin respawn timing, gem amounts, Salt Lake rate | ECO-006, ECO-011, WLD-004..008 | Pit, Lake, Ruin, OuterDomains pages | Arch: lifespans would need expiry events for territories. Rule: premium faucet size. |
| U-23 | Robbery "burn" rule | MSN-006, SOC-011 | Resources page | Arch: none. Rule: loot efficiency. |
| U-24 | Recall threshold ("too close"); Observation mission | MSN-005, MSN-010 | Combat page | Arch: `game/missions.go Recall` would get a distance/time check. Rule: commitment of attacks. |
| U-25 | Battle-report layout | CMB-012 | Gameplay video | Arch: none (report is JSON). Rule: UX only. |
| U-26 | Specialisation costs, the Defender values, the full list | TECH-020, HERO-004 | Spec page (EN/ES) | Arch: none. Rule: see C-14. |
| U-27 | Science prerequisite graph, costs and level counts | TECH-001, TECH-003..019 | ScienceDescription and per-science pages | Arch: none (pack requirements). Rule: pacing. |
| U-28 | Healer resurrection numbers | ARMY-005 | UnitsDescription page | Arch: resurrection would add a post-battle phase in `rules/combat.go`. Rule: attrition. |
| U-29 | Whether the reference game and its mobile apps still operate in 2026 | GAME_OVERVIEW.md, Unknowns item 3 | Official site and stores, once reachable | Arch: none. Rule: none; matters only for market positioning. |
| U-30 | Linked-account (shared IP) interaction rules | SOC-031 | Rules page | Arch: anti-multi-account enforcement would need device/IP linkage and interaction checks. Rule: fair play. |

---

## 2. Source conflicts

| ID | Topic | Claim A (source) | Claim B (source) | What we did | How to resolve | Impact if B (or A) is right |
|---|---|---|---|---|---|---|
| C-01 | Number of races | "one of four races" (SRC-A1-03, SRC-A1-36, SRC-A1-38) | 7 races across Light/Dark (SRC-A1-05, SRC-A1-06, SRC-A2-18) | 4 original races, 2 per faction | Read live Race page + registration | None for us (data-driven). |
| C-02 | Knights' bonus | gold, carriers, warehouses (SRC-A1-04, SRC-A1-05) | crops (forum SRC-A1-31); gold + grain (RU fan SRC-A1-42); A2 lists crops too (ARMY-013, SRC-A2-17) | Not copied; no AC race has this profile | Read Race page in full | None. |
| C-03 | Deadland native land | Demons (SRC-A1-05) | Undead (one Battle snippet, SRC-A1-26) | Not implemented (RACE-009) | Read Battle/Kingdom tables | None. |
| C-04 | Faction penalty stacking | −50% vs own race, −25% vs own faction (SRC-A1-06) | −25% faction + −25% race = −50% total (forum SRC-A1-31) | Same race −50% (not stacked with −25%); different race, same faction −25% (`game/missions.go factionMods`) | Read Factions page; test in game | Numerically identical either way for the same-race case. |
| C-05 | Developer and release date | Elyland (SRC-A1-36, -37); Steam 16 Sep 2014 (SRC-A1-03) | Gravvit Ltd / Forbes Consult (SRC-A1-39); Metacritic 14 Mar 2014 (SRC-A1-41) | Irrelevant to design | Steam page | None. |
| C-06 | Extra cities | up to 8 more (SRC-A2-47, EN/IT/new RU) | up to 7 more (older RU, SRC-A1-14 mirrors) | `params.maxCities` 9 = 1 + 8 (but see I-01) | City page revisions | Changing to 8 total is a one-param change. |
| C-07 | City cell counts | level-5 city 39–43 cells (EN) | level-4 city 34–38 (RU) (both SRC-A1-14) | Cells not implemented | City page | None today. |
| C-08 | Overflow decay scope | per domain (EN, SRC-A1-04) | per city (ES/RU mirrors, SRC-A1-04) | Per city (our only domain with storage) | Resources page | None (territories have no storage). |
| C-09 | Fortifications | kill attackers in a pre-round (SRC-A2-06, SRC-A2-07) | raise the domain's defence level (forum SRC-A2-14 / SRC-A1-44) | AC `walls` add defense points (closer to B) | Combat page in full | If we wanted A: new pre-round in `rules/combat.go Resolve`. |
| C-10 | Healer numbers | "1 to 3" | table 1 (Elves 2) → 4 (Elves 5) (both SRC-A2-03) | Not implemented | UnitsDescription | None today. |
| C-11 | Hero hiring rank | Captain (current wiki, SRC-A2-19) | Suzerain, not under attack (2011 blog SRC-A2-35) | No rank gate; War Hall level gates slots | Heroes page revision history | Adding a gate needs named ranks (not implemented). |
| C-12 | Dungeon levels | 1–9 (EN, SRC-A2-26) | 1–7 (VN mirror) | Not implemented | Dungeon page | None. |
| C-13 | Alchemy levels and details | levels 1–6, L6 monumental buildings (SRC-A1-24) | 7 levels, +3 perfect upgrades at L7; EN vs ES differ (SRC-A2-65) | Not implemented | Alchemy pages (EN/ES/RU) | None. |
| C-14 | What a "specialisation" is | one **player** specialisation; a second cancels the first (A1, TECH-020, SRC-A1-12) | each **hero** has a 3-level specialty (A2, HERO-004, SRC-A2-45) — same Spec page, different reading | Player-level, mutually exclusive "Paths" (`exclusiveGroup: calling`), permanent | Read Spec page (EN and ES) in full | If hero-level: specialisations move to `HeroConfig` (new field) and heroes become economic governors (HERO-004). |
| C-15 | Faction rosters by language | ES omits Chimeras (SRC-A2-18) | other mirrors include them (SRC-A1-06) | n/a | Mirrors | None. |
| C-16 | Gate timing | "opens 5 minutes and closes 6 hours" read as ambiguous (A1, BLD-002) | read as open/close transition times (A2, WLD-010) | Not implemented | Gate page | None today. |
| C-17 | Time cap of bonuses | construction time reduction ≤ 80% (BLD-000) | training acceleration ≤ 90% (ARMY-008) | One cap for all durations: 80% (`rules/modifiers.go MinTimeFactor` 0.2) | Not a true conflict (different scopes) | A separate training cap would need a per-target floor. |

---

## 3. Assumptions: PROPOSED values in the Ashen Crowns pack

All values below are **PROPOSED** (our design), even where a research ID inspired the value. A research anchor means "the shape or number was taken from research"; it does not make the value verified for our game. Validation method for every row: playtest plus economy telemetry once worlds run; numbers are changed in the pack, not the code.

| ID | Parameter / data (AC) | Value | Research anchor | Why we chose it | Impact if changed |
|---|---|---|---|---|---|
| A-01 | `worldSpeed` | 1 (standard); speed worlds multiply it | SOC-043 (server types exist) | One pack serves normal and speed worlds | Speed worlds scale timers and flows; see I-08 for what is not scaled. |
| A-02 | `economyTickSeconds` | 600 | — | Population/desertion/decay granularity; path independence | Shorter ticks = smoother but more CPU per settle. |
| A-03 | `buildQueueMax` / `trainQueueMax` / `researchQueueMax` | 3 / 5 / 2 | — | Allows short offline plans without premium queues | Session length and return frequency. |
| A-04 | `cancelRefundPct` | 80 | — | Mistakes cost a little | Exploit risk if 100. |
| A-05 | `mapWidth` × `mapHeight` | 48 × 48 | — (map size UNKNOWN, WLD-001) | Dense neighbourhoods for a small MVP population | Too small ⇒ crowding (`world_full`); too big ⇒ isolation. |
| A-06 | `neutralDensityPct` | 12 | WLD-005 (ruins exist) | PvE targets within a short march | PvE availability. |
| A-07 | `newbieProtectionHours` | 72 | SOC-003 (protection exists; values differ) | Simple, readable shield; ends when the player attacks | Early churn from being farmed vs. shield abuse. |
| A-08 | `minTravelSeconds` | 900 | MSN-003 (15-min minimum) | Gives defenders time to react | Raid tempo. |
| A-09 | `combatRounds` | 20 | CMB-001 | Matches reference structure | With defender retreat 100, attacks must annihilate within 20 rounds (I-12). |
| A-10 | `lootPct` | 50 | MSN-006 (50% on outer domains) | Raids hurt but never wipe out unprotected stock | Raiding profitability. |
| A-11 | `popGrowthMode` / `popGrowthPctPerHour` | `free_housing` / 12 | POP-002 (research says 0.04 × idle; we chose a different mode) | Growth slows as housing fills: easy to read, no runaway | Switching to `idle` (engine-supported) changes early growth shape. |
| A-12 | `foodResource` / `foodPerPopPerHour` / `starvationPctPerTick` | grain / 0.12 / 1 | None (U-08: no evidence that people eat) | Gives grain a steady sink and a reason to farm | Removing food upkeep would leave grain mostly for training and trade. |
| A-13 | `upkeepResource` / `desertionPctPerHour` / `missionUpkeepMult` | gold / 0.5 / 2 | ECO-015, ARMY-010, MSN-003 | Army size is capped by the gold economy | Larger or smaller standing armies. |
| A-14 | `overflowDecayPctPerHour` | 1 | ECO-013 | Soft cap; offline overflow is not lost at once | Hoarding behaviour. |
| A-15 | `marketFeePct` / `marketMaxOpenOrders` | 5 / 10 | ECO-017 (tax exists; base UNKNOWN) | Mild sink; limits spam | Trade volume. |
| A-16 | `exchangeRates` / `exchangeFeePct` | timber 1, grain 1, stone 1.2, iron 1.6, gold 3 / 30 | ECO-018 ("worse rates") | NPC exchange is a worse fallback than the player market; also defines the market price band (10x) | Liquidity; abuse resistance of the price band. |
| A-17 | `allianceCreateCost` / `allianceMaxMembers` | 500 gold / 100 | SOC-010 (5–100 members) | Upper bound from research; no minimum | Alliance size and power concentration. |
| A-18 | `cityCapBuilding` / `citiesPerCapLevels` / `maxCities` | keep / 5 / 9 | WLD-003 (1 + 8 cities) | Keep level paces expansion | See I-01 (only 6 reachable). |
| A-19 | `territoriesPerLevel` / `territoryYieldPerHour` / `territoryYieldCapHours` / `territoryBonusPct` | 1 / 25 / 24 / 0 | ECO-011 (contested nodes) | Map control pays; uncollected yield stops after a day to reward check-ins | Map competition intensity. |
| A-20 | `heroHireCost` / `heroBuilding` / `heroesPerBuildingLevel` / `maxHeroes` | 300 gold + 300 grain (× n+1) / warhall / 1 / 7 | HERO-002 (max 7) | Rising cost per hero | Hero availability. |
| A-21 | `heroWoundHours` | 6 | CMB-011 (stun exists; duration UNKNOWN) | Loss costs time, not the hero | Hero risk. |
| A-22 | `attackScoreFloorPct` | 0 (off) | SOC-001 (rank rule) | Small MVP worlds; protection by shield instead | Turning it on protects weak players but can block revenge. |
| A-23 | `defenseCapPct` / `defenderRetreatPct` | 80 / 100 | RACE-009 says max Defence 50 (different concept); CMB-005 | Walls + defense stay meaningful; defenders hold to the end | Defensive strength; see I-12. |
| A-24 | `sameRaceAttackPct` / `sameFactionAttackPct` | −50 / −25 | FAC-002 | Steers war to the other faction | Intra-faction conflict level. |
| A-25 | `maxAttacksPerTargetPerDay` | 2 | MSN-004 | Anti-farming | Raid pressure on one target. |
| A-26 | `maxReinforcingPlayers` | 3 | SOC-015 | Matches "3 + owner" | Defensive cooperation. |
| A-27 | `scoutDetectBasePct` | 30 | MSN-010 (shape differs) | Scouting is never fully safe | Intel value. |
| A-28 | All building, tech, unit, neutral and quest numbers; start state (600 timber, 600 grain, 500 stone, 250 iron, 400 gold, 20 starglass, 40 people, 10 militia, 2 scouts) | see pack | U-03, U-04 (tables UNKNOWN) | Original balance | Full rebalance possible without code changes. |
| A-29 | `heroes.maxLevel` / `xpBase` / `xpGrowth` / `skillPointsPerLevel` / `xpPerScore` | 30 / 200 / 1.35 / 2 / 0.5; skill points = 1 + 2 × (level − 1) | HERO-005 (3 + 2/level), HERO-008 (cap 12) | Longer hero progression than the reference | Hero power ceiling. |
| A-30 | Specialisation techs (`exclusiveGroup: calling`) | three "Paths", require keep 5, permanent | TECH-020 (one at a time) / C-14 | Simple, build-defining choice | Respec would need a cancel rule (I-13). |

---

## 4. Implementation notes (found while checking code against research)

These are observations about our current engine and pack. They are documentation only; no code was changed. Each lists the evidence (code location) and the impact.

| ID | Observation | Evidence | Impact | Suggested follow-up |
|---|---|---|---|---|
| I-01 | AC can reach only **6 cities**, not `maxCities` 9 | `game/missions.go cityCap` = min(`maxCities`, 1 + max keep level / `citiesPerCapLevels`); keep `maxLevel` 25 ⇒ 1 + 5 = 6 | Design intent (1 + 8 cities, WLD-003) not reached | Decide: lower `citiesPerCapLevels` to 3, or set `maxCities` 6 and document it. |
| I-02 | `params.territoryBonusPct` is declared and validated but never read | Only in `pack/types.go` and `pack/validate.go` | Setting it has no effect (AC sets 0, engine default 5) | Implement or remove from the format. |
| I-03 | Market Hall's `market_fee` modifier has no effect | `AcceptOrder` uses `c.mods` (race + techs); building modifiers are only in `cityMods`. The fee is computed with the **accepting** player's mods but deducted from the seller's proceeds | Market Hall level only adds carry; descriptions promise lower fees | Decide whose mods apply (seller's city is natural) and fix in code. |
| I-04 | Building `carry` bonus (Market Hall +3%/level) is used for the cargo check at dispatch but not for loot capacity at arrival | `SendMission` uses `cityMods(ci)`; `battleCity` loot uses `attMods` = owner player mods + hero | Small inconsistency in loot | Store the dispatch capacity on the mission. |
| I-05 | Watchtower `scout` bonus helps only the defending city | `scoutCity`: attacker power uses player mods + hero; defender uses `cityMods(ci)` | Watchtower description says "spies and counter-espionage" | Same fix pattern as I-04, or update the text. |
| I-06 | `Neutral.Capturable` is not read by game code | No reference outside `pack/` | Any cleared neutral tile can be occupied while it is cleared, including `barrow` (capturable: false) | Enforce the flag in `occupy`/`settle`. |
| I-07 | `transport` missions may target **any** city (no ally or protection check) | `SendMission` case `transport` checks only that the target is a city | Resource pushing between accounts bypasses the market price band (anti-abuse rule R-MKT-6 in code comments) | Restrict to own/allied cities, or log for moderation. |
| I-08 | Some time limits are not scaled by world speed | Newbie shield (`Join`), 24 h attack window (`checkPvP`), `heroWoundHours`; territory yield cap is in game hours | On a ×10 world the shield lasts 72 real hours (720 game hours) | Decide per limit whether it is "real time" or "game time". |
| I-09 | Starglass has **no sink** | No cost in AC uses `starglass`; no service code | Players accumulate a currency with no use | PROPOSED sinks in `GAME_DESIGN_DOCUMENT.md` §17. |
| I-10 | Lore is a per-city stock, not a pooled flow | `lore` is a normal resource (`capped: false`); research pays from the queuing city | Multi-city players must move lore by transport | Accept (simple) or add a "player-wide resource" flag. |
| I-11 | `bonelord` text says "tireless and unpaid" but upkeep is 0.4 gold/h | AC pack | Text/data mismatch | Fix text or set upkeep 0. |
| I-12 | With `defenderRetreatPct` 100 an attacker wins only by destroying every defender within 20 rounds | `rules/combat.go Resolve` | Strong defence bias | Intended for MVP; revisit after playtest. |
| I-13 | Specialisation choice is permanent | `game/city.go exclusiveTaken` blocks other techs of the group once researched or queued | No respec; reference cancels the old one (TECH-020) | Optional respec rule later. |
| I-14 | Several rules have no dedicated integration test: same-race/faction penalty, attack cap per target, reinforcing-player cap, specialisation exclusivity | `game/game_test.go` and `api/api_test.go` assert other rule codes (`target_protected`, `friendly_target`, `hero_limit`, …) but not `attack_limit`, `too_many_reinforcing_players`, `specialisation_taken` | Regressions could go unnoticed | Add tests (see PRD Definition of Done). |
| I-15 | Code comments cite rule IDs in `GAME_RULES.md` (e.g. R-CMB-3) but that file is not in `docs/` | `rules/*.go`, `game/*.go` comments | Readers cannot follow the references | Write `docs/spec/GAME_RULES.md` or update comments. |
| I-16 | `pack.Unit.Upkeep` comment says "food per hour"; the paying resource is `params.upkeepResource` (gold in AC) | `pack/types.go` | Documentation only | Update comment. |
| I-17 | Faction IDs (`dawnpact`, `duskbound`) have no localized names | `Race.Faction` is a plain string | UI must show raw IDs or invent names | Add faction names to `strings` (PROPOSED in GDD). |

---

## 5. Verification backlog (priority order)

1. **Combat** (U-06, C-09): open Combat EN/RU, Calculator, UnitsDescription. Decides whether we add randomness or fortification rounds.
2. **Economy formulas** (U-07, U-10, U-03): Resources, Population, building pages. Decides modifier stacking and storage behaviour.
3. **Specialisations** (C-14, U-26): Spec page EN/ES.
4. **Protection and PvP** (U-15, U-16): Rank, Quests, Ansients pages.
5. **Heroes** (U-14): Heroes, HeroClasses, Skills, Stunned, Dungeon.
6. **Gameplay video** (SRC-A2-62): first OBSERVED evidence for UI, reports and map.
7. **Code follow-ups** I-01, I-03, I-07, I-08, I-14 (engineering; not research).
