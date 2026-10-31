# Buildings, Technologies, Population, Quests: research notes (A1)

Access date for all evidence: 2026-10-31. Source IDs refer to `_sources_A1.md`.
**Evidence caveat:** every page was read only through search-engine extracts, because direct fetching was blocked in this session. **No per-level cost or time tables were captured.** All such numbers are UNKNOWN unless stated.

## Building catalogue (names as listed on the encyclopedia Buildings page, SRC-A1-09)

| Category (wiki) | Buildings found | Notes |
|---|---|---|
| Initial / special | Main Building, City Gate | Present in every city from the start |
| Producing (need Workers or Alchemists) | Woodcutter Cabin, Windmill, Stone Quarry, Iron Plant, Gold Plant, Trading Post, Alchemist Lab, Chalice of Abundance | Most sit on matching deposit cells |
| General / municipal | Warehouse, Cottage, Market, Terraformer, Tavern (Capital only) | |
| Military (combat units) | Slave Dwelling, Barracks, Stable, Pegasus Bastion, Shooting Range, Abbey, Mercenary Order, School of Magic | Names from one snippet. May be race-specific variants. |
| Special (non-combat units) | Merchant Guild, Spy Guild, Counterspy Guild, Expeditionary Corps, Colonisation Centre, Visionary Tower, Mirror Tower | |
| Defensive | Fortification, Tower, Magic Tower | |
| Outer-domain | Main Building of Mine, Pit, Salt Lake, Grail | Mines allow no construction |
| Clan (A2 scope) | Clan Castle, Clan Warehouse | Listed for completeness |

---

## A. Buildings

### BLD-000 — General construction rules
- Description: Buildings occupy cells. They can be constructed, upgraded, deleted or repaired.
- Game purpose: Core progression loop.
- Inputs: Resources, a free cell, Builders, prerequisite science.
- Rules & constraints: Fewer Builders make construction, upgrade and repair take longer. Bonuses can shorten time by **at most 80%**. Debuffs have no cap. If a building's cell becomes unavailable (for example after a Capital transfer), the building is deleted automatically. The same building type looks different per race.
- Output: Building level changes.
- Related systems: POP-005, TECH-020 (Architect), ECO-016.
- Evidence: SRC-A1-09: "the less a construction site is filled with Builders the longer the building construction/upgrade/repair will last"; time reduction "no more than 80%"; "if the cell a building was standing on become unavailable… the building is automatically deleted".
- Status: VERIFIED (high)
- Test plan: Build time with full Builders = T. Half Builders gives T' > T (curve TBD). Stacked bonuses of 95% are clamped to 80%.

### BLD-001 — Main Building (city core)
- Description: Initial building in every city. It produces wood. Its level drives the city's cell count, storage, living places and growth.
- Game purpose: Central progression gate.
- Inputs: Resources for upgrades (UNKNOWN).
- Rules & constraints: Starts at level 1, max level **5** (RU wiki). Upgrades increase cells inside the City Wall up to the square's limit. It contributes to the Warehouse limit, living places and population growth. Max level is a prerequisite for villages. The Main Building window shows living places, Cache capacity and the storage allocation chart.
- Output: Wood, storage, housing, growth, cells.
- Related systems: BLD-003, BLD-004, ECO-012, POP-002.
- Evidence: SRC-A1-24/42 (RU Alchemy / central building): "Initially, every city has a level 1 central building. The maximum is level 5" (translated). SRC-A1-14: cells grow "as the Main Building upgrades". SRC-A1-04, SRC-A1-08.
- Status: VERIFIED (med; max level from RU only)
- Test plan: Upgrading Main Building 1→2 increases the cell count, storage and housing. Level 6 is rejected.

### BLD-002 — City Gate
- Description: Initial defensive building. Opening it unlocks outward actions.
- Game purpose: A deliberate "leave the safe phase" step.
- Inputs: "Gate" science.
- Rules & constraints: After the Gate is opened, the player can launch combat missions, found new cities and develop villages. Snippet: "opens 5 minutes and closes 6 hours" (meaning ambiguous: opening and closing transition times?).
- Output: Gate state.
- Related systems: TECH-003, BLD-004, QST-005.
- Evidence: SRC-A1-28: "after the Gate is open you will be able to launch combat missions, found new cities and develop villages". SRC-A1-11.
- Status: VERIFIED (med); timing UNKNOWN
- Test plan: With the Gate closed, combat and colonisation missions are rejected.

