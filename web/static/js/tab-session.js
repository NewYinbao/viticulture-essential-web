// Move the previous browser-wide login into the first tab that opens the new UI.
// Subsequent tabs start independently; only non-sensitive UI preferences stay shared.
export function migrateTabSession() {
  const tokenKey = 'vineyard-ee-token';
  const nameKey = 'vineyard-ee-name';
  const previousToken = localStorage.getItem(tokenKey);
  if (!sessionStorage.getItem(tokenKey) && previousToken) {
    sessionStorage.setItem(tokenKey, previousToken);
    sessionStorage.setItem(nameKey, localStorage.getItem(nameKey) || '');
  }
  localStorage.removeItem(tokenKey);
  localStorage.removeItem(nameKey);
}
