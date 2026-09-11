# Release Notes：公共棋盘内卡片面板

提交 `1f0dd8b`（2026-09-11）把公共行动从遮罩弹窗迁入棋盘，并用可检查的卡片替代下拉表单。

## 中文

### 新增与改进

- 棋盘内展开工人、格位、奖励、建筑、手牌和资源选择。
- 八座基础建筑显示费用、折扣、拥有状态和前置条件。
- 种植、订单与酿酒改用卡片选择，并提供实时容量、酒种和品质预览。
- 手机布局增加面板内滚动、固定确认区和庄园/行动往返。
- 取消、SSE 更新和过时 revision 会安全清理草稿。

### 验证与限制

- 真实服务/浏览器验证了基础公共行动、桌面和 390px 布局及过时状态处理。
- 本版尚无规则速查、新手导览或扩展模块。
- 本提交没有语义化版本号或 Git tag。

## English

### Added and changed

- Replaced full-screen public-action dialogs with board-local card panels.
- Added card-based worker, slot, reward, building, hand, field, grape, and wine selection.
- Added live building-cost, planting-capacity, order, and winemaking previews.
- Preserved access to the estate and hand, with mobile panel scrolling and stale-draft cleanup.

### Validation and scope

- Real-server browser checks covered the core action flows, desktop/mobile viewports, cancellation, SSE refreshes, and stale revisions.
- Rules reference, beginner guidance, and expansion modules were not yet part of this commit.
