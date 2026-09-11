import { buildings } from './building-catalog.js';
import { localCard } from './card-i18n.js';
import { cardArt } from './card-art.js';
import { art, worker, actionArt } from './graphics.js';
import {
  defaultLarge,
  privateSpace,
  freeSlots,
  hasBonus,
  winePreview,
  vineRequirements,
} from './action-options.js';
import { hint } from './hints.js';

const node = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
const names = { red: '红', white: '白', blush: '桃红', sparkling: '起泡' };
const bonusLabels = {
  draw_vine: '+1 葡萄藤',
  draw_order: '+1 订单',
  tour: '+1 金币',
  build: '少付 1 金币',
  train: '少付 1 金币',
  plant: '最多种 2 张藤',
  harvest: '最多收获 2 块田',
  make_wine: '最多酿 3 瓶',
  fill_order: '+1 胜利分',
  sell_grapes: '+1 胜利分',
  summer_visitor: '最多打出 2 张',
  winter_visitor: '最多打出 2 张',
};
let active;
export const hasActionPanel = () => !!active;
export function closeActionPanel({ restoreFocus = true } = {}) {
  if (!active) return;
  const { panel, host, source, trigger } = active;
  active = null;
  panel.remove();
  host.classList.remove('editing-action');
  source.hidden = false;
  document.querySelector('#pass').hidden = false;
  document.querySelector('.table-nav a').href = '#action-area';
  document.dispatchEvent(new Event('action-panel-close'));
  document.dispatchEvent(new CustomEvent('action-panel-state', { detail: null }));
  if (restoreFocus) document.querySelector(trigger)?.focus({ preventScroll: true });
}
document.addEventListener('keydown', (event) => {
  if (event.key === 'Escape' && active && !active.sending) {
    event.preventDefault();
    closeActionPanel();
  }
});

