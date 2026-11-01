## Pack `ashen-crowns` 1.2.0

Pack hash `a8f307219632…`. World speed ×1. Income value is Σ net/h × exchange rate.

### Race `thornkin`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `thornweaver`; first army at h11.5; trained 24, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 129.3 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 87% | 359.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 2 | 63/150 | 95% | 450.7 | 0.0 | 12 | 14 | 83 | wolf_den L1 |
| 4 | 82/150 | 98% | 573.7 | 0.0 | 12 | 17 | 95 | wolf_den L1 |
| 8 | 133/214 | 100% | 659.9 | 0.0 | 12 | 25 | 155 | wolf_den L1 |
| 12 | 164/214 | 100% | 827.7 | 102.9 | 13 | 32 | 190 | wolf_den L1 |
| 16 | 183/214 | 100% | 960.4 | 77.2 | 14 | 37 | 215 | wolf_den L1 |
| 24 | 300/428 | 100% | 1128.9 | 94.4 | 36 | 43 | 305 | crater L2 |

Queue idle 93.7%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets 1.3/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 94% of the upkeep resource's production at the end

**Careless bot (ignores wages)** — target unit `thornweaver`; first army at h11.5; trained 39, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 129.3 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 87% | 359.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 2 | 63/150 | 95% | 450.7 | 0.0 | 12 | 14 | 83 | wolf_den L1 |
| 4 | 82/150 | 98% | 573.7 | 0.0 | 12 | 17 | 95 | wolf_den L1 |
| 8 | 133/214 | 100% | 659.9 | 0.0 | 12 | 25 | 155 | wolf_den L1 |
| 12 | 164/214 | 100% | 765.7 | 104.2 | 14 | 31 | 189 | wolf_den L1 |
| 16 | 183/214 | 100% | 736.6 | 228.8 | 28 | 31 | 231 | crater L1 |
| 24 | 248/315 | 100% | 697.0 | 232.5 | 51 | 33 | 307 | crater L2 |

Queue idle 96.2%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 96% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets -19.7/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 232% of the upkeep resource's production at the end

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| wolf_den L1 | 308 | thornweaver | 4 | 25.0 | 460 | 250 | 37 | 98 | 4.7 | 6.0 |
| wolf_den L2 | 770 | thornweaver | 12 | 25.0 | 1380 | 600 | 110 | 145 | 9.5 | 6.0 |
| brigand_camp L1 | 810 | militia | 19 | 26.3 | 1102 | 488 | 78 | 120 | 9.2 | 4.0 |
| wolf_den L3 | 1733 | thornweaver | 27 | 29.6 | 3105 | 1540 | 248 | 373 | 8.3 | 6.0 |
| brigand_camp L2 | 2025 | militia | 44 | 29.5 | 2552 | 1220 | 181 | 286 | 8.9 | 4.0 |
| barrow L1 † | 3024 | militia | 30 | 26.7 | 1740 | 422 | 124 | -166 | never | 2.0 |
| wolf_den L4 | 3465 | thornweaver | 55 | 29.1 | 6325 | 3080 | 504 | 735 | 8.6 | 6.0 |
| crater L1 † | 4800 | militia | 27 | 29.6 | 1566 | 365 | 111 | -210 | never | 1.0 |
| brigand_camp L3 | 5033 | militia | 119 | 29.4 | 6902 | 3180 | 491 | 660 | 10.5 | 4.0 |
| wolf_den L5 | 6930 | thornweaver | 110 | 30.0 | 12650 | 5900 | 1008 | 1097 | 11.5 | 6.0 |
| barrow L2 † | 8064 | militia | 77 | 28.6 | 4466 | 1054 | 317 | -541 | never | 2.0 |
| crater L2 † | 12000 | militia | 60 | 30.0 | 3480 | 730 | 247 | -561 | never | 1.0 |
| brigand_camp L4 | 13310 | militia | 227 | 30.0 | 13166 | 6360 | 936 | 1475 | 8.9 | 4.0 |
| barrow L3 † | 23140 | militia | 203 | 29.6 | 11774 | 2508 | 837 | -1814 | never | 2.0 |
| crater L3 † | 26400 | militia | 125 | 29.6 | 7250 | 1458 | 515 | -1203 | never | 1.0 |
| brigand_camp L5 | 27200 | militia | 425 | 29.9 | 24650 | 11960 | 1752 | 2838 | 8.7 | 4.0 |
| barrow L4 † | 48480 | militia | 415 | 29.9 | 24070 | 5014 | 1711 | -3894 | never | 2.0 |
| crater L4 † | 59700 | militia | 274 | 29.9 | 15892 | 2430 | 1129 | -3451 | never | 1.0 |
| barrow L5 † | 94560 | militia | 754 | 30.0 | 43732 | 10025 | 3108 | -6203 | never | 2.0 |
| crater L5 † | 116700 | militia | 520 | 30.0 | 30160 | 4250 | 2143 | -6941 | never | 1.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

