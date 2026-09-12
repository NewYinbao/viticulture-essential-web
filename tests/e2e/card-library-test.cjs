const assert = require('node:assert/strict'),
  fs = require('node:fs'),
  path = require('node:path'),
  net = require('node:net');
const { spawn, execFileSync } = require('node:child_process');
const { chromium } = require('playwright');
const { ROOT, GO, temp, executable } = require('./runtime.cjs');
(async () => {
  const dir = temp('viticulture-library-');
  let server, browser;
  try {
    const bin = process.env.VITICULTURE_TEST_EXE || path.join(dir, executable('server'));
    if (!process.env.VITICULTURE_TEST_EXE)
      execFileSync(GO, ['build', '-o', bin, './cmd/viticulture'], { cwd: ROOT });
    const port = await new Promise((r) => {
      const s = net.createServer();
      s.listen(0, '127.0.0.1', () => {
        const p = s.address().port;
        s.close(() => r(p));
      });
    });
    const base = 'http://127.0.0.1:' + port;
    server = spawn(bin, ['-addr', '127.0.0.1:' + port, '-data', path.join(dir, 'data')], {
      stdio: 'ignore',
      windowsHide: true,
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
    const groups = await (await fetch(base + '/api/cards')).json();
    assert.deepEqual(
      groups.map((g) => g.cards.length),
      [190, 40, 80, 36, 11],
    );
    assert.equal(groups.flatMap((g) => g.cards).filter((c) => c.requiresTuscany).length, 4);
    browser = await chromium.launch({
      headless: true,
      ...(process.env.PLAYWRIGHT_CHANNEL ? { channel: process.env.PLAYWRIGHT_CHANNEL } : {}),
    });
    const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
    const errors = [];
    page.on('pageerror', (e) => errors.push(e.message));
    await page.goto(base);
    await page.locator('#welcome [data-card-library]').click();
    await page.locator('.library-summary').filter({ hasText: '357' }).waitFor();
    assert.equal(await page.locator('.library-card').count(), 30);
    await page.getByRole('button', { name: '葡萄藤 42', exact: true }).click();
    assert.match(await page.locator('.library-summary').innerText(), /筛选结果 42/);
    await page.getByRole('button', { name: '下一页', exact: true }).click();
    assert.equal(await page.locator('.library-card').count(), 12);
    await page.getByRole('searchbox', { name: '搜索卡牌' }).fill('霞多丽');
    assert.equal(await page.locator('.library-card').count(), 4);
    await page.keyboard.press('Escape');
    assert.ok(await page.locator('#card-library').isHidden());
    async function api(route, body, token) {
      const r = await fetch(base + route, {
        method: body ? 'POST' : 'GET',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: 'Bearer ' + token } : {}),
        },
        ...(body ? { body: JSON.stringify(body) } : {}),
      });
      assert.equal(r.status, 200);
      return r.json();
    }
    const owner = await api('/api/create', { name: '图鉴房主', password: 'password-123' });
    await api('/api/join', { name: '图鉴访客', password: 'password-123', code: owner.code });
    const lobby = await api('/api/state', null, owner.token);
    await api('/api/action', { type: 'start', revision: lobby.revision }, owner.token);
    await page.evaluate((token) => sessionStorage.setItem('vineyard-ee-token', token), owner.token);
    await page.reload();
    await page.locator('#game').waitFor({ state: 'visible' });
    const before = await api('/api/state', null, owner.token);
    await page.locator('#game [data-card-library]').click();
    await page.locator('.library-summary').filter({ hasText: '190' }).waitFor();
    assert.equal(await page.getByLabel('图鉴范围').inputValue(), 'current');
    await page.getByLabel('图鉴范围').selectOption('all');
    assert.match(await page.locator('.library-summary').innerText(), /357/);
    await page.getByLabel('牌组', { exact: true }).selectOption('structures');
    assert.match(await page.locator('.library-summary').innerText(), /筛选结果 36/);
    await page
      .locator('.library-art')
      .evaluateAll((imgs) => Promise.all(imgs.map((i) => i.decode())));
    const out = path.join(ROOT, 'artifacts/card-features');
    fs.mkdirSync(out, { recursive: true });
    await page.screenshot({ path: path.join(out, 'library-desktop.png') });
    await page.setViewportSize({ width: 390, height: 844 });
    await page.screenshot({ path: path.join(out, 'library-mobile.png') });
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
    await page.keyboard.press('Escape');
    const after = await api('/api/state', null, owner.token);
    assert.equal(after.revision, before.revision);
    assert.deepEqual(errors, []);
    console.log(
      'PASS: immutable 357-card library, category totals, search/pagination, pregame/in-game entry, current/all scopes, mobile, zero gameplay mutation.',
    );
  } finally {
    if (browser) await browser.close();
    if (server && server.exitCode === null)
      await new Promise((r) => {
        server.once('exit', r);
        server.kill();
      });
  }
})().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
