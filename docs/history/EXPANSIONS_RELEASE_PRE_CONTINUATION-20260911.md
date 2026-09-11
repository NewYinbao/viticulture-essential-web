# 扩展开发版交付记录

> 当前状态（2026-09-11）：`partial`。Tuscany Essential 特殊工人与36张结构牌模块已进入可配置/可开局的实现阶段；Moor、Rhine 仍未开放。本文早期“所有扩展选项均禁用”的段落仅保留为历史记录，不能覆盖本状态。

日期：2026-09-11；状态：**partial**。

**当前阶段：Tuscany 特殊工人与建筑模块可玩；Moor、Rhine 仍未完成并由服务端拒绝。** Tuscany 主板也已可配置，但三扩展完整交付仍未完成。本交付保留实际领域代码、配置保护、隔离发布程序和验证证据，不能当作完整扩展可玩版本。

## 最新主板集成验收

本阶段最新修复已进入独立扩展 EXE：School 的训练费用/特殊工人选择、Charmat 的 1 红 1 白边界、Tavern 的精确红白资源选择与 EE Statue 终局阈值均已通过 Go 回归；浏览器验收仍是代表流程，不等同于36张逐张自然局和24组合自然终局。

本轮完成主行动表单、原创折叠影响力地图和逐格奖励，补春秋／金币先领奖提示，修复Organizer／Planner代表规则恢复链并接通Planner／Manager操作；本阶段补齐特殊工人 Politico 复杂奖励与隐藏行动格规则，接入结构牌官方 EE 草拟，并修复 Farmer 隐藏奖励归一化、Tuscany 替代奖励重复结算和待决选择恢复。已重建原生Windows开发版，普通／审计Go回归、vet、Node测试及双Edge配置／开局／重启恢复冒烟通过。Moor、Rhine仍不可从正式大厅开局；完整构建哈希、复现方式和精确边界以 [主板操作验收](tuscany-ui-acceptance.md) 为准。

## 可用性与实际代码

| 项目 | 当前结果 |
| --- | --- |
| EE 本体 | 可开局；全套 Go 回归及原生双浏览器派工／恢复验证通过 |
| 开局配置 | 持久化与公开同步；房主／大厅／revision／开局锁定；24 组合结构校验；已实现的 EE/Tuscany 主板＋建筑/特殊工人组合可开局，Moor/Rhine 仍拒绝 |
| Tuscany 主板与影响力 | 领域主板＋主行动界面和代表恢复测试，已可配置；未完成全部跨季组合与合法开局验收 |
| Tuscany 36 建筑 | 已实现目录、有限牌堆、EE 官方四轮开局草拟、Tuscany side2/影响力获取、结构垫/空田建造、拆除、12个私人行动、增强/年末效果及定向 Go/Node 测试；独立 EXE 双 Edge 已验证结构草拟/建造、隐私与重启恢复；全量自然终局尚未验收 |
| Tuscany 11 特殊工人 | 本阶段已实现并验收两类牌池选择、训练与加价、身份/次年可用、11 项能力、Politico复杂奖励及访客/Planner/Messenger状态；独立双 Edge 配置/开局/重启恢复冒烟通过，完整扩展自然终局仍未验收 |
| Moor 40 访客 | 未实现；完整清晰卡面／勘误未核实 |
| Rhine 80 访客 | 未实现；完整卡面及 4 张 Tuscany 依赖过滤未核实；不会混入 EE/Moor |
| 扩展规则说明与新手引导 | 已实现按真实View配置分主题、Tuscany已核实通则、当前访客分支／费用／数量、真实选择定位、多人私有响应与切房恢复；逐张扩展卡说明及完整玩法引导仍未完成 |

主板开发代码覆盖逐格奖励、交易、售酒、影响力、个人过季奖励、个人冬末回收及清理、下一年起床位和 25 分终局。Planner／Organizer／Manager 有代表恢复分支测试，不覆盖所有访客与跨季组合。详见 [进度](expansions-progress.md)。

本轮新增影响力奖励与第二次交易的实际操作表单，接通已有领域待决选择；使用真实Go `Apply` 产生的测试局验证放星收益、弃两张手牌换葡萄、过季抽牌、个人收入与下年起床。结构阶段另以真实领域测试验证36张目录、有限牌堆、建造/拆除、私人行动、增强与年末效果；普通交付不会错误触发酒标工厂隐藏选择，酒标工厂改为独立私人行动。完整扩展仍未交付。规则帮助复用已有访客牌面／分支费用，女王多人回复已在两个独立浏览器中验证手牌定位、确认定位、非响应者隐私，以及待决存档重启恢复。详见 [追加帮助验收记录](expansions-help-acceptance.md)。

## 交付文件与使用边界

- 工程：`C:\Users\admin\Games\Viticulture\essential-web`
- 独立 EXE：`C:\Users\admin\Games\Viticulture\essential-web\dist\Viticulture-Expansions.exe`
- 本阶段构建 SHA-256：`5724522AB1BD5AC6F227D1F155CD80ED90B8E202048210EC0C843B838015228D`
- 独立启动器：`C:\Users\admin\Games\Viticulture\essential-web\Start-Viticulture-Expansions.cmd`
- 构建脚本：`C:\Users\admin\Games\Viticulture\essential-web\scripts\build-expansions.ps1`
- 机器结果：`C:\Users\admin\Games\Viticulture\essential-web\artifacts\expansions\implementation-result.json`

