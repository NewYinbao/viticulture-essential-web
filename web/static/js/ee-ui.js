import { cardArt } from './card-art.js';
import { openActionPanel } from './action-panel.js';
export { cardArt } from './card-art.js';
export { openActionPanel as setupEE } from './action-panel.js';
import { renderVisitor } from './visitor-ui.js';
import { localCard } from './card-i18n.js';
import { renderTuscanyChoice } from './tuscany-choice.js';
import { tuscanyInputState, renderTuscanyInputs } from './tuscany-inputs.js';
import { availableBuildings } from './building-catalog.js';
const $ = (s) => document.querySelector(s);
const n = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
const typeNames = {
  vine: '葡萄藤',
  order: '订单',
  summer: '夏季访客',
  winter: '冬季访客',
  structure: '结构牌',
  mama: '家族起始牌 A',
  papa: '家族起始牌 B',
  red: '红',
  white: '白',
  blush: '桃红',
  sparkling: '起泡',
};
const gifts = {
  trellis: '棚架',
  irrigation: '灌溉塔',
  yoke: '轭',
  medium_cellar: '中酒窖',
  cottage: '小屋',
  windmill: '磨坊',
  tasting_room: '品酒室',
  worker: '额外工人',
  vp: '1 胜利分',
};
function panel() {
  let e = $('#ee-choice');
  if (!e) {
    e = n('section', null, 'panel ee-choice');
    e.id = 'ee-choice';
    $('#game').insertBefore(e, $('#lobby-panel'));
  }
  return e;
}
export function renderEE(v, act, online) {
  renderDiscard(v, act, online);
  const box = panel(),
    metaOpen = box.querySelector('.table-info')?.open,
    familyOpen = box.querySelector('.ee-family')?.open;
  box.replaceChildren();
  const meta = n('details', null, 'table-info');
  meta.open = !!metaOpen;
  meta.append(n('summary', '牌堆与家族信息'));
  const head = n('div', null, 'ee-heading');
  const counts = n('div', null, 'ee-deck-counts');
  for (const [type, c] of Object.entries(v.deckCounts || {})) {
    if (type === 'structure' && !v.config?.structures) continue;
    counts.append(n('span', `${typeNames[type] || type} ${c.deck} / 弃 ${c.discard}`));
  }
  head.append(counts);
  meta.append(head);
  box.append(meta);
  box.hidden = v.phase === 'lobby';
  if (v.phase === 'finished' && v.winnerIds?.length > 1)
    $('#turn-label').textContent =
      '共同获胜：' +
      v.players
        .filter((p) => v.winnerIds.includes(p.id))
        .map((p) => p.name)
        .join('、');
  const me = v.players.find((p) => p.id === v.youId);
  if (v.config?.specialWorkers && v.specialWorkerPool?.length) {
    const panel = n('details', null, 'ee-family');
    panel.open = v.phase === 'setup' || !!familyOpen;
    panel.append(n('summary', '本局特殊工人 · 公开牌池'));
    const cards = n('div', null, 'ee-parent-row');
    for (const id of v.specialWorkerPool) {
      const w = (v.specialWorkerCatalog || []).find((item) => item.id === id);
      const card = n('article', null, 'ee-parent-card');
      card.append(
        n('strong', w?.name || id),
        n('small', w?.description || '特殊能力由服务器按规则结算'),
      );
      cards.append(card);
    }
    panel.append(cards);
    if (v.phase === 'setup') box.append(panel);
    else meta.append(panel);
  }
  if (me?.mama?.id) {
    const family = n('details', null, 'ee-family');
    family.open = v.phase === 'setup' || !!familyOpen;
    family.append(
      n('summary', '家族传承 · ' + localCard(me.mama).name + ' ＆ ' + localCard(me.papa).name),
    );
    const row = n('div', null, 'ee-parent-row');
    for (const original of [me.mama, me.papa]) {
      const c = localCard(original);
      const a = n('article', null, 'ee-parent-card');
      a.append(
        cardArt(c),
        n('strong', c.name),
        n('small', c.englishName, 'card-english'),
        n('small', c.description || '开局资源见当前选择'),
      );
      row.append(a);
    }
    family.append(row);
    if (v.phase === 'setup') box.append(family);
    else meta.append(family);
  }
  if (v.phase === 'finished') {
    const result = n('section', null, 'final-results');
    result.append(
      n(
        'h3',
        '最终结算 · ' +
          v.players
            .filter((p) => (v.winnerIds || [v.winnerId]).includes(p.id))
            .map((p) => p.name)
            .join('、') +
          ' 获胜',
      ),
    );
    const table = n('table');
    const head = n('tr');
    for (const text of ['庄主', '胜利分', '金币', '酒总值', '葡萄总值']) head.append(n('th', text));
    table.append(head);
    for (const p of v.players) {
      const row = n('tr');
      for (const text of [
        p.name,
        p.vp,
        p.coins,
        (p.wines || []).reduce((s, w) => s + w.value, 0),
        (p.grapes || []).reduce((s, g) => s + g.value, 0),
      ])
        row.append(n('td', text));
      table.append(row);
    }
    result.append(
      table,
      n('small', '依次比较胜利分、金币、酒总值、葡萄总值；完全相同则共同获胜。'),
    );
    box.append(result);
  }
  const c = v.pendingChoice;
  if (!c) {
    renderVisitor(document.createElement('div'), v, act, online);
    return;
  }
  const rules = n('button', '当前选择的规则', 'choice-rules-link');
  rules.type = 'button';
  rules.dataset.ruleTopic = 'current';
  box.append(rules);
  if (c.visitor || ['visitor', 'planner'].includes(c.kind)) {
    const visitor = n('section', null, 'visitor-choice');
    box.append(visitor);
    renderVisitor(visitor, v, act, online);
    return;
  }
  renderVisitor(document.createElement('div'), { ...v, pendingChoice: null }, act, online);
  const your = c.playerId === v.youId;
  box.append(
    n(
      'h3',
      your
        ? '请完成当前选择'
        : '等待 ' + (v.players.find((p) => p.id === c.playerId)?.name || '玩家') + ' 完成选择',
    ),
  );
  if (!your) {
    box.append(n('p', '其他玩家的手牌与选择内容不会向你公开。', 'muted'));
    return;
  }
  const form = n('form', null, 'ee-choice-form');
  const err = n('p', null, 'ee-warning');
  box.append(form, err);
  let locked = false;
  const submit = async (a) => {
    if (locked || !online) return;
    locked = true;
    for (const b of form.querySelectorAll('button')) b.disabled = true;
    const ok = await act({
      type: 'choose',
      choiceId: c.id,
      revision: v.revision,
      ...a,
    });
    if (!ok) {
      locked = false;
      for (const b of form.querySelectorAll('button')) b.disabled = !online;
      err.textContent = '提交未生效，请检查资源或局面是否已经变化。';
    }
    return ok;
  };
  if (renderTuscanyChoice(form, v, online, submit)) return;
  if (renderPoliticoChoice(form, v, online, submit)) return;
  if (
    ['messenger', 'special_mafioso', 'structure_fermentation', 'structure_mercado'].includes(c.kind)
  ) {
    const kind = c.kind;
    const spaceId =
      kind === 'messenger' || kind === 'special_mafioso'
        ? c.actionSpace
        : kind === 'structure_fermentation'
          ? 'make_wine'
          : 'fill_order';
    const space = v.spaces.find((s) => s.id === spaceId);
    const choose = n(
      'button',
      kind === 'messenger'
        ? '选择当前资源并执行信使预约'
        : kind === 'special_mafioso'
          ? '选择再次行动的资源'
          : kind === 'structure_fermentation'
            ? '选择葡萄 · 酿造1瓶'
            : '选择订单用酒',
      'primary',
    );
    choose.type = 'button';
    choose.disabled = !online || !space;
    choose.dataset.choiceResources = kind;
    choose.onclick = () =>
      openActionPanel(
        space,
        v,
        (action) =>
          submit({
            ...action,
            type: 'choose',
            choiceId: c.id,
            option:
              kind === 'messenger'
                ? 'execute'
                : kind === 'special_mafioso'
                  ? 'repeat'
                  : kind === 'structure_fermentation'
                    ? 'make'
                    : 'use',
          }),
        {
          continuation: true,
          continuationSlot: kind === 'messenger' ? c.actionSlot : undefined,
          continuationBonus: kind === 'messenger' ? c.specialBonus : undefined,
          maxWines: kind === 'structure_fermentation' ? 1 : undefined,
          allowedCardIDs: kind === 'structure_mercado' ? c.cardIds || [] : undefined,
        },
      );
    const skip = n('button', '跳过', 'secondary');
    skip.type = 'button';
    skip.disabled = !online;
    skip.dataset.choice = 'skip';
    skip.onclick = () => submit({ option: 'skip' });
    form.append(choose);
    if (kind !== 'messenger') form.append(skip);
    return;
  }
  if (c.kind === 'structure_visitor_bonus' && c.playerId === v.youId) {
    const wineInputs = [],
      grapeInputs = [];
    const wineGroup = n('div', null, 'choice-resource-group');
    wineGroup.append(n('p', '酒馆/品酒吧：选择要弃置的酒（1瓶）'));
    for (const w of me?.wines || []) {
      const label = n('label', null, 'check');
      const input = document.createElement('input');
      input.type = 'radio';
      input.name = 'structure-bonus-wine';
      input.value = w.id;
      label.append(
        input,
        document.createTextNode((typeNames[w.type] || w.type) + '酒 品质' + w.value),
      );
      wineGroup.append(label);
      wineInputs.push(input);
    }
    const grapeGroup = n('div', null, 'choice-resource-group');
    grapeGroup.append(n('p', '酒馆：选择要弃置的红/白葡萄（2颗）'));
    for (const g of me?.grapes || []) {
      if (!['red', 'white'].includes(g.color)) continue;
      const label = n('label', null, 'check');
      const input = document.createElement('input');
      input.type = 'checkbox';
      input.name = 'structure-bonus-grape';
      input.value = g.id;
      label.append(
        input,
        document.createTextNode((typeNames[g.color] || g.color) + '葡萄 品质' + g.value),
      );
      grapeGroup.append(label);
      grapeInputs.push(input);
    }
    if ((c.schema || []).some((field) => field.name === 'wineIds')) form.append(wineGroup);
    if ((c.schema || []).some((field) => field.name === 'grapeIds')) form.append(grapeGroup);
    for (const option of c.options || []) {
      const b = n('button', c.labels?.[option] || option, 'primary');
      b.type = 'button';
      b.disabled = !online;
      b.onclick = () =>
        submit({
          option,
          wineIds: wineInputs.filter((x) => x.checked).map((x) => x.value),
          grapeIds: grapeInputs.filter((x) => x.checked).map((x) => x.value),
        });
      form.append(b);
    }
    return;
  }
  if (c.kind === 'discard' || c.kind === 'structure_barn') {
    form.remove();
    err.remove();
    box.append(n('p', '请在下方原手牌区选择弃牌。'));
  } else {
    const specialLabels = {
      special_train: {
        grande: '重新培训大工人',
        regular:
          c.actionSpace === 'school_training' ? '普通工人（免费 · 本年可用）' : '普通工人（4金币）',
      },
      special_farmer: { skip: '不额外领取奖励' },
      special_mafioso: { repeat: '再执行一次行动', skip: '不重复行动' },
      special_politico: {
        repeat_bonus: '支付1金币，再取一次奖励',
        skip: '不重复奖励',
      },
      special_merchant: {
        vine: '抽1张藤蔓牌',
        order: '抽1张订单牌',
        summer: '抽1张夏季访客',
        winter: '抽1张冬季访客',
      },
      special_innkeeper: { skip: '不交换访客' },
    };
    const workerName = (id) => (v.specialWorkerCatalog || []).find((w) => w.id === id)?.name || id;
    const bonusNames = {
      coin: '1金币',
      vp: '1分',
      draw_vine: '抽1张葡萄藤',
      draw_order: '抽1张订单',
      discount: '折扣1金币',
      plant: '种至多2藤',
      harvest: '收获至多2田',
      make_wine: '酿至多3瓶酒',
      visitor: '再打出至多2名访客',
      influence: '额外放置/移动影响力',
      trade: '额外交易',
      build_tour: '再次建造或导览',
    };
    for (const option of c.options || []) {
      let text =
        c.labels?.[option] ||
        specialLabels[c.kind]?.[option] ||
        {
          summer_summer: '夏季访客＋夏季访客',
          summer_winter: '夏季访客＋冬季访客',
          winter_winter: '冬季访客＋冬季访客',
          'summer+summer': '夏季访客＋夏季访客',
          'summer+winter': '夏季访客＋冬季访客',
          'winter+winter': '冬季访客＋冬季访客',
        }[option] ||
        typeNames[option] ||
        option;
      if (c.kind === 'special_train' && !['regular', 'grande'].includes(option))
        text =
          '特殊工人 · ' +
          workerName(option) +
          (c.actionSpace === 'school_training' ? '（1金币 · 本年可用）' : '（次年可用）');
      if (c.kind === 'special_farmer' && option !== 'skip')
        text = '选择奖励：' + (bonusNames[option] || option);
      if (c.kind === 'special_professore' && option !== 'skip') {
        const [space, slot] = option.split(':');
        text =
          '收回 ' +
          (v.spaces?.find((s) => s.id === space)?.name || space) +
          ' 第' +
          slot +
          '格普通工人';
      }
      if (c.kind === 'special_oracle') {
        const card = v.hand?.find((item) => item.id === option);
        text = '弃置本次抽到的 ' + (card ? localCard(card).name : option);
      }
      if (c.kind === 'papa') {
        const o = v.parentOptions;
        text =
          option === 'gift'
            ? '接受赠礼：' + (gifts[o?.gift] || o?.gift)
            : '改为额外 ' + o?.coins + ' 金币';
      }
      if (c.kind === 'tuscany_next_wake')
        text =
          '第' +
          option +
          '行 · ' +
          (v.wakeSlots?.find((w) => String(w.slot) === option)?.bonus || '预约下一年');
      if (c.kind === 'tuscany_upkeep') text = '领取 ' + (me?.income || 0) + ' 金币年收入，继续预约';
      const b = n('button', text, 'primary');
      b.type = 'button';
      b.disabled = !online;
      b.dataset.choice = option;
      b.onclick = () => {
        const pair = option.split(/[_+,]/);
        submit({
          option,
          ...(pair.length === 2 && pair.every((x) => ['summer', 'winter'].includes(x))
            ? { colors: pair }
            : {}),
        });
      };
      form.append(b);
    }
  }
}
function select(form, name, label, items) {
  const l = n('label', label),
    s = n('select');
  s.name = name;
  s.required = true;
  for (const [value, text] of items) {
    const o = n('option', text);
    o.value = value;
    s.append(o);
  }
  l.append(s);
  form.append(l);
  return s;
}
function checks(form, name, label, items) {
  form.append(n('p', label));
  for (const [value, text] of items) {
    const l = n('label', null, 'check'),
      i = n('input');
    i.type = 'checkbox';
    i.name = name;
    i.value = value;
    l.append(i, document.createTextNode(text));
    form.append(l);
  }
}
let politicoDraft = null;
const politicoLabels = {
  plant: '最多选择2张葡萄藤并指定田地。',
  harvest: '最多选择2块本年尚未收获的田地。',
  make_wine: '为葡萄分配酒瓶；配方按酒窖规则校验。',
  visitor: '选择要再次打出的夏季或冬季访客。',
  influence: '完成额外一次影响力放置或移动。',
  trade: '完成额外一次交易。',
  build_tour: '再次建造（少付1金币）或导览（多得1金币）。',
};
function renderPoliticoChoice(form, v, online, submit) {
  const c = v.pendingChoice;
  if (c?.kind !== 'special_politico' || !c.specialBonus || !c.options.includes('confirm')) {
    politicoDraft = null;
    return false;
  }
  const key = JSON.stringify([v.code, v.youId, c.id]);
  if (politicoDraft?.key !== key)
    politicoDraft = {
      key,
      values: {
        cards: [],
        fields: [],
        assignment: {},
        mode: 'tour',
        color: 'summer',
        cardId: '',
        tuscany: {},
      },
    };
  const d = politicoDraft.values;
  form.replaceChildren();
  form.append(
    n(
      'p',
      'Politico：' + (politicoLabels[c.specialBonus] || '再次获得奖励') + ' 需支付1金币。',
      'muted',
    ),
  );
  const p = v.players.find((x) => x.id === v.youId);
  const vineyard = (v.hand || []).filter((card) => card.type === 'vine');
  const payload = () => {
    if (c.specialBonus === 'plant')
      return {
        cardIds: d.cards,
        fields: d.cards.map((id) => Number(d.fields[id])),
      };
    if (c.specialBonus === 'harvest') return { fields: d.fields.map(Number) };
    if (c.specialBonus === 'make_wine')
      return {
        recipes: [1, 2, 3]
          .map((b) =>
            (p?.grapes || [])
              .map((_, i) => (d.assignment[i] === String(b) ? i : -1))
              .filter((i) => i >= 0),
          )
          .filter((r) => r.length),
      };
    if (c.specialBonus === 'visitor') return { color: d.color, cardId: d.cardId };
    if (c.specialBonus === 'influence' || c.specialBonus === 'trade')
      return c.specialBonus === 'influence'
        ? {
            influence: tuscanyInputState('influence', v, d.tuscany).payload.influence,
          }
        : { trades: tuscanyInputState('trade', v, d.tuscany).payload.trades };
    if (c.specialBonus === 'build_tour') return { mode: d.mode, building: d.building || '' };
    return {};
  };
  const addButton = (text, option, disabled = false) => {
    const b = n('button', text, option === 'confirm' ? 'primary' : '');
    b.type = 'button';
    b.disabled = !online || disabled;
    b.onclick = () => submit({ option, ...payload() });
    form.append(b);
  };
  if (c.specialBonus === 'influence' || c.specialBonus === 'trade') {
    const state = tuscanyInputState(c.specialBonus, v, d.tuscany);
    renderTuscanyInputs(form, state, d.tuscany, online, () =>
      renderPoliticoChoice(form, v, online, submit),
    );
    form.append(n('p', state.reason || '已就绪；确认后支付1金币并执行。', 'muted'));
    addButton('确认并执行', 'confirm', !!state.reason);
    addButton('放弃重复奖励', 'skip');
    return true;
  }
  if (c.specialBonus === 'plant') {
    checks(
      form,
      'politico-vine',
      '选择葡萄藤（1–2张）',
      vineyard.map((card) => [card.id, localCard(card).name]),
    );
    for (const input of form.querySelectorAll('input[name="politico-vine"]')) {
      input.checked = d.cards.includes(input.value);
      input.onchange = () => {
        if (input.checked && !d.cards.includes(input.value)) d.cards.push(input.value);
        if (!input.checked) d.cards = d.cards.filter((id) => id !== input.value);
        renderPoliticoChoice(form, v, online, submit);
      };
    }
    for (const id of d.cards) {
      const label = n(
        'label',
        (localCard(vineyard.find((card) => card.id === id))?.name || id) + ' → 田地',
      );
      const select = n('select');
      select.required = true;
      for (const f of p?.fields || []) {
        const o = n('option', '田地' + (f.index + 1));
        o.value = f.index;
        select.append(o);
      }
      select.value = d.fields[id] ?? '';
      select.onchange = () => {
        d.fields[id] = select.value;
      };
      label.append(select);
      form.append(label);
    }
  } else if (c.specialBonus === 'harvest') {
    checks(
      form,
      'politico-field',
      '选择田地（1–2块）',
      (p?.fields || []).map((f) => [String(f.index), '田地' + (f.index + 1)]),
    );
    for (const input of form.querySelectorAll('input[name="politico-field"]')) {
      input.checked = d.fields.includes(Number(input.value));
      input.onchange = () => {
        const field = Number(input.value);
        if (input.checked && !d.fields.includes(field)) d.fields.push(field);
        if (!input.checked) d.fields = d.fields.filter((x) => x !== field);
        renderPoliticoChoice(form, v, online, submit);
      };
    }
  } else if (c.specialBonus === 'make_wine') {
    form.append(n('p', '每颗葡萄选择酒瓶0=不使用、1–3=对应酒瓶。'));
    for (const [i, grape] of (p?.grapes || []).entries()) {
      const label = n('label', (grape.color === 'red' ? '红' : '白') + grape.value + ' · 酒瓶'),
        select = n('select');
      [
        ['0', '不使用'],
        ['1', '1'],
        ['2', '2'],
        ['3', '3'],
      ].forEach(([value, text]) => {
        const o = n('option', text);
        o.value = value;
        select.append(o);
      });
      select.value = d.assignment[i] || '0';
      select.onchange = () => {
        if (select.value === '0') delete d.assignment[i];
        else d.assignment[i] = select.value;
      };
      label.append(select);
      form.append(label);
    }
  } else if (c.specialBonus === 'visitor') {
    const color = select(form, 'politico-color', '季节', [
      ['summer', '夏季访客'],
      ['winter', '冬季访客'],
    ]);
    color.value = d.color;
    color.onchange = () => {
      d.color = color.value;
      renderPoliticoChoice(form, v, online, submit);
    };
    const cards = (v.hand || []).filter((card) => card.type === d.color),
      card = select(form, 'politico-card', '访客', [
        ['', '请选择'],
        ...cards.map((x) => [x.id, localCard(x).name]),
      ]);
    card.value = d.cardId;
    card.onchange = () => {
      d.cardId = card.value;
    };
  } else if (c.specialBonus === 'build_tour') {
    const mode = select(form, 'politico-mode', '再次奖励', [
      ['tour', '导览（+1金币）'],
      ['build', '建造（少付1金币）'],
    ]);
    mode.value = d.mode;
    mode.onchange = () => {
      d.mode = mode.value;
      renderPoliticoChoice(form, v, online, submit);
    };
    if (d.mode === 'build') {
      const building = select(form, 'politico-building', '建筑', [
        ['', '请选择'],
        ...availableBuildings(v).map((b) => [b[0], b[1] + ' · ' + b[2] + '金币']),
      ]);
      building.value = d.building || '';
      building.onchange = () => {
        d.building = building.value;
      };
    }
  }
  const ready =
    c.specialBonus === 'plant'
      ? d.cards.length > 0 &&
        d.cards.length <= 2 &&
        d.cards.every((id) => d.fields[id] !== undefined && d.fields[id] !== '')
      : c.specialBonus === 'harvest'
        ? d.fields.length > 0 && d.fields.length <= 2
        : c.specialBonus === 'make_wine'
          ? Object.values(d.assignment).length > 0
          : c.specialBonus === 'visitor'
            ? !!d.cardId
            : c.specialBonus === 'build_tour'
              ? d.mode === 'tour' || !!d.building
              : true;
  addButton('确认并执行', 'confirm', !ready);
  addButton('放弃重复奖励', 'skip');
  return true;
}
// Choice-scoped state survives SSE, not another room/player/choice.
let discard = null;
function renderDiscard(v, act, online) {
  const c = v.pendingChoice,
    barn = c?.kind === 'structure_barn',
    own = (c?.kind === 'discard' || barn) && c.playerId === v.youId;
  const key = own ? JSON.stringify([v.code, v.youId, v.year, c.id]) : null;
  const fresh = key !== discard?.key;
  $('#discard-prompt')?.remove();
  $('#hand').classList.toggle('discard-hand', own);
  if (!own) {
    discard = null;
    return;
  }
  if (fresh) discard = { key, ids: new Set(), busy: false, error: '' };
  const d = discard,
    valid = new Set((v.hand || []).map((c) => c.id));
  d.view = v;
  d.online = online;
  for (const id of d.ids) if (!valid.has(id)) d.ids.delete(id);
  const prompt = n('section', null, 'discard-prompt');
  prompt.id = 'discard-prompt';
  prompt.setAttribute('role', 'region');
  prompt.setAttribute('aria-labelledby', 'discard-title');
  const title = n('h3', barn ? '谷仓：可弃2张牌得1分' : '年末需要弃牌');
  title.id = 'discard-title';
  const status = n('p');
  status.id = 'discard-status';
  status.setAttribute('role', 'status');
  const hint = n('p', '点击下方原手牌选择，再点取消；键盘 Tab 定位，空格 / Enter 选择。', 'muted');
  const error = n('p', d.error, 'ee-warning');
  error.setAttribute('role', 'alert');
  const send = n('button', '确认弃牌', 'primary');
  send.id = 'confirm-discard';
  send.type = 'button';
  const skip = barn ? n('button', '跳过谷仓', 'secondary') : null;
  if (skip) {
    skip.type = 'button';
    skip.disabled = !online || d.busy;
    skip.onclick = () =>
      act({
        type: 'choose',
        choiceId: c.id,
        revision: v.revision,
        option: 'skip',
      });
  }
  prompt.append(title, status, hint, send, ...(skip ? [skip] : []), error);
  $('#hand-panel').insertBefore(prompt, $('#hand'));
  const cards = [...document.querySelectorAll('#hand .card')];

  const update = () => {
    status.textContent =
      (barn ? '选择2张牌换1分 · ' : '需要弃置 ' + c.count + ' 张 · ') +
      d.ids.size +
      ' / ' +
      c.count +
      ' 张（保留不超过7张）';
    send.disabled = !online || d.busy || d.ids.size !== c.count;
    send.textContent = d.busy ? '正在提交…' : '确认弃牌';
    prompt.setAttribute('aria-busy', String(d.busy));
    for (const card of cards) {
      const chosen = d.ids.has(card.dataset.cardId);
      card.classList.toggle('discard-selected', chosen);
      card.setAttribute('aria-pressed', String(chosen));
      card.setAttribute('aria-disabled', String(!online || d.busy));
    }
  };
  for (const card of cards) {
    card.tabIndex = 0;
    card.setAttribute('role', 'button');
    card.setAttribute('aria-describedby', 'discard-status');
    const toggle = () => {
      if (!online || d.busy) return;
      const id = card.dataset.cardId;
      if (d.ids.has(id)) d.ids.delete(id);
      else d.ids.add(id);
      update();
    };
    card.onclick = toggle;
    card.onkeydown = (e) => {
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault();
        if (!e.repeat) toggle();
      }
    };
  }
  send.onclick = async () => {
    if (discard !== d || d.busy || !online || d.ids.size !== c.count) return;
    d.busy = true;
    d.error = '';
    update();
    let ok = false;
    try {
      ok = await act({
        type: 'choose',
        choiceId: c.id,
        revision: v.revision,
        option: barn ? 'take' : undefined,
        cardIds: [...d.ids],
      });
    } catch (e) {
      d.error = e.message;
    } finally {
      d.busy = false;
      if (discard === d) {
        if (ok) d.ids.clear();
        else d.error = d.error || '提交未生效，请检查连接或最新局面后重试。';
        renderDiscard(d.view, act, d.online);
      }
    }
  };
  update();
  if (fresh) prompt.scrollIntoView({ block: 'start' });
}
