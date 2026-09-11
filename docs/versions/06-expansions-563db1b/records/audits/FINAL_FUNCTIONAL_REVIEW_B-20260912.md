> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# Viticulture Essential Web：最终功能复核

日期：2026-09-12
复核对象：`C:\Users\admin\Games\Viticulture\essential-web` 当前完整工作区（含未提交与未跟踪内容）
复核模式：只读；未修改生产源码、未停止既有服务、未覆盖旧 `dist`。

## 结论

上一版复核确认的两项问题已在最终冻结快照中修复，并由本次独立定向复审关闭。最终快照的常规 Go/HTTP 回归、真实 Edge 流程和隐藏信息检查均通过；本次未发现新的 P1/P2 问题。

1. **已关闭 P1：公共 Action JSON 可伪造 `BonusOverride`。** `Room.Apply` 现在忽略客户端值，HTTP 两板落盘/重放回归通过。
2. **已关闭 P2：Planner × Farmer 培训折扣预检丢预约。** Planner 现在与实际执行共用保存的 worker/bonus 投影，且 UI 只向预约拥有者显示折扣价格。

本次还复核了 Planner 特殊工人嵌套选择、Messenger 重启恢复、Moor 结构互动、Rhine 代表性牌效果、回滚与公开视图隔离；未确认新的 P1/P2 问题。但访客卡效果尚未达到逐卡穷举级别，详见覆盖边界。

## 确认问题

### 已关闭 P1 — `BonusOverride` 可由普通 HTTP Action 伪造

**历史影响**：任何已通过身份验证的玩家都可以在普通派工 JSON 中提交 `bonusOverride`，在非奖励格重复抽牌或获得折扣。该问题已修复。

**代码证据**：

- HTTP handler 将请求 JSON 直接解码为 `game.Action`，随后调用 `copy.Apply`：`internal/server/http.go:243-279`。
- `BonusOverride` 是公共 JSON 字段：`internal/game/model.go:152-188`。
- `Room.Apply` 现在在事务开始处清空客户端提交的 `BonusOverride`：`internal/game/engine.go:9-18`。
- Planner 等内部续步仍由服务端从已保存状态规范化恢复：`internal/game/visitor_special.go:347-355`。

**关闭验证**（最终快照，生产源码只读）：

- `TestApplyIgnoresClientBonusOverride`：EE tour、EE training、Messenger continuation、Tuscany 非奖励格均通过。
- `TestHTTPDiscardsForgedBonusAndPersistsOnlyLegalReward`：EE/Tuscany 两板真实 HTTP 请求、存档读取和重复提交均通过。
- 直接复现请求现在只执行合法基础行动；伪造值不会增加奖励。

**关闭条件**：已满足公共边界清除、内部续步服务端重建、EE/Tuscany HTTP 落盘与重放回归；没有把私有奖励状态放进普通公开行动投影。

### 已关闭 P2 — Planner × Farmer 培训折扣被预检丢弃

**历史影响**：玩家在夏季用 Farmer 预约 Tuscany 冬季 `train` 非奖励格、选择 `discount` 且只有 3 金币时，预约曾被静默移除。该问题已修复。

**代码证据**：

- 进入执行季节时，`startNextPlanner` 在 `plannerCanExecute` 与 Rhine 例外均失败时直接删除预约：`internal/game/visitor_special.go:304-317`。
- `plannerCanExecute`、Planner context 和实际执行现在共用规范化投影：`internal/game/visitor_special.go:333-355`、`:477-485`。
- 预约拥有者的 Planner choice 只额外显示 `SpecialBonus=discount`；其他玩家仍只看到 `playerId/kind`，不泄露 `PlannedWorkerState`。

**关闭验证**（最终快照，独立 TEMP EXE/数据/端口）：

- 使用现有 `plannerSpecialFixture("farmer")`，将玩家金币设为 3。
- 预约 `farmer → train(slot 2)`，当场选择 `discount`；确认 `Planned[0].SpecialWorker.FarmerBonus == "discount"`。
- `TestPlannerFeasibilityUsesLockedFarmerDiscount` 通过；同时验证没有锁定折扣时 3 金币仍不可行。
- Edge 实际流程通过：当场选 Farmer discount、重启服务、推进到冬季、拥有者看到 3 金币培训控件、执行后金币为 0、普通工人数恰增 1、预约清空。
- Edge 结果：`status=passed`，四项检查通过，`pageErrors=[]`。

**关闭条件**：已共享 Planner action/context 投影，并补齐 3 金币 Go/HTTP/Edge 回归与 actor-only UI 价格投影。

## 已验证通过的关键路径

- 最终冻结快照逐文件 SHA-256 完全匹配；总 digest：`91036c232fa80d3697d6f45c380ace1e86fe3e27e5d79f6c6163679230c0e438`。最终 `dist/Viticulture-Continuation.exe` SHA-256：`513f41b507991384ed93c7de1a8e1c3261847822d0f4de8021e380aeadaacd46`。
- `go test ./... -count=1`：通过。
- `go vet ./...`：通过。
- `go test -tags=ee_rule_audit,audit_diff ./... -count=1`：通过。
- `npm run test:units`：15/15 通过。
- 最终完整浏览器验收已启动并已完成 fixture export；本次报告不把尚在运行的全量 24 局结果当作定向修复关闭依据。
- 独立临时服务与 Microsoft Edge 实测：Planner Farmer、Messenger/Traveler、Moor 互动、Rhine HTTP/浏览器代表路径均通过；重启恢复路径通过，已检查的页面错误为空。
- Planner 特殊工人嵌套上下文：Farmer/Professore/Innkeeper 当场选择，Mafioso/Politico/Merchant 资格锁定，Oracle 延迟到真实抽牌，失败执行原子回滚，公开视图不泄露私有 `PlannedWorkerState`。

## 规则与组件基线

规则证据的本地提取文件为 `C:/Users/admin/Games/Viticulture/essential-web/artifacts/expansions/evidence/tuscany-essential.txt:212`（本地证据）：特殊工人“when you place”能力先于其他事情触发；Farmer、Chef、Innkeeper、Professore、Politico、Oracle、Merchant、Traveler、Messenger 的牌文位于同文件 `:236-282`。本次实现对 Planner × Farmer 的“当场选择、未来基础行动结算”属于对官方文字的实现解释；官方材料未找到专门的 Planner × Farmer 问答裁定。

组件数量与产品模块采用 Stonemaier 官方页面作为外部基线：

- [Tuscany Essential Edition](https://europe.stonemaiergames.com/products/viticulture-tuscany)：36 张结构牌、11 张特殊工人牌，并提供 Game Rules 入口。
- [Moor Visitors](https://europe.stonemaiergames.com/products/viticulture-moor-visitors)：40 张 Moor 访客牌。
- [Visit from the Rhine Valley](https://europe.stonemaiergames.com/products/viticulture-rhine-valley)：80 张 Rhine 访客牌，且官方说明其牌背独立、不能与其他访客牌混用。

## 覆盖边界与发布建议

当前验收不是逐卡穷举：24 次自然对局策略不主动打访客；Rhine 的 179 分支主要是 UI/render sweep；Moor/Rhine/结构/特殊工人包含代表性真实效果与嵌套恢复，但不是每个替代选项与组合的独立 Go 断言。库存报告也明确标记 `exhaustiveVisitorBrowserCoverage=false`。

定向复审结论：本报告上一版的 P1/P2 已关闭，最终快照未发现新的已确认 P1/P2。发布前仍应等待全量验收完成，并保留现有逐卡覆盖边界说明；除此之外，本报告未对生产源码作任何修改。