// A local, non-modal workspace. The estate and hand stay interactive.
export function openActionPanel(space, view, act) {
  closeActionPanel({ restoreFocus: false });
  const wake = space === 'wake';
  const p = view.players.find((player) => player.id === view.youId);
  const host = document.querySelector(wake ? '#wake-panel' : '#board');
  const source = document.querySelector(wake ? '#wake-options' : '#spaces');
  const panel = node('section', null, 'action-panel');
  panel.id = 'action-panel';
  panel.setAttribute('aria-labelledby', 'action-panel-title');
  const session = {
    panel,
    host,
    source,
    trigger: wake ? '#wake-options [data-slot="5"]' : '[data-space="' + space.id + '"]',
    sending: false,
  };
  active = session;
  host.classList.add('editing-action');
  source.hidden = true;
  document.querySelector('#pass').hidden = true;
  host.append(panel);
  document.querySelector('.table-nav a').href = '#action-panel';
  const head = node('div', null, 'action-panel-head');
  const back = node('button', '← 返回棋盘', 'action-back');
  back.type = 'button';
  back.onclick = () => closeActionPanel();
  const title = node('h3', wake ? '选择起床奖励' : space.name);
  title.id = 'action-panel-title';
  title.tabIndex = -1;
  head.append(
    back,
    title,
    art(wake ? 'cottage' : actionArt[space.id] || 'estate', 'action-heading-art'),
  );
  const assets = node('div', null, 'action-assets');
  assets.append(
    node('b', p.coins + ' 金币'),
    node('span', p.vp + ' 胜利分'),
    node('span', '工人 ' + p.workers + ' + 大工人 ' + Number(p.largeWorker)),
  );
  const estate = node('a', '查看庄园 ↗');
  estate.href = '#estate-panel';
  assets.append(estate);
  const form = node('form');
  const content = node('div');
  const preview = node('div', null, 'action-preview');
  preview.setAttribute('role', 'status');
  const error = node('p', null, 'ee-warning');
  error.id = 'action-error';
  error.setAttribute('role', 'alert');
  const footer = node('div', null, 'action-panel-footer');
  const readiness = node('span', null, 'action-readiness');
  const cancel = node('button', '取消');
  cancel.type = 'button';
  cancel.onclick = () => closeActionPanel();
  const send = node('button', wake ? '领取奖励' : '确认派遣', 'primary');
  send.id = 'confirm-action';
  send.type = 'submit';
  footer.append(readiness, cancel, send);
  panel.append(head, assets, form);
  form.append(content, preview, error, footer);
  const draft = {
    worker: wake ? '' : defaultLarge(space, p) ? 'large' : 'normal',
    slot: 0,
    declineBonus: 'no',
    color: 'summer',
    mode: '',
    building: '',
    cardId: '',
    field: null,
    cards: [],
    fields: [],
    plantFields: {},
    inspectVine: '',
    grapes: [],
    wines: [],
    bottle: 1,
    assignments: {},
  };
  if (space.id === 'yoke')
    draft.mode = p.fields.some((f) => !f.sold && !f.harvested && f.vines.length)
      ? 'harvest'
      : 'uproot';
  if (space.id === 'sell_grapes')
    draft.mode = p.grapes.length
      ? 'sell_grapes'
      : p.fields.some((f) => !f.sold && !f.vines.length)
        ? 'sell_field'
        : 'buy_field';
  const bonus = () => !wake && hasBonus(space, draft.slot, draft.declineBonus === 'yes');
  const kind = () => draft.mode || space.id;
  const limit = () => (space.id === 'yoke' ? 1 : bonus() ? 2 : 1);
  const recipes = () =>
    [1, 2, 3]
      .map((b) => p.grapes.flatMap((g, i) => (draft.assignments[i] === b ? [i] : [])))
      .filter((r) => r.length);
  const toggle = (list, value) =>
    list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
  const group = (label, compact = false) => {
    const section = node('section', null, 'action-choice-section');
    section.append(node('h4', label));
    const grid = node('div', null, 'action-card-grid' + (compact ? ' compact-choices' : ''));
    grid.setAttribute('role', 'group');
    grid.setAttribute('aria-label', label);
    section.append(grid);
    content.append(section);
    return grid;
  };
  const choice = (grid, name, value, label, options = {}) => {
    const button = node('button', null, 'action-choice ' + (options.cls || ''));
    button.type = 'button';
    button.dataset.group = name;
    button.dataset.value = String(value);
    const selected = options.selected ?? draft[name] === value;
    button.setAttribute('aria-pressed', String(selected));
    if (options.image) button.append(options.image);
    button.append(node('strong', label));
    if (options.meta) button.append(node('span', options.meta, 'choice-meta'));
    button.append(
      node(
        'small',
        (options.reason?.length > 10 ? '条件不足' : options.reason) ||
          (selected ? '✓ 已选' : options.state || '可选择'),
        'choice-state',
      ),
    );
    hint(
      button,
      (options.tip || label) + (options.reason ? '\n' + options.reason : ''),
      !!options.reason,
    );
    button.onclick = () => {
      if (session.sending) return;
      if (options.reason) {
        if (options.inspect) {
          options.inspect();
          redraw({ name, value: String(value) });
          const info = panel.querySelector('#vine-inspection');
          info?.scrollIntoView({ block: 'nearest' });
          info?.focus({ preventScroll: true });
        }
        return;
      }
      error.textContent = '';
      if (options.change) options.change();
      else draft[name] = value;
      redraw({ name, value: String(value) });
    };
    if (options.inspect && options.reason) {
      button.setAttribute('aria-disabled', 'false');
      button.classList.add('inspectable');
      button.setAttribute('aria-label', label + ' · 查看种植条件');
    }
    grid.append(button);
    return button;
  };
  const fieldLoad = (field) => field.vines.reduce((sum, c) => sum + c.red + c.white, 0);
  const plantRoom = (field, card) =>
    field.capacity -
    fieldLoad(field) -
    draft.cards
      .filter((id) => id !== card.id && draft.plantFields[id] === field.index)
      .reduce((sum, id) => {
        const c = view.hand.find((c) => c.id === id);
        return sum + c.red + c.white;
      }, 0);
  const vineReason = (c) => vineRequirements(p, c).reason;
  const buildingReason = ([id, , cost]) =>
    p.buildings.includes(id)
      ? '已建造'
      : id === 'large_cellar' && !p.buildings.includes('medium_cellar')
        ? '需要中酒窖'
        : p.coins < cost - Number(bonus())
          ? '还差 ' + (cost - Number(bonus()) - p.coins) + ' 金币'
          : '';
  const cardChoice = (grid, name, original, options = {}) => {
    const c = localCard(original);
    return choice(grid, name, c.id, c.name, {
      cls: 'hand-choice ' + c.type,
      image: cardArt(c, 'choice-art'),
      meta:
        c.type === 'vine'
          ? '红 ' + c.red + ' · 白 ' + c.white
          : c.type === 'order'
            ? (c.requirements || []).map((w) => names[w.type] + ' ≥' + w.value).join(' · ')
            : c.type === 'summer'
              ? '夏季访客'
              : '冬季访客',
      tip:
        c.name +
        '\n' +
        c.description +
        (c.type === 'order' ? '\n奖励 ' + c.points + ' 分 · 年收入 +' + c.income : ''),
      ...options,
    });
  };
  const fieldReason = (f, mode) =>
    mode === 'buy_field'
      ? !f.sold
        ? '尚未出售'
        : p.coins < f.capacity
          ? '还差 ' + (f.capacity - p.coins) + ' 金币'
          : ''
      : f.sold
        ? '已出售'
        : mode === 'sell_field'
          ? f.vines.length
            ? '需先拔除葡萄藤'
            : ''
          : !f.vines.length
            ? '尚未种植'
            : mode === 'harvest' && f.harvested
              ? '本年已收获'
              : '';
  const fieldChoice = (grid, name, f, options = {}) =>
    choice(grid, name, f.index, '田地 ' + (f.index + 1), {
      image: art('field', 'choice-art'),
      meta: '容量 ' + fieldLoad(f) + '/' + f.capacity,
      tip:
        '田地 ' +
        (f.index + 1) +
        '\n' +
        (f.vines.map((c) => localCard(c).name).join('、') || '空田地'),
      ...options,
    });
  const grapeChoice = (grid, g, i, options) =>
    choice(grid, 'grapes', i, names[g.color] + '葡萄', {
      image: art('grape-' + g.color, 'resource-art'),
      meta: '品质 ' + g.value,
      ...options,
    });
  function invalid() {
    if (wake) return '';
    if (!view.legal.canPlace) return '等待你的行动回合';
    if (!p.workers && !p.largeWorker) return '没有待命工人，仅查看条件';
    if (
      draft.worker === 'normal' &&
      (!p.workers || (!privateSpace(space) && !freeSlots(space).length))
    )
      return '需要大工人';
    if (kind() === 'build')
      return !draft.building
        ? '选择一座建筑'
        : buildingReason(buildings.find((b) => b[0] === draft.building));
    if (['summer_visitor', 'winter_visitor', 'fill_order'].includes(kind())) {
      if (!draft.cardId) return '选择一张手牌';
      if (view.cardReasons?.[draft.cardId]) return view.cardReasons[draft.cardId];
      if (kind() === 'fill_order') {
        const order = view.hand.find((c) => c.id === draft.cardId);
        const wines = p.wines.filter((w) => draft.wines.includes(w.id));
        if (wines.length !== order.requirements.length)
          return '选择所需的 ' + order.requirements.length + ' 瓶酒';
        for (const req of [...order.requirements].sort((a, b) => b.value - a.value)) {
          const i = wines.findIndex((w) => w.type === req.type && w.value >= req.value);
          if (i < 0) return '所选酒不满足订单';
          wines.splice(i, 1);
        }
      }
    }
    if (kind() === 'plant') {
      if (!draft.cards.length) {
        const vines = view.hand.filter((c) => c.type === 'vine');
        return !vines.length
          ? '没有葡萄藤手牌'
          : vines.some((c) => !vineReason(c))
            ? '选择葡萄藤'
            : '暂无可种植的藤，点卡查看条件';
      }
      if (draft.cards.length > limit()) return '本次最多种 ' + limit() + ' 张藤';
      for (const id of draft.cards) {
        const card = view.hand.find((c) => c.id === id);
        const field = p.fields.find((f) => f.index === draft.plantFields[id]);
        if (vineReason(card)) return vineReason(card);
        if (!field) return '为葡萄藤选择田地';
        if (plantRoom(field, card) < card.red + card.white) return '田地总容量不足';
      }
    }
    if (kind() === 'harvest' && space.id !== 'yoke') {
      if (!draft.fields.length) return '选择要收获的田地';
      if (draft.fields.length > limit()) return '本次最多收获 ' + limit() + ' 块田地';
    }
    if (['buy_field', 'sell_field', 'uproot'].includes(kind()) || space.id === 'yoke') {
      const field = p.fields.find((f) => f.index === draft.field);
      if (!field) return '选择一块田地';
      if (fieldReason(field, kind())) return fieldReason(field, kind());
      if (kind() === 'uproot' && !draft.cardId) return '选择要拔回的藤';
    }
    if (kind() === 'sell_grapes' && !draft.grapes.length) return '选择要出售的葡萄';
    if (kind() === 'make_wine') {
      const r = recipes();
      if (!r.length) return '选择葡萄加入当前酒瓶';
      if (Object.values(draft.assignments).some((b) => b > (bonus() ? 3 : 2)))
        return '第三瓶需要奖励格';
      if (winePreview(p, r).some((s) => /配方不成立|无可用酒槽/.test(s)))
        return '请调整配方或检查酒窖';
    }
    return '';
  }
  function redraw(focus) {
    document.dispatchEvent(new Event('action-panel-close'));
    content.replaceChildren();
    preview.replaceChildren();
    if (wake) {
      const grid = group('选择访客');
      for (const [value, label] of [
        ['summer', '夏季访客'],
        ['winter', '冬季访客'],
      ])
        choice(grid, 'color', value, label, { image: cardArt({ type: value }, 'choice-art') });
    } else {
      const workers = group('派遣工人', true);
      choice(workers, 'worker', 'normal', '普通工人', {
        image: worker(view, p.id, false),
        meta: '剩余 ' + p.workers,
        reason: !p.workers
          ? '已用完'
          : !privateSpace(space) && !freeSlots(space).length
            ? '普通格已满'
            : '',
        tip: '派遣一名普通工人，占用空闲行动格。',
      });
      choice(workers, 'worker', 'large', '大工人', {
        image: worker(view, p.id, true),
        meta: '剩余 ' + Number(p.largeWorker),
        reason: !p.largeWorker ? '已使用' : '',
        tip: '普通格已满时仍可进入；溢出位置没有行动格奖励。',
      });
      if (space.capacity >= 2 && !privateSpace(space)) {
        const slots = group('行动格', true);
        choice(slots, 'slot', 0, '自动', {
          meta: freeSlots(space).length ? '首个空格' : '大工人位',
        });
        for (let i = 1; i <= space.capacity; i++)
          choice(slots, 'slot', i, '第 ' + i + ' 格', {
            meta: i === 1 ? '★ 奖励格' : '普通格',
            reason: freeSlots(space).includes(i) ? '' : '已占用',
          });
        if (hasBonus(space, draft.slot)) {
          const rewards = group('行动格奖励', true);
          choice(rewards, 'declineBonus', 'no', '★ ' + (bonusLabels[space.id] || '领取奖励'));
          choice(rewards, 'declineBonus', 'yes', '放弃奖励');
        }
      }
      if (draft.mode) {
        const modes = group('操作', true);
        const options =
          space.id === 'yoke'
            ? [
                ['harvest', '收获田地'],
                ['uproot', '拔藤回手'],
              ]
            : [
                ['sell_grapes', '出售葡萄'],
                ['sell_field', '出售空田'],
                ['buy_field', '买回田地'],
              ];
        for (const [value, label] of options)
          choice(modes, 'mode', value, label, {
            reason:
              value === 'sell_grapes'
                ? p.grapes.length
                  ? ''
                  : '没有葡萄'
                : p.fields.some((f) => !fieldReason(f, value))
                  ? ''
                  : '没有可用田地',
            change: () => {
              draft.mode = value;
              draft.field = null;
              draft.cardId = '';
            },
          });
      }
      if (kind() === 'build') {
        const grid = group('选择建筑');
        for (const b of buildings) {
          const [id, label, cost, picture, effect] = b,
            price = cost - Number(bonus()),
            reason = buildingReason(b),
            owned = p.buildings.includes(id);
          choice(grid, 'building', id, label, {
            cls: 'building-choice',
            image: art(picture, 'choice-art'),
            meta: price + ' 金币' + (bonus() ? ' · 已减 1' : ''),
            reason,
            state: '可建造',
            tip:
              label +
              '\n' +
              effect +
              '\n本次 ' +
              price +
              ' 金币（原价 ' +
              cost +
              (bonus() ? '，奖励减 1' : '') +
              '）\n' +
              (owned ? '已建造' : '尚未建造') +
              (reason && !owned ? ' · ' + reason : ''),
          });
        }
      } else if (['summer_visitor', 'winter_visitor', 'fill_order'].includes(kind())) {
        const type =
          kind() === 'fill_order' ? 'order' : kind() === 'summer_visitor' ? 'summer' : 'winter';
        const grid = group(type === 'order' ? '选择订单' : '选择要打出的访客');
        for (const c of view.hand.filter((c) => c.type === type))
          cardChoice(grid, 'cardId', c, { reason: view.cardReasons?.[c.id] || '' });
        if (kind() === 'fill_order') {
          const wines = group('交付的酒');
          for (const w of p.wines)
            choice(wines, 'wines', w.id, names[w.type] + '酒', {
              image: art('wine', 'resource-art'),
              meta: '品质 ' + w.value,
              selected: draft.wines.includes(w.id),
              change: () => {
                draft.wines = toggle(draft.wines, w.id);
              },
            });
        }
      } else if (kind() === 'plant') {
        const grid = group('选择葡萄藤 · 最多 ' + limit() + ' 张');
        const vines = view.hand.filter((c) => c.type === 'vine');
        if (!vines.length) {
          const empty = node('div', null, 'plant-empty');
          empty.append(
            art('vine', 'resource-art'),
            node('strong', '还没有葡萄藤手牌'),
            node('span', '先去葡萄藤市场抽牌，再来查看种植条件。'),
          );
          grid.append(empty);
        }
        for (const c of vines) {
          const requirements = vineRequirements(p, c);
          const button = cardChoice(grid, 'cards', c, {
            reason: requirements.reason,
            selected: draft.cards.includes(c.id),
            inspect: () => {
              draft.inspectVine = c.id;
            },
            change: () => {
              draft.cards = toggle(draft.cards, c.id);
              draft.inspectVine = c.id;
            },
          });
          const badges = node('span', null, 'vine-checks');
          for (const check of requirements.checks.filter((c) => c.needed))
            badges.append(
              node(
                'span',
                check.label.startsWith('田地容量')
                  ? check.met
                    ? '✓ 容量 ' + requirements.value
                    : '容量不足'
                  : (check.met ? '✓ ' : '缺 ') + check.label,
                check.met ? 'met' : 'missing',
              ),
            );
          button.append(badges);
          if (draft.inspectVine === c.id) button.dataset.inspected = 'true';
        }
        if (draft.inspectVine) {
          const card = vines.find((c) => c.id === draft.inspectVine),
            requirements = vineRequirements(p, card);
          const info = node('section', null, 'vine-inspection');
          info.id = 'vine-inspection';
          info.tabIndex = -1;
          info.setAttribute('aria-label', '葡萄藤种植条件');
          const heading = node('div', null, 'section-head');
          heading.append(node('h4', localCard(card).name + ' · 种植条件'));
          const rules = node('button', '种植规则 ↗');
          rules.type = 'button';
          rules.dataset.ruleTopic = 'planting';
          heading.append(rules);
          info.append(heading);
          const structures = node('div', null, 'vine-checks');
          for (const check of requirements.checks.slice(0, 2))
            structures.append(
              node(
                'span',
                check.label + ' · ' + (!check.needed ? '无需' : check.met ? '✓ 已有' : '缺少'),
                check.met ? 'met' : 'missing',
              ),
            );
          info.append(
            structures,
            node(
              'p',
              '这张藤占用 ' +
                requirements.value +
                ' 容量（红 ' +
                card.red +
                ' + 白 ' +
                card.white +
                '）',
            ),
          );
          const fields = node('div', null, 'vine-field-checks');
          for (const f of requirements.fields)
            fields.append(
              node(
                'span',
                '田地 ' +
                  (f.index + 1) +
                  ' · ' +
                  (f.sold
                    ? '已出售'
                    : '空余 ' +
                      f.free +
                      (f.free >= requirements.value
                        ? ' ✓'
                        : ' · 差 ' + (requirements.value - f.free))),
                !f.sold && f.free >= requirements.value ? 'met' : 'missing',
              ),
            );
          info.append(fields);
          content.append(info);
        }
        for (const id of draft.cards) {
          const c = view.hand.find((c) => c.id === id),
            fields = group(localCard(c).name + ' → 田地', true);
          for (const f of p.fields)
            fieldChoice(fields, 'field-' + id, f, {
              selected: draft.plantFields[id] === f.index,
              reason: f.sold ? '已出售' : plantRoom(f, c) < c.red + c.white ? '容量不足' : '',
              change: () => {
                draft.plantFields[id] = f.index;
              },
            });
        }
      } else if (['buy_field', 'sell_field', 'uproot', 'harvest'].includes(kind())) {
        const grid = group(
          kind() === 'harvest' ? '选择田地 · 最多 ' + limit() + ' 块' : '选择田地',
        );
        const multi = kind() === 'harvest' && space.id !== 'yoke';
        for (const f of p.fields)
          fieldChoice(grid, multi ? 'fields' : 'field', f, {
            reason: fieldReason(f, kind()),
            meta: ['buy_field', 'sell_field'].includes(kind())
              ? (kind() === 'buy_field' ? '花费 ' : '获得 ') + f.capacity + ' 金币'
              : '已种 ' + fieldLoad(f) + '/' + f.capacity,
            selected: multi ? draft.fields.includes(f.index) : draft.field === f.index,
            change: () => {
              if (multi) draft.fields = toggle(draft.fields, f.index);
              else {
                draft.field = f.index;
                draft.cardId = '';
              }
            },
          });
        if (kind() === 'uproot' && draft.field != null) {
          const vines = group('选择拔回手牌的藤');
          for (const c of p.fields.find((f) => f.index === draft.field).vines)
            cardChoice(vines, 'cardId', c);
        }
      } else if (kind() === 'sell_grapes') {
        const grid = group('选择出售的葡萄');
        p.grapes.forEach((g, i) =>
          grapeChoice(grid, g, i, {
            selected: draft.grapes.includes(i),
            state: '+' + Math.ceil(g.value / 3) + ' 金币',
            change: () => {
              draft.grapes = toggle(draft.grapes, i);
            },
          }),
        );
        if (draft.grapes.length)
          preview.textContent =
            '获得 ' +
            draft.grapes.reduce((sum, i) => sum + Math.ceil(p.grapes[i].value / 3), 0) +
            ' 金币';
      } else if (kind() === 'make_wine') {
        const bottles = group('先选酒瓶，再添加葡萄 · 最多 ' + (bonus() ? 3 : 2) + ' 瓶', true);
        for (let b = 1; b <= 3; b++)
          choice(bottles, 'bottle', b, '第 ' + b + ' 瓶', {
            meta: Object.values(draft.assignments).filter((v) => v === b).length + ' 颗葡萄',
            reason:
              b === 3 && !bonus() && !Object.values(draft.assignments).includes(3)
                ? '需要奖励格'
                : '',
            tip: '单颗红／白酿红／白酒；红白各一酿桃红；两红一白酿起泡。',
          });
        const grid = group('葡萄 · 再点一次移出');
        p.grapes.forEach((g, i) =>
          grapeChoice(grid, g, i, {
            selected: draft.assignments[i] === draft.bottle,
            state: draft.assignments[i] ? '已入第 ' + draft.assignments[i] + ' 瓶' : '待分配',
            tip: '点击加入第 ' + draft.bottle + ' 瓶；再次点击可移出。',
            change: () => {
              if (draft.assignments[i] === draft.bottle) delete draft.assignments[i];
              else draft.assignments[i] = draft.bottle;
            },
          }),
        );
        const bottlesUsed = [1, 2, 3].filter((b) => Object.values(draft.assignments).includes(b));
        winePreview(p, recipes()).forEach((text, i) =>
          preview.append(node('p', text.replace(/^第 \d+ 瓶：/, '第 ' + bottlesUsed[i] + ' 瓶：'))),
        );
      } else {
        preview.textContent = space.description;
      }
    }
    const reason = invalid();
    readiness.textContent = reason || '✓ 已就绪';
    readiness.classList.toggle('ready', !reason);
    send.disabled = !!reason || session.sending;
    document.dispatchEvent(
      new CustomEvent('action-panel-state', {
        detail: { space: wake ? 'wake' : space.id, reason },
      }),
    );
    if (focus)
      [...content.querySelectorAll('[data-group]')]
        .find((e) => e.dataset.group === focus.name && e.dataset.value === focus.value)
        ?.focus({ preventScroll: true });
  }
  form.onsubmit = async (event) => {
    event.preventDefault();
    if (invalid() || session.sending) return;
    const action = wake
      ? { type: 'wake', slot: 5, color: draft.color }
      : {
          type: 'place',
          space: space.id,
          large: draft.worker === 'large',
          slot: draft.slot,
          declineBonus: draft.declineBonus === 'yes',
        };
    action.revision = view.revision;
    if (!wake) {
      if (draft.mode) action.mode = draft.mode;
      if (draft.field != null) action.field = draft.field;
      if (draft.cardId) action.cardId = draft.cardId;
      if (draft.building) action.building = draft.building;
      if (kind() === 'plant') {
        action.cardIds = draft.cards;
        action.fields = draft.cards.map((id) => draft.plantFields[id]);
      }
      if (kind() === 'harvest' && space.id !== 'yoke') action.fields = draft.fields;
      if (kind() === 'sell_grapes') action.grapes = draft.grapes;
      if (kind() === 'make_wine') action.recipes = recipes();
      if (kind() === 'fill_order') action.wineIds = draft.wines;
    }
    session.sending = true;
    send.disabled = cancel.disabled = back.disabled = true;
    try {
      if (await act(action)) {
        if (active === session) closeActionPanel({ restoreFocus: false });
      } else if (!error.textContent) error.textContent = '操作未生效，请检查选择。';
    } catch (err) {
      error.textContent = err.message;
    } finally {
      session.sending = false;
      send.disabled = !!invalid();
      cancel.disabled = back.disabled = false;
    }
  };
  redraw();
  title.focus({ preventScroll: true });
  panel.scrollIntoView({ block: 'start' });
}
