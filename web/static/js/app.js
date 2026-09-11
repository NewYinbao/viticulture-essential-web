import { openActionPanel, closeActionPanel, hasActionPanel } from './action-panel.js';
import { actionReason } from './action-options.js';
import { hint, installHints } from './hints.js';
import { localCard } from './card-i18n.js';
import { renderEE, cardArt, setupEE } from './ee-ui.js';
import { art, worker, decorateTable, renderEstate, actionArt } from './graphics.js';
const $ = (s) => document.querySelector(s);
installHints();
const el = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
const types = {
  vine: '葡萄藤',
  order: '订单',
  summer: '夏季访客',
  winter: '冬季访客',
  red: '红',
  white: '白',
  blush: '桃红',
  sparkling: '起泡',
};
const buildings = {
  trellis: '棚架 · 2金币',
  irrigation: '灌溉 · 3金币',
  medium_cellar: '中酒窖 · 4金币',
  large_cellar: '大酒窖 · 6金币',
  cottage: '小屋 · 4金币',
  windmill: '磨坊 · 5金币',
  tasting_room: '品酒室 · 6金币',
  yoke: '轭 · 2金币',
};
const phases = {
  setup: '开局 · 家族传承',
  fall: '秋季 · 访客到来',
  year_end: '年末 · 整理手牌',
  lobby: '好友集结',
  wake: '春日 · 起床顺序',
  summer: '夏季 · 葡萄园',
  winter: '冬季 · 酿酒坊',
  finished: '收官 · 年度佳酿',
};
let token = localStorage.getItem('vineyard-ee-token') || '',
  state = null,
  source = null,
  busy = false,
  online = false;
