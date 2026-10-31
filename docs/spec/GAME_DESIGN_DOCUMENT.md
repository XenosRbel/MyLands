# Game Design Document: Ashen Crowns

| Field | Value |
|---|---|
| Owner | Agent B (Game Designer) |
| Status | DRAFT v1 |
| Date | 2026-10-31 |
| Content | `packs/ashen-crowns/pack.json` v1.0.0 (title "Ashen Crowns", subtitle "Realms of the Sundered Sky", locales EN + RU) |
| Engine | Realmforge, commit `dfa660e`, re-checked at `ed6acc7` (no game-rule changes) |
| Companion docs | `GAME_RULES.md` (normative formulas), `PRD.md`, `../research/MECHANICS_CATALOG.md`, `../research/UNKNOWN_AND_ASSUMPTIONS.md`, `../research/SOURCE_REGISTER.md` |

## 0. Conventions

- **[R: ID]** — the *mechanic's shape* is derived from research on the reference genre game (ID in `MECHANICS_CATALOG.md`). Research evidence came from search-engine excerpts only (no page opened directly), so the reference rule itself is not established fact. **Numbers are still PROPOSED** unless the text says "value from research".
- **[P]** — PROPOSED: our own design, not derived from research.
- **(not implemented)** — described for completeness or as future design; the engine does not do it today.
- All numbers are quoted from the AC pack at world speed 1. Formulas are the engine's (`server/internal/rules`). Rounding follows the code (costs use round-half-away-from-zero; durations round up).

---

## 1. Vision and pillars

*A slow-burning realm builder where every hour away still counts, war is costly, and the sky's shards are earned, never bought.* [P]

| Pillar | Meaning | Basis |
|---|---|---|
| Fair persistence | The realm keeps working offline; nothing is lost just by logging off | [P]; soft storage cap [R: ECO-013] |
| Economy before war | Armies cost wages every hour, so a war chest must be built first | [R: ECO-015] |
| Rock-paper-scissors armies | Four main classes counter each other in a loop | [R: CMB-004] |
| Map that matters | Lairs, craters and territories create local contests | [R: ECO-011, WLD-005] (shape) |
| Banners together | Alliances, diplomacy and reinforcement are central | [R: SOC-010, SOC-015] |
| No pay-to-win | No real-money purchases; premium shards are earned in play | [P] (deliberate contrast to genre criticism, SRC-A1-39) |

## 2. Setting

**Premise** (pack string `lore.intro`): "When the sky shattered, the old crowns fell. From the ash rise new houses, hungry for starglass." The player is greeted with: "Welcome, ruler. Your keep stands among ashes — rebuild it." (`tutorial.welcome`). [P]

**The world.** A broken realm under a "sundered sky". Shards of the fallen sky — **starglass** — lie in **Starfall Craters**, guarded by stone golems and, at the deepest sites, young **Ash Drakes**. Old **barrows** hold grave-goods and lore under the watch of wraiths. Wolves and brigands haunt the roads; hill trolls lurk in the camps. [P] (texts from pack `neutrals` and `units`)

**Lands** (pack `terrains`): Plains, Forest, Hills, Iron Ridge, Ashland, Marsh (passable); Mountains and Lake (impassable). [P]

### 2.1 Factions and races

Two pacts divide the new houses. The pack stores the faction IDs `dawnpact` and `duskbound`; display names "the Dawnpact" and "the Duskbound" are PROPOSED (no localized names exist yet, `UNKNOWN_AND_ASSUMPTIONS.md` I-17). Two opposing factions with races grouped under them is [R: FAC-001]; all names, lore and bonuses are [P].

| Race (ID) | Faction | Fantasy (pack text) | Modifiers | Race unit |
|---|---|---|---|---|
| Thornkin (`thornkin`) | dawnpact | "Forest-dwellers bound to old groves. Masters of timber and the bow." | timber production +15%, ranged attack +10%, march speed +5% | Thornweaver (ranged) |
| Ironvow Houses (`ironvow`) | dawnpact | "Sworn knightly houses. Stubborn defenders and fine riders." | defense +8%, cavalry attack +10%, stone production +10% | Bannerguard (infantry) |
| Ashborn (`ashborn`) | duskbound | "Fire-touched clans of the burnt lands. Fierce raiders and smiths." | attack +8%, iron production +15%, loot +10% | Cinderhound (beast) |
| Gravewardens (`gravewarden`) | duskbound | "Keepers of barrow and crypt. Their soldiers ask little pay." | army upkeep −20%, hit points +8%, lore production +15% | Bone Lord (mage) |

Race is chosen when joining a world [R: RACE-001] and cannot be changed [P] (the reference reportedly allows one change, ARMY-001). Race asymmetry through economy and army bonuses is [R: RACE-001, ARMY-013]; the specific bonuses are [P].

---

## 3. Core loop

```
 produce ──► grow people ──► build / research ──► train
    ▲                                              │
    │                                              ▼
 collect yield, claim quests ◄── act on the map (scout, clear lairs,
    ▲                             occupy, raid, reinforce, settle)
    └──────── trade, ally, compete ◄───────────────┘
```

Shape [R: GAME_OVERVIEW §3 core loop]; our version drops cells, villages, gates and deposits (see `MECHANICS_CATALOG.md`) and adds food and territory yield [P].

