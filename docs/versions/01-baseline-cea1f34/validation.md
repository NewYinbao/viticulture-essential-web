# 验证记录：既有 EE 实现快照

- 提交：`cea1f3490423af384f06c8d439fd5d1e59dedea4`
- 证据来源：该提交内的 `README.md`、`RULE_FIX_REPORT.md` 与审计文档。

## 当时记录为通过的检查

| 检查 | 当时记录 |
| --- | --- |
| `go test -count=1 ./...` | PASS |
| `go test -tags ee_rule_audit -count=1 ./...` | PASS |
| `go vet ./...` | PASS |
| Windows amd64 构建 | PASS |
| 规则修复 Chromium/HTTP 定向脚本 | 5 项通过 |
| Windows 原生冒烟 | health、内嵌页面、认证、建房、加入、开始、落盘和重启恢复通过 |

快照 README 还保存了更早的 RC1 验收记录：76 卡初始效果/选项测试、141/142 个普通初始选项夹具、19 个特殊场景，以及一局无资源注入的自然完整对局。它们是提交前形成的既有证据，不代表在创建本 Git 快照时全部重新运行。

## 明确未通过或未执行

- `go test -tags audit_diff -count=1 ./...`：FAIL，只有 `TestAuditDiffPlantMustNotOfferUproot` 失败；该失败是当时保留的已知规则差异。
- 规则修复浏览器验证只重跑了五项定向场景，没有重新运行广泛浏览器套件。
- 手机相关结论来自 390px 浏览器视口，不是实体手机。
- 没有逐一穷举所有访客嵌套、人数、牌序和异常存档组合。

## 证据边界

本页只复述该提交中存在的验证记录，没有使用 `f9227db` 之后的通过结果。尤其不能把后续 Plant 修复后的全绿 `audit_diff`、提示 UI、扩展验收或 CI 结果归到本版本。
