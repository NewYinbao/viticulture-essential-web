# Planner × 特殊工人放置时序增量报告

日期：2026-09-12。实现基线为 Messenger 隔离快照 `C:\Users\admin\AppData\Local\Temp\viticulture-messenger-b2pebu3v`，工作副本为 `C:\Users\admin\AppData\Local\Temp\viticulture-planner-placement-20260912`。生产源码未由本任务修改；根任务随后自行合入，最终又做了只读语义对照。

## 交付

- `planner-special-timing.patch`：严格相对 Messenger 快照的增量补丁。
- SHA-256：`40CBC49073E31AFCCA107720B26E02CC7848087CC09101EBA1DE64FFF6A38872`。
- `planner-special-browser-result.json`：Microsoft Edge 1.63 的同局可见交互结果。
- 新回归：`internal/game/planner_special_timing_test.go`、`internal/game/planner_special_browser_fixture_test.go`、`tests/e2e/planner-special-browser-test.cjs`。

## 完成的时序

- Farmer 在 Planner 预约放置当场产生选择；所选奖励私密、持久化在预约中，未来基础行动成功时一次结算。未来不会再次出现 Farmer 放置选择，客户端也不能伪造 `BonusOverride`。
- Professore 在预约当场选择并收回本季普通工人；Innkeeper 在预约当场付款并随机取得同行动对手的指定季节访客。二者未来执行基础行动后不会重复触发。
- Mafioso 的非奖励格资格、Politico 的奖励格资格和具体奖励、Merchant 的“其他玩家均已进入下一季”资格均在预约放置时锁定；实际重复行动／重复奖励／抽任意牌仍在未来基础行动后发生。
- Oracle 只在未来基础行动真实抽牌时多抽一张并弃一张。
- Planner 放置时的特殊选择使用持久化 `planner_placement` child context；外层访客、第二张访客和已排队响应者经 `Parent`／`ParentChoices` 恢复。
- 预约公开投影仍只显示玩家、格、位置和工人身份；`PlannedWorkerState` 在 `publicPlans` 中清空。
- Planner 表单将普通、灰色、大工和待命特殊工合并为单一工人选择，避免“大工 + 特殊工”这类矛盾组合，并使仅剩特殊工的真实浏览器路径可操作。

## 规则依据与解释边界

`artifacts/expansions/evidence/tuscany-essential.txt` 第 5–6 页规定：涉及 “when you place” 的特殊工人能力在其他事情之前触发；Farmer、Professore、Innkeeper 的牌文均以放置为触发。Mafioso／Politico／Merchant 的后半段明确发生在基础行动以后；Oracle 以实际抽牌为触发。

没有找到专门回答 Planner × Farmer 的官方逐问裁定。Farmer 的 plant／harvest／make wine／visitor 等奖励依赖未来基础行动的参数，因而本实现把“当场触发”解释为当场选择并锁定奖励，再随未来基础行动成功执行；不会脱离基础行动提前领取，也不会延迟到未来才决定。

## 验证

在全新 Messenger 快照副本 `C:\Users\admin\AppData\Local\Temp\viticulture-planner-special-verify-20260912`：

- `git apply --check --whitespace=error planner-special-timing.patch`：通过。
- 实际 `git apply --whitespace=error`：通过。
- `go test ./internal/game -run Planner -count=1`：通过。
- `go test ./internal/game -run SpecialWorker -count=1`：通过。
- `go test ./internal/game -run Messenger -count=1`：通过。
- `go test ./internal/server -count=1`：通过。
- `node --check web/static/js/visitor-ui.js`：通过。
- `node --check tests/e2e/planner-special-browser-test.cjs`：通过。
- `node tests/tuscany-ui-unit-test.cjs`：5/5 通过。

真实 Windows Microsoft Edge 1.63（headless）完成同一房间的可见流程：Planner 表单选择 Farmer → 当场选择订单奖励 → 停止并重启临时测试服务器 → 逐玩家点击“结束本季”进入秋季 → 点击预约执行。结果为基础抽 1 张加锁定 Farmer 奖励抽 1 张，共 2 张；无第二次 Farmer choice，`pageErrors=[]`。测试只操作临时 EXE 和临时数据目录。

独立快照中的 Go 回归还覆盖：两次 JSON round-trip、外层双访客恢复、Professore／Innkeeper 当场效果、Mafioso／Politico／Merchant 资格锁定、Oracle 延后、失败执行对计划／上下文／choice 的原子回滚，以及预约私有状态不进入公开视图。

本隔离验证未重跑整个 `./internal/game`；Messenger 原始快照的全套曾有两项与本补丁无关的既有失败（Truss/Harvest Machine 字段选择与空手牌 null）。根任务在合入其它修复后报告全包 audit Go 已通过。
