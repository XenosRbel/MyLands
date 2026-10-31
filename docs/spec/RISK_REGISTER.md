# Risk Register

Owner: Agent F (QA / Security Engineer), with Agent E (UI/UX) and Agent G (Release / Platform) as contributors
Status: v1, 2026-10-31. Review it at every milestone and before each store submission.

Inputs:
- `docs/spec/PLATFORM_REQUIREMENTS.md` (Yandex YG-xx and Google Play GP-xx requirements, risks R-1 to R-11)
- `docs/spec/SECURITY.md` (findings S-1 to S-13)
- `docs/spec/TEST_STRATEGY.md`
- `docs/spec/UI_UX_SPEC.md` §9
- `docs/KNOWN_LIMITATIONS.md`
- The code at commit `dfa660e`

**Scales.** Likelihood (L) and Impact (I) are each 1 (low) to 3 (high). **Score = L × I**:
- 6–9 is **high** and needs an owner action before the next release gate.
- 3–4 is medium.
- 1–2 is low.

**Status values:**
- **Open:** no mitigation in place.
- **Partial:** some mitigation exists.
- **Mitigated:** controls exist and are tested; the risk is monitored.
- **Accepted:** a conscious decision, recorded in `docs/DECISIONS.md`.

**Owners are roles:**
- Product owner (PO)
- Backend
- Frontend
- Game design (GD)
- UI/UX (Agent E)
- QA/Sec (Agent F)
- Release/Platform (Agent G)
- Legal / owner

---

## 1. Technical risks

| ID | Risk | L | I | Score | Mitigation | Owner | Status |
|---|---|---|---|---|---|---|---|
| TR-01 | **Performance at scale unknown.** Only one baseline exists: 200 players, 648 req/s, p95 ≤ 190 ms, join p95 1.2 s, all on one shared host (`docs/reports/LOAD_TEST_2026-10-31.md`). Every GET for overview or city takes the player's row locks and writes the settle, and clients poll every 15 s, so lock contention and write amplification may limit concurrent players per instance. | 2 | 3 | **6** | Extend `cmd/loadtest` with the L1–L6 scenarios (TEST_STRATEGY §6) on a dedicated database host and with 1 000+ players. Re-run after the spawn-query fix. If needed, add read-only fast paths that compute settled values without persisting them, increase the poll interval, or add push (TR-05). | Backend, QA/Sec | Partial |
| TR-02 | **Single PostgreSQL instance** is a single point of failure for state, sessions and the event queue. | 2 | 3 | **6** | Managed PostgreSQL with a standby. PITR and WAL archiving (SECURITY §13). `/readyz` already reports database loss. | Release/Platform | Open |
| TR-03 | **Data loss between backups.** There is no PITR; only logical dumps exist (`scripts/backup.sh`, `verify-backup.sh`). | 2 | 3 | **6** | Daily dump plus WAL archiving (RPO ≤ 5 min), off-site encrypted storage, a quarterly restore drill. | Release/Platform | Partial |
| TR-04 | **The rate limiter is in memory and per instance.** Limits multiply with instance count, and behind a proxy every IP shares one bucket (S-4). | 3 | 2 | **6** | Trusted real-IP header. Limits at the proxy or in shared storage (D-007 already defers Redis to this point). Stricter buckets for authentication. | Backend | Open |
| TR-05 | **Polling instead of push.** Each player sends about 0.13 requests per second while visible, events are delayed by up to 15 s, and nothing arrives while the tab is hidden or the app is backgrounded. | 3 | 2 | **6** | Add SSE or WebSocket for notifications (the CSP host must be wss, YG-15) and Android push for incoming attacks. Keep polling as the fallback. | Backend, Frontend | Accepted for v0.1 |
| TR-06 | **Argon2id memory DoS.** 19 MiB per hash, no concurrency bound (S-4). | 2 | 3 | **6** | Hashing semaphore; per-login back-off; load scenario L5. | Backend | Open |
| TR-07 | **A poison event blocks a player** with 500s until it is parked after 5 worker attempts. | 2 | 2 | 4 | Already parked and no longer blocking (tested). Add an alert on `rf_events_parked > 0` and `rf_events_overdue`. Add an admin UI for retry. | Backend | Partial |
| TR-08 | **Live pack migration corrupts a running world.** A rename or removal goes wrong. | 1 | 3 | 3 | Dry run plus a refusal on missing ids (tested). `world_pack_history`. Backup before migrating. | Backend, GD | Mitigated |
| TR-09 | **Bleeding-edge toolchain.** React 19.3, TypeScript 7, Vite 8, Vitest 5, Go 1.27, Capacitor 8.5. Ecosystem breakage or plugin incompatibility is possible. | 2 | 2 | 4 | Lockfiles committed. Renovate with CI. Pin Capacitor 8.x (PLATFORM_REQUIREMENTS §3.2). | Frontend | Partial |
| TR-10 | **Untested deployment artifacts.** The Docker image and compose file were never run, the Android project was never generated, and no AAB has been built (KNOWN_LIMITATIONS). | 3 | 2 | **6** | A CI job that builds the image, starts compose, runs `play.mjs`, and builds an AAB on the internal track. | Release/Platform | Open |
| TR-11 | **Browser E2E is not in CI**, and two integration tests can `t.Skip` on a map seed without a level-1 lair. | 2 | 2 | 4 | Run E2E in CI; seed a lair in setup instead of skipping (TEST_STRATEGY §5). | QA/Sec | Open |
| TR-12 | **Observability gaps.** `/metrics` exists but is public (S-3). There are no alerts, no tracing, and request logs have no account id. | 2 | 2 | 4 | A private metrics endpoint, alerting rules (5xx rate, overdue events, transaction retries, database connections) and structured audit logs. | Release/Platform | Open |
| TR-13 | **Unbounded table growth.** `sessions` (never pruned), `notifications`, `reports`, `resource_ledger`, `messages`. | 2 | 2 | 4 | Retention jobs in the worker (sessions, read notifications older than 30 days); partition the ledger by month; soak scenario L6. | Backend | Open |
| TR-14 | **Client clock skew.** Some UI checks use `Date.now()` instead of server time (UI_UX_SPEC gap 11). | 2 | 1 | 2 | Replace those checks with `api.now()`. | Frontend | Open |
| TR-15 | **Security findings S-1 to S-5** (session not revoked on world-screen logout, admin bootstrap squatting, public `/metrics`, argon DoS, non-idempotent admin actions). | 2 | 3 | **6** | Fix them as listed in SECURITY §11. Add the tests in SECURITY §14. | Backend, Frontend | Open |

