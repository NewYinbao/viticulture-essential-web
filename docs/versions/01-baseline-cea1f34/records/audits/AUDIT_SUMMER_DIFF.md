> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# 夏季38访客只读差分审计

## 范围、证据与复验

逐项审阅 visitor_summer.go、visitor_sequence.go、visitor_interaction.go、visitor_special.go、visitor_feasibility.go，追入 visitor.go、engine.go、model.go、ee.go。对照 /tmp/ViticultureBoard.java 的 resolveYellow、分支生成和最终结算，以及 /tmp/PlayerBoard.java、/tmp/ViticultureChip.java。卡OCR为 /tmp/ee-audit/SummerVisitors_EssEd_2nd-Page-000NNN.txt；FAQ为 /tmp/ee-audit/faq.txt（整个网页压成第1行）。Java未编译运行，结论为静态推导；Go均经 Room.apply 事务入口真实复现。

仅新增本报告及 audit_summer_diff_test.go。测试由 //go:build audit_diff 隔离，断言正确规则，故当前应失败。没有修改生产/已有测试/存档。

工具链 /opt/viticulture-toolchain/go/bin/go：
- go test -count=1 .：PASS（0.555s）。
- go test -tags audit_diff -run TestAuditSummerDiff -count=1 -v .：下列规则断言失败。
- Organizer首次夹具误选已占用格，已改新测试为槽位1并重跑，得到正确规则失败 Workers 2→3。

## 发现1：零执行仍可结算 up-to 分支，Planter无藤领取1金币

**严重度：中；EE错误。** 最小局面：手中仅summer-09，三块空田，没有任何藤手牌或已种藤。正常打出Planter，再提交 option=plant，不选CardIDs/Fields。

期望：无可完成的种植/拔藤分支，应拒绝打出或至少拒绝空种植。实际：两步均成功，Coins 50→51，田和藤手牌始终为空。复现 TestAuditSummerDiffPlanterEmptyReward。

- Go visitor_summer.go:110–120：112调用种植，115无条件加金币；visitor.go:201–210,231仅检查数量上限和数组等长，空数组直接成功；visitor_feasibility.go:216–278缺summer-09起始能力检查。
- Java ViticultureBoard.java:10371–10379：仅canPlant开放种植，仅hasPlantedVine开放拔藤；4566–4577给1金币后进入Plant2VinesOptional。在上述完全无藤局面不会开放收益分支。10277的通用Choice_0不是免费获1金币的依据。
- EE夏09卡面：Plant up to 2 ... and gain ...。FAQ第1行明确：If a card says “up to” (i.e., Plant up to 2 vines), you must fulfill the card to some extent. In that example, you must actually plant at least 1 vine.

同根因经事务入口验证：

|卡|空提交|Go行号|Java行号|实际|
|---|---|---|---|---|
|07 Novice Guide|make无Recipes|summer:92–97；visitor:255–259|Board4534–4544；10355–10359|接受空酿酒|
|28 Homesteader|plant无藤|summer:262–287；visitor:201–210|Board4838–4853；10446–10455|接受空种植|
|35 Artisan|plant无藤|summer:301–308；visitor:201–210|Board4936–4949|接受空种植|
|32 Producer|retrieve无Targets/Fields|sequence:177–185|Board4879–4882；10337–10341|付2取回0工人|
|38 Wedding Party|pay无Targets|summer:322–338|Board4964–4965|付给0人|

注意：这些扩展项依据EE FAQ最低执行量，**不声称Java所有Optional状态也严格禁止0**。Java自身允许Optional和通用不执行出口，Producer初始检查也只有现金。Landscaper抽1后种0与Planter直接种0应分开：前者至少实际执行了抽牌部分。以上是一个根因，不夸大成六个独立漏洞。

## 发现2：Landscaper交换已种藤错误地再次要求种植建筑

**严重度：中；EE错误，与Java直接不同。** 局面：无irrigation，田0已有Merlot红3/Irrigation=true，田1有Sangiovese红1，容量5和6。可用此前的Horticulturist或Sharecropper合法种出这个局面。

操作：夏11 swap，Fields=[0,1]、CardIDs=[merlot,sangio]。期望：交换成功，容量均合法，这不是重新Plant。实际：返回“缺种植建筑”，事务回滚。复现 TestAuditSummerDiffLandscaperNoStructures。

