# Army, Missions and Combat — Research Notes (Agent A2)

Subject: public materials about the browser MMO "My Lands: Black Gem Hunting" (MLG).
Purpose: reconstruct *game systems* only, for an ORIGINAL design. Unit/race/building names below are
quoted only to identify the reference mechanic; our game must use its own names, numbers and texts.

Access date for every source: **2026-10-31**. Source IDs refer to `_sources_A2.md`.

## Method note (read first)

- Direct page fetching was **not possible** from this research environment: `steam.mlgame.org`,
  `us.mlgame.org`, `mlgame.co.uk`, `store.steampowered.com` and `en.wikipedia.org` were all refused by the
  session's network policy (DNS/CONNECT 403). All evidence was gathered through a web-search tool that
  returns indexed excerpts of the official encyclopedia (English, Ukrainian-server English, Russian, and
  other language mirrors) and the official forum.
- Consequence: evidence is "excerpt-level". Confidence is capped at **med** unless the same rule appears
  in two or more independent mirrors/languages with consistent wording (then **high**). Numbers taken
  from a single excerpt where table headers were not visible are marked **low**.
- The encyclopedia has many mirrors (steam.mlgame.org/wiki-en, mlgame.co.uk/wiki, uk.mlgame.org/wiki-uk,
  ru.mlgame.org/wiki-ru, ...). Pages are of different ages (some several years old); the search tool
  flagged several as 1–5+ years since last edit. Server-specific coefficients exist (see ARMY-003).

---

## A. Army and units

### ARMY-001 — Races and the 8-category unit roster
- Description: Each playable race fields 8 unit categories. Categories seen in the encyclopedia: Carriers,
  Warriors (infantry), Cavalry, Airborne (flyers), Archers, Healers, Mercenaries, Mages. Monsters are a
  separate NPC "race".
- Game purpose: Gives every race a comparable army skeleton (combat line + support + logistics) while
  allowing race flavour per category.
- Inputs: player race; studied war sciences.
- Rules & constraints: "There are 8 types of units for each race" (also stated in RU/UA versions).
  Playable races mentioned across sources: Elves, Knights, Wizards (Light faction), Dark Elves, Demons,
  Undead, Chimeras (Dark faction). A character may change race once.
