Ниже готовый набор промтов для генерации графики «Пепельных Корон». Сами промты на английском: генераторы изображений (Midjourney, SDXL/Flux, GPT-Image) понимают его лучше. Названия ассетов совпадают с ID из `packs/ashen-crowns/pack.json`, так что готовые файлы подключаются в Game Pack без правки кода.

## 1. Общий стиль (добавлять в начало каждого промта)

```
Art style bible: hand-painted 2D fantasy strategy game art, classic browser MMO strategy look,
warm painterly digital illustration, rich but slightly muted palette (ember orange #E0A44A,
ash brown #2D2520, bone ivory #F1E6D6, moss green, steel blue accents), soft top-left light,
subtle dark outline, clean readable silhouette at small size, consistent scale and perspective
across the whole set, high detail but not noisy, no text, no letters, no watermark, no logo,
original design, game asset, centered, isolated.
Setting: "Ashen Crowns" — a medieval-fantasy realm rebuilt after the sky shattered; ash-covered
lands, glowing starglass shards fallen from the sky, rival noble houses.
```

Негативный промт для SD/Flux:

```
text, letters, watermark, signature, logo, UI frame, border, blurry, lowres, jpeg artifacts,
photorealistic, 3d render, plastic, anime, chibi, extra limbs, cropped, multiple objects,
busy background, gradient background
```

## 2. Шаблоны по типам ассетов

**Иконки ресурсов** (128×128, PNG с прозрачностью):

```
[STYLE] Game resource icon of {SUBJECT}, single object, 3/4 view, bold silhouette,
readable at 24px, transparent background, square composition.
```

| ID | {SUBJECT} |
|---|---|
| `timber` | a bundle of three cut logs tied with rope |
| `grain` | a golden wheat sheaf |
| `stone` | a stack of grey cut stone blocks |
| `iron` | two dark iron ingots with a faint orange forge glow |
| `gold` | a small pile of gold coins with an embossed crown |
| `lore` | a rolled parchment scroll with a wax seal and a quill |
| `starglass` | a glowing violet-blue crystal shard with tiny star sparks |

**Здания** (512×512, изометрия 2:1, PNG с прозрачностью, одинаковый масштаб):

```
[STYLE] Isometric 2:1 game building sprite, {SUBJECT}, standing on a small patch of
ash-dusted ground, same camera angle and scale as the rest of the set, transparent background,
level-3 appearance (sturdy, partly upgraded).
```

| ID | {SUBJECT} |
|---|---|
| `keep` | a stone castle keep with banners, central tower, the heart of the city |
| `cottages` | a cluster of timber-framed cottages with smoking chimneys |
| `lumberyard` | a lumber mill with log piles and a saw shed |
| `farmstead` | a farmhouse with a windmill and golden fields |
| `quarry` | an open stone quarry with a wooden crane |
| `ironworks` | a smelting forge with chimneys and glowing furnace |
| `tradepost` | a small merchant stall with awning and crates |
| `mint` | a fortified stone mint with coin presses |
| `scriptorium` | a scholars' hall with tall arched windows and book stacks |
| `academy` | a domed academy with observatory telescope |
| `warehouse` | a large wooden warehouse with sacks and barrels |
| `vault` | a hidden stone vault door in a hillside |
| `walls` | a segment of stone city wall with gate and towers |
| `watchtower` | a tall wooden watchtower with a signal fire |
| `barracks` | military barracks with training dummies and weapon racks |
| `range` | an archery range with targets |
| `stables` | horse stables with hay and paddock |
| `aerie` | a cliff-top aerie with griffin nests |
| `workshop` | a siege workshop with an unfinished ballista |
| `sanctum` | a mystic sanctum with floating runes and violet glow |
| `market` | a busy market hall with colourful stalls |
| `warhall` | a grand hall of heroes with statues and braziers |

Для уровней одного здания (по желанию): `… level 1: small wooden version` / `… level 10: large stone, gilded roof` / `… level 20: monumental, glowing starglass details`.

**Юниты** (512×512, портрет по пояс, нейтральный тёмный фон):

```
[STYLE] Fantasy unit portrait, waist-up, 3/4 view, {SUBJECT}, dramatic rim light,
plain dark warm background (#221C18), facing right, consistent framing across the set.
```

| ID | {SUBJECT} |
|---|---|
| `militia` | a levy spearman in padded jacket |
| `pikeman` | a disciplined pikeman in half-plate |
| `bowman` | a longbow archer with green hood |
| `crossbow` | a crossbowman with steel cap |
| `outrider` | a light cavalry rider on a fast horse |
| `knight` | a heavy lancer knight on armoured horse |
| `griffin` | a rider on a griffin, wings spread |
| `ballista` | a siege ballista with two crew |
| `battlemage` | a battle mage casting ember fire |
| `scout` | a cloaked scout with spyglass |
| `wagon` | a supply wagon with draft horse |
| `settler` | a settler family with cart and tools |
| `thornweaver` | an elven-like forest archer with thorn-vine bow (Thornkin) |
| `bannerguard` | a knightly shield-wall soldier with tall banner (Ironvow) |
| `cinderhound` | a fire-touched hound beast with ember fur (Ashborn) |
| `bonelord` | a necromancer knight in bone armour (Gravewardens) |
| `wolf` | a grey wolf (neutral) |
| `brigand` | a forest brigand (neutral) |
| `troll` | a hill troll (neutral) |
| `wraith` | a barrow wraith (neutral) |
| `golem` | a stone golem with starglass core (neutral) |
| `drake` | a young ash drake (neutral) |

