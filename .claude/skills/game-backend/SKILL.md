---
name: game-backend
description: Implement authoritative multiplayer game rules in Go with ledger integrity, deterministic outcomes, and concurrency safety.
---
# Game backend
1. Read `AGENT.md` and `docs/spec/GAME_RULES.md`.
2. Locate pure rules under `server/internal/rules`.
3. Reject invalid client resource values.
4. Apply game changes inside the documented transaction boundary.
5. Lock entities in the documented order.
6. Write all currency changes to the resource ledger.
7. Require idempotency keys for state-changing commands.
8. Test repeated and concurrent requests.
9. Update the OpenAPI contract.
10. Record migrations and rule identifiers.
