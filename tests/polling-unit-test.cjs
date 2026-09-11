const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
function harness(fetch) {
  const timers = new Map(); let id = 0;
  const context = { fetch, AbortController, setTimeout(fn, ms) { timers.set(++id, {fn, ms}); return id; }, clearTimeout(id) { timers.delete(id); } };
  vm.createContext(context);
  vm.runInContext(fs.readFileSync(path.join(__dirname, '../web/static/js/polling.js'), 'utf8').replace('export function', 'function'), context);
  return { poll: context.pollState, timers, fire(ms) { const item = [...timers].find(([, x]) => x.ms === ms); assert.ok(item, 'timer exists'); timers.delete(item[0]); item[1].fn(); } };
}
const flush = async () => { for (let i = 0; i < 10; i++) await Promise.resolve(); };
test('single flight, credentials, delivery, close prevents late state and retries', async () => {
  let resolve, signal, count = 0; const states = [], errors = [];
  const h = harness((url, options) => { count++; assert.equal(url, '/api/state'); assert.equal(options.headers.Authorization, 'Bearer private'); assert.equal(options.cache, 'no-store'); signal = options.signal; return new Promise(r => resolve = r); });
  const p = h.poll('private', s => states.push(s), e => errors.push(e), 25);
  assert.equal(count, 1); assert.equal([...h.timers.values()].filter(x => x.ms === 25).length, 0);
  resolve({ok: true, json: async () => ({turn: 1})}); await flush(); assert.equal(states[0].turn, 1);
  h.fire(25); assert.equal(count, 2); p.close(); p.close(); assert.equal(signal.aborted, true);
  resolve({ok: true, json: async () => ({turn: 2})}); await flush();
  assert.equal(states.length, 1); assert.equal(errors.length, 0); assert.equal(h.timers.size, 0);
});
test('HTTP and network errors retry, then recover', async () => {
  let count = 0; const states = [], errors = [];
  const h = harness(async () => { count++; if (count === 1) return {ok: false, status: 503}; if (count === 2) throw Error('offline'); return {ok: true, json: async () => ({ok: true})}; });
  const p = h.poll('token', s => states.push(s), e => errors.push(e), 25);
  await flush(); assert.match(errors[0].message, /503/); h.fire(25); await flush(); assert.match(errors[1].message, /offline/);
  h.fire(25); await flush(); assert.equal(states.length, 1); p.close(); assert.equal(h.timers.size, 0);
});
test('ten second timeout aborts pending fetch and retries; close suppresses abort errors', async () => {
  const errors = []; let signal;
  const h = harness((url, options) => { signal = options.signal; return new Promise((resolve, reject) => signal.addEventListener('abort', () => reject(Error('aborted')), {once: true})); });
  const p = h.poll('token', () => assert.fail('unexpected state'), e => errors.push(e), 25);
  h.fire(10000); assert.equal(signal.aborted, true); await flush(); assert.equal(errors.length, 1);
  h.fire(25); p.close(); await flush(); assert.equal(errors.length, 1); assert.equal(h.timers.size, 0);
});