---

## 4. Entities

| Entity | Scope | Key state | Basis |
|---|---|---|---|
| Account | Global | login or guest, locale | [P] |
| Player | One per account per world | race, techs (player-wide), stats, score, shield end, alliance and role | [R: RACE-001] |
| City | Owned by a player; first is the capital | position, stock per resource, population, building levels, garrison | [R: ECO-024] |
| Building | Per city, one instance per type, with a level | level 0..max | [P] (no cells) |
| Technology | Per player, with a level | level 0..max | [R: TECH-001] |
| Unit stack | In a city or on a mission | unit type → count | [P] |
| Hero | Per player, based in a city | class, level, XP, skills, items, status, wound timer | [R: HERO-001] (shape) |
| Item | Per player, optionally equipped | definition, slot | [R: HERO-009] (shape) |
| Tile | World map | terrain, neutral and its level, cleared-until, owner, city | [P] |
| Territory | A tile owned by a player | owner, owned-since, yield collected at, stationed troops | [R: ECO-011] (shape) |
| Mission | Player-owned | kind, units, hero, cargo, depart/arrive/return times, status, result | [R: MSN-001] (shape) |
| Report | World | battle or scout data, recipients with side | [R: CMB-012] |
| Alliance | World | name, tag, members, roles, relations | [R: SOC-010] |
| Market order | World | sell/buy resource and amounts, escrow, status | [R: SOC-020] |
| Quest | Pack | goal type and amount, prerequisite quests, rewards | [R: QST-001] |

---

## 5. Resources

| Resource | Kind | Storage | Trade / loot | Main sources | Main sinks | Basis |
|---|---|---|---|---|---|---|
| Timber | basic | capped (base 1,500) | tradable, lootable | Lumberyard; Forest territories | almost every building; ranged units; wagons | [R: ECO-001] (role) |
| Grain | food | capped (base 1,500) | tradable, lootable | Farmstead; Plains territories | population eats 0.12/person/h; units | [R: ECO-002] (role); eating is [P] |
| Stone | basic | capped (base 1,500) | tradable, lootable | Quarry; Hills territories; craters | walls, keep, vault, specialisation | [R: ECO-003] |
| Iron | basic | capped (base 1,000) | tradable, lootable | Ironworks; Iron Ridge territories | units, military techs | [R: ECO-004] |
| Gold | rare (currency) | capped (base 3,000) | tradable, lootable | Trading Post (idle tax), Mint, Ashland territories, lairs, quests | **army wages**, keep, heroes, alliances | [R: ECO-005, ECO-015] |
| Lore | research | uncapped, per city | not tradable, not lootable | Scriptorium; barrows; quests | technologies; mages | [R: ECO-007] (role); per-city stock is [P] |
| Starglass | premium | uncapped | **not tradable, not lootable, not transportable** | start 20, quests, barrows, craters | none yet (§17) | [R: ECO-006] (shape: premium, not stealable); earned-only is [P] |

**Storage rule.** Each city stores its own resources [R: ECO-024]. Capacity = (base + Σ storage effects) × (1 + storage%). Stock may exceed capacity (from loot, transport, quests or production); the surplus **decays 1% per hour** [R: ECO-013, value from research]. Production does not stop at the cap [P] (`UNKNOWN_AND_ASSUMPTIONS.md` U-10).

**Protection.** The Hidden Vault hides an amount of every resource from raiders [R: ECO-014] (as a building [P]).

---

## 6. Construction

Rules [P] unless marked:
- One instance of each building type per city, with levels. No cells or layout (cell-based city is not implemented; [R: BLD-003] not adopted).
- Cost(L) = round(base × growth^(L−1)); time(L) = ceil(baseSeconds × growth^(L−1) × timeFactor ÷ worldSpeed).
- timeFactor = max(0.2, 1 + Σ build_time%/100): reductions are capped at **80%** [R: BLD-000, value from research].
- Queue: 3 jobs per city, run one after another. Only the last job can be cancelled, refunding 80%.
- Effects grow as v(L) = perLevel × L × growth^(L−1).