### BLD-003 — City cells, Capital and City Expansion
- Description: A city is a grid of cells.
- Game purpose: Spatial constraint on building count.
- Inputs: Main Building level, the founding square's level, City Expansion charges.
- Rules & constraints: The Capital always has **51 cells**. Certain buildings are Capital-only. A new city starts with **22 cells** and grows with Main Building level, capped by the square's max (a level-5 city has 39–43 cells; RU says a level-4 city has 34–38). The **City Expansion** science adds 1 cell per charge, up to **43** cells per city, and costs **25,000 SP** per charge. City level can reach 5. Cells may hold obstacles (lake, trees, rocks) that need a Terraformer.
- Output: Available cells.
- Related systems: BLD-001, BLD-006, TECH-002.
- Evidence: SRC-A1-09/14: "The Capital, unlike other Cities, always has 51 cells"; "any new city has 22 cells at the start". SRC-A1-11: "the limit is 43 cells in a city"; "25,000 of Science Points".
- Status: VERIFIED (high)
- Test plan: A new city has 22 cells. An expansion charge adds 1. The 22nd expansion past 43 is rejected.

### BLD-004 — Villages (4 per city)
- Description: Four village areas around each city, with extra build cells.
- Game purpose: Mid-game economic expansion and neighbour contention.
- Inputs: Main Building at max level, the Gate opened, "Villages" science.
- Rules & constraints: Producing buildings in villages get **+30%**. Neighbours can also build in shared villages; their buildings are shown in red but do not block free cells. After the Villages science, village cells including deposits become visible.
- Output: Extra producing cells.
- Related systems: ECO-010, TECH-004.
- Evidence: SRC-A1-13: "there are 4 villages in every City"; "Producing Buildings in villages work 30% faster". SRC-A1-36 (Wikipedia) also says 4 villages.
- Status: VERIFIED (high)
- Test plan: Village build orders are rejected until all three prerequisites hold. Village producers output 1.3× the in-wall equivalent.

### BLD-005 — Deposit-cell restriction ("land-use" science)
- Description: Resource cells accept only the matching producing building, until a land-management science is studied.
- Game purpose: Map-driven layout planning.
- Inputs: Cell type.
- Rules & constraints: RU fan source: on resource cells you can build only the matching producing buildings. After the "Землеустройство" (land-use) science, any building can go there.
- Output: Placement validation.
- Related systems: ECO-009.
- Evidence: SRC-A1-42 (RU fan, 2010–2012).
- Status: VERIFIED by a fan source only (low)
- Test plan: Placing a Cottage on a stone deposit is rejected before the science and allowed after it.

### BLD-006 — Terraformer
- Description: Temporary building that clears obstacle cells.
- Game purpose: Unlocks extra cells at a time cost.
- Inputs: An obstacle cell (lake, trees, rocks).
- Rules & constraints: Self-destroys when finished, freeing the cell. Terraformation time: level 1 takes 4 days, level 2 takes 1 day.
- Output: Free cell.
- Related systems: BLD-003.
- Evidence: SRC-A1-23: "After the terraformation have been completed the Terraformer self-destroys"; time table L1 = 4 days, L2 = 1 day.
- Status: VERIFIED (med)
- Test plan: After 4 days at L1 the cell is free and the building is gone.

### BLD-010 — Woodcutter Cabin
- Description: Wood producer.
- Game purpose: Wood income.
- Inputs: Workers.
- Rules & constraints: Elves get the best base rate. Ancient Magic gives +25%. Per-level output UNKNOWN.
- Output: Wood/h.
- Related systems: ECO-001.
- Evidence: SRC-A1-04, SRC-A1-07.
- Status: VERIFIED (high for its existence and function)
- Test plan: See ECO-001.

### BLD-011 — Windmill
- Description: The only crop producer, on a crop deposit.
- Inputs: Workers. Rules: finite deposit. The Windmill's own window sells the "+25% Windmills" BG service.
- Output: Crops/h. Related systems: ECO-002.
- Game purpose: Crop income.
- Evidence: SRC-A1-04 ("Crops are produced in: Windmill"), SRC-A1-07.
- Status: VERIFIED (high)
- Test plan: See ECO-002.

