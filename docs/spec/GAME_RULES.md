# Realmforge — Game Rules (normative)

Status: normative specification of engine behaviour as implemented in `server/internal/rules` and `server/internal/game` (pack format v1).
Author role: System/Economy Designer (Agent C). Date: 2026-10-31.
Companion documents: `BALANCE_MODEL.md` (numbers for the shipped packs), `PLATFORM_REQUIREMENTS.md`, research notes in `docs/research/`.

This document is the single source of truth for game rules. Code comments reference rule IDs (`R-…`) defined here. When this document and the code disagree, the disagreement is a bug in one of them and must be logged in Appendix D. Appendix D also lists every disagreement found while this version was written.

---

## 0. Conventions

### 0.1 Rule IDs
`R-<AREA>-<n>`. IDs are stable. A retired rule keeps its ID and is marked *Retired*. Areas:

| Area | Section | Area | Section |
|---|---|---|---|
| TIME | §1 time model | MOV, MSN | §8 movement, missions |
| MOD | §2 modifiers | PVP | §9 PvP restrictions |
| RES, ECO | §3 economy | CMB | §10 combat |
| POP | §4 population | SCT | §11 scouting |
| ARMY-UPKEEP, ARMY | §5 army economy | PVE | §12 neutrals |
| BLD, TECH | §6 construction, research | TER, CITY | §13 territories, cities |
| SCORE | §7 score | HERO | §14 heroes |
| MKT | §15 market, NPC exchange | SOC | §16 alliances, social |
| QST | §17 quests | CON | §18 conflict resolution |
| AX | §19 anti-exploit | | |

### 0.2 Status labels
Every rule with a gameplay value has a status:

- **VERIFIED (ID)**: the mechanic, and the value where one is given, match a research item marked VERIFIED in `docs/research` (the ID is the research item). The research came only from search-engine excerpts of "My Lands" pages, so VERIFIED means "consistent with the excerpt", not "confirmed by play".
- **ADAPTED (ID)**: the research verified the mechanic exists, but our formula, value or scope is our own design.
- **PROPOSED**: our design. The research has no evidence, or marks the item UNKNOWN or INFERRED.

### 0.3 Notation and rounding
- `P.x` is pack parameter `params.x` after engine defaults (`pack.Defaults()`) are applied. `B.x`, `U.x`, `T.x` are building, unit and tech fields.
- `ws` is the **effective world speed**: `P.worldSpeed × worlds.speed` (R-TIME-8).
- `floor`, `ceil` are the usual functions. `round` is **round half away from zero** (Go `math.Round`), e.g. `round(22.5) = 23`, unlike banker's rounding.
- Integer division `a*b/100` on int64 values truncates (`floor` for non-negative values).
- `clamp(x, lo, hi) = min(hi, max(lo, x))`.
- Units: resource amounts in whole **units** (stored as milli-units, R-TIME-2), rates in **units per real hour**, durations in **real seconds** unless noted. A "game-hour" is `1/ws` real hours.

### 0.4 Worked-example basis
Unless stated otherwise, examples use the **Ashen Crowns** pack (`packs/ashen-crowns/pack.json` v1.0.0), `ws = 1`. Every number was computed with the engine formulas re-implemented in `/tmp/claude-0/-home-user/1547afcd-c6d4-5979-a753-a17e66a9906c/scratchpad/balance.py` (`examples` mode). That script mirrors `rules/*.go` exactly, including rounding.

---

## 1. Time model

### R-TIME-1 — Server authority and lazy settlement
- **Rule.** All game state changes happen on the server. A client sends intents only. The economic state of a city is valid as of a stored instant `econ_updated_at`. Before any read or command, the player is **settled**. Settling locks the player and their cities, applies every pending event with `due_at ≤ now` (R-TIME-5), and integrates the economy of **every** city of the player up to each event instant and finally to `now`.
- **Piecewise-constant integration.** Inside an integration interval, buildings, techs, garrison and away-upkeep are constant. Every change to them is a scheduled event (job completion, mission arrival or return, delivery), so the interval is split at that event's instant. Population, garrison (desertion) and overflow decay change only at tick boundaries (R-TIME-4). Rates are recomputed after any tick that changed population or garrison.
- **Path independence.** For any partition of `[t0, t1]`, the result is identical up to ≤ 1 milli-unit per resource per settlement (R-TIME-2). How often a player is observed does not change outcomes.
- **Status:** PROPOSED (architecture).
- **Code:** `rules.AdvanceCity`, `game.settleLocked`.

### R-TIME-2 — Fixed-point storage
- **Rule.** Stocks are stored as integers in milli-units: `1 unit = 1000 milli`. Accrual over `Δms` milliseconds at net rate `r` units/h adds `floor(r × 1000 × Δms / 3 600 000)` milli. Remainders below 1 milli are dropped per segment. Displayed and spendable amounts are `floor(stock / 1000)` (`rules.Whole`). Affordability compares milli: `stock_r ≥ cost_r × 1000`.
- **Edge cases.** Negative net with a zero stock leaves the stock at 0 (never negative). The floor applies to negative deltas too: `floor(−0.4) = −1` milli.
- **Example.** Net grain `+31.2/h` over a 7 s segment: `31.2 × 1000 × 7000 / 3 600 000 = 60.67` → **+60 milli**.
- **Status:** PROPOSED.
- **Code:** `rules.Accrue`, `rules.Milli`, `rules.CanAfford`.

### R-TIME-3 — Durations and the time-factor floor
- **Formula.** For level `L ≥ 1`:
  `t(L) = max(1, ceil( base × growth^(L−1) × f / ws − 10⁻⁹ ))` seconds, where
  `f = max(0.2, 1 + Σpct(target)/100)` and `target ∈ {build_time, research_time, train_time}`.
- **Inputs.** `base = time.baseSeconds ∈ [1, 2 592 000]`, `growth ∈ [1, 5]` (validate.go). Only `pct` modifiers count. `flat` modifiers on time targets are ignored.
- **Clamp.** Bonuses can shorten a duration by at most 80% (`f ≥ 0.2`). Penalties (`f > 1`) are not capped.
- **Rounding.** `ceil` with a 10⁻⁹ guard against float noise. Minimum 1 s.
- **Example.** Keep L10: `120 × 1.42⁹ = 2816.93` → **2817 s**. With Path of the Architect 3 (−45%) and Keep 10 (−20%), `f = 0.35` → `ceil(985.9)` = **986 s**. With `ws = 2` and no modifiers: **1409 s**. Path of the Architect 3 plus Keep 25 (−95%) gives `f = 0.05`, floored to **0.2**.
- **Status:** VERIFIED (BLD-000: "time reduction no more than 80%", debuffs uncapped).
- **Code:** `rules.DurationAt`, `Mods.TimeFactor`, `MinTimeFactor`.

### R-TIME-4 — Absolute economy ticks
- **Rule.** Tick boundaries are the instants `k × P.economyTickSeconds` since the Unix epoch, UTC. `P.economyTickSeconds ∈ [60, 3600]` must divide 86 400. A tick at boundary `b` runs after resources have accrued up to `b`. A tick advances `tickHours = P.economyTickSeconds / 3600 × ws` game-hours.
- **Tick order (fixed):** (1) starvation or growth (R-POP-2/3), (2) clamp to housing (R-POP-1), (3) desertion (R-ARMY-UPKEEP-2), (4) overflow decay (R-RES-5).
- **Edge cases.** A settlement that ends exactly on a boundary runs that tick. One that ends 1 ms before does not, and the next settlement runs it. A city created at time `t` gets its first tick at the next boundary after `t`, which can be less than a full tick.
- **Status:** PROPOSED.
- **Code:** `rules.AdvanceCity` (`t.Truncate(tick).Add(tick)`), `rules.TickStep`.

### R-TIME-5 — Events and ordering
- **Event kinds:** `job` (build, research or train completion), `mission_arrive`, `mission_return`, `delivery` (market proceeds).
- **Ownership.** Each event belongs to the player whose state it changes:
  - `mission_arrive` at a **city** belongs to the city's owner. At a tile, it belongs to the mission owner.
  - `mission_return` belongs to the mission owner.
  - `delivery` belongs to the recipient.
  - `job` belongs to the job owner.
- **Ordering.** A player's pending events are processed in ascending `(due_at, id)`. `id` is a monotonically increasing sequence, so **for equal due times the event scheduled first wins** (R-CON-2). Before each event, all of the player's cities are advanced to `due_at`, and `now` is set to `due_at` for the handler.
- **Due condition.** An event is due when `due_at ≤ settlement instant`, inclusive.
- **Status:** PROPOSED.
- **Code:** `game.settleLocked`, `game.schedule`.

### R-TIME-6 — Exactly-once processing
- **Rule.** An event's effects and its status change to `done` commit in one transaction with the player lock held. Further guarantees:
  - `(kind, ref_id, due_at)` is unique, so the same event cannot be scheduled twice.
  - Handlers are idempotent with respect to the referenced row. A job that is no longer `queued`, or a mission no longer `outbound` (arrival) or `returning` (return), makes the event a no-op. This is how cancellation and recall neutralise pending events.
  - If a handler fails, the transaction rolls back. The worker records `attempts+1` and `last_error`. After 5 failed worker attempts, the event is parked as `failed` and skipped (admin `RetryEvent` re-queues it).
- **Status:** PROPOSED.
- **Code:** `settleLocked`, `Service.Process`, migration `0001_init.sql`.

### R-TIME-7 — Worker versus lazy settlement
- **Rule.** Two paths reach the same `settleLocked` code. A **command** settles the acting player to `now`. The **worker** polls players that have due events and settles each one with `FOR UPDATE SKIP LOCKED`, so a player busy in a command is skipped and picked up on the next poll. Players without events are never advanced by the worker. Their state is integrated the next time anything locks them. By R-TIME-1 the result is the same.
- **Observable consequence.** Stored values such as `players.score` (used in rankings and in the anti-bullying check, R-PVP-1) can be stale for inactive players.
- **Status:** PROPOSED.
- **Code:** `Service.RunWorker`, `DuePlayers`, `Process`.

### R-TIME-8 — World speed
- **Formula.** `ws = P.worldSpeed × worlds.speed`, with `P.worldSpeed ∈ [0.1, 1000]` and `worlds.speed ∈ (0, 1000]`.
- **Scaled by `ws`:**
  - production, consumption and upkeep rates (×ws)
  - tick length in game-hours (×ws)
  - build, research and train durations (÷ws)
  - travel time and minimum travel time (÷ws)
  - neutral respawn (÷ws)
  - territory yield accrual (×ws)
