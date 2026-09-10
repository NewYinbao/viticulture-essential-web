export { setupEE } from './action-dialog.js';
import { renderVisitor } from './visitor-ui.js';
import { localCard } from './card-i18n.js';
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
export function cardArt(c, cls = '') {
  const e = n('img', null, cls);
  const k = parseInt(c.id?.split('-').pop() || '1', 10) || 1;
  e.src =
    '/art/handdrawn/' +
    (['summer', 'winter', 'mama', 'papa'].includes(c.type)
      ? 'portrait-' + ((k - 1) % 12)
      : c.type === 'vine'
        ? 'vine-' + ((k - 1) % 4)
        : 'order') +
    '.svg';
  e.alt = '原创手绘风格插画（部分卡牌共享主题图）';
  return e;
}
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
  for (const [type, c] of Object.entries(v.deckCounts || {}))
    counts.append(n('span', `${typeNames[type] || type} ${c.deck} / 弃 ${c.discard}`));
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
    const ok = await act({ type: 'choose', choiceId: c.id, revision: v.revision, ...a });
    if (!ok) {
      locked = false;
      for (const b of form.querySelectorAll('button')) b.disabled = !online;
      err.textContent = '提交未生效，请检查资源或局面是否已经变化。';
    }
  };
  if (c.kind === 'discard') {
    form.remove();
    err.remove();
    box.append(n('p', '请在下方原手牌区选择弃牌。'));
  } else
    for (const option of c.options || []) {
      let text =
        c.labels?.[option] ||
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
      if (c.kind === 'papa') {
        const o = v.parentOptions;
        text =
          option === 'gift'
            ? '接受赠礼：' + (gifts[o?.gift] || o?.gift)
            : '改为额外 ' + o?.coins + ' 金币';
      }
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
// Choice-scoped state survives SSE, not another room/player/choice.
let discard = null;
function renderDiscard(v, act, online) {
  const c = v.pendingChoice,
    own = c?.kind === 'discard' && c.playerId === v.youId;
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
  const title = n('h3', '年末需要弃牌');
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
  prompt.append(title, status, hint, send, error);
  $('#hand-panel').insertBefore(prompt, $('#hand'));
  const cards = [...document.querySelectorAll('#hand .card')];

  const update = () => {
    status.textContent =
      '需要弃置 ' + c.count + ' 张 · 已选 ' + d.ids.size + ' / ' + c.count + ' 张（保留不超过7张）';
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
      ok = await act({ type: 'choose', choiceId: c.id, revision: v.revision, cardIds: [...d.ids] });
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
