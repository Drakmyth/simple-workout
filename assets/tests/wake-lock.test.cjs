const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const vm = require("node:vm");

const script = readFileSync("assets/static/app.js", "utf8");
const flush = () => new Promise(setImmediate);

function setup({ enabled = true, supported = true, denied = false, pending = false } = {}) {
  const document = new EventTarget();
  document.visibilityState = "visible";
  document.querySelectorAll = () => [];
  document.querySelector = () => ({ hasAttribute: () => enabled, querySelector: () => null });
  const window = new EventTarget();
  const locks = [];
  let requests = 0;
  let resolve;
  const navigator = supported ? { wakeLock: { request: async (type) => {
    assert.equal(type, "screen");
    requests++;
    if (denied) throw new Error("Denied");
    if (pending) await new Promise((done) => { resolve = done; });
    const lock = new EventTarget();
    lock.released = false;
    lock.release = async () => {
      lock.released = true;
      lock.dispatchEvent(new Event("release"));
    };
    locks.push(lock);
    return lock;
  } } } : {};
  vm.runInNewContext(script, { document, window, navigator });
  return { document, window, locks, requests: () => requests, resolve: () => resolve() };
}

test("requests on load, reacquires on foreground, and releases on departure", async () => {
  const state = setup();
  await flush();
  assert.equal(state.requests(), 1);
  state.document.dispatchEvent(new Event("visibilitychange"));
  assert.equal(state.requests(), 1);
  state.document.visibilityState = "hidden";
  await state.locks[0].release(); // Browsers release wake locks when hidden.
  state.document.dispatchEvent(new Event("visibilitychange"));
  assert.equal(state.requests(), 1);
  state.document.visibilityState = "visible";
  state.document.dispatchEvent(new Event("visibilitychange"));
  await flush();
  assert.equal(state.requests(), 2);
  state.window.dispatchEvent(new Event("pagehide"));
  assert.equal(state.locks[1].released, true);
  state.window.dispatchEvent(new Event("pageshow"));
  await flush();
  assert.equal(state.requests(), 3);
});

test("unsupported, disabled, and denied requests do not interrupt playback", async () => {
  assert.equal(setup({ enabled: false }).requests(), 0);
  assert.equal(setup({ supported: false }).requests(), 0);
  const denied = setup({ denied: true });
  await flush();
  assert.equal(denied.requests(), 1);
  assert.equal(denied.locks.length, 0);
});

test("releases a pending request that resolves after departure", async () => {
  const state = setup({ pending: true });
  state.window.dispatchEvent(new Event("pagehide"));
  state.resolve();
  await flush();
  assert.equal(state.locks[0].released, true);
});
