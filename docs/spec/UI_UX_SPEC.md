# UI / UX Specification: Realmforge web client

Owner: Agent E (UI/UX Designer)
Status: v1. Describes the client **as implemented** at commit `dfa660e`. Where the code falls short of the target, the gap is stated explicitly (section 9).
Date: 2026-10-31

Sources read: `client/src/**` (`App.tsx`, `main.tsx`, `screens/*.tsx`, `ui/kit.tsx`, `game/context.tsx`, `game/pack.ts`, `api/client.ts`, `platform/adapter.ts`, `i18n/ui.ts`, `styles.css`), `client/e2e/play.mjs`, `server/internal/api/routes.go`, `server/internal/api/http.go`, and `docs/spec/PLATFORM_REQUIREMENTS.md`.

---

## 1. Principles

1. **The server is authoritative.** The client renders server state and sends commands. It never predicts resource totals, battle outcomes or completion. Every number on screen comes from the last API response. The only client-side estimates are the countdown timers and the mission ETA preview (section 7.13).
2. **The pack decides the look and the screens.** The active Game Pack supplies names, descriptions, palette, fonts, icon glyph names and the list of enabled screens (`pack.ui.screens`). The engine UI (`i18n/ui.ts`) only supplies generic strings such as "Queue" and "Cancel". `applyTheme(pack)` writes palette and font tokens onto `:root`.
3. **Commands are safe to repeat.** Every state-changing call goes through `command()`. It creates one `Idempotency-Key` per user action and reuses it on automatic retries, so a retry can never apply twice (section 3.2).
4. **Nothing is ever double-submitted.** Every command button is an `<Action>`. It is disabled and shows `aria-busy` while its request is in flight.
5. **Destructive actions need confirmation.** Attacks, cancellations that refund less than 100 %, leaving or kicking from an alliance, blocking a player and deleting the account all open a `Confirm` modal first (section 3.1).

---

## 2. Application lifecycle

### 2.1 Boot sequence (`main.tsx`, `App.tsx`)

| Step | What happens | UI state |
|---|---|---|
| 1 | `data-motion` is read from `localStorage["rf.motion"]` and set on `<html>` before React mounts. | Blank page |
| 2 | `detectPlatform()` picks `YandexAdapter` (build mode `yandex`), `AndroidAdapter` (`window.Capacitor.isNativePlatform()`) or `WebAdapter`. Then `platform.init()` runs. In Yandex mode, `init()` loads `/sdk.js` and calls `YaGames.init()`. | Blank page |
| 3 | React mounts `<App>`. `booting = true`. | Full-screen spinner (`.boot`) |
| 4 | If a token is stored: `GET /api/v1/me`. If `rf.world` is set: `GET /api/v1/worlds`, then `GET /api/v1/worlds/{wid}/pack` for that world if it is still joined. | Spinner |
| 5 | `booting = false`, then `platform.signalReady()`. This is Yandex `LoadingAPI.ready()`, YG-02. | Auth screen, world screen or game shell |

A token that turns out to be invalid fails silently and leads to the auth screen. A `401` anywhere clears the token (`api.setToken(null)`). `App` listens through `api.onAuthChange` and drops back to the auth screen.

### 2.2 Local persistence

Every read and write is wrapped in `try/catch`. Private mode or blocked storage gives a session that lasts only until reload. Nothing else breaks.

| Key | Written by | Value | Purpose |
|---|---|---|---|
| `rf.token` | `api.setToken` | Opaque bearer token | Session restore. Removed on logout, on a 401 and on account deletion. |
| `rf.world` | `App` (enter world) | World id | Re-enter the last world after reload |
| `rf.screen` | `GameProvider.go()` | Screen id | Restore the last screen, e.g. the map |
| `rf.city` | `GameProvider.setCityId()` | City id | Restore the selected city. If the stored city is not in the overview, the client falls back to the first city. |
| `rf.locale` | `App.setLocale` | `en` / `ru` | UI language. It wins over the platform locale. |
| `rf.motion` | Profile screen | `"reduce"` or `""` | Reduced-motion preference |

Precedence for the locale: `rf.locale`, then `platform.locale()` (Yandex `i18n.lang` or `navigator.language`), then `en`. Only locales in `supportedLocales = ["en","ru"]` are accepted. Changing the language also sends `PUT /api/v1/me/locale` (fire-and-forget) so that the server knows the locale.

---

## 3. Shared components and behaviour (`ui/kit.tsx`, `api/client.ts`)

### 3.1 Components

