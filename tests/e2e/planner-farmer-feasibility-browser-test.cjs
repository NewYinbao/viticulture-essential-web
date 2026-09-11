const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { authorizeFixtureStore } = require('../helpers/fixture-auth.cjs');

const repo = process.env.VITICULTURE_SOURCE_ROOT || path.resolve(__dirname, '../..');
const fixture = JSON.parse(
  fs.readFileSync(process.env.PLANNER_FEASIBILITY_BROWSER_FIXTURES, 'utf8'),
);
const testCase = fixture.case;
const data = process.env.PLANNER_FEASIBILITY_BROWSER_DATA;
const output = process.env.PLANNER_FEASIBILITY_BROWSER_OUTPUT;
fs.mkdirSync(data, { recursive: true });
fs.mkdirSync(output, { recursive: true });
fs.writeFileSync(
  path.join(data, 'ee-state-v1.json'),
  JSON.stringify(authorizeFixtureStore(fixture)),
);

const report = {
  browser: 'Microsoft Edge',
  headless: true,
  checks: [],
  pageErrors: [],
};
let child;
let browser;
let base;

async function start() {
  child = spawn(process.env.PLANNER_FEASIBILITY_HTTP_EXE, ['-addr', '127.0.0.1:0', '-data', data], {
    cwd: repo,
    windowsHide: true,
  });
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timer = setTimeout(() => reject(new Error('server startup timeout: ' + log)), 15000);
    child.once('error', reject);
    const consume = (chunk) => {
      log += chunk;
      const match = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (match) {
        clearTimeout(timer);
        resolve('http://127.0.0.1:' + match[1]);
      }
    };
    child.stdout.on('data', consume);
    child.stderr.on('data', consume);
  });
}

async function stop() {
  if (child && child.exitCode === null)
    await new Promise((resolve) => {
      child.once('exit', resolve);
      child.kill();
    });
  child = null;
}

async function load(page, token) {
  await page.goto(base, { waitUntil: 'domcontentloaded' });
  await page.evaluate((value) => localStorage.setItem('vineyard-ee-token', value), token);
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.locator('#code-label').waitFor({ state: 'visible' });
}

async function state(page, token = testCase.tokens[0]) {
  const response = await page.request.get(base + '/api/state', {
    headers: { Authorization: 'Bearer ' + token },
  });
  assert.equal(response.status(), 200, await response.text());
  return response.json();
}

async function clickAction(page, selector, dialog = false) {
  if (dialog) page.once('dialog', (prompt) => prompt.accept());
  const control = page.locator(selector);
  await control.waitFor({ state: 'visible' });
  const [response] = await Promise.all([
    page.waitForResponse(
      (candidate) =>
        candidate.url().endsWith('/api/action') && candidate.request().method() === 'POST',
    ),
    control.click(),
  ]);
  assert.equal(response.status(), 200, await response.text());
  return response.json();
}

(async () => {
  try {
    await start();
    browser = await chromium.launch({ channel: 'msedge', headless: true });
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    page.on('pageerror', (error) => report.pageErrors.push(error.message));
    await load(page, testCase.tokens[0]);

    await page.locator('[data-visitor-option="plan"]').click();
    await page
      .locator('.visitor-resource')
      .filter({ has: page.locator('h4', { hasText: '选择行动' }) })
      .locator('[data-value="train"]')
      .click();
    await page.locator('select[name="slot"]').selectOption('2');
    await page
      .locator('.visitor-resource')
      .filter({ has: page.locator('h4', { hasText: '选择工人' }) })
      .locator('[data-value="special:farmer"]')
      .click();
    let response = await clickAction(page, '[data-visitor-submit]');
    assert.equal(response.pendingChoice?.kind, 'special_farmer');
    response = await clickAction(page, '[data-choice="discount"]');
    assert.equal(response.pendingChoice, null);
    assert.equal(response.players.find((p) => p.id === testCase.playerId).coins, 3);
    assert.equal(response.planned[0].workerType, 'farmer');
    assert.equal(response.planned[0].specialWorkerState, undefined);
    report.checks.push('3-coin Farmer training discount is selected without early payment');

    await stop();
    await start();
    await load(page, testCase.tokens[0]);
    let current = await state(page);
    assert.equal(current.players.find((p) => p.id === testCase.playerId).coins, 3);
    report.checks.push('Farmer discount reservation survives server restart');

    for (let step = 0; step < 30; step++) {
      current = await state(page);
      if (current.pendingChoice?.kind === 'planner') break;
      const actor = current.pendingChoice?.playerId || current.turnId;
      const index = current.players.findIndex((player) => player.id === actor);
      assert(index >= 0, 'active player has no fixture token');
      await load(page, testCase.tokens[index]);
      if (current.pendingChoice) {
        await clickAction(page, '[data-choice]:not([disabled])');
      } else {
        await clickAction(page, '#pass', true);
      }
    }
    current = await state(page);
    assert.equal(current.pendingChoice?.kind, 'planner');
    assert.equal(current.planned.length, 1, 'feasibility probe discarded the reservation');
    report.checks.push('winter feasibility keeps training affordable only by locked discount');

    await load(page, testCase.tokens[0]);
    const regular = page
      .locator('.visitor-resource')
      .filter({ has: page.locator('h4', { hasText: '培训工人' }) })
      .locator('[data-value="regular"]');
    await regular.waitFor({ state: 'visible' });
    assert.equal(await regular.getAttribute('aria-disabled'), 'false');
    await regular.click();
    response = await clickAction(page, '[data-visitor-submit]');
    const player = response.players.find((p) => p.id === testCase.playerId);
    assert.equal(player.coins, 0);
    assert.equal(player.totalWorkers, testCase.totalWorkers + 1);
    assert.equal((response.planned || []).length, 0);
    assert.equal(response.pendingChoice, null);
    report.checks.push('real Planner UI trains one regular worker for the locked 3-coin price');

    assert.deepEqual(report.pageErrors, []);
    report.status = 'passed';
  } catch (error) {
    report.status = 'failed';
    report.error = error.stack;
    process.exitCode = 1;
    console.error(error);
  } finally {
    if (browser) await browser.close();
    await stop();
    fs.writeFileSync(
      path.join(output, 'planner-farmer-feasibility-browser-result.json'),
      JSON.stringify(report, null, 2),
    );
    console.log(JSON.stringify(report));
  }
})();
