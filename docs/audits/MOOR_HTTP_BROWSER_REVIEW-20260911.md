# Moor 三条真实交互验收

以已合并 worker-followup 与 Rhine v3 的主目录源码建立独立 TEMP 快照，构建独立 Windows EXE，使用临时存档和 Edge 浏览器运行。三个起始局面是明确标记的种子夹具，不属于自然对局。

游戏的合法提交以及 Truss 的非法酿酒配方均通过实际可见界面控件发出；HTTP 脚本仅读取双方状态及提交错误玩家、过期或已完成的 choice。没有通过 JS 合成游戏回调，也没有运行中改写得分或存档。

## 结果

三条全部通过，累计验证 35 次失败请求前后双方完整 View 相同。每个私密选择均确认对手 pendingChoice 只有 kind/playerId，没有 choice ID、资源、订单、选项或访客续步；对手界面没有可提交私密选择的控件。浏览器脚本错误为零。

- Councilman：从公共夏访客格真实打牌，先选择 Broker；抽到可填红1订单后进入 Mercado；真正终止进程并用同一存档重启一次。恢复后订单选择器只列本次新订单，完成订单得2分，再由 Broker 根据剩余白酒类型得1分；Inn 对两张访客各发一次金币，最终3分、2金币。外层第二访客选择保留并可跳过。
- Truss：从公共冬访客格真实打牌，选择两田收获三颗葡萄。通过配方界面尝试用红2+白2酿桃红4，会因后续必须陈酿两颗葡萄而被拒；状态完全不变。改用单颗红2酿造后，继续把剩余白2、红1分别陈酿为白3、红2，返回外层并结束。
- Fruit Dealer：实际付4金币把卡附在田地，再使用 Yoke 收获。选择奖励得到一次1分，葡萄为红2白1，轭标记已使用；不重复发奖励。

截图覆盖 Councilman 重启前、对手等待界面、恢复后的订单选择器、Truss 拒绝耗尽葡萄以及三条最终状态。输出文件由 `VITICULTURE_MOOR_OUTPUT` 指定，结果为 `moor-interactions-result.json`。

## 本轮发现并修复的缺陷

Truss 空手牌回到“第二张访客”选择时，`Room.View` 返回 `hand: null`，`hints.js` 对其调用 `filter`，使选择区域消失，浏览器报错。没有添加假手牌绕过；后端现在统一返回空数组。新增 Go 回归验证原态和 JSON 恢复后的空手牌及 second 选择，修复前断言得到 null，修复后通过。此问题也影响其他空手牌访客续步。

`go test -tags=ee_rule_audit,audit_diff ./... -count=1` 全部通过。

本次实际测试 EXE 的 SHA-256：`4789098CA01BBE731688508AFD42A07997A2E98236BAB0D90EC8A96CD6060CF7`。

## 复跑

设置 `MOOR_INTERACTION_FIXTURES`，运行 `go test ./internal/game -run TestMoorInteractionBrowserFixtureExport -count=1` 导出夹具。运行 `tests/e2e/moor-interactions-browser-test.cjs` 时设置 `VITICULTURE_MOOR_FIXTURES` 和 `VITICULTURE_MOOR_EXE`，可选 `VITICULTURE_MOOR_OUTPUT`。运行器只启停自己的子进程并使用自己的临时存档。

本次有界验收不涵盖 Truss 与 Harvest Machine 的替代收获组合，也不作为最终独立 review；该组合由主任务另行处理。
