> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# 36 张 Tuscany 结构卡卡面复核与浏览器触发（2026-09-11）

逐张查看出版社实际卡面在 Boardspace 的编号图片，未使用旧阶段审查结论作为规则依据。下面列出独立手工转述；编号链接可直接定位卡面。所有卡的建造分均为 1 分。Go 测试 `TestPrintedStructureCardCostsAndCategories` 对照手工抄录费用、派工虚线框和残余循环图标，36 项通过。

| 卡面编号/名称 | 建造费用 | 类型 | 印刷效果核对 | 浏览器触发 |
|---|---:|---|---|---|
| [01 木桶](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000001.jpg) | 2 | 派工 | 1酒陈酿两次，抽1订单 | `CARD01` 通过 |
| [02 渡槽](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000002.jpg) | 3 | 增强 | 忽略种藤所需建筑 | `CARD02` 通过 |
| [03 酒窖洞](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000003.jpg) | 2 | 派工 | 至多2酒各陈酿一次，抽1订单 | `CARD03` 通过 |
| [04 贸易站](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000004.jpg) | 2 | 派工 | 交易一次，放置/移动1星 | `CARD04` 通过 |
| [05 商店](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000005.jpg) | 5 | 派工 | 交付1订单，放置/移动1星 | `CARD05` 通过 |
| [06 压酒机](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000006.jpg) | 4 | 派工 | 酿至多2酒，放置/移动1星 | `CARD06` 通过 |
| [07 学校](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000007.jpg) | 7 | 派工 | 免费训练本年可用工人，得1金币；特殊另加1及学院费 | `CARD07` 通过 |
| [08 酒吧](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000008.jpg) | 4 | 派工 | 弃1酒得2分 | `CARD08` 通过 |
| [09 庭院](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000009.jpg) | 3 | 增强 | 每酿1桃红/起泡得2金币 | `CARD09` 通过 |
| [10 餐厅](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000010.jpg) | 8 | 派工 | 弃1酒及1葡萄，得3金币和3分 | `CARD10` 通过 |
| [11 客房](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000011.jpg) | 3 | 派工 | 弃2访客得2分 | `CARD11` 通过 |
| [12 咖啡馆](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000012.jpg) | 3 | 派工 | 弃1葡萄得3金币和1分 | `CARD12` 通过 |
| [13 蒸馏器](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000013.jpg) | 2 | 年末 | 年末额外陈酿葡萄 | `CARD13` 通过 |
| [14 市场](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000014.jpg) | 5 | 增强 | 抽到的订单可立即交付 | `CARD14` 通过 |
| [15 工作室](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000015.jpg) | 5 | 增强 | 之后每建1建筑得1分 | `CARD15` 通过 |
| [16 谷仓](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000016.jpg) | 5 | 增强 | 每年夏末可弃2手牌得1分 | `CARD16` 通过 |
| [17 学院](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000017.jpg) | 4 | 增强 | 对手每训练1工人付你1金币 | `CARD17` 通过 |
| [18 凉亭](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000018.jpg) | 3 | 增强 | 每次导览可放置/移动1星 | `CARD18` 通过 |
| [19 工坊](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000019.jpg) | 3 | 增强 | 之后建筑费用减1金币 | `CARD19` 通过 |
| [20 阳台](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000020.jpg) | 5 | 增强 | 每交付1订单额外得1分 | `CARD20` 通过 |
| [21 酒廊](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000021.jpg) | 3 | 增强 | 每交付1订单额外得2金币 | `CARD21` 通过 |
| [22 酒标工厂](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000022.jpg) | 3 | 派工 | 付3金币交付1订单，额外得2分 | `CARD22` 通过 |
| [23 收割机](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000023.jpg) | 2 | 增强 | 可将一次收获改为收获全部田地 | `CARD23` 通过 |
| [24 发酵罐](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000024.jpg) | 4 | 增强 | 收获一块或多块田后可酿1酒 | `CARD24` 通过 |
| [25 查玛法罐](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000025.jpg) | 3 | 增强 | 1红1白可选酿起泡酒 | `CARD25` 通过 |
| [26 旅店](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000026.jpg) | 4 | 增强 | 每打出1访客得1金币 | `CARD26` 通过 |
| [27 品酒吧](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000027.jpg) | 5 | 增强 | 每打出1访客可弃1酒得2分 | `CARD27` 通过 |
| [28 酒馆](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000028.jpg) | 5 | 增强 | 每打出1访客可弃2葡萄得3分 | `CARD28` 通过 |
| [29 宴会厅](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000029.jpg) | 2 | 增强 | 移星也获得该地区奖励 | `CARD29` 通过 |
| [30 顶层阁楼](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000030.jpg) | 3 | 增强 | 每酿1品质7以上酒得1分 | `CARD30` 通过 |
| [31 喷泉](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000031.jpg) | 4 | 增强 | 对手每次导览，你得1金币 | `CARD31` 通过 |
| [32 调酒器](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000032.jpg) | 3 | 派工 | 至多1桃红及1起泡，得1分 | `CARD32` 通过 |
| [33 仓库](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000033.jpg) | 2 | 年末 | 年末额外陈酿酒 | `CARD33` 通过 |
| [34 雕像](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000034.jpg) | 9 | 年末 | 年末得1分，不触发最终年 | `CARD34` 通过 |
| [35 码头](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000035.jpg) | 3 | 年末 | 年末抽1订单 | `CARD35` 通过 |
| [36 筒仓](https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000036.jpg) | 2 | 年末 | 年末抽1藤 | `CARD36` 通过 |

