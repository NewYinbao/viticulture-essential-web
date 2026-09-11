> 提交快照 · [f9227db](https://github.com/NewYinbao/viticulture-essential-web/commit/f9227db)。这是该版原有记录，不使用后续版本的测试结果补写。

# 构建与验证

Go 1.25+。发行程序没有 Node 运行时依赖；开发 UI 测试需要 Node 22+ 和 Chromium。

```powershell
go test -tags "ee_rule_audit audit_diff" ./...
go vet ./...
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

## 本次本地结果

- 全部默认测试、ee_rule_audit、audit_diff，以及 go vet 通过。
- usability-test：7 组操作体验场景，桌面／手机视口、工人选择、手牌筛选、拔藤和酿酒提交、局面更新及恢复入口通过。
- rules-fix-browser-test：5 组规则与访客交互回归通过。
- visitor-browser-test：141 个初始选项场景通过，含多个需要续步或多人回复的访客；这是夹具测试，不是 141 局自然游戏。
- natural-full-game-test：未注入资源的一局随机牌序游戏完成（11 年），验证 HTTP/SSE 手牌隔离、过期 revision、轮次、断线和重启恢复、20 分年末终局。
- discard-browser-test：实际建房、父母选择、四季流程、桌面和手机弃牌验证通过。
- Windows 本机 exe 的 health、内嵌网页、鉴权、创建／加入／开始、落盘和重启恢复全部通过，使用临时存档。

本地 WSL 未配置 C 编译器，不能运行 Go race detector；已在 GitHub Actions 的 Ubuntu 作业配置 `go test -race`。CI 配置存在不代表远端已执行成功，需以仓库实际运行结果为准。

修改网页后必须重新构建程序并刷新浏览器，网页资源由 go:embed 内嵌。自然对局牌序随机；失败时先检查 artifacts 报告区分策略覆盖不足、测试失效和规则错误，不能删除断言绕过问题。
