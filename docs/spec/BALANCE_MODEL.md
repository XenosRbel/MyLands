# Realmforge — Balance Model

Status: balance reference for the shipped packs **Ashen Crowns v1.0.0** (primary) and **Void Frontier v0.1.0** (comparison). Date: 2026-10-31.
Rules: every formula used here is defined in `GAME_RULES.md`. Rule IDs are given in brackets.

**How the tables were produced.** Every table in this document is generated from the pack JSON by
`/tmp/claude-0/-home-user/1547afcd-c6d4-5979-a753-a17e66a9906c/scratchpad/balance.py`, which re-implements `server/internal/rules` exactly, including Go's round-half-away-from-zero and the `ceil(x − 1e−9)` duration guard. Regenerate after any pack change:

```
python3 balance.py tables packs/ashen-crowns/pack.json packs/void-frontier/pack.json > tables.md
python3 build_balance_md.py   # splices tables.md into this document
```

Conventions:
- Costs are `round(base × growth^(L−1))` [R-BLD-2]. Times are `ceil(base × growth^(L−1))` with **no modifiers and world speed 1** [R-TIME-3]. In a real city, Keep levels (−2%/level), Academy and specialisations shorten them, down to the 0.2 floor.
- Effect values are `perLevel × L × growth^(L−1)` [R-ECO-1]. `modifier` effects scale **linearly**: value × L.
- "Reference value" uses `params.exchangeRates` as a common value scale (Ashen Crowns: timber 1, grain 1, stone 1.2, iron 1.6, gold 3). `lore` and premium resources have no rate and are listed separately.

---

## 1. Parameters

Ranges are from `pack/validate.go` (`params()`). "◆" marks a Void Frontier value that differs from Ashen Crowns. Values not set in a pack are the engine defaults from `pack.Defaults()`.

| Param | Type | Ashen Crowns | Allowed range (validate.go) | Unit | Gameplay impact | Void Frontier |
|---|---|---|---|---|---|---|
| `worldSpeed` | float | 1 | [0.1, 1000] | × | Multiplies all flows, timers, travel, respawn, tick length in game-hours; also × world row `speed`. | 1 |
| `economyTickSeconds` | int | 600 | [60, 3600 (must divide 86400)] | s | Granularity of population growth, starvation, desertion, overflow decay. | 600 |
| `buildQueueMax` | int | 3 | [1, 20] | jobs/city | Parallel planning depth for construction (jobs run sequentially). | 3 |
| `trainQueueMax` | int | 5 | [1, 50] | batches/city | Training batches that may be queued. | 5 |
| `researchQueueMax` | int | 2 | [1, 20] | jobs/player | Research jobs that may be queued (player-wide). | 2 |
| `cancelRefundPct` | int | 80 | [0, 100] | % | Refund on cancelling the last queued job; 100−x is a cancellation tax. | 80 |
| `mapWidth` | int | 48 | [8, 512] | tiles | World size; with density sets the PvE supply. | 32 ◆ |
| `mapHeight` | int | 48 | [8, 512] | tiles | World size. | 32 ◆ |
| `neutralDensityPct` | int | 12 | [0, 60] | % of passable tiles | PvE object density. | 12 |
| `newbieProtectionHours` | int | 72 | [0, 720] | h | PvP + scouting immunity after joining (ends early on own attack). | 72 |
| `minTravelSeconds` | int | 900 | [1, 3600] | s | Floor on one-way travel (÷ worldSpeed). | 300 ◆ |
| `combatRounds` | int | 20 | [1, 20] | rounds | Max rounds; attacker loses on timeout. | 8 ◆ |
| `lootPct` | int | 50 | [0, 100] | % | Share of unprotected stock that can be plundered (before carry cap). | 40 ◆ |
| `scoutDetectBasePct` | int | 30 | [0, 100] | % | Minimum detection chance on successful scouting. | 30 |
| `popGrowthPctPerHour` | float | 12 | [0, 100] | %/h | Population growth speed (of free housing or of idle). | 8 ◆ |
| `foodPerPopPerHour` | float | 0.12 | [0, 10] | food/pop/h | Population food consumption. | 0.1 ◆ |
| `starvationPctPerTick` | float | 1 | [0, 50] | %/tick | Population loss per tick while food stock is 0 and net < 0. | 1 |
| `foodResource` | resource id | grain | must exist |  | Consumed by population (and by army if upkeepResource unset). | biomass ◆ |
| `premiumResource` | resource id | starglass | optional |  | Declared premium currency (currently informational). | quanta ◆ |
| `marketFeePct` | int | 5 | [0, 50] | % | Sink on player trades, charged to seller proceeds. | 2 ◆ |
| `marketMaxOpenOrders` | int | 10 | [1, 100] | orders | Escrow exposure per player. | 10 |
| `exchangeRates` | map | timber 1, grain 1, stone 1.2, iron 1.6, gold 3 | >0 per resource | value/unit | Reference values for NPC exchange and market price band. | alloy 1, biomass 1, energy 1.2 ◆ |
| `exchangeFeePct` | int | 30 | [0, 90] | % | NPC exchange sink. | 25 ◆ |
| `allianceCreateCost` | amounts | gold 500 | 0..1e12 | units | Alliance founding sink (paid from capital). | — ◆ |
| `allianceMaxMembers` | int | 100 | [2, 500] | players | Alliance size cap. | 30 ◆ |
| `cityCapBuilding` | building id | keep | optional |  | Max level across cities gates extra cities. | command ◆ |
| `citiesPerCapLevels` | int | 5 | [1, 100] | levels | One extra city per N levels. | 5 |
| `settlerUnit` | unit id | settler | optional |  | Consumed to found a city. | colony ◆ |
| `territoryCapBuilding` | building id | keep | optional |  | Max level across cities gates territory count. | command ◆ |
| `territoriesPerLevel` | int | 1 | [0, 20] | tiles/level | Territory cap slope. | 1 |
| `territoryBonusPct` | int | 0 | [0, 100] | % | NOT USED by the engine (see GAME_RULES D-2). | 5 ◆ |
| `heroHireCost` | amounts | gold 300, grain 300 | 0..1e12 | units | Base hire cost, × (heroes owned + 1). | alloy 400 ◆ |
| `heroBuilding` | building id | warhall | required if classes |  | Hero cap and hiring location. | crew ◆ |
| `heroesPerBuildingLevel` | int | 1 | [0, 10] | heroes/level | Hero cap slope. | 1 |
| `heroWoundHours` | int | 6 | [0, 168] | h (real) | Hero downtime after a lost battle. | 6 |
| `attackScoreFloorPct` | int | 0 | [0, 100] | % | Anti-bullying: may not attack players below x% of own score (0 = off). | 0 |
| `upkeepResource` | resource id | gold | optional |  | Resource paying army upkeep (default foodResource). | biomass ◆ |
| `desertionPctPerHour` | float | 0.5 | [0, 100] | %/h | Garrison loss while upkeep unpaid. | 0.5 |
| `missionUpkeepMult` | float | 2 | [0, 10] | × | Upkeep multiplier for units away from home. | 1 ◆ |
| `popGrowthMode` | enum | free_housing | idle or free_housing |  | Base of population growth. | idle ◆ |
| `overflowDecayPctPerHour` | float | 1 | [0, 100] | %/h | 0 = hard storage cap; >0 = soft cap with surplus decay. | 0 ◆ |
| `defenseCapPct` | float | 80 | [0, 99] | % | Maximum damage reduction from defense (+wall). | 90 ◆ |
| `defenderRetreatPct` | int | 100 | [1, 100] | % losses | Defender withdraws at this cumulative loss (100 = never). | 100 |
| `sameRaceAttackPct` | float | -50 | [-95, 0] | % attack | Penalty attacking own race. | 0 ◆ |
| `sameFactionAttackPct` | float | -25 | [-95, 0] | % attack | Penalty attacking own faction (other race). | 0 ◆ |
| `maxAttacksPerTargetPerDay` | int | 2 | [0, 1000] | attacks/24h | Per attacker per target tile (0 = unlimited). | 0 ◆ |
| `maxReinforcingPlayers` | int | 3 | [0, 100] | players | Distinct foreign players stationed in one city. | 3 |
| `maxHeroes` | int | 7 | [0, 50] | heroes | Absolute hero cap. | 7 |
| `maxCities` | int | 9 | [1, 100] | cities | Absolute city cap. | 9 |
| `territoryYieldPerHour` | float | 25 | [0, 1e6] | units/h | Territory resource faucet. | 20 ◆ |
| `territoryYieldCapHours` | int | 24 | [1, 720] | h | Uncollected yield stops accruing after this. | 24 |

