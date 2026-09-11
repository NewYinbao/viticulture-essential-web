# 最终功能复核报告

## 结论

针对最新冻结版，上一轮确认的两项 P1 缺陷均已独立复核关闭；本轮未发现新的可复现问题。复核仅针对指定修复及其边界，没有修改生产源码、提交或推送。

## 最新冻结对象

- 生产只读工作区：`C:\Users\admin\Games\Viticulture\essential-web`。
- 临时复核副本：`C:\Users\admin\Documents\Codex\2026-09-12\viticulture-final-functional-review-a\work\snapshot-final-fix-20260912-013000`。
- 源码 manifest digest：`91036c232fa80d3697d6f45c380ace1e86fe3e27e5d79f6c6163679230c0e438`；已按 `tests/e2e/runtime.cjs` 的 195 个源文件重新计算匹配。
- 生产 Continuation EXE SHA-256：`513f41b507991384ed93c7de1a8e1c3261847822d0f4de8021e380aeadaacd46`；已复制到临时副本后使用，未覆盖生产 EXE。
- 已阅读既有三份 API/发布/密码续作文档，以及修复审计：`docs/audits/PLANNER_FEASIBILITY-20260912.md`、`docs/audits/BONUS_OVERRIDE_BOUNDARY-20260912.md`。

## 已关闭问题

### P1 — Planner Farmer 折扣预约曾被可行性探测静默丢弃

修复定位（最新冻结副本绝对行号）：

`C:\Users\admin\Documents\Codex\2026-09-12\viticulture-final-functional-review-a\work\snapshot-final-fix-20260912-013000\internal\game\visitor_special.go:334-360` 抽出共享的 Planner context/action/bonus 投影；`:482-489` 的 probe 使用同一投影，保留 `WorkerType`、`TriggerSeat`、`PlannedWorker` 和锁定的 Farmer bonus；`:319-322` 仅向当前 Planner 行动者的 Choice 设置 `SpecialBonus`。

精确回归 `internal/game/planner_farmer_feasibility_test.go:5-52` 通过：没有锁定折扣时 3 金币不可行；有锁定 `discount` 时可行，Planner Choice 对本人显示 `SpecialBonus=discount`，对其他玩家的 view 只有 `playerId` 与 `kind`，不泄露折扣；执行后金币为 0、普通工人增加 1、预约和 Choice 均正确清空。

真实 Microsoft Edge/Playwright HTTP 流程也通过：UI 选择 3 金币 Farmer training discount，服务器重启后预约和选择仍在，冬季探测不再丢弃，最终只训练 1 名普通工人。

### P1 — HTTP 客户端曾可伪造 `BonusOverride` 获取普通行动奖励

修复定位（最新冻结副本绝对行号）：

`C:\Users\admin\Documents\Codex\2026-09-12\viticulture-final-functional-review-a\work\snapshot-final-fix-20260912-013000\internal\game\engine.go:9-18` 在公共 `Room.Apply` 边界清空客户端传入的 `BonusOverride`，随后才进入事务和 `applyUnsafe`；合法 Farmer/Planner、Rhine Virtuoso 等内部 continuation 仍从服务端保存状态重新设置该字段。

精确回归 `internal/game/bonus_override_boundary_test.go:17-76` 通过，覆盖两人 EE 导览、EE 培训、Messenger continuation、Tuscany 非奖励格；`internal/server/bonus_override_http_test.go:12-75` 通过真实 `httptest.NewServer` 覆盖 EE/Tuscany，伪造请求只能产生合法基础奖励并能正确持久化，重放仍返回 409。

## 验证结果

- `/opt/viticulture-toolchain/go/bin/go test ./... -count=1`：通过。
- `/opt/viticulture-toolchain/go/bin/go vet ./...`：通过。
- 三项指定精确用例全部通过：
  - `TestApplyIgnoresClientBonusOverride`：4 个边界子用例通过。
  - `TestHTTPDiscardsForgedBonusAndPersistsOnlyLegalReward`：EE/Tuscany 通过。
  - `TestPlannerFeasibilityUsesLockedFarmerDiscount`：通过。
- `TestPlannerFarmerFeasibilityBrowserFixtureExport`：通过。
- `tests/e2e/planner-farmer-feasibility-browser-test.cjs`：通过；Microsoft Edge headless，4 项检查通过，`pageErrors=[]`。
- 浏览器流程使用 `-addr 127.0.0.1:0` 随机 loopback 端口、临时数据目录和临时输出目录；未读取真实存档或停止既有服务。

## 当前结论与边界

在本轮指定的 BonusOverride 公共边界、Planner Farmer 可行性、UI 价格显示、服务器重启持久化和隐藏 Choice 投影范围内，没有已知未解决的可复现问题。

- 本轮没有重复 24 自然局；root 正在运行完整最终验收。
- 未运行 `go test -race`：本机 WSL 缺少 C 编译器，故不作 race 结论。
- 浏览器验证是指定 Planner 场景，不代表所有访客卡/特殊工人组合穷举；自然局策略也不主动打出访客。
- 未读取或使用真实 `runtime`、`ee-data`，未进行公网 Cloudflare/TLS 部署审计。