- **Not scaled:**
  - costs, capacities, stocks, combat
  - newbie protection (`P.newbieProtectionHours`, real hours)
  - hero wounds (`P.heroWoundHours`, real hours)
  - the 24-hour attack-frequency window and the message rate window
- **Example.** At `ws = 2`, a Lumberyard L1 yields 64 timber per real hour, and Keep L10 takes 1409 s.
- **Status:** PROPOSED. See D-13 for the unscaled real-hour timers.
- **Code:** `Service.Pack` (multiplies), `ComputeRates`, `TickStep`, `DurationAt`, `TravelSeconds`, `battleNeutral`, `territoryYield`.

### R-TIME-9 — Locking discipline
- **Rule.** A transaction locks at most **one player aggregate**: the `players` row, then that player's `cities` in id order, then the rows it touches (events, jobs, missions, heroes, items, tiles, orders). Effects on other players never lock those players. They are carried by an event owned by the affected player (R-TIME-5), or by a row lock on a shared row (mission, market order, tile, alliance).
- Data about other players needed for combat (race, techs, alliance, score, protection) is read **without locking** as a snapshot (`pctx.other`).
- Deadlocks and serialization failures are retried up to 4 times with quadratic back-off (`InTx`).
- **Status:** PROPOSED.

---

## 2. Modifiers

### R-MOD-1 — Sources and scope
| Source | Scaling | Scope | Used for |
|---|---|---|---|
| Race `modifiers` | × 1 | player-wide | economy, combat, travel, everything |
| Tech `modifier` effects | × tech level | player-wide | everything |
| Building `modifier` effects | × building level | **that city only** | economy, build/research/train time, defence of that city, dispatch speed/carry from that city |
| Hero (class + perLevel×(L−1) + skills×level + items) | — | the hero's mission or home-city defence | combat, travel, carry, loot, scout, XP |
| Faction penalty (R-PVP-4) | × 1 | one battle | attacker `attack.*` |

- A building's `modifier` effect scales **linearly** with level (`value × L`). The effect `growth` field is not used for modifiers. Only `produce`, `storage`, `protect`, `housing` and `idle_tax` use `v(L)` (R-ECO-1).
- Hero and territory modifiers never enter the city economy (see D-2).
- When the attacker's army arrives, its modifiers are the owner's race + techs + hero, **without** the departure city's building modifiers (see D-3).
- **Status:** PROPOSED.
- **Code:** `pctx.recomputeMods`, `pctx.cityMods`, `pctx.other`, `heroMods`.

### R-MOD-2 — Application order (the only order)
- **Formula.** For targets `T = {t₁, t₂, …}` (e.g. `attack.militia`, `attack.infantry`, `attack.*`):
  `value = max(0, (base + Σ_{t∈T} flat_t) × max(0, 1 + Σ_{t∈T} pct_t / 100))`.
- **Single exception:** gross production adds `production.*` flat bonuses after the percentage (R-ECO-2, D-24).
- Bonuses are **additive** within each kind, across all sources. They never compound multiplicatively. Flat is added before the percentage is applied.
- **Constraints.** `pct` values in a pack must be in `[−95, 1000]` per modifier (validate.go). Sums are unbounded but clamped at ≥ 0 by the formula.
- **Example.** An Ashborn militia (base attack 8) with Tactics 3 (+9), a Marshal hero at L5 (+5 innate, +4 per-level) and an Iron Blade (+3), plus Ashborn's own `attack.*` +8:
  - `Σpct = 8 + 9 + 9 + 3 = 29`, giving attack `8 × 1.29` = **10.32**.
  - An Ironvow militia (defence 10) wearing a Leather Jerkin: `(10 + 2) × 1.08` = **12.96**.
- **Status:** ADAPTED (ECO-010: individual bonuses verified, how they combine was INFERRED).
- **Code:** `Mods.Apply`.

### R-MOD-3 — Duration modifiers
`f = max(0.2, 1 + Σpct/100)`. See R-TIME-3.

### R-MOD-4 — Target grammar (closed set)
| Target | Read by |
|---|---|
| `production.<res\|*>` | gross production (R-ECO-2) |
| `storage.<res\|*>` | storage cap (R-RES-3) |
| `protect.<res\|*>` | loot protection (R-RES-7) |
| `housing` | housing cap |
| `pop_growth` | growth rate (pct only) |
| `build_time`, `research_time`, `train_time` | R-TIME-3 |
| `upkeep` | army upkeep (R-ARMY-UPKEEP-1) |
| `attack.*`, `defense.*`, `hp.*` (`.unitId`, `.class`, `.*`) | battle stats (R-CMB-1) |
| `speed` (pct only) | travel (R-MOV-1), floored at −90% |
| `carry` | carrying capacity (R-MSN-5, R-CMB-7) |
| `loot` | loot percentage (R-CMB-7) |
| `scout` | scouting power (R-SCT-1) |
| `wall` | defensive wall points (R-CMB-1) |
| `hero_xp` | hero XP gains (R-HERO-2) |
| `market_fee` | market fee (R-MKT-5; see D-1) |

Unknown targets make the pack invalid.

---

## 3. Economy

### R-RES-1 — Resource classes
- Each resource has `kind ∈ {basic, food, rare, premium}` and flags `capped`, `tradable` and `lootable`, plus `baseStorage ≥ 0`.
- Premium resources are never tradable, lootable or transportable (validate.go and `SendMission`). They come from quests, PvE rewards and administration.
- Uncapped resources (e.g. `lore`, `starglass`) have no storage limit and do not decay.
- **Status:** ADAPTED (ECO-006 premium gems; ECO-007 science points as a separate currency).

### R-RES-2 — Spending
- **Rule.** A cost is paid in full or not at all. Under the player lock: if `CanAfford` (R-TIME-2) fails, the command fails with `insufficient_resources`, listing the missing whole units (`ceil(missing_milli / 1000)`). Otherwise the cost is deducted and written to `resource_ledger`.
- Resources are local to the city that pays. There is no player-wide pool.
- **Status:** VERIFIED (ECO-024: domain-local economy).
- **Code:** `pctx.spend`, `rules.Spend` (panics on overspend: accounting bugs must be loud).

### R-RES-3 — Storage capacity
- **Formula** (capped resources only):
  `cap_r = floor( (baseStorage_r + Σ_b v_storage,b(L_b) [resource r or "*"] + Σflat(storage.r, storage.*)) × max(0, 1 + Σpct(storage.r, storage.*)/100) )`, in units. As everywhere (R-MOD-2), flat is added **before** the percentage.
- **Example.** Warehouse L3: `1200 × 3 × 1.12² = 4515.84`. Adding base 1500 gives 6015.84. With Masonry 2 (+6%): `floor(6376.79)` = **6376**.
- **Status:** VERIFIED (ECO-012: a Warehouse limit summed over buildings). Capacity values are PROPOSED.
- **Code:** `ComputeRates`.

### R-RES-4 — Hard-cap mode (`P.overflowDecayPctPerHour = 0`)
- **Accrual.** Positive accrual of a capped resource stops at `cap`. A stock already above the cap (e.g. after a grant) is not reduced, but cannot grow further. Consumption still reduces it.
- **Grants** (quest rewards, deliveries, returning cargo, transport arrival, refunds, market receipt, exchange, territory yield): the new stock is `min(stock + grant, max(cap, stock))`. The **excess is lost** and reported (`lostToStorage`).
- **Example (Void Frontier, cap 3500).** Alloy 3400 plus a returning cargo of 300 → stock 3500, **lost 200**.
- **Status:** VERIFIED (ECO-012: "production halts when the limit is reached").
- **Code:** `rules.Accrue`, `rules.Grant`, `pctx.caps`.

### R-RES-5 — Soft-cap mode with overflow decay (`P.overflowDecayPctPerHour > 0`)
- **Rule.** The cap does not stop production, and grants are never clipped (`caps = nil`). At every tick (R-TIME-4), each capped resource loses part of its surplus `s = stock − cap × 1000` (milli):
  `loss = ceil( s × P.overflowDecayPctPerHour / 100 × tickHours )` milli, when `s > 0`.
- **Example (Ashen Crowns, 1%/h, 10-min tick).** Timber 3700 with cap 2700: surplus 1000 → loss per tick `ceil(1 000 000 × 0.01 / 6)` = 1667 milli = **1.667 timber**.
- **Edge case.** Decay acts on the stock after that tick's accrual. Uncapped resources never decay.
- **Status:** VERIFIED (ECO-013: "1% of resource lacking storage space is plundered per hour"). ECO-012 and ECO-013 conflict in the research, so both are offered as pack modes. Ashen Crowns uses ECO-013 and Void Frontier uses ECO-012.
- **Code:** `rules.TickStep` step 3.

### R-RES-6 — Non-negativity
No operation produces a negative stock. Accrual clamps at 0, spending is pre-checked, and decay only removes surplus.

### R-RES-7 — Protected stock (vault)
- **Formula.** `prot_r = floor( (Σ v_protect(L) [r or *] + Σflat(protect.r, protect.*)) × max(0, 1 + Σpct(protect.r, protect.*)/100) )`, defined only when the raw sum or the flat bonus is > 0.
- Protected units cannot be looted (R-CMB-7).
- **Example.** Vault L2: `150 × 2 × 1.1` = **330** of each resource.
- **Status:** VERIFIED (ECO-014: cache/theft protection exists). Values are PROPOSED.

### R-ECO-1 — Effect value by level
- **Formula.** `v(L) = perLevel × L × g^(L−1)` with `g = growth` (0 or 1 means linear), and `v(0) = 0`. No rounding is applied here. Rounding happens where the value is used.
- **Example.** Lumberyard (`perLevel 32, g 1.06`): `v(1) = 32`, `v(5) = 32 × 5 × 1.06⁴` = **202.00 timber/h**.
- **Status:** PROPOSED.
- **Code:** `rules.EffectValue`.

### R-ECO-2 — Employment and gross production
- **Formulas.**
  - `W = Σ_b workers_b × L_b`, summed over `produce` effects only (workers needed).
  - `employed = min(pop, W)`, `idle = pop − employed`.
  - `e = employed / W`, or `1` if `W = 0` (employment).
  - `gross_r = max(0, (Σ_b v_produce,b,r(L_b) × e + Σ_b v_idle_tax,b,r(L_b) × idle) × max(0, 1 + Σpct(production.r, production.*)/100) + Σflat(production.r, production.*)) × ws`.
