// Public-state hints only. The server validates every submitted action.
export function freeSlots(space) {
  return Array.from({ length: space.capacity }, (_, i) => i + 1).filter(
    (slot) => !(space.occupied || []).some((seat) => seat.slot === slot),
  );
}

export function privateSpace(space) {
  return ['gain_coin', 'yoke'].includes(space.id);
}

export function defaultLarge(space, player) {
  return (
    !!player.largeWorker &&
    (player.workers === 0 || (!privateSpace(space) && !freeSlots(space).length))
  );
}

export function actionReason(space, view) {
  if (view.actionReasons?.[space.id]) return view.actionReasons[space.id];
  const p = view.players.find((p) => p.id === view.youId);
  if (space.disabledReason) return space.disabledReason;
  const placements = view.workerPlacements?.[space.id];
  if (placements && !placements.length) return '没有可进入的行动格，检查工人和通行费';
  if (!placements && !p.workers && !p.largeWorker) return '没有待命工人，可结束本季';
  if (space.id === 'yoke' && (!p.buildings.includes('yoke') || p.yokeUsed))
    return p.yokeUsed ? '本年已使用轭' : '需要先建造轭';
  if (!placements && !privateSpace(space) && !freeSlots(space).length && !p.largeWorker)
    return '普通格已满，且大工人已使用';
  const type = {
    plant: 'vine',
    fill_order: 'order',
    summer_visitor: 'summer',
    winter_visitor: 'winter',
  }[space.id];
  if (
    type &&
    !(view.hand || []).some(
      (c) =>
        c.type === type ||
        (view.config?.visitors === 'ee_moor' && type === 'summer' && c.id === 'moor-winter-10'),
    )
  )
    return '没有对应类型的手牌';
  if (space.id === 'make_wine' && !p.grapes.length) return '需要先收获葡萄';
  if (space.id === 'harvest' && !p.fields.some((f) => !f.sold && !f.harvested && f.vines.length))
    return '没有可收获的田地';
  return '';
}

// Server physical slots are authoritative, including two-player Tuscany bonuses.
export function bonusKey(space, slot = 0, decline = false) {
  const resolved = Number(slot) || freeSlots(space)[0] || -1;
  if (decline || privateSpace(space) || resolved < 1 || resolved > space.capacity) return '';
  if (space.bonusSlots != null) return space.bonusSlots[resolved] || '';
  if (space.capacity < 2 || resolved !== 1) return '';
  return (
    {
      tour: 'coin',
      build: 'discount',
      train: 'discount',
      fill_order: 'vp',
      sell_grapes: 'vp',
      summer_visitor: 'visitor',
      winter_visitor: 'visitor',
    }[space.id] || space.id
  );
}
export const bonusNames = {
  coin: '+1金币',
  discount: '少付1金币',
  vp: '+1胜利分',
  influence: '额外放／移1星',
  trade: '额外交易1次',
  visitor: '最多打2张访客',
  plant: '最多种2张藤',
  harvest: '最多收获2田',
  make_wine: '最多酿3瓶',
  draw_vine: '+1葡萄藤',
  draw_order: '+1订单',
  build_tour: '建造少付1／导览多得1金币',
};
export function bonusLabel(space, slot = 0) {
  const key = bonusKey(space, slot);
  return key ? bonusNames[key] || '未识别奖励（请核对服务器）' : '无奖励';
}
export function hasBonus(space, slot = 0, decline = false) {
  return !!bonusKey(space, slot, decline);
}

export function winePreview(player, recipes, recipeTypes = []) {
  const wines = [...player.wines],
    names = {
      red: '红酒',
      white: '白酒',
      blush: '桃红酒',
      sparkling: '起泡酒',
    };
  const cap = player.buildings.includes('large_cellar')
    ? 9
    : player.buildings.includes('medium_cellar')
      ? 6
      : 3;
  return recipes.map((indices, i) => {
    const grapes = indices.map((index) => player.grapes[index]);
    const red = grapes.filter((g) => g.color === 'red').length;
    const white = grapes.length - red;
    const type =
      red === 1 && white === 0
        ? 'red'
        : red === 0 && white === 1
          ? 'white'
          : red === 1 && white === 1
            ? recipeTypes[i] === 'sparkling' && player.buildings.includes('charmat')
              ? 'sparkling'
              : 'blush'
            : red === 2 && white === 1
              ? 'sparkling'
              : '';
    const prefix = `第 ${i + 1} 瓶：`;
    if (recipeTypes[i] === 'sparkling' && type !== 'sparkling')
      return prefix + '配方不成立；起泡需红白各一（查玛法罐）或两红一白';
    if (!type) return prefix + '配方不成立；需单颗红／白、红白各一，或两红一白';
    const min = type === 'blush' ? 4 : type === 'sparkling' ? 7 : 1;
    const raw = grapes.reduce((sum, g) => sum + g.value, 0);
    let value = Math.min(cap, raw);
    while (wines.some((w) => w.type === type && w.value === value)) value--;
    if (value < min) return prefix + names[type] + '无可用酒槽（检查酒窖等级、最低品质和占位）';
    wines.push({ type, value });
    return prefix + names[type] + ` · 品质 ${value}` + (value < raw ? '（按酒窖／占位降级）' : '');
  });
}

// Planting requirements are shown together, including every missing structure.
export function vineRequirements(player, card) {
  const value = card.red + card.white;
  const fields = player.fields.map((field) => ({
    index: field.index,
    free: field.capacity - field.vines.reduce((sum, vine) => sum + vine.red + vine.white, 0),
    sold: field.sold,
    structure: field.structure,
  }));
  const checks = [
    {
      label: '棚架',
      needed: card.trellis,
      met:
        !card.trellis ||
        player.buildings.includes('trellis') ||
        player.buildings.includes('aqueduct'),
    },
    {
      label: '灌溉',
      needed: card.irrigation,
      met:
        !card.irrigation ||
        player.buildings.includes('irrigation') ||
        player.buildings.includes('aqueduct'),
    },
    {
      label: '田地容量 ≥' + value,
      needed: true,
      met: fields.some((f) => !f.sold && !f.structure && f.free >= value),
    },
  ];
  return {
    value,
    fields,
    checks,
    reason: checks
      .filter((c) => !c.met)
      .map((c) =>
        c.label === '棚架' || c.label === '灌溉'
          ? '缺少' + c.label
          : fields.every((f) => f.sold)
            ? '田地均已出售'
            : fields.every((f) => f.sold || f.structure)
              ? '田地已出售或被结构占用'
              : '田地容量不足',
      )
      .join('；'),
  };
}
