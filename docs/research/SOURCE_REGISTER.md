# Source Register (consolidated)

Project: Realmforge / *Ashen Crowns*. This register lists every external source used by the research notes in `docs/research/`. The research reconstructs the **game systems** of the browser MMO *My Lands: Black Gem Hunting* so that we can design an original game. No names, art or texts of the reference game are reused in our product.

| Field | Value |
|---|---|
| Consolidated on | **2026-10-31** |
| Consolidated by | Agent B (Game Designer) |
| Merged from | `_sources_A1.md` (Agent A1: economy, races, buildings, sciences, population, quests; 47 entries) and `_sources_A2.md` (Agent A2: army, combat, world, heroes, social; 81 entries). Both files are kept in place unchanged. |
| ID scheme | Original IDs are kept: `SRC-A1-nn` and `SRC-A2-nn`. Every research note cites these IDs. No ID was renumbered, merged or deleted. |
| Access date of every entry | 2026-10-31 |
| Pages opened directly | **0** (all evidence comes from search-engine excerpts; see below) |

## 1. Method

1. Both research agents tried to open the official encyclopedia (the developer-run wiki with many language mirrors), the official forum, the Steam store page and secondary sites directly.
2. Direct fetching was refused by the research session's network policy (DNS not resolved, or proxy CONNECT 403). See section 3.
3. All evidence was therefore collected with a **web-search tool**. It returns indexed titles, URLs and short excerpts, plus a machine summary of those excerpts. Quotes in the research notes are taken from those excerpts. Agent A2 reports 36 searches; Agent A1 did not report a count.
4. Each finding in the research notes carries a status (VERIFIED, OBSERVED, INFERRED, PROPOSED, UNKNOWN) and a confidence (high / med / low). The confidence rules are the agents' own:
   - **high**: two or more mirrors or independent sources agree, and a direct excerpt quote exists.
   - **med**: a single official-encyclopedia excerpt (Agent A2 caps everything at med unless two mirrors or languages agree).
   - **low**: forum or fan sources, non-English mirrors only, column headers missing, or wording that is unclear in the summary.
5. Reliability labels in the table: **Official** = encyclopedia or store page run by the publisher/developer; **Community** = forum posts and fan sites; **Secondary** = Wikipedia, Wikidata, review and database sites.

## 2. The main limitation (read before using any finding)

- **Nothing was read end to end.** Every rule, number and quote is an excerpt as indexed by a search engine. Tables (per-level costs, times, outputs, capacities) were usually cut off, so most numeric details are UNKNOWN.
- **Excerpts may be stale.** The search tool flagged many pages as roughly 300 to 2,600 days old. Some URLs pin an old page version (for example `version=64`).
- **Mirrors disagree.** The encyclopedia runs separate mirrors per server and language (`steam.mlgame.org/wiki-en`, `mlgame.co.uk/wiki`, `mlgame.co.uk/wiki-uk`, `uk.mlgame.org/wiki-uk`, `mlgame.us/wiki`, `mlgame.us/wiki-us`, `us.mlgame.org/wiki-us`, plus ru/fr/es/it/bg/cz/hu/ok mirrors). Different mirrors document different game versions.
- **"VERIFIED" in our research means "verified against excerpts under the confidence rules above".** It does not mean a person read the live page or played the game. No finding has the status OBSERVED, because no gameplay was watched or played (SRC-A2-62 is a candidate video that was not viewed).
- Consequence for design: our pack values are PROPOSED unless a document explicitly ties a value to a research ID, and even then the value must be re-verified before anyone states it as a fact about the reference game.

## 3. Hosts that were blocked and must be re-verified

### 3.1 Reported as blocked by the research agents

| Host | Reported by | Failure reported | Used by (examples) |
|---|---|---|---|
| `steam.mlgame.org` | A1, A2 | DNS failed / refused | SRC-A1-01, -04, -09, -11, -25, -27, -29; SRC-A2-01, -03, -21, -22, -31 |
| `us.mlgame.org` | A1, A2 | DNS failed / refused | SRC-A1-02, -06, -28; SRC-A2-18, -24, -39, -40 |
| `uk.mlgame.org` | A1 | DNS failed | SRC-A1-04, -05, -09, -14, -17, -19, -21, -24; SRC-A2-06, -20, -26, -27, -47, -55, -60, -61 |
| `mlgame.co.uk` | A2 | refused | most `SRC-A2-*` encyclopedia rows; SRC-A1-05, -07, -08, -12, -13, -15..-17, -20..-24, -26, -28 |
| `store.steampowered.com` | A1, A2 | DNS failed / refused | SRC-A1-03, SRC-A2-38 |
| `steamcommunity.com` | A1 | DNS failed | SRC-A1-35, SRC-A2-69 |
| `en.wikipedia.org` | A1, A2 | DNS failed / refused | SRC-A1-36, SRC-A2-36 |
| `mylandsgame1.blogspot.com` (A2: `*.blogspot.com`) | A1, A2 | DNS failed / refused | SRC-A1-43, SRC-A2-35 |
| Proxy status endpoint | A1 | denied | (diagnostics only) |

### 3.2 Cited but not reported as tried directly (status of access unknown)

These hosts appear in the register but neither agent reports opening them. Treat their evidence as excerpt-only as well:
`mlgame.us`, `mlgame.fr` / `fr.mlgame.org`, `es.mlgame.org`, `it.mlgame.org`, `ru.mlgame.org`, `mlgame.ru`, `hu.mlgame.org`, `bg.mlgame.org`, `ok.mlgame.org`, `m.mlgame.co.uk`, `www.forum.mlgame.co.uk`, `www.wikidata.org`, `www.mobygames.com`, `mmos.com`, `mmohuts.com`, `www.metacritic.com`, `play.google.com`, `www.appbrain.com`, `steamdb.info`, `rawg.io`, `www.youtube.com`, `adminu.ru`, `foolgamer.ucoz.ru`, `my-lands-game.blogspot.com`, `mmoglobus.ru`.

### 3.3 Re-verification procedure

