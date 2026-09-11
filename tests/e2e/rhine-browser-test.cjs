const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs'),
  path = require('node:path'),
  os = require('node:os'),
  assert = require('node:assert/strict');
const repo = process.env.VITICULTURE_SOURCE_ROOT || path.resolve(__dirname, '../..');
const { authorizeFixtureStore } = require(path.join(repo, 'tests/helpers/fixture-auth.cjs'));
const bundle = authorizeFixtureStore(
  JSON.parse(fs.readFileSync(process.env.RHINE_HTTP_FIXTURES, 'utf8')),
);
const data = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-rhine-http-'));
const out =
  process.env.RHINE_HTTP_OUTPUT ||
  fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-rhine-http-evidence-'));
fs.mkdirSync(out, { recursive: true });
fs.writeFileSync(path.join(data, 'ee-state-v1.json'), JSON.stringify(bundle));
const report = { method: bundle.method, cases: [], pageErrors: [], data, status: 'running' };
let child,
  browser,
  page,
  base,
  port = 0;
async function start() {
  child = spawn(process.env.RHINE_HTTP_EXE, ['-addr', '127.0.0.1:' + port, '-data', data], {
    windowsHide: true,
    cwd: repo,
  });
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timer = setTimeout(() => reject(Error('startup timeout ' + log)), 15000);
    child.once('error', reject);
    child.stdout.on('data', (chunk) => {
      log += chunk;
      const match = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (match) {
        port = Number(match[1]);
        clearTimeout(timer);
        resolve('http://127.0.0.1:' + port);
      }
    });
    child.stderr.on('data', (chunk) => {
      log += chunk;
    });
  });
}
async function stop() {
  if (child && child.exitCode === null)
    await new Promise((resolve) => {
      child.once('exit', resolve);
      child.kill();
    });
}
async function open(token) {
  await page.goto(base);
  await page.evaluate((token) => localStorage.setItem('vineyard-ee-token', token), token);
  await page.reload();
  await page.locator('[data-visitor-submit]').waitFor();
}
async function state(token) {
  const r = await page.request.get(base + '/api/state', {
    headers: { Authorization: 'Bearer ' + token },
  });
  assert.equal(r.status(), 200);
  return r.json();
}
async function submit(selector = '[data-visitor-submit]') {
  const pending = page.waitForResponse(
    (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
  );
  await page.locator(selector).click();
  const r = await pending;
  assert.equal(r.status(), 200, await r.text());
  return r.json();
}
async function branch(value) {
  await page.locator('[data-visitor-option="' + value + '"]').click();
}
async function restart(token) {
  await stop();
  await start();
  await open(token);
}
(async () => {
  try {
    await start();
    browser = await chromium.launch({ channel: 'msedge', headless: true });
    page = await browser.newPage({ viewport: { width: 1440, height: 1100 } });
    page.on('pageerror', (e) => report.pageErrors.push(e.message));
    for (const c of bundle.cases) {
      await open(c.token);
      const before = await state(c.token);
      let after;
      if (c.kind === 'plant') {
        const old = before.hand.map((x) => x.id);
        await branch('draw');
        after = await submit();
        const drawn = after.hand.filter((x) => !old.includes(x.id)).map((x) => x.id);
        assert.equal(drawn.length, 4);
        const choiceId = after.pendingChoice.id;
        const other = await state(c.otherToken);
        assert(
          !drawn.some((id) => JSON.stringify(other).includes(id)),
          'opponent saw privately drawn cards',
        );
        await restart(c.token);
        after = await state(c.token);
        assert.equal(after.pendingChoice.id, choiceId);
        assert.deepEqual(
          after.hand.filter((x) => drawn.includes(x.id)).map((x) => x.id),
          drawn,
        );
        assert.equal(
          await page.locator('#hand [data-card-id="v0"].visitor-eligible').count(),
          0,
          'old vine was selectable',
        );
        for (const id of drawn.slice(0, 2))
          await page.locator('#hand [data-card-id="' + id + '"]').click();
        after = await submit();
        assert.equal(after.hand.length, before.hand.length + 2);
        assert(!after.pendingChoice);
        assert(old.every((id) => after.hand.some((x) => x.id === id)));
      } else if (c.kind === 'virtuoso') {
        await branch('action');
        await page.locator('[data-manager-space="summer_visitor"]').click();
        await page.getByRole('button', { name: '不领取奖励', exact: true }).click();
        await page.locator('#hand [data-card-id="rhine-summer-docent"]').click();
        after = await submit();
        assert.equal(after.pendingChoice.visitor.cardId, 'rhine-summer-docent');
        await restart(c.token);
        await branch('coins');
        after = await submit();
        const p = after.players.find((x) => x.id === c.ownerId),
          original = before.players.find((x) => x.id === c.ownerId);
        assert.equal(p.coins, original.coins + 3);
        assert.equal(p.workers, original.workers);
        assert(!after.pendingChoice);
      } else if (c.kind === 'trainer') {
        await branch('train');
        await page.getByRole('button', { name: /重新培训大工人/ }).click();
        after = await submit();
        const p = after.players.find((x) => x.id === c.ownerId),
          original = before.players.find((x) => x.id === c.ownerId);
        assert.equal(p.coins, original.coins - 3);
        assert(p.largeWorker);
        assert(!p.grandeRemoved);
        assert.equal(p.totalWorkers, original.totalWorkers + 1);
        assert(!after.pendingChoice);
      } else if (c.kind === 'tutor') {
        await branch('exchange');
        await page.getByRole('button', { name: '冬季访客 · 大工人', exact: true }).click();
        after = await submit();
        assert(after.players.find((x) => x.id === c.ownerId).grandeRemoved);
        assert.equal(after.spaces.find((x) => x.id === 'winter_visitor').occupied.length, 0);
        for (const id of ['v0', 'v1'])
          await page.locator('#hand [data-card-id="' + id + '"]').click();
        after = await submit();
        const checkpoint = after.pendingChoice.id;
        await restart(c.token);
        assert.equal((await state(c.token)).pendingChoice.id, checkpoint);
        await page.getByRole('button', { name: '夏访客', exact: true }).first().click();
        await page.getByRole('button', { name: '订单', exact: true }).nth(1).click();
        after = await submit();
        assert.equal(after.hand.length, before.hand.length);
        assert(!after.pendingChoice);
        assert(after.players.find((x) => x.id === c.ownerId).grandeRemoved);
      } else if (c.kind === 'son') {
        const original = before.players.find((x) => x.id === c.ownerId);
        await branch('harvest');
        await page.getByRole('button', { name: '红 1 · 白 1', exact: true }).click();
        after = await submit();
        const p = after.players.find((x) => x.id === c.ownerId);
        assert.equal(p.workers, original.workers);
        assert.equal(after.turnId, c.ownerId);
        assert.equal(p.grapes.length, 2);
        assert(p.fields[0].harvested);
        await page.locator('[data-space="gain_coin"]').click();
        after = await submit('#confirm-action');
        assert.equal(after.players.find((x) => x.id === c.ownerId).workers, original.workers - 1);
        assert(!after.pendingChoice);
      }
      assert(!after.players.some((p) => Object.hasOwn(p, 'hand')), 'public hand leak');
      await page.screenshot({ path: path.join(out, c.code + '.png'), fullPage: true });
      report.cases.push({
        case: c.kind,
        status: 'passed',
        beforeRevision: before.revision,
        afterRevision: after.revision,
      });
      console.log('PASS', c.kind);
    }
    assert.deepEqual(report.pageErrors, []);
    report.status = 'passed';
  } catch (e) {
    report.status = 'failed';
    report.error = e.stack;
    process.exitCode = 1;
    console.error(e);
    if (page)
      await page
        .screenshot({ path: path.join(out, 'failure.png'), fullPage: true })
        .catch(() => {});
  } finally {
    await browser?.close();
    await stop();
    fs.writeFileSync(path.join(out, 'rhine-http-result.json'), JSON.stringify(report, null, 2));
    console.log(out);
  }
})().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
