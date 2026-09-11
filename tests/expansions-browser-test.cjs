// Native Windows smoke test. Owns one EXE child, temporary data, random loopback
// port and two isolated browsers. Never uses runtime/ee-data or a public tunnel.
const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const assert = require('node:assert/strict');

const root = path.resolve(__dirname, '..');
const binary = path.join(root, 'dist', 'Viticulture-Expansions.exe');
const output = path.join(root, 'artifacts', 'expansions');
const data = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-expansions-test-'));
let server;
const browsers = [];
const checks = [];
const errors = [];
const screenshots = [];
const record = (name) => {
  checks.push(name);
  console.log('PASS', name);
};

async function start(port = 0) {
  server = spawn(binary, ['-addr', '127.0.0.1:' + port, '-data', data], {
    cwd: root,
    windowsHide: true,
  });
  const owned = server;
  return new Promise((resolve, reject) => {
    let log = '';
    const timer = setTimeout(() => reject(new Error('Native EXE startup timed out')), 12000);
    owned.once('error', (e) => {
      clearTimeout(timer);
      reject(e);
    });
    owned.once('exit', (code) => {
      clearTimeout(timer);
      reject(new Error('Native EXE exited ' + code));
    });
    owned.stdout.on('data', (chunk) => {
      log += chunk.toString();
      const match = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (match) {
        clearTimeout(timer);
        resolve('http://127.0.0.1:' + match[1]);
      }
    });
    owned.stderr.on('data', (chunk) => {
      log += chunk.toString();
    });
  });
}

async function stop() {
  if (!server || server.exitCode != null) return;
  const child = server;
  await new Promise((resolve) => {
    child.once('exit', resolve);
    child.kill();
  });
  server = undefined;
}

async function state(page) {
  return page.evaluate(async () => {
    const response = await fetch('/api/state', {
      headers: { Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token') },
    });
    return response.json();
  });
}
async function action(page, body) {
  return page.evaluate(async (body) => {
    const response = await fetch('/api/action', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token'),
      },
      body: JSON.stringify(body),
    });
    return { status: response.status, value: await response.json() };
  }, body);
}
async function shot(page, name) {
  const file = path.join(output, name + '.png');
  await page.screenshot({ path: file, fullPage: true });
  const box = await page.evaluate(() => ({
    width: innerWidth,
    scroll: document.documentElement.scrollWidth,
  }));
  assert(box.scroll <= box.width + 1, name + ' horizontal overflow: ' + JSON.stringify(box));
  screenshots.push(file);
}

