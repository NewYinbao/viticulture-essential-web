const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const net = require('node:net');
const { spawn, execFileSync } = require('node:child_process');
const { chromium } = require('playwright');
const { ROOT, GO, temp, executable } = require('./runtime.cjs');
(async () => {
  const dir = temp('viticulture-tabs-');
  let server, browser;
  try {
    const binary = process.env.VITICULTURE_TEST_EXE || path.join(dir, executable('server'));
    if (!process.env.VITICULTURE_TEST_EXE)
      execFileSync(GO, ['build', '-o', binary, './cmd/viticulture'], { cwd: ROOT });
    const port = await new Promise((resolve) => {
      const s = net.createServer();
      s.listen(0, '127.0.0.1', () => {
        const port = s.address().port;
        s.close(() => resolve(port));
      });
    });
    const base = 'http://127.0.0.1:' + port;
    server = spawn(binary, ['-addr', '127.0.0.1:' + port, '-data', path.join(dir, 'data')], {
      stdio: 'ignore',
    });
    let ready = false;
    for (let i = 0; i < 100; i++) {
      try {
        if ((await fetch(base + '/api/health')).ok) {
          ready = true;
          break;
        }
      } catch {}
      await new Promise((r) => setTimeout(r, 100));
    }
    assert.ok(ready);
    browser = await chromium.launch({
      headless: true,
      ...(process.env.PLAYWRIGHT_CHANNEL ? { channel: process.env.PLAYWRIGHT_CHANNEL } : {}),
    });
    const context = await browser.newContext(),
      a = await context.newPage();
    async function state(page) {
      return page.evaluate(async () => {
        const token = sessionStorage.getItem('vineyard-ee-token');
        const r = await fetch('/api/state', { headers: { Authorization: 'Bearer ' + token } });
        if (!r.ok) throw Error(await r.text());
        return r.json();
      });
    }
    async function enter(page, name, code) {
      await page.locator('#nickname').fill(name);
      await page.locator('#player-password').fill('test-password-123');
      if (code) await page.locator('#room-code').fill(code);
      await page.locator(code ? '#join-room' : '#create-room').click();
      await page.locator('#game').waitFor({ state: 'visible' });
    }
    await a.goto(base);
    await enter(a, '标签甲');
    const initialA = await state(a),
      code = initialA.code;
    const next = context.waitForEvent('page');
    await a.locator('#game [data-open-account]').click();
    const b = await next;
    await b.waitForLoadState();
    assert.equal(await b.evaluate(() => window.opener === null), true);
    assert.equal(await b.evaluate(() => sessionStorage.getItem('vineyard-ee-token')), null);
    assert.equal(await b.locator('#room-code').inputValue(), code);
    await enter(b, '标签乙', code);
    const tokenA = await a.evaluate(() => sessionStorage.getItem('vineyard-ee-token')),
      tokenB = await b.evaluate(() => sessionStorage.getItem('vineyard-ee-token'));
    assert.notEqual(tokenA, tokenB);
    assert.equal(await a.evaluate(() => localStorage.getItem('vineyard-ee-token')), null);
    await a.reload();
    await b.reload();
    await a.locator('#game').waitFor({ state: 'visible' });
    await b.locator('#game').waitFor({ state: 'visible' });
    const sa = await state(a),
      sb = await state(b);
    assert.equal(sa.youId, initialA.youId);
    assert.notEqual(sa.youId, sb.youId);
    await a.locator('#start-game').click();
    await a.waitForFunction(
      () => document.querySelector('#season-title').textContent !== '好友集结',
    );
    for (const page of [a, b]) {
      const view = await state(page);
      assert.ok(view.players.every((p) => !Object.hasOwn(p, 'hand')));
      assert.ok(Array.isArray(view.hand));
    }
    const handA = (await state(a)).hand.map((c) => c.id),
      handB = (await state(b)).hand.map((c) => c.id);
    assert.ok(handA.length > 0 && handB.length > 0);
    assert.ok(handA.every((id) => !handB.includes(id)));
    await a.locator('#password-settings').evaluate((e) => (e.open = true));
    await a.locator('#current-password').fill('test-password-123');
    await a.locator('#new-password').fill('new-password-456');
    await a.locator('#confirm-password').fill('new-password-456');
    await a.locator('#password-form button[type=submit]').click();
    await a.waitForFunction((old) => sessionStorage.getItem('vineyard-ee-token') !== old, tokenA);
    assert.equal(await b.evaluate(() => sessionStorage.getItem('vineyard-ee-token')), tokenB);
    assert.equal((await state(b)).youId, sb.youId);
    await b.locator('#lock-session').click();
    await b.locator('#welcome').waitFor({ state: 'visible' });
    assert.equal(await b.evaluate(() => sessionStorage.getItem('vineyard-ee-token')), null);
    assert.equal((await state(a)).youId, sa.youId);
    const migration = await browser.newContext();
    const old = await migration.newPage();
    await old.goto(base);
    const currentToken = await a.evaluate(() => sessionStorage.getItem('vineyard-ee-token'));
    await old.evaluate((t) => {
      localStorage.setItem('vineyard-ee-token', t);
      localStorage.setItem('vineyard-ee-name', '标签甲');
    }, currentToken);
    await old.reload();
    await old.locator('#game').waitFor({ state: 'visible' });
    assert.equal(await old.evaluate(() => localStorage.getItem('vineyard-ee-token')), null);
    assert.equal(
      await old.evaluate(() => sessionStorage.getItem('vineyard-ee-token')),
      currentToken,
    );
    const fresh = await migration.newPage();
    await fresh.goto(base);
    assert.equal(await fresh.evaluate(() => sessionStorage.getItem('vineyard-ee-token')), null);
    await migration.close();
    const out = path.join(ROOT, 'artifacts/multi-account');
    fs.mkdirSync(out, { recursive: true });
    await a.screenshot({ path: path.join(out, 'account-a.png'), fullPage: true });
    console.log(
      'PASS: same-browser independent tabs, noopener, refresh identity, private hands, password rotation/logout isolation, one-time legacy migration.',
    );
  } finally {
    if (browser) await browser.close();
    if (server && server.exitCode === null) {
      await new Promise((resolve) => {
        server.once('exit', resolve);
        server.kill();
      });
    }
  }
})().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
