# World Map and Exploration — Research Notes (Agent A2)

Subject: public materials about "My Lands: Black Gem Hunting" (MLG). Systems only; names are cited to
identify the reference mechanic and must not be reused.
Access date for all sources: **2026-10-31**. Source IDs → `_sources_A2.md`.
Method caveat: all evidence is search-excerpt level (direct page fetch blocked by network policy); see
the method note in `ARMY_AND_COMBAT.md`. Confidence capped at med unless multiple mirrors agree.

---

### WLD-001 — Map hierarchy: Empire → Kingdom → Square
- Description: The world ("Empire map") is a grid of Kingdoms. A Kingdom is addressed by two numbers
  (e.g. "36-11"); a third number addresses a Square inside the Kingdom. Domains (cities, outer domains,
  ruins, etc.) sit on squares.
- Game purpose: Coarse-to-fine navigation; local neighbourhoods (kingdoms) create competition clusters.
- Inputs: coordinates (kingdom X-Y, square N).
- Rules & constraints: kingdoms containing your domains are highlighted on the Empire map; the current
  kingdom is outlined. Capital coordinates of any player are visible via ratings window / contacts.
- Output: addressable map locations.
- Related systems: MSN-001 (missions target domains on the Kingdom Map), WLD-002.
- Evidence: SRC-A2-46 ("Kingdom coordinates are specified with two figures, for example, 36-11. With the
  third figure of coordinates a Square in the Kingdom is specified").
- Status: VERIFIED (med). Map dimensions and squares-per-kingdom UNKNOWN.
- Test plan: coordinate parser/formatter round-trip; every domain belongs to exactly one (kingdom, square).

### WLD-002 — Domain types (taxonomy)
- Description: "Domain" is the generic owned/attackable location. Types seen: City (Capital + other
  cities), Villages (attached to a city), Outer Domains (Pits, Stone/Iron/Gold Mines, Salt Lakes / flying
  islands, Grail), Ruins (monster domains, incl. Archruin), Camps (mobile), Barracks & Underground estates,
  Clan Castles.
- Game purpose: Varied targets with distinct risk/reward and rules.
- Evidence: SRC-A2-12 (OuterDomains); SRC-A2-01 (overview: "Ruins, Pits, Salt Lakes, and Mines");
  SRC-A2-30, SRC-A2-10, SRC-A2-54, SRC-A2-31.
- Status: VERIFIED (med).
- Test plan: each domain type has a defined rule profile (capturable? has gate/tunnel? rank rule?
  lifespan? building cells?).

### WLD-003 — Cities: capital, extra cities, cells, villages
- Description: First city = Capital: always 51 building cells, cannot be removed; capital status can be
  moved via a "Capital Transfer" science. Additional cities start with 22 cells; their size/level depends
  on the square chosen. Up to 8 more cities can be founded via science (Colonisation science → Colonisation
  Centre building; Settler sent on a City Foundation mission). Each city has 4 villages; village buildings
  work 30% faster; neighbouring cities can compete for villages.
- Game purpose: Expansion ("second cities") as a mid-game goal; square choice rewards exploration.
- Inputs: science level, settler, target square quality.
- Rules & constraints: max cities = 1 + 8 (older RU page says +7). Buildings inside the City Wall cannot be
  destroyed by enemies (MSN-008).
- Output: new city with its own garrison, buildings, villages.
- Related systems: WLD-007 (exploration to find squares), MSN-001 (City Foundation), economy.
- Evidence: SRC-A2-47 (City: "always has 51 cells"; "up to 8 more cities"; "size and level will depend on the
  square"); SRC-A2-15 (Villages: "There are 4 villages in every City"; 30% faster); SRC-A2-22 (Colonisation).
- Status: VERIFIED (high for 51 cells / +8 — EN, IT, new RU agree; med for others).
- Test plan: founding the 10th city fails; capital deletion fails; new city cell count derived from square
  attribute.

### WLD-004 — Outer Domains & seizure
- Description: Secondary domains (Pits, Mines, Salt Lakes, Grail) that can be seized after studying the
  Outer Domains science. Invisible until found by Explorers.
- Rules & constraints: seizure requires no remaining defensive buildings; on seizure level-1 buildings are
  destroyed, higher-level ones partially damaged; robbery of an outer domain yields at most 50%; Rank rule
  does not apply (any rank may attack Mine/Pit/Salt Lake/Camp). Max number owned: UNKNOWN.
- Evidence: SRC-A2-12 ("Outer Domains and Ruins are invisible until your Explorers discover them"); SRC-A2-27.
- Status: VERIFIED (med).
- Test plan: undiscovered domain not targetable; seizure blocked by defensive building.

### WLD-005 — Ruins (monster domains, Black Gem source)
- Description: NPC monster domains of levels (up to 7, plus Archruin). Black Gems can be stolen from
  them; artefacts too if a hero participates.
- Rules & constraints: 1–7 ruins per kingdom plus 1 Archruin. After a successful attack a ruin disappears
  and a new one spawns elsewhere. On Mining/Tactical servers ruins of levels 5–7 can be "aggressive" (a
  monster hero leads attacks on players). Monsters refuse to attack Archruins (level 7). Explorers only find
  ruins of level = rank, rank+1, and all lower; level-7 ruins findable by any rank.
- Game purpose: PvE content, renewable premium-currency faucet, rank-gated difficulty.
- Evidence: SRC-A2-48 (Ruin: "There might be 1 up to 7 ruins in each kingdom (plus 1 Archruin)"); SRC-A2-33;
  SRC-A2-55 (Quests: monsters refuse to attack Arch-Ruins); SRC-A2-23 (artefact theft needs hero).
- Status: VERIFIED (med). Ruin garrison sizes, gem amounts UNKNOWN.
- Test plan: kingdom spawn invariant (1..7 + 1 arch); defeat → despawn + respawn elsewhere; discovery filter
  by rank.

### WLD-006 — Black Gem economy on the map
- Description: Premium currency obtainable for real money or in-game: stolen from monsters in Ruins,
  produced by the main building of Salt Lakes (hovering islands), produced by the Grail, bought at auction
  for gold, gifted between players. Spent on premium services.
- Rules & constraints: Salt Lake: 4 building cells + villages, main building produces Black Gems, no
  Gate/Tunnel → attackable at any time. Grail: main building produces Black Gems at 200/hour; relocates
  randomly once a day, destroying its buildings when it moves. City teleport reported to cost 99 BG, instant,
  blocked by incoming attacks within 6 h or any missions (low).
- Game purpose: Contested map hot-spots generate PvP around the premium resource (the game's namesake).
- Evidence: SRC-A2-01 ("stolen from Monsters in Ruins, produced in Salt Lakes of flying islands, and bought
  at the auction for the in-game Gold"); SRC-A2-50 (Lake); SRC-A2-51 (Grail 200/h); SRC-A2-52 (BlackPearl:
  premium, gifting).
- Status: VERIFIED (med); teleport INFERRED (low — source page not identified).
- Test plan: Salt Lake has no gate → attack allowed anytime; Grail relocation clears buildings; gem
  production accrues per hour.

### WLD-007 — Exploration mission
- Description: Explorers go to empty squares to discover Ruins/Outer Domains or evaluate a square for a
  city.
- Rules & constraints: death chance 2–10% depending on chosen speed (slower = safer); rank filter (WLD-005).
  Premium includes automatic exploration.
- Evidence: SRC-A2-33 ("Explorers are sent with the 'Exploration' mission to the empty Kingdom Squares in
  order to find Ruins and Outer Domains or to choose a good square"); SRC-A2-52.
- Status: VERIFIED (med).
- Test plan: explorer loss probability within [2%,10%] and monotonic in speed.

### WLD-008 — Pits (temporary underground domains)
- Description: Underground domains with empty cells for building; 5 levels; finite lifespan — when it
  ends, the Pit collapses with its resources and buildings. Protected by a Tunnel (not a Gate). Clan Head can
  build Clan Warehouse / Clan Reservation in a Pit.
- Evidence: SRC-A2-49 ("Pits can be of 5 levels"; collapse at end of lifespan); SRC-A2-59.
- Status: VERIFIED (med). Lifespan length UNKNOWN.
- Test plan: pit expiry deletes everything inside.

### WLD-009 — Mines
- Description: Stone, Iron, Gold Mines are underground outer domains with a Tunnel; produce stone/iron/gold.
  Any rank may attack mines (rank rule exempt).
- Evidence: SRC-A2-12; SRC-A2-28; SRC-A2-27.
- Status: VERIFIED (med).

### WLD-010 — Gates and Tunnels (time-window defence)
- Description: Cities have a Gate; Pits/Mines a Tunnel. A closed gate blocks incoming attacks (no battle).
- Rules & constraints: open takes 5 minutes; close takes 6 hours; "Ancient Magic" can close a Gate in 30
  minutes but not once an attack has been launched; forum says speed-up not usable for tunnels.
  Robbery of clan warehouse possible whenever the tunnel is open.
- Game purpose: Player-controlled protection with commitment cost (slow close) — anti-griefing lever.
- Evidence: SRC-A2-53 (Gate: "opens 5 minutes and closes 6 hours"); SRC-A2-49; SRC-A2-14 (forum: no battle if
  gate closed before arrival).
- Status: VERIFIED (med). What a closed gate prevents for the owner (e.g. outgoing missions?) UNKNOWN.
- Test plan: state machine OPEN→CLOSING(6h)→CLOSED; CLOSED blocks arrivals; quick-close forbidden after an
  incoming attack is launched.

### WLD-011 — Camps (mobile forward bases)
- Description: Mobile domain used as a springboard for quick attacks; created by the hero mission
  "Encamping" on a square suitable for a city.
- Rules & constraints: requires a hero; troops/hero cannot leave and hero can't cast while founding;
  carried gold becomes the camp bank = troops' salary for its lifetime; cannot be upgraded or built on;
  destroyed by Domain Destruction or by owner (not while missions outgoing); count limited by Camp science;
  targetable by Robbery/Slaves Capture; rank rule exempt.
- Evidence: SRC-A2-30 ("Camp is a mobile domain that is used as springboard for quick attacks").
- Status: VERIFIED (med).
- Test plan: encamp on unsuitable square → fails with message; camp bank decreases by upkeep; destruction
  rules.

### WLD-012 — Barracks / Underground estates
- Description: Hidden underground domains seized with "Underground domains" science; 43 cells, 8–18 with
  level-7 military buildings; neutral garrison; give training discounts. "Underground estates" page exists
  (details not retrieved).
- Evidence: SRC-A2-10; SRC-A2-54.
- Status: VERIFIED (med) / UNKNOWN for estates.

### WLD-013 — Clan Castle (contested clan objective)
- Description: Clan-owned domain with an Internal Gate. Attack sequence: fortification round → rounds vs
  Internal Gate (defender troops absent; towers fire; gate "assimilates 100000 attack points in the first
  round") → defender troops join once gate strength hits 0 (e.g. gate falls round 17 → troops join round 18);
  20-round limit. Seizure only if no defending buildings remain; all buildings except Main Building can be
  destroyed. Main Building yields War and Peace points per hour → buy Clan Standards, which other clans can
  steal via Robbery. Clan members and their allies can launch joint Seizure.
- Evidence: SRC-A2-31.
- Status: VERIFIED (med); numbers med.
- Test plan: troops cannot be engaged while internal gate HP > 0; seizure blocked by any defensive building.

### WLD-014 — Dungeon (prison building) — map interaction
- See HERO-007. Captured heroes return by Escape mission at Fast march speed.

### WLD-015 — Monster attacks on players
- Description: Monsters (incl. aggressive ruins) raid players; a quest reward gives protection from new
  monster attacks (not from ones already en route); Ancients monster army reinforces newbie cities.
- Evidence: SRC-A2-27; SRC-A2-40; SRC-A2-48.
- Status: VERIFIED (med).
- Test plan: protection flag blocks new monster mission creation only.

---

## Unknowns & discrepancies

1. **Map size**, number of kingdoms, squares per kingdom, distance metric and travel-time formula: UNKNOWN.
2. **Max cities**: +8 (EN/IT/new RU) vs +7 (older RU). Treat +8 as current.
3. **Races count**: Steam store text says players choose "one of four races"; wiki Factions page lists 7
   race names across Light/Dark. Possibly races added later or some are NPC/premium. Verify on Race page.
4. Outer-domain ownership limits, Pit lifespan, ruin respawn timing, Black Gem quantities in ruins: UNKNOWN.
5. City teleport (99 BG) — source page not identified; low confidence.
6. Server types (Mining, Tactical) change attack rules; the full list of server types is UNKNOWN.
7. Black Gem dedicated article exists per Resources page but was not retrieved.
8. Verification route: allow hosts mlgame.co.uk / steam.mlgame.org and read pages Kingdom, City, Ruin,
   Pit, Lake, Grail, Gate, Camp, OuterDomains, Clan_castle; watch gameplay videos to OBSERVE map UI.
