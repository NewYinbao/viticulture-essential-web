# Go EE 核心规则差分审计

## 结论及范围
只读审计生产代码、既有测试和 Java。仅新增本报告与 audit_core_diff_test.go（build tag audit_diff），不改生产/既有测试/存档。
发现 **4 个已复现核心交互错误/差异**，另有 **1 个官方 FAQ 确认、Java 也有同样简化的可选奖励缺口**。父母牌 18+18 资源表逐项未发现错误。以下 Java 行号以 /tmp/ViticultureBoard.java、/tmp/PlayerBoard.java、/tmp/ViticultureChip.java 为准；Constants 使用项目 references/ViticultureConstants.java。
已复现表示运行 Go 测试并对照 Java 分支，未运行完整 Java 客户端。Boardspace 不是唯一规则裁判，该 Java 使用 Tuscany 主板，必须排除扩展特有差异。

## A. 已复现问题（双方精确行号）

### A1 高：取回轭上的工人后，空轭仍锁死
- Go ee.go:318–322 按 YokeUsed 拒绝进入，365–380 使用后置 true；visitor_sequence.go:196–205 Producer 取回只增加可用工人并删除 Seat，不清 YokeUsed。visitor_interaction.go:113–129 Motivator 从轭取回大工人亦遗漏；仅 engine.go:72 年终重置。
- Java ViticultureBoard.java:10890–10905 允许 Producer 取私板工人；PlayerBoard.java:479 workerCells 包括 yokeWorker；ViticultureBoard.java:6341–6355 执行取回，9335–9341 只要求轭已建且 yokeWorker 空。Motivator 的 5468–5474 移走 grande 也会释放格子。
- 轭是私人占位行动，而非风车/品酒室式年度奖励。取回后应能再放工人；年度同田最多收获一次仍保留，可拔藤或收另一田。
- TestAuditDiffYokeReopensAfterProducer：取回成功，后续 placement 却返回“需要未使用的轭”。Motivator 路径为静态确认，未另立复现。

### A2 高：Producer 禁止取回整个触发行动区域上的所有工人
- Go visitor_sequence.go:177–189 注释称排除当前工人，实际用 id == r.Context.Space 拒绝整区，连以前回合放的工人也禁止。
- Java ViticultureBoard.java:10860–10864 明确写官方裁定只排除打出此卡的那个工人；10874–10879 精确排除 lastDroppedWorker 和 lastDroppedWorkerIndex，而非整个行动区。
- 3 人局，自己之前已放在夏访客格2，以另一个格打 Producer，应可取格2旧工人。TestAuditDiffProducerMayRetrieveEarlierWorkerOnSameAction 返回“只能取其他行动工人”。测试直接调用效果函数，夹具只包含已存在的旧 Seat，无 HTTP/存档副作用。
- 与 A1 独立：A1 是取回后状态不释放，A2 是合法目标过度过滤。

### A3 中：Landscaper 交换已种藤，错误重验两田所有藤的建筑
- Go visitor_sequence.go:92–121 已实现 swap，**不是遗漏交换功能**；111–117 错误重验交换后两田所有藤的棚架/灌溉。
- Java ViticultureBoard.java:10558–10570 canSwap 仅检交换后容量；10576–10598 生成交换；4600–4611 为同一张 EE Landscaper。
- 可达场景：用 Horticulturist/Sharecropper 无视建筑种下需棚架的藤，再用 Landscaper 交换两田藤，两田均不超容量。交换不是重新种植，不应重验建筑。
- TestAuditDiffSwapNeedsCapacityNotBuildings 返回“缺种植建筑”。用例使用可达局面夹具，未运行之前的完整访客链。

