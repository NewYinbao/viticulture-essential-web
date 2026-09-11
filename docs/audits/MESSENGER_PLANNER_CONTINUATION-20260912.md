# Messenger / Traveler / Planner continuation patch

## Deliverables

- `messenger-planner.patch` is a minimal patch generated against the current production v5 tree on 2026-09-12. Production source files were not edited.
- `messenger-browser-result.json` records the real Microsoft Edge run.
- SHA-256 for `messenger-planner.patch`: `D8A13ADE0823E962DD01D0E2D9779EF5F1BB842BCCC792E589FA1167D2776FB5`.

## Implemented behavior

- EE Traveler may reserve an open summer slot while placed in winter. Tuscany keeps its ordered future-season rule.
- EE Messenger may reserve a winter slot from summer. Tuscany Messenger reservations also work in later seasons.
- A future Messenger reservation stores only the worker position, action space, slot, and applicable slot identity. It does not preselect or consume grapes, cards, coins, structures, or rewards.
- At the player's first turn in the target season, Messenger creates a persisted `messenger` continuation. The action panel asks for inputs from the player's current state and executes through the existing atomic action transaction.
- An impossible Messenger action removes the reservation, grants no slot reward, and ends that turn. A successful action receives the actual reserved slot reward.
- EE season advancement and Tuscany's Planner queue both activate due Messenger reservations. Messenger runs after due Planner entries when both exist.
- Rhine Son's first-winter harvest can enable a previously impossible winemaking Messenger reservation. The harvest choice and restored Messenger continuation remain persisted and consume no extra turn.
- Planner feasibility now covers Tuscany planting, building, structure drawing, visitor plays, training, trading, influence, field flipping, and build/tour actions. Summer Visitor 29 considers ready special and gray workers as well as regular and Grande workers. Training probes the real selectable worker pool, including Grande.
- A failed Planner reservation keeps every worker type placed/used. No dedicated Planner-special-worker failure ruling was found; this is a combined-rules inference from Planner placing a worker and special workers acting like regular workers in other respects.
- The training action continuation exposes `trainingWorker`, so a Planner-triggered training action can select the worker it actually trains.

The separate placement-trigger timing for Farmer, Professore, Innkeeper, Soldato, Mafioso, Politico, and Merchant is intentionally handed to the follow-on worker using this TEMP as its baseline. The primary rule says a special-worker ability that occurs "when you place" triggers before anything else; that follow-on change is not represented in this patch.

## Primary evidence used

- `artifacts/expansions/evidence/tuscany-essential.txt`, pages 5-6: special-worker placement timing and worker-specific text.
- The same evidence file at line 274 onward: Traveler may use hidden player-count slots in the preceding season; Messenger takes the action at the beginning of the target season without placing another worker; an impossible action receives no bonus and the turn ends.
- Essential Edition Planner clarification: place a worker on an action in a future season and take that action at the beginning of that season.

## Verification

The patch was applied to a fresh copy of the current production tree at `C:\Users\admin\AppData\Local\Temp\viticulture-messenger-verify-20260912` before verification.

- `git apply --check --whitespace=error messenger-planner.patch`: passed.
- `go test ./internal/game -run Messenger -count=1`: passed.
- `go test ./internal/game -run EETraveler -count=1`: passed.
- `go test ./internal/game -run Planner -count=1`: passed.
- `go test ./internal/game -run RhineSonHarvest -count=1`: passed.
- `go test ./internal/server -count=1`: passed.
- JavaScript syntax checks for `action-panel.js`, `ee-ui.js`, and `visitor-ui.js`: passed.
- `node tests/tuscany-ui-unit-test.cjs`: 5/5 passed.
- Real Windows Microsoft Edge, headless: passed summer reservation without wine inputs, server-process restart and save restore, winter first-turn continuation, and execution using the live grape state; no page errors.

The standalone Edge runner is `tests/e2e/messenger-browser-test.cjs`. It accepts:

- `MESSENGER_BROWSER_FIXTURES`: fixture JSON produced by `go test ./internal/game -run TestMessengerBrowserFixtureExport`, with this variable set to the desired output file.
- `MESSENGER_BROWSER_DATA`: temporary server data directory.
- `MESSENGER_BROWSER_OUTPUT`: report directory.
- `MESSENGER_HTTP_EXE`: temporary Windows server executable.
- `VITICULTURE_SOURCE_ROOT`: optional patched source root; defaults to the repository root inferred from the script.

The full `internal/game` suite in the original isolated TEMP also exposed two pre-existing failures outside this patch: `TestTrussHarvestMachineAgesActualFieldCount` (no field selected) and `TestEmptyPrivateHandRemainsJSONArrayAfterRestore` (`nil` versus empty JSON array). Targeted coverage and the fresh integration checks above pass.