### Race `ironvow`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `bannerguard`; first army at h10.8; trained 20, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 127.6 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 87% | 339.6 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 2 | 63/150 | 100% | 393.0 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 478.6 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 566.1 | 166.7 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 615.1 | 92.4 | 15 | 26 | 167 | wolf_den L1 |
| 16 | 180/214 | 100% | 812.6 | 116.7 | 16 | 30 | 187 | brigand_camp L1 |
| 24 | 241/315 | 100% | 1041.9 | 111.5 | 32 | 39 | 275 | crater L1 |

Queue idle 94.4%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets -2.2/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 112% of the upkeep resource's production at the end

**Careless bot (ignores wages)** — target unit `bannerguard`; first army at h8.8; trained 25, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 127.6 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 87% | 339.6 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 2 | 63/150 | 100% | 393.0 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 478.6 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 566.1 | 166.7 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 634.3 | 141.0 | 18 | 26 | 175 | brigand_camp L1 |
| 16 | 180/214 | 100% | 744.1 | 131.9 | 23 | 30 | 209 | crater L1 |
| 24 | 241/315 | 100% | 1028.4 | 134.9 | 37 | 39 | 290 | crater L2 |

Queue idle 94.4%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets -6.7/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 135% of the upkeep resource's production at the end

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| wolf_den L1 | 308 | militia | 10 | 30.0 | 580 | 250 | 43 | 33 | 17.7 | 6.0 |
| wolf_den L2 | 770 | militia | 26 | 26.9 | 1508 | 600 | 112 | 82 | 18.4 | 6.0 |
| brigand_camp L1 | 810 | militia | 19 | 26.3 | 1102 | 488 | 82 | 116 | 9.5 | 4.0 |
| wolf_den L3 | 1733 | militia | 57 | 29.8 | 3306 | 1540 | 246 | 309 | 10.7 | 6.0 |
| brigand_camp L2 | 2025 | militia | 44 | 29.5 | 2552 | 1220 | 190 | 277 | 9.2 | 4.0 |
| barrow L1 † | 3024 | militia | 30 | 26.7 | 1740 | 422 | 130 | -172 | never | 2.0 |
| wolf_den L4 | 3465 | militia | 116 | 29.3 | 6728 | 3080 | 501 | 608 | 11.1 | 6.0 |
| crater L1 † | 4800 | militia | 27 | 29.6 | 1566 | 365 | 117 | -215 | never | 1.0 |
| brigand_camp L3 | 5033 | militia | 117 | 29.9 | 6786 | 3180 | 505 | 646 | 10.5 | 4.0 |
| wolf_den L5 | 6930 | militia | 231 | 29.9 | 13398 | 5900 | 998 | 896 | 15.0 | 6.0 |
| barrow L2 † | 8064 | militia | 77 | 28.6 | 4466 | 1054 | 333 | -556 | never | 2.0 |
| crater L2 † | 12000 | militia | 60 | 30.0 | 3480 | 730 | 259 | -573 | never | 1.0 |
| brigand_camp L4 | 13310 | militia | 225 | 29.8 | 13050 | 6360 | 972 | 1499 | 8.7 | 4.0 |
| barrow L3 † | 23140 | militia | 201 | 29.9 | 11658 | 2508 | 868 | -1846 | never | 2.0 |
| crater L3 † | 26400 | militia | 125 | 29.6 | 7250 | 1458 | 540 | -1228 | never | 1.0 |
| brigand_camp L5 | 27200 | militia | 424 | 29.7 | 24592 | 11960 | 1832 | 2824 | 8.7 | 4.0 |
| barrow L4 † | 48480 | militia | 414 | 29.7 | 24012 | 5014 | 1788 | -3906 | never | 2.0 |
| crater L4 † | 59700 | militia | 274 | 29.9 | 15892 | 2430 | 1184 | -3505 | never | 1.0 |
| barrow L5 † | 94560 | militia | 750 | 30.0 | 43500 | 10025 | 3240 | -6265 | never | 2.0 |
| crater L5 † | 116700 | militia | 518 | 29.9 | 30044 | 4250 | 2238 | -6971 | never | 1.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

