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
