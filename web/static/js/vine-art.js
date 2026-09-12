// Original modular vine illustration. Only normalized card attributes enter SVG markup.
function normalize(card) {
  const quantity = (value) => (Number.isInteger(value) && value >= 0 && value <= 4 ? value : 0);
  const red = quantity(card.red),
    white = quantity(card.white);
  return {
    red,
    white,
    trellis: card.trellis === true,
    irrigation: card.irrigation === true,
    shape: red && white ? 'forked' : 'arched',
  };
}
function yieldCircles(c) {
  return [
    ['red', '#852f51', '#fff0f3'],
    ['white', '#5b691c', '#ffffe5'],
  ]
    .filter(([key]) => c[key] > 0)
    .map(
      ([key, color, fill], i) =>
        '<g data-resource="' +
        key +
        '"><circle cx="' +
        (48 + i * 88) +
        '" cy="352" r="36" fill="' +
        fill +
        '" stroke="' +
        color +
        '" stroke-width="5"/><text x="' +
        (48 + i * 88) +
        '" y="354" text-anchor="middle" dominant-baseline="central" font-family="Arial, sans-serif" font-size="58" font-weight="700" fill="' +
        color +
        '">' +
        c[key] +
        '</text></g>',
    )
    .join('');
}
function grape(color, x, y, scale = 1) {
  let palette =
    color === 'red' ? ['#813b57', '#592d49', '#bc687b'] : ['#c2cc51', '#8b9d3c', '#edf19b'];
  let positions = [
    [0, 0],
    [30, 3],
    [-18, 26],
    [14, 30],
    [42, 29],
    [-7, 58],
    [27, 62],
    [8, 87],
    [17, 109],
  ];
  return `<g transform="translate(${x} ${y}) scale(${scale})"><path d="M15 -27 Q1 -18 12 0" fill="none" stroke="#5c7031" stroke-width="9" stroke-linecap="round"/>${positions.map(([a, b], i) => `<circle cx="${a + 2}" cy="${b + 3}" r="20" fill="${palette[1]}"/><circle cx="${a}" cy="${b}" r="18" fill="${palette[0]}"/><ellipse cx="${a - 5}" cy="${b - 6}" rx="5" ry="7" fill="${palette[2]}" opacity=".7"/>`).join('')}</g>`;
}
function illustration(c) {
  const count = (c.red > 0) + (c.white > 0);
  return `<svg viewBox="0 0 600 400" role="img" aria-label="${c.red ? '红葡萄 ' : ''}${c.white ? '白葡萄 ' : ''}${c.trellis ? '棚架 ' : ''}${c.irrigation ? '灌溉水渠' : ''}" xmlns="http://www.w3.org/2000/svg"><rect width="600" height="400" fill="#faf0cf"/><circle cx="510" cy="65" r="32" fill="#e9c56f"/><path d="M0 230 Q130 160 290 218 T600 170 V330 H0Z" fill="#c2c996"/><path d="M0 277 Q200 201 400 251 T600 225 V400 H0Z" fill="#96ae78"/><path d="M0 307 Q240 265 600 299 V400 H0Z" fill="#d9ab73"/>${c.irrigation ? `<g data-layer="irrigation"><path d="M600 248 Q395 283 402 328 T245 400 H600Z" fill="#c08758"/><path d="M600 262 Q431 288 439 327 T322 400 H548 Q513 351 521 322 T600 299Z" fill="#3288ae"/><path d="M580 280 Q459 303 474 326 T398 390" fill="none" stroke="#97dbe0" stroke-width="9" stroke-linecap="round"/></g>` : ''}${c.trellis ? `<g data-layer="trellis" stroke="#7c5039" stroke-width="5" stroke-linejoin="round"><path d="M92 336 V56 L120 49 V336Z M423 326 V49 L451 55 V326Z" fill="#a97448"/><path d="M61 94 V66 L479 66 V94Z" fill="#bd8850"/></g>` : ''}<g data-layer="vine"><path d="${c.shape === 'forked' ? 'M255 345 Q241 254 255 209 M255 244 Q214 184 157 160 M255 244 Q320 177 357 134' : 'M230 347 Q200 248 252 166 Q288 105 362 132'}" stroke="#795338" stroke-width="24" fill="none" stroke-linecap="round"/><path d="M248 168 Q187 107 172 78 Q236 59 276 116 Q302 62 356 82 Q345 136 304 153 Q352 154 384 198 Q312 212 271 174 Q223 225 172 199 Q204 167 248 168Z" fill="#497148"/><path d="M255 159 Q255 111 232 89 M274 157 L334 99" stroke="#6f904e" stroke-width="6" fill="none" stroke-linecap="round"/></g><g data-layer="grapes">${count === 2 ? grape('red', 185, 194, 0.86) + grape('white', 326, 189, 0.86) : c.red ? grape('red', 265, 195, 1.04) : c.white ? grape('white', 265, 195, 1.04) : ''}</g><path d="M189 354 Q233 346 275 356" stroke="#b68a58" stroke-width="8" stroke-linecap="round" fill="none"/>${yieldCircles(c)}</svg>`;
}

export function vineArt(card, cls = '') {
  const c = normalize(card);
  const image = document.createElement('img');
  image.className = (cls + ' vine-art').trim();
  image.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(illustration(c));
  image.alt = [
    '葡萄藤',
    c.red ? '红葡萄产量 ' + c.red : '',
    c.white ? '白葡萄产量 ' + c.white : '',
    c.trellis ? '需要棚架' : '',
    c.irrigation ? '需要灌溉' : '',
    !c.trellis && !c.irrigation ? '无需棚架或灌溉' : '',
  ]
    .filter(Boolean)
    .join('，');
  image.draggable = false;
  return image;
}
export function vineDescription(text) {
  const p = document.createElement('p');
  p.className = 'vine-description';
  for (const part of String(text || '').split(/(红葡萄|白葡萄|棚架|灌溉|[0-9]+)/)) {
    if (!part) continue;
    if (/^[0-9]+$/.test(part)) {
      const n = document.createElement('span');
      n.className = 'vine-quantity';
      n.textContent = part;
      p.append(n);
    } else if (/^(红葡萄|白葡萄|棚架|灌溉)$/.test(part)) {
      const n = document.createElement('strong');
      n.textContent = part;
      p.append(n);
    } else p.append(document.createTextNode(part));
  }
  return p;
}
