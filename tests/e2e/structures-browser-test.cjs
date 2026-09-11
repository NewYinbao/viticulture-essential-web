// 36 printed-card trigger fixtures; explicitly seeded, never natural-game evidence.
// Every action is submitted by visible UI controls in an isolated real browser.
const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs'),
  path = require('node:path'),
  os = require('node:os'),
  assert = require('node:assert/strict');
const { pbkdf2Sync, randomBytes } = require('node:crypto');
const repo = path.resolve(__dirname, '../..');
const root =
  process.env.VITICULTURE_STRUCTURES_OUTPUT ||
  fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-structures-evidence-'));
fs.mkdirSync(root, { recursive: true });
const bundle = JSON.parse(
  fs.readFileSync(
    process.env.VITICULTURE_STRUCTURES_FIXTURES ||
      path.join(repo, 'artifacts/expansions/structure-fixtures.json'),
    'utf8',
  ),
);
const data = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-structure-ui-'));
// Protect explicitly seeded fixture seats under the current persisted auth contract.
// Authentication attacks are covered by separate tests; this suite tests game controls.
bundle.Passwords = {};
const salt = randomBytes(16);
const record = {
  Salt: salt.toString('hex'),
  Hash: pbkdf2Sync('fixture-password-123', salt, 600000, 32, 'sha256').toString('hex'),
};
for (const session of Object.values(bundle.Sessions)) bundle.Passwords[session.PlayerID] = record;
fs.writeFileSync(path.join(data, 'ee-state-v1.json'), JSON.stringify(bundle));
let child, browser, page, base, currentCase;
const result = {
  method: bundle.method,
  cases: [],
  browserErrors: [],
  status: 'running',
};
async function start() {
  child = spawn(
    process.env.VITICULTURE_STRUCTURES_EXE || path.join(repo, 'dist/Viticulture-Expansions.exe'),
    ['-addr', '127.0.0.1:0', '-data', data],
    { windowsHide: true, cwd: repo },
  );
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timer = setTimeout(() => reject(Error('Startup timeout ' + log)), 15000);
    child.once('error', reject);
    child.stdout.on('data', (chunk) => {
      log += chunk;
      const m = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (m) {
        clearTimeout(timer);
        resolve('http://127.0.0.1:' + m[1]);
      }
    });
    child.stderr.on('data', (chunk) => (log += chunk));
  });
}
async function state() {
  return page.evaluate(async () => {
    const response = await fetch('/api/state', {
      headers: {
        Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token'),
      },
    });
    if (!response.ok) throw Error(await response.text());
    return response.json();
  });
}
async function send(selector) {
  const response = page.waitForResponse(
    (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
  );
  await page.locator(selector).click();
  const r = await response;
  assert.equal(r.status(), 200, await r.text());
  const v = await r.json();
  return v;
}
async function groups(entries = []) {
  for (const [name, value] of entries) {
    if (name.startsWith('tuscany:'))
      await page.locator('[data-tuscany-input="' + name.slice(8) + '"]').selectOption(value);
    else await page.locator('[data-group="' + name + '"][data-value="' + value + '"]').click();
  }
}
async function step(s) {
  if (s.kind === 'place' || s.kind === 'resource') {
    const selector =
      s.kind === 'place'
        ? '[data-space="' + s.space + '"]'
        : '[data-choice-resources="' + s.space + '"]';
    await page.locator(selector).click();
    await page.locator('#action-panel').waitFor();
    const decline = page.locator('[data-group="declineBonus"][data-value="yes"]');
    if (await decline.count()) await decline.click();
    await groups(s.groups || []);
    if (['charmat', 'mixer', 'workshop', 'fermentation_tank'].includes(currentCase.card))
      await page.screenshot({
        path: path.join(root, currentCase.code + '-' + currentCase.card + '-action.png'),
        fullPage: true,
      });
    await send('#confirm-action');
    await page.waitForFunction(() => !document.querySelector('#action-panel'));
  } else if (s.kind === 'influence') {
    for (const [key, value] of s.groups)
      await page.locator('[data-tuscany-input="' + key + '"]').selectOption(value);
    await send('[data-tuscany-confirm]');
  } else if (s.kind === 'pass') {
    page.once('dialog', (dialog) => dialog.accept());
    await send('#pass');
  } else if (s.kind === 'choice') await send('[data-choice="' + s.space + '"]');
  else if (s.kind === 'visitor') {
    await page.locator('[data-visitor-option="' + s.space + '"]').click();
    await send('[data-visitor-submit]');
  } else if (s.kind === 'visitorBonus') {
    for (const [type, id] of s.groups)
      await page.locator('input[name="structure-bonus-' + type + '"][value="' + id + '"]').check();
    const label = s.space === 'tap_room' ? '弃1瓶酒得2分' : '弃2颗红/白葡萄得3分';
    const response = page.waitForResponse(
      (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
    );
    await page.getByRole('button', { name: label, exact: true }).click();
    assert.equal((await response).status(), 200);
  } else if (s.kind === 'barn') {
    for (const [id] of s.groups) await page.locator('#hand [data-card-id="' + id + '"]').click();
    await send('#confirm-discard');
  } else throw Error('Unknown UI step ' + s.kind);
}
function check(c, v) {
  const p = v.players.find((p) => p.id === c.ownerId);
  for (const [key, value] of Object.entries(c.expect)) {
    let actual;
    if (key === 'hand') actual = p.id === v.youId ? (v.hand || []).length : p.handCount;
    else if (key === 'wines' || key === 'grapes') actual = (p[key] || []).length;
    else if (key === 'wineValues') actual = p.wines.map((x) => x.value);
    else if (key === 'grapeValues') actual = p.grapes.map((x) => x.value);
    else if (key === 'wineTypes') actual = p.wines.map((x) => x.type);
    else if (key === 'harvested') actual = p.fields.filter((f) => f.harvested).length;
    else if (key === 'planted') actual = p.fields.reduce((s, f) => s + f.vines.length, 0);
    else if (key === 'building') actual = p.buildings.includes(value) ? value : null;
    else if (key === 'pisa') actual = p.influence.pisa;
    else if (key === 'year') actual = v.year;
    else actual = p[key];
    assert.deepEqual(actual, value, c.card + ' ' + key);
  }
}
(async () => {
  try {
    await start();
    browser = await chromium.launch({
      ...(process.platform === 'win32' ? { channel: 'msedge' } : {}),
      headless: true,
    });
    page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    page.on('pageerror', (e) => result.browserErrors.push(e.message));
    for (const c of bundle.cases) {
      currentCase = c;
      await page.goto(base);
      await page.evaluate((token) => localStorage.setItem('vineyard-ee-token', token), c.token);
      await page.reload();
      await page.waitForFunction(
        (code) => document.querySelector('#code-label')?.textContent === '房间 ' + code,
        c.code,
      );
      const before = await state();
      assert.equal(
        await page.locator('#influence-map [data-region]').count(),
        7,
        'Side 2 lost influence map',
      );
      for (const action of c.steps) await step(action);
      const after = await state();
      check(c, after);
      assert(!after.players.some((p) => Object.hasOwn(p, 'hand')), 'Public players leaked hand');
      const evidence = {
        card: c.card,
        code: c.code,
        steps: c.steps,
        expected: c.expect,
        beforeRevision: before.revision,
        afterRevision: after.revision,
        status: 'passed',
      };
      result.cases.push(evidence);
      if (c.card === 'banquet_hall') await page.locator('#influence-map > summary').click();
      const width = await page.evaluate(() => ({
        actual: document.documentElement.scrollWidth,
        viewport: innerWidth,
      }));
      assert(width.actual <= width.viewport + 1, 'Page overflow ' + c.card);
      await page.screenshot({
        path: path.join(root, c.code + '-' + c.card + '.png'),
        fullPage: true,
      });
      console.log('PASS printed-card UI', c.card);
    }
    assert.equal(result.cases.length, 36);
    assert.deepEqual(result.browserErrors, []);
    result.status = 'passed';
  } catch (error) {
    result.status = 'failed';
    result.error = error.stack;
    process.exitCode = 1;
    console.error(error);
    if (page)
      await page
        .screenshot({ path: path.join(root, 'failure.png'), fullPage: true })
        .catch(() => {});
  } finally {
    if (browser) await browser.close();
    if (child && child.exitCode == null)
      await new Promise((resolve) => {
        child.once('exit', resolve);
        child.kill();
      });
    fs.writeFileSync(
      path.join(root, 'structures-browser-result.json'),
      JSON.stringify(result, null, 2),
    );
    console.log('Structure fixture evidence:', root);
  }
})();
