# EE development API contract

**EE all-card RC1, not official certification**. Ruleset ee-base-v1, 2–6 human players.

- Existing create/join/state/events/action endpoints remain. Actions require session token and current revision; stale revision returns 409, invalid action 400, with no committed mutation.
- Phases: lobby → setup → wake → summer → fall → winter → year_end → wake/finished.
- legal.canChoose controls pending choices. Only head choice owner may submit type=choose,choiceId,option,revision. Other players see only playerId/kind, not actionable ID/options.
- Papa choice: kind=papa, options gift|coins. Own parentOptions={gift,coins,baseCoins} gives building ID/worker/vp, alternative additional cash, and already-granted base cash. Public players[].mama/papa/papaResolved identify parents. Mama hand and base cash already granted; gift and alternative coins are mutually exclusive.
- Fall: kind=fall, options summer|winter; cottage adds another queued choice.
- Discard: kind=discard,count; submit exactly count distinct own cardIds, no option needed.
- Wake: type=wake,slot=1..7; slot 5 requires color=summer|winter.
- Placement: type=place,space,large,slot. Slot 0 auto-selects, positive selects explicitly, -1 is grande overflow only when normal slots full. Slot 1 bonus only with 3+ players. gain_coin unlimited, yoke private once/year requiring building.
- Plant: cardId/field or cardIds/fields (bonus up to 2). Harvest: field or fields (bonus up to 2). All indices zero-based.
- Wine: recipes=[[grapeIndex],...] uses pre-action grape indices, max 2 wines (bonus 3); legacy grapes makes one recipe.
- Order: cardId/wineIds. Build: building. Sell grapes: grapes indices.
- Field trade: space=sell_grapes,mode=sell_field|buy_field,field. Yoke: mode=harvest|uproot,field plus cardId for uproot; plant-space also supports mode=uproot.
- State contains own hand, other handCount, deckCounts[type].deck/discard only. No deck sequence/opponent hand IDs. Save files contain private state and must not be shared.
- winnerIds contains exact ties; winnerId is first top-ranked player for compatibility.

## 当前验收状态
76访客已实现。后续协议与卡牌选项见 VISITOR_API.md。公开视图含 planned（玩家/行动/格/工人，不含未来私有选牌）及 Caravan revealed。Cottage count=2 使用组合选项一次提交。规则约定和未穷举边界见 README.md。
Windows目标：Viticulture-EE-AllCards-RC1.exe，端口3012，独立data-ee-allcards；旧版未覆盖。
