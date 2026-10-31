# Visual production stages — authoritative execution plan

Updated: 2026-11-01
Branch: improve/game-experience-audit-20261031
Reference: the approved My Lands-inspired castle-screen screenshot.

## Immutable visual requirements

1. Preserve the approved muted gold, purple, parchment and landscape palette.
2. Keep a named left navigation panel.
3. Keep the top resource strip, clock, specialisation and hero portrait.
4. Keep a permanent right construction and troop-training column.
5. Keep a large illustrated city viewport.
6. Keep the bottom command dock and illustrated troop roster.
7. Do not ship SVG as final player-visible game artwork.
8. Use PNG or WebP with consistent painted art direction.
9. Do not treat SVG-to-PNG conversion as an artistic replacement.
10. Preserve authoritative server actions and Game Pack rules.
11. Include screenshot review before accepting a stage.
12. Keep separate completion statuses for visual quality and CI.

## Stage gates

| Stage | Deliverable | Status |
|---|---|---|
| 0 | Asset inventory, disallowed-art detector and policies | In progress |
| 1 | Unique raster research icons and progression states | Implemented in code; final art acceptance pending |
| 2 | Raster resource and top-HUD icon atlas | Implemented base; responsive and artistic review pending |
| 3 | Raster bottom HUD, globe and command controls | Not started |
| 4 | Painted coherent city scene, not vector landscape | Not started |
| 5 | Level-variant building portraits and footprints | Not started |
| 6 | Painted queue cards and progress controls | Not started |
| 7 | Hero, army and equipment layout and artwork | Not started |
| 8 | World map terrain and neutral locations | Not started |
| 9 | Game dialogs, market, reports, mail and diplomacy | Not started |
| 10 | Replace all remaining unit and entity SVG art | Not started |
| 11 | Animations, optional audio and reduced motion | Not started |
| 12 | Full desktop/mobile image comparison and review | Not started |

A stage is complete only when its assets, wiring, tests, screenshots and documentation pass.
CI success alone is not visual acceptance.

## Current known art violations

- Inline SVG glyph renderer: client/src/ui/kit.tsx.
- Inline research seals: client/src/game/ResearchEmblem.tsx.
- SVG world globe: client/public/ui/world-globe.svg.
- SVG city environment: client/public/ui/city-landscape.svg.
- SVG decoration: client/public/ui/medieval-corners.svg.
- SVG textures: client/public/ui/parchment-grain.svg and dark-wood.svg.
- SVG entry landscape: client/public/ui/realm-landscape.svg.
- Ten SVG unit illustrations under packs/ashen-crowns/assets/units/.
- Four SVG neutral-site illustrations under packs/ashen-crowns/assets/neutrals/.
- Race and faction SVG emblems under packs/ashen-crowns/assets/emblems/.

The original painted building, resource, hero and equipment WebP art is not listed as a violation.
A raster file may still fail art review if it merely embeds old flat SVG imagery.

## Workflow for each stage

1. Collect every image reference used by the affected screen.
2. Produce original raster sprites and record their source.
3. Verify dimensions, alpha edges, artistic consistency and filename stability.
4. Connect the sprites to current UI elements without changing game commands.
5. Run client, server, Docker and browser tests.
6. Review screenshots at 1360 × 860, 1680 × 900 and 390 × 844.
7. Record actual remaining differences and completed changes.
8. Advance only when the stage has been reviewed.

## Release gate

The final strict raster audit must pass with zero SVG asset URLs and zero inline SVG icons.
This cannot be declared complete while legacy SVG assets remain visible.

## Research stage — implementation pass

- Every research in the Ashen Crowns pack now has a unique raster art identity.
- All 134 technology-level combinations have individually generated WebP files.
- The art changes at each new research level, not only at levels 4 and 7.
- The research tree and side panel both use the generated raster art.
- The detail panel compares the current and next level visually.
- Locked, completed, available, pending, and selected states remain distinct.
- A client build check verifies that all 134 generated WebP assets exist and have the correct file header.
- The browser screenshot sheet now covers all technology levels.
- Existing server research costs, effects, exclusivity, and requirements were not changed.
- Human art-direction review is still needed before final sign-off.
- The next implementation area is the top resource and kingdom HUD.

A procedural WebP icon is still a prototype until its artistic quality matches the approved resource pack.

## Top HUD — implementation pass

- Retained the approved resource strip, clock, selected kingdom specialisation and hero position.
- Enlarged the seven existing painted resource icons from the Game Pack.
- Reused actual research WebP artwork for the selected kingdom specialisation.
- Expanded hero portrait and level badge proportions on large desktop screens.
- Preserved real resource balances, production rates and storage warnings from server state.
- Browser QA waits for all resource images to load and rejects missing or vector icons.
- Pending: compare screenshots at 1360x860 and 1680x900 and correct text collisions.
- Pending: painted emblem/frame quality still needs explicit reference approval.
