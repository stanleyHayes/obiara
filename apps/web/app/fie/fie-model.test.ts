import { describe, expect, it } from "vitest";

import {
  connectionMessage,
  fieHomeReducer,
  initialFieHomeState,
} from "./fie-model";

describe("Fie home connection state", () => {
  it("retains queued work while constrained", () => {
    const queued = fieHomeReducer(initialFieHomeState, {
      type: "queue-action",
    });
    expect(queued.queuedActions).toBe(2);
    expect(connectionMessage(queued).detail).toContain("2 safe actions");
  });

  it("clears queued work only after an online transition", () => {
    const online = fieHomeReducer(initialFieHomeState, {
      type: "connection",
      mode: "online",
    });
    expect(online.queuedActions).toBe(0);
    expect(connectionMessage(online).label).toBe("Connected");
  });

  it("announces offline status assertively", () => {
    const offline = fieHomeReducer(initialFieHomeState, {
      type: "connection",
      mode: "offline",
    });
    expect(connectionMessage(offline).live).toBe("assertive");
  });
});
