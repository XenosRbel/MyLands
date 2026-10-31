# Races and Factions: research notes (A1)

Access date for all evidence: 2026-10-31. Source IDs refer to `_sources_A1.md`.
**Evidence caveat:** every source page was read only through search-engine extracts, because direct fetching was blocked in this session. See the method notes in `_sources_A1.md`. Race names below are the original game's names, kept for reference only. Our game must use its own names.

---

### RACE-001 — Playable race selection (one race per account/server)
- Description: At start the player picks one race. The race fixes building looks, the unit roster, production specialties and its faction.
- Game purpose: Long-term identity and asymmetric economy and army choices. Drives replayability and faction conflict.
- Inputs: Player choice at account or server registration.
- Rules & constraints: The Steam text says "one of four races". The encyclopedia lists up to 7 races across versions (see RACE-002..008). One character per server ("Only one character is allowed per server", SRC-A1-18 Rules snippet). Whether race can be changed later is UNKNOWN.
- Output: Race-dependent building base rates, units, visuals and faction membership.
- Related systems: FAC-001 (factions), ECO-001..006 (production bonuses), BLD-000 (race visuals), army (A2/other).
- Evidence: SRC-A1-03 (Steam): "player chooses one of four races to lead it to the victory". SRC-A1-09: "The same type of buildings looks different for each race". SRC-A1-04: "Race Bonus is considered in initial characteristics of each race's buildings".
- Status: VERIFIED (high: Steam text plus wiki agree on race selection; the number of races is version dependent)
- Test plan: Create an account for each race and assert that base production rates differ per race only for the specialty resources. Assert that race is immutable unless a "change race" feature is defined.

### RACE-002 — Elves (Light)
- Description: Forest race. Highest wood production. Friendly with Knights.
- Game purpose: Wood-economy specialist.
- Inputs: Race choice.
- Rules & constraints: "greatest rate of Wood production". Home terrain is Magic Forest, which gives +25 Defence to Elves (SRC-A1-26). Forum: Elf healers resurrect one extra unit (low).
- Output: Higher base wood rate in Woodcutter-type buildings (exact % UNKNOWN).
- Related systems: ECO-001 Wood, RACE-009 native land bonus.
- Evidence: SRC-A1-05 (mlgame.co.uk Race): "greatest rate of Wood production". SRC-A1-31 (forum): healers "treat one more unit".
- Status: VERIFIED (high for the wood specialty; low for the healer bonus)
- Test plan: Elf wood producer at level L outputs more than a non-Elf producer at level L with the same terrain and bonuses.

### RACE-003 — Dark Elves / Drow (Dark)
- Description: Elves exiled to caves of the Cursed Woods and mutated. Allied with Demons.
- Game purpose: Wood and iron specialist with a mage-army edge.
- Inputs: Race choice.
- Rules & constraints: "produce Wood and Iron well". Mages train faster. Combat mages are strong against enemy Towers. Home terrain is Cursed Forest (+25 Def).
- Output: Wood and iron rate bonus; mage training-time bonus (values UNKNOWN).
- Related systems: ECO-001, ECO-004, unit training.
- Evidence: SRC-A1-05: "produce Wood and Iron well"; mages "train faster". SRC-A1-31: "drow mages reduce defensive tower damage and are produced faster".
- Status: VERIFIED (med)
- Test plan: Drow wood/iron rates exceed the baseline. Drow mage training time is below the baseline.

### RACE-004 — Knights (Light)
- Description: Human race. Gold and logistics specialist.
- Game purpose: Trade and gold economy.
- Inputs: Race choice.
- Rules & constraints: "highest Gold production efficiency and the greatest capacity of the Carriers' backpacks". Warehouses have the highest capacity (SRC-A1-04 summary). Unit line starts with the Peasant (SRC-A1-20). Home terrain is Holy Land per the faction/terrain listing (exact mapping low).
- Output: Gold rate bonus, carrier capacity bonus, warehouse capacity bonus (values UNKNOWN).
- Related systems: ECO-005 Gold, ECO-012 storage, ECO-020 transport.
- Evidence: SRC-A1-05 (mlgame.co.uk/wiki Race), SRC-A1-04. A forum (SRC-A1-31) instead says Knights get "increased crop production", which **conflicts**. A RU fan source (SRC-A1-42) says Knights have gold and grain bonuses plus capacious warehouses.
- Status: VERIFIED (high for gold and carriers; med for warehouse; crops disputed)
- Test plan: Knight gold producer output exceeds baseline. Carrier per-unit capacity exceeds other races. Warehouse capacity per level exceeds other races.

