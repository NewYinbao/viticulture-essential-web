// Isolated browser fixtures: this checks visibility, not a legal/natural game.
const fs = require("node:fs"),
  path = require("node:path"),
  os = require("node:os"),
  http = require("node:http"),
  crypto = require("node:crypto"),
  assert = require("node:assert/strict");
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");
const root = path.resolve(__dirname, "../.."),
  web = path.join(root, "web/static");
const out =
  process.env.VITICULTURE_VISIBILITY_OUTPUT ||
  fs.mkdtempSync(
    path.join(os.tmpdir(), "viticulture-module-visibility-evidence-"),
  );
fs.mkdirSync(out, { recursive: true });
const cases = [];
for (const board of ["ee", "tuscany"])
  for (const structures of [false, true])
    for (const specialWorkers of [false, true])
      for (const visitors of ["ee", "ee_moor", "rhine"])
        cases.push({ board, structures, specialWorkers, visitors });
const fixture = (config, index) => {
  const p = {
    id: "p",
    name: "门控测试庄主",
    coins: 30,
    vp: 5,
    income: 2,
    workers: 2,
    largeWorker: true,
    totalWorkers: 3,
    handCount: 2,
    handCounts: {
      vine: 1,
      order: 1,
      summer: 0,
      winter: 0,
      structure: config.structures ? 1 : 0,
    },
    fields: [
      {
        index: 0,
        capacity: 5,
        sold: false,
        harvested: false,
        vines: [],
        fruitDealer: true,
      },
      { index: 1, capacity: 6, sold: false, harvested: false, vines: [] },
    ],
    buildings: ["trellis"],
    structureSlots: config.structures ? ["cask", ""] : [],
    specialWorkers: config.specialWorkers ? ["chef"] : [],
    specialWorkerUsed: { chef: false },
    specialWorkerReady: { chef: 1 },
    grapes: [{ id: "g", color: "red", value: 2 }],
    wines: [{ id: "w", type: "red", value: 2 }],
    wake: 2,
    season: "summer",
    passed: false,
    papaResolved: true,
    moorContracts: [
      {
        id: "moor-summer-06",
        name: "大赞助人",
        description: "终局收入不足4时失4分",
      },
    ],
  };
  const sp = (id, name, season = "summer") => ({
    id,
    name,
    season,
    description: "隔离界面夹具",
    capacity: 2,
    occupied: [],
    bonus: "",
    bonusSlots: {},
  });
  const spaces = [
    sp("plant", "种植"),
    sp("summer_visitor", "夏季访客"),
    sp("build", "建造"),
    sp("gain_coin", "获取金币", "any"),
  ];
  if (config.board === "tuscany")
    spaces.push(sp("influence", "放置影响力"), sp("trade", "交易"));
  if (config.structures)
    spaces.push(
      sp("structure_action:p:cask", "木桶", "any"),
      sp("structure_destroy:p", "拆除结构", "any"),
    );
  const hand = [
    {
      id: "vine-test",
      name: "测试葡萄藤",
      type: "vine",
      red: 1,
      white: 0,
      implemented: true,
      description: "需要棚架",
    },
    {
      id: "order-test",
      name: "测试订单",
      type: "order",
      requirements: [{ type: "red", value: 1 }],
      points: 2,
      income: 1,
      description: "订单",
    },
  ];
  if (config.structures)
    hand.push({
      id: "structure-wine_cave",
      structureId: "wine_cave",
      name: "酒窖洞",
      type: "structure",
      structureCost: 2,
      description: "结构牌夹具",
      implemented: true,
    });
  return {
    code: "MODULE" + index,
    config,
    phase: "summer",
    year: 1,
    revision: 1,
    youId: "p",
    hostId: "p",
    turnId: "p",
    players: [p, { ...structuredClone(p), id: "q", name: "邻座" }],
    hand,
    spaces,
    wakeSlots: [
      { slot: 2, playerId: "p", bonus: "当前起床奖励" },
      { slot: 3, playerId: "q", bonus: "当前起床奖励" },
    ],
    pendingChoice: null,
    pendingCount: 0,
    legal: {
      canPlace: true,
      canPass: true,
      canStart: false,
      canWake: false,
      canConfigure: false,
    },
    passwordSet: true,
    log: ["隔离24组合UI夹具；非自然对局"],
    rulesNotes: ["旧全局说明：Tuscany、Moor、Rhine 与特殊工人、结构牌。"],
    deckCounts: {
      vine: { deck: 30, discard: 0 },
      order: { deck: 30, discard: 0 },
      summer: { deck: 30, discard: 0 },
      winter: { deck: 30, discard: 0 },
      structure: { deck: 30, discard: 0 },
    },
    specialWorkerPool: ["chef", "professor"],
    specialWorkerCatalog: [
      { id: "chef", name: "厨师", description: "退回对手工人" },
      { id: "professor", name: "教授", description: "取回本季工人" },
    ],
    expansionAvailability: {
      tuscany: true,
      structures: true,
      specialWorkers: true,
      ee_moor: true,
      rhine: true,
    },
    ruleSupport: { playable: true, reason: "" },
    actionReasons: {},
    cardReasons: {},
    optionReasons: {},
    influenceRegions: [
      { id: "lucca", name: "卢卡", points: 1, reward: "coin" },
    ],
    winnerIds: [],
  };
};
const views = new Map(
  cases.map((config, i) => ["MODULE" + i, fixture(config, i)]),
);
const report = {
  fixture: true,
  naturalGame: false,
  sourceHashes: {},
  cases: [],
  failures: [],
  errors: [],
  crossRoom: [],
  output: out,
};
for (const name of fs
  .readdirSync(path.join(web, "js"))
  .filter((f) => f.endsWith(".js")))
  report.sourceHashes["web/static/js/" + name] = crypto
    .createHash("sha256")
    .update(fs.readFileSync(path.join(web, "js", name)))
    .digest("hex");
