# Source Register — Agent A2 (Army, World, Heroes, Social)

All access dates: **2026-10-31**.

**Access method (important limitation for every row):** direct HTTP fetching of `steam.mlgame.org`,
`us.mlgame.org`, `mlgame.co.uk`, `store.steampowered.com`, `en.wikipedia.org` and `*.blogspot.com` was
refused by this environment's network policy (DNS not resolved / proxy CONNECT 403). Evidence was taken
from **search-engine indexed excerpts** of these pages (WebSearch, often restricted to the
`mlgame.org` / `mlgame.co.uk` / `mlgame.us` domains). Quotes are therefore short excerpts as indexed, not
full-page reads; table headers were sometimes missing. Re-verify by opening pages directly once
the hosts are allowed. Reliability: "official" = MLG Encyclopaedia (developer-run wiki, multiple language
mirrors); "community" = official forum player posts or third-party sites.

| ID | URL | Title | Access date | What it evidences | Reliability | Limitations |
|---|---|---|---|---|---|---|
| SRC-A2-01 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=MyLands (index https://steam.mlgame.org/wiki-en/) | Encyclopaedia MyLands: MyLands (overview/index) | 2026-10-31 | Mission taxonomy (combat / hero / non-combat); Black Gem sources; explorable site types | Official | Index page not fetchable; only excerpts |
| SRC-A2-02 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Units (also ru.mlgame.org/wiki-ru Units, ok.mlgame.org) | Encyclopaedia MyLands: Units | 2026-10-31 | 8 unit types per race; gold salary; 0.5%/h desertion; double salary on missions; speed coefficients 1/1.25/1.5/2; training accel ≤90%; science-upgrade cost difference; worked example 465 carriers | Official | Older article; EN/RU mirrors |
| SRC-A2-03 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=UnitsDescription (mirror mlgame.co.uk/wiki/…UnitsDescription) | Encyclopaedia MyLands: UnitsDescription | 2026-10-31 | Unit categories & 4 levels; pursuer/victim loop; carrier capacity; healer resurrection; mercenary fortification immunity; mage tower reduction | Official | First category header not visible; healer numbers inconsistent in excerpt |
| SRC-A2-04 | https://mlgame.co.uk/wiki/Wiki.jsp?page=UnitsParameters | Encyclopaedia MyLands: UnitsParameters | 2026-10-31 | Example Warrior row; "Rating" = total resource cost; server coefficient | Official | Column headers missing in excerpt |
| SRC-A2-05 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=UnitParametrs | Энциклопедия Mylands: UnitParametrs (RU) | 2026-10-31 | Defence parameter starts at 0 | Official (RU) | Translation; excerpt only |
| SRC-A2-06 | https://uk.mlgame.org/wiki/Wiki.jsp?page=Combat (also m.mlgame.co.uk mobile v36, mlgame.co.uk/wiki-uk) | Encyclopaedia MyLands: Combat | 2026-10-31 | Fortification round; retreat thresholds; 20-round cap; defender wins ties; reinforcements; 2 attacks/24h; squad award formula; 15-min minimum; recall | Official | Multiple versions |
| SRC-A2-07 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Combat (also mlgame.ru/wiki-ru) | Энциклопедия Mylands: Combat (RU) | 2026-10-31 | Damage algorithm (sum attack → reduce by defence % → subtract health; overflow redistribution); pursuer extra strike | Official (RU) | Flagged as ≥368 days old |
| SRC-A2-08 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Calculator (also uk/es/ok mirrors) | Encyclopaedia MyLands: Calculator | 2026-10-31 | Battle calculator inputs/outputs; average of 30 simulations | Official | — |
| SRC-A2-09 | https://mlgame.co.uk/forum/viewtopic.php?t=1640 | Forum: "Carriers" | 2026-10-31 | Definitions of Attack/Health/Defence | Community | Player-written guide |
| SRC-A2-10 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Barracks | Encyclopaedia MyLands: Barracks | 2026-10-31 | Underground barracks domains; 43 cells; training discounts | Official | 50% figure from RU only |
| SRC-A2-11 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=TrainMercenaries | Encyclopaedia MyLands: TrainMercenaries | 2026-10-31 | Building levels 1–9; queue 75→4,000; time 8 min→50 s | Official | Single excerpt |
| SRC-A2-12 | https://mlgame.co.uk/wiki/Wiki.jsp?page=OuterDomains | Encyclopaedia MyLands: OuterDomains | 2026-10-31 | Seizure prerequisites; building damage on seizure; invisibility until explored; 50% robbery cap; mines with tunnels | Official | — |
| SRC-A2-13 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Espionage | Encyclopaedia MyLands: Espionage | 2026-10-31 | Spy mission, counterspies, access levels 1–4, range by science | Official | — |
| SRC-A2-14 | https://mlgame.co.uk/forum/viewtopic.php?f=17&t=729 | Forum: "Defense basics" | 2026-10-31 | Tower/Magic Tower/Fortification behaviour; closed gate → no battle; one building type per destruction attack | Community | Conflicts with official on fortifications |
| SRC-A2-15 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Villages | Encyclopaedia MyLands: Villages | 2026-10-31 | 4 villages per city; +30% speed; wall protection; shared villages | Official | — |
| SRC-A2-16 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Buildings | Encyclopaedia MyLands: Buildings | 2026-10-31 | Workers/builders die when building destroyed by enemy | Official | — |
| SRC-A2-17 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Race | Encyclopaedia MyLands: Race | 2026-10-31 | Race bonuses (Knights, Elves, Dark Elves); one race change | Official | Not all races' text retrieved |
| SRC-A2-18 | https://us.mlgame.org/wiki-us/Wiki.jsp?page=Factions (also mlgame.co.uk/wiki-uk) | Encyclopaedia MyLands: Factions | 2026-10-31 | Light/Dark faction rosters; Renegade; faction loyalty rating; post-resurrection counting | Official | ES version omits Chimeras |
| SRC-A2-19 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Heroes | Encyclopaedia MyLands: Heroes | 2026-10-31 | Hero attributes, hiring, max 7, XP rules & server multipliers, level cap 12 | Official | Partly dated |
| SRC-A2-20 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=HeroClasses (RU: ru.mlgame.org/wiki-ru/…HeroClasses) | Encyclopaedia MyLands: HeroClasses | 2026-10-31 | Skill frame convention; Whisperer/Spellcaster rank rule | Official | Class list not in excerpt |
| SRC-A2-21 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Skills | Encyclopaedia MyLands: Skills | 2026-10-31 | 3 basic skills; 2 per level; active vs passive | Official | — |
| SRC-A2-22 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=ScienceDescription | Encyclopaedia MyLands: ScienceDescription | 2026-10-31 | Sciences: Hero Commander, Diplomacy, Colonisation, defensive buildings, Accelerated Studying | Official | Costs not retrieved |
| SRC-A2-23 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Artefacts | Encyclopaedia MyLands: Artefacts | 2026-10-31 | Sets (9 of 12), activation rules, artefact theft needs hero, crystal theft | Official | — |
| SRC-A2-24 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Artefact_store (also us.mlgame.org Artifact_store) | Encyclopaedia MyLands: Artefact_store | 2026-10-31 | Vault, hero bag, trader 30 (older 24) slots, alchemy altar slots | Official | Version differences |
| SRC-A2-25 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Stunned | Encyclopaedia MyLands: Stunned | 2026-10-31 | Stun effects; mana regen rules; failed-seizure exception | Official | Duration not stated |
| SRC-A2-26 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Dungeon (also mlgame.co.uk/wiki-uk, VN mirror) | Encyclopaedia MyLands: Dungeon | 2026-10-31 | Capture-chance formula; level coeff 1→30; capacity 1→15; ransom & tax; release paths | Official | ~890 days old; VN shows L1–7 only |
| SRC-A2-27 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rank (also mlgame.us/wiki-us Rank) | Encyclopaedia MyLands: Rank | 2026-10-31 | Rank Rule; exemptions; revenge; compassion spell; rank points formula & thresholds; monster protection quest | Official | Ambiguous Suzerain wording |
| SRC-A2-28 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Resources (also steam wiki-en Resources v64) | Encyclopaedia MyLands: Resources | 2026-10-31 | Gold consumed by troops; Black Gem special resource; robbery burn/clan exception; mine outputs | Official | — |
| SRC-A2-29 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Market | Encyclopaedia MyLands: Market | 2026-10-31 | Player market for gold; merchants; exchange tab; robbable stock | Official | RU says "Black pearl" |
| SRC-A2-30 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Camp (mobile v10) | Encyclopaedia MyLands: Camp | 2026-10-31 | Camp domain; Encamping mission rules; targetable by robbery/slaves | Official | Camp relocation not covered |
| SRC-A2-31 | https://steam.mlgame.org/wiki-en/Wiki.jsp?page=Clan_castle (also mlgame.co.uk) | Encyclopaedia MyLands: Clan_castle | 2026-10-31 | Internal gate rounds (100000 attack points absorbed), seizure, War & Peace points, standards | Official | EN versions differ in wording |
| SRC-A2-32 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Assist | Encyclopaedia MyLands: Assist | 2026-10-31 | Assistance 2× speed; joint defence | Official | — |
| SRC-A2-33 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Exploration (also steam wiki-en) | Encyclopaedia MyLands: Exploration | 2026-10-31 | Explorer purpose; 2–10% death by speed; rank-based ruin discovery | Official | — |
| SRC-A2-34 | https://mlgame.co.uk/forum/viewtopic.php?t=151 | Forum: "Attacked by a 5* player?" | 2026-10-31 | 2011 claim that any rank can attack mines/pits | Community | Old (2011) |
| SRC-A2-35 | https://mylandsgame1.blogspot.com/2011/10/my-lands-mlgame-heroes.html | Blog: "My Lands (mlgame) - Heroes" | 2026-10-31 | 2011 hire rules (Suzerain; not under attack) | Community | Old; conflicts with current wiki |
| SRC-A2-36 | https://en.wikipedia.org/wiki/My_Lands | Wikipedia: My Lands | 2026-10-31 | Developer/publisher; "4 races"; official wiki link | Secondary | Fetch blocked; search summary only |
| SRC-A2-37 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Population | Encyclopaedia MyLands: Population | 2026-10-31 | Excess population flees | Official | — |
| SRC-A2-38 | https://store.steampowered.com/app/290730/My_Lands_Black_Gem_Hunting/ | Steam store page | 2026-10-31 | "choose one of four races"; dark vs light fight for Black Gem; clans & alliances; 56% of 580 reviews positive (as indexed) | Official (publisher) | Fetch blocked; search snippet only |
| SRC-A2-39 | https://us.mlgame.org/ | Official site | 2026-10-31 | — (not reachable) | Official | Blocked; no evidence used |
| SRC-A2-40 | https://us.mlgame.org/wiki-us/Wiki.jsp?page=Ansients | Encyclopaedia MyLands: Ansients | 2026-10-31 | Ancients monster reinforcement protecting newbies (shield icon) | Official | Excerpt partly via RU text |
| SRC-A2-41 | https://mlgame.co.uk/forum/viewtopic.php?t=703 | Forum: "Info on Known Quests" | 2026-10-31 | Slave capture tips; captives may die in slow transit | Community | Anecdotal |
| SRC-A2-42 | https://mlgame.co.uk/wiki/Wiki.jsp?page=MoveResources | Encyclopaedia MyLands: MoveResources | 2026-10-31 | Transport between own domains; merchants; open gates; 3% merchant death | Official | — |
| SRC-A2-43 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Mes | Encyclopaedia MyLands: Mes (Messages) | 2026-10-31 | Mailbox UI | Official | Report format not described |
| SRC-A2-44 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=TrainAirborne | Encyclopaedia MyLands: TrainAirborne | 2026-10-31 | Airborne base training time 20 min | Official | Single figure |
| SRC-A2-45 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Spec (ES mirror es.mlgame.org) | Encyclopaedia MyLands: Spec | 2026-10-31 | Hero specialisations (Cleric, Miner, Architect, Benefactor) with 3 levels | Official | Header garbled in excerpt |
| SRC-A2-46 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Kingdom | Encyclopaedia MyLands: Kingdom | 2026-10-31 | Coordinates (kingdom X-Y + square) | Official | Map size not given |
| SRC-A2-47 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=City (IT/RU mirrors) | Encyclopaedia MyLands: City | 2026-10-31 | Capital 51 cells; +8 cities; 22 cells; square-dependent size; capital transfer | Official | Old RU says +7 |
| SRC-A2-48 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Ruin | Encyclopaedia MyLands: Ruin | 2026-10-31 | 1–7 ruins/kingdom + Archruin; aggressive ruins; respawn | Official | ~2,600 days old |
| SRC-A2-49 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Pit (steam wiki-en Pit) | Encyclopaedia MyLands: Pit | 2026-10-31 | 5 levels; lifespan collapse; tunnel 5 min / 6 h | Official | Lifespan value missing |
| SRC-A2-50 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Lake | Encyclopaedia MyLands: Lake (Salt Lake) | 2026-10-31 | Flying island, 4 cells, BG production, no gate | Official | — |
| SRC-A2-51 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Grail | Encyclopaedia MyLands: Grail | 2026-10-31 | 200 BG/hour; daily random relocation destroys buildings | Official | — |
| SRC-A2-52 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=BlackPearl | Encyclopaedia MyLands: BlackPearl | 2026-10-31 | Premium currency uses, premium perks, gifting | Official | — |
| SRC-A2-53 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Gate (RU mirror) | Encyclopaedia MyLands: Gate | 2026-10-31 | Gate open 5 min / close 6 h; Ancient Magic 30 min; not after attack launched | Official | — |
| SRC-A2-54 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Underground_estates | Encyclopaedia MyLands: Underground_estates | 2026-10-31 | Existence of underground estates | Official | No content retrieved |
| SRC-A2-55 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Quests (mlgame.co.uk/wiki Quests) | Encyclopaedia MyLands: Quests | 2026-10-31 | Counsellor quests, reward guards, Ancients 60 days, medals, random events 19:00, monsters refuse Arch-ruins, Temple ≤5 | Official | Mixed versions |
| SRC-A2-56 | https://steamdb.info/app/290730/ ; https://mmos.com/review/my-lands-black-gem-hunting | SteamDB / MMOs.com review | 2026-10-31 | Listed in search as secondary descriptions (not used for rules) | Secondary | Not opened |
| SRC-A2-57 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Tournament | Encyclopaedia MyLands: Tournament | 2026-10-31 | Flags Hunting & Treasure Hunt rules and prizes | Official | Versions differ |
| SRC-A2-58 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Clan | Encyclopaedia MyLands: Clan | 2026-10-31 | Clan requirements; 5–100 members; types; roles | Official | — |
| SRC-A2-59 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Clan_warehouse | Encyclopaedia MyLands: Clan_warehouse | 2026-10-31 | Clan warehouse/reservation in Pit; loot routing; robbable | Official | Table values may be dated |
| SRC-A2-60 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Auction (mlgame.co.uk/wiki Auction) | Encyclopaedia MyLands: Auction | 2026-10-31 | BG auction: 10M rating, 5,000 step, +1h anti-snipe, 48h pay, 7-day ban | Official | — |
| SRC-A2-61 | https://uk.mlgame.org/wiki-uk/Wiki.jsp?page=Rating | Encyclopaedia MyLands: Rating | 2026-10-31 | Rating window, tournament tab | Official | Old |
| SRC-A2-62 | https://www.youtube.com/watch?v=g5vrbLmC0MY | YouTube: "My Lands: Black Gem Hunting - First Impressions Gameplay" | 2026-10-31 | Candidate for OBSERVED evidence (not viewed) | Community | Not accessed |
| SRC-A2-63 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Spells | Encyclopaedia MyLands: Spells | 2026-10-31 | Book of Spells, casting, replacement rule, cooldown | Official | No cost table |
| SRC-A2-64 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=SpellName | Encyclopaedia MyLands: SpellName | 2026-10-31 | Spell effect examples; default range one kingdom | Official | — |
| SRC-A2-65 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Alchemy (uk/es/it mirrors) | Encyclopaedia MyLands: Alchemy | 2026-10-31 | Perfection recipe, runes ≤5, epic crystals, 7 levels | Official | EN/ES differ |
| SRC-A2-66 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=HeroTemp | Энциклопедия Mylands: HeroTemp (RU) | 2026-10-31 | Monster XP per ruin level table | Official (RU) | Excerpt only |
| SRC-A2-67 | https://mlgame.co.uk/forum/viewtopic.php?p=4767 | Forum: "Hero XP" | 2026-10-31 | Arena ~100 XP/h at L7 | Community | Anecdotal |
| SRC-A2-68 | https://mlgame.us/forum/viewtopic.php?t=681 | Forum: "best Hero for beginner" | 2026-10-31 | Class names (Strategist, Governor, Scientist, Wizard); commander cost 20M science points | Community | Player opinion |
| SRC-A2-69 | https://steamcommunity.com/app/290730/discussions/0/622955136048237252 | Steam discussion: "Managing Heroes" | 2026-10-31 | Hero management discussion (listed, not used for rules) | Community | Not opened |
| SRC-A2-70 | https://mlgame.co.uk/forum/viewtopic.php?t=1261 | Forum: "suzerain" | 2026-10-31 | Suzerain = level 5; 650M points | Community | — |
| SRC-A2-71 | https://mlgame.us/forum/viewtopic.php?f=11&t=194 | Forum: "How to choose the race!" | 2026-10-31 | Faction loyalty can be negative (claim) | Community | Player claim |
| SRC-A2-72 | https://www.forum.mlgame.co.uk/viewtopic.php?p=7555 | Forum: "Hero Captured" | 2026-10-31 | Anecdote of 5-day imprisonment | Community | Anecdotal |
| SRC-A2-73 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Brotherhood | Encyclopaedia MyLands: Brotherhood | 2026-10-31 | Sword Brotherhood bounty vs aggressors (7 days) | Official | Mechanics thin |
| SRC-A2-74 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Serv (RU ru.mlgame.org/wiki-ru Serv) | Encyclopaedia MyLands: Serv | 2026-10-31 | Server types; Tactical vs Mining differences; Faction War mode; XP multipliers | Official | Heroic/Strategic/Operative undocumented |
| SRC-A2-75 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Monarchic | Encyclopaedia MyLands: Monarchic | 2026-10-31 | Monarchic clan type (roles) | Official | Excerpt only |
| SRC-A2-76 | https://mlgame.co.uk/wiki/Wiki.jsp?page=JointAttack | Encyclopaedia MyLands: JointAttack | 2026-10-31 | One joiner; 50/50 loot; cascade withdrawal; per-player rank check; 3+owner defence | Official | — |
| SRC-A2-77 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Alliance | Энциклопедия Mylands: Alliance (RU) | 2026-10-31 | Alliance ≤7 members; cap by Diplomacy level | Official (RU) | RU only |
| SRC-A2-78 | https://mlgame.co.uk/wiki/Wiki.jsp?page=Mentor | Encyclopaedia MyLands: Mentor | 2026-10-31 | Mentor/apprentice; daily resource send | Official | — |
| SRC-A2-79 | https://mlgame.co.uk/wiki-uk/Wiki.jsp?page=Rules (uk/us/es/it mirrors) | Encyclopaedia MyLands: Rules | 2026-10-31 | Conduct rules; gag 1–7 days; bans; shared-IP rules | Official | Versions differ |
| SRC-A2-80 | https://fr.mlgame.org/wiki-fr/Wiki.jsp?page=Add_contact ; https://fr.mlgame.org/wiki-fr/Wiki.jsp?page=Filtr_Mes | Encyclopédie MyLands (FR): Add_contact, Filtr_Mes | 2026-10-31 | Contacts and message filters | Official (FR) | — |
| SRC-A2-81 | https://ru.mlgame.org/wiki-ru/Wiki.jsp?page=Tournament | Энциклопедия Mylands: Tournament (RU) | 2026-10-31 | Tournament versions/prizes | Official (RU) | — |

## Page-visit tally
- Distinct wiki/forum/store pages evidenced via indexed excerpts: ~75 (rows above, excluding SRC-A2-39,
  -56, -62, -69 which were listed but yielded no rule evidence).
- Pages opened directly in a browser/fetcher: 0 (blocked by network policy).
- Searches run: 36.