### RACE-005 — Demons (Dark)
- Description: Inhabit the Deadlands (scorched volcanic land). Allied with Drow.
- Game purpose: Stone specialist with a strong mercenary.
- Inputs: Race choice.
- Rules & constraints: Mine stone faster. Their mercenary has the strongest attack among mercenaries. Forum: assassins about +50% attack (low). Home terrain is Deadland (+25 Def; one Battle snippet maps Deadland to Undead, see discrepancies).
- Output: Stone rate bonus.
- Related systems: ECO-003 Stone.
- Evidence: SRC-A1-05: "mine stone faster"; the "Demoness" mercenary has "the strongest attack among mercenaries" (FR mirror).
- Status: VERIFIED (med)
- Test plan: Demon stone producers exceed baseline output.

### RACE-006 — Wizards (Light)
- Description: Live in the Magic Woods next to the Elves. Magic-oriented.
- Game purpose: Magic and spell economy plus crops/iron.
- Inputs: Race choice (availability per server UNKNOWN).
- Rules & constraints: "expend half as much mana on spells". Heroes have natural spell resistance. "greatest rate of Crops and Iron production". Allied with Elves and Knights.
- Output: Spell mana cost ×0.5. Crop and iron rate bonus.
- Related systems: ECO-002 Crops, ECO-004 Iron, hero mana (A2).
- Evidence: SRC-A1-05 (uk.mlgame.org / FR mirror): "expend half as much mana on spells". SRC-A1-06: "Elves, Knights and Wizards – present the Light Faction".
- Status: VERIFIED (med; appears in only some mirrors, so likely a later addition)
- Test plan: Spell cost for a Wizard hero is exactly 50% of the base cost.

### RACE-007 — Undead (Dark)
- Description: Dark-faction race in some versions.
- Game purpose: Fast rebuilding (wood and stone).
- Inputs: Race choice.
- Rules & constraints: "increased production of wood and stone". This helps them rebuild cities quickly.
- Output: Wood and stone rate bonus.
- Related systems: ECO-001, ECO-003.
- Evidence: SRC-A1-05 (FR mirror) and SRC-A1-06 (faction list includes "Undeads").
- Status: VERIFIED (low; FR mirror only)
- Test plan: As for RACE-002, for wood and stone.

### RACE-008 — Chimeras (Dark)
- Description: Underground-oriented race, introduced on a new server.
- Game purpose: Underground-domain specialist.
- Inputs: Race choice (server-specific).
- Rules & constraints: Colonise underground domains faster. Produce Gold from gold deposits in Pits and Mines. No movement penalty underground. Forum: doubled extraction in mines and in pit Main Buildings.
- Output: Faster underground expansion. ×2 extraction in Mines and Pits (forum).
- Related systems: ECO-011 Outer Domains, BLD-030 Pits.
- Evidence: SRC-A1-05 (FR), SRC-A1-34 forum announcement "New server with a new race".
- Status: VERIFIED (low)
- Test plan: Chimera pit Main Building output equals 2× the baseline.

### RACE-009 — Native-land (home terrain) bonus
- Description: Each race gets a defensive bonus on its home terrain type.
- Game purpose: Ties geography to race. Gives an incentive to settle on home terrain.
- Inputs: Terrain of the square where the battle or domain is located.
- Rules & constraints: Magic Forest, Cursed Forest, Holy Land and Deadland each give "+25 Defence" to matching races. Max Defence is 50. Does not work for Renegades.
- Output: +25 Defence for units on native terrain.
- Related systems: ECO-010 terrain production modifiers, FAC-003 Renegade.
- Evidence: SRC-A1-26 (Battle page): native lands give "+25 Defence"; "Max Defence value cannot exceed 50". SRC-A1-05/06: the bonus is "not working if player is a Renegade".
- Status: VERIFIED (med; the exact race-to-terrain mapping is partly unclear)
- Test plan: A unit with base defence D on native land has min(D+25, 50). A Renegade gets no bonus.

