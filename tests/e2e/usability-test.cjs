// Isolated fixtures test public UI behavior, not full-game strategy.
const { ROOT, GO, temp, executable } = require('./runtime.cjs');
const { chromium } = require('playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const net = require('node:net');
const { execFileSync, spawn } = require('node:child_process');
const tmp = temp('viticulture-usability-'), out = path.join(ROOT, 'artifacts/usability');
fs.mkdirSync(out, { recursive: true });
const report = { fixtures: true, checks: [], errors: [], screenshots: [] };
let server, browser, base;
const delay = (ms) => new Promise((r) => setTimeout(r, ms));
async function start() {
  server = spawn(path.join(tmp, executable('server')), ['-addr', new URL(base).host, '-data', path.join(tmp, 'data')], { stdio: 'ignore' });
  for (let i = 0; i < 100; i++) { try { if ((await fetch(base + '/api/health')).ok) return; } catch {} await delay(50); }
  throw Error('Server readiness failed');
}
async function stop() { if (server && server.exitCode === null) { const p = server; await new Promise((r) => { p.once('exit', r); p.kill(); }); } server = null; }
async function api(url, body, token) {
  const res = await fetch(base + url, { method: body ? 'POST' : 'GET', headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: 'Bearer ' + token } : {}) }, ...(body ? { body: JSON.stringify(body) } : {}) });
  const data = await res.json(); assert.equal(res.status, 200, JSON.stringify(data)); return data;
}
(async () => {
  try {
    execFileSync(GO, ['build', '-o', path.join(tmp, executable('server')), './cmd/viticulture'], { cwd: ROOT });
    const port = await new Promise((resolve) => { const s = net.createServer(); s.listen(0, '127.0.0.1', () => { const p = s.address().port; s.close(() => resolve(p)); }); });
    base = 'http://127.0.0.1:' + port; await start();
    const owner = await api('/api/create', { name: '山丘庄主' });
    for (const name of ['河谷庄主', '林间庄主']) await api('/api/join', { name, code: owner.code });
    const lobby = await api('/api/state', null, owner.token);
    await api('/api/action', { type: 'start', revision: lobby.revision }, owner.token);
    await stop();
    const file = path.join(tmp, 'data/ee-state-v1.json'), original = JSON.parse(fs.readFileSync(file));
    const template = original.Rooms[owner.code], store = { Rooms: {}, Sessions: {} };
    const catalog = JSON.parse(fs.readFileSync(path.join(ROOT, 'internal/game/cards/ee_cards.json')));
    const vines = catalog.filter((c) => c.type === 'vine' && !c.trellis && !c.irrigation);
    function seed(code) {
      const room = structuredClone(template);
      Object.assign(room, { code, phase: 'summer', year: 1, revision: 1, choices: [], context: null, resume: '', planned: [], turnId: room.players[0].id });
      room.spaces.forEach((s) => s.occupied = []);
      room.players.forEach((p, i) => { Object.assign(p, { wake: i + 1, coins: 12, workers: 2, largeWorker: true, passed: false, hand: [], buildings: [], grapes: [], wines: [] }); p.fields.forEach((f) => { f.vines = []; f.harvested = false; f.sold = false; }); });
      store.Rooms[code] = room; store.Sessions[code] = { Code: code, playerId: room.turnId }; return room;
    }
    const plant = seed('PLANT'); plant.players[0].hand = [vines[0], catalog.find((c) => c.type === 'order')];
    plant.players[1].hand = [vines[1], vines[2], catalog.find((c) => c.type === 'winter')];
    plant.spaces.find((s) => s.id === 'tour').occupied = [{ slot: 2, playerId: plant.players[1].id, large: false }];
    const yoke = seed('YOKE'); yoke.players[0].buildings = ['yoke']; yoke.players[0].fields[0].vines = [vines[0]]; yoke.players[0].fields[1].vines = [vines[1]];
    const wine = seed('WINE'); wine.phase = 'winter'; wine.players[0].buildings = ['medium_cellar']; wine.players[0].grapes = [{ id: 'r', color: 'red', value: 5 }, { id: 'w', color: 'white', value: 4 }];
    wine.spaces.find((s) => s.id === 'make_wine').occupied = [{ slot: 1, playerId: wine.players[1].id, large: false }, { slot: 2, playerId: wine.players[2].id, large: false }];
    const stale = seed('STALE'); stale.players[0].hand = [vines[0]];
    const paid = seed('PAID'); paid.players[0].coins = 2;
    paid.context = { actorId: paid.turnId, returnTurnId: paid.turnId, space: 'summer_visitor' };
    paid.choices = [{ id: 'cost-choice', kind: 'visitor', playerId: paid.turnId, options: ['buy', 'sell'], visitor: { cardId: 'summer-02', stage: 'effect', actorId: paid.turnId } }];
    const selection = seed('SELECT'); selection.players[0].hand = [vines[0], vines[1]];
    selection.context = { actorId: selection.turnId, returnTurnId: selection.turnId, space: 'summer_visitor' };
    selection.choices = [{ id: 'cards-choice', kind: 'visitor', playerId: selection.turnId, options: ['coins', 'vp'], visitor: { cardId: 'summer-15', stage: 'effect', actorId: selection.turnId } }];
    fs.writeFileSync(file, JSON.stringify(store)); await start();
    browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    page.on('pageerror', (e) => report.errors.push(e.message));
    await page.goto(base);
    async function load(code) { await page.evaluate((c) => localStorage.setItem('vineyard-ee-token', c), code); await page.reload(); await page.locator('#code-label').waitFor({ state: 'visible' }); }
    async function screenshot(name) { const dest = path.join(out, name + '.png'); await page.screenshot({ path: dest, fullPage: false }); report.screenshots.push(dest); }
    await load('PLANT');
    assert.equal(await page.locator('[data-space=tour] .worker-slot').nth(0).getAttribute('aria-label'), '奖励');
    assert.equal(await page.locator('[data-space=tour] .worker-slot').nth(1).locator('img').count(), 1);
    assert.match(await page.locator('[data-space=gain_coin] .slots').innerText(), /不限人数/);
    assert.match(await page.locator('[data-space=yoke]').getAttribute('data-hint'), /轭/);
    assert(await page.locator('[data-space=yoke]').isDisabled());
    const opponent = page.locator('.player').nth(1);
    assert.match(await opponent.locator('.public-hand-counts').innerText(), /葡萄藤 2/);
    report.checks.push('slot indices, unlimited coin, private yoke and public hand counts');
    await page.locator('[data-filter=vine]').click(); assert.equal(await page.locator('#hand .card:visible').count(), 1);
    await page.locator('[data-filter=all]').click(); assert.equal(await page.locator('#hand .card:visible').count(), 2);
    await page.locator('[data-space=plant]').click();
    assert.equal(await page.locator('#ee-action-dialog [name=mode]').count(), 0);
    const checkbox = page.locator('#ee-action-dialog [name=plantCards]').first();
    assert(await page.locator('#ee-action-dialog select[name^=field-]').first().isDisabled());
    await checkbox.check(); assert(await page.locator('#ee-action-dialog select[name^=field-]').first().isEnabled());
    assert.equal(await page.locator('#ee-action-dialog [name=declineBonus]').inputValue(), 'no');
    await page.locator('#ee-action-dialog button').filter({ hasText: '取消' }).click();
    await page.locator('[data-space=draw_vine]').click();
    assert.equal(await page.locator('#action-form [name=declineBonus]').inputValue(), 'no');
    await page.locator('#close-dialog').click();
    report.checks.push('hand filters, plant-only action, field selection and usable bonus defaults');
    await page.evaluate(() => window.scrollTo(0, 0)); await screenshot('desktop-table');
    await page.setViewportSize({ width: 390, height: 844 });
    await screenshot('mobile-table');
    await page.locator('.table-nav a[href="#hand-panel"]').click();
    assert(await page.locator('#hand-filters').isVisible()); await screenshot('mobile-hand');
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
    const nav = await page.locator('.table-nav').boundingBox(); assert(nav.y + nav.height <= 844);
    report.checks.push('390px layout has no horizontal overflow; fixed navigation reaches hand');
    await load('YOKE'); await page.locator('[data-space=yoke]').click();
    await page.locator('#ee-action-dialog [name=mode]').selectOption('uproot');
    await page.locator('#ee-action-dialog [name=field]').selectOption('1');
    assert.equal(await page.locator('#ee-action-dialog [name=cardId]').inputValue(), vines[1].id);
    assert.equal(await page.locator('#ee-action-dialog [name=cardId] option').count(), 1);
    const uprootResponse = page.waitForResponse((r) => r.url().endsWith('/api/action'));
    await page.locator('#ee-action-dialog button.primary').click(); const uprooted = await (await uprootResponse).json();
    assert.equal(uprooted.hand[0].id, vines[1].id);
    report.checks.push('yoke uproot options follow field; real server accepts chosen vine');
    await load('WINE'); await page.locator('[data-space=make_wine]').click();
    assert.equal(await page.locator('#ee-action-dialog [name=worker]').inputValue(), 'large');
    await page.locator('[name=recipe-0]').selectOption('1'); await page.locator('[name=recipe-1]').selectOption('1');
    assert.match(await page.locator('.action-preview').innerText(), /桃红酒 · 品质 6/);
    assert.match(await page.locator('.action-preview').innerText(), /最多 2 瓶/); await screenshot('mobile-wine-dialog');
    const wineResponse = page.waitForResponse((r) => r.url().endsWith('/api/action'));
    await page.locator('#ee-action-dialog button.primary').click(); const res = await wineResponse;
    assert.equal(res.status(), 200); const made = await res.json(); assert.equal(made.players[0].wines[0].type, 'blush'); assert.equal(made.players[0].wines[0].value, 6);
    report.checks.push('full board defaults to large worker; wine preview matches server result');
    await load('STALE');
    await page.route('**/api/events?*', (route) => route.abort()); await page.reload();
    await page.locator('#code-label').waitFor({ state: 'visible' });
    await page.unroute('**/api/events?*'); await page.reload();
    await page.locator('[data-space=plant]').click();
    await api('/api/action', { type: 'place', space: 'gain_coin', revision: 1 }, 'STALE');
    await page.waitForFunction(() => !document.querySelector('#ee-action-dialog').open);
    assert.match(await page.locator('#toast').innerText(), /局面已更新/);
    report.checks.push('incoming new revision closes stale action dialog');
    page.once('dialog', (d) => d.accept()); await page.locator('#switch-table').click();
    assert(await page.locator('#resume-room').isVisible()); await page.locator('#resume-room').click();
    await page.locator('#code-label').waitFor({ state: 'visible' });
    assert.match(await page.locator('#code-label').innerText(), /STALE/);
    report.checks.push('return-to-entry preserves explicit resume path');
    await load('PAID');
    const unavailable = page.locator('[data-visitor-option=buy]');
    assert.equal(await unavailable.getAttribute('aria-disabled'), 'true');
    await unavailable.hover(); assert.match(await page.locator('#context-hint').innerText(), /还差 7 金币/);
    await unavailable.focus(); assert.equal(await unavailable.getAttribute('aria-describedby'), 'context-hint');
    assert.equal(await page.locator('[data-visitor-option=sell]').getAttribute('aria-pressed'), 'true');
    assert(await page.locator('[data-visitor-submit]').isEnabled());
    assert.equal(await page.locator('.visitor-details').getAttribute('open'), null);
    await screenshot('mobile-visitor-prerequisite');
    await load('SELECT');
    assert.equal(await page.locator('[data-visitor-option=vp]').getAttribute('aria-disabled'), 'true');
    assert(await page.locator('[data-visitor-submit]').isDisabled());
    const eligible = page.locator('#hand .visitor-eligible');
    await eligible.nth(0).click(); assert(await page.locator('[data-visitor-submit]').isDisabled());
    await eligible.nth(1).click(); assert(await page.locator('[data-visitor-submit]').isEnabled());
    assert.match(await page.locator('.form-readiness').innerText(), /已就绪/);
    report.checks.push('unaffordable options have hover/focus reasons; legal option defaults; confirmation reacts to resource selection');
    const broken = await page.evaluate(() => [...document.images].filter((i) => i.complete && !i.naturalWidth).map((i) => i.src));
    assert.deepEqual(broken, []); assert.deepEqual(report.errors, []); report.completed = true;
    console.log(JSON.stringify(report, null, 2));
  } catch (e) { report.failure = e.stack; console.error(e); process.exitCode = 1; }
  finally { if (browser) await browser.close(); await stop(); fs.writeFileSync(path.join(out, 'result.json'), JSON.stringify(report, null, 2)); }
})();
