// Fixture-only credentials. Call only on a synthetic test snapshot, never a
// user's save. Game-state fixtures exercise normal authenticated HTTP routes.
const { randomBytes, pbkdf2Sync } = require('node:crypto');
const salt = randomBytes(16);
const credential = {
  Salt: salt.toString('hex'),
  Hash: pbkdf2Sync('test-password-123', salt, 600000, 32, 'sha256').toString('hex'),
};
function authorizeFixtureStore(store) {
  store.Passwords ||= {};
  for (const room of Object.values(store.Rooms)) {
    for (const player of room.players) store.Passwords[player.id] ||= { ...credential };
  }
  return store;
}
module.exports = { authorizeFixtureStore };
