# Platform Requirements: Yandex Games and Google Play

Owner: Agent G (Release / Platform Engineer)
Status: DRAFT v1. Research snapshot, not legal advice.
Verification date (access date for all sources): **2026-10-31**

---

## 1. Summary and verification date

### 1.1 What this document is

This document lists the current publishing requirements for:

- **Yandex Games**, as an HTML5 build of our React/TypeScript/Vite client.
- **Google Play**, as an Android app built from the same client with Capacitor.

For each requirement it records how our server-authoritative MMO (Go backend) will meet it, or where the gap is. It also proposes one `PlatformAdapter` interface so that game code stays platform-agnostic.

> **Technical compatibility is not the same as passing moderation.** Every item below is a necessary condition, not a sufficient one. Yandex moderators and Google Play reviewers judge content, UX and policy fit by hand. A build that meets every technical item can still be rejected. For example, Yandex can reject a duplicate game or a non-working loading screen, and Google can reject an app it treats as "webview spam".

### 1.2 Evidence levels (read this first)

Several official hosts **could not be fetched directly from this research session**: yandex.ru, yandex.com, support.google.com, play.google.com and capacitorjs.com were blocked by the session's network policy. To verify those items properly, the owner or the next agent needs a session with these hosts allowed. Each requirement carries one evidence tag:

| Tag | Meaning |
|---|---|
| **[F]** | Fetched and read directly from the official source on 2026-10-31. This covers developer.android.com pages, the Capacitor npm packages (registry.npmjs.org tarballs) and the Capacitor docs Markdown source (raw.githubusercontent.com/ionic-team/capacitor-docs). |
| **[S]** | Official page found through a web search. Only search-engine excerpts or summaries of the page were seen; the page itself was not opened. The content is probably right, but wording, numbering and numbers **must be re-checked** on the live page before submission. |
| **[3P]** | Third-party source only (blogs, vendor guides). Lower confidence. |
| **UNVERIFIED** | No official confirmation found. Do not rely on it until it has been checked. |

### 1.3 Key findings (TL;DR)

**Yandex Games**
- **Guest play is mandatory.** The game must be playable without registration. Yandex ID is the only login allowed. Our server must create a guest account keyed to the Yandex player ID and keep guest progress. (YG-08)
- **Our backend must be approved in the CSP tab.** Connections to it are https/wss only and must be justified. The alternative for complex multiplayer is the **iframe route**, which needs a separate approval through support. (YG-15, YG-16)
- **Pause rules.** Sound and gameplay must stop on focus loss, within about 2 s, and during fullscreen ads. Because the server keeps simulating the world, a paused client must not be punished. (YG-04, YG-05)
- **Load timing.** `LoadingAPI.ready()` must be called exactly when the game becomes playable. The `GameplayAPI.start/stop` markers are optional but must be correct if used. (YG-02, YG-03)
- **Ads and payments.** Both go only through the SDK. Purchases must be enabled by Yandex first. Yandex search excerpts say **cloud saves are required for games with in-app purchases**; this needs reconciling with our server-side saves. (YG-06, YG-09, YG-10)
- **Build format.** Upload a ZIP of at most 100 MB uncompressed, with `index.html` at the root and the language detected automatically from `ysdk.environment.i18n.lang`. (YG-13, YG-14)

**Google Play**
- **Target and signing.** Target **API 36** for new apps and updates (required since 2026-08-31). Ship an **AAB** and enroll in **Play App Signing**. (GP-01 to GP-03)
- **Billing.** Use **Google Play Billing** for gems and other digital items, outside the US-specific programs. The Play Billing Library must be version 8 or later. (GP-04, GP-11)
- **Disclosures.** The Data safety form and a privacy policy are required. Account deletion must be available both in the app and through a web link. In-game chat must offer report and block. (GP-08 to GP-10, GP-16)
- **New personal developer accounts** have to run a closed test with 12 testers for 14 consecutive days before production access. This is widely reported but was not confirmed on an official page in this pass. (GP-13)
- **Minimum functionality / webview spam.** The game must bundle its assets and behave like a real app, not a website wrapper. (GP-12)
- **Capacitor 8.6.0** (latest stable, 2026-10-28) defaults to minSdk 24 and compile/target SDK 36, needs Node 22 or later and Android Studio 2025.2.1 or later, and builds an AAB by default. (Section 3.2)

---

## 2. Yandex Games

### 2.1 Requirement list

Requirement numbers in brackets, such as "1.19", follow the Yandex requirement IDs seen in search excerpts. Yandex renumbers items between page versions; excerpts showed versions dated between 2025-12-16 and 2026-10-09. **Re-map all IDs against the live page.**

