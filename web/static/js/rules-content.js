import { bonusLabel } from './action-options.js';
import { baseBuildings as buildings } from './building-catalog.js';
import { cardInfo } from './card-i18n.js';
import { choiceFields, visitorCost, visitorOptionText } from './visitor-ui.js';

// Lobby and in-game module descriptions. The server remains authoritative for
// the exact legal action and the current saved configuration.
export const expansionDescriptions = [
  [
    '新手组合',
    '新手可先使用 EE 本体；熟悉后选择 Tuscany 四季主板，或独立加入建筑卡、特殊工人。访客牌组在 EE、EE+Moor、Rhine 中三选一。',
  ],
  [
    'Tuscany 主板',
    '替换本体主板，增加四季工人行动、影响力地图和资源交易。玩家过季时立即领取自己的奖励；冬末逐人结算。终局门槛改为 25 分，之后结算影响力。',
  ],
  [
    'Tuscany 建筑与工人',
    '两个模块各自独立。36 张建筑卡带来私有行动或持续、年末收益；11 种特殊工人在开局抽出两种供全桌培训。可与 EE 或 Tuscany 主板、任一合法访客牌组搭配。',
  ],
  [
    'Moor 与 Rhine',
    'Moor 的40张访客加入本体访客；Rhine 的80张访客替换整套旧访客，不能与本体或Moor混洗。Rhine 的4张专用牌只在启用 Tuscany 主板时加入，与建筑卡、特殊工人开关无关。',
  ],
];

