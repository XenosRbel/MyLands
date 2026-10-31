# Game Pack artwork audit

Updated: 2026-11-01
Game Pack: ashen-crowns
Scope: original assets and runtime UI integration

## Coverage

| Entity family | Art paths | Style review |
|---|---:|---|
| Resources | 7/7 | Painted WebP present |
| Buildings | 22/22 | Painted WebP present |
| Units | 22/22 | 12 painted WebP, 10 textured SVG illustrations |
| Hero classes | 4/4 | Painted WebP present |
| Hero equipment | 11/11 | Painted WebP icons present |
| Terrain | 8/8 | Painted WebP; seamless tiling unverified |
| Neutral locations | 4/4 | Textured SVG illustrations |
| Races | 4/4 | SVG emblems present |

Complete paths do not mean equivalent artistic quality.
The ten newer units and four neutral sites still require visual art-direction acceptance.

## Implemented improvements

1. Replaced the original 14 low-detail SVG placeholders in place.
2. Removed the flat circular backgrounds from those illustrations.
3. Added layered materials, shading, texture, and transparent backgrounds.
4. Reused painted item images for hero inventory and equipment.
5. Reused the existing painted keep image for map city markers.
6. Reused building illustrations for temporary research tier art.
7. Reused painted Game Pack art for the command dock.
8. Added a structural check against the original placeholder format.
9. Added an E2E art gallery at 120px and 48px sizes.
10. Kept all Game Pack asset paths stable.

## Not yet accepted

- The 14 redrawn SVGs are not hand-painted WebP-quality portraits.
- They require side-by-side comparison with existing unit art.
- Unique painted technology art by branch and level remains missing.
- Building level variation and staged construction visuals remain missing.
- Terrain seamless-tiling and a coherent world-map texture atlas remain unverified.
- Combat, ability, construction, and reward animations are not complete.
- Final visual-reference parity with My Lands is not established.

## Quality gates

1. Check SVG geometry and actual browser rendering.
2. Inspect all illustrations at 48px, 96px, and 256px sizes.
3. Check alpha edges on wood and parchment backgrounds.
4. Check whether silhouettes remain recognizable without labels.
5. Check that an entity never uses another entity's image by mistake.
6. Check production builds and the server's Game Pack validation.
7. Review the CI artifact `20-art-consistency.png`.
8. Keep release acceptance open until the artwork passes human review.

The art pipeline and checks do not replace a visual-design review.
