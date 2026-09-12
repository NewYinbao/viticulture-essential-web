const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('playwright');
const root = path.resolve(__dirname, '../..');
const cards = JSON.parse(
  fs.readFileSync(path.join(root, 'internal/game/cards/ee_cards.json')),
).filter((c) => c.type === 'vine');
(async () => {
  const browser = await chromium.launch({
    headless: true,
    ...(process.env.PLAYWRIGHT_CHANNEL ? { channel: process.env.PLAYWRIGHT_CHANNEL } : {}),
  });
  try {
    const page = await browser.newPage({ viewport: { width: 1200, height: 900 } });
    await page.route('http://vine-art.test/**', async (route) => {
      const pathname = new URL(route.request().url()).pathname;
      if (pathname === '/')
        return route.fulfill({
          contentType: 'text/html',
          body: '<!doctype html><html><meta charset="utf-8"><link rel="stylesheet" href="/css/style.css"><link rel="stylesheet" href="/css/action-panel.css"><link rel="stylesheet" href="/css/vine-art.css"><body><main id="samples" style="display:flex;flex-wrap:wrap;gap:20px;padding:24px"></main></body></html>',
        });
      if (!/^\/(js|css)\/[a-z0-9-]+\.(js|css)$/.test(pathname)) return route.abort();
      return route.fulfill({
        contentType: pathname.endsWith('.js') ? 'text/javascript' : 'text/css',
        body: fs.readFileSync(path.join(root, 'web/static', pathname)),
      });
    });
    await page.goto('http://vine-art.test/');
    const results = await page.evaluate(async (cards) => {
      const { cardArt } = await import('/js/card-art.js');
      const { vineDescription } = await import('/js/vine-art.js');
      const results = [];
      for (const c of cards) {
        const image = cardArt(c, 'card-art');
        await image.decode();
        const svg = new DOMParser().parseFromString(
          decodeURIComponent(image.src.split(',').slice(1).join(',')),
          'image/svg+xml',
        );
        const numbers = [...svg.querySelectorAll('[data-resource]')].map((g) => [
          g.getAttribute('data-resource'),
          g.textContent,
        ]);
        const p = vineDescription(c.description);
        results.push({
          id: c.id,
          loaded: image.naturalWidth > 0,
          invalid: !!svg.querySelector('parsererror'),
          trellis: !!svg.querySelector('[data-layer="trellis"]'),
          irrigation: !!svg.querySelector('[data-layer="irrigation"]'),
          numbers,
          text: p.textContent,
          emphasis: p.querySelectorAll('strong').length,
          circlesLeft: [...svg.querySelectorAll('[data-resource] circle')].every(
            (n) => Number(n.getAttribute('cx')) < 200,
          ),
        });
        if (['vine-01', 'vine-09', 'vine-25', 'vine-39'].includes(c.id)) {
          const article = document.createElement('article');
          article.className = 'card vine';
          const title = document.createElement('h4');
          title.textContent = c.name;
          article.append(image, title, p);
          document.querySelector('#samples').append(article);
          const button = document.createElement('button');
          button.className = 'action-choice hand-choice vine';
          button.style.width = '150px';
          button.append(cardArt(c, 'choice-art'));
          document.querySelector('#samples').append(button);
        }
      }
      const hostile = '<img src=x onerror=alert(1)> 白葡萄*2';
      const p = vineDescription(hostile);
      results.push({
        safe: p.textContent === hostile && !p.querySelector('img'),
        quantity: [...p.querySelectorAll('.vine-quantity')].at(-1).textContent,
      });
      return results;
    }, cards);
    const safety = results.pop();
    assert.deepEqual(safety, { safe: true, quantity: '2' });
    for (let i = 0; i < cards.length; i++) {
      const c = cards[i],
        r = results[i];
      assert.equal(r.loaded, true, c.id);
      assert.equal(r.invalid, false, c.id);
      assert.equal(r.trellis, c.trellis, c.id);
      assert.equal(r.irrigation, c.irrigation, c.id);
      assert.equal(r.text, c.description, c.id);
      assert.ok(r.emphasis >= 2, c.id);
      assert.equal(r.circlesLeft, true, c.id);
      assert.deepEqual(
        r.numbers,
        [
          ['red', c.red],
          ['white', c.white],
        ]
          .filter(([, n]) => n > 0)
          .map(([k, n]) => [k, String(n)]),
        c.id,
      );
    }
    await page.locator('img').evaluateAll((imgs) => Promise.all(imgs.map((i) => i.decode())));
    const ratios = await page.locator('.vine-art').evaluateAll((imgs) =>
      imgs.map((i) => {
        const r = i.getBoundingClientRect();
        return r.width / r.height;
      }),
    );
    assert.ok(ratios.every((r) => Math.abs(r - 1.5) < 0.04));
    const out = path.join(root, 'artifacts/vine-art');
    fs.mkdirSync(out, { recursive: true });
    await page.screenshot({ path: path.join(out, 'cards.png'), fullPage: true });
    await page.setViewportSize({ width: 390, height: 844 });
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
    console.log(
      'PASS: 42 vine cards, exact resource values and prerequisites, left circles, safe rich text, hand/action aspect ratios, 390px layout.',
    );
  } finally {
    await browser.close();
  }
})().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
