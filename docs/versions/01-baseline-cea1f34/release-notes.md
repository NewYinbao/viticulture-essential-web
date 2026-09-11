# Release Notes：既有 EE 实现快照

提交 `cea1f34`（2026-09-10）保存了目录重构前的 Essential Edition Web 实现，作为后续版本比较和历史追溯的基线。

## 中文

### 包含内容

- EE 本体 2–6 人局域网游玩、固定卡库、独立手牌、SSE 同步和本地 JSON 存档。
- 76 张访客及多步选择、多人响应和规则专项测试。
- 原生网页界面、中文卡名与内嵌本地 SVG。
- 规则审计、差异测试和修复记录的原始版本。

### 已知限制

- 这是开发快照，不是正式语义化版本或 Git tag。
- Plant 无轭拔藤仍是已知未决差异；当时的 `audit_diff` 因此不是全绿。
- 不含 Tuscany、Moor、Rhine、特殊工人扩展、玩家密码或公网分享。

## English

### Included

- A pre-restructure snapshot of the 2–6 player Essential Edition web implementation.
- Fixed decks, private hands, SSE updates, local JSON saves, 76 visitor effects, and the original browser UI.
- The original audit, rule-fix, and verification records used by later work.

### Known limitations

- This commit is a development snapshot, not a tagged or semantic release.
- The public Plant action still exposed uprooting; the corresponding audit-diff test intentionally remained failing.
- Tuscany, Moor, Rhine, special-worker modules, player passwords, and public sharing were outside this version.