### Race `ashborn`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `cinderhound`; first army at h17.9; trained 14, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 124.5 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 96% | 331.4 | 0.0 | 12 | 12 | 75 | wolf_den L1 |
| 2 | 63/150 | 100% | 382.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 463.7 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 556.1 | 166.7 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 618.5 | 51.5 | 12 | 26 | 158 | wolf_den L1 |
| 16 | 180/214 | 100% | 714.0 | 45.7 | 12 | 31 | 183 | wolf_den L1 |
| 24 | 240/315 | 100% | 903.2 | 140.4 | 26 | 39 | 278 | crater L1 |

Queue idle 94.6%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 95% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets -6.6/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 140% of the upkeep resource's production at the end

**Careless bot (ignores wages)** — target unit `cinderhound`; first army at h17.9; trained 26, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 124.5 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 96% | 331.4 | 0.0 | 12 | 12 | 75 | wolf_den L1 |
| 2 | 63/150 | 100% | 382.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 463.7 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 556.1 | 166.7 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 618.5 | 51.5 | 12 | 26 | 158 | wolf_den L1 |
| 16 | 180/214 | 100% | 714.0 | 45.7 | 12 | 31 | 183 | wolf_den L1 |
| 24 | 240/315 | 100% | 904.5 | 260.1 | 38 | 40 | 330 | crater L2 |

Queue idle 94.3%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace
- ⚠ bottleneck: `gold` nets -24.5/h at the end, yet buildings or units cost it
- ⚠ upkeep consumes 260% of the upkeep resource's production at the end

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| wolf_den L1 | 308 | militia | 10 | 30.0 | 580 | 250 | 43 | 33 | 17.7 | 6.0 |
| wolf_den L2 | 770 | militia | 25 | 28.0 | 1450 | 600 | 108 | 86 | 16.9 | 6.0 |
| brigand_camp L1 | 810 | militia | 17 | 29.4 | 986 | 488 | 73 | 125 | 7.9 | 4.0 |
| wolf_den L3 | 1733 | militia | 57 | 28.1 | 3306 | 1540 | 246 | 365 | 9.1 | 6.0 |
| brigand_camp L2 | 2025 | militia | 44 | 29.5 | 2552 | 1220 | 190 | 277 | 9.2 | 4.0 |
| barrow L1 † | 3024 | militia | 28 | 28.6 | 1624 | 422 | 121 | -163 | never | 2.0 |
| wolf_den L4 | 3465 | militia | 112 | 29.5 | 6496 | 3080 | 484 | 680 | 9.6 | 6.0 |
| crater L1 † | 4800 | militia | 27 | 25.9 | 1566 | 365 | 117 | -157 | never | 1.0 |
| brigand_camp L3 | 5033 | militia | 114 | 29.8 | 6612 | 3180 | 492 | 717 | 9.2 | 4.0 |
| wolf_den L5 | 6930 | militia | 224 | 29.9 | 12992 | 5900 | 968 | 1048 | 12.4 | 6.0 |
| barrow L2 † | 8064 | militia | 74 | 29.7 | 4292 | 1054 | 320 | -540 | never | 2.0 |
| crater L2 † | 12000 | militia | 58 | 29.3 | 3364 | 730 | 251 | -506 | never | 1.0 |
| brigand_camp L4 | 13310 | militia | 218 | 29.8 | 12644 | 6360 | 942 | 1650 | 7.7 | 4.0 |
| barrow L3 † | 23140 | militia | 194 | 29.9 | 11252 | 2508 | 838 | -1694 | never | 2.0 |
| crater L3 † | 26400 | militia | 121 | 29.8 | 7018 | 1458 | 523 | -1156 | never | 1.0 |
| brigand_camp L5 | 27200 | militia | 411 | 29.9 | 23838 | 11960 | 1776 | 3057 | 7.8 | 4.0 |
| barrow L4 † | 48480 | militia | 400 | 30.0 | 23200 | 5014 | 1728 | -3674 | never | 2.0 |
| crater L4 † | 59700 | militia | 265 | 29.8 | 15370 | 2430 | 1145 | -3295 | never | 1.0 |
| barrow L5 † | 94560 | militia | 733 | 29.9 | 42514 | 10025 | 3167 | -5853 | never | 2.0 |
| crater L5 † | 116700 | militia | 504 | 29.4 | 29232 | 4250 | 2177 | -6521 | never | 1.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

