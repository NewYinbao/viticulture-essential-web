# Release Notes：扩展、玩家密码与上下文界面

提交 `563db1b`（2026-09-12）完成 Tuscany Essential、建筑、特殊工人、Moor 与 Rhine 访客模块，加入座位密码、按配置显示的上下文界面及可选 Cloudflare 分享。

## 中文

### 新增

- 24 种 EE/Tuscany × 建筑 × 特殊工人 × 访客牌组开局配置。
- Tuscany 四季主板、影响力与个人过季；36 张建筑和 11 类特殊工人。
- Moor 40 张访客加入 EE 牌库；Rhine 80 张访客作为替换牌组，并按主板过滤依赖牌。
- 房间座位密码、同名验证恢复、改密、退出、旧会话撤销和旧存档安全登记。
- 按本局配置显示的行动、规则、新手提示和专用卡牌控件。
- 独立 Continuation 启动器/存档，以及显式开启的 Cloudflare Quick Tunnel 分享。

### 关键修复

- 阻止客户端伪造行动格奖励。
- 修复 Planner × Farmer 折扣预约的可行性、私有投影和重启恢复。
- 修复多项建筑、特殊工人、Moor/Rhine 嵌套、结构占田、培训与年末恢复边界。

### 验证

- 完整验收 12/12；24 种真实配置与自然终局、36 张建筑 UI 场景、密码 6 组流程和模块显示检查通过。
- 两份独立功能复核和一份工程审查完成；GitHub Actions run 34629523327 成功。
- 覆盖并非逐卡浏览器穷举；自然局策略不主动打访客，详细边界见本版本 [validation.md](validation.md)。

### 范围

- 不含 World、Bordeaux、旧 Tuscany 额外模块、促销或 Automa。
- 旧服务与旧存档不会自动升级；续作程序使用独立端口和目录。
- 本提交没有语义化版本号或 Git tag。

## English

### Added

- 24 configurations across the Essential Edition/Tuscany board, Structures, Special Workers, and EE, Moor, or Rhine visitors.
- Tuscany seasons and influence, all 36 structure cards, all 11 special-worker types, 40 Moor visitors, and 80 Rhine replacement visitors.
- Seat-bound player passwords with secure legacy-seat enrollment, session rotation, password changes, and logout.
- Configuration-aware controls, rules, and beginner guidance, plus an isolated Continuation launcher and opt-in Cloudflare Quick Tunnel sharing.

### Fixed

- Ignored client-supplied action-space bonus overrides.
- Preserved Planner/Farmer discounts through feasibility checks, private projection, persistence, and restart.
- Closed reviewed structure, worker, nested-visitor, field-occupancy, training, and year-end recovery issues.

### Validation and scope

- The 12-step final suite passed, including all 24 configurations and natural finishes, 36 seeded structure UI scenarios, password flows, module visibility. Independent reviews and GitHub Actions run 34629523327 also passed.
- Visitor effects were not exhaustively exercised in browsers; see [validation.md](validation.md) for exact limits.
- No semantic version or Git tag was created for this commit.
