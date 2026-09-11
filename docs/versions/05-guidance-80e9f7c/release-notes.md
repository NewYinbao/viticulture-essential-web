# Release Notes：种植检查、规则速查与新手引导

提交 `80e9f7c`（2026-09-11）让种植条件可在失败前检查，并加入不打断棋盘操作的规则速查和可选新手引导。

## 中文

### 新增

- 缺藤、缺设施、田地容量不足或无工人时仍可打开种植条件检查。
- 藤卡、建筑需求和三块田地余量的关联说明。
- 八主题 EE 规则速查，以及从行动面板定位相关规则的入口。
- 四步界面导览和依据当前局面的短提示；偏好只保存在当前浏览器。

### 验证与限制

- 新增三组真实浏览器场景，当版累计 14 组体验检查。
- 条件检查和帮助页不会提交游戏行动或清除当前行动草稿。
- 本提交没有语义化版本号或 Git tag；范围仍为 EE 本体。

## English

### Added

- Read-only planting requirement inspection for missing cards, buildings, workers, and field capacity.
- An eight-topic Essential Edition rules reference that stays alongside the board.
- An optional four-step tour and contextual guidance based on season, turn, pending choices, and resources.

### Validation and scope

- Three new real-browser scenarios brought the recorded usability total to 14.
- Help and inspection flows do not submit game actions or discard an in-progress action draft.
- Expansion configuration, player passwords, and public sharing were outside this commit.
