# Viticulture Essential Edition 规则审计

范围：基础 EE 多人，38 夏卡 + 38 冬卡。不是 Tuscany/Plus。本文件是规范与后端验收清单，不代表实现已通过。仅交付文档，不修改 Go、web 或卡 JSON。

## 1. 来源与证据边界

- **R** EE 规则书：https://boardspace.net/viticulture/english/VitiRulebook_EssEd_2nd_r6.pdf ，现有 /tmp/ee-audit/rules.pdf。
- **F** 出版商 FAQ：https://stonemaiergames.com/games/viticulture/faq/ ，现有 /tmp/ee-audit/faq.html、faq.txt。FAQ 明说部分访客条目属于旧版，不能覆盖 EE 新卡面。
- **Snn/Wnn** 卡面照片及 OCR：/tmp/ee-audit/SummerVisitors_EssEd_2nd-Page-0000nn.jpg/.txt 与 WinterVisitors_EssEd_2nd-Page-0000nn.jpg/.txt。nn 为表中序号。完整来源 URL 见 /tmp/ee-audit/image-urls.json。远端根 https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/ ，夏目录 summer-cropped/，冬目录 wintervisitor-cropped/。
- **J** /tmp/ViticultureBoard.java、/tmp/ViticultureChip.java、/tmp/PlayerBoard.java。源：https://github.com/ddyer0/boardspace.net/tree/main/client/boardspace-java/boardspace-games/viticulture 。Board 夏卡 4436–4970，冬卡 5507–6141。
- **I** 出版商链接库存表：https://docs.google.com/spreadsheets/d/1mCRAuf99t6tuiPRWHbaffc07waRcMcDtx7FVwrSqcbw/edit#gid=0 ，本地 /tmp/ee-audit/inventory.csv。它证明名单和抽牌颜色，不是效果数据库。

本轮读取全部 76 张已有 OCR，交叉核对 J 的数值和状态以及 I 的颜色；没有声称完成全部图片目视复核。OCR 经常丢失金币/VP/收入图标。Queen 失分图标仍须补核，下文明确标注。商业图片不复制进项目。表中每行的 S/W 是具体卡面来源，J 的同编号 case 为数字交叉证据。

### 不得照搬 Java

- Board 1285 排除 Queen；6140–6141 没有 Queen 实现。基础 EE 仍必须有 Queen。
- Board 4796–4798 Importer 有 TODO，直接抽 3 冬卡，遗漏对手交牌分支。
- LimitPoints、无限工人、第五种结构牌、Tuscany 季节行动及回放 revision 不是基础 EE。
- Vendor 在 J 中自动给对手抽牌，卡面是 may，必须允许拒绝。
-  doTrainWorker 第二参数是折扣不是最终价格：Scholar/Governess 传 1 实际付 3；Guest Speaker 传 3 实际付 1。
- J 的 Accessor、Zymnologist 是名称拼写问题，卡名为 Assessor、Zymologist。

## 2. 通用语义

金币写 $；胜利分 VP；收入是年收入轨道。三者不能混用。藤=绿，订单=紫，夏访客=黄，冬访客=蓝。

或为互斥，选二必须两个不同分支。F 允许卡上动作按玩家选择顺序执行，但 Then 条件在前项后判断。支付和强制数量不能部分做却拿完整奖励；至多数量与 may 不等于免付成本。VP 最低 -5，收入 0–5，金币不能负数。

卡只豁免明确限制：免种植建筑不免容量，免容量不免建筑，免费大酒窖仍先要中酒窖。工人总上限 6，含大工人、不含临时灰工；通常培训工人次年可用，Governess 例外。卡上收获/建造/酿酒不拿主板格奖励，建筑触发另行判断。

双访客奖励格必须先完整结算第一张，再选择第二张；可使用第一张刚得到的访客（F）。正在打出的卡不是剩余手牌，不能拿自身付弃牌成本。

## 3. 76 张访客逐卡规范

### 夏季 38 张

