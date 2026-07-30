import { describe, expect, it } from "vitest";
import { storyPermissions } from "./story-model";

describe("Anansesɛm client permission boundary", () => {
  it("never grants a transition from fabricated local state", () => {
    expect(
      storyPermissions({
        passageCount: 0,
        yourTurn: true,
        yourGrant: false,
        bothGranted: false,
      }),
    ).toEqual({ canAdd: true, canGrant: false, canPublish: false });
  });

  it("respects retained turn, capacity, and current-draft grants", () => {
    expect(
      storyPermissions({
        passageCount: 40,
        yourTurn: true,
        yourGrant: true,
        bothGranted: true,
      }),
    ).toEqual({ canAdd: false, canGrant: false, canPublish: true });
  });
});
