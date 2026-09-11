import { art } from './graphics.js';
import { actionReason, vineRequirements } from './action-options.js';
import { ruleTopics, actionTopics, actionLessons } from './rules-content.js';

const $ = (selector) => document.querySelector(selector);
const node = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
const read = (key) => {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
};
const save = (key, value) => {
  try {
    localStorage.setItem(key, value);
  } catch {}
};
let enabled = read('vineyard-ee-guide') === 'on';
let step = Math.max(0, Math.min(4, Number(read('vineyard-ee-tour-step')) || 0));
let state,
  connected = false,
  panelState,
  help,
  guide,
  returnFocus,
  topic = 'overview';

function button(text, action, cls) {
  const b = node('button', text, cls);
  b.type = 'button';
  b.onclick = action;
  return b;
}
function focusTarget(selector) {
  const target = $(selector);
  if (!target || !target.getClientRects().length) return;
  if (!target.matches('button, a, input, select, summary, [tabindex]')) target.tabIndex = -1;
  target.scrollIntoView({ block: 'center', behavior: 'instant' });
  target.focus({ preventScroll: true });
}
function clearHighlight() {
  document.querySelectorAll('.guide-target').forEach((e) => e.classList.remove('guide-target'));
}
function closeRules() {
  help.hidden = true;
  document
    .querySelectorAll('[data-rules-open]')
    .forEach((b) => b.setAttribute('aria-expanded', 'false'));
  if (returnFocus?.isConnected && returnFocus.getClientRects().length) {
    returnFocus.focus({ preventScroll: true });
    returnFocus.scrollIntoView({ block: 'nearest' });
  }
  paintGuide();
}
function showRules(id = 'overview', trigger = document.activeElement) {
  if (help.hidden) returnFocus = trigger;
  topic = ruleTopics.some((t) => t.id === id) ? id : 'overview';
  const anchor = $('#game').hidden ? $('#welcome') : $('#action-area');
  anchor.before(help);
  help.hidden = false;
  paintRules();
  document
    .querySelectorAll('[data-rules-open]')
    .forEach((b) => b.setAttribute('aria-expanded', 'true'));
  clearHighlight();
  help.scrollIntoView({ block: 'start' });
  $('#rules-help-title').focus({ preventScroll: true });
}
function paintRules() {
  help.replaceChildren();
  const head = node('div', null, 'help-heading');
  const title = node('h3', '规则说明 · EE 本体');
  title.id = 'rules-help-title';
  title.tabIndex = -1;
  head.append(title, button('返回游戏 ×', closeRules, 'help-close'));
  const nav = node('nav', null, 'rule-topics');
  nav.setAttribute('aria-label', '规则主题');
  for (const item of ruleTopics) {
    const b = button(item.title, () => {
      topic = item.id;
      paintRules();
      help.querySelector('[data-topic="' + topic + '"]').focus({ preventScroll: true });
    });
    b.dataset.topic = item.id;
    b.setAttribute('aria-pressed', String(topic === item.id));
    nav.append(b);
  }
  const item = ruleTopics.find((t) => t.id === topic);
  const intro = node('div', null, 'rule-intro');
  intro.append(art(item.art, 'rule-art'), node('p', item.intro));
  const cards = node('div', null, 'rule-cards');
  for (const [title, text] of item.cards) {
    const card = node('article');
    card.append(node('h4', title), node('p', text));
    cards.append(card);
  }
  const foot = node('div', null, 'help-foot');
  const sources = node('span', '规则来源：');
  for (const [label, url] of [
    [
      'EE 说明书',
      'https://shared.steamstatic.com/store_item_assets/steam/apps/414235/manuals/Viticulture_EE_Rules.pdf?t=1551153655',
    ],
    ['官方 FAQ', 'https://stonemaiergames.com/games/viticulture/faq/'],
  ]) {
    const link = node('a', label);
    link.href = url;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    sources.append(link);
  }
  const toggle = button(enabled ? '关闭新手引导' : '开启新手引导', () => {
    setEnabled(!enabled);
    paintRules();
  });
  toggle.setAttribute('aria-pressed', String(enabled));
  foot.append(sources, toggle);
  help.append(head, nav, intro, cards, foot);
}
function setEnabled(value) {
  enabled = value;
  save('vineyard-ee-guide', enabled ? 'on' : 'off');
  paintGuide();
  if (enabled && !guide.hidden && !panelState && help.hidden)
    guide.scrollIntoView({ block: 'nearest' });
}
function advice() {
  if (!state)
    return {
      title: '欢迎来到酒庄',
      text: '创建或加入房间后，会带你认识季节、资产、手牌和派工。',
      target: '#nickname',
      topic: 'overview',
    };
  if (!connected)
    return {
      title: '连接恢复后继续',
      text: '先等待重新连接。规则说明仍可查看。',
      target: '#connection',
      topic: 'overview',
    };
  const p = state.players.find((p) => p.id === state.youId);
  if (state.phase === 'lobby')
    return {
      title: '先让朋友落座',
      text: state.legal.canStart
        ? '朋友已到齐时，房主可以开始第一年。'
        : '分享房间码，至少两人落座后由房主开始。',
      target: state.legal.canStart ? '#start-game' : '#code-label',
      topic: 'overview',
    };
  if (state.phase === 'finished')
    return {
      title: '酒庄完成这一局',
      text: '先比较胜利分，再依次比较金币、酒和葡萄总品质。',
      target: '.final-results',
      topic: 'overview',
    };
  if (state.pendingChoice) {
    if (state.pendingChoice.playerId !== state.youId)
      return {
        title: '等待其他庄主选择',
        text: '你可以先查看自己的手牌，准备接下来的行动。',
        target: '#hand-panel',
        topic: 'orders',
      };
    const kind = state.pendingChoice.kind;
    return {
      title: kind === 'discard' ? '年末整理手牌' : '先完成当前选择',
      text:
        kind === 'discard'
          ? '点选多余手牌，弃至 7 张后确认。'
          : kind === 'papa'
            ? '比较父亲的赠礼与额外金币，选择适合起始手牌的一项。'
            : kind === 'fall'
              ? '选择想要的访客颜色：黄牌用于夏季，蓝牌用于冬季。'
              : '先选效果，再补齐所需卡牌或资源，最后确认。',
      target: kind === 'discard' ? '.discard-prompt' : '#ee-choice',
      topic:
        kind === 'papa' ? 'overview' : kind === 'fall' || kind === 'discard' ? 'seasons' : 'orders',
    };
  }
  if (state.turnId !== state.youId)
    return {
      title: '等候你的回合',
      text: '趁现在查看手牌与庄园，规划下一位工人去哪里。',
      target: '#hand-panel',
      topic: 'workers',
    };
  if (state.phase === 'wake')
    return {
      title: '春季 · 选起床顺序',
      text: '早起优先行动，晚起拿奖励；先想想今年最需要什么。',
      target: '#wake-options',
      topic: 'seasons',
    };
  const legal = (id) => {
    const space = state.spaces.find(
      (s) => s.id === id && (s.season === state.phase || s.season === 'any'),
    );
    return state.legal.canPlace && space && !actionReason(space, state);
  };
  const action = (id, title, text) => ({
    title,
    text,
    target: '[data-space="' + id + '"]',
    topic: actionTopics[id] || 'workers',
  });
  if (state.phase === 'summer') {
    if (!p.workers && !p.largeWorker)
      return {
        title: '本季已无待命工人',
        text: '可以结束本季，按流程进入秋季与冬季。',
        target: '#pass',
        topic: 'seasons',
      };
    const vines = state.hand.filter((c) => c.type === 'vine');
    if (vines.length && !vines.some((c) => !vineRequirements(p, c).reason))
      return action('plant', '先看看藤缺什么', '打开种植，再点藤卡查看建筑和田地条件。');
    if (legal('plant'))
      return action(
        'plant',
        '夏季 · 种下葡萄藤',
        '选择满足条件的藤与田地；给冬季收获和酿酒留些工人。',
      );
    if (!vines.length && !p.fields.some((f) => f.vines.length) && legal('draw_vine'))
      return action('draw_vine', '先拿葡萄藤', '去葡萄藤市场抽牌，再检查是否需要棚架或灌溉。');
    if (p.coins < 3 && legal('tour'))
      return action('tour', '给建设准备金币', '导览可以补充金币；别把冬季要用的工人全派完。');
    return {
      title: '准备冬季的生产',
      text: '田上有藤后，记得留工人收获、酿酒、交单；也可按手牌选择其他行动。',
      target: '#spaces',
      topic: 'seasons',
    };
  }
  if (state.phase === 'winter') {
    if (legal('fill_order'))
      return action('fill_order', '有订单可以交付', '选择订单与对应的酒，获得胜利分和年收入。');
    if (legal('make_wine'))
      return action('make_wine', '把葡萄酿成酒', '按订单需求搭配葡萄，确认前看酒种与品质预览。');
    if (legal('harvest'))
      return action(
        'harvest',
        '先收获田里的葡萄',
        '每块田一年只能收获一次；藤会留着供以后年份使用。',
      );
    if (p.wines.length && legal('draw_order'))
      return action(
        'draw_order',
        '给现有的酒找订单',
        '抽订单后比较酒种与品质；交付是常用得分方式。',
      );
    if (legal('winter_visitor'))
      return action('winter_visitor', '看看冬季访客', '选项会标注所需条件，先确认费用与资源。');
    return {
      title: '检查还有哪些行动可用',
      text: '绿色边框表示可派遣；本年不再行动时，可结束冬季等待结算。',
      target: '#pass',
      topic: 'workers',
    };
  }
  return {
    title: '按当前选择继续',
    text: '先完成屏幕上的选择，再进入下一个季节。',
    target: '#ee-choice',
    topic: 'seasons',
  };
}
const tour = [
  {
    title: '先认识季节',
    text: '春季选顺序，夏季建设，秋季抽访客，冬季收获酿酒。',
    target: '#season-track',
    topic: 'seasons',
  },
  {
    title: '看看自己的庄园',
    text: '金币、工人、田地和酒窖都在这里。派工前先看可用资源。',
    target: '#estate-panel',
    topic: 'workers',
  },
  {
    title: '认识你的手牌',
    text: '绿藤、紫订单、黄夏访客、蓝冬访客。只有你能看到自己的牌面。',
    target: '#hand-panel',
    topic: 'orders',
  },
];
function paintActionLesson() {
  const panel = $('#action-panel');
  if (!panel) return;
  let box = panel.querySelector('.action-lesson');
  if (!enabled || !panelState) {
    box?.remove();
    return;
  }
  if (!box) {
    box = node('div', null, 'action-lesson');
    panel.querySelector('form').before(box);
  }
  box.replaceChildren(
    node(
      'span',
      '新手提示 · ' + (actionLessons[panelState.space] || '先选工人和资源，再确认派遣。'),
    ),
  );
  const link = button('规则', () => showRules(actionTopics[panelState.space] || 'workers', link));
  box.append(link);
}
function paintGuide() {
  clearHighlight();
  document.querySelectorAll('[data-guide-toggle]').forEach((b) => {
    b.setAttribute('aria-pressed', String(enabled));
    b.textContent = enabled ? '引导：开' : '新手引导';
  });
  const welcome = $('#welcome-guide');
  welcome.hidden = !enabled;
  welcome.textContent = '新手引导已开启 · 加入房间后跟着短提示操作，随时可以关闭。';
  guide.hidden = !enabled || !state || $('#game').hidden;
  paintActionLesson();
  if (guide.hidden) return;
  const learning = step < 4 && !['lobby', 'finished'].includes(state.phase);
  const current = learning && step < 3 ? tour[step] : advice();
  const focusKey = Object.keys(document.activeElement?.dataset || {}).find((key) =>
    key.startsWith('guide'),
  );
  guide.replaceChildren();
  guide.dataset.mode = learning ? 'tour' : 'context';
  guide.dataset.step = String(step);
  guide.dataset.phase = state.phase;
  const text = node('div', null, 'guide-copy');
  text.append(
    node('small', learning ? '认识桌面 ' + (step + 1) + ' / 4' : '随局提示'),
    node('h3', current.title),
    node('p', current.text),
  );
  const controls = node('div', null, 'guide-controls');
  const locate = button('看这里 ↗', () => focusTarget(current.target));
  locate.dataset.guideLocate = '';
  const rules = button('相关规则', () => showRules(current.topic, rules));
  controls.append(locate, rules);
  if (learning) {
    if (step > 0) {
      const previous = button('上一步', () => {
        step--;
        save('vineyard-ee-tour-step', step);
        paintGuide();
      });
      previous.dataset.guidePrevious = '';
      controls.append(previous);
    }
    const next = button(step === 3 ? '开始随局提示' : '下一步 →', () => {
      step++;
      save('vineyard-ee-tour-step', step);
      paintGuide();
    });
    next.dataset.guideNext = '';
    controls.append(next);
    const skip = button('跳过导览', () => {
      step = 4;
      save('vineyard-ee-tour-step', step);
      paintGuide();
    });
    skip.dataset.guideSkip = '';
    controls.append(skip);
  } else {
    const restart = button('重看导览', () => {
      step = 0;
      save('vineyard-ee-tour-step', step);
      paintGuide();
    });
    restart.dataset.guideRestart = '';
    controls.append(restart);
  }
  const close = button('关闭', () => setEnabled(false));
  close.setAttribute('aria-label', '关闭新手引导');
  controls.append(close);
  guide.append(text, controls);
  if (focusKey)
    (
      [...controls.querySelectorAll('button')].find((b) => focusKey in b.dataset) ||
      controls.querySelector('[data-guide-restart]') ||
      controls.querySelector('[data-guide-next]')
    )?.focus({ preventScroll: true });
  const target = $(current.target);
  if (help.hidden && !panelState && target?.getClientRects().length)
    target.classList.add('guide-target');
}
export function renderHelp(view, online) {
  state = view;
  connected = online;
  if (!view && !help.hidden) $('#welcome').before(help);
  paintGuide();
}
export function initHelp() {
  help = node('section', null, 'rules-help panel');
  help.id = 'rules-help';
  help.hidden = true;
  help.setAttribute('aria-labelledby', 'rules-help-title');
  $('#welcome').before(help);
  guide = node('section', null, 'beginner-guide');
  guide.id = 'beginner-guide';
  guide.hidden = true;
  guide.setAttribute('aria-label', '新手引导');
  $('#status').after(guide);
  document.addEventListener('click', (event) => {
    const rule = event.target.closest('[data-rules-open], [data-rule-topic]');
    if (rule)
      showRules(rule.dataset.ruleTopic || actionTopics[panelState?.space] || 'overview', rule);
    if (event.target.closest('[data-guide-toggle]')) setEnabled(!enabled);
  });
  document.addEventListener(
    'keydown',
    (event) => {
      if (event.key === 'Escape' && !help.hidden) {
        event.stopImmediatePropagation();
        event.preventDefault();
        closeRules();
      }
    },
    true,
  );
  document.addEventListener('action-panel-state', (event) => {
    panelState = event.detail;
    paintGuide();
  });
  paintGuide();
}