**Parameter design notes (Ashen Crowns):**
- `combatRounds 20` and timeout = attacker loss match the research (CMB-001). Void Frontier uses 8, which favours defenders more: a slow attacker hits the round limit sooner.
- `minTravelSeconds 900` matches the 15-minute minimum mission duration (MSN-003). Void Frontier's 300 makes nearby raids three times faster.
- `upkeepResource gold` with `missionUpkeepMult 2` and `desertionPctPerHour 0.5` reproduce the salary model (ECO-015). Void Frontier pays upkeep in its food resource, so army size is limited by food rather than by a separate currency.
- `overflowDecayPctPerHour 1` is the soft cap (ECO-013): storage never blocks production, the surplus melts at 1%/h. Void Frontier uses the hard cap (ECO-012).
- `maxAttacksPerTargetPerDay 2` (MSN-004) and the faction penalties −50/−25 (FAC-002) are the verified values. Void Frontier disables both.
- `territoryBonusPct` is unused by the engine (GAME_RULES D-2). `maxCities 9` is unreachable with Keep max 25 / 5 (D-16).

## 2. Resources

**Ashen Crowns**

| ID | Name | Kind | Capped | Base storage | Tradable | Lootable | Exchange value |
|---|---|---|---|---|---|---|---|
| `timber` | Timber | basic | yes | 1500 | yes | yes | 1 |
| `grain` | Grain | food | yes | 1500 | yes | yes | 1 |
| `stone` | Stone | basic | yes | 1500 | yes | yes | 1.2 |
| `iron` | Iron | basic | yes | 1000 | yes | yes | 1.6 |
| `gold` | Gold | rare | yes | 3000 | yes | yes | 3 |
| `lore` | Lore | basic | no | ∞ | no | no | — |
| `starglass` | Starglass | premium | no | ∞ | no | no | — |

**Void Frontier**

| ID | Name | Kind | Capped | Base storage | Tradable | Lootable | Exchange value |
|---|---|---|---|---|---|---|---|
| `alloy` | Alloy | basic | yes | 2000 | yes | yes | 1 |
| `biomass` | Biomass | food | yes | 2000 | yes | yes | 1 |
| `energy` | Energy | basic | yes | 2000 | yes | yes | 1.2 |
| `data` | Data | basic | no | ∞ | no | no | — |
| `quanta` | Quanta | premium | no | ∞ | no | no | — |

## 3. Buildings — Ashen Crowns

| Building | Max | Req. | L1 cost / time | L5 cost / time | L10 cost / time | L20 cost / time | Effect L1/L5/L10/L20 | Score/lvl |
|---|---|---|---|---|---|---|---|---|
| `keep` | 25 | — | timber 120, stone 120, gold 40 / 2m00s | timber 608, stone 608, gold 203 / 8m08s | timber 4,613, stone 4,613, gold 1,538 / 46m57s | timber 266,021, stone 266,021, gold 88,674 / 1d02h | +60 housing / +350.96 housing / +853.99 housing / +2,528.22 housing<br>build_time -2% / build_time -10% / build_time -20% / build_time -40% | 20 |
| `cottages` | 25 | — | timber 60, stone 20 / 1m00s | timber 244, stone 81 / 3m38s | timber 1,408, stone 469 / 18m10s | timber 46,949, stone 15,650 / 7h34m | +90 housing / +568.11 housing / +1,520.53 housing / +5,446.08 housing | 4 |
| `lumberyard` | 25 | — | timber 40, stone 30 / 50s | timber 145, stone 109 / 2m52s | timber 726, stone 545 / 13m16s | timber 18,187, stone 13,640 / 4h47m | +32 timber/h (6 workers) / +202.00 timber/h (30 workers) / +540.63 timber/h (60 workers) / +1,936.38 timber/h (120 workers) | 4 |
| `farmstead` | 25 | — | timber 45, stone 20 / 50s | timber 163, stone 73 / 2m52s | timber 817, stone 363 / 13m16s | timber 20,460, stone 9,094 / 4h47m | +36 grain/h (6 workers) / +227.25 grain/h (30 workers) / +608.21 grain/h (60 workers) / +2,178.43 grain/h (120 workers) | 4 |
| `quarry` | 25 | — | timber 50, grain 20 / 1m00s | timber 181, grain 73 / 3m26s | timber 908, grain 363 / 15m55s | timber 22,734, grain 9,094 / 5h44m | +26 stone/h (6 workers) / +164.12 stone/h (30 workers) / +439.26 stone/h (60 workers) / +1,573.31 stone/h (120 workers) | 4 |
| `ironworks` | 25 | keep 2 | timber 70, stone 50 / 1m20s | timber 269, stone 192 / 4m42s | timber 1,446, stone 1,033 / 22m41s | timber 41,834, stone 29,882 / 8h48m | +20 iron/h (8 workers) / +126.25 iron/h (40 workers) / +337.90 iron/h (80 workers) / +1,210.24 iron/h (160 workers) | 5 |
| `tradepost` | 20 | keep 2 | timber 80, stone 60 / 1m30s | timber 354, stone 265 / 5m46s | timber 2,267, stone 1,700 / 31m00s | timber 93,128, stone 69,846 / 14h56m | +0.060 gold/h per idle / +0.300 gold/h per idle / +0.600 gold/h per idle / +1.200 gold/h per idle | 5 |
| `mint` | 20 | keep 4, ironworks 2 | timber 120, stone 120, iron 60 / 2m30s | timber 530, stone 530, iron 265 / 9m37s | timber 3,400, stone 3,400, iron 1,700 / 51m40s | timber 139,693, stone 139,693, iron 69,846 / 1d00h | +14 gold/h (10 workers) / +88.37 gold/h (50 workers) / +236.53 gold/h (100 workers) / +847.17 gold/h (200 workers) | 6 |
| `scriptorium` | 20 | keep 2 | timber 90, stone 70, gold 30 / 2m00s | timber 398, stone 309, gold 133 / 7m41s | timber 2,550, stone 1,983, gold 850 / 41m20s | timber 104,769, stone 81,487, gold 34,923 / 19h55m | +10 lore/h (5 workers) / +68.02 lore/h (25 workers) / +199.90 lore/h (50 workers) / +863.14 lore/h (100 workers) | 6 |
| `academy` | 15 | scriptorium 3 | timber 200, stone 200, gold 100 / 5m00s | timber 1,013, stone 1,013, gold 506 / 22m07s | timber 7,689, stone 7,689, gold 3,844 / 2h21m | — | research_time -3% / research_time -15% / research_time -30% / — | 8 |
| `warehouse` | 25 | — | timber 60, stone 40 / 1m00s | timber 230, stone 154 / 3m32s | timber 1,240, stone 826 / 17m01s | timber 35,858, stone 23,905 / 6h36m | +1,200 storage (*) / +9,441.12 storage (*) / +33,276.95 storage (*) / +206,706.28 storage (*) | 3 |
| `vault` | 15 | warehouse 3 | timber 80, stone 120 / 2m00s | timber 354, stone 530 / 7m41s | timber 2,267, stone 3,400 / 41m20s | — | +150 protected (*) / +1,098.08 protected (*) / +3,536.92 protected (*) / — | 3 |
| `walls` | 20 | keep 3 | stone 150, timber 50 / 3m00s | stone 663, timber 221 / 12m12s | stone 4,250, timber 1,417 / 1h10m | stone 174,616, timber 58,205 / 1d15h | wall +2 (flat) / wall +10 (flat) / wall +20 (flat) / wall +40 (flat) | 6 |
| `watchtower` | 15 | keep 2 | timber 90, stone 90 / 2m00s | timber 398, stone 398 / 7m41s | timber 2,550, stone 2,550 / 41m20s | — | scout +8% / scout +40% / scout +80% / — | 3 |
| `barracks` | 20 | keep 2 | timber 100, stone 60, iron 20 / 2m00s | timber 442, stone 265, iron 88 / 7m41s | timber 2,833, stone 1,700, iron 567 / 41m20s | timber 116,410, stone 69,846, iron 23,282 / 19h55m | train_time -2% / train_time -10% / train_time -20% / train_time -40% | 6 |
| `range` | 20 | barracks 2 | timber 140, stone 60, iron 40 / 2m30s | timber 619, stone 265, iron 177 / 9m37s | timber 3,967, stone 1,700, iron 1,133 / 51m40s | timber 162,975, stone 69,846, iron 46,564 / 1d00h | train_time -1% / train_time -5% / train_time -10% / train_time -20% | 6 |
| `stables` | 20 | barracks 3, farmstead 4 | timber 160, stone 100, iron 80, grain 80 / 3m00s | timber 707, stone 442, iron 354, grain 354 / 11m32s | timber 4,533, stone 2,833, iron 2,267, grain 2,267 / 1h01m | timber 186,257, stone 116,410, iron 93,128, grain 93,128 / 1d05h | train_time -1% / train_time -5% / train_time -10% / train_time -20% | 7 |
| `aerie` | 15 | stables 3, beastlore 1 | timber 200, stone 200, iron 120, gold 80 / 5m00s | timber 1,013, stone 1,013, iron 608, gold 405 / 22m07s | timber 7,689, stone 7,689, iron 4,613, gold 3,075 / 2h21m | — | train_time -1% / train_time -5% / train_time -10% / — | 8 |
| `workshop` | 15 | barracks 5, engineering 1 | timber 250, stone 150, iron 150 / 5m00s | timber 1,266, stone 759, iron 759 / 22m07s | timber 9,611, stone 5,767, iron 5,767 / 2h21m | — | train_time -1% / train_time -5% / train_time -10% / — | 8 |
| `sanctum` | 15 | scriptorium 5, arcana 1 | stone 250, gold 150, iron 60 / 6m00s | stone 1,266, gold 759, iron 304 / 26m32s | stone 9,611, gold 5,767, iron 2,307 / 2h50m | — | train_time -1% / train_time -5% / train_time -10% / — | 8 |
| `market` | 15 | keep 3 | timber 150, stone 100, gold 50 / 3m00s | timber 663, stone 442, gold 221 / 11m32s | timber 4,250, stone 2,833, gold 1,417 / 1h01m | — | market_fee -4% / market_fee -20% / market_fee -40% / —<br>carry +3% / carry +15% / carry +30% / — | 4 |
| `warhall` | 7 | keep 3 | timber 200, stone 200, gold 150 / 5m00s | timber 1,311, stone 1,311, gold 983 / 25m19s | — | — | — (gates heroes) | 8 |

