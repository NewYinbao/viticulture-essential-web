import { bonusLabel } from './action-options.js';
import { regionRewards } from './tuscany-inputs.js';
import { structureBuildings } from './building-catalog.js';
const seasons = {
  spring: '春季',
  summer: '夏季',
  fall: '秋季',
  winter: '冬季',
  year_end: '个人年末',
  ready: '已预约下年',
};
const node = (tag, text, cls) => {
  const n = document.createElement(tag);
  if (text != null) n.textContent = text;
  if (cls) n.className = cls;
  return n;
};
export function regionStanding(view, region) {
  const counts = view.players.map((p) => ({
    name: p.name,
    id: p.id,
    count: p.influence?.[region.id] || 0,
  }));
  const top = Math.max(0, ...counts.map((p) => p.count));
  const leaders = counts.filter((p) => p.count === top && top > 0);
  return {
    counts,
    label: !top
      ? '暂无星'
      : leaders.length > 1
        ? '并列最多 · 终局均不得分'
        : leaders[0].name + ' 暂时领先 · 终局' + region.points + '分',
  };
}
export function renderTuscanyBoard(view) {
  let panel = document.querySelector('#tuscany-board');
  if (view.config?.board !== 'tuscany' || view.phase === 'lobby') {
    panel?.remove();
    return;
  }
  const mapOpen = !!panel?.querySelector('#influence-map')?.open;
  const focus = panel?.contains(document.activeElement)
    ? document.activeElement.closest('details')?.id
    : null;
  const open = new Set(
    [...(panel?.querySelectorAll('details[open]') || [])].map((e) => e.dataset.season),
  );
  if (!panel) {
    panel = node('section', null, 'panel tuscany-board');
    panel.id = 'tuscany-board';
    document.querySelector('#board').before(panel);
  }
  panel.replaceChildren();
  panel.append(node('h2', 'Tuscany · 四季与影响力'));
  if (!view.expansionAvailability?.tuscany)
    panel.append(
      node(
        'p',
        '开发局面展示：主板仍未开放合法开局；建筑卡、特殊工人与扩展访客不可用。',
        'ee-warning',
      ),
    );
  const players = node('div', null, 'tuscany-personal-seasons');
  for (const p of view.players) {
    const line = node('article');
    line.dataset.playerId = p.id;
    const row = view.wakeSlots?.find((w) => w.slot === p.wake);
    line.append(
      node('strong', p.name + (p.id === view.youId ? ' · 你' : '')),
      node(
        'p',
        (seasons[p.season] || '待选起床') +
          ' · 本年起床 ' +
          (p.wake || '未选') +
          (p.passed ? ' · 已过季，等待全员' : ''),
      ),
    );
    line.append(node('small', row?.bonus || '首年选择第2–7行；选行不立即领奖'));
    if (p.nextWake) line.append(node('p', '下年已预约第' + p.nextWake + '行'));
    players.append(line);
  }
  panel.append(players);
  const rules = node('button', '起床与个人过季规则');
  rules.type = 'button';
  rules.dataset.ruleTopic = 'seasons';
  panel.append(rules);
  const overview = node('div', null, 'tuscany-season-overview');
  for (const [season, title] of Object.entries(seasons).slice(0, 4)) {
    const details = node('details');
    details.dataset.season = season;
    details.open = open.has(season);
    details.append(
      node('summary', title + (view.phase === season ? ' · 当前行动季' : ' · 查看行动与奖励')),
    );
    for (const s of (view.spaces || []).filter((s) => s.season === season)) {
      const line = node('p');
      line.append(node('strong', s.name + '：'), node('span', s.description));
      line.append(
        node(
          'small',
          Array.from(
            { length: s.capacity },
            (_, i) => '第' + (i + 1) + '格 ' + bonusLabel(s, i + 1),
          ).join('；'),
        ),
      );
      details.append(line);
    }
    overview.append(details);
  }
  panel.append(overview);
  const map = node('details', null, 'tuscany-influence');
  map.id = 'influence-map';
  map.open = mapOpen;
  const total = view.players.reduce(
    (sum, p) => sum + Object.values(p.influence || {}).reduce((n, v) => n + v, 0),
    0,
  );
  map.append(node('summary', '公开影响力地图 · 已放 ' + total + ' 星 · 展开查看地区与领先者'));
  if (view.config?.structures) {
    map.append(node('p', 'Tuscany Side 2 · 结构牌和影响力地图同时生效。'));
    const structures = node('div', null, 'tuscany-regions');
    for (const p of view.players) {
      const own = [
        ...(p.structureSlots || []).filter(Boolean),
        ...(p.fields || []).map((f) => f.structure).filter(Boolean),
      ];
      const line = node('article');
      line.append(node('h4', p.name + ' · 已建结构 ' + own.length));
      line.append(
        node(
          'p',
          own.length
            ? own
                .map((id) => structureBuildings.find((item) => item[0] === id)?.[1] || id)
                .join('、')
            : '暂无结构；结构牌可从本局结构牌堆抽取。',
        ),
      );
      structures.append(line);
    }
    map.append(structures);
  }
  map.append(node('p', '每人6星；放置领奖，移动不领奖（宴会厅除外）。地区独占多数才得终局分。'));
  const grid = node('div', null, 'tuscany-regions');
  for (const r of view.influenceRegions || []) {
    const region = node('article');
    region.dataset.region = r.id;
    const standing = regionStanding(view, r);
    region.append(
      node('h4', r.name + ' · ' + r.points + '分'),
      node('p', '首次放星：' + (regionRewards[r.reward] || '奖励尚未识别')),
    );
    for (const p of standing.counts)
      region.append(node('span', p.name + '：★ ' + p.count, 'tuscany-star-count'));
    region.append(node('small', standing.label));
    grid.append(region);
  }
  map.append(grid);
  const link = node('button', '放星／移星规则');
  link.type = 'button';
  link.dataset.ruleTopic = 'influence';
  map.append(link);
  panel.append(map);
  if (focus) panel.querySelector('#' + focus + ' > summary')?.focus({ preventScroll: true });
}
