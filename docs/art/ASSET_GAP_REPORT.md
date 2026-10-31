# Asset coverage report — 2026-10-31

## Verified against repository tree
The Ashen Crowns pack now resolves artwork for all 82 required game entities.
The checked entities include resources, buildings, units, terrain, neutral sites, hero classes, hero items, and races.
The report checks file existence, not artistic quality or runtime rendering.

| Category | Coverage |
|---|---:|
| Resources | 7/7 |
| Buildings | 22/22 |
| Units | 22/22 |
| Hero classes | 4/4 |
| Hero items | 11/11 |
| Terrain | 8/8 |
| Neutral locations | 4/4 |
| Race emblems | 4/4 |

Two additional faction emblems are available.

## Changes
- Added ten original SVG illustrations for missing units.
- Added four original SVG neutral-site illustrations.
- Added four race emblems and two faction emblems.
- Configured each unit icon to use its artwork.
- Added neutral-site artwork to the world-map canvas.
- Added `npm run check:art`.
- Made `npm run build` run the asset check first.

## Still incomplete
- The new vector artwork is simpler than the existing painted sprites.
- A consistent hand-painted replacement pass is still needed.
- UI 9-slice frames remain missing.
- The terrain images do not tile seamlessly.
- Building upgrade-level variants remain missing.
- Login, combat, report, and quest animations need a complete design pass.
- Original My Lands visual equivalence has not been verified.

## Validation
The repository tree and Game Pack artwork paths were checked through GitHub.
No missing target files were found in that check.
A browser render test and a production build remain unverified.