### Race `gravewarden`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `bonelord`; first army at h22.2; trained 10, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 126.6 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 96% | 333.5 | 0.0 | 12 | 12 | 75 | wolf_den L1 |
| 2 | 63/150 | 100% | 384.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 465.8 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 554.8 | 133.3 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 613.6 | 41.2 | 12 | 26 | 158 | wolf_den L1 |
| 16 | 180/214 | 100% | 716.7 | 39.8 | 12 | 31 | 182 | wolf_den L1 |
| 24 | 241/315 | 100% | 1004.4 | 28.7 | 22 | 40 | 283 | crater L1 |

Queue idle 94.2%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace

**Careless bot (ignores wages)** — target unit `bonelord`; first army at h22.2; trained 10, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 40/150 | 100% | 126.6 | 0.0 | 12 | 7 | 55 | wolf_den L1 |
| 1 | 52/150 | 96% | 333.5 | 0.0 | 12 | 12 | 75 | wolf_den L1 |
| 2 | 63/150 | 100% | 384.5 | 0.0 | 12 | 13 | 79 | wolf_den L1 |
| 4 | 82/150 | 100% | 465.8 | 0.0 | 12 | 15 | 87 | wolf_den L1 |
| 8 | 125/214 | 100% | 554.8 | 133.3 | 12 | 21 | 133 | wolf_den L1 |
| 12 | 159/214 | 100% | 613.6 | 41.2 | 12 | 26 | 158 | wolf_den L1 |
| 16 | 180/214 | 100% | 716.7 | 39.8 | 12 | 31 | 182 | wolf_den L1 |
| 24 | 241/315 | 100% | 1004.4 | 28.7 | 22 | 40 | 283 | crater L1 |

Queue idle 94.2%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 94% of the time, so saving up, not timers, sets the pace

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| wolf_den L1 | 308 | militia | 10 | 30.0 | 580 | 250 | 43 | 33 | 17.7 | 6.0 |
| wolf_den L2 | 770 | militia | 24 | 29.2 | 1392 | 600 | 104 | 90 | 15.5 | 6.0 |
| brigand_camp L1 | 810 | militia | 17 | 29.4 | 986 | 488 | 73 | 125 | 7.9 | 4.0 |
| wolf_den L3 | 1733 | militia | 56 | 28.6 | 3248 | 1540 | 242 | 369 | 8.8 | 6.0 |
| brigand_camp L2 | 2025 | bonelord | 14 | 28.6 | 2464 | 1220 | 81 | 435 | 5.7 | 4.0 |
| barrow L1 † | 3024 | militia | 28 | 28.6 | 1624 | 422 | 121 | -163 | never | 2.0 |
| wolf_den L4 | 3465 | militia | 111 | 29.7 | 6438 | 3080 | 480 | 688 | 9.4 | 6.0 |
| crater L1 † | 4800 | bonelord | 8 | 25.0 | 1408 | 365 | 46 | -33 | never | 1.0 |
| brigand_camp L3 | 5033 | militia | 114 | 29.8 | 6612 | 3180 | 492 | 717 | 9.2 | 4.0 |
| wolf_den L5 | 6930 | militia | 224 | 29.9 | 12992 | 5900 | 968 | 1048 | 12.4 | 6.0 |
| barrow L2 † | 8064 | bonelord | 24 | 29.2 | 4224 | 1054 | 138 | -318 | never | 2.0 |
| crater L2 † | 12000 | bonelord | 18 | 27.8 | 3168 | 730 | 104 | -254 | never | 1.0 |
| brigand_camp L4 | 13310 | militia | 215 | 29.8 | 12470 | 6360 | 929 | 1715 | 7.3 | 4.0 |
| barrow L3 † | 23140 | militia | 194 | 29.9 | 11252 | 2508 | 838 | -1694 | never | 2.0 |
| crater L3 † | 26400 | bonelord | 40 | 30.0 | 7040 | 1458 | 230 | -884 | never | 1.0 |
| brigand_camp L5 | 27200 | militia | 407 | 29.7 | 23606 | 11960 | 1758 | 3191 | 7.4 | 4.0 |
| barrow L4 † | 48480 | militia | 396 | 29.8 | 22968 | 5014 | 1711 | -3541 | never | 2.0 |
| crater L4 † | 59700 | militia | 264 | 29.9 | 15312 | 2430 | 1140 | -3289 | never | 1.0 |
| barrow L5 † | 94560 | militia | 727 | 29.8 | 42166 | 10025 | 3141 | -5681 | never | 2.0 |
| crater L5 † | 116700 | militia | 497 | 30.0 | 28826 | 4250 | 2147 | -6545 | never | 1.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

