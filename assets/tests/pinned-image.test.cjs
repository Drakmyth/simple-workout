const { test } = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const vm = require("node:vm");
const script = readFileSync("assets/static/app.js", "utf8");

function setup({ enabled = true, supported = true, guidanceHeight = 200, fontSize = 16 } = {}) {
  const classes = new Set();
  const properties = {};
  let update;
  const header = {};
  const guidance = { clientHeight: guidanceHeight };
  const action = { getBoundingClientRect: () => ({ height: 80 }) };
  const playback = {
    hasAttribute: (name) => enabled && name === "data-keep-image-visible",
    querySelector: (selector) => ({ ".playback-set-header": header, ".playback-guidance": guidance })[selector],
    getBoundingClientRect: () => ({ top: 0 }),
    style: { setProperty: (name, value) => { properties[name] = value; } },
    classList: { add: (name) => classes.add(name), remove: (name) => classes.delete(name) },
  };
  const document = {
    documentElement: { clientHeight: 800 },
    querySelectorAll: () => [],
    querySelector: (selector) => selector === "[data-playback]" ? playback : action,
  };
  const window = { scrollY: 0, addEventListener: () => {} };
  if (supported) window.ResizeObserver = class {
    constructor(callback) { update = callback; }
    observe() {}
  };
  vm.runInNewContext(script, {
    document, window, navigator: {},
    getComputedStyle: () => ({ paddingTop: "20px", fontSize: `${fontSize}px` }),
  });
  return { classes, properties, guidance, update: () => update() };
}

test("pins when guidance has space and reserves space for the Next action", () => {
  const state = setup();
  assert.equal(state.classes.has("pinned-image"), true);
  assert.equal(state.properties["--set-height"], "684px");
});

test("falls back for short screens and large text, and restores when space returns", () => {
  const state = setup({ guidanceHeight: 100 });
  assert.equal(state.classes.has("pinned-image"), false);
  state.guidance.clientHeight = 200;
  state.update();
  assert.equal(state.classes.has("pinned-image"), true);
  assert.equal(setup({ fontSize: 24 }).classes.has("pinned-image"), false);
});

test("disabled or unsupported enhancement preserves normal scrolling", () => {
  assert.equal(setup({ enabled: false }).classes.size, 0);
  assert.equal(setup({ supported: false }).classes.size, 0);
});