|来源|名称|效果|
|---|---|---|
|S01|Surveyor|每块拥有的空田得 2金币；或每块已种植田得 1 VP。售出田不计。|
|S02|Broker|付 9金币 得 3 VP；或失 2 VP 得 6金币。|
|S03|Wine Critic|抽 2 冬卡；或弃 1 瓶价值至少 7 的酒得 4 VP。|
|S04|Blacksmith|建 1 建筑减 2金币；若原价 5金币 或 6金币，再得 1 VP。|
|S05|Contractor|三选二：得 1 VP；原价建 1 建筑；种 1 藤。|
|S06|Tour Guide|得 4金币；或收获 1 田。|
|S07|Novice Guide|得 3金币；或酿至多 2 瓶酒。|
|S08|Uncertified Broker|失 3 VP 得 9金币；或付 6金币 得 2 VP。|
|S09|Planter|种至多 2 藤并得 1金币；或拔除并弃 1 藤得 2 VP。|
|S10|Buyer|付 2金币 放置 1 个价值 1 的红或白葡萄；或弃 1 葡萄得 2金币 和 1 VP。|
|S11|Landscaper|抽 1 藤并种至多 1 藤；或交换田上的 2 藤。交换后田地须合法，新抽藤可种。|
|S12|Architect|建 1 建筑减 3金币；或每座原价 4金币 的已建建筑得 1 VP。|
|S13|Uncertified Architect|失 1 VP 免费建原价 2金币/3金币 的 1 建筑；或失 2 VP 免费建任意 1 建筑。|
|S14|Patron|得 4金币；或抽 1 订单和 1 冬卡。|
|S15|Auctioneer|弃 2 手牌得 4金币；或弃 4 手牌得 3 VP。|
|S16|Entertainer|付 4金币 抽 3 冬卡；或弃 1 酒及 3 访客得 3 VP。|
|S17|Vendor|抽 1 藤、1 订单、1 冬卡；每名对手可以抽 1 夏卡。|
|S18|Handyman|所有玩家可各建 1 建筑减 2金币；每名实际这样做的对手使出牌者得 1 VP。不是旧版拆建筑升级。|
|S19|Horticulturist|种 1 藤，可无所需建筑；或拔除并弃 2 藤得 3 VP。|
|S20|Peddler|弃 2 手牌，抽四种牌各 1；基础 EE 没有第五种结构牌。|
|S21|Banker|得 5金币；每名对手可失 1 VP 从银行得 3金币。|
|S22|Overseer|原价建 1 建筑并种 1 藤；若是价值 4 的藤，得 1 VP。|
|S23|Importer|抽 3 冬卡，除非全部对手合计给你 3 张访客。|
|S24|Sharecropper|种 1 藤，可无所需建筑；或拔除并弃 1 藤得 2 VP。|
|S25|Grower|种 1 藤；然后若自己共种有至少 6 张藤，得 2 VP。不是历史累计种植次数。|
|S26|Negotiator|弃 1 葡萄加 1 收入；或弃 1 酒加 2 收入。不是金币/VP。|
|S27|Cultivator|种 1 藤，可超过该田容量；仍需建筑与田地归属合法。|
|S28|Homesteader|建 1 建筑减 3金币；或种至多 2 藤；可失 1 VP 两项都做。|
|S29|Planner|放置可用工人到未来季节行动；该季开始才执行。现在只占位，不给行动收益。|
|S30|Agriculturist|种 1 藤；然后若该田有至少 3 种不同品种的藤，得 2 VP。不能按实例 ID 数品种。|
|S31|Swindler|每名对手可给你 2金币；每名不付者使你得 1 VP。对手不失 VP。|
|S32|Producer|付 2金币 从其他行动收回至多 2 名自己的工人，本年可再用；不包括正在执行此卡的工人。|
|S33|Organizer|公鸡移至空起床行，取该行奖励，然后结束本季。|
|S34|Sponsor|抽 2 藤；或得 3金币；可失 1 VP 两项都做。|
|S35|Artisan|三选一：得 3金币；建 1 建筑减 1金币；种至多 2 藤。|
|S36|Stonemason|付 8金币 建任意 2 座不同未建建筑，忽略原价；仍先中窖再大窖。|
|S37|Volunteer Crew|所有玩家可各种 1 藤；每名实际种植的对手使你得 2金币。|
|S38|Wedding Party|给至多 3 名不同对手各 2金币；每名收款者使你得 1 VP。|

### 冬季 38 张

