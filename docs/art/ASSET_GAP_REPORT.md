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

## Visual quality correction — 2026-11-01

File existence is not an art-quality acceptance test.
The first report claimed 22/22 units only because paths existed.
Ten unit pictures are low-detail SVG placeholders.
Four neutral location images are also SVG placeholders.
This is a visual release blocker.

### Replace with consistent illustrated assets

- Units: griffin, ballista, wagon, settler, wolf, brigand, troll, wraith, golem, drake.
- Neutral sites: wolf_den, brigand_camp, barrow, crater.
- Hero inventory already has original painted WebP assets.
- The map now renders the painted keep for city markers.
- Research tier nodes now reuse existing painted building art.
- Research still needs dedicated illustrated technology icons.

### Art direction and acceptance

- Match the existing hand-painted WebP pack, not flat iconography.
- Keep subdued lighting, readable silhouettes, and natural materials.
- Use consistent perspective and transparent backgrounds for units.
- Distinguish faction troops from neutral monsters visually.
- Provide one large illustration and one derived small icon per entity.
- Validate 48px, 96px, and 256px rendering.
- Do not stretch, recolor, or silently reuse a different entity image.
- Check every final image in army, map, hero, and research screens.
- Reject a release if any placeholder remains in the illustrated entity set.

### Implementation status

Map city markers, hero items, and research illustration wiring were improved.
The 14 low-quality SVG illustrations have not yet been replaced.
No claim of 1:1 visual parity or complete art coverage is justified.

## Illustrated SVG pass — 2026-11-01

All fourteen original flat placeholder SVG files were redrawn in place.
The new artwork uses transparent backgrounds, layered materials, shaded silhouettes, and texture filters.
Illustrations were adjusted to be readable at small UI sizes.
They no longer use the generic circular badge template.

Affected units: griffin, ballista, wagon, settler, wolf, brigand, troll, wraith, golem, drake.
Affected sites: wolf_den, brigand_camp, barrow, crater.

The Game Pack asset paths stay unchanged.
The existing icon and artwork consumers automatically receive the new SVGs.
The art integrity script now rejects old flat 320-pixel placeholder backgrounds.

Important limitation: the new SVGs are detailed vector illustrations, not hand-painted WebP portraits.
They still require aesthetic comparison against the painted art.
Do not classify the painted-art parity requirement as passed.
