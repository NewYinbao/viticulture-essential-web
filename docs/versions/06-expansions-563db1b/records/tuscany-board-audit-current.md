> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

> 本文件为先前阶段的范围或验收快照，文中“未实现/未开放”不代表续作当前状态。当前实现和验证结论见 [最新交付记录](expansions-release.md)。

# Tuscany 主板有界规则审计（当前开发路径）

结论：**partial。主板开发路径已有代表规则、UI与恢复验证，未开放生产 Tuscany 开局，不代表完整扩展已完成。**

本报告替换此前缺少代码标识符且夹带 setup-failed 输出的损坏版本；此前版本留在 `artifacts/expansions/_final-doc-backup/`。以下通过结论以本次重新执行的实际日志为准，不沿用未经复核的红绿对照声称。

## 已有证据与边界

- 项目已有 `artifacts/expansions/evidence/tuscany-essential.pdf`（8页）、同名txt，以及 `tuscany-page2.png`、`tuscany-page3.png`、`tuscany-page4.png`、`tuscany-page6.png`、`tuscany-page7.png`。主板行动／过季／影响力／交易证据主要在第2–4页。
- EE规则书已有 `ee-official.pdf` 及第11页访客澄清图；已有审计将其用于 Planner、Organizer、Manager 的基础适配。
- 现有 `faq-official.txt` 用于行动与格奖励的先后顺序。参考入口：[出版社 FAQ](https://stonemaiergames.com/games/viticulture/faq/)；[Tuscany 规则入口](https://stonemaiergames.com/games/viticulture/tuscany-essential-edition/)。
- 本次最终验收主要是源码、规则测试和真实浏览器重跑，没有重新逐像素核实全部官方图标或取得所有缺失卡面。程序通过不等于官方全规则签核。

## 本轮修复

### Organizer 第5行不再套用本体颜色校验

`internal/game/visitor_special.go` 的 `summer-33` 分支仅在非Tuscany时要求第5行提交 `Color`。Tuscany由实际季节奖励的持久化选择继续处理。

`TestTuscanyAuditOrganizerNoLegacyColorAndManagerResume` 覆盖夏季Organizer与冬季Manager→Organizer，第5行抽牌／放星、外层第二张访客、自动过季、冬末和下年起床，关键点JSON恢复并拒绝非法选择。

### Planner 不删除合法的先领金币预约

`plannerCanExecute` 对 Tuscany 实际金币格另行探测 `BonusFirst: true`，但不自动替玩家领奖或选择顺序。零金币、零收入的Banker可以先拿金币再支付买收入；无奖励格不能借用其他格金币。

`TestTuscanyAuditPlannerCoinBeforeRestoration` 从夏季预约经过秋季、冬季恢复执行；默认后领与主动弃奖励仍在资源不足时拒绝。成功路径不额外花工人、不重复领钱。

### 主行动可用性与私有牌面提示

`availability.go` 使用实际行动季，修复春秋提示缺省；建造／培训按可达物理格判断折扣；访客考虑真正空着的金币奖励格。`view.go` 的 `visitorCardReasons` 区分原始资源与金币先领，仅给当前本人提供自己的卡牌提示。

`TestTuscanyViewPhysicalSlotAndSeasonHints`、`TestTuscanyViewCoinBeforeVisitorHintsAndPrivacy` 覆盖上述边界及只读性／对手隐私。UI当前格与奖励顺序校验由 `action-panel.js` 消费此信息，领域 `Apply` 仍为最终权威。

## 其余精确测试

`internal/game/tuscany_audit_test.go` 的六个顶层测试包含：

- `TestTuscanyAuditPlannerCoinBeforeRestoration`
- `TestTuscanyAuditOrganizerNoLegacyColorAndManagerResume`
- `TestTuscanyAuditNaturalActionGame`
- `TestTuscanyAuditPersonalSeasonRewards`
- `TestTuscanyAuditFinalWinterWaitAndMajorities`
- `TestTuscanyAuditTradeInvalidReceiptRollback`

覆盖个人先过季领奖、全员结束当前季后才能在新季派工，夏／秋／冬各起床行的奖励与恢复，终局等待全部冬末、区域多数／并列、超40分不截断，非法交易接收项回滚及资源守恒。原 `tuscany_test.go` 的人数、格奖励、酒类、交易／星续接测试保留。

## 开发领域完整循环重跑

从开发专用起床初态与有限EE牌堆开始，之后只通过真实 `Apply` 推进并逐步JSON保存恢复。不中途修改资源、得分或季节。本次实际结果：

| 人数 | 终局年份 | Apply／恢复步数 |
| --- | --- | --- |
| 2 | 25 | 474 |
| 3 | 25 | 747 |
| 4 | 25 | 1045 |
| 5 | 14 | 717 |
| 6 | 22 | 1347 |

日志：`artifacts/expansions/tuscany-domain-final.log`。这是领域可行性冒烟，不是完整动作遍历、全部访客或从正式大厅开始的扩展浏览器全局验收。

## 最终验证与仍缺项

最终 `go test ./...`、`go test -tags=ee_rule_audit,audit_diff ./...`、`go vet ./...` 均实际通过；日志 `artifacts/expansions/go-regression-final.log`。代表Windows UI操作、主行动截图、Planner／Manager和实际待决重启见 [主板操作验收](tuscany-ui-acceptance.md)。

仍待裁定／验收：多预约顺序、Organizer／Manager全部嵌套和特殊起床占用情况、第7行任意牌与小屋的隐藏信息选择时机、交易换入已占用1值葡萄格的适用裁定、全部售酒奖励时机与访客跨季组合。36建筑、11特殊工人、Moor／Rhine、side2和全扩展组合仍未实现或未验收；不能归结为主板名称改成四季后便已完成。