| ID | Requirement | Source (accessed 2026-10-31) | Our project: how we satisfy it / gap | Verification method |
|---|---|---|---|---|
| YG-01 | The Yandex Games SDK must be connected and initialized. This is "a prerequisite for successful moderation". Load it with `<script src="/sdk.js">` for archive builds (`https://sdk.games.s3.yandex.net/sdk.js` if self-hosted), then `await YaGames.init()`. | [S] https://yandex.com/dev/games/doc/en/sdk/sdk-about ; https://yandex.com/dev/games/doc/en/requirements/1/19 | Add a `YandexAdapter.init()` that loads `/sdk.js` and awaits `YaGames.init()` before any other SDK call. **Gap:** not implemented. | Run in draft mode. The console must show "Yandex SDK initialized" [S]. Run the debug panel. |
| YG-02 | Call `LoadingAPI.ready()` "when the user can start playing", and not on a fixed timer. Moderators check that the Game Ready indicator matches actual readiness. | [S] https://yandex.com/dev/games/doc/en/requirements/1/19 ; https://yandex.com/dev/games/doc/en/concepts/moderation | Call `signalReady()` when the main menu or lobby is interactive (assets loaded, UI responsive). Keep it separate from the world connection, which may fail; see YG-15. **Gap:** define the "ready" event in the client boot sequence. | Check the Game Ready indicator in the debug panel. Test manually on a slow-network profile. |
| YG-03 | `GameplayAPI.start()` / `GameplayAPI.stop()` are **optional**, but if used, timing must exactly match the "Gameplay" section. Call `start` when gameplay starts or resumes and `stop` on pause or end. Moderators check level start and end, menu and shop opening, ads, and focus loss and return. | [S] https://yandex.com/dev/games/doc/en/sdk/sdk-game-events ; https://yandex.com/dev/games/doc/en/requirements/1/19 | Decision needed. Recommendation: **implement** the markers. Call `start` when entering world or battle view and `stop` when opening shop, menus or ads, and on `game_api_pause`. Tie this to the client UI state machine. | Scripted manual test of each scenario listed in the excerpt. |
| YG-04 | When the game loses focus, **sound stops**. Moderators test minimizing the window, switching tabs and the tab-selection menu. Up to about 2 s of delay is tolerated. On iOS, sound may keep playing in the tab menu. The SDK emits `game_api_pause` / `game_api_resume` for ads, purchase windows, tab switches and minimize. The pause handler should mute sound and pause gameplay. | [S] https://yandex.com/dev/games/doc/en/requirements/1/3 ; https://yandex.com/dev/games/doc/en/sdk/sdk-events (requirements 1.3, 4.7) | The adapter subscribes with `ysdk.on('game_api_pause' / 'game_api_resume')` and also listens to `visibilitychange` as a fallback. Pausing mutes the audio bus and freezes client input and animation. **MMO gap:** the server world keeps running; see the risk section (2.3, R-3). | Desktop Chrome and Yandex Browser, and mobile Android and iOS: minimize, switch tab, open the tab switcher. |
| YG-05 | During fullscreen ads (interstitial or rewarded), "game sound and gameplay must be paused". If a user clicks an ad and comes back, progress is kept. | [S] https://yandex.com/dev/games/doc/en/concepts/requirements ; https://yandex.com/dev/games/doc/en/sdk/sdk-adv | The ad callbacks `onOpen` and `onClose` call the same pause and resume path as YG-04. Progress is server-authoritative, so an ad click cannot lose it. | Debug-panel ad test, plus an ad-click test. |
| YG-06 | Ads are shown **only through the Yandex SDK**. No third-party ads, including static images and text. Yandex controls how often interstitials appear. Rewarded ads have no frequency cap. Do not show ads during active gameplay, because accidental clicks count as ad fraud; use natural breaks. | [S] https://yandex.com/dev/games/doc/en/concepts/requirements ; https://yandex.com/dev/games/doc/en/sdk/sdk-adv ; https://yandex.com/dev/games/doc/en/requirements/4/4 (wording UNVERIFIED) | Show interstitials only at safe breaks, for example after a battle summary or when returning to the city screen. Rewarded ads are opt-in buttons, and the **server grants the reward** after the client reports `onRewarded`. This is still forgeable, so cap the value and rate-limit it server-side. **No other ad network in the Yandex build.** | Code review: no other ad SDK in the Yandex bundle. Manual check of ad placement. |
| YG-07 | Progress is saved immediately after a player action or a save-button press. Refreshing the page or rotating the device on mobile must not lose progress. | [S] https://yandex.com/dev/games/doc/en/requirements/1/9 | Satisfied by design: the server is authoritative and every action is persisted server-side. Pending client-side UI state, such as unsent build queues, should go to the server immediately. | Refresh mid-action. Rotate on mobile. Reconnect. |
| YG-08 | **No mandatory registration.** The game must start and be playable without registration or third-party login. Only Yandex ID login is allowed, through the SDK. Guest play must keep progress. Any authorization prompt must explain its benefits and let the player decline and keep playing. Gating entry or play behind login is "Incorrect". | [S] https://yandex.com/dev/games/doc/en/requirements/1/2 ; https://yandex.com/dev/games/doc/en/sdk/sdk-player | **Major design impact.** On Yandex, there is no email/password or Google login. The server creates or loads an account keyed to `player.getUniqueID()` (available for all players) and issues our own session token. Login is offered through `ysdk.auth.openAuthDialog()` with a benefit text and a "Not now" option. **Gap:** server-side Yandex identity binding and guest-to-authorized upgrade. See R-1 and R-2. | Play from a fresh browser profile without logging in. Confirm progress survives a refresh. Decline the auth prompt. |
| YG-09 | Cloud saves (`player.setData` / `getData`) are optional in general ("in the cloud or locally"). If used, enable "game uses cloud save" in the draft. Limits: 200 KB per player; `setData` rate limit of 100 requests per 5 minutes. Excerpts state cloud saves "are required for games with in-app purchases" (item 1.13.3); **exact wording UNVERIFIED**. | [S] https://yandex.com/dev/games/doc/en/requirements/1/9 ; https://yandex.com/dev/games/doc/en/sdk/sdk-player | Our source of truth is our own Go server. **Gap and risk:** if 1.13.3 truly requires SDK cloud saves for IAP games, mirror a small binding record via `setData`, such as `{serverAccountId, lastSyncAt}`. Ask Yandex support whether server-side persistence satisfies 1.13.3. See R-4. | Read live requirement 1.13.3. Get written confirmation from Yandex support. |
| YG-10 | In-app purchases go **only through the SDK** (`ysdk.payments`). Purchases must be enabled by Yandex: email games-partners@yandex-team.com, wait for confirmation, then add products and confirm "Purchases are enabled" in Settings. Implement consumption (`consume`). Check unprocessed purchases (`getPurchases`) on startup. Breaking these rules "may result in the disabling of in-app purchases … or in the app's depublishing". `signed: true` returns purchase data only in a `signature` field for server-side validation. | [S] https://yandex.com/dev/games/doc/en/sdk/sdk-purchases ; https://yandex.com/dev/games/doc/en/requirements/1/12 ; [3P] defold-yagames README | Flow: `purchase({id, developerPayload: serverOrderId})` with `signed:true`, then send the signature to our server. The server verifies it with the Yandex secret key, grants the item idempotently by purchase token, returns OK, and only then does the client call `consume`. On boot, call `getPurchases()` and resend any unconsumed ones. **Gap:** the signature algorithm and key handling are **UNVERIFIED** (on the `#signature` section, not retrieved). | Sandbox or test purchases in draft. Kill the tab mid-purchase and confirm the grant on next launch. |
| YG-11 | Correct display during window resize and screen rotation. No disproportionate deformation. If only one orientation is supported, select it in the draft; Yandex then shows a rotate prompt. The active field must not be cut off. | [S] https://yandex.com/dev/games/doc/en/requirements/1/10 ; https://yandex.com/dev/games/doc/en/release-notes | Responsive React layout, with a canvas scaled to the viewport and safe-area aware. Decide the supported orientations (recommendation: landscape plus portrait on mobile, or declare landscape-only). | Device matrix test with resize and rotate. |
| YG-12 | No browser scrollbar and no swipe-to-refresh (1.10.2). On mobile, a long tap must not select an area or open the context menu. On desktop, interacting with the game field must not highlight it or open the context menu. | [S] https://yandex.com/dev/games/doc/en/concepts/requirements ; https://yandex.com/dev/games/doc/en/release-notes | CSS: `overscroll-behavior: none`, `user-select: none`, `-webkit-touch-callout: none`, `touch-action` on the game root, `contextmenu` preventDefault on the game field, `html,body{overflow:hidden}`. Keep text selection inside chat input fields only. | Mobile Chrome, Safari and Yandex Browser: long-press and pull-down tests. |
| YG-13 | **Automatic language detection** through `ysdk.environment.i18n.lang` at launch, with a fallback language set. This applies even to single-language games. Every language declared in "Game translated into" is tested. Store texts and media must be in the declared language. | [S] https://yandex.com/dev/games/doc/en/requirements/2/14 ; https://yandex.com/dev/games/doc/en/concepts/languages-and-domains ; https://yandex.com/dev/games/doc/en/console/add-new-game | `getLocale()` in the adapter returns `i18n.lang`, mapped to our supported locales with fallbacks (for example be/kk/uk → ru, others → en). Server-sent strings must be keyed so they are localizable. **Gap:** decide the launch locales (ru and en recommended). | Switch languages in the debug panel (works only with the panel enabled). |
| YG-14 | Upload a ZIP archive with a single `index.html` at the root. Uncompressed content must be **100 MB or less**. File and folder names must not contain spaces or Cyrillic characters. | [S] https://yandex.com/dev/games/doc/en/console/add-new-game ; https://yandex.com/dev/games/doc/en/concepts/requirements | Use a Vite build with `base: './'` (relative paths) and hashed ASCII filenames, zipped by CI. Add a CI size gate that fails above 90 MB. **Gap:** the CI packaging job. | CI script checks the unzipped size, the root `index.html` and filename characters. |
| YG-15 | **External hosts and CSP.** Hosts the game connects to must be declared in the draft's CSP tab with a reason and are reviewed by Yandex. Only https and wss are allowed; plain http and ws are refused. No paths, ports or IP addresses. A host is not approved if "almost all game data" loads from it. The game must tell the user about connection errors and offer a fix such as a retry button. CSP is not enforced in dev mode but is enforced in draft mode. | [S] https://yandex.com/dev/games/doc/en/console/add-new-game ; https://yandex.com/dev/games/doc/en/concepts/local-launch | Declare `api.<ourdomain>` (https) and `ws.<ourdomain>` (wss) on port 443, with the justification "server-authoritative multiplayer state; game assets are in the archive". **All static assets ship in the archive** (or on a Yandex-approved CDN, UNVERIFIED). Show a connection-error screen with a Retry button. **Gap:** domain, TLS, CSP declaration. | Test in **draft mode**, not dev, so CSP is active. Block the network to test the retry UI. |
| YG-16 | **Iframe hosting** is for games needing "complex server architecture (for example, multiplayer)". Request it through the support chat with the game name, developer account, an https link and a justification. It needs separate approval. (One locale says to email games-partners@yandex-team.com; which channel is current is UNVERIFIED.) | [S] https://yandex.com/dev/games/doc/en/console/add-new-game | **Fallback only.** The preferred route is archive plus CSP-approved wss (YG-15). If Yandex refuses the CSP host, request the iframe route. | Decide after the CSP review outcome. |
| YG-17 | Links in text or media must be added through the SDK and lead **only to our own games in the Yandex Games catalog**, and only at inactive moments such as menus or between levels. Third-party links are grounds for rejection, including on the loading screen. | [S] https://yandex.com/dev/games/doc/en/requirements/8/4/1 ; https://yandex.com/dev/games/doc/en/concepts/moderation | The Yandex build must hide links to Discord, the website, Google Play, social networks and the support web form. Use a build flag (`platform === 'yandex'`). Whether the privacy policy and terms of service may be linked is **UNVERIFIED**; ask support. | Grep the Yandex bundle for external URLs. Manual UI review. |
| YG-18 | No duplicates. The game must not copy, in full or in part, another game in the catalog. | [S] https://yandex.com/dev/games/doc/en/requirements/3/6 | The game is original IP. Do not publish multiple near-identical variants or regional clones. | Pre-submission catalog search for similar titles. |
| YG-19 | "The use of interactive artificial intelligence (AI) in games is prohibited." | [S] https://yandex.com/dev/games/doc/en/concepts/requirements (excerpt dated 2026-10-09) | No LLM-driven NPCs or chat in the Yandex build. Deterministic bots or AI opponents are probably fine, since only a third-party interpretation was found (UNVERIFIED). | Review design docs for generative-AI features. |
| YG-20 | Loading screen: no third-party links, unauthorized logos or copyright-infringing elements. | [S] https://yandex.com/dev/games/doc/en/concepts/moderation | Loading screen shows only our own logo and art, and no engine or vendor logos. | Visual review. |
| YG-21 | Moderators test on the browsers and operating systems listed in the requirements; failing on any of them means rejection. | [S] https://yandex.com/dev/games/doc/en/concepts/moderation | Browser and OS matrix: **UNVERIFIED** (the list was not retrieved). Plan at minimum: Yandex Browser, Chrome, Safari iOS, Firefox, Edge; desktop and mobile. | Fill in the list from the live page. Run the test matrix. |
| YG-22 | At least one clearly defined gameplay element (goal, win/lose, competition). Quoted as requirement 2.4. | [3P] GitHub issue quoting 2.4; UNVERIFIED | Met by the game design (strategy goals, battles). | Read live 2.4. |
| YG-23 | Content rules (section 8): for example, no shocking or unpleasant content, and promotional-material rules. | [S] https://yandex.com/dev/games/doc/en/requirements/8/3/6 ; https://yandex.com/dev/games/doc/en/console/promo-rules | Fantasy violence level, icons and screenshots need a content review. | Art and marketing review against section 8. |
| YG-24 | Moderation "usually takes 3–5 working days". Yandex reports about 10k new-game applications a month, so delays are possible. If a submission is older than 100 days, re-check the requirements. | [S] https://yandex.com/dev/games/doc/en/concepts/quick-start ; [3P] wnhub.io news | Plan at least 2 weeks of moderation buffer, plus re-submission cycles. | — |
| YG-25 | **Signed player identity.** `ysdk.getPlayer({ signed: true })` returns `player.signature`, made of two base64 parts (signature and profile data), meant for "authorizing the user and saving game state on your server". Rate limit: 20 requests per 5 min. ID is the only field guaranteed for all players; name and avatar are available only for authorized players. `getUniqueID()` is permanent and differs from the deprecated `getID()`. | [S] https://yandex.com/dev/games/doc/en/sdk/sdk-player | The server verifies the signature to bind a session to a Yandex unique ID. **UNVERIFIED:** (a) the algorithm (believed to be HMAC-SHA256, not confirmed); (b) whether guests get a signed payload; (c) the Turkish docs suggest the secret key is available only after purchases are connected. See R-2. | Prototype against a draft game. Read the `#signature` section live. |
| YG-26 | For archive uploads, the SDK wraps `localStorage` to make it reliable. | [S] https://yandex.com/dev/games/doc/en/sdk/sdk-player (excerpt) | Store only the session token and settings in localStorage. The server is the source of truth. | — |

