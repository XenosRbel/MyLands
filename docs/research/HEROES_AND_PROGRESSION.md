# Heroes and Progression — Research Notes (Agent A2)

Subject: public materials about "My Lands: Black Gem Hunting" (MLG). Systems only; do not reuse names.
Access date for all sources: **2026-10-31**. Source IDs → `_sources_A2.md`.
Method caveat: search-excerpt-level evidence only (direct fetch blocked) — see `ARMY_AND_COMBAT.md`.

---

### HERO-001 — Hero acquisition (hiring)
- Description: Heroes are hired in a Tavern building.
- Game purpose: Mid-game unlock that personalises the army.
- Inputs: science "Heroes Managing", Tavern built, player rank.
- Rules & constraints: requires the science + Tavern + Captain rank (current EN wiki). Cannot hire while
  the Capital is under attack (2011 blog). Some classes have extra rank rules: a specific spell-oriented
  class ("Whisperer" in EN / "Spellcaster" in RU) is hireable only at Suzerain rank (RU wording: Captain,
  Master, Marshal ranks cannot hire it). Commander-class heroes need the "Hero Commander" science (forum:
  ~20 million science points).
- Output: hero with random name, gender, class, specialty, mana, experience, level 1.
- Related systems: HERO-002, HERO-010 (ranks).
- Evidence: SRC-A2-19 ("study the 'Heroes Managing' science, construct the Tavern building, and be of
  Captain rank"; "random Name from the list, Class, Specialty, some Mana, some Experience, Level");
  SRC-A2-20 ("only players with Suzerain rank can hire Whisperer hero"); SRC-A2-35 (2011 blog: Suzerain
  rank required; not under attack); SRC-A2-68 (forum 20M science points).
- Status: VERIFIED (med). Rank requirement conflicts over time (see Unknowns).
- Test plan: hire rejected without science/Tavern/rank or while capital has incoming attack.

### HERO-002 — Hero slots (max heroes)
- Rule: max 7 heroes: 5 via levels of "Heroes Managing" science, +1 via "Hero Commander" science, +1 via
  a special artefact set ("+1 ... but no more than 7").
- Evidence: SRC-A2-19; SRC-A2-22 ("Changes the parameters of science 'Heroes Managing' by +1, but no more than 7").
- Status: VERIFIED (high — EN + RU agree).
- Test plan: slots = min(7, sciLevelSlots + commander + setBonus).

### HERO-003 — Hero classes
- Description: Each hero has a class that determines its skill tree. Names seen: Commander (EN/RU),
  Whisperer/Spellcaster (EN/RU — likely same class), "Wizard" (forum), plus forum-reported early classes
  Strategist, Governor, Scientist. Full class list UNKNOWN.
- Evidence: SRC-A2-20; SRC-A2-22; SRC-A2-68 (forum: "Strategist, Governor, Scientist"; wizard relies on
  active casts).
- Status: INFERRED (low) for full list; VERIFIED (med) that classes exist and gate skills.
- Test plan: class → allowed skill set mapping; class-specific hiring rules.

### HERO-004 — Specialty (specialisation) with 3 levels
- Description: Each hero has a specialty with 3 levels providing economy/military bonuses. Examples:
  Cleric (+1/+3/+5% Resurrector resurrection; healer training time −50/−60/−70%), Miner (+20/40/60% iron,
  stone, gold production), Architect (−50/65/80% build/upgrade/repair time), Benefactor (cottage capacity,
  population growth).
- Game purpose: Heroes double as governors (economic buffs) not only war leaders.
- Evidence: SRC-A2-45 (Spec page; ES mirror shows same Cleric values).
- Status: VERIFIED (med; high for Cleric — EN+ES agree).
- Test plan: bonus active only while hero is in the domain (see HERO-005); level → value table lookup.

### HERO-005 — Skills: basic/active/passive, per-level skill points
- Description: A new level-1 hero has 3 basic skills at level 1; each level makes 2 skill upgrades
  available. Active skills (round-framed) add spells to the Book of Spells; passive skills (square-framed)
  are permanently on but only while the hero is in the domain and not stunned.
- Rules & constraints: skills cannot be studied while an incoming attack is underway.
- Evidence: SRC-A2-21 ("A new, level 1 Hero has 3 Basic Skills of level 1"; "2 Skills per each level");
  SRC-A2-20 (frame convention); SRC-A2-63.
- Status: VERIFIED (high for frame convention — EN + RU agree; med for skill counts).
- Test plan: skill points = 3 initial + 2 × (level − 1); study blocked under attack.

### HERO-006 — Spells and mana
- Description: Spells live in a Book of Spells (from active skills or consumable scrolls). Each spell has
  level, mana cost, cooldown ("preparation time after last use"), and optional duration. Re-casting the
  same spell replaces the active effect unless the existing one is higher level. Spells may be cast during
  battle. Default range = one kingdom when not stated.
- Examples (effects only): a necromancy spell converts killed troops (both sides) into mercenary units up
  to 20% of killed; a reveal spell shows all true attacks on the current domain; an explorer spell instantly
  explores a chosen cell.
- Mana: regenerates over time; no regen while stunned, in Sanctuary, or on a mission.
- Evidence: SRC-A2-63 (Spells); SRC-A2-64 (SpellName: Necromancy ≤20%; "True View"; "Explorer's eye");
  SRC-A2-25 ("Mana is not regenerating if hero is stunned, in the Sanctuary or in mission").
- Status: VERIFIED (med). Mana regen rate, max mana, per-spell costs UNKNOWN.
- Test plan: cast rejected if mana < cost or cooldown active; lower-level recast cannot overwrite higher.

### HERO-007 — Stun, capture, dungeon, ransom, release
- Description: A hero on the losing side is Stunned: cannot be sent on missions; passive skills, spells
  and artefacts inactive; no mana regen; no XP gain. A defeated hero may be imprisoned in the winner's
  Dungeon.
- Rules & constraints: capture chance = (gold cost of attacker's killed units after resurrection &
  necromancy ÷ gold cost of attacker's participating units) × dungeon-level coefficient; ≥100% ⇒ certain.
  Coefficient 1 (L1) → 30 (L9); prisoner capacity 1 → 15. Ransom paid in crops + iron, depends on hero
  level, +10% tax; L8/L9 dungeons add +50%/+100%. Release options: "Hero Release" mission, ransom, or the
  owner lets go (even with gate closed); freed hero returns via "Escape" mission (Fast march speed).
  Hero whose battle is won but whose seizure fails is not stunned. Stun duration UNKNOWN.
- Game purpose: High-stakes hero risk; creates rescue/ransom gameplay.
- Evidence: SRC-A2-25 (Stunned); SRC-A2-26 (Dungeon: coefficients, ransom, tax); SRC-A2-72 (forum
  "Hero Captured" anecdote, 5 days imprisonment).
- Status: VERIFIED (med). Older VN mirror lists only levels 1–7 (discrepancy).
- Test plan: capture probability formula unit test; clamp at 100%; ransom = base(level) × 1.1 × levelBonus.

### HERO-008 — Hero experience and levels
- Description: XP mainly from killed enemy troops, proportional to their gold value with a server-type and
  win/lose multiplier. If the attacker wins: XP = gold value of slain enemy troops × 0.0002 (Mining server)
  or × 0.0005 (War server); other cases (defender, loss) use other multipliers not retrieved. Worked
  example (War server): 10,000 level-1 mercenaries at 75 gold each ⇒ 375 XP (consistent with 0.0005). Monster XP per ruin level from a table (RU: 2 XP at L1 → 25 XP at L5 on
  battle servers). Fractions truncated. Multiple players on one side split XP by troop cost at battle start.
  Stunned heroes gain none. Level cap 12 without "Hero Skills" science; XP stops at the science cap. Other XP
  sources: clan castle arena (forum ~100 XP/h at L7), a rune (+10% battle XP), "Accelerated Studying"
  science (+battle XP), Legendary heroes +100% on a specific server type.
- Evidence: SRC-A2-19; SRC-A2-66 (HeroTemp RU table); SRC-A2-67 (forum Hero XP); SRC-A2-65 (Alchemy rune);
  SRC-A2-22 (Accelerated Studying).
- Evidence (multipliers): SRC-A2-74/SRC-A2-19 ("War server: If the attacker wins, his hero gains experience
  = gold value of slain enemy troops 0.0005").
- Status: VERIFIED (med) incl. attacker-win multipliers; other-case multipliers and XP-to-level curve UNKNOWN.
- Test plan: XP = floor(Σ killed_gold × k_server × k_outcome / share); no XP when stunned; cap enforcement.

### HERO-009 — Artefacts, sets, inventory, Vault
- Description: Heroes equip artefacts (levelled items). Artefacts belong to at most one Set; equipping any
  9 of a set's 12 pieces gives an extra set effect. Equipped artefacts work only while the hero is active
  (not stunned, not in Sanctuary) and affect a domain only while the hero is there (science bonuses excepted).
  Inventory: Hero Bag; overflow to the Vault (Capital-only building; holds artefacts, scrolls, bottles,
  crystals, runes). Artefact trader shows 30 item slots (older: 24).
- Sources of artefacts: robbing Ruins with a hero; tournaments; trader/store; auction (SOC).
- Evidence: SRC-A2-23 ("Additional effects work when any 9 of 12 artifacts of the same set are equipped");
  SRC-A2-24 (Artefact store, Vault); SRC-A2-19.
- Status: VERIFIED (med). Equipment slot layout (“doll”) UNKNOWN.
- Test plan: set bonus triggers at ≥9 distinct pieces of a set; deactivation on stun/sanctuary/leave domain.

### HERO-010 — Alchemy: crystals, perfection, runes
- Description: Alchemy science (7 levels) unlocks: perfecting artefacts (5 crystals + 2 same-level
  artefacts + 1 scroll → a "perfect" artefact), stealing Magic Crystals, Runes (passive bonus engraved on
  artefacts; up to 5), Epic Crystals for upgrading perfect artefacts (first upgrade picks 1 of 3 attributes),
  and +3 perfect-artefact upgrade levels at science level 7. Robbery with a hero can steal 1 random crystal
  from the defender hero's bag.
- Evidence: SRC-A2-65 (Alchemy EN; ES differs in detail); SRC-A2-23.
- Status: VERIFIED (med); EN vs ES discrepancies.
- Test plan: recipe validation; rune cap 5.

### HERO-011 — Player rank (account progression tier)
- Description: Rank (level) from Rank Points = Main Rating + Army Strength Rating × 80. Main Rating = gold
  equivalent (market average) of buildings, army, cities/pits/mines and resources. Thresholds: Captain 60M,
  Master/Magister 125M, Marshal 350M, Suzerain 650M (Suzerain = level 5; lower ranks unnamed in excerpt).
  Rank can drop (abandon domains, destroy buildings, lose troops).
- Game purpose: Gates hero hiring, event entry, and PvP targeting (see SOC-PVP).
- Evidence: SRC-A2-27 (Rank: "Rank_Points = Main_Raiting + (Army_Strength_Rating * Coefficient)", coeff 80);
  SRC-A2-70 (forum "Level 5 is Suzerain").
- Status: VERIFIED (med).
- Test plan: rank recomputed from ratings; thresholds monotonic; demotion when points fall.

### HERO-012 — Sciences as the main progression tree (army-relevant subset)
- Description: Sciences gate units (war sciences per unit category), defensive buildings (Tower,
  Fortification, Magic Tower), heroes (Heroes Managing, Hero Skills, Hero Commander, Accelerated
  Studying), expansion (Colonisation, Capital Transfer, Outer Domains, Underground domains, Camp),
  Espionage, Sanctuary, Resurrection, Clans, Diplomacy, Alchemy.
- Evidence: SRC-A2-22 (ScienceDescription); SRC-A2-02.
- Status: VERIFIED (med). Costs/levels UNKNOWN (other agent's scope: economy/tech).

---

## Unknowns & discrepancies

1. **Hire rank**: current wiki says Captain; 2011 blog says Suzerain → rules changed over time.
2. **Class list** not retrieved; "Whisperer" (EN) vs "Spellcaster" (RU) likely one class (translation).
3. Stun duration, mana pool/regen numbers, spell costs, XP curve per level: UNKNOWN.
4. Equipment slots per hero ("doll") and artefact level range: UNKNOWN (tournament prizes mention up to
   level 5 artefacts).
5. Dungeon table: EN shows levels 1–9; older VN version shows 1–7.
6. Alchemy details differ between EN and ES pages.
7. XP multipliers known only for "attacker wins" on Mining (0.0002) and War (0.0005) servers.
8. Verify via HeroClasses, Skills, Spells, Heroes, Dungeon pages once hosts are reachable, and via
   gameplay videos (OBSERVED hero window).
