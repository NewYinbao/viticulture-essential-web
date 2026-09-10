# EE 后端访客 API 与验收状态

## 范围与真实缺口

本轮仅修改后端 Go 和本 API 文档；没有修改 web/、ee_cards.json，也没有启动或替换现有服务。
规则依据为 EE_RULE_AUDIT.md、出版商 FAQ 和对应卡图 OCR。**76/76 张访客已注册效果并通过每卡初始选项执行测试；不是完整 EE 验收。**

本次补齐 Planner summer-29、Organizer summer-33、Queen winter-11；Queen 官方卡图的紫色圆内数字 **1 VP** 由主代理亲自目视确认，不再沿用旧审计中“数量待核”的阻塞状态。卡图来源：
https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/wintervisitor-cropped/WinterVisitors_EssEd_2nd-Page-000011.jpg

其他待验收风险：每卡测试覆盖全部已公开的初始选项及默认后续执行，但**并未穷举全部数值、费用/资源边界、多人所有分支组合**。Manager 的任意嵌套组合未穷举。多步骤已确认步骤逐请求持久化，后续强制步骤无法完成时没有取消/整体撤回协议；不能把单请求原子性当成整张卡跨请求可撤回。这些仍需规则/交互验收，不应作为已完整原版发布。

## HTTP 命令与事务

POST ，，JSON 必带当前 。
- 打牌：、、，可带  / 。
- 响应：、、 必须为当前  中的一项。
- 当前  独占响应权。其他玩家只见 ，不见该选择 ID/选项。
- 旧 revision 返回 409；无权限、旧 choiceId、无效资源/选项返回 400；保存失败返回 500，权威状态不变。
-  本身也对失败回滚，HTTP 另在候选副本上执行，保存成功才提交并 SSE 通知。不要直接调用低层 effect/helper 并期望其自身事务化。
- 队列、Context（出牌人、返回回合、剩余第二访客、托管牌）、暂存中的 Innkeeper、公开 Caravan 均保存于 ee-state-v1.json。Context/托管牌不放入玩家视图。

例：。

## 参数约定

|参数|含义|
|---|---|
|cardId|打出的访客，或交付的订单；第二访客 play 的 ID|
|cardIds|弃/给/种/取的卡 ID 列表，种植时与 fields 对齐|
|fields|田地零基索引列表；Producer 特例见下|
|field|单田字段：Tour Guide / Harvest Expert 使用|
|building / buildings|建筑键；Stonemason 必须两个不同合法未建建筑，顺序重要|
|recipes|葡萄索引二维列表；所有配方基于同一次提交前的葡萄列表，不随消耗重编号|
|grapes|弃葡萄索引列表（单葡萄效果须恰好1）|
|wineIds|酒 ID 列表；订单按需求配对；单酒效果须恰好1|
|targetIds|对手 ID；Wedding Party/Governor 至多3且去重；Producer 为行动空间 ID|
|color|Buyer: red/white；起床第5行: summer/winter|
|colors|Innkeeper: 两个不同弃牌堆，vine/summer/order/winter|
|space|Manager 选择夏季行动；不消耗工人、不授予格奖励|

建筑键：yoke, trellis, irrigation, medium_cellar, large_cellar, cottage, windmill, tasting_room。
 用于种植，不能混同订单 。至多种植/酿酒/收获允许空数组，强制1藤/1田仍须完成。

## 后续阶段

不要假定一次 choose 就结束回合；始终重新读取 pendingChoice。
- 双访客：第一张及其多人响应全部结束后才出现 ， 或 ，可用新抽到的访客。
- 选二顺序： 等表示本次先执行 build，再出现 plant 待决；下一步使用新的 choiceId 和对应参数。Homesteader / Scholar 的  分别是 build_plant / draw_train 的兼容别名。
- Landscaper draw 后 stage=plant，可提交空 cardIds/fields 不种。
- Master Vintner fill 首步 wineIds 为需陈酿的1酒；后续 stage=fill 再提交订单 cardId 和用于交付的 wineIds。
- Importer：出牌人 resolve 后，对手按固定座位次序 give（1至剩余数量的访客）或 skip；提交移至私有托管区，不先交给出牌人。最终出牌人 stage=settle/resolve，恰好3张才全部转移，否则全部退还原玩家并抽3冬卡。数字化协议不是卡面唯一协商方式；满3后剩余对手须 skip。每次提交可持久化恢复。
- Mentor：各玩家 make/skip；只有实际酿酒的对手产生奖励。奖励立即插队交还出牌人选择 **vine/summer**，抽完再问下一名玩家。
- 所有玩家可参与卡：本人也有 reply；奖励只数实际参与的对手。无资源者可 skip。已 pass 不被重新激活。Governor 持夏卡目标没有 skip。
- Producer：targetIds 为至多2个其他行动空间，fields 为对应 seat.slot；空数组仍支付2金币。当前行动空间不允许取回。
- Innkeeper：本卡先离手暂存，不盖住旧冬弃牌顶；选两个不同非空弃牌堆并取顶后，本卡才入冬弃牌。绝不能选回自身。暂存同样持久化。
- Caravan：先 reveal，公开 （四类各1，牌堆空时取实际可得牌）；stage=take 选择两张 cardIds，其余弃牌，清空公开区。
- Manager：space 为夏季行动，提交该行动参数；若 summer_visitor，cardId 必须是自己的夏访客，完整结算该嵌套卡。Organizer 保留冬季外层；Planner 因本年没有未来季节而拒绝，不能预约下一年。

