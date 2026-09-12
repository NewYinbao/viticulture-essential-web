import { vineArt } from './vine-art.js';
import { illustratedCard } from './card-illustrations.js';
export function cardArt(c, cls = '') {
  return c.type === 'vine' ? vineArt(c, cls) : illustratedCard(c, cls);
}
