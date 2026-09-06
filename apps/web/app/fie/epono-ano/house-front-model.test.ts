import { describe, expect, it } from "vitest";

import {
  closesIn,
  houseFrontSummary,
  soonestFirst,
  type RestingPod,
} from "./house-front-model";

const now = new Date("2026-09-06T12:00:00.000Z");

function pod(podId: string, closesAt: string, opened = false): RestingPod {
  return { podId, closesAt, opened };
}

describe("closesIn", () => {
  it("says the time left in words rather than a timestamp", () => {
    expect(closesIn("2026-09-06T12:30:00.000Z", now)).toBe(
      "Closes within the hour",
    );
    expect(closesIn("2026-09-06T15:00:00.000Z", now)).toBe("Closes in 3 hours");
    expect(closesIn("2026-09-06T13:00:00.000Z", now)).toBe("Closes in 1 hour");
    expect(closesIn("2026-09-09T12:00:00.000Z", now)).toBe("Closes in 3 days");
    expect(closesIn("2026-09-07T12:00:00.000Z", now)).toBe("Closes in 1 day");
  });

  it("says a pod past its time is closed rather than counting backwards", () => {
    expect(closesIn("2026-09-05T12:00:00.000Z", now)).toBe("Closed");
  });

  it("says nothing at all when the date makes no sense", () => {
    // Better silence than a confident wrong number under somebody's pod.
    expect(closesIn("not a date", now)).toBe("");
  });
});

describe("houseFrontSummary", () => {
  it("says an empty house front plainly", () => {
    // Nothing resting is the ordinary state of most days, not a failure, and
    // it should not read like one.
    expect(houseFrontSummary([])).toBe("Nothing is resting here yet.");
  });

  it("counts what is still unopened, because that is the part that matters", () => {
    expect(houseFrontSummary([pod("a", "2026-09-07T12:00:00.000Z")])).toBe(
      "One pod, unopened.",
    );
    expect(
      houseFrontSummary([
        pod("a", "2026-09-07T12:00:00.000Z"),
        pod("b", "2026-09-08T12:00:00.000Z"),
      ]),
    ).toBe("2 pods, unopened.");
    expect(
      houseFrontSummary([
        pod("a", "2026-09-07T12:00:00.000Z", true),
        pod("b", "2026-09-08T12:00:00.000Z"),
      ]),
    ).toBe("2 pods · 1 still unopened.");
  });

  it("does not nag when everything has been heard", () => {
    expect(
      houseFrontSummary([pod("a", "2026-09-07T12:00:00.000Z", true)]),
    ).toBe("One pod, already heard.");
    expect(
      houseFrontSummary([
        pod("a", "2026-09-07T12:00:00.000Z", true),
        pod("b", "2026-09-08T12:00:00.000Z", true),
      ]),
    ).toBe("2 pods, all heard.");
  });
});

describe("soonestFirst", () => {
  it("puts the pod closing first at the top", () => {
    const ordered = soonestFirst([
      pod("later", "2026-09-09T12:00:00.000Z"),
      pod("sooner", "2026-09-07T12:00:00.000Z"),
    ]);
    expect(ordered.map((entry) => entry.podId)).toEqual(["sooner", "later"]);
  });

  it("does not mutate what it was given", () => {
    const pods = [
      pod("later", "2026-09-09T12:00:00.000Z"),
      pod("sooner", "2026-09-07T12:00:00.000Z"),
    ];
    soonestFirst(pods);
    expect(pods.map((entry) => entry.podId)).toEqual(["later", "sooner"]);
  });
});
