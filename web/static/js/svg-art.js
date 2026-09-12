// Shared original SVG motifs. Geometry is fixed; card text never enters SVG markup.
export const motifs = {
  vine: '<path d="M49 90 Q34 60 52 29" fill="none" stroke="#805634" stroke-width="8"/><path d="M48 44 Q10 45 20 15 Q48 17 48 44 M51 35 Q80 7 89 29 Q79 51 51 35" fill="#608348"/><g fill="#8b4262" stroke="#653047" stroke-width="2"><circle cx="45" cy="51" r="10"/><circle cx="62" cy="51" r="10"/><circle cx="39" cy="67" r="10"/><circle cx="56" cy="68" r="10"/><circle cx="49" cy="83" r="10"/></g>',
  grape:
    '<path d="M52 24Q43 5 60 7" stroke="#60773a" stroke-width="6" fill="none"/><path d="M54 23Q74 0 89 22Q73 36 54 23" fill="#738b46"/><g fill="#894768" stroke="#63334b" stroke-width="2"><circle cx="36" cy="35" r="15"/><circle cx="64" cy="35" r="15"/><circle cx="27" cy="60" r="15"/><circle cx="54" cy="59" r="15"/><circle cx="46" cy="83" r="15"/></g>',
  coins:
    '<g fill="#e6bb57" stroke="#947137" stroke-width="3"><ellipse cx="44" cy="77" rx="32" ry="13"/><path d="M12 62V77M76 62V77"/><ellipse cx="44" cy="62" rx="32" ry="13"/><path d="M12 47V62M76 47V62"/><ellipse cx="44" cy="47" rx="32" ry="13"/><circle cx="69" cy="25" r="19"/></g><circle cx="69" cy="25" r="11" fill="none" stroke="#fff0ad" stroke-width="3"/>',
  house:
    '<rect x="19" y="42" width="65" height="49" rx="3" fill="#e0bb83" stroke="#805c41" stroke-width="3"/><path d="M9 44L51 10L94 44Z" fill="#a46143" stroke="#805c41" stroke-width="3"/><path d="M44 91V60H62V91" fill="#674f3c"/><path d="M26 54H37V69H26ZM68 54H78V69H68Z" fill="#91bab1"/>',
  hammer:
    '<path d="M33 86L63 22" stroke="#7c573b" stroke-width="13" stroke-linecap="round"/><path d="M35 8L72 20L84 39L69 40L62 30L27 22Z" fill="#70918d" stroke="#3c605e" stroke-width="3"/>',
  shovel:
    '<path d="M49 14V68" stroke="#93653e" stroke-width="10"/><path d="M31 7H67V23Q48 41 31 23Z" fill="none" stroke="#684a32" stroke-width="6"/><path d="M31 61H69V80L50 99L31 80Z" fill="#7c9a91" stroke="#42645f" stroke-width="3"/>',
  field:
    '<path d="M4 47L53 18L96 49L51 89Z" fill="#b99460" stroke="#7c7046" stroke-width="3"/><path d="M20 49L60 26M32 62L74 36M45 76L89 49" stroke="#4f7740" stroke-width="9" stroke-linecap="round"/>',
  basket:
    '<path d="M12 44L23 91H78L91 44Z" fill="#c39a60" stroke="#86613d" stroke-width="4"/><path d="M26 46V32C26 5 75 5 75 32V46M18 60H85M20 76H81M39 46V89M59 46V89" fill="none" stroke="#86613d" stroke-width="4"/>',
  bottle:
    '<path d="M40 8H60V29L73 44V90H27V44L40 29Z" fill="#49694b" stroke="#304f3b" stroke-width="3"/><path d="M39 8H61V20H39Z" fill="#aa624a"/><rect x="30" y="52" width="40" height="24" rx="2" fill="#efe2b7"/><circle cx="50" cy="64" r="7" fill="#91475f"/>',
  press:
    '<path d="M15 91V15H86V91M8 92H93M50 5V51M29 51H72" fill="none" stroke="#7e563a" stroke-width="8"/><path d="M22 56H81L74 83H29Z" fill="#bb8953" stroke="#805d3d" stroke-width="3"/><path d="M50 25H93" stroke="#9f7550" stroke-width="6"/>',
  order:
    '<path d="M19 7H73L87 23V93H19Z" fill="#fff5d7" stroke="#a18554" stroke-width="3"/><path d="M72 8V25H86M31 40H71M31 52H64M31 64H71" fill="none" stroke="#b6a17b" stroke-width="4"/><circle cx="69" cy="79" r="12" fill="#9b4d53"/>',
  cards:
    '<rect x="9" y="17" width="54" height="70" rx="6" transform="rotate(-12 36 52)" fill="#d7b776" stroke="#957850" stroke-width="3"/><rect x="36" y="10" width="54" height="75" rx="6" fill="#fff1d4" stroke="#957850" stroke-width="3"/><path d="M64 27Q41 35 57 48Q65 31 77 34Q88 55 64 65" fill="#6c894d"/>',
  worker:
    '<circle cx="50" cy="21" r="16" fill="#dba16f" stroke="#8c5c41" stroke-width="3"/><path d="M32 43H68L85 66L72 73L64 62L68 95H32L36 62L27 73L15 66Z" fill="#5e8294" stroke="#395b70" stroke-width="3"/>',
  book: '<path d="M50 22Q25 7 5 17V82Q28 72 50 90Q72 72 95 82V17Q75 7 50 22Z" fill="#eee0b5" stroke="#846d4c" stroke-width="4"/><path d="M50 23V88M16 33L37 37M16 45L37 49M63 37L83 33M63 49L83 45" fill="none" stroke="#b19b76" stroke-width="3"/>',
  star: '<path d="M50 6L63 34L95 38L71 60L78 94L50 78L21 94L27 60L5 38L37 34Z" fill="#e2b756" stroke="#a17b35" stroke-width="4"/>',
  map: '<path d="M5 20L34 9L66 23L95 12V84L66 95L34 81L5 91Z" fill="#d5d8a8" stroke="#839368" stroke-width="3"/><path d="M34 10V81M66 23V94" stroke="#a7af82" stroke-width="3"/><path d="M12 57Q42 27 62 56T92 45" fill="none" stroke="#68a8b0" stroke-width="7"/>',
  arrows:
    '<path d="M9 35H82L66 18M90 65H17L34 82" fill="none" stroke="#75886a" stroke-width="10" stroke-linecap="round" stroke-linejoin="round"/>',
  sun: '<circle cx="50" cy="50" r="22" fill="#e5b84d"/><path d="M50 6V17M50 83V94M6 50H17M83 50H94M18 18L27 27M73 73L82 82M18 82L27 73M73 27L82 18" stroke="#d69a3e" stroke-width="6"/>',
  snow: '<path d="M50 6V94M12 27L88 73M12 73L88 27M37 17L50 30L63 17M37 83L50 70L63 83" stroke="#78a5b1" stroke-width="6" fill="none" stroke-linecap="round"/>',
  barrel:
    '<path d="M25 8H75Q96 51 75 94H25Q4 51 25 8Z" fill="#b68451" stroke="#7b5635" stroke-width="3"/><path d="M20 25H80M16 73H84" stroke="#586e66" stroke-width="9"/><path d="M41 12Q30 50 41 90M59 12Q70 50 59 90" stroke="#855d37" stroke-width="3" fill="none"/>',
  glass:
    '<path d="M23 10H77L72 44Q50 71 28 44Z" fill="#f6ebd2" stroke="#67857b" stroke-width="4"/><path d="M28 27H72L69 43Q50 60 31 43Z" fill="#98465e"/><path d="M50 56V88M31 91H69" stroke="#67857b" stroke-width="5"/>',
  shears:
    '<path d="M31 68L79 12M67 68L22 12" stroke="#668581" stroke-width="7"/><circle cx="50" cy="44" r="7" fill="#cfb56f"/><ellipse cx="26" cy="79" rx="14" ry="17" fill="none" stroke="#935239" stroke-width="7"/><ellipse cx="72" cy="79" rx="14" ry="17" fill="none" stroke="#935239" stroke-width="7"/>',
};
export function motif(name, x, y, size) {
  return `<g transform="translate(${x} ${y}) scale(${size / 100})">${motifs[name] || motifs.cards}</g>`;
}
export function sceneSvg(objects, tint = '#f4e9cd') {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 180"><rect width="300" height="180" rx="8" fill="${tint}"/><path d="M0 133Q85 102 168 128T300 111V180H0Z" fill="#d1d6ae"/><ellipse cx="153" cy="157" rx="113" ry="12" fill="#8b7543" opacity=".12"/>${objects}</svg>`;
}
export function svgImage(svg, cls = '', label = '') {
  const image = document.createElement('img');
  image.className = cls;
  image.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg);
  image.alt = label;
  image.draggable = false;
  return image;
}