| Building | Category | Max | Requires | Effect per level (v(L) formula inputs) | Basis |
|---|---|---|---|---|---|
| Keep | core | 25 | — | housing 60 (growth 1.04); build time −2%; every 5 levels +1 city; +1 territory per level | [R: BLD-001] (central building gates progress) |
| Cottages | economy | 25 | — | housing 90 (growth 1.06) | [R: BLD-019] |
| Lumberyard | economy | 25 | — | timber 32/h, 6 workers (growth 1.06) | [R: BLD-010] |
| Farmstead | economy | 25 | — | grain 36/h, 6 workers (1.06) | [R: BLD-011] |
| Quarry | economy | 25 | — | stone 26/h, 6 workers (1.06) | [R: BLD-012] |
| Ironworks | economy | 25 | Keep 2 | iron 20/h, 8 workers (1.06) | [R: BLD-013] |
| Trading Post | economy | 20 | Keep 2 | 0.06 gold/h per idle person | [R: BLD-015, ECO-005] (idle × tax, INFERRED in research) |
| Mint | economy | 20 | Keep 4, Ironworks 2 | gold 14/h, 10 workers (1.06) | [R: BLD-014] |
| Scriptorium | science | 20 | Keep 2 | lore 10/h, 5 workers (1.08) | [R: BLD-017] |
| Academy | science | 15 | Scriptorium 3 | research time −3% | [P] |
| Warehouse | storage | 25 | — | storage +1,200 all capped resources (1.12) | [R: BLD-018] |
| Hidden Vault | storage | 15 | Warehouse 3 | protects 150 of every resource (1.10) | [R: ECO-014] |
| Walls | military | 20 | Keep 3 | +2 defense points to every city defender | [R: BLD-024] (as defence bonus, cf. C-09) |
| Watchtower | military | 15 | Keep 2 | scout strength +8% (currently counts only when defending against scouts, I-05) | [P] |
| Barracks | military | 20 | Keep 2 | train time −2%; trains infantry, scouts, wagons | [R: BLD-022] |
| Archery Range | military | 20 | Barracks 2 | train time −1%; ranged | [R: BLD-022] |
| Stables | military | 20 | Barracks 3, Farmstead 4 | train time −1%; cavalry | [R: BLD-022] |
| Aerie | military | 15 | Stables 3, Beast Lore 1 | train time −1%; winged beasts | [P] |
| Siege Workshop | military | 15 | Barracks 5, Engineering 1 | train time −1%; siege | [P] |
| Sanctum | military | 15 | Scriptorium 5, Arcana 1 | train time −1%; battle mages | [P] |
| Market Hall | economy | 15 | Keep 3 | market fee −4% (relative); carry +3% | [R: BLD-020]; fee effect currently inactive (I-03) |
| War Hall | military | 7 | Keep 3 | +1 hero slot per level (max 7) | [R: BLD-021, HERO-002] |

Sample values (Keep only modifying time): Keep 2 costs 180 timber, 180 stone, 60 gold (167 s); Keep 5 costs 608 / 608 / 203 (449 s); Keep 10 costs 4,613 / 4,613 / 1,538; Keep 25 costs 2,020,093 / 2,020,093 / 673,364.

---

## 7. Sciences and specialisations

Technologies are researched with **lore plus a resource**, paid from the city that queues them; effects apply to the whole player [R: TECH-001] (shape) / [P] (lore per city). Research queue: 2 per player. Time reductions share the 80% cap.

| Tech | Branch | Max | Requires | Effect per level | Basis |
|---|---|---|---|---|---|
| Forestry | economy | 10 | Scriptorium 1 | timber +5% | [P] |
| Crop Rotation | economy | 10 | Scriptorium 1 | grain +5% | [P] |
| Masonry | economy | 10 | Scriptorium 1 | stone +5%, storage +3% | [P] |
| Metallurgy | economy | 10 | Scriptorium 2, Ironworks 1 | iron +5% | [P] |
| Exchequer | civic | 10 | Scriptorium 3 | army upkeep −4%, gold +3% | [R: TECH-008] |
| Civic Husbandry | civic | 10 | Scriptorium 2 | population growth +10%, housing +3% | [P] |
| Drill Tactics | military | 10 | Barracks 2, Scriptorium 2 | attack +3% | [P] |
| Armoursmithing | military | 10 | Ironworks 2, Scriptorium 2 | defense +2%, hit points +2% | [P] |
| Logistics | military | 10 | Scriptorium 2 | march speed +4%, carry +5% | [P] |
| Espionage | military | 10 | Watchtower 1 | scout strength +10% | [R: TECH-012] |
| Beast Lore | military | 5 | Stables 2, Logistics 2 | unlocks Aerie; beast HP +4% | [P] |
| Engineering | military | 5 | Masonry 3, Drill Tactics 2 | unlocks Siege Workshop; siege attack +5% | [P] |
| Arcana | arcane | 5 | Scriptorium 4 | unlocks Sanctum; mage attack +5% | [P] |
| Plunder Doctrine | military | 5 | Drill Tactics 3 | loot +4% | [P] |
| Heroic Sagas | arcane | 5 | War Hall 1 | hero XP +10% | [P] |

### 7.1 Specialisations: the Paths

A player chooses **one Path** (exclusive group `calling`), each with 3 levels, available at Keep 5. Choosing one blocks the others permanently (no respec today, I-13). One specialisation at a time with three levels is [R: TECH-020]; permanence and the specific Paths are [P] (see conflict C-14 about whether the reference ties specialisations to the player or to heroes).

| Path | Effect per level (max L3) | L1 cost (L2, L3 ×1.8 each) | L1 time |
|---|---|---|---|
| Path of the Architect | construction time −15% (−45% at L3) | 400 lore, 400 stone, 200 gold | 30 min |
| Path of the Steward | all production +6% (+18%) | 400 lore, 400 grain, 200 gold | 30 min |
| Path of the Warlord | attack +6% (+18%), training time −8% (−24%) | 400 lore, 400 iron, 200 gold | 30 min |

---

## 8. Population