### BLD-012 — Stone Quarry
- Description: Stone producer on a stone deposit. Game purpose: stone income. Inputs: Workers. Rules: finite deposit. Demons are best. Output: Stone/h. Related systems: ECO-003, ECO-009.
- Evidence: SRC-A1-04.
- Status: VERIFIED (high)
- Test plan: See ECO-003 and ECO-009.

### BLD-013 — Iron Plant
- Description: Iron producer on an iron deposit. Game purpose: iron income. Inputs: Workers. Rules: finite deposit. Output: Iron/h. Related systems: ECO-004.
- Evidence: SRC-A1-04.
- Status: VERIFIED (high)
- Test plan: See ECO-004.

### BLD-014 — Gold Plant
- Description: Gold producer on a gold deposit. Game purpose: gold income. Inputs: Workers. Rules: finite deposit. Knights are best. Output: Gold/h. Related systems: ECO-005.
- Evidence: SRC-A1-04, SRC-A1-09 (grid lists "Gold Plant").
- Status: VERIFIED (high)
- Test plan: See ECO-005.

### BLD-015 — Trading Post
- Description: Gold producer driven by idle population ("tax").
- Game purpose: Turns spare population into gold. Gives a reason to keep idle people.
- Inputs: Idle population. Trader spec raises the tax.
- Rules & constraints: Formula UNKNOWN (see ECO-005).
- Output: Gold/h.
- Related systems: POP-003.
- Evidence: SRC-A1-08, SRC-A1-12.
- Status: VERIFIED (med); formula INFERRED
- Test plan: Income is monotonic in idle count.

### BLD-016 — Chalice of Abundance
- Description: Recovery building for a depleted deposit.
- Game purpose: Sustainability loop.
- Inputs: A depleted deposit cell after demolishing the producer.
- Rules & constraints: Can be built only on depleted deposits. Recovers at "certain rate" (UNKNOWN). Self-destroys at full recovery.
- Output: Restored deposit.
- Related systems: ECO-009.
- Evidence: SRC-A1-22.
- Status: VERIFIED (high)
- Test plan: See ECO-009.

### BLD-017 — Alchemist Lab
- Description: Produces Science Points. Staffed by Alchemists.
- Game purpose: Research engine.
- Inputs: Alchemists.
- Rules & constraints: At least one is required to study sciences. Village and Ancient Magic bonuses apply per lab. Forum: labs are cheap, so whole villages are often filled with them.
- Output: SP (global pool).
- Related systems: ECO-007, TECH-001.
- Evidence: SRC-A1-10, SRC-A1-09 ("filled with Workers or Alchemists"), SRC-A1-30 (forum).
- Status: VERIFIED (high)
- Test plan: With 0 labs, research is blocked.

### BLD-018 — Warehouse
- Description: Raises the domain storage limit.
- Inputs: Resources. Rules: limit = sum over Warehouses (count × levels) + Main Building. Knights get the highest capacity. Values UNKNOWN.
- Output: Capacity. Related systems: ECO-012.
- Game purpose: Storage growth.
- Evidence: SRC-A1-04.
- Status: VERIFIED (high for the rule)
- Test plan: Adding a Warehouse increases the cap by its level value.

### BLD-019 — Cottage
- Description: Housing. Adds living places and population growth.
- Game purpose: Population capacity and growth.
- Inputs: Resources.
- Rules & constraints: Growth is limited to the capacity of all Cottages plus the Main Building. BG service: +10 or +25 growth/h per Cottage. Benefactor spec: +8/12/16 growth/h per Cottage. Base values UNKNOWN.
- Output: Living places, growth/h.
- Related systems: POP-002.
- Evidence: SRC-A1-08: "Population Growth is limited to all Cottages and Main Building's capacity". SRC-A1-07, SRC-A1-12.
- Status: VERIFIED (high)
- Test plan: Growth per hour equals base + service + spec, clamped at capacity.

