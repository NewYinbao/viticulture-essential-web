import { expansionDescriptions, configurationSummary } from './rules-content.js';

const node = (tag, text) => {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  return e;
};

export function renderExpansions(view, act, online) {
  let box = document.querySelector('#expansion-config');
  if (view.phase !== 'lobby') {
    box?.remove();
    return;
  }
  if (!box) {
    box = node('section');
    box.id = 'expansion-config';
    document.querySelector('#lobby-panel').insertBefore(box, document.querySelector('#start-game'));
  }
  const detailsOpen = !!box.querySelector('details')?.open;
  const focused = box.contains(document.activeElement) ? document.activeElement.id : '';
  box.replaceChildren();
  const config = view.config || {
    board: 'ee',
    visitors: 'ee',
    structures: false,
    specialWorkers: false,
  };
  const canEdit = !!view.legal.canConfigure && online;
  box.append(node('h4', '开局规则'));
  const summary = node(
    'p',
    '本局已保存：' +
      configurationSummary(view) +
      (view.phase === 'lobby' ? '；开局后锁定。' : '；已锁定。'),
  );
  summary.id = 'expansion-summary';
  summary.setAttribute('role', 'status');
  box.append(summary);
  const row = node('div');
  row.className = 'expansion-options';
  const selectors = {};
  for (const [key, label, entries] of [
    [
      'board',
      '主板',
      [
        ['ee', 'EE 本体'],
        ['tuscany', 'Tuscany 四季主板'],
      ],
    ],
    [
      'visitors',
      '访客牌组',
      [
        ['ee', 'EE 本体 · 76 张'],
        ['ee_moor', 'EE + Moor · 116 张'],
        ['rhine', 'Rhine · EE 主板 76 张／Tuscany 主板 80 张'],
      ],
    ],
  ]) {
    const wrap = node('label', label);
    const select = node('select');
    select.id = 'config-' + key;
    for (const [value, title] of entries) {
      const available = value === 'ee' || !!view.expansionAvailability?.[value];
      const option = node('option', title + (available ? '' : ' · 暂不可用'));
      option.value = value;
      option.disabled = value !== 'ee' && !view.expansionAvailability?.[value];
      select.append(option);
    }
    select.value = config[key];
    select.disabled = !canEdit;
    wrap.append(select);
    row.append(wrap);
    selectors[key] = select;
  }
  for (const [key, label] of [
    ['structures', 'Tuscany 建筑卡（Side 2／结构垫）'],
    ['specialWorkers', 'Tuscany 特殊工人'],
  ]) {
    const wrap = node('label');
    const input = node('input');
    input.type = 'checkbox';
    input.id = 'config-' + key;
    input.checked = config[key];
    input.disabled = !canEdit || !view.expansionAvailability?.[key];
    wrap.append(
      input,
      node('span', label + (view.expansionAvailability?.[key] ? '' : ' · 未完成，暂不可用')),
    );
    row.append(wrap);
    selectors[key] = input;
  }
  box.append(row);
  const preview = node('p');
  preview.id = 'expansion-preview';
  preview.setAttribute('role', 'status');
  preview.hidden = true;
  box.append(preview);
  const draftConfig = () => ({
    board: selectors.board.value,
    visitors: selectors.visitors.value,
    structures: selectors.structures.checked,
    specialWorkers: selectors.specialWorkers.checked,
  });
  for (const input of Object.values(selectors))
    input.onchange = () => {
      preview.hidden = false;
      preview.textContent =
        '尚未保存：' +
        configurationSummary({ config: draftConfig() }) +
        '。保存成功后全桌同步；Rhine 替换其他访客，不混洗。';
    };
  const note = node(
    'p',
    '建筑卡与特殊工人可独立启用。Tuscany 主板搭配建筑卡时使用 Side 2；EE 主板搭配建筑卡时先草拟4张。Rhine 替换其他访客，不能混洗。',
  );
  note.className = 'muted';
  box.append(note);
  const details = node('details');
  details.open = detailsOpen;
  details.append(node('summary', '各模块作用与搭配'));
  for (const [title, text] of expansionDescriptions) {
    details.append(node('h5', title), node('p', text));
  }
  box.append(details);
  const save = node('button', canEdit ? '保存开局规则' : '仅房主可在大厅配置');
  save.id = 'save-expansion-config';
  save.disabled = !canEdit;
  save.onclick = () =>
    act({
      type: 'configure',
      config: {
        board: selectors.board.value,
        visitors: selectors.visitors.value,
        structures: selectors.structures.checked,
        specialWorkers: selectors.specialWorkers.checked,
      },
    });
  box.append(save);
  if (focused) document.getElementById(focused)?.focus({ preventScroll: true });
}
