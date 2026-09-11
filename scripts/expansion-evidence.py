"""Read downloaded publisher pages without importing third-party dependencies."""
from html import unescape
from html.parser import HTMLParser
from pathlib import Path
import re

class Text(HTMLParser):
    def __init__(self):
        super().__init__()
        self.parts = []
        self.skip = 0
    def handle_starttag(self, tag, attrs):
        if tag in ('script', 'style'): self.skip += 1
        if tag in ('p', 'div', 'li', 'h1', 'h2', 'h3', 'tr'): self.parts.append('\n')
    def handle_endtag(self, tag):
        if tag in ('script', 'style'): self.skip = max(0, self.skip - 1)
    def handle_data(self, data):
        if not self.skip: self.parts.append(data)

for path in Path('artifacts/expansions/evidence').glob('*.html'):
    src = path.read_text(encoding='utf-8')
    parser = Text()
    parser.feed(src)
    out = '\n'.join(' '.join(s.split()) for s in ''.join(parser.parts).splitlines() if s.strip())
    path.with_suffix('.txt').write_text(out, encoding='utf-8')
    links = sorted(set(unescape(x) for x in re.findall(r'href=[\"\']([^\"\']+)', src) if any(k in x for k in ('dropbox', '.pdf', 'uploads/201', 'uploads/202', 'spreadsheets'))))
    print(path.name, len(src), 'bytes')
    print('\n'.join(links))