// Concise Chinese reference for Essential Edition multiplayer, written for this UI.
// Sources: Stonemaier Games EE rulebook pp. 4–17 and official FAQ; see docs/rules-review.md.
export const ruleTopics = [
  {
    id: 'overview',
    title: '入门目标',
    art: 'estate',
    intro: '从葡萄藤到订单，一步步经营你的酒庄。',
    cards: [
      [
        '赢得胜利分',
        '种藤 → 收获葡萄 → 酿酒 → 交付订单，是最容易上手的得分路线。访客和部分建筑也能得分。',
      ],
      [
        '20 分与终局',
        '达到 20 分后仍要完成当前年份，再比较最终分数。同分依次比较金币、酒总品质、葡萄总品质。',
      ],
      [
        '认识四类手牌',
        '绿牌是葡萄藤，紫牌是订单；黄牌是夏季访客，蓝牌是冬季访客。牌背类型和数量公开，牌面仅自己可见。',
      ],
      [
        '开局家族',
        '母亲给起始手牌与工人；父亲给金币，再让你选择赠礼或额外金币。先看已有资源，再选路线。',
      ],
    ],
  },
  {
    id: 'seasons',
    title: '四季流程',
    art: 'cottage',
    intro: '工人是一整年共用的，要给冬天留人。',
    cards: [
      [
        '春 · 选择起床顺序',
        '早起先行动，晚起有奖励。每个起床位置一年只能选一次；顺序决定夏季和冬季的行动先后。',
      ],
      [
        '夏 · 建设与种植',
        '抽藤、导览赚钱、建造、种植、打夏季访客，或出售葡萄／交易田地。结束夏季会把剩余工人留给冬季。',
      ],
      [
        '秋 · 邀请访客',
        '选择抽一张夏季或冬季访客，不用工人。有小屋可额外抽一张，两张的颜色可以自由搭配。',
      ],
      [
        '冬与年末',
        '收获、酿酒、交单、培训或打冬季访客。年末陈年、回收工人、领取年收入，并把手牌弃至 7 张；临时工人归还。',
      ],
    ],
  },
  {
    id: 'workers',
    title: '工人与奖励',
    art: 'train',
    intro: '看准空格，也记得留住大工人。',
    cards: [
      [
        '一次行动用一名工人',
        '轮到你时，派一名待命工人并完成主行动；已派出的工人通常到年末才回来。也可结束本季。',
      ],
      [
        '人数决定格位',
        '2 人每项行动只有 1 格，没有格位奖励；3–4 人有 2 格，5–6 人有 3 格。第 1 格提供奖励，可选择放弃。',
      ],
      [
        '大工人',
        '有空格时照常占格并领取可用奖励；普通格全满时仍能进入，但溢出位置不领取格位奖励。',
      ],
      [
        '赚零钱与培训',
        '“获得 1 金币”不限人数。培训通常花 4 金币，新工人下一年可用；普通工人与大工人合计最多 6 名。',
      ],
    ],
  },
  {
    id: 'planting',
    title: '种植与田地',
    art: 'trellis',
    intro: '建筑和容量都满足，藤才能落地。',
    cards: [
      [
        '先检查设施',
        '只需拥有藤牌要求的棚架、灌溉，或两者。没有标记就不需要该设施；每次种植不用重新付建筑费用。',
      ],
      [
        '红白数值合计占容量',
        '田地容量为 5、6、7。一块田里所有藤的红值＋白值之和不能超限；红 2 白 1 的藤占 3 容量。',
      ],
      [
        '点卡查看缺什么',
        '“种植葡萄藤”即使没有可种牌也能打开。点藤卡查看全部缺项与各田空余容量；条件满足后再选藤、选田并确认。',
      ],
      [
        '田地买卖与拔藤',
        '只能卖空田，售价和买回价均等于容量。拔藤需要自家轭或相应访客；公共种植行动不能拔藤。',
      ],
    ],
  },
  {
    id: 'harvest',
    title: '收获与陈年',
    art: 'field',
    intro: '藤留在田里，收获的是葡萄。',
    cards: [
      ['每田每年一次', '选择有藤、未出售且本年未收获的田地。普通收获最多 1 块，奖励格最多 2 块。'],
      [
        '分别合计红白',
        '一块田里所有藤的红值相加产出一颗红葡萄，白值相加产出一颗白葡萄。数值为 0 的颜色不产葡萄。',
      ],
      [
        '品质与占位',
        '葡萄最高品质 9，同色同品质只能放一颗。收获时目标位置已满，会向下寻找空位；没有空位则无法保留该颗。',
      ],
      [
        '年末陈年',
        '葡萄和酒品质通常提高 1，上限 9。酒还受酒窖等级限制；位置被占或到达上限时留在原位。',
      ],
    ],
  },
  {
    id: 'wine',
    title: '酿酒与酒窖',
    art: 'cellar',
    intro: '先选酒瓶，再把葡萄加入配方。',
    cards: [
      [
        '红酒／白酒',
        '一颗红葡萄酿一瓶红酒，一颗白葡萄酿一瓶白酒。普通酿酒行动最多 2 瓶，奖励格最多 3 瓶。',
      ],
      [
        '桃红／起泡',
        '一红＋一白酿桃红，品质至少 4；两红＋一白酿起泡，品质至少 7。配方品质由葡萄数值相加。',
      ],
      [
        '酒窖等级',
        '小酒窖上限 3，中酒窖上限 6，大酒窖上限 9。桃红需要中酒窖，起泡需要大酒窖；建大酒窖须先有中酒窖。',
      ],
      [
        '看预计结果',
        '酿酒会消耗所选葡萄。酒窖或占位可能让品质下降；若低于该酒种最低品质，配方不能完成。确认前看预览。',
      ],
    ],
  },
  {
    id: 'orders',
    title: '订单与访客',
    art: 'order',
    intro: '匹配酒种与品质，再收下得分。',
    cards: [
      [
        '一枚酒对应一个要求',
        '订单每个酒符号需一瓶同种、品质不低于要求的酒。不能把两瓶低品质酒合并抵一瓶高品质酒。',
      ],
      [
        '得分与年收入',
        '交付后消耗订单与所选酒，立即得分并提高年收入。收入在年末领取，上限每年 5 金币。',
      ],
      [
        '按季节打访客',
        '夏季打黄牌，冬季打蓝牌。先完成当前访客选择，才能继续派工；需满足所选效果的费用和条件。',
      ],
      [
        '奖励格的第二张',
        '可先完成第一张访客，再决定是否打第二张。“最多”并不表示可以完全不执行主行动。',
      ],
    ],
  },
  {
    id: 'buildings',
    title: '建筑速查',
    art: 'windmill',
    intro: '每座建筑只能建一次；建造奖励格可少付 1 金币。',
    cards: buildings.map(([, name, cost, , effect]) => [name + ' · ' + cost + ' 金币', effect]),
  },
];
export const actionTopics = {
  influence: 'influence',
  trade: 'trade',
  sell_wine: 'trade',
  flip_field: 'planting',
  build_tour: 'buildings',
  plant: 'planting',
  yoke: 'planting',
  sell_grapes: 'planting',
  harvest: 'harvest',
  make_wine: 'wine',
  fill_order: 'orders',
  summer_visitor: 'orders',
  winter_visitor: 'orders',
  build: 'buildings',
  train: 'workers',
  gain_coin: 'workers',
  wake: 'seasons',
};
export const actionLessons = {
  influence: '先放完自己的六颗星，之后只能移动；移动不再领取地区即时奖励。',
  trade: '先选付出与获得的资源；奖励交易在第一次完成后继续选择，可使用刚获得的资源。',
  sell_wine: '选择一瓶酒；红白酒得1分、桃红2分、起泡4分。额外放星按当前待决选择继续。',
  flip_field: '只可出售自己没有藤的田地，或按面值买回田地；具体格位奖励见行动格。',
  build_tour: '建造或导览二选一；对应奖励格建造少付1金币，或导览多得1金币。',
  plant: '先看藤的建筑要求和容量；缺少条件的藤卡也可以点开查看。',
  build: '建筑只买一次。先看藤缺什么设施，再比较当前费用。',
  harvest: '按田地收获，藤不会消失；同一块田每年只能收获一次。',
  make_wine: '先选一瓶，再点葡萄组成配方；下方会预览酒种与品质。',
  fill_order: '先点订单，再选满足酒种和最低品质的酒。',
  summer_visitor: '先选访客卡，确认后再选择效果；奖励格可接着打第二张。',
  winter_visitor: '先选访客卡，确认后再选择效果；费用和资源要满足所选分支。',
  sell_grapes: '可出售葡萄、出售空田或买回田地；先选操作，再点资源。',
  yoke: '自家轭每年用一次，也消耗工人；收获或拔藤二选一。',
  train: '新工人下一年才可用；大工人也计入 6 名工人的上限。',
  wake: '这格奖励一张访客，不消耗工人；先选想要的颜色。',
  tour: '导览补充金币，别忘了给冬季留待命工人。',
  draw_vine: '抽到的藤可能需要设施。拿到牌后可以打开种植查看条件。',
  draw_order: '订单决定需要哪种酒、多少品质，可以据此规划酿酒。',
  gain_coin: '这项行动不限人数，每派一名工人获得 1 金币。',
};