**Герои** (512×640, портрет-«карточка»):

```
[STYLE] Heroic character portrait, chest-up, {SUBJECT}, confident pose, detailed face,
ornate but readable costume, dark vignette background.
```

| ID | {SUBJECT} |
|---|---|
| `marshal` | a battle-scarred marshal with commander's cloak |
| `warden` | a shield-bearing warden in heavy armour |
| `pathfinder` | a ranger pathfinder with map and compass |
| `sage` | a wise sage with glowing tome |

**Предметы героя** (128×128, PNG с прозрачностью), шаблон как у иконок ресурсов:

| ID | {SUBJECT} |
|---|---|
| `iron_blade` | an iron blade |
| `ember_sword` | a sword with glowing ember edge |
| `skyfall_lance` | a lance tipped with starglass |
| `leather_jerkin` | a leather jerkin |
| `warden_plate` | heavy plate armour |
| `starglass_mail` | chainmail woven with starglass |
| `scout_hood` | a scout's hood |
| `horned_helm` | a horned helmet |
| `lucky_coin` | a lucky coin |
| `wind_charm` | a feather wind charm |
| `sage_tome` | a sage's tome |

Модификатор редкости в конец промта: `common: plain` / `rare: blue glow` / `epic: violet glow` / `legendary: golden radiant glow`.

**Тайлы карты** (256×256, вид строго сверху, бесшовные):

```
[STYLE] Seamless tileable top-down map terrain texture, {SUBJECT}, orthographic,
no perspective, no objects crossing the edges, flat even lighting.
```

| ID | {SUBJECT} |
|---|---|
| `plains` | green-gold grassland |
| `forest` | dense forest canopy |
| `hills` | rocky brown hills |
| `ironridge` | grey ore-veined rocks |
| `ashland` | cracked ash ground with embers |
| `marsh` | murky marsh with reeds |
| `mountains` | snowy peaks (impassable) |
| `lake` | deep blue water |

**Объекты на карте** (256×256, изометрия, PNG с прозрачностью):

| ID | {SUBJECT} |
|---|---|
| `wolf_den` | a wolf den cave |
| `brigand_camp` | a brigand tent camp |
| `barrow` | an ancient burial barrow with standing stones |
| `crater` | a smoking starfall crater with glowing crystals |

Плюс иконка «город игрока» в четырёх цветах рас.

**Гербы** (512×512, PNG с прозрачностью):

```
[STYLE] Heraldic emblem on a shield, {SUBJECT}, flat painted heraldry, centered.
```

| ID | {SUBJECT} |
|---|---|
| `thornkin` | a thorned oak, green |
| `ironvow` | a crossed lance and gauntlet, blue |
| `ashborn` | a burning crown, red-orange |
| `gravewarden` | a skull under a lantern, violet |
| `dawnpact` (фракция) | a rising sun |
| `duskbound` (фракция) | a crescent moon over ash |

**Интерфейс** (9-slice-элементы, PNG с прозрачностью):

```
[STYLE] Game UI kit element, {SUBJECT}, dark carved wood and blackened bronze frame with
subtle ember-gold trim and parchment inner area, designed for 9-slice scaling (plain stretchable
center, decorated corners), no text, transparent outside the frame.
```

| Файл | {SUBJECT} |
|---|---|
| `ui/panel` | a large window panel |
| `ui/modal` | a dialog window |
| `ui/button_primary` | a primary button, gold |
| `ui/button_secondary` | a secondary button, bronze |
| `ui/resource_slot` | a resource bar slot |
| `ui/tab` | a tab |
| `ui/progress_bar` | an empty progress bar frame |
| `ui/tooltip` | a tooltip frame |

**Фоны** (1920×1080 и 1080×1920):

```
[STYLE] Wide painted key art background, {SUBJECT}, atmospheric, space left for UI,
no characters in the foreground.
```

| Файл | {SUBJECT} |
|---|---|
| `bg/login` | a ruined kingdom under a shattered night sky with falling starglass |
| `bg/city` | a green valley with a walled medieval city seen from above, warm sunset |

## 3. Правила, чтобы пак получился цельным

1. Сначала сгенерируйте 3–4 «эталона»: `keep`, `militia`, `timber`, `plains`. Остальное делайте с ними как с референсом: style reference / IP-Adapter в SD, `--sref` в Midjourney. Тот же seed для каждой серии.
2. Одна серия — одни настройки: одинаковые ракурс, свет и масштаб.
3. Не пишите в промтах названия чужих игр и «in the style of …». Иначе можно получить узнаваемые чужие элементы, а нам нужна оригинальная графика.
4. Именование и структура файлов:

   ```
   packs/ashen-crowns/assets/
     resources/timber.png …
     buildings/keep.png …
     units/militia.png …
     heroes/marshal.png …
     items/iron_blade.png …
     terrain/plains.png …
     neutrals/wolf_den.png …
     emblems/thornkin.png …
     ui/… bg/…
   ```

5. Зафиксируйте происхождение графики (какой генератор и с какой лицензией) в `theme.assetsLicense` пака. Это нужно для модерации в Яндекс Играх и Google Play.

## Что потом потребуется от меня

Иконки ресурсов и предметов можно подключить уже сейчас, прописав пути в `theme.icons` пака. Для остального нужны доработки клиента:
- картинки зданий, юнитов и героев на карточках;
- текстурные тайлы на карте;
- 9-slice-рамки интерфейса.

Если хотите, сохраню этот промт в репозиторий как `docs/art/ART_PROMPTS.md` и сделаю эти доработки, когда графика будет готова.