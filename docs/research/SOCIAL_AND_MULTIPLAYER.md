# Social and Multiplayer Systems — Research Notes (Agent A2)

Subject: public materials about "My Lands: Black Gem Hunting" (MLG). Systems only; do not reuse names.
Access date for all sources: **2026-10-31**. Source IDs → `_sources_A2.md`.
Method caveat: search-excerpt-level evidence only (direct fetch blocked) — see `ARMY_AND_COMBAT.md`.

---

## A. PvP rules and protection

### SOC-001 — Rank-based attack restriction ("Rank Rule")
- Description: Players may attack Ruins/Cities of players with a higher (or, by implication, equal) rank;
  on Mining and Tactical servers they cannot attack Ruins/Cities of lower-rank players.
- Rules & constraints: exemptions — Outer Domains and Camps; any rank may attack any Mine, Pit, Salt Lake or
  Camp. In joint attacks rank is checked separately for initiator and each joiner. The phrase "the required
  rank is Suzerain for Mining Servers and Tactical servers" is ambiguous (see Unknowns).
- Game purpose: Anti-bullying: stronger players cannot farm weaker ones; weaker can punch up.
- Inputs: attacker rank, target owner rank, target domain type, server type.
- Output: allow/deny mission launch.
- Related systems: HERO-011 (rank), SOC-002, SOC-012.
- Evidence: SRC-A2-27 ("On Mining Server and Tactical Server you cannot attack Ruins and Cities of players
  with the ranks (levels) lower than yours"; "Rank Rule doesn't cover Outer Domains and Camps").
- Status: VERIFIED (med).
- Test plan: matrix test over (attacker rank × target rank × domain type × server type).

### SOC-002 — Revenge exception
- Rule: if a lower-rank player attacked your city, you get a one-time right to attack any city of that
  attacker within 24 h of the attack.
- Evidence: SRC-A2-27 (US mirror: "you receive a one-time possibility to get revenge ... during 24 hours").
- Status: VERIFIED (med).
- Test plan: revenge token created on being attacked by lower rank; consumed on use; expires at 24 h.

### SOC-003 — Newbie protection layers
- Description: (a) A "compassion" spell is cast on a player's Capital after reaching a certain rank (effect
  not documented; spell list associates Compassion with making enemy troops retreat — INFERRED);
  (b) "Ancients": very large monster reinforcement placed in new players' assistance slots for 60 days
  (shield icon on the city); (c) a quest reward that blocks new monster attacks; (d) Gate closing (WLD-010).