export const ruleSources = {
  ee: [
    'EE 说明书',
    'https://shared.steamstatic.com/store_item_assets/steam/apps/414235/manuals/Viticulture_EE_Rules.pdf?t=1551153655',
  ],
  faq: ['官方 FAQ', 'https://stonemaiergames.com/games/viticulture/faq/'],
  tuscany: [
    'Tuscany 规则入口',
    'https://stonemaiergames.com/games/viticulture/tuscany-essential-edition/',
  ],
  ee_moor: [
    'Moor 出版社说明',
    'https://stonemaiergames.com/games/viticulture/moor-visitors-expansion/',
  ],
  rhine: [
    'Rhine 出版社说明',
    'https://stonemaiergames.com/games/viticulture/visit-from-the-rhine-valley/',
  ],
};
export const seasonNames = {
  spring: '春季',
  summer: '夏季',
  fall: '秋季',
  winter: '冬季',
  year_end: '个人年末',
  ready: '已预约下年',
};
export function configOf(view) {
  return { board: 'ee', visitors: 'ee', structures: false, specialWorkers: false, ...view?.config };
}
export function configurationSummary(view) {
  const c = configOf(view);
  const parts = [
    c.board === 'tuscany' ? 'Tuscany 四季主板' : 'EE 本体主板',
    c.visitors === 'rhine'
      ? 'Rhine 替换访客'
      : c.visitors === 'ee_moor'
        ? 'EE + Moor 访客'
        : 'EE 本体访客',
  ];
  if (c.structures) parts.push('建筑卡');
  if (c.specialWorkers) parts.push('特殊工人');
  return parts.join(' · ');
}
export function unsupportedModules(view) {
  const c = configOf(view),
    a = view?.expansionAvailability || {};
  return [
    [c.board === 'tuscany' && !a.tuscany, 'Tuscany 主板'],
    [c.structures && !a.structures, '建筑卡'],
    [c.specialWorkers && !a.specialWorkers, '特殊工人'],
    [c.visitors !== 'ee' && !a[c.visitors], c.visitors === 'rhine' ? 'Rhine' : 'Moor'],
  ]
    .filter(([on]) => on)
    .map(([, name]) => name);
}
const section = (id, title, art, intro, cards) => ({ id, title, art, intro, cards });
const rewardNames = {
  winter: '抽1冬季访客',
  summer: '抽1夏季访客',
  vine: '抽1藤',
  order: '抽1订单',
  coin: '1金币',
  coins2: '2金币',
};