### BLD-020 — Market
- Description: Trading building (resources ↔ Gold) with an Exchange tab and the BG Auction tab.
- Inputs: Market science. Rules: see ECO-017..019. Output: trades. Related systems: TECH-006.
- Game purpose: Trade.
- Evidence: SRC-A1-16, SRC-A1-11.
- Status: VERIFIED (high)
- Test plan: See ECO-017.

### BLD-021 — Tavern (Capital only)
- Description: Hero hiring building.
- Game purpose: Hero acquisition (A2 scope).
- Inputs: Heroes Managing science, Captain rank, gold.
- Rules & constraints: Only one, Capital only. Hero offers refresh daily.
- Output: Heroes.
- Related systems: TECH-011.
- Evidence: SRC-A1-25 (RU Tavern; EN Heroes: "study the 'Heroes Managing' science, construct the Tavern building, and be of Captain rank").
- Status: VERIFIED (med)
- Test plan: A second Tavern is rejected. A Tavern outside the Capital is rejected.

### BLD-022 — Military training buildings
- Description: Train combat units for combat missions.
- Game purpose: Army production (A2 / combat scope).
- Inputs: Resources, population, sciences.
- Rules & constraints: Names seen: Slave Dwelling, Barracks, Stable, Pegasus Bastion, Shooting Range, Abbey, Mercenary Order, School of Magic. These are probably race variants of a smaller set of slots (UNKNOWN). Fan advice: build many level-1 buildings early, since level 1 is cheaper and faster. Note that "Barracks" is also the name of an underground domain type (SRC-A1-15).
- Output: Units.
- Related systems: POP-009.
- Evidence: SRC-A1-09 snippet listing; SRC-A1-42.
- Status: VERIFIED (med for names; mapping UNKNOWN)
- Test plan: Unit training is allowed only in the building of the matching type.

### BLD-023 — Special-unit buildings
- Description: Train non-combat units.
- Game purpose: Logistics, intelligence, expansion.
- Inputs: Their sciences.
- Rules & constraints:
  - Merchant Guild trains Merchants (Merchantry science).
  - Spy Guild and Counterspy Guild (Espionage).
  - Expeditionary Corps trains Explorers. Exploration mission death chance is 2–10% depending on speed.
  - Colonisation Centre (Colonisation science) supports Settlers and new cities.
  - Visionary Tower trains Visionaries (Vision science).
  - Mirror Tower trains Mirrors (Reflection science). Each Mirror kills one Visionary with 30% probability.
- Output: Special units.
- Related systems: ECO-020, TECH-005/012.
- Evidence: SRC-A1-09, SRC-A1-11, SRC-A1-27, SRC-A1-23 (Vision/Mirror snippet).
- Status: VERIFIED (med)
- Test plan: A mission requires the matching unit type. Explorer losses fall in the 2–10% band by speed.

### BLD-024 — Defensive buildings
- Description: Fortification, Tower, Magic Tower.
- Game purpose: Defence of cities and domains.
- Inputs: Fortification, Towers and Magic Towers sciences.
- Rules & constraints: Fortifications raise the domain's defence level. Towers deal damage. Magic Towers deal no damage but add attack to defending troops (RU). Forum: a domain can be seized only when no defensive buildings remain. Forum tower guidance: 2 towers at level 2/3/4/5/7 for cities of 1–5 stars.
- Output: Defence.
- Related systems: TECH-010, combat (A2).
- Evidence: SRC-A1-09 ("Fortification · Tower · Magic Tower"), SRC-A1-23 (Magic Tower RU), SRC-A1-44, SRC-A1-30.
- Status: VERIFIED (med); forum numbers low
- Test plan: A seizure attempt fails while any defensive building stands (if adopted).

### BLD-025 — Outer-domain Main Buildings
- Description: Each outer domain has a Main Building that produces its resource (mine resource, BG for Salt Lake and Grail). Pits have a few free cells with +100%.
- Game purpose: Resource nodes.
- Inputs: Domain ownership.
- Rules & constraints: See ECO-011. Life span information is shown in the Main Building.
- Output: Resources or BG.
- Related systems: ECO-011.
- Evidence: SRC-A1-15.
- Status: VERIFIED (high)
- Test plan: See ECO-011.

