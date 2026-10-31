# MyLands recovery plan

Date: 2026-10-31
Status: Active
Scope: Ashen Crowns game, common engine, Yandex Games, Android.

## Delivery rules

1. Fix confirmed defects in code.
2. Add regression coverage for each defect.
3. Run checks before marking tasks complete.
4. Keep changes within the existing pull request.
5. Keep unverified claims separate from tested results.
6. Keep user instructions within 20 words per sentence.
7. Preserve server authority for all game actions.
8. Preserve Game Pack independence.

## P0: Restore repeatable builds

- [x] Copy Game Packs into the Docker client stage.
- [x] Add asset-path validation to the client build.
- [x] Add a Docker build job to GitHub Actions.
- [x] Add server Game Pack validation to the CI quality gate.
- [x] Remove invalid faction artwork keys from Game Pack maps.
- [x] Verify the production Docker build in GitHub Actions.
- [ ] Verify the production Docker build on a local workstation.
- [ ] Verify local startup through Docker Compose.
- [ ] Verify the container serves the client, packs, and API.
- [ ] Verify the database starts from an empty volume.
- [ ] Verify the initial player journey with browser E2E.

Acceptance: production Docker build and first-login browser tests pass.

## P0: Secure multiplayer economy

- [x] Restrict transports to owned or allied cities.
- [x] Add a regression test for unrelated transports.
- [x] Release market escrow into the city before successful raid plunder.
- [x] Add a regression test for released escrow orders.
- [ ] Confirm escrow release and plunder with database integration tests.
- [ ] Verify ledger integrity across place, accept, cancel, and plunder.
- [ ] Add concurrency tests for battles and market orders.
- [ ] Review cross-account transfers and alliance farming.
- [ ] Review privileged operations and audit records.

Acceptance: exploits cannot bypass resource availability or ownership checks.

## P1: Research and specialisation

- [x] Build the tiered research interface.
- [x] Add per-tier research icon variation.
- [x] Display active specialisation bonuses.
- [x] Permit exclusive Path changes with confirmation.
- [x] Test server-side Path modifier switching.
- [x] Document official science rules and current differences.
- [ ] Implement player-wide science production and progress.
- [ ] Add laboratory-specific production bonuses.
- [ ] Add manual confirmation for level upgrades.
- [ ] Migrate existing research jobs without resource loss.
- [ ] Run database integration tests for research switching.
- [ ] Validate all research states across desktop and mobile.

See `docs/spec/RESEARCH_SPECIALISATIONS.md`.
See `docs/spec/SCIENCE_ENGINE_MIGRATION.md`.

## P1: Art and visual game design

- [x] Resolve required artwork paths for all 82 gameplay entities.
- [x] Add original art for missing units and neutral locations.
- [x] Add faction and race emblems.
- [ ] Replace placeholder vector images with consistent painted sprites.
- [ ] Create transparent nine-slice UI panels and buttons.
- [ ] Make terrain textures tile without visible seams.
- [x] Restyle desktop navigation and game HUD with muted medieval surfaces.
- [x] Add browser screenshots for all primary screens.
- [ ] Review and approve final browser screenshots after all visual fixes.
- [x] Apply consistent frames, buttons, panels, typography, cards, and modals across existing game screens.
- [x] Add illustrated army-card heraldry and hero equipment styling.
- [x] Add original scalable panel and parchment texture assets.
- [x] Rework city, army, hero, and map layouts around illustrated game content.
- [x] Add a bottom command dock with actual army portraits.
- [x] Add tabbed research branches and restore readable parchment details.
- [x] Add illustrated race selection and a subdued game entry background.
- [x] Restyle shared modal dialogs as light parchment windows.
- [ ] Complete full-screen illustration and artwork for every gameplay state.
- [ ] Add building-level variants and construction effects.
- [x] Extend the E2E screenshot checklist to missions and reports.
- [x] Run desktop and mobile screenshots in CI.
- [ ] Complete final visual review after recent changes.
- [x] Compare the city and research screenshots with public My Lands references.
- [ ] Complete per-screen reference comparisons.
- [ ] Keep original game brands and third-party assets out of the product.

