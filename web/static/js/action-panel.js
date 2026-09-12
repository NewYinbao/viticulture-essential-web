import { buildingCardArt } from './card-illustrations.js';
import { cardText } from './card-text.js';
import { tuscanyInputState, renderTuscanyInputs } from './tuscany-inputs.js';
import { buildings, structureBuildings, availableBuildings } from './building-catalog.js';
import { localCard } from './card-i18n.js';
import { cardArt } from './card-art.js';
import { art, worker, actionArt } from './graphics.js';
import {
  defaultLarge,
  privateSpace,
  freeSlots,
  hasBonus,
  bonusKey,
  bonusLabel,
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
export function openActionPanel(space, view, act, options = {}) {
  const allowedCardIDs = options.allowedCardIDs;
  const continuation = !!options.continuation;
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
  const send = node(
    'button',
    wake ? '领取奖励' : continuation ? '确认选择' : '确认派遣',
    'primary',
  );
  send.id = 'confirm-action';
  send.type = 'submit';
  footer.append(readiness, cancel, send);
  panel.append(head, assets, form);
  form.append(content, preview, error, footer);
  const draft = {
    tuscany: {},
    bonusFirst: 'after',
    worker: wake ? '' : defaultLarge(space, p) ? 'large' : 'normal',
    workerType: '',
    specialWorker: 'regular',
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
    recipeTypes: {},
    harvestAll: false,
    structureTarget: '',
  };
  const futureReservation = () =>
    !continuation && draft.workerType === 'messenger' && space.season !== view.phase;
  const placementsFor = (
    type = draft.workerType,
    large = draft.worker === 'large',
    gray = draft.worker === 'gray',
  ) =>
    view.workerPlacements?.[space.id]?.filter(
      (item) => item.workerType === type && item.large === large && !!item.gray === gray,
    );
  const chosenPlacement = () =>
    placementsFor()?.find((item) => draft.slot === 0 || item.slot === draft.slot);
  if (
    !wake &&
    !continuation &&
    !placementsFor()?.length &&
    view.workerPlacements?.[space.id]?.length
  ) {
    const first = view.workerPlacements[space.id][0];
    draft.workerType = first.workerType;
    draft.worker = first.gray ? 'gray' : first.large ? 'large' : 'normal';
  }
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
  if (space.id === 'flip_field')
    draft.mode = p.fields.some((f) => !f.sold && !f.vines.length) ? 'sell_field' : 'buy_field';
  if (space.id === 'build_tour') draft.mode = 'tour';
  if (space.id === 'harvest' && p.buildings.includes('harvest_machine')) draft.mode = 'harvest';
  const placementSpace = () => ({
    ...space,
    capacity: Math.max(space.capacity || 0, ...(placementsFor() || []).map((item) => item.slot)),
  });
  const printedBonus = () =>
    wake
      ? ''
      : continuation
        ? draft.declineBonus === 'yes'
          ? ''
          : options.continuationBonus || ''
        : bonusKey(
            placementSpace(),
            chosenPlacement()?.slot ?? draft.slot,
            draft.declineBonus === 'yes',
          );
  const bonus = () => !!printedBonus() && (printedBonus() !== 'coin' || kind() === 'tour');
  const trainingToll = () =>
    view.config?.structures
      ? view.players.filter((q) => q.id !== p.id && q.buildings?.includes('academy')).length
      : 0;
  const placementToll = () => (continuation ? 0 : chosenPlacement()?.toll || 0);
  const farmerTrainingDiscount = () =>
    !continuation &&
    draft.workerType === 'farmer' &&
    draft.declineBonus !== 'yes' &&
    printedBonus() !== 'discount';
  const trainingPrice = (id) =>
    4 +
    (['regular', 'grande'].includes(id) ? 0 : 1) -
    (printedBonus() === 'discount' || farmerTrainingDiscount() ? 1 : 0) +
    trainingToll() +
    placementToll();
  const trainingReason = (id) => {
    if (p.totalWorkers >= 6) return '工人已达6名上限';
    if (id !== 'grande' && p.totalWorkers - (p.grandeRemoved ? 0 : 1) >= 5)
      return '普通和特殊工人合计已达5名';
    if (id !== 'regular' && p.specialWorkers?.includes(id)) return '已拥有此工人';
    if (p.coins < trainingPrice(id)) return '还差' + (trainingPrice(id) - p.coins) + '金币';
    return '';
  };
  const structureAction = !!space.id?.startsWith('structure_action:');
  const destroyAction = !!space.id?.startsWith('structure_destroy:');
  const structureID = structureAction ? space.id.split(':').slice(-1)[0] : '';
  const structure = structureBuildings.find((b) => b[0] === structureID);
  const resourceState = () =>
    tuscanyInputState(structureID === 'trading_post' ? 'trade' : space.id, view, draft.tuscany);
  const kind = () =>
    space.id === 'harvest' && draft.mode === 'all' ? 'harvest' : draft.mode || space.id;
  const cardReason = (id) => {
    const table = view.visitorCardReasons;
    const visitor = ['summer_visitor', 'winter_visitor'].includes(space.id);
    if (!visitor || !table) return view.cardReasons?.[id] || '';
    return (
      table[printedBonus() === 'coin' && draft.bonusFirst === 'before' ? 'coinBefore' : 'base']?.[
        id
      ] || ''
    );
  };
  const limit = () => (space.id === 'yoke' ? 1 : bonus() ? 2 : 1);
  const wineLimit = () => options.maxWines || (bonus() ? 3 : 2);
  const recipes = () =>
    [1, 2, 3]
      .map((b) => p.grapes.flatMap((g, i) => (draft.assignments[i] === b ? [i] : [])))
      .filter((r) => r.length);
  const recipeTypes = () =>
    [1, 2, 3]
      .filter((b) => Object.values(draft.assignments).includes(b))
      .map((b) => draft.recipeTypes[b] || '');
  const renderCharmat = () => {
    if (!p.buildings.includes('charmat')) return;
    const grid = group('第 ' + draft.bottle + ' 瓶 · 查玛法罐');
    for (const [value, label] of [
      ['', '常规配方'],
      ['sparkling', '红白各一 → 起泡'],
    ])
      choice(grid, 'recipeType', value, label, {
        selected: (draft.recipeTypes[draft.bottle] || '') === value,
        change: () => {
          draft.recipeTypes[draft.bottle] = value;
        },
      });
  };
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
    if (options.meta) button.append(cardText(options.meta, 'span', 'choice-meta'));
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
  const structureOwned = (id) =>
    (p.structureSlots || []).includes(id) || p.fields.some((f) => f.structure === id);
  const buildingPrice = ([id, , cost]) =>
    Math.max(0, cost - Number(bonus()) - Number(p.buildings.includes('workshop')));
  const buildingReason = ([id, , cost]) =>
    p.buildings.includes(id) || (structureBuildings.some((b) => b[0] === id) && structureOwned(id))
      ? '已建造'
      : id === 'large_cellar' && !p.buildings.includes('medium_cellar')
        ? '需要中酒窖'
        : p.coins < cost - Number(bonus()) - (p.buildings.includes('workshop') ? 1 : 0)
          ? '还差 ' +
            Math.max(
              0,
              cost - Number(bonus()) - (p.buildings.includes('workshop') ? 1 : 0) - p.coins,
            ) +
            ' 金币'
          : '';
  const cardChoice = (grid, name, original, options = {}) => {
    if (allowedCardIDs && !allowedCardIDs.includes(original.id)) return null;
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
    f.structure
      ? '已被结构占用，须先拆除'
      : mode === 'buy_field'
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
    const special = draft.workerType;
    const previousOrFuture = special && ['traveler', 'messenger'].includes(special);
    if (
      (!continuation &&
        (!view.legal.canPlace ||
          (space.season !== view.phase && space.season !== 'any' && !previousOrFuture))) ||
      (continuation && (!view.legal.canChoose || view.pendingChoice?.playerId !== view.youId))
    )
      return '等待你的行动回合';
    if (!continuation && space.disabledReason) return space.disabledReason;
    if (!continuation) {
      if (draft.workerType) {
        const used = p.specialWorkerUsed?.[draft.workerType];
        const ready = (p.specialWorkerReady?.[draft.workerType] || 0) <= (view.year || 0);
        if (!p.specialWorkers?.includes(draft.workerType) || used || !ready)
          return '该特殊工人未待命';
      } else if (draft.worker === 'large' && !p.largeWorker) return '大工人已使用';
      if (!draft.workerType && !p.workers && !p.largeWorker) return '没有待命工人，仅查看条件';
      if (placementsFor() && !chosenPlacement()) return '此工人无法进入所选行动格';
      if (!placementsFor() && draft.slot > 0 && !freeSlots(space).includes(draft.slot))
        return '行动格已占用';
      if (futureReservation()) return '';
      if (
        kind() === 'sell_wine' &&
        (draft.wines.length !== 1 || !p.wines.some((w) => w.id === draft.wines[0]))
      )
        return '选择一瓶自有葡萄酒';
      if (
        !draft.workerType &&
        draft.worker === 'normal' &&
        (!p.workers || (!privateSpace(space) && !freeSlots(space).length))
      )
        return '需要大工人';
    }
    if (destroyAction) {
      if (!draft.building) return '选择要拆除的结构';
      return '';
    }
    if (space.id === 'train' && trainingReason(draft.specialWorker))
      return trainingReason(draft.specialWorker);
    if (structureAction) {
      if (!structure) return '结构行动定义不存在';
      if (structureID === 'guest_house' && draft.cards.length !== 2) return '选择2张访客牌';
      if (['cask', 'wine_bar'].includes(structureID) && draft.wines.length !== 1)
        return '选择1瓶酒';
      if (structureID === 'wine_cave' && (draft.wines.length < 1 || draft.wines.length > 2))
        return '选择1至2瓶酒';
      if (['wine_press', 'mixer'].includes(structureID) && !recipes().length)
        return '选择至少1瓶酒的葡萄配方';
      if (['wine_press', 'mixer'].includes(structureID)) {
        if (
          winePreview(p, recipes(), recipeTypes()).some((text) =>
            /配方不成立|无可用酒槽/.test(text),
          )
        )
          return '请调整配方或检查酒窖';
        if (structureID === 'mixer') {
          const types = recipes().map((recipe, i) => {
            const red = recipe.filter((index) => p.grapes[index].color === 'red').length;
            const white = recipe.length - red;
            return red === 1 && white === 1
              ? recipeTypes()[i] === 'sparkling'
                ? 'sparkling'
                : 'blush'
              : red === 2 && white === 1
                ? 'sparkling'
                : '';
          });
          if (types.some((type) => !type) || new Set(types).size !== types.length)
            return '至多酿1瓶桃红和1瓶起泡酒';
        }
      }
      if (structureID === 'school') {
        if (p.totalWorkers >= 6) return '工人已达6名上限';
        const fee = view.players.filter(
          (other) => other.id !== p.id && other.buildings.includes('academy'),
        ).length;
        if (p.coins < fee) return '还需 ' + (fee - p.coins) + ' 金币缴纳学院费';
      }
      if (['shop', 'label_factory'].includes(structureID) && !draft.cardId)
        return '选择1张订单和对应的酒';
      if (structureID === 'ristorante' && (draft.wines.length !== 1 || !draft.grapeId))
        return '选择1瓶酒和1颗红/白葡萄';
      if (structureID === 'cafe' && !draft.grapeId) return '选择1颗红/白葡萄';
      if (structureID === 'trading_post' && resourceState().reason) return resourceState().reason;
      return '';
    }
    if (['influence', 'trade'].includes(space.id)) return resourceState().reason;

    if (kind() === 'build') {
      if (!draft.building) return '选择一座建筑';
      if (structureBuildings.some((b) => b[0] === draft.building)) {
        const card = view.hand.find(
          (c) => c.type === 'structure' && c.structureId === draft.building,
        );
        if (!card) return '需先从结构牌堆抽取并持有该结构牌';
        if (!draft.structureTarget) return '选择结构垫空位或空田地';
        if (draft.structureTarget.startsWith('field:')) {
          const field = p.fields[Number(draft.structureTarget.slice(6))];
          if (!field || field.sold || field.vines.length || field.structure)
            return '只能选择未出售且空的田地';
        } else if (draft.structureTarget.startsWith('mat:')) {
          const slot = Number(draft.structureTarget.slice(4));
          if ((p.structureSlots || [])[slot]) return '结构垫该位置已占用';
        }
        return buildingReason(buildings.find((b) => b[0] === draft.building));
      }
      return buildingReason(buildings.find((b) => b[0] === draft.building));
    }
    if (['summer_visitor', 'winter_visitor', 'fill_order'].includes(kind())) {
      if (!draft.cardId) return '选择一张手牌';
      if (cardReason(draft.cardId)) return cardReason(draft.cardId);
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
        if (field.structure) return '田地已被结构占用';
        if (plantRoom(field, card) < card.red + card.white) return '田地总容量不足';
      }
    }
    if (kind() === 'harvest' && space.id !== 'yoke') {
      if (draft.mode === 'all') return '';
      if (!draft.fields.length) return '选择要收获的田地';
      if (draft.fields.length > limit()) return '本次最多收获 ' + limit() + ' 块田地';
    }
    if (kind() === 'harvest' && draft.harvestAll)
      return p.fields.some((f) => !f.sold && !f.harvested && f.vines.length)
        ? ''
        : '没有可收获田地';
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
      if (Object.values(draft.assignments).some((b) => b > wineLimit()))
        return '本次最多酿造 ' + wineLimit() + ' 瓶';
      if (winePreview(p, r, recipeTypes()).some((s) => /配方不成立|无可用酒槽/.test(s)))
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
        choice(grid, 'color', value, label, {
          image: cardArt({ type: value }, 'choice-art'),
        });
    } else {
      if (!continuation) {
        const workers = group('派遣工人', true);
        for (const id of p.specialWorkers || []) {
          const def = (view.specialWorkerCatalog || []).find((w) => w.id === id);
          const ready =
            (p.specialWorkerReady?.[id] || 0) <= (view.year || 0) && !p.specialWorkerUsed?.[id];
          choice(workers, 'workerType', id, def?.name || id, {
            image: worker(view, p.id, false),
            meta: ready ? '可用 · 特殊能力' : '本年不可用',
            reason: !ready
              ? '尚未可用或本年已使用'
              : placementsFor(id, false)?.length === 0
                ? '无法进入此行动格'
                : '',
            tip: def?.description || '特殊工人',
            change: () => {
              draft.workerType = id;
              draft.worker = 'normal';
              draft.slot = 0;
            },
          });
        }
        choice(workers, 'worker', 'normal', '普通工人', {
          selected: draft.worker === 'normal' && !draft.workerType,
          image: worker(view, p.id, false),
          meta: '剩余 ' + (p.workers - Number(!!p.grayWorkerAvailable)),
          reason: !(p.workers - Number(!!p.grayWorkerAvailable))
            ? '已用完'
            : placementsFor('', false, false)?.length === 0 ||
                (!placementsFor('', false, false) &&
                  !privateSpace(space) &&
                  !freeSlots(space).length)
              ? '普通格已满'
              : '',
          tip: '派遣一名普通工人，占用空闲行动格。',
          change: () => {
            draft.worker = 'normal';
            draft.workerType = '';
            draft.slot = 0;
          },
        });
        if (p.grayWorkerAvailable)
          choice(workers, 'worker', 'gray', '灰色临时工', {
            selected: draft.worker === 'gray' && !draft.workerType,
            meta: '本年可用',
            reason: placementsFor('', false, true)?.length === 0 ? '无可用行动格' : '',
            tip: '临时工，本年结束归还；不能用于牺牲永久工人的访客。',
            change: () => {
              draft.worker = 'gray';
              draft.workerType = '';
              draft.slot = 0;
            },
          });
        choice(workers, 'worker', 'large', '大工人', {
          selected: draft.worker === 'large' && !draft.workerType,
          image: worker(view, p.id, true),
          meta: '剩余 ' + Number(p.largeWorker),
          reason: !p.largeWorker
            ? '已使用'
            : placementsFor('', true, false)?.length === 0
              ? '无法进入此行动格'
              : '',
          tip: '普通格已满时仍可进入；溢出位置没有行动格奖励。',
          change: () => {
            draft.worker = 'large';
            draft.workerType = '';
            draft.slot = 0;
          },
        });
        if (space.capacity >= 1 && !privateSpace(space)) {
          const slots = group('行动格', true);
          choice(slots, 'slot', 0, '自动', {
            meta: freeSlots(space).length ? '首个空格' : '大工人位',
          });
          for (let i = 1; i <= placementSpace().capacity; i++)
            choice(slots, 'slot', i, '第 ' + i + ' 格', {
              meta: bonusLabel(placementSpace(), i),
              reason: placementsFor()
                ? placementsFor().some((item) => item.slot === i)
                  ? ''
                  : '此工人不可进入'
                : freeSlots(space).includes(i)
                  ? ''
                  : '已占用',
            });
          if (
            !futureReservation() &&
            bonusKey(placementSpace(), chosenPlacement()?.slot ?? draft.slot)
          ) {
            const rewards = group('行动格奖励', true);
            choice(
              rewards,
              'declineBonus',
              'no',
              '★ ' + bonusLabel(placementSpace(), chosenPlacement()?.slot ?? draft.slot),
            );
            choice(rewards, 'declineBonus', 'yes', '放弃奖励');
          }
        }
      }
      if (futureReservation()) {
        const reserved = node('section', null, 'action-choice-section');
        reserved.append(
          node('h4', '预定未来行动'),
          node(
            'p',
            '现在只锁定信使的位置；到该季第一次轮到你时，再按当时资源完成行动并选择格奖励。',
          ),
        );
        content.append(reserved);
      } else {
        if (continuation && options.continuationBonus) {
          const rewards = group('预约行动格奖励', true);
          choice(
            rewards,
            'declineBonus',
            'no',
            '★ ' + bonusLabel(space, options.continuationSlot || 0),
          );
          choice(rewards, 'declineBonus', 'yes', '放弃奖励');
        }
        if (space.id === 'train' && (view.config?.specialWorkers || p.grandeRemoved)) {
          const training = group('培训类别', true);
          for (const id of [
            'regular',
            ...(p.grandeRemoved ? ['grande'] : []),
            ...(view.config?.specialWorkers ? view.specialWorkerPool || [] : []),
          ]) {
            const def = (view.specialWorkerCatalog || []).find((w) => w.id === id);
            choice(
              training,
              'specialWorker',
              id,
              id === 'regular'
                ? '普通工人'
                : id === 'grande'
                  ? '重新培训大工人'
                  : '特殊工人 · ' + (def?.name || id),
              {
                meta:
                  trainingPrice(id) +
                  ' 金币' +
                  (farmerTrainingDiscount() ? '（需选农夫折扣）' : '') +
                  (placementToll() ? '（含通行费' + placementToll() + '）' : '') +
                  ' · 次年可用',
                reason: trainingReason(id),
                tip:
                  id === 'regular'
                    ? '培训一名普通工人。'
                    : (def?.description || '特殊工人') + '；培训费用比普通工人多1金币。',
              },
            );
          }
        }
        if (printedBonus() === 'coin' && ['summer_visitor', 'winter_visitor'].includes(space.id)) {
          const order = group('金币奖励结算顺序', true);
          choice(order, 'bonusFirst', 'before', '先领1金币，再打访客');
          choice(order, 'bonusFirst', 'after', '先打访客，再领1金币');
        }
        if (space.id === 'harvest' && p.buildings.includes('harvest_machine')) {
          const modes = group('收获方式', true);
          choice(modes, 'mode', 'harvest', '选择田地', {
            selected: draft.mode !== 'all',
            meta: '按当前行动格限制数量',
            tip: '选择要收获的田地；按当前行动格的收获数量限制结算。',
          });
          choice(modes, 'mode', 'all', '收割机：全部田地', {
            selected: draft.mode === 'all',
            meta: '收获所有可收获田地',
            tip: '同时收获所有未出售、未收获且有葡萄藤的田地。',
          });
        } else if (draft.mode) {
          const modes = group('操作', true);
          const options =
            space.id === 'build_tour'
              ? [
                  ['build', '建造固定建筑'],
                  ['tour', '导览获得2金币'],
                ]
              : space.id === 'flip_field'
                ? [
                    ['sell_field', '出售空田'],
                    ['buy_field', '买回田地'],
                  ]
                : space.id === 'yoke'
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
              reason: ['build', 'tour'].includes(value)
                ? ''
                : value === 'sell_grapes'
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
        if (['influence', 'trade'].includes(space.id)) {
          renderTuscanyInputs(content, resourceState(), draft.tuscany, true, (key) => {
            redraw();
            content
              .querySelector('[data-tuscany-input="' + key + '"]')
              ?.focus({ preventScroll: true });
          });
          preview.textContent =
            space.id === 'influence'
              ? '首次放星领奖；六星用尽后移动不领奖。额外星另行结算。'
              : '完成一次交易；如有额外交易，使用更新后的资源继续选择。';
        } else if (kind() === 'sell_wine') {
          const grid = group('出售一瓶葡萄酒');
          for (const w of p.wines)
            choice(grid, 'wines', w.id, names[w.type] + '酒', {
              image: art('wine', 'resource-art'),
              meta:
                '品质' +
                w.value +
                ' · 获得' +
                { red: 1, white: 1, blush: 2, sparkling: 4 }[w.type] +
                '分',
              selected: draft.wines.includes(w.id),
              change: () => {
                draft.wines = [w.id];
              },
            });
        } else if (destroyAction) {
          const grid = group('选择要拆除的结构');
          const ids = [...(p.structureSlots || []), ...p.fields.map((f) => f.structure)].filter(
            Boolean,
          );
          for (const id of ids) {
            const b = structureBuildings.find((item) => item[0] === id);
            choice(grid, 'building', id, b?.[1] || id, {
              image: art('estate', 'choice-art'),
              meta: '拆除后不返还建造分',
            });
          }
        } else if (structureAction) {
          if (structureID === 'guest_house') {
            const cards = group('选择2张访客 · 再点一次取消');
            for (const card of view.hand.filter((c) => ['summer', 'winter'].includes(c.type)))
              cardChoice(cards, 'cards', card, {
                selected: draft.cards.includes(card.id),
                change: () => {
                  draft.cards = toggle(draft.cards, card.id);
                },
              });
          }
          if (structureID === 'trading_post') {
            renderTuscanyInputs(content, resourceState(), draft.tuscany, true, (key) => {
              redraw();
              content
                .querySelector('[data-tuscany-input="' + key + '"]')
                ?.focus({ preventScroll: true });
            });
          }
          if (['cask', 'wine_cave', 'wine_bar', 'ristorante'].includes(structureID)) {
            const wines = group('选择葡萄酒');
            for (const w of p.wines)
              choice(wines, 'wines', w.id, names[w.type] + '酒', {
                image: art('wine', 'resource-art'),
                meta: '品质 ' + w.value,
                selected: draft.wines.includes(w.id),
                change: () => {
                  draft.wines = ['cask', 'wine_bar', 'ristorante'].includes(structureID)
                    ? [w.id]
                    : toggle(draft.wines, w.id);
                },
              });
          }
          if (['cafe', 'ristorante'].includes(structureID)) {
            const grapes = group('选择红/白葡萄');
            p.grapes.forEach((g) =>
              choice(grapes, 'grapeId', g.id, names[g.color] + '葡萄', {
                image: art('grape-' + g.color, 'resource-art'),
                meta: '品质 ' + g.value,
                selected: draft.grapeId === g.id,
                change: () => {
                  draft.grapeId = g.id;
                },
              }),
            );
          }
          if (['shop', 'label_factory'].includes(structureID)) {
            const orders = group('选择订单');
            for (const c of view.hand.filter((c) => c.type === 'order'))
              cardChoice(orders, 'cardId', c);
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
          if (['wine_press', 'mixer'].includes(structureID)) {
            const bottles = group('选择第几瓶');
            for (let b = 1; b <= 2; b++)
              choice(bottles, 'bottle', b, '第 ' + b + ' 瓶', {
                selected: draft.bottle === b,
              });
            renderCharmat();
            const grapes = group('选择酿酒配方');
            p.grapes.forEach((g, i) =>
              choice(grapes, 'grapes', i, names[g.color] + '葡萄', {
                image: art('grape-' + g.color, 'resource-art'),
                meta: '品质 ' + g.value,
                selected: Object.prototype.hasOwnProperty.call(draft.assignments, i),
                change: () => {
                  const used = Object.entries(draft.assignments).find(([key]) => Number(key) === i);
                  if (used) delete draft.assignments[used[0]];
                  else draft.assignments[i] = draft.bottle;
                },
              }),
            );
            preview.textContent =
              structureID === 'mixer' ? '调酒器至多酿1瓶桃红和1瓶起泡酒。' : '压酒机最多酿造2瓶。';
          }
        } else if (kind() === 'build') {
          const grid = group('选择建筑');
          for (const b of availableBuildings(view)) {
            const [id, label, cost, , effect] = b,
              price = buildingPrice(b),
              reason = buildingReason(b),
              owned = p.buildings.includes(id);
            choice(grid, 'building', id, label, {
              cls: 'building-choice',
              image: buildingCardArt(id, 'choice-art', label),
              meta: price + ' 金币' + (cost !== price ? ' · 已减 ' + (cost - price) : ''),
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
                (cost !== price ? '，减免 ' + (cost - price) : '') +
                '）\n' +
                (owned ? '已建造' : '尚未建造') +
                (reason && !owned ? ' · ' + reason : ''),
              change: () => {
                draft.building = id;
                if (structureBuildings.some((item) => item[0] === id)) {
                  const first = (p.structureSlots || []).findIndex((slot) => !slot);
                  draft.structureTarget = first >= 0 ? 'mat:' + first : '';
                } else draft.structureTarget = '';
              },
            });
          }
          if (structureBuildings.some((b) => b[0] === draft.building)) {
            const locations = group('结构放置位置');
            for (let i = 0; i < 2; i++)
              choice(locations, 'structureTarget', 'mat:' + i, '结构垫 · 第' + (i + 1) + '格', {
                selected: draft.structureTarget === 'mat:' + i,
                reason: (p.structureSlots || [])[i] ? '已占用' : '',
              });
            for (const f of p.fields)
              choice(
                locations,
                'structureTarget',
                'field:' + f.index,
                '空田地 · 第' + (f.index + 1) + '块',
                {
                  selected: draft.structureTarget === 'field:' + f.index,
                  reason: f.sold ? '已出售' : f.vines.length || f.structure ? '田地不为空' : '',
                },
              );
          }
        } else if (['summer_visitor', 'winter_visitor', 'fill_order'].includes(kind())) {
          const type =
            kind() === 'fill_order' ? 'order' : kind() === 'summer_visitor' ? 'summer' : 'winter';
          const grid = group(type === 'order' ? '选择订单' : '选择要打出的访客');
          for (const c of view.hand.filter(
            (c) =>
              c.type === type ||
              (view.config?.visitors === 'ee_moor' &&
                type === 'summer' &&
                c.id === 'moor-winter-10'),
          ))
            cardChoice(grid, 'cardId', c, { reason: cardReason(c.id) });
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
                      : f.structure
                        ? '结构占用'
                        : '空余 ' +
                          f.free +
                          (f.free >= requirements.value
                            ? ' ✓'
                            : ' · 差 ' + (requirements.value - f.free))),
                  !f.sold && !f.structure && f.free >= requirements.value ? 'met' : 'missing',
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
                reason: f.sold
                  ? '已出售'
                  : f.structure
                    ? '结构占用，须先拆除'
                    : plantRoom(f, c) < c.red + c.white
                      ? '容量不足'
                      : '',
                change: () => {
                  draft.plantFields[id] = f.index;
                },
              });
          }
        } else if (kind() === 'harvest' && space.id === 'harvest' && draft.mode === 'all') {
          preview.textContent = '将同时收获所有未出售、未收获且有葡萄藤的田地。';
        } else if (['buy_field', 'sell_field', 'uproot', 'harvest'].includes(kind())) {
          if (
            space.id === 'yoke' &&
            kind() === 'harvest' &&
            p.buildings.includes('harvest_machine')
          ) {
            const modes = group('收割机');
            choice(modes, 'harvestAll', false, '选择田地', {
              selected: !draft.harvestAll,
            });
            choice(modes, 'harvestAll', true, '收获所有田地', {
              selected: draft.harvestAll,
            });
          }
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
          const bottles = group('先选酒瓶，再添加葡萄 · 最多 ' + wineLimit() + ' 瓶', true);
          for (let b = 1; b <= (options.maxWines || 3); b++)
            choice(bottles, 'bottle', b, '第 ' + b + ' 瓶', {
              meta: Object.values(draft.assignments).filter((v) => v === b).length + ' 颗葡萄',
              reason:
                b === 3 && !bonus() && !Object.values(draft.assignments).includes(3)
                  ? '需要奖励格'
                  : '',
              tip: '单颗红／白酿红／白酒；红白各一酿桃红；两红一白酿起泡。',
            });
          renderCharmat();
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
          winePreview(p, recipes(), recipeTypes()).forEach((text, i) =>
            preview.append(
              node('p', text.replace(/^第 \d+ 瓶：/, '第 ' + bottlesUsed[i] + ' 瓶：')),
            ),
          );
        } else {
          preview.textContent = space.description;
        }
      }
    }
    const reason = invalid();
    readiness.textContent = reason || '✓ 已就绪';
    readiness.classList.toggle('ready', !reason);
    send.disabled = !!reason || session.sending;
    document.dispatchEvent(
      new CustomEvent('action-panel-state', {
        detail: {
          space: wake ? 'wake' : space.id,
          reason,
          target: ['influence', 'trade'].includes(space.id) ? resourceState().target : null,
        },
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
          gray: draft.worker === 'gray',
          workerType: draft.workerType,
          slot: continuation
            ? options.continuationSlot || 0
            : (chosenPlacement()?.slot ?? draft.slot),
          declineBonus: draft.declineBonus === 'yes',
        };
    action.revision = view.revision;
    if (!wake && !futureReservation()) {
      if (draft.mode) action.mode = draft.mode;
      if (draft.field != null) action.field = draft.field;
      if (draft.structureTarget?.startsWith('field:')) {
        action.mode = 'field';
        action.field = Number(draft.structureTarget.slice(6));
      }
      if (draft.structureTarget?.startsWith('mat:')) {
        action.mode = 'mat';
        action.structureSlot = Number(draft.structureTarget.slice(4));
      }
      if (draft.cardId) action.cardId = draft.cardId;
      if (draft.building) action.building = draft.building;
      if (space.id === 'train') action.specialWorker = draft.specialWorker;
      if (kind() === 'plant') {
        action.cardIds = draft.cards;
        action.fields = draft.cards.map((id) => draft.plantFields[id]);
      }
      if (kind() === 'harvest' && space.id !== 'yoke') action.fields = draft.fields;
      if (kind() === 'sell_grapes') action.grapes = draft.grapes;
      if (kind() === 'make_wine') {
        action.recipes = recipes();
        action.recipeTypes = recipeTypes();
      }
      if (draft.harvestAll) action.harvestAll = true;
      if (['fill_order', 'sell_wine'].includes(kind())) action.wineIds = draft.wines;
      if (structureAction) {
        if (structureID === 'guest_house') action.cardIds = draft.cards;
        if (
          ['cask', 'wine_cave', 'wine_bar', 'ristorante', 'shop', 'label_factory'].includes(
            structureID,
          )
        )
          action.wineIds = draft.wines;
        if (['cafe', 'ristorante'].includes(structureID)) action.grapeId = draft.grapeId;
        if (['wine_press', 'mixer'].includes(structureID)) {
          action.recipes = recipes();
          action.recipeTypes = recipeTypes();
        }
        if (structureID === 'trading_post') Object.assign(action, resourceState().payload);
      }
      if (destroyAction && draft.building) action.building = draft.building;
      if (['influence', 'trade'].includes(space.id)) Object.assign(action, resourceState().payload);
      if (printedBonus() === 'coin') action.bonusFirst = draft.bonusFirst === 'before';
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