1. Allow at least `*.mlgame.org`, `mlgame.co.uk`, `mlgame.us`, `mlgame.ru`, `mlgame.fr`, `store.steampowered.com`, `steamcommunity.com` and `en.wikipedia.org` in the session network settings.
2. Open each URL in the register and record: the page revision/date, the exact wording of each cited excerpt, and any table that was missing.
3. For each research finding, keep or lower the status. **Never raise a status only because a page now opens**: raise it only when the full page actually supports the claim (and to OBSERVED only after watching or playing).
4. Priority order (most design impact first): Combat, Units/UnitsParameters, Resources, Population, Factions, Rank, Heroes, Market, City, Villages, Spec, Gate.
5. Watch SRC-A2-62 (gameplay video) to obtain the first OBSERVED evidence (UI, battle reports, map).

## 4. Same page cited under two IDs

The two agents worked in parallel, so several pages carry one ID per agent. Both IDs remain valid. Cite either; prefer the one used by the research note you quote.

| Topic (encyclopedia page) | A1 ID | A2 ID(s) |
|---|---|---|
| Steam store page | SRC-A1-03 | SRC-A2-38 |
| Official US site (not reachable) | SRC-A1-02 | SRC-A2-39 |
| Encyclopedia index / overview | SRC-A1-01 | SRC-A2-01 |
| Resources | SRC-A1-04 | SRC-A2-28 |
| Race | SRC-A1-05 | SRC-A2-17 |
| Factions | SRC-A1-06 | SRC-A2-18 |
| BlackPearl (premium gems) | SRC-A1-07 | SRC-A2-52 |
| Population | SRC-A1-08 | SRC-A2-37 |
| Buildings | SRC-A1-09 | SRC-A2-16 |
| ScienceDescription | SRC-A1-11 | SRC-A2-22 |
| Spec (specialisations) | SRC-A1-12 | SRC-A2-45 |
| Villages | SRC-A1-13 | SRC-A2-15 |
| City | SRC-A1-14 | SRC-A2-47 |
| OuterDomains, Lake, Grail, Barracks | SRC-A1-15 | SRC-A2-12, -50, -51, -10 |
| Market | SRC-A1-16 | SRC-A2-29 |
| Auction | SRC-A1-17 | SRC-A2-60 |
| Quests | SRC-A1-19 | SRC-A2-55 |
| Units / UnitsParameters / UnitsDescription | SRC-A1-20 | SRC-A2-02, -04, -03 |
| Rating / Rank | SRC-A1-21 | SRC-A2-61, -27 |
| Alchemy | SRC-A1-24 | SRC-A2-65 |
| Heroes / Skills / HeroClasses / Tavern | SRC-A1-25 | SRC-A2-19, -21, -20 |
| Kingdom / Battle | SRC-A1-26 | SRC-A2-46 |
| Exploration | SRC-A1-27 | SRC-A2-33 |
| Gate | SRC-A1-28 | SRC-A2-53 |
| Clan_warehouse / Clan_castle | SRC-A1-29 | SRC-A2-59, -31 |
| Forum "How to choose the race!" | SRC-A1-31 | SRC-A2-71 |
| Forum "best Hero for beginner" | SRC-A1-32 | SRC-A2-68 |
| Forum "Info on Known Quests" | SRC-A1-33 | SRC-A2-41 |
| Forum "Defense basics" | SRC-A1-44 | SRC-A2-14 |
| Wikipedia: My Lands | SRC-A1-36 | SRC-A2-36 |
| SteamDB / MMOs.com | SRC-A1-47, SRC-A1-39 | SRC-A2-56 |

## 5. Register (all entries, original IDs)

Rows are copied verbatim from the two agent registers; only the **Agent** column was added. "Not reachable" means the page was not opened; any evidence listed came from search excerpts.

