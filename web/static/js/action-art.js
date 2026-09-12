import { motif, sceneSvg, svgImage } from './svg-art.js';
// Foreground tools describe the action, rather than repeating a generic estate.
export const actionScenes = {
  draw_vine: ['cards', 'vine'],
  tour: ['house', 'coins'],
  build: ['house', 'hammer'],
  plant: ['vine', 'shovel'],
  summer_visitor: ['worker', 'sun'],
  sell_grapes: ['grape', 'coins'],
  draw_order: ['order', 'cards'],
  harvest: ['basket', 'grape'],
  make_wine: ['press', 'bottle'],
  fill_order: ['order', 'bottle'],
  train: ['worker', 'book'],
  winter_visitor: ['worker', 'snow'],
  influence: ['map', 'star'],
  trade: ['coins', 'arrows', 'grape'],
  flip_field: ['field', 'arrows', 'coins'],
  build_tour: ['house', 'hammer', 'coins'],
  sell_wine: ['bottle', 'coins'],
  gain_coin: ['coins'],
  yoke: ['basket', 'shears'],
  draw_structure: ['cards', 'house'],
};
export function boardActionArt(space, cls = 'action-art') {
  const parts = actionScenes[space.id] || ['house'];
  const objects =
    parts.length === 1
      ? motif(parts[0], 89, 26, 128)
      : motif(parts[0], 38, 24, 133) +
        motif(parts[1], 171, 67, 89) +
        (parts[2] ? motif(parts[2], 212, 20, 48) : '');
  return svgImage(
    sceneSvg(objects, space.season === 'winter' ? '#e6edf0' : '#f4e9cd'),
    cls,
    space.name,
  );
}
