import { tuscanyInputState, renderTuscanyInputs } from './tuscany-inputs.js';
let draft;
// Shared resource form for domain continuations and primary worker actions.
export function renderTuscanyChoice(form, view, online, submit) {
  const c = view.pendingChoice;
  if (!['tuscany_influence', 'tuscany_trade', 'structure_influence'].includes(c?.kind)) {
    draft = null;
    return false;
  }
  if (c.playerId !== view.youId) return false;
  const key = JSON.stringify([view.code, view.youId, c.id]);
  if (draft?.key !== key) draft = { key, values: {} };
  const kind = c.kind === 'structure_influence' ? 'influence' : c.kind.slice('tuscany_'.length);
  function draw(focus) {
    form.replaceChildren();
    const state = tuscanyInputState(kind, view, draft.values);
    renderTuscanyInputs(form, state, draft.values, online, draw);
    const reason = document.createElement('p');
    reason.setAttribute('role', 'status');
    reason.textContent = !online
      ? '连接恢复后才能提交。'
      : state.reason || '已就绪；确认执行当前奖励。';
    const confirm = document.createElement('button');
    confirm.type = 'submit';
    confirm.dataset.tuscanyConfirm = '';
    confirm.textContent = '确认当前奖励';
    confirm.disabled = !online || !!state.reason;
    const skip = document.createElement('button');
    skip.type = 'button';
    skip.dataset.choice = 'skip';
    skip.textContent = '放弃本次额外奖励';
    skip.disabled = !online || !c.options.includes('skip');
    skip.onclick = () => submit({ option: 'skip' });
    form.append(reason, confirm, skip);
    form.onsubmit = (e) => {
      e.preventDefault();
      if (!confirm.disabled && !tuscanyInputState(kind, view, draft.values).reason)
        submit({ option: kind === 'influence' ? 'place' : 'trade', ...state.payload });
    };
    if (focus)
      form.querySelector('[data-tuscany-input="' + focus + '"]')?.focus({ preventScroll: true });
    document.dispatchEvent(
      new CustomEvent('tuscany-choice-state', {
        detail: { code: view.code, youId: view.youId, choiceId: c.id, target: state.target },
      }),
    );
  }
  draw();
  return true;
}
