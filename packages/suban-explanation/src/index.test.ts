import { describe, expect, it } from "vitest";

import { initialSubanState, isPrivacySafe, subanReducer } from "./index";

describe("suban explanation", () => {
  it("keeps every contribution visible and explains decay and suppression", () => {
    expect(initialSubanState.events).toHaveLength(3);
    expect(
      initialSubanState.events.filter((event) => event.decays),
    ).toHaveLength(2);
    expect(initialSubanState.events.at(-1)?.effect).toContain("does not erase");
    expect(initialSubanState.markState).toBe("suppressed");
  });

  it("creates a pending opaque appeal without changing mark history", () => {
    const explained = subanReducer(initialSubanState, {
      type: "appeal-reason",
      value: "This finding is missing relevant context.",
    });
    const appealed = subanReducer(explained, { type: "submit-appeal" });
    expect(appealed.appealRef).toBe("appeal•••5T6");
    expect(appealed.appealState).toBe("pending");
    expect(appealed.events).toEqual(initialSubanState.events);
    expect(appealed.markState).toBe("suppressed");
  });

  it("rejects short appeals and excludes restricted data", () => {
    const short = subanReducer(
      subanReducer(initialSubanState, {
        type: "appeal-reason",
        value: "wrong",
      }),
      { type: "submit-appeal" },
    );
    expect(short.appealState).toBe("none");
    expect(isPrivacySafe(short)).toBe(true);
  });
});