启动器留给用户手动运行：默认 `0.0.0.0:3014`，存档 `runtime/expansions-data`。本次未运行该启动器、未自动启动永久服务器。静态 JS/CSS 和中文说明嵌入 EXE，无新增在线运行依赖。原创影响力示意地图已嵌入；逐张扩展卡牌面和其他未实现模块资源尚未完成。

构建（PowerShell，在工程根目录）：

```powershell
& ./scripts/build-expansions.ps1
```

脚本优先使用可用 Windows Go，否则调用已有 WSL `/opt/viticulture-toolchain/go/bin/go`，输出固定为扩展版 EXE。也可用 `-Go` 指定 Windows Go。本机默认 WSL 回退路线已实际构建成功，不修改 PATH。

## 验证结果与复现

前轮特殊工人测试、结构阶段定向 Go 测试及目录/帮助/UI Node 测试均已加入；原有默认测试和审计断言保留。以下检查继续通过：

```powershell
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go test ./...
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go test "-tags=ee_rule_audit,audit_diff" ./...
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go vet ./...
& 'C:/Program Files/nodejs/node.exe' --test tests/polling-unit-test.cjs
& 'C:/Program Files/nodejs/node.exe' tests/expansions-browser-test.cjs
```

本次浏览器脚本已在最终独立 EXE、随机回环端口、TEMP 存档和双 Edge 中真实启用 structures+specialWorkers，并验证 EE 四轮结构草拟、配置、开局、私有 View、SSE/轮询、建造与重启恢复；这仍不是结构模块全量验收。本结构阶段新增的 Go/Node 结果与明确未验收项记录在 `artifacts/expansions/structures-acceptance.json`；24组合自然开局至终局及36张结构代表动作仍待执行，不能把配置冒烟报告冒充全量结构验收。

浏览器报告：`artifacts/expansions/browser-result.json`。Go/Node 输出：`artifacts/expansions/check-results.json`。桌面 1440×1000、手机 390×844 截图：`lobby-desktop.png`、`lobby-mobile.png`、`game-desktop.png`、`game-mobile.png`、`guide-mobile.png`，均在 `artifacts/expansions`，已实际查看。检查横向溢出、可见引导目标、规则面板焦点恢复；修复大厅徽章文字排版。本体截图不构成扩展玩法验收。

## 规则证据与阻塞

- [Tuscany 出版社页](https://stonemaiergames.com/games/viticulture/tuscany-essential-edition/)及[出版社规则书的下载镜像](https://games-for-you.oozmi-cdn.com/products/4b617307-9552-47af-9d86-f41840ebcb3a/pdfs/tuscany-rules-english.pdf)：实际下载 8 页 PDF，检查主板、起床、影响力及特殊工人；图像不作为产品资源。
- [EE 规则书](https://shared.steamstatic.com/store_item_assets/steam/apps/414235/manuals/Viticulture_EE_Rules.pdf?t=1551153655)：实际下载 20 页 PDF，渲染阅读访客澄清。
- [Moor 出版社页](https://stonemaiergames.com/games/viticulture/moor-visitors-expansion/)及[Rhine 出版社页](https://stonemaiergames.com/games/viticulture/visit-from-the-rhine-valley/)：核实数量、替换规则和官网示例；完整卡效审计未完成。Moor 整套照片有模糊／遮挡，Rhine 示例不足以实现全部牌。
- [官方 FAQ](https://stonemaiergames.com/games/viticulture/faq/)已读取；其 BGG 综合 FAQ 入口返回 403，Dropbox 目录未取得完整清晰牌面。Moor 40、Rhine 80 及四张依赖牌的完整规则证据仍是客观缺口；36张结构牌本阶段已逐牌记录来源与效果。

特殊工人基础规则已取得，主板界面与扩展引导也仍有实现工作，不能把全部未完成项归因于资料不足。未核实牌效没有猜写、没有套本体效果、没有开放成可用牌。

## 保护核对

与 `artifacts/expansions/source-before-expansions-20260911-135950.zip` 对照保留 Cloudflare 工作，详见 `artifacts/expansions/preservation-result.json`。原 `dist/Viticulture.exe` 未覆盖。没有访问或更改用户对局／`runtime/ee-data`，没有停止或重启 3013 PUBLIC 服务或本体 EXE，没有改 Windows 防火墙／服务／全局 PATH／账号或 Hermes 配置，没有提交或推送 Git。

备份对照覆盖 16 个保护文件与 17 个原有 Go 测试文件。既有 `tests/cloudflare-browser-test.cjs`、`tests/polling-unit-test.cjs` 不在该备份内，未作备份字节比较；本次未编辑这两份文件。原本体 EXE 大小与修改时间仍为 12,375,552 字节、2026-09-11 13:24:25。
