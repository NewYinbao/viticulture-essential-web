// Text is tokenized into DOM nodes, never parsed as HTML.
const keywords =
  /(红葡萄酒|白葡萄酒|红葡萄|白葡萄|红\/白葡萄|葡萄藤|桃红酒|起泡酒|红酒|白酒|葡萄|金币|胜利分|年收入|收入|影响力星|订单|夏季访客|冬季访客|访客牌|工人|棚架|灌溉|中酒窖|大酒窖|建筑|田地|每年一次|每年最多|每年|至多|至少|必须|不能|本年|下一年|[0-9]+(?:\.[0-9]+)?)/g;
export function cardText(text, tag = 'p', cls = '') {
  const el = document.createElement(tag);
  el.className = ('card-description ' + cls).trim();
  for (const part of String(text ?? '').split(keywords)) {
    if (!part) continue;
    if (/^[0-9]+(?:\.[0-9]+)?$/.test(part)) {
      const n = document.createElement('span');
      n.className = 'card-number';
      n.textContent = part;
      el.append(n);
    } else if (new RegExp('^(?:' + keywords.source + ')$').test(part)) {
      const n = document.createElement('strong');
      n.textContent = part;
      el.append(n);
    } else el.append(document.createTextNode(part));
  }
  return el;
}
