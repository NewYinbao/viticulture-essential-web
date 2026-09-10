# 冬季38访客 Go / Boardspace Java 差分审计

## 结论、范围与证据边界

阅读 visitor_winter.go、visitor_sequence.go、visitor_interaction.go、visitor_special.go、visitor_feasibility.go，追踪 visitor.go / engine.go / ee.go 共用逻辑；对照 /tmp/ViticultureBoard.java、/tmp/PlayerBoard.java、/tmp/ViticultureChip.java 及 /tmp/ee-audit/ 卡面OCR与FAQ。下文 Java 行号均指 /tmp/ViticultureBoard.java。

发现2项明确实现错误 W1/W2，1项 Manager 嵌套官方裁定/Java 差异 W3，1项 Go/Java 共同偏离所提供FAQ的最低执行量问题 W4。W2跨全体陈酿、单瓶陈酿和预检三处。Java未实际编译运行，期望由列出的代码推导；Go通过内存 Room.apply 事务入口实际复现。仅新增本报告与 audit_winter_diff_test.go，不修改生产、已有测试、存档。

## W1 Harvest Expert #18 把收获并抽藤错拆成互斥项

**P1，确定，正常使用必然少拿收益。**

- Go visitor_winter.go:26 声明 harvest/draw/build 三选一；:205-214 中 draw 提前返回，harvest只收获。
- Java :5789-5808；Choice_A在 :5794-5795 同时安排 Harvest1Optional 和抽1张green，Choice_B付1建yoke。
- 卡OCR /tmp/ee-audit/WinterVisitors_EssEd_2nd-Page-000018.txt:12-14 明确可读 Harvest 1 field and / draw；图标数字OCR不可靠，不以OCR数字为唯一依据。
- 可复现初态：可收获田0，藤牌堆非空；打winter-18后提交 option=harvest, field=0。
- 期望：田收获且手牌+1藤。实际：田收获，手牌增量0。选择draw又无法接续收获。建yoke的付1金币正确。
- 实测 TestAuditWinterDiffHarvestExpert 失败：hand delta=0。

## W2 Zymologist越级酒被错误冻结，影响#04/#14/#15及年末

**P1，确定Java差异：合法特殊资源不能继续陈酿，阻断交单与得分。**

- Go engine.go:125-132把酒窖转换为全局上限；:152-168中 w.Value >= cellar(p) 禁止所有已高于酒窖上限的酒移动。
- #04 visitor_winter.go:85-88、#15 :195-198两次调用该函数；#14 :173-188复制同一禁止逻辑；visitor_feasibility.go:283-305，特别:287-288又在预检重复。engine.go:65-66年末也受影响。
- Java canAge :3736-3743仅禁止跨入缺失酒窖边界：3→4、6→7，不禁止已经越过边界的酒移动。全体陈酿:3746-3777；单瓶:3779-3812。#04 :5564-5565；#15 :5767-5768；#14 :5754。PlayerBoard.java:1232-1234,1356定义双酒窖/中酒窖状态。
- 可达初态：无中/大酒窖，先用#33把红葡萄4酿成红酒4，5、6槽空；该酿酒初态已用控制测试验证合法。
- 输入A：#04或#15选择age。期望4→5→6，实际仍4。没有大酒窖仍不得6→7；已有越级7/8可继续向9移动，不能简单把所有后续边界都取消。
- 输入B：无酒窖，红酒ID=zym-red4品质4，红5订单；保留至少2金币使#14打出时升级分支可行，再选 option=fill,wineIds=[zym-red4]。期望先陈酿至5并进入交单，实际报“酒不能陈酿”。若金币不足，预检也会错误拒绝打牌。
- 实测 TestAuditWinterDiffAgeBeyondCellar（#04/#15两个子用例）actual=4；TestAuditWinterDiffMasterVintnerAgeBeyondCellar报上述错误。

## W3 Manager→Producer错误排除外层整个行动，不能回收旧工人

**P2，明确Java差异；Java标注official ruling，与卡面other actions字面有张力，应确认采用官方裁定。**