- Game purpose: Onboarding safety without a hard invulnerability timer.
- Evidence: SRC-A2-27 ("After a player reaches certain rank, the spell of compassion is cast upon the
  player's Capital"); SRC-A2-40 (Ancients); SRC-A2-55 (Quests: "Ancients stay at your Assistance slots
  for 60 days"); SRC-A2-53.
- Status: VERIFIED for existence (med); compassion mechanics UNKNOWN.
- Test plan: new account has protective NPC reinforcement for 60 days; removal after expiry.

### SOC-004 — Attack frequency cap
- Rule: max 2 attacks on the same domain per 24 h (see MSN-004). Status VERIFIED (med). Evidence SRC-A2-06.

### SOC-005 — Factions and faction loyalty rating
- Description: Races belong to Light or Dark faction; a player who lost faction support is a "Renegade".
  Faction Loyalty = gold value of enemy units killed against the opposite faction or Renegades; attackers
  only; counted after resurrection; assisting troops' faction ignored; assisting monsters don't count;
  joint-attack points split by attack share at battle start. Forum claims it can go negative when attacking
  own faction (low).
- Game purpose: Steers PvP toward cross-faction conflict.
- Evidence: SRC-A2-18 ("Gold equivalent of units killed in battle against enemy faction player ... or
  against Renegade of any faction"); SRC-A2-71 (forum, negative rating).
- Status: VERIFIED (med); negative values INFERRED (low).
- Test plan: same-faction kill gives 0 (or negative per design); post-resurrection counting.

### SOC-006 — Sword Brotherhood (bounty system)
- Rule: hire and reward other players for damaging aggressors who successfully attacked you within the
  last 7 days.
- Evidence: SRC-A2-73 (Brotherhood page).
- Status: VERIFIED (med); payment mechanics UNKNOWN.
- Test plan: bounty only on aggressors within 7-day window.

### SOC-007 — Temple protection (Tactical servers)
- Rule: in the Temple a player can ask for protection from at most 5 players (Tactical servers).
- Evidence: SRC-A2-74 (Serv page); SRC-A2-55.
- Status: VERIFIED (low); mechanism UNKNOWN.

---

## B. Clans, alliances and joint operations

### SOC-010 — Clans
- Description: Player organisation requiring the "Clans" science and a Diplomatic Corps building; join by
  registering or invitation from the clan head. 5–100 members. Two governance types: Monarchic and
  Democratic. Roles: Clan Head; Treasurer and Secretary (assigned by head); Treasurer can admit/expel and
  access the clan warehouse; War Minister can admit/expel but not act against the Head.
- Game purpose: Long-term social retention; collective objectives.
- Evidence: SRC-A2-58 (Clan: "A clan can consist of 5 to 100 members"; "two types of Clans: Monarchic and
  Democratic"); SRC-A2-75 (Monarchic page).
- Status: VERIFIED (med).
- Test plan: membership bounds; role permission matrix.

### SOC-011 — Clan warehouse & reservation
- Description: Clan Head can build a Clan Warehouse and Clan Reservation in a Pit. Robbery loot by members
  goes to the warehouse instead of being "burned"; head distributes contents at discretion. Others can rob it
  while the Pit tunnel is open.
- Evidence: SRC-A2-59; SRC-A2-28 ("burnt during robberies but not for the players in clan").
- Status: VERIFIED (med).
- Test plan: member robbery routes surplus loot into clan storage; storage robbable when tunnel open.

### SOC-012 — Clan Castle, War & Peace points, Standards
- Description: Clan Castle main building yields War and Peace points hourly; points buy Clan Standards,
  which other clans can steal by Robbery. Castle siege rules in WLD-013. Arena in castle gives hero XP to
  allied-slot heroes. 3 players can jointly attack; 3 can defend in ally slots.
- Evidence: SRC-A2-31; SRC-A2-76 (JointAttack).
- Status: VERIFIED (med).

### SOC-013 — Alliances (Diplomacy)
- Description: Diplomacy science allows building the Diplomatic Corps and forming an alliance with several
  players; ally count grows with science level; RU page: alliance up to 7 members.
- Evidence: SRC-A2-22 (Diplomacy science text); SRC-A2-77 (RU Alliance page).
- Status: VERIFIED (med); 7-member cap low (RU only).
- Test plan: ally count ≤ f(science level) ≤ 7.

### SOC-014 — Joint attack
- Rule: initiator invites allies/clan-mates; only ONE player can actually join; rank rule checked per player;
  resources and Black Gems split 50/50; if initiator withdraws all squads, joined squads are withdrawn too.
- Evidence: SRC-A2-76 ("Only one player will actually be able to join the attack"; "split in half").
- Status: VERIFIED (med).
- Test plan: second joiner rejected; loot split; cascade withdrawal.

### SOC-015 — Joint defence (Assistance)
- Rule: up to 3 players plus the owner can defend a domain jointly; sent via "Assistance" from the domain's
  map menu; assistance travels 2× attack speed; damage shared proportionally (CMB-007).
- Evidence: SRC-A2-32; SRC-A2-76.
- Status: VERIFIED (med).
- Test plan: 4th assisting player rejected.

### SOC-016 — Mentor / Apprentice
- Rule: mentors advise apprentices, may send resources to them once a day, and can view their cities.
- Evidence: SRC-A2-78 (Mentor page).
- Status: VERIFIED (med).
- Test plan: once-per-day transfer limiter.

---

## C. Economy between players

### SOC-020 — Market (player-to-player resources for gold)
- Description: Players sell resources to each other for Gold at the Market. Selling needs no Merchants;
  buying needs Merchants (trained in a Merchant Guild) plus Market building and science. "Exchange" tab
  swaps resources at an inflated rate (NPC-like sink). Resources on sale can be stolen by robbers like
  warehouse stock.
- Game purpose: Specialisation by race (resource bonuses) → trade.
- Evidence: SRC-A2-29 ("No Merchants are needed to sell resources"; market stock robbable).
- Status: VERIFIED (med). Fees, price bounds UNKNOWN.
- Test plan: buy requires merchants; listed stock counts as robbable.

### SOC-021 — Auction (premium currency for gold)
- Description: Black Gems are auctioned for in-game Gold. Access requires Main Rating ≥ 10,000,000; min
  bid increment 5,000 Gold; any bid in the last hour extends the lot by 1 hour; winner must pay within 48 h
  or gets a 7-day auction ban; players cannot sell Black Gems themselves (lots supplied by the game —
  INFERRED).
- Game purpose: Lets non-paying players obtain premium currency; gold sink.
- Evidence: SRC-A2-60 (Auction: "Players can't sell the Black Gems!").
- Status: VERIFIED (med).
- Test plan: anti-snipe extension; payment deadline → ban; access gate by rating.

### SOC-022 — Resource transfers & gifting
- Description: Resources Transportation moves goods between own domains only; requires Merchants and an
  open Gate/Tunnel at both ends; merchants die after completion with 3% probability. Black Gems can be gifted.
- Evidence: SRC-A2-42 (MoveResources); SRC-A2-52 (BlackPearl).
- Status: VERIFIED (med).
- Test plan: transfer blocked if either gate closed; 3% merchant loss RNG.

---

## D. Communication

### SOC-030 — Messages (mail), contacts, reports
- Description: In-game mailbox (envelope button), contact list ("Add to contacts"), message filters,
  saved spy reports, system notices (random events). "Report" button for abuse/spam. Premium raises mail
  and contact capacity. Message retention period UNKNOWN. Real-time chat: not confirmed in reachable
  excerpts (UNKNOWN).
- Evidence: SRC-A2-43 (Mes); SRC-A2-79 (Rules); SRC-A2-80 (FR Add_contact / Filtr_Mes); SRC-A2-52 (Premium).
- Status: VERIFIED (med) for mail; UNKNOWN for chat.
- Test plan: contact add/remove; filter states; capacity limit by premium.

### SOC-031 — Conduct rules & sanctions
- Rule: abuse, coarse language, spam, multi-accounting banned; sanctions: "gag" (writing ban) for 1–7 days
  on reasonable report, bans for repeated reports; unreasonable reporting → warning. Players sharing a
  computer/IP must notify admins and their characters may not interact (no attacks, joint attacks,
  assistance or trade between them — INFERRED from excerpt).
- Evidence: SRC-A2-79 (Rules).
- Status: VERIFIED (med).
- Test plan: moderation tooling requirements; interaction block between linked accounts.

---

## E. Events, quests, tournaments, server modes

### SOC-040 — Quests (Counsellor)
- Description: Quests given by an advisor NPC; skippable without losing later quests; reward claimed by
  returning to the advisor; some rewards (troops, instant construction) blocked if a battle in the Capital is
  < 2 h away; same-value timed bonuses extend instead of stack; quests grant medals/orders.
- Evidence: SRC-A2-55 (Quests).
- Status: VERIFIED (med).
- Test plan: reward claim guard on imminent battle; bonus extension logic.

### SOC-041 — Tournaments
- Flags Hunting: 10 days; ranked by number of Flag Sets collected; tie → who collected faster; prizes =
  artefacts (L4 for 1st, L3 for 2–6, L2 + L3 scroll for 7–26 in one version; newer version L5 set item for
  1st); top-100 get a medal. Treasure Hunt: Suzerain only; +3 points for revealing a fragment in a new cell,
  +1 in an explored cell; full treasure = 200 points to first finder; prizes L5 (1st), L4 (2–3), L3 (4–10);
  ≥3 points outside top-100 → L3 scroll. Rewards claimed via advisor tab; Vault required.
- Evidence: SRC-A2-57 (Tournament EN); SRC-A2-81 (Tournament RU); SRC-A2-61 (Rating).
- Status: VERIFIED (med); prize tables differ by version.
- Test plan: leaderboard ordering with time tie-break; eligibility filter by rank.

### SOC-042 — Random events
- Rule: sometimes at 19:00 players get a positive or negative random event, notified by in-game mail.
- Evidence: SRC-A2-55.
- Status: VERIFIED (low).

### SOC-043 — Server types and Faction War mode
- Description: Servers differ in rules: Mining, Combat/War, Faction War, Economic, Tactical, Heroic,
  Strategic, Operative (RU list). Tactical vs Mining: no troop salary in domains, doubled resource output,
  richer deposits, Temple protection ≤5 players, daily ruin-destruction limit 20. Hero XP multiplier on attacker
  win: 0.0002 Mining vs 0.0005 War. Faction War: entry after a 3-quest chain on Mining/Combat server; goal
  destroy all enemy-faction capitals; survivors split prize pool; server closes 3 days then removed.
- Evidence: SRC-A2-74 (Serv EN/RU); SRC-A2-19 (Heroes XP multipliers).
- Status: VERIFIED (med).
- Test plan: ruleset configuration object per server type; config-driven tests.

### SOC-044 — Premium / monetisation touchpoints
- Description: Black Gems buy premium (auto-exploration, research/construction queues, bigger mail/contact
  limits), extendable anytime; gifting allowed.
- Evidence: SRC-A2-52.
- Status: VERIFIED (med).

---

## Unknowns & discrepancies

1. Rank-rule wording on Mining/Tactical servers ("required rank is Suzerain") is ambiguous — could mean the
   restriction applies only below Suzerain, or that attacking needs Suzerain. Verify on Rank page.
2. Compassion spell effect and the rank at which it is applied: UNKNOWN.
3. Real-time chat existence/channels: not confirmed. Message retention: UNKNOWN (FAQ has the question).
4. Alliance cap 7 from RU page only; EN says "several players".
5. Linked-account interaction ban: excerpt wording ambiguous between "linked accounts" and "allies";
   we interpreted as linked accounts (allies obviously can assist/joint-attack).
6. Market fees/limits and auction lot supply: UNKNOWN.
7. Tournament prize tables vary between page versions; event calendar/frequency UNKNOWN.
8. Faction rosters differ by language (ES omits Chimeras). Steam lists "four races".
9. Verification route: open Rank, Rules, Market, Auction, JointAttack, Clan, Tournament, Serv pages;
   Steam community discussions; YouTube gameplay to OBSERVE chat/mail UI.
