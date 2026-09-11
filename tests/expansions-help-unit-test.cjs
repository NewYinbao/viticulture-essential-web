const test = require('node:test');
const assert = require('node:assert/strict');
const { pathToFileURL } = require('node:url');
const path = require('node:path');
const moduleURL = pathToFileURL(path.join(__dirname, '../web/static/js/rules-content.js'));
const rules = import(moduleURL.href);

test('24 released module combinations select isolated topics without obsolete warnings', async () => {
  const { rulesForView, ruleTopics } = await rules;
  const original = JSON.stringify(ruleTopics);
  let count = 0;
  for (const board of ['ee', 'tuscany'])
    for (const visitors of ['ee', 'ee_moor', 'rhine'])
      for (const structures of [false, true])
        for (const specialWorkers of [false, true]) {
          const v = {
            config: { board, visitors, structures, specialWorkers },
            expansionAvailability: {
              ee: true,
              tuscany: true,
              structures: true,
              specialWorkers: true,
              ee_moor: true,
              rhine: true,
            },
          };
          const result = rulesForView(v),
            ids = result.topics.map((t) => t.id);
          assert.equal(ids.includes('influence'), board === 'tuscany');
          assert.equal(ids.includes('trade'), board === 'tuscany');
          assert.equal(ids.includes('structures'), structures);
          assert.equal(ids.includes('special-workers'), specialWorkers);
          assert.equal(ids.includes('visitors'), visitors !== 'ee');
          const overview = JSON.stringify(result.topics.find((t) => t.id === 'overview'));
          assert.match(overview, board === 'tuscany' ? /25 分/ : /20 分/);
          if (board === 'tuscany') assert.doesNotMatch(overview, /20 分/);
          assert.equal(!!result.warning, false);
          assert.doesNotMatch(JSON.stringify(result.topics), /仍禁止开局|逐牌说明尚未可用/);
          if (visitors === 'rhine')
            assert.match(
              JSON.stringify(result.topics.find((t) => t.id === 'visitors')),
              /替换.*不能混入/,
            );
          count++;
        }
  assert.equal(count, 24);
  assert.equal(JSON.stringify(ruleTopics), original, 'switching rooms changed the EE reference');
  assert.deepEqual(rulesForView(undefined).topics, ruleTopics);
});

test('live availability warnings remain authoritative after release', async () => {
  const { rulesForView } = await rules;
  const result = rulesForView({
    config: { board: 'ee', visitors: 'rhine' },
    expansionAvailability: { rhine: false },
    ruleSupport: { playable: false, reason: '访客 example 尚未完成实现' },
  });
  assert.match(result.warning, /example/);
});

test('Label Factory is the twelfth private action, not a persistent enhancement', async () => {
  const { rulesForView } = await rules;
  const cards = rulesForView({ config: { structures: true } }).topics.find(
    (t) => t.id === 'structures',
  ).cards;
  assert.match(cards.find((c) => c[0] === '私有行动建筑')[1], /酒标工厂/);
  assert.doesNotMatch(cards.find((c) => c[0] === '持续、年末与拆除')[1], /酒标工厂/);
});

test('Tuscany rewards and actions use actual public board on either side', async () => {
  const { rulesForView } = await rules;
  const view = {
    config: { board: 'tuscany' },
    wakeSlots: [{ slot: 5, bonus: '冬：影响力' }],
    influenceRegions: [{ name: '锡耶纳', reward: 'coin', points: 2 }],
    spaces: [{ name: '本局交易', description: '真实主板行动', season: 'summer' }],
  };
  let text = JSON.stringify(rulesForView(view));
  assert.match(text, /第5行奖励.*冬：影响力/);
  assert.match(text, /锡耶纳.*1金币.*2分/);
  assert.match(text, /本局交易.*真实主板行动/);
  text = JSON.stringify(rulesForView({ ...view, config: { ...view.config, structures: true } }));
  assert.match(text, /side 2/);
  assert.match(text, /锡耶纳.*1金币.*2分/);
});

test('private pending branch help reuses concrete costs and hides opponent details', async () => {
  const { pendingRule } = await rules;
  const view = {
    code: 'HELP',
    youId: 'guest',
    players: [{ id: 'guest', coins: 5, fields: [] }],
    optionReasons: { cards: '需要2张不同手牌' },
    pendingChoice: {
      id: 'choice',
      kind: 'visitor',
      playerId: 'guest',
      options: ['vp', 'cards', 'coins'],
      visitor: { cardId: 'winter-11', stage: 'reply', actorId: 'host' },
    },
  };
  const text = JSON.stringify(
    pendingRule(view, { choiceId: 'choice', option: 'cards', reason: '还需选1张' }),
  );
  assert.match(text, /多人响应/);
  assert.match(text, /已选.*给出牌者 2 张手牌/);
  assert.match(text, /支付 3 金币/);
  assert.match(text, /还需选1张/);
  const publicText = JSON.stringify(pendingRule({ ...view, youId: 'host' }));
  assert.doesNotMatch(publicText, /winter-11|3 金币|2 张手牌|还需选1张/);
  assert.match(publicText, /只有当前响应者/);
});

test('pending quantities, personal upkeep, and unknown expansion cards are explicit', async () => {
  const { pendingRule } = await rules;
  const v = { youId: 'me', players: [{ id: 'me', income: 4 }] };
  const choice = (kind, extra = {}) => ({
    ...v,
    pendingChoice: { kind, playerId: 'me', ...extra },
  });
  assert.match(JSON.stringify(pendingRule(choice('discard', { count: 3 }))), /恰好3张/);
  assert.match(JSON.stringify(pendingRule(choice('tuscany_upkeep'))), /4金币/);
  assert.match(JSON.stringify(pendingRule(choice('tuscany_trade'))), /第一次交易已结算/);
  const unknown = JSON.stringify(
    pendingRule(choice('visitor', { options: ['coins'], visitor: { cardId: 'rhine-unverified' } })),
  );
  assert.match(unknown, /尚未核实/);
  assert.doesNotMatch(unknown, /无需额外|已就绪/);
});

test('released worker and structure continuations have concise private contextual rules', async () => {
  const { pendingRule } = await rules;
  const view = { youId: 'me', players: [{ id: 'me' }] };
  for (const kind of [
    'messenger', 'special_train', 'special_farmer', 'special_professore',
    'special_innkeeper', 'special_mafioso', 'special_politico', 'special_oracle',
    'special_merchant', 'structure_draft', 'structure_mercado',
  ]) {
    const pendingChoice = { kind, playerId: 'me', options: [] };
    const rule = pendingRule({ ...view, pendingChoice });
    assert.doesNotMatch(JSON.stringify(rule), /尚未支持|尚无经过核实/);
    assert(rule.cards[0][1].length < 100, kind + ' should stay concise');
    const hidden = pendingRule({ ...view, youId: 'opponent', pendingChoice });
    assert.match(hidden.intro, /等待/);
    assert.doesNotMatch(JSON.stringify(hidden), /先知|信使|1金币/);
  }
  assert.match(JSON.stringify(pendingRule({ ...view, pendingChoice: { kind: 'messenger', playerId: 'me' } })), /当前手牌/);
  assert.match(JSON.stringify(pendingRule({ ...view, pendingChoice: { kind: 'special_farmer', playerId: 'me' } })), /先锁定.*成功执行/);
});
