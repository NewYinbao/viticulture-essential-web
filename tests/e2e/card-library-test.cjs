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
    const visuals = await page.evaluate(async () => {
      const { cardArt } = await import('/js/card-art.js');
      const { cardText } = await import('/js/card-text.js');
      const { localCard } = await import('/js/card-i18n.js');
      const { boardActionArt, actionScenes } = await import('/js/action-art.js');
      const { renderEstate } = await import('/js/graphics.js');
      const groups = await (await fetch('/api/cards')).json();
      const failures = [];
      const parse = (img) =>
        new DOMParser().parseFromString(decodeURIComponent(img.src.split(',')[1]), 'image/svg+xml');
      for (const c of groups.flatMap((g) => g.cards)) {
        const img = cardArt(localCard(c));
        try {
          await img.decode();
        } catch {
          failures.push(c.id + ': decode');
        }
        const xml = parse(img);
        if (xml.querySelector('parsererror')) failures.push(c.id + ': XML');
        if (
          c.type === 'order' &&
          [...xml.querySelectorAll('text')].map((n) => Number(n.textContent)).join() !==
            c.requirements.map((w) => w.value).join()
        )
          failures.push(c.id + ': wine values');
      }
      const scenes = Object.entries(actionScenes).map(([id]) => boardActionArt({ id, name: id }));
      for (const img of scenes) await img.decode();
      const source = '至少2白葡萄，获得3金币；不能 <img src=x onerror=alert(1)>';
      const rich = cardText(source);
      const cellar = renderEstate(
        {
          players: [{ id: 'p', name: '测试庄主' }],
          youId: 'p',
          turnId: 'p',
          hostId: 'p',
          config: {},
        },
        {
          id: 'p',
          name: '测试庄主',
          vp: 0,
          coins: 0,
          income: 0,
          handCount: 0,
          handCounts: {},
          workers: 2,
          totalWorkers: 2,
          largeWorker: false,
          buildings: [],
          fields: [{ index: 0, capacity: 5, vines: [] }],
          grapes: [],
          wines: [],
        },
        {},
        { red: '红', white: '白', blush: '桃红', sparkling: '起泡' },
      );
      return {
        failures,
        sceneCount: new Set(scenes.map((i) => i.src)).size,
        preserved: rich.textContent === source,
        injected: rich.querySelectorAll('img').length,
        bold: [...rich.querySelectorAll('strong')].map((n) => n.textContent),
        numbers: [...rich.querySelectorAll('.card-number')].map((n) => n.textContent),
        cellar: [...cellar.querySelectorAll('.quality-row')].map((row) =>
          [...row.querySelectorAll('.quality-slot')].map((slot) => slot.getAttribute('aria-label')),
        ),
        cellarNote: cellar.querySelector('.cellar-note').textContent,
      };
    });
    assert.deepEqual(visuals.failures, []);
    assert.equal(visuals.sceneCount, 20);
    assert.equal(visuals.preserved, true);
    assert.equal(visuals.injected, 0);
    assert.ok(visuals.bold.includes('白葡萄') && visuals.bold.includes('不能'));
    assert.deepEqual(visuals.numbers, ['2', '3', '1']);
    assert.match(visuals.cellar[4][0], /桃红酒最低品质为 4/);
    assert.match(visuals.cellar[2][3], /缺少中酒窖/);
    assert.match(visuals.cellar[2][6], /需先建中酒窖，再建大酒窖/);
    assert.match(visuals.cellarNote, /中酒窖.*大酒窖/);
    await page.locator('#welcome [data-card-library]').click();
    await page.locator('.library-summary').filter({ hasText: '357' }).waitFor();
    assert.equal(await page.locator('.library-card').count(), 30);
    const libraryNames = await page.locator('.library-card h3').allTextContents();
    assert.equal(new Set(libraryNames).size, libraryNames.length);
    await page.getByRole('button', { name: '葡萄藤 42', exact: true }).click();
    assert.match(await page.locator('.library-summary').innerText(), /筛选结果 42/);
    assert.ok((await page.locator('.library-card').count()) < 30);
    assert.ok((await page.locator('.card-quantity').count()) > 0);
    await page.getByRole('searchbox', { name: '搜索卡牌' }).fill('霞多丽');
    assert.equal(await page.locator('.library-card').count(), 1);
    assert.equal(await page.locator('.card-quantity').count(), 1);
    assert.equal(await page.locator('.card-quantity').innerText(), '×4');
    assert.ok(
      await page
        .locator('.card-quantity')
        .evaluateAll((badges) => badges.every((b) => /^×\d+$/.test(b.textContent))),
    );
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
    await page.waitForTimeout(400);
    const before = await api('/api/state', null, owner.token);
    const building = page.locator('.player.you .structure[data-hint]').first();
    await building.evaluate((element) =>
      element.dispatchEvent(new PointerEvent('pointerover', { bubbles: true })),
    );
    await page.screenshot({ path: path.join(ROOT, 'artifacts/card-features/estate-hint.png') });
    await page.locator('#context-hint:visible').waitFor();
    assert.match(await page.locator('#context-hint').innerText(), /基础费用.*金币/);
    assert.match(await page.locator('#context-hint').innerText(), /葡萄藤/);
    await building.focus();
    assert.equal(await building.getAttribute('aria-describedby'), 'context-hint');
    await page.keyboard.press('Escape');
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
    await page.getByLabel('牌组', { exact: true }).selectOption('ee');
    await page
      .locator('#card-library')
      .getByRole('button', { name: /夏访客/ })
      .click();
    await page
      .locator('.library-art')
      .evaluateAll((imgs) => Promise.all(imgs.map((i) => i.decode())));
    await page.screenshot({ path: path.join(out, 'visitors-desktop.png') });
    await page.setViewportSize({ width: 390, height: 844 });
    await page.screenshot({ path: path.join(out, 'library-mobile.png') });
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
    assert.ok(
      await page.locator('#card-library').evaluate((el) => el.scrollWidth <= el.clientWidth),
    );
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
