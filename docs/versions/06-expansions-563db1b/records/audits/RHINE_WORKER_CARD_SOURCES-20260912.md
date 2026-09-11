> 历史记录 · [所属版本](../../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# Rhine 最后四张规则取证与工程接入依据

核查日期：2026-09-11 至 2026-09-12。研究者只读生产源码，本文记录规则证据，不把代码存在等同于规则通过。实际实现、测试和发布状态由整合报告记录。

## 结论

Laborer、Supervisor、Trainer、Tutor 均已找到实际 Rhine 牌面的照片；Trainer、Tutor 另取得作者或出版社工作人员的逐牌规则答复。可以继续实现，不需要以无效果奖励或 EE ID 别名代替。工人的精确身份、回到未培训供应池、跨年后不会自动复活，是实现必须满足的部分。

## 可复核的实物证据

BGG 用户 Alfredo_VB 同批上传西班牙版《Visitantes del valle del Rin》盒面、盒背、独有牌背及卡牌布局。检索元数据保存在 `bgg-images.json`。不能仅以 BGG 归类证明任意照片都是 Rhine；此组照片有盒面和牌背互证，卡牌也混列 Rhine 独有访客。

| 图片 | 定位与本地缓存 |
| --- | --- |
| [Rhine 冬季卡面](https://boardgamegeek.com/image/9636551/viticulture-visit-from-the-rhine-valley) | `bgg-image-9636551.jpg`，原图 3024×4032。Tutor/MENTOR 第 3 行第 3 列；Laborer/PEÓN 第 4 行第 6 列；Supervisor/SUPERVISORA 第 5 行第 2 列。 |
| [Trainer 清晰卡面](https://boardgamegeek.com/image/9636560/viticulture-visit-from-the-rhine-valley) | `bgg-image-9636560.jpg`。INSTRUCTOR 位于第一行第 3 列。 |
| [夏季卡面与 Wine Engineer](https://boardgamegeek.com/image/9636557/viticulture-visit-from-the-rhine-valley) | `bgg-image-9636557.jpg`。INGENIERO VINÍCOLA 位于第 3 行第 5 列。 |
| [同批 Rhine 盒面](https://boardgamegeek.com/image/9636544/viticulture-visit-from-the-rhine-valley) | `bgg-image-9636544.jpg`，标题、设计者和出版者清晰。 |
| [同批 Rhine 牌背](https://boardgamegeek.com/image/9636547/viticulture-visit-from-the-rhine-valley) | `bgg-image-9636547.jpg`，夏冬均为 Rhine 风景及花环专用背面。 |

这些是玩家拍摄的正式产品照片，属于可直接检查印刷事实的材料，不代表上传者拥有规则裁定权。

## Laborer（冬）

**印刷效果：**二选一：收获至多 2 块田；或酿造至多 3 瓶酒。也可失去 1 VP 执行两项。实物卡面与已有英文转录吻合。

补充英文照片来自 [Board Game Gumbo 的 Rhine 评测](https://boardgamegumbo.wordpress.com/2019/04/15/spice-it-up-with-viticulture-visit-from-the-rhine-valley/) 中 `img_5714.jpg`，本地 `gumbo-9.jpg`。该孤立英文照片的版次不能单凭标题认定，因此最初未拿它独自确认 Rhine；上述西班牙实物整组现在补足这个缺口。

**实现建议：**独立 `rhine-winter-*` ID；按卡面分支执行真实收获/酿造，失去 VP 的双项按所选合法次序及引擎续步结算。不得因为同名同效果将 Rhine 卡从牌堆去重。Harvest/Fermentation 等结构触发仍须保留本张访客的后续阶段。

## Supervisor（冬）

**印刷效果：**酿造至多 2 瓶酒；每瓶本次实际酿成的气泡酒额外获得 1 VP。

除上述 SUPERVISORA 实物图，同一 [Gumbo 评测](https://boardgamegumbo.wordpress.com/2019/04/15/spice-it-up-with-viticulture-visit-from-the-rhine-valley/) 的 `img_5799.jpg`（本地 `gumbo-7.jpg`）可读英文牌面，并与 Bride-to-be、Lovebirds、Grape Whisperer、Enthusiast 等 Rhine 牌同框。

**实现建议：**按本次实际新增的气泡酒计分；已有库存、酿造失败或改成红白酒不得贡献额外 VP。独立 Rhine 卡 ID，不以 EE 引擎 alias 充当规则完成证明。

## Trainer（冬）

**印刷事实：**付 3 币培训 1 名当年即可使用的工人；或失去 1 名仍可用的工人，获得 2 VP。西班牙卡明确有“仍可用”限定。出版者《Viticulture World》规则末页还重印这张不兼容 World 的牌，其英文内容可交叉核对；[出版者 World 规则入口](https://stonemaiergames.com/games/viticulture/viticulture-world/) 链接规则文件。

**逐牌作者裁定：**[Jamey Stegmaier，2020-08-09，Trainer 答复](https://boardgamegeek.com/thread/2481397/article/35537987#35537987)。作者说明此牌针对庄园板上的永久工人，失去后退回可再培训的供应池；不应适用于灰色临时工。该排除优先于仅用通则“任意工人”得出的相反推断。作者以设计意图表述灰工排除，语气有保留，本文保留这一证据性质，不将其改写为印刷原句。

缓存：`bgg-thread-2481397.json`。公开只读接口为 `https://api.geekdo.com/api/articles?threadid=2481397`，含 2 帖。作者身份 `bgg-user-323693.json` 对应 `jameystegmaier`。

**确定的操作限制：**必须仍未派遣；只能永久工；灰工不可牺牲；先失去该工人才得 2 VP；返回未培训供应池，允许未来再次培训。

## Tutor（冬）

**印刷事实：**二选一：失去 1 VP 并付 1 币培训 1 名工人；或失去 1 名工人，弃 X 张手牌，再抽取 X 张任意种类的牌。该牌没有 Trainer 的“仍可用”限定。西班牙 MENTOR 明确使用任意工人措辞；培训分支没有“当年可用”，所以适用普通培训时机。

**作者裁定：**[Jamey Stegmaier，2020-01-03](https://boardgamegeek.com/thread/2341105/article/33701959#33701959) 明确失去的工人返回未培训池，而非整个游戏永久销毁。

**出版社工作人员补充：**[Joe Aubrey，2021-03-22](https://boardgamegeek.com/thread/2341105/article/37309445#37309445) 确认前文解释：可选择已派遣工人或尚可用的工人；弃牌与新抽牌的颜色不必匹配。[同日另一答复](https://boardgamegeek.com/thread/2341105/article/37309842#37309842) 确认移除已派遣工人会腾出该行动位。他的身份可由 [Stonemaier 员工页](https://stonemaiergames.com/about/staff/joe-aubrey/) 以及 BGG `bgg-user-1609837.json` 交叉核对。

缓存：`bgg-thread-2341105.json`，完整 12 帖。普通玩家猜测与 Jamey/Joe 答复分别保留作者，不混淆证据等级。

## 工人类别与实现推导

以下内容是把逐牌裁定与通则合并后的**明确推导**，不声称作者逐项列举了每一种组合：

- [Jamey 在 2015-01-23 的通则答复](https://boardgamegeek.com/thread/1306946/article/18095809#18095809) 明确永久工人包含普通、大工人、特殊工人；灰色临时工不计入 6 名永久工人的上限。缓存 `bgg-thread-1306946.json`。
- 因此 Trainer 可牺牲仍可用的普通、大、特殊工人。Tutor 同样可以失去这些永久工人，但可选择已派遣者。Tutor 的灰工排除由相同的“失去→未培训池”机制和 Trainer 对 lose 的作者解释推导；尚未找到专门点名 Tutor 灰工的作者句子。
- 失去大工人后，必须记录其处于未培训池，不能在下一年自动恢复；该大工人应可按培训规则重新获得。失去特殊工人后，同类型应重新成为可培训选择。退回池与从游戏销毁不是同一行为。
- 没有证据支持人为增加“至少保留 1 名永久工人”门槛，因此不能补造此限制。
- Tutor 移除已派遣工人须精确匹配玩家、行动区域和座位；同种工人不能靠数量差猜测。预约到未来季节的工人被移除时，应清理其预约，以免继续执行免费行动。
- 灰色临时工必须有显式身份。不能把余下的唯一灰工算作普通永久工，亦不能因普通工与灰工同一计数产生可牺牲灰工的漏洞。
- 培训仍须兼容特殊工人额外费用、Academy、School 以及当年是否可用规则。UI 要显示可重新培训的大工人与特殊工人，不能只在后台支持。

## Wine Engineer 的补充裁定

新证据解决之前重复效果的范围问题。

**卡面：**付 2 币，二选一：获得 1 个 4 值酒；或将 1 至 2 颗葡萄陈酿至多 3 次。西班牙卡面照片 `bgg-image-9636557.jpg` 的 OR 及末尾次数从句可直接阅读。

**作者 2019 答复：**[Jamey，2019-02-19](https://boardgamegeek.com/thread/2153880/article/31292084#31292084) 明确是同两颗葡萄各自最多陈酿 3 次，不是总计只有 3 个陈酿次数分摊给两颗。缓存 `bgg-thread-2153880.json`。

**作者 2021 答复：**[出版者 Rhine 页面旧评论](https://stonemaiergames.com/games/viticulture/visit-from-the-rhine-valley/comment-page-1/) 中 Samuel Hung 于 2021-05-02 询问做 3 次支付 2 还是 6 币；Jamey 同日回答需要 6 币。原文保存在 `official-older-comments.txt`。这条较新答复明确了重复成本，但没有说明可换新葡萄或交替执行产酒分支。

**建议执行两条裁定的交集，已与实现代理对齐：**

1. 产酒分支付 2 币获得 1 个 4 值酒，执行一次；仍遵守卡面没有豁免的酒窖条件。
2. 陈酿分支整张牌选定至多 2 颗葡萄。每轮付 2 币，将所选葡萄陈酿一次；至多 3 轮，所以最多付 6 币。
3. 后续轮只允许最初那至多 2 颗葡萄中仍可陈酿的子集；不能换第三颗，也不能改为产酒。
4. 费用不足、没有仍可陈酿的已选葡萄时不能继续扣钱。更新后的葡萄身份必须随选择持久化，服务重启后仍有效。

这是对卡文与两条作者答复的合并解释；“可换葡萄”“可重复产酒”“混用两分支”均没有来源支持，不予放宽。

## 尚需工程验收，不属于证据已完成的部分

规则研究可以闭环，但以下实现必须由整合测试证明：Trainer 灰工拒绝且状态回滚；Tutor 精确移除可用/已派遣/预约工；大工与特殊工失去后的跨年和重新培训；School 的当年可用、Academy 的费用；存档迁移；Wine Engineer 固定葡萄身份、每轮费用、至多三轮及重启恢复。本文不替这些代码验收预先盖章。

附：`bgg-image-4300430.jpg` 虽被归在 Rhine 图库，实际多为 EE 英德访客对照，未用作本次四张的证据；低分辨率冬牌总览与卡文被遮挡的图片同样没有用于推断看不到的规则。