### BLD-026 — Improved / "perfect" / monumental buildings (Alchemy)
- Description: Alchemy science levels allow improving buildings with magic crystals. Level 6 unlocks "monumental buildings" and masterpiece crystals. RU fan wiki: most improved buildings need BG to upgrade.
- Game purpose: Late-game sink.
- Inputs: Crystals, BG.
- Rules & constraints: Details UNKNOWN.
- Output: Enhanced buildings.
- Related systems: TECH-016.
- Evidence: SRC-A1-24 (IT/RU Alchemy), SRC-A1-42 (my-lands.fandom.com/ru "Совершенные здания", seen in search results).
- Status: VERIFIED (low)
- Test plan: TBD if adopted.

---

## B. Sciences and specialisations

### TECH-001 — Science system
- Description: Sciences are studied with Science Points from Alchemist Labs, chosen in the Science Window.
- Game purpose: Tech tree gating buildings, units and features.
- Inputs: SP. Some sciences need a resource payment on completion.
- Rules & constraints: At least 1 lab required. SP pooled. Surplus SP lost. When a science needs confirmation, its button glows green; the player pays the listed resources from a chosen domain. Sciences have levels; higher levels increase effects (for example Market range, Heroes count). Prerequisite graph UNKNOWN.
- Output: Unlocks and bonuses.
- Related systems: ECO-007, BLD-017.
- Evidence: SRC-A1-10: "If any science requires your confirmation on completion of the process, the button will shine green".
- Status: VERIFIED (high)
- Test plan: A science completes only after SP ≥ cost and (if required) confirmation plus payment.

### TECH-002 — "Charged" sciences (repeatable consumables)
- Description: Some sciences are studied to a "charged" state, then spent on a target.
- Game purpose: Repeatable SP sink with player choice.
- Inputs: SP.
- Rules & constraints: City Expansion costs 25,000 SP per charge and gives +1 cell to a chosen city (max 43). Capital Transfer charges, then the player selects the new Capital; buildings on cells the new Capital no longer owns are removed.
- Output: Applied effect.
- Related systems: BLD-003.
- Evidence: SRC-A1-11/10: "requires 25,000 of Science Points to get studied and 'charged'". SRC-A1-14 (Capital Transfer).
- Status: VERIFIED (high)
- Test plan: A charge is consumed on use. A second use needs a recharge.

### TECH-003..019 — Named sciences (catalogue)

| ID | Science | Effect found | Evidence | Status |
|---|---|---|---|---|
| TECH-003 | Gate | Allows opening the City Gate | SRC-A1-11 | VERIFIED (high) |
| TECH-004 | Villages | Village cells visible and buildable | SRC-A1-11, -13 | VERIFIED (high) |
| TECH-005 | Colonisation | Colonisation Centre, own several cities (up to 7–8 more) | SRC-A1-11, -14 | VERIFIED (high) |
| TECH-006 | Market | Market building. Trade range by level. | SRC-A1-11, -16 | VERIFIED (high) |
| TECH-007 | Merchantry | Merchant Guild and Merchants | SRC-A1-09/11 | VERIFIED (med) |
| TECH-008 | Exchequer | Reduces troops' base salary (up to 96%) | SRC-A1-11, -20 | VERIFIED (med) |
| TECH-009 | Cache | Theft-protected resource amount | SRC-A1-11, -14 | VERIFIED (med) |
| TECH-010 | Construction, Repair, Towers, Fortification, Magic Towers | Construction/repair improvements, defensive buildings | SRC-A1-11, -44 | VERIFIED (med; effects UNKNOWN) |
| TECH-011 | Heroes Managing, Hero Skills | Hire heroes (more per level). Without Hero Skills, hero max level is 12. | SRC-A1-11, -25 | VERIFIED (med) |
| TECH-012 | Vision, Reflection, Espionage, Diplomacy | Visionary/Mirror towers, spies, diplomacy | SRC-A1-11 | VERIFIED (med) |
| TECH-013 | Mission Control | Number of simultaneous missions | SRC-A1-11 | VERIFIED (med) |
| TECH-014 | Underground domains | Seize underground "Barracks" domains | SRC-A1-15 | VERIFIED (med) |
| TECH-015 | Clan Castles | Clan-wide science, up to 5 castles (A2) | SRC-A1-11 | VERIFIED (med) |
| TECH-016 | Alchemy (levels 1–6) | L1: magic crystals, improve buildings and artifacts. Later: runes, recipes, enchantments, ancient runes. L6: monumental buildings. | SRC-A1-24 | VERIFIED (low; non-EN) |
| TECH-017 | Land-use ("Землеустройство") | Any building on resource cells | SRC-A1-42 | VERIFIED (low) |
| TECH-018 | Resurrection | Seen in the guide's science order; effect UNKNOWN (likely unit revival) | SRC-A1-35 | INFERRED (low) |
| TECH-019 | Capital Transfer, City Expansion | See TECH-002 | SRC-A1-11, -14 | VERIFIED (high) |

