# HTTP 与玩家视角

规则集 `ee-base-v1`，JSON API。接口路径和旧存档字段保持兼容。

| 方法与路径 | 用途 |
| --- | --- |
| GET /api/health | 存活检查 |
| POST /api/create | `{name}` 创建房间，返回 code/token |
| POST /api/join | `{name,code}` 加入房间，返回 code/token |
| GET /api/state | 当前玩家视角 |
| GET /api/events | 按玩家投影的 SSE 完整状态 |
| POST /api/action | 带 revision 的行动或选择 |

会话使用 `Authorization: Bearer <token>`；GET 也支持 token 查询参数以供 EventSource 使用。不要把真实 token 写进 issue、截图、日志或源码。

状态包括 code/year/phase/revision/turnId/youId、公开 players/spaces、本人 hand 和 pendingChoice。`players[].handCount` 是总数，`handCounts` 为 vine/order/summer/winter 的公开数量；其他人的牌面、牌 ID 和牌堆顺序不公开。仅选择队列的首位拥有者能看到并提交 choiceId/options。

流程：lobby → setup → wake → summer → fall → winter → year_end → wake/finished。legal 控制当前玩家可执行的操作，但不表示每个具体资源组合都合法。

行动提交携带最新 revision；旧 revision 返回 409，规则错误返回 400，失败不会提交部分状态。普通 place 使用 space/large/slot；slot=0 自动，slot=-1 只允许没有空格时的大工人溢出。3 人以上第 1 格提供奖励；`declineBonus:true` 可放弃奖励。

常用参数：

- start、pass；wake 带 slot，起床第 5 行还带 color=summer|winter。
- plant 带 cardIds/fields（同长度，田地索引从 0 开始），不接受 uproot。
- yoke 带 mode=harvest|uproot、field；uproot 还需要 cardId。
- harvest 带 fields；make_wine 带 recipes，每个子数组是行动前葡萄列表的索引。
- sell_grapes 带 grapes；出售／买回田地使用同一 space，mode=sell_field|buy_field 和 field。
- build 带 building；fill_order 带 cardId/wineIds。
- choose 带 choiceId 和当前选择所需 option/cardIds 等参数。Papa 的 gift 与 coins 二选一；秋季有小屋时一次选择两张访客的颜色组合；年末 discard 必须提交要求数量的不同 cardIds。

旧的 [访客 API 细节](history/VISITOR_API.md)和[访客 UI 约定](history/VISITOR_UI_CONTRACT.md)保留作详细背景，最新规则结论以 [rules-review.md](rules-review.md) 为准。
