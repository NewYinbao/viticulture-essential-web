import { localCard } from './card-i18n.js';
import { defaultLarge, privateSpace, hasBonus, winePreview } from './action-options.js';

const names = { red: '红', white: '白' };
const node = (tag, text, cls) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (cls) e.className = cls;
  return e;
};
function select(parent, name, label, items) {
  const wrap = node('label', label),
    input = node('select');
  input.name = name;
  for (const [value, label] of items) {
    const option = node('option', label);
    option.value = value;
    input.append(option);
  }
  input.required = true;
  wrap.append(input);
  parent.append(wrap);
  return input;
}
function checks(parent, name, label, items) {
  parent.append(node('p', label));
  if (!items.length) parent.append(node('p', '暂无可用资源', 'empty'));
  for (const [value, text] of items) {
    const wrap = node('label', null, 'check'),
      input = node('input');
    input.type = 'checkbox';
    input.name = name;
    input.value = value;
    wrap.append(input, document.createTextNode(text));
    parent.append(wrap);
  }
}

export function setupEE(space, view, act) {
  let dialog = document.querySelector('#ee-action-dialog');
  if (!dialog) {
    dialog = node('dialog');
    dialog.id = 'ee-action-dialog';
    document.body.append(dialog);
  }
  dialog.replaceChildren();
  const wake = space === 'wake',
    p = view.players.find((p) => p.id === view.youId);
  const title = node('h2', wake ? '起床奖励 · 选择访客' : space.name);
  title.id = 'ee-action-title';
  dialog.setAttribute('aria-labelledby', title.id);
  const form = node('form'),
    fields = node('div'),
    error = node('p', null, 'ee-warning');
  error.setAttribute('role', 'alert');
  dialog.append(title, form);
  form.append(fields, error);
  let mode, slot, decline, detail, summary;
  const bonus = () => !wake && hasBonus(space, slot?.value, decline?.value === 'yes');
  const recipeGroups = () => {
    const groups = {};
    p.grapes.forEach((g, i) => {
      const value = form.elements['recipe-' + i]?.value;
      if (value) (groups[value] ??= []).push(i);
    });
    return Object.values(groups);
  };
  const refreshSummary = () => {
    if (!summary) return;
    const kind = mode?.value || space.id;
    summary.replaceChildren();
    if (kind === 'make_wine') {
      const recipes = recipeGroups(),
        limit = bonus() ? 3 : 2;
      summary.append(node('p', `本次最多 ${limit} 瓶 · 已配置 ${recipes.length} 瓶`));
      for (const text of winePreview(p, recipes)) summary.append(node('p', text));
      if (recipes.length > limit)
        summary.append(node('p', '超出本次瓶数，请减少配方或选择可用奖励格。', 'ee-warning'));
    } else if (kind === 'plant' || kind === 'harvest') {
      summary.textContent = `本次最多 ${space.id === 'yoke' ? 1 : bonus() ? 2 : 1} ${kind === 'plant' ? '张藤' : '块田地'}`;
    }
  };
  if (wake) {
    select(fields, 'color', '选择一种访客', [
      ['summer', '夏季访客'],
      ['winter', '冬季访客'],
    ]);
  } else {
    const worker = select(fields, 'worker', '派遣工人', [
      ...(p.workers > 0 ? [['normal', `普通工人 · 剩余 ${p.workers} 位`]] : []),
      ...(p.largeWorker ? [['large', '大工人 · 普通格满时仍可进入']] : []),
    ]);
    worker.value = defaultLarge(space, p) ? 'large' : 'normal';
    if (space.capacity >= 2 && !privateSpace(space)) {
      slot = select(fields, 'slot', '行动格', [
        [0, '自动选择首个空格'],
        ...Array.from({ length: space.capacity }, (_, i) => [
          i + 1,
          `第 ${i + 1} 格${i === 0 ? ' · 奖励格' : ''}${space.occupied.some((o) => o.slot === i + 1) ? ' · 已占用' : ''}`,
        ]),
      ]);
      for (const option of slot.options)
        if (space.occupied.some((o) => o.slot === Number(option.value))) option.disabled = true;
      decline = select(fields, 'declineBonus', '行动格奖励', [
        ['no', '领取奖励（如所选格可领取）'],
        ['yes', '放弃奖励，仍执行行动'],
      ]);
    }
    const modeItems =
      space.id === 'sell_grapes'
        ? [
            ['sell_grapes', '出售葡萄'],
            ['sell_field', '出售空田地'],
            ['buy_field', '买回田地'],
          ]
        : space.id === 'yoke'
          ? [
              ['harvest', '收获一块田地'],
              ['uproot', '拔藤回手牌'],
            ]
          : [];
    if (modeItems.length) mode = select(fields, 'mode', '操作', modeItems);
    if (space.id === 'sell_grapes' && !p.grapes.length)
      mode.value = p.fields.some((f) => !f.sold && !f.vines.length) ? 'sell_field' : 'buy_field';
    detail = node('div');
    summary = node('div', null, 'action-preview');
    summary.setAttribute('role', 'status');
    fields.append(detail, summary);
    const rebuild = () => {
      detail.replaceChildren();
      error.textContent = '';
      const kind = mode?.value || space.id;
      if (['buy_field', 'sell_field', 'uproot'].includes(kind) || space.id === 'yoke') {
        const available = p.fields.filter((f) =>
          kind === 'buy_field'
            ? f.sold
            : kind === 'sell_field'
              ? !f.sold && !f.vines.length
              : kind === 'uproot'
                ? f.vines.length > 0
                : !f.sold && !f.harvested && f.vines.length > 0,
        );
        const field = select(
          detail,
          'field',
          '选择田地',
          available.map((f) => [
            f.index,
            `田地 ${f.index + 1} · ${f.capacity} 金币${f.sold ? ' · 已出售' : ''}`,
          ]),
        );
        if (!available.length) detail.append(node('p', '没有符合这项操作的田地。', 'empty'));
        if (kind === 'uproot') {
          const cards = node('div');
          detail.append(cards);
          const update = () => {
            cards.replaceChildren();
            const f = available.find((f) => String(f.index) === field.value);
            select(
              cards,
              'cardId',
              '拔回手牌的葡萄藤',
              (f?.vines || []).map((c) => [
                c.id,
                localCard(c).name + ` · 红 ${c.red} / 白 ${c.white}`,
              ]),
            );
          };
          field.onchange = update;
          update();
        }
      } else if (kind === 'plant') {
        detail.append(node('p', '勾选葡萄藤，并指定田地。田地容量按藤牌上的红白数值之和计算。'));
        const vines = view.hand.filter((c) => c.type === 'vine');
        if (!vines.length) detail.append(node('p', '没有葡萄藤手牌。', 'empty'));
        for (const c of vines) {
          const wrap = node('label', null, 'check'),
            input = node('input');
          input.type = 'checkbox';
          input.name = 'plantCards';
          input.value = c.id;
          const missing = [
            c.trellis && !p.buildings.includes('trellis') ? '棚架' : '',
            c.irrigation && !p.buildings.includes('irrigation') ? '灌溉' : '',
          ].filter(Boolean);
          const available = p.fields.filter(
            (f) =>
              !f.sold &&
              f.vines.reduce((s, v) => s + v.red + v.white, 0) + c.red + c.white <= f.capacity,
          );
          input.disabled = !!missing.length || !available.length;
          wrap.append(
            input,
            document.createTextNode(
              localCard(c).name +
                ` · 红 ${c.red} / 白 ${c.white}` +
                (missing.length
                  ? ' · 需要' + missing.join('、')
                  : !available.length
                    ? ' · 田地容量不足'
                    : ''),
            ),
          );
          detail.append(wrap);
          const field = select(
            detail,
            'field-' + c.id,
            '种入田地',
            available.map((f) => [
              f.index,
              `田地 ${f.index + 1} · 已种 ${f.vines.reduce((s, v) => s + v.red + v.white, 0)} / 容量 ${f.capacity}`,
            ]),
          );
          field.disabled = true;
          input.onchange = () => {
            field.disabled = !input.checked;
          };
        }
      } else if (kind === 'harvest') {
        checks(
          detail,
          'fields',
          '选择要收获的田地',
          p.fields
            .filter((f) => !f.sold && !f.harvested && f.vines.length)
            .map((f) => [f.index, '田地 ' + (f.index + 1)]),
        );
      } else if (kind === 'make_wine') {
        detail.append(
          node('p', '单颗葡萄酿红／白酒；红白各一酿桃红；两红一白酿起泡。分配后可查看预计品质。'),
        );
        p.grapes.forEach((g, i) => {
          const selectGrape = select(
            detail,
            'recipe-' + i,
            `${names[g.color]}葡萄 · 品质 ${g.value}`,
            [
              ['', '不使用'],
              ['1', '第 1 瓶'],
              ['2', '第 2 瓶'],
              ['3', '第 3 瓶'],
            ],
          );
          selectGrape.required = false;
        });
      } else if (kind === 'sell_grapes') {
        checks(
          detail,
          'grapes',
          '选择葡萄（品质 1–3 / 4–6 / 7–9 分别获得 1 / 2 / 3 金币）',
          p.grapes.map((g, i) => [
            i,
            `${names[g.color]}葡萄 · 品质 ${g.value} → ${Math.ceil(g.value / 3)} 金币`,
          ]),
        );
      }
      refreshSummary();
    };
    if (mode) mode.onchange = rebuild;
    form.addEventListener('change', refreshSummary);
    rebuild();
  }
  const buttons = node('div', null, 'ee-buttons'),
    cancel = node('button', '取消'),
    send = node('button', '确认派遣', 'primary');
  cancel.type = 'button';
  cancel.onclick = () => dialog.close();
  send.type = 'submit';
  buttons.append(cancel, send);
  form.append(buttons);
  form.onsubmit = async (event) => {
    event.preventDefault();
    send.disabled = true;
    error.textContent = '';
    try {
      const data = new FormData(form),
        action = wake
          ? { type: 'wake', slot: 5, color: data.get('color') }
          : {
              type: 'place',
              space: space.id,
              large: data.get('worker') === 'large',
              declineBonus: data.get('declineBonus') === 'yes',
            };
      action.revision = view.revision;
      if (!wake) {
        if (data.has('slot')) action.slot = Number(data.get('slot'));
        const kind = data.get('mode') || space.id;
        if (data.has('mode')) action.mode = kind;
        if (data.has('field')) action.field = Number(data.get('field'));
        if (data.has('cardId')) action.cardId = data.get('cardId');
        if (kind === 'plant') {
          action.cardIds = data.getAll('plantCards');
          action.fields = action.cardIds.map((id) => Number(data.get('field-' + id)));
          if (!action.cardIds.length) throw Error('至少选择一张葡萄藤');
          if (action.cardIds.length > (bonus() ? 2 : 1)) throw Error('种植数量超过所选行动格上限');
        }
        if (kind === 'harvest' && space.id !== 'yoke') {
          action.fields = data.getAll('fields').map(Number);
          if (!action.fields.length) throw Error('至少选择一块田地');
          if (action.fields.length > (bonus() ? 2 : 1)) throw Error('收获数量超过所选行动格上限');
        }
        if (kind === 'sell_grapes') {
          action.grapes = data.getAll('grapes').map(Number);
          if (!action.grapes.length) throw Error('至少选择一颗葡萄');
        }
        if (kind === 'make_wine') {
          action.recipes = recipeGroups();
          if (!action.recipes.length) throw Error('至少选择一个酿酒配方');
          if (action.recipes.length > (bonus() ? 3 : 2)) throw Error('配方数量超过所选行动格上限');
        }
      }
      if (await act(action)) dialog.close();
      else error.textContent = '操作未生效，请根据提示检查选择。';
    } catch (err) {
      error.textContent = err.message;
    } finally {
      send.disabled = false;
    }
  };
  dialog.showModal();
}
