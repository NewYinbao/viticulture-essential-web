# Release Notes：工程重构与 EE 规则收口

提交 `f9227db`（2026-09-11）完成服务、规则、存储、网页和测试目录的工程重构，并关闭三项有规则依据的 EE 问题。

## 中文

### 新增与改进

- 将游戏领域、HTTP 服务、JSON 存储和内嵌网页拆成独立包与目录。
- 新增统一构建/检查脚本、项目文档和 GitHub Actions 配置。
- 修复公共 Plant 行动错误提供拔藤、Papa 赠礼选择顺序和对手手牌类型数量显示。
- 改善桌面/窄屏导航、手牌筛选、工人格位、行动前置条件、种植与酿酒预览，以及过时表单处理。

### 验证与限制

- 默认测试、规则审计、差异审计和 vet 当时均记录为通过。
- 真实浏览器覆盖基础操作、规则修复、访客初始选项、弃牌和一局自然终局。
- 本提交没有语义化版本号或 Git tag；归档编号 02 只表示历史顺序。
- 范围仍是 EE 本体，不含扩展、密码和公网分享。

## English

### Added and changed

- Split the game domain, HTTP server, JSON storage, embedded web client, tests, scripts, and documentation into explicit project areas.
- Fixed public Plant uprooting, Papa choice order, and public hand-type counts while preserving private card identities.
- Improved navigation, hand filtering, worker-slot rendering, action prerequisites, planting/winemaking previews, and stale-form recovery.

### Validation and scope

- The commit records passing default, rule-audit, audit-diff, vet, browser, native-smoke, and one natural-game run.
- Coverage remained Essential Edition only. No semantic version or Git tag was created.