## Pack `void-frontier` 0.1.0

Pack hash `469f2e879306…`. World speed ×1. Income value is Σ net/h × exchange rate.

### Race `concord`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `aegis`; first army at h5.8; trained 61, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 60/165 | 100% | 148.4 | 5.0 | 10 | 7 | 49 | - |
| 1 | 60/165 | 100% | 576.6 | 1.7 | 10 | 16 | 85 | - |
| 2 | 61/165 | 100% | 628.9 | 1.7 | 10 | 17 | 89 | - |
| 4 | 61/165 | 90% | 720.8 | 1.5 | 10 | 20 | 101 | - |
| 8 | 61/220 | 85% | 725.0 | 4.0 | 17 | 23 | 145 | derelict L1 |
| 12 | 61/220 | 85% | 712.4 | 8.9 | 31 | 24 | 191 | derelict L1 |
| 16 | 61/220 | 85% | 701.6 | 13.1 | 43 | 25 | 231 | derelict L2 |
| 24 | 61/220 | 85% | 676.4 | 22.8 | 71 | 25 | 315 | derelict L2 |

Queue idle 96.5%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 96% of the time, so saving up, not timers, sets the pace
- ⚠ population stalled from hour 16 at 61 of 220 housing (growth mode "idle": growth needs idle people, and every one has a job)

**Careless bot (ignores wages)** — target unit `aegis`; first army at h5.8; trained 61, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 60/165 | 100% | 148.4 | 5.0 | 10 | 7 | 49 | - |
| 1 | 60/165 | 100% | 576.6 | 1.7 | 10 | 16 | 85 | - |
| 2 | 61/165 | 100% | 628.9 | 1.7 | 10 | 17 | 89 | - |
| 4 | 61/165 | 90% | 720.8 | 1.5 | 10 | 20 | 101 | - |
| 8 | 61/220 | 85% | 725.0 | 4.0 | 17 | 23 | 145 | derelict L1 |
| 12 | 61/220 | 85% | 712.4 | 8.9 | 31 | 24 | 191 | derelict L1 |
| 16 | 61/220 | 85% | 701.6 | 13.1 | 43 | 25 | 231 | derelict L2 |
| 24 | 61/220 | 85% | 676.4 | 22.8 | 71 | 25 | 315 | derelict L2 |

Queue idle 96.5%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 96% of the time, so saving up, not timers, sets the pace
- ⚠ population stalled from hour 16 at 61 of 220 housing (growth mode "idle": growth needs idle people, and every one has a job)

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| nest L1 | 448 | trooper | 10 | 30.0 | 600 | 300 | 10 | 110 | 5.5 | 8.0 |
| nest L2 | 1792 | sniper | 40 | 30.0 | 2560 | 600 | 27 | -195 | never | 8.0 |
| derelict L1 † | 3500 | trooper | 18 | 27.8 | 1080 | 153 | 18 | -165 | never | 2.0 |
| nest L3 | 4032 | trooper | 96 | 29.2 | 5760 | 900 | 96 | -878 | never | 8.0 |
| nest L4 | 7168 | trooper | 170 | 30.0 | 10200 | 1200 | 170 | -2030 | never | 8.0 |
| derelict L2 † | 14000 | trooper | 64 | 29.7 | 3840 | 306 | 64 | -898 | never | 2.0 |
| derelict L3 † | 31500 | trooper | 140 | 30.0 | 8400 | 459 | 140 | -2201 | never | 2.0 |
| derelict L4 † | 56000 | trooper | 247 | 29.6 | 14820 | 612 | 247 | -4022 | never | 2.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

