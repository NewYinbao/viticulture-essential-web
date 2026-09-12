// Natural configuration smoke tests: only each acting player's authenticated View
// and public rules inform decisions. No save reads, fixtures, deck seeds or score edits.
// API actions drive the policy; two real browsers observe HTTP/SSE and the end screen.
// This proves configuration lifecycle coverage, not every card or worker ability.
const { chromium } = require('playwright');
const { spawn, execFileSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const assert = require('node:assert/strict');
const repo = path.resolve(__dirname, '../..');
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-natural-matrix-'));
const output = process.env.VITICULTURE_NATURAL_OUTPUT || temporary;
fs.mkdirSync(output, { recursive: true });
const binary =
  process.env.VITICULTURE_NATURAL_EXE ||
  path.join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
const results = {
  scope:
    'Two-player natural unseeded games over EE/Tuscany and optional structures/special workers.',
  method:
    'Real create/configure/join/start HTTP, private-View policy, browser SSE, normal final-year scoring.',
  limitations: [
    'The policy does not play visitors.',
    'Special worker training and private placement are exercised; all abilities require separate tests.',
    'This is configuration coverage, not exhaustive card/branch coverage.',
  ],
  games: [],
};
let child,
  browser,
  base,
  pages = [];
async function start() {
  child = spawn(binary, ['-addr', '127.0.0.1:0', '-data', path.join(temporary, 'data')], {
    cwd: repo,
    windowsHide: true,
  });
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timer = setTimeout(() => reject(Error('Server startup timeout: ' + log)), 15000);
    child.once('error', reject);
    child.stdout.on('data', (chunk) => {
      log += chunk;
      const match = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (match) {
        clearTimeout(timer);
        resolve('http://127.0.0.1:' + match[1]);
      }
    });
    child.stderr.on('data', (chunk) => {
      log += chunk;
    });
  });
}
async function request(route, token, body) {
  const response = await fetch(base + route, {
    method: body ? 'POST' : 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: 'Bearer ' + token } : {}),
    },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const value = await response.json();
  assert.equal(response.status, 200, JSON.stringify({ route, body, value }));
  return value;
}
const mine = (view) => view.players.find((p) => p.id === view.youId);
const has = (p, building) => p.buildings.includes(building);
const regularCount = (p) => p.workers + Number(p.largeWorker);
function availablePlacement(view, spaceID, extra = {}) {
  const p = mine(view);
  const space = view.spaces.find((s) => s.id === spaceID);
  if (!space || ![view.phase, 'any'].includes(space.season)) return null;
  const free =
    space.capacity === 0 ||
    Array.from({ length: space.capacity }, (_, i) => i + 1).some(
      (slot) => !(space.occupied || []).some((seat) => seat.slot === slot),
    );
  const large = p.workers === 0 || !free;
  if (large ? !p.largeWorker : p.workers === 0) return null;
  return { type: 'place', space: spaceID, large, declineBonus: true, ...extra };
}
function plantable(view) {
  const p = mine(view);
  for (const card of view.hand.filter((c) => c.type === 'vine')) {
    if (card.trellis && !has(p, 'trellis') && !has(p, 'aqueduct')) continue;
    if (card.irrigation && !has(p, 'irrigation') && !has(p, 'aqueduct')) continue;
    const field = p.fields.find(
      (f) =>
        !f.sold &&
        !f.structure &&
        f.vines.reduce((sum, c) => sum + c.red + c.white, 0) + card.red + card.white <= f.capacity,
    );
    if (field) return { cardId: card.id, field: field.index };
  }
  return null;
}
const fixedCosts = { windmill: 5, tasting_room: 6, trellis: 2, irrigation: 3 };
function economicalAction(view, buildSpace, tourSpace) {
  const p = mine(view),
    planted = p.fields.some((f) => f.vines.length),
    wines = p.wines.length;
  const tour = () =>
    availablePlacement(view, tourSpace, tourSpace === 'build_tour' ? { mode: 'tour' } : {});
  const build = (building, extra = {}) =>
    availablePlacement(view, buildSpace, {
      building,
      ...(buildSpace === 'build_tour' ? { mode: 'build' } : {}),
      ...extra,
    });
  if (wines && has(p, 'tasting_room') && !p.tastingUsed) return tour();
  const wanted = ['windmill', ...(wines ? ['tasting_room'] : [])];
  if (!planted && view.hand.some((c) => c.type === 'vine')) wanted.push('trellis', 'irrigation');
  for (const id of wanted)
    if (!has(p, id) && p.coins >= fixedCosts[id]) {
      const action = build(id);
      if (action) return action;
    }
  if (view.config.structures && wines && !(p.structureSlots || []).some(Boolean)) {
    const cards = view.hand
      .filter((c) => c.type === 'structure' && c.structureCost <= p.coins)
      .sort((a, b) => a.structureCost - b.structureCost);
    if (cards.length) {
      const action = build(cards[0].id, { mode: 'mat', structureSlot: 0 });
      if (action) return action;
    }
  }
  return tour();
}
function trainingAction(view) {
  const p = mine(view);
  if (p.totalWorkers >= 5) return null;
  const toll = view.players.filter((q) => q.id !== p.id && has(q, 'academy')).length;
  const special =
    view.config.specialWorkers && !(p.specialWorkers || []).length;
  if (p.coins < (special ? 5 : 4) + toll) return null;
  return availablePlacement(view, 'train', {
    specialWorker: special ? view.specialWorkerPool[0] : 'regular',
  });
}
function strategy(view, game) {
  const p = mine(view),
    workers = regularCount(p),
    planted = p.fields.some((f) => f.vines.length);
  const done = game.actions.filter(
    (a) =>
      a.actor === p.id &&
      a.year === view.year &&
      a.phase === view.phase &&
      a.action.type === 'place',
  ).length;
  if (view.phase === 'winter') {
    if (view.config.board === 'ee' && !p.wines.length) {
      if (p.grapes.length) {
        const a = availablePlacement(view, 'make_wine', { recipes: [[0]] });
        if (a) return a;
      }
      const field = p.fields.find((f) => !f.sold && !f.harvested && f.vines.length);
      if (field) {
        const a = availablePlacement(view, 'harvest', { fields: [field.index] });
        if (a) return a;
      }
    }
    const train = trainingAction(view);
    if (train) return train;
    // Use trained special identities without invoking an optional board ability.
    const special = (p.specialWorkers || []).find(
      (id) => !p.specialWorkerUsed?.[id] && (p.specialWorkerReady?.[id] || 0) <= view.year,
    );
    if (special) return { type: 'place', space: 'gain_coin', workerType: special };
    return { type: 'pass' };
  }
  if (!workers) return { type: 'pass' };
  if (view.config.board === 'tuscany') {
    if (view.phase === 'spring') {
      if (done >= 1) return { type: 'pass' };
      if (!planted && !view.hand.some((c) => c.type === 'vine')) {
        const a = availablePlacement(view, 'draw_vine');
        if (a) return a;
      }
      return economicalAction(view, 'build', 'tour') || { type: 'pass' };
    }
    if (view.phase === 'summer') {
      if (done >= 1) return { type: 'pass' };
      const vine = plantable(view);
      if (vine && (!planted || !p.millUsed)) {
        const a = availablePlacement(view, 'plant', vine);
        if (a) return a;
      }
      if (p.totalWorkers < 5 && p.coins >= 4 && workers <= 2) return { type: 'pass' };
      if (p.coins >= 3) {
        const a = availablePlacement(view, 'trade', { trades: [{ give: 'coins', receive: 'vp' }] });
        if (a) return a;
      }
      return { type: 'pass' };
    }
    if (view.phase === 'fall') {
      if (!p.wines.length && p.grapes.length) {
        const a = availablePlacement(view, 'make_wine', { recipes: [[0]] });
        if (a) return a;
      }
      if (!p.wines.length && planted) {
        const field = p.fields.find((f) => !f.sold && !f.harvested && f.vines.length);
        if (field) {
          const a = availablePlacement(view, 'harvest', { fields: [field.index] });
          if (a) return a;
        }
      }
      if (done || (p.totalWorkers < 5 && p.coins >= 4 && workers <= 1)) return { type: 'pass' };
      return economicalAction(view, 'build_tour', 'build_tour') || { type: 'pass' };
    }
  }
  // EE summer: preserve enough workers for the real harvest/wine/training chain.
  const reserve =
    !p.wines.length && planted
      ? p.grapes.length
        ? 1
        : 2
      : p.totalWorkers < 5 && p.coins >= 4
        ? 1
        : 0;
  if (done && workers <= reserve) return { type: 'pass' };
  if (has(p, 'tasting_room') && p.wines.length && !p.tastingUsed) {
    const a = availablePlacement(view, 'tour');
    if (a) return a;
  }
  const vine = plantable(view);
  if (vine && has(p, 'windmill') && (!planted || !p.millUsed)) {
    const a = availablePlacement(view, 'plant', vine);
    if (a) return a;
  }
  if (!view.hand.some((c) => c.type === 'vine') && !planted) {
    const a = availablePlacement(view, 'draw_vine');
    if (a) return a;
  }
  return economicalAction(view, 'build', 'tour') || { type: 'pass' };
}
function choice(view, seat) {
  const c = view.pendingChoice;
  const action = { type: 'choose', choiceId: c.id };
  if (c.kind === 'papa') return { ...action, option: 'coins' };
  if (c.kind === 'discard') {
    const priority = { winter: 0, summer: 1, order: 2, structure: 3, vine: 4 };
    const cards = [...view.hand].sort((a, b) => priority[a.type] - priority[b.type]);
    return { ...action, cardIds: cards.slice(0, c.count).map((card) => card.id) };
  }
  if (c.kind === 'tuscany_next_wake') {
    const preferred = seat === 0 ? '6' : view.config.structures ? '4' : '2';
    return { ...action, option: c.options.includes(preferred) ? preferred : c.options[0] };
  }
  if (c.options.includes('skip')) return { ...action, option: 'skip' };
  return { ...action, option: c.options[0] };
}
async function runGame(config, index) {
  const game = { config, actions: [], thresholds: [], browserErrors: [], end: null };
  results.games.push(game);
  const first = await request('/api/create', null, {
    name: '自然山丘' + index,
    password: 'natural-game-123',
  });
  const second = await request('/api/join', null, {
    name: '自然河谷' + index,
    password: 'natural-game-456',
    code: first.code,
  });
  const tokens = [first.token, second.token];
  let host = await request('/api/state', tokens[0]);
  const ids = [host.youId, (await request('/api/state', tokens[1])).youId];
  host = await request('/api/action', tokens[0], {
    type: 'configure',
    config,
    revision: host.revision,
  });
  assert.deepEqual(host.config, config);
  const contexts = await Promise.all(
    [0, 1].map(() => browser.newContext({ viewport: { width: 1440, height: 1000 } })),
  );
  pages = await Promise.all(contexts.map((context) => context.newPage()));
  for (let i = 0; i < 2; i++) {
    pages[i].on('pageerror', (error) => {
      game.browserErrors.push(error.message);
    });
    await pages[i].goto(base);
    await pages[i].evaluate((token) => sessionStorage.setItem('vineyard-ee-token', token), tokens[i]);
    await pages[i].reload();
    await pages[i].locator('#code-label').waitFor();
  }
  await request('/api/action', tokens[0], { type: 'start', revision: host.revision });
  for (let step = 0; step < 2000; step++) {
    const publicView = await request('/api/state', tokens[0]);
    assert(publicView.year <= 50, 'Natural strategy exceeded 50 years');
    const actor = publicView.pendingChoice?.playerId || publicView.turnId;
    const seat = Math.max(0, ids.indexOf(actor));
    // Fetch only the acting player's private state for decision-making.
    const view = await request('/api/state', tokens[seat]);
    assert(view.players.every((p) => !Object.hasOwn(p, 'hand')));
    assert(!Object.hasOwn(view, 'decks'));
    const target = config.board === 'tuscany' ? 25 : 20;
    if (!game.thresholds.length && view.players.some((p) => p.vp >= target))
      game.thresholds.push({ year: view.year, phase: view.phase, revision: view.revision });
    if (view.phase === 'finished') {
      assert(game.thresholds.length);
      assert.equal(view.year, game.thresholds[0].year, 'Must finish the threshold year normally');
      assert(view.winnerIds.length);
      if (config.board === 'tuscany') assert(view.players.every((p) => p.season === 'ready'));
      game.end = {
        year: view.year,
        revision: view.revision,
        winnerIds: view.winnerIds,
        scores: view.players.map((p) => ({ id: p.id, vp: p.vp, coins: p.coins })),
      };
      for (const id of ['build', 'plant', 'harvest', 'make_wine'])
        assert(
          game.actions.some((a) => a.action.space === id),
          'Missing natural production action ' + id,
        );
      if (config.structures)
        assert(
          view.players.some((p) => (p.structureSlots || []).some(Boolean)),
          'No structure built naturally',
        );
      if (config.specialWorkers)
        assert(
          game.actions.some((a) => a.action.workerType),
          'No trained special worker placed',
        );
      await pages[0].waitForFunction(() =>
        document.querySelector('#season-title')?.textContent.includes('收官'),
      );
      await pages[0].screenshot({
        path: path.join(output, 'natural-' + index + '-finished.png'),
        fullPage: true,
      });
      console.log(
        'PASS natural configuration',
        index,
        JSON.stringify(config),
        'year',
        view.year,
        'steps',
        step,
      );
      break;
    }
    let action;
    if (view.pendingChoice) action = choice(view, seat);
    else if (view.phase === 'wake') {
      const preferred = seat === 0 ? 6 : config.board === 'tuscany' && config.structures ? 4 : 2;
      const slots = view.wakeSlots.filter(
        (slot) => !slot.playerId && (config.board !== 'tuscany' || slot.slot !== 1),
      );
      action = {
        type: 'wake',
        slot: slots.some((s) => s.slot === preferred) ? preferred : slots[0].slot,
      };
    } else action = strategy(view, game);
    const before = mine(view);
    const after = await request('/api/action', tokens[seat], {
      ...action,
      revision: view.revision,
    });
    const next = mine(after);
    if (action.type === 'place') {
      if (action.workerType)
        assert(next.specialWorkerUsed[action.workerType], 'Special identity not consumed');
      else
        assert.equal(
          regularCount(before) - regularCount(next),
          1,
          'Placement must spend one real worker',
        );
    }
    game.actions.push({
      actor: view.youId,
      year: view.year,
      phase: view.phase,
      action,
      revision: after.revision,
      scores: after.players.map((p) => ({ id: p.id, vp: p.vp })),
    });
  }
  assert(game.end, 'No natural terminal state');
  assert.deepEqual(game.browserErrors, [], 'Browser runtime errors');
  for (const context of contexts) await context.close();
  pages = [];
  fs.writeFileSync(
    path.join(output, 'natural-matrix-result.json'),
    JSON.stringify(results, null, 2),
  );
}
(async () => {
  try {
    if (!process.env.VITICULTURE_NATURAL_EXE)
      execFileSync(process.env.GO_BIN || 'go', ['build', '-o', binary, './cmd/viticulture'], {
        cwd: repo,
      });
    await start();
    browser = await chromium.launch({
      ...(process.platform === 'win32' ? { channel: 'msedge' } : {}),
      headless: true,
    });
    const visitors = (process.env.VITICULTURE_NATURAL_VISITORS || 'ee').split(',');
    let index = 0;
    for (const visitor of visitors)
      for (const board of ['ee', 'tuscany'])
        for (const structures of [false, true])
          for (const specialWorkers of [false, true])
            await runGame({ board, structures, specialWorkers, visitors: visitor }, index++);
    results.status = 'passed';
  } catch (error) {
    results.status = 'failed';
    results.error = error.stack;
    process.exitCode = 1;
    console.error(error);
  } finally {
    if (browser) await browser.close();
    if (child && child.exitCode == null)
      await new Promise((resolve) => {
        child.once('exit', resolve);
        child.kill();
      });
    fs.writeFileSync(
      path.join(output, 'natural-matrix-result.json'),
      JSON.stringify(results, null, 2),
    );
    console.log('Natural matrix evidence:', output);
  }
})();
