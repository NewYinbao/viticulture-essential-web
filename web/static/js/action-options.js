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
  const p = view.players.find((p) => p.id === view.youId);
  if (space.disabledReason) return space.disabledReason;
  if (!p.workers && !p.largeWorker) return '没有待命工人，可结束本季';
  if (space.id === 'yoke' && (!p.buildings.includes('yoke') || p.yokeUsed))
    return p.yokeUsed ? '本年已使用轭' : '需要先建造轭';
  if (!privateSpace(space) && !freeSlots(space).length && !p.largeWorker)
    return '普通格已满，且大工人已使用';
  const type = {
    plant: 'vine',
    fill_order: 'order',
    summer_visitor: 'summer',
    winter_visitor: 'winter',
  }[space.id];
  if (type && !(view.hand || []).some((c) => c.type === type)) return '没有对应类型的手牌';
  if (space.id === 'make_wine' && !p.grapes.length) return '需要先收获葡萄';
  if (space.id === 'harvest' && !p.fields.some((f) => !f.sold && !f.harvested && f.vines.length))
    return '没有可收获的田地';
  return '';
}

export function hasBonus(space, slot = 0, decline = false) {
  return (
    !decline &&
    !privateSpace(space) &&
    space.capacity >= 2 &&
    (Number(slot) || freeSlots(space)[0] || -1) === 1
  );
}

export function winePreview(player, recipes) {
  const wines = [...player.wines],
    names = { red: '红酒', white: '白酒', blush: '桃红酒', sparkling: '起泡酒' };
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
            ? 'blush'
            : red === 2 && white === 1
              ? 'sparkling'
              : '';
    const prefix = `第 ${i + 1} 瓶：`;
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
