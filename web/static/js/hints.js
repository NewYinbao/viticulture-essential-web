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
    if (owner && owner !== next) owner.removeAttribute('aria-describedby');
    owner = next;
    bubble.textContent = next.dataset.hint;
    next.setAttribute('aria-describedby', bubble.id);
    bubble.hidden = false;
    // A dialog's top layer also needs to contain its tooltip.
    const container = next.closest('dialog') || document.body;
    if (bubble.parentElement !== container) container.append(bubble);
    const box = next.getBoundingClientRect(),
      width = bubble.offsetWidth,
      height = bubble.offsetHeight;
    bubble.style.left = Math.max(8, Math.min(innerWidth - width - 8, box.left)) + 'px';
    bubble.style.top =
      (box.bottom + height + 16 < innerHeight
        ? box.bottom + 8
        : Math.max(8, box.top - height - 8)) + 'px';
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
  document.addEventListener('scroll', hide, true);
  document.addEventListener('close', hide, true);
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
