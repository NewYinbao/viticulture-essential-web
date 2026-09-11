> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# Moor 与公共工人交互追加验收

本次在主目录已合并 Moor/Rhine 的独立 TEMP 快照上审查并修改，未操作生产服务或存档。这里只记录本次已复现并修复的具体缺陷，不替代全工程最终独立 review。

## 修复

- Builder（moor-summer-12）把扩展建筑在 Buildings 和摆放位置里的两份记录重复计数。真实调用 buildStructure 建三座，旧结果计成六座；现在按建筑 ID 去重。三座拒绝得分且整个房间回滚，补齐三座固定建筑后才能获得两分。
- Chef 零金币顶走唯一 Soldato 后无需付通行费，旧 workerPlacements 在模拟顶替前过滤掉这项合法行动。现在逐个工人/位置计算顶替后仍存在的 Soldato，并把实际 toll 提供给 UI。
- Chef 的 API 自动位置以前先遇到对手 Chef 就失败，即使另有空位或可顶替普通工人。现在跳过不能顶替的 Chef。根代理当前 UI 已提交选出的具体位置，因此这里修的是 API 默认选择，不声称旧 UI 必然失败。
- 公共培训价格与可用性以前遗漏 Soldato 入场费。现在显示总支付金额及短注“含通行费1”；差一金币时行动提示明确说明。
- Farmer 在非奖励格选择培训折扣后，resumeSpecialPending 忽略 BonusOverride，仍收四金币。现在正确收三金币；API 先选培训类型时也保留后续 Farmer 奖励选择。面板提前显示“三金币（需选农夫折扣）”。
- 特殊工人被选中时，普通工人卡不再同时显示选中边框。

## 验证

`go test -tags=ee_rule_audit,audit_diff ./... -count=1` 与 `go vet ./...` 通过。新增回归覆盖真实建筑计数、失败事务回滚、Chef 有/无剩余 Soldato、Chef 自动空位/顶替、培训通行费、Farmer 普通培训及特殊工人选类续步。

`tests/e2e/worker-interactions-browser-test.cjs` 使用显式种子夹具、独立服务器、真实 Edge 浏览器。所有游戏提交均点击界面控件；状态 API 仅用于读取断言。这六个夹具不属于自然对局：

| 夹具 | 界面和结果断言 |
| --- | --- |
| CHEFBUMP | 零金币 Chef 顶走 Soldato，导览获得两金币 |
| SPECIALONLY | 只有 Chef 待命仍可确认导览，普通工人未被错误选中 |
| DISCOUNTTRAIN | 四金币站折扣格可培训 Chef，直接完成，无第二次 special_train |
| TRAINSHORT | 四金币加通行费不够，棋盘明确显示还差一金币，无状态变化 |
| TRAINEXACT | 五金币显示包含一金币通行费，培训成功且余额零 |
| FARMERTRAIN | 三金币在非奖励格选 Farmer，再通过奖励按钮取折扣，完成培训 |

六项全部通过，浏览器无脚本错误。证据目录为运行器指定的 `VITICULTURE_WORKER_OUTPUT`，包含结果 JSON 和面板/结果截图。

本次测试二进制 SHA-256：`4CEC32618CDD705B577E7FA6220237FD241AB852ED2F9776931BAB65366D8706`。

导出夹具：设置 `WORKER_BROWSER_FIXTURES`，运行 `go test ./internal/game -run TestWorkerBrowserFixtureExport -count=1`。浏览器运行器需要 `VITICULTURE_WORKER_FIXTURES`、`VITICULTURE_WORKER_EXE`，可选 `VITICULTURE_WORKER_OUTPUT`。
