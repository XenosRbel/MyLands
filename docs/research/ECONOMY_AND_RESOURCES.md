# Economy and Resources: research notes (A1)

Access date for all evidence: 2026-10-31. Source IDs refer to `_sources_A1.md`.
**Evidence caveat:** every page was read only through search-engine extracts, because direct fetching was blocked in this session. Numbers are recorded only where a snippet stated them. Otherwise they are UNKNOWN.

## Resource summary (as found)

| Resource | Type | Producers (per wiki) | Main sinks |
|---|---|---|---|
| Wood | basic | City Main Building, Woodcutter Cabin | Construction, units, sciences (exact costs UNKNOWN) |
| Crops | basic (food-like) | Windmill | Units, construction (exact use UNKNOWN; no evidence that population eats) |
| Stone | basic | Stone Quarry, Stone Mine Main Building, sometimes Pit Main Building | Construction (scarce early per fan guide) |
| Iron | basic | Iron Plant, Iron Mine Main Building, sometimes Pit Main Building | Units, construction |
| Gold | currency/basic | Trading Post, Gold Plant, Gold Mine Main Building, sometimes Pit Main Building | Army salary, market purchases, hero hiring, BG auction |
| Black Gems (BG) | premium / special | Salt Lake and Grail Main Buildings, Ruins (monsters), Auction (for Gold), real money | Ancient Magic / local and global services, instant completion |
| Science Points | research currency | Alchemist Labs (all domains, pooled) | Sciences, "charged" sciences |
| Population (idle, Workers, Builders, Alchemists) | human resource | Main Building, Cottages, idle reproduction | Staffing producers and construction, unit training (units cost population) |
| Mana | hero resource | UNKNOWN (mana bottles exist as loot) | Spells (Wizards pay 50%) |
| Artifacts / crystals / runes | items | Ruins, quests, tournaments, Trader, recipes | Building and artifact improvement (Alchemy) |

---

### ECO-001 — Wood
- Description: Basic construction resource.
- Game purpose: Core build input.
- Inputs: Producing buildings staffed with Workers.
- Rules & constraints: Produced in the Main Building of a City and in the Woodcutter Cabin. Elves have the greatest wood rate. Drow and Undead also get wood bonuses.
- Output: Wood per hour into the local domain store.
- Related systems: ECO-010 modifiers, ECO-012 storage, BLD-010.
- Evidence: SRC-A1-04 (uk.mlgame.org Resources): "Wood is produced in: Main Building of City, Woodcutter Cabin." SRC-A1-43 (fan blog) agrees.
- Status: VERIFIED (high)
- Test plan: A city with only a Main Building still gains wood > 0 per hour. Adding a staffed Woodcutter Cabin increases the hourly wood.

### ECO-002 — Crops
- Description: Agricultural resource, produced by the Windmill only.
- Game purpose: Basic input. Its exact role as food is UNKNOWN.
- Inputs: A Windmill on a crop deposit, staffed.
- Rules & constraints: Only the Windmill produces crops. Crop deposits are finite (ECO-009). Wizards have the best crop rate. Market price range is 8–20 Gold (avg 14) per the Rating page.
- Output: Crops per hour.
- Related systems: ECO-009, ECO-022.
- Evidence: SRC-A1-04: "Crops are produced in: Windmill." SRC-A1-21: "Crops possible cost is from 8 to 20… (8+20)/2=14".
- Status: VERIFIED (high for the producer; med for the price band)
- Test plan: Only Windmill-type buildings yield crops. The market rejects crop prices outside [8, 20] gold per unit (if we adopt price bands).

### ECO-003 — Stone
- Description: Basic construction resource.
- Game purpose: Construction input, scarce early.
- Inputs: A Stone Quarry on a stone deposit. Stone Mine (outer domain). Pit Main Building (sometimes).
- Rules & constraints: Demons mine stone faster. The deposit is finite. A fan guide says stone is among the most lacking resources early.
- Output: Stone per hour.
- Related systems: ECO-009, ECO-011.
- Evidence: SRC-A1-04 (uk.mlgame.org): Stone Quarries, Stone Mine Main Building, sometimes Pit. SRC-A1-42 (RU fan): stone is "one of the most valuable and lacking resources" at start (paraphrased translation).
- Status: VERIFIED (high)
- Test plan: As ECO-001. Deposit depletion test in ECO-009.

