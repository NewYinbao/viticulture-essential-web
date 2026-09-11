> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

> 本文件为先前阶段的范围或验收快照，文中“未实现/未开放”不代表续作当前状态。当前实现和验证结论见 [最新交付记录](expansions-release.md)。

# Tuscany 主板操作与最终回归验收

更新时间：2026-09-11T16:52:35+08:00

## 结论先行

**当前为部分实现的扩展开发版，不是完整扩展发布版。正式大厅已可配置 EE/Tuscany 主板、36张建筑和特殊工人模块；Moor、Rhine 的未完成保护仍保留。**

本轮交付的是已写入源码、已嵌入 Windows EXE 并通过开发局操作验证的 Tuscany 主板界面与规则恢复修复。开发夹具不改变生产开局限制。

## 实际完成

- 春夏秋冬主板和对应行动显示；个人季节、起床行与奖励预览。
- 真实放星／移星、资源交易、买卖田地、建造／导览、售酒的主行动表单，接通服务器 Action，而非只显示说明。
- 原创本地 SVG 影响力示意图、每区各玩家星数、独占／并列计分说明。地图默认折叠，避免手机行动表单被长地图推到很远；不是商业原图或等比例地理地图。
- 按真实物理格显示奖励、自动选位、奖励放弃；金币格区分先领奖／后领奖，不给大工人溢出位借用其他格奖励。
- 修复春秋行动提示缺省、第二格建造折扣遗漏、访客先领金币的可行性提示；自己的牌面提示不投影给等待对手。
- Planner 支持对应预约动作参数、放弃奖励与金币先后；Manager 显示 Tuscany 前季主行动，不额外消耗工人。Organizer 去除错误套用本体第5行颜色参数的校验。
- Planner 的可行性探测不再删除“先拿格子金币后打 Banker”这一合法预约。
- 引导跟随实际行动、所需资源与确认状态；访客提示不再把 Tuscany 金币格说成双访客格。

关键实现：`internal/game/tuscany*.go`、`availability.go`、`view.go`、`visitor_special.go`；`web/static/js/tuscany-board.js`、`tuscany-inputs.js`、`action-panel.js`、`action-options.js`、`visitor-ui.js`、`help.js`。

## 本次实跑验证

| 验证 | 结果与边界 |
| --- | --- |
| Go 普通回归 | `go test ./...` 通过 |
| Go 规则审计标签 | `go test -tags=ee_rule_audit,audit_diff ./...` 通过 |
| 静态 Go 检查 | `go vet ./...` 通过 |
| JS 单元测试 | 12 项通过，覆盖配置帮助、轮询、逐格奖励、星移动、交易、地区多数、访客提示 |
| 原生本体双浏览器 | 7 组检查通过：UI 建房／配置／加入／开局、父母／起床、双方派工、隐私与重启 |
| 原生帮助双浏览器 | 7 组检查通过：24 个保存配置的主题隔离、实际待决选择与女王回复、断线和待决重启 |
| 原生 Tuscany 主行动双浏览器 | 6 组检查通过，详见下节；使用开发季节初态及真实 Apply 生成的续接选择 |
| 既有本体规则浏览器断言 | 5 个案例通过；只用临时适配器改为本次 EXE 和 Edge，保留原有断言，不宣称原 Chromium 命令原样运行 |
| 本体自然完整游戏 | 无预置资源、无修改存档；UI 建房／加入／开局，之后由合法 HTTP 策略推进，浏览器 SSE 同步；第 11 年自然结束。不是全鼠标或完整扩展游戏 |
| Tuscany 领域完整循环 | 2–6 人从开发起床初态，通过真实 Apply 和每步 JSON 恢复自然终局；不是生产 lobby／HTTP／浏览器完整扩展开局验收 |
| 视觉核验 | 查看最终构建桌面主行动／展开地图、手机售酒确认及本体自然终局截图；代表截图无页面横向溢出、确认按钮无遮挡。手机为 390×844 视口模拟，并非实体手机 |