// This is the existing EE reference with module-specific replacements. Never
// mutate the EE catalogue: room changes and old saves must return to pure EE.
export function rulesForView(view) {
  const c = configOf(view),
    tuscany = c.board === 'tuscany';
  const topics = ruleTopics.map((t) => ({ ...t, cards: t.cards.map((card) => [...card]) }));
  const find = (id) => topics.find((t) => t.id === id);
  const sources = [ruleSources.ee, ruleSources.faq];
  if (tuscany || c.structures || c.specialWorkers) sources.push(ruleSources.tuscany);
  if (c.visitors !== 'ee' && ruleSources[c.visitors]) sources.push(ruleSources[c.visitors]);
  if (tuscany) {
    find('overview').cards[1] = [
      '25 分与影响力终局',
      '所有玩家结束冬季后，若有人达到25分，结算影响力地图后比较最终分数。每人都须结束冬季；个人回收工人可能为其他玩家腾出格位。影响力分数不能先计入25分门槛。',
    ];
    Object.assign(find('seasons'), {
      title: '个人四季与起床',
      intro: '春、夏、秋、冬都有工人行动；个人过季立即领奖，全员过季后才开始下一季派工。',
      cards: [
        [
          '春季与首年起床',
          '首年依座位顺序选空的第2–7行，不能选第1行；选行时不领奖。春季可抽藤、导览、建造或放置影响力星。',
        ],
        [
          '夏季',
          '个人结束春季立即领取本行夏季奖励，工人留在原格。等全员进入夏季后，按起床顺序派工：夏访客、种藤、交易、买卖田地。',
        ],
        [
          '秋季',
          '个人进入秋季按起床行领奖；不是人人自动抽访客。小屋额外抽1张自选访客。全员进入秋季后可抽订单、收获、酿酒、建造或导览。',
        ],
        [
          '冬季',
          '个人进入冬季按起床行领奖；全员进入后可打冬访客、培训、售酒或交订单。每季都可用自家行动与赚1金币。',
        ],
        [
          '个人冬末顺序',
          '结束冬季时依次回收自己的工人、陈酿葡萄和酒、手牌弃至7张、领年收入、预约下年起床位。工人回收后，仍在冬季的玩家可使用空出的格位。',
        ],
        [
          '下年起床',
          '预约下年仍可选的第2–7行；本年在第7行并取得首位标记者，下年必须选第1行。所有人结束冬季后才进入下一年。',
        ],
        ...(view?.wakeSlots || []).map((w) => [`第${w.slot}行奖励`, w.bonus]),
      ],
    });
    find('workers').cards[1] = [
      '人数与逐格奖励',
      '2人用每项行动第1格，3–4人用前2格，5–6人用前3格。Tuscany 在2人局也可能有奖励，各格奖励不同；看实际行动格，不能沿用本体“2人无奖励”。',
    ];
    find('harvest').cards[0] = [
      '每田每年一次',
      '选择有藤、未出售且本年未收获的田地。秋季收获通常1田；是否多收获或得金币取决于所占格位。',
    ];
    find('wine').cards[0] = [
      '红酒／白酒',
      '一颗红葡萄酿一瓶红酒，一颗白葡萄酿一瓶白酒。秋季普通酿酒至多2瓶，对应奖励格至多3瓶。',
    ];
    find('buildings').intro =
      '本体固定建筑仍各只能建一次；春季建造的折扣与秋季建造／导览奖励依实际格位而定。';
    topics.push(
      section(
        'influence',
        '影响力地图',
        'estate',
        '每人6颗星；只有从供应区首次放星才领取即时奖励。',
        [
          [
            '放置与移动',
            '先放完自己的6颗星，之后把自己某地区的一颗星移到另一地区。不能移动别人的星；移动不再次领奖。额外一星的奖励逐颗结算，第6颗放完后下一颗应移动。',
          ],
          [
            '区域多数',
            '终局每个地区独占最多星的玩家获得该区分数；并列最多都不得分。两人不计地图终局分是可选变体，本版本未提供该变体开关。',
          ],
          ...(c.structures
            ? [['建筑主板 side 2', '启用建筑牌堆与对应主板奖励；影响力地区仍按本局地图结算。']]
            : []),
          ...(view?.influenceRegions || []).map((r) => [
            r.name,
            `首次放星：${rewardNames[r.reward] || '待核实'}；终局独占多数：${r.points}分。`,
          ]),
        ],
      ),
    );
    topics.push(
      section('trade', '交易与售酒', 'order', '交易按资源组交换，售酒按酒种计分。', [
        [
          '交易资源',
          '一次可付出3金币、1胜利分、任意2张手牌或1颗任意品质葡萄，换取其中一种资源组。换来的葡萄为品质1，可选红白；换来的2张牌可分别选牌类。',
        ],
        [
          '额外交易',
          '交易奖励允许再交易一次，可以使用第一次所得；先完成第一次，再选择第二次，或放弃额外交易。',
        ],
        [
          '出售一瓶酒',
          '弃1瓶红／白酒得1分、桃红得2分、起泡得4分，酒的品质不影响得分。对应奖励格还可放置或移动1颗影响力星。',
        ],
      ]),
    );
  }
  if (c.structures) {
    find('overview').cards[2] = [
      '手牌与建筑牌',
      '绿藤、紫订单、黄夏访客、蓝冬访客外，再加入橙色建筑牌；橙牌也占手牌上限，只有持牌者能看牌面。',
    ];
    topics.push(
      section(
        'structures',
        'Tuscany 建筑卡',
        'windmill',
        '36张橙色建筑牌已接入有限牌堆、建造/拆除、结构垫与空田地放置，并按牌面类型触发行动、持续或年末效果。',
        [
          [
            '独立使用与草拟',
            '可独立于Tuscany主板启用。无扩展主板时，选完父母后每人抽4张建筑牌，每轮选1张、其余传给右邻，直到各选4张；计入手牌。',
          ],
          [
            '建造与占地',
            '使用建造行动，支付牌面费用（工坊减1），建在结构垫空位或自有未出售空田；田里不能有藤或其他建筑。每张结构建成得1分，工作室使后续建造任何建筑再得1分。',
          ],
          [
            '私有行动建筑',
            '木桶、酒窖洞、贸易站、商店、压酒机、学校、酒吧、餐厅、客房、咖啡馆、酒标工厂、调酒器拥有私有行动格；每年每格1名工人，须先完成牌面支付/选择，再取得中部奖励。',
          ],
          [
            '持续、年末与拆除',
            '渡槽、庭院、市场、学院、凉亭、工坊、阳台、酒廊、收割机、发酵罐、查玛法罐、旅店、品酒吧、酒馆、宴会厅、顶层阁楼、喷泉为持续效果；蒸馏器、谷仓、仓库、雕像、码头、筒仓在规定年末/季末触发。拆除不倒扣建成分。',
          ],
          [
            '核实边界',
            '逐牌中文短描述、费用、类型和规则来源随服务器结构目录投影；关键可选分支（结构影响力、酒标工厂、访客奖励、谷仓、发酵罐）会保存为 choice 后再继续行动。雕像年末分不触发最终年门槛。',
          ],
        ],
      ),
    );
  }
  if (c.specialWorkers) {
    find('workers').cards[3] = [
      '赚零钱与培训',
      '“获得1金币”不限人数。培训普通工人通常4金币，特殊工人额外1金币，通常次年可用。普通、大工人及特殊工人合计最多6名，灰色临时工不计入永久上限。',
    ];
    topics.push(
      section(
        'special-workers',
        'Tuscany 特殊工人',
        'train',
        '11种特殊工人开局抽2种供全桌培训；当前牌池可在家族信息旁查看。',
        [
          [
            '选取与培训费用',
            '在选父母前随机选2种；2人局排除旅店老板。培训特殊工人总比当前普通培训费用多1金币，包括折扣、免费或其他资源培训情形。每人每种至多1名。',
          ],
          [
            '总数与生效时机',
            '包括大工人在内总数最多6名；通常下一年才可用。特殊能力可选，涉及派遣的能力通常先于行动，除非卡面另有时机。',
          ],
          ['农夫 Farmer', '派在主板时，从该行动全部奖励中选取总计1项奖励，不受人数开放格位影响。'],
          [
            '黑手党 Mafioso',
            '派在无奖励格，行动完成后可再做同一行动；仍须满足每次行动条件，不能重复收获同一田。',
          ],
          ['厨师 Chef', '可把对手占格的工人退回其待命区并占该格，不能挤走厨师。'],
          [
            '旅店老板 Innkeeper',
            '派遣时可付1金币给同一行动上有工人的对手，选夏／冬颜色后随机拿其1张访客。2人局不使用。',
          ],
          ['教授 Professore', '派遣时可从本季主板行动格收回自己的1名普通工人，本年可再次派遣。'],
          [
            '士兵 Soldato',
            '在主板行动上时，对手派入同一行动须付你1金币；即使全部格位占满，对手仍可付费进入。',
          ],
          ['政客 Politico', '派在主板奖励格，完成行动与奖励后可付1金币，再拿一次该格奖励。'],
          ['先知 Oracle', '因先知抽牌时，同类多抽1张，再弃掉本次抽到的1张；每回合至多多抽1张。'],
          ['商人 Merchant', '所有对手已过季后派在主板，行动完成后可抽1张任意类别牌。'],
          ['旅行者 Traveler', '可派在本年之前季节的空行动格并立即执行，不受人数开放格位限制。'],
          [
            '信使 Messenger',
            '先预约未来季节行动格；该季第一个回合再选当前资源执行，不另派工人。无法执行则不能领奖，且该回合结束。',
          ],
        ],
      ),
    );
  }
  if (c.visitors !== 'ee')
    topics.push(
      section(
        'visitors',
        c.visitors === 'rhine' ? 'Rhine 访客牌组' : 'Moor 访客牌组',
        'order',
        c.visitors === 'rhine'
          ? 'Rhine 替换全部其他访客，侧重酿酒经营。'
          : 'Moor 将40张访客加入本体76张访客。',
        [
          [
            '牌堆搭配',
            c.visitors === 'rhine'
              ? '共40夏季、40冬季访客，不能混入本体或Moor。未启用Tuscany主板时移除老将军、建筑重组者、影响者、说客，共使用76张；启用主板使用80张。'
              : '夏冬各增加20张，仍按访客颜色使用；可与Tuscany三个模块独立搭配，不能同时使用Rhine替换牌组。',
          ],
          [
            '选牌与分步结算',
            '悬停手牌可查看中文效果；打出后按卡片选择分支与资源，条件不足会给出短原因。涉及其他玩家或后续步骤时，当前选择的规则会跟随响应者与阶段更新；刷新后可继续未完成的选择。',
          ],
        ],
      ),
    );
  const blocked = unsupportedModules(view);
  if (tuscany)
    topics.push(
      section(
        'actions',
        '本局行动与奖励',
        'train',
        '以下主行动来自服务器本局主板；逐格奖励不能从EE格位推断。',
        (view?.spaces || []).map((s) => [
          s.name,
          `${seasonNames[s.season] || '各季'}：${s.description}。` +
            Array.from(
              { length: s.capacity || 0 },
              (_, i) => '第' + (i + 1) + '格：' + bonusLabel(s, i + 1),
            ).join('；'),
        ]),
      ),
    );
  return {
    title: configurationSummary(view),
    topics,
    sources,
    warning:
      view?.ruleSupport?.reason ||
      (blocked.length ? `${blocked.join('、')} 尚未完成；本配置暂不可开局。` : ''),
  };
}

