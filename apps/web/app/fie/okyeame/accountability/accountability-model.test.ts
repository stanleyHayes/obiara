import { describe, expect, it } from "vitest";

import {
  accountabilityProjectionContainsSensitiveData,
  capabilityCards,
  initialAppealState,
  submitAppeal,
} from "./accountability-model";

describe("AI accountability projection", () => {
  it("publishes a version, consent basis and bounded review state", () => {
    expect(
      capabilityCards.every(
        (card) =>
          card.version &&
          card.consentBasis &&
          card.evaluation &&
          card.redTeam &&
          /^\d{4}-\d{2}-\d{2}$/.test(card.lastReviewed),
      ),
    ).toBe(true);
  });

  it("keeps an unreleased ranker paused", () => {
    expect(
      capabilityCards.find((card) => card.id === "matching-ranker"),
    ).toMatchObject({
      status: "paused",
      version: "not-released",
    });
  });

  it("creates a human appeal reference only for a published capability", () => {
    expect(submitAppeal(initialAppealState, "unknown")).toEqual(
      initialAppealState,
    );
    expect(submitAppeal(initialAppealState, "okyeame-help")).toMatchObject({
      status: "submitted",
      capabilityId: "okyeame-help",
      reference: "appeal-okyeame-help",
    });
  });

  it("does not expose raw or sensitive system fields", () => {
    expect(accountabilityProjectionContainsSensitiveData()).toBe(false);
  });
});
