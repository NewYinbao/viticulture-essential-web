# 构建与测试

历史结果按提交保存在[版本索引](../README.md)，当前指南只说明如何运行。

需要 Go 1.25+；浏览器测试还需要 Node.js 22+。从项目根目录执行：

```sh
go test -tags "ee_rule_audit audit_diff" ./...
go vet ./...
npm ci
npx playwright install chromium
npm run test:units
npm run format:check
npm run test:ui
npm run test:rules-ui
npm run test:visitors
npm run test:game
npm run test:discard
```

Windows 构建续作程序：

```powershell
powershell -NoProfile -File scripts/check.ps1
powershell -NoProfile -File scripts/build-continuation.ps1
npm run test:expansions
```

完整扩展验收需要 Windows Edge。测试使用隔离服务与测试存档，报告写入 artifacts；不要把真实存档传给测试。部分 E2E 可通过 GO_BIN 指定 Go 路径。Linux 可执行 go test -race ./...，需可用 C 编译器。

最新功能提交的实际结果、CI 和非穷举边界见[563db1b 验证记录](../versions/06-expansions-563db1b/validation.md)。