const stages = {
  effect: '选择效果',
  reply: '多人响应',
  reward: '选择奖励',
  second: '第二张访客',
  plant: '后续种藤',
  make: '后续酿酒',
  fill: '后续交单',
  build: '后续建造',
  take: '选取翻牌',
};
const fieldNames = {
  cards: '手牌',
  wines: '酒',
  grapes: '葡萄',
  plant: '藤与田地',
  fields: '田地',
  recipes: '酿酒配方',
  building: '建筑',
  buildings: '建筑',
  targets: '玩家',
  space: '行动',
  placementSlot: '格位',
  worker: '工人',
  wake: '起床行',
  color: '颜色',
  colors: '牌类',
  seats: '已派工人',
  revealed: '公开翻牌',
  manager: '前季行动',
  swap: '交换的藤',
  uproot: '拔除的藤',
};
export function pendingRule(view, selection) {
  const c = view?.pendingChoice;
  if (!c) return null;
  if (c.playerId !== view.youId)
    return section('current', '当前待决', 'order', '等待当前响应者完成选择。', [
      [
        '现在不能操作',
        '只有当前响应者可提交。你可以查看自己的手牌；不展示对方的牌面、分支、费用或目标。',
      ],
    ]);
  const p = view.players.find((p) => p.id === view.youId);
  if (c.visitor || c.kind === 'planner') {
    const info =
        c.title && c.description
          ? { name: c.title, description: c.description }
          : cardInfo(c.visitor?.cardId),
      selected = selection && selection.choiceId === c.id ? selection.option : null;
    const cards = [];
    if (c.visitor?.cardId && !info)
      return section('current', '当前待决', 'order', '本牌效果尚未核实', [
        [
          '未提供可执行说明',
          '本地没有本牌的已核实中文效果与分支。该模块仍未完成，不能套用本体访客说明。',
        ],
      ]);
    if (info) cards.push([info.name, info.description]);
    if (c.kind === 'planner')
      cards.push(['预约行动', '执行已占的预约格，不再消耗工人；仍须满足本次行动的成本和条件。']);
    for (const o of c.options || []) {
      const fields = choiceFields(c, o)
        .filter(
          (f) =>
            configOf(view).specialWorkers ||
            (f.type === 'trainingWorker' && p?.grandeRemoved) ||
            !['specialWorker', 'trainingWorker'].includes(f.type),
        )
        .map((f) => `${fieldNames[f.type] || '所需资源'} ${f.min ?? 1}–${f.max ?? 1}`)
        .join('；');
      cards.push([
        (selected === o ? '已选 · ' : '') + visitorOptionText(c, o, p),
        [
          visitorCost(c, o, p, view),
          fields && '需要选择：' + fields,
          view.optionReasons?.[o] && '暂不可选：' + view.optionReasons[o],
        ]
          .filter(Boolean)
          .join('。') || '此分支无需额外选资源，确认当前步骤执行。',
      ]);
    }
    if (selection?.choiceId === c.id && selection.reason)
      cards.push(['尚不能确认', selection.reason]);
    return section(
      'current',
      '当前待决',
      'order',
      `${info?.name || (c.kind === 'planner' ? '预约' : '访客')} · ${stages[c.visitor?.stage] || '后续选择'}`,
      cards,
    );
  }
  const entries = {
    messenger: [
      '执行信使预约',
      '打开预约行动，使用当前手牌、葡萄和金币选择资源；不再消耗工人。成功后取得该格奖励，完成这次回合。',
    ],
    special_train: [
      '选择培训的工人',
      '从当前可培训类型中选择。费用、折扣和本年或次年可用以按钮说明为准；大工人与特殊工人都计入永久工人上限。',
    ],
    special_farmer: [
      '农夫选择奖励',
      '从该行动奖励中总计选1项，或跳过。预约派工时先锁定选择，到目标季成功执行基础行动时结算。',
    ],
    special_professore: [
      '教授收回工人',
      '可收回列出的1名已派普通工人，本年再次使用；也可以跳过。选择后继续当前行动或预约流程。',
    ],
    special_innkeeper: [
      '旅店老板交换访客',
      '选择列出的对手与牌色，付对方1金币，随机取得其1张该色访客；不能先看牌再换色，也可跳过。',
    ],
    special_mafioso: [
      '黑手党再次行动',
      '基础行动已完成。重新选择当前资源，可再做1次相同行动，不再取得格位奖励；仍须支付费用，或选择跳过。',
    ],
    special_politico: [
      '政客重复奖励',
      '可付1金币，再取得本格奖励。需要资源的奖励须先选好资源再确认；跳过则不支付。',
    ],
    special_oracle: [
      '先知弃置抽牌',
      '只从本次抽到并列出的候选中弃1张，其余留下；不能弃旧手牌代替。',
    ],
    special_merchant: ['商人抽牌', '行动已完成，选择1种允许的牌类，抽取1张。'],
    structure_draft: [
      '草拟建筑牌',
      '从当前传来的牌组中选1张，余牌向右传递；共4轮，选到的建筑加入自己的手牌。',
    ],
    structure_mercado: [
      '市场即时交付',
      '只能交付本次新抽到并列出的订单。选择满足订单的酒再确认，也可以跳过。',
    ],
    papa: ['家族赠礼', '比较父亲赠礼与额外金币；这一步不消耗工人，选定后继续其他玩家的家族选择。'],
    fall: ['秋季抽访客', '按可选按钮抽夏季／冬季访客；小屋的额外一张可与第一张同色或不同色。'],
    discard: [
      '年末弃牌',
      `须从自己的手牌弃置恰好${c.count}张，保留不超过7张；在原手牌区选择后确认。`,
    ],
    tuscany_draw: [
      '过季抽牌奖励',
      '从本次奖励允许的牌类中选1张，不消耗工人。小屋或其他后续奖励会分别继续提示。',
    ],
    tuscany_influence: [
      '额外影响力星',
      '选1个目标区域；六颗尚未放完时从供应放入，否则必须选择自己已有星的另一区域作为起点。移动不领奖，也可放弃这次奖励。',
    ],
    structure_influence: [
      '结构奖励',
      '贸易站、商店或压酒机结算后，可放置或移动1颗影响力星；在没有Tuscany影响力主板时不会出现该奖励。',
    ],
    structure_label_factory: ['酒标工厂', '交付订单后可支付3金币额外获得2分；选择跳过则不支付。'],
    structure_barn: ['谷仓', '夏季结束可弃置任意2张手牌换1分；手牌仍按原手牌区选择。'],
    structure_fermentation: ['发酵罐', '收获至少1块田后，可在收获动作完成后酿造1瓶酒或跳过。'],
    structure_visitor_bonus: [
      '访客结构奖励',
      '打出访客牌后，品酒吧可弃1瓶酒得2分，酒馆可弃2颗红/白葡萄得3分；可分别跳过。',
    ],
    tuscany_trade: [
      '第二次交易',
      '第一次交易已结算。可以使用刚得到的资源，再做1次交易，也可以跳过；费用以当前资源为准。',
    ],
    tuscany_upkeep: [
      '领取年收入',
      `回收工人、陈酿与弃牌已经处理。确认后领取本人的${p?.income || 0}金币年收入，再预约下年起床位。`,
    ],
    tuscany_next_wake: [
      '预约下一年',
      '从服务器列出的空起床行中选择；取得首位标记者必须选第1行。预约后等全员结束冬季。',
    ],
  };
  const entry = entries[c.kind] || [
    '选择尚未支持',
    '此阶段尚无经过核实的操作说明，不能用普通访客提示代替。',
  ];
  return section('current', '当前待决', 'order', entry[0], [entry]);
}

// Contextual copy must not promise an EE double-card bonus on Tuscany's coin slot.
export function actionLessonForView(view, id) {
  if (configOf(view).board === 'tuscany' && ['summer_visitor', 'winter_visitor'].includes(id))
    return '先选择实际格位：金币奖励可在访客前或后结算；只有标注双访客的格子才能接着打第二张。';
  return actionLessons[id] || '先选工人和资源，再确认派遣。';
}