|来源|名称|效果|
|---|---|---|
|W01|Merchant|付 3金币 放置价值 1 的红白葡萄各 1；或交付 1 订单额外得 1 VP。F 明确新版。|
|W02|Crusher|得 3金币 并抽 1 夏卡；或抽 1 订单并酿至多 2 瓶。|
|W03|Judge|抽 2 夏卡；或弃 1 瓶价值至少 4 的酒得 3 VP。|
|W04|Oenologist|所有酒陈酿两次；或付 3金币 酒窖升一级。|
|W05|Marketer|抽 2 夏卡并得 1金币；或交付 1 订单额外得 1 VP。|
|W06|Crush Expert|得 3金币 并抽 1 订单；或酿至多 3 瓶。不要照用旧版无限酿酒 FAQ。|
|W07|Uncertified Teacher|失 1 VP 免费培训 1 工人；或每名已有总计 6 工人的对手使你得 1 VP。|
|W08|Teacher|酿至多 2 瓶；或付 2金币 培训 1 工人。|
|W09|Benefactor|抽 1 藤和 1 夏卡；或弃 2 访客得 2 VP。|
|W10|Assessor|每张剩余手牌得 1金币；或弃全部剩余手牌（至少 1）得 2 VP。本卡不计。|
|W11|Queen|右手邻座必须三选一：失 VP、给你任意 2 手牌、付你 3金币。失分数量在现有 OCR 不清楚，J 未实现，须补核，不伪造数值。|
|W12|Harvester|收获至多 2 田，并选择得 2金币 或 1 VP。不是收获与奖励二选一。|
|W13|Professor|付 2金币 培训 1 工人；或自己已有总计 6 工人时得 2 VP。|
|W14|Master Vintner|酒窖升一级，正常价减 2金币；或将 1 酒陈酿一次并交付 1 订单。|
|W15|Uncertified Oenologist|所有酒陈酿两次；或失 1 VP 免费酒窖升一级。|
|W16|Promoter|弃 1 葡萄或 1 酒得 1 VP。|
|W17|Mentor|所有玩家可各酿至多 2 瓶；每名实际酿酒的对手使你抽 1 藤或 1 夏卡。|
|W18|Harvest Expert|三选一：收获 1 田；抽 1 藤；付 1金币 建 Yoke。不是同时收获并抽牌。|
|W19|Innkeeper|打出时取两个不同弃牌堆的顶牌各 1 张。不能取回自己。|
|W20|Jack-of-All-Trades|三选二：收获 1 田；酿至多 2 瓶；交付 1 订单。允许玩家安排顺序。|
|W21|Politician|自己 VP 小于 0 时得 6金币；否则抽 1 藤、1 夏卡、1 订单。|
|W22|Supervisor|酿至多 2 瓶，每瓶新酿起泡酒得 1 VP。|
|W23|Scholar|抽 2 订单；或付 3金币 培训 1 工人；可失 1 VP 两项都做。|
|W24|Reaper|收获至多 3 田；实际收获 3 田得 2 VP。|
|W25|Motivator|每名玩家可收回自己的大工人；每名这样做的对手使你得 1 VP。|
|W26|Bottler|酿至多 3 瓶，每种实际新酿酒类型得 1 VP，不是每瓶得分。|
|W27|Craftsman|三选二：抽 1 订单；正常付费酒窖升一级；得 1 VP。升级不免费。|
|W28|Exporter|三选一：酿至多 2 瓶；交付 1 订单；弃 1 葡萄得 2 VP。|
|W29|Laborer|收获至多 2 田；或酿至多 3 瓶；可失 1 VP 两项都做，顺序由玩家选。|
|W30|Designer|原价建 1 建筑，然后共有至少 6 座已建建筑时得 2 VP。|
|W31|Governess|付 3金币 培训 1 工人，本年可用；或弃 1 酒得 2 VP。|
|W32|Manager|不放工人，执行先前季节 1 行动，不拿格奖励。基础 EE 不引入 Tuscany 秋季主板。|
|W33|Zymologist|酿至多 2 瓶价值至少 4 的酒，即使酒窖未升级；仍需正确配方、葡萄和空槽。|
|W34|Noble|付 1金币 加 1 收入；或减 2 收入得 2 VP。|
|W35|Governor|指定至多 3 名不同对手，各给你 1 夏卡；每名不能给者使你得 1 VP。不能任意拒交保牌。|
|W36|Taster|弃 1 酒得 4金币；若弃前该酒为所有玩家酒窖中价值唯一最高，再得 2 VP，并列不算。|
|W37|Caravan|公开翻四种牌堆顶牌，各 1；取其中 2 张，其余弃掉。|
|W38|Guest Speaker|所有玩家可各付 1金币 培训 1 工人；每名实际培训的对手使你得 1 VP，通常次年可用。|

