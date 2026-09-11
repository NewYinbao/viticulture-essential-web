// One bounded request at a time; close aborts both pending fetch and future retries.
export function pollState(token, onState, onError, interval = 1500) {
  let stopped = false,
    timer,
    controller;
  async function tick() {
    controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 10000);
    try {
      const response = await fetch('/api/state', {
        headers: { Authorization: 'Bearer ' + token },
        cache: 'no-store',
        signal: controller.signal,
      });
      if (!response.ok) {
        const error = new Error('State request failed: ' + response.status);
        error.status = response.status;
        throw error;
      }
      const state = await response.json();
      if (!stopped) onState(state);
    } catch (error) {
      if (!stopped) onError(error);
    } finally {
      clearTimeout(timeout);
      if (!stopped) timer = setTimeout(tick, interval);
    }
  }
  tick();
  return {
    close() {
      stopped = true;
      clearTimeout(timer);
      controller?.abort();
    },
  };
}
