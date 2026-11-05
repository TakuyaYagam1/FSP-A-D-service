const test = require("node:test");
const assert = require("node:assert/strict");
const { buildRoutePassPNG, routePassViewModel } = require("./route-pass.js");

const pass = {
  route_slot_id: "123e4567-e89b-12d3-a456-426614174000",
  case_id: "123e4567-e89b-12d3-a456-426614174001",
  kind: "import",
  origin: "Орел",
  destination: "Тула",
  carrier_username: "carrier",
  inspector_username: "inspector",
  issued_at: "2026-11-05T17:00:00Z",
  slot_code: "private-code",
};

test("step 4 stays locked until ready, then owner alone may issue", () => {
  const waiting = routePassViewModel({ ready: false, issued: false }, true);
  assert.equal(waiting.enabled, false);
  assert.equal(waiting.canIssue, false);
  assert.equal(waiting.canDownload, false);

  const ownerReady = routePassViewModel({ ready: true, issued: false }, true);
  assert.equal(ownerReady.enabled, true);
  assert.equal(ownerReady.canIssue, true);
  assert.equal(ownerReady.canDownload, false);

  const participantReady = routePassViewModel({ ready: true, issued: false }, false);
  assert.equal(participantReady.enabled, true);
  assert.equal(participantReady.canIssue, false);
  assert.match(participantReady.status, /владельц/);
});

test("issued pass stays available when approvals expire or the page reloads", () => {
  const expired = { ready: false, issued: true, pass };
  for (const owner of [true, false]) {
    const view = routePassViewModel(expired, owner);
    assert.equal(view.enabled, true);
    assert.equal(view.canIssue, false);
    assert.equal(view.canDownload, true);
    assert.match(view.status, /выпущен/);
  }
});

test("PNG contains only server issuance metadata, never the slot code", () => {
  const text = [];
  const context = {
    fillRect() {}, beginPath() {}, moveTo() {}, lineTo() {}, stroke() {},
    measureText(value) { return { width: String(value).length * 18 }; },
    fillText(value) { text.push(value); },
  };
  const previousDocument = global.document;
  global.document = { createElement(tag) {
    assert.equal(tag, "canvas");
    return {
      getContext(kind) { assert.equal(kind, "2d"); return context; },
      toDataURL(kind) { assert.equal(kind, "image/png"); return "data:image/png;base64,iVBORw0KGgo="; },
    };
  } };
  try {
    const result = buildRoutePassPNG(pass);
    assert.match(result, /^data:image\/png;base64,/);
    const drawn = text.join(" | ");
    for (const label of ["МАРШРУТ ГОТОВ", "Орел", "Тула", "carrier", "inspector", pass.case_id, pass.route_slot_id]) {
      assert.ok(drawn.includes(label), `missing ${label}: ${drawn}`);
    }
    assert.ok(!drawn.includes("private-code"), drawn);
    assert.ok(!drawn.includes("ПОДПИС"), drawn);
    assert.ok(!drawn.includes("НЕ СОГЛАСОВАНО"), drawn);
  } finally {
    global.document = previousDocument;
  }
});

test("invalid or unissued pass cannot become a PNG", () => {
  assert.throws(() => buildRoutePassPNG({ ...pass, issued_at: "not-a-date" }), /выпуска/);
  assert.throws(() => buildRoutePassPNG({ ...pass, route_slot_id: "" }), /пропуска/);
});
