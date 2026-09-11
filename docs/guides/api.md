# HTTP 与玩家视角

JSON API。旧存档缺省使用本体规则；未设密码的旧座位需先完成本机验证码登记。

| 方法与路径 | 用途 |
| --- | --- |
| GET /api/health | 存活检查 |
| POST /api/create | `{name,password}` 创建房间，返回 code/token |
| POST /api/join | `{name,code,password}` 加入或验证密码恢复原座位，返回 code/token |
| POST /api/password | 设置或修改本人座位密码，返回轮换后的 token |
| POST /api/logout | 撤销当前会话 |
| GET /api/state | 当前玩家视角 |
| GET /api/events | 按玩家投影的 SSE 完整状态 |
| POST /api/action | 带 revision 的行动或选择 |

会话使用 `Authorization: Bearer <token>`；GET 也支持 token 查询参数以供 EventSource 使用。不要把真实 token 写进 issue、截图、日志或源码。

密码为 8–128 个 Unicode 字符。密码记录属于房间中的玩家座位，同名不构成认证。已存在座位登录和改密会撤销该座位所有旧 token；SSE 收到 `session-ended` 后必须清空私密状态。`POST /api/password` 提交 `password`、已保护座位的 `currentPassword`；旧座位首次登记还须 `enrollmentCode`，此码仅由服务器本机控制台提供。未完成登记时，state/events/action 返回 403，不返回手牌。登录限流返回 429 与 Retry-After；认证失败不改变游戏状态。密码、哈希和登记码均不在游戏 View 中，View 仅含本人的 `passwordSet` 布尔值。

状态包括 code/year/phase/revision/turnId/youId、公开 players/spaces、本人 hand 和 pendingChoice。空手牌固定为 `[]`。`players[].handCount` 是总数，`handCounts` 为各牌种的公开数量；其他人的牌面、牌 ID 和牌堆顺序不公开。仅选择队列的首位拥有者能看到并提交 choiceId/options。对手只收到待决玩家与选择种类；预约仅公开位置和工人身份，隐藏资源参数。

actionReasons / cardReasons / optionReasons 是当前玩家的操作提示；键分别对应行动、本人手牌和当前选择选项，非空值表示已知不满足的前置条件。空值不是任意资源组合都合法的承诺。提示计算不执行行动，也不提供未来牌序信息。

`bonusOverride` 只用于服务端内部续步和存档。公共行动入口会忽略客户端提交的值，奖励必须由真实格位或已保存的特殊工人选择确定；Planner 预检与实际执行共用同一份预约状态投影。

`workerPlacements` 按行动空间列出服务端判断可进入的工人、格位、奖励与通行费；最终仍以带 revision 的行动验证为准。公开 `config` 包含 `board:ee|tuscany`、`structures`、`specialWorkers`、`visitors:ee|ee_moor|rhine`。房主仅在大厅可提交 `type:configure` 与完整 config；开局后锁定。Rhine 替换其他访客，Tuscany 专用 4 张牌按主板过滤。

流程：lobby → setup → wake → summer → fall → winter → year_end → wake/finished。legal 控制当前玩家可执行的操作，但不表示每个具体资源组合都合法。

Tuscany 使用春、夏、秋、冬及个人过季/冬末回收流程；是否显示影响力、结构牌、特殊工人、牌组规则和引导应完全依据本局 config，不从全局目录推断。

行动提交携带最新 revision；旧 revision 返回 409，规则错误返回 400，失败不会提交部分状态。普通 place 使用 space/large/slot；slot=0 自动，slot=-1 只允许没有空格时的大工人溢出。3 人以上第 1 格提供奖励；`declineBonus:true` 可放弃奖励。

常用参数：

- start、pass；wake 带 slot，起床第 5 行还带 color=summer|winter。
- plant 带 cardIds/fields（同长度，田地索引从 0 开始），不接受 uproot。
- yoke 带 mode=harvest|uproot、field；uproot 还需要 cardId。
- harvest 带 fields；make_wine 带 recipes，每个子数组是行动前葡萄列表的索引。
- sell_grapes 带 grapes；出售／买回田地使用同一 space，mode=sell_field|buy_field 和 field。
- build 带 building；fill_order 带 cardId/wineIds。
- choose 带 choiceId 和当前选择所需 option/cardIds 等参数。Papa 的 gift 与 coins 二选一；秋季有小屋时一次选择两张访客的颜色组合；年末 discard 必须提交要求数量的不同 cardIds。

旧的 [访客 API 细节](../versions/01-baseline-cea1f34/records/VISITOR_API.md)和[访客 UI 约定](../versions/01-baseline-cea1f34/records/VISITOR_UI_CONTRACT.md)保留作详细背景，本体重构结论见 [rules-review.md](../versions/02-refactor-f9227db/records/rules-review.md)，扩展与密码变更见[当前功能版本](../versions/06-expansions-563db1b/release-notes.md)及其验证记录。