- **Exception to R-MOD-2.** For production only, `flat` is added **after** the percentage (as the `ComputeRates` comment says). For example, a +10 flat timber bonus is not boosted by Forestry. See D-24.
- **Units.** Units per real hour (float, unrounded). Employment is shared evenly: every producer runs at the same fraction `e`. Players cannot prioritise.
- **Example (fresh Ashen Crowns city, Ironvow, pop 40).** `W = 18`, `employed 18`, `idle 22`, `e = 1`. Timber **32**, grain **36**, stone `26 × 1.10` = **28.6**/h. With pop 12: `e = 0.667`, giving timber 21.33, grain 24.0, stone 19.07.
- **Status:** ADAPTED (POP-004: producers must be staffed, VERIFIED; linear scaling with fill was INFERRED. ECO-005/BLD-015: Trading Post income from idle people, formula INFERRED).
- **Code:** `ComputeRates`.

### R-ECO-3 — Consumption and net rate
- `cons_food = pop × P.foodPerPopPerHour × ws`.
- `cons_upkeepRes += upkeep` (R-ARMY-UPKEEP-1). If the upkeep resource is the food resource, both terms add up.
- `net_r = gross_r − cons_r`.
- **Example.** Pop 40 × 0.12 = **4.8 grain/h**. The start army of 10 militia × 0.3 and 2 scouts × 0.2 costs **3.4 gold/h**, so net gold is **−3.4/h**.
- **Status:** PROPOSED. POP-002 found no evidence that population eats. Food upkeep was chosen for legibility.

### R-ECO-4 — Housing
`housing = floor( (Σ v_housing(L) + Σflat(housing)) × max(0, 1 + Σpct(housing)/100) )`. Example: Keep L1 (60) + Cottages L1 (90) = **150**.
- **Status:** VERIFIED (POP-002: growth limited by the living places of Cottages and Main Building). Values are PROPOSED.

---

## 4. Population

### R-POP-1 — Housing cap
After each tick, `pop = clamp(pop, 0, housing)`. Population above housing (possible after migration, or in a settled city) disappears **at the next tick**. There is no gradual flight (see D-19).
- **Status:** ADAPTED (POP-001: a hard cap exists; POP-006: gradual flight was VERIFIED but not implemented).

### R-POP-2 — Growth
- **Formula.** When starvation does not apply and `pop < housing`:
  - `base = housing − pop` in mode `free_housing`, or `base = idle` in mode `idle`.
  - `g = base × P.popGrowthPctPerHour × max(0, 1 + pct(pop_growth)/100) / 100 × tickHours + carry_pop`.
  - `pop += floor(g)` and `carry_pop = g − floor(g)`.
- **Example (Ashen Crowns, free_housing 12%/h, 10-min tick).** Tick 1: `(150 − 40) × 0.12 / 6 = 2.2`, so pop becomes 42 with carry 0.2. After six ticks (1 h), pop is **52** with carry 0.6. After 24 h it is 144 (BALANCE_MODEL §8).
- **Status:**
  - Mode `idle`: VERIFIED (POP-002: "idle reproduction per hour = 0.04 × idle". The research value is 4%/h. Void Frontier uses 8%).
  - Mode `free_housing` and the Ashen Crowns value of 12%: PROPOSED (gentler onboarding: growth does not stall when every inhabitant works).

### R-POP-3 — Starvation
- **Rule.** At a tick, if food stock `≤ 0` **and** `net_food < 0` **and** `pop > 0`:
  `pop −= ceil(pop × P.starvationPctPerTick / 100)`, `carry_pop = 0`, and no growth happens in that tick.
- **Example.** Pop 400, housing 628, grain 0, net −12/h: 400 → 396 → 392 → … → **376** after one hour (each tick removes `ceil(1%)` = 4).
- Starvation stops at the first tick where net food ≥ 0 or the stock is > 0.
- **Status:** PROPOSED.

### R-POP-4 — Carry and determinism
Fractional growth (`carry.pop`) and desertion (`carry.desert.<unit>`) remainders are stored per city (`econ_carry`) and survive between settlements. Combined with R-TIME-4, this makes population independent of how often the player is observed.
- **Status:** PROPOSED.

---

## 5. Army economy and training

### R-ARMY-UPKEEP-1 — Upkeep
- **Formula.** `upkeep = max(0, (Σ_{u∈garrison} n_u × U.upkeep_u + away + flat(upkeep)) × max(0, 1 + pct(upkeep)/100)) × ws`, paid in `R = P.upkeepResource` (defaults to `P.foodResource`), in units/h.
  `away = P.missionUpkeepMult × Σ_{missions not done} Σ_u upkeep_units_u × U.upkeep_u`.
- `upkeep_units` is the mission's composition **at departure** and is not reduced by losses (see D-6). Units stationed in an ally's city are paid by the **sender's** origin city. Foreign troops stationed in a city cost the host nothing.
- **Example.** 50 militia on a mission in Ashen Crowns (`missionUpkeepMult 2`): `50 × 0.3 × 2` = **30 gold/h**. For Gravewarden (−20%): **24 gold/h**.
- **Status:** VERIFIED (ECO-015: army salary in gold; double salary on missions). The double comes from MSN-003's round-trip and speed coefficient, simplified to a flat multiplier.

### R-ARMY-UPKEEP-2 — Desertion
- **Rule.** At a tick, if upkeep-resource stock `≤ 0` **and** its net `< 0`, each **garrison** stack (home units only) loses
  `d = n × P.desertionPctPerHour / 100 × tickHours + carry_u`, with `floor(d)` units removed and the fraction carried.
- Units on missions never desert. Desertion lowers upkeep, so it is self-limiting.
- **Example (Ashen Crowns, 0.5%/h).** 10 militia without gold: `10 × 0.005 / 6 = 0.00833` per tick, so the first deserter leaves after **120 ticks = 20 h**. After 24 h, 9 militia remain (verified by simulation).
- **Status:** VERIFIED (ECO-015: "0.5% of unsupplied units per hour").

### R-ARMY-1 — Trainability
A unit is trainable in a city iff:
- `trainedIn ≠ ""` and that building is at level ≥ 1 in the city;
- the race is allowed;
- all `requires` are met using **current** levels;
- the train queue has fewer than `P.trainQueueMax` jobs.

Units without `trainedIn` are neutral-only guards.

### R-ARMY-2 — Training batches
- **Formulas.**
  - `each = DurationAt(U.trainSeconds, growth 1, L 1, f = timeFactor(train_time), ws)`.
  - `batch time = each × qty`.
  - `cost = U.cost × qty`, saturating at 10¹⁵ per resource.
  - `qty ∈ [1, 100 000]`.
- The batch is paid up front and starts at the end of the city's train queue. **All units appear at once** when the batch completes. There is no per-unit trickle.
- **Example.** Barracks 2 (−4%) in a city with Range 2 (−2%): `f = 0.94`, so a militia takes `ceil(45 × 0.94)` = **43 s**. 25 militia take **1075 s** and cost grain 750, timber 500, iron 125.
- **Status:** PROPOSED.
- **Code:** `pctx.Train`.

### R-ARMY-3 — Unit stats
Attack, defence (a damage-reduction **percentage**), HP, speed (tiles/h), carry (units), upkeep (units/h of R), scout power, class and class bonuses. Ranges come from validate.go: `hp ∈ (0, 1e6]`, `speed ∈ (0, 1000]`, bonus `pct ∈ [−90, 500]`.

---

## 6. Construction and research

### R-BLD-1 — Next level and eligibility
- The next level is `targetLevel + 1`, where `targetLevel = max(current, highest queued level of that building)`.
- An upgrade is blocked (in order) by: `race` → `requirements` → `queue_full` → `resources`.
  - Requirements are checked against **current** levels, not queued ones. You cannot queue Ironworks (needs Keep 2) while Keep 2 is still under construction.
- At `maxLevel` the error is `max_level`.
- **Status:** VERIFIED (BLD-000) for sequential upgrades. Requirement trees are PROPOSED.

### R-BLD-2 — Cost
- **Formula.** `cost_r(L) = round(base_r × growth^(L−1))`, round half away from zero, with `growth ∈ [1, 5]` and `base ∈ [0, 1e12]`.
- **Example.** Keep L2: `{timber 180, stone 180, gold 60}`. Lumberyard L2: `40 × 1.38 = 55.2` → 55 and `30 × 1.38 = 41.4` → 41. Rounding edge: `15 × 1.5 = 22.5` → **23** (not 22).
- **Status:** PROPOSED. ECO-016 numbers are UNKNOWN.

### R-BLD-3 — Duration
`t(L)` as in R-TIME-3 with `target = build_time` and that city's modifiers. Example: Keep L2 in a fresh city (Keep 1 gives −2%) takes `ceil(120 × 1.42 × 0.98)` = **167 s**.

### R-BLD-4 — Queue
- The queue holds at most `P.buildQueueMax` jobs, including the one currently running.
- A new job starts at `queueEnd = max(now, max finishes_at of queued jobs)`. Jobs run **sequentially**.
- The cost is paid at enqueue time.
- The duration is fixed at enqueue time from the modifiers in effect then. Later bonuses do not retime queued jobs.
- **Status:** ADAPTED (POP-005/BLD-000: Builders limit parallelism; we use a fixed queue).

### R-BLD-5 — Completion
On its `job` event, `level = max(level, job.level)`. Effects apply from that instant: the economy interval is split there (R-TIME-1).

### R-BLD-6 — Cancellation
- Only the **last** job of a queue can be cancelled. Cancelling it cannot shift the times of other jobs.
- **Refund formula.** `refund_r = floor(cost_r × P.cancelRefundPct / 100)` (integer division). The refund is granted to the job's city under R-RES-4/5. The pending event is cancelled.
- **Example.** Cancelling Lumberyard L5 (`timber 145, stone 109`) refunds **timber 116, stone 87**.
- A running job that is the only job is the last job, so it can be cancelled.
- **Status:** PROPOSED.

### R-TECH-1 — Research
- Techs are player-wide (`players.techs`). The queue is player-wide with `P.researchQueueMax` slots.
- The cost is paid by the city that issues the command. The duration uses that city's `research_time` modifiers (e.g. its Academy).
- Requirements of type `building` are checked in **that** city.
- Completion triggers `recomputeMods`. The effect applies to all cities from that instant.
- **Status:** VERIFIED (ECO-007: pooled science currency). Structure is PROPOSED.