### ECO-004 — Iron
- Description: Basic military and construction resource.
- Game purpose: Army input.
- Inputs: Iron Plant, Iron Mine Main Building, sometimes Pit Main Building.
- Rules & constraints: Drow and Wizards have iron bonuses. The deposit is finite.
- Output: Iron per hour.
- Related systems: ECO-009, ECO-011.
- Evidence: SRC-A1-04: "Iron Plants, the Main Building of an Iron Mine, and sometimes a Pit's Main Building."
- Status: VERIFIED (high)
- Test plan: As ECO-001.

### ECO-005 — Gold
- Description: Main soft currency, also counted as a resource.
- Game purpose: Pays army salary and trade. Funds BG auctions and hero hiring.
- Inputs: Trading Post (boosted by idle population), Gold Plant (on a gold deposit), Gold Mine Main Building, sometimes Pit Main Building. Chimeras extract gold underground.
- Rules & constraints: Knights have the highest gold efficiency. Idle population "increases the Gold income via the Trading Posts" (formula UNKNOWN). The Trader specialisation raises Trading Post tax by +0.01/+0.02/+0.03 per level, which suggests Trading Post income is tax × idle people (INFERRED).
- Output: Gold per hour.
- Related systems: POP-003 idle population, ECO-015 salary, ECO-017 market, TECH-020 Trader spec.
- Evidence: SRC-A1-04: Gold produced in "Trading Post, Gold Plant, the Main Building of a Gold Mine, and sometimes… Pit". SRC-A1-08: "Idle Population increases the Gold income via the Trading Posts." SRC-A1-12: Trader "Trading Post tax… 0.01 to 0.03".
- Status: VERIFIED (high for producers; Trading Post formula INFERRED, low)
- Test plan: Trading Post income rises monotonically with idle population. With 0 idle people, income is 0 or base (to be decided).

### ECO-006 — Black Gems (premium "special resource")
- Description: Premium currency, called "a special resource" usable for in-game services and enhancements.
- Game purpose: Monetisation and endgame objective (title theme: "hunting" for gems).
- Inputs: (a) controlling a Salt Lake or Grail, whose Main Building produces BG; (b) robbing Monsters in Ruins (all monsters must be slain); (c) buying at the Auction for Gold; (d) real-money purchase; (e) gifts from other players via a "Present BG" service.
- Rules & constraints: BG cannot be taken from other players. Players cannot sell BG at the auction. BG is bound to the player account (Wikipedia). A portal wallet and a server wallet exist (SRC-A1-18). Grail mining speed is 200 BG per hour. The Salt Lake rate is UNKNOWN.
- Output: A BG balance spendable on services (ECO-021).
- Related systems: ECO-011, ECO-019, ECO-021, Ruins (A2/combat).
- Evidence: SRC-A1-07 (BlackPearl): "special resource that can be used to purchase in-game services and enhancements"; "steal the Black Gems from Monsters who live in Ruins"; Grail "200 BG per hour" (SRC-A1-15). SRC-A1-17: BG cannot be sold. SRC-A1-36: "bound to player's account".
- Status: VERIFIED (high for the sources; med for the Grail rate)
- Test plan: BG is not plunderable in any raid. The auction offers no "sell BG" path. A Grail held for 1 h credits 200 BG (if adopted).

### ECO-007 — Science Points (SP)
- Description: Research currency produced by Alchemist Labs.
- Game purpose: Paces the tech tree.
- Inputs: Labs in all domains, staffed by Alchemists.
- Rules & constraints: SP are pooled across all domains. More labs mean faster research. Village and Ancient Magic bonuses apply per lab: multiply, round, then sum. Surplus SP beyond what a science needs is lost. Some sciences need resource payment on completion.
- Output: Science progress.
- Related systems: TECH-001..
- Evidence: SRC-A1-10: "Science Points are produced by Alchemist Labs from all domains"; "multiplied by bonus first, then rounded off, and then – all the values are summed"; "the surplus will be lost".
- Status: VERIFIED (high)
- Test plan: Two labs with 10.4 effective output each give 20 (rounded per lab), not 21. Excess SP at completion is discarded.

