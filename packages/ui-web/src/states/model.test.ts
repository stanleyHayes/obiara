import assert from "node:assert/strict";
import test from "node:test";

import { stateSemantics, type ObiaraStateKind } from "./model.ts";

test("announces failures assertively and all other states politely", () => {
  assert.equal(stateSemantics("error").live, "assertive");
  assert.equal(stateSemantics("permission-denied").role, "alert");
  for (const kind of [
    "empty",
    "offline",
    "queued",
    "low-bandwidth",
  ] satisfies ObiaraStateKind[]) {
    assert.equal(stateSemantics(kind).live, "polite");
    assert.equal(stateSemantics(kind).role, "status");
  }
});

test("marks only loading as busy and suppresses its action", () => {
  const loading = stateSemantics("loading");
  assert.equal(loading.busy, true);
  assert.equal(loading.actionAllowed, false);
  assert.equal(stateSemantics("error").actionAllowed, true);
});