### A4 中：主板 Plant 额外允许无需轭的 uproot
- Go ee.go:405–407 给 plant 增加 Mode=uproot 入口，直接 uproot，无需建轭或可种手牌。engine.go:305–321 正常消耗工人占 Plant 格；visitor_sequence.go:237–253 Manager 复用 performPlacement，也继承额外权限。
- Java ViticultureBoard.java:3355–3371 PlantWorker 只进入种植状态；3383–3395 在 PlayerYokeWorker 才提供 HarvestOrUproot/Uproot。6358–6383 的普通拔藤回手与 Go ee.go:476–485 一致，错误在入口权限。
- TestAuditDiffPlantMustNotOfferUproot 经过真实 Room.apply，无轭、有田上藤、请求主板 Plant uproot，Go 接受。
- 确定是额外动作的代码差异；按 Java EE 共享基础动作应去除该入口。但本地官方规则 PDF 损坏，未做到规则书原文交叉认证，不把此项夸称为已读原版 PDF 确认。

## B. FAQ 确认的共同简化：固定奖励不能放弃
- 官方快照 /tmp/ee-audit/faq.txt 全部在第1行；问题标题 When can I place a worker on the middle action space in a 3-6 player game? 明确必须做行动但 bonus is optional，可在行动前后领取。
- Go model.go:130–150 Action 没有奖励放弃字段；ee.go:356 按格位自动 bonus，455–472 强制折扣、额外抽牌、钱/VP，400–402 买卖田强制 VP。数量型加成可少做，固定收益却不能拒绝。
- 三人占抽藤奖励位不能只抽1张；拒绝 VP 可推迟终局，不应一概认为无关紧要。
- Java ViticultureBoard.java:3223–3237、3262–3280 非市场抽牌也直接按 bonus 抽2。**这不是 Java 比 Go 更完整的差分，而是两者对官方 FAQ 的共同简化。** 未为不存在的 API 编造失败测试。

## C. 已核对正确 / 不应重复误报

| 范围 | Go | Java / 官方 FAQ 与结论 |
|---|---|---|
| 初始2普通+1大工人 | model.go:230–235 | PlayerBoard.java:667–672 一致 |
| 训练延迟、总工人6 | engine.go:421–430、67–69 | references/ViticultureConstants.java:654 上限6；普通训练不立即增加 Workers，年末恢复；灰工人不增 TotalWorkers |
| grande 优先空位、溢出无奖励 | ee.go:324–356 | ViticultureBoard.java:9607–9643 空格优先，无可放才 overflow；普通工人不能溢出 |
| 轭的私人性 | ee.go:318–323 | PlayerBoard.java:458 每人独立 yokeWorker；Go 聚合 Space 但按玩家标记，基本私人性正确，仅取回重置有 A1 |
| 2 / 3–4 / 5–6 人格数 | model.go:240–247、ee.go:356 | ViticultureBoard.java:9559–9564、9614 人数限格；Go 1/2/3格，2人无奖励 |
| 夏天结束不回收工人 | engine.go:43–55、67–69 | FAQ Do I get my workers back at the end of summer? 明确直到年末，Go 正确 |
| 秋季小屋双牌同时决定颜色 | engine.go:45–51、ee.go:189–192 | FAQ Do I have to draw 2 of the same type of visitor card if I have the cottage? 必须同时决定，不能看首张再选；Go 一次三组合正确。Java 3915–3935 为 Tuscany 转季，不能硬套 |
| 年末资源/训练恢复/年度标记 | engine.go:60–79 | ViticultureBoard.java:3874–3878 有对应效果；Go 最终年也结算后排名，控制测试通过 |
| 起始标记逆时针 | engine.go:122，春季266–268 | FAQ Why do players choose their wake-up times clockwise… 明确相反方向；当前正确。旧 EE_RULE_AUDIT.md:178 仍描述 +1 旧错误，不是当前缺陷 |
| 18 Mama | ee.go:115、137–142 | PlayerBoard.java:675–764 逐一一致，17/18两卡+2钱；ViticultureChip.java:732–740 编号1–18 |
| 18 Papa | ee.go:123、143–145、173–187 | PlayerBoard.java:775–915 全部基础钱、建筑/工人/VP、替代钱一致；15/16新工人立即可用，17/18给1VP；ViticultureChip.java:745–753 编号对应 |
| 种藤/卖田/拔藤结果 | engine.go:372–404、ee.go:381–399、476–485 | 校验未卖田、容量和建筑，只卖空田；轭拔藤回手与 Java 6358–6383 一致。交换重验建筑才是 A3 |
| 平分比较 | engine.go:76–118、ee.go:258–266 | PlayerBoard.java:1000–1009 顺序同为VP/钱/酒值/葡萄值。Go 完全相同并列比 Java 2054–2067 单赢家更稳，不能照搬 Java |

