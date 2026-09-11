const { chromium } = require("playwright");
const { spawn } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { authorizeFixtureStore } = require("../helpers/fixture-auth.cjs");

const repo =
  process.env.VITICULTURE_SOURCE_ROOT || path.resolve(__dirname, "../..");
const fixture = JSON.parse(
  fs.readFileSync(process.env.MESSENGER_BROWSER_FIXTURES, "utf8"),
);
const bundle = authorizeFixtureStore(fixture);
const data = process.env.MESSENGER_BROWSER_DATA;
const output = process.env.MESSENGER_BROWSER_OUTPUT;
fs.mkdirSync(data, { recursive: true });
fs.mkdirSync(output, { recursive: true });
fs.writeFileSync(path.join(data, "ee-state-v1.json"), JSON.stringify(bundle));
const report = {
  browser: "Microsoft Edge",
  headless: true,
  checks: [],
  pageErrors: [],
};
let child;
let browser;
let base;

async function start() {
  child = spawn(
    process.env.MESSENGER_HTTP_EXE,
    ["-addr", "127.0.0.1:0", "-data", data],
    {
      cwd: repo,
      windowsHide: true,
    },
  );
  base = await new Promise((resolve, reject) => {
    let log = "";
    const timer = setTimeout(
      () => reject(new Error("server startup timeout: " + log)),
      15000,
    );
    child.once("error", reject);
    const consume = (chunk) => {
      log += chunk;
      const match = log.match(/http:\/\/127\.0\.0\.1:(\d+)/);
      if (match) {
        clearTimeout(timer);
        resolve("http://127.0.0.1:" + match[1]);
      }
    };
    child.stdout.on("data", consume);
    child.stderr.on("data", consume);
  });
}

async function stop() {
  if (child && child.exitCode === null)
    await new Promise((resolve) => {
      child.once("exit", resolve);
      child.kill();
    });
  child = null;
}

async function load(page, token) {
  await page.goto(base, { waitUntil: "domcontentloaded" });
  await page.evaluate(
    (value) => localStorage.setItem("vineyard-ee-token", value),
    token,
  );
  await page.reload({ waitUntil: "domcontentloaded" });
  await page.locator("#code-label").waitFor({ state: "visible" });
}

async function state(page, token) {
  const response = await page.request.get(base + "/api/state", {
    headers: { Authorization: "Bearer " + token },
  });
  assert.equal(response.status(), 200, await response.text());
  return response.json();
}

async function clickAction(page, selector) {
  const pending = page.waitForResponse(
    (response) =>
      response.url().endsWith("/api/action") &&
      response.request().method() === "POST",
  );
  await page.locator(selector).click();
  const response = await pending;
  assert.equal(response.status(), 200, await response.text());
  return response.json();
}

(async () => {
  try {
    await start();
    browser = await chromium.launch({ channel: "msedge", headless: true });
    const pages = await Promise.all(
      bundle.case.tokens.map(async () => {
        const page = await browser.newPage({
          viewport: { width: 1440, height: 1000 },
        });
        page.on("pageerror", (error) => report.pageErrors.push(error.message));
        return page;
      }),
    );
    await load(pages[0], bundle.case.tokens[0]);
    await pages[0].locator('[data-space="make_wine"]').click();
    await pages[0]
      .locator('[data-group="workerType"][data-value="messenger"]')
      .click();
    assert.equal(await pages[0].locator('[data-group="grapes"]').count(), 0);
    assert.match(
      await pages[0].locator("#action-panel").innerText(),
      /现在只锁定信使的位置/,
    );
    await pages[0].locator('[data-group="slot"][data-value="1"]').click();
    let view = await clickAction(pages[0], "#confirm-action");
    assert.equal(view.messengerPlans.length, 1);
    assert.equal(view.messengerPlans[0].space, "make_wine");
    report.checks.push(
      "summer UI reserves only the Messenger position without wine inputs",
    );

    await stop();
    await start();
    await load(pages[0], bundle.case.tokens[0]);
    await pages[0].waitForFunction(
      () => !document.querySelector("#pass")?.disabled,
    );
    pages[0].once("dialog", (dialog) => dialog.accept());
    view = await clickAction(pages[0], "#pass");
    assert.equal(view.phase, "fall");
    report.checks.push("reservation survives a real server process restart");

    for (let i = 0; i < pages.length; i++) {
      await load(pages[i], bundle.case.tokens[i]);
      await pages[i].waitForFunction(
        () => !document.querySelector('[data-choice="winter"]')?.disabled,
      );
      await clickAction(pages[i], '[data-choice="winter"]');
    }
    await load(pages[0], bundle.case.tokens[0]);
    view = await state(pages[0], bundle.case.tokens[0]);
    assert.equal(view.phase, "winter");
    assert.equal(view.pendingChoice.kind, "messenger");
    assert.equal(view.pendingChoice.actionSpace, "make_wine");
    await pages[0].evaluate(() => localStorage.setItem('vineyard-ee-guide', 'on'));
    await pages[0].reload({ waitUntil: 'domcontentloaded' });
    await pages[0].locator('[data-choice-resources="messenger"]').waitFor();
    assert.match(await pages[0].locator('#beginner-guide').innerText(), /执行信使预约/);
    assert.equal(await pages[0].locator('[data-guide-locate]').isEnabled(), true);
    await pages[0].locator('[data-guide-locate]').click();
    assert.equal(await pages[0].evaluate(() => document.activeElement?.dataset.choiceResources), 'messenger');
    report.checks.push('restored Messenger guide explains the current choice and focuses its real control');
    await pages[0].locator('[data-choice-resources="messenger"]').click();
    assert.equal(await pages[0].locator('[data-group="worker"]').count(), 0);
    await pages[0].locator('[data-group="grapes"][data-value="0"]').click();
    view = await clickAction(pages[0], "#confirm-action");
    const me = view.players.find((player) => player.id === view.youId);
    assert.equal(me.wines.length, 1);
    assert.equal((view.messengerPlans || []).length, 0);
    assert.notEqual(view.turnId, view.youId);
    report.checks.push(
      "winter first turn opens continuation controls and executes using live hand state",
    );
    assert.deepEqual(report.pageErrors, []);
    report.status = "passed";
  } catch (error) {
    report.status = "failed";
    report.failure = error.stack;
    process.exitCode = 1;
  } finally {
    await browser?.close();
    await stop();
    fs.writeFileSync(
      path.join(output, "messenger-browser-result.json"),
      JSON.stringify(report, null, 2),
    );
    console.log(JSON.stringify(report, null, 2));
  }
})();
