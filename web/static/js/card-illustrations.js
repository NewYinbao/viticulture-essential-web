import { motif, sceneSvg, svgImage } from './svg-art.js';
const extras = {
  cup: '<path d="M12 24H69V63Q39 90 12 63Z" fill="#eee6d3" stroke="#9f754a" stroke-width="4"/><path d="M69 29Q101 23 88 53L69 55" fill="none" stroke="#9f754a" stroke-width="7"/><ellipse cx="41" cy="26" rx="27" ry="7" fill="#704a37"/><path d="M6 88H85" stroke="#9f754a" stroke-width="5"/>',
  plate:
    '<circle cx="50" cy="52" r="32" fill="#f4ead5" stroke="#ae8c56" stroke-width="5"/><circle cx="50" cy="52" r="23" fill="none" stroke="#cebb92" stroke-width="3"/><path d="M5 14V88M14 14V36H5M95 14V88M95 14Q76 32 95 49" fill="none" stroke="#718983" stroke-width="5"/>',
  bed: '<path d="M8 90V26M91 90V46M8 49H91V77H8" stroke="#856043" stroke-width="6" fill="#d6b97c"/><rect x="17" y="35" width="24" height="14" rx="6" fill="#fff2d8"/><path d="M45 48H87V70H45Z" fill="#8ca099"/>',
  tank: '<ellipse cx="50" cy="15" rx="28" ry="10" fill="#d5d7c3" stroke="#6c8a81" stroke-width="3"/><path d="M22 15V77Q50 94 78 77V15" fill="#9cb6a6" stroke="#6c8a81" stroke-width="4"/><path d="M38 25V69M21 85H80M70 63H94V76" fill="none" stroke="#dce3c9" stroke-width="6"/>',
  fountain:
    '<path d="M49 15V82M10 88H90" stroke="#a6936c" stroke-width="10"/><path d="M16 53Q48 90 83 53ZM30 24Q50 47 72 24Z" fill="#c3b38f" stroke="#8d805f" stroke-width="3"/><path d="M41 19Q50 1 59 19M26 60V73M73 60V73" fill="none" stroke="#64a5b7" stroke-width="4"/>',
  easel:
    '<path d="M24 94L50 6L78 94M15 66H86" stroke="#92643f" stroke-width="6"/><rect x="20" y="18" width="64" height="48" fill="#fff0cf" stroke="#986c44" stroke-width="3"/><path d="M30 56L48 33L61 45L75 29V57Z" fill="#809957"/>',
  boat: '<path d="M9 65H94L74 90H28Z" fill="#a86e42" stroke="#785335" stroke-width="3"/><path d="M48 10V64M55 18L85 56H55Z" fill="#f4dfac" stroke="#8a7954" stroke-width="4"/><path d="M6 97H94" stroke="#6ba6b1" stroke-width="5"/>',
  wheel:
    '<circle cx="50" cy="50" r="35" fill="#a87f4b" stroke="#7a5d3d" stroke-width="5"/><circle cx="50" cy="50" r="25" fill="#dfc392"/><path d="M50 19V81M19 50H81M28 28L72 72M28 72L72 28" stroke="#8f673d" stroke-width="5"/>',
  bell: '<path d="M24 67Q29 46 29 35Q48 7 71 35Q71 52 81 67Z" fill="#e4bd5a" stroke="#9c7840" stroke-width="4"/><circle cx="51" cy="76" r="9" fill="#9c7840"/>',
  scroll:
    '<path d="M18 15H82V86H18Z" fill="#f5e4b9" stroke="#ae8a58" stroke-width="3"/><path d="M11 14H89M11 88H89M32 36H68M32 49H60M32 62H68" stroke="#b69663" stroke-width="5"/>',
  crown:
    '<path d="M14 74L8 23L35 45L50 12L66 45L94 23L85 74Z" fill="#e5bd53" stroke="#a3803c" stroke-width="4"/><circle cx="50" cy="61" r="7" fill="#a95061"/>',
  scales:
    '<path d="M50 12V90M29 91H73M14 30H86M23 30L10 63H38ZM78 30L65 63H92Z" fill="#dac086" stroke="#967441" stroke-width="4"/>',
  flower:
    '<g fill="#c78080"><circle cx="50" cy="25" r="17"/><circle cx="73" cy="42" r="17"/><circle cx="61" cy="67" r="17"/><circle cx="32" cy="63" r="17"/><circle cx="26" cy="36" r="17"/></g><circle cx="49" cy="46" r="14" fill="#e4c258"/><path d="M49 77V99" stroke="#71944d" stroke-width="5"/>',
};
function prop(name, x, y, size) {
  return extras[name]
    ? `<g transform="translate(${x} ${y}) scale(${size / 100})">${extras[name]}</g>`
    : motif(name, x, y, size);
}
const buildings = {
  cask: ['barrel', 'bottle'],
  aqueduct: ['bridge', 'vine'],
  wine_cave: ['cave', 'barrel'],
  trading_post: ['stall', 'arrows'],
  shop: ['stall', 'bottle'],
  wine_press: ['press', 'grape'],
  school: ['school', 'book'],
  wine_bar: ['hall', 'glass'],
  patio: ['pergola', 'coins'],
  ristorante: ['hall', 'plate'],
  guest_house: ['house', 'bed'],
  cafe: ['stall', 'cup'],
  distiller: ['tank', 'bottle'],
  mercado: ['stall', 'order'],
  studio: ['house', 'easel'],
  barn: ['barn', 'cards'],
  academy: ['school', 'bell'],
  gazebo: ['pergola', 'star'],
  workshop: ['barn', 'hammer'],
  veranda: ['balcony', 'order'],
  wine_parlor: ['hall', 'coins'],
  label_factory: ['factory', 'scroll'],
  harvest_machine: ['wheel', 'shears'],
  fermentation_tank: ['tank', 'grape'],
  charmat: ['tank', 'glass'],
  inn: ['house', 'bed'],
  tap_room: ['barrel', 'glass'],
  tavern: ['hall', 'grape'],
  banquet_hall: ['hall', 'plate'],
  penthouse: ['balcony', 'bottle'],
  fountain: ['fountain', 'coins'],
  mixer: ['press', 'glass'],
  storehouse: ['barn', 'barrel'],
  statue: ['statue', 'star'],
  dock: ['dock', 'boat'],
  silo: ['silo', 'vine'],
  trellis: ['pergola', 'vine'],
  irrigation: ['bridge', 'fountain'],
  medium_cellar: ['cave', 'barrel'],
  large_cellar: ['cave', 'bottle'],
  cottage: ['house', 'worker'],
  windmill: ['mill', 'vine'],
  tasting_room: ['hall', 'glass'],
  yoke: ['barn', 'shears'],
};
function buildingShape(kind, seed) {
  const wall = ['#dcb680', '#d2b794', '#dfbf8e'][seed % 3];
  if (kind === 'bridge')
    return '<path d="M19 141V72H204V141H176V99Q152 66 129 99V141H103V99Q79 66 56 99V141Z" fill="#c6ae84" stroke="#8c7754" stroke-width="4"/><path d="M16 65H209M23 153Q90 136 211 151" stroke="#78a9b1" stroke-width="9" fill="none"/>';
  if (kind === 'cave')
    return (
      '<path d="M24 154L39 87L88 29L150 28L202 84L221 154Z" fill="#a7a78d"/><path d="M68 153V102Q118 41 168 102V153Z" fill="#575c4b"/>' +
      motif('barrel', 84, 88, 75)
    );
  if (kind === 'pergola')
    return '<path d="M41 157V63H194V157M22 64L119 23L215 64Z" fill="#bb875a" stroke="#806146" stroke-width="6"/><path d="M61 90H175" stroke="#718348" stroke-width="11"/>';
  if (kind === 'statue')
    return (
      '<path d="M62 152H174V137H62Z" fill="#a99f80"/>' +
      motif('worker', 68, 32, 102).replaceAll('#5e8294', '#9ca58d').replaceAll('#dba16f', '#b9b69a')
    );
  if (kind === 'dock')
    return '<path d="M20 114H217V130H20M44 130V162M176 130V162" fill="#ae8554" stroke="#7a5a39" stroke-width="5"/><path d="M4 157H220" stroke="#74a6ad" stroke-width="16"/>';
  if (['press', 'barrel', 'wheel', 'tank', 'fountain'].includes(kind))
    return prop(kind, 44, 19, 145);
  if (kind === 'silo')
    return '<path d="M69 151V48Q119 10 167 48V151Z" fill="#c6b183" stroke="#8c7854" stroke-width="4"/><path d="M69 48L117 17L167 48M83 71H154M83 95H154M83 119H154" stroke="#8c7854" stroke-width="4" fill="none"/>';
  const roof =
    kind === 'factory'
      ? '<path d="M29 74L66 44V73L107 43V74L159 43V74H204V95H29Z" fill="#ad7050"/>'
      : '<path d="M23 76L117 23L211 76Z" fill="#ae6b48" stroke="#82563b" stroke-width="4"/>';
  return `<rect x="39" y="74" width="157" height="82" fill="${wall}" stroke="#94754d" stroke-width="4"/>${roof}<path d="M98 156V104H131V156" fill="#715e45"/><path d="M57 91H80V118H57ZM151 91H175V118H151Z" fill="#8cafa8" stroke="#8a754f" stroke-width="3"/>${kind === 'balcony' ? '<path d="M31 119H204M40 104V132M58 104V132M77 104V132M154 104V132M175 104V132M194 104V132" stroke="#826744" stroke-width="5"/>' : ''}${kind === 'stall' ? '<path d="M27 82H209L198 105H37Z" fill="#7a956b"/>' : ''}${kind === 'mill' ? '<path d="M117 17V109M73 63H161M87 33L147 93M87 93L147 33" stroke="#e8d6a5" stroke-width="12"/>' : ''}${kind === 'school' ? prop('bell', 100, 35, 36) : ''}${kind === 'barn' ? '<path d="M48 149L188 80M48 80L188 149" stroke="#f0d8a1" stroke-width="5"/>' : ''}`;
}
function hash(text) {
  let h = 0;
  for (const c of text || '') h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return h;
}
export function buildingCardArt(id, cls = '', label = '建筑') {
  const seed = hash(id),
    [kind, tool] = buildings[id] || ['house', 'hammer'];
  return svgImage(
    sceneSvg(buildingShape(kind, seed) + prop(tool, 211, 95, 76), '#f5e2c7'),
    (cls + ' illustrated-art').trim(),
    label,
  );
}
const professions = [
  [
    /blacksmith|铁匠|stone|石匠|contract|承包|craft|工匠|手艺|handyman|杂务|artisan|工坊/i,
    'hammer',
    'apron',
  ],
  [/architect|建筑师|设计|designer|survey|测量|planner|规划/i, 'scroll', 'cap'],
  [/teacher|教授|profess|scholar|学者|教|mentor|导师|speaker|讲师|学校/i, 'book', 'scholar'],
  [/queen|女王|noble|贵族|govern|总督|politic|政治|politico|council|议员/i, 'crown', 'formal'],
  [/judge|评审|评估|assess|law|法官|律师/i, 'scales', 'formal'],
  [/chef|厨|cook|餐|banquet|宴会/i, 'plate', 'chef'],
  [/innkeeper|旅店|客房|hotel|旅馆|host|招待/i, 'bed', 'apron'],
  [/crush|压榨|oenolog|酿酒|vintner|zymolog|发酵|bottl|装瓶/i, 'press', 'apron'],
  [/critic|酒评|taster|品酒|sommelier/i, 'glass', 'formal'],
  [/harvest|收割|收获|reaper|劳工|laborer/i, 'basket', 'straw'],
  [
    /plant|种植|grow|栽培|hortic|园艺|agricult|农艺|cultivat|耕作|farmer|佃农|sharecrop|landscap|园景|homestead|拓荒/i,
    'vine',
    'straw',
  ],
  [
    /merchant|商|broker|经纪|bank|银行|buyer|采购|patron|赞助|sponsor|资助|benefact|恩人|account|会计|vendor|negotiat|谈判|auction|拍卖/i,
    'coins',
    'formal',
  ],
  [/tour|导游|travel|旅行|messeng|信使|coach|马车|caravan|商队|sailor|水手/i, 'map', 'cap'],
  [/entertain|艺人|wedding|婚宴|party|宾客|flower|花|danc|舞/i, 'flower', 'festive'],
  [/soldato|士兵|警|guard|守卫/i, 'star', 'cap'],
  [/oracle|预言|巫|占卜/i, 'star', 'scholar'],
];
function effectProp(text) {
  for (const [re, prop] of [
    [/培训|工人/, 'worker'],
    [/种植|葡萄藤/, 'vine'],
    [/建造|建筑/, 'house'],
    [/交付|订单/, 'order'],
    [/酿|陈酿/, 'barrel'],
    [/收获|葡萄/, 'grape'],
    [/金币|收入/, 'coins'],
    [/影响力|分/, 'star'],
  ])
    if (re.test(text || '')) return prop;
  return 'cards';
}
function portrait(c) {
  const seed = hash(c.id),
    title = [c.name, c.englishName].filter(Boolean).join(' ');
  const role = professions.find(([re]) => re.test(title));
  const tool = role?.[1] || effectProp(c.description),
    dress = role?.[2] || (c.type === 'mama' ? 'festive' : c.type === 'papa' ? 'formal' : 'cap');
  const coat = ['#687e95', '#9b645a', '#738658', '#ad855a', '#766b8b'][seed % 5];
  const skin = ['#e1ad80', '#cc936a', '#b87e59', '#e8bf98'][(seed >>> 3) % 4];
  const hair = ['#5b4534', '#8e7253', '#bcb096', '#704c36'][(seed >>> 5) % 4];
  const sly = /swindler|骗子|thief|贼|小偷|mafioso/i.test(title);
  const hat =
    dress === 'straw'
      ? '<path d="M76 54Q126 39 175 54M96 47L102 22H151L159 47" fill="#dfc176" stroke="#9c8551" stroke-width="5"/>'
      : dress === 'chef'
        ? '<path d="M97 49V25Q83 4 105 10Q126 -4 144 10Q168 2 162 28V49Z" fill="#fff3d9" stroke="#c3b493" stroke-width="3"/>'
        : dress === 'scholar'
          ? '<path d="M82 31L127 13L174 31L129 50Z" fill="#655477"/>'
          : dress === 'cap'
            ? '<path d="M96 44Q111 13 154 35L171 50H94Z" fill="' + coat + '"/>'
            : dress === 'formal'
              ? '<path d="M104 43V19H152V43M88 46H166" fill="' +
                coat +
                '" stroke="' +
                coat +
                '" stroke-width="5"/>'
              : '';
  const body = `<path d="M63 176L73 125Q89 104 111 104H147Q178 109 185 128L193 176Z" fill="${coat}" stroke="#665b4a" stroke-width="3"/><path d="M111 106L129 130L146 106" fill="#f3e4c4"/>${dress === 'apron' || dress === 'chef' ? '<path d="M101 129H156L164 177H94Z" fill="#e8d8b0"/>' : ''}<rect x="116" y="86" width="24" height="29" rx="8" fill="${skin}"/><ellipse cx="128" cy="65" rx="33" ry="39" fill="${skin}" stroke="#9a775a" stroke-width="2"/><path d="M94 62Q85 23 124 20Q161 18 162 59L149 40Q121 51 105 39Z" fill="${hair}"/>${hat}<path d="M109 61L118 60M137 60L146 61" stroke="#5d4b3a" stroke-width="3" stroke-linecap="round"/><path d="M126 63L124 77H131" stroke="#b3815e" stroke-width="2" fill="none"/><path d="${sly ? 'M116 88L140 82' : 'M116 85Q129 94 142 84'}" fill="none" stroke="#895841" stroke-width="3" stroke-linecap="round"/>${dress === 'scholar' ? '<circle cx="114" cy="64" r="10" fill="none" stroke="#7b6b52" stroke-width="2"/><circle cx="141" cy="64" r="10" fill="none" stroke="#7b6b52" stroke-width="2"/><path d="M124 63H131" stroke="#7b6b52"/>' : ''}`;
  return sceneSvg(
    prop(effectProp(c.description), 218, 18, 55) + body + prop(tool, 15, 96, 80),
    c.type === 'winter'
      ? '#e3ebef'
      : c.type === 'mama' || c.type === 'papa'
        ? '#efe3d0'
        : '#f5e8c7',
  );
}
function orderArt(c) {
  const colors = { red: '#873853', white: '#a6af45', blush: '#c98282', sparkling: '#d6b346' };
  const req = Array.isArray(c.requirements) ? c.requirements : [];
  let contents = motif('order', 10, 25, 82);
  req.slice(0, 3).forEach((w, i) => {
    const color = colors[w.type] || colors.red,
      x = 94 + i * 65;
    contents += `<g transform="translate(${x} 22)"><path d="M16 0H37V27L49 41V117H4V41L16 27Z" fill="${color}" stroke="#645d40" stroke-width="3"/><path d="M15 1H38V12H15Z" fill="#dfcf9e"/>${w.type === 'sparkling' ? '<circle cx="17" cy="43" r="4" fill="#fff1b9"/><circle cx="35" cy="60" r="3" fill="#fff1b9"/>' : ''}<circle cx="26" cy="99" r="21" fill="#fff5db" stroke="${color}" stroke-width="3"/><text x="26" y="109" text-anchor="middle" font-size="29" font-weight="700" fill="${color}">${Number.isInteger(w.value) && w.value >= 0 && w.value <= 9 ? w.value : 0}</text></g>`;
  });
  return sceneSvg(contents, '#eee5d7');
}
export function illustratedCard(c, cls = '') {
  if (c.type === 'structure') return buildingCardArt(c.structureId || c.id, cls, c.name);
  const svg = c.type === 'order' ? orderArt(c) : portrait(c);
  return svgImage(svg, (cls + ' illustrated-art').trim(), c.name || '卡牌插画');
}