### 2.2 SDK integration plan (Yandex)

All Yandex calls live in `client/src/platform/yandex/` behind `PlatformAdapter` (section 4). Game code never calls `ysdk` directly.

1. **Boot.** `index.html` includes `<script src="/sdk.js"></script>`; use the absolute S3 URL only for self-hosted test pages. `init()` awaits `YaGames.init()`. No SDK call is made before this resolves [S sdk-about].
2. **Locale.** Read `ysdk.environment.i18n.lang`, then load the i18n bundle (YG-13).
3. **Pause and resume wiring.** Use `ysdk.on('game_api_pause', …)` and `ysdk.on('game_api_resume', …)`, plus a `document.visibilitychange` fallback, to call `onPause` and `onResume` subscribers. Those mute audio, freeze rendering and input, and send `client_away` to the server (YG-04, R-3).
4. **Identity (guest-first).**
   - `player = await ysdk.getPlayer({ signed: true })`, or unsigned if signed is unavailable for guests (UNVERIFIED).
   - `POST /auth/yandex { uniqueId, signature? }` returns our `sessionToken`. The server upserts an account keyed by `yandex:<uniqueId>`.
   - If `!player.isAuthorized()`, play continues as a guest. At a natural moment, such as after the tutorial or before a purchase, show our own benefit dialog ("Save progress across devices"). If the player accepts, call `ysdk.auth.openAuthDialog()`, then `getPlayer` again and `POST /auth/yandex/upgrade`. Never block play (YG-08).
