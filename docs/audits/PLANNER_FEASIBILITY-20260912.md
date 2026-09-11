# Planner feasibility follow-up review

Date: 2026-09-12

## Confirmed defect

`plannerCanExecute` probed future Planner actions after clearing `Action.WorkerType` and constructing a context without the persisted `PlannedWorkerState`. It also omitted the Farmer bonus that `resolvePlanner` restores at actual execution. A Farmer reservation on the nonbonus training slot with the locked `discount` choice was therefore discarded when the player had 3 coins, although actual execution legally trains a regular worker for 3 coins.

The reproduction uses the normal `Apply` flow to reserve the Farmer, selects `discount`, serializes and restores the room, sets the future balance to 3 coins, and starts the winter Planner queue. Before the patch, `TestPlannerFeasibilityUsesLockedFarmerDiscount` failed with:

```
Planner discarded a train action affordable only with the locked Farmer discount
```

## Fix reviewed

The patch factors the Planner execution projection into two helpers shared by `startNextPlanner`, `resolvePlanner`, and the isolated feasibility probe:

- the future action receives the persisted worker identity and only the Farmer bonus locked in `PlannedWorkerState`;
- the probe receives the same Planner context fields as real execution, including `TriggerSeat`, `WorkerType`, and `PlannedWorker`;
- client-provided `BonusOverride` remains cleared before the locked Farmer bonus is applied.

The regression includes a negative boundary: Farmer identity without a persisted discount remains unaffordable at 3 coins. The change does not grant a generic discount to all special workers or infer bonuses from worker type.

The required real-browser reproduction exposed a second layer after the server fix: `publicPlans` correctly hides the private Farmer decision, but the active Planner choice did not project that decision back to its owning player. `visitor-ui.js` therefore priced regular training at 4 coins and marked it unavailable. The follow-up patch puts the locked effective bonus in the actor-only `Choice.SpecialBonus`, keeps the nonactor choice view limited to `playerId` and `kind`, and derives the Planner training price from that effective bonus. No private state is added to the public planned placement.

## Validation

- `git apply --check --whitespace=error planner-feasibility.patch`: passed against a fresh read-only-derived production snapshot.
- `git apply --check --whitespace=error planner-feasibility-browser.patch`: passed against the current production tree after the server patch.
- `go test ./internal/game -run 'Planner|Messenger|SpecialWorker|RhineSonHarvest|EETraveler' -count=1`: passed in the implementation TEMP.
- `/opt/viticulture-toolchain/go/bin/go test ./... -count=1`: passed against a fresh snapshot of the final production tree after applying the browser follow-up patch; `internal/game`, `internal/server`, and `internal/sharing` all passed.
- Microsoft Edge via Playwright 1.63: passed a real HTTP flow from the visible Planner UI through Farmer selection, discount selection, server restart, summer-to-winter transitions, Planner feasibility, regular-worker selection, and execution. The final player balance was 0 coins, total workers increased by exactly 1, the reservation was drained, and `pageErrors` was empty.
- The earlier Planner Farmer Edge flow for a locked card-draw reward also passed after the UI projection change.

Server patch SHA-256: `A2B8694BEF546CBACC9006F0959D459733D4583B586EDAD2E0AB55085B877098`

Browser/UI follow-up patch SHA-256: `7C93826399495AFDDC2F8F91BC1ED70BA4945276DE853EDD37CD86D5B2B88820`

Browser result SHA-256: `2D093DA7B2E9F1F6638DE8EE180BA85C34B76F2257DC4BD2EF7FC3D8180182A1`
