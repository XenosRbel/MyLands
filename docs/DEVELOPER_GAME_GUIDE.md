# Developer guide

## Architecture
The Go server owns gameplay state.
The React client renders server responses.
Game Packs define names, art, and balance settings.

## Local checks
1. Run `make packs`.
2. Run `make lint`.
3. Run `make test`.
4. Start the development environment.
5. Run `make e2e`.

## Change a game screen
1. Read `AGENT.md`.
2. Read the target Game Pack.
3. Inspect the target screen.
4. Preserve all existing commands.
5. Add loading, error, empty, and success states.
6. Verify both supported locales.
7. Test keyboard and touch input.
8. Test reduced motion.
9. Update the implementation status.

## Change a game rule
1. Read `docs/spec/GAME_RULES.md`.
2. Add deterministic rule tests.
3. Update the authoritative server implementation.
4. Record every resource change in the ledger.
5. Add retry and concurrency tests.
6. Update the OpenAPI contract.
7. Record the rule identifier in the change summary.

## Skill discovery
Claude Code uses project skills under `.claude/skills/`.
Codex can use project skills under `.agents/skills/`.
The project instructions live in `AGENT.md`.
Follow `docs/STYLE_AND_WRITING.md` when writing instructions.

## Validation status
This guide does not certify the build.
Run the listed checks in a configured development environment.
