# Roadmap — next increments

Stage plan with exit criteria: [`spec/ROADMAP.md`](spec/ROADMAP.md). Status: [`IMPLEMENTATION_STATUS.md`](IMPLEMENTATION_STATUS.md).

## Priority 1 — before any public test
1. **Re-verify research live** once `*.mlgame.org`, `store.steampowered.com`, `steamcommunity.com`, `yandex.ru`, `support.google.com`, `capacitorjs.com` are allowed: upgrade/downgrade statuses in `MECHANICS_CATALOG.md`, `PLATFORM_REQUIREMENTS.md` `[S]` rows. (task T-020)
2. **Load test** (k6 or Go harness): 1 000 concurrent players, 50 commands/s, worker lag < 2 s, p95 < 150 ms. (T-021)
3. **Balance pass** with a simulation harness over `rules` (first-day progression, PvE difficulty curve, early gold upkeep pressure). (T-022)
4. **Production deployment**: TLS terminator, managed PostgreSQL with PITR, log shipping, metrics dashboard, alerting on `rf_events_overdue`/`rf_events_parked`. (T-023)

## Priority 2 — platform
5. Yandex Games: server-side verification of player identity (signature), cloud-save policy decision vs server saves, CSP host registration, pause/resume QA, moderation checklist. (T-030)
6. Android: generate Capacitor project on a machine with Android SDK, target API 36, AAB signing, data-safety form, in-app account deletion (implemented) + web deletion URL, closed testing track. (T-031)

## Priority 3 — game depth (research-derived, not yet implemented)
7. Healers/resurrection after battle (CMB-009), hero stun/capture/ransom (HERO-007), spells & mana (HERO-006).
8. Joint attack (SOC-014), premium-currency auction (ECO-019/SOC-021), clan warehouse (SOC-011).
9. Cell-based city layout & multiple building instances (BLD-003, D-012) — format v2.
10. Fog of war / visibility, SSE push for incoming attacks, events/tournaments.
11. Art pass (original sprites), sound, onboarding polish.
