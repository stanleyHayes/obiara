import assert from "node:assert/strict";
import test from "node:test";

import { mobileStateSemantics, type MobileStateKind } from "./model.ts";

test("uses assertive announcements only for blocked outcomes", () => {
  assert.equal(mobileStateSemantics("error").liveRegion, "assertive");
  assert.equal(
    mobileStateSemantics("permission-denied").liveRegion,
    "assertive",
  );
  for (const kind of [
    "empty",
    "offline",
    "queued",
    "low-bandwidth",
  ] satisfies MobileStateKind[]) {
    assert.equal(mobileStateSemantics(kind).liveRegion, "polite");
  }
});

test("marks only loading busy and keeps recovery actions deliberate", () => {
  assert.deepEqual(mobileStateSemantics("loading"), {
    liveRegion: "polite",
    busy: true,
    actionAllowed: false,
  });
  assert.equal(mobileStateSemantics("offline").actionAllowed, true);
});
