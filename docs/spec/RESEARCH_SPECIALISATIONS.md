# Research and specialisations

Updated: 2026-10-31

## Research sources

- My Lands science rules: https://mlgame.ru/wiki-ru/Wiki.jsp?page=Science
- My Lands specialisations: https://mlgame.ru/wiki-ru/Wiki.jsp?page=Spec
- English confirmation: https://us.mlgame.org/wiki-us/Wiki.jsp?page=Spec

Use these sources for gameplay research.
Do not reuse the original game's artwork or interface files.

## Verified My Lands rules

1. A player can hold one specialisation.
2. Starting a different specialisation cancels the previous one.
3. Specialisations have three levels.
4. Each level changes its bonuses.
5. The specialisation badge appears in the player's domains.
6. Sciences use scientist production from laboratories.
7. Science production can receive several location-dependent bonuses.
8. Some science completions require player interaction.

## Ashen Crowns implementation

The Game Pack defines three original Paths.
Their bonuses are server-side modifiers.
The server applies these modifiers to gameplay calculations.

| Path | Level I | Level II | Level III |
|---|---|---|---|
| Architect | -15% build time | -30% build time | -45% build time |
| Steward | +6% production | +12% production | +18% production |
| Warlord | +6% attack, -8% train time | +12%, -16% | +18%, -24% |

These values are original Ashen Crowns balancing choices.
They do not reproduce My Lands values.

A Path requires Keep level five.
A Path has three research levels.
Changing Paths removes previous bonuses when research begins.
The new Path activates after the research job completes.
Cancelled research does not restore the previous Path.
Research deductions use the existing resource ledger.
The game uses the existing authoritative server clock.

## Interface contract

1. Show the current Path beside the hero shortcut.
2. Show the Path level and current modifiers in its tooltip.
3. Open the research screen when the Path badge is selected.
4. Show technologies grouped by branch.
5. Show separate visual nodes for research levels.
6. Show completed, available, queued, and locked states.
7. Show costs, prerequisites, and effects for selected research.
8. Confirm a change that cancels the current Path.
9. Respect reduced-motion preferences.
10. Support Russian and English labels.

## Differences requiring further work

- My Lands produces science points through laboratories.
- Ashen Crowns currently spends resources and queues timed jobs.
- My Lands has more specialisations than this Game Pack.
- Laboratory production and science-point storage are not implemented.
- Some My Lands science interactions are not implemented.
- Exact My Lands numerical balance is not adopted.
- No original My Lands sprites or user interface files are included.

## Acceptance criteria

- A player never receives bonuses from two Paths.
- A new Path removes the previous Path at research start.
- All Path modifiers affect server-side calculations.
- Cancelled research cannot restore removed bonuses.
- Both locales describe the consequences of switching.
- Research progression survives reload and reconnection.
- The research tree remains usable with reduced motion.

## Verification status

The code and Game Pack were inspected through GitHub.
The full database integration test suite has not been verified.
Do not mark this design complete until the listed tests pass.

## Reference research families

Source: https://mlgame.ru/wiki-ru/Wiki.jsp?page=ScienceDescription

The original game groups technologies by their function.
The Ashen Crowns technology graph requires a similar breadth.

| Reference family | Current Ashen Crowns status | Required work |
|---|---|---|
| Basic city permissions | Partial | Validate gates, expansion, and multiple construction slots |
| Military units | Partial | Connect level upgrades to unit characteristics |
| City defence | Partial | Connect fortifications and protection to technology |
| Building unlocks | Partial | Audit every construction prerequisite |
| Economy upgrades | Partial | Verify production and storage modifiers |
| Hero sciences | Partial | Enforce hero limits and level caps |
| Player-intervention technology | Missing | Add a confirmation and settlement workflow |
| Global science production | Missing | Add shared science points across all cities |
| Change research focus | Missing | Add preserved progress and focus switching |
| Specialisation | Implemented for three Paths | Test switching, bonuses, and per-level presentation |

Research mechanics must remain configurable through Game Packs.
Do not hard-code the original game's technology names.
Do not insert unverified original game costs.

## Specialisation decisions

The three existing Paths remain the initial implemented set.
The game needs no hard-coded specialization ID in server rules.
Gameplay modifiers are defined by each Path's effects.
The interface can identify a Path by its exclusive group.

Research mode changes require an explicit Game Pack version.
Existing worlds retain their pinned rules until a migration is approved.
