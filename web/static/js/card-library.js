import { localCard } from './card-i18n.js';
import { cardArt } from './card-art.js';
import { vineDescription } from './vine-art.js';
const types = {
  vine: '葡萄藤',
  order: '订单',
  summer: '夏访客',
  winter: '冬访客',
  mama: 'Mama',
  papa: 'Papa',
  structure: '建筑',
  worker: '特殊工人',
};
const node = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
let view = null,
  groups = null,
  loading = null,
  dialog,
  grid,
  summary,
  filters,
  search,
  scope,
  groupSelect,
  pager,
  message;
let category = 'all',
  page = 0,
  origin;
const pageSize = 30;
export function updateCardLibrary(next) {
  view = next;
  if (dialog?.open) render();
}
function included(card, group) {
  if (scope.value === 'all' || !view) return true;
  const c = view.config || {};
  if (group === 'structures') return !!c.structures;
  if (group === 'workers') return !!c.specialWorkers;
  if (group === 'moor') return c.visitors === 'ee_moor';
  if (group === 'rhine')
    return c.visitors === 'rhine' && (!card.requiresTuscany || c.board === 'tuscany');
  return !(c.visitors === 'rhine' && ['summer', 'winter'].includes(card.type));
}
function render() {
  if (!groups) return;
  const enabled = groups.flatMap((g) =>
    g.cards
      .filter((c) => included(c, g.id))
      .map((c) => ({ ...localCard(c), group: g.id, groupName: g.name })),
  );
  scope.querySelector('[value="current"]').disabled = !view;
  const query = search.value.trim().toLowerCase();
  const matching = enabled.filter(
    (c) =>
      (groupSelect.value === 'all' || groupSelect.value === c.group) &&
      (category === 'all' || c.type === category) &&
      (!query ||
        [c.name, c.englishName, c.description].some((s) =>
          (s || '').toLowerCase().includes(query),
        )),
  );
  summary.textContent = `${scope.value === 'current' && view ? '本局配置' : '全部已实现内容'}牌库总计 ${enabled.length} 张 · 筛选结果 ${matching.length} 张`;
  message.textContent = '固定牌库构成，含同名副本；不表示当前剩余，也不显示任何玩家的持牌。';
  filters.replaceChildren();
  for (const [key, label] of [['all', '全部'], ...Object.entries(types)]) {
    const count = enabled.filter(
      (c) =>
        (groupSelect.value === 'all' || c.group === groupSelect.value) &&
        (key === 'all' || c.type === key),
    ).length;
    if (!count && key !== 'all') continue;
    const b = node('button', `${label} ${count}`);
    b.type = 'button';
    b.setAttribute('aria-pressed', String(category === key));
    b.onclick = () => {
      category = key;
      page = 0;
      render();
      [...filters.children].find((n) => n.textContent === b.textContent)?.focus();
    };
    filters.append(b);
  }
  page = Math.min(page, Math.max(0, Math.ceil(matching.length / pageSize) - 1));
  grid.replaceChildren();
  for (const c of matching.slice(page * pageSize, (page + 1) * pageSize)) {
    const article = node('article', null, 'library-card card ' + c.type);
    article.dataset.cardId = c.id;
    const img = cardArt(c, 'library-art');
    img.loading = 'lazy';
    article.append(
      img,
      node('small', c.groupName + ' · ' + (types[c.type] || c.type), 'library-source'),
      node('h3', c.name),
    );
    if (c.englishName && c.englishName !== c.name) article.append(node('small', c.englishName));
    article.append(c.type === 'vine' ? vineDescription(c.description) : node('p', c.description));
    if (c.type === 'order') article.append(node('p', `奖励 ${c.points} 分 · 年收入 +${c.income}`));
    if (c.type === 'structure') article.append(node('p', `费用 ${c.structureCost} 金币`));
    if (c.requiresTuscany) article.append(node('small', '需要 Tuscany 主板', 'library-source'));
    grid.append(article);
  }
  if (!matching.length) grid.append(node('p', '没有匹配的卡牌。'));
  pager.replaceChildren();
  for (const [delta, label] of [
    [-1, '上一页'],
    [1, '下一页'],
  ]) {
    const b = node('button', label);
    b.type = 'button';
    b.disabled = delta < 0 ? page === 0 : (page + 1) * pageSize >= matching.length;
    b.onclick = () => {
      page += delta;
      render();
      grid.scrollIntoView({ block: 'start' });
    };
    pager.append(b);
  }
  pager.append(node('span', `${page + 1} / ${Math.max(1, Math.ceil(matching.length / pageSize))}`));
}
async function open(button) {
  origin = button;
  category = 'all';
  page = 0;
  search.value = '';
  groupSelect.value = 'all';
  scope.value = view && view.phase !== 'lobby' ? 'current' : 'all';
  dialog.showModal();
  search.focus();
  message.textContent = '正在读取卡牌图鉴…';
  try {
    if (!groups) {
      loading ||= fetch('/api/cards')
        .then(async (r) => {
          if (!r.ok) throw Error('卡牌图鉴暂不可用');
          return r.json();
        })
        .finally(() => (loading = null));
      groups = await loading;
    }
    if (dialog.open) render();
  } catch (e) {
    message.textContent = e.message;
  }
}
export function initCardLibrary() {
  dialog = node('dialog', null, 'card-library');
  dialog.id = 'card-library';
  dialog.setAttribute('aria-labelledby', 'card-library-title');
  const header = node('header', null, 'library-heading'),
    title = node('h2', '全卡牌图鉴');
  title.id = 'card-library-title';
  const close = node('button', '关闭');
  close.type = 'button';
  close.onclick = () => dialog.close();
  header.append(title, close);
  dialog.append(header);
  const controls = node('div', null, 'library-controls');
  search = node('input');
  search.type = 'search';
  search.placeholder = '搜索卡名、职业或效果';
  search.setAttribute('aria-label', '搜索卡牌');
  scope = node('select');
  scope.setAttribute('aria-label', '图鉴范围');
  scope.add(new Option('全部已实现卡牌', 'all'));
  scope.add(new Option('本局配置', 'current'));
  groupSelect = node('select');
  groupSelect.setAttribute('aria-label', '牌组');
  for (const [id, label] of [
    ['all', '所有牌组'],
    ['ee', 'Essential Edition'],
    ['moor', 'Moor'],
    ['rhine', 'Rhine'],
    ['structures', 'Tuscany 建筑'],
    ['workers', '特殊工人'],
  ])
    groupSelect.add(new Option(label, id));
  for (const e of [search, scope, groupSelect]) {
    e.addEventListener(e === search ? 'input' : 'change', () => {
      page = 0;
      category = 'all';
      render();
    });
    controls.append(e);
  }
  summary = node('p', null, 'library-summary');
  summary.setAttribute('aria-live', 'polite');
  message = node('p', null, 'library-note');
  filters = node('div', null, 'library-filters');
  grid = node('div', null, 'library-grid');
  pager = node('nav', null, 'library-pager');
  pager.setAttribute('aria-label', '图鉴分页');
  dialog.append(controls, summary, message, filters, grid, pager);
  document.body.append(dialog);
  dialog.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();
      dialog.close();
    }
  });
  dialog.addEventListener('close', () => origin?.isConnected && origin.focus());
  for (const button of document.querySelectorAll('[data-card-library]'))
    button.addEventListener('click', () => open(button));
}