## Cottage 与年末方向

Cottage 秋季为**单个 count=2**选择：summer_summer / summer_winter / winter_winter。确认前不抽任何牌；锁定组合后同一事务抽两张。普通秋季仍 count=1，summer/winter。旧前端逐张选择必须升级才能操作 Cottage；本轮按限制未修改前端。
春季选起床保持座位 +1，年末起始标记改为 -1；三人 A/B/C，A 的下一年为 C。

## 已注册卡和初始 option

下表是运行实现清单，不是独立规则证明；数值权威见审计文件。

|ID|效果摘要|初始 option|
|---|---|---|
|summer-01|每块自有空田得2金币或每块已种田得1分|coins, vp|
|summer-02|付9金币得3分或失2分得6金币|buy, sell|
|summer-03|抽2冬访客或弃品质至少7的酒得4分|draw, discard|
|summer-04|建造减2金币，原价5/6建筑额外1分|build|
|summer-05|三选二，顺序自选：得1分、原价建造、种1藤|vp_build, build_vp, vp_plant, plant_vp, build_plant, plant_build|
|summer-06|得4金币或收获1田|coins, harvest|
|summer-07|得3金币或酿至多2酒|coins, make|
|summer-08|失3分得9金币或付6金币得2分|sell, buy|
|summer-09|种至多2藤得1金币或拔除弃置1藤得2分|plant, uproot|
|summer-10|付2金币得品质1葡萄或弃1葡萄得2金币及1分|buy, discard|
|summer-11|抽1藤后可种1藤，或交换不同田地的两藤|draw, swap|
|summer-12|建造减3金币或每座原价4建筑得1分|build, vp|
|summer-13|失1分免费建原价2/3建筑或失2分免费建任意建筑|small, any|
|summer-14|得4金币或抽1订单及1冬访客|coins, draw|
|summer-15|弃2手牌得4金币或弃4手牌得3分|coins, vp|
|summer-16|付4金币抽3冬访客或弃1酒及3访客得3分|draw, discard|
|summer-17|抽藤、订单、冬访客各1；每个对手可抽1夏访客|resolve|
|summer-18|所有玩家可减2金币建造，每个参与对手令你得1分|resolve|
|summer-19|无视建筑种1藤或拔除弃置2藤得3分|plant, uproot|
|summer-20|弃2手牌抽四类各1|exchange|
|summer-21|得5金币；对手可失1分得3金币|resolve|
|summer-22|原价建造与种1藤；价值4藤得1分，顺序自选|build_plant, plant_build|
|summer-23|对手合计给3访客，否则退还提交并抽3冬访客|resolve|
|summer-24|无视建筑种1藤或拔除弃置1藤得2分|plant, uproot|
|summer-25|种1藤，总共至少6藤得2分|plant|
|summer-26|弃1葡萄得1收入或弃1酒得2收入|grape, wine|
|summer-27|种1藤可超容量，仍需建筑|plant|
|summer-28|建造减3金币或种至多2藤；失1分可依次做两项|build, plant, both, build_plant, plant_build|
|summer-29|放可用工人到冬季格，冬初执行|plan|
|summer-30|种1藤，该田至少3不同藤种得2分|plant|
|summer-31|每个对手可付你2金币，不付则你得1分|resolve|
|summer-32|付2金币取回其他行动上至多2工人|retrieve|
|summer-33|移到空起床行，领奖后结束本季|move|
|summer-34|抽2藤或得3金币，失1分可做两项|draw, coins, both|
|summer-35|得3金币或建造减1金币或种至多2藤|coins, build, plant|
|summer-36|付8金币免费建2座建筑|build|
|summer-37|所有玩家可种1藤，每个参与对手令你得2金币|resolve|
|summer-38|向至多3个不同对手各付2金币，每个得1分|pay|
|winter-01|付3金币得红白品质1葡萄各1或交订单额外1分|buy, fill|
|winter-02|得3金币抽1夏访客或抽1订单酿至多2酒|coins, make|
|winter-03|抽2夏访客或弃品质至少4的酒得3分|draw, discard|
|winter-04|全部酒陈酿两次或付3金币升级酒窖一级|age, upgrade|
|winter-05|抽2夏访客及1金币或交订单额外1分|draw, fill|
|winter-06|得3金币抽1订单或酿至多3酒|coins, make|
|winter-07|失1分免费培训或每个拥有6工人的对手得1分|train, vp|
|winter-08|酿至多2酒或付2金币培训|make, train|
|winter-09|抽1藤及1夏访客或弃2访客得2分|draw, discard|
|winter-10|每张剩余手牌得1金币或弃全部手牌(至少1张)得2分|coins, discard|
|winter-11|右邻座失1分、给2手牌或付3金币|resolve|
|winter-12|收获至多2田并得2金币或1分|coins, vp|
|winter-13|付2金币培训或已有6工人得2分|train, vp|
|winter-14|升级酒窖减2金币或陈酿1酒后交订单|upgrade, fill|
|winter-15|全部酒陈酿两次或失1分免费升级酒窖|age, upgrade|
|winter-16|弃1葡萄或酒得1分|grape, wine|
|winter-17|所有玩家可酿至多2酒，每个参与对手让你抽1藤/夏访客|resolve|
|winter-18|收获1田或抽1藤或付1金币建轭|harvest, draw, build|
|winter-19|取两个不同弃牌堆顶牌，不能取回自身|take|
|winter-20|三选二，顺序自选：收获1田、酿至多2酒、交1订单|harvest_make, make_harvest, harvest_fill, fill_harvest, make_fill, fill_make|
|winter-21|负分得6金币，否则抽藤、夏访客、订单各1|resolve|
|winter-22|酿至多2酒，每瓶起泡酒得1分|make|
|winter-23|抽2订单或付3金币培训，失1分可都做|draw, train, both, draw_train, train_draw|
|winter-24|收获至多3田，收获3田得2分|harvest|
|winter-25|所有玩家可取回大工人，每个参与对手令你得1分|resolve|
|winter-26|酿至多3酒，每种酿出的酒得1分|make|
|winter-27|选两项：抽1订单、原价升级酒窖、得1分|draw_upgrade, upgrade_draw, draw_vp, vp_draw, upgrade_vp, vp_upgrade|
|winter-28|酿至多2酒或交订单或弃1葡萄得2分|make, fill, discard|
|winter-29|收获至多2田或酿至多3酒；失1分可都做|harvest, make, harvest_make, make_harvest|
|winter-30|原价建1建筑，建筑至少6座得2分|build|
|winter-31|付3金币培训立即可用或弃1酒得2分|train, discard|
|winter-32|执行夏季1行动，不放工人且不拿格奖励|action|
|winter-33|酿至多2瓶品质至少4的酒，无视酒窖限制|make|
|winter-34|付1金币得1收入或失2收入得2分|buy, sell|
|winter-35|选至多3对手各给你1夏访客，无法给则你得1分|select|
|winter-36|弃1酒得4金币，严格高于所有其他酒则得2分|discard|
|winter-37|公开四类牌顶各1，取2其余弃置|reveal|
|winter-38|所有玩家可付1金币培训，每个参与对手令你得1分|resolve|

