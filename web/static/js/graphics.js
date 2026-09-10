// Original offline artwork; public state only.
const node = (tag, text, cls) => {
  const n = document.createElement(tag);
  if (text != null) n.textContent = text;
  if (cls) n.className = cls;
  return n;
};
export const actionArt = {
  draw_vine: 'market',
  tour: 'estate',
  build: 'cottage',
  plant: 'trellis',
  summer_visitor: 'summer',
  sell_grapes: 'vine',
  draw_order: 'order',
  harvest: 'field',
  make_wine: 'cellar',
  fill_order: 'wine',
  train: 'train',
  winter_visitor: 'winter',
};
export function art(name, cls = '') {
  const img = node('img', null, cls);
  img.src = '/art/' + name + '.svg';
  img.alt = '';
  img.setAttribute('aria-hidden', 'true');
  img.draggable = false;
  return img;
}
export function worker(v, id, large = false) {
  const i = Math.max(
    0,
    v.players.findIndex((p) => p.id === id),
  );
  const n = art('worker-' + i, 'meeple' + (large ? ' large' : ''));
  n.removeAttribute('aria-hidden');
  n.alt = (v.players[i]?.name || '玩家') + (large ? '的大工人' : '的普通工人');
  return n;
}
export function decorateTable(v) {
  document.body.dataset.season = v.phase;
  const track = document.querySelector('#season-track');
  track.replaceChildren();
  for (const [id, title, desc] of [
    ['wake', '春', '选择顺序'],
    ['summer', '夏', '经营葡萄园'],
    ['fall', '秋', '选择访客'],
    ['winter', '冬', '酿酒与交付'],
  ]) {
    const n = node('div', null, 'season-step' + (v.phase === id ? ' current' : ''));
    n.append(node('b', title), node('span', desc));
    if (v.phase === id) n.setAttribute('aria-current', 'step');
    track.append(n);
  }
  document.querySelector('#year-seal').textContent =
    v.phase === 'lobby' ? '待开局' : String(v.year).padStart(2, '0');
}
function quality(p, type, wine, types) {
  const row = node('div', null, 'quality-row ' + type);
  row.append(node('span', (types[type] || type) + (wine ? '酒' : '葡萄'), 'quality-label'));
  const slots = node('div', null, 'quality-slots');
  const max = wine
    ? p.buildings.includes('large_cellar')
      ? 9
      : p.buildings.includes('medium_cellar')
        ? 6
        : 3
    : 9;
  const min = wine ? (type === 'blush' ? 4 : type === 'sparkling' ? 7 : 1) : 1;
  for (let value = 1; value <= 9; value++) {
    const items = (wine ? p.wines : p.grapes).filter(
      (x) => (wine ? x.type : x.color) === type && x.value === value,
    );
    const locked = value > max || value < min;
    const slot = node(
      'span',
      null,
      'quality-slot' + (locked ? ' locked' : '') + (items.length ? ' filled' : ''),
    );
    slot.append(node('span', value, 'quality-number'));
    if (items.length) slot.append(art((wine ? 'bottle-' : 'grape-') + type, 'resource-token'));
    const label =
      (types[type] || type) +
      (wine ? '酒' : '葡萄') +
      '品质' +
      value +
      '：' +
      (items.length ? '已有资源' : locked ? '不可用' : '空槽');
    slot.title = label;
    slot.setAttribute('aria-label', label);
    slots.append(slot);
  }
  row.append(slots);
  return row;
}

export function renderEstate(v, p, buildings, types) {
  const you = p.id === v.youId,
    box = node(
      'article',
      null,
      'player' + (you ? ' you' : '') + (p.id === v.turnId ? ' active' : ''),
    );
  const heading = node('div', null, 'player-title');
  heading.append(
    worker(v, p.id),
    node('h3', p.name + (you ? ' · 你的酒庄' : '') + (p.id === v.hostId ? ' · 房主' : '')),
    node('span', p.id === v.turnId ? '行动中' : p.passed ? '已结束本季' : '等待', 'turn-pill'),
  );
  box.append(heading);
  const stats = node('div', null, 'stats');
  for (const [label, value] of [
    ['胜利分', p.vp],
    ['金币', p.coins],
    ['年收入', p.income],
    ['手牌', p.handCount],
  ]) {
    const a = node('span');
    a.append(node('b', value), node('small', label));
    stats.append(a);
  }
  box.append(stats);
  const publicHand = node('div', null, 'public-hand-counts');
  for (const type of ['vine', 'order', 'summer', 'winter'])
    publicHand.append(node('span', (types[type] || type) + ' ' + (p.handCounts?.[type] || 0)));
  publicHand.setAttribute('aria-label', '公开手牌类型数量');
  box.append(publicHand);
  const reserve = node('div', null, 'worker-reserve');
  reserve.append(node('span', '待命工人', 'mini-label'));
  for (let i = 0; i < p.workers; i++) reserve.append(worker(v, p.id));
  if (p.largeWorker) reserve.append(worker(v, p.id, true));
  reserve.append(
    node(
      'small',
      '普通 ' + p.workers + ' / 大 ' + (p.largeWorker ? 1 : 0) + ' · 总数 ' + p.totalWorkers,
    ),
  );
  box.append(reserve);
  const content = node(you ? 'div' : 'details', null, 'estate-content');
  if (!you) content.append(node('summary', '查看公开庄园 · 田地 / 建筑 / 酒窖'));
  const fields = node('div', null, 'fields');
  for (const f of p.fields) {
    const e = node('div', null, 'field' + (f.harvested ? ' harvested' : ''));
    e.append(
      art(f.vines?.length ? 'field' : 'fallow', 'field-art'),
      node('strong', '田地 ' + (f.index + 1) + (f.sold ? ' · 已出售' : '')),
      node('span', '容量 ' + f.capacity, 'field-capacity'),
      node(
        'small',
        f.vines?.length ? f.vines.map((c) => '红' + c.red + ' 白' + c.white).join(' / ') : '待种植',
      ),
    );
    if (f.harvested) e.append(node('span', '本年已收获', 'harvest-stamp'));
    fields.append(e);
  }
  content.append(fields);
  const structures = node('div', null, 'structures'),
    icons = {
      trellis: 'trellis',
      irrigation: 'irrigation',
      medium_cellar: 'cellar',
      large_cellar: 'cellar',
      cottage: 'cottage',
      windmill: 'windmill',
      tasting_room: 'estate',
      yoke: 'field',
    };
  for (const [id, label] of Object.entries(buildings)) {
    const built = p.buildings.includes(id),
      n = node('div', null, 'structure' + (built ? ' built' : ''));
    n.append(
      art(icons[id], 'building-art'),
      node('span', label.split(' · ')[0]),
      node('small', built ? '已建成' : label.split(' · ')[1] + ' · 未建'),
    );
    structures.append(n);
  }
  content.append(
    node('h4', '酒庄建筑', 'board-caption'),
    structures,
    node('h4', '压榨槽 · 葡萄品质', 'board-caption'),
  );
  for (const type of ['red', 'white']) content.append(quality(p, type, false, types));
  content.append(node('h4', '私人酒窖 · 葡萄酒品质', 'board-caption'));
  for (const type of ['red', 'white', 'blush', 'sparkling'])
    content.append(quality(p, type, true, types));
  content.append(node('small', '斜线槽尚不可用：受酒窖等级 / 酒种最低品质限制。', 'cellar-note'));
  box.append(content);
  return box;
}