Cumulative investment to max level (no modifiers):

| Building | Σ cost L1→max (all resources) | Σ base time L1→max | Reference value Σ (exchange) |
|---|---|---|---|
| `keep` (max 25) | timber 6,060,040, stone 6,060,040, gold 2,020,015 | 21d05h | 19,392,133 |
| `cottages` (max 25) | timber 916,312, stone 305,437 | 5d17h | 1,282,836 |
| `lumberyard` (max 25) | timber 330,455, stone 247,843 | 3d12h | 627,867 |
| `farmstead` (max 25) | timber 371,765, stone 165,230 | 3d12h | 570,041 |
| `quarry` (max 25) | timber 413,071, grain 165,230 | 4d04h | 578,301 |
| `ironworks` (max 25) | timber 787,305, stone 562,361 | 6d13h | 1,462,138 |
| `tradepost` (max 20) | timber 299,902, stone 224,928 | 2d04h | 569,816 |
| `mint` (max 20) | timber 449,853, stone 449,853, iron 224,928 | 3d15h | 1,349,561 |
| `scriptorium` (max 20) | timber 337,391, stone 262,414, gold 112,462 | 2d21h | 989,674 |
| `academy` (max 15) | timber 174,759, stone 174,759, gold 87,380 | 2d00h | 646,610 |
| `warehouse` (max 25) | timber 674,833, stone 449,888 | 4d21h | 1,214,699 |
| `vault` (max 15) | timber 46,639, stone 69,957 | 12h52m | 130,587 |
| `walls` (max 20) | stone 562,316, timber 187,440 | 5d12h | 862,219 |
| `watchtower` (max 15) | timber 52,469, stone 52,469 | 12h52m | 115,432 |
| `barracks` (max 20) | timber 374,877, stone 224,928, iron 74,977 | 2d21h | 764,754 |
| `range` (max 20) | timber 524,829, stone 224,928, iron 149,951 | 3d15h | 1,034,664 |
| `stables` (max 20) | timber 599,805, stone 374,877, iron 299,902, grain 299,902 | 4d08h | 1,829,403 |
| `aerie` (max 15) | timber 174,759, stone 174,759, iron 104,855, gold 69,903 | 2d00h | 761,947 |
| `workshop` (max 15) | timber 218,447, stone 131,069, iron 131,069 | 2d00h | 585,440 |
| `sanctum` (max 15) | stone 218,447, gold 131,069, iron 52,429 | 2d10h | 739,230 |
| `market` (max 15) | timber 87,446, stone 58,297, gold 29,150 | 19h19m | 244,852 |
| `warhall` (max 7) | timber 8,614, stone 8,614, gold 6,461 | 2h40m | 38,334 |

Observations:
- Producers grow output by `L × 1.06^(L−1)` and cost by `1.38^(L−1)`. For the Lumberyard, payback (cost value ÷ extra output per hour) is 2.9 h at L2, 5.6 h at L5, 16.9 h at L10, 53 h at L15, 172 h at L20 and 566 h at L25. Beyond about L15 an upgrade takes days to pay back. This is the intended soft ceiling.
- Producers employ `workers × L` people. A Lumberyard at L20 needs 120 workers, so housing (Cottages, Keep) becomes the real limit. Workers also remove idle people, who are the Trading Post's tax base (see §8).
- The Keep is the gate building: Keep 2 unlocks the economy and military, Keep 3 Walls/Market/Warhall, Keep 4 the Mint, Keep 5 specialisations and settlers (and a 2nd city).

## 4. Technologies — Ashen Crowns

| Tech | Branch | Max | Requires | L1 cost / time | L3 cost / time | Lmax cost / time | Effect per level → at max | Exclusive |
|---|---|---|---|---|---|---|---|---|
| `forestry` | economy | 10 | scriptorium 1 | lore 60, timber 60 / 3m00s | lore 135, timber 135 / 6m19s | lore 2,307, timber 2,307 / 1h25m | production.timber +5% → production.timber +50% | — |
| `agriculture` | economy | 10 | scriptorium 1 | lore 60, grain 60 / 3m00s | lore 135, grain 135 / 6m19s | lore 2,307, grain 2,307 / 1h25m | production.grain +5% → production.grain +50% | — |
| `masonry` | economy | 10 | scriptorium 1 | lore 70, stone 70 / 3m20s | lore 158, stone 158 / 7m01s | lore 2,691, stone 2,691 / 1h34m | production.stone +5% → production.stone +50%; storage.* +3% → storage.* +30% | — |
| `metallurgy` | economy | 10 | scriptorium 2, ironworks 1 | lore 90, iron 60 / 4m00s | lore 203, iron 135 / 8m25s | lore 3,460, iron 2,307 / 1h53m | production.iron +5% → production.iron +50% | — |
| `exchequer` | civic | 10 | scriptorium 3 | lore 120, gold 80 / 5m00s | lore 288, gold 192 / 11m15s | lore 6,197, gold 4,131 / 3h12m | upkeep -4% → upkeep -40%; production.gold +3% → production.gold +30% | — |
| `husbandry` | civic | 10 | scriptorium 2 | lore 80, grain 80 / 3m20s | lore 180, grain 180 / 7m01s | lore 3,075, grain 3,075 / 1h34m | pop_growth +10% → pop_growth +100%; housing +3% → housing +30% | — |
| `tactics` | military | 10 | barracks 2, scriptorium 2 | lore 100, iron 60 / 5m00s | lore 240, iron 144 / 11m15s | lore 5,164, iron 3,098 / 3h12m | attack.* +3% → attack.* +30% | — |
| `armour` | military | 10 | ironworks 2, scriptorium 2 | lore 100, iron 80 / 5m00s | lore 240, iron 192 / 11m15s | lore 5,164, iron 4,131 / 3h12m | defense.* +2% → defense.* +20%; hp.* +2% → hp.* +20% | — |
| `logistics` | military | 10 | scriptorium 2 | lore 90, grain 90 / 4m00s | lore 203, grain 203 / 8m25s | lore 3,460, grain 3,460 / 1h53m | speed +4% → speed +40%; carry +5% → carry +50% | — |
| `espionage` | military | 10 | watchtower 1 | lore 80, gold 40 / 3m20s | lore 180, gold 90 / 7m01s | lore 3,075, gold 1,538 / 1h34m | scout +10% → scout +100% | — |
| `beastlore` | military | 5 | stables 2, logistics 2 | lore 200, grain 200 / 10m00s | lore 512, grain 512 / 22m30s | lore 1,311, grain 1,311 / 50m38s | hp.beast +4% → hp.beast +20% | — |
| `engineering` | military | 5 | masonry 3, tactics 2 | lore 250, timber 250, iron 100 / 10m00s | lore 640, timber 640, iron 256 / 22m30s | lore 1,638, timber 1,638, iron 655 / 50m38s | attack.siege +5% → attack.siege +25% | — |
| `arcana` | arcane | 5 | scriptorium 4 | lore 300, gold 150 / 12m00s | lore 768, gold 384 / 27m00s | lore 1,966, gold 983 / 1h00m | attack.magic +5% → attack.magic +25% | — |
| `plunder` | military | 5 | tactics 3 | lore 150, gold 100 / 6m40s | lore 384, gold 256 / 15m00s | lore 983, gold 655 / 33m45s | loot +4% → loot +20% | — |
| `heroics` | arcane | 5 | warhall 1 | lore 180, gold 120 / 8m20s | lore 461, gold 307 / 18m45s | lore 1,180, gold 786 / 42m12s | hero_xp +10% → hero_xp +50% | — |
| `path_architect` | calling | 3 | keep 5 | lore 400, stone 400, gold 200 / 30m00s | lore 1,296, stone 1,296, gold 648 / 1h07m | lore 1,296, stone 1,296, gold 648 / 1h07m | build_time -15% → build_time -45% | calling |
| `path_steward` | calling | 3 | keep 5 | lore 400, grain 400, gold 200 / 30m00s | lore 1,296, grain 1,296, gold 648 / 1h07m | lore 1,296, grain 1,296, gold 648 / 1h07m | production.* +6% → production.* +18% | calling |
| `path_warlord` | calling | 3 | keep 5 | lore 400, iron 400, gold 200 / 30m00s | lore 1,296, iron 1,296, gold 648 / 1h07m | lore 1,296, iron 1,296, gold 648 / 1h07m | attack.* +6% → attack.* +18%; train_time -8% → train_time -24% | calling |

