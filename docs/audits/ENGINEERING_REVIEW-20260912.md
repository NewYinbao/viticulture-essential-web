# 工程质量专项审查报告

审查对象：`C:\Users\admin\Games\Viticulture\essential-web` 当前工作树（含未跟踪文件）。

## 结论

未发现需要改变规则引擎职责划分、Go/JS 行为契约或生产运行流程的新增功能缺陷。确认了 4 项有实际维护收益的局部问题，并在只读生产副本之外的 TEMP 副本完成修复验证：

1. 10 个 Go 文件未通过 `gofmt -l`。格式修复严格采用 gofmt 输出，无逻辑改动；10 个文件的预期 SHA-256 与 `artifacts/continuation-20260911/gofmt-expected.json` 的 `expectedAfter` 全部匹配。
2. `npm run test:units` 确实产生 `MODULE_TYPELESS_PACKAGE_JSON` 警告。TEMP 副本的 `package.json` 增加 `"type": "module"` 后警告消失；现有 `.cjs` CommonJS 脚本仍可被 `require` 正常加载。
3. `tests/e2e/discard-browser-test.cjs` 与 `tests/e2e/screenshot-stages.cjs` 的证据备注仍写“76 visitor effects and expansions remain unimplemented”，与当前交付范围冲突。备注已改为准确描述各自只覆盖弃牌/自然局截图，扩展与访客覆盖由专门验收套件记录。
4. `Start-Viticulture-Expansions.cmd` 指向旧的 `dist/Viticulture-Expansions.exe`，却提示 Moor/Rhine 仍禁用；这不能代表当前源码或最新 Continuation EXE。提示已改为明确的 legacy development build，不自动升级，并指向 `Start-Viticulture-Continuation.cmd` 与 `scripts/build-continuation.ps1`。监听端口、存档目录和旧 EXE 均未改动。

另外，`scripts/check.ps1` 已增加 Go 格式检查；README 与 `docs/testing.md` 的基础验证入口改为调用该检查脚本，并把本地结果明确写为包含 gofmt。没有批量重排旧式 `.cjs` 单行测试代码，因为那会制造噪声而没有对应维护收益。

## 格式修复文件

`internal/game/placement_view.go`、`internal/game/special_workers.go`、`internal/game/visitor.go`、`internal/game/visitor_moor.go`、`internal/game/visitor_moor_availability.go`、`internal/game/visitor_rhine_catalog.go`、`internal/game/visitor_rhine_http_test.go`、`internal/game/visitor_rhine_scenarios_test.go`、`internal/game/visitor_sequence.go`、`internal/game/workers.go`。

生产快照的 195 个逐文件 SHA-256 仍全部匹配 `final-source-snapshot.json`；生产工作树未被本专项写入。补丁共涉及 17 个文件，未包含构建产物、存档、真实服务或提交操作。

## 验证

- 生产快照核对：195/195 文件匹配，digest `91036c232fa80d3697d6f45c380ace1e86fe3e27e5d79f6c6163679230c0e438`。
- 生产快照只读检查：`go test ./...`、`go vet ./...`、网页 Prettier 检查和 4 组 JS 单测通过；JS 单测原先的模块类型警告已被复现。
- TEMP 修复副本：显式生产包集合的 `go test`、`go vet`、`gofmt -l cmd internal web` 通过；4 组 JS 单测通过；全部 JS/CJS `node --check` 通过；CommonJS `require('./tests/e2e/runtime.cjs')` 通过且无模块类型警告。
- `git apply --check --whitespace=error engineering-review.patch`：通过。
- 10/10 gofmt 预期输出哈希通过。

未重复执行 24 自然局或完整浏览器矩阵；该专项只做工程/维护性局部验证。未运行 race detector（环境没有 WSL C 编译器），未构建或覆盖任何旧 EXE。

## 补丁

- 原补丁保留于独立审查任务的本地 outputs，未作为源码文件提交。
- SHA-256：`3cc0611d5e35002d3ba233a61f50a7d92dbee83df72f4f8632b9b8390c3c386b`


## 合入后复核补记

主任务已合入上述补丁。独立工程任务再次确认模块声明、检查入口、旧启动器与两份报告备注均已更新，生产 gofmt 检查和 10/10 预期输出哈希匹配，结论为本专项无已知未解决问题。

最终源码 digest：`5a60cb1aee782e5f4fa8ae15d9988d64367474d3d45d077dd1bda7ec31c7232e`。主任务随后在生产源目录通过全包 Go/审计/vet、全部 JavaScript 单元/格式/语法检查，并用 TEMP 里的重建 EXE 通过 Farmer 3 金币训练的真实 Edge 跨季与重启流程。重建 EXE 与完整功能验收 EXE 的 SHA-256 完全相同，原完整浏览器证据仍对应同一程序。