---

## 2. Product risks

| ID | Risk | L | I | Score | Mitigation | Owner | Status |
|---|---|---|---|---|---|---|---|
| PR-01 | **Guest progress loss.** A guest's only credential is the token in `localStorage`. Clearing storage, a private window, switching device or a 30-day absence (TTL) loses the account permanently. On Yandex this conflicts with YG-07 ("progress must not be lost"). | 3 | 3 | **9** | Bind guests to a platform identity: the Yandex `getUniqueID` plus a signed payload (YG-25), and an Android device or account id. Prompt guests to upgrade after the tutorial (UI exists in Profile). Consider a sliding TTL for guests. | Backend, Release/Platform | Open |
| PR-02 | **No art assets.** Vector glyphs and a flat canvas map may not meet store expectations, or Yandex moderation's view of quality and a "complete game" (YG-22, GP-12 minimum functionality). | 3 | 3 | **9** | Commission original art (licensed, recorded in `theme.assetsLicense`); store screenshots and promo material; an isometric or illustrated map later (D-019 revisit). | PO, UI/UX | Open |
| PR-03 | **Retention without push and with harsh absence penalties.** Starvation and desertion after long absences (tested at 7 days) can make returning players quit. | 2 | 3 | **6** | Push notifications on Android; a "while you were away" summary screen; tune starvation and desertion (GD-03). | PO, GD | Open |
| PR-04 | **Empty worlds at launch.** A persistent MMO needs a critical mass; the map and market are dead without players. | 3 | 2 | **6** | Fewer, faster worlds (the speed multiplier exists); NPC lairs as PvE content; scheduled world openings; cross-promotion. | PO | Open |
| PR-05 | **Only English and Russian.** Yandex needs auto-detection with fallbacks (YG-13). Only 8 server error codes are localized, so Russian users see English errors. | 2 | 2 | 4 | Localize every error code; map be/kk/uk to ru; translate pack texts. | UI/UX | Partial |
| PR-06 | **Accessibility gaps.** No keyboard tile selection, no focus trap, partial tab pattern (UI_UX_SPEC §9). | 2 | 2 | 4 | Fix gaps 7–10; add axe and keyboard E2E checks. | UI/UX, Frontend | Open |
| PR-07 | **Moderation workload.** Chat reports have no tooling (S-8); a single operator cannot moderate at scale. | 2 | 2 | 4 | Admin moderation UI (delete, mute), reporter and reason fields, a profanity filter, a community policy. | PO, Backend | Open |
| PR-08 | **Monetization absent by design.** No revenue path in v0.1. Adding payments later triggers platform reviews (YG-10, GP-04, GP-11). | 2 | 2 | 4 | Intentional (AGENT.md: no real-money features without review). Plan billing per PLATFORM_REQUIREMENTS §3.6 before beta. | PO | Accepted |
| PR-09 | **Void Frontier is a demo**, not a complete game. Shipping it as a second title would risk "duplicate" or "incomplete" rejection. | 2 | 2 | 4 | Ship one title (Ashen Crowns) first; keep Void Frontier internal. | PO | Accepted |