### ECO-008 — Mana (hero resource)
- Description: Heroes have mana for spells. Mana bottles exist as loot.
- Game purpose: Spell economy (combat scope).
- Inputs: UNKNOWN (regeneration source not found).
- Rules & constraints: Wizards spend 50% mana on spells.
- Output: Spell casting.
- Related systems: RACE-006, heroes (A2).
- Evidence: SRC-A1-05: "expend half as much mana on spells". SRC-A1-24: "mana and experience bottles" among loot. SRC-A1-25: each hero has "some Mana".
- Status: VERIFIED that it exists (med). Mechanics UNKNOWN.
- Test plan: Spell cost deducts mana. Wizard cost is ×0.5.

### ECO-009 — Finite resource deposits and depletion
- Description: Stone, Iron, Gold and Crop deposits contain a limited amount of resource.
- Game purpose: Forces expansion, relocation and recovery decisions.
- Inputs: A deposit cell and a producing building on it.
- Rules & constraints: After full depletion, the producer's rate drops to **1 per hour**. To restore it, demolish the producer and build a Chalice of Abundance, which recovers the deposit at some rate (UNKNOWN) and self-destroys when done. Wood is not listed as deposit-bound (the Main Building produces wood). Deposit sizes are UNKNOWN.
- Output: A deposit remaining-amount counter.
- Related systems: BLD-016 Chalice of Abundance.
- Evidence: SRC-A1-04: "each Stone, Iron, Gold or a Crop deposit contains a limited amount of resource"; "rate downfall to 1 per hour". SRC-A1-22: "It's necessary to demolish the Producing Building and construct a Chalice of Abundance… After the full recovery of a deposit, the Chalice self-destroys." SRC-A1-18 repeats this.
- Status: VERIFIED (high)
- Test plan: Drain a deposit to 0 and assert the rate is 1/h. Build a recovery structure and assert the deposit refills and the structure self-removes at full.

### ECO-010 — Production modifier stack
- Description: Final production = base (race-specific, building level) × modifiers.
- Game purpose: Spatial and strategic optimisation.
- Inputs: Race base rate, building level, workers filled, terrain, location (village/pit), Ancient Magic service, specialisations.
- Rules & constraints:
  - Terrain bonus or penalty per square (values seen in a table range from -50% to +200%).
  - Village: +30% vs inside the city wall.
  - Pit: +100% for all producing buildings.
  - Ancient Magic (BG service): +25% per resource type, all domains.
  - Race: baked into base rates.
  - For science labs, bonuses apply per lab, then rounding (ECO-007).
  - Whether modifiers add or multiply is UNKNOWN for resources.
- Output: Hourly production.
- Related systems: RACE-*, ECO-021, BLD-*.
- Evidence: SRC-A1-04: "faster in villages by +30%"; "All Producing Buildings in Pits have a production bonus +100%"; terrain table "+70%, -50% … +200%". SRC-A1-07: "Production in Woodcutter Cabins of all domains +25% (Ancient Magic)". SRC-A1-26: "Each terrain has certain bonuses and penalties: bonus on Resources Production".
- Status: VERIFIED (high for the individual values; INFERRED for how they combine)
- Test plan: Golden-value table tests for each modifier alone, then combined. The stacking rule must be a design decision.

### ECO-011 — Outer Domains as resource sources
- Description: Non-city territories found by Explorers and seized: Stone/Gold/Iron Mines, Pits, Salt Lakes, Grails, underground "Barracks".
- Game purpose: Contested resource nodes and expansion beyond cities.
- Inputs: Exploration finds the domain. Seizure needs an army (A2). Some types need a science ("Underground domains").
- Rules & constraints:
  - Mines: no construction possible. Resources must be carried out by Merchants.
  - Pits: some free cells, with a +100% production bonus.
  - Salt Lake: hovering island with 4 cells and villages; its Main Building produces BG.
  - Grail: 4 cells, visible to all players on appearance, 200 BG/h.
  - Mines and Pits have a Tunnel; Lake and Grail have no gate or tunnel.
  - Domains have a life span. It can be extended by 30 days except for Salt Lake and Grail.
  - Outer domains can be re-seized, unlike cities.
