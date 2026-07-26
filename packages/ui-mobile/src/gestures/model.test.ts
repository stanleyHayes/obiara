import assert from "node:assert/strict";
import test from "node:test";

import {
  HOLD_DURATION_MS,
  SOW_RELEASE_DISTANCE,
  initialGatherState,
  initialHoldState,
  initialSowState,
  initialStoneState,
  reduceGather,
  reduceHold,
  reduceSow,
  reduceStone,
} from "./model.ts";

test("Hold pauses safely on release and never completes early", () => {
  const holding = reduceHold(initialHoldState, { type: "start" });
  const partial = reduceHold(holding, { type: "advance", milliseconds: 400 });
  const paused = reduceHold(partial, { type: "release" });
  assert.deepEqual(paused, { status: "paused", elapsedMs: 400 });
  assert.deepEqual(reduceHold(paused, { type: "cancel" }), initialHoldState);
  assert.equal(
    reduceHold(reduceHold(paused, { type: "start" }), {
      type: "advance",
      milliseconds: HOLD_DURATION_MS,
    }).status,
    "completed",
  );
});

test("Sow only releases a staged recording after deliberate intent", () => {
  const staged = reduceSow(initialSowState, { type: "finish-recording" });
  const shortDrag = reduceSow(staged, {
    type: "drag",
    distance: SOW_RELEASE_DISTANCE - 1,
  });
  assert.equal(reduceSow(shortDrag, { type: "release" }).status, "staged");
  const ready = reduceSow(staged, {
    type: "drag",
    distance: SOW_RELEASE_DISTANCE,
  });
  assert.equal(reduceSow(ready, { type: "release" }).status, "sown");
  assert.equal(reduceSow(staged, { type: "confirm" }).status, "sown");
});

test("Stone supports equivalent slow drag, long press, and button paths", () => {
  const pressing = reduceStone(initialStoneState, { type: "start" });
  assert.equal(
    reduceStone(pressing, { type: "advance", milliseconds: 1_000 }).status,
    "settled",
  );
  assert.equal(
    reduceStone(initialStoneState, { type: "drag", distance: 120 }).status,
    "settled",
  );
  assert.equal(
    reduceStone(initialStoneState, { type: "confirm" }).status,
    "settled",
  );
});

test("Gather clamps and rounds accessible adjustments deterministically", () => {
  assert.deepEqual(
    reduceGather(initialGatherState, { type: "adjust", delta: 2 }),
    {
      amount: 1,
      completed: false,
    },
  );
  const lowered = reduceGather(initialGatherState, {
    type: "adjust",
    delta: -0.2,
  });
  assert.equal(lowered.amount, 0.3);
  assert.equal(reduceGather(lowered, { type: "confirm" }).completed, true);
});
