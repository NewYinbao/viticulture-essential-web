import { buildings } from './building-catalog.js';

const cardTypes = [
  ['vine', '葡萄藤'],
  ['order', '订单'],
  ['summer', '夏访客'],
  ['winter', '冬访客'],
  ['structure', '结构牌'],
];
const regions = [
  ['lucca', '卢卡'],
  ['pisa', '比萨'],
  ['firenze', '佛罗伦萨'],
  ['livorno', '利沃诺'],
  ['siena', '锡耶纳'],
  ['arezzo', '阿雷佐'],
  ['grosseto', '格罗塞托'],
];

// Every input maps to an explicit, persisted server operation. Button groups
// allow repeated draw colors and keep the estate visible while choosing.
export function renderRhineField(f, ctx) {
  if (!f.type.startsWith('rhine_')) return false;
  const { p, v, d, controls, readers, validators, rerender, node, chooser, online } = ctx;
  if (v.config?.visitors !== 'rhine') return false;
  if (f.type.startsWith('rhine_') && v.config?.visitors !== 'rhine') return false;
  const choose = (key, label, entries, parent = controls) => {
    const box = node('div', null, 'visitor-options');
    parent.append(node('h4', label), box);
    for (const [value, text, reason, description] of entries) {
      const button = node('button', text, d.values[key] === value ? 'primary' : '');
      button.type = 'button';
      button.disabled = !!reason || !online || d.busy;
      button.title = reason || description || text;
      button.setAttribute('aria-pressed', String(d.values[key] === value));
      button.onclick = () => {
        d.values[key] = value;
        rerender();
      };
      box.append(button);
    }
    validators.push(() => {
      if (!entries.some(([value, , reason]) => value === d.values[key] && !reason))
        throw Error('请选择' + label);
    });
    return () => d.values[key];
  };
  if (f.type === 'rhine_worker_loss') {
    const read = choose(
      f.name,
      '移除永久工人',
      (f.workers || []).map((w) => [
        w.id,
        w.label,
        '',
        '退回培训池，可重新培训；灰色临时工不可移除',
      ]),
    );
    readers.push((a) => {
      a.workerId = read();
    });
  } else if (f.type === 'rhine_draw') {
    const count = f.max;
    const picks = Array.from({ length: count }, (_, i) =>
      choose(
        f.name + i,
        `第 ${i + 1} 张`,
        cardTypes.filter(([id]) => id !== 'structure' || v.config?.structures),
      ),
    );
    readers.push((a) => {
      a.colors = picks.map((read) => read());
    });
  } else if (f.type === 'rhine_field') {
    const read = choose(
      f.name,
      '标记田地',
      p.fields
        .filter((x) => !x.sold && !x.structure)
        .map((x) => [x.index, `田地 ${x.index + 1} · 容量 ${x.capacity}`]),
    );
    readers.push((a) => {
      a.field = read();
    });
  } else if (f.type === 'rhine_yield') {
    const field = p.fields[p.rhineSonField];
    const red = (field?.vines || []).reduce((sum, c) => sum + c.red, 0),
      white = (field?.vines || []).reduce((sum, c) => sum + c.white, 0);
    const total = Math.min(3, red + white),
      entries = [];
    for (let r = 0; r <= total; r++)
      if (r <= red && total - r <= white)
        entries.push([`${r},${total - r}`, `红 ${r} · 白 ${total - r}`]);
    const read = choose(f.name, '收获葡萄', entries);
    readers.push((a) => {
      a.grapes = read().split(',').map(Number);
    });
  } else if (f.type === 'rhine_plan') {
    const seasons = ['spring', 'summer', 'fall', 'winter'];
    const spaces = v.spaces.filter(
      (s) =>
        s.season !== 'any' &&
        (v.config?.board === 'tuscany'
          ? seasons.indexOf(s.season) > seasons.indexOf(v.phase)
          : s.season === 'winter'),
    );
    const space = choose(
      f.name + 'Space',
      '预约行动',
      spaces.map((s) => [s.id, s.name]),
    );
    const target = spaces.find((s) => s.id === space());
    const slots = target
      ? Array.from({ length: target.capacity }, (_, i) => [
          i + 1,
          `第 ${i + 1} 格`,
          target.occupied?.some((x) => x.slot === i + 1) ? '已占用' : '',
        ])
      : [];
    if (target && slots.every((x) => x[2]) && v.triggerPlacement?.Seat?.large)
      slots.push([-1, '大工人额外位置']);
    const slot = choose(f.name + 'Slot', '预约位置', slots);
    readers.push((a) => {
      a.space = space();
      a.slot = slot();
    });
  } else if (f.type === 'rhine_purchase') {
    for (const [color, label, price] of [
      ['vine', '葡萄藤', 2],
      ['winter', '冬访客', 4],
    ]) {
      const key = f.name + color;
      d.values[key] ??= 0;
      const row = node('div', null, 'visitor-options');
      const minus = node('button', '−'),
        plus = node('button', '+');
      minus.type = plus.type = 'button';
      minus.disabled = d.values[key] <= 0;
      plus.disabled = d.values[key] >= Math.floor(p.coins / price);
      minus.onclick = () => {
        d.values[key] = Math.max(0, d.values[key] - 1);
        rerender();
      };
      plus.onclick = () => {
        d.values[key]++;
        rerender();
      };
      row.append(minus, node('span', `${label} ${d.values[key]} 张 · 每张 ${price} 金币`), plus);
      controls.append(row);
    }
    validators.push(() => {
      if ((d.values[f.name + 'vine'] || 0) + (d.values[f.name + 'winter'] || 0) < (f.min || 0))
        throw Error('本次尚未出售葡萄藤，至少购买1张牌');
      if ((d.values[f.name + 'vine'] || 0) * 2 + (d.values[f.name + 'winter'] || 0) * 4 > p.coins)
        throw Error('购买总价超过金币');
    });
    readers.push((a) => {
      a.colors = [
        ...Array(d.values[f.name + 'vine'] || 0).fill('vine'),
        ...Array(d.values[f.name + 'winter'] || 0).fill('winter'),
      ];
    });
  } else if (f.type === 'rhine_value') {
    const read = choose(
      f.name,
      '价值 / 花费',
      Array.from({ length: 9 }, (_, i) => [
        i + 1,
        String(i + 1),
        i + 1 > p.coins ? '金币不足' : '',
      ]),
    );
    readers.push((a) => {
      a.slot = read();
    });
  } else if (f.type === 'rhine_wine') {
    const value = f.valueFromBuilding
      ? buildings.find((b) => b[0] === d.values.building)?.[2]
      : f.valueFromPurchase
        ? d.values.rhineValue
        : f.value;
    const limit = p.buildings.includes('large_cellar')
      ? 9
      : p.buildings.includes('medium_cellar')
        ? 6
        : 3;
    const entries = [
      ['red', '红酒', 1],
      ['white', '白酒', 1],
      ['blush', '桃红酒', 4],
      ['sparkling', '起泡酒', 7],
    ].map(([type, name, minimum]) => {
      let actual = Math.min(value || 0, limit, 9);
      while (actual >= minimum && p.wines.some((w) => w.type === type && w.value === actual))
        actual--;
      const reason = !value
        ? '先选择价值来源'
        : value < minimum
          ? `最低价值 ${minimum}`
          : limit < minimum
            ? minimum === 7
              ? '需要大酒窖'
              : '需要中酒窖'
            : actual < minimum
              ? '没有空酒槽'
              : '';
      return [type, reason ? name : `${name} · 价值 ${actual}`, reason];
    });
    const read = choose(f.name, '获得酒', entries);
    readers.push((a) => {
      a.color = read();
    });
  } else if (f.type === 'rhine_build') {
    const available = buildings.filter(
      ([id]) =>
        !p.buildings.includes(id) &&
        (buildings.findIndex((b) => b[0] === id) < 8 ||
          (v.config?.structures && v.hand?.some((c) => c.structureId === id))),
    );
    const read = choose(
      f.name,
      '建造建筑',
      available.map(([id, name, value]) => {
        const cost = Math.max(
          0,
          value - (f.discount || 0) - (p.buildings.includes('workshop') ? 1 : 0),
        );
        const reason =
          id === 'large_cellar' && !p.buildings.includes('medium_cellar')
            ? '需要中酒窖'
            : cost > p.coins
              ? `还差 ${cost - p.coins} 金币`
              : '';
        return [id, `${name} · ${cost} 金币`, reason, buildings.find((b) => b[0] === id)?.[4]];
      }),
    );
    readers.push((a) => {
      a.building = read();
    });
  } else if (f.type === 'rhine_destroy') {
    const read = choose(
      f.name,
      '拆除建筑',
      p.buildings.map((id) => {
        const b = buildings.find((b) => b[0] === id);
        return [
          id,
          `${b?.[1] || id} · 标价 ${b?.[2] ?? '?'} 金币`,
          id === 'medium_cellar' && p.buildings.includes('large_cellar') ? '须先拆大酒窖' : '',
        ];
      }),
    );
    readers.push((a) => {
      a.building = read();
    });
  } else if (f.type === 'rhine_influence') {
    const total = Object.values(p.influence || {}).reduce((a, b) => a + b, 0);
    const from =
      total >= 6
        ? choose(
            f.name + 'From',
            '移出区域',
            regions.filter(([id]) => p.influence?.[id] > 0),
          )
        : () => '';
    const to = choose(f.name + 'To', total >= 6 ? '移入区域' : '放置区域', regions);
    validators.push(() => {
      if (from() === to()) throw Error('须移动到不同区域');
    });
    readers.push((a) => {
      a.influence = [{ from: from(), to: to() }];
    });
  } else if (f.type === 'rhine_stars') {
    const tokens = [];
    for (const owner of v.players) {
      if (f.own && owner.id !== p.id) continue;
      for (const [from, count] of Object.entries(owner.influence || {})) {
        for (let i = 0; i < count; i++)
          tokens.push([
            `${owner.id}|${from}|${i}`,
            `${owner.name} · ${regions.find((x) => x[0] === from)?.[1] || from} · ${i + 1}`,
          ]);
      }
    }
    const selected = chooser(
      f.name,
      f.operation === 'move_stars' ? '移动星星' : '收回星星',
      tokens,
      true,
      f.min,
      f.max,
    );
    const chosen = selected();
    const destinations = new Map();
    if (f.operation === 'move_stars')
      for (const key of chosen)
        destinations.set(key, choose('to:' + key, tokens.find((x) => x[0] === key)?.[1], regions));
    readers.push((a) => {
      a.rhineStars = selected().map((key) => {
        const [playerId, from] = key.split('|');
        const to = destinations.get(key)?.() || '';
        if (f.operation === 'move_stars' && (!to || to === from))
          throw Error('为每颗星选择不同的目标区域');
        return { playerId, from, to };
      });
    });
    // The ordinary chooser updates selection in place. A compact button opens
    // destination cards only when needed, preserving selected source tokens.
    if (f.operation === 'move_stars') {
      const refresh = node('button', '选择移动目标');
      refresh.type = 'button';
      refresh.onclick = rerender;
      controls.append(refresh);
    }
  } else return false;
  return true;
}