Notes:
- Lore comes only from the Scriptorium (10/h at L1, 5 workers) plus quest rewards (t05 +100, t06 +150) and Barrow ruins. Forestry L1 (60 lore) therefore needs about 6 h of a level-1 Scriptorium. This is the intended early pacing gate.
- The `calling` specialisations (Architect −15% build time per level, Steward +6% production, Warlord +6% attack and −8% train time) are mutually exclusive [R-TECH-2] and cost the same.

## 5. Units — Ashen Crowns

| Unit | Class | Race | Atk | Def % | HP | EHP¹ | Speed | Carry | Upkeep (gold/h) | Cost | Cost value² | Train | Bonuses | Engine power³ | Atk/100 value | EHP/100 value | Power/100 value | Upkeep h to equal cost⁴ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `militia` | infantry | all | 8 | 10 | 40 | 44.4 | 5 | 15 | 0.3 | grain 30, timber 20, iron 5 | 58 | 45s (barracks) | +40% vs ranged | 72 | 13.8 | 76.6 | 124.1 | 64 |
| `pikeman` | infantry | all | 14 | 18 | 60 | 73.2 | 4.5 | 20 | 0.6 | grain 50, timber 30, iron 30 | 128 | 1m30s (barracks) | +40% vs ranged, +15% vs cavalry | 192 | 10.9 | 57.2 | 150.0 | 71 |
| `bowman` | ranged | all | 14 | 5 | 35 | 36.8 | 5 | 10 | 0.4 | grain 30, timber 45, iron 5 | 83 | 1m00s (range) | +40% vs beast | 66 | 16.9 | 44.4 | 80.1 | 69 |
| `crossbow` | ranged | all | 22 | 8 | 45 | 48.9 | 4.5 | 10 | 0.7 | grain 40, timber 50, iron 35 | 146 | 1m40s (range) | +40% vs beast | 135 | 15.1 | 33.5 | 92.5 | 70 |
| `outrider` | cavalry | all | 18 | 8 | 70 | 76.1 | 10 | 40 | 1 | grain 70, timber 30, iron 40 | 164 | 2m00s (stables) | +40% vs infantry | 182 | 11.0 | 46.4 | 111.0 | 55 |
| `knight` | cavalry | all | 30 | 20 | 110 | 137.5 | 8 | 30 | 1.8 | grain 100, timber 40, iron 90, gold 20 | 344 | 3m20s (stables) | +40% vs infantry | 550 | 8.7 | 40.0 | 159.9 | 64 |
| `griffin` | beast | all | 34 | 12 | 120 | 136.4 | 14 | 25 | 2.5 | grain 160, iron 60, gold 60 | 436 | 5m00s (aerie) | +40% vs cavalry | 552 | 7.8 | 31.3 | 126.6 | 58 |
| `ballista` | siege | all | 45 | 4 | 90 | 93.8 | 2.5 | 0 | 2 | timber 200, iron 100, gold 40 | 480 | 6m00s (workshop) | +20% vs infantry | 441 | 9.4 | 19.5 | 91.9 | 80 |
| `battlemage` | magic | all | 40 | 6 | 60 | 63.8 | 5 | 5 | 3 | grain 80, gold 120, lore 40 | 440 + lore 40 | 6m00s (sanctum) | — | 276 | 9.1 | 14.5 | 62.7 | 49 |
| `scout` | scout | all | 0 | 2 | 15 | 15.3 | 16 | 0 | 0.2 | grain 20, gold 10 | 50 | 40s (barracks) | — | 3 | 0.0 | 30.6 | 6.0 | 83 |
| `wagon` | transport | all | 0 | 4 | 60 | 62.5 | 6 | 300 | 0.4 | timber 80, grain 20 | 100 | 1m20s (barracks) | — | 24 | 0.0 | 62.5 | 24.0 | 83 |
| `settler` | settler | all | 0 | 2 | 50 | 51.0 | 3 | 500 | 1 | timber 800, grain 800, stone 600, gold 400 | 3,520 | 30m00s (keep) | — | 10 | 0.0 | 1.4 | 0.3 | 1,173 |
| `thornweaver` | ranged | thornkin | 20 | 10 | 50 | 55.6 | 6 | 10 | 0.8 | grain 40, timber 70, lore 5 | 110 + lore 5 | 1m50s (range) | +50% vs beast | 150 | 18.2 | 50.5 | 136.4 | 46 |
| `bannerguard` | infantry | ironvow | 14 | 28 | 80 | 111.1 | 4 | 20 | 0.9 | grain 60, stone 30, iron 50 | 176 | 2m00s (barracks) | +40% vs ranged | 336 | 8.0 | 63.1 | 190.9 | 65 |
| `cinderhound` | beast | ashborn | 26 | 8 | 90 | 97.8 | 12 | 20 | 1.4 | grain 120, iron 40 | 184 | 3m00s (stables) | +50% vs cavalry | 306 | 14.1 | 53.2 | 166.3 | 44 |
| `bonelord` | magic | gravewarden | 30 | 15 | 90 | 105.9 | 5 | 10 | 0.4 | grain 40, iron 60, lore 40 | 136 + lore 40 | 4m00s (barracks) | — | 405 | 22.1 | 77.9 | 297.8 | 113 |

