# Game Overview: *My Lands: Black Gem Hunting* (reference game)

Access date: 2026-10-31. Source IDs refer to `_sources_A1.md`. Evidence comes from search-engine extracts only, because direct page fetching was blocked in this session.
Purpose: describe the **reference game's systems** so we can design an original game. We do not reuse its names, art or text.

## 1. Identity

| Field | Finding | Evidence | Status |
|---|---|---|---|
| Title | *My Lands* (browser). The Steam client is titled *My Lands: Black Gem Hunting* (Steam App ID 290730). | SRC-A1-03, SRC-A1-47 | VERIFIED (high) |
| Genre | MMO real-time strategy: "military economic real-time strategy" with a fantasy setting. Heavy city-building and economy focus. | SRC-A1-03 ("free online RTS"), SRC-A1-11/encyclopedia main page snippet, SRC-A1-39 | VERIFIED (high) |
| Business model | Free-to-play with a premium currency (Black Gems). Reviewers call it pay-to-win. | SRC-A1-03, SRC-A1-07, SRC-A1-39 | VERIFIED (high) |
| Platforms | Browser (originally Flash client with a Java backend). Windows Steam client (19 languages). Android and iOS versions existed (one Android listing removed in 2018). | SRC-A1-36, SRC-A1-38, SRC-A1-41 | VERIFIED (med) |
| Developer / publisher | Elyland (Wikipedia, Wikidata). Steam and MMOs.com credit Gravvit Ltd and Forbes Consult Ltd as developers. | SRC-A1-36, -37, -39 | VERIFIED (med; credits conflict) |
| History | Browser release in 2010. "Best browser-based game of 2011" (#ConfOG, per Wikipedia). Steam release Sep 16, 2014 (Metacritic lists Mar 14, 2014). Many regional servers and wikis (EN/UK/US/RU/FR/ES/IT/BG/CZ/HU…). Later races (Wizards, Undead, Chimeras) added on some servers. | SRC-A1-36, -03, -41, -34 | VERIFIED (med) |
| Reception | Steam: Mixed, 56% positive of 580 reviews (at time of index). Pros: extensive science trees, clan play. Cons: slow construction, pay-to-win, dated graphics. | SRC-A1-03, SRC-A1-39 | VERIFIED (med) |
| Pitch | Dark and light forces fight for the Black Gem and dominion. Cooperation, clans, alliances and team strategy are the foundation. "No to grind": cities and heroes develop while offline. | SRC-A1-03, SRC-A1-38 | VERIFIED (high) |

## 2. World structure
- **Kingdom map** of squares with coordinates (for example 36-11). Each square has a terrain giving production and army modifiers. (SRC-A1-26)
- **Domains**: Cities (Capital plus up to 7–8 more). Each city has 4 villages. Outer Domains are Mines, Pits, Salt Lakes, Grails and underground domains. Ruins are owned by monsters. Outer domains and ruins are hidden until discovered by Explorers. (SRC-A1-14, -15, -27)
- Cities cannot be seized. Outer domains can be re-seized. (SRC-A1-14)

## 3. Core loop (reconstructed)
1. **Produce**: staffed producing buildings yield wood, crops, stone, iron and gold into the local domain storage. (ECO-001..005)
2. **Grow**: population grows from the Main Building, Cottages and idle reproduction. People are assigned as Workers, Builders and Alchemists. (POP-*)
3. **Build and upgrade** on limited cells with Builders over real time. Storage and housing caps push further upgrades. (BLD-000..)
4. **Research** with Science Points from labs to unlock buildings, units, the Gate, villages, colonisation and the market. (TECH-*)
5. **Expand**: open the Gate, then build in villages (+30%), explore, colonise new cities, and seize outer domains (Pits +100%, BG lakes). (BLD-002/004, ECO-011)
6. **Fight and plunder**: train armies (they cost resources and population and are paid in gold), rob monster ruins for artifacts and BG, attack the enemy faction. (ECO-015, ECO-023, FAC-*)
7. **Trade and accelerate**: the market converts surpluses. BG (earned or bought) buys production boosts and instant completion. (ECO-017..021)
8. **Compete**: Main Rating, Army Strength, Rank and Faction Loyalty. Clans and alliances are covered by Agent A2. (ECO-022, FAC-004)

## 4. Key systems map

| System | Our file / IDs | Confidence of the reconstruction |
|---|---|---|
| Races (up to 7) and Light/Dark factions with intra-faction attack penalties | RACES_AND_FACTIONS.md RACE-*, FAC-* | med-high for structure, low for numbers |
| Five basic resources plus gold currency, premium BG, Science Points, population | ECONOMY_AND_RESOURCES.md | high for producers, low for rates |
| Finite deposits with a recovery building | ECO-009, BLD-016 | high |
| Per-domain storage, overflow decay of 1%/h | ECO-012/013 | high |
| Location modifiers: terrain (−50%..+200%), village +30%, pit +100%, paid +25% | ECO-010 | high for values, unknown stacking |
| Cell-based city with Capital of 51 cells, new cities 22→43 | BLD-003 | high |
| Labs → pooled SP → sciences (some "charged"), one specialisation at a time | TECH-* | high |
| Population: 0.04/h idle reproduction, 150k cap, overcrowding flight formula | POP-* | med-high |
| Gold army upkeep with desertion of 0.5%/h | ECO-015 | med |
| Optional Counsellor quests with manual claim | QST-* | high |

## 5. Design takeaways for our original game (non-binding)
- The reference game's distinctive levers are **finite deposits plus recovery**, **domain-local storage with decaying overflow**, **population as a three-way labour pool** (workers, builders, idle-as-tax-base) and **location multipliers** (village, pit, terrain).
- Its main criticisms (slow timers, pay-to-win) point to the risky areas: real-time construction pacing and premium acceleration.
- Many numeric tables remain UNKNOWN. Our balance must be designed from scratch, not copied.

## Unknowns & discrepancies
1. Steam and press say **four races**. Some wiki mirrors list **seven** (Wizards, Undead, Chimeras added). Current live roster UNKNOWN.
2. Developer credit conflicts (Elyland vs Gravvit/Forbes Consult). Release date conflicts (Steam Sep 16 2014 vs Metacritic Mar 14 2014).
3. Current operational status of servers and the mobile apps in 2026 is UNKNOWN (official sites unreachable in this session).
4. The Steam store page, the official site and the EN wiki index could not be opened (DNS blocked). Facts come from search snippets. Re-verify once `store.steampowered.com`, `us.mlgame.org` and `steam.mlgame.org` are allowed.
