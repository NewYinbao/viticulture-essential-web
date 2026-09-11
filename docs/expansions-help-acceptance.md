> 本文件为先前阶段的范围或验收快照，文中“未实现/未开放”不代表续作当前状态。当前实现和验证结论见 [最新交付记录](expansions-release.md)。

# 规则说明与新手引导追加验收

2026-09-11。整包仍为 **partial**。Tuscany 特殊工人与建筑配置已开放并接入真实状态；Moor、Rhine 仍禁用。此文件记录真实实现和本轮测试，不把代表流程或开发存档的帮助显示称为完整扩展成功开局。

## 已落地

- 大厅已保存规则摘要、编辑预览、全桌可见与开局锁定说明；Rhine明确替换访客，未完成选项继续禁用。
- 复用 `web/static/js/rules-content.js` 生成按配置隔离的局内内容；纯本体不展示扩展主题。Tuscany替换20分、夏冬派工、秋季抽牌等不适用说明，加入个人四季、起床、即时与年末时机、影响力、交易和售酒。实际起床行、行动、地图值读取公共View；建筑side2不套用side1地图。
- 建筑通则与独立模块草拟、占地、派工、持续／年末／拆除；特殊工人开局选取、两人排除旅店老板、培训额外1金币、总数上限、11项已核实基础能力。建筑和特殊工人已接入配置与真实代码；全量自然局与逐牌浏览器覆盖仍未完成，不以代表流程冒充完整验收。
- 当前选择入口：复用原有本体中文卡库、访客分支标签、费用、数量及服务器不可用原因；所选分支改变时同步规则和提示。未核实扩展卡显示缺口，不套本体牌效。
- 引导优先处理本人待决选择、多人响应和断线；需要手牌时定位原手牌控件，就绪后定位确认按钮。非响应者没有可执行定位，不展示另一人的牌面或分支。
- Tuscany 已有开发代码的影响力与额外交易选择接通表单；实际消费自有资源并取得收益。过季抽牌、年收入与下年起床位都有明确中文提示。
- 开启、关闭、跳过、按房间重看；切房移除旧主题和高亮，重连按服务器View重建，断线停止操作提示。规则面板为页面内区域，不是覆盖选牌的弹层；键盘关闭恢复原入口或其重绘后的对应入口。
- 常用文案和样式内嵌于独立EXE，测试没有请求外部资源。官网链接仅作参考。

## 实际执行与范围

1. `go test ./...`、`go test -tags=ee_rule_audit,audit_diff ./...`、`go vet ./...` 通过，原默认及审计断言未删改。
2. `TestExpansionHelpViewMatrix` 对24配置、两个身份验证真实HTTP与SSE投影一致、公共配置、未完成原因、私有选择隐藏、读取不改局面、存档恢复。这里测试的是保存的开发配置，不绕过生产configure/start保护。
3. Node原轮询3项、新帮助4项通过；新帮助含24主题组合、不污染EE源内容、服务器地图／起床数据、side2隔离、具体费用和数量、未知卡拒绝套用通用说明。
4. 原生Windows `tests/expansions-browser-test.cjs`：独立EXE、随机回环端口、新TEMP、两个独立Edge；真实UI建房／保存本体配置／加入／开局／家族／起床／双方派工，SSE与轮询、锁定、隐私、自己的EXE重启恢复、本体帮助回归通过。
5. `tests/expansions-help-browser-test.cjs`：读取Go测试生成的全新开发存档，经实际HTTP View显示24配置；两独立Edge分别SSE／轮询。真实界面切房、创建新本体房，旧扩展主题和目标清除。实际提交已由Go `Apply` 产生的放星奖励、第二次交易、过季抽牌、年末收入／下年起床待决；女王多人回复显示具体成本与数量，验证对手隐私、选手牌和确认按钮的定位；仅重启自己EXE后恢复待决并完成回复。
6. 桌面1440×1000、手机390×844截图已实际查看。检查横向溢出、目标是否存在／可见／启用；手机确认按钮中心通过`elementFromPoint`检验没有被引导或规则面板盖住。调整手机关闭按钮避免挤窄。开发Tuscany主板完整派工界面仍缺失，放星截图的待决表单不代表完整地图或主板验收。

截图：`artifacts/expansions/help-modules-desktop.png`、`help-rhine-mobile.png`、`help-influence-desktop.png`、`help-reply-mobile.png`、`help-ready-mobile.png`、`help-ready-viewport.png`。报告：`help-browser-result.json`、`help-check-results.json`、`browser-result.json`。

## 仍必须保留的未完成项

- 全部36建筑、40 Moor、80 Rhine的逐牌官方卡面／勘误、中文效果、分支、数量、费用与条件；Rhine四张Tuscany依赖牌名单及过滤仍未核实。
- 特殊工人真实选取、培训与能力操作尚未实现，所以没有该模块完整真实待决引导验收。
- Tuscany完整派工前端、原创地图／SVG、全部格位可用性、完整Planner／Organizer／Manager及扩展访客嵌套选择链与引导。
- 真正启用各独立模块／混合模块的双浏览器合法开局到终局、扩展有限牌堆与隐私恢复的全量验收。当前23种扩展配置仍拒绝开局；保存本体配置同步与开发配置显示测试不能冒充成功改变可玩扩展配置。

因此追加帮助交付也只能视为**部分完成**，不得把整包或帮助需求标记complete。原有主任务全部未完成项继续保留于 `expansions-progress.md` 和 `implementation-result.json`。

## 复现

在工程根目录PowerShell执行，使用已有WSL Go；测试导出器只存在于`_test.go`，生产EXE没有新增测试接口或开局后门。

```powershell
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go test ./...
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go test '-tags=ee_rule_audit,audit_diff' ./...
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec /opt/viticulture-toolchain/go/bin/go vet ./...
wsl.exe --cd /mnt/c/Users/admin/Games/Viticulture/essential-web --exec env VITICULTURE_HELP_FIXTURE=/mnt/c/Users/admin/Games/Viticulture/essential-web/artifacts/expansions/help-fixtures.json /opt/viticulture-toolchain/go/bin/go test ./internal/game -run TestExpansionHelpBrowserFixtures -count=1
& ./scripts/build-expansions.ps1
& 'C:/Program Files/nodejs/node.exe' --test tests/polling-unit-test.cjs tests/expansions-help-unit-test.cjs
& 'C:/Program Files/nodejs/node.exe' tests/expansions-browser-test.cjs
& 'C:/Program Files/nodejs/node.exe' tests/expansions-help-browser-test.cjs
```

没有启动3014永久服务，没有触碰3013公网、本体EXE进程、原存档或Hermes配置。只生成 `dist/Viticulture-Expansions.exe`；原 `dist/Viticulture.exe` 保持不变。
