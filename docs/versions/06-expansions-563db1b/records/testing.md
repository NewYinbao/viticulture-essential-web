> 历史快照：563db1b；测试范围以当时记录为准。

# 构建与验证

Go 1.25+。发行程序没有 Node 运行时依赖；开发 UI 测试需要 Node 22+ 和 Chromium。

```powershell
powershell -File scripts/check.ps1
powershell -File scripts/build.ps1
python tests/native/smoke.py
npm ci
npx playwright install chromium
npm run format:check
npm run test:ui
npm run test:rules-ui
npm run test:visitors
npm run test:game
npm run test:discard
```

Linux 使用 `go build -trimpath -o dist/viticulture ./cmd/viticulture`。E2E 自动从 PATH 找 Go；特殊安装可设置 GO_BIN。临时服务绑定随机本地端口，存档写系统临时目录，报告统一写 artifacts；不依赖任何开发者的绝对路径。`ee-browser-test.cjs` 是供其他测试调用的已运行服务冒烟脚本，可用 E2E_BASE_URL 指定地址。

## 本轮本地结果

最新稳定功能快照、最终构建和独立审查结果见 [最终验收与审查](audits/CONTINUATION_FINAL-20260912.md)。以下专项结果与新增的 24 组合整套验收分别保留，不能相互替代覆盖范围。

- 全部默认测试、ee_rule_audit、audit_diff、go vet，以及 gofmt 检查通过。
- usability-test：14 组操作体验场景，桌面／手机视口、工人选择、手牌筛选、拔藤和酿酒提交、局面更新、恢复入口、条件气泡与确认按钮联动通过。
- rules-fix-browser-test：5 组规则与访客交互回归通过。
- visitor-browser-test：141 个初始选项场景通过，含多个需要续步或多人回复的访客；这是夹具测试，不是 141 局自然游戏。
- special-browser-test：19 个 Planner、Organizer、Queen 和冬初预约场景通过。
- natural-full-game-test：未注入资源的一局随机牌序游戏完成（11 年），验证 HTTP/SSE 手牌隔离、过期 revision、轮次、断线和重启恢复、20 分年末终局。
- discard-browser-test：实际建房、父母选择、四季流程、桌面和手机弃牌验证通过。
- Windows 本机 exe 的 health、内嵌网页、鉴权、创建／加入／开始、落盘和重启恢复全部通过，使用临时存档。

本地 WSL 未配置 C 编译器，不能运行 Go race detector；已在 GitHub Actions 的 Ubuntu 作业配置 `go test -race`。CI 配置存在不代表远端已执行成功，需以仓库实际运行结果为准。

修改网页后必须重新构建程序并刷新浏览器，网页资源由 go:embed 内嵌。自然对局牌序随机；失败时先检查 artifacts 报告区分策略覆盖不足、测试失效和规则错误，不能删除断言绕过问题。

## 续作扩展的整套验收

在 Windows 上构建新的续作 EXE 后执行：

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
node tests/expansions-full-browser-test.cjs
```

入口把 EXE 复制到 TEMP，测试服务只监听随机回环端口，使用自建临时存档。默认使用原生 Windows Go；缺省安装时可通过 WSL 工具链导出领域夹具；自定义路径使用 `GO_BIN` 或 `VITICULTURE_WSL_GO`。浏览器使用已安装的 Edge（`channel: msedge`）。报告写到 `artifacts/expansions/full-browser/full-acceptance.json`；可通过 `VITICULTURE_FULL_EXE`、`VITICULTURE_FULL_OUTPUT` 指定 EXE 与证据目录。

报告明确区分正常配置/自然局、seeded 规则场景和合成 UI 显示测试，并记录受测 EXE 与源码哈希；源码在运行期间变化会使整套结果失败。24 自然局不打访客，36 建筑和复杂访客另有真实浏览器场景，不能相互替代。完整覆盖范围和最新结果以 [交付记录](expansions-release.md) 为准。

纯 JavaScript 规则/目录/视图回归：

```powershell
node tests/expansions-help-unit-test.cjs
node tests/expansions-catalogue-unit-test.cjs
node tests/tuscany-ui-unit-test.cjs
node tests/polling-unit-test.cjs
```
