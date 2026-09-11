export function cardArt(c, cls = '') {
  const e = document.createElement('img');
  e.className = cls;
  const k = parseInt(c.id?.split('-').pop() || '1', 10) || 1;
  e.src =
    '/art/handdrawn/' +
    (['summer', 'winter', 'mama', 'papa'].includes(c.type)
      ? 'portrait-' + ((k - 1) % 12)
      : c.type === 'vine'
        ? 'vine-' + ((k - 1) % 4)
        : c.type === 'structure'
          ? 'order'
          : 'order') +
    '.svg';
  e.alt = '原创手绘风格插画（部分卡牌共享主题图）';
  return e;
}
