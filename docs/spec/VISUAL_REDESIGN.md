# Visual reconstruction — Ashen Crowns

Updated: 2026-10-31
Branch: improve/game-experience-audit-20261031
Status: In progress

## Goal

Create a familiar, game-first experience for veteran browser-strategy players.
Keep the approved muted palette.
Use original art, names, textures, and ornamentation.
Preserve all existing server commands and Game Pack rules.
Do not claim visual parity without screenshot comparison.

## Visual identity

- Wood: dark walnut and charcoal-brown frames.
- Parchment: aged beige and quiet bronze.
- Interaction: muted olive-green action buttons.
- Warning: restrained red.
- Text: ivory on dark wood and brown ink on parchment.
- Motion: slow reveal, subtle hover, and meaningful progress.
- Navigation: dedicated persistent game controls.
- Illustrations: character, unit, and location art remain original.

## Screen implementation

| Screen | Layout | Controls | Assets | Browser review |
|---|---|---|---|---|
| City | World-first stage, building tray, right queue rail, quick actions | Existing build and navigation | Existing sprites | Pending |
| Army | Select-one-unit roster, illustrated detail chamber, training dock | Existing training commands | Existing unit portraits | Pending |
| Heroes | Selectable hero roster, large portrait, equipment slots, skill rail | Existing hero commands | Existing hero and item icons | Pending |
| World map | Atlas frame, illustrated selected-location inspector, mission controls | Existing missions and map pan | Existing terrain and sites | Pending |
| Research | Tiered technology tree and specialization | Existing research commands | Glyphs per tier | Pending |
| Economy | Unified framed statistics and actions | Existing controls | Resource artwork | Pending |
| Market | Unified game panels | Existing trading controls | Resource artwork | Pending |
| Alliances | Unified game panels | Existing diplomacy actions | Existing icons | Pending |
| Missions and reports | Unified lists, status tags, report modal | Existing actions | Existing icons | Pending |
| Quests and mail | Unified framed panels | Existing actions | Existing icons | Pending |
| Login and world join | Game entry framing | Existing authentication | Original crest | Pending |

## UI architecture

The game UI remains React and CSS.
The desktop shell includes a bottom command dock with real stationed units.
A compact left icon rail keeps secondary destinations accessible.
The top bar retains resources, kingdom specialisation, and hero shortcuts.
Game state remains server-authoritative.
World navigation uses the existing screen context.
Game Pack art remains resolved through PackIndex.
The shared look lives in `client/src/game/realm-theme.css`.
Research-specific changes live in `client/src/game/experience.css`.
The new screen layouts live under `client/src/screens/`.

## Interaction integrity

1. Keep every server command intact.
2. Retain loading and error states.
3. Keep keyboard navigation.
4. Respect reduced-motion settings.
5. Preserve Russian and English string keys.
6. Show server-provided prerequisites and blocked states.
7. Display actual player hero portraits.
8. Keep audio opt-in.
9. Do not infer completed gameplay from a successful build.
10. Test mobile overflow after each layout change.

## Screenshot review

The workflow `.github/workflows/visual-gameplay.yml` uses Docker Compose and Playwright.
It creates screenshots for desktop and mobile.
It uploads screenshots as a CI artifact.
A screenshot is evidence of rendering, not proof of visual equivalence.

Review each screenshot for:
1. No horizontal or vertical overflow inside modal controls.
2. Visible resource bar and right-side player shortcuts.
3. Readable labels in both locales.
4. No missing textures or broken sprites.
5. Clear selection and disabled states.
6. No overlapping artwork and actions.
7. Readable layout at 1360 × 860 and 390 × 844.
8. Consistent art scale for terrain, heroes, and army.
9. Stable browser console during the player journey.
10. Meaningful interaction through each panel.

## Known art and layout gaps

- Terrain sprites are not seamless.
- Ten unit illustrations still use simpler SVG artwork.
- Research uses glyph icons rather than a painted tier atlas.
- Building progression lacks painted level variants.
- Combat animation is not built.
- The city stage uses predetermined plots, not free construction placement.
- Screens need human comparison against reference images.
- Source reference screenshots show a game-first world viewport and a bottom action dock.
- A browser build passing does not prove that all art matches the reference.
- The new browser visual workflow needs passing runtime results.

Do not mark the redesign complete before these gaps are reviewed.

## Browser screenshot review — first successful run

Workflow: https://github.com/XenosRbel/MyLands/actions/runs/41241133060

The browser journey completed.
The screenshot artifact contains twenty desktop and mobile images.
The tested layout uses the new city, army, hero, and map screens.
The game still requires a final screenshot pass after later fixes.

### Defects found and addressed

- The city had a large empty black margin. The ground now fills the stage.
- Research details used dark text against a dark card. The inspector now uses parchment.
- The research page stacked every branch vertically. Tabs now show one branch.
- The map inspector initially had no selected location. The owned city is selected when visible.
- The entry form lacked faction artwork. Race emblems are now displayed.
- The scene lacked a lower command dock. The desktop shell now shows one.
- The original E2E used stale building selectors. They were corrected.

### Open quality work

- Review rendered screenshots after all fixes.
- Compare Russian and English interface captures.
- Verify long translated strings at small widths.
- Replace remaining low-detail unit SVGs with painted illustrations.
- Produce seamless terrain art and level-specific building art.
- Add hero skill and research illustrations with unique tier designs.
- Review performance of illustrated maps on low-end mobile hardware.

The completed browser workflow proves navigation and rendering at tested viewports.
It does not certify complete similarity with the original My Lands.
