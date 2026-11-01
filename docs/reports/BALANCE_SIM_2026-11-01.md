# Balance simulation: first day

Date: 2026-11-01. Task: [T-022](../../tasks/T-022.md).
Status: baseline. The numbers describe a simple bot, not human players.

## What was built

- `server/internal/sim` plays a city with a bot player.
- The bot calls the same rules as the server: `AdvanceCity`, `CostAt`, `DurationAt`, `Resolve`.
- No database is used. A 24-hour run for one race takes tens of milliseconds.
- `server/cmd/sim` runs every race of a pack and writes the tables.
- Run it with `make sim`. The command writes [BALANCE_SIM_TABLES.md](BALANCE_SIM_TABLES.md).
- Two runs of the same pack and flags give identical output. A test checks this.

## Bot policy

The policy is generic. It names no resource, building or unit.

1. Rate each legal building level by income gain divided by cost.
2. Value income with `exchangeRates`. A resource without a rate counts as 1.
3. Add a bonus for housing when population nears the cap.
4. Add a bonus for storage when a stock is above 85% of its cap.
5. Add a bonus for food when the food balance is negative.
6. Buy the best affordable level. Wait only if the favourite is within 15 minutes.
7. After hour 4, build the chain for the unit with the best power per cost.
8. Spend 40% of stock per decision on that unit.
9. The wage-aware bot trains only while upkeep income covers the new units.

## What the simulation does not cover

- Research, heroes, items, quests, market and NPC exchange.
- PvP, alliances, second cities and territories.
- Rewards from PvE missions. The bot does not send armies. It only checks what its army could beat.
- Human choices. A better player gets different numbers.

The PvE table uses race modifiers only. Techs, heroes and items are absent.
Rewards for resources without `exchangeRates` count as 1 per unit. Those rows carry the mark †.

## Findings

Each finding is an observation from the tables. Each one needs a designer decision.

### Ashen Crowns

1. **Income grows 7 to 9 times in a day.**
   The value of net income rises from about 129 to between 900 and 1130 per hour.
2. **Saving, not timers, sets the pace.**
   The build queue is empty 94% to 96% of the time for every race.
   A bot waits 8 to 39 minutes for timber and stone before each purchase at hour 10.
3. **Gold is the bottleneck.**
   The wage-aware Thornkin bot nets +1.3 gold per hour at hour 24.
   Upkeep takes 94% of its gold production.
   A careless Thornkin bot reaches 232% and Ashborn reaches 260%.
   The bot invested little in gold buildings, so part of this is bot policy.
   The pack pays upkeep in gold and produces gold in few places.
4. **No desertion happens in 24 hours.**
   The gold stock never reaches zero in this horizon.
   The pressure shows as a rising wage share, not as desertion.
   A longer horizon is needed to test desertion.
5. **The first army arrives late for some races.**
   The first trained army arrives between hour 8.8 and hour 22.2.
   Gravewarden is last, at hour 22.2.
   Thornkin needs lore for its target unit, so the bot builds a lore producer first.
6. **Wolf Den level 1 is the entry point.**
   Four Thornweavers clear it with 25% losses. The army costs 460 value units.
   The army pays back after about 4.7 missions.
7. **Ruins and craters do not pay back by this valuation.**
   Their net value per mission is negative for every level.
   Their rewards include lore and starglass, which have no exchange rate.
   The result is a lower bound. Add rates before judging these sites.

### Void Frontier

1. **Population stalls at 61 of 220 from hour 16.**
   The pack uses growth mode `idle`. Growth needs idle people.
   Every inhabitant has a job, so nobody reproduces.
   Housing stays free. The bot built many staffed producers, so part of this is bot policy.
   A player faces the same rule. Decide whether this is intended.
2. **The first army arrives near hour 6.**
   This is much earlier than in Ashen Crowns.
3. **Nests above level 1 do not pay back.**
   Rewards are small against army cost.
   The mark † applies. Rates for the reward resources are missing.

## Verification

| Check | Result |
|---|---|
| `go vet ./internal/sim ./cmd/sim` | Pass |
| `go test -race -count=1 ./internal/sim` | Pass, 10 tests |
| Determinism: two `make sim` runs compared with `cmp` | Identical |
| Invariants on both packs and all six races: no negative stock, population within housing, queue within limit | Pass |

## Limits of these findings

- The thresholds in the warnings are PROPOSED review prompts. They are not game rules.
- The 30% loss limit and the 6-tile distance are PROPOSED defaults.
- Findings 3 and 5 for Ashen Crowns and finding 1 for Void Frontier depend on the bot policy.
- No finding here changes a rule, a pack or an evidence status.
