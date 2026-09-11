// Three seeded gameplay fixtures, real UI requests, real process restart.
const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs'),
  path = require('node:path'),
  os = require('node:os');
const assert = require('node:assert/strict');
const { pbkdf2Sync, randomBytes } = require('node:crypto');
const repo = path.resolve(__dirname, '../..');
const out =
  process.env.VITICULTURE_MOOR_OUTPUT ||
  fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-moor-evidence-'));
fs.mkdirSync(out, { recursive: true });
const bundle = JSON.parse(fs.readFileSync(process.env.VITICULTURE_MOOR_FIXTURES, 'utf8'));
const data = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-moor-ui-'));
const salt = randomBytes(16);
const password = {
  Salt: salt.toString('hex'),
  Hash: pbkdf2Sync('fixture-password-123', salt, 600000, 32, 'sha256').toString('hex'),
};
bundle.Passwords = {};
for (const session of Object.values(bundle.Sessions)) bundle.Passwords[session.PlayerID] = password;
fs.writeFileSync(path.join(data, 'ee-state-v1.json'), JSON.stringify(bundle));
const result = {
  method: 'explicit fixtures; visible UI gameplay; HTTP invalid submissions only',
  status: 'running',
  cases: [],
  errors: [],
  rejected: 0,
  restarts: 0,
};
let child, browser, page, opponentPage, base, current;
async function start() {
  child = spawn(
    process.env.VITICULTURE_MOOR_EXE || path.join(repo, 'dist/Viticulture-Expansions.exe'),
    ['-addr', '127.0.0.1:0', '-data', data],
    { windowsHide: true, cwd: repo },
  );
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timeout = setTimeout(() => reject(Error('Startup timeout: ' + log)), 15000);
    child.once('error', reject);
    child.stdout.on('data', (chunk) => {
      log += chunk;
      const match = log.match(/http:\/\/127\.0\.0\.1:\d+/);
      if (match) {
        clearTimeout(timeout);
        resolve(match[0]);
      }
    });
    child.stderr.on('data', (chunk) => {
      log += chunk;
    });
  });
}
async function stop() {
  if (child && child.exitCode == null)
    await new Promise((resolve) => {
      child.once('exit', resolve);
      child.kill();
    });
}
async function state(token = current.token) {
  const response = await fetch(base + '/api/state', {
    headers: { Authorization: 'Bearer ' + token },
  });
  assert.equal(response.status, 200, await response.clone().text());
  return response.json();
}
async function login(target, token) {
  await target.goto(base);
  await target.evaluate((value) => localStorage.setItem('vineyard-ee-token', value), token);
  await target.reload();
  await target.waitForFunction(
    (code) => document.querySelector('#code-label')?.textContent === '房间 ' + code,
    current.code,
  );
}
async function views() {
  return Promise.all([state(), state(current.opponentToken)]);
}
async function rejectUnchanged(payload, token = current.token) {
  const before = await views();
  const response = await fetch(base + '/api/action', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
    body: JSON.stringify(payload),
  });
  assert(
    response.status >= 400 && response.status < 500,
    'invalid request accepted: ' + (await response.text()),
  );
  assert.deepEqual(await views(), before, 'rejected request changed player-visible state');
  result.rejected++;
}
async function privacy() {
  const [owner, other] = await views();
  const choice = owner.pendingChoice;
  assert(choice?.id, 'actor lacks actionable choice ID');
  assert.deepEqual(Object.keys(other.pendingChoice).sort(), ['kind', 'playerId']);
  for (const key of ['id', 'cards', 'cardIds', 'options', 'visitor', 'optionFields'])
    assert(!Object.hasOwn(other.pendingChoice, key));
  assert(!JSON.stringify(other).includes('new-private-order'));
  assert(!JSON.stringify(other).includes('old-private-order'));
  assert.equal(
    await opponentPage.locator('[data-visitor-submit], [data-choice-resources]').count(),
    0,
    'opponent rendered actionable private choice',
  );
  await rejectUnchanged(
    { type: 'choose', choiceId: choice.id, option: choice.options[0] },
    current.opponentToken,
  );
  await rejectUnchanged({
    type: 'choose',
    choiceId: choice.id + '-stale',
    option: choice.options[0],
  });
}
async function submit(selector, expected = 200) {
  const pending = (await state()).pendingChoice;
  if (pending?.id) await privacy();
  const before = await views();
  const response = page.waitForResponse(
    (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
  );
  await page.locator(selector).click();
  const reply = await response,
    payload = reply.request().postDataJSON();
  assert.equal(reply.status(), expected, await reply.text());
  if (expected === 200 && payload.type === 'choose') await rejectUnchanged(payload);
  if (expected !== 200) {
    assert.deepEqual(await views(), before, 'failed UI resource choice changed state');
    result.rejected++;
    await page.screenshot({
      path: path.join(out, current.code + '-rejected-resource.png'),
      fullPage: true,
    });
  }
  return state();
}
async function option(value) {
  await page.locator('[data-visitor-option="' + value + '"]').click();
}
async function hand(id) {
  await page.locator('#hand [data-card-id="' + id + '"]').click();
}
async function resource(value) {
  await page.locator('.visitor-resource button[data-value="' + value + '"]').click();
}
async function group(name, value) {
  await page.locator('[data-group="' + name + '"][data-value="' + value + '"]').click();
}
async function place(space, card, slot = 1) {
  await page.locator('[data-space="' + space + '"]').click();
  await page.locator('#action-panel').waitFor();
  if (card) await group('cardId', card);
  if (slot) await group('slot', slot);
  return submit('#confirm-action');
}
async function finishSecond() {
  const v = await state();
  assert.equal(v.pendingChoice?.visitor?.stage, 'second');
  await option('skip');
  return submit('[data-visitor-submit]');
}
async function councilman() {
  await place('summer_visitor', current.card);
  await option('play');
  await hand('moor-winter-05');
  await submit('[data-visitor-submit]');
  const beforeRestart = await views();
  assert.equal(beforeRestart[0].pendingChoice.kind, 'structure_mercado');
  await privacy();
  await page.screenshot({ path: path.join(out, 'COUNCIL-before-restart.png'), fullPage: true });
  await opponentPage.screenshot({
    path: path.join(out, 'COUNCIL-opponent-private-choice.png'),
    fullPage: true,
  });
  await stop();
  await start();
  result.restarts++;
  await login(page, current.token);
  await login(opponentPage, current.opponentToken);
  assert.deepEqual(await views(), beforeRestart, 'process restart lost pending nested choice');
  await page.locator('[data-choice-resources="structure_mercado"]').click();
  assert.equal(
    await page
      .locator('#action-panel [data-group="cardId"][data-value="old-private-order"]')
      .count(),
    0,
  );
  await group('cardId', 'new-private-order');
  await group('wines', 'red');
  await page.screenshot({
    path: path.join(out, 'COUNCIL-restored-order-picker.png'),
    fullPage: true,
  });
  await submit('#confirm-action');
  let v = await state();
  assert.equal(v.pendingChoice.visitor.cardId, 'moor-winter-05');
  assert.equal(v.players.find((p) => p.id === current.playerId).vp, 2);
  await option('vp');
  await submit('[data-visitor-submit]');
  v = await finishSecond();
  const p = v.players.find((p) => p.id === current.playerId);
  assert.equal(p.vp, 3);
  assert.equal(p.coins, 2);
  assert.equal(p.wines.length, 1);
  assert.equal(v.pendingChoice, null);
  assert.deepEqual(
    v.hand.map((c) => c.id),
    ['old-private-order'],
  );
}
async function truss() {
  await place('winter_visitor', current.card);
  await option('harvest');
  await resource('0');
  await resource('1');
  await submit('[data-visitor-submit]');
  assert.equal((await state()).pendingChoice.kind, 'structure_fermentation');
  await page.locator('[data-choice-resources="structure_fermentation"]').click();
  await group('grapes', '0');
  await group('grapes', '1');
  // A legal blush recipe would leave only one grape for mandatory age-X=2.
  await submit('#confirm-action', 400);
  await group('grapes', '1');
  await submit('#confirm-action');
  let v = await state();
  assert.equal(v.pendingChoice.visitor.stage, 'age');
  assert.equal(v.pendingChoice.visitor.moorCount, 2);
  await option('age');
  await resource('0');
  await resource('1');
  await submit('[data-visitor-submit]');
  v = await finishSecond();
  const p = v.players.find((p) => p.id === current.playerId);
  assert.deepEqual(
    p.wines.map((w) => [w.type, w.value]),
    [['red', 2]],
  );
  assert.deepEqual(
    p.grapes.map((g) => [g.color, g.value]),
    [
      ['white', 3],
      ['red', 2],
    ],
  );
  assert.equal(p.fields.filter((f) => f.harvested).length, 2);
  assert.equal(v.pendingChoice, null);
}
async function fruit() {
  await place('summer_visitor', current.card, 2);
  await option('attach');
  await submit('[data-visitor-submit]');
  let v = await state(),
    p = v.players.find((p) => p.id === current.playerId);
  assert.equal(p.fields[0].fruitDealer, true);
  assert.equal(p.coins, 0);
  assert.equal(p.vp, 0);
  await page.locator('[data-space="yoke"]').click();
  await group('field', '0');
  await submit('#confirm-action');
  v = await state();
  assert.equal(v.pendingChoice.visitor.cardId, 'moor-summer-04');
  assert.equal(v.pendingChoice.visitor.stage, 'reward');
  await option('vp');
  await submit('[data-visitor-submit]');
  v = await state();
  p = v.players.find((p) => p.id === current.playerId);
  assert.equal(p.vp, 1);
  assert.equal(p.coins, 0);
  assert.equal(p.fields[0].harvested, true);
  assert.equal(p.yokeUsed, true);
  assert.deepEqual(
    p.grapes.map((g) => [g.color, g.value]),
    [
      ['red', 2],
      ['white', 1],
    ],
  );
  assert.equal(v.pendingChoice, null);
}
(async () => {
  try {
    await start();
    browser = await chromium.launch({
      ...(process.platform === 'win32' ? { channel: 'msedge' } : {}),
      headless: true,
    });
    page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    opponentPage = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    for (const tab of [page, opponentPage])
      tab.on('pageerror', (error) => result.errors.push(error.message));
    for (current of bundle.cases) {
      await login(page, current.token);
      await login(opponentPage, current.opponentToken);
      if (current.code === 'COUNCIL') await councilman();
      else if (current.code === 'TRUSS') await truss();
      else await fruit();
      await page.screenshot({ path: path.join(out, current.code + '.png'), fullPage: true });
      result.cases.push({ code: current.code, status: 'passed' });
      console.log('PASS Moor HTTP/UI', current.code);
    }
    assert.equal(result.cases.length, 3);
    assert.equal(result.restarts, 1);
    assert.deepEqual(result.errors, []);
    result.status = 'passed';
  } catch (error) {
    result.status = 'failed';
    result.error = error.stack;
    process.exitCode = 1;
    console.error(error);
    if (page)
      await page
        .screenshot({ path: path.join(out, 'failure.png'), fullPage: true })
        .catch(() => {});
  } finally {
    if (browser) await browser.close();
    await stop();
    fs.writeFileSync(
      path.join(out, 'moor-interactions-result.json'),
      JSON.stringify(result, null, 2),
    );
    console.log('Evidence:', out);
  }
})();
