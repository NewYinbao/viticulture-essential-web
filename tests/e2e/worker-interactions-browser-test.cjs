// Deliberately seeded fixtures. All gameplay submissions use real browser controls.
const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs'),
  path = require('node:path'),
  os = require('node:os');
const assert = require('node:assert/strict');
const { pbkdf2Sync, randomBytes } = require('node:crypto');
const repo = path.resolve(__dirname, '../..');
const output =
  process.env.VITICULTURE_WORKER_OUTPUT ||
  fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-worker-evidence-'));
fs.mkdirSync(output, { recursive: true });
const bundle = JSON.parse(fs.readFileSync(process.env.VITICULTURE_WORKER_FIXTURES, 'utf8'));
const data = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-worker-ui-'));
const salt = randomBytes(16);
const password = {
  Salt: salt.toString('hex'),
  Hash: pbkdf2Sync('fixture-password-123', salt, 600000, 32, 'sha256').toString('hex'),
};
bundle.Passwords = {};
for (const session of Object.values(bundle.Sessions)) bundle.Passwords[session.PlayerID] = password;
fs.writeFileSync(path.join(data, 'ee-state-v1.json'), JSON.stringify(bundle));
const result = {
  method: 'isolated seeded fixtures; UI submissions; read-only state assertions',
  status: 'running',
  cases: [],
  errors: [],
};
let child, browser, page;
async function state() {
  return page.evaluate(async () => {
    const response = await fetch('/api/state', {
      headers: { Authorization: 'Bearer ' + sessionStorage.getItem('vineyard-ee-token') },
    });
    if (!response.ok) throw Error(await response.text());
    return response.json();
  });
}
(async () => {
  try {
    child = spawn(
      process.env.VITICULTURE_WORKER_EXE || path.join(repo, 'dist/Viticulture-Expansions.exe'),
      ['-addr', '127.0.0.1:0', '-data', data],
      { windowsHide: true, cwd: repo },
    );
    const base = await new Promise((resolve, reject) => {
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
    browser = await chromium.launch({
      ...(process.platform === 'win32' ? { channel: 'msedge' } : {}),
      headless: true,
    });
    page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    page.on('pageerror', (error) => result.errors.push(error.message));
    for (const test of bundle.cases) {
      await page.goto(base);
      await page.evaluate((token) => sessionStorage.setItem('vineyard-ee-token', token), test.token);
      await page.reload();
      await page.waitForFunction(
        (code) => document.querySelector('#code-label')?.textContent === '房间 ' + code,
        test.code,
      );
      const before = await state();
      const actor = before.players.find((p) => p.id === test.playerId);
      if (test.code === 'TRAINSHORT') {
        const action = page.locator('[data-space="train"]');
        assert.equal(await action.getAttribute('aria-disabled'), 'true');
        assert.match(await action.getAttribute('data-hint'), /还差 1 金币/);
        assert.equal((await state()).revision, before.revision);
        await page.screenshot({ path: path.join(output, test.code + '.png'), fullPage: true });
        result.cases.push({ code: test.code, status: 'passed' });
        console.log('PASS worker interaction', test.code);
        continue;
      }
      await page.locator('[data-space="' + test.space + '"]').click();
      await page.locator('#action-panel').waitFor();
      if (['CHEFBUMP', 'SPECIALONLY', 'FARMERTRAIN'].includes(test.code)) {
        assert.equal(
          await page
            .locator('[data-group="worker"][data-value="normal"]')
            .getAttribute('aria-pressed'),
          'false',
          'special worker selection also highlighted regular worker',
        );
      }
      if (test.code === 'DISCOUNTTRAIN') {
        const choice = page.locator('[data-group="specialWorker"][data-value="chef"]');
        assert.match(await choice.innerText(), /4 金币/);
        await choice.click();
      }
      if (test.code.startsWith('TRAIN')) {
        assert.match(
          await page.locator('[data-group="specialWorker"][data-value="regular"]').innerText(),
          /5 金币（含通行费1）/,
        );
      }
      if (test.code === 'FARMERTRAIN') {
        assert.match(
          await page.locator('[data-group="specialWorker"][data-value="regular"]').innerText(),
          /3 金币（需选农夫折扣）/,
        );
      }
      await page.screenshot({ path: path.join(output, test.code + '-panel.png'), fullPage: true });
      {
        assert.equal(
          await page.locator('#confirm-action').isEnabled(),
          true,
          'legal action was hidden',
        );
        const response = page.waitForResponse(
          (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
        );
        await page.locator('#confirm-action').click();
        const reply = await response;
        assert.equal(reply.status(), 200, await reply.text());
        let after = await reply.json();
        if (test.code === 'FARMERTRAIN') {
          assert.equal(after.pendingChoice?.kind, 'special_farmer');
          const next = page.waitForResponse(
            (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
          );
          await page.locator('[data-choice="discount"]').click();
          const resolved = await next;
          assert.equal(resolved.status(), 200, await resolved.text());
          after = await resolved.json();
        }
        const player = after.players.find((p) => p.id === test.playerId);
        assert.equal(
          after.pendingChoice,
          null,
          'public training or Chef unexpectedly inserted another choice',
        );
        if (test.code === 'DISCOUNTTRAIN') {
          assert.equal(player.coins, 0);
          assert.equal(player.totalWorkers, actor.totalWorkers + 1);
          assert(player.specialWorkers.includes('chef'));
          assert.equal(player.specialWorkerReady.chef, before.year + 1);
        } else if (test.code === 'TRAINEXACT' || test.code === 'FARMERTRAIN') {
          assert.equal(player.coins, 0);
          assert.equal(player.totalWorkers, actor.totalWorkers + 1);
        } else {
          assert.equal(player.coins, actor.coins + 2);
          assert.equal(player.workers, 0);
          assert.equal(player.specialWorkerUsed.chef, true);
          assert.equal(player.largeWorker, false);
        }
      }
      await page.screenshot({ path: path.join(output, test.code + '.png'), fullPage: true });
      result.cases.push({ code: test.code, status: 'passed' });
      console.log('PASS worker interaction', test.code);
    }
    assert.equal(result.cases.length, 6);
    assert.deepEqual(result.errors, []);
    result.status = 'passed';
  } catch (error) {
    result.status = 'failed';
    result.error = error.stack;
    process.exitCode = 1;
    console.error(error);
    if (page)
      await page
        .screenshot({ path: path.join(output, 'failure.png'), fullPage: true })
        .catch(() => {});
  } finally {
    if (browser) await browser.close();
    if (child && child.exitCode == null)
      await new Promise((resolve) => {
        child.once('exit', resolve);
        child.kill();
      });
    fs.writeFileSync(
      path.join(output, 'worker-interactions-result.json'),
      JSON.stringify(result, null, 2),
    );
    console.log('Evidence:', output);
  }
})();