| Component | Behaviour | Accessibility |
|---|---|---|
| `Action` | Async button. While running: `disabled`, `aria-busy=true`, inline spinner. An error becomes an error toast through `errorMessage(locale, code, message)`. The optional `confirm` text opens `Confirm` (danger style) before running. | Native `<button>` |
| `Modal` | Overlay plus dialog. On open it focuses the dialog container and remembers the previously focused element. `Escape` closes it, and so does a mouse-down on the backdrop. On close, focus returns to the remembered element. | `role="dialog"`, `aria-modal="true"`, `aria-label={title}`, `tabIndex=-1`; the close button has `aria-label="Close"` |
| `Confirm` | A `Modal` with Cancel and Confirm buttons. Confirm has `autoFocus` and the danger style when `danger`. | Inherits from `Modal` |
| `Tabs` | Button group | `role="tablist"`, `role="tab"`, `aria-selected` |
| `Timer` | Countdown to an ISO instant on **server time** (`api.now()`), ticking each second. Calls `onDone` once, 600 ms after it reaches zero, so the server is certainly past the due time. | Plain text. Not a live region, to avoid announcing every second. |
| `Progress` | Bar | `role="progressbar"`, `aria-valuenow/min/max`, `aria-label` |
| `Loading` | Spinner plus "Loading…" | `aria-live="polite"` |
| `ErrorBox` | Localized error text and an optional Retry button | `role="alert"` |
| `Empty` | "Nothing here yet" or custom children | — |
| `useLoad` | Fetch hook with `AbortController`. It cancels on unmount and when its dependencies change, and returns `{data, error, reload, setData}`. | — |
| Toasts | At most 4 visible. Info and ok toasts last 3.5 s, errors 6 s. Fixed bottom-right on desktop, above the bottom bar on mobile. | Container `aria-live="polite"` |
| `Icon` | Original SVG glyphs drawn in code. Decorative icons are `aria-hidden`; titled ones get `role="img"`. | — |

### 3.2 API client contract (`api/client.ts`)

- **Auth header:** `Authorization: Bearer <token>` when a token exists.
- **Commands:** `command(path, body, key = newIdempotencyKey())` sends `POST` with `Idempotency-Key`, which is 32 hex characters from `crypto.getRandomValues`.
- **Retries:** GETs and keyed commands retry up to 2 times, with backoff of 300 ms × attempt². Retries happen only on network errors, `429` and `5xx`, and always reuse the same key. Domain errors (`4xx`) are never retried. Un-keyed POSTs (auth, join, upgrade, locale, delete) are never retried.
- **401:** the token is cleared and the app returns to the auth screen.
- **Server time:** the offset is estimated from the `Date` header (RTT/2, smoothed 0.7/0.3) and set exactly from `serverTime` fields in `overview` and `cities/{id}` responses (`syncServerTime`). `api.now()` = `Date.now()` + offset.
- **Error envelope:** `{error:{code,message,details}}` becomes `ApiError(status, code, message, details)`. The UI maps `code` to localized text (section 8).

### 3.3 Freshness model (polling)

| What | How often | Paused when hidden? |
|---|---|---|
| Overview (`GET /worlds/{wid}/overview`) and current city (`GET /cities/{cid}`) | Every **15 s** (`setInterval` in `GameProvider`), plus right after every command the UI performs (`refresh()`) and when any `Timer` reaches zero | **Yes.** `platform.onPause` sets `paused`. Pause comes from `visibilitychange` (all adapters) and from Yandex `game_api_pause`. Resume refreshes immediately and calls `gameplayStart()`. |
| Missions list | On mount, on each mission `Timer` completion, and after a recall | No interval |
| Reports, market, alliance, heroes, mail, territories | On mount, on tab or filter change, and after an action on that screen | No interval |
| Quests | On mount, and again whenever `city.serverTime` changes (that is, after every 15 s poll) | Follows the main poll |
| Map viewport | On mount and whenever the viewport origin changes; also after a mission is sent | No interval |

There is no WebSocket or push channel. An incoming attack becomes visible within about 15 s, through the bell badge (unread notifications) and the Missions screen.

---

## 4. Layout and responsive behaviour

### 4.1 Shell regions (`Shell.tsx`, `styles.css`)

```
Desktop (> 860 px)                          Mobile (≤ 720 px)
┌──────────────────────────────────────┐    ┌───────────────────────┐
│ topbar: title · city · resbar · bell │    │ title · city · bell   │
├────────┬─────────────────────────────┤    │ resbar (full width)   │
│sidenav │ main.content                │    ├───────────────────────┤
│ (all   │  tutorial bar               │    │ main.content          │
│ pack   │  screen                     │    │ (bottom padding 80 px │
│ screens│                             │    │  + safe-area)         │
│ +badges│                             │    ├───────────────────────┤
└────────┴─────────────────────────────┘    │ bottomnav: City Map   │
                                            │ Army Missions More    │
                                            └───────────────────────┘
```

