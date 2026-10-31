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
- [ ] Verify the production Docker build completes.
- [ ] Verify local startup through Docker Compose.
- [ ] Verify the container serves the client, packs, and API.
- [ ] Verify the database starts from an empty volume.
- [ ] Verify the initial player journey with browser E2E.

Acceptance: production Docker build and first-login browser tests pass.

## P0: Secure multiplayer economy

- [x] Restrict transports to owned or allied cities.
- [x] Add a regression test for unrelated transports.
- [ ] Prevent market escrow from sheltering resources during raids.
- [ ] Verify ledger integrity across place, accept, cancel, and plunder.
- [ ] Add concurrency tests for battles and market orders.
- [ ] Review cross-account transfers and alliance farming.
- [ ] Review privileged operations and audit records.

Acceptance: exploits cannot bypass resource availability or ownership checks.

## P1: Art and visual game design

- [x] Resolve required artwork paths for all 82 gameplay entities.
- [x] Add original art for missing units and neutral locations.
- [x] Add faction and race emblems.
- [ ] Replace placeholder vector images with consistent painted sprites.
- [ ] Create transparent nine-slice UI panels and buttons.
- [ ] Make terrain textures tile without visible seams.
- [ ] Replace administrative navigation with a coherent game HUD.
- [ ] Complete city, map, battles, research, army, heroes, and reports.
- [ ] Add building-level variants and construction effects.
- [ ] Validate desktop and mobile screen captures.
- [ ] Compare each screen against licensed visual references.
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
- Not run locally: Docker build, browser E2E, database concurrency tests.
- CI Docker build job added. No passing run is verified yet.

## Current rule

Do not mark the project complete while any P0 task remains open.
