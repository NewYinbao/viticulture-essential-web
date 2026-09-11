const { ROOT, GO, temp, executable, sourceHashes } = require('./runtime.cjs');
const { chromium } = require('playwright');
const { execFileSync, spawn } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const net = require('node:net');
const assert = require('node:assert/strict');
const tmp = temp('viticulture-password-');
const out = process.env.VITICULTURE_PASSWORD_OUTPUT || path.join(ROOT, 'artifacts/continuation-20260911/password');
const binary = process.env.VITICULTURE_EXE || path.join(tmp, executable('server'));
const password = 'private-cellar-2026';
const checks = [], errors = [];
let server, browser, base;
const enrollmentCodes = new Map();
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
async function start() {
  server = spawn(binary, ['-addr', new URL(base).host, '-data', path.join(tmp, 'data')], { windowsHide: true, stdio: ['ignore', 'pipe', 'ignore'] });
  // This reads only our owned test server console; codes are never requested
  // through HTTP or written to the browser-evidence report.
  let localOutput = '';
  server.stdout.setEncoding('utf8');
  server.stdout.on('data', (text) => {
    localOutput += text;
    for (const match of localOutput.matchAll(/旧座位验证码 ([a-f0-9]{12}) \/ (\w+) \/ (.*?)（/g)) enrollmentCodes.set(match[2] + ':' + match[3], match[1]);
  });
  for (let i = 0; i < 100; i++) {
    try { if ((await fetch(base + '/api/health')).ok) return; } catch {}
    await delay(50);
  }
  throw new Error('Server failed to start');
}
async function stop() {
  if (server?.exitCode === null) await new Promise((resolve) => { server.once('exit', resolve); server.kill(); });
  server = null;
}
async function state(page) {
  return page.evaluate(async () => (await fetch('/api/state', { headers: { Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token') } })).json());
}
async function submit(page, url, selector, expected = 200) {
  const response = page.waitForResponse((r) => r.url().endsWith(url) && r.request().method() === 'POST');
  await page.locator(selector).click();
  const r = await response;
  const result = await r.json();
  assert.equal(r.status(), expected, JSON.stringify(result));
  return result;
}
async function login(page, name, code, secret, expected = 200) {
  await page.locator('#nickname').fill(name);
  await page.locator('#room-code').fill(code);
  await page.locator('#player-password').fill(secret);
  return submit(page, '/api/join', '#join-room', expected);
}
async function shot(page, name) {
  await page.screenshot({ path: path.join(out, name + '.png'), fullPage: true });
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), 'horizontal overflow');
}
(async () => {
  fs.mkdirSync(out, { recursive: true });
  let failure;
  try {
    if (!process.env.VITICULTURE_EXE) execFileSync(GO, ['build', '-o', binary, './cmd/viticulture'], { cwd: ROOT });
    const port = await new Promise((resolve) => { const socket = net.createServer(); socket.listen(0, '127.0.0.1', () => { const port = socket.address().port; socket.close(() => resolve(port)); }); });
    base = 'http://127.0.0.1:' + port;
    await start();
    browser = await chromium.launch({ headless: true, ...(process.platform === 'win32' ? { channel: 'msedge' } : { args: ['--no-sandbox'] }) });
    const pages = [];
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
      const page = await context.newPage();
      page.on('pageerror', (e) => errors.push(e.message));
      await page.goto(base);
      await page.waitForFunction(() => typeof document.querySelector('#create-room')?.onclick === 'function');
      pages.push(page);
    }
    const [host, guest, stranger] = pages;
    await host.locator('#nickname').fill('山丘庄主');
    await host.locator('#create-room').click();
    await host.locator('#toast').filter({ hasText: '密码需要' }).waitFor();
    await host.locator('#player-password').fill(password);
    const created = await submit(host, '/api/create', '#create-room');
    await host.locator('#game').waitFor({ state: 'visible' });
    await login(guest, '河谷庄主', created.code, 'guest-cellar-2026');
    await guest.locator('#game').waitFor({ state: 'visible' });
    await host.waitForFunction(() => !document.querySelector('#start-game').disabled);
    await submit(host, '/api/action', '#start-game');
    const original = await state(host);
    assert(original.hand.length > 0 && original.passwordSet);
    await login(stranger, '山丘庄主', created.code, 'incorrect-password', 401);
    assert.equal(await stranger.locator('#game').isVisible(), false);
    assert.equal(await stranger.evaluate(() => localStorage.getItem('vineyard-ee-token')), null);
    assert.deepEqual(await state(host), original);
    checks.push('password required; same-name intruder rejected without seat mutation or private hand');
    await login(stranger, '山丘庄主', created.code, password);
    await stranger.locator('#game').waitFor({ state: 'visible' });
    await host.locator('#welcome').waitFor({ state: 'visible' });
    assert.equal(await host.locator('#hand .card').count(), 0);
    assert.equal(await host.evaluate(() => localStorage.getItem('vineyard-ee-token')), null);
    assert.deepEqual((await state(stranger)).hand, original.hand);
    assert.equal((await state(stranger)).youId, original.youId);
    checks.push('correct password resumes started game; old SSE browser locks and clears hand');
    await stranger.locator('#password-settings summary').click();
    await stranger.locator('#current-password').fill(password);
    await stranger.locator('#new-password').fill('new-cellar-password');
    await stranger.locator('#confirm-password').fill('mismatch-password');
    await stranger.locator('#password-form button').click();
    await stranger.locator('#password-error').filter({ hasText: '不一致' }).waitFor();
    await stranger.locator('#confirm-password').fill('new-cellar-password');
    await submit(stranger, '/api/password', '#password-form button');
    await stranger.locator('#toast').filter({ hasText: '密码已保存' }).waitFor();
    assert.deepEqual((await state(stranger)).hand, original.hand);
    const saved = await stranger.evaluate(() => ({ ...localStorage }));
    assert(!JSON.stringify(saved).includes('new-cellar-password'));
    assert.equal(await stranger.locator('#new-password').inputValue(), '');
    await submit(stranger, '/api/logout', '#lock-session');
    await stranger.locator('#welcome').waitFor({ state: 'visible' });
    await login(stranger, '山丘庄主', created.code, password, 401);
    await login(stranger, '山丘庄主', created.code, 'new-cellar-password');
    await stranger.locator('#game').waitFor({ state: 'visible' });
    checks.push('password confirmation/change, logout and new-password recovery; no plaintext storage');
    await stop(); await start(); await stranger.reload();
    await stranger.locator('#game').waitFor({ state: 'visible' });
    assert.deepEqual((await state(stranger)).hand, original.hand);
    checks.push('owned server restart preserves password/session/hand');
    await stranger.setViewportSize({ width: 390, height: 844 });
    await stranger.locator('#password-settings summary').click();
    await shot(stranger, 'password-mobile');
    await host.setViewportSize({ width: 1440, height: 1000 });
    await shot(host, 'entry-desktop');
    // Only the test's own save is edited to emulate an old passwordless seat.
    await stop();
    const file = path.join(tmp, 'data/ee-state-v1.json');
    const snapshot = JSON.parse(fs.readFileSync(file));
    delete snapshot.Passwords[original.youId];
    fs.writeFileSync(file, JSON.stringify(snapshot));
    await start(); await stranger.reload();
    await stranger.locator('#password-settings[open]').waitFor();
    assert.equal(await stranger.locator('#game').isVisible(), false);
    assert.equal(await stranger.locator('#hand .card').count(), 0);
    assert.equal(await stranger.locator('#legacy-access').isVisible(), true);
    assert.equal(await stranger.locator('#current-password').isVisible(), false);
    await login(host, '山丘庄主', created.code, 'old-seat-password', 403);
    await stranger.locator('#new-password').fill('old-seat-password');
    await stranger.locator('#confirm-password').fill('old-seat-password');
    await stranger.locator('#enrollment-code').fill('wrong-code');
    await submit(stranger, '/api/password', '#password-form button', 403);
    assert.equal(await stranger.locator('#game').isVisible(), false);
    await stranger.locator('#enrollment-code').fill(enrollmentCodes.get(created.code + ':山丘庄主'));
    await shot(stranger, 'legacy-enrollment-mobile');
    await submit(stranger, '/api/password', '#password-form button');
    await stranger.locator('#toast').filter({ hasText: '密码已保存' }).waitFor();
    await login(host, '山丘庄主', created.code, 'old-seat-password');
    await host.locator('#game').waitFor({ state: 'visible' });
    checks.push('legacy tokens see no hand; local one-time code plus original session enrolls and revokes old sessions');
    // Polling must also stop permanently when a session is revoked.
    await guest.context().route('**/api/transport', (route) => route.fulfill({ contentType: 'application/json', body: '{"poll":true}' }));
    await guest.reload(); await guest.locator('#connection').filter({ hasText: '轮询' }).waitFor();
    await stranger.locator('#welcome').waitFor({ state: 'visible' });
    await login(stranger, '河谷庄主', created.code, 'guest-cellar-2026');
    await guest.locator('#welcome').waitFor({ state: 'visible' });
    assert.equal(await guest.evaluate(() => localStorage.getItem('vineyard-ee-token')), null);
    checks.push('polling session revocation clears private UI and returns to login');
    assert.deepEqual(errors, []);
    console.log('PASS', checks);
  } catch (error) { failure = error.stack; throw error; }
  finally {
    await browser?.close(); await stop();
    fs.writeFileSync(path.join(out, 'browser-result.json'), JSON.stringify({ status: failure ? 'failed' : 'passed', checks, errors, failure, binary, sourceHashes: sourceHashes(), testData: tmp }, null, 2));
  }
})().catch((error) => { console.error(error); process.exitCode = 1; });
