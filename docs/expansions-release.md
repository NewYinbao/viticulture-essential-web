# 扩展交付记录

更新：2026-09-12。本轮实现、完整功能验收及独立复审已完成，工程收尾已合入并复验。旧阶段记录已归档至 [历史发布记录](history/EXPANSIONS_RELEASE_PRE_CONTINUATION-20260911.md)。

## 当前实现

| 范围 | 已实现行为 |
| --- | --- |
| EE 本体 | 2–6 人、有限牌库、独立手牌、正常对局与存档 |
| Tuscany Essential | 四季主板、个人过季、起床奖励、影响力与终局计分 |
| 建筑模块 | 36 张结构牌、草拟、建造/拆除、私人行动、持续及年末触发 |
| 特殊工人模块 | 11 类工人、两类公开培训牌池、费用/身份/跨季能力 |
| Moor | 40 张新增访客，加入 EE 访客牌库 |
| Rhine | 80 张替换访客；EE 主板过滤 4 张 Tuscany 主板依赖牌 |
| 密码 | 房间座位密码、同名验证、改密、退出、旧会话撤销、旧存档首次设密 |
| UI 与帮助 | 公共棋盘内行动面板、卡片选择和条件提示；按本局配置显示规则与引导 |

开局配置为主板 2 种 × 建筑开关 × 特殊工人开关 × 访客牌组 3 种，共 24 种。服务端对实际选入的全部访客检查真实效果实现，缺失时拒绝开局；Rhine 不与 EE/Moor 混洗。World、Bordeaux、旧 Tuscany 额外模块、促销和 Automa 不在本轮范围内。

## 程序与启动

新程序：`dist/Viticulture-Continuation.exe`，启动器：`Start-Viticulture-Continuation.cmd`。默认地址 `http://localhost:3015`，监听 `0.0.0.0:3015`，独立存档 `runtime/continuation-data`。

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture-Continuation.cmd
```

静态资源、中文卡面、规则与引导均内嵌，不需要在线加载商业美术。只创建新的续作程序；原有正在运行的服务、存档和两个旧 EXE 保留。旧服务不会自动获得密码功能，需要房主自行安排切换。旧存档安全接续方式见 [密码说明](玩家密码与续作记录-20260911.md)。

## 验收证据如何阅读

- Go 普通/审计测试与 vet：验证规则、错误回滚、有限牌堆、恢复和公私投影。
- `tests/expansions-full-browser-test.cjs`：统一驱动真实配置、自然对局、逐建筑触发、代表工人/访客、密码及模块显示检查，记录源码与 EXE 哈希。
- 24 配置测试从正常建房开始，在两份独立浏览器中配置、加入、开局并拒绝非法变更。
- 24 自然局使用随机正常牌堆及玩家 View 策略，到达正式终局；策略不打访客，不能证明全部访客组合。
- 36 建筑使用明确标记的 seeded 场景，在真实 Windows 服务和浏览器里逐张提交、核对实际资源变化。
- Moor/Rhine 的逐牌 Go 效果测试、分支 UI 渲染和代表真实 HTTP/重启场景分别记录；不把渲染检查写成所有分支实际游玩。
- 桌面与手机 viewport 检查不等于实体手机验证；本机无 WSL C 编译器，未运行 race。

本地生成证据位于 `artifacts/expansions/` 和 `artifacts/continuation-20260911/`，不随源码提交。卡牌/工人覆盖表逐项记录来源、实际效果与测试引用。已完成两份独立功能复审，以及另外一份代码风格与工程审查；“未发现未解决问题”只针对受审快照和已执行检查，不保证软件绝对无错误。

## 规则依据与解释边界

优先使用出版社规则、卡面和设计者答复；规则取证与开源实现交叉核对记录保留在本地证据及专项审查文档，未复制参考项目源码或商业卡图。基础入口：[Tuscany Essential](https://stonemaiergames.com/games/viticulture/tuscany-essential-edition/)、[Moor](https://stonemaiergames.com/games/viticulture/moor-visitors-expansion/)、[Rhine](https://stonemaiergames.com/games/viticulture/visit-from-the-rhine-valley/)。

Rhine Tutor 按牌面允许永久失去自有工人（不含临时灰工），不另加“必须待命”限制；弃抽 X 为 0 时仍真实永久损失工人。此处没有找到专门 X=0 裁定，记录为字面规则解释。Wine Engineer 固定先选最多两颗葡萄，每轮付费陈酿，最多三轮；制造一瓶 4 值酒是另一分支。

Planner 预放置特殊工人时，放置触发能力先执行；农夫在预约时选定奖励，随未来成功的基础行动结算。没有找到专门说明 Planner × Farmer 数量奖励时机的官方裁定，这一处理结合“放置时选择”与“执行基础行动才能领取奖励”解释，不把它写成独立官方 FAQ。

失败的 Planner 预约保留工人已放置/已使用状态。该处理依据 Planner 已经完成放置、特殊工人其余行为遵循普通工人的组合规则；没有找到专门的失败退款裁定，故保留为明确的规则解释。

## 最终版本与交付

完整验收 12/12 步通过；两份功能复审关闭已确认问题，工程修复后 Go/审计/vet、JavaScript 单元/语法/格式及真实 Edge 复验通过。详细结果与非穷举边界见 [最终验收与审查](audits/CONTINUATION_FINAL-20260912.md)。

最终源码 manifest digest：`5a60cb1aee782e5f4fa8ae15d9988d64367474d3d45d077dd1bda7ec31c7232e`。新的 Continuation EXE SHA-256：`513f41b507991384ed93c7de1a8e1c3261847822d0f4de8021e380aeadaacd46`，与完整功能验收时的二进制完全相同。格式修复前后映射单独保留，没有改写原始验收证据。

源码远端为私有仓库 [NewYinbao/viticulture-essential-web](https://github.com/NewYinbao/viticulture-essential-web)，分支 `main`。构建程序和临时测试数据不随 Git 提交；本机使用上面的新启动器即可启动续作版本。
