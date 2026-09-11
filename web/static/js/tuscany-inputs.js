import { localCard } from './card-i18n.js';
export const regionRewards = {
  winter: '抽1冬访客',
  summer: '抽1夏访客',
  vine: '抽1藤',
  order: '抽1订单',
  coin: '1金币',
  coins2: '2金币',
  structure: '抽1结构牌',
};
const groups = [
  ['coins', '3金币'],
  ['vp', '1胜利分'],
  ['cards', '2张手牌'],
  ['grape', '1颗葡萄'],
];
const colors = [
  ['vine', '葡萄藤'],
  ['order', '订单'],
  ['summer', '夏季访客'],
  ['winter', '冬季访客'],
  ['structure', '结构牌'],
];
// Pure projection of Go Trade / InfluenceMove; never uses an opponent's hand.
export function tuscanyInputState(kind, view, d) {
  const p = view.players.find((p) => p.id === view.youId),
    fields = [];
  let reason = '',
    target = '';
  const need = (key, label, items) => {
    fields.push({ key, label, items });
    if (!items.some(([id]) => id === d[key]) && !reason) {
      reason = '请选择：' + label;
      target = '[data-tuscany-input="' + key + '"]';
    }
  };
  if (!p) return { fields, reason: '未找到自己的玩家状态', payload: {}, target };
  if (kind === 'influence') {
    const regions = view.influenceRegions || [];
    const moving = Object.values(p.influence || {}).reduce((s, n) => s + n, 0) >= 6;
    const awards = !moving || (view.config?.structures && p.buildings?.includes('banquet_hall'));
    if (moving)
      need(
        'from',
        '移出星的区域',
        regions
          .filter((r) => p.influence?.[r.id] > 0)
          .map((r) => [r.id, r.name + ' · ' + p.influence[r.id] + '星']),
      );
    need(
      'to',
      moving
        ? awards
          ? '移入区域（宴会厅领奖）'
          : '移入区域（移动不领奖）'
        : '放星区域（领取即时奖励）',
      regions
        .filter((r) => !moving || r.id !== d.from)
        .map((r) => [
          r.id,
          r.name +
            ' · ' +
            (!awards ? '移动不领奖' : regionRewards[r.reward] || '奖励待核实') +
            ' · 终局' +
            r.points +
            '分',
        ]),
    );
    if (!regions.length) reason = '此主板的影响力区域尚未实现。';
    return {
      fields,
      reason,
      target,
      payload: {
        influence: [{ ...(moving ? { from: d.from } : {}), to: d.to }],
      },
    };
  }
  need('give', '付出资源', groups);
  need('receive', '获得资源', groups);
  const t = { give: d.give, receive: d.receive };
  if (d.give === 'coins' && p.coins < 3) {
    reason = '需要3金币。';
    target = '[data-tuscany-input="give"]';
  }
  if (d.give === 'vp' && p.vp <= -5) {
    reason = '胜利分不能低于−5。';
    target = '[data-tuscany-input="give"]';
  }
  if (d.give === 'cards') {
    const cards = (view.hand || []).map((c) => [c.id, localCard(c).name]);
    need('card1', '付出第1张手牌', cards);
    need(
      'card2',
      '付出第2张手牌',
      cards.filter(([id]) => id !== d.card1),
    );
    t.cardIds = [d.card1, d.card2];
  }
  if (d.give === 'grape') {
    need(
      'grape',
      '付出葡萄',
      (p.grapes || []).map((g) => [g.id, (g.color === 'red' ? '红' : '白') + g.value]),
    );
    t.grapeId = d.grape;
  }
  if (d.receive === 'cards') {
    need('color1', '抽第1张牌', colors);
    need('color2', '抽第2张牌', colors);
    t.colors = [d.color1, d.color2];
  }
  if (d.receive === 'grape') {
    need('color', '获得品质1葡萄', [
      ['red', '红葡萄'],
      ['white', '白葡萄'],
    ]);
    t.color = d.color;
  }
  return { fields, reason, target, payload: { trades: [t] } };
}
export function renderTuscanyInputs(host, state, d, online, change) {
  const controls = document.createElement('div');
  controls.className = 'tuscany-choice-inputs';
  for (const { key, label, items } of state.fields) {
    const wrap = document.createElement('label'),
      select = document.createElement('select');
    wrap.textContent = label;
    select.name = key;
    select.dataset.tuscanyInput = key;
    select.disabled = !online;
    select.required = true;
    for (const [value, text] of [['', '请选择'], ...items]) {
      const option = document.createElement('option');
      option.value = value;
      option.textContent = text;
      select.append(option);
    }
    select.value = d[key] || '';
    select.onchange = () => {
      d[key] = select.value;
      change(key);
    };
    wrap.append(select);
    controls.append(wrap);
  }
  host.append(controls);
}