## 验证

visitor_test.go：76张卡逐卡子测试，全部覆盖已公开初始 option 和默认后续。每个已实现卡初始选项都测伪造选项/非本人拒绝和回滚。
另测：多成本后半失败回滚、权限、revision/choiceId 重放、保存失败、Guest Speaker 多人中途重启/恢复/拒绝、Importer 托管成功/退款/重启、Mentor 正确颜色和奖励插队/已pass、Innkeeper 不能取回自身与暂存恢复、Caravan 公开区恢复、已确认关键金币/VP/收入数值、支付下限。
既有两项审计回归为年末方向和 Cottage 不提前曝光。

本轮执行结果将在最终校验后追加。测试使用隔离临时目录/构造合法动作资源夹具，不是浏览器真实完整对局或 Windows 可执行验证。

## 三张特殊访客：本次接入与边界

### Queen winter-11
- 初始 resolve 后仅固定座位右邻座（Players 下标 -1，循环；不是起床顺序）获得 reply。
- reply 的合法 option 动态筛选：vp 失 **1 VP**；cards 由目标提交恰好两个不同 cardIds（任意手牌类型）；coins 目标付3金币给出牌者。
- VP=-5 时不允许 vp，金币不足3不允许 coins，手牌不足2不允许 cards。不能跳过、部分交付、欠债或降到-6。
- 三项全部不能完成时，**打牌前原子拒绝**，保留牌和工人；依据出版商 FAQ “You must be able to do the minimum requirement on the card” 处理强制要求，不是伪造 Queen 特例裁定。此边界由通则推导，未找到单卡独立 FAQ。
- 中途保存恢复、非法目标/选牌/重发都沿用服务器事务；完成后恢复出牌者外层而非从响应者 next。

