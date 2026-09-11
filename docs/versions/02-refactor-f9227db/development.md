# 开发记录：工程重构与 EE 规则收口

- 提交：`f9227db2ed09b8c37df9fe4eabb08ee497589a30`
- 日期：2026-09-11 00:33:31 +08:00
- 原提交标题：`refactor: separate game server storage and web; reconcile EE rules and improve play flow`
- 归档编号：02；编号不表示 semver，也没有对应 Git tag。

## 工程重构

本提交把原先集中在仓库根目录的实现整理为清楚的运行边界：

- `cmd/viticulture` 负责程序入口、监听参数和关闭流程。
- `internal/game` 负责领域模型、行动、季节、资源、工人、待决选择和访客状态机。
- `internal/server` 负责 HTTP、会话、revision、SSE 和应用事务。
- `internal/store` 负责 JSON 快照读写。
- `web/static` 与 `web/embed.go` 负责可内嵌的原生前端和 SVG。
- 测试、构建脚本及文档分别迁到 `tests`、`scripts`、`docs`。

重构保留既有 API 路径和存档字段，并增加持久化/回滚验证。请求在候选状态成功落盘后才提交并发送 SSE；规则错误恢复原状态。

## 规则与隐私修正

- 根据 EE 说明书移除公共 Plant 行动的拔藤入口；拔藤仍可由轭或允许该效果的访客执行。
- Papa 赠礼选择队列改为从随机起始玩家开始按顺时针顺序，而不是加入顺序。
- 对手视图增加四类公开手牌数量 `handCounts`，但仍隐藏牌面、牌 ID 和牌堆顺序。
- 保留行动格奖励可放弃、`up to` 至少执行一次、田地年度收获限制及订单酒种匹配等已核对行为。

## 操作体验

- 移除重复版本横幅，折叠牌堆/家族区，并增加桌面和窄屏导航。
- 手牌按类型筛选并显示数量；恢复入口可返回上次座位。
- 工人按真实格位绘制，奖励格、无限金币行动和私人轭分别显示。
- 表单在工人、轭或基础资源不足时给出直接原因；大工人、奖励默认值、田地/藤选择和酿酒预览得到修正。
- 新 revision 到达或收到 409 时清理旧表单并刷新状态。

## 本版记录入口

- [rules-review.md](records/rules-review.md)：Plant、Papa 与手牌公开规则依据。
- [ui-review.md](records/ui-review.md)：当版 UI 问题与修改。
- [testing.md](records/testing.md)：当版构建和验证记录。
- `../01-baseline-cea1f34/records/audits/`：首个提交已经存在、在本提交中仅移动的规则审计；本阶段的 `rules-review.md` 链接回这些历史审计。

当版架构差异可由提交 `f9227db` 中的 `docs/architecture.md` 查阅；归档后的当前架构说明属于 `docs/guides/`。

后续 `43bd5ed` 才加入统一的行动/卡牌/选项原因投影和高亮提示；本版本不包含那些能力。