- Go visitor_sequence.go:109–121，尤其115–117遍历交换后两田全部藤并检查设施，连未移动的藤也会导致拒绝。engine.go:190–205确保回滚，所以不报半交换污染。
- Java ViticultureBoard.java:4600–4611、10387–10389；canSwap在10558–10570只检查交换后容量；10576–10605枚举合法交换，不检查建筑。
- EE夏11卡面：switch 2 ... on your fields，无重新满足种植设施要求。FAQ第1行：You only need a Trellis to plant vines. Once a vine is planted, it stays on your field permanently (unless you decide to uproot it). 该段有旧版Handyman背景，只用其设施不持续约束已种藤原则；当前卡面和Java canSwap为直接证据。

## 发现3：Organizer可凭空再造第7行灰工人

**严重度：高，破坏工人经济不变量；EE错误。** 局面：A春季选第7行领取唯一临时灰工人；夏季用Organizer移到第4行。第7行虽然空了，灰工人仍归A。B随后经回收/重洗再取得该Organizer，移到第7行。

期望：B可进入空行，但不能再领已经被A拿走的唯一灰工人。实际：B普通可用Workers 2→3，A仍持有之前工人。复现 TestAuditSummerDiffOrganizerCannotMintGrayWorker；B用大工人打牌，不混淆普通工人放置消耗。测试直接布置再次取得此牌的局面并移除弃牌堆中的该牌，不冒充完整自然抽牌路径。

- Go visitor_special.go:85–98只查公鸡所在行；129–164尤其162–163无条件Workers++；engine.go:251–264春季第7行同样直接Workers++；model.go:46–70,91–115无灰工人唯一领取状态。
- Java ViticultureBoard.java:4027–4041，4032先检查current.topChip()==ViticultureChip.GrayMeeple；4033注释特别说明Organizer再用此行时工人未必还在；4034移除真实棋子、4035登记owner。
- 版本边界：Java起床表包含Tuscany季节奖励，不能整表照搬；此处只引用summer唯一灰工人存在性。EE同样仅一个临时灰工人，不把Tuscany冬季第7行起始玩家标记当作EE奖励。
- 同一玩家跨季回收Organizer再进入该行同样受此根因影响；夏季立即pass不能证明不可达。

## 38张逐卡覆盖索引

Go缩写：S=visitor_summer.go，Q=visitor_sequence.go，I=visitor_interaction.go，P=visitor_special.go；Java均为 ViticultureBoard.java 的行号。下表“未发现”仅表示本轮有限静态/局部动态检查无新的可证明错误，不表示全部组合被穷尽。