- **Housing** comes from the Keep and Cottages (× housing%) [R: BLD-019, POP-002] (sources).
- **Growth (AC mode `free_housing`)** [P]: every 10-minute tick, population grows by (housing − population) × 12%/h × (1 + growth%) × tick length; fractions carry over. Growth slows as the city fills.
- The engine also supports mode **`idle`** (growth ∝ idle people), closer to the reference rule of 0.04 × idle per hour [R: POP-002]; Void Frontier uses it at 8%/h. AC does not.
- **Employment** [R: POP-004] (staffing) / [P] (automatic): workers needed = Σ workers per level × level over producing buildings. If people are short, every producer runs at the same fraction (employment = people ÷ workers needed). No manual assignment and no builders.
- **Idle people** (population − employed) pay the Trading Post's gold tax [R: POP-003, ECO-005].
- **Food** [P]: each person eats 0.12 grain per hour. With grain at 0 and a deficit, 1% of the population is lost each tick.
- Population above housing is removed at the next tick [P] (reference uses a gradual flight formula, [R: POP-006], not adopted).
- Units do **not** cost population [P] (reference does, [R: POP-009], not adopted).

Start: 40 people, housing 150 (Keep 1: 60, Cottages 1: 90), 18 jobs (three producers × 6).

---

## 9. Armies and the counter loop

### 9.1 Upkeep

Every unit has a gold wage per hour [R: ECO-015, ARMY-010]. Units away (on missions or stationed in territories or allied cities) cost **double** [R: ECO-015, value from research]. If gold is at 0 and the balance is negative, each garrison stack deserts at **0.5% per hour** [R: ECO-015, value from research]. Exchequer and the Gravewarden race reduce wages.

### 9.2 Units (trainable)

Defense is a damage-reduction percentage (capped at 80% after walls and modifiers). Speed is tiles per hour. Training time is per unit; a batch completes at once.

| Unit | Class | Atk | Def | HP | Spd | Carry | Wage/h | Cost | Time | Trained in / requires | Bonus |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Militia | infantry | 8 | 10 | 40 | 5 | 15 | 0.3 | 30 grain, 20 timber, 5 iron | 45 s | Barracks | +40% vs ranged |
| Pikeman | infantry | 14 | 18 | 60 | 4.5 | 20 | 0.6 | 50 grain, 30 timber, 30 iron | 90 s | Barracks 3, Armoursmithing 1 | +40% vs ranged, +15% vs cavalry |
| Bowman | ranged | 14 | 5 | 35 | 5 | 10 | 0.4 | 30 grain, 45 timber, 5 iron | 60 s | Archery Range | +40% vs beast |
| Crossbowman | ranged | 22 | 8 | 45 | 4.5 | 10 | 0.7 | 40 grain, 50 timber, 35 iron | 100 s | Range 3, Metallurgy 2 | +40% vs beast |
| Outrider | cavalry | 18 | 8 | 70 | 10 | 40 | 1.0 | 70 grain, 30 timber, 40 iron | 120 s | Stables | +40% vs infantry |
| Knight | cavalry | 30 | 20 | 110 | 8 | 30 | 1.8 | 100 grain, 40 timber, 90 iron, 20 gold | 200 s | Stables 4, Armoursmithing 3 | +40% vs infantry |
| Griffin Rider | beast | 34 | 12 | 120 | 14 | 25 | 2.5 | 160 grain, 60 iron, 60 gold | 300 s | Aerie | +40% vs cavalry |
| Ballista | siege | 45 | 4 | 90 | 2.5 | 0 | 2.0 | 200 timber, 100 iron, 40 gold | 360 s | Siege Workshop | +20% vs infantry |
| Battle Mage | magic | 40 | 6 | 60 | 5 | 5 | 3.0 | 80 grain, 120 gold, 40 lore | 360 s | Sanctum | — |
| Scout | scout | 0 | 2 | 15 | 16 | 0 | 0.2 | 20 grain, 10 gold | 40 s | Barracks | scout strength 1 |
| Supply Wagon | transport | 0 | 4 | 60 | 6 | 300 | 0.4 | 80 timber, 20 grain | 80 s | Barracks | — |
| Settlers | settler | 0 | 2 | 50 | 3 | 500 | 1.0 | 800 timber, 800 grain, 600 stone, 400 gold | 1,800 s | Keep 5 | founds a city |
| Thornweaver | ranged (Thornkin) | 20 | 10 | 50 | 6 | 10 | 0.8 | 40 grain, 70 timber, 5 lore | 110 s | Range 2 | +50% vs beast |
| Bannerguard | infantry (Ironvow) | 14 | 28 | 80 | 4 | 20 | 0.9 | 60 grain, 30 stone, 50 iron | 120 s | Barracks 2 | +40% vs ranged |
| Cinderhound | beast (Ashborn) | 26 | 8 | 90 | 12 | 20 | 1.4 | 120 grain, 40 iron | 180 s | Stables 2 | +50% vs cavalry |
| Bone Lord | magic (Gravewarden) | 30 | 15 | 90 | 5 | 10 | 0.4 | 40 grain, 60 iron, 40 lore | 240 s | Barracks 4, Scriptorium 3 | — |

Labels: the four-class pursuit loop is [R: CMB-004]; categories such as carriers, mages and siege-breakers echo [R: ARMY-001, ARMY-004, ARMY-007]; all stats and costs are [P].

### 9.3 Counter loop

```
 infantry ──+40%──► ranged ──+40%──► beast ──+40%──► cavalry ──+40%──► infantry
```

