import { buildings } from './building-catalog.js';
import { localCard } from './card-i18n.js';
import { hint } from './hints.js';

export function cardMatchesSeason(card, season, config) {
  return (
    card.type === season ||
    (config?.visitors === 'ee_moor' && season === 'summer' && card.id === 'moor-winter-10')
  );
}

// Each custom control uses the existing keyboard-accessible resource-card picker.
// Private draw pools arrive only in the current responder's pending choice.
export function renderMoorField(f, { v, p, c, controls, chooser, readers, validators }) {
  if (v.config?.visitors !== 'ee_moor' || !f.type.startsWith('moor')) return false;
  const pick = (
    name,
    label,
    items,
    multi = false,
    min = multi ? f.min : 1,
    max = multi ? f.max : 1,
  ) => chooser(name, label, items, multi, min, max);
  if (f.type === 'moorNumber') {
    const read = pick(f.name, f.label || '次数', [
      [1, '1次'],
      [2, '2次'],
      [3, '3次'],
    ]);
    readers.push((a) => (a[f.name] = Number(read())));
  } else if (f.type === 'moorAgeOrder') {
    const read = pick('mode', '执行顺序', [
      ['upgrade_first', '先升级，再陈酿'],
      ['age_first', '先陈酿，再升级'],
    ]);
    readers.push((a) => (a.mode = read()));
  } else if (f.type === 'moorTraining') {
    const read = pick(
      f.name,
      '培训工人',
      (f.options || []).map((x) => [x.id, x.label]),
    );
    readers.push((a) => (a.specialWorker = read()));
  } else if (f.type === 'moorPrivateCards') {
    const cards = c.cards || [],
      multi = f.name !== 'cardId';
    const read = pick(
      f.name,
      '刚抽到的牌 · 仅自己可见',
      cards.map((x) => [x.id, localCard(x).name]),
      multi,
    );
    for (const b of controls.lastElementChild.querySelectorAll('button')) {
      const card = cards.find((x) => x.id === b.dataset.value);
      if (card) {
        b.classList.add('card', card.type);
        hint(b, localCard(card).description || card.name);
      }
    }
    readers.push((a) => (a[f.name] = read()));
  } else if (f.type === 'moorAnyField') {
    const read = pick(
      'field',
      '放置水果商',
      p.fields.map((x) => [
        String(x.index),
        '田地' + (x.index + 1) + ' · 容量' + x.capacity,
        x.sold ? '田地已出售' : x.structure ? '田地有结构' : x.fruitDealer ? '已有水果商' : '',
      ]),
    );
    readers.push((a) => (a.field = Number(read())));
  } else if (f.type === 'moorOpponentField') {
    const options = [];
    for (const q of v.players) {
      if (q.id === p.id) continue;
      for (const field of q.fields) {
        if (!field.sold && field.vines?.length && !field.structure) {
          const key = q.id + ':' + field.index;
          options.push([key, q.name + ' · 田地' + (field.index + 1)]);
        }
      }
    }
    const read = pick('opponentField', '选择对手的田地', options);
    readers.push((a) => {
      const [id, field] = read().split(':');
      a.targetIds = [id];
      a.field = Number(field);
    });
  } else if (f.type === 'moorSeats') {
    const options = f.options || [];
    const read = pick(
      'moorSeats',
      '取回先前季节的2名工人',
      options.map((x, i) => [
        String(i),
        x.label + ' · ' + (x.large ? '大工人' : x.workerType || '普通工人'),
      ]),
      true,
    );
    readers.push((a) => {
      const selected = read().map((i) => options[Number(i)]);
      a.targetIds = selected.map((x) => x.space);
      a.fields = selected.map((x) => x.slot);
    });
  } else if (f.type === 'moorDrawColors') {
    const colors = [
      ['vine', '葡萄藤'],
      ['summer', '夏季访客'],
      ['order', '订单'],
      ['winter', '冬季访客'],
    ];
    if (v.config?.structures) colors.push(['structure', '结构牌']);
    const reads = [0, 1, 2].map((i) => pick('moorColor' + i, '第' + (i + 1) + '张', colors));
    readers.push((a) => (a.colors = reads.map((read) => read())));
  } else if (f.type === 'moorBuildings') {
    const hand = v.hand || [],
      multi = f.name === 'buildings';
    const options = buildings
      .filter(
        ([id]) =>
          buildings.findIndex((x) => x[0] === id) < 8 || hand.some((c) => c.structureId === id),
      )
      .map(([id, label, cost, , desc], i) => {
        const structure =
          i >= 8 ||
          ![
            'trellis',
            'irrigation',
            'yoke',
            'medium_cellar',
            'large_cellar',
            'cottage',
            'windmill',
            'tasting_room',
          ].includes(id);
        const owned =
          p.buildings.includes(id) ||
          p.structureSlots?.includes(id) ||
          p.fields.some((x) => x.structure === id);
        const price = Math.max(0, cost - (f.discount || 0));
        return {
          id: structure ? 'structure-' + id : id,
          label: label + ' · ' + price + '金币',
          desc,
          reason: owned
            ? '已拥有'
            : !multi && id === 'large_cellar' && !p.buildings.includes('medium_cellar')
              ? '需要中酒窖'
              : p.coins < price
                ? '金币不足'
                : '',
        };
      });
    const read = pick(
      f.name,
      multi ? '选择3座建筑（按建造顺序）' : '建造建筑',
      options.map((x) => [x.id, x.label, x.reason]),
      multi,
    );
    for (const b of controls.lastElementChild.querySelectorAll('button')) {
      const x = options.find((x) => x.id === b.dataset.value);
      if (x) hint(b, [x.desc, x.reason || '尚未建造'].join(' · '), !!x.reason);
    }
    readers.push((a) => (a[f.name] = read()));
  } else return false;
  return true;
}