- Output: Resource and BG income.
- Related systems: ECO-006, ECO-020, BLD-030.
- Evidence: SRC-A1-15 (OuterDomains/Mine): "it's necessary to train a Merchant"; Salt Lake "4 cells for buildings construction and villages"; Grail "can be seen by all players the moment it appears"; "Life Span of Outer Domain, except for Salt Lake and Grail, can be increased by 30 days". SRC-A1-14: "Cities cannot be seized, as opposed to Outer Domains".
- Status: VERIFIED (high)
- Test plan: A mine domain refuses build orders. Resources accrue in the mine and can be moved only by a transport unit. The domain expires at the end of its life span unless extended.

### ECO-012 — Storage per domain (Warehouse limit)
- Description: Each domain stores its own resources, up to a limit.
- Game purpose: Pushes spending and transport. Raid target.
- Inputs: Number and levels of Warehouses plus the Main Building level.
- Rules & constraints: Resources are stored in the domain where they are produced. Production halts when the limit is reached. The limit can be re-allocated between resources in the Main Building's chart, but a resource's allocation cannot go below its current stock. Knights' warehouses hold the most. Per-level capacities are UNKNOWN.
- Output: Per-resource capacity.
- Related systems: ECO-013, ECO-014, BLD-012 Warehouse.
- Evidence: SRC-A1-04: "The Warehouse limit is defined by the number and the levels of Warehouses and Main Building in sum." The production-halt and re-allocation rules appear in the same snippets.
- Status: VERIFIED (high for the rule; capacity values UNKNOWN)
- Test plan: When stock equals the cap, the next tick adds 0. Re-allocation below current stock is rejected.

### ECO-013 — Over-storage plunder and hard cap
- Description: Stock can exceed the limit (via transfers and loot), but the surplus decays.
- Game purpose: Soft cap. Prevents hoarding.
- Inputs: Stock above the limit.
- Rules & constraints: 1% of the surplus is lost per hour. Any amount of one resource above 999,999,999 is lost immediately.
- Output: Decay.
- Related systems: ECO-012.
- Evidence: SRC-A1-04: "1% of resource lacking storage space is plundered per hour"; "Stored resources can exceed the limits, but… surplus will be gradually plundered". SRC-A1-18 agrees.
- Status: VERIFIED (high). Scope ("domain" vs "city") is disputed between mirrors (med).
- Test plan: Stock = cap + 10,000, so after 1 h the surplus is 9,900 (if compounding per hour). Stock is never above 999,999,999.

### ECO-014 — Cache (theft protection)
- Description: A science that protects a fixed amount of resources from theft.
- Game purpose: Anti-farming protection for weaker players.
- Inputs: Cache science level.
- Rules & constraints: Capacity depends on the science level and is shown in the Main Building. Values UNKNOWN.
- Output: A protected stock amount.
- Related systems: raids (A2), TECH-0xx.
- Evidence: SRC-A1-14/11: "Cache capacity depends on the science level and is specified inside of the Main Building."
- Status: VERIFIED (med)
- Test plan: A raid steals at most max(0, stock − cache).

### ECO-015 — Army salary (gold upkeep) and desertion
- Description: Troops in garrisons and sanctuaries draw gold salary.
- Game purpose: Main continuous gold sink. Caps army size by economy.
- Inputs: Units and their salaries; Exchequer science; items.
- Rules & constraints: If a city cannot pay, units desert at 0.5% of unsupplied units per hour. On missions, salary is double for the round trip and depends on the speed coefficient. Maximum total reduction is about 99.74%; the Exchequer science alone gives 96%.
- Output: Gold drain per hour; desertion.
- Related systems: ECO-005, TECH Exchequer.
- Evidence: SRC-A1-20: desertion "0.5% of unsupplied units per hour"; "maximum possible reduction is about 99.74%… Exchequer science providing 96%".
- Status: VERIFIED (med)
- Test plan: With gold = 0 and N unpaid units, after 1 h the count is N × 0.995.