- Test plan (catalogue): Each unlock is a boolean gate checked by the build or mission validator. Each level-scaled effect is tested against a table.

### TECH-020 — Specialisations (player-level)
- Description: Player picks one specialisation with 3 levels.
- Game purpose: Build-defining choice.
- Inputs: Study (cost UNKNOWN).
- Rules & constraints: Only one at a time. Starting a second **cancels** the previous one. Known values:
  - Architect: build, upgrade and repair time −50/65/80%.
  - Trader: market tax −2/5/9%; Trading Post tax +0.01/0.02/0.03.
  - Benefactor: cottage growth +8/12/16 per hour.
  - Defender: tower attack and fortification damage up (values UNKNOWN).
- Output: Passive bonuses.
- Related systems: BLD-000, ECO-017, POP-002.
- Evidence: SRC-A1-12: "a player can have a single specialisation… second Specialisation studying… will cancel the previous"; tables as above.
- Status: VERIFIED (high for Architect and Trader; med for others)
- Test plan: Switching specialisation removes the old bonuses immediately. Architect L3 plus other bonuses stays clamped at 80% total (BLD-000), if the cap applies to the sum.

### TECH-021 — Recommended early science order (community)
- Description: The guide's order: Construction 1 → Market 1 → Market 2 ("for the stone quest") → Fortification 1 → Cache 1 → Resurrection 1 → Merchantry 1 → Villages 1 → Gate.
- Game purpose: Shows the onboarding flow and that quests reference sciences.
- Inputs/Output: n/a.
- Rules & constraints: A community suggestion, not a rule.
- Related systems: QST-005.
- Evidence: SRC-A1-35 (Polish copy snippet).
- Status: VERIFIED as community advice (low)
- Test plan: Use as a reference for our onboarding pacing playtest.

---

## C. Population

### POP-001 — Domain population and hard cap
- Description: Each domain has its own population.
- Game purpose: Labour pool for the economy.
- Rules & constraints: Max **150,000** per domain. The excess disappears.
- Inputs: Growth (POP-002). Output: Population.
- Related systems: all producers.
- Evidence: SRC-A1-08: "Each domain can shelter maximum 150 000 population".
- Status: VERIFIED (med)
- Test plan: Population never exceeds 150,000.

### POP-002 — Population growth
- Description: Growth comes from the Main Building, Cottages, and idle-people reproduction.
- Game purpose: Steady labour income.
- Inputs: Buildings, idle count, services and specialisation.
- Rules & constraints: Idle reproduction per hour = **0.04 × idle people**. Total growth is limited by the living places of Cottages plus the Main Building. Base Cottage and Main Building growth values are UNKNOWN.
- Output: New idle people per hour.
- Related systems: BLD-019, ECO-021, TECH-020.
- Evidence: SRC-A1-08: "Idle Population reproduction per hour = 0,04*(number of idle people)".
- Status: VERIFIED (high)
- Test plan: 1,000 idle people with no buildings gives +40/h, clamped at capacity.

### POP-003 — Idle population
- Description: People without a job.
- Game purpose: Buffer that reproduces, feeds Trading Post gold, and can be captured by enemies.
- Rules & constraints: Idle people flee first on overcrowding. They increase Trading Post gold. They can be captured (raid scope).
- Inputs: Unassigned people. Output: Growth, gold.
- Related systems: BLD-015, POP-006.
- Evidence: SRC-A1-08.
- Status: VERIFIED (med)
- Test plan: On overpopulation, the idle count drops before the worker count.

