import { expect, test } from "vitest";

import { initialInteractionState, interactionReducer } from "./interactions";

test("Hold is release-free until its deliberate threshold completes", () => {
  const holding = interactionReducer(initialInteractionState, {
    type: "hold-start",
  });
  expect(holding.hold).toBe("holding");
  expect(interactionReducer(holding, { type: "hold-release" }).hold).toBe(
    "ready",
  );
  expect(interactionReducer(holding, { type: "hold-complete" }).hold).toBe(
    "paused",
  );
});

test("Sow cannot send without an explicit review and confirmation", () => {
  const staged = interactionReducer(initialInteractionState, {
    type: "sow-stage",
  });
  expect(staged.sow).toBe("staged");
  expect(interactionReducer(staged, { type: "sow-confirm" }).sow).toBe(
    "staged",
  );
  const confirming = interactionReducer(staged, { type: "sow-review" });
  expect(confirming.sow).toBe("confirming");
  expect(interactionReducer(confirming, { type: "sow-confirm" }).sow).toBe(
    "sent",
  );
});

test("Stone release cancels while completion places it", () => {
  const settling = interactionReducer(initialInteractionState, {
    type: "stone-start",
  });
  expect(interactionReducer(settling, { type: "stone-release" }).stone).toBe(
    "ready",
  );
  expect(interactionReducer(settling, { type: "stone-complete" }).stone).toBe(
    "placed",
  );
});

test("Gather always has a discrete non-gesture alternative", () => {
  const gathered = interactionReducer(initialInteractionState, {
    type: "gather-set",
    distance: "near",
  });
  expect(gathered.gather).toBe("near");
  expect(
    interactionReducer(gathered, {
      type: "gather-set",
      distance: "spacious",
    }).gather,
  ).toBe("spacious");
});