[R: CMB-004] (reference loop Warriors → Archers → Airborne → Cavalry → Warriors). Our bonus multiplies damage against the countered class [P]; the reference uses an extra strike. Siege (+20% vs infantry) and mages (no counter, high attack, high wage) sit outside the loop [P].

### 9.4 Battle resolution

[R: CMB-001, CMB-003, CMB-005, CMB-006] (shape) / [P] (formula):
- Up to **20 rounds** [R: CMB-001, value from research]. Both sides fire simultaneously from start-of-round counts.
- Damage from stack *s* to stack *t* = count_s × attack_s × (count_t ÷ total enemy count) × (1 + bonus vs t's class) × (1 − min(80, defense_t)/100). Partial damage carries over as wounds; damage beyond a destroyed stack is lost.
- Walls add 2 defense points per level to city defenders (before the 80 cap).
- Retreat: the attacker sets a loss threshold per mission (1–100%); defenders never retreat in AC (`defenderRetreatPct` 100). If both sides reach their threshold in the same round, the defender holds [R: CMB-006].
- The attacker wins only if the defender is destroyed or retreats and the attacker did not retreat. A timeout is a defeat for the attacker [R: CMB-001].
- Combat is deterministic: the same armies always give the same result [P].

---

## 10. Heroes

| Rule | AC value | Basis |
|---|---|---|
| Hiring | In a city with a War Hall; cost (heroes owned + 1) × (300 gold + 300 grain); the player chooses class and name | [R: HERO-001] (building-gated hiring); details [P] |
| Slots | War Hall level, max 7 | [R: HERO-002, value from research] |
| Classes | Marshal (attack +5%, +1% per level), Warden (defense +3%, HP +3%; +1% HP per level), Pathfinder (speed +8%, loot +5%, scout +10%; +1% speed per level), Sage (hero XP +15%, mage attack +8%; +1% mage attack per level) | [R: HERO-003] (classes exist) / [P] |
| Skills (max 5 each) | Spear Wall (infantry def +3%), Volley (ranged atk +4%), Charge (cavalry atk +4%), Beastmaster (beast atk +4%), Forced March (speed +3%), Quartermaster (carry +5%), Field Medic (HP +3%; Warden, Sage), Pyromancy (mage atk +5%; Sage) | [P] |
| Skill points | 1 + 2 × (level − 1) | [R: HERO-005] ("2 per level") / [P] (start value) |
| Equipment | 4 slots: weapon, armor, helm, trinket; items of rarity common → legendary drop from neutrals (e.g. Iron Blade, Ember Sword, Skyfall Lance, Starglass Mail, Sage's Tome) | [R: HERO-009] (artefacts from ruins) / [P] |
| Experience | floor(enemy score destroyed × 0.5 × (1 + hero XP%)), plus the neutral's level XP on a win | [R: HERO-008] (XP from enemy value) / [P] |
| Levels | Max 30; XP to reach level L = Σ round(200 × 1.35^(k−1)) for k < L (L2: 200, L5: 1,327, L10: 7,939, L20: 170,551, L30: 3,440,086) | [P] |
| Role | A hero joins one mission and adds its modifiers to that army; an idle, unwounded hero at home defends its city (highest level first) | [P] |
| Defeat | Wounded for 6 hours (attacking hero on a lost battle; defending hero when the attacker wins) | [R: CMB-011] (defeat penalty) / [P] (wound instead of stun or capture) |

Not implemented: hero capture and ransom [R: HERO-007], spells and mana [R: HERO-006], economic hero specialties [R: HERO-004], artefact sets [R: HERO-009].

---

## 11. Map and territories

- **World:** 48 × 48 tiles, generated from the world seed with terrain weights Plains 30, Forest 24, Hills 14, Iron Ridge 7, Mountains 7, Ashland 6, Marsh 6, Lake 6 [P]. Flat grid with Euclidean distance (reference uses a kingdom/square hierarchy [R: WLD-001], not adopted).
- **Neutrals** occupy 12% of passable tiles. Their level grows with distance from the centre, so the centre (where new players spawn) is safer [P].
- **Spawn:** a free passable tile near the centre, at least 3 tiles (Chebyshev) from any city [P].
- **Travel:** time = max(900 s, distance ÷ (slowest unit speed × (1 + speed%)) × 3,600 s), both divided by world speed. The 15-minute minimum mirrors [R: MSN-003, value from research].
- **Territories** [R: ECO-011, WLD-004] (contested outer nodes) / [P] (rules):
  - Send troops on an **occupy** mission to a free tile with no active neutral. Troops stay stationed (paying double wages).
  - Limit: 1 territory per Keep level.
  - Yield: 25 units/hour of the terrain's resource (Plains grain, Forest timber, Hills stone, Iron Ridge iron, Ashland gold; Marsh none), accruing for at most 24 hours until **collected** into a city.
  - Other players can attack the stationed garrison and, with *capture* enabled, take the tile if they win and are under their own limit.
  - Abandoning a territory sends own troops home.
- **Cities:** found with Settlers on a free tile. City limit = 1 + Keep level ÷ 5 (Keep 5: 2 cities … Keep 25: 6 cities); `maxCities` 9 is not reachable with Keep max 25 (I-01). Cities can never be captured [R: ECO-011] ("Cities cannot be seized, as opposed to Outer Domains", SRC-A1-14). A new city starts with level 1 of each starting building, half the starting population and the resources carried.

---

## 12. PvE

| Neutral | Kind | Levels 1 → 5 guards | Rewards L1 → L5 | Drops | Respawn | Basis |
|---|---|---|---|---|---|---|
| Wolf Den | lair | 8 → 180 wolves | 150 grain + 100 timber → 3,000 grain + 2,000 timber + 300 gold | Leather Jerkin, Lucky Coin, Scout's Hood, Wind Charm | 4 h | [P] |
| Brigand Camp | lair | 10 brigands → 200 brigands + 10 trolls | 120 gold + 80 iron → 2,600 gold + 1,700 iron + 1,200 stone | Iron Blade, Lucky Coin, Horned Helm, Ember Sword | 6 h | [P] |
| Ancient Barrow | ruin | 6 wraiths → 140 wraiths + 10 golems | 120 lore + 100 gold + **2 starglass** → 2,800 lore + 2,400 gold + **25 starglass** | Sage's Tome, Warden's Plate, Skyfall Lance, Starglass Mail | 12 h | [R: WLD-005, ECO-023] (monster ruins yield premium currency and items) |
| Starfall Crater | node | 2 golems → 40 golems + 3 drakes | **5 starglass** + 300 stone → **50 starglass** + 3,500 stone | Starglass Mail | 24 h | [R: WLD-006] (premium hot-spots) / [P] |

Rules [P]: attack a neutral with an *attack* mission; guards have no modifiers; same-race/faction penalties do not apply; rewards and drops ride home with the survivors regardless of carry capacity; anything above storage then decays at home. A cleared neutral is inactive until it respawns. Neutrals never attack players (monster raids [R: WLD-015] not implemented).

---

## 13. PvP and protections

| Rule | AC value | Basis |
|---|---|---|
| Newbie shield | 72 h from joining; blocks being attacked or scouted; ends early when the player attacks another player | [R: SOC-003] (newbie protection exists) / [P] |
| Friendly fire | No attacks on own alliance or alliances you have declared *ally* or *nap* | [R: SOC-013] (shape) / [P] |
| Same race / same faction | Attacker's attack −50% vs same race, −25% vs same faction | [R: FAC-002, values from research] |
| Attack cap | Max 2 attacks on the same target tile per attacker per 24 h | [R: MSN-004, value from research] |
| Weak-target floor | Engine supports a score floor; **off** in AC | [R: SOC-001] (rank rule) |
| Loot | Up to 50% of unprotected stock of lootable resources (× loot%), limited by carry capacity; proportional across resources | [R: MSN-006] (50% cap) |
| Vault | Hidden Vault amounts cannot be looted | [R: ECO-014] |
| City capture | Impossible | [R: ECO-011] |
| Territory capture | Win an attack with *capture* on | [R: MSN-009] (shape) |
| Joint defence | Up to 3 other players may station troops in a city | [R: SOC-015, value from research] |
| Scouting | Success if your scout strength > defender's; double strength reveals buildings and population; detection chance 30% + 70% × defender ÷ (attacker + defender); failure loses the scouts | [R: MSN-010] (shape) / [P] |
| Incoming warning | Defender sees an incoming attack/scout and its arrival time, not its composition | [P] |
| Recall | Any outbound mission can be recalled; stationed troops can be withdrawn | [R: MSN-005] (shape) |

Not implemented: renegades and faction loyalty [R: FAC-003, FAC-004], revenge rights [R: SOC-002], joint attack [R: SOC-014], gates [R: WLD-010], building destruction and slave capture [R: MSN-007, MSN-008].

---

## 14. Economy and trade

- **Production formula** [P]: gross = (Σ building output × employment + idle tax × idle) × (1 + Σ production%) + flat. Percentages add up before multiplying (reference stacking is UNKNOWN, [R: ECO-010]).
- **Player market** [R: ECO-017, SOC-020] (shape) / [P]: place an order selling one tradable resource for another (gold included); the offer is escrowed; another player accepts it whole; the seller receives the payment minus a 5% fee. Max 10 open orders. Price must be within 10× of the reference value. No building or merchant needed.
- **NPC exchange** [R: ECO-018] (worse rates): reference values timber 1, grain 1, stone 1.2, iron 1.6, gold 3; received = floor(amount × value_from ÷ value_to × 0.7). A deliberate sink.
- **Transport** [R: MSN-011]: wagons (carry 300) or any units carry cargo to a city; arrival grants it (premium excluded).
- **Sinks overview:** wages (gold), food (grain), construction and research (all), exchange fee (30%), market fee (5%), overflow decay (1%/h), cancel loss (20%), raid losses.

---

## 15. Social

[R: SOC-010, SOC-013, SOC-030] (shapes) / [P] (values):
- **Alliances:** found for 500 gold (paid from the capital); name, 2–5 character tag, description; up to 100 members; roles leader, officer, member. Leaders and officers invite; the leader removes anyone, officers only plain members; the leader sets roles and diplomacy. A leader must hand over leadership before leaving; the last member leaving disbands it.
- **Diplomacy:** per alliance, towards another alliance: *ally* (no attacks, may reinforce), *nap* (no attacks), *war*, *neutral*. Relations are one-directional declarations.
- **Mail:** private messages by player name, alliance broadcast, inbox/sent/alliance boxes, block a player, report a message to moderators; max 60 messages per hour.
- **Reports and notifications:** battle and scout reports for each participant; notifications for jobs, missions, invites, trades.
- **Rankings:** by score = Σ building score × level + Σ tech score × level + Σ home unit score; alliance score = sum of members.
- Not implemented: chat, contacts, mentors, clan warehouse, clan castle [R: SOC-011, SOC-012, SOC-016].

---

## 16. Progression

| Axis | Range in AC | Gate |
|---|---|---|
| Keep | 1 → 25 | resources, time |
| Other buildings | up to 7–25 | Keep and building requirements |
| Cities | 1 → 6 | Keep level ÷ 5, Settlers |
| Territories | 0 → 25 | Keep level |
| Technologies | 15 techs to level 5–10, plus one Path to level 3 | Scriptorium, buildings, other techs |
| Heroes | 0 → 7, each to level 30 | War Hall level, XP |
| Army | unit tiers unlocked by buildings and techs | §9 |
| Standing | score ranking | — |

Design note [P]: costs grow 1.38–1.6× per level while outputs grow about 1.06–1.12× per level, so every level is more expensive per unit of benefit; growth naturally shifts from vertical (levels) to horizontal (more cities and territories) in the mid-game.

---

## 17. Player journey

Times below are computed from pack formulas at world speed 1, assuming resources are on hand.

### 17.1 First session — first 10 minutes [P]

| Time | Player does | Game rule |
|---|---|---|
| 0:00 | Guest or account login; choose a world, a name and one of four races | Shield 72 h starts; capital near the map centre; start stock 600 timber, 600 grain, 500 stone, 250 iron, 400 gold, 20 starglass; 10 Militia, 2 Scouts |
| 0:00 | Queue Lumberyard 2 (67 s), Farmstead 2 (67 s), Cottages 2 (82 s) | Quests t01–t03 [R: QST-001, QST-004] |
| ~0:01–0:04 | Claim each reward as the building completes (+150 timber, +100 stone; +200 grain; +150 timber, +150 grain) | Manual claim [R: QST-001] |
| ~0:04 | Queue Keep 2 (180 timber, 180 stone, 60 gold; 167 s) | t04 |
| ~0:06 | Queue Scriptorium (116 s) and Barracks (116 s) | t05, t07 |
| ~0:08 | Claim 100 lore; research Forestry (60 lore, 60 timber; 180 s) | t06; research queue runs in parallel |
| ~0:10 | Barracks done; train 10 Militia (45 s each, 7.5 min) | t08 |

The six tutorial buildings take 615 s of build time in total.

### 17.2 First hour [P]

- Send scouts to a nearby lair (t09). Every mission takes at least 15 minutes each way.
- Attack a low-level Wolf Den or Brigand Camp near the centre with the militia (t10: +250 gold, +5 starglass on claim). The clear counts when the army returns, so t10 takes at least 30 minutes.
- Unlock side quests: market trade (t11), join or found an alliance (t12).

### 17.3 First day [P]

- Raise the Keep to 5 (m01: +1,000 stone, +500 gold, +10 starglass). Keep 3 opens Walls, Market Hall and War Hall; Keep 4 opens the Mint.
- Hire the first hero (t13) and occupy the first territory (t14).
- At Keep 5: choose a Path; train Settlers.
- Economy watch-points: gold has no producer until the Trading Post (Keep 2) or Mint (Keep 4); the starting army costs 3.4 gold/h against 400 gold, so the first day is safe, but every new soldier raises the wage bill [R: ECO-015].

### 17.4 Mid term (days 2–14) [P]

- Found the second city (m03: rule two cities). Specialise cities (one military, one economic) and move goods with wagons.
- Win 10 battles (m02). Grow heroes, collect items, hold several territories for daily yield.
- The shield ends at 72 h (earlier if the player attacks someone); alliances and diplomacy become the main protection.

### 17.5 Long term (weeks +) [P]

- Keep 15–25, up to 6 cities, up to 25 territories, 7 heroes towards level 30, techs at max.
- Alliance wars over territories and craters; ranking competition.
- No world end condition exists; seasonal worlds are an open question (PRD §12).
- Balance risk: from Keep 21, a Keep level costs more timber than a city can store even with Warehouse at the same level and Masonry 10 (Keep 21 costs ~399k timber vs ~318k capacity); the overflow rule lets stock exceed the cap but it decays at 1%/h. To be reviewed in playtests.

---

## 18. Tutorial and quests

Quests are optional, chained and claimed manually [R: QST-001, QST-004]; rewards go to the chosen city, else the capital [R: QST-002] (no battle-proximity guard). The client shows a tutorial bar with the shield timer and the number of claimable quests [P].

| Quest | Goal | Requires | Reward | Tutorial |
|---|---|---|---|---|
| t01 First Timber | Lumberyard 2 | — | 150 timber, 100 stone | yes |
| t02 Bread for All | Farmstead 2 | t01 | 200 grain | yes |
| t03 Room to Grow | Cottages 2 | t02 | 150 timber, 150 grain | yes |
| t04 Raise the Keep | Keep 2 | t03 | 250 stone, 100 gold | yes |
| t05 House of Scholars | Scriptorium 1 | t04 | 100 lore | yes |
| t06 First Insight | Forestry 1 | t05 | 150 lore, 200 timber | yes |
| t07 Call to Arms | Barracks 1 | t04 | 150 iron, 200 grain | yes |
| t08 Muster the Levy | Train 10 units | t07 | 300 grain, 150 gold | yes |
| t09 Eyes in the Wild | Send 1 scouting mission | t08 | 100 gold | yes |
| t10 Clear the Wilds | Clear 1 neutral | t09 | 250 gold, 5 starglass | yes |
| t11 Open for Trade | Accept 1 market order | t08 | 200 gold | no |
| t12 Banners Together | Join or found an alliance | t08 | 300 gold, 5 starglass | no |
| t13 A Champion Rises | Hire a hero | t10 | 300 gold | no |
| t14 Stake a Claim | Hold 1 territory | t10 | 500 stone, 500 timber | no |
| m01 Lord of the Keep | Keep 5 | t04 | 1,000 stone, 500 gold, 10 starglass | no |
| m02 Seasoned Commander | Win 10 battles | t10 | 1,000 gold, 10 starglass | no |
| m03 A Second Seat | Rule 2 cities | m01 | 2,000 timber, 2,000 stone, 2,000 grain, 20 starglass | no |

Onboarding order (economy → science → army → map) echoes the community early-game flow [R: QST-005, TECH-021]; the content is [P].

---

## 19. Events (not implemented)

No live events exist in the engine or pack. Future design, all [P] unless marked:
- **Starfall nights:** a timed world event that spawns extra Starfall Craters near the centre for 24 h. Inspired by random world events [R: SOC-042].
- **Seasonal ladders:** time-boxed rankings (most neutrals cleared, most territory-hours) with cosmetic rewards. Inspired by tournaments [R: SOC-041].
- Requirements: a world-event scheduler, time-boxed leaderboards, and reward delivery through the quest/claim pipeline.

---

## 20. Balance principles

| # | Principle | Lever in AC | Basis |
|---|---|---|---|
| B1 | Armies are paid continuously; size is capped by gold income | wages, ×2 away, 0.5%/h desertion | [R: ECO-015] |
| B2 | Soft caps instead of walls | 1%/h overflow decay | [R: ECO-013] |
| B3 | No single best army | four-class loop at +40%; siege and mages off-loop with high wages | [R: CMB-004] / [P] |
| B4 | Defence is easier than attack | walls, defenders never retreat, ties go to the defender, 20-round limit | [R: CMB-006, CMB-001] / [P] |
| B5 | Raids hurt but do not erase | 50% loot cap, vault, 2 attacks per target per day | [R: MSN-006, ECO-014, MSN-004] |
| B6 | War points outward | −50% same race, −25% same faction | [R: FAC-002] |
| B7 | Speed-ups never reach zero | 80% cap on all time reductions | [R: BLD-000] |
| B8 | Race bonuses are flavour, not destiny | single bonuses ≤ 15%, upkeep −20% the largest | [P] |
| B9 | Diminishing returns per level | cost growth 1.38–1.6 vs output growth 1.06–1.12 | [P] |
| B10 | Premium never buys power | starglass not tradable/lootable; no power sink | [P] |
| B11 | Everything is pack data | rebalance without code changes; live worlds migrated | [P] |

Known risks: defence bias (I-12), late Keep costs vs storage (§17.5), early gold dependence on quests, starglass without sink (I-09).

---

## 21. Monetisation

- **MVP: none.** No real-money sales, no ads, no subscriptions. [P]
- **Starglass** is the premium currency of the setting. It is **earned only in-game**: 20 at start, 50 from quests (t10 5, t12 5, m01 10, m02 10, m03 20), 2–25 per Ancient Barrow clear and 5–50 per Starfall Crater clear. It cannot be traded, looted, transported or bought. [R: ECO-006] (premium, not stealable) / [P] (earned only).
- No auction of premium currency [R: ECO-019, SOC-021] (not adopted) and no paid boosts or instant completion [R: ECO-021] (not adopted).
- **Spending (planned, not implemented)** [P], constraints first: no combat stats, no production boosts, no time skips beyond what non-spenders can get. Candidates: cosmetic city banners and name colours; alliance emblems; renaming a hero; a starglass-forged legendary *appearance* for items. Any future real-money offer must be designed so that it cannot buy starglass that converts into power.
- Platform payment rules (Yandex SDK, Google Play Billing) apply only if real money is ever added (`PLATFORM_REQUIREMENTS.md`).

---

## 22. Open design questions

1. Respec for Paths (I-13) — keep permanent, or allow a costly switch as the reference does [R: TECH-020]?
2. Should defenders have a retreat option (per city) [R: CMB-005]?
3. Which starglass sinks keep the no-pay-to-win promise (§21)?
4. Should the city cap reach 1 + 8 [R: WLD-003] (I-01)?
5. Should population use `idle` growth mode, closer to [R: POP-002]?
6. World end / seasons?
