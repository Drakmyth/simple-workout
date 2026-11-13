const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const vm = require("node:vm");
const script = readFileSync("assets/static/app.js", "utf8");

function setup({ seconds = 2, enabled = true, visible = true, denied = false } = {}) {
  let tick;
  let plays = 0;
  let stopped = false;
  const display = { dataset: { restSeconds: String(seconds) } };
  const audio = { play: () => {
    plays++;
    return denied ? Promise.reject(new Error("Blocked")) : Promise.resolve();
  } };
  const playback = {
    hasAttribute: () => false,
    querySelector: (selector) => selector === "[data-rest-timer]" ? display :
      selector === "[data-rest-sound]" && enabled ? audio : null,
  };
  const document = {
    visibilityState: visible ? "visible" : "hidden",
    querySelectorAll: () => [], querySelector: () => playback,
  };
  vm.runInNewContext(script, {
    document, navigator: {}, window: {},
    setInterval: (callback) => { tick = callback; return 1; },
    clearInterval: () => { stopped = true; },
  });
  return { tick: () => { if (!stopped) tick?.(); }, plays: () => plays, display };
}

test("plays once when a positive rest reaches zero, never before", () => {
  const state = setup();
  state.tick();
  assert.equal(state.plays(), 0);
  state.tick();
  assert.equal(state.plays(), 1);
  assert.equal(state.display.textContent, "0 seconds");
  state.tick();
  assert.equal(state.plays(), 1);
});

test("zero-length, disabled, and hidden rests are silent", () => {
  for (const options of [{ seconds: 0 }, { enabled: false }, { visible: false }]) {
    const state = setup(options);
    state.tick(); state.tick();
    assert.equal(state.plays(), 0);
  }
});

test("blocked audio leaves the countdown completed without an unhandled rejection", async () => {
  const state = setup({ seconds: 1, denied: true });
  state.tick();
  await new Promise(setImmediate);
  assert.equal(state.plays(), 1);
  assert.equal(state.display.textContent, "0 seconds");
});
