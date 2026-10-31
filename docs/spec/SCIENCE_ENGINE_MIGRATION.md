# Science engine migration

Status: Proposed architecture. Not implemented.
Updated: 2026-10-31.

## Reference sources

- Science: https://mlgame.ru/wiki-ru/Wiki.jsp?page=Science
- Science catalog: https://mlgame.ru/wiki-ru/Wiki.jsp?page=ScienceDescription
- Specialisations: https://mlgame.ru/wiki-ru/Wiki.jsp?page=Spec

The mechanics below use public rules as design references.
All art, text, names, and interface files remain original.

## Verified reference mechanics

1. Laboratories produce scientific output in multiple domains.
2. Science contributes to one shared player-wide research resource.
3. Location bonuses affect each laboratory before aggregation.
4. Temporary bonuses can modify laboratory output.
5. Quest rewards can increase total scientific output.
6. A player chooses which technology receives scientific progress.
7. Some technology completions await a player decision.
8. Military upgrades can require additional resources.
9. A player can choose one specialisation.
10. Starting a different specialisation cancels the previous one.

## Differences in the current implementation

Ashen Crowns has a resource-based research queue.
Research costs are deducted from one city.
The server uses a clock-based job completion time.
The research system lacks a player-wide science pool.
It lacks per-building scientific output and stored research progress.
It lacks awaiting-confirmation states for technology upgrades.

Do not claim this migration is already complete.

## Proposed authoritative state

Store research state in player-owned tables.

- `player_science`: player_id, balance_milli, carry_milli, updated_at.
- `science_progress`: player_id, tech_id, target_level, progress_milli, state.
- `science_focus`: player_id, tech_id, changed_at.
- `science_events`: player_id, source, quantity_milli, reference, created_at.

States: idle, researching, awaiting_confirmation, completed, cancelled.
Use database constraints for uniqueness and nonnegative balances.
Use idempotency keys for science selection and confirmation commands.

## Production algorithm

1. Lock the player.
2. Settle all owned cities to the same server timestamp.
3. Calculate each laboratory's base production.
4. Apply each laboratory's location bonuses.
5. Apply each laboratory's temporary bonuses.
6. Round each laboratory's production as the configured formula requires.
7. Sum the laboratory outputs.
8. Add player-wide quest bonuses.
9. Deposit production into the player-wide science pool.
10. Apply research progress to the selected technology.
11. Stop at any required player confirmation.
12. Write all debits and credits to a science ledger.

Use fixed-point values.
Persist fractional production carry.
Never use browser time to settle science.
Never allow science balances below zero.
Keep deterministic time-step simulation.

## Research selection

1. Validate the target technology.
2. Check the player's eligibility.
3. Check the selected research level.
4. Store the research focus inside the player transaction.
5. Preserve existing progress according to the Game Pack policy.
6. Return authoritative progress and expected completion time.

Changing research focus must not duplicate progress.
Only one technology can receive new points at one time.
The client must not grant science or complete research.

## Player confirmation

Some upgrades require a separate confirmation command.

1. Pause the selected technology at its completion threshold.
2. Set its state to awaiting_confirmation.
3. Calculate the upgrade cost from authoritative state.
4. Include relevant units across cities, missions, and queues.
5. Accept the player's selected payment city.
6. Revalidate the cost inside the transaction.
7. Deduct resources and write ledger entries.
8. Apply the upgrade exactly once.
9. Resume the research system after confirmation.

The first migration phase should support one generic confirmation type.
Later phases can add unit conversions and city upgrades.

## Specialisations

The existing Ashen Crowns Paths use three levels.
The selected Path modifies actual server calculations.
Starting a different Path removes the previous Path.
The previous bonuses stop when replacement research begins.
The replacement bonuses start after research completes.
A cancelled replacement does not restore the old Path.

A selected Path requires a visible badge.
The interface must explain every active bonus.
The interface must warn before changing Paths.

## Migration safety

1. Add new database tables through append-only migrations.
2. Keep existing research jobs valid during the transition.
3. Add a versioned research mode to the Game Pack.
4. Leave existing worlds on the legacy research mode.
5. Migrate one test world to the new mode.
6. Verify player resource reconciliation.
7. Switch new worlds to the science mode only after testing.
8. Preserve a rollback path for the new mode.

World Game Packs are pinned by hash.
Updating a pack file does not automatically update existing worlds.

## Required tests

- Fixed production across time intervals.
- Multi-city production aggregation.
- Per-laboratory rounding and carry.
- Temporary and quest bonus stacking.
- Progress switching without duplication.
- Offline progression with clock skew.
- Concurrent commands on one player.
- Research confirmation idempotency.
- Military upgrade cost across all player holdings.
- Specialisation switching and cancellation.
- Resource ledger reconciliation.
- Restart, disconnect, and rollback behavior.
- Russian and English interface parity.

## Current delivery gate

Keep the legacy research model active until migration tests pass.
Do not silently change an established world's research economy.