### Race `swarm`

**Wage-aware bot (trains only while upkeep income covers the new units)** — target unit `hover`; first army at h5.9; trained 43, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 60/150 | 100% | 157.4 | 4.3 | 10 | 7 | 49 | - |
| 1 | 60/150 | 100% | 601.2 | 1.5 | 10 | 16 | 85 | - |
| 2 | 61/150 | 100% | 648.7 | 1.5 | 10 | 17 | 89 | - |
| 4 | 61/150 | 90% | 742.4 | 1.3 | 10 | 20 | 101 | - |
| 8 | 61/200 | 90% | 736.4 | 3.2 | 15 | 23 | 139 | derelict L1 |
| 12 | 61/200 | 90% | 723.2 | 7.4 | 26 | 23 | 172 | derelict L1 |
| 16 | 61/200 | 80% | 730.8 | 10.8 | 32 | 26 | 202 | derelict L1 |
| 24 | 61/200 | 80% | 705.6 | 19.7 | 53 | 27 | 269 | derelict L2 |

Queue idle 96.0%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 96% of the time, so saving up, not timers, sets the pace
- ⚠ population stalled from hour 16 at 61 of 200 housing (growth mode "idle": growth needs idle people, and every one has a job)

**Careless bot (ignores wages)** — target unit `hover`; first army at h5.9; trained 43, deserted 0.

| Hour | Pop/Housing | Employment | Income value/h | Wage % | Garrison | Build levels | Score | Strongest neutral cleared |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 60/150 | 100% | 157.4 | 4.3 | 10 | 7 | 49 | - |
| 1 | 60/150 | 100% | 601.2 | 1.5 | 10 | 16 | 85 | - |
| 2 | 61/150 | 100% | 648.7 | 1.5 | 10 | 17 | 89 | - |
| 4 | 61/150 | 90% | 742.4 | 1.3 | 10 | 20 | 101 | - |
| 8 | 61/200 | 90% | 736.4 | 3.2 | 15 | 23 | 139 | derelict L1 |
| 12 | 61/200 | 90% | 723.2 | 7.4 | 26 | 23 | 172 | derelict L1 |
| 16 | 61/200 | 80% | 730.8 | 10.8 | 32 | 26 | 202 | derelict L1 |
| 24 | 61/200 | 80% | 705.6 | 19.7 | 53 | 27 | 269 | derelict L2 |

Queue idle 96.0%, population at housing cap 0.0%, storage cap 0.0%, food-starved 0.00 h, upkeep-starved 0.00 h.

- ⚠ resource-bound: the build queue is empty 96% of the time, so saving up, not timers, sets the pace
- ⚠ population stalled from hour 16 at 61 of 200 housing (growth mode "idle": growth needs idle people, and every one has a job)

**PvE difficulty** — cheapest single-unit army with ≤30% losses, race modifiers only, 6 tiles one way.

| Neutral | Guard power | Unit | Army | Loss % | Train value | Reward | Wages | Net/run | Payback runs | Runs/day |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|
| nest L1 | 448 | ravager | 6 | 16.7 | 540 | 300 | 4 | 206 | 2.6 | 8.0 |
| nest L2 | 1792 | ravager | 24 | 29.2 | 2160 | 600 | 17 | -48 | never | 8.0 |
| derelict L1 † | 3500 | ravager | 9 | 22.2 | 810 | 153 | 6 | -33 | never | 2.0 |
| nest L3 | 4032 | ravager | 52 | 28.8 | 4680 | 900 | 36 | -484 | never | 8.0 |
| nest L4 | 7168 | ravager | 94 | 28.7 | 8460 | 1200 | 66 | -1294 | never | 8.0 |
| derelict L2 † | 14000 | ravager | 34 | 29.4 | 3060 | 306 | 24 | -617 | never | 2.0 |
| derelict L3 † | 31500 | ravager | 77 | 28.6 | 6930 | 459 | 54 | -1577 | never | 2.0 |
| derelict L4 † | 56000 | ravager | 134 | 29.9 | 12060 | 612 | 94 | -3088 | never | 2.0 |

† The reward includes a resource without `exchangeRates`; the simulator values it at 1 per unit, so net value is a lower bound.

