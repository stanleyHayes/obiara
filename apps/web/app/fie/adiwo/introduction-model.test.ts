import { describe, expect, it } from "vitest";

import {
  askSummary,
  canAskThrough,
  initialAsk,
  type AskState,
} from "./introduction-model";

describe("asking to be introduced through a circle", () => {
  it("offers the ask only to settled members", () => {
    // The API refuses anyone else, so showing the button would promise
    // something the next request cannot keep.
    for (const membership of ["member", "host", "owner"]) {
      expect(canAskThrough(membership)).toBe(true);
    }
    for (const membership of ["none", "requested", "expelled", "left"]) {
      expect(canAskThrough(membership)).toBe(false);
    }
  });

  it("says nobody plainly rather than dressing it up", () => {
    const empty: AskState = {
      ...initialAsk,
      stage: "asked",
      found: 0,
      requestId: "src_1",
    };
    expect(askSummary(empty)).toContain("Nobody in this circle");
  });

  it("counts people, in words that fit the number", () => {
    const one: AskState = {
      ...initialAsk,
      stage: "asked",
      found: 1,
      requestId: "src_1",
    };
    const several: AskState = { ...one, found: 4 };
    expect(askSummary(one)).toBe("One person from this circle could meet you.");
    expect(askSummary(several)).toBe(
      "4 people from this circle could meet you.",
    );
  });

  it("says nothing before an ask has landed", () => {
    expect(askSummary(initialAsk)).toBe("");
    expect(askSummary({ ...initialAsk, stage: "asking" })).toBe("");
    expect(askSummary({ ...initialAsk, stage: "failed", error: "nope" })).toBe(
      "",
    );
  });
});
