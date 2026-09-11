// Short inline state, detail on pointer hover / keyboard focus / touch.
export function hint(element, message, blocked = false) {
  element.dataset.hint = message;
  element.classList.add('has-hint');
  element.classList.toggle('is-blocked', blocked);
  element.setAttribute('aria-disabled', String(blocked));
}

let bubble;
export function installHints() {
  bubble = document.createElement('div');
  bubble.id = 'context-hint';
  bubble.setAttribute('role', 'tooltip');
  bubble.hidden = true;
  document.body.append(bubble);
  let owner;
  const hide = () => {
    if (owner) owner.removeAttribute('aria-describedby');
    owner = null;
    bubble.hidden = true;
  };
  const show = (target) => {
    const next = target?.closest?.('[data-hint]');
    if (!next || !next.dataset.hint) return hide();
    const scroller = next.closest('.action-panel form > div:first-child');
    if (scroller) {
      const area = scroller.getBoundingClientRect(),
        target = next.getBoundingClientRect();
      if (target.bottom <= area.top || target.top >= area.bottom) return hide();
    }
    if (owner && owner !== next) owner.removeAttribute('aria-describedby');
    owner = next;
    bubble.textContent = next.dataset.hint;
    next.setAttribute('aria-describedby', bubble.id);
    bubble.hidden = false;
    // A dialog's top layer also needs to contain its tooltip.
    const container = next.closest('dialog') || document.body;
    if (bubble.parentElement !== container) container.append(bubble);
    const bounds = next.closest('.action-panel')?.getBoundingClientRect();
    bubble.style.maxWidth = bounds ? Math.min(290, bounds.width - 24) + 'px' : '';
    const box = next.getBoundingClientRect(),
      width = bubble.offsetWidth,
      height = bubble.offsetHeight;
    const left = Math.max(8, bounds ? bounds.left + 8 : 8);
    const right = Math.min(innerWidth - 8, bounds ? bounds.right - 8 : innerWidth - 8);
    const top = Math.max(8, bounds ? bounds.top + 8 : 8);
    const bottom = Math.min(innerHeight - 8, bounds ? bounds.bottom - 8 : innerHeight - 8);
    bubble.style.left = Math.max(left, Math.min(right - width, box.left)) + 'px';
    bubble.style.top =
      Math.max(
        top,
        Math.min(
          bottom - height,
          box.bottom + height + 8 < bottom ? box.bottom + 8 : box.top - height - 8,
        ),
      ) + 'px';
  };
  document.addEventListener('pointerover', (e) => show(e.target));
  document.addEventListener('focusin', (e) => show(e.target));
  document.addEventListener('pointerout', (e) => {
    if (!owner?.contains(e.relatedTarget)) hide();
  });
  document.addEventListener('focusout', hide);
  document.addEventListener('pointerdown', (e) => {
    if (e.pointerType === 'touch') show(e.target);
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') hide();
  });
  document.addEventListener(
    'scroll',
    () => {
      const focused = owner === document.activeElement ? owner : null;
      hide();
      if (focused)
        requestAnimationFrame(() => {
          if (focused === document.activeElement) show(focused);
        });
    },
    true,
  );
  document.addEventListener('close', hide, true);
  document.addEventListener('action-panel-close', hide);
}

// Necessary resource counts only; effects can have special sequencing/discounts.
// Exact card-specific rule prerequisites are supplied by the server.
export function optionResourceReason(view, choice, fields) {
  const p = view.players.find((p) => p.id === view.youId);
  for (const f of fields) {
    const minimum = f.min ?? 1;
    if (!minimum) continue;
    if (f.type === 'cards') {
      const cards = view.hand.filter(
        (c) =>
          !f.filter ||
          c.type === f.filter ||
          (f.filter === 'visitors' && ['summer', 'winter'].includes(c.type)) ||
          (f.filter === 'season' && c.type === view.phase),
      );
      if (cards.length < minimum) return `还缺 ${minimum - cards.length} 张所需手牌`;
    }
    if (f.type === 'wines' && p.wines.filter((w) => w.value >= (f.minValue || 1)).length < minimum)
      return f.minValue > 1 ? `需要品质 ≥${f.minValue} 的酒` : '没有足够的酒';
    if (f.type === 'grapes' && p.grapes.length < minimum) return '没有足够的葡萄';
    if (
      f.type === 'fields' &&
      p.fields.filter((f) => !f.sold && !f.harvested && f.vines.length).length < minimum
    )
      return '没有可收获田地';
    if (f.type === 'uproot' && p.fields.reduce((n, f) => n + f.vines.length, 0) < minimum)
      return `田上至少需要 ${minimum} 张藤`;
    if (f.type === 'swap' && p.fields.filter((f) => f.vines.length).length < 2)
      return '需要两块不同田地上的藤';
  }
  return '';
}
