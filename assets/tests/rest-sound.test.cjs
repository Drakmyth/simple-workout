const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const vm = require("node:vm");
const script = readFileSync("assets/static/app.js", "utf8");

function setup({ seconds = 2, enabled = true, visible = true, denied = false, rest = true } = {}) {
  const classes = new Set();
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
    classList: { add: (name) => classes.add(name), remove: (name) => classes.delete(name) },
    querySelector: (selector) => selector === "[data-rest-timer]" ? (rest ? display : null) :
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
  return { tick: () => { if (!stopped) tick?.(); }, plays: () => plays, display, classes };
}

test("rest countdown cue starts immediately and clears at zero even without sound", () => {
  const state = setup({ enabled: false });
  assert.equal(state.classes.has("rest-counting"), true);
  state.tick();
  assert.equal(state.classes.has("rest-counting"), true);
  state.tick();
  assert.equal(state.classes.has("rest-counting"), false);
  assert.equal(state.display.textContent, "0 seconds");
});

test("zero-length rests and non-rest screens do not get the countdown cue", () => {
  assert.equal(setup({ seconds: 0 }).classes.has("rest-counting"), false);
  assert.equal(setup({ rest: false }).classes.has("rest-counting"), false);
});

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