| ID | Agent | URL | Title | Access date | What it evidences | Reliability | Limitations |
|---|---|---|---|---|---|---|---|
| SRC-A1-01 | A1 | https://steam.mlgame.org/wiki-en/ | Encyclopaedia MyLands (EN, Steam server): index | 2026-10-31 | Intended starting index. **Not reachable** (DNS blocked). Its sub-pages were found through search. | Official | Could not be opened directly |
| SRC-A1-02 | A1 | https://us.mlgame.org/ | Official site (US) | 2026-10-31 | **Not reachable** (DNS blocked) | Official | No content obtained |
| SRC-A1-03 | A1 | https://store.steampowered.com/app/290730/My_Lands_Black_Gem_Hunting/ | My Lands: Black Gem Hunting on Steam | 2026-10-31 | Pitch text ("free online RTS… dark and light forces fight for Black Gem"), choice of "one of four races", F2P, Windows, release date Sep 16 2014, Mixed reviews (56% of 580). Search snippet only. | Official store | Snippet only. Review count changes over time. |
| SRC-A1-04 | A1 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Resources&version=64 ; mirrors https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Resources , https://mlgame.us/wiki-us/Wiki.jsp?page=Resources , https://mlgame.co.uk/wiki/Wiki.jsp?page=Resources , https://es.mlgame.org/wiki/Wiki.jsp?page=Resources , https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Resources | Encyclopaedia MyLands: Resources | 2026-10-31 | Producers of each resource, finite deposits (drop to 1/h), village +30%, Pit +100%, storage rules, overflow plunder at 1%/h, hard cap of 999,999,999, terrain modifiers from -50% to +200%, race bonus built into building base rates | Official | Snippets. The URL is a pinned *version=64*, so it may not be the latest. Mirrors disagree on the scope of overflow. |
| SRC-A1-05 | A1 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Race ; https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Race ; https://mlgame.fr/wiki-fr/Wiki.jsp?page=Race ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Race | Encyclopaedia MyLands: Race | 2026-10-31 | Race list, homelands and production specialties (Elves wood; Drow wood+iron, faster mages; Knights gold, carrier capacity, warehouses; Demons stone, strongest mercenary; Wizards crops+iron, half mana cost, spell resistance; Undead wood+stone; Chimeras underground) | Official | Cached copies are about 530–770+ days old. Some details appear only on the FR mirror. |
| SRC-A1-06 | A1 | https://us.mlgame.org/wiki-us/Wiki.jsp?page=Factions ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Factions ; https://mlgame.fr/wiki-fr/Wiki.jsp?page=Fraction | Encyclopaedia MyLands: Factions | 2026-10-31 | Light = Elves, Knights, Wizards. Dark = Dark Elves, Demons, Undeads, Chimeras. Attack penalty -50% vs own race and -25% vs own faction. Renegade status for 7 days. Faction Loyalty rating. | Official | Snippets. The roster is version dependent. |
| SRC-A1-07 | A1 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=BlackPearl ; https://mlgame.us/wiki-us/Wiki.jsp?page=BlackPearl ; https://fr.mlgame.org/wiki-fr/Wiki.jsp?page=BlackPearl | Encyclopaedia MyLands: BlackPearl (Black Gems) | 2026-10-31 | BG is the premium "special resource". Sources: Salt Lake/Grail, Ruins, Auction, real money. Cannot be taken from players. "Present BG". Ancient Magic services (+25% production, cottage growth +10/+25 per hour, instant completion except under siege). Local services apply per domain. | Official | About 1,400 days old. Prices not captured. |
| SRC-A1-08 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Population | Encyclopaedia MyLands: Population | 2026-10-31 | 150,000 cap per domain. Growth from Main Building + Cottages + idle reproduction at 0.04 × idle per hour. Flight formula N×(0.01717+0.00002828×N). Idle people flee first. "Distribute Builders/Workers". Idle people raise Trading Post gold. | Official | Snippets. Revision age unknown. |
| SRC-A1-09 | A1 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Buildings ; https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Buildings ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Buildings ; https://mlgame.fr/wiki/Wiki.jsp?page=Buildings ; https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Buildings | Encyclopaedia MyLands: Buildings | 2026-10-31 | Building categories and names. Producing buildings need Workers or Alchemists. Builders and repair. Time bonus cap of 80%. Buildings deleted if their cell is lost. Race-specific visuals. | Official | No per-level tables captured |
| SRC-A1-10 | A1 | https://mlgame.us/wiki/Wiki.jsp?page=Science ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Science | Encyclopaedia MyLands: Science | 2026-10-31 | Alchemist Lab required. Science Points shared across domains. More labs mean faster research. Bonuses apply per lab (multiply, round, then sum). Green "confirm" button with resource payment from a chosen domain. Surplus SP lost. | Official | Snippets |
| SRC-A1-11 | A1 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=ScienceDescription ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=ScienceDescription ; https://mlgame.us/wiki-us/Wiki.jsp?page=ScienceDescription ; http://it.mlgame.org/wiki/Wiki.jsp?page=ScienceDescription | Encyclopaedia MyLands: ScienceDescription | 2026-10-31 | Named sciences and what each unlocks: Gate, Villages, Market, Colonisation, Exchequer (salary), Heroes Managing, Cache, Repair, Towers, Fortification, Diplomacy, Espionage, Merchantry, City Expansion (25,000 SP, max 43 cells), Capital Transfer, Clan Castles (max 5), Mission Control, Hero Skills (hero cap 12), Vision, Reflection, Magic Towers, Underground domains | Official | Mirrors list different science sets (version drift) |
| SRC-A1-12 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Spec ; https://es.mlgame.org/wiki/Wiki.jsp?page=Spec | Encyclopaedia MyLands: Spec (Specialisations) | 2026-10-31 | One specialisation per player; starting a second cancels the first. Architect -50/65/80% build time. Trader market tax -2/5/9%, Trading Post tax +0.01/0.02/0.03. Benefactor cottage growth +8/12/16 per hour. Defender boosts towers and fortifications. | Official | Not every specialisation captured |
| SRC-A1-13 | A1 | https://mlgame.us/wiki-us/Wiki.jsp?page=Villages ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Villages | Encyclopaedia MyLands: Villages | 2026-10-31 | 4 villages per city. Prerequisites: Main Building at max level, open Gate, "Villages" science. +30% production. Villages are shared with neighbours (others' buildings shown red). | Official | Snippets |
| SRC-A1-14 | A1 | https://mlgame.us/wiki/Wiki.jsp?page=City ; https://mlgame.us/wiki-us/Wiki.jsp?page=City ; https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=City ; https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=City ; https://bg.mlgame.org/wiki/Wiki.jsp?page=City | Encyclopaedia MyLands: City | 2026-10-31 | Capital: 51 cells, max city level 5, cannot be removed, receives quest rewards, can be transferred. New city starts with 22 cells and grows with Main Building level, capped by its square. Up to 7 (or 8) more cities via Settler. Cities cannot be seized. | Official | Mirrors conflict: "7 more" vs "8 more" cities |
| SRC-A1-15 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=OuterDomains ; https://mlgame.us/wiki/Wiki.jsp?page=Mine ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Lake ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Grail ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Barracks | Encyclopaedia MyLands: Outer Domains, Mine, Lake, Grail, Barracks (underground domain) | 2026-10-31 | Domain types: Stone/Gold/Iron Mines (no building; resources hauled by Merchants), Pits (a few cells, +100%), Salt Lakes (4 cells, BG), Grails (4 cells, visible to all, 200 BG/h). Life span, with +30 days extension except Lake and Grail. | Official | The Salt Lake BG rate was not found |
| SRC-A1-16 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Market ; https://steam.mlgame.org/wiki-ru/Wiki.jsp?page=Market | Encyclopaedia MyLands: Market | 2026-10-31 | Sell or buy resources for Gold. Needs the Market science, the Market building and Merchants (Merchant Guild) to buy, but no Merchants to sell. Partial lot purchases. "Exchange" tab at worse rates. BG bought in the "Auction" tab. | Official | Tax and price bounds not captured |
| SRC-A1-17 | A1 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Auction ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Auction | Encyclopaedia MyLands: Auction | 2026-10-31 | BG auction for Gold. Requires Main Rating of 10,000,000. Bid capped at 4× the player's total gold. Minimum raise of 5,000 Gold. Players cannot sell BG. | Official | Snippets |
| SRC-A1-18 | A1 | https://mlgame.us/wiki/Wiki.jsp?page=Mistakes ; https://mlgame.co.uk/wiki/Wiki.jsp?page=Mistakes | Encyclopaedia MyLands: Mistakes | 2026-10-31 | Newbie pitfalls: quests optional, reward collected from the Counsellor, deposits deplete, overflow plundered, portal vs server BG wallet | Official | Old (about 890–2,000 days) |
| SRC-A1-19 | A1 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Quests ; https://ok.mlgame.org/wiki-ok/Wiki.jsp?page=Quests ; https://mlgame.us/wiki/Wiki.jsp?page=Quests | Encyclopaedia MyLands: Quests | 2026-10-31 | Counsellor gives quests; return to collect the reward. Rewards delivered to the Capital. Some rewards are blocked if a Capital battle is less than 2 h away. Protection rewards (from monster attacks and negative events). Quests not obligatory. | Official | Quest list not captured |
| SRC-A1-20 | A1 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Units ; https://mlgame.co.uk/wiki/Wiki.jsp?page=UnitsParameters ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=UnitsDescription ; https://steam.mlgame.org/wiki-en/Wiki.jsp?page=UnitsDescription | Encyclopaedia MyLands: Units / UnitsParameters / UnitsDescription | 2026-10-31 | Gold salary for garrisons. Desertion 0.5% per hour of unpaid units. Doubled salary on missions. Max salary reduction 99.74% (Exchequer 96%). Unit costs in resources and population. Peasant as a basic unit. A server coefficient applies. | Official | Per-unit cost tables not captured with names |
| SRC-A1-21 | A1 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rating ; https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Rating ; https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rank ; https://mlgame.us/wiki-us/Wiki.jsp?page=Rank ; https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Rating | Encyclopaedia MyLands: Rating, Rank | 2026-10-31 | Main Rating = gold value of buildings, army, domains and resources at the market-average rate (avg = (min+max)/2). Crops 8–20, giving 14. Clan-warehouse resources excluded. Rank = Main Rating + Army Strength × 80. Ranks run from Esquire (0) to Suzerain (650,000,000). | Official | About 2,000 days old. Min/max for other resources not captured. |
| SRC-A1-22 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Recovery | Encyclopaedia MyLands: Recovery (Chalice of Abundance) | 2026-10-31 | Chalice is built only on a depleted deposit after demolishing the producer. It restores the deposit at some rate, then self-destroys. | Official | Rate not captured |
| SRC-A1-23 | A1 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Terraformer ; https://hu.mlgame.org/wiki-hu/Wiki.jsp?page=Terraformer | Encyclopaedia MyLands: Terraformer | 2026-10-31 | Clears obstacle cells (lake, trees, rocks), then self-destroys. Level 1 takes 4 days, level 2 takes 1 day. | Official | Snippet |
| SRC-A1-24 | A1 | https://mlgame.co.uk/wiki-hr/Wiki.jsp?page=Alchemist ; https://it.mlgame.org/wiki/Wiki.jsp?page=Alchemy ; https://mlgame.ru/wiki-ru/Wiki.jsp?page=Alchemy ; https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Alchemy | Encyclopaedia MyLands: Alchemist, Alchemy | 2026-10-31 | Alchemy science levels unlock magic crystals and building/artifact improvement, then later loot types. Level 6 brings monumental buildings. RU: central building starts at level 1, max 5, and affects storage, housing and growth. | Official (non-EN) | Translation risk |
| SRC-A1-25 | A1 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Heroes ; https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Skills ; https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=HeroClasses ; https://ok.mlgame.org/wiki-ok/Wiki.jsp?page=Tavern | Encyclopaedia MyLands: Heroes, Skills, HeroClasses, Tavern | 2026-10-31 | Hiring needs the Heroes Managing science, a Tavern (one only, Capital only) and Captain rank. Daily offers for gold. Skills are active or passive; 3 basic skills at level 1, then 2 skills per level. | Official | Mostly for A2 / combat scope. Included for links only. |
| SRC-A1-26 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Kingdom ; https://mlgame.us/wiki-us/Wiki.jsp?page=Battle | Encyclopaedia MyLands: Kingdom, Battle | 2026-10-31 | Each square has a terrain giving production and army modifiers ("Terrain Information"). Native lands give +25 Defence to the matching race. Max Defence 50. 20-round battles. | Official | Per-terrain production table not captured |
| SRC-A1-27 | A1 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Exploration ; https://mlgame.us/wiki-us/Wiki.jsp?page=Exploration | Encyclopaedia MyLands: Exploration | 2026-10-31 | Explorers come from the Expeditionary Corps. Exploration of empty squares finds Ruins and Outer Domains or city sites. Death chance 2–10% depending on speed. | Official | Snippet |
| SRC-A1-28 | A1 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Gate ; https://us.mlgame.org/wiki-us/Wiki.jsp?page=Gate | Encyclopaedia MyLands: Gate | 2026-10-31 | Gate (initial building) needs the Gate science. When open it enables combat missions, new cities and villages. Snippet: "opens 5 minutes and closes 6 hours". | Official | The timing statement is ambiguous |
| SRC-A1-29 | A1 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Clan_warehouse ; https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Clan_castle | Encyclopaedia MyLands: Clan_warehouse, Clan_castle | 2026-10-31 | Clan warehouse storage runs from 100,000 (level 1) to 25,000,000 (level 9). Clan castle living places run from 1,000 (level 1) to 3,000 (level 5). | Official | Clan systems belong to A2. Recorded here for the resource link only. |
| SRC-A1-30 | A1 | https://mlgame.us/forum/viewtopic.php?t=123 | Forum: "Build your city!" | 2026-10-31 | Player advice: labs are cheap, tower counts by city star level, plan buildings around the army | Community | About 2015. Opinion. |
| SRC-A1-31 | A1 | https://mlgame.us/forum/viewtopic.php?f=11&t=194 | Forum: "How to choose the race! – PDF version included" | 2026-10-31 | Player view of race bonuses: Elf healers heal +1 unit, Demon assassins about +50% attack, Drow mages reduce tower damage and train faster, Knights get crops. Penalty stacking. | Community | Opinion. Conflicts with the wiki on the Knights bonus. |
| SRC-A1-32 | A1 | https://mlgame.us/forum/viewtopic.php?t=681 | Forum: "best Hero for beginner" | 2026-10-31 | Hero classes mentioned (Wizard, Governor, Trader). Trader can cover army salary. | Community | Opinion |
| SRC-A1-33 | A1 | https://mlgame.co.uk/forum/viewtopic.php?t=703 | Forum: "Info on Known Quests" | 2026-10-31 | Quests are chained ("passage of a previous quest activates the next") | Community | 2012 |
| SRC-A1-34 | A1 | https://uk.mlgame.org/forum/viewtopic.php?p=10961 | Forum: "New server with a new race is now open!" | 2026-10-31 | New race (Chimeras) released on a new server, with doubled extraction in mines and pit Main Buildings | Community/official announcement | Date unknown |
| SRC-A1-35 | A1 | https://steamcommunity.com/sharedfiles/filedetails/?id=320588419 (RU: ?id=320604110) | Steam guide: "Fast start in My Lands" | 2026-10-31 | Follow the Counsellor quests. Rush the Main Building to levels 4–5 with a service. Early science order: Construction 1 → Market 1 → Market 2 → Fortification 1 → Cache 1 → Resurrection 1 → Merchantry 1 → Villages 1 → Gate. | Community | Page now shows a removal notice. Evidence comes from snippets of cached and translated copies. |
| SRC-A1-36 | A1 | https://en.wikipedia.org/wiki/My_Lands | Wikipedia: My Lands | 2026-10-31 | 2010 browser launch, Elyland. Four races (elves, drow, demons, knights). BG is the account-bound currency. Gold sources. 4 villages per city. Best browser game 2011 (#ConfOG). | Secondary | Snippet only. May be dated. |
| SRC-A1-37 | A1 | https://www.wikidata.org/wiki/Q4044698 | Wikidata: My Lands | 2026-10-31 | Developer Elyland. Country UK/Russia. | Secondary | Snippet |
| SRC-A1-38 | A1 | https://www.mobygames.com/game/132575/my-lands-black-gem-hunting/ | MobyGames | 2026-10-31 | Store-text mirror: "one of four races", "No to 'grind'", offline development via skills | Secondary | Snippet |
| SRC-A1-39 | A1 | https://mmos.com/review/my-lands-black-gem-hunting | MMOs.com review | 2026-10-31 | MMO RTS focused on city building and economy. Pros and cons: big science trees, clan play, slow builds, pay-to-win. Developer credit Gravvit Ltd / Forbes Consult Ltd. | Secondary | Opinion |
| SRC-A1-40 | A1 | https://mmohuts.com/review/my-lands-black-gem-hunting | MMOHuts review | 2026-10-31 | Long-running timer-based pacing | Secondary | Opinion |
| SRC-A1-41 | A1 | https://www.metacritic.com/game/my-lands/ ; https://play.google.com/store/apps/details?id=com.mlgame.mylands1 ; https://www.appbrain.com/app/my-lands/com.mlgame.mylands | Metacritic / Google Play / AppBrain | 2026-10-31 | Mobile and iOS versions. Metacritic date 14 Mar 2014 conflicts with Steam's 16 Sep 2014. An Android app was removed in 2018 and another listing remains. | Secondary | Snippet |
| SRC-A1-42 | A1 | https://adminu.ru/2012/02/bystryj-start-v-my-lands-ch-1/ ; https://foolgamer.ucoz.ru/publ/kakie_zdanija_stroit_v_my_lands_sovety_po_zastrojke_gorodov/1-1-0-7 ; http://my-lands-game.blogspot.com/p/server-i-rasa.html ; https://mmoglobus.ru/rasy-my-lands | Russian fan guides | 2026-10-31 | Producing buildings only on matching resource cells until the "land-use" science ("Землеустройство"). Level-1 buildings are cheaper and faster. Stone is scarce early. Knights have large warehouses. | Community (RU) | 2010–2012. Translation risk. |
| SRC-A1-43 | A1 | https://mylandsgame1.blogspot.com/search/label/Resources | Fan blog: Resources | 2026-10-31 | Same producer list as the wiki. Race bonus in base stats. Terrain modifies production. | Community | Fetch blocked. Search summary only. |
| SRC-A1-44 | A1 | https://mlgame.co.uk/forum/viewtopic.php?f=17&t=729 | Forum: "Defense basics" | 2026-10-31 | Domain seizure is possible only with no defensive buildings left. Fortifications raise the defence level. | Community | Opinion |
| SRC-A1-45 | A1 | https://mlgame.us/forum/viewtopic.php?p=2763 ; https://mlgame.us/forum/viewtopic.php?t=59 | Forum: "New Local Services", "Game too slow" | 2026-10-31 | Treasure Chest → "Local services" → Purchase. Example price: 7-day cottage growth boost for 49 BG. | Community | Old prices |
| SRC-A1-46 | A1 | https://mlgame.co.uk/forum/viewtopic.php?p=3626 | Forum: "More Gold Production To All Races Except Knights" | 2026-10-31 | Player discussion of per-Trading-Post gold output | Community | Opinion. Numbers not captured. |
| SRC-A1-47 | A1 | https://steamdb.info/app/290730/ ; https://rawg.io/games/my-lands-black-gem-hunting | SteamDB / RAWG | 2026-10-31 | App ID 290730. Achievement referencing hero skills. | Secondary | Snippet |
| SRC-A2-01 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=MyLands (index https://steam.mlgame.org/wiki-en/) | Encyclopaedia MyLands: MyLands (overview/index) | 2026-10-31 | Mission taxonomy (combat / hero / non-combat); Black Gem sources; explorable site types | Official | Index page not fetchable; only excerpts |
| SRC-A2-02 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Units (also ru.mlgame.org/wiki-ru Units, ok.mlgame.org) | Encyclopaedia MyLands: Units | 2026-10-31 | 8 unit types per race; gold salary; 0.5%/h desertion; double salary on missions; speed coefficients 1/1.25/1.5/2; training accel ≤90%; science-upgrade cost difference; worked example 465 carriers | Official | Older article; EN/RU mirrors |
| SRC-A2-03 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=UnitsDescription (mirror mlgame.co.uk/wiki/…UnitsDescription) | Encyclopaedia MyLands: UnitsDescription | 2026-10-31 | Unit categories & 4 levels; pursuer/victim loop; carrier capacity; healer resurrection; mercenary fortification immunity; mage tower reduction | Official | First category header not visible; healer numbers inconsistent in excerpt |
| SRC-A2-04 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=UnitsParameters | Encyclopaedia MyLands: UnitsParameters | 2026-10-31 | Example Warrior row; "Rating" = total resource cost; server coefficient | Official | Column headers missing in excerpt |
| SRC-A2-05 | A2 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=UnitParametrs | Энциклопедия Mylands: UnitParametrs (RU) | 2026-10-31 | Defence parameter starts at 0 | Official (RU) | Translation; excerpt only |
| SRC-A2-06 | A2 | https://uk.mlgame.org/wiki/Wiki.jsp?page=Combat (also m.mlgame.co.uk mobile v36, mlgame.co.uk/wiki-uk) | Encyclopaedia MyLands: Combat | 2026-10-31 | Fortification round; retreat thresholds; 20-round cap; defender wins ties; reinforcements; 2 attacks/24h; squad award formula; 15-min minimum; recall | Official | Multiple versions |
| SRC-A2-07 | A2 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Combat (also mlgame.ru/wiki-ru) | Энциклопедия Mylands: Combat (RU) | 2026-10-31 | Damage algorithm (sum attack → reduce by defence % → subtract health; overflow redistribution); pursuer extra strike | Official (RU) | Flagged as ≥368 days old |
| SRC-A2-08 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Calculator (also uk/es/ok mirrors) | Encyclopaedia MyLands: Calculator | 2026-10-31 | Battle calculator inputs/outputs; average of 30 simulations | Official | — |
| SRC-A2-09 | A2 | https://mlgame.co.uk/forum/viewtopic.php?t=1640 | Forum: "Carriers" | 2026-10-31 | Definitions of Attack/Health/Defence | Community | Player-written guide |
| SRC-A2-10 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Barracks | Encyclopaedia MyLands: Barracks | 2026-10-31 | Underground barracks domains; 43 cells; training discounts | Official | 50% figure from RU only |
| SRC-A2-11 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=TrainMercenaries | Encyclopaedia MyLands: TrainMercenaries | 2026-10-31 | Building levels 1–9; queue 75→4,000; time 8 min→50 s | Official | Single excerpt |
| SRC-A2-12 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=OuterDomains | Encyclopaedia MyLands: OuterDomains | 2026-10-31 | Seizure prerequisites; building damage on seizure; invisibility until explored; 50% robbery cap; mines with tunnels | Official | — |
| SRC-A2-13 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Espionage | Encyclopaedia MyLands: Espionage | 2026-10-31 | Spy mission, counterspies, access levels 1–4, range by science | Official | — |
| SRC-A2-14 | A2 | https://mlgame.co.uk/forum/viewtopic.php?f=17&t=729 | Forum: "Defense basics" | 2026-10-31 | Tower/Magic Tower/Fortification behaviour; closed gate → no battle; one building type per destruction attack | Community | Conflicts with official on fortifications |
| SRC-A2-15 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Villages | Encyclopaedia MyLands: Villages | 2026-10-31 | 4 villages per city; +30% speed; wall protection; shared villages | Official | — |
| SRC-A2-16 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Buildings | Encyclopaedia MyLands: Buildings | 2026-10-31 | Workers/builders die when building destroyed by enemy | Official | — |
| SRC-A2-17 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Race | Encyclopaedia MyLands: Race | 2026-10-31 | Race bonuses (Knights, Elves, Dark Elves); one race change | Official | Not all races' text retrieved |
| SRC-A2-18 | A2 | https://us.mlgame.org/wiki-us/Wiki.jsp?page=Factions (also mlgame.co.uk/wiki-uk) | Encyclopaedia MyLands: Factions | 2026-10-31 | Light/Dark faction rosters; Renegade; faction loyalty rating; post-resurrection counting | Official | ES version omits Chimeras |
| SRC-A2-19 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Heroes | Encyclopaedia MyLands: Heroes | 2026-10-31 | Hero attributes, hiring, max 7, XP rules & server multipliers, level cap 12 | Official | Partly dated |
| SRC-A2-20 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=HeroClasses (RU: ru.mlgame.org/wiki-ru/…HeroClasses) | Encyclopaedia MyLands: HeroClasses | 2026-10-31 | Skill frame convention; Whisperer/Spellcaster rank rule | Official | Class list not in excerpt |
| SRC-A2-21 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Skills | Encyclopaedia MyLands: Skills | 2026-10-31 | 3 basic skills; 2 per level; active vs passive | Official | — |
| SRC-A2-22 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=ScienceDescription | Encyclopaedia MyLands: ScienceDescription | 2026-10-31 | Sciences: Hero Commander, Diplomacy, Colonisation, defensive buildings, Accelerated Studying | Official | Costs not retrieved |
| SRC-A2-23 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Artefacts | Encyclopaedia MyLands: Artefacts | 2026-10-31 | Sets (9 of 12), activation rules, artefact theft needs hero, crystal theft | Official | — |
| SRC-A2-24 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Artefact_store (also us.mlgame.org Artifact_store) | Encyclopaedia MyLands: Artefact_store | 2026-10-31 | Vault, hero bag, trader 30 (older 24) slots, alchemy altar slots | Official | Version differences |
| SRC-A2-25 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Stunned | Encyclopaedia MyLands: Stunned | 2026-10-31 | Stun effects; mana regen rules; failed-seizure exception | Official | Duration not stated |
| SRC-A2-26 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Dungeon (also mlgame.co.uk/wiki-uk, VN mirror) | Encyclopaedia MyLands: Dungeon | 2026-10-31 | Capture-chance formula; level coeff 1→30; capacity 1→15; ransom & tax; release paths | Official | ~890 days old; VN shows L1–7 only |
| SRC-A2-27 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rank (also mlgame.us/wiki-us Rank) | Encyclopaedia MyLands: Rank | 2026-10-31 | Rank Rule; exemptions; revenge; compassion spell; rank points formula & thresholds; monster protection quest | Official | Ambiguous Suzerain wording |
| SRC-A2-28 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Resources (also steam wiki-en Resources v64) | Encyclopaedia MyLands: Resources | 2026-10-31 | Gold consumed by troops; Black Gem special resource; robbery burn/clan exception; mine outputs | Official | — |
| SRC-A2-29 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Market | Encyclopaedia MyLands: Market | 2026-10-31 | Player market for gold; merchants; exchange tab; robbable stock | Official | RU says "Black pearl" |
| SRC-A2-30 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Camp (mobile v10) | Encyclopaedia MyLands: Camp | 2026-10-31 | Camp domain; Encamping mission rules; targetable by robbery/slaves | Official | Camp relocation not covered |
| SRC-A2-31 | A2 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Clan_castle (also mlgame.co.uk) | Encyclopaedia MyLands: Clan_castle | 2026-10-31 | Internal gate rounds (100000 attack points absorbed), seizure, War & Peace points, standards | Official | EN versions differ in wording |
| SRC-A2-32 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Assist | Encyclopaedia MyLands: Assist | 2026-10-31 | Assistance 2× speed; joint defence | Official | — |
| SRC-A2-33 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Exploration (also steam wiki-en) | Encyclopaedia MyLands: Exploration | 2026-10-31 | Explorer purpose; 2–10% death by speed; rank-based ruin discovery | Official | — |
| SRC-A2-34 | A2 | https://mlgame.co.uk/forum/viewtopic.php?t=151 | Forum: "Attacked by a 5* player?" | 2026-10-31 | 2011 claim that any rank can attack mines/pits | Community | Old (2011) |
| SRC-A2-35 | A2 | https://mylandsgame1.blogspot.com/2011/10/my-lands-mlgame-heroes.html | Blog: "My Lands (mlgame) - Heroes" | 2026-10-31 | 2011 hire rules (Suzerain; not under attack) | Community | Old; conflicts with current wiki |
| SRC-A2-36 | A2 | https://en.wikipedia.org/wiki/My_Lands | Wikipedia: My Lands | 2026-10-31 | Developer/publisher; "4 races"; official wiki link | Secondary | Fetch blocked; search summary only |
| SRC-A2-37 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Population | Encyclopaedia MyLands: Population | 2026-10-31 | Excess population flees | Official | — |
| SRC-A2-38 | A2 | https://store.steampowered.com/app/290730/My_Lands_Black_Gem_Hunting/ | Steam store page | 2026-10-31 | "choose one of four races"; dark vs light fight for Black Gem; clans & alliances; 56% of 580 reviews positive (as indexed) | Official (publisher) | Fetch blocked; search snippet only |
| SRC-A2-39 | A2 | https://us.mlgame.org/ | Official site | 2026-10-31 | — (not reachable) | Official | Blocked; no evidence used |
| SRC-A2-40 | A2 | https://us.mlgame.org/wiki-us/Wiki.jsp?page=Ansients | Encyclopaedia MyLands: Ansients | 2026-10-31 | Ancients monster reinforcement protecting newbies (shield icon) | Official | Excerpt partly via RU text |
| SRC-A2-41 | A2 | https://mlgame.co.uk/forum/viewtopic.php?t=703 | Forum: "Info on Known Quests" | 2026-10-31 | Slave capture tips; captives may die in slow transit | Community | Anecdotal |
| SRC-A2-42 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=MoveResources | Encyclopaedia MyLands: MoveResources | 2026-10-31 | Transport between own domains; merchants; open gates; 3% merchant death | Official | — |
| SRC-A2-43 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Mes | Encyclopaedia MyLands: Mes (Messages) | 2026-10-31 | Mailbox UI | Official | Report format not described |
| SRC-A2-44 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=TrainAirborne | Encyclopaedia MyLands: TrainAirborne | 2026-10-31 | Airborne base training time 20 min | Official | Single figure |
| SRC-A2-45 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Spec (ES mirror es.mlgame.org) | Encyclopaedia MyLands: Spec | 2026-10-31 | Hero specialisations (Cleric, Miner, Architect, Benefactor) with 3 levels | Official | Header garbled in excerpt |
| SRC-A2-46 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Kingdom | Encyclopaedia MyLands: Kingdom | 2026-10-31 | Coordinates (kingdom X-Y + square) | Official | Map size not given |
| SRC-A2-47 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=City (IT/RU mirrors) | Encyclopaedia MyLands: City | 2026-10-31 | Capital 51 cells; +8 cities; 22 cells; square-dependent size; capital transfer | Official | Old RU says +7 |
| SRC-A2-48 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Ruin | Encyclopaedia MyLands: Ruin | 2026-10-31 | 1–7 ruins/kingdom + Archruin; aggressive ruins; respawn | Official | ~2,600 days old |
| SRC-A2-49 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Pit (steam wiki-en Pit) | Encyclopaedia MyLands: Pit | 2026-10-31 | 5 levels; lifespan collapse; tunnel 5 min / 6 h | Official | Lifespan value missing |
| SRC-A2-50 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Lake | Encyclopaedia MyLands: Lake (Salt Lake) | 2026-10-31 | Flying island, 4 cells, BG production, no gate | Official | — |
| SRC-A2-51 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Grail | Encyclopaedia MyLands: Grail | 2026-10-31 | 200 BG/hour; daily random relocation destroys buildings | Official | — |
| SRC-A2-52 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=BlackPearl | Encyclopaedia MyLands: BlackPearl | 2026-10-31 | Premium currency uses, premium perks, gifting | Official | — |
| SRC-A2-53 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Gate (RU mirror) | Encyclopaedia MyLands: Gate | 2026-10-31 | Gate open 5 min / close 6 h; Ancient Magic 30 min; not after attack launched | Official | — |
| SRC-A2-54 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Underground_estates | Encyclopaedia MyLands: Underground_estates | 2026-10-31 | Existence of underground estates | Official | No content retrieved |
| SRC-A2-55 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Quests (mlgame.co.uk/wiki Quests) | Encyclopaedia MyLands: Quests | 2026-10-31 | Counsellor quests, reward guards, Ancients 60 days, medals, random events 19:00, monsters refuse Arch-ruins, Temple ≤5 | Official | Mixed versions |
| SRC-A2-56 | A2 | https://steamdb.info/app/290730/ ; https://mmos.com/review/my-lands-black-gem-hunting | SteamDB / MMOs.com review | 2026-10-31 | Listed in search as secondary descriptions (not used for rules) | Secondary | Not opened |
| SRC-A2-57 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Tournament | Encyclopaedia MyLands: Tournament | 2026-10-31 | Flags Hunting & Treasure Hunt rules and prizes | Official | Versions differ |
| SRC-A2-58 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Clan | Encyclopaedia MyLands: Clan | 2026-10-31 | Clan requirements; 5–100 members; types; roles | Official | — |
| SRC-A2-59 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Clan_warehouse | Encyclopaedia MyLands: Clan_warehouse | 2026-10-31 | Clan warehouse/reservation in Pit; loot routing; robbable | Official | Table values may be dated |
| SRC-A2-60 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Auction (mlgame.co.uk/wiki Auction) | Encyclopaedia MyLands: Auction | 2026-10-31 | BG auction: 10M rating, 5,000 step, +1h anti-snipe, 48h pay, 7-day ban | Official | — |
| SRC-A2-61 | A2 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rating | Encyclopaedia MyLands: Rating | 2026-10-31 | Rating window, tournament tab | Official | Old |
| SRC-A2-62 | A2 | https://www.youtube.com/watch?v=g5vrbLmC0MY | YouTube: "My Lands: Black Gem Hunting - First Impressions Gameplay" | 2026-10-31 | Candidate for OBSERVED evidence (not viewed) | Community | Not accessed |
| SRC-A2-63 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Spells | Encyclopaedia MyLands: Spells | 2026-10-31 | Book of Spells, casting, replacement rule, cooldown | Official | No cost table |
| SRC-A2-64 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=SpellName | Encyclopaedia MyLands: SpellName | 2026-10-31 | Spell effect examples; default range one kingdom | Official | — |
| SRC-A2-65 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Alchemy (uk/es/it mirrors) | Encyclopaedia MyLands: Alchemy | 2026-10-31 | Perfection recipe, runes ≤5, epic crystals, 7 levels | Official | EN/ES differ |
| SRC-A2-66 | A2 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=HeroTemp | Энциклопедия Mylands: HeroTemp (RU) | 2026-10-31 | Monster XP per ruin level table | Official (RU) | Excerpt only |
| SRC-A2-67 | A2 | https://mlgame.co.uk/forum/viewtopic.php?p=4767 | Forum: "Hero XP" | 2026-10-31 | Arena ~100 XP/h at L7 | Community | Anecdotal |
| SRC-A2-68 | A2 | https://mlgame.us/forum/viewtopic.php?t=681 | Forum: "best Hero for beginner" | 2026-10-31 | Class names (Strategist, Governor, Scientist, Wizard); commander cost 20M science points | Community | Player opinion |
| SRC-A2-69 | A2 | https://steamcommunity.com/app/290730/discussions/0/622955136048237252 | Steam discussion: "Managing Heroes" | 2026-10-31 | Hero management discussion (listed, not used for rules) | Community | Not opened |
| SRC-A2-70 | A2 | https://mlgame.co.uk/forum/viewtopic.php?t=1261 | Forum: "suzerain" | 2026-10-31 | Suzerain = level 5; 650M points | Community | — |
| SRC-A2-71 | A2 | https://mlgame.us/forum/viewtopic.php?f=11&t=194 | Forum: "How to choose the race!" | 2026-10-31 | Faction loyalty can be negative (claim) | Community | Player claim |
| SRC-A2-72 | A2 | https://www.forum.mlgame.co.uk/viewtopic.php?p=7555 | Forum: "Hero Captured" | 2026-10-31 | Anecdote of 5-day imprisonment | Community | Anecdotal |
| SRC-A2-73 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Brotherhood | Encyclopaedia MyLands: Brotherhood | 2026-10-31 | Sword Brotherhood bounty vs aggressors (7 days) | Official | Mechanics thin |
| SRC-A2-74 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Serv (RU ru.mlgame.org/wiki-ru Serv) | Encyclopaedia MyLands: Serv | 2026-10-31 | Server types; Tactical vs Mining differences; Faction War mode; XP multipliers | Official | Heroic/Strategic/Operative undocumented |
| SRC-A2-75 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Monarchic | Encyclopaedia MyLands: Monarchic | 2026-10-31 | Monarchic clan type (roles) | Official | Excerpt only |
| SRC-A2-76 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=JointAttack | Encyclopaedia MyLands: JointAttack | 2026-10-31 | One joiner; 50/50 loot; cascade withdrawal; per-player rank check; 3+owner defence | Official | — |
| SRC-A2-77 | A2 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Alliance | Энциклопедия Mylands: Alliance (RU) | 2026-10-31 | Alliance ≤7 members; cap by Diplomacy level | Official (RU) | RU only |
| SRC-A2-78 | A2 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Mentor | Encyclopaedia MyLands: Mentor | 2026-10-31 | Mentor/apprentice; daily resource send | Official | — |
| SRC-A2-79 | A2 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Rules (uk/us/es/it mirrors) | Encyclopaedia MyLands: Rules | 2026-10-31 | Conduct rules; gag 1–7 days; bans; shared-IP rules | Official | Versions differ |
| SRC-A2-80 | A2 | https://fr.mlgame.org/wiki-fr/Wiki.jsp?page=Add_contact ; https://fr.mlgame.org/wiki-fr/Wiki.jsp?page=Filtr_Mes | Encyclopédie MyLands (FR): Add_contact, Filtr_Mes | 2026-10-31 | Contacts and message filters | Official (FR) | — |
| SRC-A2-81 | A2 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Tournament | Энциклопедия Mylands: Tournament (RU) | 2026-10-31 | Tournament versions/prizes | Official (RU) | — |

## 6. Tally

| Measure | A1 | A2 | Total |
|---|---|---|---|
| Register entries | 47 | 81 | 128 |
| Entries listed but yielding no rule evidence | SRC-A1-01, -02 (index/official site not reachable) | SRC-A2-39, -56, -62, -69 | 6 |
| Pages opened directly | 0 | 0 | **0** |

## 7. Change log

| Date | Change |
|---|---|
| 2026-10-31 | Register consolidated from `_sources_A1.md` and `_sources_A2.md`. IDs unchanged. Blocked-host list and duplicate-page map added. Source files left in place. |