本体自然完整游戏还验证：每次派工真实消耗工人、建造付费、种植／收获／酿酒／访客、达到20分后年末结束、非本人操作拒绝、过时 revision 拒绝、HTTP/SSE 私有手牌、断线期间错过动作后追上、待决访客时重启且保留原令牌与完整 View。

## Tuscany 主行动实际浏览器路径

1. 春季在第2格放星，取得地区奖励，继续第二颗奖励星；另一玩家在建造第2格按折扣建棚架。
2. 夏季金币换分，进入第二次交易时停止测试 EXE，再从同一临时存档恢复；用刚得的分换预选颜色两张牌。另一玩家出售空田并得相应格奖励。
3. 秋季第1格导览得对应奖励；另一玩家在第2格建造，不误领第1格奖励。
4. 冬季手机界面出售自己的红酒，得分并完成奖励放星。
5. Manager 选择前季影响力行动，放星而不多花工人。
6. Planner 在真实预约到期时先领金币，再打 Banker 买收入，恢复正常回合。

第一端使用 SSE，第二端在本地拦截传输配置启用现有轮询；没有打开公网隧道。所有游戏数据来自新建 TEMP 测试目录，没有使用用户正在进行的对局。

## 可复现命令与证据

在项目根目录，WSL Go 路径为 `/opt/viticulture-toolchain/go/bin/go`。先运行 Go 回归和构建，再在原生 Windows 执行：

```text
node --test tests/expansions-help-unit-test.cjs tests/polling-unit-test.cjs tests/tuscany-ui-unit-test.cjs
node tests/expansions-browser-test.cjs
node tests/expansions-help-browser-test.cjs
node tests/tuscany-ui-browser-test.cjs
```

帮助／主行动测试需要开发夹具。生成时环境变量须用绝对路径，因为 Go 测试工作目录是包目录：

```bash
VITICULTURE_HELP_FIXTURE=/mnt/c/Users/admin/Games/Viticulture/essential-web/artifacts/expansions/help-fixtures.json /opt/viticulture-toolchain/go/bin/go test ./internal/game -run TestExpansionHelpBrowserFixtures -count=1
```

报告均在 `artifacts/expansions/`：`final-regression-result.json`、`browser-result.json`、`help-browser-result.json`、`tuscany-ui-browser-result.json`、`ee-rules-final.json`、`ee-natural-final/result.json`、`tuscany-domain-final.log`。

## 发布与保护

- 独立程序：`dist/Viticulture-Expansions.exe`，12849152 字节。
- SHA-256：`2CBA8811C309C6843EC9C69EFD805E89E046E63A7312CE9AFE7879386A7ADE59`。
- 启动器：`Start-Viticulture-Expansions.cmd`；手动启动后地址 `http://localhost:3014`，绑定 `0.0.0.0:3014`，独立存档 `runtime/expansions-data`。
- **该启动器允许正式配置 EE/Tuscany、建筑和特殊工人；Moor/Rhine 仍由服务端拒绝。默认仍是 EE 本体，测试使用独立临时存档，不会写入用户存档。**
- 未替换 `dist/Viticulture.exe`；未通过本次验收启动长期服务或改动原 PUBLIC 服务。测试 EXE、浏览器由各测试自行停止。
- 未提交或推送 Git；当前源码仍包括既有未提交改动。静态资源已嵌入独立 EXE，无新增在线前端依赖。

## 仍未完成，不能标为可玩

- 36 建筑的24组合自然开局至终局、36张逐牌真实浏览器触发与完整嵌套恢复覆盖；源码已实现规则目录、官方 EE 草拟、建造／拆除／占田／年末流程及 Tuscany side 2。
- 11 特殊工人的全量自然终局浏览器覆盖及所有访客／预约组合；训练、能力和配置已接入源码。
- Moor 40、Rhine 80 的完整清晰卡效证据、所有分支、中文牌面、有限牌堆与 Rhine 四张依赖牌过滤。
- 主板所有待决与访客跨季嵌套组合、独立模块和合法混合，从生产合法开局到终局的全链路验收。
- 上述未实现效果对应的完整规则说明和新手引导。

因此本轮可以验收主板代表开发路径与本体回归，**不能验收“完整 Tuscany／Moor／Rhine 已实现”。**