### R-TECH-2 — Specialisations (`exclusiveGroup`)
- Techs that share an `exclusiveGroup` are mutually exclusive. If any other tech of the group has level > 0 **or is queued**, research fails with `specialisation_taken`.
- The choice is permanent. There is no respec in v1.
- Example: the `calling` group (Architect / Steward / Warlord).
- **Status:** ADAPTED (HERO-004 and the Trader specialisation in ECO-017 show exclusive specialisations in the original).

---

## 7. Score

### R-SCORE-1
- **Formula.** `score = Σ_cities Σ_b B.score × L_b + Σ_t T.score × L_t + Σ_cities Σ_u U.score × n_u(home)`.
- The score is recomputed at each save. Units **away** do not count (see D-9).
- It is used for rankings, the anti-bullying floor (R-PVP-1) and hero XP (`score lost`, R-HERO-2).
- **Status:** ADAPTED (ECO-022: rank points = rating + army strength).

---

## 8. Movement and missions

### R-MOV-1 — Travel time
- **Formula.**
  `dist = √((x₁−x₂)² + (y₁−y₂)²)` tiles (Euclidean).
  `v = min_{u in mission, n>0} U.speed_u × max(0.1, 1 + Σpct(speed)/100)` tiles/h.
  `t = max( ceil(P.minTravelSeconds / ws), ceil(dist / v × 3600 / ws), 1 )` seconds.
- **Inputs.** The speed modifiers at dispatch are those of the departure city (race, techs, buildings) plus the hero.
- **Example.** From (10,10) to (16,12): `dist 6.325`. Militia (speed 5): **4554 s**. With Thornkin +5%: **4337 s**. 2 tiles with scouts (speed 16) gives 450 s, raised to the minimum **900 s**. A mixed army containing a ballista (speed 2.5) with Logistics 1 (+4%): **8758 s**.
- **Status:** ADAPTED (MSN-003: minimum mission duration of 15 min one way, VERIFIED, which is why Ashen Crowns uses 900 s; MSN-002 speed tiers are not implemented).

### R-MOV-2 — Return time
- After a resolved arrival, the return trip takes the **outbound duration** (`arrive_at − depart_at`).
- A recalled **outbound** mission returns in `now − depart_at`, the time already travelled.
- A withdrawn **stationed** mission returns in its outbound duration.
- A mission with no surviving units "returns" instantly (travel 0). Its hero is back immediately.

### R-MSN-1 — Mission kinds and common validation
- **Kinds:** `attack`, `scout`, `transport`, `reinforce`, `settle`, `occupy`.
- **Common checks at dispatch:**
  - known kind;
  - `retreatPct ∈ [0, 100]`;
  - ≥ 1 unit, each count in `[1, 10⁹]`, with that many units **at home** in the city;
  - target ≠ departure tile;
  - target inside the map and passable.
- Units leave the garrison immediately. Away upkeep starts (R-ARMY-UPKEEP-1).
- The arrival event is scheduled per R-TIME-5. A defender sees an incoming `attack` notification at dispatch.

### R-MSN-2 — Hero attachment
- Optional. The hero must be owned, `idle`, not wounded, and based in the departure city.
- The hero's modifiers apply to speed and carry at dispatch, and to combat, scout, loot and XP at arrival.
- The hero is `mission` until the mission returns or ends.

### R-MSN-3 — Scout missions
Only `class = scout` units may be sent. A target owned by another player under newbie protection is rejected (`target_protected`). Each dispatch increments the `scouts_sent` stat. Resolution: R-SCT.

### R-MSN-4 — Attack targets
Valid targets, in order:
1. A city: PvP checks R-PVP-1.
2. A tile with an **active** neutral (not cleared): PvE.
3. A tile owned by another player (territory): PvP checks R-PVP-1.
4. Anything else is `no_target`.

`capture = true` asks to take a territory on victory (R-TER-4).

### R-MSN-5 — Transport
- **Rule.** Targets a city: own, or anyone's (no alliance check).
- `capacity = floor((Σ n_u × U.carry_u + flat(carry)) × max(0, 1 + pct(carry)/100))`, using departure-city modifiers + hero.
- `Σ cargo ≤ capacity`. Cargo cannot include premium resources.
- Cargo is escrowed from the city at dispatch (R-MSN-9).
- On arrival, the cargo is granted to the target city (caps apply, R-RES-4). The units then return empty after the outbound duration.
- **Example.** 5 wagons (`carry 300`) from a city with Market 2 (+6%): `floor(1500 × 1.06)` = **1590**.
- **Status:** VERIFIED (ECO-020 carriers; MSN-011 transportation).

### R-MSN-6 — Reinforce
- The target must be a city that is own or **allied** (same alliance, or the sender's alliance has `ally` status toward the target's alliance; R-SOC-4). Non-aggression (`nap`) does not allow reinforcing.
- **Arrival at an own city:** relocation. Units join the garrison, the mission is `done`, and the hero is re-based there.
- **Arrival at a foreign city:** if `distinct other players already stationed ≥ P.maxReinforcingPlayers`, the mission is rejected and sent back. Otherwise it becomes `stationed` and defends (R-CMB-8) until recalled.
- **Status:** VERIFIED (MSN-012, SOC-015: up to 3 helpers). The 2× assistance speed is **not** implemented (see D-15).

### R-MSN-7 — Settle (found a city)
- **Dispatch:** needs `P.settlerUnit ≥ 1`. The tile must be free: no city, no active neutral, not owned by another player. `cities < cityCap` (R-CITY-1). Optional cargo ≤ capacity.
- **Arrival:** all checks are repeated. On failure: `settle_failed` and return.
- **On success:**
  - One settler is consumed.
  - A city is created with **every start building at level 1**, `pop = floor(start.population / 2)`, and stock = cargo (set directly; caps are not applied).
  - The remaining units become its garrison, and the hero is re-based there.
  - The tile's ownership is cleared.
- **Status:** VERIFIED (WLD-003: settler and city-foundation mission). Contents are PROPOSED.