---

## 3. Game-design risks

| ID | Risk | L | I | Score | Mitigation | Owner | Status |
|---|---|---|---|---|---|---|---|
| GD-01 | **Balance is unvalidated.** Most numbers are PROPOSED. Research could not open primary sources (KNOWN_LIMITATIONS), so unit stats, costs and the travel formula are not derived from evidence. | 3 | 2 | **6** | A balance simulator on `internal/rules` (deterministic, already unit-tested); closed-beta telemetry from `resource_ledger` and `stats`; pack versioning plus live migration. | GD | Open |
| GD-02 | **Large players farm small ones.** Snowballing. | 2 | 3 | **6** | Newbie protection; attack caps per target per day; score floor (`target_too_weak`); loot protection; premium not lootable. Tune with telemetry. | GD | Partial |
| GD-03 | **Offline punishment versus pause rules.** The world keeps simulating while the client is paused (PLATFORM R-3); attacks resolve while the player is away. | 3 | 2 | **6** | Async design with no real-time battles; incoming attacks visible early (missions screen, notifications); defensive buildings; consider a vacation mode. | GD | Partial |
| GD-04 | **Multi-account abuse.** Unlimited guests feed a main account through the market (within the 10× band), transports and attacks (S-12). | 3 | 2 | **6** | Rate-limit guest creation; transport tax or caps outside an alliance; an abuse detection dashboard over ledger flows; ban tooling (exists). | GD, QA/Sec | Partial |
| GD-05 | **No fog of war.** Full map visibility makes target selection trivial and reduces the value of scouting. | 2 | 2 | 4 | A design decision for v0.1 (KNOWN_LIMITATIONS). Add per-player visibility later if testing shows a problem. | GD | Accepted |
| GD-06 | **Deterministic combat is predictable**: there are no damage ranges, so outcomes can be computed exactly. | 2 | 1 | 2 | Accepted for transparency. The scout and intel layer limits information. | GD | Accepted |
| GD-07 | **Economy inflation.** Sinks may be insufficient. | 2 | 2 | 4 | The NPC exchange fee, market fee, upkeep, storage caps and overflow decay act as sinks. Monitor ledger totals per day. | GD | Partial |

---

## 4. Platform risks (Yandex Games and Google Play)

These come from `PLATFORM_REQUIREMENTS.md`. Requirement IDs are shown in brackets.