## 本轮新增修复

1. Workshop 的“建筑”同时涵盖固定建筑和结构牌。修复固定建筑及访客折扣收费，卡片报价和缺钱说明同步。
2. Mixer 只能酿至多 1 桃红和 1 起泡，不能同类各酿两瓶；服务端原子校验与面板提交条件一致。
3. Fermentation Tank 在轭、单田和访客收获也触发。一次收获多田只给一次奖励；奖励选择不覆盖外层 `PendingAction`。可选结构奖励若将导致排队的强制访客步骤无法完成，会完整回滚该选择。
4. Charmat 的能力可在公共酿酒、Wine Press、Mixer、Fermentation Tank 和访客酿酒使用。显式 `recipeTypes` 让每瓶独立选择常规/红白起泡，支持同批一桃红一起泡；未拥有能力、品质/酒窖不足或类型不符均拒绝。
5. Academy 在访客培训也收费。统一 `visitorTrain(p,cost,now,workerType...)` 的 cost 表示实际基础费用，另计特殊工人 1 金币和对手学院费；即时特殊工人只改变自己的就绪年份。
6. 修正文案：Wine Bar 不重复暗示总得4分；Studio 包含固定建筑；School 可训练特殊工人；Barn 是夏末增强效果。
7. Side 2 不再提前返回而隐藏影响力地区；七区、Lucca 抽结构奖励正常显示。移星时拥有 Banquet Hall 的即时奖励提示正确；移除原来仅画六区的不完整示意图。

## 浏览器验收方法

`internal/game/structure_browser_fixture_test.go` 通过环境变量 `STRUCTURE_BROWSER_FIXTURES` 导出明确标注的测试夹具。`tests/e2e/structures-browser-test.cjs` 启动自己的临时服务、加载这些夹具，通过真实 Edge 浏览器和可见控件逐一提交行动，不在页面注入替代渲染或直接调用行动 API。API 仅用于读取本人的 View 和响应结果。

全部 36 项主触发通过，包括 12 张私人行动和 24 张增强/残余结构。断言分别检查卡面所要求的金币、分数、牌数、酿酒类型、陈酿品质、工人、收获数量或影响力；不只检查 HTTP 200。Side 2 七个地区节点在全部用例存在，页面没有横向溢出和 JavaScript 异常。脚本自动接受其发起的“结束季节”浏览器确认框。

本轮最终运行二进制 SHA256：`CF7F583A2DBFDC69881B1241C4B00E6CCA0DF767681D8FFB67CDD6CBA8AB0E95`。完整结果和截图位于执行副本 `structures-browser-evidence/structures-browser-result.json`，状态 `passed`、36 条记录。还保存了 Charmat、Mixer、Workshop 和 Fermentation 的实际资源面板截图。

这 36 条为 seeded fixture 主触发覆盖，并非自然局，也不能视为每张卡所有组合和分支均已穷尽。自然局另见 8 组合验收记录。带 `ee_rule_audit,audit_diff` 标签的游戏包测试通过。Moor/Rhine 合并后的嵌套访客和 24 组合、全工程测试，以及全新独立 review 仍应由主任务继续。
