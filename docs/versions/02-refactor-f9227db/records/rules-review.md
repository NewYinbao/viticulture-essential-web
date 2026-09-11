> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# EE 本体规则核对（2026-09-11）

范围：Essential Edition 本体，2–6 位真人。参考开源实现时排除 Tuscany、Plus 和 Automa 规则。此次核对是针对代码与既有审计的修正，并非对所有访客组合的官方认证。

## 实际查阅的来源

- [出版方规则入口](https://stonemaiergames.com/games/viticulture/rules/)与[官方 FAQ](https://stonemaiergames.com/games/viticulture/faq/)。官网通过普通 HTTP 下载读取；研究工具的页面提取多次返回 502。
- [出版方编写的 EE 英文说明书，Steam 手册镜像](https://shared.steamstatic.com/store_item_assets/steam/apps/414235/manuals/Viticulture_EE_Rules.pdf?t=1551153655)：20 页扫描 PDF，文件元数据为 2016 年。已渲染并目视核对第 3、4、8 页；不是仅凭搜索摘要判断。
- [Boardspace 的 Viticulture Java 实现](https://github.com/ddyer0/boardspace.net/tree/main/client/boardspace-java/boardspace-games/viticulture)和[其游戏说明](https://boardspace.net/english/about_viticulture.html)：辅助比较 Plant/Yoke 分派及终局检查。其默认使用 Tuscany Essential，并包含独立变体，不能代替本体说明书。

研究 PDF、网页副本与参考 Java 文件只保存在被 Git 忽略的本地研究目录。本次没有复制开源 Java 代码，也没有把说明书图像加入应用。

## 已确认并修复

| 问题 | 依据 | 结果与验证 |
| --- | --- | --- |
| 公共 Plant 行动允许 mode=uproot | 说明书第 8 页把拔藤限定于轭和允许拔藤的访客；Boardspace 也分别分派 Plant 与 Yoke | 后端拒绝这种请求；种植表单移除拔藤选项。原有 TestAuditDiffPlantMustNotOfferUproot 从失败变为通过，轭／访客拔藤保留 |
| Papa 赠礼按加入顺序选择 | 第 4 页：从随机起始玩家开始，顺时针选择 | 发牌不变，选择队列按 SpringLeader 轮转。新增六个起始座位（含跨尾部）的回归；持久化所有权测试改为选择真正的非拥有者 |
| 仅公开对手手牌总数 | 第 3 页 Note 2 明确手牌数量与类型是公开信息 | players[].handCounts 展示藤／订单／夏访客／冬访客数量；牌面和 ID 仍隐藏。测试验证数量及无身份泄露、无视图副作用 |

## 核对后保持的行为

官方 FAQ 明确：必须执行主行动，行动格奖励可以放弃；含 up to 的行动至少执行一次；同一田地每年最多收获一次；每个订单符号须由一枚相同酒种、品质足够的酒满足。既有修复已包含这些规则，本次保留相关审计并全部运行。UI 默认领取可用奖励，但仍提供放弃选项。

2 人每个行动只开放一格且无格位奖励；3–4 人开放两格；5–6 人开放三格。大工人占空格时正常得到奖励，满格溢出不获得奖励。UI 现在按真实 slot 画工人，不按占位数组的插入顺序画。

## 尚需进一步裁定／独立核对

1. Planner 多个预约按预约先后结算、Organizer 在外层双访客结算结束后 pass，是原有公开实现约定。当前测试验证实现自洽，不能据此声称特殊组合与纸面游戏完全等价。
2. 玩家在年内达到 20 分后又跌回 20 分以下：说明书的一般终局表述可能被理解为立即锁定最后一年；当前代码在年末看分数，参考的 Boardspace 也检查年末最大分数。本次没有找到针对这一边界的明确官方裁定，未擅自改变。
3. 年末代码先排队弃至 7 张，再统一陈年、回收工人及领取收入；纸面说明书按陈年、回收、收入、弃牌列出。在目前没有年末插入行动的流程中，最终资源结果相同，但弃牌界面的公开资源更新时间与说明书顺序不同。

既有逐卡审计见 [audits](../../01-baseline-cea1f34/records/audits)，此前的实现记录见 [history](../../01-baseline-cea1f34/records)。其中旧的“Plant 待裁定／失败”结论已被本次说明书证据和修正替代。
