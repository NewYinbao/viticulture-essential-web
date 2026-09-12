// Real, unseeded create/join/configure/start and setup actions through two browsers.
// This covers configuration/help lifecycle, not natural endings or every card.
const { chromium } = require('playwright');
const { spawn } = require('node:child_process');
const { createHash } = require('node:crypto');
const fs = require('node:fs'),
  path = require('node:path'),
  os = require('node:os');
const assert = require('node:assert/strict');
const repo = path.resolve(__dirname, '../..');
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-config-http-'));
const output = process.env.VITICULTURE_CONFIG_OUTPUT || temporary;
const binary =
  process.env.VITICULTURE_CONFIG_EXE || path.join(repo, 'dist/Viticulture-Expansions.exe');
fs.mkdirSync(output, { recursive: true });
const result = {
  method:
    'unseeded HTTP server; real UI create/join/configure/start/setup; two isolated Edge contexts, desktop and mobile',
  limitations: [
    'Configuration lifecycle and setup coverage; visitor/card effects and natural endings are separate suites.',
    'Screenshots need visual inspection in addition to DOM assertions.',
  ],
  binary,
  binarySHA256: createHash('sha256').update(fs.readFileSync(binary)).digest('hex'),
  cases: [],
  errors: [],
  rejected: 0,
  screenshots: [],
};
let child, browser, base, current, pureRoom;
async function start() {
  child = spawn(binary, ['-addr', '127.0.0.1:0', '-data', path.join(temporary, 'data')], {
    cwd: repo,
    windowsHide: true,
  });
  base = await new Promise((resolve, reject) => {
    let log = '';
    const timeout = setTimeout(() => reject(Error('Server startup timeout: ' + log)), 15000);
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
}
async function state(token) {
  const response = await fetch(base + '/api/state', {
    headers: { Authorization: 'Bearer ' + token },
  });
  assert.equal(response.status, 200, await response.clone().text());
  return response.json();
}
async function uiAction(page, action) {
  const response = page.waitForResponse(
    (r) => r.url().endsWith('/api/action') && r.request().method() === 'POST',
  );
  await action();
  const reply = await response;
  assert.equal(reply.status(), 200, await reply.text());
  return reply.json();
}
async function rejectUnchanged(tokens, actor, payload, expected) {
  const before = await Promise.all(tokens.map(state));
  const response = await fetch(base + '/api/action', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + tokens[actor] },
    body: JSON.stringify(payload),
  });
  assert.equal(response.status, expected, await response.text());
  assert.deepEqual(
    await Promise.all(tokens.map(state)),
    before,
    'rejected config changed a private/public View',
  );
  result.rejected++;
}
async function screenshot(page, name) {
  const file = name + '.png';
  await page.screenshot({ path: path.join(output, file), fullPage: false });
  result.screenshots.push(file);
}
async function noInvisibleHighlight(page) {
  assert.deepEqual(
    await page
      .locator('.guide-target')
      .evaluateAll((elements) =>
        elements
          .filter(
            (el) =>
              !el.getClientRects().length ||
              el.closest('[hidden]') ||
              getComputedStyle(el).visibility === 'hidden',
          )
          .map((el) => el.id),
      ),
    [],
  );
}
async function inspectRules(page, config, label, takeShot = false) {
  await page.locator('#game [data-rules-open]').click();
  const panel = page.locator('#rules-help');
  await panel.waitFor({ state: 'visible' });
  const ids = await panel
    .locator('[data-topic]')
    .evaluateAll((es) => es.map((e) => e.dataset.topic));
  for (const [topic, on] of [
    ['influence', config.board === 'tuscany'],
    ['trade', config.board === 'tuscany'],
    ['actions', config.board === 'tuscany'],
    ['structures', config.structures],
    ['special-workers', config.specialWorkers],
    ['visitors', config.visitors !== 'ee'],
  ])
    assert.equal(ids.includes(topic), on, label + ': topic ' + topic);
  const texts = {};
  for (const id of ids) {
    await panel.locator(`[data-topic="${id}"]`).click();
    texts[id] = await panel.innerText();
    assert.doesNotMatch(texts[id], /逐牌说明尚未可用|仍禁止开局|仍按未完成状态禁用/);
  }
  assert.match(texts.overview, config.board === 'tuscany' ? /25 分/ : /20 分/);
  if (config.board === 'tuscany') assert.match(texts.seasons, /个人|全员/);
  if (config.visitors === 'rhine') assert.match(texts.visitors, /76张.*80张/);
  if (config.visitors === 'ee_moor') assert.match(texts.visitors, /40张.*76张/);
  const links = await panel.locator('.help-foot a').evaluateAll((es) => es.map((e) => e.href));
  assert.equal(
    links.some((s) => s.includes('/moor-visitors-expansion')),
    config.visitors === 'ee_moor',
  );
  assert.equal(
    links.some((s) => s.includes('rhine')),
    config.visitors === 'rhine',
  );
  await panel
    .locator(`[data-topic="${config.visitors !== 'ee' ? 'visitors' : 'overview'}"]`)
    .click();
  await panel.scrollIntoViewIfNeeded();
  if (takeShot) await screenshot(page, label + '-rules');
  await page.keyboard.press('Escape');
  await panel.waitFor({ state: 'hidden' });
  assert.equal(
    await page.evaluate(() => document.activeElement?.matches('#game [data-rules-open]')),
    true,
    'rules close did not restore opener focus',
  );
  await noInvisibleHighlight(page);
  return { ids, sources: links };
}
async function inspectGame(page, config, label) {
  assert.equal(await page.locator('#expansion-config').count(), 0, 'lobby editor survived start');
  assert.equal(await page.locator('#tuscany-board').count(), config.board === 'tuscany' ? 1 : 0);
  assert.equal(
    await page.locator('#hand-filters [data-filter="structure"]').count(),
    config.structures ? 1 : 0,
  );
  assert.equal(
    (await page.locator('.ee-deck-counts').textContent()).includes('结构'),
    config.structures,
  );
  if (!config.specialWorkers)
    assert.equal(await page.locator('.special-worker-reserve').count(), 0);
  assert.equal(
    (await page.locator('.ee-family').allTextContents()).join(' ').includes('本局特殊工人'),
    config.specialWorkers,
  );
  if (!config.structures) assert.equal(await page.locator('[data-space^="structure_"]').count(), 0);
  assert.equal(
    await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1),
    true,
    label + ': horizontal page overflow',
  );
  await noInvisibleHighlight(page);
}
async function chooseSetup(pages, tokens, record) {
  for (let count = 0; count < 20; count++) {
    const views = await Promise.all(tokens.map(state));
    const pending = views[0].pendingChoice;
    if (!pending && views[0].phase === 'wake') return;
    assert(pending, 'unexpected setup state ' + views[0].phase);
    const seat = views.findIndex((v) => v.youId === pending.playerId),
      own = views[seat],
      other = views[1 - seat];
    assert.deepEqual(Object.keys(other.pendingChoice).sort(), ['kind', 'playerId']);
    const choice = own.pendingChoice;
    assert(['papa', 'structure_draft'].includes(choice.kind), choice.kind);
    const option = choice.kind === 'papa' ? 'coins' : choice.options[0];
    const locator = pages[seat].locator(`#ee-choice [data-choice="${option}"]`);
    await locator.waitFor({ state: 'visible' });
    await pages[seat].locator('[data-guide-rules]').click();
    await pages[seat].locator('#rules-help [data-topic="current"]').waitFor();
    await pages[seat].keyboard.press('Escape');
    await pages[seat].locator('[data-guide-locate]').click();
    assert.equal(
      await pages[seat].evaluate(() => document.activeElement?.matches('#ee-choice [data-choice]')),
      true,
      'setup guide failed to locate actual choice',
    );
    await uiAction(pages[seat], () => locator.click());
    record.setup.push(choice.kind);
  }
  throw Error('setup failed to finish within 20 genuine choices');
}
async function run(config, index) {
  current = { config, index, setup: [], views: [], rules: [] };
  result.cases.push(current);
  const contexts = await Promise.all(
    [
      { width: 1440, height: 1000 },
      { width: 390, height: 844 },
    ].map((viewport) => browser.newContext({ viewport })),
  );
  const pages = await Promise.all(contexts.map((c) => c.newPage()));
  for (const [seat, page] of pages.entries()) {
    page.on('pageerror', (error) => result.errors.push({ index, seat, message: error.message }));
    page.on('dialog', (dialog) => dialog.accept());
    await page.goto(base);
    await page.locator('#nickname').fill(`配置验收${index}-${seat}`);
    await page.locator('#player-password').fill('configuration-test-123');
  }
  await pages[0].locator('#create-room').click();
  await pages[0].locator('#code-label').waitFor();
  const hostToken = await pages[0].evaluate(() => sessionStorage.getItem('vineyard-ee-token'));
  const initial = await state(hostToken);
  if (index === 0) pureRoom = { code: initial.code, name: '配置验收0-1' };
  await pages[1].locator('#room-code').fill(initial.code);
  await pages[1].locator('#join-room').click();
  await pages[1].locator('#code-label').waitFor();
  const tokens = [
    hostToken,
    await pages[1].evaluate(() => sessionStorage.getItem('vineyard-ee-token')),
  ];
  for (const page of pages) {
    await page.locator('#game [data-guide-toggle]').click();
    await page.locator('#beginner-guide').waitFor({ state: 'visible' });
  }
  await pages[0].waitForFunction(() => !document.querySelector('#start-game').disabled);
  const before = await state(tokens[0]);
  await rejectUnchanged(tokens, 1, { type: 'configure', config, revision: before.revision }, 400);
  await rejectUnchanged(
    tokens,
    0,
    { type: 'configure', config: { ...config, visitors: 'moor+rhine' }, revision: before.revision },
    400,
  );
  await pages[0].locator('#config-board').selectOption(config.board);
  await pages[0].locator('#config-visitors').selectOption(config.visitors);
  await pages[0].locator('#config-structures').setChecked(config.structures);
  await pages[0].locator('#config-specialWorkers').setChecked(config.specialWorkers);
  const configured = await uiAction(pages[0], () =>
    pages[0].locator('#save-expansion-config').click(),
  );
  assert.deepEqual(configured.config, config);
  await pages[1].waitForFunction(
    (c) =>
      document.querySelector('#config-board')?.value === c.board &&
      document.querySelector('#config-visitors')?.value === c.visitors &&
      document.querySelector('#config-structures')?.checked === c.structures &&
      document.querySelector('#config-specialWorkers')?.checked === c.specialWorkers,
    config,
  );
  assert.equal(await pages[1].locator('#save-expansion-config').isDisabled(), true);
  for (const [seat, page] of pages.entries()) {
    assert.deepEqual((await state(tokens[seat])).config, config);
    current.rules.push(await inspectRules(page, config, `${index}-lobby-${seat}`));
  }
  await rejectUnchanged(tokens, 0, { type: 'configure', config, revision: before.revision }, 409);
  const started = await uiAction(pages[0], () => pages[0].locator('#start-game').click());
  await rejectUnchanged(tokens, 0, { type: 'configure', config, revision: started.revision }, 400);
  await chooseSetup(pages, tokens, current);
  assert.equal(
    current.setup.filter((x) => x === 'structure_draft').length,
    config.board === 'ee' && config.structures ? 8 : 0,
  );
  for (let n = 0; n < 2; n++) {
    const views = await Promise.all(tokens.map(state)),
      seat = views.findIndex((v) => v.youId === views[0].turnId);
    const slot = views[seat].wakeSlots.find((w) => w.slot >= 2 && !w.playerId).slot;
    await pages[seat].locator('#wake-options button:not(:disabled)').first().waitFor();
    // Choose safe non-choice wake rows, avoiding extra pending choices in EE.
    const wake = pages[seat]
      .locator(`#wake-options button`)
      .filter({ hasText: new RegExp('^' + slot + '(?:\\s|\\D)') });
    const buttons = await pages[seat]
      .locator('#wake-options button')
      .evaluateAll((es) => es.map((e) => ({ text: e.textContent, slot: e.dataset.slot })));
    const bySlot = pages[seat].locator(`#wake-options button[data-slot="${slot}"]`);
    await uiAction(pages[seat], () =>
      bySlot.count().then((count) => (count ? bySlot.click() : wake.first().click())),
    );
    current.wakeButtons = buttons;
  }
  const views = await Promise.all(tokens.map(state));
  for (let seat = 0; seat < 2; seat++) {
    const view = views[seat];
    assert.equal(
      view.players.every((p) => !Object.hasOwn(p, 'hand')),
      true,
    );
    assert.equal(Object.hasOwn(view, 'decks'), false);
    assert.equal(view.specialWorkerPool?.length || 0, config.specialWorkers ? 2 : 0);
    if (!config.structures)
      assert.equal(
        view.hand.some((c) => c.type === 'structure'),
        false,
      );
    for (const card of view.hand.filter((c) => ['summer', 'winter'].includes(c.type))) {
      assert.equal(card.implemented, true);
      assert.equal(card.id.startsWith('rhine-'), config.visitors === 'rhine');
      if (config.visitors === 'ee') assert.equal(card.id.startsWith('moor-'), false);
    }
    current.views.push({
      phase: view.phase,
      revision: view.revision,
      deckCounts: view.deckCounts,
      handCounts: view.players.find((p) => p.id === view.youId).handCounts,
    });
    await inspectGame(pages[seat], config, `${index}-${seat}`);
    await inspectRules(pages[seat], config, `${index}-started-${seat}`, seat === 0);
  }
  const active = views.findIndex((v) => v.youId === views[0].turnId),
    actor = pages[active];
  // The active player gets tour lessons; the other gets waiting guidance.
  if (await actor.locator('[data-guide-restart]').count())
    await actor.locator('[data-guide-restart]').click();
  while (Number(await actor.locator('#beginner-guide').getAttribute('data-step')) > 0)
    await actor.locator('[data-guide-previous]').click();
  current.firstLesson = await actor.locator('#beginner-guide .guide-copy').innerText();
  assert.equal(current.firstLesson.includes('四季都有工人行动'), config.board === 'tuscany');
  await actor.locator('[data-guide-next]').click();
  await actor.locator('[data-guide-next]').click();
  current.handLesson = await actor.locator('#beginner-guide .guide-copy').innerText();
  assert.equal(current.handLesson.includes('橙色建筑牌'), config.structures);
  await actor.locator('[data-guide-skip]').click();
  await actor.locator('[data-space="gain_coin"]').click();
  await actor.locator('#confirm-action').waitFor();
  await actor.locator('[data-guide-locate]').click();
  assert.equal(await actor.evaluate(() => document.activeElement?.id), 'confirm-action');
  const coinsBefore = views[active].players.find((p) => p.id === views[active].youId).coins;
  const after = await uiAction(actor, () => actor.locator('#confirm-action').click());
  assert.equal(after.players.find((p) => p.id === after.youId).coins, coinsBefore + 1);
  await pages[1].reload();
  await pages[1].locator('#code-label').waitFor();
  await pages[1].locator('#beginner-guide').waitFor({ state: 'visible' });
  assert.deepEqual((await state(tokens[1])).config, config);
  await inspectGame(pages[1], config, `${index}-mobile-reloaded`);
  await inspectRules(pages[1], config, `${index}-mobile-reloaded`);
  await pages[1].locator('#beginner-guide').scrollIntoViewIfNeeded();
  await screenshot(pages[1], `${index}-mobile-guide`);
  if (
    config.board === 'tuscany' &&
    config.structures &&
    config.specialWorkers &&
    config.visitors !== 'ee'
  ) {
    const page = pages[1];
    await page.locator('[data-filter="structure"]').click();
    await page.locator('#game [data-rules-open]').click();
    await page.locator('#rules-help [data-topic="special-workers"]').click();
    await page.locator('#switch-table').click();
    await page.locator('#nickname').fill(pureRoom.name);
    await page.locator('#player-password').fill('configuration-test-123');
    await page.locator('#room-code').fill(pureRoom.code);
    await page.locator('#join-room').click();
    await page.waitForFunction(
      (code) => document.querySelector('#code-label')?.textContent === '房间 ' + code,
      pureRoom.code,
    );
    assert.equal(
      await page
        .locator(
          '#rules-help [data-topic="special-workers"], #rules-help [data-topic="structures"], #rules-help [data-topic="visitors"], #rules-help [data-topic="influence"]',
        )
        .count(),
      0,
    );
    assert.equal(
      await page.locator('#hand-filters [data-filter="all"]').getAttribute('aria-pressed'),
      'true',
    );
    await inspectGame(
      page,
      { board: 'ee', structures: false, specialWorkers: false, visitors: 'ee' },
      `${index}-cross-room`,
    );
    await inspectRules(
      page,
      { board: 'ee', structures: false, specialWorkers: false, visitors: 'ee' },
      `${index}-cross-room`,
    );
    current.crossRoom =
      'Actual authenticated join from all modules to previously created pure EE; stale topics, filters and invisible highlights cleared';
  }
  assert.equal(result.errors.length, 0, JSON.stringify(result.errors));
  current.status = 'passed';
  for (const context of contexts) await context.close();
  fs.writeFileSync(
    path.join(output, 'configuration-http-result.json'),
    JSON.stringify(result, null, 2),
  );
  console.log('PASS actual HTTP configuration', index, JSON.stringify(config));
}
(async () => {
  try {
    await start();
    browser = await chromium.launch({
      headless: true,
      ...(process.platform === 'win32' ? { channel: 'msedge' } : {}),
    });
    let index = 0;
    for (const visitors of (process.env.VITICULTURE_CONFIG_VISITORS || 'ee,ee_moor,rhine').split(
      ',',
    ))
      for (const board of ['ee', 'tuscany'])
        for (const structures of [false, true])
          for (const specialWorkers of [false, true])
            await run({ board, structures, specialWorkers, visitors }, index++);
    result.status = 'passed';
  } catch (error) {
    result.status = 'failed';
    result.error = error.stack;
    process.exitCode = 1;
    console.error(error);
  } finally {
    if (browser) await browser.close();
    if (child && child.exitCode == null)
      await new Promise((resolve) => {
        child.once('exit', resolve);
        child.kill();
      });
    fs.writeFileSync(
      path.join(output, 'configuration-http-result.json'),
      JSON.stringify(result, null, 2),
    );
    console.log('Evidence:', output);
  }
})();