5. **Ready.** When the main menu is interactive, call `ysdk.features.LoadingAPI.ready()`. The world WebSocket connects after that. A connection failure shows a Retry screen; the game is still "ready", and the screen is interactive (YG-02, YG-15).
6. **Gameplay markers.** Call `ysdk.features.GameplayAPI.start()` when entering the world or battle view, and `stop()` when opening the shop or menus, before ads and on pause. Do not call `start()` on resume if the UI is in a stopped state (YG-03).
7. **Saves.** The server is authoritative. Optional `player.setData({ link: {acct, ts} })` acts as a cloud-save mirror if YG-09 / 1.13.3 demands it. Stay well below 100 `setData` calls per 5 min and 200 KB.
8. **Ads.** `ysdk.adv.showFullscreenAdv({callbacks:{onOpen,onClose,onError,onOffline}})` runs only at safe breaks. `ysdk.adv.showRewardedVideo({callbacks:{onOpen,onRewarded,onClose,onError}})` follows, and the server grants the reward with a rate limit. Pause and resume happen around both (YG-05, YG-06).
9. **Payments.** `payments = await ysdk.getPayments({ signed: true })`. Then `payments.purchase({ id, developerPayload })` sends the signature to the server, the server grants the item, and the client calls `payments.consumePurchase(token)`. At boot, call `payments.getPurchases()` and reconcile. Show prices from `payments.getCatalog()`. Method names are taken from excerpts and must be re-checked against the live sdk-purchases page (YG-10).
10. **Links.** Hide every external link in the Yandex build. `ysdk.features.GamesAPI` cross-promo links to our own games are optional (YG-17).

### 2.3 Risks specific to a server-authoritative MMO on Yandex Games

| # | Risk | Impact | Mitigation |
|---|---|---|---|
| R-1 | **Mandatory auth is forbidden (YG-08)**, but an MMO needs persistent identity. | Rejection if we gate play behind login. | Create guest accounts keyed by `getUniqueID()`, with a later upgrade to an authorized Yandex ID. Never show our own login or registration on Yandex. Guest accounts are persistent server accounts. |
| R-2 | **Trust in guest identity.** Signed payloads may not be available to guests, and the secret key may require purchases to be enabled (UNVERIFIED). | Account takeover or spoofing of `uniqueId`. | Issue our own random session token at first contact and store it in the SDK-wrapped localStorage. Bind `uniqueId` to the account only on first login. Require a signed payload before the account can buy anything or before a guest is merged with an authorized identity. Confirm the details with Yandex. |
| R-3 | **Pause rules versus a live world.** Tab switches and ads must pause sound and gameplay, but the server world continues. | Unfair losses while paused, or moderators may see "gameplay continuing". | Design the game to be pause-tolerant: async or timer-based strategy actions, no twitch combat. When the client reports `away`, the server stops the player's real-time interactions (for example, auto-resolve or defensive stance). Show a "Reconnecting / Resuming" overlay on return. Never show ads during real-time battles. |
| R-4 | **Cloud saves possibly required for IAP games (1.13.3).** | Purchase rejection or IAP disabled. | Mirror a minimal binding record via `setData`. Ask support in writing whether server persistence is acceptable. |
| R-5 | **The CSP host may not be approved**, or connections may be restricted to wss/https on standard ports. | No connectivity, so the game cannot run. | Use one API hostname and one WS hostname on 443, with valid TLS and no IPs. Ship all assets in the archive. The fallback is the iframe route (YG-16). |
| R-6 | **Origin and CORS on our Go server.** The game is served from Yandex domains (several TLDs) and the CDN hosting archive games. | Requests blocked by our own Origin checks. | Allow-list the Yandex game origins seen in draft mode (exact origin pattern UNVERIFIED; capture it from the browser in draft). Validate the session token rather than relying on Origin alone. |
| R-7 | **Server outage during moderation** leads to "game does not work" and rejection. | Rejection. | Keep a dedicated stable moderation environment, monitoring, and the retry UI (YG-15). Never deploy breaking server changes while a build is in moderation. |
| R-8 | **Cross-platform mentions.** Links or mentions of Google Play, Discord or the website break YG-17. | Rejection. | Use a build flag and a CI check that greps for URLs. |
| R-9 | **Ad-reward and IAP fraud** through client-side callbacks. | Economic exploits. | Grant everything server-side, idempotently. Rate-limit rewarded ads. Use signed purchases. |
| R-10 | **Interactive AI ban (YG-19).** | Rejection. | No generative AI features in the Yandex build. |
| R-11 | **Moderation queue and requirement drift.** Yandex docs change often; excerpts show versions dated 2025-12 to 2026-10. | Schedule slip. | Re-verify this document against the live pages within 2 weeks before submission. |

---

## 3. Google Play

### 3.1 Requirement list

