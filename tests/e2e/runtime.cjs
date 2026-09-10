const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');

const ROOT = path.resolve(__dirname, '../..');
const GO = process.env.GO_BIN || 'go';
const temp = (prefix) => fs.mkdtempSync(path.join(os.tmpdir(), prefix));
const executable = (name) => name + (process.platform === 'win32' ? '.exe' : '');
function sourceHashes() {
  const files = ['go.mod'];
  function walk(dir) {
    for (const entry of fs.readdirSync(path.join(ROOT, dir), { withFileTypes: true })) {
      const relative = path.join(dir, entry.name);
      if (entry.isDirectory()) walk(relative); else files.push(relative);
    }
  }
  for (const dir of ['cmd', 'internal', 'web']) walk(dir);
  return Object.fromEntries(files.sort().map((file) => [file, crypto.createHash('sha256').update(fs.readFileSync(path.join(ROOT, file))).digest('hex')]));
}
module.exports = { ROOT, GO, temp, executable, sourceHashes };