### R-MSN-8 — Occupy
- **Dispatch:** the tile must have no city and no active neutral, and must not be owned by another player. Owned by nobody requires `territories < territoryCap`. Occupying an own territory adds a garrison.
- **Arrival:** checks are repeated. On success, a free tile becomes owned (`owner_since = yield_collected_at = now`) and the mission becomes `stationed` (the territory's garrison).
- **Status:** ADAPTED (WLD-004/MSN-009: outer domains and seizure).

### R-MSN-9 — Cargo escrow
Cargo for `transport` and `settle` is spent at dispatch (ledger `mission_cargo`). It is in transit, cannot be looted, and is lost only if the mission is destroyed. A mission has no combat, so in practice it is never lost.

### R-MSN-10 — Arrival resolution order
On `mission_arrive`, if the mission is no longer `outbound` the event is a no-op (recalled). Otherwise:

**A. Target is a city** (event owned by the city owner):

| Kind | Resolution |
|---|---|
| transport | Grant cargo, then return. |
| reinforce | R-MSN-6. |
| scout | R-SCT. |
| attack | City battle (R-CMB-8), then loot, then return. |
| other | Return. |

**B. Target is a tile** (event owned by the mission owner). The tile row is locked, then the first matching case wins:
1. A city now stands there (founded after departure): `target_changed`, return.
2. `scout`: tile intel (R-SCT-5).
3. `attack` and the neutral is active: PvE battle (R-PVE).
4. `attack` and the tile is owned by another player: territory battle (R-TER-4).
5. `occupy` (or `reinforce` on an own territory; unreachable, see D-5): R-MSN-8.
6. `settle`: R-MSN-7.
7. Otherwise: `nothing_found`, return. Example: the neutral was cleared by someone else in the meantime.

### R-MSN-11 — Return processing
On `mission_return` (only if `returning`):
- Units rejoin the origin city.
- Cargo (loot, rewards) is granted under caps (R-RES-4/5).
- Deferred stats (`missions_won`, `neutrals_cleared`) are credited.
- Dropped items are granted to the player.
- The hero becomes `idle`, gains `heroXp` (R-HERO-2) and, if `heroWounded`, gets `woundedUntil = now + P.heroWoundHours`.
- The mission becomes `done` and away upkeep is recomputed.

### R-MSN-12 — Recall
- An owner may recall a mission that is `outbound` (normatively: while `now < arrive_at`; the code does not enforce this, see D-22) or `stationed`.
- An `outbound` recall returns in `now − depart` (R-MOV-2). The pending arrival becomes a no-op.
- A `stationed` recall returns in the outbound duration.
- `returning` and `done` missions cannot be recalled.
- **Status:** ADAPTED (MSN-005: recall allowed until "too close", threshold UNKNOWN; we allow it until arrival).

---

## 9. PvP restrictions

### R-PVP-1 — PvP gate at dispatch
An attack on another player's city or territory is rejected if any of these holds:
1. The target is the attacker (`own_target`).
2. The two are **friendly** (R-SOC-3): `friendly_target`.
3. The target is under newbie protection (`protected_until > now`): `target_protected`.
4. `P.attackScoreFloorPct > 0` and `target.score × 100 < attacker.score × floor`: `target_too_weak`.
5. `P.maxAttacksPerTargetPerDay > 0` and the attacker already launched that many `attack` missions to the same `(x, y)` with `depart_at > now − 24 h` (recalled ones count): `attack_limit`.

- **Example (Ashen Crowns).** A third attack on the same city within 24 h is rejected.
- **Status:**
  - (5): VERIFIED (MSN-004/SOC-004: "no more than 2 attacks on the same domain within 24 hours").
  - (4): ADAPTED (SOC-001 rank rule). It is off in both packs.
  - (3): PROPOSED (SOC-003: the original had compassion spells and Ancients, not a timer).
- **Arrival.** Checks 1–5 are not repeated when an attack arrives at a city. Friendliness is re-checked for territory battles (see D-7).

### R-PVP-2 — Attacking ends own protection
Passing the PvP gate (R-PVP-1) sets the attacker's `protected_until = null` immediately. Scouting does not end protection.
- **Status:** PROPOSED.

### R-PVP-3 — Newbie protection
- Starts at join: `protected_until = join + P.newbieProtectionHours`, in real hours.
- A protected player cannot be attacked or scouted.
- A protected player **can** attack neutrals and scout tiles.

### R-PVP-4 — Same-race and same-faction penalty
- If the attacker's race equals the defender's race, add `pct(attack.*) += P.sameRaceAttackPct`.
- Otherwise, if both races have the same non-empty `faction`, add `P.sameFactionAttackPct`.
- The two do **not** stack. The penalty is additive with other attack bonuses (R-MOD-2).
- It applies in city and territory battles, never in PvE.
- **Example.** Ashborn (+8% attack) attacking Ashborn: `8 − 50` = **−42%**. Attacking Gravewarden (same faction, `duskbound`): **−17%**.
- **Status:** VERIFIED (FAC-002: −50% own race, −25% own faction). Not stacking is our choice: stacking was INFERRED from a forum.

---

## 10. Combat

Battles are deterministic. Identical inputs always give identical results. The research (CMB-003/CMB-010) indicates the original used random attack ranges. We deliberately chose determinism for replayability and testability (PROPOSED).

### R-CMB-1 — Stacks and modifiers
- One stack per (owner group, unit type). Group 0 is the attacker or the home garrison. Groups 1..n are each stationed reinforcement, in mission-id order.
- For unit `u` with modifiers `M`, targets are `stat.<unitId>`, `stat.<class>`, `stat.*`:
  - `A = M.apply(U.attack)`
  - `D = M.apply(U.defense) + wall`
  - `HP = max(1, M.apply(U.hp))`
- **Wall** (city defence only): `wall = (0 + Σflat(wall)) × (1 + Σpct(wall)/100)` percentage points, using the defended city's modifiers. It is added to **every** defending stack's `D`, garrison and reinforcements alike.
- **Modifier sets:**
  - attacker: owner race + techs + hero + faction penalty;
  - home garrison: city modifiers (race + techs + that city's buildings) + home hero;
  - reinforcements: their owner's race + techs;
  - neutral guards: none.
- **Example.** Walls L5 gives `wall = 10`, so a defending militia has `D = 10 + 10 = 20`.
- **Status:** ADAPTED (CMB-002/CMB-008: fortifications exist; the forum interpretation "raises the city's defence level" was chosen over a pre-round kill).

### R-CMB-2 — Rounds
- Up to `P.combatRounds` rounds (`[1, 20]`). A round runs only while both sides have units.
- Both sides compute damage from the counts **at the start of the round** and apply it simultaneously. A unit killed this round still strikes this round.
- **Status:** VERIFIED (CMB-001: max 20 rounds). Simultaneity is PROPOSED.

### R-CMB-3 — Damage distribution
- **Formula.** Damage from attacking stack `s` to target stack `t`:
  `dmg(s→t) = n_s × A_s × (n_t / N) × (1 + bonus_s(class_t)/100) × (1 − min(capDef, max(0, D_t))/100)`
  - `N = Σ` counts of all living enemy stacks (all groups);
  - `bonus_s(c)` = the first matching class bonus of `s` against class `c`, else 0;
  - `capDef = P.defenseCapPct ∈ [0, 99]`.
- A target stack receives `Σ_s dmg(s→t)`. Damage is spread **in proportion to unit count**, not to strength, across all groups of the side, so reinforcements share losses.
- **Overflow** damage that exceeds what a stack can absorb in that round is **lost**, not redistributed (see D-14).
- **Example (round 1, 40 militia + 20 bowmen vs 55 brigands + 15 wolves).**
  - Militia → brigands: `40 × 8 × 55/70 × (1 − 0.08)` = 231.31.
  - Bowmen → wolves: `20 × 14 × 15/70 × 1.40 × (1 − 0.04)` = 80.64.
  - Total attacker damage this round: **580.18**.
- **Status:** VERIFIED (CMB-003: defence reduces attack by a percentage; CMB-007: proportional to group size). Class bonuses are ADAPTED from CMB-004 (pursuer/victim, implemented as a percentage damage bonus instead of an extra strike).

### R-CMB-4 — Kills and wounds
- **Formula.** For stack `t` with incoming damage `X` and carried wounds `w`:
  `k = floor((X + w) / HP)`.
  - If `k ≥ n`: the stack is destroyed and `w = 0`.
  - Otherwise: `n −= k` and `w = X + w − k × HP`.
- Wounds carry over between rounds within one battle and are discarded afterwards. Surviving units are fully healed.
- **Example.** 460.8 damage on 45 wolves (HP 35): `k = 13`, `w = 5.8`.
- **Status:** VERIFIED (CMB-003: residual damage is carried). Healing after battle is PROPOSED.

### R-CMB-5 — Retreat
- After each round, a side **retreats** if it still has units, its threshold is < 100, and `cumulative losses × 100 ≥ threshold × starting count` (counted in units, not value).
- The attacker's threshold is the mission's `retreatPct` (1..100; 0 or absent means 100, never). The defender's threshold is `P.defenderRetreatPct`.
- If both qualify in the same round, **only the attacker retreats** (the defender holds). The battle ends at the first retreat.
- A retreating defender keeps its survivors in place (city garrison) but counts as beaten.
- **Example.** 40 outriders with `retreatPct 30` against 60 pikemen: 12 outriders are lost in round 1 (30%). The attacker retreats and the defender wins.
- **Status:** VERIFIED (CMB-005/CMB-006).

### R-CMB-6 — Victory
- `attackerWon ⇔ attacker has units ∧ attacker did not retreat ∧ (defender has no units ∨ defender retreated)`.
- A timeout (all rounds used) is an attacker **defeat**.
- An undefended target (no defending units) is an immediate attacker win with 0 rounds.
- `scoreLost_side = Σ losses_u × U.score_u` feeds hero XP.
- **Example.** 60 Ironvow militia (D 10.8) against Wolf Den L3 (45 wolves): 4 rounds, **attacker wins**, 16 militia lost (44 left), 45 wolves killed. Round damages: 460.8 / 281.0, 407.0 / 199.8, 368.6 / 131.1, 345.6 / 62.4.
- **Status:** VERIFIED (CMB-006).

### R-CMB-7 — Loot
- **Formula** (attacker won a **city** battle):
  - `capacity = floor((Σ_survivors n × carry + flat(carry)) × max(0, 1 + pct(carry)/100))`, using attacker arrival modifiers (no building modifiers; see D-3).
  - `lp = min(100, (P.lootPct + flat(loot)) × (1 + pct(loot)/100))`.
  - `avail_r = floor(max(0, stock_r − prot_r) × lp / 100)` for lootable resources.
  - If `Σavail ≤ capacity`, take `avail`. Otherwise take `floor(avail_r × capacity / Σavail)` for each `r`.
- Loot is spent from the city (ledger `plundered`) and carried home as cargo. The stock is read as whole units after settlement to the arrival instant.
- **Example.** City with timber 2400, grain 1800, stone 900, iron 300, gold 500 and a Vault L2 (330 protected each). The Ashborn attacker has `lp = 55`.
  - `avail` = timber 1138, grain 808, stone 313, iron 0, gold 93 (Σ 2352).
  - 40 outriders carry 1600, so the loot is **timber 774, grain 549, stone 212, gold 63** (Σ 1598; flooring loses 2).
- **Status:** ADAPTED (MSN-006 robbery: "only 50%" in outer domains; carry-limited).

### R-CMB-8 — Battle participants
| Battle | Attacker | Defender |
|---|---|---|
| City | mission units | garrison (city modifiers + wall + home hero) + every `stationed` mission at that city (owner modifiers + wall), groups by mission id |
| Neutral | mission units | level guards, no modifiers |
| Territory | mission units | every `stationed` mission on the tile (owner modifiers, no wall, no hero) |

**Home hero:** the highest-level `idle`, unwounded hero based in the city. Its modifiers are merged into the garrison's.

### R-CMB-9 — After the battle
- **Survivors.** Each group's survivors are written back. A reinforcement group with no survivors is closed (returned with travel 0).
- **Territory battles:** if the attacker won, every surviving defending group is sent home (travel = its outbound duration).
- **Attacker hero:** XP = `floor((defender scoreLost × xpPerScore + flat(hero_xp)) × max(0, 1 + pct(hero_xp)/100))`, plus the neutral level's `xp` if a PvE battle was won. It is credited on return. The hero is wounded if the attacker did not win.
- **Defending home hero:** gains XP from the attacker's `scoreLost` immediately. It is wounded if the attacker won.
- **Stats.** The attacker gets `missions_won` (and `neutrals_cleared`) on return. A defender who wins gets `defenses_won`.
- **Status:** ADAPTED (CMB-011: the losing hero is stunned; HERO-008: XP from killed troops' value).

### R-CMB-10 — Power indicator (UI only)
`power = round(Σ n × (A + D) × HP / 10)`. It is not used in resolution.

---

## 11. Scouting

### R-SCT-1 — Success
- **Formula.**
  - `att = M_att.apply(Σ n_u × U.scout_u, "scout")`, using mission units and attacker arrival modifiers;
  - `def = M_city.apply(Σ garrison n_u × U.scout_u, "scout")`, using **home garrison only** (stationed foreign scouts do not count) and the defended city's modifiers.
- Success iff `att > 0 ∧ att > def`.
- On **failure**, all scouts are lost, and the defender is always notified (receives the report).
- **Example.** 10 scouts against 6 defending scouts: success.

### R-SCT-2 — Intel level
- On success, `intel = 2` if `att ≥ 2 × def`, else `1`.
- Intel 1 shows resources, garrison and foreign troops.
- Intel 2 also shows buildings and population.
- **Status:** ADAPTED (MSN-010: spies vs counterspies with access levels 1–4; no counterspies means full access).

### R-SCT-3 — Detection
- **Formula.** On success: `P(detect) = (b + (100 − b) × def / (att + def)) / 100`, with `b = P.scoutDetectBasePct`. Detected iff `roll < P(detect)`, where `roll = RNG(SeedFrom(missionId, 7)).Float()` (R-PVE-4).
- A detected defender receives the report.
- **Example (b = 30).** att 10, def 4: P = 0.50. att 10, def 6: P = 0.5625. No defenders: P = 0.30.
- **Status:** PROPOSED.

### R-SCT-4 — Incoming visibility
A defender sees incoming `attack` and `scout` missions targeting their cities (arrival time, owner) but **not** their composition, cargo, hero or result.

### R-SCT-5 — Tile scouting
Scouting a tile always succeeds with intel 2 and is never detected. It shows the active neutral's level, guards and rewards, and the territory's owner and total garrison.

---

## 12. Neutrals (PvE)

### R-PVE-1 — Placement
- The map is generated once from the world seed (SplitMix64, R-PVE-4).
- **Terrain:** weighted pick per tile, scanning rows y = 0.., then x = 0...
- **Neutral:** on a passable tile, with probability `P.neutralDensityPct%`, a neutral type is picked uniformly. Its level is
  `L = 1 + floor(d × maxL × 0.999)`, where `d = distance to map centre / distance from centre to corner`, then lowered by 1 with probability 1/4 when `L > 1`.
- The level is fixed per tile.
- **Example (48×48, 5 levels).** Tile (40,40): `d = 0.702`, giving L4 (or L3 with 25% probability).
- **Status:** ADAPTED (WLD-005: ruins scale in level). Distance-from-centre scaling is PROPOSED.

### R-PVE-2 — Engagement
- Attacking an active neutral fights the level's guards with no modifiers (R-CMB-8).
- On victory:
  - the tile is cleared until `now + floor(respawnSeconds / ws)`;
  - the level `rewards` are added to the mission cargo **in full, regardless of carry capacity** (see D-8);
  - item drops are rolled (R-PVE-3);
  - the hero gains XP (R-CMB-9).
- On defeat, nothing changes for the neutral. Guards do not stay wounded and fully regenerate.
- **Status:** ADAPTED (WLD-005: in the original, a ruin disappears and respawns elsewhere; ours respawns in place).

### R-PVE-3 — Drops
- For each listed drop, in order: drop iff `rng.Intn(100) < chance`, with `rng = RNG(SeedFrom(missionId, 11))`.
- Drops are independent, and several items can drop.
- Items are granted on return.
- **Status:** VERIFIED (ECO-023: items from ruins). Chances are PROPOSED.

### R-PVE-4 — Randomness
- All randomness is seeded and reproducible:
  - map generation: world seed;
  - scouting detection: `(missionId, 7)`;
  - drops: `(missionId, 11)`.
- Combat is fully deterministic.
- Spawn-point choice (`findSpawn`) is the only non-seeded choice: a database `random()` pick among the legal tiles of one ring.
- **Status:** PROPOSED.

---

## 13. Territories and cities

### R-TER-1 — Territory cap
- `territoryCap = (max level of P.territoryCapBuilding over own cities) × P.territoriesPerLevel`, or 0 if no building is configured.
- It is checked when a free tile is occupied (dispatch and arrival) and when a tile is captured (arrival).
- **Example (Ashen Crowns).** Keep 3 allows **3** territories.

### R-TER-2 — Occupation and garrison
See R-MSN-8. A territory is defended only by stationed missions. With no garrison, any attacker wins at once.
- **Status:** ADAPTED (MSN-009).

### R-TER-3 — Yield
- **Formula.** For an owned tile whose terrain has a `resource`:
  `pending = floor( min(hours_since(yield_collected_at) × ws, P.territoryYieldCapHours) × P.territoryYieldPerHour )` units of that resource.
- `Collect` grants `pending` to a chosen own city (caps apply) and sets `yield_collected_at = now`. The fractional remainder is lost.
- Terrains without a resource yield nothing.
- **Example (Ashen Crowns, 25/h, cap 24 h).** After 7.5 h: **187** units. The maximum is **600**.
- **Status:** PROPOSED. The VERIFIED outer-domain production (ECO-011) is simplified to a flat yield.

### R-TER-4 — Capture (attack on another player's territory)
- **Dispatch.** PvP gate R-PVP-1.
- **Arrival.**
  1. If the parties became friendly in flight, the mission returns (`target_became_friendly`).
  2. Faction penalties apply.
  3. Battle against all stationed groups.
- **On victory with `capture = true` and `territories < cap`:**
  - ownership passes to the attacker (`owner_since = now`);
  - `yield_collected_at` is **not** reset, so **the captor inherits the uncollected yield**;
  - the attacking mission becomes the new garrison (`stationed`).
- **On victory without capture, or at the cap:** the attacker returns (`captureFailed: territory_limit`). The former owner keeps the tile, now ungarrisoned.
- **Status:** ADAPTED (MSN-009 domain seizure).

### R-TER-5 — Abandon
- The owner may release a territory: ownership and yield timestamps are cleared.
- The owner's own stationed missions there are sent home (travel = outbound duration). Uncollected yield is lost.

### R-CITY-1 — City cap
- `cityCap = min(P.maxCities, 1 + floor(maxLevel(P.cityCapBuilding) / P.citiesPerCapLevels))`, or 1 if no building is configured.
- **Example (Ashen Crowns).** Keep 4 gives 1 city, Keep 5 gives **2**, Keep 25 gives 6 (so `maxCities 9` is unreachable; see D-16).
- **Status:** ADAPTED (WLD-003: up to 1 + 8 cities, VERIFIED. Gating by building level is PROPOSED).

### R-CITY-2 — Founding
See R-MSN-7. The capital is the city created at join (`is_capital`). Rewards and refunds fall back to the capital when the target city no longer exists.

### R-CITY-3 — Spawn
- A new player's capital is placed on a free tile close to the map centre with jitter: passable, no neutral (even a cleared one), no owner, and Chebyshev distance ≥ 3 from every city.
- The server walks square rings outward from the centre. It picks a random legal tile of the first ring that has one. Rings inside the last used ring are searched only when the outer rings are full.
- Joins of one world run one at a time. A per-world advisory lock keeps two simultaneous joins from choosing neighbouring tiles.
- Join throughput is limited by this lock: about 70 joins per second on the load-test host (T-021).
- Initial state comes from `start`: resources, buildings with their levels, units, population.
- **Status:** PROPOSED.

---

## 14. Heroes

### R-HERO-1 — Hiring and cap
- Hiring requires `P.heroBuilding ≥ 1` in the hiring city.
- `heroCap = min(P.maxHeroes, maxLevel(P.heroBuilding) × P.heroesPerBuildingLevel)`.
- Hire cost: `P.heroHireCost × (n + 1)`, where `n` = heroes currently owned.
- **Example (Ashen Crowns).** The 1st hero costs gold 300 + grain 300; the 3rd costs gold 900 + grain 900. Warhall 3 allows 3 heroes.
- **Status:** VERIFIED (HERO-002: max 7 heroes). Building gating is ADAPTED (HERO-001 Tavern).

### R-HERO-2 — Experience and level
- **Formula.** `xpTotal(L) = Σ_{k=1}^{L−1} round(xpBase × xpGrowth^(k−1))`, with `xpTotal(1) = 0`.
  `level = max{L ≤ maxLevel : xp ≥ xpTotal(L)}`.
- XP gains: R-CMB-9 (`floor(scoreLost × xpPerScore × hero_xp modifiers)` + neutral `xp` on a PvE win). Negative gains are ignored.
- XP keeps accumulating after `maxLevel`, but the level is capped.
- **Example (Ashen Crowns: base 200, growth 1.35).** L2 needs 200, L3 470, L5 1327, L10 7939. Clearing Wolf Den L3 (45 wolves, score 1 each, `xpPerScore 0.5`): `floor(22.5)` = 22 + 90 = **112 XP**. A Sage (+15%) gets `floor(25.875)` = 25 + 90 = 115.
- **Status:** VERIFIED (HERO-008: XP from the value of killed troops). The curve is PROPOSED (research: UNKNOWN).

### R-HERO-3 — Hero modifiers
`M_hero = class.modifiers × 1 + class.perLevel × (level − 1) + Σ_skills skill.modifiers × skillLevel + Σ_equipped item.modifiers`. They apply only to the hero's mission or home-city defence (R-MOD-1).
- **Example.** A Marshal at L5 has `attack.* +9%`.

### R-HERO-4 — Skill points
- `points = 1 + (level − 1) × skillPointsPerLevel − Σ skill levels`.
- Learning costs 1 point. A skill is limited by its `maxLevel` and its `classes` restriction.
- **Status:** ADAPTED (HERO-005: 2 skill upgrades per level, VERIFIED (med); Ashen Crowns uses 2).

### R-HERO-5 — Items
- An item is equipped on an **idle** hero into its declared slot. A slot holds at most one item, and the previous occupant is unequipped.
- Items are owned by the player and can be moved between heroes while they are idle.
- An item whose definition is gone after a pack migration cannot be equipped.

### R-HERO-6 — Wounds
- A wounded hero (`woundedUntil > now`) cannot be sent on missions and does not defend.
- Wounds last `P.heroWoundHours` real hours.
- **Status:** ADAPTED (CMB-011: stun).

---

## 15. Market and NPC exchange

### R-MKT-1 — Tradability
Both resources must be `tradable` and not premium, and must differ.

### R-MKT-2 — Placing and cancelling orders (escrow)
- Placing an order `(sell r₁ × a₁, buy r₂ × a₂)`, with `a ∈ [1, 10⁹]`, escrows `a₁` of `r₁` from the city (ledger `market_escrow`).
- Cancelling an open order returns the escrow **in full** to that city (caps apply).
- **Status:** VERIFIED (ECO-017/SOC-020: player market exists). Escrow is PROPOSED.

### R-MKT-3 — Open-order cap
At most `P.marketMaxOpenOrders` open orders per player.

### R-MKT-4 — Accepting an order
- Only whole orders can be filled (no partial fills in v1). A player cannot accept their own order. The order row is locked, so concurrent accepts of the same order cannot both succeed.
- In one transaction, the buyer:
  1. pays `a₂` of `r₂`;
  2. receives `a₁` of `r₁` (caps apply);
  3. marks the order `filled`;
  4. schedules a **delivery** event for the seller.
- **Status:** ADAPTED (ECO-017: the original allowed partial lots; deferred).

### R-MKT-5 — Fee and delivery
- **Formulas.**
  - `fee% = (P.marketFeePct + flat) × (1 + pct(market_fee)/100)`, using the **buyer's** player-wide modifiers (see D-1).
  - `proceeds = a₂ − ceil(a₂ × fee% / 100)`.
- Proceeds are delivered to the seller's order city (or capital) by a `delivery` event due immediately. It is applied the next time the seller is settled (worker or command), in the seller's own lock.
- **Example (5%).** An order buying 1000 gold pays the seller **950**. With an 8% fee reduction, `fee% = 5 × 0.92 = 4.6000000000000005` in float64, so `ceil(46.00000000000001)` = **47**, and the seller gets **953**, not 954. There is no epsilon guard here, unlike R-TIME-3 (see D-23).
- **Status:** VERIFIED (ECO-017: market tax exists). The value is PROPOSED.

### R-MKT-6 — Price band
- **Formula.** When both resources have `exchangeRates`: `ratio = (a₂ × rate₂) / (a₁ × rate₁)` must be in `[0.1, 10]`.
- **Example.** Selling 1000 timber (rate 1) for gold (rate 3): `a₂ ∈ [33.3, 3333.3]`, so between **34 and 3333 gold**.
- **Status:** ADAPTED (ECO-002/ECO-022: the original had per-resource price bands).

### R-MKT-7 — NPC exchange
- **Formula.** `received = floor(amount × rate_from / rate_to × (1 − P.exchangeFeePct/100))`, `amount ∈ [1, 10⁹]`. The result must be ≥ 1. It is instant, with no counterparty. Received goods are subject to caps.
- **Example (Ashen Crowns, 30%).** 1000 timber gives **233 gold**. 100 gold gives **210 timber**. A round trip loses 51%.
- **Status:** VERIFIED (ECO-018: an exchange tab "at worse rates"). Rates are PROPOSED.

---

## 16. Alliances and social

### R-SOC-1 — Founding
- A player not in an alliance pays `P.allianceCreateCost` from the capital.
- Name: 3–24 characters. Tag: 2–5 alphanumerics. Both are unique (case-insensitive) among active alliances of the world.
- The founder becomes `leader`.

### R-SOC-2 — Membership
- Leaders and officers invite players by name. A player can have one pending invite per alliance.
- The invitee accepts under the alliance row lock. The member cap `P.allianceMaxMembers` is checked there, so concurrent accepts cannot exceed it.
- Joining requires not being in another alliance.
- **Status:** VERIFIED (SOC-010 clans). The size value is PROPOSED.

### R-SOC-3 — Friendliness (blocks attacks)
- `friendly(a, b)` holds iff any of:
  - `a = b`;
  - same alliance;
  - `a`'s alliance has a relation `ally` or `nap` **toward** `b`'s alliance (directional: the declaring side is bound).
- Friendly targets cannot be attacked (R-PVP-1).
- **Status:** VERIFIED (SOC-013: alliances/diplomacy).

### R-SOC-4 — Alliance for reinforcement
`allied(a, b)`: `a = b`, the same alliance, or `a`'s alliance has `ally` toward `b`'s alliance. `nap` does not allow reinforcement.

### R-SOC-5 — Diplomacy
- Only the leader sets relations toward another active alliance of the world: `ally`, `nap`, `war` or `neutral` (delete).
- `war` has no mechanical effect in v1 (informational).

### R-SOC-6 — Roles, leaving and disbanding
- Roles: `leader`, `officer`, `member`.
- The leader can promote or demote any member. Promoting someone to leader demotes the caller to officer.
- Kick: the leader can kick anyone. An officer can kick plain members only.
- A leader must hand over leadership before leaving, unless they are the last member.
- When the last member leaves, the alliance is **disbanded** and its relations (in both directions) are deleted.

### R-SOC-7 — Messages
- Private message (by name) or alliance broadcast. Subject ≤ 120 characters, body 1–4000 characters.
- At most 60 messages per sender per real hour (`rate_limited`).
- A recipient can block senders. Blocked senders' messages are refused and hidden.
- Received messages can be reported for moderation.

### R-SOC-8 — Reinforcement cap
At most `P.maxReinforcingPlayers` distinct foreign players may be stationed in one city. The cap is enforced at arrival (R-MSN-6).
- **Status:** VERIFIED (SOC-015: "up to 3 players plus the owner").

---

## 17. Quests

### R-QST-1 — Status machine
`locked → active → complete → claimed`:
- `locked`: some quest in `requires` is not **claimed**.
- `active`: progress < goal.
- `complete`: progress ≥ goal.
- `claimed`: a row exists in `quest_progress`.

Quests are optional.
- **Status:** VERIFIED (QST-001: claim the reward from the Counsellor after completion; QST-004: chains).

### R-QST-2 — Progress sources (closed set)
| Goal | Progress source |
|---|---|
| `building_level` | max level over own cities |
| `tech_level` | tech level |
| `units_trained` (optional `id`) | cumulative stat |
| `missions_won`, `neutrals_cleared` | cumulative stat, credited on mission return |
| `market_trade` | cumulative, **buyer side only** (see D-11) |
| `hero_hired` | cumulative stat |
| `scouts_sent` | counted at dispatch |
| `join_alliance` | 1 while a member |
| `cities` | current count |
| `territories` | current count |

Progress is retroactive: stats from before unlocking count.

### R-QST-3 — Claim exactly once
- Claiming requires status `complete`. The `quest_progress (player, quest)` primary key is inserted first. A second claim, even a concurrent one, fails with `already_claimed`.

### R-QST-4 — Reward delivery
- Rewards are granted to the chosen own city, or the capital, under caps (R-RES-4/5). Premium rewards are allowed.
- **Status:** VERIFIED (QST-002). The "no claim within 2 h of a battle" restriction is not implemented (PROPOSED to omit).

---

## 18. Conflict resolution between rules

### R-CON-1 — Settle before act
- Every command first applies all of the player's events due at or before `now`, then acts on the settled state.
- A command at instant `t` therefore never sees a job that finished before `t` as pending, nor resources not yet produced.

### R-CON-2 — Same-instant events
- Events of one player due at the same instant run in **scheduling order** (event id).
- **Two arrivals at the same city at the same `due_at`** (attacks, transports, reinforcements, from any players) are all the defender's events, so the one **scheduled first resolves first**. A later attacker fights the survivors and plunders what is left.
- **Job completion versus arrival at the same instant:** whichever was scheduled first applies first. A Walls level finishing at the same instant as an attack helps only if its job was queued before the attack was launched.
- **Arrival versus recall:**
  - **Tile targets:** the arrival event belongs to the mission owner. A recall settles the owner first, so if the arrival is already due it resolves first and the recall fails (`not_recallable`).
  - **City targets:** the arrival event belongs to the defender, and the attacker's settlement does not process it. A recall issued after `arrive_at` but before the defender has been settled (worker latency) still succeeds, and the battle never happens (see D-22).
  - **Normative intent:** a recall is valid only while `now < arrive_at`.

### R-CON-3 — Economy versus event at the same instant
- Resources accrue up to `due_at` with the **old** rates, then the event applies.
- If the event instant is a tick boundary, the tick (R-TIME-4) runs **before** the event.

### R-CON-4 — Cross-player effects
- A battle at a city is resolved in the defender's transaction. The attacker's units, cargo and hero results are written to the mission row and applied by the attacker's own `mission_return` event.
- Market proceeds use a `delivery` event.
- No transaction mutates two players' cities.

### R-CON-5 — Lock contention
- The same mission, order, tile or alliance row is serialised by row locks.
- The first transaction wins. The second sees the new status and fails with a domain error: `order_closed`, `not_recallable`, `already_claimed` or `alliance_full`.

### R-CON-6 — Validation at dispatch versus arrival
- Spatial and capacity conditions are re-checked at arrival: target changed, tile occupied, territory cap, city cap, settler present, reinforcement cap.
- PvP diplomatic conditions are re-checked only for territories (see D-7).

---

## 19. Anti-exploit rules

| ID | Rule |
|---|---|
| R-AX-1 | Server authority: clients send intents. All costs, times, outcomes and RNG are computed server-side. |
| R-AX-2 | Mutating HTTP commands require an `Idempotency-Key` (8–80 chars). A replay returns the stored response. Reusing a key with a different body returns 422. Keys expire after 24 h. |
| R-AX-3 | All spends happen under the player lock with pre-checks (R-RES-2). Over-spend panics, and concurrency tests assert no double spend. |
| R-AX-4 | Premium resources are not tradable, lootable or transportable (R-RES-1). |
| R-AX-5 | Market: price band 10× of reference (R-MKT-6) against gift-trading between accounts, open-order cap, no self-accept, row-locked fills. |
| R-AX-6 | PvP: friendly-fire block, newbie protection, attack-frequency cap, optional score floor (R-PVP-1). |
| R-AX-7 | Recall costs the elapsed travel time (no teleport). Only the last queued job can be cancelled, at an 80% refund (no free re-ordering or banking). |
| R-AX-8 | Defenders see incoming missions but not their composition (R-SCT-4). A failed scouting is always detected and the scouts are lost. |
| R-AX-9 | Bounds on every numeric input: units ≤ 10⁹ per mission, training ≤ 100 000 per batch, trade amounts ≤ 10⁹. Cost scaling saturates at 10¹⁵. Pack amounts ≤ 10¹². |
| R-AX-10 | Messages: rate limit of 60/h, block list, reporting. API rate limits per account and per IP. |
| R-AX-11 | Failing events are parked after 5 worker attempts and cannot block a world (R-TIME-6). Admin grants are audited. |
| R-AX-12 | Sinks to counter inflation: NPC exchange fee, market fee, cancellation tax, overflow decay, army upkeep, alliance cost, escalating hero hire cost. |

Known exploit risks are listed in Appendix D (D-9, D-10).

---

## Appendix A — Index of rule IDs referenced from code

| Code reference | Location | Defined here |
|---|---|---|
| GAME_RULES.md §2 | `rules/modifiers.go` (Mods) | §2, R-MOD-2 |
| R-TIME-1 | `rules/simulate.go` (AdvanceCity) | §1 |
| R-TIME-3 | `rules/modifiers.go` (MinTimeFactor) | §1 |
| §3 | `rules/economy.go` (ComputeRates) | §3 (R-ECO-1..4, R-RES-3, R-RES-7) |
| R-RES-4 | `rules/economy.go` (Grant) | §3 |
| R-RES-5 | `rules/economy.go` (TickStep) | §3 |
| R-POP-1..4 | `rules/economy.go` (TickStep) | §4 |
| R-ARMY-UPKEEP-2 | `rules/economy.go` (TickStep) | §5 |
| R-ARMY-2 | `game/city.go` (Train) | §5 |
| R-BLD-1..6 | `game/city.go` (Build) | §6 |
| R-CMB-1 | `rules/combat.go` (BuildStack) | §10 |
| R-CMB-2..6 | `rules/combat.go` (Resolve) | §10 |
| R-CMB-3 | `rules/combat.go` (damage) | §10 |
| R-CMB-7 | `rules/combat.go` (Loot) | §10 |
| R-SCT-1..3 | `rules/combat.go` (Scout) | §11 |
| R-SCT-4 | `game/missions.go` (Missions) | §11 |
| R-MOV-1 | `rules/world.go` (TravelSeconds) | §8 |
| R-MSN-1..12 | `game/missions.go` (SendMission) | §8 |
| R-PVP-2 | `game/missions.go` (checkPvP) | §9 |
| R-SOC-1..8 | `game/social.go` | §16 |
| R-SOC-3 | `game/missions.go` (friendly) | §16 |
| R-QST-1..4 | `game/economy.go` | §17 |
| R-MKT-1..7 | `game/economy.go` | §15 |
| R-MKT-6 | `game/economy.go` (PlaceOrder) | §15 |
| R-TER-1..5 | `game/economy.go` | §13 |
| R-HERO-3 | `game/heroes.go` (heroMods) | §14 |

## Appendix B — Research traceability summary

| Research item | Status in research | Engine rule | Our status |
|---|---|---|---|
| BLD-000 (≤80% time reduction) | VERIFIED high | R-TIME-3 | VERIFIED |
| ECO-012 (production halts at limit) | VERIFIED high | R-RES-4 (hard-cap mode) | VERIFIED |
| ECO-013 (1%/h surplus loss) | VERIFIED high | R-RES-5 (soft-cap mode) | VERIFIED |
| ECO-014 (cache) | VERIFIED med | R-RES-7 | VERIFIED concept, values PROPOSED |
| ECO-015 (gold salary, 0.5%/h desertion, double on missions) | VERIFIED med | R-ARMY-UPKEEP-1/2 | VERIFIED |
| POP-002 (0.04 × idle per hour) | VERIFIED high | R-POP-2 `idle` mode | VERIFIED (mode), value per pack |
| POP-004 (staffing) | VERIFIED high; scaling INFERRED | R-ECO-2 | ADAPTED |
| POP-006 (gradual flight) | VERIFIED med | R-POP-1 | not implemented (D-19) |
| CMB-001 (20 rounds, timeout = attacker loss) | VERIFIED high | R-CMB-2/6 | VERIFIED |
| CMB-003 (defence %, proportional spread, residual damage) | VERIFIED med | R-CMB-3/4 | VERIFIED except overflow (D-14) |
| CMB-004 (pursuer/victim) | VERIFIED high | class bonuses | ADAPTED |
| CMB-005/006 (retreat, both retreat means defender wins) | VERIFIED high/med | R-CMB-5/6 | VERIFIED |
| CMB-007 (multiple armies per side) | VERIFIED high | R-CMB-8 | VERIFIED |
| FAC-002 (−50% race, −25% faction) | VERIFIED high | R-PVP-4 | VERIFIED (no stacking) |
| MSN-003 (15-minute minimum, round-trip pay) | VERIFIED med | R-MOV-1, R-ARMY-UPKEEP-1 | VERIFIED |
| MSN-004 (2 attacks per target per 24 h) | VERIFIED med | R-PVP-1(5) | VERIFIED |
| MSN-005 (recall) | VERIFIED med | R-MSN-12 | ADAPTED |
| MSN-012/SOC-015 (assistance, 3 helpers, 2× speed) | VERIFIED med | R-MSN-6, R-SOC-8 | cap VERIFIED, speed missing (D-15) |
| HERO-002 (max 7 heroes) | VERIFIED high | R-HERO-1 | VERIFIED |
| HERO-005 (2 skill points per level) | VERIFIED med | R-HERO-4 | VERIFIED (Ashen Crowns) |
| HERO-008 (XP from slain troops' value) | VERIFIED med | R-HERO-2 | ADAPTED (score instead of gold) |
| WLD-003 (1 + 8 cities) | VERIFIED high | R-CITY-1 | VERIFIED cap; see D-16 |
| QST-001/002 (claim after completion) | VERIFIED | R-QST-1..4 | VERIFIED |
| ECO-017/018 (market, exchange at worse rates) | VERIFIED | R-MKT | VERIFIED concept, values PROPOSED |

---

## Appendix D — Discrepancies between code comments, behaviour and this spec (found 2026-10-31)

No code was changed. Each item needs a decision from the owner: fix the code or accept the behaviour and keep this spec.

| ID | Severity | Finding |
|---|---|---|
| D-1 | High (dead feature) | The Market building's `market_fee −4%/level` never applies. `AcceptOrder` computes the fee with `c.mods` (race + techs only), and building modifiers live only in `cityMods`. The fee is also computed from the **buyer's** modifiers, though it reduces the **seller's** proceeds. |
| D-2 | Medium | `params.territoryBonusPct` (default 5, documented "production bonus per owned territory") is validated but never read. The `rules.EconInput.Mods` comment says "race + techs (+ hero/territory)", but no hero or territory modifier ever enters the economy. |
| D-3 | Medium | Attacker modifiers at arrival omit the departure city's building modifiers: `attMods = owner race+techs + hero`. Dispatch uses `cityMods` (with buildings). So Market `carry +3%/lvl` raises transport capacity but not loot capacity. |
| D-4 | Medium | `Neutral.capturable` is never read. Any cleared neutral tile (including `barrow`, `capturable: false`) can be occupied. When the neutral respawns on an owned territory, the owner and the neutral coexist, and attacks there fight the neutral, not the owner. |
| D-5 | Low | The `reinforce` on own territory branch in `missionArrive` is unreachable: `SendMission` rejects `reinforce` without a target city. `occupy` does the job. |
| D-6 | Medium (economy) | Away upkeep uses `upkeep_units` fixed at departure. Units killed on a mission keep costing `upkeep × missionUpkeepMult` until the mission is done (return trip). Stationed reinforcements cost the sender indefinitely (intended, but undocumented in code). |
| D-7 | Medium | `battleCity` and `scoutCity` do not re-check friendliness or newbie protection at arrival. An alliance or NAP formed while an attack is in flight does not stop it. `battleTerritory` does re-check friendliness. Inconsistent. |
| D-8 | Medium (balance) | PvE rewards are added to cargo regardless of carrying capacity. Player loot is carry-capped. Premium `starglass` rewards travel as cargo although premium cargo is forbidden for transport. |
| D-9 | Medium (exploit) | Score counts only units at home. Sending the army away lowers the attacker's own score, which would bypass `attackScoreFloorPct` (currently 0 in both packs). |
| D-10 | Medium (exploit) | Market escrow is neither stored in the city nor lootable, and cancelling refunds 100%. Players can hide stock from raids in overpriced (≤ 10×) open orders. Suggest a cancellation fee or making escrow lootable. |
| D-11 | Low | The `market_trade` stat is credited only to the buyer. A player who only sells never completes `t11_trade`. |
| D-12 | Low | A settled city gets every `start.buildings` entry at level 1 (start levels ignored via `_ = l`) and `start.population / 2`. Its cargo is set as stock without storage caps. The code has no comment on this. Now specified in R-MSN-7. |
| D-13 | Low | `heroWoundHours`, `newbieProtectionHours` and the 24 h attack window are real time and are not divided by `ws`, unlike respawn and territory yield. In speed worlds, protection lasts relatively much longer. |
| D-14 | Spec deviation | Overflow damage on a destroyed stack is discarded. Research CMB-003 (VERIFIED) says it is redistributed to remaining groups. The `damage` comment matches the code. |
| D-15 | Spec deviation | MSN-012 (VERIFIED): assistance travels at 2× attack speed. Not implemented. |
| D-16 | Balance | Ashen Crowns `maxCities 9` is unreachable: Keep max 25 / `citiesPerCapLevels 5` gives at most 6 cities. Void Frontier: Command max 20 gives at most 5. |
| D-17 | Balance | Ashen Crowns hero skill points (59 at L30) exceed learnable skill levels (Marshal 30, Warden 35, Sage 40). Points saturate around L15–L20. |
| D-18 | Low | Pack `modes` (`peaceful`, `speed`) are validated but have no engine effect. `params.premiumResource` is never read. Premium handling relies on `resource.kind`. |
| D-19 | Spec deviation | Population above housing is removed at the next tick (hard clamp). POP-006 (VERIFIED) describes gradual flight. |
| D-20 | Low (comments) | `pack.Unit.Upkeep` is commented "food per hour" but is paid in `upkeepResource` (gold in Ashen Crowns). `Params.PopGrowthPctPerHour` is commented "of free housing per hour" but is "of idle" in `idle` mode. `speedOf(units, mods)` ignores `mods` (harmless: the speed % is applied in `TravelSeconds`). |
| D-21 | Low (ops) | A handler error blocks all commands of the affected player (`lockPlayer` fails) until the worker parks the event after 5 attempts. Only worker attempts are counted. |
| D-22 | Medium (exploit) | `Recall` does not check `now < arrive_at`. For city targets, the arrival event is the defender's, so an attacker can recall a mission whose arrival is already due but not yet processed. This avoids a lost battle within the worker latency window. The return time (`now − depart`) is then longer than the outbound trip. |
| D-23 | Low | `ceil` without an epsilon guard in the market fee (and in travel and overflow decay) can round float noise up. Example: a 4.6% fee on 1000 gives 47, not 46. `DurationAt` uses a 10⁻⁹ guard. The others should too, or use integer arithmetic. |
| D-24 | Low (consistency) | `modifiers.go` says the order is always `(base + Σflat) × (1 + Σpct)`, but `ComputeRates` applies production as `base × (1 + Σpct) + Σflat`. No shipped pack uses flat production modifiers yet, so the decision is open: keep the exception (specified in R-ECO-2) or align it. |

## Appendix E — Resolution of Appendix D discrepancies (2026-10-31)

| # | Discrepancy | Resolution |
|---|---|---|
| Market fee ignored building modifiers | **Fixed**: fee uses the accepting city's modifiers (`cityMods`), so Market Hall levels reduce it. Ceil has a 1e-9 float guard. |
| Attacker modifiers skipped home-city buildings at arrival | **Fixed**: the departure city's building modifiers are merged into attacker modifiers at arrival. |
| City attack landed after players became allies | **Fixed**: `battleCity` re-checks friendliness and returns the army (`target_became_friendly`). |
| Recall after due time (exploit) | **Fixed**: outbound missions with `now ≥ arrive_at` return `409 already_arrived`; regression test added. |
| Production applied flat after pct | **Fixed**: `(base + Σflat) × (1 + Σpct/100)` like every other modifier; unit test added. |
| `Neutral.capturable` unused, neutrals respawn on territories | **Fixed**: owned territories are pacified (`neutralActive` false when owned); non-capturable sites (e.g. barrows) cannot be occupied (`not_capturable`). |
| Ashen Crowns `maxCities 9` unreachable | **Fixed in pack 1.1.0**: `citiesPerCapLevels 3` → 1 + ⌊25/3⌋ = 9. |
| Starglass had no sink | **Fixed in pack 1.1.0**: specialisation Paths (15), settlers (10) and hero hire (5) cost starglass. |
| `territoryBonusPct` unused | Open: parameter retained for format compatibility; territories use yield (D-013). To be removed in format v2. |
| Upkeep of dead units abroad until return | By design (D-010). |
| Escrowed market orders hide stock from raids; 100 % cancel refund | Open balance item (B-flag); candidate fix: cancel fee or lootable escrow. |
| Neutral rewards ignore carry capacity | By design (PROPOSED): treasure is collected regardless of carts; documented. |