## D. 生命周期：确定差异，不直接当 EE 错误
1. Go ee.go:243–256 先所有人弃至7，再 engine.go:62–79 回收、收入、陈年；Java ViticultureBoard.java:3874–3884 每人冬季退出时回收、陈年、收入、弃牌。纯本体年终没有额外收益选择时这些操作基本可交换，未复现最终状态差异。流程时点不同，不夸大为胜负错误。
2. Java 3879–3898 达其阈值时跳过弃牌及下年起床，Go 最终年仍弃牌。Java 是 Tuscany，也可能仅优化终局无影响步骤，不能直接判定 Go 必须删除弃牌。
3. Go engine.go:76 年末按当前VP>=20检查，不在中途锁存 FinalYear。Java 2025–2041、3879 也是检查当前最大分。中途达到再花回阈值以下的问题，不能仅凭没有即时锁存就报错，正式 EE 裁定仍宜补有效规则 PDF。

## E. Tuscany、可选玩法、信息呈现
- **20 vs 25 不是 Go 错误**：Go engine.go:76 为 EE 20；references/ViticultureConstants.java:653 的25和 ViticultureBoard.java:2036–2044 influence 结分、3942–4004 含橙牌/星的多季起床奖、9559–9564 四季主板均显示 Tuscany。春秋无主板放工、无星/结构牌/特殊工人不是本体遗漏。
- references/ViticultureConstants.java:42–51 明列 Option：GreenMarket、DraftPapa、UnlimitedWorkers、ContinuousPlay、DrawWithReplacement、LimitPoints、PurpleMarket、DraftStructures、ExtraSpecial。Go 不实现不属 EE 必需缺失，不能强加父母二选一、无上限工人或有放回抽牌。
- Go model.go:260–274 隐去对手手牌、公开数量和庄园资源，ee.go:104–112 非选择者只见待决玩家和 kind，基本隐私正确。父母开局资源展示不据此认定泄密。
- **弃牌顶信息待核对，不列确定缺口**：Go ee.go:94–98 / model.go:274 全局 view 只给弃牌数量；Innkeeper 可能有专用 schema，不能只凭无全局字段认定盲选。Java 5815 及 references/ViticultureConstants.java:803 有 Pick2Discards。尚须 viewer/UI/schema 核对，可能仅可选 UX。

## F. 验证与限制
- /opt/viticulture-toolchain/go/bin/go test ./... ：PASS，ok vineyard 0.559s。
- /opt/viticulture-toolchain/go/bin/go test -tags audit_diff -run ^TestAuditDiff -count=1 -v . ：4个预期规则测试 FAIL（A1–A4），1个基础工人/终局/格数控制 PASS。红灯表示已复现现有差异，build tag 保证默认测试不受影响。
- 专用 read_file/search_files/write_file 因本任务工具环境路径映射/固定 D盘工作目录失效；先尝试失败后，用 terminal Python 只读读取及仅写两个获准文件。未修改生产、既有测试、存档。
- references/rules.pdf 空，/tmp/ee.pdf 实为 HTML，/tmp/ee-audit/rules.pdf 交叉引用/页树损坏而无法 pdftotext。使用本地官方 FAQ 与 Java，不能声称完成 PDF 逐页认证。
- 报告第一次写入遇工具外层 shell 对 Markdown 反引号做替换，已重写为不含反引号文本并回读核验；失败命令仅为不存在的路径/标签调用，没有执行生产写入。