### POP-004 — Workers and Alchemists (staffing)
- Description: Producing buildings must be filled with Workers (Labs with Alchemists) to produce.
- Game purpose: Population is the bottleneck for production.
- Rules & constraints: Output presumably scales with fill (INFERRED). Each building has a job count (UNKNOWN).
- Inputs: Idle people. Output: Production.
- Related systems: POP-007.
- Evidence: SRC-A1-09: "Producing buildings differ from other buildings because they have to be filled with Workers or Alchemists". SRC-A1-08: "Workers are needed to maintain Producing Buildings".
- Status: VERIFIED (high); scaling INFERRED
- Test plan: A producer with 0 workers yields 0.

### POP-005 — Builders
- Description: People assigned to construction sites.
- Game purpose: Construction speed is tied to population.
- Rules & constraints: Fewer Builders make the job take longer. They become idle when the job ends.
- Inputs: Idle people. Output: Construction progress.
- Related systems: BLD-000.
- Evidence: SRC-A1-09.
- Status: VERIFIED (high)
- Test plan: When construction finishes, the builder count returns to idle.

### POP-006 — Overcrowding and flight
- Description: If people exceed living places, the excess flees.
- Game purpose: Soft housing cap.
- Rules & constraints: Per hour, fleeing = N × (0.01717 + 0.00002828 × N), where N is the excess. Examples: 100 excess gives about 2% flee per hour; 10,000 excess gives about 30% per hour. (The formula gives 2 people and 3,000 people respectively, i.e. about 2% and 30% **of the excess**. The wiki wording says "of population".) Idle people leave first, then employed.
- Inputs: Excess population. Output: Population loss.
- Related systems: POP-001, POP-003.
- Evidence: SRC-A1-08: "100 excessive people cause 2% of population flee per hour, 10,000 excessive people cause 30%".
- Status: VERIFIED (med); interpretation of the "%" base is ambiguous
- Test plan: N=100 gives 2.0/h. N=10,000 gives about 3,000/h. Idle people are consumed first.

### POP-007 — Distribute Builders / Distribute Workers
- Description: UI actions that auto-assign idle people to vacant jobs.
- Game purpose: Low-friction management.
- Rules & constraints: Requires vacant places and idle population in the domain.
- Inputs: Idle people, vacancies. Output: Assignment.
- Related systems: POP-004/005.
- Evidence: SRC-A1-08: "drive people to work with the help of buttons 'Distribute Builders' and 'Distribute Workers'".
- Status: VERIFIED (med)
- Test plan: After distribution, min(idle, vacancies) people are moved.

### POP-008 — Building destruction effects on people
- Description: Deleting a building or finishing work makes its people idle. Enemy destruction kills them.
- Game purpose: War cost.
- Inputs/Output: Building removal leads to people becoming idle or dying.
- Rules & constraints: As described.
- Related systems: POP-003.
- Evidence: SRC-A1-09: "Workers and Builders from those buildings become Idle"; when destroyed by an enemy, "the Workers and Builders will die".
- Status: VERIFIED (med)
- Test plan: A self-demolish keeps the headcount. Enemy destruction reduces it.

### POP-009 — Units consume population
- Description: Training units costs resources **and population**. Peasants are the basic human infantry unit.
- Game purpose: Couples army size to population.
- Rules & constraints: Per-unit values UNKNOWN.
- Inputs: Population, resources. Output: Units.
- Related systems: BLD-022.
- Evidence: SRC-A1-20: tables show "the resources and population needed to train each unit". Peasant is "the most numerous infantry unit of human army".
- Status: VERIFIED (med)
- Test plan: Training is blocked if population is short.

---

## D. Quests, tutorial and early game

### QST-001 — Counsellor quests
- Description: An NPC adviser (Counsellor) issues quests.
- Game purpose: Onboarding and directed progression.
- Inputs: Completion of the quest condition.
- Rules & constraints: Quests are **optional**. After completing one, the player must return to the Counsellor to claim the reward.
- Output: Rewards.
- Related systems: QST-002..005.
- Evidence: SRC-A1-19: "After completing a quest given by the Counsellor, turn to the Counsellor again in order to receive the Reward". SRC-A1-18: "Quests are not obligatory."
- Status: VERIFIED (high)
- Test plan: A completed quest stays "claimable" until claimed. No auto-claim.