---

## Factions

### FAC-001 — Two factions: Light vs Dark
- Description: Races are grouped into two opposing factions.
- Game purpose: Macro-conflict framing and diplomacy defaults.
- Inputs: Race (faction derived from race).
- Rules & constraints: Light = Elves, Knights, Wizards. Dark = Dark Elves, Demons, Undeads, Chimeras (latest mirror). Older sources say Light = Knights, Elves and Dark = Demons, Drow.
- Output: Faction tag per player, used by penalties and ratings.
- Related systems: FAC-002, FAC-003, FAC-004.
- Evidence: SRC-A1-06 (us.mlgame.org Factions): "Elves, Knights and Wizards – present the Light Faction". SRC-A1-36 (Wikipedia) lists four races. SRC-A1-03: "dark and light forces fight for Black Gem".
- Status: VERIFIED (high)
- Test plan: Faction is computed from race. No race belongs to both factions.

### FAC-002 — Same-race / same-faction attack penalty
- Description: Attacking your own kind weakens your attack.
- Game purpose: Steers aggression toward the opposing faction.
- Inputs: Attacker race and faction; defender race and faction.
- Rules & constraints: -50% attack vs own race; -25% vs own faction. A forum describes this as stacking: -25% faction plus -25% race = -50% total (interpretation).
- Output: Attack multiplier.
- Related systems: combat (A2), FAC-003.
- Evidence: SRC-A1-06: "attacks a player of his own race receives a -50% Attack Penalty… own faction receives -25%". SRC-A1-31 (forum) gives the stacking interpretation.
- Status: VERIFIED (high for the values; stacking is INFERRED)
- Test plan: Same race: attack ×0.5. Same faction, different race: ×0.75. Opposite faction: ×1.0.

### FAC-003 — Renegade status
- Description: A player who attacks a domain of their own race becomes a Renegade.
- Game purpose: Social punishment for infighting.
- Inputs: Attack on an own-race domain.
- Rules & constraints: Lasts 7 days. Renegade domains have red titles on the map. Renegades can be attacked by their own race or faction without penalty. They lose the native-land bonus.
- Output: Temporary status flag.
- Related systems: FAC-002, RACE-009.
- Evidence: SRC-A1-06: "becomes Renegade for 7 days… red coloured titles on the Kingdom Map".
- Status: VERIFIED (med)
- Test plan: After an own-race attack, the flag is set for 7×24 h and the penalty is waived for attackers of the renegade.

### FAC-004 — Faction Loyalty rating
- Description: Rating of the gold-equivalent value of enemy-faction (or Renegade) units killed.
- Game purpose: Rewards cross-faction PvP.
- Inputs: Units killed, valued at the market-average gold rate.
- Rules & constraints: Attackers only (initiator and joint players). Units resurrected by the defender do not count.
- Output: Loyalty points.
- Related systems: ECO-030 market-average valuation, rating (A2).
- Evidence: SRC-A1-06, SRC-A1-21.
- Status: VERIFIED (med)
- Test plan: Killing 10 enemy units of rating R gives the attacker 10R loyalty. Resurrected units subtract.

---

## Unknowns & discrepancies
1. **Number of races:** Steam, MobyGames and MMOs.com say "four races". Wikipedia names elves, drow, demons, knights. Wiki mirrors add Wizards, Undead and Chimeras. These appear to be later or server-specific additions. Which races exist on current servers is UNKNOWN. Verify on the live wiki Race page (steam.mlgame.org) and the registration screen.
2. **Exact bonus percentages** for every race specialty are UNKNOWN. The wiki says bonuses are baked into the base rates of buildings. Verify through the in-game building info of each race or per-race building tables.
3. **Knights' bonus:** the wiki says gold, carriers and warehouses. A forum says crops. A RU fan site says gold and grain. Unresolved.
4. **Deadland mapping:** one Battle snippet pairs Deadland with Undead. Race pages place Demons in the Deadlands. The exact native-land table needs a direct read.
5. **Penalty stacking** (-50% vs -25%-plus-25%) is a forum interpretation.
6. Whether race can be changed and whether faction can be switched apart from Renegade status: UNKNOWN.
7. All evidence is from search snippets (fetch blocked). Re-verify by fetching SRC-A1-05 and SRC-A1-06 directly.
