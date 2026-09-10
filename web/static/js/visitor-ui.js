import { cardInfo, localCard } from './card-i18n.js';
const node = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
const names = {
  coins: '获得金币',
  vp: '胜利分',
  buy: '支付费用换取收益',
  sell: '交换资源换取收益',
  draw: '抽牌',
  discard: '弃置换取收益',
  build: '建造',
  small: '失1分建小型建筑',
  any: '失2分建任意建筑',
  harvest: '收获',
  make: '酿酒',
  plant: '种藤',
  uproot: '拔除并弃藤',
  exchange: '弃牌换牌',
  grape: '弃葡萄',
  wine: '弃酒',
  both: '支付额外费用，两项都做',
  pay: '付款',
  resolve: '开始结算',
  select: '选择目标',
  play: '再打出一张访客',
  skip: '不执行 / 跳过',
  fill: '交付订单',
  age: '陈酿',
  upgrade: '升级酒窖',
  train: '培训',
  draw_upgrade: '抽订单＋升级酒窖',
  draw_vp: '抽订单＋得分',
  upgrade_vp: '升级酒窖＋得分',
  retrieve: '收回工人',
  cards: '交出手牌',
  give: '交出卡牌',
  summer: '夏季访客',
  winter: '冬季访客',
  vine: '葡萄藤',
  order: '订单',
  red: '红葡萄',
  white: '白葡萄',
  blush: '桃红酒',
  sparkling: '起泡酒',
  summer_summer: '夏季＋夏季',
  summer_winter: '夏季＋冬季',
  winter_winter: '冬季＋冬季',
  draw_train: '先抽订单 → 再培训',
  train_draw: '先培训 → 再抽订单',
  upgrade_draw: '先升级酒窖 → 再抽订单',
  vp_draw: '先得分 → 再抽订单',
  vp_upgrade: '先得分 → 再升级酒窖',
};
Object.assign(names, {
  plan: '预约冬季行动',
  move: '迁移起床行并结束本季',
  execute: '执行冬初预约',
  swap: '交换不同田地的两张藤',
  take: '选取卡牌',
  reveal: '公开翻牌',
  action: '执行夏季行动',
  continuation: '继续后续步骤',
  vp_build: '先得分 → 再建造',
  build_vp: '先建造 → 再得分',
  vp_plant: '先得分 → 再种藤',
  plant_vp: '先种藤 → 再得分',
  build_plant: '先建造 → 再种藤',
  plant_build: '先种藤 → 再建造',
  harvest_make: '先收获 → 再酿酒',
  make_harvest: '先酿酒 → 再收获',
  harvest_fill: '先收获 → 再交单',
  fill_harvest: '先交单 → 再收获',
  make_fill: '先酿酒 → 再交单',
  fill_make: '先交单 → 再酿酒',
});
const buildings = {
  trellis: ['棚架', 2],
  irrigation: ['灌溉塔', 3],
  yoke: ['轭', 2],
  medium_cellar: ['中酒窖', 4],
  large_cellar: ['大酒窖', 6],
  cottage: ['小屋', 4],
  windmill: ['磨坊', 5],
  tasting_room: ['品酒室', 6],
};
const field = (name, type, min = 1, max = 1, filter = '') => ({ name, type, min, max, filter });
const card = (n, filter = '') => field('cardIds', 'cards', n, n, filter);
const wine = (minValue = 1) => ({ ...field('wineIds', 'wines'), minValue });
const plant = (max = 1) => field('plant', 'plant', 1, max);
const harvest = (max = 1) => field('fields', 'fields', 1, max);
const make = (max = 2) => field('recipes', 'recipes', 1, max);
const build = () => field('building', 'building');
const grape = () => field('grapes', 'grapes');
const fill = () => [field('cardId', 'cards', 1, 1, 'order'), field('wineIds', 'wines', 1, 3)];
let draft = null;
// Compatibility mapping for the concrete Go options. Never infer a cost from a label.
export function choiceFields(c, option) {
  if (c.kind === 'planner') return [];
  const id = c.visitor?.cardId,
    stage = c.visitor?.stage;
  if (c.optionFields?.[option]) return c.optionFields[option];
  let out = [];
  if (stage === 'second' && option === 'play') out = [field('cardId', 'cards', 1, 1, 'season')];
  else if (stage === 'reward' || option === 'skip') out = [];
  else if (stage === 'reply') {
    if (id === 'summer-18' && option === 'build') out = [build()];
    if (id === 'summer-37' && option === 'plant') out = [plant()];
    if (id === 'winter-11' && option === 'cards') out = [card(2)];
    if (id === 'winter-17' && option === 'make') out = [{ ...make(), min: 1 }];
    if (id === 'winter-35' && option === 'give') out = [card(1, 'summer')];
    if (id === 'summer-23' && option === 'give')
      out = [field('cardIds', 'cards', 1, 3, 'visitors')];
  } else {
    if (option === 'plant')
      out = [plant(['summer-09', 'summer-28', 'summer-35'].includes(id) ? 2 : 1)];
    if (option === 'uproot')
      out = [field('uproot', 'uproot', id === 'summer-19' ? 2 : 1, id === 'summer-19' ? 2 : 1)];
    if (option === 'build' && id !== 'winter-18')
      out = [id === 'summer-36' ? field('buildings', 'buildings', 2, 2) : build()];
    if (id === 'summer-13') out = [{ ...build(), small: option === 'small' }];
    if (id === 'summer-28' && ['both', 'build_plant'].includes(option)) out = [build()];
    if (id === 'summer-28' && option === 'plant_build') out = [plant(2)];
    if (option === 'harvest')
      out = ['summer-06', 'winter-18'].includes(id)
        ? [field('field', 'fields')]
        : [harvest(id === 'winter-24' ? 3 : 1)];
    if (id === 'winter-12') out = [harvest(2)];
    if (option === 'make')
      out = [make(['winter-06', 'winter-26', 'winter-29'].includes(id) ? 3 : 2)];
    if (id === 'winter-02' && stage === 'effect' && option === 'make') out = [];
    if (option === 'fill') out = id === 'winter-14' && stage !== 'fill' ? [wine()] : fill();
    if (option === 'grape') out = [grape()];
    if (option === 'wine') out = [wine()];
    if (option === 'discard') {
      if (['summer-03', 'winter-03', 'winter-31', 'winter-36'].includes(id))
        out = [wine(id === 'summer-03' ? 7 : id === 'winter-03' ? 4 : 1)];
      if (['summer-10', 'winter-28'].includes(id)) out = [grape()];
      if (id === 'summer-16') out = [wine(), card(3, 'visitors')];
      if (id === 'winter-09') out = [card(2, 'visitors')];
    }
    if (id === 'summer-10' && option === 'buy') out = [field('color', 'color')];
    if (id === 'summer-15') out = [card(option === 'coins' ? 2 : 4)];
    if (id === 'summer-20') out = [card(2)];
    if (id === 'summer-38' || id === 'winter-35')
      out = [field('targetIds', 'targets', id === 'summer-38' ? 1 : 0, 3)];
  }
  if (['summer-05', 'summer-22', 'winter-20', 'winter-29'].includes(id)) {
    const first = option.split('_')[0];
    out =
      first === 'build'
        ? [build()]
        : first === 'plant'
          ? [plant()]
          : first === 'harvest'
            ? [{ ...harvest(id === 'winter-29' ? 2 : 1), min: 1 }]
            : first === 'make'
              ? [make(id === 'winter-29' ? 3 : 2)]
              : first === 'fill'
                ? fill()
                : [];
  }
  if (id === 'summer-29')
    out = [field('space', 'space'), field('slot', 'placementSlot'), field('large', 'worker')];
  if (id === 'summer-33')
    out = [field('slot', 'wake'), { ...field('color', 'color'), options: ['summer', 'winter'] }];
  if (id === 'summer-11' && option === 'swap') out = [field('swap', 'swap', 2, 2)];
  if (id === 'summer-11' && stage === 'plant') out = [{ ...plant(), min: 0 }];
  if (id === 'summer-32') out = [field('seats', 'seats', 1, 2)];
  if (id === 'winter-19')
    out = [{ ...field('colors', 'colors', 2, 2), options: ['vine', 'summer', 'order', 'winter'] }];
  if (id === 'winter-37' && stage === 'take') out = [field('revealed', 'revealed', 2, 2)];
  if (id === 'winter-32') out = [field('manager', 'manager')];
  // Older server emits broad schema for all branches; only a field-specific server schema
  // may replace a mapped field. New optionFields is preferred above.
  const schema = c.schema || c.fields || [];
  if (!out.length && schema.length) return schema;
  return out.map((f) => ({ ...f, ...schema.find((s) => s.name === f.name) }));
}
export function buildingCost(c, price) {
  const id = c.visitor?.cardId;
  return Math.max(
    0,
    price -
      ({
        'summer-04': 2,
        'summer-12': 3,
        'summer-13': 99,
        'summer-18': 2,
        'summer-28': 3,
        'summer-35': 1,
        'summer-36': 99,
      }[id] || 0),
  );
}
export function visitorCost(c, option, p) {
  if (c.kind === 'planner')
    return '冬初执行原预约格行动及格奖励，不再消耗工人；培训原价 4 金币（奖励格减 1），其他行动按通常规则结算。';
  const id = c.visitor?.cardId,
    stage = c.visitor?.stage,
    first = option === 'both' ? (id === 'summer-28' ? 'build' : 'draw') : option.split('_')[0],
    notes = [];
  if (
    stage === 'effect' &&
    ((['summer-28', 'winter-23'].includes(id) &&
      ['both', 'build_plant', 'plant_build', 'draw_train', 'train_draw'].includes(option)) ||
      (id === 'winter-29' && option.includes('_')) ||
      (id === 'summer-34' && option === 'both'))
  )
    notes.push('本步骤先失去 1 胜利分；后续步骤不重复扣分');
  if (id === 'summer-29')
    notes.push(
      '现在额外消耗 1 名可用工人并占住冬季格；现在不执行、不收费、不拿奖励。冬初再选择酿酒、订单等参数并支付行动费用；不可执行时耗尽预约且不退工人',
    );
  if (id === 'summer-33')
    notes.push(
      '免费移到空起床行：1 无奖励；2 抽1藤；3 抽1订单；4 得1金币；5 抽1张所选夏/冬访客；6 得1胜利分；7 仅当唯一灰工人尚未领取时获得本年临时工人。本次外层剩余访客结算后结束本季',
    );
  if (id === 'winter-11')
    notes.push(
      stage === 'reply'
        ? {
            vp: '你失去 1 胜利分（不能低于 −5）；出牌者不获得此分',
            cards: '你交给出牌者恰好 2 张不同手牌（任意类型）',
            coins: '你支付 3 金币给出牌者',
          }[option] || '必须完成一个合法分支，不能跳过'
        : '固定座位右邻座必须选择：失去 1 胜利分、给你 2 张任意手牌，或付你 3 金币；不是按起床顺序',
    );
  if (id === 'summer-36') notes.push('本次合计支付 8 金币，所选两座建筑不另收费');
  if (id === 'summer-13')
    notes.push('本次失去 ' + (option === 'small' ? 1 : 2) + ' 胜利分，建筑免费');
  if (id === 'winter-23' && first === 'train') notes.push('本步骤支付 3 金币培训');
  if (id === 'winter-23' && ['both', 'draw_train'].includes(option))
    notes.push('后续培训另需 3 金币');
  if (id === 'winter-27' && option.includes('upgrade'))
    notes.push(
      (first === 'upgrade' ? '本步骤' : '后续步骤') +
        '原价升级酒窖：' +
        (p.buildings.includes('medium_cellar') ? 6 : 4) +
        ' 金币',
    );
  return notes.join('。');
}
function compatibleFilter(card, filter, v) {
  return (
    !filter ||
    filter === card.type ||
    (filter === 'visitors' && ['summer', 'winter'].includes(card.type)) ||
    (filter === 'season' && card.type === (v.phase === 'winter' ? 'winter' : 'summer'))
  );
}
function normalize(f) {
  const aliases = {
    cardId: 'cards',
    cardIds: 'cards',
    wineIds: 'wines',
    grapes: 'grapes',
    fields: 'fields',
    field: 'fields',
    building: 'building',
    buildings: 'buildings',
    targetIds: 'targets',
    recipes: 'recipes',
    color: 'color',
    colors: 'colors',
    slot: 'wake',
    space: 'space',
    large: 'worker',
  };
  return {
    ...f,
    type: f.type === 'placementSlot' ? f.type : aliases[f.name] || f.type,
    min: f.min ?? 1,
    max: f.max ?? 1,
  };
}
export function renderVisitor(box, v, act, online) {
  const c = v.pendingChoice;
  if (!c?.visitor && !['visitor', 'planner'].includes(c?.kind)) {
    draft = null;
    return false;
  }
  box.replaceChildren();
  box.dataset.choiceId = c.id;
  box.dataset.revision = v.revision;
  if (c.playerId !== v.youId) {
    box.append(
      node(
        'p',
        '等待 ' +
          (v.players.find((p) => p.id === c.playerId)?.name || '玩家') +
          ' 响应；只有该玩家能查看及操作自己的手牌。',
      ),
    );
    return true;
  }
  const key = JSON.stringify([v.code, v.youId, c.id]);
  if (draft?.key !== key)
    draft = { key, option: c.options?.[0] || '', values: {}, busy: false, error: '' };
  const d = draft;
  d.view = v;
  d.online = online;
  d.box = box;
  if (!(c.options || []).includes(d.option)) d.option = c.options?.[0] || '';
  const p = v.players.find((p) => p.id === v.youId),
    info = cardInfo(c.visitor?.cardId),
    title = node('h3', info?.name || '访客后续行动');
  box.append(title);
  if (info)
    box.append(
      node('small', info.englishName + ' · ' + c.visitor.cardId, 'card-english'),
      node('p', info.description, 'visitor-rules'),
    );
  box.append(
    node(
      'p',
      '当前响应：' +
        p.name +
        ' · ' +
        ({
          effect: '选择效果',
          reply: '多人响应',
          reward: '选择奖励',
          second: '第二张访客',
          plant: '后续种藤',
          make: '后续酿酒',
          fill: '后续交付',
        }[c.visitor?.stage] ||
          c.visitor?.stage ||
          '选择') +
        ' · 队列剩余 ' +
        (v.pendingCount || 1),
      'visitor-stage',
    ),
  );
  const cost = visitorCost(c, d.option, p);
  if (cost) box.append(node('p', cost, 'visitor-cost'));
  const options = node('div', null, 'visitor-options');
  options.setAttribute('role', 'group');
  options.setAttribute('aria-label', '访客效果选项');
  box.append(options);
  const rerender = () => {
    if (draft === d) renderVisitor(d.box, d.view, act, d.online);
  };
  for (const option of c.options || []) {
    const b = node(
      'button',
      c.labels?.[option] ||
        (c.visitor?.cardId === 'winter-11' && c.visitor?.stage === 'reply'
          ? { vp: '失去 1 胜利分', cards: '给出牌者 2 张手牌', coins: '付出牌者 3 金币' }[option]
          : null) ||
        names[option] ||
        option,
      d.option === option ? 'primary' : '',
    );
    b.type = 'button';
    b.dataset.visitorOption = option;
    b.setAttribute('aria-pressed', String(d.option === option));
    b.disabled = !online || d.busy;
    b.onclick = () => {
      d.option = option;
      d.values = {};
      d.error = '';
      rerender();
    };
    options.append(b);
  }
  const form = node('form', null, 'visitor-form'),
    controls = node('fieldset');
  controls.disabled = !online || d.busy;
  form.append(controls);
  box.append(form);
  const readers = [],
    validators = [],
    fields = choiceFields(c, d.option).map(normalize);
  let unsupported = false;
  if (c.kind === 'planner') {
    const plan = v.planned?.[0],
      sp = (v.spaces || []).find((x) => x.id === plan?.space);
    controls.append(node('h4', '冬初预约：' + (sp?.name || '预约信息缺失')));
    if (!sp || plan.playerId !== v.youId) {
      unsupported = true;
      controls.append(
        node(
          'p',
          '服务器未提供有效的公开预约 planned；无法安全推断行动，请更新后端。',
          'ee-warning',
        ),
      );
    }
    const bonus = plan?.slot === 1 && sp?.capacity >= 2;
    const label = node('label', '放弃行动格奖励（仍执行行动）'),
      decline = node('input');
    decline.type = 'checkbox';
    decline.name = 'declineBonus';
    decline.checked = !!d.values.declineBonus;
    decline.onchange = () => {
      d.values.declineBonus = decline.checked;
    };
    label.prepend(decline);
    controls.append(label);
    readers.push((a) => (a.declineBonus = decline.checked));
    readers.push((a) => (a.space = plan.space));
    switch (plan?.space) {
      case 'harvest':
        fields.push({ ...harvest(bonus ? 2 : 1), min: 1 });
        break;
      case 'make_wine':
        fields.push({ ...make(bonus ? 3 : 2), min: 1 });
        break;
      case 'fill_order':
        fields.push(...fill());
        break;
      case 'winter_visitor':
        fields.push(field('cardId', 'cards', 1, 1, 'winter'));
        break;
      case 'draw_order':
      case 'train':
        break;
      default:
        unsupported = true;
    }
  }
  if (fields.some((f) => f.type === 'manager')) {
    const summer = (v.spaces || []).filter((x) => x.season === 'summer');
    if (!summer.some((x) => x.id === d.values.space)) d.values.space = summer[0]?.id || '';
    const row = node('div', null, 'visitor-options');
    controls.append(node('h4', '不放工人执行夏季行动（无行动格奖励）'), row);
    for (const sp of summer) {
      const b = node('button', sp.name, d.values.space === sp.id ? 'primary' : '');
      b.type = 'button';
      b.onclick = () => {
        d.values = { space: sp.id };
        rerender();
      };
      row.append(b);
    }
    readers.push((a) => (a.space = d.values.space));
    fields.splice(0, fields.length);
    switch (d.values.space) {
      case 'build':
        fields.push(build());
        break;
      case 'plant':
        fields.push(plant());
        break;
      case 'summer_visitor':
        fields.push(field('cardId', 'cards', 1, 1, 'summer'));
        break;
      case 'sell_grapes':
        {
          const modes = [
            ['sell_grapes', '出售葡萄'],
            ['sell_field', '出售空田'],
            ['buy_field', '买回田地'],
          ];
          d.values.mode ??= 'sell_grapes';
          const modeRow = node('div', null, 'visitor-options');
          controls.append(modeRow);
          for (const [id, label] of modes) {
            const b = node('button', label, d.values.mode === id ? 'primary' : '');
            b.type = 'button';
            b.onclick = () => {
              d.values = { space: d.values.space, mode: id };
              rerender();
            };
            modeRow.append(b);
          }
          readers.push((a) => (a.mode = d.values.mode));
          fields.push(
            d.values.mode === 'sell_grapes'
              ? field('grapes', 'grapes', 1, p.grapes.length)
              : field('tradeField', 'tradeField'),
          );
        }
        break;
    }
  }

  // Clear only visitor bindings; annual discard has its own independent lifecycle.
  for (const e of document.querySelectorAll('#hand .card')) {
    e.classList.remove('visitor-selected', 'visitor-eligible');
    if (e.dataset.visitorBound) {
      e.onclick = null;
      e.onkeydown = null;
      e.removeAttribute('role');
      e.removeAttribute('aria-pressed');
      e.tabIndex = -1;
      delete e.dataset.visitorBound;
    }
  }
  const put = (name, value) => {
    d.values[name] = value;
  };
  function chooser(name, label, items, multi = false, min = 1, max = 1) {
    const section = node('section', null, 'visitor-resource'),
      heading = node('h4', label),
      status = node('small');
    section.append(heading, status);
    controls.append(section);
    const exists = new Set(items.map((i) => String(i[0])));
    let value = d.values[name];
    if (multi) {
      value = (Array.isArray(value) ? value : []).filter((x) => exists.has(String(x)));
    } else if (!exists.has(String(value))) value = items[0]?.[0] ?? '';
    put(name, value);
    const update = () => {
      status.textContent = multi
        ? '已选 ' + d.values[name].length + ' · 要求 ' + min + '–' + max
        : '请选择一项';
      for (const b of section.querySelectorAll('button'))
        b.setAttribute(
          'aria-pressed',
          String(
            multi
              ? d.values[name].includes(b.dataset.value)
              : String(d.values[name]) === b.dataset.value,
          ),
        );
    };
    for (const [val, text] of items) {
      const b = node('button', text, 'resource-chip');
      b.type = 'button';
      b.dataset.value = String(val);
      b.onclick = () => {
        if (multi) {
          const a = d.values[name];
          put(
            name,
            a.includes(String(val)) ? a.filter((x) => x !== String(val)) : [...a, String(val)],
          );
        } else put(name, String(val));
        update();
      };
      section.append(b);
    }
    if (!items.length) section.append(node('p', '暂无可用资源', 'ee-warning'));
    validators.push(() => {
      const size = multi ? d.values[name].length : d.values[name] !== '' ? 1 : 0;
      if (size < min || size > max) throw Error(label + '：请选择 ' + min + '–' + max + ' 项');
    });
    update();
    return () => d.values[name];
  }
  function select(name, label, items, parent = controls) {
    const l = node('label', label),
      s = node('select');
    s.name = name;
    for (const [val, text] of items) {
      const o = node('option', text);
      o.value = String(val);
      s.append(o);
    }
    if (d.values[name] != null) s.value = String(d.values[name]);
    if (s.selectedIndex < 0 && items.length) s.selectedIndex = 0;
    put(name, s.value);
    s.onchange = () => put(name, s.value);
    l.append(s);
    parent.append(l);
    return () => d.values[name];
  }
  function handSelect(f, onChange = () => {}) {
    const eligible = (v.hand || []).filter(
      (x) =>
        compatibleFilter(x, f.filter, v) &&
        (!(f.name === 'cardId' && ['season', 'summer', 'winter'].includes(f.filter)) ||
          x.implemented),
    );
    const valid = new Set(eligible.map((x) => x.id));
    const old = Array.isArray(d.values[f.name])
      ? d.values[f.name]
      : d.values[f.name]
        ? [d.values[f.name]]
        : [];
    put(
      f.name,
      old.filter((x) => valid.has(x)),
    );
    const hint = node('p', null, 'visitor-hand-hint');
    hint.setAttribute('role', 'status');
    controls.append(hint);
    const update = () => {
      hint.textContent =
        '在下方原手牌区点选' +
        (f.filter === 'order' ? '订单' : f.filter === 'vine' ? '葡萄藤' : '卡牌') +
        '：已选 ' +
        d.values[f.name].length +
        ' / ' +
        f.min +
        '–' +
        f.max +
        ' 张';
      for (const e of document.querySelectorAll('#hand .card')) {
        if (!valid.has(e.dataset.cardId)) continue;
        e.classList.add('visitor-eligible');
        const yes = d.values[f.name].includes(e.dataset.cardId);
        e.classList.toggle('visitor-selected', yes);
        e.setAttribute('aria-pressed', String(yes));
      }
    };
    for (const e of document.querySelectorAll('#hand .card')) {
      if (!valid.has(e.dataset.cardId)) continue;
      e.tabIndex = 0;
      e.dataset.visitorBound = 'true';
      e.setAttribute('role', 'button');
      e.setAttribute('aria-disabled', String(!online || d.busy));
      const toggle = () => {
        if (d.busy || !online) return;
        const a = d.values[f.name],
          id = e.dataset.cardId;
        put(f.name, a.includes(id) ? a.filter((x) => x !== id) : f.max === 1 ? [id] : [...a, id]);
        update();
        onChange();
      };
      e.onclick = toggle;
      e.onkeydown = (event) => {
        if (event.key === ' ' || event.key === 'Enter') {
          event.preventDefault();
          if (!event.repeat) toggle();
        }
      };
    }
    if (!eligible.length) controls.append(node('p', '没有符合要求的手牌', 'ee-warning'));
    validators.push(() => {
      const len = d.values[f.name].length;
      if (len < f.min || len > f.max) throw Error('手牌须选 ' + f.min + '–' + f.max + ' 张');
    });
    update();
    return () => d.values[f.name];
  }
  for (const f of fields) {
    const multi = !['field', 'cardId', 'building'].includes(f.name);
    if (f.type === 'cards') {
      const read = handSelect(f);
      readers.push((a) => (a[f.name] = f.name === 'cardId' ? read()[0] : read()));
    } else if (f.type === 'plant') {
      const destinations = node('div', null, 'visitor-plant-fields');
      let maps = {};
      const updateFields = () => {
        destinations.replaceChildren();
        maps = {};
        for (const id of read())
          maps[id] = select(
            'plant-' + id,
            '将 ' + localCard(v.hand.find((x) => x.id === id)).name + ' 种入',
            p.fields
              .filter((x) => !x.sold)
              .map((x) => [x.index, '田地 ' + (x.index + 1) + ' · 容量 ' + x.capacity]),
            destinations,
          );
      };
      const read = handSelect({ ...f, name: 'cardIds', filter: 'vine' }, updateFields);
      controls.append(destinations);
      updateFields();
      readers.push((a) => {
        a.cardIds = read();
        a.fields = read().map((id) => {
          const value = maps[id]();
          if (value === '') throw Error('没有可种植田地');
          return Number(value);
        });
      });
    } else if (f.type === 'uproot' || f.type === 'swap') {
      const items = p.fields.flatMap((x) =>
        (x.vines || []).map((card) => [
          card.id,
          '田地 ' + (x.index + 1) + ' · ' + localCard(card).name,
        ]),
      );
      const read = chooser(
        'uproot',
        f.type === 'swap' ? '选择不同田地的两藤交换' : '选择田上葡萄藤（拔除并弃置）',
        items,
        true,
        f.min,
        f.max,
      );
      readers.push((a) => {
        a.cardIds = read();
        a.fields = read().map((id) => p.fields.find((x) => x.vines.some((y) => y.id === id)).index);
      });
    } else if (f.type === 'revealed') {
      const cards = v.revealed || c.cards || [];
      const read = chooser(
        'revealed',
        '选择本次公开牌（其余弃置）',
        cards.map((x) => {
          const card = localCard(x);
          return [x.id, card.name + ' / ' + card.englishName + ' · ' + card.description];
        }),
        true,
        f.min,
        f.max,
      );
      readers.push((a) => (a.cardIds = read()));
    } else if (f.type === 'seats') {
      const seats = (v.spaces || []).flatMap((sp) =>
        (sp.occupied || [])
          .filter(
            (x) =>
              x.playerId === v.youId &&
              !(
                sp.id === v.triggerPlacement?.Space &&
                x.slot === v.triggerPlacement.Seat.slot &&
                x.large === v.triggerPlacement.Seat.large
              ),
          )
          .map((x) => ({
            id: sp.id + ':' + x.slot,
            space: sp.id,
            slot: x.slot,
            label: sp.name + ' · ' + (x.large ? '大工人' : '普通工人') + ' · 格 ' + x.slot,
          })),
      );
      const read = chooser(
        'seats',
        '收回自己的工人（不含触发此牌者）',
        seats.map((x) => [x.id, x.label]),
        true,
        1,
        2,
      );
      readers.push((a) => {
        const chosen = read().map((id) => seats.find((x) => x.id === id));
        a.targetIds = chosen.map((x) => x.space);
        a.fields = chosen.map((x) => x.slot);
      });
    } else if (f.type === 'tradeField') {
      const read = chooser(
        'tradeField',
        '选择交易田地',
        p.fields
          .filter((x) => (d.values.mode === 'buy_field' ? x.sold : !x.sold && !x.vines.length))
          .map((x) => [x.index, '田地 ' + (x.index + 1) + ' · ' + x.capacity + ' 金币']),
      );
      readers.push((a) => (a.field = Number(read())));
    } else if (f.type === 'fields') {
      const read = chooser(
        f.name,
        '选择可收获田地',
        p.fields
          .filter((x) => !x.sold && !x.harvested && x.vines?.length)
          .map((x) => [x.index, '田地 ' + (x.index + 1) + ' · ' + x.vines.length + ' 藤']),
        multi,
        f.min,
        f.max,
      );
      readers.push((a) => (a[f.name] = multi ? read().map(Number) : Number(read())));
    } else if (f.type === 'wines') {
      const read = chooser(
        f.name,
        '选择葡萄酒',
        p.wines
          .filter((x) => x.value >= (f.minValue || 1))
          .map((x) => [x.id, (names[x.type] || x.type) + ' · 品质 ' + x.value]),
        true,
        f.min,
        f.max,
      );
      readers.push((a) => (a[f.name] = read()));
    } else if (f.type === 'grapes') {
      const read = chooser(
        f.name,
        '选择葡萄',
        p.grapes.map((x, i) => [i, (names[x.color] || x.color) + ' · 品质 ' + x.value]),
        true,
        f.min,
        f.max,
      );
      readers.push((a) => (a[f.name] = read().map(Number)));
    } else if (['building', 'buildings'].includes(f.type)) {
      const isMulti = f.type === 'buildings';
      const read = chooser(
        f.name,
        isMulti ? '选择两座建筑（中酒窖先于大酒窖）' : '选择未建建筑',
        Object.entries(buildings)
          .filter(
            ([id, b]) =>
              !p.buildings.includes(id) &&
              (!f.small || b[1] <= 3) &&
              (isMulti || id !== 'large_cellar' || p.buildings.includes('medium_cellar')),
          )
          .map(([id, b]) => [
            id,
            b[0] + ' · 原价 ' + b[1] + ' / 本次 ' + buildingCost(c, b[1]) + ' 金币',
          ]),
        isMulti,
        f.min,
        f.max,
      );
      readers.push(
        (a) =>
          (a[f.name] = isMulti
            ? [...read()].sort((x, y) =>
                x === 'medium_cellar' ? -1 : y === 'medium_cellar' ? 1 : 0,
              )
            : read()),
      );
    } else if (f.type === 'targets') {
      const read = chooser(
        f.name,
        '选择不同对手',
        v.players
          .filter((x) => x.id !== v.youId)
          .map((x) => [x.id, x.name + ' · ' + x.coins + ' 金币']),
        true,
        f.min,
        f.max,
      );
      readers.push((a) => (a[f.name] = read()));
    } else if (f.type === 'recipes') {
      controls.append(
        node('h4', '为每颗葡萄选择酒瓶配方'),
        node(
          'p',
          '红/白：单颗；桃红：一红一白且总值≥4；起泡：两红一白且总值≥7。酒窖与空槽由服务器检查。',
        ),
      );
      const picks = p.grapes.map((g, i) =>
        select('recipe-' + (g.id || i), (names[g.color] || g.color) + '葡萄 · 品质 ' + g.value, [
          ['', '保留'],
          ...Array.from({ length: f.max }, (_, j) => [String(j + 1), '第 ' + (j + 1) + ' 瓶']),
        ]),
      );
      readers.push((a) => {
        const groups = {};
        picks.forEach((read, i) => {
          const val = read();
          if (val) (groups[val] ??= []).push(i);
        });
        a.recipes = Object.values(groups);
        if (a.recipes.length < f.min || a.recipes.length > f.max) throw Error('酿酒瓶数不符合要求');
        for (const r of a.recipes) {
          const red = r.filter((i) => p.grapes[i].color === 'red').length,
            white = r.length - red,
            sum = r.reduce((s, i) => s + p.grapes[i].value, 0);
          if (!(
            r.length === 1 ||
            (red === 1 && white === 1 && sum >= 4) ||
            (red === 2 && white === 1 && sum >= 7)
          ))
            throw Error('配方不合法：检查葡萄颜色、颗数及最低品质');
        }
      });
    } else if (f.type === 'color' || f.type === 'colors') {
      const read = chooser(
        f.name,
        c.visitor?.cardId === 'summer-33' ? '选择颜色（仅第5行使用；其他行忽略）' : '选择颜色',
        (f.options || ['red', 'white'])
          .filter((x) => c.visitor?.cardId !== 'winter-19' || v.deckCounts?.[x]?.discard > 0)
          .map((x) => [
            x,
            (names[x] || x) +
              (c.visitor?.cardId === 'winter-19' ? ' · 弃牌 ' + v.deckCounts[x].discard : ''),
          ]),
        f.type === 'colors',
        f.min,
        f.max,
      );
      readers.push((a) => (a[f.name] = read()));
    } else if (f.type === 'placementSlot') {
      const read = select('slot', '预约行动格', [
        [0, '自动选择空格'],
        [1, '第1格（3人以上为奖励格）'],
        [2, '第2格'],
        [3, '第3格'],
        [-1, '满格时大工人溢出（无奖励）'],
      ]);
      readers.push((a) => (a.slot = Number(read())));
    } else if (f.type === 'wake') {
      const read = chooser(
        f.name,
        '选择空起床行',
        (v.wakeSlots || [])
          .filter((x) => !v.players.some((p) => p.wake === x.slot))
          .map((x) => [x.slot, '第 ' + x.slot + ' 行 · ' + x.bonus]),
      );
      readers.push((a) => (a[f.name] = Number(read())));
    } else if (f.type === 'space') {
      const read = chooser(
        f.name,
        '选择行动',
        (v.spaces || [])
          .filter((x) => (f.options ? f.options.includes(x.id) : x.season === 'winter'))
          .map((x) => [x.id, x.name]),
      );
      readers.push((a) => (a[f.name] = read()));
    } else if (f.type === 'worker') {
      const read = chooser(f.name, '选择工人', [
        ...(p.workers > 0 ? [[false, '普通工人']] : []),
        ...(p.largeWorker ? [[true, '大工人']] : []),
      ]);
      readers.push((a) => (a[f.name] = String(read()) === 'true'));
    } else {
      unsupported = true;
      controls.append(
        node(
          'p',
          '此步骤需要尚未支持的界面字段：' +
            f.name +
            ' (' +
            f.type +
            ')。请更新接口契约；不会发送空参数冒充执行。',
          'ee-warning',
        ),
      );
    }
  }
  if (c.visitor?.cardId === 'winter-10' && d.option === 'discard')
    controls.append(
      node(
        'p',
        '此选项会弃置你目前全部 ' + (v.hand || []).length + ' 张手牌；确认后不可撤销。',
        'ee-warning',
      ),
    );
  const error = node('p', d.error, 'ee-warning');
  error.setAttribute('role', 'alert');
  const send = node('button', d.busy ? '提交中…' : '确认当前步骤', 'primary');
  send.type = 'submit';
  send.disabled = !online || d.busy || unsupported || !d.option;
  send.dataset.visitorSubmit = 'true';
  form.append(error, send);
  form.onsubmit = async (event) => {
    event.preventDefault();
    if (draft !== d || d.busy || !d.online || unsupported) return;
    try {
      validators.forEach((check) => check());
      const a = { type: 'choose', choiceId: c.id, revision: d.view.revision, option: d.option };
      readers.forEach((read) => read(a));
      d.busy = true;
      d.error = '';
      send.disabled = true;
      controls.disabled = true;
      const ok = await act(a);
      if (!ok) d.error = '提交未生效。选择已保留，请检查上方服务器错误或最新资源后重试。';
    } catch (e) {
      d.error = e.message;
    } finally {
      d.busy = false;
      if (draft === d && d.view.pendingChoice?.id === c.id) rerender();
    }
  };
  return true;
}