| Breakpoint | Change |
|---|---|
| `> 860px` | Two-column layouts (`.grid-2`, `.map-layout`). Sidebar navigation is 200 px wide (`--nav-w`). |
| `≤ 860px` | `.grid-2` and `.map-layout` collapse to one column. The tile panel moves below the map. The sidebar stays. |
| `≤ 720px` | The sidebar is hidden. The fixed **bottom bar** shows 4 primary screens (`city, map, army, missions`, filtered by the pack's enabled screens) plus **More**. The top bar wraps, and the resource bar takes a full row (`order:3`). Content gets bottom padding of `80px + env(safe-area-inset-bottom)`. Toasts move above the bottom bar. Building cards go to 2 columns. Tables shrink to 0.82 rem. |

Other rules:
- **More sheet:** a bottom sheet with a 3-column grid of every remaining enabled screen, with badges. Tapping the backdrop closes it, and choosing an item navigates and closes it. The More button has `aria-expanded`, and a dot badge appears when `questsReady + unreadReports > 0`.
- **Safe areas:** `viewport-fit=cover`, and `env(safe-area-inset-bottom)` padding on the bottom bar and the sheet.
- **No horizontal overflow:** `e2e/play.mjs` asserts `scrollWidth - innerWidth ≤ 1` at 390×844. The map sizes cells from its container through `ResizeObserver` (cell = max(16, floor(min(width, 640)/15))) and never from the window.
- **Touch:** the map canvas has `touch-action:none`, with pointer-capture drag to pan.

### 4.2 Accessibility baseline

- A visible `:focus-visible` outline (2 px, accent colour) everywhere.
- Navigation: `<nav aria-label="Main">`, and the active item carries `aria-current="page"` (sidebar only; see gaps).
- The bell has `aria-label={t("notifications")}`. The resource bar has `aria-label="Resources"`, and each item has a `title` tooltip with the name, amount, cap and rate.
- Reduced motion: the Profile toggle sets `[data-motion="reduce"]`, which disables every transition and animation.
- Forms use `<label>` wrapping, `required`, `minLength` and `maxLength` that mirror the server rules, and `autoComplete` hints (`username`, `current-password`, `new-password`).
- Errors use `role="alert"`; loading and toasts use `aria-live="polite"`.

---

## 5. Navigation model

`Screen = city | economy | research | army | heroes | map | missions | reports | market | alliance | quests | mail | profile | settings` (`settings` renders `ProfileScreen`).
- `go(screen, focus?)` sets the screen, persists `rf.screen`, optionally sets `mapFocus` (used by "go to coordinates" links) and scrolls to the top.
- The sidebar lists every enabled screen except `settings`.
- Badges: **Reports** shows `unreadReports`, **Quests** shows `questsReady`, and **Mail** and the bell show `unreadNotifications`.

---

## 6. Global states and patterns

| State | Pattern |
|---|---|
| Loading (first load) | `<Loading/>`. While the shell has no city yet, the whole content area shows Loading. |
| Load error | `<ErrorBox error retry/>` with a localized message and a Retry button |
| Empty | `<Empty/>`, sometimes with a call to action (e.g. "Map →" on empty territories) |
| Disabled with a reason | The server's `CityView` gives each building, tech and unit a machine-readable `next.blocked` / `blocked` value: `resources`, `requirements`, `queue_full`, `race`, or `specialisation_taken`. The UI shows a tag on the card (`blockedLabel`), disables the button, sets the button `title` to the reason, lists unmet requirements (`Reqs`) and marks missing cost amounts in red (`Res .short`). |
| Pending | `Action` is disabled with `aria-busy` and a spinner. The auth and join forms disable their submit button while `busy`. |
| Success | An `ok` toast (e.g. "Lumberyard → 2"), then `refresh()` |
| Domain error | An error toast with localized text, or the server message as fallback. Nothing is retried. |
| Rate limited (429) | Retried twice automatically. If it still fails: "Too many actions — slow down a little". |

---

## 7. Screens

Each subsection lists the purpose, the user actions, the **exact API calls** (method and path, relative to `/api/v1`) and the screen's states. "Refresh" means `GET /worlds/{wid}/overview` plus `GET /cities/{cid}`.

### 7.1 Auth and guest (`AuthScreen`)

- **Purpose:** get a session without forcing registration. Guest play is mandatory on Yandex (YG-08).
- **Actions:**
  - Create an account, which is the default mode.
  - Sign in.
  - Switch between those two modes.
  - **Play as guest**.
- **API:**
  - `POST /auth/register` with `{login, password}`
  - `POST /auth/login` with `{login, password}`
  - `POST /auth/guest` with `{}`

  Each returns `{token, expiresAt, account}`, and the client stores the token.
- **Validation (mirrors the server):**
  - Login: 3–64 characters. The server lowercases it and requires `[a-z0-9_.@-]`.
  - Password: 8–128 characters.
- **States:**
  - Pending: `busy` disables both submit buttons.
  - Error: an inline `role="alert"` box, for example "invalid login or password" or `login_taken`. A banned account gets `403 banned` with an end date.
  - The auth endpoints are IP rate-limited, and a 429 shows the localized rate-limit text.
- **Responsive:** a single centred card on all widths.
- **A11y:** labelled inputs, `autoComplete` hints, and errors announced through `role="alert"`.
- **Not used yet:** `POST /auth/platform` (dev and Yandex identity) has no UI. `platform.identity()` returns `null` in every adapter.

### 7.2 World select and join (`WorldScreen`)

- **Purpose:** choose a world and, for a world the player has not joined, pick a ruler name and a race.
- **Actions:** Enter (if joined), Join (opens the join form), Cancel, Log out.
- **API:**
  - `GET /worlds` returns the list with `joined`, `players`, `speed`, `packId` and `packVersion`.
  - `GET /worlds/{wid}/pack` returns the full pack JSON. The server sends an ETag; the client does not send `If-None-Match` yet.
  - `POST /worlds/{wid}/join` with `{name, race}` returns `{playerId}`.
- **States:**
  - Loading: spinner.
  - Load error: an `error-box` with the raw message, which is not localized.
  - Join errors (localized when the code is known):
    - `name_taken` (409)
    - `invalid_request` for a name outside `^[\p{L}\p{N}_ .'-]{3,24}$`
    - `world_closed` (422)
  - Pending: Enter, Join and Submit are disabled while `busy`.
  - Join is disabled until a race is selected. The first race is pre-selected.
  - The default ruler name is the display name with disallowed characters removed, truncated to 24 characters.
- **Race picker:** a radio group (`fieldset`/`legend`) with the race colour as the border and the name colour, plus the race description.
- **Persistence:** entering a world writes `rf.world`.
- **Gap:** an empty world list renders nothing (there is no Empty state). "Log out" here clears the token **without calling `POST /auth/logout`**, so the server session stays valid until its TTL ends (see SECURITY.md, finding S-1).

### 7.3 Shell, navigation, resource bar and tutorial bar (`Shell`)

- **Top bar:**
  - The game title, which is `pack.meta.title` localized.
  - A city selector (`<select aria-label="City">`) when the player has more than one city, otherwise the city name.
  - The resource bar.
  - The bell (navigates to Mail) with an unread badge.
- **Resource bar:** one item per pack resource showing an icon, the amount (`fmt`: 12.3k, 2.5M) and the net rate per hour. Item states:
  - `.full`: amount ≥ storage cap.
  - `.neg`: net rate < 0.
  - The last item is population and housing.
  - All numbers come from the last `GET /cities/{cid}`. They do **not** tick between polls; see gaps.
- **Sidebar and bottom bar:** see section 4.1.
- **Tutorial bar:** shown when overview and city are loaded. It can be dismissed with ✕, but only for the current mount (not persisted). It contains:
  - The newbie-protection pill with a `Timer`, while `player.protectedUntil` is in the future.
  - A "Quests: N" pill when the next tutorial quest exists and `questsReady > 0`. The pill navigates to Quests.
- **Content:** before the first city loads, it shows `Loading`, or `ErrorBox` with Retry if the refresh failed.

### 7.4 City (`CityScreen`, `QueuePanel`)

- **Purpose:** see every building grouped by pack category, and the build, research and train queues.
- **Actions:** open a building card (opens the building modal); cancel the **last** job in a queue.
- **API:**
  - Data comes from the shared refresh, `GET /cities/{cid}`.
  - `POST /cities/{cid}/jobs/cancel` with `{jobId}` returns `{refund}`.
- **Building cards** (`<button class="bcard">`) show:
  - The name.
  - "Level N" and "→ target" when an upgrade is queued.
  - A state tag: Build/Upgrade when available, or the blocked reason.
  - `.unbuilt` styling at level 0.
- **Queue panel:**
  - Each job shows its item and target level (or × quantity).
  - The **running** job (`startsAt ≤ now`) shows a `Timer` that triggers a refresh when it ends. Queued jobs show "Queue".
  - Only the last job has Cancel. Cancel needs **confirmation** ("Cancel?") because the refund is partial (80 % in the test pack). The server answers `not_last_in_queue` for any other job.
- **States:** an empty queue shows "Queue is empty". The layout is two columns on desktop: buildings, then an aside with three queues.
- **Mobile:** building cards in 2 columns; the queues stack below.

### 7.5 Building modal (`BuildingModal`)

- **Content:**
  - Description, current level and max level.
  - Effects at the current level (`effectText`): production per hour with workers, storage, hidden (protected) amount, housing, idle tax, modifiers.
  - Next level: cost (red where the player is short), duration and unmet requirements.
- **Action:** `Build → N` / `Upgrade → N`, which sends `POST /cities/{cid}/build` with `{building}`. It is disabled when `next.blocked` is set. Success shows a toast and refreshes; the modal stays open and re-renders from the refreshed city.
- **Max level:** "Max level" text instead of the action.
- **A11y:** follows the `Modal` contract (focus on open, Escape, focus restore). The e2e test uses `getByRole("dialog")`.
- **Errors:**
  - `insufficient_resources` (422) with `details.missing`
  - `requirements_not_met`
  - `queue_full`
  - `rate_limited`

### 7.6 Economy (`EconomyScreen`)

- **Purpose:** explain the city economy: gross production, consumption, net, storage cap and protected amount for each resource. Also population, workers needed, employment %, idle inhabitants, upkeep, and owned territories.
- **Actions:**
  - **Collect** territory yield.
  - **Abandon** a territory, which needs **confirmation**.
  - Go to the map, from the empty state.
- **API:**
  - `GET /worlds/{wid}/territories`
  - `POST /cities/{cid}/collect` with `{x, y}`
  - `POST /worlds/{wid}/territories/abandon` with `{x, y}`
- **States:**
  - Each resource row has a storage progress bar.
  - Warnings: "Employment < 100 %" and "Net < 0".
  - Territories: Loading, ErrorBox with Retry, Empty with a "Map →" link.
  - Collect is disabled when nothing is pending.
- **Mobile:** the table shrinks; one column.

### 7.7 Research (`ResearchScreen`)

- **Purpose:** show the tech tree by branch.
- **Card content:** level and max, target level, description, exclusive-group tag (★ group, for specialisation), effects, cost, duration and unmet requirements.
- **Action:** `Research → N`, which sends `POST /cities/{cid}/research` with `{tech}`. It is disabled with a `title` when blocked. The `specialisation_taken` reason shows as "Locked".
- **Queue:** the research `QueuePanel` sits at the top.

### 7.8 Army (`ArmyScreen`)

- **Purpose:** train units and see the home and away garrisons, plus foreign troops stationed in the city.
- **Unit card content:**
  - Class and description.
  - Stats with icons: attack, defence %, HP, speed, carry, scout.
  - Class bonuses and upkeep per hour.
  - Home and away counts.
  - Cost per unit, time per unit and unmet requirements.
- **Actions:**
  - A quantity field (1–100 000).
  - A **Max (n)** button that sets the affordable maximum, computed client-side from current stock.
  - **Train**, which sends `POST /cities/{cid}/train` with `{unit, quantity}`. It is disabled when the unit is not trainable or the player cannot afford the quantity.
- **States:** "Foreign troops" shows Empty or a list of `player: units`.

### 7.9 Heroes (`HeroesScreen`)

- **Purpose:** hire heroes, spend skill points, equip items.
- **API:**
  - `GET /worlds/{wid}/heroes` returns `{heroes, inventory, maxHeroes, hireCost}`.
  - `POST /cities/{cid}/heroes` with `{class, name}` (hire).
  - `POST /worlds/{wid}/heroes/{hid}/skills` with `{skill}`.
  - `POST /worlds/{wid}/heroes/{hid}/equip` with `{itemId}`.
  - `POST /worlds/{wid}/heroes/{hid}/unequip` with `{slot}`.
- **States:**
  - Hire is disabled at `heroes.length ≥ maxHeroes` or when the name is shorter than 3 characters.
  - Skill buttons are disabled without skill points or at max skill level.
  - Equip and unequip are disabled while the hero is on a mission. The server also returns `hero_busy`.
  - A wounded hero shows a danger tag with a `Timer`.
  - Each hero has an XP progress bar.
- **Gap:** the equip `<select>` calls `command` directly, without `Action`. It has no busy state, and its errors toast the raw message rather than the localized one.

### 7.10 Map, tile panel and mission dialog (`MapScreen`, `TilePanel`, `MissionDialog`)

- **Purpose:** explore the world grid, inspect tiles and dispatch missions.

**Map:**
- **Data:** `GET /worlds/{wid}/map?x&y&w=15&h=15`. The server allows viewports up to 64×64.
- **Rendering:** a flat 15×15 `<canvas>` drawn in code with no image assets:
  - Terrain colour from the pack.
  - Owned territory: a dashed outline in the race colour, or the accent colour for your own.
  - Cities: a house glyph in the race colour or accent.
  - Neutral lairs: a dark disc with the level number, at 35 % alpha when cleared.
  - Impassable tiles: darkened.
  - The selected tile: a 2 px outline.
- **Header:** the viewport coordinates.
- **Interaction:**
  - Pointer-drag pans by whole cells (pointer capture).
  - A click without movement selects a tile.
  - **Keyboard:** the canvas is focusable (`tabIndex=0`) with `aria-label="Map — arrow keys to pan, click a tile"`. Arrow keys pan 3 tiles (`preventDefault` stops page scroll).
  - Buttons: "Your city" (re-centre), and ◀ ▲ ▼ ▶ to pan 5 tiles, each with an `aria-label`.
  - `mapFocus` from other screens (e.g. mission coordinates) sets the initial centre.
- **States:** an `ErrorBox` with Retry above the map on a load error.

**Tile panel:**
- **Content:** coordinates, terrain (and its yield per hour), city name, owner [alliance] · race. For a lair: name, level, guards, rewards and "cleared".
- **Action set by tile type:**

  | Tile | Actions offered |
  |---|---|
  | Own capital | none |
  | Any city | scout, transport, then reinforce (own city) or attack plus reinforce (other city) |
  | Active lair | attack, scout |
  | Territory owned by another player | attack, scout |
  | Free land | occupy, settle, scout |
  | Impassable tile | none |

- **Gap:** the action set is derived client-side. The server still validates every request.

**Mission dialog** (wide `Modal`):
- **Unit picker:** a row per available home unit with a count input (clamped 0..home) and a Max button. Scout missions list only scout-class units.
- **Cargo** (transport and settle): one input per non-premium resource, clamped to stock, with a header showing the total and carry capacity.
- **Hero** (attack, settle, reinforce): a `<select>` of idle, unwounded heroes from this city. Data comes from `GET /worlds/{wid}/heroes`.
- **Attack options:** a Capture checkbox (enemy territory only) and "Retreat at %" (1–100).
- **Preview:** ETA computed client-side as `max(minTravelSeconds, distance/slowest × 3600)/worldSpeed`, plus total carry. This is an estimate; the server-returned `arriveAt` is authoritative.
- **Send:** `POST /cities/{cid}/missions` with `{kind, x, y, units, heroId?, cargo?, capture?, retreatPct?}`.
  - Disabled with no units, or when cargo exceeds capacity.
  - **Attack requires confirmation** ("Attack this target? Troops may be lost."), and the Send button uses the danger style.
  - On success: a toast, a refresh, the dialog closes and the map reloads.
- **Server errors:**
  - `target_protected`
  - `friendly_target`
  - `target_too_weak`
  - `attack_limit`
  - `over_capacity`
  - `city_limit`
  - `territory_limit`
  - `own_target`
  - `invalid_request`

### 7.11 Missions (`MissionsScreen`)

- **Purpose:** list own missions (outbound, stationed, returning) and **incoming** hostile missions.
- **API:**
  - `GET /worlds/{wid}/missions`
  - `POST /worlds/{wid}/missions/{mid}/recall` with `{}`
- **Rows:**
  - An icon: skull for incoming, eye for scout, sword for attack, flag for others.
  - Kind, then target coordinates as a link that opens the map focused there.
  - For incoming missions: an "Incoming · owner" danger tag.
  - Units, shown as "?" for incoming. The server hides the attacker's composition.
  - The status timer: "arrives in" or "returns in", or "stationed".
  - Cargo.
- **Recall:** offered for own outbound or stationed missions. It has **no confirmation**.
- **Refresh:** when a timer ends, the list reloads and the shared refresh runs.
- **States:** Loading, ErrorBox with Retry, Empty.

### 7.12 Reports and report modal (`ReportsScreen`, `ReportModal`)

- **API:**
  - `GET /worlds/{wid}/reports?limit=50` (server maximum 100; `before` cursor supported but not used by the UI).
  - `GET /worlds/{wid}/reports/{rid}`. **Opening a report marks it read on the server.**
- **List:**
  - Battle or scout icon, then target, attacker and defender (or the neutral).
  - A Victory or Defeat tag computed from the viewer's side.
  - Local date and time.
  - `.unread` styling.
- **Battle report modal:**
  - Attacker vs defender, or neutral with level.
  - A table of start, losses and survivors per side.
  - Rounds, retreat flag, wall bonus.
  - Loot, rewards, item drops, territory captured.
- **Scout report modal:** success or failure, "detected", then resources, garrison, buildings, neutral guards.
- **On close:** the list reloads and the shared refresh runs, which updates the badge.
- **Gap:** the words "retreat" and "detected" are hard-coded English.

### 7.13 Quests and tutorial (`QuestsScreen`)

- **API:**
  - `GET /worlds/{wid}/quests`
  - `POST /worlds/{wid}/quests/{qid}/claim` with `{cityId}`
- **Tabs:** In progress (`active` and `complete`) and Claimed. Locked quests are hidden.
- **Rows:** name, a Tutorial tag, description, a 💡 hint (active only), `progress/goal`, rewards.
- **Actions:**
  - **Claim reward** for completed quests. The server enforces exactly-once claims: a second claim returns `already_claimed` (409).
  - **Go to** the quest's target screen for active quests.
- **Refresh:** reloads after a claim and on every 15 s poll.

### 7.14 Market (`MarketScreen`)

- **API:**
  - `GET /worlds/{wid}/market[?sell=<res>]`
  - `POST /cities/{cid}/market/orders` with `{sell, sellAmount, buy, buyAmount}`
  - `POST /cities/{cid}/market/accept` with `{orderId}`
  - `POST /worlds/{wid}/market/{oid}/cancel`
  - `POST /cities/{cid}/exchange` with `{from, to, amount}`
- **Place order:**
  - The sell and buy selects list only tradable, non-premium resources.
  - Disabled when sell = buy or when stock < sell amount.
  - Server errors: `price_out_of_band` (outside 1:10..10:1 of the reference value), `too_many_orders`, `insufficient_resources`.
- **NPC exchange:**
  - A live preview of what the player receives: `floor(amount × rate_from/rate_to × (1 − fee))`.
  - Disabled when the output would be < 1 or stock is short.
- **Order book:**
  - Filter by the sell resource.
  - Rows: sell, buy (red if the player cannot afford it), player.
  - Actions: **Accept** (disabled if unaffordable) or **Cancel** (own orders).
  - A concurrent second accept returns `order_closed` (409).
- **States:** Loading, ErrorBox, Empty.
- **Gap:** accepting or cancelling an order has no confirmation.

### 7.15 Alliance and rankings (`AllianceScreen`)

- **Tabs:** Alliance (mine), Relations (all alliances), Rankings.
- **API:**
  - `GET /worlds/{wid}/alliances/{aid}`
  - `GET /worlds/{wid}/alliances`
  - `GET /worlds/{wid}/alliance-invites`
  - `GET /worlds/{wid}/rankings` (default 50, max 100, `offset` supported)
  - `POST /worlds/{wid}/alliances` with `{name, tag, description}`
  - `POST /worlds/{wid}/alliance/invites` with `{player}`
  - `POST /worlds/{wid}/alliance-invites/{iid}/respond` with `{accept}`
  - `POST /worlds/{wid}/alliance/members/{pid}/role` with `{role: member|officer|leader|kick}`
  - `POST /worlds/{wid}/alliance/relations` with `{allianceId, status: ally|nap|war|neutral}`
  - `POST /worlds/{wid}/alliance/leave`
- **Without an alliance:** a pending-invite strip (Accept / Decline) and a create form. The name must be ≥ 3 characters and the tag 2–5. The form shows the creation cost.
- **With an alliance:**
  - A member table with role-gated actions:
    - The leader can promote or demote, **make another member leader (confirm)** and kick (confirm).
    - An officer can kick members (confirm).
  - An invite field for the leader and officers.
  - The relations list.
  - **Leave (confirm)**. When the leader is the only member, leaving disbands the alliance.
- **Relations tab:** a list of alliances. The leader sees ally / nap / war / neutral buttons per alliance.
- **Rankings tab:** rank, name, race, alliance, score. The viewer's row is highlighted with `.me`.
- **Gap:** declaring **war** has no confirmation. The Relations and Rankings tabs have no pagination UI.

### 7.16 Mail, notifications and compose (`MailScreen`, `Compose`)

- **Tabs:** Notifications, Inbox, Sent, Alliance chat. A **Compose** button is always visible.
- **API:**
  - `GET /worlds/{wid}/notifications` returns `{items, unread}` (max 100).
  - `POST /worlds/{wid}/notifications/read` with `{upTo}` ("Mark all read").
  - `GET /worlds/{wid}/messages?box=inbox|sent|alliance`
  - `POST /worlds/{wid}/messages` with `{to, alliance, subject, body}`
  - `POST /worlds/{wid}/messages/{id}/read`
  - `POST /worlds/{wid}/messages/{id}/report`
  - `POST /worlds/{wid}/blocks` with `{playerId, block}`
- **Notifications:** human-readable lines for `build_done`, `research_done`, `train_done`, `report`, `incoming_attack`, `message` and `mission_returned`. Any other kind shows its raw name with `_` replaced by a space. `.unread` styling.
- **Messages:**
  - Each message is a `<details>` element (keyboard-operable natively).
  - Opening an unread inbox message marks it read (fire-and-forget).
  - Inside: the body as React text, never HTML. Then **Reply ↩**, **Report abuse** (toast confirms) and **Block** (needs **confirmation**). Block also hides the sender's messages from the inbox.
- **Compose modal:**
  - Fields: To (hidden for alliance chat), Subject (≤ 120), Message (≤ 4000).
  - Send is disabled with an empty body, or with a recipient shorter than 3 characters.
  - Server errors: `rate_limited` (more than 60 messages per hour), `forbidden` ("this player does not accept your messages"), `not_found`.
- **Store requirement:** report and block are the minimum Google Play requires for UGC (GP-16).
- **Gap:** there is no profanity filter, no "unblock" UI and no list of blocked players.

### 7.17 Profile, settings and account deletion (`ProfileScreen`)

- **Profile panel:**
  - Ruler name, race, score, alliance tag, cities (n/cap), territories (n/cap).
  - The protection timer, using server time.
  - Server time in UTC.
  - World and pack version.
  - **Choose a world**, which leaves to the world screen.
- **Settings:**
  - **Language**: options are the intersection of the pack locales and `["en","ru"]`. Changing it calls `PUT /me/locale` with `{locale}` and stores `rf.locale`.
  - **Reduce animations**: a checkbox stored in `rf.motion`.
- **Account:**
  - Shows the login or the display name.
  - **Guest upgrade:** a warning text, then login and password fields. **Save account** sends `POST /auth/upgrade` with `{login, password}`. The server keeps the account id and all progress. Errors: `login_taken`, `not_guest`.
  - **Log out:** `POST /auth/logout`, then the token is cleared. A logout failure is ignored.
  - **Delete account:** a danger `Action` with **confirmation** ("Delete your account permanently? This cannot be undone."), then `DELETE /me`, then the token is cleared. The server anonymizes the account immediately; see SECURITY.md §8. This meets the in-app deletion requirement of GP-10. The web-link deletion route does not exist yet.
- **Gaps:**
  - After a successful upgrade, the in-memory `account` still says guest, so the form stays visible until reload.
  - There is no typed-phrase confirmation and no grace period, both of which PLATFORM_REQUIREMENTS §3.5 recommends.
  - There is no password change and no "log out everywhere".

---

## 8. Error code to UX mapping

The codes below are localized in `i18n/ui.ts` (`en`, `ru`). Every other code falls back to the server's English `message`.

| Code | HTTP | UX |
|---|---|---|
| `insufficient_resources` | 422 | Error toast. Missing amounts are already red in the cost list. |
| `requirements_not_met` | 422 | Error toast. Unmet requirements are listed. |
| `queue_full` | 422 | Error toast. The card shows a "Queue" tag. |
| `target_protected` | 422 | Error toast in the mission dialog |
| `friendly_target` | 422 | Error toast |
| `already_claimed` | 409 | Error toast |
| `rate_limited` | 429 | Automatic retry ×2, then a toast |
| `network` | — (client) | Automatic retry ×2, then a toast or ErrorBox with Retry |
| `unauthenticated` | 401 | Token cleared, back to the auth screen |
| `banned` | 403 | Auth error box ("account suspended until …") |
| `idempotency_key_reused` | 422 | Should never surface: keys are unique per action |
| Others, e.g. `order_closed`, `price_out_of_band`, `attack_limit`, `not_last_in_queue`, `hero_busy` | 4xx | Server message in English. **Gap:** not localized. |

---

## 9. Current UX gaps (honest list)

**Game presentation**
1. **No fog of war.** The map shows every city, owner, alliance tag, race and lair level in any 15×15 viewport of a joined world. Scouting is only needed for garrison and resource details. Incoming mission composition is hidden.
2. **Polling, not push.** State refreshes every 15 s (paused while hidden). Incoming attacks, filled market orders and messages appear with up to about 15 s delay. Without a WebSocket, there are no "attack incoming" alerts while the tab is hidden, and no OS push notifications on Android.
3. **The map is a flat canvas with original vector glyphs.** There is no isometric view, no art, no animation, no minimap and no zoom (only pan); movement and territory borders are not drawn. The 15×15 viewport is fixed.
4. **No art assets yet.** Every icon is an original SVG glyph drawn in code (`ui/kit.tsx`). Buildings, units, heroes and races have no illustrations. The `/packs/{pack}/assets/...` route exists, but packs ship no assets.
5. **English and Russian only.** The engine UI strings exist only in `en` and `ru`. Only 8 error codes are localized; all other server messages show in English. A few labels are hard-coded English: `aria-label`s "City", "Close", "Class", "Filter", "Stats", "Pan left/up/down/right", and the report words "retreat" and "detected". Some building category headings are pack ids with a capital first letter.
6. **Resource counters do not tick** between polls. The bar updates only every 15 s or after an action.

**Accessibility**

7. **No keyboard tile selection.** Arrow keys pan the map, but selecting a tile (and so sending a mission from the map) needs a pointer. Fix: a keyboard cursor with Enter to select, plus a coordinate entry field.
8. **Modals do not trap focus.** `Tab` can leave the dialog. The mobile **More sheet** has no `role="dialog"`, no Escape handling and no focus management.
9. **Partial tablist pattern.** `Tabs` lacks arrow-key navigation and `role="tabpanel"` linkage. The bottom bar does not set `aria-current`.
10. **Motion and contrast.** `prefers-reduced-motion` is not honoured automatically; only the manual toggle works. No contrast audit has been run on pack palettes, and a pack palette can produce low-contrast text.

**Timers and confirmations**

11. **Some timers use the device clock.** `QueuePanel` (running vs queued), the tutorial-bar protection check, the hero "wounded" check and the mission-dialog idle-hero filter use `Date.now()` instead of `api.now()`. A device whose clock is skewed by more than about 1 s can mislabel a job as queued or show a stale tag.
12. **Missing confirmations:**
    - Mission recall.
    - Market accept and cancel.
    - Declaring war or other relation changes.
    - Declining an invite.

    Account deletion has a confirm modal, but no typed phrase and no grace period.
13. **Account handling:**
    - Logout on the world screen does not revoke the server session.
    - After a guest upgrade, the in-memory account is not refreshed.
    - There is no password change, no session list and no "log out everywhere".

**Missing screens and states**

14. **No admin UI.** The admin API exists: inspect, grant, ban, failed events, audit log, reported messages and pack migration.
15. **No city rename UI** (the API exists: `POST /cities/{cid}/rename`). No unblock or blocked-player list. No alliance description editing.
16. **No dedicated offline or connection-error screen.** Yandex YG-15 requires a retry UI for connection errors. Today failures surface as ErrorBox and Retry inside a screen, or as toasts. The boot sequence falls back to the auth screen on network failure, which can mislead a logged-in user.
17. **Empty world list** renders nothing.
18. **Yandex and Android integration is thin.** `gameplayStart`/`gameplayStop` are called on pause and resume only, not when menus open (YG-03 timing if used). There is no Android back-button handling to close dialogs (GP-06). There is no in-game benefits prompt for Yandex auth (YG-08).
19. **External fonts.** Pack fonts load at runtime from `fonts.googleapis.com`, which conflicts with "all assets in the archive" (YG-15) and with offline use. The fonts must be bundled.

---

## 10. UI acceptance checklist (release gate)

| # | Check | How verified today |
|---|---|---|
| U1 | Guest, then join, then build, reload, and the level persists | `e2e/play.mjs` (desktop 1360×860) |
| U2 | Every enabled screen renders an `h1` without console errors | `e2e/play.mjs`, which loops over 10 screens and fails on any `pageerror` or console error |
| U3 | Mobile 390×844: bottom bar, More sheet, no horizontal overflow | `e2e/play.mjs` |
| U4 | Tutorial quest claim through the UI | `e2e/play.mjs` (conditional on a claimable quest) |
| U5 | Retried command reuses its Idempotency-Key; 401 clears the token; domain errors are not retried | `client/src/__tests__/client.test.ts` |
| U6 | Keyboard-only: complete one mission dispatch | **Not met** (gap 7) |
| U7 | Screen reader: modal title announced; focus restored after close | Manual, not automated |
| U8 | Both locales: no untranslated engine strings on the main path | Manual. Partially met (gap 5). |
| U9 | Yandex draft mode: `LoadingAPI.ready` timing; pause on tab switch stops polling | Manual, not done |
| U10 | Android back button closes a modal | **Not met** (gap 18) |