## 4. 多人响应与延后结算

### Queen

右手邻座按固定座位而非起床序判断；若 Players 的 +1 约定为顺时针，右手为 -1。暂停原行动，把权限交给目标玩家。付款须有 3金币，交牌须有至少 2 手牌，失分须满足 -5 下限。具体交哪些牌由目标选，不是出牌者选。确认转移后才恢复出牌者外层动作，不从目标座位调用 next()。

现有 OCR 的失分图标不清楚，Java 没有 Queen 实现；失分数值与三分支都不可执行时的官方边界尚未签核。不得假造数值、制造欠债或让 VP 低于 -5。私有选牌列表不能广播给其他玩家。

### Guest Speaker 等所有玩家可执行的卡

适用于 Guest Speaker、Handyman、Volunteer Crew、Mentor、Motivator。建议 continuation 保存出牌者、参与者序列、当前响应者、子步骤、返回动作。包括出牌者本人，但奖励只数对手；每人单独接受/拒绝，不能由出牌者替全桌作决定。无钱、工人满或无合法动作，只跳过本人。

先实际完成动作，再给出牌者奖励。Guest Speaker 接受后培训失败不能给 VP；每名玩家仅一次，重发不能重复扣款。默认新工人不能当年用。Mentor 对手酿好后由出牌者选择绿色/黄色奖励，抽完再继续队列。Motivator 收回大工人不会撤销已 pass 状态，不能让结束本季者重新行动。全部响应结束后恢复原出牌者外层动作，不能按最后响应者推进回合。不要套用 J 的 LimitPoints 任意封顶。

### Importer、Governor、Swindler、Banker、Vendor、Wedding Party

Importer 需要对手合计交 3 张访客的集体选择，建议私有提交并托管，凑齐才统一转移；未凑齐撤销托管并抽 3 冬卡。不能先抢 1–2 张再额外抽 3。协商/承诺顺序属于数字化协议，不冒充卡面唯一顺序。

Governor 先选至多 3 目标，再让各目标私有选夏卡；确实没有才得 VP，不是目标失 VP。Swindler 独立付款或拒绝；Banker 独立选择失 VP 从银行拿钱；Vendor 保留拒绝抽牌权。Wedding Party 由出牌者选不同收款人并一次检查总支付能力。

### Organizer

1. 先选空起床行；旧行释放、新行占用，WakeSlots 与 Player.Wake 同步。
2. 取新行奖励。第 5 行先选夏/冬，第 7 行获得本年临时工，第 1 行无资源奖励。
3. 所有奖励的待决抽牌/选牌完成后才 pass，不能直接 next() 吃掉后续选择。新排名作用于后续顺序，不给自己本季额外一回合。
4. 双访客格内打出时保留外层剩余步骤与待 pass，不要覆盖上下文。第二访客与强制 pass 的最终优先级仍须官方裁定补核，不能凭 Java 自动过季逻辑断言。
5. Manager 冬季执行夏季访客再打 Organizer 属于嵌套链，不可直接把全局 Phase 改成 summer。一个人 pass 不代表全桌换季。

### Planner

现在扣可用工人并占未来格；未来季开始、正常轮流摆工人前执行预约。执行时才检查当时资源和位置奖励，不提前给收益，也不能执行时再扣一次工人。多个预约和其待决选择完成才进入正常起床序。基础 EE 没有秋季摆工人行动；不可复制 Tuscany 季节编号开放不存在的格。多个预约具体顺序及无法执行的边界须补核验收。

## 5. 隐藏信息时序

### Cottage：基础 EE 的已明确 FAQ 裁定

F 明确可抽任意两色组合，但必须同时抽，不能看第一张后决定第二张。正确事务为：选择夏夏/夏冬/冬冬 → 锁定两色 → 抽取 → 仅本人看到结果。也可分两步提交颜色，但第二色确认前第一张不得进入可见 Hand、快照、SSE、日志或撤销界面。

已有测试 TestAuditCottageCannotRevealBeforeSecondColorCommitted 允许原子双颜色方案拒绝旧单颜色命令。不要将修复只放前端，API 本身必须不泄漏。

