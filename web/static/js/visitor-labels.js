// Compact labels keep common costs and outcomes visible when rules are collapsed.
const effects = {
  'summer-02': { buy: '付 9 币 → 3 分', sell: '失 2 分 → 6 币' },
  'summer-03': { draw: '抽 2 冬访客', discard: '弃 ≥7 酒 → 4 分' },
  'summer-06': { coins: '得 4 金币', harvest: '收获 1 田' },
  'summer-07': { coins: '得 3 金币', make: '酿至多 2 酒' },
  'summer-08': { buy: '付 6 币 → 2 分', sell: '失 3 分 → 9 币' },
  'summer-09': { plant: '种至多 2 藤 + 1 币', uproot: '弃田上 1 藤 → 2 分' },
  'summer-10': { buy: '付 2 币 → 品质 1 葡萄', discard: '弃葡萄 → 2 币 + 1 分' },
  'summer-15': { coins: '弃 2 牌 → 4 币', vp: '弃 4 牌 → 3 分' },
  'summer-16': { draw: '付 4 币 → 3 冬访客', discard: '弃 1 酒 + 3 访客 → 3 分' },
  'summer-26': { grape: '弃葡萄 → 1 收入', wine: '弃酒 → 2 收入' },
  'summer-34': { draw: '抽 2 藤', coins: '得 3 金币', both: '失 1 分 → 两项' },
  'summer-36': { build: '付 8 币 → 建 2 座建筑' },
  'winter-01': { buy: '付 3 币 → 红白各 1 葡萄', fill: '交订单 + 1 分' },
  'winter-03': { draw: '抽 2 夏访客', discard: '弃 ≥4 酒 → 3 分' },
  'winter-04': { age: '酒陈年 2 次', upgrade: '付 3 币 → 升级酒窖' },
  'winter-08': { make: '酿至多 2 酒', train: '付 2 币 → 培训' },
  'winter-09': { draw: '抽藤 + 夏访客', discard: '弃 2 访客 → 2 分' },
  'winter-13': { train: '付 2 币 → 培训', vp: '已有 6 工人 → 2 分' },
  'winter-15': { age: '酒陈年 2 次', upgrade: '失 1 分 → 升级酒窖' },
  'winter-18': { harvest: '收获 1 田 + 抽藤', build: '付 1 币 → 建轭' },
  'winter-23': { draw: '抽 2 订单', train: '付 3 币 → 培训', both: '失 1 分 → 两项' },
  'winter-28': { make: '酿至多 2 酒', fill: '交订单', discard: '弃葡萄 → 2 分' },
  'winter-31': { train: '付 3 币 → 即用工人', discard: '弃酒 → 2 分' },
  'winter-34': { buy: '付 1 币 → 1 收入', sell: '失 2 收入 → 2 分' },
};
export function visitorLabel(choice, option, player) {
  if (choice.visitor?.stage !== 'effect') return '';
  const id = choice.visitor.cardId;
  if (id === 'summer-01') {
    const count = player.fields.filter(
      (f) => !f.sold && (option === 'coins' ? !f.vines.length : f.vines.length),
    ).length;
    return option === 'coins' ? `得 ${count * 2} 金币` : `得 ${count} 分`;
  }
  if (id === 'winter-10')
    return option === 'coins' ? `得 ${player.handCount} 金币` : '弃全部手牌 → 2 分';
  return effects[id]?.[option] || '';
}