- Output: a fixed 8-slot roster per race.
- Related systems: ARMY-002, ARMY-013, SOC (factions), HERO.
- Evidence: SRC-A2-02 (Units: "There are 8 types of units for each race"); SRC-A2-03 (UnitsDescription:
  numbered categories Warriors, Cavalry, Airborne, Archers, Healers, Mercenaries, Mages; carrier unit names
  in first section); SRC-A2-17 (Race page; "Each character can change their current race to any other one
  time"); SRC-A2-18 (Factions: Light = Elves, Knights, Wizards; Dark = Dark Elves, Demons, Undeads, Chimeras).
- Status: VERIFIED (high for "8 types"; med for exact category list — first category header "Carriers" was
  inferred from unit names, not seen as a header).
- Test plan: data-validation test: every race definition has exactly 8 unit categories, one per category
  enum; race-change counter allows exactly one change.

### ARMY-002 — Unit levels (tiers) inside each category
- Description: Each combat category has 4 levels (e.g. light → medium → heavy cavalry); higher levels are
  unlocked by levelling the matching war science.
- Game purpose: Vertical progression of the army without adding new categories.
- Inputs: war-science level for that category.
- Rules & constraints: UnitsDescription shows "four levels" per category (Cavalry: Light, Medium, Heavy...).
  Upgrading the war science to level 2+ converts existing units of that type and requires paying the cost
  difference between old and new level for every unit owned (incl. units on march, in Resurrector, in
  training). Owning zero units makes the upgrade free; releasing units first reduces the cost. Cost is
  taken from the warehouses of the selected domain.
- Output: all units of the category move to the new level.
- Related systems: ARMY-008, ARMY-009.
- Evidence: SRC-A2-03 (four levels per category); SRC-A2-02/SRC-A2-22 ("pay off the difference between the
  previous and next levels of units"; "If you don't have any Warriors, the upgrade costs nothing").
- Status: VERIFIED (med).
- Test plan: given N units at level k, upgrading science to k+1 charges N × (cost(k+1) − cost(k)) from
  selected domain; with N=0 charge is 0; units in transit/healing/queue are counted.

### ARMY-003 — Core unit parameters
- Description: Units have Attack (a min–max range), Health, Defence, Salary (gold upkeep), training time,
  population need, resource cost, and a "Rating" value = total resource cost of the unit. Carriers add
  Capacity; Healers add resurrection count; Mages add tower-damage reduction.
- Game purpose: Small, comparable stat block to drive combat and economy.
- Inputs: unit type & level; server coefficient.
- Rules & constraints: Attack = damage dealt; Health = damage a unit can withstand; Defence = share of
  damage repelled. Defence for all troops starts at 0 (RU page). Example row (column headers partly
  hidden): level-1 Warrior "2 - 3 | 10 | 0,2 | ... | 185" → attack 2–3, likely health 10, salary 0.2,
  rating 185. Page notes a server coefficient must be applied to listed values.
- Output: per-unit stat block.
- Related systems: CMB-003, ARMY-010.
- Evidence: SRC-A2-09 (forum "Carriers": "[Attack] – damage to be delivered ... [Health] – damage a unit can
  withstand [Defence] – damage to be repelled"); SRC-A2-05 (Defence starts at 0); SRC-A2-04
  (UnitsParameters row; "Rating in the table – is the cost of all resources for a certain unit";
  "Server coefficient should be taken into account").
- Status: VERIFIED for parameter set (med); numeric example INFERRED (low — headers not visible).
- Test plan: schema test that every unit has attack_min ≤ attack_max, health > 0, defence ≥ 0,
  salary ≥ 0; rating equals sum of resource costs; server multiplier applied consistently.

### ARMY-004 — Carriers (logistics unit)
- Description: Non-combat unit with zero attack whose special stat Capacity limits resources carried or
  slaves convoyed during missions.
- Game purpose: Makes looting and transport a planned logistics problem.
- Inputs: number of carriers × capacity per carrier (race-dependent).
- Rules & constraints: "Carrier has zero Attack Value"; capacity = "amount of resources carried or slaves
  convoyed". Knights race has the largest carrier capacity.
- Output: maximum haul for Robbery/Slaves Capture/Transport.
- Related systems: MSN-006, MSN-007, MSN-011, ARMY-013.
- Evidence: SRC-A2-03 (quote above); SRC-A2-17 ("greatest capacity of the Carriers' backpacks among other races").
- Status: VERIFIED (med). Capacity numbers UNKNOWN.
- Test plan: haul = min(available_loot, Σ carrier capacity); carriers contribute 0 attack in combat.

### ARMY-005 — Healers (post-battle resurrection)
- Description: Healers resurrect a capped number of killed units after combat.
- Game purpose: Sustain/attrition reduction; support role that rewards mixed armies.
- Inputs: surviving healers × resurrect-per-healer (by level, race bonus).
- Rules & constraints: Resurrect own party first, then allies' party; process "from the left to the right
  ... one by one from each group" (round-robin across groups, e.g. 10 Warriors + 100 Carriers killed with 20
  resurrections → 10 + 10). Table: 1 (Elves 2) at level 1 up to 4 (Elves 5) at level 4.
  Separate building "Resurrector" (unlocked by Resurrection science) also revives with a % chance; a hero
  "Cleric" specialisation adds +1/+3/+5 % resurrection there.
- Output: units restored to the army; kills counted for rating only after resurrection.
- Related systems: CMB-009, HERO-004, SOC (faction rating).
- Evidence: SRC-A2-03 ("If 10 Warriors and 100 Carriers were killed ... 10 Carriers and 10 Warriors will be
  resurrected"); SRC-A2-45 (Spec page: Cleric +1,3,5%); SRC-A2-18 (killed units
  counted after resurrection round).
- Status: VERIFIED (med). Per-level numbers med/low (snippet also gave inconsistent "1 to 3").
- Test plan: deterministic round-robin allocation test; own-before-ally priority; cap = Σ healer capacity.

### ARMY-006 — Mercenaries ignore fortifications
- Description: Mercenary category is not hit in the pre-battle Fortifications round.
- Game purpose: Dedicated siege-opener unit type.
- Inputs: unit category flag.
- Rules & constraints: "Aren't attacked by Fortifications". Fortification losses are distributed over all
  groups "except mercenaries".
- Output: fortification damage skips mercenary groups.
- Related systems: CMB-002, CMB-008.
- Evidence: SRC-A2-03; SRC-A2-06/SRC-A2-07 (Combat: fortification round excludes mercenaries).
- Status: VERIFIED (high — EN and RU agree).
- Test plan: run fortification round with mixed army; assert mercenary count unchanged.

### ARMY-007 — Mages reduce tower damage
- Description: Mages decrease damage attacking troops take from Defensive Towers and Magic Towers.
- Game purpose: Counter to static defence; siege specialist.
- Inputs: number/level of mages in attacking army.
- Rules & constraints: table values "10 to 40" reduction by level; Dark Elf mages "20 to 80"; Dark Elf mages
  also train faster. Lore says rulers use mages "only during sieges" because expensive. Unit of the
  reduction (absolute vs. %) UNKNOWN.
- Output: reduced tower damage to the attacker.
- Related systems: CMB-008, ARMY-013.
- Evidence: SRC-A2-03; SRC-A2-17 (Dark Elves "combat Mages are the best units to oppose enemy Towers").
- Status: VERIFIED (med); numbers low.
- Test plan: tower damage with k mages present < tower damage with 0 mages; reduction is capped ≥ 0.

### ARMY-008 — Training pipeline
- Description: War science → build matching military building → train units in a queue → units appear in
  the city garrison.
- Game purpose: Gate military growth behind research and construction.
- Inputs: science level, building level, resources, population, time.
- Rules & constraints: "For each type of units training the appropriate war science has to be studied";
  then the military building can be built. Building level raises queue size and lowers per-unit time
  (Mercenaries building example: queue 75 at L1 → 4,000 at L9; per-unit time 8 min → 50 s). Training can be
  accelerated by up to 90% (Units page). Airborne example: 20 min per unit (base).
- Output: units added to garrison.
- Related systems: ARMY-002, ARMY-012.
- Evidence: SRC-A2-02; SRC-A2-11 (TrainMercenaries table); SRC-A2-44 (TrainAirborne ("One unit training time,
  minutes: 20")).
- Status: VERIFIED (med); numbers med/low.
- Test plan: cannot enqueue without science+building; queue length ≤ building cap; per-unit time decreases
  monotonically with building level; acceleration never > 90%.

### ARMY-009 — Unit leveling cost on science upgrade
- See ARMY-002 (merged rule). Status VERIFIED (med).

### ARMY-010 — Upkeep: gold salary and desertion
- Description: Troops are paid salary in gold (not food). If gold is insufficient, unpaid units desert
  gradually.
- Game purpose: Soft cap on army size tied to economy.
- Inputs: unit counts × salary; gold stock.
- Rules & constraints: troops in garrison and Sanctuary consume salary; deserting rate "0.5%" of unpaid
  units per hour (worked example: 1 unit with no gold deserts after 200 hours → fractional accumulation).
  Units on missions cost double salary (round trip) times a mission-speed coefficient (MSN-002).
- Output: gold drain per hour; desertion events.
- Related systems: MSN-002, MSN-003, economy (other agent).
- Evidence: SRC-A2-02 ("When sending troops in combat or non-combat mission the troops will take double
  salary"); SRC-A2-02/RU Units (0.5%/h, 200 h example); SRC-A2-28 (Resources: "Gold is consumed by troops").
- Status: VERIFIED (high for "upkeep paid in gold" — EN + RU agree; med for the 0.5%/h figure).
- Test plan: with gold=0 and 1 unit, desertion accumulator reaches 1 unit at t=200 h; with enough gold
  no desertion; mission doubles salary.

### ARMY-011 — Sanctuary (troop hiding)
- Description: Troops can be hidden in a Sanctuary (after the Sanctuary science) to avoid battle.
- Game purpose: Defensive option against overwhelming attacks; still costs upkeep.
- Rules & constraints: troops in Sanctuary "do not fight" but still consume salary. Hero in Sanctuary:
  artefacts inactive, mana does not regenerate.
- Related systems: ARMY-010, HERO-006.
- Evidence: SRC-A2-02; SRC-A2-25 (Stunned page mana rule).
- Status: VERIFIED (med).
- Test plan: hidden units excluded from battle participants; upkeep still charged.

### ARMY-012 — Barracks (captured underground training domains)
- Description: Special underground domains ("Barracks") that, once seized, reduce resource cost, training
  time and population need for a matching unit type.
- Rules & constraints: requires "Underground domains" science; 43 cells, 8–18 cells hold level-7 military
  buildings; neutral (not monster) garrison; RU page: up to 50% reduction of base training time.
- Related systems: WLD (domains), ARMY-008.
- Evidence: SRC-A2-10.
- Status: VERIFIED (med); 50% figure low (RU only).
- Test plan: owning a Barracks of type X lowers cost/time for X only; stacking rules UNKNOWN → design decision.

### ARMY-013 — Race specialisations affecting the army
- Description: Races get asymmetric bonuses: Knights (gold, carrier capacity, warehouses, crops), Elves
  (wood; strong healers, +1 resurrection), Dark Elves (wood & iron; better and faster-training mages).
- Evidence: SRC-A2-17; SRC-A2-03 (Elves healer table +1).
- Status: VERIFIED (med).
- Test plan: race modifiers applied only to the stated categories; no modifier leaks to other races.

### ARMY-014 — Monsters (NPC army)
- Description: Monsters are an NPC race garrisoning Ruins and other sites, and attack players
  (a quest reward grants protection from new monster attacks; "Ancients" monster reinforcement protects
  newcomers).
- Evidence: SRC-A2-02 (Monsters special NPC race); SRC-A2-27 (monster protection quest); SRC-A2-40 (Ancients).
- Status: VERIFIED (med).
- Test plan: see WLD-005 and SOC-PVP rules.

---

## B. Missions

### MSN-001 — Mission taxonomy
- Description: Missions are launched from the Kingdom Map by clicking a target domain. Three groups:
  - Combat: Robbery, Slaves Capture, Buildings Destruction, Domain Seizure.
  - Hero missions: Domain Destruction, Encamping, Hero Release.
  - Non-combat: Espionage, Exploration, City Foundation, Observation, Resources Transportation,
    Troops Relocation. Also Assistance (reinforce) and Escape (hero return) appear elsewhere.
- Game purpose: One unified "send army from A to B with an intent" interface.
- Inputs: source domain, target, units, hero(es), speed, retreat threshold.
- Output: travel → resolution at target → return.
- Related systems: all MSN-*, CMB-*.
- Evidence: SRC-A2-01 (MyLands overview page lists the three groups); SRC-A2-06 ("Combat Missions are
  launched from the Kingdom Map after clicking a target domain").
- Status: VERIFIED (med).
- Test plan: enum coverage test; each mission type validates allowed target types.

### MSN-002 — Mission speed tiers
- Description: Player picks Slow / Normal / Accelerated / Fast; faster = costlier.
- Rules & constraints: coefficients Slow 1, Normal 1.25, Accelerated 1.5, Fast 2 (multiply troop pay).
  Explorer death chance 2–10% depends on speed (slower = safer). Assistance missions travel twice as fast as
  attacks. Escape moves at Fast march speed. Distance→time formula: UNKNOWN. Base unit speeds: UNKNOWN.
- Evidence: SRC-A2-02 ("For Slow Speed coefficient is 1, For Normal – 1.25, For Accelerated – 1.5, For
  Fast – 2"; same in RU); SRC-A2-32 (Assist); SRC-A2-33 (Exploration 2–10%).
- Status: VERIFIED (high for coefficients — EN+RU agree; med for others).
- Test plan: cost(speed) ratio matches coefficient table; explorer death chance monotonic in speed.

### MSN-003 — Squad award (mission pay) and minimum duration
- Description: Sending troops costs gold: salary × units × round-trip duration × speed coefficient.
- Rules & constraints: "Squad award = 1 unit salary × number of units × mission duration (both ways) ×
  coefficient"; minimum mission duration 15 minutes one way. Worked example: 465 Carriers, 3,884 s Fast
  mission ≈ 200.67 gold (salary unit/time base not shown; consistent with a per-hour salary).
- Evidence: SRC-A2-06 (Attack Launch Window); SRC-A2-02 (worked example).
- Status: VERIFIED (med). Time unit of salary INFERRED (per hour).
- Test plan: formula unit test incl. 15-min floor; reproduce worked example within rounding once salary
  units are fixed in our design.

### MSN-004 — Attack frequency cap per target
- Rule: a player can't launch more than 2 attacks on the same domain within 24 hours.
- Evidence: SRC-A2-06.
- Status: VERIFIED (med).
- Test plan: third attack on same target inside a rolling 24 h window is rejected.

### MSN-005 — Mission withdrawal (recall)
- Rule: combat missions can be withdrawn until they come "too close to the target" (threshold UNKNOWN).
- Evidence: SRC-A2-06 (search excerpt "Combat Missions can be withdrawn until they didn't come too close").
- Status: VERIFIED (med); threshold UNKNOWN.
- Test plan: recall allowed while remaining_time > threshold; rejected afterwards.

### MSN-006 — Robbery
- Description: Combat mission whose success transfers resources (and possibly artefacts/crystals) from the
  target to the attacker, limited by carrier capacity.
- Rules & constraints: Outer Domains: only 50% of resources or people can be stolen. Market stock can be
  stolen as well as warehouse stock. Excess gold without warehouse room is plundered. Resources/people are
  "usually burnt during robberies but not for the players in clan" (exact burn rule UNKNOWN). Artefact
  can be stolen from a Ruin only if a Hero takes part. With 1 hero, 1 random Crystal can be stolen from the
  defender hero's bag.
- Evidence: SRC-A2-12 (50%); SRC-A2-29 (Market); SRC-A2-28 (Resources); SRC-A2-23 (Artefacts); SRC-A2-03.
- Status: VERIFIED (med); burn rule INFERRED/UNKNOWN.
- Test plan: loot ≤ carrier capacity; ≤50% for outer domains; hero-gated artefact theft.

### MSN-007 — Slaves Capture
- Description: Combat mission that takes population from the target; carriers convoy captives.
- Rules & constraints: captives add to population; if population exceeds housing, excess flees gradually.
  Camps can be targeted by Robbery and Slaves Capture. Forum: captives can die if convoyed slowly (anecdotal).
- Evidence: SRC-A2-03; SRC-A2-30 (Camp); SRC-A2-37 (Population); SRC-A2-41 (forum quest guide).
- Status: VERIFIED (med); death-in-transit INFERRED (low).
- Test plan: captured ≤ carrier capacity; overflow population decays.

### MSN-008 — Buildings Destruction
- Description: Combat mission that damages buildings in villages/outer domains (not inside City Wall).
- Rules & constraints: buildings inside the City Wall cannot be destroyed; destroyed buildings kill their
  Workers and Builders (own demolition frees them). Forum: each attack targets one building type.
- Evidence: SRC-A2-15 (Villages); SRC-A2-16 (Buildings); SRC-A2-14 (forum Defense basics).
- Status: VERIFIED (med); one-type-per-attack INFERRED (low, forum).
- Test plan: target building inside wall → rejected; destruction removes assigned workers.

### MSN-009 — Domain Seizure
- Description: Combat mission that transfers ownership of a non-capital domain.
- Rules & constraints: requires Outer Domains science; succeeds only if no defensive buildings (Tower,
  Magic Tower, Fortification) remain; after seizure, level-1 buildings are destroyed fully and higher-level
  ones partially; a hero whose battle is won but seizure fails is not stunned. See WLD-004.
- Evidence: SRC-A2-12; SRC-A2-14; SRC-A2-25.
- Status: VERIFIED (med).
- Test plan: seizure with any defensive building present → battle may win but ownership unchanged.

### MSN-010 — Espionage
- Description: Spies (trained in a Spy Guild) are sent to a domain; resolution is instant as a fight
  between spies and defender's counterspies; result is an access level 1–4 controlling report detail.
- Rules & constraints: range limited by Espionage science level; no counterspies → level 4 access
  ("all information about the player's city"); lower levels show only resources/cell positions and hide
  troops/sciences. Spy reports can be saved in messages. Observation mission exists but details UNKNOWN.
- Evidence: SRC-A2-13 ("If there is no Counterspies in the domain which is spied on, then Spies will get
  level 4 access").
- Status: VERIFIED (med).
- Test plan: access level is a function of spy vs counterspy outcome; level 4 when counterspies=0;
  report fields filtered by level.

### MSN-011 — Resources Transportation / Troops Relocation
- Description: Non-combat moves between own domains (and possibly allies). Capacity-limited by carriers.
- Evidence: SRC-A2-01 (mission list); SRC-A2-03 (carrier capacity); SRC-A2-42 (MoveResources page exists).
- Status: VERIFIED (existence, med); detailed rules UNKNOWN.
- Test plan: transported amount ≤ capacity; delivered on arrival.

### MSN-012 — Assistance (reinforcement)
- Description: Send troops to help another domain; travels at 2× attack speed; joins defence as an
  additional army (see CMB-007).
- Evidence: SRC-A2-32 ("Assistance mission speed is twice as fast as attack missions").
- Status: VERIFIED (med).
- Test plan: travel_time(assist) = travel_time(attack)/2.

---

## C. Combat resolution

### CMB-001 — Battle structure
- Description: Automatic, round-based. Optional "Fortifications" pre-round, then up to 20 combat rounds;
  after battle, resurrection/necromancy rounds.
- Rules & constraints: max 20 rounds; if the attacker hasn't won by round 20 it is a defeat ("troops are
  exhausted"). Clan Castle internal gate also uses 20 rounds.
- Evidence: SRC-A2-06 (EN) and SRC-A2-07 (RU) ("Максимальное количество раундов в бою - 20"); SRC-A2-31.
- Status: VERIFIED (high).
- Test plan: simulation never exceeds 20 rounds; round-20 stalemate → attacker loses.

### CMB-002 — Fortifications round
- Rule: defender fortifications kill a fixed number of attacking troops before combat; losses are spread in
  proportion to group size regardless of unit strength; mercenaries exempt.
- Evidence: SRC-A2-06, SRC-A2-07 ("В этом раунде Укрепления защитника убивают определенное количество войск").
- Status: VERIFIED (high). Kill amount formula UNKNOWN.
- Conflict: forum guide (SRC-A2-14) says Fortifications "raise the city's defence level" — see Unknowns.
- Test plan: proportional distribution test; mercenary exemption.

### CMB-003 — Damage calculation per round
- Description (RU algorithm): (1) sum the whole army's attack incl. bonuses/penalties + Tower and Magic
  Tower attack; (2) reduce by a % equal to defenders' Defence; (3) subtract from health. Attack is
  distributed across enemy unit groups proportionally to their size; overflow damage a group cannot absorb
  is redistributed proportionally to remaining groups. Pursuers add a separate extra strike on victims.
- Evidence: SRC-A2-07 ("уменьшаем атаку на %, равный защите обороняющихся войск"; "Остаточный урон
  распределяется на другие группы").
- Status: VERIFIED (med; RU only, excerpt-level). Randomness: attack is a min–max range (ARMY-003) and
  the calculator averages 30 simulations ⇒ battles are stochastic (INFERRED, med).
- Test plan: deterministic test with fixed RNG seed; damage = attack × (1 − defence%); overflow conserved.

### CMB-004 — Pursuer / victim counter loop
- Rule: Warriors → pursue Archers; Archers → pursue Airborne; Airborne → pursue Cavalry; Cavalry → pursue
  Warriors. Pursuers deal normal damage to all enemies and additionally strike their victim category again.
- Evidence: SRC-A2-03 ("Warriors pursue Archers and are victim for Cavalry"); SRC-A2-07 (riders strike
  warriors a second time).
- Status: VERIFIED (high — EN + RU agree). Size of extra strike UNKNOWN.
- Test plan: with victims present, pursuer deals extra damage only to the victim category; no extra if
  victim absent.

### CMB-005 — Retreat thresholds
- Rule: each army has a loss % at which it withdraws; checked at end of each round; retreated army skips
  later rounds. Defender's threshold set in the city's main building; attacker's in the mission-launch
  window. Calculator displays all thresholds.
- Evidence: SRC-A2-06; SRC-A2-08.
- Status: VERIFIED (high).
- Test plan: army retreats at end of the first round its cumulative loss% ≥ threshold.

### CMB-006 — Victory condition
- Rule: the side under its loss limit while the other reaches its limit wins; if both reach the limit in
  the same round, the defender wins; attacker loses on round-20 timeout.
- Evidence: SRC-A2-06.
- Status: VERIFIED (med).
- Test plan: tie-break test → defender.

### CMB-007 — Multiple armies per side
- Rule: several armies on one side add their attack; damage received is split proportionally across all
  groups of all armies, respecting pursuer/victim pairs.
- Evidence: SRC-A2-06, SRC-A2-07.
- Status: VERIFIED (high).
- Test plan: two allied armies vs one; loss split proportional to group sizes.

### CMB-008 — Defensive buildings
- Description: Tower (damage by level), Magic Tower (damage scales with own garrison strength — zero if
  garrison empty, per forum), Fortification (pre-round kills). Each unlocked by a Mage-branch science.
- Evidence: SRC-A2-16, SRC-A2-22 (sciences unlock Tower/Fortification/Magic Tower); SRC-A2-14 (forum:
  "This tower deals damage based on the strength of your army").
- Status: VERIFIED (existence, med); damage rules INFERRED (low, forum).
- Test plan: Magic Tower damage = f(garrison) with f(0)=0.

### CMB-009 — Resurrection & necromancy after battle; kill counting
- Rule: after combat rounds, resurrection (healers) and necromancy rounds occur; kills for rating and for
  dungeon capture chance are counted after them.
- Evidence: SRC-A2-18 ("killed units are counted up after the round of resurrection"); SRC-A2-26.
- Status: VERIFIED (med). Necromancy mechanics UNKNOWN (linked to Undead/Skeleton units).
- Test plan: rating points computed on net losses after resurrection.

### CMB-010 — Battle calculator
- Rule: in-game predictor fed with both sides' troops and defensive buildings; predicts losses, gold
  equivalent, loot, attacker profit; result = average of 30 simulations.
- Evidence: SRC-A2-08.
- Status: VERIFIED (high: several mirrors).
- Test plan: our simulator exposes N-run Monte-Carlo average; variance > 0 for random attack ranges.

### CMB-011 — Hero outcome in battle
- Rule: hero on losing side is Stunned (cannot go on missions, skills/artefacts inactive, no mana regen);
  defeated hero can be imprisoned (100% chance stated in Dungeon context). See HERO-006/HERO-007.
- Evidence: SRC-A2-25, SRC-A2-26.
- Status: VERIFIED (med).

### CMB-012 — Battle / spy reports
- Description: Results are delivered as messages; spy reports can be saved. Battle-report layout UNKNOWN.
- Evidence: SRC-A2-13; SRC-A2-43 (Mes page exists).
- Status: INFERRED (low).
- Test plan: report contains participants, per-group losses before/after resurrection, loot.

---

## Unknowns & discrepancies

1. **Direct access blocked** — every rule above is from search excerpts. Verify by opening the pages
   (SRC-A2-02..08) in a browser once the hosts are allowed.
2. **Numeric unit tables** (attack/health/defence/salary/speed/capacity per unit and level) — only one
   partial Warrior row seen; column labels inferred. Verify on UnitsParameters (EN) / UnitParametrs (RU).
3. **Travel time formula and unit movement speeds** — not found. Only speed-tier coefficients known.
4. **Fortifications**: official Combat page says they kill troops in a pre-round; forum guide says they
   raise defence level. Treat official page as primary; verify in-game help.
5. **Healer resurrection numbers** — snippet gave both "1–3" and table "1 (Elves 2) → 4 (Elves 5)".
6. **Mage reduction units** (absolute vs %) and **pursuer extra-strike size** unknown.
7. **Robbery burn rule** ("burnt ... but not for players in clan") — mechanism unclear.
8. **Mission recall threshold** and **Observation** mission behaviour unknown.
9. **Battle report format** not documented in reachable excerpts; watch YouTube gameplay to OBSERVE.
10. Pages carry different versions; server-specific coefficients exist (Mining/Tactical servers etc.).