let handFilter = 'all';
let toastTimer;
function toast(text) {
  $('#toast').textContent = text;
  $('#toast').hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => ($('#toast').hidden = true), 4500);
}
async function api(path, body) {
  const opts = { headers: {} };
  if (body !== undefined) {
    opts.method = 'POST';
    opts.headers['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(body);
  }
  if (token) opts.headers.Authorization = 'Bearer ' + token;
  const r = await fetch(path, opts);
  const data = await r.json();
  if (!r.ok) {
    const error = new Error(data.error || '请求失败');
    error.status = r.status;
    throw error;
  }
  return data;
}
function setConnection(text) {
  $('#connection').textContent = text;
}
function myPlayer() {
  return state.players.find((p) => p.id === state.youId);
}
function playerName(id) {
  return state.players.find((p) => p.id === id)?.name || '—';
}
function render(v) {
  if (state && v.revision !== state.revision && !busy) {
    if (hasActionPanel()) {
      closeActionPanel({ restoreFocus: false });
      toast('局面已更新，请重新选择行动。');
    }
  }
  const focusedCard = document.activeElement?.dataset.cardId;
  for (const s of v.spaces || []) s.occupied ??= [];
  for (const p of v.players || []) {
    p.buildings ??= [];
    p.grapes ??= [];
    p.wines ??= [];
    p.fields ??= [];
    for (const f of p.fields) f.vines ??= [];
  }
  state = v;
  decorateTable(v);
  $('#welcome').hidden = true;
  $('#game').hidden = false;
  $('#season-title').textContent =
    (v.phase === 'lobby' ? '' : `第 ${v.year} 年 · `) + phases[v.phase];
  $('#code-label').textContent = '房间 ' + v.code;
  $('#turn-label').textContent =
    v.phase === 'finished'
      ? '最高排名：' + playerName(v.winnerId)
      : v.turnId
        ? v.turnId === v.youId
          ? '轮到你行动'
          : '等待 ' + playerName(v.turnId) + ' 行动'
        : '房主：' + playerName(v.hostId);
  $('#status').textContent =
    v.phase === 'lobby'
      ? '分享邀请链接和房间码，等待好友加入。'
      : v.turnId === v.youId
        ? v.pendingChoice
          ? '请先完成下方选择。'
          : v.phase === 'wake'
            ? '轮到你选择起床顺序。'
            : `轮到你 · 普通工人 ${myPlayer().workers} / 大工人 ${myPlayer().largeWorker ? 1 : 0}`
        : '';
  $('#lobby-panel').hidden = v.phase !== 'lobby';
  $('#start-game').disabled = !v.legal.canStart || !online;
  $('#wake-panel').hidden = v.phase !== 'wake';
  $('#board').hidden = !['summer', 'winter'].includes(v.phase);
  $('#pass').disabled = !v.legal.canPass || !online;
  $('#hand-panel').hidden = v.phase === 'lobby';
  const wake = $('#wake-options');
  wake.replaceChildren();
  for (const s of v.wakeSlots || []) {
    const b = el('button');
    b.append(
      el('strong', s.slot),
      el('span', s.bonus),
      el('span', s.playerId ? playerName(s.playerId) : '选择'),
    );
    b.disabled = !!s.playerId || !v.legal.canWake || !online;
    if (s.playerId) b.prepend(worker(v, s.playerId, false));
    b.dataset.slot = s.slot;
    b.onclick = () =>
      s.slot === 5 ? setupEE('wake', state, act) : act({ type: 'wake', slot: s.slot });
    wake.append(b);
  }
  const board = $('#spaces');
  board.replaceChildren();
  for (const s of v.spaces || []) {
    if (s.season !== v.phase && s.season !== 'any') continue;
    const b = el('button', null, 'space');
    b.dataset.space = s.id;
    b.append(art(actionArt[s.id] || 'estate', 'action-art'));
    b.append(el('strong', s.name), el('span', s.description, 'desc'));
    const seats = el('div', null, 'worker-slots');
    const occupied =
      s.id === 'yoke' ? s.occupied.filter((o) => o.playerId === v.youId) : s.occupied;
    const drawSeat = (o, label, overflow = false) => {
      const seat = el('span', null, 'worker-slot' + (overflow ? ' overflow-seat' : ''));
      if (o) seat.append(worker(v, o.playerId, o.large));
      else seat.textContent = label;
      seat.title = o ? playerName(o.playerId) + (o.large ? ' · 大工人' : ' · 普通工人') : label;
      seat.setAttribute('aria-label', seat.title);
      seats.append(seat);
    };
    if (s.id !== 'gain_coin')
      for (let slot = 1; slot <= s.capacity; slot++)
        drawSeat(
          occupied.find((o) => o.slot === slot),
          s.capacity >= 2 && slot === 1 ? '奖励' : '空',
        );
    for (const o of occupied.filter((o) => o.slot <= 0)) drawSeat(o, '', true);
    b.append(seats);
    const normal = occupied.filter((o) => o.slot > 0).length;
    b.append(
      el(
        'span',
        s.id === 'gain_coin'
          ? '不限人数 · 每位工人获得 1 金币'
          : s.id === 'yoke'
            ? '自家行动 · 每年一次'
            : `普通格 ${normal}/${s.capacity}` + (s.capacity >= 2 ? ' · 第 1 格有奖励' : ''),
        'slots',
      ),
    );
    const reason = actionReason(s, v);
    b.disabled = !v.legal.canPlace || !online;
    const ready = v.legal.canPlace && online && !reason;
    b.classList.toggle('action-ready', ready);
    b.append(el('span', reason ? '条件不足' : ready ? '可派遣' : '等待', 'availability-badge'));
    hint(b, reason || (ready ? '选择工人与资源' : '等待你的行动回合'), !!reason);
    const cardType = {
      plant: 'vine',
      fill_order: 'order',
      summer_visitor: 'summer',
      winter_visitor: 'winter',
    }[s.id];
    const focusCards = (on) =>
      document
        .querySelectorAll('#hand .card')
        .forEach((card) =>
          card.classList.toggle(
            'action-linked',
            on && !!cardType && card.dataset.cardType === cardType,
          ),
        );
    b.onpointerenter = () => focusCards(true);
    b.onpointerleave = () => focusCards(false);
    b.onfocus = () => focusCards(true);
    b.onblur = () => focusCards(false);
    b.onclick = () => {
      if (ready) openAction(s);
    };
    board.append(b);
  }
  const choosingCards = !!v.pendingChoice && v.pendingChoice.playerId === v.youId;
  if (choosingCards) handFilter = 'all';
  const filters = $('#hand-filters');
  filters.replaceChildren();
  for (const [type, label] of [
    ['all', '全部'],
    ['vine', '葡萄藤'],
    ['order', '订单'],
    ['summer', '夏访客'],
    ['winter', '冬访客'],
  ]) {
    const count = (v.hand || []).filter((c) => type === 'all' || c.type === type).length;
    const button = el('button', label + ' ' + count);
    button.type = 'button';
    button.dataset.filter = type;
    button.setAttribute('aria-pressed', String(handFilter === type));
    button.disabled = choosingCards && type !== 'all';
    button.onclick = () => {
      handFilter = type;
      render(state);
    };
    filters.append(button);
  }
  const hand = $('#hand');
  hand.replaceChildren();
  for (const original of v.hand || []) {
    const c = localCard(original);
    const e = el('article', null, 'card ' + c.type);
    e.dataset.cardId = c.id;
    e.dataset.cardType = c.type;
    e.hidden = handFilter !== 'all' && c.type !== handFilter;
    e.append(cardArt(c, 'card-art'));
    e.append(
      el('span', types[c.type] || c.type, 'type'),
      el('h4', c.name),
      el('small', c.englishName, 'card-english'),
      el('p', c.description),
    );
    if (c.type === 'vine') e.append(el('b', `红 ${c.red} · 白 ${c.white}`));
    if (c.type === 'order')
      e.append(
        el('b', (c.requirements || []).map((w) => `${types[w.type]}酒 ≥ ${w.value}`).join(' + ')),
        el('p', `奖励 ${c.points}分 · 收入 +${c.income}`),
      );
    if (['summer', 'winter'].includes(c.type) && !c.implemented)
      e.append(el('small', '效果尚未实现 / 不可打出', 'ee-warning'));
    if (v.legal.canPlace && !choosingCards) {
      const actionID = {
        vine: 'plant',
        order: 'fill_order',
        summer: 'summer_visitor',
        winter: 'winter_visitor',
      }[c.type];
      const space = v.spaces.find((s) => s.id === actionID && s.season === v.phase);
      if (space) {
        const reason = v.cardReasons?.[c.id] || actionReason(space, v);
        const label = reason
          ? '条件不足'
          : { vine: '可种植', order: '可交付', summer: '可打出', winter: '可打出' }[c.type];
        e.classList.toggle('card-ready', !reason);
        const badge = el('span', label, 'card-state');
        badge.tabIndex = 0;
        hint(badge, reason || '在对应行动中选择此牌', !!reason);
        e.append(badge);
      }
    }
    hand.append(e);
  }
  if (!v.hand?.length) hand.append(el('p', '暂无手牌', 'empty'));
  const ps = $('#players');
  ps.replaceChildren();
  for (const p of [...v.players].sort((a, b) => (b.id === v.youId) - (a.id === v.youId)))
    ps.append(renderEstate(v, p, buildings, types));
  const logs = $('#log');
  logs.replaceChildren(...[...(v.log || [])].reverse().map((s) => el('li', s)));
  $('#rules-text').replaceChildren(...v.rulesNotes.map((n) => el('p', n)));
  renderEE(v, act, online);
  if (focusedCard) {
    const card = [...hand.children].find((e) => e.dataset.cardId === focusedCard);
    if (card?.tabIndex === 0) card.focus({ preventScroll: true });
  }
}
async function connect() {
  if (!token) return;
  try {
    const v = await api('/api/state?token=' + encodeURIComponent(token));
    online = true;
    render(v);
    source?.close();
    source = new EventSource('/api/events?token=' + encodeURIComponent(token));
    source.onopen = () => {
      online = true;
      setConnection('● 已连接本地服务器');
      if (state) render(state);
    };
    source.onmessage = (e) => {
      const next = JSON.parse(e.data),
        changed = !online || !state || state.code !== next.code || state.revision !== next.revision;
      online = true;
      if (changed) render(next);
    };
    source.onerror = () => {
      online = false;
      setConnection('○ 断线，正在自动重连…');
      if (state) render(state);
    };
  } catch (e) {
    online = false;
    setConnection('连接失败');
    toast(e.message);
  }
}
async function enter(create) {
  if (busy) return;
  const name = $('#nickname').value.trim(),
    code = $('#room-code').value.trim().toUpperCase();
  if (!name) return toast('请填写昵称');
  busy = true;
  try {
    const r = await api(create ? '/api/create' : '/api/join', create ? { name } : { name, code });
    token = r.token;
    localStorage.setItem('vineyard-ee-token', token);
    localStorage.setItem('vineyard-ee-name', name);
    await connect();
  } catch (e) {
    toast(e.message);
  } finally {
    busy = false;
  }
}
async function act(a) {
  if (busy) return;
  busy = true;
  try {
    const v = await api('/api/action', { revision: state.revision, ...a });
    render(v);
    return true;
  } catch (e) {
    toast(e.message);
    const error = $('#action-error');
    if (error) error.textContent = e.message;
    if (e.status === 409) {
      try {
        const latest = await api('/api/state');
        closeActionPanel({ restoreFocus: false });
        render(latest);
        toast('局面已经变化，已刷新。请按最新局面重新选择。');
      } catch {}
    }
    return false;
  } finally {
    busy = false;
  }
}
$('#create-room').onclick = () => enter(true);
$('#join-room').onclick = () => enter(false);
$('#start-game').onclick = () => act({ type: 'start' });
$('#pass').onclick = () => {
  if (
    confirm(
      state.phase === 'summer'
        ? `确定结束夏季？剩余 ${myPlayer().workers + (myPlayer().largeWorker ? 1 : 0)} 位工人将保留到冬季。`
        : '确定结束冬季？本年不再派遣工人，等待年末结算。',
    )
  )
    act({ type: 'pass' });
};
$('#switch-table').onclick = () => {
  if (
    !confirm('返回入口不会退出座位。原会话仍可刷新恢复；加入新房间会替换本浏览器保存的会话。继续？')
  )
    return;
  source?.close();
  closeActionPanel({ restoreFocus: false });
  $('#resume-room').hidden = !token;
  online = false;
  $('#welcome').hidden = false;
  $('#game').hidden = true;
  setConnection('已返回入口');
};
$('#invite').onclick = async () => {
  const u = new URL(location.href);
  u.search = 'room=' + state.code;
  try {
    await navigator.clipboard.writeText(u.href);
    toast('邀请链接已复制。请确保链接中的地址是服务器局域网IP。');
  } catch {
    prompt('复制邀请链接（如为localhost，请改为服务器局域网IP）', u.href);
  }
};
function openAction(s) {
  openActionPanel(s, state, act);
}
$('#resume-room').hidden = !token;
$('#resume-room').onclick = () => connect();
$('#nickname').value = localStorage.getItem('vineyard-ee-name') || '';
const invitation = new URL(location.href).searchParams.get('room');
if (invitation) $('#room-code').value = invitation.toUpperCase();
if (token && !invitation) connect();
