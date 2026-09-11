const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const source = fs.readFileSync(path.join(root, 'internal/game/special_workers.go'), 'utf8');
const ids = [...source.matchAll(/\{ID:\s*"([a-z]+)"\s*,\s*Name:\s*"([^"]+)"\s*,\s*Description:\s*"([^"]+)"\s*,\s*RuleSource:\s*"([^"]+)"\}/g)];
const expected = ['farmer', 'mafioso', 'chef', 'innkeeper', 'professore', 'soldato', 'politico', 'oracle', 'merchant', 'traveler', 'messenger'];

assert.equal(ids.length, 11, 'Tuscany special-worker catalogue must contain exactly 11 entries');
assert.deepEqual(ids.map(([, id]) => id), expected, 'special-worker IDs/order changed unexpectedly');
for (const [, id, name, description, ruleSource] of ids) {
  assert.ok(name.length > 0, `${id} needs a display name`);
  assert.match(description, /[\u3400-\u9fff]/, `${id} needs a Chinese rules description`);
  assert.match(ruleSource, /Tuscany Essential/);
  assert.match(ruleSource, /第6页/);
}

const expansionSource = fs.readFileSync(path.join(root, 'internal/game/expansions.go'), 'utf8');
assert.match(expansionSource, /SpecialWorkers/);
assert.doesNotMatch(expansionSource, /SpecialWorkers[^\n]+未完成/);

const structureSource = fs.readFileSync(path.join(root, 'internal/game/structures.go'), 'utf8');
const structures = [...structureSource.matchAll(/\{"([a-z_]+)",\s*"([^"]+)",\s*(\d+),\s*(true|false),\s*"(action|enhancement|residual)",\s*"([^"]+)",\s*structureRuleSource\}/g)];
const structureIds = ['cask','aqueduct','wine_cave','trading_post','shop','wine_press','school','wine_bar','patio','ristorante','guest_house','cafe','distiller','mercado','studio','barn','academy','gazebo','workshop','veranda','wine_parlor','label_factory','harvest_machine','fermentation_tank','charmat','inn','tap_room','tavern','banquet_hall','penthouse','fountain','mixer','storehouse','statue','dock','silo'];
assert.equal(structures.length, 36, 'Tuscany structure catalogue must contain exactly 36 entries');
assert.deepEqual(structures.map(([, id]) => id), structureIds, 'structure IDs/order changed unexpectedly');
assert.equal(new Set(structures.map(([, id]) => id)).size, 36, 'structure IDs must be unique');
for (const [, id, name, cost, , category, description] of structures) {
  assert.ok(name.length > 0, `${id} needs a display name`);
  assert.ok(Number(cost) > 0, `${id} needs a positive cost`);
  assert.match(description, /[\u3400-\u9fff]/, `${id} needs a Chinese rules description`);
  assert.ok(['action', 'enhancement', 'residual'].includes(category), `${id} has invalid structure category`);
}
assert.match(structureSource, /const structureRuleSource = "Tuscany Essential .*第7页/);
const effectChecks = {
  cask: ['陈酿2级', '抽1张订单'], aqueduct: ['忽略棚架与灌溉'], wine_cave: ['至多2瓶', '陈酿1级'],
  trading_post: ['一换一交易'], shop: ['交付订单'], wine_press: ['至多2瓶酒'], school: ['培训1名本年可用工人', '普通免费', '特殊加1金币'],
  wine_bar: ['弃置1瓶酒', '得2分'], patio: ['桃红或起泡酒', '2金币'], ristorante: ['弃置1瓶酒', '1颗红或白葡萄'],
  guest_house: ['弃置2张访客牌'], cafe: ['弃置1颗红或白葡萄'], distiller: ['所有葡萄', '陈酿1级'],
  mercado: ['抽到订单牌', '立即交付'], studio: ['建造一座建筑', '含固定建筑和结构牌', '额外得1分'], barn: ['夏季结束', '弃2张牌'],
  academy: ['培训工人', '1金币'], gazebo: ['导览时', '影响力星'], workshop: ['少付1金币'],
  veranda: ['交付订单', '额外得1分'], wine_parlor: ['交付订单', '额外得2金币'], label_factory: ['支付3金币', '交付1张订单'],
  harvest_machine: ['所有未收获田地'], fermentation_tank: ['收获至少1块田地', '酿造1瓶酒'], charmat: ['1红葡萄和1白葡萄', '起泡酒'],
  inn: ['访客牌', '1金币'], tap_room: ['弃1瓶酒', '得2分'], tavern: ['弃2颗红/白葡萄', '得3分'],
  banquet_hall: ['移动影响力星', '地区奖励'], penthouse: ['品质至少7', '额外得1分'], fountain: ['对手导览', '1金币'],
  mixer: ['至多酿造1瓶桃红和1瓶起泡酒'], storehouse: ['所有酒', '陈酿1级'], statue: ['年末得1分'], dock: ['年末抽1张订单牌'], silo: ['年末抽1张葡萄藤'],
};
for (const [id, phrases] of Object.entries(effectChecks)) {
  const row = structures.find(([, cardId]) => cardId === id);
  assert(row, `missing structure ${id}`);
  for (const phrase of phrases) assert(row[6].includes(phrase), `${id} missing concrete effect: ${phrase}`);
}
assert.match(fs.readFileSync(path.join(root, 'internal/game/decks.go'), 'utf8'), /StructureCatalog\(\)/, 'structure deck must be config-driven');
console.log(`✔ catalogue: ${ids.length} special workers + ${structures.length} structures with Chinese descriptions, typed effects and rule sources`);
