import { describe, expect, it } from "vitest";
import { ampeReducer, initialAmpeState } from "./ampe-model";

describe("private Ampe pulse", () => {
  it("requires ready, private choice and lock before reveal", () => {
    const choosing = ampeReducer(initialAmpeState, { type: "ready" });
    expect(ampeReducer(choosing, { type: "lock" })).toEqual(choosing);
    const chosen = ampeReducer(choosing, { type: "choose", gesture: "apart" });
    const locked = ampeReducer(chosen, { type: "lock" });
    expect(locked.stage).toBe("locked");
    expect(ampeReducer(locked, { type: "reveal" }).stage).toBe("revealed");
  });

  it("keeps a locked choice hidden and safe through reconnect", () => {
    const locked = { stage: "locked", choice: "together" } as const;
    const reconnecting = ampeReducer(locked, { type: "connection-lost" });
    expect(reconnecting.choice).toBe("together");
    expect(ampeReducer(reconnecting, { type: "reconnected" }).stage).toBe(
      "locked",
    );
  });
});