### ECO-016 — Costs of construction, research, units (consumption)
- Description: Buildings, upgrades, repairs, sciences and units cost resources. Units also cost population.
- Game purpose: Primary sinks.
- Inputs: Resources from the selected domain.
- Rules & constraints: Payment comes from the domain selected in a drop-down; "payment can be done from any of your domains" (for sciences). A science-confirm cost "rises with how many units of that type you already have" (unit upgrades). A server coefficient is applied to unit tables. Most numeric costs are UNKNOWN. RU fan wiki: "perfect" buildings need BG to upgrade.
- Output: Stock deduction.
- Related systems: BLD-*, TECH-*, POP-*.
- Evidence: SRC-A1-10, SRC-A1-20 ("cost tables show the resources and population needed to train each unit"), SRC-A1-42.
- Status: VERIFIED (med for structure); numbers UNKNOWN
- Test plan: An action is rejected if the selected domain lacks resources. Resources deduct only from that domain.

### ECO-017 — Market (resource trade) [resource side only]
- Description: Sell resources to other players, or buy their lots for Gold.
- Game purpose: Converts specialty surplus. Builds race interdependence.
- Inputs: Market science, Market building. Merchants (Merchant Guild) to buy; none needed to sell.
- Rules & constraints: Partial lots can be bought. Trading range depends on the Market science level ("with other Markets situated in your Kingdom or farther"). Market tax exists; the Trader specialisation lowers it by 2/5/9%. The base tax is UNKNOWN.
- Output: Gold ↔ resources.
- Related systems: ECO-018, ECO-020, TECH Market/Merchantry. Social and economic market details are covered by A2.
- Evidence: SRC-A1-16: "Resources can be sold to other players or bought from them for Gold"; "No Merchants are needed to sell resources". SRC-A1-11: Market science range. SRC-A1-12: Trader tax.
- Status: VERIFIED (high)
- Test plan: Buying without transport units fails. Selling succeeds. Partial purchase leaves the remainder listed.

### ECO-018 — Exchange tab (resource-to-resource swap)
- Description: Swap resources directly at rates worse than the player market.
- Game purpose: Liquidity fallback.
- Inputs: Market building.
- Rules & constraints: Rates UNKNOWN ("worse rates").
- Output: Resource conversion.
- Related systems: ECO-017.
- Evidence: SRC-A1-16 summary: "The Exchange tab swaps resources, but at worse rates."
- Status: VERIFIED (low; summary wording)
- Test plan: Exchange rate is below the market-average parity.

### ECO-019 — Black Gem Auction
- Description: BG sold for Gold at auction (Market "Auction" tab).
- Game purpose: Gold sink. Lets non-payers obtain premium currency.
- Inputs: Gold. Main Rating ≥ 10,000,000.
- Rules & constraints: Bid ≤ 4 × the player's total gold. Each bid must exceed the current one by at least 5,000 Gold. Players cannot sell BG.
- Output: BG.
- Related systems: ECO-006, rating.
- Evidence: SRC-A1-17 (uk.mlgame.org Auction).
- Status: VERIFIED (med)
- Test plan: A bid below current + 5,000 is rejected. A bid above 4× gold is rejected. A player under the rating threshold is denied access.

### ECO-020 — Transport units (Merchants / Carriers)
- Description: Units that move purchased or extracted resources.
- Game purpose: Logistics constraint and raid exposure.
- Inputs: Merchant Guild training. Merchantry science.
- Rules & constraints: Required to buy on the Market and to haul resources out of Mines. Knights' carriers have the greatest backpack capacity. Per-unit capacity is UNKNOWN.
- Output: Resource movement.
- Related systems: ECO-011, ECO-017, BLD-020 Merchant Guild.
- Evidence: SRC-A1-16, SRC-A1-15, SRC-A1-05 ("greatest capacity of the Carriers' backpacks").
- Status: VERIFIED (high)
- Test plan: Amount bought ≤ merchants × capacity.