### Planner summer-29
- plan 提交 space（仅本年 winter 行动）、slot（0自动，1为3+人奖励格，-1为满格大工溢出）及 large。立刻扣额外可用工人并占格，不提前执行、不提前支付行动费用或得奖励。打牌所用工人不算可预约工人。
- Room.planned 持久化公开预约信息；不预存未来私有卡牌选择。全部秋季选择完成后才开始冬初预约；kind=planner，option=execute，提交预约行动的正常参数。space 可省略，不能改成别的行动。服务器固定原格奖励，不再扣工人。
- 预约冬访客可触发多人回复、奖励、双访客和 Manager 嵌套；全部完成才执行下一预约，全部预约完成才进入起床顺序首个正常冬季回合。
- 多预约采用**预约先后顺序**串行；这是明确的数字化顺序约定，尚未找到出版商多预约顺序专门裁定。
- 冬初最低行动已不可能（例如无钱培训、无可收获田、无合法酒/订单）时跳过该预约，无收益且不退工人；与参考引擎的可执行检查交叉一致。被 Producer/Professor 提前取回的预约取消，不会无工人执行。
- 冬访客可执行预检只到可合法开始打牌：其他访客已有的跨请求强制子步骤可行性限制仍存在，并非新增完整的76卡效果搜索器。

### Organizer summer-33
- move 提交空起床行 slot=1..7，第5行同一请求带 color=summer/winter；旧行释放，新行占用，WakeSlots 和 Player.Wake 同步。
- 奖励：1无资源，2藤，3订单，4一金币，5自选访客，6一VP，7本年额外普通工人（不增加 TotalWorkers）。非法行/颜色原子拒绝。
- 奖励后 Context.passAfter 持久化，**待本次外层放置的剩余双访客及其多人回复完成后提交 pass**；不覆盖外层 remaining/space，不提前切全局 Phase。Manager 冬季打本卡仍在冬季结算、最后结束冬季。
- 此“保留双访客后再 pass”的优先级是当前明确实现约定；旧审计要求的出版商独立签核尚未获得，不能把测试通过说成该规则争议已由官方裁定。

### 本次验证
新增 visitor_special_test.go，覆盖 Queen 三分支/右邻座/已pass/-5/全不能执行/伪造选牌/持久化；Organizer 七行奖励/占用与颜色拒绝/双访客/Manager/持久化；Planner 不提前执行/额外工人/冬初奖励/权限与重放/持久化/多人访客和下一预约/无钱耗尽预约/冬季拒绝/大工溢出/取回取消。

实测通过：go test ./...；go test -tags ee_rule_audit ./...；go vet ./...。未修改 web/ 或 ee_cards.json；未启动/替换服务。

## 集成后最终修复
- Importer空Escrow在JSON副本/重启后写入前初始化；真HTTP回归通过，无nil map panic。
- 强制访客出牌前及选二首步骤加入可行性预检，随机抽牌前检验后续培训/升级费用；不允许看牌后撤回或免费skip。旧版已卡死存档不自动修复。
- 公共view.planned投影已补齐。浏览器142普通选项及19特殊场景均完成；自然完整对局另测。
- 全卡版是RC候选，未穷举Manager嵌套及全部数值边界；规则约定和美术范围见README.md。

HTTP: POST /api/action; Authorization: Bearer TOKEN; JSON {type:"choose",choiceId:"当前ID",option:"当前选项",revision:当前revision,...参数}。
打牌: {type:"place",space:"summer_visitor"或"winter_visitor",cardId:"卡ID",large:false,slot:0,revision:当前revision}。