Footnotes:
1. EHP (effective HP) = `HP / (1 − min(defenseCapPct, D)/100)`. It ignores walls and modifiers.
2. Cost value = Σ cost × exchange rate. Lore is listed separately.
3. Engine power = `(A + D) × HP / 10` per unit [R-CMB-10].
4. Hours of upkeep (valued at the upkeep resource's exchange rate) that equal the purchase value. Lower means upkeep dominates the lifetime cost sooner.

Neutral-only guard units:

| Guard unit | Class | Atk | Def % | HP | EHP | Engine power | Score (→ hero XP × xpPerScore) |
|---|---|---|---|---|---|---|---|
| `wolf` | beast | 7 | 4 | 35 | 36.5 | 38 | 1 |
| `brigand` | infantry | 10 | 8 | 45 | 48.9 | 81 | 1 |
| `troll` | infantry | 30 | 20 | 220 | 275.0 | 1100 | 5 |
| `wraith` | magic | 26 | 30 | 90 | 128.6 | 504 | 4 |
| `golem` | siege | 40 | 40 | 300 | 500.0 | 2400 | 8 |
| `drake` | beast | 80 | 35 | 600 | 923.1 | 6900 | 20 |

Observations:
- Per value, the strongest generalists are Bannerguard (Ironvow, 191 power / 100 value) and **Bonelord (Gravewarden, 298)**. The Bonelord also costs only 0.4 gold/h upkeep, and Gravewarden has −20% upkeep on top. **Flag:** probably overtuned. Suggested fixes: upkeep 1.2–1.5, or attack 24.
- Battlemage (63 power/100 value, 3 gold/h upkeep, no class bonus) is weak for its cost and only shines with Arcana or a Sage hero. **Flag:** consider a class bonus or upkeep 2.
- Bowmen are value-inefficient as generalists (80) but carry +40% vs beasts. They are the intended answer to Wolf Dens (see §6).
- With Walls 20 (+40 points), Armour 10 and an Ironvow Bannerguard, defence reaches about 76% (`28 × 1.28 + 40`), close to the 80% `defenseCapPct`. The cap binds only at the very end of progression, as intended.

## 6. Neutrals — Ashen Crowns

| Neutral (kind, respawn) | Lvl | Guards | Guard power | Guard score | Rewards | Reward value⁵ | XP | Drops | Min militia to win (losses)⁶ | Min bowman to win (losses)⁶ | Min outrider to win (losses)⁶ | Min knight to win (losses)⁶ | 2× min militia: losses | Reward value ÷ value of militia lost (min / 2× min)⁷ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `wolf_den` (lair, 4h00m) | 1 | wolf 8 | 308 | 8 | grain 150, timber 100 | 250 | 20 | leather_jerkin 15% | 7 (6) | 5 (3) | 4 (2) | 2 (1) | 14 → 2 lost | 0.7× / 2.2× |
|  | 2 | wolf 20 | 770 | 20 | grain 350, timber 250 | 600 | 45 | leather_jerkin 20%, lucky_coin 8% | 18 (12) | 12 (10) | 9 (6) | 5 (4) | 36 → 5 lost | 0.9× / 2.1× |
|  | 3 | wolf 45 | 1,733 | 45 | grain 800, timber 500, gold 80 | 1,540 | 90 | scout_hood 15% | 39 (32) | 27 (24) | 20 (16) | 11 (10) | 78 → 13 lost | 0.8× / 2.0× |
|  | 4 | wolf 90 | 3,465 | 90 | grain 1,600, timber 1,000, gold 160 | 3,080 | 160 | wind_charm 8% | 77 (68) | 54 (49) | 39 (35) | 23 (18) | 154 → 27 lost | 0.8× / 2.0× |
|  | 5 | wolf 180 | 6,930 | 180 | grain 3,000, timber 2,000, gold 300 | 5,900 | 260 | wind_charm 15% | 153 (144) | 107 (105) | 78 (71) | 45 (40) | 306 → 54 lost | 0.7× / 1.9× |
| `brigand_camp` (lair, 6h00m) | 1 | brigand 10 | 810 | 10 | gold 120, iron 80 | 488 | 25 | iron_blade 15% | 13 (9) | 10 (8) | 5 (4) | 3 (2) | 26 → 4 lost | 0.9× / 2.1× |
|  | 2 | brigand 25 | 2,025 | 25 | gold 300, iron 200 | 1,220 | 55 | iron_blade 25%, lucky_coin 10% | 30 (26) | 25 (21) | 13 (10) | 8 (5) | 60 → 10 lost | 0.8× / 2.1× |
|  | 3 | brigand 55, wolf 15 | 5,033 | 70 | gold 700, iron 450, stone 300 | 3,180 | 110 | horned_helm 10% | 78 (71) | 63 (56) | 35 (31) | 20 (17) | 156 → 28 lost | 0.8× / 2.0× |
|  | 4 | brigand 110, troll 4 | 13,310 | 130 | gold 1,400, iron 900, stone 600 | 6,360 | 190 | ember_sword 6% | 148 (136) | 123 (114) | 64 (56) | 37 (31) | 296 → 54 lost | 0.8× / 2.0× |
|  | 5 | brigand 200, troll 10 | 27,200 | 250 | gold 2,600, iron 1,700, stone 1,200 | 11,960 | 300 | ember_sword 12% | 279 (261) | 231 (222) | 120 (111) | 69 (63) | 558 → 102 lost | 0.8× / 2.0× |
| `barrow` (ruin, 12h00m) | 1 | wraith 6 | 3,024 | 24 | lore 120, gold 100, starglass 2 | 300 + lore 120, starglass 2 | 40 | sage_tome 4% | 20 (15) | 16 (14) | 10 (8) | 6 (4) | 40 → 7 lost | 0.3× / 0.7× |
|  | 2 | wraith 16 | 8,064 | 64 | lore 300, gold 250, starglass 4 | 750 + lore 300, starglass 4 | 90 | warden_plate 10% | 51 (42) | 42 (38) | 26 (21) | 15 (11) | 102 → 17 lost | 0.3× / 0.8× |
|  | 3 | wraith 35, troll 5 | 23,140 | 165 | lore 700, gold 600, starglass 8 | 1,800 + lore 700, starglass 8 | 170 | warden_plate 15%, sage_tome 6% | 133 (126) | 111 (99) | 66 (61) | 38 (37) | 266 → 48 lost | 0.2× / 0.6× |
|  | 4 | wraith 70, troll 12 | 48,480 | 340 | lore 1,400, gold 1,200, starglass 14 | 3,600 + lore 1,400, starglass 14 | 280 | skyfall_lance 5% | 273 (253) | 226 (219) | 136 (123) | 78 (73) | 546 → 96 lost | 0.2× / 0.6× |
|  | 5 | wraith 140, golem 10 | 94,560 | 640 | lore 2,800, gold 2,400, starglass 25 | 7,200 + lore 2,800, starglass 25 | 420 | skyfall_lance 10%, starglass_mail 3% | 499 (459) | 413 (407) | 254 (237) | 146 (139) | 998 → 177 lost | 0.3× / 0.7× |
| `crater` (node, 1d00h) | 1 | golem 2 | 4,800 | 16 | starglass 5, stone 300 | 360 + starglass 5 | 50 | — | 18 (14) | 15 (11) | 9 (7) | 5 (4) | 36 → 5 lost | 0.4× / 1.2× |
|  | 2 | golem 5 | 12,000 | 40 | starglass 10, stone 600 | 720 + starglass 10 | 110 | — | 41 (36) | 34 (32) | 21 (17) | 12 (10) | 82 → 13 lost | 0.3× / 1.0× |
|  | 3 | golem 11 | 26,400 | 88 | starglass 18, stone 1,200 | 1,440 + starglass 18 | 200 | starglass_mail 2% | 87 (75) | 71 (66) | 44 (37) | 26 (18) | 174 → 27 lost | 0.3× / 0.9× |
|  | 4 | golem 22, drake 1 | 59,700 | 196 | starglass 30, stone 2,000 | 2,400 + starglass 30 | 340 | starglass_mail 4% | 189 (165) | 153 (149) | 95 (88) | 55 (48) | 378 → 61 lost | 0.3× / 0.7× |
|  | 5 | golem 40, drake 3 | 116,700 | 380 | starglass 50, stone 3,500 | 4,200 + starglass 50 | 520 | starglass_mail 8% | 357 (308) | 288 (276) | 181 (164) | 104 (94) | 714 → 115 lost | 0.2× / 0.6× |

5. Reward value counts only priced resources. `lore` and `starglass` are listed separately.
6. The smallest single-type army **without modifiers** that wins (binary search over the exact combat rules [R-CMB]). Losses at that size are in brackets.
7. Reward value ÷ purchase value of the militia lost, at the minimum army and at twice the minimum. Above 1× the clear pays for its losses, before XP and drops are counted.

Observations:
- At the bare minimum army, every lair is roughly break-even or worse (0.7–0.9×). With 2× the minimum (Lanchester effect), losses fall 2.5–3× and lairs pay 1.9–2.2×. This is the intended "bring enough force" lesson.
- **Barrows and craters** pay only 0.2–1.2× in tradable value. Their purpose is premium `starglass`, `lore` and rare items: Barrow L1 gives 2 starglass + 120 lore for about 7 militia lost. This is acceptable because premium resources have no market value. Watch starglass inflation: about 50 per Crater L5 clear every 24 h.
- Guard power grows faster than rewards in the mixed-guard levels: Brigand L1 to L5 is ×33.6 power but ×24.5 rewards. Higher levels sit farther from the map centre [R-PVE-1], so they are mid-game content.
- Hero XP: Wolf Den L3 gives 22 + 90 = 112 XP per clear, so about 12 clears reach hero L5 and about 71 reach L10 (curve in §7).
- PvE rewards are not limited by carry capacity (GAME_RULES D-8), so wagons are not needed for PvE.

## 7. Heroes — Ashen Crowns

| Level | Total XP to reach | XP for this step | Skill points earned (cumulative) | Example class bonus (marshal) | Hire cost of n-th hero |
|---|---|---|---|---|---|
| 1 | 0 | 0 | 1 | attack.* +5% | gold 300, grain 300 |
| 2 | 200 | 200 | 3 | attack.* +6% | gold 600, grain 600 |
| 3 | 470 | 270 | 5 | attack.* +7% | gold 900, grain 900 |
| 4 | 835 | 365 | 7 | attack.* +8% | gold 1,200, grain 1,200 |
| 5 | 1,327 | 492 | 9 | attack.* +9% | gold 1,500, grain 1,500 |
| 6 | 1,991 | 664 | 11 | attack.* +10% | gold 1,800, grain 1,800 |
| 7 | 2,888 | 897 | 13 | attack.* +11% | gold 2,100, grain 2,100 |
| 8 | 4,099 | 1,211 | 15 | attack.* +12% | — |
| 9 | 5,733 | 1,634 | 17 | attack.* +13% | — |
| 10 | 7,939 | 2,206 | 19 | attack.* +14% | — |

Max level 30: total XP 3,440,086; skill points at max = 59; skills × maxLevel available = 40.

The curve grows 35% per level. L10 needs 7,939 XP and L30 needs 3.44 M XP, so in practice L30 is unreachable in a season. Skill points saturate at about L15–L20 (D-17).

## 8. Starting conditions and early economy check

### 8.1 Ashen Crowns start

- Buildings: keep 1, cottages 1, lumberyard 1, farmstead 1, quarry 1, warehouse 1
- Units: militia 10, scout 2
- Population: 40
- Resources: timber 600, grain 600, stone 500, iron 250, gold 400, starglass 20
- Newbie protection: 72 h

### 8.2 Hour-1 economy of a fresh city (Ironvow)

Race `ironvow` (modifiers: defense.* +8%, attack.cavalry +10%, production.stone +10%, build_time -2%).

Housing 150, workers needed 18, employed 18, idle 22, employment 1.00, army upkeep 3.40 gold/h.

| Resource | Stock t=0 | Gross /h | Consumption /h | Net /h | Storage cap | Hours to cap (or to 0) |
|---|---|---|---|---|---|---|
| `timber` | 600 | 32.00 | 0.00 | +32.00 | 2700 | 65.6 h to cap |
| `grain` | 600 | 36.00 | 4.80 | +31.20 | 2700 | 67.3 h to cap |
| `stone` | 500 | 28.60 | 0.00 | +28.60 | 2700 | 76.9 h to cap |
| `iron` | 250 | 0.00 | 0.00 | +0.00 | 2200 | — |
| `gold` | 400 | 0.00 | 3.40 | -3.40 | 4200 | **117.6 h to 0** |
| `lore` | 0 | 0.00 | 0.00 | +0.00 | ∞ | — |
| `starglass` | 20 | 0.00 | 0.00 | +0.00 | ∞ | — |

Simulation from a tick boundary (exact engine algorithm, no player actions):

| t | Population | Idle | timber | grain | stone | iron | gold | lore |
|---|---|---|---|---|---|---|---|---|
| 1 h | 52 | 34 | 632 | 631 | 529 | 250 | 397 | 0 |
| 2 h | 63 | 45 | 664 | 660 | 557 | 250 | 393 | 0 |
| 6 h | 97 | 79 | 792 | 765 | 672 | 250 | 380 | 0 |
| 12 h | 124 | 106 | 984 | 901 | 843 | 250 | 359 | 0 |
| 24 h | 144 | 126 | 1,368 | 1,136 | 1,186 | 250 | 318 | 0 |

Same city as Gravewarden (−20% upkeep):

Race `gravewarden` (modifiers: upkeep -20%, hp.* +8%, production.lore +15%, build_time -2%).

Housing 150, workers needed 18, employed 18, idle 22, employment 1.00, army upkeep 2.72 gold/h.

| Resource | Stock t=0 | Gross /h | Consumption /h | Net /h | Storage cap | Hours to cap (or to 0) |
|---|---|---|---|---|---|---|
| `timber` | 600 | 32.00 | 0.00 | +32.00 | 2700 | 65.6 h to cap |
| `grain` | 600 | 36.00 | 4.80 | +31.20 | 2700 | 67.3 h to cap |
| `stone` | 500 | 26.00 | 0.00 | +26.00 | 2700 | 84.6 h to cap |
| `iron` | 250 | 0.00 | 0.00 | +0.00 | 2200 | — |
| `gold` | 400 | 0.00 | 2.72 | -2.72 | 4200 | **147.1 h to 0** |
| `lore` | 0 | 0.00 | 0.00 | +0.00 | ∞ | — |
| `starglass` | 20 | 0.00 | 0.00 | +0.00 | ∞ | — |

Simulation from a tick boundary (exact engine algorithm, no player actions):

| t | Population | Idle | timber | grain | stone | iron | gold | lore |
|---|---|---|---|---|---|---|---|---|
| 1 h | 52 | 34 | 632 | 631 | 526 | 250 | 397 | 0 |
| 2 h | 63 | 45 | 664 | 660 | 552 | 250 | 395 | 0 |
| 6 h | 97 | 79 | 792 | 765 | 656 | 250 | 384 | 0 |
| 12 h | 124 | 106 | 984 | 901 | 812 | 250 | 367 | 0 |
| 24 h | 144 | 126 | 1,368 | 1,136 | 1,124 | 250 | 335 | 0 |

### 8.3 Gold balance and the Trading Post

The start army costs 3.4 gold/h. **No building produces gold until a Trading Post (Keep 2) or Mint (Keep 4 + Ironworks 2) exists**, so a fresh Ashen Crowns city runs a gold deficit from minute one. The table below shows the upkeep-resource balance once housing is full, by Trading Post level (each level taxes 0.06 gold/h per idle inhabitant [R-ECO-2]):

| Trading post level | Keep | Housing (= pop when full) | Idle at full housing | gold gross /h | Upkeep /h (start army) | Net gold /h | Extra militia supportable |
|---|---|---|---|---|---|---|---|
| 0 | 1 | 150 | 132 | 0.00 | 3.40 | -3.40 | 0 |
| 1 | 2 | 214 | 196 | 11.76 | 3.40 | +8.36 | 28 |
| 2 | 2 | 214 | 196 | 23.52 | 3.40 | +20.12 | 67 |
| 3 | 2 | 214 | 196 | 35.28 | 3.40 | +31.88 | 106 |
| 5 | 2 | 214 | 196 | 58.80 | 3.40 | +55.40 | 185 |

**Assessment (deliberate pressure, flagged risk):**
- **Deliberate.** The deficit pushes the player through the tutorial path (Keep 2, then Trading Post) and makes gold the army-size limiter, as in the original salary model (ECO-015). Starting gold (400) covers the start army for about 118 h (147 h for Gravewarden). Tutorial quests add 600 gold (t04 +100, t08 +150, t09 +100, t10 +250).
- **Risk 1: early over-training.** Each militia adds 0.3 gold/h at home and 0.6 gold/h on a mission. A new player who trains 50 militia before building a Trading Post and sends them raiding pays 30 + 3.4 gold/h, so the 400 starting gold is gone in about 12 h. Desertion then starts silently (first deserter after about 20 h with 10 militia; scales with army size). Mitigation proposals:
  - a UI warning when net gold < 0;
  - a tutorial step for the Trading Post right after Keep 2;
  - or `start.buildings.tradepost = 1`.
- **Risk 2: the Trading Post / producer tension.** The idle tax depends on **idle** people. Every producer level hires 5–10 workers and removes 0.3–0.6 gold/h per Trading Post level. Players who max their producers can unintentionally bankrupt their army. This is intended depth, but it must be explained in the UI (show "gold lost by employing").
- **Risk 3: the soft cap hides gold shortfalls.** Overflow decay means storage never stops gold, so a negative gold balance is the only economic failure state. It should be the most visible alarm in the city UI.

## 9. Void Frontier comparison (pack-driven balance)

The engine is the same. Only the pack differs. Key differences:

| Aspect | Ashen Crowns | Void Frontier | Effect |
|---|---|---|---|
| Army upkeep | gold, ×2 on missions | biomass (= food), ×1 | Food is the single limiter, with no early deficit |
| Population growth | `free_housing` 12%/h | `idle` 8%/h (POP-002 mode) | Void Frontier grows slower once people are employed |
| Storage | soft cap, 1%/h decay | hard cap, excess lost | Void Frontier punishes overflow and needs Depots |
| Combat | 20 rounds, cap 80% | 8 rounds, cap 90% (default) | Void Frontier: more timeouts, so defender-favoured |
| Travel floor | 900 s | 300 s | Void Frontier raids are faster |
| PvP limits | 2 attacks / target / 24 h; race/faction penalties | unlimited; none | Void Frontier is free-for-all |
| Loot | 50% | 40% | Void Frontier is less punishing |
| Heroes | 4 classes, 2 points/level, L30 | 2 classes, 1 point/level, L20 | Void Frontier is simpler |

Buildings and units by role:

| Role | Ashen Crowns: building, L1 cost / time, L10 cost / time, L10 effect | Void Frontier: same |
|---|---|---|
| core | `keep`: timber 120, stone 120, gold 40 / 2m00s; L10 timber 4,613, stone 4,613, gold 1,538 / 46m57s; +853.99 housing; build_time -20% | `command`: alloy 150, energy 100 / 2m00s; L10 alloy 4,250, energy 2,833 / 41m20s; +500 housing |
| housing | `cottages`: timber 60, stone 20 / 1m00s; L10 timber 1,408, stone 469 / 18m10s; +1,520.53 housing | `habitat`: alloy 80 / 1m00s; L10 alloy 2,267 / 20m40s; +1,551.33 housing |
| producer A | `lumberyard`: timber 40, stone 30 / 50s; L10 timber 726, stone 545 / 13m16s; +540.63 timber/h (60 workers) | `refinery`: energy 50 / 50s; L10 energy 1,417 / 17m14s; +542.96 alloy/h (40 workers) |
| food producer | `farmstead`: timber 45, stone 20 / 50s; L10 timber 817, stone 363 / 13m16s; +608.21 grain/h (60 workers) | `biodome`: alloy 50 / 50s; L10 alloy 1,417 / 17m14s; +620.53 biomass/h (40 workers) |
| producer C | `quarry`: timber 50, grain 20 / 1m00s; L10 timber 908, grain 363 / 15m55s; +439.26 stone/h (60 workers) | `reactor`: alloy 60 / 1m00s; L10 alloy 1,700 / 20m40s; +465.40 energy/h (40 workers) |
| science | `scriptorium`: timber 90, stone 70, gold 30 / 2m00s; L10 timber 2,550, stone 1,983, gold 850 / 41m20s; +199.90 lore/h (50 workers) | `lab`: alloy 100, energy 100 / 2m00s; L10 alloy 2,833, energy 2,833 / 41m20s; +186.16 data/h (40 workers) |
| storage | `warehouse`: timber 60, stone 40 / 1m00s; L10 timber 1,240, stone 826 / 17m01s; +33,276.95 storage (*) | `depot`: alloy 60 / 1m00s; L10 alloy 1,700 / 20m40s; +35,369.22 storage (*) |
| protection | `vault`: timber 80, stone 120 / 2m00s; L10 timber 2,267, stone 3,400 / 41m20s; +3,536.92 protected (*) | `bunker`: alloy 120 / 2m00s; L10 alloy 3,400 / 41m20s; +2,000 protected (*) |
| wall | `walls`: stone 150, timber 50 / 3m00s; L10 stone 4,250, timber 1,417 / 1h10m; wall +20 (flat) | `shields`: energy 200, alloy 100 / 3m00s; L10 energy 5,667, alloy 2,833 / 1h01m; wall +30 (flat) |
| hero building | `warhall`: timber 200, stone 200, gold 150 / 5m00s; L7 timber 3,355, stone 3,355, gold 2,517 / 56m58s; — | `crew`: alloy 300, energy 300 / 5m00s; L5 alloy 1,326, energy 1,326 / 19m13s; — |

| Unit role | Ashen Crowns | Void Frontier |
|---|---|---|
| basic infantry | `militia` A8/D10/HP40, spd 5, upkeep 0.3, cost value 58, power/100 value 124.1 | `trooper` A10/D12/HP45, spd 6, upkeep 0.5, cost value 60, power/100 value 165.0 |
| ranged | `bowman` A14/D5/HP35, spd 5, upkeep 0.4, cost value 83, power/100 value 80.1 | `sniper` A16/D4/HP30, spd 7, upkeep 0.4, cost value 64, power/100 value 93.8 |
| cavalry | `outrider` A18/D8/HP70, spd 10, upkeep 1, cost value 164, power/100 value 111.0 | `hover` A25/D15/HP100, spd 11, upkeep 1.2, cost value 162, power/100 value 246.9 |
| scout | `scout` A0/D2/HP15, spd 16, upkeep 0.2, cost value 50, power/100 value 6.0 | `probe` A0/D1/HP10, spd 20, upkeep 0.1, cost value 30, power/100 value 3.3 |
| transport | `wagon` A0/D4/HP60, spd 6, upkeep 0.4, cost value 100, power/100 value 24.0 | `hauler` A0/D3/HP60, spd 8, upkeep 0.3, cost value 60, power/100 value 30.0 |
| settler | `settler` A0/D2/HP50, spd 3, upkeep 1, cost value 3520, power/100 value 0.3 | `colony` A0/D2/HP80, spd 4, upkeep 1, cost value 2520, power/100 value 0.6 |

### 9.1 Void Frontier tables

Buildings:

| Building | Max | Req. | L1 cost / time | L5 cost / time | L10 cost / time | L20 cost / time | Effect L1/L5/L10/L20 | Score/lvl |
|---|---|---|---|---|---|---|---|---|
| `command` | 20 | — | alloy 150, energy 100 / 2m00s | alloy 663, energy 442 / 7m41s | alloy 4,250, energy 2,833 / 41m20s | alloy 174,616, energy 116,410 / 19h55m | +50 housing / +250 housing / +500 housing / +1,000 housing | 15 |
| `habitat` | 20 | — | alloy 80 / 1m00s | alloy 354 / 3m51s | alloy 2,267 / 20m40s | alloy 93,128 / 9h57m | +100 housing / +607.75 housing / +1,551.33 housing / +5,053.90 housing | 4 |
| `refinery` | 20 | — | energy 50 / 50s | energy 221 / 3m13s | energy 1,417 / 17m14s | energy 58,205 / 8h18m | +35 alloy/h (4 workers) / +212.71 alloy/h (20 workers) / +542.96 alloy/h (40 workers) / +1,768.87 alloy/h (80 workers) | 4 |
| `biodome` | 20 | — | alloy 50 / 50s | alloy 221 / 3m13s | alloy 1,417 / 17m14s | alloy 58,205 / 8h18m | +40 biomass/h (4 workers) / +243.10 biomass/h (20 workers) / +620.53 biomass/h (40 workers) / +2,021.56 biomass/h (80 workers) | 4 |
| `reactor` | 20 | — | alloy 60 / 1m00s | alloy 265 / 3m51s | alloy 1,700 / 20m40s | alloy 69,846 / 9h57m | +30 energy/h (4 workers) / +182.33 energy/h (20 workers) / +465.40 energy/h (40 workers) / +1,516.17 energy/h (80 workers) | 4 |
| `lab` | 15 | command 2 | alloy 100, energy 100 / 2m00s | alloy 442, energy 442 / 7m41s | alloy 2,833, energy 2,833 / 41m20s | — | +12 data/h (4 workers) / +72.93 data/h (20 workers) / +186.16 data/h (40 workers) / — | 4 |
| `depot` | 20 | — | alloy 60 / 1m00s | alloy 265 / 3m51s | alloy 1,700 / 20m40s | alloy 69,846 / 9h57m | +1,500 storage (*) / +10,980.75 storage (*) / +35,369.22 storage (*) / +183,477.27 storage (*) | 4 |
| `bunker` | 10 | depot 2 | alloy 120 / 2m00s | alloy 530 / 7m41s | alloy 3,400 / 41m20s | — | +200 protected (*) / +1,000 protected (*) / +2,000 protected (*) / — | 4 |
| `shields` | 15 | command 3 | energy 200, alloy 100 / 3m00s | energy 884, alloy 442 / 11m32s | energy 5,667, alloy 2,833 / 1h01m | — | wall +3 (flat) / wall +15 (flat) / wall +30 (flat) / — | 4 |
| `dock` | 20 | command 2 | alloy 120, energy 80 / 2m00s | alloy 530, energy 354 / 7m41s | alloy 3,400, energy 2,267 / 41m20s | alloy 139,693, energy 93,128 / 19h55m | train_time -2% / train_time -10% / train_time -20% / train_time -40% | 4 |
| `crew` | 5 | command 3 | alloy 300, energy 300 / 5m00s | alloy 1,326, energy 1,326 / 19m13s | — | — | — (gates heroes) | 4 |

Technologies:

| Tech | Branch | Max | Requires | L1 cost / time | L3 cost / time | Lmax cost / time | Effect per level → at max | Exclusive |
|---|---|---|---|---|---|---|---|---|
| `nanofab` | industry | 10 | lab 1 | data 80 / 3m00s | data 180 / 6m19s | data 3,075 / 1h25m | production.alloy +6% → production.alloy +60% | — |
| `hydroponics` | industry | 10 | lab 1 | data 80 / 3m00s | data 180 / 6m19s | data 3,075 / 1h25m | production.biomass +6% → production.biomass +60% | — |
| `weapons` | military | 10 | lab 2 | data 120, energy 100 / 5m00s | data 288, energy 240 / 11m15s | data 6,197, energy 5,164 / 3h12m | attack.* +4% → attack.* +40% | — |
| `warp` | military | 10 | lab 2 | data 100 / 4m00s | data 225 / 8m25s | data 3,844 / 1h53m | speed +6% → speed +60% | — |
| `doctrine_raider` | doctrine | 3 | command 4 | data 400 / 20m00s | data 1,296 / 45m00s | data 1,296 / 45m00s | loot +10% → loot +30% | doctrine |
| `doctrine_builder` | doctrine | 3 | command 4 | data 400 / 20m00s | data 1,296 / 45m00s | data 1,296 / 45m00s | build_time -15% → build_time -45% | doctrine |

Units:

| Unit | Class | Race | Atk | Def % | HP | EHP¹ | Speed | Carry | Upkeep (biomass/h) | Cost | Cost value² | Train | Bonuses | Engine power³ | Atk/100 value | EHP/100 value | Power/100 value | Upkeep h to equal cost⁴ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `trooper` | infantry | all | 10 | 12 | 45 | 51.1 | 6 | 20 | 0.5 | alloy 30, biomass 30 | 60 | 45s (dock) | +30% vs ranged | 99 | 16.7 | 85.2 | 165.0 | 120 |
| `sniper` | ranged | all | 16 | 4 | 30 | 31.2 | 7 | 5 | 0.4 | alloy 40, energy 20 | 64 | 1m00s (dock) | +30% vs beast | 60 | 25.0 | 48.8 | 93.8 | 160 |
| `hover` | cavalry | all | 25 | 15 | 100 | 117.6 | 11 | 40 | 1.2 | alloy 90, energy 60 | 162 | 2m30s (dock) | +30% vs infantry | 400 | 15.4 | 72.6 | 246.9 | 135 |
| `probe` | scout | all | 0 | 1 | 10 | 10.1 | 20 | 0 | 0.1 | energy 25 | 30 | 30s (dock) | — | 1 | 0.0 | 33.7 | 3.3 | 300 |
| `hauler` | transport | all | 0 | 3 | 60 | 61.9 | 8 | 400 | 0.3 | alloy 60 | 60 | 1m00s (dock) | — | 18 | 0.0 | 103.1 | 30.0 | 200 |
| `colony` | settler | all | 0 | 2 | 80 | 81.6 | 4 | 600 | 1 | alloy 900, biomass 900, energy 600 | 2,520 | 30m00s (command) | — | 16 | 0.0 | 3.2 | 0.6 | 2,520 |
| `ravager` | beast | swarm | 22 | 6 | 70 | 74.5 | 12 | 25 | 0.7 | biomass 90 | 90 | 1m30s (dock) | +40% vs cavalry | 196 | 24.4 | 82.7 | 217.8 | 129 |
| `aegis` | infantry | concord | 14 | 30 | 90 | 128.6 | 5 | 10 | 0.9 | alloy 80, energy 40 | 128 | 2m00s (dock) | — | 396 | 10.9 | 100.4 | 309.4 | 142 |

Neutrals:

| Neutral (kind, respawn) | Lvl | Guards | Guard power | Guard score | Rewards | Reward value⁵ | XP | Drops | Min trooper to win (losses)⁶ | Min sniper to win (losses)⁶ | Min hover to win (losses)⁶ | 2× min trooper: losses | Reward value ÷ value of trooper lost (min / 2× min)⁷ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `nest` (lair, 3h00m) | 1 | crawler 8 | 448 | 8 | biomass 200, alloy 100 | 300 | 20 | — | 8 (4) | 7 (4) | 3 (2) | 16 → 2 lost | 1.2× / 2.5× |
|  | 2 | crawler 32 | 1,792 | 32 | biomass 400, alloy 200 | 600 | 40 | — | 30 (20) | 25 (21) | 12 (8) | 60 → 9 lost | 0.5× / 1.1× |
|  | 3 | crawler 72 | 4,032 | 72 | biomass 600, alloy 300 | 900 | 60 | — | 67 (46) | 55 (52) | 28 (18) | 134 → 22 lost | 0.3× / 0.7× |
|  | 4 | crawler 128 | 7,168 | 128 | biomass 800, alloy 400 | 1,200 | 80 | — | 119 (81) | 98 (92) | 49 (34) | 238 → 38 lost | 0.2× / 0.5× |
| `derelict` (ruin, 12h00m) | 1 | sentinel 2 | 3,500 | 12 | data 150, quanta 3 | 0 + data 150, quanta 3 | 50 | exosuit 5% | 14 (7) | 13 (10) | 6 (2) | 28 → 3 lost | 0.0× / 0.0× |
|  | 2 | sentinel 8 | 14,000 | 48 | data 300, quanta 6 | 0 + data 300, quanta 6 | 100 | exosuit 10% | 53 (24) | 46 (34) | 22 (10) | 106 → 12 lost | 0.0× / 0.0× |
|  | 3 | sentinel 18 | 31,500 | 108 | data 450, quanta 9 | 0 + data 450, quanta 9 | 150 | exosuit 15% | 117 (52) | 100 (72) | 48 (22) | 234 → 26 lost | 0.0× / 0.0× |
|  | 4 | sentinel 32 | 56,000 | 192 | data 600, quanta 12 | 0 + data 600, quanta 12 | 200 | exosuit 20% | 208 (91) | 177 (128) | 85 (39) | 416 → 45 lost | 0.0× / 0.0× |

**Flag (Void Frontier).** Nest rewards scale ×4 from L1 to L4 while guards scale ×16. Nests at L3 and above lose value even with twice the minimum army (0.5–0.7×). Suggest rewards of about `200 × L^1.6` biomass. Derelict rewards are only `data` and `quanta` (unpriced), which is fine for a premium source.

Hero XP curve:

| Level | Total XP to reach | XP for this step | Skill points earned (cumulative) | Example class bonus (captain) | Hire cost of n-th hero |
|---|---|---|---|---|---|
| 1 | 0 | 0 | 1 | attack.* +6% | alloy 400 |
| 2 | 150 | 150 | 2 | attack.* +7% | alloy 800 |
| 3 | 360 | 210 | 3 | attack.* +8% | alloy 1,200 |
| 4 | 654 | 294 | 4 | attack.* +9% | alloy 1,600 |
| 5 | 1,066 | 412 | 5 | attack.* +10% | alloy 2,000 |
| 6 | 1,642 | 576 | 6 | attack.* +11% | alloy 2,400 |
| 7 | 2,449 | 807 | 7 | attack.* +12% | alloy 2,800 |
| 8 | 3,578 | 1,129 | 8 | attack.* +13% | — |
| 9 | 5,159 | 1,581 | 9 | attack.* +14% | — |
| 10 | 7,373 | 2,214 | 10 | attack.* +15% | — |

Max level 20: total XP 223,736; skill points at max = 20; skills × maxLevel available = 10.

### 9.2 Void Frontier start and early economy (Concord)

- Buildings: command 1, habitat 1, refinery 1, biodome 1, reactor 1, depot 1
- Units: trooper 8, probe 2
- Population: 60
- Resources: alloy 800, biomass 800, energy 600, quanta 10
- Newbie protection: 72 h

Race `concord` (modifiers: housing +10%, production.energy +10%).

Housing 165, workers needed 12, employed 12, idle 48, employment 1.00, army upkeep 4.20 biomass/h.

| Resource | Stock t=0 | Gross /h | Consumption /h | Net /h | Storage cap | Hours to cap (or to 0) |
|---|---|---|---|---|---|---|
| `alloy` | 800 | 35.00 | 0.00 | +35.00 | 3500 | 77.1 h to cap |
| `biomass` | 800 | 40.00 | 10.20 | +29.80 | 3500 | 90.6 h to cap |
| `energy` | 600 | 33.00 | 0.00 | +33.00 | 3500 | 87.9 h to cap |
| `data` | 0 | 0.00 | 0.00 | +0.00 | ∞ | — |
| `quanta` | 10 | 0.00 | 0.00 | +0.00 | ∞ | — |

Simulation from a tick boundary (exact engine algorithm, no player actions):

| t | Population | Idle | alloy | biomass | energy | data |
|---|---|---|---|---|---|---|
| 1 h | 63 | 51 | 835 | 830 | 633 | 0 |
| 2 h | 68 | 56 | 870 | 859 | 666 | 0 |
| 6 h | 89 | 77 | 1,010 | 971 | 798 | 0 |
| 12 h | 135 | 123 | 1,220 | 1,120 | 996 | 0 |
| 24 h | 165 | 153 | 1,640 | 1,356 | 1,392 | 0 |

Unlike Ashen Crowns, every resource of a fresh Void Frontier city is positive. The upkeep resource (biomass) nets +29.8/h, so there is no early bankruptcy risk. The pressure instead comes from the hard storage cap: alloy hits the 3,500 cap after about 77 h, after which production is wasted.

## 10. Summary of balance flags

| # | Pack | Finding | Suggested action |
|---|---|---|---|
| B-1 | Ashen Crowns | Gold deficit until a Trading Post exists (§8.3) | Keep it as pressure; add a UI alarm and a tutorial step, or start with Trading Post 1 |
| B-2 | Ashen Crowns | Bonelord 298 power/100 value at 0.4 upkeep | Raise upkeep to about 1.2 or lower attack |
| B-3 | Ashen Crowns | Battlemage weak (63) with the highest upkeep | Add a class bonus or cut upkeep |
| B-4 | Ashen Crowns | `maxCities 9` unreachable (6 max) | Set `citiesPerCapLevels 3`, or `maxCities 6` |
| B-5 | Ashen Crowns | Skill points exceed learnable skills after about L15–L20 | Add skills or set `skillPointsPerLevel 1` |
| B-6 | Ashen Crowns | Starglass faucet from Craters/Barrows (up to 50 per clear) | Monitor per-day issuance per player |
| B-7 | Void Frontier | Nest rewards do not scale with guard count | Rescale rewards (§9.1) |
| B-8 | both | PvE rewards ignore carry; market escrow hides stock from loot | See GAME_RULES D-8, D-10 |
