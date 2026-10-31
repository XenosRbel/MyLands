# Asset gap report (after import of Resource Pack v0.1)

Integrated (`theme.icons` / `theme.art` in `packs/ashen-crowns/pack.json` 1.2.0):

| Category | Count | Used in client |
|---|---|---|
| Resource icons | 7/7 | resource bar, costs, rewards |
| Item icons | 11/11 | hero equipment and inventory |
| Building sprites | 22/22 | isometric city scene, building cards, building dialog |
| Unit portraits | 12/22 | army cards (others fall back to text) |
| Hero class portraits | 4/4 | heroes screen |
| Terrain textures | 8/8 | world map canvas, city ground |

Not used / still missing:
- **UI frame crops** (panel, dialog, buttons, tabs, banner…): delivered with an opaque black background instead of transparency, so they cannot be layered; not imported. Need transparent 9-slice exports.
- Unit portraits: `wagon`, `settler`, `wolf`, `brigand`, `troll`, `wraith`, `golem`, `drake`.
- Neutral map sites: `wolf_den`, `brigand_camp`, `barrow`, `crater` (map shows numbered level markers instead).
- Race and faction emblems; player-city markers in race colours (map draws vector house markers).
- Login and city backgrounds; building level variants (1/10/20); animations; hover/disabled states.
- Terrain tiles are not truly seamless; the map shows visible tile seams. A seamless set (or an isometric map style) is recommended.