### ECO-021 — Ancient Magic (BG-paid services)
- Description: Time-limited paid boosts, either local (per domain) or global (all domains).
- Game purpose: Monetisation and acceleration.
- Inputs: BG.
- Rules & constraints:
  - Examples: production +25% per building type in all domains (Woodcutter Cabins, Quarries, Mines, Windmills, shops); cottage population growth +10/h or +25/h; instant completion of all constructions and upgrades (not under siege).
  - Local services apply separately to each domain. A service can be prolonged any time. Some can be bought from the relevant building's window.
  - Forum example price: 7-day cottage growth for 49 BG (old).
- Output: Temporary modifiers.
- Related systems: ECO-010, POP-002.
- Evidence: SRC-A1-07: "Instant completion of all the constructions and upgrades (impossible under siege)"; "Population Growth acceleration in each Cottage (+10 per hour or +25 per hour)". SRC-A1-45 (forum price).
- Status: VERIFIED (med); prices low
- Test plan: A service is active for its duration, then removed. Instant-complete is disabled while the domain is under siege.

### ECO-022 — Market-average valuation (gold equivalent)
- Description: Each resource has min and max possible gold prices. Average = (min+max)/2. Used to value armies, buildings and holdings for ratings.
- Game purpose: A common currency for rankings and scoring.
- Inputs: Price band per resource.
- Rules & constraints: Crops band 8–20, so avg 14. Other bands UNKNOWN. Rank points = Main Rating + Army Strength × 80.
- Output: Rating values.
- Related systems: FAC-004, ECO-019 (rating gate).
- Evidence: SRC-A1-21.
- Status: VERIFIED (med)
- Test plan: valuation(crops, 100) = 1,400 gold.

### ECO-023 — Loot / items from Ruins (resource source)
- Description: Ruins are owned by Monsters. Robbing them yields artifacts and BG. Alchemy levels unlock further loot types (crystals, runes, recipes, enchantments).
- Game purpose: PvE resource faucet.
- Inputs: An army and a Robbery mission.
- Rules & constraints: The chance of an artifact depends on Ruin level. Level 5–6 ruins can have monster heroes.
- Output: Artifacts, BG, possibly resources (UNKNOWN).
- Related systems: ECO-006, TECH Alchemy.
- Evidence: SRC-A1-07, SRC-A1-24, Ruin/Artifacts pages (https://mlgame.fr/wiki-fr/Wiki.jsp?page=Ruin , https://mlgame.us/wiki-us/Wiki.jsp?page=Artifacts) seen in search results.
- Status: VERIFIED (med)
- Test plan: Loot table per ruin level, probabilities as configured.

### ECO-024 — Domain-local economy
- Description: Every domain (city, village group, outer domain) has its own stock, population and storage. Only Science Points are global.
- Game purpose: Makes logistics and geography matter.
- Inputs: —
- Rules & constraints: See ECO-012 and ECO-016. Pay "from the domain selected in the drop-down".
- Output: —
- Related systems: all.
- Evidence: SRC-A1-04 (resources stored in the producing domain), SRC-A1-10.
- Status: VERIFIED (high)
- Test plan: Production in domain A never appears in domain B without a transfer.

---

## Unknowns & discrepancies
1. **Per-level production rates, deposit sizes, warehouse capacities, base market tax, min/max prices** (except crops 8–20): all UNKNOWN. These sit in tables that snippets did not capture. Verify by fetching the Resources, Market, Rating and building pages directly once `*.mlgame.org` is allowed.
2. **Crops as food:** no evidence that population or army consume crops over time. Upkeep found is Gold only. Treat "food" as UNKNOWN.
3. **Over-storage scope:** the EN wiki ties it to the domain. The ES and RU mirrors reportedly say "city".
4. **Modifier stacking** (additive vs multiplicative) for resources is UNKNOWN. Only the per-lab science rule is documented.
5. **Ancient Magic vs "global services":** the wiki is inconsistent on naming. BG prices are mostly unknown. The forum price (49 BG) is old.
6. **Salt Lake BG rate** UNKNOWN. The Grail rate (200 BG/h) comes from one snippet.
7. **Trading Post formula** (idle people × tax) is INFERRED from the specialisation wording.
8. All evidence is snippet-based. The confidence ceiling is "med" unless two or more mirrors agree.