| ID | Risk | L | I | Score | Mitigation | Owner | Status |
|---|---|---|---|---|---|---|---|
| PL-01 | **Yandex rejects mandatory login** (YG-08, R-1). | 1 | 3 | 3 | Guest play is the default (`/auth/guest`, "Play as guest"); the e2e test plays as a guest. No email/password prompt should be shown in the Yandex build. **Gap:** the auth screen currently defaults to "Create account"; the Yandex build should start straight as a guest. | Frontend, Release/Platform | Partial |
| PL-02 | **Yandex identity cannot be trusted or bound** (YG-25, R-2). The signature algorithm and key availability are UNVERIFIED; the server returns 501 for `platform=yandex`. | 3 | 3 | **9** | Implement signature verification once it has been confirmed against the live docs; bind `uniqueId`; require a signed identity before purchases or a merge. Until then guests are not portable (PR-01). | Backend, Release/Platform | Open |
| PL-03 | **The CSP host is not approved, or assets load from external hosts** (YG-15, YG-16, R-5). Pack fonts load from `fonts.googleapis.com` at runtime. | 2 | 3 | **6** | Self-host the fonts in the archive; one https API host on 443; a retry UI; fallback to the iframe route. | Frontend, Release/Platform | Open |
| PL-04 | **Pause rules** (YG-04, YG-05, R-3). | 2 | 2 | 4 | `onPause` stops polling and calls `gameplayStop`; Yandex `game_api_pause` is wired; there is no audio yet. Test in draft mode. | Frontend | Partial |
| PL-05 | **`LoadingAPI.ready` timing and GameplayAPI markers** (YG-02, YG-03). | 2 | 2 | 4 | `signalReady()` is called after session restore. Markers are sent on pause and resume only; extend them to menus and modals or drop them. Verify in the debug panel. | Frontend | Partial |
| PL-06 | **Yandex rejects external links or a missing connection-error UI** (YG-17, YG-15). | 2 | 2 | 4 | Build flag to strip links (none exist today); a dedicated offline/retry screen (UI_UX_SPEC gap 16). | Frontend | Partial |
| PL-07 | **Cloud saves may be mandatory for games with in-app purchases** (YG-09, R-4). | 2 | 2 | 4 | No IAP in v0.1. If added, mirror a binding record via `setData` and get written confirmation from Yandex support. | Release/Platform | Accepted (no IAP) |
| PL-08 | **Server outage during moderation** (R-7). | 2 | 3 | **6** | A dedicated, stable moderation environment; no breaking deploys while a build is in review; monitoring (TR-12). | Release/Platform | Open |
| PL-09 | **Google Play "webview spam" / minimum functionality** (GP-12). | 2 | 3 | **6** | Bundle all assets in the AAB (no `server.url`); native back button; offline screen; haptics or push; real art (PR-02). | Release/Platform, Frontend | Open |
| PL-10 | **Google Play account deletion** (GP-10). The in-app path exists (`DELETE /me`); the **web deletion link does not**. | 3 | 3 | **9** | Build a public deletion-request page, a grace period and a retention policy (SECURITY §9). | Backend, PO | Partial |
| PL-11 | **Google Play UGC requirements** (GP-16): report and block exist; terms of use, moderation tooling and a profanity filter are missing (S-8). | 3 | 3 | **9** | Write the terms and community policy; build the admin moderation actions; add the filter; make report and block clearly labelled (done). | PO, Backend | Partial |
| PL-12 | **Data safety form and privacy policy** (GP-08, GP-09) do not exist. The Google Fonts call and IP handling must be declared. | 3 | 3 | **9** | Owner writes and hosts the privacy policy; data inventory (SECURITY §9); self-host fonts. | Legal / owner | Open |
| PL-13 | **Closed testing with 12 testers for 14 days** for new personal developer accounts (GP-13, UNVERIFIED). This delays production access by at least 3 weeks. | 3 | 2 | **6** | Start the closed track early; recruit 15 or more testers; consider an organization account. | PO, Release/Platform | Open |
| PL-14 | **Target API 36, AAB, Play App Signing, 16 KB page size** (GP-01 to GP-03, GP-05). | 1 | 3 | 3 | Capacitor 8 defaults to 36/36; AAB is the default; enroll in signing; audit plugins for `.so` files. | Release/Platform | Partial |
| PL-15 | **Android back button and edge-to-edge** (GP-06). Not handled; modals do not close on back. | 3 | 1 | 3 | `@capacitor/app` `backButton` closes the top modal, then navigates; `appCategory=game`. | Frontend | Open |
| PL-16 | **CORS / origin for Yandex and Android** (R-6, GP-18). Yandex origins are not in the default allow-list. | 2 | 2 | 4 | Capture the draft-mode origins; set `RF_ALLOWED_ORIGINS` per environment (`https://localhost` is already allowed for Capacitor). | Release/Platform | Partial |
| PL-17 | **Requirement drift.** Several YG and GP rows are `[S]` or UNVERIFIED (search excerpts only). | 3 | 2 | **6** | Re-verify against the live pages within 2 weeks before submission (R-11). | Release/Platform | Open |
| PL-18 | **Interactive AI ban on Yandex** (YG-19). | 1 | 2 | 2 | The game contains no generative-AI features. | PO | Mitigated |