- Go Manager visitor_sequence.go:237-253正确保留外层Context.Space；Producer :186-189以行动ID相等整类拒绝，不是仅排除本次触发工人。
- Java :10860-10864明确注释官方裁定：可回收任何工人，除了用于打此牌的工人，并点名Manager嵌套时触发工人在blue card行动。:10865-10885，特别:10877-10878以lastDroppedWorker及index排除实际触发工人（revision>=153）。Manager入口:6072-6075，执行:7132-7136。
- 初态：3人局，同玩家较早在winter_visitor slot0放过工人；本次slot1打winter-32，选summer_visitor并打summer-32 Producer。
- 输入：option=retrieve,targetIds=[winter_visitor],fields=[0]，有2金币。
- 期望旧slot0可取回，刚放slot1不可。实际“只能取其他行动工人”，事务回滚此次选择。
- 实测 TestAuditWinterDiffManagerProducerOtherWorker失败。
- 控制 TestAuditWinterDiffManagerBonusPreservedControl通过：Manager嵌套普通夏访客后仍有外层冬访客第二张，Context.Space仍winter_visitor。没有把已修复bonus丢失再报一遍。

## W4 up to N空执行与提供FAQ冲突，不是Java独有差分

**P2，按提供FAQ应修；Go与Java均宽松，不可用Java直接反证FAQ。**

- /tmp/ee-audit/faq.html:531：If a card says “up to” ... you must fulfill the card to some extent ... actually plant at least 1 vine。同段要求最低效果可完成。:533规定默认可任意顺序执行。
- Go visitor.go:244-259对空Fields/Recipes无条件成功；visitor_feasibility.go:96-107,153-157明确把up to当作0可行。#12 visitor_winter.go:149-157允许0收获领取1分或2金币；#22/#24/#26/#33及#29相关最低量也无守卫。#17互动回复visitor_interaction.go:101-105反而明确要求至少酿1酒才奖励。
- Java #12 :5704-5717先给奖励并进入Harvest2Optional；#24 :5944-5948；#33 :6076-6078；:10181甚至明注zymologist - might make no wines。因此本项是参考实现宽松策略与FAQ冲突，不声称Java会拒绝。
- 输入：3田全部无藤，打#12，option=vp,fields=[]。FAQ期望拒绝打牌或拒绝空收获，不得积分；实际整牌成功+1VP。
- 实测 TestAuditWinterDiffHarvesterMinimumFAQ失败，actual VP delta=1。
- 修复需区分“全员may不参与”、做了才给奖励、OR分支和AND必须一起完成，不能简单删除所有skip。

## 其他边界与不计入确定缺陷的差异

1. #33红酒4槽已有酒：Go visitor_winter.go:238-245检查最终成酒>=4，拒绝下挤成3；Java :10743-10754仅要求原始葡萄>=4，然后:8895-8909调用向下找空槽放酒逻辑，可能成3。卡文说酒value4 or greater，Go有文本依据，不以Java放宽直接判Go错。错误返回前临时建筑/葡萄变动由engine.go:189-206事务回滚，不是可提交免费大酒窖漏洞。
2. #14无可陈酿酒却可交单：Java :5754回退FillWineOptional；Go visitor_winter.go:173-194要求陈酿，visitor_feasibility.go:373-375要求余下交单可行。卡文及FAQ最低效果支持Go更严格，不报缺失宽松回退。Java :10032全局Choice_0也不表示本体必须允许放弃。
3. Queen #11：ViticultureChip.java:437-448包含本体Queen；Java :10190-10191注queen, too mean无分支。Go visitor_special.go:19,21-43,100-128已实现右邻座三选一，:47-52检查最低可行性。不能以Java有意删除Queen判断本体不需Queen。
4. #33无酒窖酿red4/blush4/sparkling7，三个控制测试全部通过，临时大酒窖没有留下。engine.go:492-522检查混酿比例与4/7门槛；Java :10760-10832对应。
5. #22/#26仅本次酒得分：Go visitor_winter.go:228,238-248扫描本次追加酒；Java :7423-7444从pendingMoves聚合本次酒种，#22 :7446起按起泡酒计数。未发现旧酒重复计分。
6. #36唯一最高包括自己的其他酒：Go visitor_winter.go:345-368；Java :7088-7090对所有玩家剩余酒比较。合法酒ID唯一，不把非法重复ID存档作为复现。
7. 不将Tuscany workshop折扣、special workers、5类抽牌、LimitPoints/UnlimitedWorkers等计入本体遗漏。

