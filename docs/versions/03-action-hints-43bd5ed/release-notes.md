# Release Notes：行动高亮与条件提示

提交 `43bd5ed`（2026-09-11）让玩家在提交动作前看到可用行动、资源缺口和访客分支限制，同时保持服务器为最终规则判定者。

## 中文

### 新增

- 行动、手牌和选择选项的当前玩家专属原因字段。
- 可派遣/条件不足的视觉状态和手牌联动高亮。
- 鼠标、键盘和手机均可访问的紧凑提示。
- 费用、建筑、酒窖、田地、工人、配方和资源数量的前置条件反馈。

### 验证与限制

- 新增领域可用性测试，并扩展真实浏览器 usability、规则和访客场景。
- 提示不执行规则、不预测随机结果；最终合法性仍由服务端行动校验决定。
- 本提交没有语义化版本号或 Git tag。

## English

### Added

- Player-scoped reason maps for actions, cards, and pending-choice options.
- Available/unavailable board highlights, matching hand-card emphasis, and compact pointer, keyboard, and touch hints.
- Pre-submit feedback for common resource, building, field, worker, recipe, and visitor prerequisites.

### Validation and limits

- Domain and browser coverage was extended for read-only hints, disabled options, focus behavior, and submit readiness.
- Hints do not execute game actions or guarantee every parameter combination; the server remains authoritative.