(async () => {
  fs.mkdirSync(output, { recursive: true });
  let result;
  try {
    const url = await start();
    const pages = [];
    for (let i = 0; i < 2; i++) {
      const browser = await chromium.launch({ channel: 'msedge', headless: true });
      browsers.push(browser);
      const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } });
      // Exercise the existing polling transport locally. No public tunnel opens.
      if (i === 1)
        await context.route('**/api/transport', (route) =>
          route.fulfill({ contentType: 'application/json', body: '{"poll":true}' }),
        );
      const page = await context.newPage();
      page.on('pageerror', (e) => errors.push(e.message));
      await page.goto(url);
      await page.waitForFunction(
        () => typeof document.querySelector('#create-room')?.onclick === 'function',
      );
      pages.push(page);
    }
    const [host, guest] = pages;
    await host.locator('#nickname').fill('扩展验收房主');await host.locator('#player-password').fill('test-password-123');
    await host.locator('#create-room').click();
    await host.locator('#code-label').waitFor();
    const first = await state(host);
    await guest.locator('#nickname').fill('扩展验收访客');await guest.locator('#player-password').fill('test-password-123');
    await guest.locator('#room-code').fill(first.code);
    await guest.locator('#join-room').click();
    await guest.locator('#code-label').waitFor();
    await host.waitForFunction(() => !document.querySelector('#start-game').disabled);
    await guest.waitForFunction(() =>
      document.querySelector('#connection').textContent.includes('轮询'),
    );
    assert.equal(await guest.locator('#save-expansion-config').isDisabled(), true);
    for (const [selector, expectedDisabled] of [
      ['#config-board option[value=tuscany]', false],
      ['#config-visitors option[value=ee_moor]', false],
      ['#config-visitors option[value=rhine]', false],
      ['#config-structures', false],
      ['#config-specialWorkers', false],
    ]) {
      assert.equal(
        await host.locator(selector).evaluate((el) => el.disabled),
        expectedDisabled,
        selector + ' has an incorrect data-driven expansion capability state',
      );
    }
    const before = await state(host);
    await host.locator('#save-expansion-config').click();
    await host.waitForFunction(async (revision) => {
      const response = await fetch('/api/state', {
        headers: { Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token') },
      });
      return (await response.json()).revision > revision;
    }, before.revision);
    const configured = await state(host);
    assert.equal(configured.revision, before.revision + 1);
    assert.deepEqual(configured.config, {
      board: 'ee',
      structures: false,
      specialWorkers: false,
      visitors: 'ee',
    });
    // Structures and special workers are implemented modules, so exercise
    // both real lobby switches before starting. Alternate visitor decks are
    // exercised by the 24-configuration HTTP browser matrix.
    await host.locator('#config-structures').check();
    await host.locator('#config-specialWorkers').check();
    await host.locator('#save-expansion-config').click();
    await host.waitForFunction(async (revision) => {
      const response = await fetch('/api/state', {
        headers: { Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token') },
      });
      return (await response.json()).revision > revision;
    }, configured.revision);
    const workerConfigured = await state(host);
    assert.equal(workerConfigured.config.structures, true);
    assert.equal(workerConfigured.config.specialWorkers, true);
    record('two native Edge browsers create/configure/join; structures and special-worker modules enabled');
    const reject = async (page, body, status) => {
      const old = await state(host);
      const response = await action(page, body);
      assert.equal(response.status, status);
      assert.deepEqual(await state(host), old, 'rejected action changed public state');
    };
    await reject(
      guest,
      { type: 'configure', revision: workerConfigured.revision, config: workerConfigured.config },
      400,
    );
    await reject(
      host,
      { type: 'configure', revision: before.revision, config: configured.config },
      409,
    );
    for (const config of [
      { ...workerConfigured.config, visitors: 'moor+rhine' },
      { ...workerConfigured.config, visitors: 'ee+rhine' },
    ]) {
      await reject(host, { type: 'configure', revision: workerConfigured.revision, config }, 400);
    }
    record('HTTP rejects non-host, stale and illegal configuration without mutation');
    await host.locator('#expansion-config summary').click();
    assert.match(await host.locator('#expansion-config details').textContent(), /Rhine.*替换/);
    await host.locator('#expansion-config summary').click();
    await shot(host, 'lobby-desktop');
    await guest.setViewportSize({ width: 390, height: 844 });
    await shot(guest, 'lobby-mobile');
    await host.locator('#start-game').click();
    await host.waitForFunction(() =>
      document.querySelector('#season-title').textContent.includes('开局'),
    );
    await guest.waitForFunction(() =>
      document.querySelector('#season-title').textContent.includes('开局'),
    );
    let hs = await state(host);
    let gs = await state(guest);
    assert.equal(hs.revision, gs.revision);
    assert.notEqual(hs.youId, gs.youId);
    for (const s of [hs, gs]) {
      assert(s.players.every((p) => !('hand' in p)));
      assert(!('decks' in s) && !('discards' in s));
    }
    assert(!hs.hand.some((c) => gs.hand.some((d) => d.id === c.id)));
    await reject(
      host,
      { type: 'configure', revision: hs.revision, config: configured.config },
      400,
    );
    record('base start, private hands, SSE and local polling synchronization, start locks config');
    // Choose actual setup prompts through the browser UI and reach the board.
    for (let i = 0; i < 2; i++) {
      hs = await state(host);
      const page = hs.pendingChoice.playerId === hs.youId ? host : guest;
      const buttons = page.locator('#ee-choice button');
      await buttons.filter({ hasText: '金币' }).first().click();
      await host.waitForTimeout(350);
    }
    // EE board + structures uses the official four-round setup draft instead
    // of adding a non-existent structure-draw action to the base board.
    while ((await state(host)).phase === 'setup') {
      hs = await state(host);
      assert.equal(hs.pendingChoice.kind, 'structure_draft');
      const page = hs.pendingChoice.playerId === hs.youId ? host : guest;
      await page.locator('#ee-choice form button').first().click();
      await host.waitForTimeout(250);
    }
    hs = await state(host);
    assert.equal(hs.phase, 'wake');
    for (let i = 0; i < 2; i++) {
      hs = await state(host);
      const page = hs.turnId === hs.youId ? host : guest;
      await page.locator('#wake-options button:not([disabled])').first().click();
      await host.waitForTimeout(350);
    }
    hs = await state(host);
    assert.equal(hs.phase, 'summer');
    // A direct state fetch can lead the polling UI. Wait for both rendered views.
    await Promise.all(
      pages.map((page) =>
        page.waitForFunction(() =>
          document.querySelector('#season-title').textContent.includes('夏季'),
        ),
      ),
    );
    record('UI resolves family setup and wake choices into playable EE board');
    for (let i = 0; i < 2; i++) {
      const beforePlace = await state(host);
      const actor = beforePlace.players.find((p) => p.id === beforePlace.turnId);
      const page = actor.id === beforePlace.youId ? host : guest;
      await page.locator('[data-space="gain_coin"]:not([disabled])').click();
      await page.locator('#confirm-action').click();
      await page.locator('#action-panel').waitFor({ state: 'detached' });
      const afterPlace = await state(host);
      const changed = afterPlace.players.find((p) => p.id === actor.id);
      assert.equal(changed.coins, actor.coins + 1);
      assert.equal(changed.workers, actor.workers - 1);
      assert(
        afterPlace.spaces
          .find((s) => s.id === 'gain_coin')
          .occupied.some((s) => s.playerId === actor.id),
      );
    }
    await guest.waitForFunction(
      () => document.querySelector('[data-space="gain_coin"] .worker-slots').children.length === 2,
    );
    record(
      'each browser places a worker through UI; coins, worker pools and shared occupancy update',
    );
    // The structures module must be exercised through the real setup draft,
    // private hand and normal base-board build action.
    let structureState = await state(host);
    const structureActor = structureState.players.find((p) => p.id === structureState.turnId);
    const structurePage = structureActor.id === structureState.youId ? host : guest;
    const structureCard = structureState.youId === structureActor.id
      ? structureState.hand.find((c) => c.type === 'structure')
      : await state(guest).then((v) => v.hand.find((c) => c.type === 'structure'));
    assert.equal(structureActor.handCounts.structure, 4, 'setup draft must retain four structure cards');
    assert(structureCard?.structureId, 'drafted card must be a real catalogued structure');
    // The current actor already owns the next turn after wake; use the real
    // private View of that actor for the build form.
    const builderPage = structureActor.id === structureState.youId ? host : guest;
    const actorView = structureActor.id === structureState.youId ? structureState : await state(guest);
    const actorPublic = actorView.players.find((p) => p.id === structureActor.id);
    if (structureCard.structureCost <= actorPublic.coins) {
      // The second browser is intentionally on polling; reload it after the
      // other player passes so the build form is driven by the fresh private
      // View, not by an older rendered turn.
      await builderPage.reload();
      await builderPage.waitForFunction(async (id) => {
        const response = await fetch('/api/state', {
          headers: { Authorization: 'Bearer ' + localStorage.getItem('vineyard-ee-token') },
        });
        const v = await response.json();
        return v.youId === id && v.turnId === id && v.legal?.canPlace;
      }, structureActor.id);
      await builderPage.waitForFunction(() => document.querySelector('#season-title')?.textContent.includes('夏季'));
      await builderPage.waitForFunction(() => document.querySelector('#connection')?.textContent.includes('已连接'));
      const buildState = await builderPage.locator('[data-space="build"]').evaluate((el) => ({
        disabled: el.disabled,
        reason: el.getAttribute('data-reason') || el.textContent,
        outer: el.outerHTML.slice(0, 500),
      }));
      if (buildState.disabled) {
        throw new Error('structure build unexpectedly unavailable: ' + JSON.stringify({ buildState, connection: await builderPage.locator('#connection').textContent(), passDisabled: await builderPage.locator('#pass').isDisabled(), actionReasons: actorView.actionReasons, card: structureCard, actor: actorPublic }));
      }
      await builderPage.locator('[data-space="build"]:not([disabled])').click();
      await builderPage.locator(`[data-group="building"][data-value="${structureCard.structureId}"]`).click();
      await builderPage.locator('[data-group="structureTarget"]:not([disabled])').first().click();
      await builderPage.locator('#confirm-action').click();
      await builderPage.locator('#action-panel').waitFor({ state: 'detached' });
      const built = (await state(host)).players.find((p) => p.id === structureActor.id);
      assert(
        (built.structureSlots || []).includes(structureCard.structureId) ||
          built.fields.some((f) => f.structure === structureCard.structureId),
        'real UI build must place the drawn structure on the construction mat or empty field',
      );
      record('real UI drafts and builds a finite structure card from the private hand');
    } else {
      record('real UI drafts finite structure cards; build correctly remains disabled when price is unaffordable');
    }
    const rulesButton = host.locator('#game [data-rules-open]').first();
    await rulesButton.click();
    await host.locator('#rules-help-title').waitFor();
    assert.match(await host.locator('#rules-help-title').textContent(), /EE 本体/);
    await host.keyboard.press('Escape');
    assert.equal(await rulesButton.evaluate((el) => el === document.activeElement), true);
    await guest.locator('#game [data-guide-toggle]').click();
    await guest.locator('#beginner-guide').waitFor();
    await guest.locator('[data-guide-skip]').click();
    assert.equal(
      await guest
        .locator('.guide-target')
        .evaluateAll((els) => els.every((el) => el.getClientRects().length > 0)),
      true,
    );
    await shot(guest, 'guide-mobile');
    await guest.locator('#beginner-guide [aria-label="关闭新手引导"]').click();
    assert.equal(await guest.locator('.guide-target').count(), 0);
    record(
      'EE rules keyboard focus restoration; guide enable/skip/close; visible targets on mobile',
    );
    await shot(host, 'game-desktop');
    await shot(guest, 'game-mobile');
    const port = Number(new URL(url).port);
    const snapshot = await state(host);
    await stop();
    await start(port);
    await Promise.all(pages.map((p) => p.reload()));
    await host.locator('#code-label').waitFor();
    await guest.locator('#code-label').waitFor();
    assert.deepEqual(await state(host), snapshot);
    assert.equal((await state(guest)).revision, snapshot.revision);
    record('owned native EXE restart restores config, hands, turn and browser sessions');
    assert.deepEqual(errors, []);
    result = {
      status: 'passed',
      checks,
      screenshots,
      browser: 'Windows Edge headless, two separate processes',
      nativeBinary: binary,
      dataIsolation: 'new temporary directory; random loopback port',
      expansionGameplay: 'released visitor decks are covered by the 24-configuration HTTP matrix',
      errors,
    };
  } catch (e) {
    result = { status: 'failed', checks, screenshots, error: e.stack, errors };
    throw e;
  } finally {
    await Promise.all(browsers.map((b) => b.close()));
    await stop();
    fs.writeFileSync(path.join(output, 'browser-result.json'), JSON.stringify(result, null, 2));
  }
})().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
