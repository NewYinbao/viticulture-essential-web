# 扩展接续审查与自然终局验证（2026-09-11）

本轮在独立 TEMP 副本审查、修复和测试，没有使用现用服务或修改玩家存档。本报告是修复作者的验证记录，不能代替合并后另一个全新任务的独立 review。

## 已证实并修复的问题

- Messenger 公开 View 泄露预约行动中的私有牌 ID；Oracle 弃牌没有回到弃牌堆。
- EE/Tuscany 年末把特殊工人计入普通工人；School 即时培训特殊工人额外增加了普通工人。
- EE 结构奖励结算无法续步；Gazebo → Manager 嵌套结算丢失第二张访客续步。
- Mafioso 重复行动复用了已消耗资源；Fermentation Tank 忽略玩家提交的酿酒配方。
- Mercado 可交付抽牌前的旧订单；Tavern / Tap Room 可结算未拥有的建筑奖励。
- 结构占地仍能种植、卖地；部分结构访客奖励绕过后续行动可行性校验。
- Guest House 缺少两张访客选牌；上述多步骤结构和特殊工人缺少实际资源控件。
- Tuscany 秋季 `build_tour` 与结构位置复用 `Action.Mode`，卡片 UI 提交 `mat/field` 后服务端拒绝。追加修复保留结构位置并分派到真实建造处理，测试覆盖庄园、田地和兼容 `build` 路径。

## School 规则纠错

旧阶段报告和固定普通 4 / 特殊 5 金币的验收结果无效。School 实际卡面为免费培训一名工人且本年可用；卡面还给出使用奖励 1 金币。普通工人基础费用为 0，特殊工人基础费用为 1，另计对手 Academy 的实际附加费；School 的 1 金币在支付和培训完成后获得，不能预支。

证据为出版社卡面在 Boardspace 的镜像：[School 第 7 张](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000007.jpg)、[Academy 第 17 张](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000017.jpg)。特殊工人的额外 1 金币规则见 [Tuscany EE 说明书第 5 页](https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/414236/manuals/Tuscany_EE_Rules.pdf?t=1551153526)。

新增 School 回归测试覆盖普通/特殊 × 有无 Academy、付款不足回滚、0 金币普通 / 1 金币特殊训练、即时可用和存档恢复。

## 8 组合自然终局

`tests/e2e/expansions-natural-test.cjs` 使用真实创建、密码入座、配置、开始和行动 API，两个浏览器接收 SSE 并检查收官界面。策略仅根据行动者的本人 View 和公开规则作决策。不读取或修改保存文件，不固定牌堆，不加分或强制终局。

| 棋盘 | 结构 | 特殊工人 | 终局年份 | 成功行动数 | 最终得分 |
|---|---|---|---:|---:|---|
| EE | 关 | 关 | 11 | 154 | 21 / 10 |
| EE | 关 | 开 | 12 | 176 | 20 / 10 |
| EE | 开 | 关 | 12 | 187 | 21 / 13 |
| EE | 开 | 开 | 12 | 195 | 20 / 12 |
| Tuscany | 关 | 关 | 16 | 281 | 26 / 19 |
| Tuscany | 关 | 开 | 16 | 300 | 25 / 21 |
| Tuscany | 开 | 关 | 15 | 278 | 25 / 11 |
| Tuscany | 开 | 开 | 16 | 307 | 26 / 17 |

各局验证自然达到 20 / 25 分阈值、在同年正常收官、有获胜者、Tuscany 所有人完成个人年末、实际建造/种植/收获/酿酒。开启结构时至少自然建成一张结构牌；开启特殊工人时必须实际培训并使用其身份放置。浏览器未出现运行异常。策略没有打访客，特殊工人主要用于私人行动；不能把此结果算作 120 张访客、36 张结构或 11 种工人能力的全分支覆盖。

可用 `VITICULTURE_NATURAL_VISITORS=ee,ee_moor,rhine` 拓展到 24 组合；须先实现并开放对应模块，不能放宽可用性校验使测试通过。脚本默认新建临时服务和数据目录，结束后关闭自己的进程。通过 `VITICULTURE_NATURAL_EXE` 可指定待验收二进制，通过 `VITICULTURE_NATURAL_OUTPUT` 指定证据输出。

本次二进制 SHA256：`7E7C1AC3FE604599F64A1C9F6FDF09D91B11D8DF2E007DE7709CD3C538178F4E`。完整 JSON 行动记录及截图在执行副本的 `natural-evidence` 下。截图捕获时页面可能保留在手牌位置；收官状态由 DOM 断言和 JSON 终态确认。

## 其他验证与范围

独立问题回归和 School、秋季建造回归通过；当前游戏包通过 `go test -tags=ee_rule_audit,audit_diff ./internal/game`。最初 17 文件修复快照通过全部 Go 测试、带标签测试和 `go vet`。密码代码并行施工后，最终全工程测试需要在主目录合并完成后重跑。此环境未安装 gcc，无法运行需要 CGO 的 Go race 检测。

`tests/e2e/expansions-review-test.cjs` 是明确标注的 fixture 测试，已通过 Guest House 双访客、Fermentation 实际收获后重启恢复并选择白葡萄、Mafioso 换新藤与新田、Mercado 仅显示新订单四条浏览器路径；这四条不能冒称自然全局。

仍需逐张实际卡面审查 36 张结构、补全逐牌触发和分支证据，待访客模块完成后补齐 24 组合，并由新的独立任务 review 合并结果。