---

## 5. Legal and IP risks

| ID | Risk | L | I | Score | Mitigation | Owner | Status |
|---|---|---|---|---|---|---|---|
| LG-01 | **Too close to the reference game.** *My Lands* (*My Lands: Black Gem Hunting*) is studied in `docs/research/`. The risk is that names, art, text or trade dress are copied. | 2 | 3 | **6** | **No My Lands assets, names or texts are used.** Pack content is original (`packs/*/pack.json` `license`: "original … CC BY 4.0"). Icons are original SVG drawn in code (`assetsLicense`). Races, buildings and units have new names. Game mechanics as such are generally not copyrightable, but avoid UI layout and trade-dress imitation. Add a pre-release review that greps for reference-game names. | GD, Legal / owner | Partial |
| LG-02 | **Title and trademark clearance not done** for "Realmforge" (engine and web title), "Ashen Crowns" (Android `appName`, pack title) and `com.realmforge.ashencrowns` (the appId is **permanent** once published). | 3 | 3 | **9** | Trademark search (EUIPO, USPTO, Rospatent/WIPO) and store and catalogue searches (YG-18) before first upload. Decide the final appId. | Legal / owner | Open |
| LG-03 | **Font licensing.** Pack fonts are Google Fonts under the SIL Open Font License, which allows bundling. They are currently loaded remotely. | 1 | 1 | 1 | When self-hosting, ship the OFL licence text with the fonts and do not sell the fonts on their own. Record the families in `assetsLicense`. | Frontend | Mitigated (license); see PL-03 |
| LG-04 | **Russian personal-data law (152-FZ, data localisation).** Yandex Games has a predominantly Russian audience, and the law reportedly requires that personal data of Russian citizens be stored first in databases located in Russia. Not legally verified. | 2 | 3 | **6** | Get legal advice before the Yandex launch. Options: host the primary database in Russia for the Yandex deployment, or minimize personal data (no email; guest ids only). | Legal / owner | Open |
| LG-05 | **GDPR / privacy compliance for EU players** (Google Play). No privacy policy, no data export, retention after deletion (S-8). | 2 | 3 | **6** | Privacy policy; data inventory; a deletion grace period and purge; a data-export endpoint; a data processing record. | Legal / owner, Backend | Open |
| LG-06 | **Pack content licensed CC BY 4.0** lets anyone reuse the game texts and data commercially with attribution. | 2 | 1 | 2 | Confirm this is intended; otherwise relicense before publishing (all rights reserved for the commercial title). | PO, Legal / owner | Open |
| LG-07 | **Research documents quote third-party sources** (`docs/research/_sources_*.md`). Fine for internal use; must not ship in the build or the store listing. | 1 | 1 | 1 | Keep `docs/` out of the client bundle (Vite only bundles `src`). | Release/Platform | Mitigated |
| LG-08 | **Age rating, chat with minors.** Chat plus an MMO raises target-audience obligations (GP-14, GP-15). | 2 | 2 | 4 | Target 13+ or 16+ (not children); answer IARC honestly; moderation tooling (PL-11). | PO | Open |

---

## 6. Top risks (score ≥ 9) to act on first

1. **PR-01 and PL-02:** guest progress portability and Yandex identity binding.
2. **PR-02:** original art assets.
3. **PL-10, PL-11 and PL-12:** Google Play deletion link, UGC moderation, privacy policy and Data safety.
4. **LG-02:** title, trademark and appId clearance.

## 7. Change log

| Date | Change | By |
|---|---|---|
| 2026-10-31 | Initial register created from the code review, PLATFORM_REQUIREMENTS, SECURITY, TEST_STRATEGY and UI_UX_SPEC | Agent F, Agent E |