Acceptance: every screen uses complete art and consistent visual language.

## P1: Player feedback, motion, and sound

- [x] Add opt-in UI cues and persistent sound preference.
- [x] Add ambient depth and motion-reduction support.
- [ ] Add original sound assets for construction and missions.
- [ ] Add progress, combat, and reward effects.
- [ ] Persist music and effect volume independently.
- [ ] Confirm browser autoplay constraints.
- [ ] Measure FPS, memory, and load time on mobile devices.

Acceptance: feedback remains readable without sound or motion.

## P1: Game rules and balance

- [ ] Compare implemented mechanics with the GDD and rule catalogue.
- [ ] Audit combat inputs, determinism, casualty accounting, and rewards.
- [ ] Audit every resource source, sink, cap, and protection effect.
- [ ] Add tests for construction, research, progression, and cooldowns.
- [ ] Document differences from My Lands without claiming equivalence.
- [ ] Evaluate missing placement, fog of war, and hero abilities.
- [ ] Validate early-game progression with simulation.

Acceptance: documented rules match tested behavior and a playable progression.

## P1: Localization and documentation

- [x] Add Russian and English sound labels.
- [x] Add initial localization tests.
- [x] Add player and developer guides.
- [x] Add project instructions for Codex and Claude Code.
- [ ] Find all hard-coded player-visible strings.
- [ ] Localize unit effects, notifications, errors, and date formats.
- [ ] Review every Russian and English screen in context.
- [ ] Add locale parity and placeholder verification.
- [ ] Apply the project ASD-STE100 writing profile to documentation.

Acceptance: both locales cover every user-visible state.

## P2: Platform and operations

- [ ] Implement Yandex platform authentication checks.
- [ ] Complete Android project generation and build.
- [ ] Validate app lifecycle, offline recovery, and store requirements.
- [ ] Replace instance-local rate limits where deployment needs scaling.
- [ ] Add restore tests and database recovery procedures.
- [ ] Run load tests against production-like configuration.
- [ ] Review original assets and dependency licenses.

Acceptance: deployment checklists and platform test cases pass.

## Verification record

- Confirmed via GitHub: missing Docker copy caused asset-check ENOENT.
- Confirmed via GitHub: 82 required entities resolve to committed art paths.
- Fixed: faction names were incorrectly added as entities in theme art maps.
- Not run locally: Docker build, browser E2E, database concurrency tests.
- CI Docker builds passed on prior branch commits.
- Full shared visual theme, original texture assets, and army redesign were added.
- Browser screenshot workflow completed successfully on commit 2d0b76a.
- Latest visual commits still need passing CI and refreshed screenshots.
- See docs/spec/VISUAL_REDESIGN.md for observed defects and implemented fixes.

## Current rule

Do not mark the project complete while any P0 task remains open.

## Screenshot-driven HUD correction — 2026-11-01

- [x] Remove global game-content width restriction for desktop.
- [x] Expand settlement scenery to full viewport height.
- [x] Replace permanently visible building dashboard with an on-demand drawer.
- [x] Anchor the bottom dock to the desktop viewport.
- [x] Rework bottom actions into framed round controls.
- [x] Keep hero and specialization controls in the top HUD.
- [x] Collapse the queue by default and allow reopening it.
- [x] Show empty troop slots as desaturated unit artwork without fabricating counts.
- [x] Rework the top resource strip to behave as continuous HUD information.
- [x] Expand the world map's main area and right-side inspector.
- [x] Add 1360x860 and 1680x900 layout regression coverage.
- [ ] Inspect the final 1680x900 screenshot with queue collapsed.
- [ ] Replace repeated ground textures with a seamless world environment.
- [ ] Approve HUD layout against public reference screenshots.
- [ ] Verify every screen under the final wide-screen changes.

Source screenshots: docs/research/VISUAL_REFERENCES.md
Detailed acceptance: docs/spec/VISUAL_REDESIGN.md