let server, browser, page;
const save = () =>
  fs.writeFileSync(
    path.join(out, "module-visibility-result.json"),
    JSON.stringify(report, null, 2),
  );
const fail = (message, detail) => report.failures.push({ message, ...detail });
const check = (actual, expected, message, detail) => {
  if (actual !== expected) fail(message, { actual, expected, ...detail });
};
(async () => {
  try {
    server = http.createServer(async (req, res) => {
      const token = (req.headers.authorization || "").replace(/^Bearer /, "");
      res.setHeader("Cache-Control", "no-store");
      if (req.url === "/api/state") {
        res.setHeader("Content-Type", "application/json");
        return res.end(
          JSON.stringify(views.get(token) || views.get("MODULE0")),
        );
      }
      if (req.url === "/api/transport") {
        res.setHeader("Content-Type", "application/json");
        return res.end('{"poll":true}');
      }
      if (req.url === "/api/join") {
        let body = "";
        for await (const chunk of req) body += chunk;
        const code = JSON.parse(body).code;
        res.setHeader("Content-Type", "application/json");
        return res.end(JSON.stringify({ token: code, code }));
      }
      if (req.url.startsWith("/api/")) {
        res.setHeader("Content-Type", "application/json");
        return res.end("{}");
      }
      const pathname = decodeURIComponent(
        new URL(req.url, "http://localhost").pathname,
      );
      const file = path.resolve(
        web,
        "." + (pathname === "/" ? "/index.html" : pathname),
      );
      if (!file.startsWith(web + path.sep)) {
        res.statusCode = 403;
        return res.end();
      }
      const ext = path.extname(file);
      res.setHeader(
        "Content-Type",
        {
          ".html": "text/html",
          ".js": "text/javascript",
          ".css": "text/css",
          ".svg": "image/svg+xml",
          ".png": "image/png",
        }[ext] || "application/octet-stream",
      );
      try {
        res.end(fs.readFileSync(file));
      } catch {
        res.statusCode = 404;
        res.end();
      }
    });
    await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
    const base = "http://127.0.0.1:" + server.address().port;
    browser = await chromium.launch({ headless: true, ...(process.platform === 'win32' ? {channel:'msedge'} : {}), args: ["--no-sandbox"] });
    page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    page.on("pageerror", (e) => report.errors.push(e.message));
    page.on("dialog", (dialog) => dialog.accept());
    async function load(code) {
      await page.goto(base);
      await page.evaluate((code) => {
        sessionStorage.setItem("vineyard-ee-token", code);
        localStorage.setItem("vineyard-ee-guide", "on");
        localStorage.setItem("vineyard-ee-tour-step", "0");
        localStorage.removeItem("vineyard-ee-tour:" + code + ":p");
      }, code);
      await page.reload();
      await page.waitForFunction(
        (code) =>
          document.querySelector("#code-label")?.textContent === "房间 " + code,
        code,
      );
    }
    async function inspect(config, code, transition = false) {
      const record = { code, config, transition };
      check((await page.locator('#rules-text').textContent()).includes('旧全局说明'), false, 'legacy global module notes never shown', record);
      check(
        await page.locator("#tuscany-board").count(),
        config.board === "tuscany" ? 1 : 0,
        "Tuscany panel gating",
        record,
      );
      check(
        await page.locator('#hand-filters [data-filter="structure"]').count(),
        config.structures ? 1 : 0,
        "structure hand-filter gating",
        record,
      );
      check(
        /结构|structure/.test(
          await page.locator(".ee-deck-counts").textContent(),
        ),
        config.structures,
        "structure deck-count gating",
        record,
      );
      check(
        (await page.locator(".special-worker-reserve").count()) > 0,
        config.specialWorkers,
        "special-worker reserve gating",
        record,
      );
      check(
        (await page.locator("#ee-choice").textContent()).includes(
          "本局特殊工人",
        ),
        config.specialWorkers,
        "special-worker public pool gating",
        record,
      );
      check(
        (await page.locator('[data-space^="structure_action:"]').count()) > 0,
        config.structures,
        "private structure action gating",
        record,
      );
      check(
        (await page.locator("#players").innerText()).includes("终局收入≥4"),
        config.visitors === "ee_moor",
        "Moor contract gating",
        record,
      );
      check(
        (await page.locator("#players").innerText()).includes("水果商"),
        config.visitors === "ee_moor",
        "Moor field badge gating",
        record,
      );
      await page.locator("#game [data-rules-open]").click();
      await page.locator("#rules-help").waitFor({ state: "visible" });
      const nav = await page
        .locator("#rules-help [data-topic]")
        .evaluateAll((es) =>
          es.map((e) => ({ id: e.dataset.topic, title: e.textContent })),
        );
      record.ruleTopics = nav;
      for (const [topic, on] of [
        ["influence", config.board === "tuscany"],
        ["trade", config.board === "tuscany"],
        ["actions", config.board === "tuscany"],
        ["structures", config.structures],
        ["special-workers", config.specialWorkers],
        ["visitors", config.visitors !== "ee"],
      ])
        check(
          nav.some((t) => t.id === topic),
          on,
          "rule-topic gating " + topic,
          record,
        );
      const vt = nav.find((x) => x.id === "visitors");
      if (vt)
        check(
          vt.title.includes(config.visitors === "rhine" ? "Rhine" : "Moor"),
          true,
          "visitor rule set name",
          record,
        );
      const links = await page
        .locator("#rules-help .help-foot a")
        .evaluateAll((es) => es.map((e) => e.href));
      record.sources = links;
      check(
        links.some((s) => s.includes("/moor-visitors-expansion")),
        config.visitors === "ee_moor",
        "Moor source gating",
        record,
      );
      check(
        links.some((s) => s.includes("rhine")),
        config.visitors === "rhine",
        "Rhine source gating",
        record,
      );
      await page.keyboard.press("Escape");
      // The same compact guide must change its season and hand lesson with config.
      if (await page.locator("[data-guide-restart]").count())
        await page.locator("[data-guide-restart]").click();
      while (
        Number(
          await page.locator("#beginner-guide").getAttribute("data-step"),
        ) > 0
      )
        await page.locator("[data-guide-previous]").click();
      record.guideFirst = await page
        .locator("#beginner-guide .guide-copy")
        .innerText();
      check(
        record.guideFirst.includes("四季都有工人行动"),
        config.board === "tuscany",
        "beginner season lesson gating",
        record,
      );
      await page.locator("[data-guide-next]").click();
      await page.locator("[data-guide-next]").click();
      record.guideHand = await page
        .locator("#beginner-guide .guide-copy")
        .innerText();
      check(
        record.guideHand.includes("橙色建筑牌"),
        config.structures,
        "beginner structure hand lesson gating",
        record,
      );
      await page.locator('[data-space="build"]').click();
      await page.locator("#action-panel").waitFor({ state: "visible" });
      check(
        (await page
          .locator('#action-panel [data-group="workerType"]')
          .count()) > 0,
        config.specialWorkers,
        "placement special-worker controls",
        record,
      );
      check(
        (await page
          .locator(
            '#action-panel [data-group="building"][data-value="wine_cave"]',
          )
          .count()) > 0,
        config.structures,
        "build structure controls",
        record,
      );
      await page.keyboard.press("Escape");
      // Exercise visitor worker-field gating even with the common schema present.
      await page.evaluate(async (code) => {
        const v = await (
          await fetch("/api/state", {
            headers: { Authorization: "Bearer " + code },
          })
        ).json();
        v.pendingChoice = {
          id: "visibility-visitor",
          kind: "visitor",
          playerId: "p",
          options: ["plan"],
          visitor: { cardId: "summer-29", stage: "effect", actorId: "p" },
          optionFields: {
            plan: [
              { name: "workerType", type: "specialWorker", min: 1, max: 1 },
            ],
          },
        };
        const { renderVisitor } = await import("/js/visitor-ui.js");
        const box = document.createElement("div");
        box.id = "visibility-visitor";
        document.body.append(box);
        renderVisitor(box, v, async () => true, true);
      }, code);
      check(
        (await page.locator("#visibility-visitor").innerText()).includes(
          "特殊工人（可选）",
        ),
        config.specialWorkers,
        "visitor special worker field gating",
        record,
      );
      await page.locator("#visibility-visitor").evaluate((e) => e.remove());
      if (!transition) report.cases.push(record);
      else report.crossRoom.push(record);
      save();
      return record;
    }
    for (let i = 0; i < cases.length; i++) {
      await load("MODULE" + i);
      await inspect(cases[i], "MODULE" + i);
      console.log("CHECK", i, JSON.stringify(cases[i]));
    }
    // Enter another room through the real visible join UI, without reloading the app.
    for (const visitors of ["ee_moor", "rhine"]) {
      const index = cases.findIndex(
        (c) =>
          c.board === "tuscany" &&
          c.structures &&
          c.specialWorkers &&
          c.visitors === visitors,
      );
      await load("MODULE" + index);
      await page.locator('[data-filter="structure"]').click();
      await page.locator("#game [data-rules-open]").click();
      await page.locator('[data-topic="special-workers"]').click();
      await page.locator("#switch-table").click();
      await page.locator("#nickname").fill("门控测试庄主");
      await page.locator("#player-password").fill("fixture-password");
      await page.locator("#room-code").fill("MODULE0");
      await page.locator("#join-room").click();
      await page.waitForFunction(
        () =>
          document.querySelector("#code-label")?.textContent === "房间 MODULE0",
      );
      check(
        await page
          .locator(
            '#rules-help [data-topic="special-workers"], #rules-help [data-topic="structures"], #rules-help [data-topic="visitors"], #rules-help [data-topic="influence"]',
          )
          .count(),
        0,
        "room-switch clears stale module rules",
        { from: visitors },
      );
      await page.keyboard.press("Escape");
      check(
        await page
          .locator('#hand-filters [data-filter="all"]')
          .getAttribute("aria-pressed"),
        "true",
        "room-switch clears structure hand filter",
        { from: visitors },
      );
      await inspect(cases[0], "MODULE0", true);
    }
    check(report.errors.length, 0, "browser runtime errors", {});
    report.status = report.failures.length ? "failed" : "passed";
    save();
    console.log(
      JSON.stringify({
        status: report.status,
        cases: report.cases.length,
        crossRoom: report.crossRoom.length,
        failures: report.failures,
        errors: report.errors,
        output: out,
      }),
    );
    if (report.failures.length) process.exitCode = 1;
  } catch (e) {
    report.status = "error";
    report.error = e.stack;
    save();
    console.error(e);
    process.exitCode = 1;
    if (page)
      await page.screenshot({
        path: path.join(out, "failure.png"),
        fullPage: true,
      });
  } finally {
    if (browser) await browser.close();
    if (server) await new Promise((resolve) => server.close(resolve));
  }
})();