| # / 英文名 | Go | Java | 本轮判断 |
|---|---|---|---|
|01 Surveyor|S46–56|4436–4448|未发现；PlayerBoard1076–1089核对已售空田排除|
|02 Broker|S57–68|4450–4464|9金币/3VP、2VP/6金币及负分底线一致|
|03 Wine Critic|S69–77|4467–4478；10419–10422|现版本品质>=7一致；不要采用旧revision的8|
|04 Blacksmith|S78–84|4482–4484；6595–6601|EE建筑最大价6，>=5等价5/6|
|05 Contractor|Q10,22–43,68–82|4487–4517；10430–10443|Go强制选两项符合EE；Java允许单项是宽松差异，不报Go缺陷|
|06 Tour Guide|S85–91|4519–4530|未发现|
|07 Novice Guide|S92–97|4534–4546|发现1：空make|
|08 Uncertified Broker|S98–109|4549–4562|未发现|
|09 Planter|S110–121|4566–4579；10371–10379|发现1：无藤仍获金币|
|10 Buyer|S122–137|4583–4597|分支数值一致；已占品质1时葡萄放置细节未独立裁定，不列确证缺陷|
|11 Landscaper|Q83–122|4600–4614；10558–10605|发现2；抽牌后种0与纯空种植分开看|
|12 Architect|S138–146|4617–4632|原价4建筑计数对应PlayerBoard1186–1203；扩展橙建筑不属EE|
|13 Uncertified Architect|S147–158|4635–4649|未发现|
|14 Patron|S159–165|4652–4675|未发现；Oracle是Tuscany扩展分支|
|15 Auctioneer|S166–178|4678–4689|弃2/4手牌成本一致|
|16 Entertainer|S179–193|4693–4706|3访客+1酒一致，已打卡不计弃牌成本|
|17 Vendor|I7,72–73,146–150|4709–4727|Go允许对手拒绝抽牌符合EE may；Java强制抽不是Go错误|
|18 Handyman|I74–80,151–152|4730–4742|折扣2与每对手1VP一致；无扩展限分选项|
|19 Horticulturist|S194–209|4746–4758；6724–6725|无视设施/拔2弃2得3VP一致|
|20 Peddler|S210–216|4761–4763|未发现；旧卡可重洗影响组合可达性|
|21 Banker|I81–85,153–155|4766–4788|5金币、对手失1VP得3金币一致|
|22 Overseer|Q12,22–43,68–82|4791–4793；6765–6770|Go双向顺序及种价值4得1VP符合EE FAQ|
|23 Importer|Q123–176|4796–4798；6210–6211|Java明确移除且只抽3，不能当缺失Go的参考标准|
|24 Sharecropper|S194–209|4800–4812|未发现|
|25 Grower|S217–231|4814–4815；6677–6685|种后累计>=6一致，额外Windmill分由基础plant处理|
|26 Negotiator|S241–256|4818–4832|收入1/2及上限5一致|
|27 Cultivator|S257–261|4834–4836|只放宽容量，不放宽设施，一致|
|28 Homesteader|S262–288|4838–4855|发现1；Go支持反向顺序符合FAQ|
|29 Planner|P53–76,167–196,216–343|4857–4863；1818–1868|未新增确证错误；见下方版本边界|
|30 Agriculturist|S217–240|4865–4867；6772–6784|Go按Name、Java按vineProfile；EE固定藤品种与数值一一对应，不用伪造Name卡报假阳性|
|31 Swindler|I65–68,86–90,156–157|4870–4876|对手拒付得1VP一致|
|32 Producer|Q177–213|4879–4882|发现1扩展；取消Planner位置的连带清理已实现|
|33 Organizer|P78–98,129–166,203–215|4884–4911；4027–4041|发现3；不重复README的外层双访客优先级约定|
|34 Sponsor|S289–300|4913–4934|未发现|
|35 Artisan|S301–309|4936–4952|发现1扩展|
|36 Stonemason|S310–321|4955–4961|付8建2，Go按选择顺序建，可先中后大窖；无新增确证错误|
|37 Volunteer Crew|I91–100,158–159|4967–4970；6787–6795|必须实种1藤才触发对手给2金币，一致|
|38 Wedding Party|S322–338|4964–4965|发现1扩展；不同对手、每人2金币、最多3人已校验|

## 避免版本误报与本轮保留项

1. Java实际排除Importer：Board1254、4796–4798、6210–6211；Queen为冬11，Board1285、4188、6140，Chip472/475对应常量。Queen不计入夏38覆盖数。Go实现了它们并不意味着应向Java残缺行为看齐。
2. Planner未来季：Java11091–11095遍历未来季，Go P54、P180只冬季；EE只有夏/冬工人季，因此Tuscany的秋季工人行动不是EE遗漏。Java ContinuousPlay（1818–1838、2417–2426）每玩家进入季节的触发与EE全桌同时转季不同。
3. 多Planner预约顺序：Java非ContinuousPlay在1833起也是遍历预约数组，不足以把README的预约先后直接判成bug。本轮不复述已有约定充数。
4. Java含Oracle、特殊工人、橙建筑、影响力/限分等扩展代码。仅对EE公共子集比较数值、数量、成本、已种藤和唯一组件不变量。
5. Buyer品质1被占、Homesteader在-5分时先种植获Windmill分再支付1分等边界值得后续官方裁定/复现；本报告不将尚未足够验证的候选混入确证列表。
6. 文件工具与terminal不处于同一挂载视图：read_file对所给WSL路径返回不存在，search_files缺rg/find；因此使用terminal内Python只读编号输出，且仅以Python创建授权的两个新增文件。没有修改另一profile或绕开存档隔离。