| ID | Requirement | Source (accessed 2026-10-31) | Our project: how we satisfy it / gap | Verification method |
|---|---|---|---|---|
| GP-01 | **Target API level.** From 2026-08-31, "New apps and app updates must target Android 16 (API level 36) or higher". An extension to 2026-11-01 can be requested. Existing apps must target API 35 or higher to stay available to new users on newer OS versions. | [F] https://developer.android.com/google/play/requirements/target-sdk (official Help mirror: https://support.google.com/googleplay/android-developer/answer/11926878, not fetched) | Capacitor 8 template: `targetSdkVersion = 36`, `compileSdkVersion = 36` [F]. Satisfied if we keep the defaults. | Inspect `android/variables.gradle`. Play Console pre-launch report. |
| GP-02 | **Android App Bundle.** "From August 2021, new apps are required to publish with the Android App Bundle." Compressed install download must be 4 GB or less. | [F] https://developer.android.com/guide/app-bundle | `npx cap build android` defaults to an **AAB** (`releaseType` defaults to `'AAB'` in the CLI source) [F]. Our size is far below the limit. | Upload to the internal track. |
| GP-03 | **Play App Signing.** "When publishing your app to Google Play for the first time, you must also configure Play App Signing." The upload key is ours; the app signing key is held by Google. The page advises validity of "at least 25 years". | [F] https://developer.android.com/studio/publish/app-signing | Generate an upload keystore (owner action) and enroll at first upload. Store the keystore and passwords in the CI secret store and never in the repo. | Play Console → Setup → App signing. |
| GP-04 | **Play Billing Library version.** The FAQ banner says "By Aug 31, 2026, all new apps and updates … must use Billing Library version 8 or later" (extension to 2026-11-01). The table on the same page lists version 7 for 2026 and version 8 for 2027, an inconsistency on the page. | [F] https://developer.android.com/google/play/billing/deprecation-faq | Use a billing plugin that bundles **PBL 8 or later**, which satisfies both readings. Candidate plugins: `cordova-plugin-purchase` 13.20.0 and `@revenuecat/purchases-capacitor` 14.0.0 (npm, 2026-10-27/28) [F for versions]. **Which PBL version each bundles is UNVERIFIED**; check their Gradle dependencies. | Inspect the merged Gradle dependency tree (`./gradlew app:dependencies`). |
| GP-05 | **16 KB memory page size.** "Starting February 1, 2027, if your app updates don't support 16 KB memory page sizes, you won't be able to release these updates." Apps using only Java/Kotlin already comply. | [F] https://developer.android.com/guide/practices/page-sizes | Capacitor core is Java/Kotlin. **Gap:** any plugin with `.so` libraries must be checked. | APK Analyzer or `zipalign -c -P 16` check; 16 KB emulator test. |
| GP-06 | **Android 16 behavior changes for apps targeting 36.** Edge-to-edge opt-out is "deprecated and disabled". Predictive back is on by default, and `onBackPressed` is not called. Orientation and resizability restrictions are ignored on screens of 600dp or wider, **except games** (by `android:appCategory`). | [F] https://developer.android.com/about/versions/16/behavior-changes-16 ; [F] Capacitor 8 upgrade guide (removed `adjustMarginsForEdgeToEdge` in favor of the System Bars plugin and CSS `env()` safe-area variables) | Set `android:appCategory="game"`. Use CSS `env(safe-area-inset-*)`. Handle the back button through `@capacitor/app` `backButton` (8.1.2 [F]) so it closes in-game dialogs. | Test on an API 36 emulator, phone and tablet. |
| GP-07 | **Android developer verification.** From 2026-09-30, verification protections apply in Brazil, Indonesia, Singapore and Thailand, expanding globally in 2027. "Google Play automatically registers 99% of apps." | [F] https://developer.android.com/developer-verification | Play-distributed: likely automatic. Owner should confirm registration status in Play Console. | Play Console → developer verification page. |
| GP-08 | **Data safety form.** All apps published on Play must complete it (apps only on internal testing are exempt), including data collected by third-party SDKs. Even apps that collect no data must complete it and link a privacy policy. | [S] https://support.google.com/googleplay/android-developer/answer/10787469 | Declare at least: user IDs (account ID), purchase history, in-app messages (chat), app interactions and gameplay, diagnostics or crash logs, IP-based approximate location if logged. Encrypted in transit: yes (TLS). Deletion: yes (GP-10). **Gap:** a data inventory owned by the backend. | Cross-check against the server data model and the SDK list. |
| GP-09 | **Privacy policy** link required, in Play Console and in-app. | [S] https://support.google.com/googleplay/android-developer/answer/10787469 | **Owner must supply** a hosted privacy policy URL that names the developer, the data collected, retention and deletion. | URL reachable, not a PDF behind a login. |
| GP-10 | **Account deletion.** Apps that allow account creation must let users request deletion of the account and associated data **in the app** and through a **web link resource**, and disclose both in the Data safety form. | [S] https://support.google.com/googleplay/android-developer/answer/13327111 | Add an in-app Settings → "Delete account" option that calls the server `DELETE /account` (with a grace period). Add a public web page where a user can request deletion, for example by game ID plus email or code. **Gap:** server endpoint, web page, retention policy. Guest accounts count as accounts. | Manual test. Fill in the Data safety deletion questions. |
| GP-11 | **Payments policy.** In-app purchases of virtual currencies and items must use Google Play's billing system. Earned (not sold) currency is exempt. Currency may be used only in the title it was bought for. In the US, from 2025-12-09 the alternative billing and external content links programs exist (fees from 2026-10-01; link-out reporting extended to 2026-12-01). Regional programs exist elsewhere (EEA, UK, Japan, Australia, India). | [S] https://support.google.com/googleplay/android-developer/answer/9858738 ; [S] https://support.google.com/googleplay/android-developer/answer/15582165 | The Android build uses **Play Billing only** (no Yandex payments, no web checkout links). Use server-side verification of purchase tokens (Google Play Developer API) plus Real-time Developer Notifications; this approach is a design recommendation, and the API docs were not fetched (UNVERIFIED). Our server grants gems idempotently and then acknowledges or consumes. Leave the alternative billing programs out of scope for v1. | Test with license testers on the closed track. |
| GP-12 | **Minimum functionality and webview spam.** Apps that are "just webviews of existing websites" are considered "webview spam, and are removed from Play". | [S] https://android-developers.googleblog.com/2020/10/developer-tips-and-guides-common-policy.html (2020; current Policy Center wording UNVERIFIED) | **Bundle all web assets inside the AAB** (never `server.url` in production; Capacitor marks it "not intended for use in production" [F]). Add native integrations: Play Billing, hardware back, offline and reconnect screen, haptics, optionally push notifications. Provide a full game experience. **Risk remains** because a review is human. | Pre-launch report. Policy review before submission. |
| GP-13 | **Testing requirement for new personal accounts.** Personal accounts created after 2023-11-13 must run a **closed test with at least 12 testers opted in for 14 consecutive days** before applying for production access (lowered from 20 testers in Dec 2024). | [3P] multiple (e.g. https://ontest.app/blog/google-play-12-testers-14-days-requirement-explained); **official Help page not fetched → UNVERIFIED** | Owner should consider an **organization account** (reported as exempt, UNVERIFIED; requires D-U-N-S). Otherwise recruit at least 15 testers (buffer) for at least 14 days and budget about 3 weeks. | Play Console Dashboard shows the requirement status. |
| GP-14 | **Content rating.** The IARC questionnaire is required; "We don't allow apps without a content rating". Redo it when content changes. | [S] https://support.google.com/googleplay/android-developer/answer/9898843 ; https://support.google.com/googleplay/android-developer/answer/188189 | Answer honestly: fantasy violence, user interaction (chat), in-app purchases. | Play Console → App content. |
| GP-15 | **Target audience and Families.** If the target audience includes children and the app shows ads, use only Families self-certified ad SDKs. Apps not for children are not required to. | [S] https://support.google.com/googleplay/android-developer/answer/9867159 ; https://support.google.com/admob/answer/6223431 | Recommendation: target audience **13+ or 16+**, not children, given chat, IAP and an MMO. That avoids the Families obligations. Ads on Android are undecided; if added, use an ad SDK of our choice (AdMob, UNVERIFIED details). | App content → Target audience. |
| GP-16 | **User-generated content.** Apps with UGC such as in-game chat need terms of use, in-app reporting of objectionable content and users, user blocking, and reasonable moderation. | [S] https://support.google.com/googleplay/android-developer/answer/9876937 ; https://support.google.com/googleplay/android-developer/answer/12923286 | **Gap:** chat needs "Report message/player" and "Block player" (server-side mute list), a profanity filter, and a moderation queue. The same features help on Yandex. | Manual UI review; the functions must be clearly labelled. |
| GP-17 | **Developer account.** Organization accounts need a D-U-N-S number. Personal accounts need ID verification. A one-time registration fee (reported as US$25). | [S] https://support.google.com/googleplay/android-developer/answer/13634885 ; fee [3P] UNVERIFIED | Owner action. | — |
| GP-18 | **WebView transport.** Capacitor Android serves the app from `https://localhost` (`androidScheme` defaults to `https`, `hostname` to `localhost`). Cleartext is disabled from API 28. External navigation opens in the system browser. | [F] Capacitor v8 config reference (capacitor-docs `versioned_docs/version-v8/main/reference/config.md`) | Our Go server must accept Origin `https://localhost` for REST/CORS and WebSocket upgrades. Use TLS everywhere. | Test against staging from a release build. |

### 3.2 Packaging plan with Capacitor

Facts [F] (npm registry and the Capacitor v8 docs source, accessed 2026-10-31):

- `@capacitor/core`, `@capacitor/cli` and `@capacitor/android` **latest = 8.6.0** (published 2026-10-28). `9.0.0-beta.0` is tagged `next`, and the docs site shows v9 as "unreleased". **Use 8.x.**
- Requirements: Node **≥ 22** (`engines` in `@capacitor/cli`; the v9 docs already say Node 24). Android Studio **2025.2.1 (Otter) or newer**. A separate JDK is not needed; Android Studio bundles it. The `capacitor-android` module compiles with Java 21.
- Template defaults: `minSdkVersion = 24`, `compileSdkVersion = 36`, `targetSdkVersion = 36`; AGP 8.13.0; Gradle 8.14.3. Supported devices are API 24 and up (Android 7+), with a WebView of Chrome 60 or later.
- Workflow: build the web app (`npm run build` for Vite), then `npx cap sync` "copy[ies] over your already built web bundle" and updates native dependencies. `npx cap build android` produces a signed AAB or APK; AAB is the default.

Plan:

1. Add a `client/capacitor.config.ts` with `appId: 'com.<owner>.<game>'` (owner decides; **permanent once published**), `appName`, and `webDir: 'dist'` (Vite output; the Capacitor default is `www`). Do **not** set `server.url` or `cleartext` in release builds.
2. Vite build mode `android`: `VITE_PLATFORM=android` selects `AndroidAdapter`, uses `base: './'`, and strips Yandex SDK references.
3. `npx cap add android` once, and commit the `android/` folder. Then edit:
   - `AndroidManifest.xml`: `android:appCategory="game"`, `INTERNET` permission (template), and the orientation decision.
   - `variables.gradle`: keep 24/36/36.
   - `versionCode` and `versionName`, driven by CI from the git tag.
4. Plugins: `@capacitor/app` (back button, pause/resume through `appStateChange`), a billing plugin (GP-04), optionally `@capacitor/haptics` and `@capacitor/splash-screen`. Audit every plugin for native `.so` files (GP-05).
5. Release build in CI:
   `npm ci && npm run build:android && npx cap sync android && npx cap build android --keystorepath $KS --keystorepass $KS_PASS --keystorealias upload --keystorealiaspass $KEY_PASS --androidreleasetype AAB`
   The flag names come from the Capacitor 8.6.0 CLI source [F]. An alternative is a Gradle `signingConfig` that reads environment variables, followed by `./gradlew :app:bundleRelease`.
6. Upload the AAB to the **internal testing** track, then **closed testing** (GP-13 clock), then production.

### 3.3 Signing and build steps (owner plus CI)

1. The owner creates an **upload key**, through Android Studio → Build → Generate Signed Bundle → Create new, or with `keytool -genkeypair`. Use a validity of 25 years or more [F]. Keep two offline backups.
2. Store the keystore (base64), store password, alias and key password as CI secrets.
3. On first upload, Play Console enrolls the app in **Play App Signing**: Google generates and holds the app signing key [F].
4. Export the upload certificate if Play support asks for it: `keytool -export -rfc -keystore upload.jks -alias upload -file upload_cert.pem` [F].
5. Record the **app signing key SHA-256** from Play Console. It may be needed later for Google APIs, such as sign-in.

### 3.4 Data safety (draft answers; owner and backend to confirm)

| Data type | Collected | Shared | Purpose | Optional? |
|---|---|---|---|---|
| User IDs (our account ID, Play purchase tokens) | Yes | No | App functionality, account management, fraud prevention | Required |
| Purchase history | Yes | No | App functionality | Required for purchasers |
| In-app messages (chat) | Yes | No | App functionality, moderation | Optional (if chat is used) |
| App interactions / gameplay | Yes | No | App functionality, analytics | Required |
| Crash logs / diagnostics | Yes, if we add crash reporting | Depends on SDK | Analytics | — |
| Approximate location from IP | Only if logged or derived | No | Fraud prevention / security | — |
| Email | Only if we add email login or recovery | No | Account management | Optional |

- Data encrypted in transit: **Yes**, TLS/WSS only.
- Users can request deletion: **Yes**, in-app plus web link (GP-10).
- Every third-party SDK (billing, analytics, ads) must be added to this inventory before submission [S].

### 3.5 Account deletion (design)

- **In-app:** Settings → Account → "Delete account". The user confirms with a typed phrase. The server marks the account `pending_deletion`, terminates sessions, and hard-deletes or anonymizes personal data after a grace period (the period is the owner's decision; it is not specified by the sources).
- **Web:** `https://<ourdomain>/delete-account` explains the steps and accepts requests (account ID plus a verification code shown in game, or email). The URL goes into the Data safety form [S].
- Purchase records that must be retained for tax or legal reasons need a stated retention period in the privacy policy (legal input required, UNVERIFIED).
- Use the same server endpoint for Yandex players. This is good practice; no Yandex rule on deletion was found in this pass.

### 3.6 Billing (design)

- Catalog: consumable gem packs, plus optionally non-consumables (cosmetics). Product IDs are shared across platforms where possible, for example `gems_100`, mapped per platform in a server table.
- Flow: client `purchase(sku)` → Play Billing UI → client sends `purchaseToken` to the server → server verifies with the Google Play Developer API using a **service account** → grants gems idempotently by `orderId`/token → client consumes or acknowledges. Unacknowledged purchases are refunded by Google after a time limit (exact window UNVERIFIED; check the Billing docs).
- Use RTDN (Pub/Sub) for refunds and voids, and have the server claw back gems (UNVERIFIED details; design intent).
- No price display that is not sourced from the store. No external payment links (GP-11).

### 3.7 Testing tracks

| Track | Purpose | Notes |
|---|---|---|
| Internal testing | Smoke-test builds; up to a small group of testers | Exempt from the Data safety display [S] |
| Closed testing | Required gate for new personal accounts: 12 or more testers opted in for 14 consecutive days [3P / UNVERIFIED] | Start as early as possible. Testers need Google accounts. Track opt-ins. |
| Open testing | Optional public beta | — |
| Production | After production access is granted (personal accounts) | Staged rollout recommended |

---

## 4. Shared adapter interface proposal (design only)

Goals:
- Game code depends only on `PlatformAdapter`.
- Implementations: `YandexAdapter`, `AndroidAdapter` (Capacitor), and `WebDevAdapter` (local dev, mocks).
- Selected at build time through `import.meta.env.VITE_PLATFORM`, so that Yandex SDK code never ships in the Android bundle and vice versa.
- Platform side effects that grant value (rewards, purchases) always end in a **server-verifiable proof**.

```ts
// client/src/platform/PlatformAdapter.ts — DESIGN SKETCH ONLY (not implemented)

export type PlatformId = 'yandex' | 'android' | 'web-dev';

export interface PlatformCapabilities {
  interstitialAds: boolean;
  rewardedAds: boolean;
  purchases: boolean;
  platformLogin: boolean;      // Yandex ID dialog; Android: our own login
  cloudSave: boolean;          // Yandex player.setData mirror
  externalLinks: boolean;      // false on Yandex (YG-17)
  accountDeletionUI: boolean;  // true on Android (GP-10)
}

/** Proof the server verifies before trusting a platform identity. */
export interface PlayerIdentity {
  platform: PlatformId;
  platformUserId: string;       // Yandex getUniqueID(); Android: our account id / device-bound guest id
  isAuthorized: boolean;        // Yandex: player.isAuthorized(); guest otherwise
  displayName?: string;         // only for authorized Yandex players
  avatarUrl?: string;
  signedPayload?: string;       // Yandex getPlayer({signed:true}).signature, if available
}

export interface PurchaseResult {
  status: 'success' | 'cancelled' | 'pending' | 'error';
  productId: string;
  /** Opaque, server-verifiable proof: Yandex signature, or Play purchaseToken. */
  receipt?: string;
  platformOrderId?: string;
  errorCode?: string;
}

export interface CatalogItem {
  productId: string;
  title: string;
  description?: string;
  priceDisplay: string;         // localized price string from the store, never hard-coded
  priceCurrency?: string;
}

export type AdResult = 'shown' | 'rewarded' | 'closed' | 'unavailable' | 'error';

export type Unsubscribe = () => void;

export interface PlatformAdapter {
  readonly id: PlatformId;
  readonly capabilities: PlatformCapabilities;

  /** Load and initialize the platform SDK. Must resolve before any other call (YG-01). */
  init(): Promise<void>;

  /** Locale chosen by the platform (Yandex: environment.i18n.lang, YG-13; Android: device locale). */
  getLocale(): string;

  /** Tell the platform the game is playable (Yandex LoadingAPI.ready, YG-02; Android: hide splash). */
  signalReady(): void;

  /** Optional gameplay markers (Yandex GameplayAPI.start/stop, YG-03). No-op elsewhere. */
  gameplayStart?(): void;
  gameplayStop?(): void;

  /** Identity, guest-first. Never blocks play on Yandex (YG-08). */
  getPlayerIdentity(): Promise<PlayerIdentity>;
  /** User-initiated login upgrade (Yandex auth dialog; Android: our own login flow). */
  requestLogin?(): Promise<PlayerIdentity>;

  /** Lifecycle: focus loss, tab hidden, ad open, app backgrounded (YG-04/05; Capacitor appStateChange). */
  onPause(handler: () => void): Unsubscribe;
  onResume(handler: () => void): Unsubscribe;

  /** Ads. The adapter pauses and resumes around them itself. Rewards are granted by the server. */
  showInterstitial?(placement: string): Promise<AdResult>;
  showRewarded?(placement: string): Promise<AdResult>;

  /** Purchases: the client never grants items; the server verifies `receipt` and grants. */
  getCatalog?(): Promise<CatalogItem[]>;
  purchase?(productId: string, serverOrderId: string): Promise<PurchaseResult>;
  /** Purchases not yet confirmed by the server (Yandex getPurchases / Play queryPurchases). */
  getPendingPurchases?(): Promise<PurchaseResult[]>;
  /** Call only after the server confirmed the grant (Yandex consumePurchase / Play consume+ack). */
  finalizePurchase?(receiptOrToken: string): Promise<void>;

  /** Optional small mirror for platforms that require cloud saves (YG-09 / 1.13.3). */
  cloudSaveMirror?(data: Record<string, unknown>): Promise<void>;

  /** External link handling. Returns false if the platform forbids it (Yandex, YG-17). */
  openExternalUrl?(url: string): Promise<boolean>;

  /** Android hardware back (GP-06). Return true if handled by the game. */
  onBackButton?(handler: () => boolean): Unsubscribe;

  /** Show the native share sheet, a rate prompt, etc. Optional. */
  share?(text: string, url?: string): Promise<void>;
}
```

Server-side counterparts (design notes, no code):
- `POST /auth/platform` takes `{platform, platformUserId, signedPayload?}` and returns a session token.
- `POST /purchases/verify` takes `{platform, productId, receipt, serverOrderId}`, grants idempotently and returns `{granted, finalize: true}`.
- `POST /ads/reward` takes `{placement, nonce}` and is rate-limited.
- `DELETE /account` handles deletion with a grace period (GP-10).

---

## 5. Pre-submission checklist and owner-supplied items

### 5.1 Must be done before Yandex Games submission

- [ ] Re-verify every `[S]` row in section 2.1 against the live yandex.ru / yandex.com pages. Update the requirement IDs.
- [ ] Implement `YandexAdapter`: SDK init, `LoadingAPI.ready`, `GameplayAPI` markers, pause and resume, guest-first identity, ads, payments.
- [ ] Server: Yandex identity binding (`yandex:<uniqueId>`), signature verification (algorithm to be confirmed), guest persistence, upgrade flow.
- [ ] Server: purchase signature verification, idempotent grants, reconciliation of unconsumed purchases.
- [ ] Decide cloud-save mirroring (YG-09 / 1.13.3). Get written confirmation from Yandex support.
- [ ] Mobile CSS hardening: no scroll, no swipe-to-refresh, no long-press selection or context menu.
- [ ] Automatic locale from `i18n.lang`, with fallbacks. Complete translations for each declared language, including store texts, screenshots and video.
- [ ] Remove all external links from the Yandex build. CI check that greps for URLs.
- [ ] Connection-error screen with Retry.
- [ ] CI: ZIP with root `index.html`, at most 100 MB uncompressed, ASCII filenames without spaces.
- [ ] CSP tab: declare the API and WS hosts with justification. Test in **draft** mode.
- [ ] Stable moderation server environment and monitoring.
- [ ] Ad placements only at safe breaks. Run the pause/resume test matrix on the browser and OS list.
- [ ] Email games-partners@yandex-team.com to enable purchases. Configure the catalog.
- [ ] Draft: orientation, cloud-save flag, languages, icon, cover, screenshots, description.
- [ ] Original-content and duplicate check. No generative AI features.

### 5.2 Must be done before Google Play submission

- [ ] Capacitor 8.6.x Android project: appId, `appCategory="game"`, target and compile SDK 36, edge-to-edge safe areas, back-button handling.
- [ ] Bundle all assets locally; no `server.url`. Add native value (billing, back button, offline and reconnect, haptics) to avoid webview-spam treatment.
- [ ] Billing plugin with PBL 8 or later. Server-side token verification with a service account. Acknowledge or consume after the grant. RTDN for refunds.
- [ ] Our own Android identity: guest account plus an optional login. Account deletion in the app and on the web.
- [ ] UGC: report message or player, block player, chat filter, moderation tooling, terms of use.
- [ ] Privacy policy URL, Data safety form, content rating questionnaire, target audience (13+ or 16+), ads declaration.
- [ ] Server: accept Origin `https://localhost` for WebView requests.
- [ ] 16 KB check of native libraries from plugins.
- [ ] Signed AAB through CI. Internal track, then closed track (12 or more testers for 14 consecutive days if a personal account), then the production access application.
- [ ] Store listing: icon 512×512, feature graphic, screenshots, short and full description. Exact asset specs are UNVERIFIED in this pass; check Play Console.

### 5.3 Credentials, accounts and decisions the owner must supply

| Item | Platform | Notes |
|---|---|---|
| Yandex ID with Yandex Games developer console access | Yandex | The console account that owns the game. |
| Legal entity or self-employed details, payout and contract acceptance for monetization | Yandex | Required documents are UNVERIFIED; check the console. |
| Purchases enablement (email to games-partners@yandex-team.com) and the **signature secret key** | Yandex | Store the key in the server secret store only. |
| Production domain(s), TLS certificates and hostnames for API and WS | Both | Needed for the CSP declaration and Android. |
| Google Play Console developer account (personal or **organization with D-U-N-S**) plus identity verification | Google | An organization account is reportedly exempt from the 12-tester rule (UNVERIFIED). |
| Google payments merchant profile (to sell IAP) | Google | Set up in Play Console. |
| Upload keystore and passwords (generated by the owner, backed up) | Google | Never commit to git. |
| Google Cloud project plus **service account JSON** linked to Play Console (Developer API for purchase verification and CI uploads) | Google | Least privilege. |
| Pub/Sub topic for RTDN | Google | Optional but recommended. |
| Final `applicationId` (package name) | Google | Permanent. |
| Privacy policy URL, terms of use URL, account deletion URL, support email | Both (Yandex link visibility to be confirmed, YG-17) | Legal review recommended. |
| 12 or more (ideally 15 or more) closed-test testers with Google accounts | Google | Only if a personal account. |
| Target age rating and audience decision; ads-on-Android decision | Both | Affects Families, Data safety and content rating. |
| Launch languages | Yandex (mandatory auto-detect) | Translations of the store listing too. |
| Store assets (icons, covers, screenshots, video) per language | Both | — |

---

## 6. Sources

All accessed **2026-10-31**. Tag meanings are in section 1.2.

### Yandex Games ([S]: the live pages were blocked in this session and must be re-verified)
1. Game requirements: https://yandex.com/dev/games/doc/en/concepts/requirements (RU: https://yandex.ru/dev/games/doc/ru/concepts/requirements)
2. Req. 1.2 Authorization: https://yandex.com/dev/games/doc/en/requirements/1/2
3. Req. 1.3 Sound outside the game: https://yandex.com/dev/games/doc/en/requirements/1/3
4. Req. 1.9 Progress saving: https://yandex.com/dev/games/doc/en/requirements/1/9
5. Req. 1.10 Correct display: https://yandex.com/dev/games/doc/en/requirements/1/10
6. Req. 1.12 Monetization: https://yandex.com/dev/games/doc/en/requirements/1/12
7. Req. 1.19 SDK methods: https://yandex.com/dev/games/doc/en/requirements/1/19
8. Req. 2.14 Automatic language detection: https://yandex.com/dev/games/doc/en/requirements/2/14
9. Req. 3.6 Game duplicates: https://yandex.com/dev/games/doc/en/requirements/3/6
10. Req. 4.4 Ad placement (not retrieved): https://yandex.com/dev/games/doc/en/requirements/4/4
11. Req. 8.3.6 Unpleasant content: https://yandex.com/dev/games/doc/en/requirements/8/3/6
12. Req. 8.4.1 Links to other games: https://yandex.com/dev/games/doc/en/requirements/8/4/1
13. SDK installation and use: https://yandex.com/dev/games/doc/en/sdk/sdk-about
14. Player data: https://yandex.com/dev/games/doc/en/sdk/sdk-player
15. In-game purchases: https://yandex.com/dev/games/doc/en/sdk/sdk-purchases
16. Ads: https://yandex.com/dev/games/doc/en/sdk/sdk-adv
17. Game loading and gameplay markup: https://yandex.com/dev/games/doc/en/sdk/sdk-game-events
18. Pause and resume events: https://yandex.com/dev/games/doc/en/sdk/sdk-events (seen via the zh/vi locale versions)
19. Upload a game (archive, CSP, iframe): https://yandex.com/dev/games/doc/en/console/add-new-game
20. Moderation: https://yandex.com/dev/games/doc/en/concepts/moderation
21. Quick start: https://yandex.com/dev/games/doc/en/concepts/quick-start
22. Languages and domains: https://yandex.com/dev/games/doc/en/concepts/languages-and-domains
23. Launch from local server (dev and draft CSP behavior): https://yandex.com/dev/games/doc/en/concepts/local-launch
24. Release notes: https://yandex.com/dev/games/doc/en/release-notes
25. Promo materials rules: https://yandex.com/dev/games/doc/en/console/promo-rules
26. [3P] wnhub.io news on the moderation load: https://wnhub.io/news/other/item-51215
27. [3P] defold-yagames (purchase handling practice): https://github.com/indiesoftby/defold-yagames

### Google Play / Android
28. [F] Target API level requirements: https://developer.android.com/google/play/requirements/target-sdk (page "last updated" 2026-10-28)
29. [S] Target API (Play Console Help, not fetched): https://support.google.com/googleplay/android-developer/answer/11926878
30. [F] About Android App Bundles: https://developer.android.com/guide/app-bundle (updated 2026-06-24)
31. [F] Sign your app / Play App Signing: https://developer.android.com/studio/publish/app-signing (updated 2026-03-06)
32. [F] Billing Library deprecation FAQ: https://developer.android.com/google/play/billing/deprecation-faq (updated 2026-10-23)
33. [F] 16 KB page sizes: https://developer.android.com/guide/practices/page-sizes (updated 2026-10-16)
34. [F] Android 16 behavior changes: https://developer.android.com/about/versions/16/behavior-changes-16
35. [F] Android developer verification: https://developer.android.com/developer-verification
36. [S] Data safety section: https://support.google.com/googleplay/android-developer/answer/10787469
37. [S] Account deletion requirements: https://support.google.com/googleplay/android-developer/answer/13327111
38. [S] Payments policy: https://support.google.com/googleplay/android-developer/answer/9858738
39. [S] US policy update (alternative billing / external links): https://support.google.com/googleplay/android-developer/answer/15582165
40. [S] Content ratings: https://support.google.com/googleplay/android-developer/answer/9898843 and https://support.google.com/googleplay/android-developer/answer/188189
41. [S] Target audience and app content: https://support.google.com/googleplay/android-developer/answer/9867159
42. [S] AdMob families guidance: https://support.google.com/admob/answer/6223431
43. [S] UGC policy: https://support.google.com/googleplay/android-developer/answer/9876937 and https://support.google.com/googleplay/android-developer/answer/12923286
44. [S] Organization accounts / D-U-N-S: https://support.google.com/googleplay/android-developer/answer/13634885
45. [S] Android Developers Blog, "webview spam" (2020): https://android-developers.googleblog.com/2020/10/developer-tips-and-guides-common-policy.html
46. [3P] 12 testers / 14 days: https://ontest.app/blog/google-play-12-testers-14-days-requirement-explained ; https://dev.to/davekurian/google-play-closed-testing-prepare-for-the-14-day-wait-before-launch-5cii

### Capacitor
47. [F] npm registry: `@capacitor/core`, `@capacitor/cli`, `@capacitor/android` (latest 8.6.0, 2026-10-28; next 9.0.0-beta.0): https://registry.npmjs.org/@capacitor/cli
48. [F] `@capacitor/cli@8.6.0` package source: Android template `variables.gradle` (minSdk 24, compile/target 36), AGP 8.13.0, Gradle 8.14.3, `engines.node >=22.0.0`, `cap build` default AAB and keystore flags; `@capacitor/android@8.6.0` build.gradle (Java 21): https://registry.npmjs.org/@capacitor/cli/-/cli-8.6.0.tgz ; https://registry.npmjs.org/@capacitor/android/-/android-8.6.0.tgz
49. [F] Capacitor v8 docs source (the capacitorjs.com site was blocked; content read from its docs repo): environment setup, Android getting started, workflow, updating to 8.0, config reference: https://raw.githubusercontent.com/ionic-team/capacitor-docs/main/versioned_docs/version-v8/main/ (corresponding site pages: https://capacitorjs.com/docs/getting-started/environment-setup, https://capacitorjs.com/docs/android, https://capacitorjs.com/docs/basics/workflow, https://capacitorjs.com/docs/updating/8-0, https://capacitorjs.com/docs/config)
50. [F] Billing plugin versions (npm): `cordova-plugin-purchase` 13.20.0, `@revenuecat/purchases-capacitor` 14.0.0, `@capacitor/app` 8.1.2

### Open verification items (UNVERIFIED)
- The exact current Yandex requirement numbering and wording, especially 1.13.3 (cloud saves for IAP games), 4.x (ads), 2.4, and the browser and OS test list.
- The Yandex signature algorithm and key availability for players and purchases, and whether guests receive signed payloads.
- Which Yandex origin(s) archive-hosted games run on (needed for our server's Origin allow-list).
- Whether Yandex allows privacy policy and terms links.
- The official Play Console Help wording of the 12-tester / 14-day rule, and whether organization accounts are exempt.
- The current Play Policy Center wording of Minimum Functionality / webview spam.
- Which PBL version each Capacitor billing plugin bundles.
- The Play Console registration fee, and the store listing asset specifications.