### Oracle：Tuscany 特殊工人，不是基础 EE 访客

不要把 Oracle 当 EE 第 77 张访客，也不要给基础秋季抽牌附加 Oracle。这里是 J 状态机兼容审计，并非本轮已核对 Tuscany 官方规则书。

J 2499–2564 的 drawCards、6390 起 drawForOracle、6526 起 SelectCardColor：同色抽牌多抽 1 后从本次候选选留；多牌堆先选额外抽牌颜色，再抽候选，再选留/弃。不能看原有抽牌后才选额外颜色。本次候选只对拥有者可见。

保留后续 continuation：Landscaper 抽藤 → Oracle 选留 → 种藤；Scholar 双做 → 抽订单/Oracle → 培训。不能用新 pending 覆盖旧待决。不能让 currentWorker 残留导致 Vendor 对手奖励或别人响应也额外抽牌；J revision 169 专门修过该交互。Cottage 为免费秋季抽牌，不沿用当前工人；Caravan 翻牌本来公开，不混成普通暗抽。

### Innkeeper 与弃牌

As you play 要在本卡盖住冬季弃牌顶前确定两个不同弃牌堆的顶牌，不能拿回自身，也不能提前入弃牌遮住原顶牌。J 5809–5816 对应此时序。建议使用独立 resolving 区。Assessor、Auctioneer、Peddler 的剩余手牌不包含正在出的本卡。

## 6. 基本规则错误及测试状态

### 已由本轮读取基线确认

- **起始玩家方向错**：engine.go finishYear 使用 (SpringLeader + 1) % n，春季选起床同样 +1。F 明确春季顺时针、年末起始标记逆时针。三人 A/B/C，A 后下一年应 C 开始。已有 TestAuditSpringLeaderMustRotateOppositeWakeSelection。
- **Cottage 泄漏**：engine.go next 连续 enqueue 两个单张 fall 选择；必须先锁定两色后展示牌。已有上述 Cottage 审计测试。
- **源代码完整性风险**：EE_SOURCES.md 旧状态明确 76 访客未实现，J 本身 Queen/Importer 缺失。不能将卡记录存在视为效果已实现。另代理正在实现，不据旧记录宣称新后端全部仍缺失。

### 必须回归，但不冒充全部已证实有 bug

每田每年只收获一次，藤不因收获消失；双收获是两田而非同田两次。工人仅年末回收；大工人计 6 上限、灰工不计。先中窖后大窖，折扣不倒赚。桃红一红一白最低 4，起泡两红一白最低 7，同色葡萄不能合成一瓶红/白；酒窖限制与空槽降级正确，Zymologist 局部豁免。

订单每图标匹配一瓶同型且至少指定值的酒，不能合瓶；收入不即时拿钱、不年末清零。行动格数量随人数，大工人溢出不拿被占奖励；访客子动作不拿主板奖励。20 VP 不是立即中止卡片/多人队列，年末结束及弃至 7 张不能跳过。并列比较 VP、现金、酒总值、葡萄总值，完全相同并列。错误请求原子回滚，重复 choice 不能重复收益，私有选牌不广播。

### 测试交接

已有 audit_base_rules_test.go 带 ee_rule_audit build tag，命令：go test -tags ee_rule_audit -run TestAudit -v .



现有 /tmp/ee-audit/base-test.log 是构建失败：缺 VisitorStep、ChoiceField、visitorDefs、resolveVisitorChoice、startVisitor，发生于后端并发编辑。不能说两条测试已运行失败/通过。本轮不修改 Go。

每卡应测全部分支、费用不足、-5/收入/工人上限、空目标、去重、回滚；多人卡加拒绝、无能力、已 pass、重发、恢复原行动。优先补核 Queen 图标和无可行分支、Organizer 双访客 pass、Importer 协商提交、Planner 延后边界。表有 76 行不等于 76 卡验收通过。

## 后续目视核实与实现更新
Queen原图已由主代理通过原生视觉查看：https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/wintervisitor-cropped/WinterVisitors_EssEd_2nd-Page-000011.jpg 。失分符号明确为紫色1 VP，现金为3，手牌为2；上文历史OCR未决仅保留审计轨迹，已解决。
三张特殊卡及Cottage/起始玩家方向已实现并通过专项测试。上文旧基线构建失败不是最终状态；以README和验收JSON为准。
