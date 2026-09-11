import { art } from './graphics.js';
import { actionReason, vineRequirements } from './action-options.js';
import {
  actionTopics,
  actionLessons,
  actionLessonForView,
  rulesForView,
  pendingRule,
  configOf,
  configurationSummary,
  seasonNames,
} from './rules-content.js';

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
  returnSelector,
  visitorState,
  roomKey,
  ruleSignature,
  topic = 'overview';

function button(text, action, cls) {
  const b = node('button', text, cls);
  b.type = 'button';
  b.onclick = action;
  return b;
}
function saveTour() {
  save('vineyard-ee-tour-step', step);
  save('vineyard-ee-tour:' + roomKey, step);
}
function visibleTarget(selector) {
  if (!selector) return null;
  return (
    [...document.querySelectorAll(selector)].find(
      (e) =>
        !e.closest('[hidden]') &&
        e.getClientRects().length &&
        getComputedStyle(e).visibility !== 'hidden' &&
        !e.matches(':disabled, [aria-disabled="true"]'),
    ) || null
  );
}
function focusTarget(selector) {
  const target = visibleTarget(selector);
  if (!target) return;
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
  const back =
    returnFocus?.isConnected && visibleTarget(returnSelector) === returnFocus
      ? returnFocus
      : visibleTarget(returnSelector) ||
        visibleTarget('#game [data-rules-open], #welcome [data-rules-open]');
  if (back) {
    back.focus({ preventScroll: true });
    back.scrollIntoView({ block: 'nearest' });
  }
  paintGuide();
}
function showRules(id = 'overview', trigger = document.activeElement) {
  if (help.hidden) {
    returnFocus = trigger;
    returnSelector = trigger?.id
      ? '#' + CSS.escape(trigger.id)
      : trigger?.dataset.ruleTopic
        ? '[data-rule-topic="' + CSS.escape(trigger.dataset.ruleTopic) + '"]'
        : trigger?.hasAttribute('data-guide-rules')
          ? '[data-guide-rules]'
          : '[data-rules-open]';
  }
  topic = id;
  const anchor = $('#game').hidden
    ? $('#welcome')
    : id === 'current'
      ? $('#ee-choice')
      : $('#action-area');
  anchor.before(help);
  help.hidden = false;
  paintRules();
  document
    .querySelectorAll('[data-rules-open]')
    .forEach((b) => b.setAttribute('aria-expanded', 'true'));
  clearHighlight();
  paintGuide();
  help.scrollIntoView({ block: 'start' });
  $('#rules-help-title').focus({ preventScroll: true });
}
function paintRules() {
  const reference = rulesForView(state);
  const current = pendingRule(state, visitorState);
  const topics = current ? [current, ...reference.topics] : reference.topics;
  if (!topics.some((t) => t.id === topic)) topic = 'overview';
  const focused = help.contains(document.activeElement)
    ? {
        id: document.activeElement.id,
        topic: document.activeElement.dataset.topic,
        close: document.activeElement.classList.contains('help-close'),
      }
    : null;
  help.replaceChildren();
  const head = node('div', null, 'help-heading');
  const title = node('h3', '规则说明 · ' + reference.title);
  title.id = 'rules-help-title';
  title.tabIndex = -1;
  head.append(title, button('返回游戏 ×', closeRules, 'help-close'));
  const nav = node('nav', null, 'rule-topics');
  nav.setAttribute('aria-label', '规则主题');
  for (const item of topics) {
    const b = button(item.title, () => {
      topic = item.id;
      paintRules();
      help.querySelector('[data-topic="' + topic + '"]').focus({ preventScroll: true });
    });
    b.dataset.topic = item.id;
    b.setAttribute('aria-pressed', String(topic === item.id));
    nav.append(b);
  }
  const item = topics.find((t) => t.id === topic);
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
  for (const [label, url] of reference.sources) {
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
  toggle.id = 'rules-guide-toggle';
  toggle.setAttribute('aria-pressed', String(enabled));
  foot.append(sources, toggle);
  help.append(head);
  if (reference.warning)
    help.append(
      node('p', reference.warning + ' 以下已核实通则不代表模块可玩。', 'rule-unavailable'),
    );
  help.append(nav, intro, cards, foot);
  if (focused) {
    const replacement = focused.topic
      ? help.querySelector('[data-topic="' + focused.topic + '"]')
      : focused.close
        ? help.querySelector('.help-close')
        : document.getElementById(focused.id);
    (replacement || title)?.focus({ preventScroll: true });
  }
}
function setEnabled(value) {
  const closingGuide = !value && guide.contains(document.activeElement);
  enabled = value;
  save('vineyard-ee-guide', enabled ? 'on' : 'off');
  paintGuide();
  if (closingGuide)
    visibleTarget('#game [data-guide-toggle], #welcome [data-guide-toggle]')?.focus({
      preventScroll: true,
    });
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
        ? configurationSummary(state) + '。确认开局规则后，房主开始第一年；开局后锁定。'
        : state.ruleSupport?.reason || '分享房间码，至少两人落座后由房主开始；规则摘要对全桌公开。',
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
        target: null,
        topic: 'current',
      };
    const kind = state.pendingChoice.kind;
    const rule = pendingRule(state, visitorState);
    const selection = visitorState?.choiceId === state.pendingChoice.id ? visitorState : null;
    const visitor = !!state.pendingChoice.visitor || kind === 'planner';
    let target = ['discard', 'structure_barn'].includes(kind)
      ? '#hand .card[role="button"]'
      : visitor
        ? selection?.ready
          ? '[data-visitor-submit]'
          : visibleTarget('#hand .visitor-eligible')
            ? '#hand .visitor-eligible'
            : visibleTarget('.visitor-resource button, .visitor-resource select')
              ? '.visitor-resource button, .visitor-resource select'
              : '.visitor-options .option-ready'
        : '#ee-choice [data-choice], #ee-choice form button:not(:disabled)';
    if (
      ['messenger', 'special_mafioso', 'structure_fermentation', 'structure_mercado'].includes(kind)
    )
      target = '#ee-choice [data-choice-resources]';
    if (['tuscany_influence', 'tuscany_trade', 'structure_influence'].includes(kind))
      target = visibleTarget('[data-tuscany-confirm]')
        ? '[data-tuscany-confirm]'
        : '#ee-choice [data-tuscany-input]';
    if (panelState)
      target = panelState.reason
        ? panelState.target || '#action-panel .action-choice:not([aria-disabled="true"])'
        : '#confirm-action';
    return {
      title: rule.intro,
      text: panelState
        ? panelState.reason || '资源已选好，确认当前步骤。'
        : visitor
          ? selection?.reason ||
            (selection?.ready
              ? '所选分支已就绪，确认当前步骤；相关规则列出本牌效果、费用与数量。'
              : '选择可用分支，并按当前表单补齐资源；相关规则列出本牌各分支。')
          : rule.cards[0][1],
      target,
      topic: 'current',
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
      text:
        configOf(state).board === 'tuscany'
          ? '首年选第2–7行，选行不立即领奖；查看本行各季奖励。'
          : '早起优先行动，晚起拿奖励；先想想今年最需要什么。',
      target: '#wake-options button:not(:disabled)',
      topic: 'seasons',
    };
  if (panelState)
    return {
      title: '完成这次行动',
      text: panelState.reason || '所需资源已选好，可以确认派遣。',
      target: panelState.reason
        ? panelState.target ||
          '#action-panel .action-choice:not([aria-disabled="true"]), #action-panel select, #action-panel input'
        : '#confirm-action',
      topic: actionTopics[panelState.space] || 'workers',
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
  if (configOf(state).board === 'tuscany') {
    const available = (state.spaces || []).find((s) => s.season === state.phase && legal(s.id));
    if (p.passed || ['ready', 'year_end'].includes(p.season))
      return {
        title: '个人过季已完成',
        text: '奖励已在自己的过季步骤结算；等待全员进入下一季，不能提前派工。',
        target: null,
        topic: 'seasons',
      };
    if (available)
      return action(
        available.id,
        `${seasonNames[state.phase]} · ${available.name}`,
        actionLessonForView(state, available.id),
      );
    return {
      title: '检查本季行动',
      text: '按实际行动条件选择；结束本季会立即结算自己的过季奖励，再等待其他玩家。',
      target: '#pass',
      topic: 'seasons',
    };
  }
  if (state.phase === 'summer') {
    if (
      !p.workers &&
      !p.largeWorker &&
      !Object.values(state.workerPlacements || {}).some((placements) => placements.length)
    )
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
  box.replaceChildren(node('span', '新手提示 · ' + actionLessonForView(state, panelState.space)));
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
  const immediate = !connected || state.pendingChoice || state.turnId !== state.youId || panelState;
  const lesson = { ...tour[step] };
  if (step === 0 && configOf(state).board === 'tuscany')
    lesson.text = '四季都有工人行动；个人过季立即领奖，全员过季后才开始下一季派工。';
  if (step === 2 && configOf(state).structures)
    lesson.text =
      '绿藤、紫订单、黄夏访客、蓝冬访客与橙色建筑牌都占手牌上限；只有你能看自己的牌面。';
  const current = learning && step < 3 && !immediate ? lesson : advice();
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
  const target = visibleTarget(current.target);
  locate.disabled = !target || !connected || !help.hidden;
  if (!target) locate.title = '当前没有可定位的操作控件；可查看相关规则。';
  const rules = button('相关规则', () => showRules(current.topic, rules));
  rules.dataset.guideRules = '';
  controls.append(locate, rules);
  if (learning) {
    if (step > 0) {
      const previous = button('上一步', () => {
        step--;
        saveTour();
        paintGuide();
      });
      previous.dataset.guidePrevious = '';
      controls.append(previous);
    }
    const next = button(step === 3 ? '开始随局提示' : '下一步 →', () => {
      step++;
      saveTour();
      paintGuide();
    });
    next.dataset.guideNext = '';
    controls.append(next);
    const skip = button('跳过导览', () => {
      step = 4;
      saveTour();
      paintGuide();
    });
    skip.dataset.guideSkip = '';
    controls.append(skip);
  } else {
    const restart = button('重看导览', () => {
      step = 0;
      saveTour();
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
  if (help.hidden && connected && target) target.classList.add('guide-target');
}
export function renderHelp(view, online) {
  const key = view ? JSON.stringify([view.code, view.youId, configOf(view)]) : null;
  if (key !== roomKey) {
    clearHighlight();
    panelState = null;
    if (
      visitorState?.code !== view?.code ||
      visitorState?.youId !== view?.youId ||
      visitorState?.choiceId !== view?.pendingChoice?.id
    )
      visitorState = null;
    returnFocus = null;
    returnSelector = null;
    topic = 'overview';
    ruleSignature = null;
    roomKey = key;
    // Keep the user's preference, but scope unfinished tours to this room.
    step = Math.max(0, Math.min(4, Number(read('vineyard-ee-tour:' + key)) || 0));
  }
  state = view;
  connected = online;
  if (
    visitorState &&
    (visitorState.code !== view?.code ||
      visitorState.youId !== view?.youId ||
      visitorState.choiceId !== view?.pendingChoice?.id)
  )
    visitorState = null;
  if (!view && !help.hidden) $('#welcome').before(help);
  refreshRules();
  paintGuide();
}
function refreshRules() {
  const signature = JSON.stringify([
    state?.config,
    state?.ruleSupport,
    state?.pendingChoice,
    state?.optionReasons,
    state?.parentOptions,
    state?.players?.find((p) => p.id === state.youId),
    visitorState,
  ]);
  if (!help.hidden && signature !== ruleSignature) {
    paintRules();
    ruleSignature = signature;
  }
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
  document.addEventListener('visitor-choice-state', (event) => {
    visitorState = event.detail;
    if (
      state?.code === visitorState?.code &&
      state?.youId === visitorState?.youId &&
      state?.pendingChoice?.id === visitorState?.choiceId
    ) {
      refreshRules();
      paintGuide();
    }
  });
  document.addEventListener('tuscany-choice-state', (event) => {
    if (event.detail.code === state?.code && event.detail.choiceId === state?.pendingChoice?.id)
      paintGuide();
  });
  paintGuide();
}
