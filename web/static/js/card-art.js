import { vineArt } from './vine-art.js';
export function cardArt(c, cls = '') {
  if (c.type === 'vine') return vineArt(c, cls);
  const e = document.createElement('img');
  e.className = cls;
  const k = parseInt(c.id?.split('-').pop() || '1', 10) || 1;
  e.src =
    '/art/handdrawn/' +
    (['summer', 'winter', 'mama', 'papa'].includes(c.type)
      ? 'portrait-' + ((k - 1) % 12)
      : 'order') +
    '.svg';
  e.alt = '原创手绘风格插画（部分卡牌共享主题图）';
  return e;
}
