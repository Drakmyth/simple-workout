const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const vm = require("node:vm");

const script = readFileSync("assets/static/app.js", "utf8");
const flush = () => new Promise(setImmediate);

function setup({ supported = true, denied = false } = {}) {
  const button = new EventTarget();
  button.hidden = true;
  button.textContent = "Copy summary";
  const status = { textContent: "" };
  const text = "Workout: Test\n\nCurl\n2 sets × 10 reps · Bodyweight\nNotes: Tough\nbut good";
  const nodes = {
    "[data-copy-summary]": button,
    "[data-summary-text]": { textContent: text },
    "[data-copy-status]": status,
  };
  const playback = { hasAttribute: () => false, querySelector: (selector) => nodes[selector] };
  const document = { querySelectorAll: () => [], querySelector: () => playback };
  let copied;
  let reset;
  const navigator = supported ? { clipboard: { writeText: async (value) => {
    if (denied) throw new Error("Denied");
    copied = value;
  } } } : {};
  vm.runInNewContext(script, {
    document, navigator, window: {},
    clearTimeout: () => {}, setTimeout: (callback) => { reset = callback; },
  });
  return { button, status, text, copied: () => copied, reset: () => reset() };
}

test("copies exact plain text and briefly confirms success", async () => {
  const state = setup();
  assert.equal(state.button.hidden, false);
  state.button.dispatchEvent(new Event("click"));
  await flush();
  assert.equal(state.copied(), state.text);
  assert.equal(state.button.textContent, "Copied");
  assert.equal(state.status.textContent, "Summary copied.");
  assert.equal(state.button.disabled, false);
  state.reset();
  assert.equal(state.button.textContent, "Copy summary");
});

test("clipboard rejection shows an error and allows retry", async () => {
  const state = setup({ denied: true });
  state.button.dispatchEvent(new Event("click"));
  await flush();
  assert.match(state.status.textContent, /Could not copy summary/);
  assert.equal(state.button.disabled, false);
});

test("unsupported clipboard leaves the progressive control hidden", () => {
  assert.equal(setup({ supported: false }).button.hidden, true);
});
