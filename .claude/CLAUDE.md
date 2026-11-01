# Claude Code project instructions

@../.agents/HARNESS.md

Read `AGENT.md` before editing any file.
Read `docs/STYLE_AND_WRITING.md` before writing documentation.
Load the relevant skill from `.claude/skills/` before editing a subsystem.
The server owns resources, timers, battles, and ownership.
Keep Game Pack data separate from game engine rules.
Use one instruction per sentence.
Keep sentences at 20 words or fewer.
Use one canonical term for each concept.
Do not claim tests passed unless they ran.
Run `make lint`, `make test`, and `make e2e` when the environment supports them.
Record missing validation in the change report.

## Raster-only game art

Never introduce new SVG files into visible game artwork.
Never render inline SVG in a game component.
Use original WebP or PNG sprites, not rasterized flat SVG placeholders.
Treat legacy SVG assets as unfinished art debt.
Run `npm run check:art` after changing client visuals.
Review raster contact sheets before approving artwork.
See `docs/art/RASTER_REDESIGN_STAGES.md` for the staged process.
