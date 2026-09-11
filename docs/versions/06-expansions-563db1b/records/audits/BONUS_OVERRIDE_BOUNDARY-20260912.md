> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# BonusOverride public-boundary fix

## Finding

`Action.BonusOverride` is part of the HTTP JSON shape, while placement code treats a non-empty value as trusted server continuation state. Before this patch, a client could submit the field directly through `Room.Apply` and claim a printed or special-worker bonus that its occupied slot did not grant.

The pre-fix regression test reproduced all four cases:

- A two-player EE tour on slot 1 gained 3 coins instead of 2.
- A two-player EE training action paid 3 coins instead of 4.
- A Messenger continuation on non-bonus training slot 2 paid 3 coins instead of 4.
- A two-player Tuscany tour on non-bonus slot 1 gained 3 coins instead of 2.

Each action was decoded from JSON before being passed to `Room.Apply`, covering the same exported action field used by the HTTP handler.

## Fix

`Room.Apply` now clears `a.BonusOverride` before taking its transactional snapshot and calling `applyUnsafe`. This is the common public boundary used by the HTTP server and direct callers.

No production internal code calls `Room.Apply` or `applyUnsafe` with a legitimate override. Farmer/Planner and Rhine Virtuoso produce `BonusOverride` only after the boundary from persisted server state, and persisted `ActionContext.PendingAction` is left unchanged.

## Verification

- Pre-fix `TestApplyIgnoresClientBonusOverride`: failed in all four subtests with the unauthorized rewards above.
- Post-fix `TestApplyIgnoresClientBonusOverride`: passed.
- Planner Oracle persisted-continuation regression: passed.
- Messenger live-resource continuation regression: passed.
- `go test ./... -count=1`: passed for every package.
- `git apply --check --whitespace=error bonus-override-boundary.patch`: passed against the current production tree.

Patch SHA-256: `FEE1B2612D04DB39A21BAE9480AB1469514E85B510F912AD346358E49C44281B`.

根任务补充 HTTP 回归：`internal/server/bonus_override_http_test.go` 使用真实 `httptest.NewServer` 网络请求，两种主板的非奖励格分别提交伪造奖励，核对合法结果落盘及过时请求拒绝。移除修复行的自有 TEMP 副本会稳定失败（EE 3币而非2、Tuscany 2订单而非1）；当前源码两项通过。
