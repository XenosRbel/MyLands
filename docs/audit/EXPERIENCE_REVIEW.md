# Game experience audit — 2026-10-31

## Confirmed implementation facts
- The client uses React 19, Vite 8, and TypeScript 7.
- The shell includes thirteen gameplay destinations.
- The client supports Russian and English string tables.
- The app uses a configurable Game Pack for visual tokens.
- Game state and commands flow through the existing game context.
- The repository contains Vitest and an E2E script.

## High-priority findings
1. **P0 — Game feel.** The main shell uses a conventional application sidebar and generic panels. Build a game HUD, illustrated screen headers, and world-first navigation.
2. **P0 — Art coverage.** The documented asset gap includes missing unit portraits, neutral sites, and emblems. Finish art coverage before visual sign-off.
3. **P0 — Exploitable mechanics.** The documented market escrow shelters stock from raids. Restrict or account for this mechanic.
4. **P0 — Economy abuse.** Transports can target arbitrary players. Add anti-farming rules and audit events.
5. **P1 — Missing gameplay.** Combat omits healing and hero spells. City construction omits placement on cells.
6. **P1 — Map presentation.** The square grid lacks fog of war. Define visibility and discovery states.
7. **P1 — Localization.** Two UI locales exist. Add semantic review and in-context screenshots.
8. **P1 — Gameplay feedback.** Add clear arrival, construction, research, and reward effects.
9. **P1 — Platform.** Complete Android and platform identity checks before publishing.
10. **P2 — Accessibility.** Keep reduced motion, keyboard navigation, contrast, and sound opt-in.

## Completed in this branch
- Added optional generated UI sounds. Default playback remains muted.
- Added theme-aware panel depth and subtle navigation transitions.
- Added motion reduction at the operating-system and game levels.
- Added development instructions for Codex and Claude Code.

## Not completed
- No full visual remake or all-screen playthrough.
- No runtime screenshot comparison or performance validation.
- No executed database integration tests in this connector session.
- No confirmed source-wide localization quality or balance verification.

## Acceptance criteria for the next implementation stage
- Every enabled screen receives a visual reference and mobile layout.
- Browser screenshots cover idle, loading, error, empty, and success states.
- A complete onboarding path works without developer intervention.
- Every spend and reward survives retry and concurrent requests.
- Resource-ledger totals match the authoritative player state.
- Both locales pass key parity and manual context checks.
- Game audio uses permission-safe playback and a persistent mute preference.
- FPS, memory, and bundle budgets are recorded for mobile.