## 38张覆盖索引

“无独立新差分”仅表示此次阅读未发现，不宣称穷尽所有组合。Java统一对应resolveBlue :5500-6135编号case，#11例外如上。Go文件简称winter=visitor_winter.go，sequence=visitor_sequence.go，interaction=visitor_interaction.go，special=visitor_special.go。

|编号|Go入口|结论|
|---|---|---|
|01–03|winter:55-84|葡萄/交单/Crusher/Judge无独立新差分|
|04|winter:85-94|W2|
|05–10|winter:95-148|收益、训练6工人、弃访客、剩余手牌计数无独立新差分|
|11|special:100-128|Queen已实现，Java有意缺失|
|12|winter:149-157|W4|
|13|winter:158-165|6工人分支无独立新差分|
|14|winter:166-194|W2，预检重复；Java宽松回退不计遗漏|
|15|winter:195-204|W2|
|16–17|interaction:160-172;101-111|弃资源/全员酿酒奖励无独立新差分|
|18|winter:205-214|W1|
|19|sequence:214-236|不同弃堆/暂存自身排除无独立新差分|
|20|sequence:68-82;22-60|三选二顺序/续步有实现；最低量受W4影响|
|21–22|winter:215-249|负分阈值正确；#22最低量W4|
|23|winter:250-275|培训3金币、失1分双做及抽前检查有实现|
|24|winter:276-282|三田2分逻辑；空收获W4|
|25|interaction:113-132;173-174|回收大工人及对手奖励/取消预约有实现|
|26|winter:223-249|本次酒种去重正确；最低量W4|
|27|winter:283-301|二选项/顺序/原价升级/抽前检查有实现|
|28|winter:302-313|弃葡萄2分/交单；酿酒最低量W4|
|29|sequence:68-82|付1分双做及顺序有实现；最低量W4|
|30–31|winter:314-328|6建筑2分、立即可用工人无独立新差分|
|32|sequence:237-253|W3；嵌套冬访客bonus保留控制通过|
|33|winter:223-249|例外酿酒通过，后续陈酿W2，最低量W4|
|34|winter:329-343|收入上限5/失2收入得2分无独立新差分|
|35|interaction:175-197;133-134|至多3不同对手/必须给夏访客有实现|
|36|winter:344-368|最高酒包含自己其他酒，无独立新差分|
|37|sequence:254-287|四类公开取2及弃其余，不要求Tuscany第五类|
|38|interaction:135-141;198-199|全员付1培训及对手每人1分无独立新差分|

## 验证与文件

新增audit_winter_diff_test.go，首行 //go:build audit_diff。
运行命令：/opt/viticulture-toolchain/go/bin/go test -tags audit_diff -run '^TestAuditWinterDiff' -v .

实际：HarvestExpert、AgeBeyondCellar（两个子用例）、MasterVintnerAgeBeyondCellar、HarvesterMinimumFAQ、ManagerProducerOtherWorker按期望断言失败；ZymologistControls（三个子用例）、ManagerBonusPreservedControl通过。这些是故意失败的缺陷复现，不是测试全绿。

不带tag运行 go test -run '^TestAuditWinterDiff' . 返回ok [no tests to run]，隔离有效。未运行整个原套件，避免无关HTTP/存档副作用。测试仅内存Room.apply，不访问服务和存档。

工具环境异常：上下文声明Windows，但terminal实为WSL；read_file/search_files/write_file无法访问该WSL路径或默认工作目录不存在，失败后使用terminal+Python仅写用户允许的两个新文件。shell末尾有Hermes cache路径警告。报告首次写入受到包装shell反引号展开影响，已重写为不含反引号的版本，并再次检查文件内容。项目无.git，不能提供git差分验证，不假称有干净git baseline。