### QST-002 — Reward delivery and restrictions
- Description: Rewards are delivered to the Capital.
- Game purpose: Anchors rewards and prevents abuse.
- Rules & constraints: Some rewards cannot be taken if a battle in the Capital is less than 2 hours away.
- Inputs/Output: Claim leads to Capital stock.
- Related systems: BLD-003.
- Evidence: SRC-A1-19 / SRC-A1-14.
- Status: VERIFIED (med)
- Test plan: A claim is rejected when an incoming battle ETA is under 2 h.

### QST-003 — Protection rewards
- Description: Some quest rewards grant timed protection from new Monster Attacks or from Negative Events.
- Game purpose: Safety window for new players (a partial substitute for newbie protection).
- Rules & constraints: Does not cancel attacks already under way.
- Inputs/Output: Timed protection flag.
- Related systems: PvE (A2).
- Evidence: SRC-A1-19: "no new Monsters Attack will be launched against you in specified period".
- Status: VERIFIED (med)
- Test plan: During protection, no new PvE attack is scheduled. Existing ones proceed.

### QST-004 — Quest chains
- Description: Completing one quest unlocks the next.
- Game purpose: Guided sequence.
- Evidence: SRC-A1-33 (2012 forum): "passage of a previous quest activates the next". SRC-A1-35 mentions a "stone quest" tied to the Market science.
- Status: VERIFIED (low)
- Inputs/Output/Rules: The chain list is UNKNOWN.
- Related systems: QST-001.
- Test plan: Quest N+1 is hidden until quest N is claimed.

### QST-005 — Typical early-game flow (community-derived)
- Description: (1) Follow the Counsellor quests and avoid off-quest buildings. (2) Upgrade the Main Building, with players often rushing levels 4–5 using a BG service. (3) Study the early sciences (TECH-021). (4) Open the Gate, then use Villages for +30%. (5) Explore, colonise new cities, rob ruins and empty cities.
- Game purpose: Shows the intended onboarding funnel.
- Evidence: SRC-A1-35 ("follow… Counselor's quests", "Main Building… level 4 and 5", Gate last), SRC-A1-30, SRC-A1-42.
- Status: INFERRED (low)
- Inputs/Rules/Output: n/a.
- Related systems: BLD-001/002/004, TECH-021.
- Test plan: Our onboarding playtest checks that milestones are reachable in this order without paying.

### QST-006 — Formal tutorial / newbie protection
- Description: No evidence found for a scripted tutorial beyond Counsellor quests, or for a newbie-shield rule.
- Status: UNKNOWN
- Evidence: Searches of the encyclopedia mirrors returned only QST-003 protections and the "Mistakes" page.
- Test plan: n/a (design decision).

---

## Unknowns & discrepancies
1. **No per-level cost, time, output, capacity or job-count tables** were captured for any building. These are UNKNOWN. Verify by fetching individual building pages (for example `Wiki.jsp?page=Warehouse`, `Cottage`, `Main_Building`) once the `*.mlgame.org` and `mlgame.co.uk` hosts are allowed.
2. **Main Building max level 5** comes from a RU page (old DB version). The EN snippet says only "city level up to five".
3. **Number of additional cities**: "up to 7 more" (EN City page) vs "up to 8 more" (another mirror).
4. **City cell counts**: level-5 city 39–43 (EN table) vs level-4 city 34–38 (RU). Both may hold, but this is unconfirmed.
5. **Gate timing** ("opens 5 minutes and closes 6 hours"): ambiguous.
6. **Military building names** may be race variants. The mapping is UNKNOWN. "Barracks" also names an underground domain.
7. **Science prerequisites, costs and level counts**: UNKNOWN. Mirrors list different science sets (version drift between the Steam and UK servers).
8. **Population flight formula**: the base of the percentage (excess vs total) is ambiguous.
9. **Tutorial and newbie protection**: UNKNOWN.
10. The land-use science and "perfect" buildings come from RU fan sources only. They may be outdated.
11. All evidence is search-snippet based. Pages opened directly: 0.
