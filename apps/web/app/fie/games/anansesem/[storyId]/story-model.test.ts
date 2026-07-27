import { describe, expect, it } from "vitest";
import { canPublish, initialStoryState, storyReducer } from "./story-model";

describe("private Anansesɛm relay", () => {
  it("requires a bounded contribution and hands over the turn", () => {
    expect(storyReducer(initialStoryState, { type: "contribute" })).toEqual(
      initialStoryState,
    );
    const drafted = storyReducer(initialStoryState, {
      type: "draft",
      value: "The path answered with a drumbeat.",
    });
    const sent = storyReducer(drafted, { type: "contribute" });
    expect(sent.turn).toBe("ama");
    expect(sent.contributions).toBe(5);
  });

  it("keeps publish consent separate and invalidates it after new writing", () => {
    const consented = storyReducer(initialStoryState, {
      type: "toggle-publish-consent",
    });
    expect(canPublish(consented)).toBe(true);
    const drafted = storyReducer(consented, {
      type: "draft",
      value: "A new ending changes the shared work.",
    });
    expect(
      storyReducer(drafted, { type: "contribute" }).yourPublishConsent,
    ).toBe(false);
  });

  it("bounds drafts and blocks consecutive turns", () => {
    const drafted = storyReducer(initialStoryState, {
      type: "draft",
      value: "a".repeat(400),
    });
    expect(drafted.draft).toHaveLength(280);
    const sent = storyReducer(drafted, { type: "contribute" });
    expect(
      storyReducer(sent, { type: "draft", value: "another turn" }),
    ).toEqual(sent);
  });
});
