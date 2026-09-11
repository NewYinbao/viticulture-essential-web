// One reproducible run of complementary acceptance suites. Seeded scenarios are
// labelled separately from normal setup / natural games; this is not a claim
// that every possible visitor combination was exercised in a browser.
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const { spawn, execFileSync } = require('node:child_process');
const { sourceHashes } = require('./e2e/runtime.cjs');
const root = path.resolve(__dirname, '..');
const out = path.resolve(process.env.VITICULTURE_FULL_OUTPUT || path.join(root, 'artifacts/expansions/full-browser'));
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'viticulture-full-acceptance-'));
const sourceBinary = path.resolve(process.env.VITICULTURE_FULL_EXE || path.join(root, 'dist/Viticulture-Continuation.exe'));
const binary = path.join(temporary, process.platform === 'win32' ? 'server.exe' : 'server');
fs.mkdirSync(out, { recursive: true });
fs.copyFileSync(sourceBinary, binary);
const hash = (file) => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const report = {
  status: 'running', startedAt: new Date().toISOString(), sourceBinary,
  binarySha256: hash(binary), sourceHashes: sourceHashes(), suites: [],
  scope: {
    normalSetup: '24 configurations, two independent browsers, real API/UI, no saved-state fixtures',
    naturalGames: '24 normal random-deck games from setup to terminal scoring; strategy does not play visitors',
    seededGameplay: '36 structure triggers plus representative worker / Moor / Rhine nested and restore paths',
    authentication: 'password, unauthorized name reuse, session revocation and legacy enrollment',
    syntheticVisibility: 'all 24 combinations and cross-room transitions; supplements real configuration suite',
  },
  exhaustiveVisitorBrowserCoverage: false,
};
const reportFile = path.join(out, 'full-acceptance.json');
const save = () => fs.writeFileSync(reportFile, JSON.stringify(report, null, 2));
save();
function run(command, args, env, label) {
  const item = { label, command, args, status: 'running', startedAt: new Date().toISOString() };
  report.suites.push(item); save();
  const log = fs.createWriteStream(path.join(out, label + '.log'));
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: root, windowsHide: true, env: { ...process.env, ...env }, stdio: ['ignore', 'pipe', 'pipe'],
    });
    for (const stream of [child.stdout, child.stderr]) stream.on('data', (chunk) => {
      process.stdout.write(chunk); log.write(chunk);
    });
    child.once('error', (error) => { log.end(); reject(error); });
    child.once('exit', (code, signal) => {
      log.end(); item.exitCode = code; item.signal = signal; item.finishedAt = new Date().toISOString();
      item.status = code === 0 ? 'passed' : 'failed'; save();
      code === 0 ? resolve() : reject(Error(label + ' failed with exit ' + code));
    });
  });
}
const linuxPath = (value) => execFileSync('wsl.exe', ['--exec', 'wslpath', '-a', '-u', value], { encoding: 'utf8', windowsHide: true }).trim();
async function exportFixtures() {
  const env = {
    STRUCTURE_BROWSER_FIXTURES: path.join(out, 'structures-fixtures.json'),
    WORKER_BROWSER_FIXTURES: path.join(out, 'workers-fixtures.json'),
    MOOR_INTERACTION_FIXTURES: path.join(out, 'moor-fixtures.json'),
    RHINE_HTTP_FIXTURES: path.join(out, 'rhine-fixtures.json'),
    MESSENGER_BROWSER_FIXTURES: path.join(out, 'messenger-fixtures.json'),
    PLANNER_SPECIAL_BROWSER_FIXTURES: path.join(out, 'planner-special-fixtures.json'),
    PLANNER_FEASIBILITY_BROWSER_FIXTURES: path.join(out, 'planner-feasibility-fixtures.json'),
  };
  const args = ['test', './internal/game', '-run', 'Test(StructureBrowserFixtureExport|WorkerBrowserFixtureExport|MoorInteractionBrowserFixtureExport|RhineWriteHTTPFixtures|MessengerBrowserFixtureExport|PlannerSpecialBrowserFixtureExport|PlannerFarmerFeasibilityBrowserFixtureExport)$', '-count=1'];
  if (process.platform === 'win32' && !process.env.GO_BIN) {
    const variables = Object.entries(env).map(([key, value]) => key + '=' + linuxPath(value));
    await run('wsl.exe', ['--cd', linuxPath(root), '--exec', 'env', ...variables,
      process.env.VITICULTURE_WSL_GO || '/opt/viticulture-toolchain/go/bin/go', ...args], {}, 'fixture-export');
  } else await run(process.env.GO_BIN || 'go', args, env, 'fixture-export');
  return env;
}
async function suite(label, script, env) {
  await run(process.execPath, [script], env, label);
}
(async () => {
  try {
    const fixtures = await exportFixtures();
    await suite('configuration-24', 'tests/e2e/configuration-http-browser-test.cjs', {
      VITICULTURE_CONFIG_EXE: binary, VITICULTURE_CONFIG_OUTPUT: path.join(out, 'configuration'),
      VITICULTURE_CONFIG_VISITORS: 'ee,ee_moor,rhine',
    });
    await suite('natural-24', 'tests/e2e/expansions-natural-test.cjs', {
      VITICULTURE_NATURAL_EXE: binary, VITICULTURE_NATURAL_OUTPUT: path.join(out, 'natural'),
      VITICULTURE_NATURAL_VISITORS: 'ee,ee_moor,rhine',
    });
    for (const [label, file, prefix, fixture] of [
      ['structures-36', 'structures-browser-test.cjs', 'VITICULTURE_STRUCTURES', fixtures.STRUCTURE_BROWSER_FIXTURES],
      ['workers', 'worker-interactions-browser-test.cjs', 'VITICULTURE_WORKER', fixtures.WORKER_BROWSER_FIXTURES],
      ['moor', 'moor-interactions-browser-test.cjs', 'VITICULTURE_MOOR', fixtures.MOOR_INTERACTION_FIXTURES],
      ['rhine', 'rhine-browser-test.cjs', 'RHINE_HTTP', fixtures.RHINE_HTTP_FIXTURES],
    ]) await suite(label, 'tests/e2e/' + file, {
      [prefix + '_EXE']: binary, [prefix + '_FIXTURES']: fixture, [prefix + '_OUTPUT']: path.join(out, label),
    });
    await suite('messenger', 'tests/e2e/messenger-browser-test.cjs', {
      MESSENGER_HTTP_EXE: binary, MESSENGER_BROWSER_FIXTURES: fixtures.MESSENGER_BROWSER_FIXTURES,
      MESSENGER_BROWSER_DATA: path.join(temporary, 'messenger-data'),
      MESSENGER_BROWSER_OUTPUT: path.join(out, 'messenger'),
    });
    await suite('password', 'tests/e2e/password-browser-test.cjs', {
      VITICULTURE_EXE: binary, VITICULTURE_PASSWORD_OUTPUT: path.join(out, 'password'),
    });
    await suite('planner-special', 'tests/e2e/planner-special-browser-test.cjs', {
      PLANNER_SPECIAL_HTTP_EXE: binary,
      PLANNER_SPECIAL_BROWSER_FIXTURES: fixtures.PLANNER_SPECIAL_BROWSER_FIXTURES,
      PLANNER_SPECIAL_BROWSER_DATA: path.join(temporary, 'planner-special-data'),
      PLANNER_SPECIAL_BROWSER_OUTPUT: path.join(out, 'planner-special'),
    });
    await suite('planner-farmer-feasibility', 'tests/e2e/planner-farmer-feasibility-browser-test.cjs', {
      PLANNER_FEASIBILITY_HTTP_EXE: binary,
      PLANNER_FEASIBILITY_BROWSER_FIXTURES: fixtures.PLANNER_FEASIBILITY_BROWSER_FIXTURES,
      PLANNER_FEASIBILITY_BROWSER_DATA: path.join(temporary, 'planner-feasibility-data'),
      PLANNER_FEASIBILITY_BROWSER_OUTPUT: path.join(out, 'planner-farmer-feasibility'),
    });
    await suite('visibility-24', 'tests/e2e/module-visibility-test.cjs', {
      VITICULTURE_VISIBILITY_OUTPUT: path.join(out, 'visibility'),
    });
    if (JSON.stringify(report.sourceHashes) !== JSON.stringify(sourceHashes()))
      throw Error('Source changed during the acceptance run; rerun against one stable snapshot');
    report.status = 'passed';
  } catch (error) {
    report.status = 'failed'; report.error = error.stack; process.exitCode = 1;
  } finally {
    report.finishedAt = new Date().toISOString(); save();
    console.log(JSON.stringify({ status: report.status, report: reportFile }));
  }
})();
