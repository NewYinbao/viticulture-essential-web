export const baseBuildings = [
  ['trellis', '棚架', 2, 'trellis', '允许种植需要棚架的葡萄藤。'],
  ['irrigation', '灌溉', 3, 'irrigation', '允许种植需要灌溉的葡萄藤。'],
  ['yoke', '轭', 2, 'field', '每年可派工一次：收获一块田地，或拔一张藤回手牌。'],
  ['medium_cellar', '中酒窖', 4, 'cellar', '可储存品质 4–6 的酒；解锁桃红酒（品质至少 4）。'],
  [
    'large_cellar',
    '大酒窖',
    6,
    'cellar',
    '可储存品质 7–9 的酒；解锁起泡酒（品质至少 7）。需先有中酒窖。',
  ],
  ['cottage', '小屋', 4, 'cottage', '每年秋季额外抽一张自选颜色的访客牌。'],
  ['windmill', '磨坊', 5, 'windmill', '种植葡萄藤时获得 1 分，每年最多触发一次。'],
  ['tasting_room', '品酒室', 6, 'estate', '拥有酒时，导览额外获得 1 分，每年最多触发一次。'],
];

export const structureBuildings = [
  ['cask', '木桶', 2, 'building', '使1瓶酒陈酿2级；奖励可抽1订单。'],
  ['aqueduct', '渡槽', 3, 'building', '种植时忽略棚架与灌溉要求。'],
  ['wine_cave', '酒窖洞', 2, 'building', '使至多2瓶酒各陈酿1级；奖励可抽1订单。'],
  ['trading_post', '贸易站', 2, 'building', '执行一换一交易；奖励可放置或移动影响力星。'],
  ['shop', '商店', 5, 'building', '交付1张订单；奖励可放置或移动影响力星。'],
  ['wine_press', '压酒机', 4, 'building', '酿造至多2瓶酒；奖励可放置或移动影响力星。'],
  [
    'school',
    '学校',
    7,
    'building',
    '培训1名本年可用工人（普通免费，特殊加1）；另计学院费，之后得1金币。',
  ],
  ['wine_bar', '酒吧', 4, 'building', '弃1瓶酒得2分。'],
  ['patio', '庭院', 3, 'building', '酿桃红或起泡酒得2金币。'],
  ['ristorante', '餐厅', 8, 'building', '弃1瓶酒和1颗红/白葡萄，得3金币和3分。'],
  ['guest_house', '客房', 3, 'building', '弃2张访客牌；奖励得2分。'],
  ['cafe', '咖啡馆', 3, 'building', '弃1颗红/白葡萄，得3金币和1分。'],
  ['distiller', '蒸馏器', 2, 'building', '年末额外使所有葡萄陈酿1级。'],
  ['mercado', '市场', 5, 'building', '抽到订单时可立即交付。'],
  ['studio', '工作室', 5, 'building', '之后每建造固定建筑或结构牌，额外得1分。'],
  ['barn', '谷仓', 5, 'building', '每年夏末可弃2张牌得1分。'],
  ['academy', '学院', 4, 'building', '对手培训工人时向你支付1金币。'],
  ['gazebo', '凉亭', 3, 'building', '导览时可放置或移动影响力星。'],
  ['workshop', '工坊', 3, 'building', '之后建造固定建筑或结构牌少付1金币。'],
  ['veranda', '阳台', 5, 'building', '每次交付订单额外得1分。'],
  ['wine_parlor', '酒廊', 3, 'building', '每次交付订单额外得2金币。'],
  ['label_factory', '酒标工厂', 3, 'building', '支付3金币交付1张订单；奖励得2分。'],
  ['harvest_machine', '收割机', 2, 'building', '收获时可改为收获所有未收获田地。'],
  ['fermentation_tank', '发酵罐', 4, 'building', '收获后可酿1瓶酒。'],
  ['charmat', '查玛法罐', 3, 'building', '1红+1白可酿起泡酒。'],
  ['inn', '旅店', 4, 'building', '每打出1张访客牌得1金币。'],
  ['tap_room', '品酒吧', 5, 'building', '打出访客后可弃1酒得2分。'],
  ['tavern', '酒馆', 5, 'building', '打出访客后可弃2红/白葡萄得3分。'],
  ['banquet_hall', '宴会厅', 2, 'building', '移动影响力星也获得地区放置奖励。'],
  ['penthouse', '顶层阁楼', 3, 'building', '酿造品质≥7的酒每瓶得1分。'],
  ['fountain', '喷泉', 4, 'building', '对手导览时你得1金币。'],
  ['mixer', '调酒器', 3, 'building', '至多酿1瓶桃红和1瓶起泡，并得1分。'],
  ['storehouse', '仓库', 2, 'building', '年末额外使所有酒陈酿1级。'],
  ['statue', '雕像', 9, 'building', '年末得1分，不能触发最终年。'],
  ['dock', '码头', 3, 'building', '年末抽1张订单。'],
  ['silo', '筒仓', 2, 'building', '年末抽1张葡萄藤。'],
];

export const buildings = [...baseBuildings, ...structureBuildings];

// Fixed buildings are always inspectable; expansion buildings come from the
// player's own cards and estate, so disabled modules never clutter base games.
export function availableBuildings(view) {
  if (!view.config?.structures) return baseBuildings;
  const p = view.players.find((player) => player.id === view.youId);
  const known = new Set([
    ...(view.hand || [])
      .filter((card) => card.type === 'structure')
      .map((card) => card.structureId),
    ...(p?.structureSlots || []),
    ...(p?.fields || []).map((field) => field.structure),
  ]);
  return [...baseBuildings, ...structureBuildings.filter(([id]) => known.has(id))];
}
