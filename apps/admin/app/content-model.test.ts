import { describe, expect, it } from "vitest";
import {
  promptErrors,
  tournamentKey,
  validCohort,
  validCompetition,
  validPack,
  validPackResult,
  validPromptResult,
  validReviewResult,
  type PromptInput,
} from "./content-model";
const pack = {
  packId: "pack-1",
  market: "gh_tw" as const,
  terminologyRef: "terms:v1",
  features: { sow: true },
  status: "draft" as const,
  version: 1,
  createdAt: "2026-08-22T12:00:00Z",
};
const prompt: PromptInput = {
  id: "ebe.1",
  version: 1,
  language: "tw",
  cue: "Cue",
  acceptedAnswers: ["answer"],
  source: {
    kind: "book",
    citation: "Citation",
    locator: "https://example.com/source",
  },
};
const cohort = {
  id: "cohort-1",
  capacity: 4 as const,
  enrolled: 4,
  joined: false,
  status: "locked" as const,
  revision: 1,
};
const competition = {
  id: "comp-1",
  revision: 2,
  status: "active" as const,
  ladder: [{ label: "Player A", played: 1, wins: 1, you: false }],
  matches: [
    {
      id: "match-1",
      round: 1,
      slot: 1,
      firstLabel: "A",
      secondLabel: "B",
      resultRecorded: false,
      youArePlaying: false,
    },
  ],
  reviews: [
    {
      id: "review-1",
      matchId: "match-1",
      status: "open" as const,
      decision: "none" as const,
      openedAt: "2026-08-22T12:00:00Z",
      yours: false,
    },
  ],
};
describe("content and programme contracts", () => {
  it("validates exact market packs and transition identity", () => {
    expect(validPack(pack)).toBe(true);
    expect(validPack({ ...pack, version: 0 })).toBe(false);
    expect(
      validPackResult(
        { ...pack, status: "published", version: 2 },
        { action: "publish", pack },
      ),
    ).toBe(true);
    expect(
      validPackResult(pack, {
        action: "draft",
        market: "gh_tw",
        terminologyRef: "terms:v1",
        features: { sow: true },
      }),
    ).toBe(true);
    expect(
      validPackResult(
        { ...pack, packId: "wrong", status: "published" },
        { action: "publish", pack },
      ),
    ).toBe(false);
  });
  it("validates prompt fields and response identity without answer projection", () => {
    expect(promptErrors(prompt)).toEqual({});
    expect(
      promptErrors({ ...prompt, acceptedAnswers: ["same", "SAME"] }),
    ).toHaveProperty("acceptedAnswers");
    expect(promptErrors({ ...prompt, id: "-bad" })).toHaveProperty("id");
    expect(
      promptErrors({
        ...prompt,
        source: {
          ...prompt.source,
          kind: "web" as PromptInput["source"]["kind"],
        },
      }),
    ).toHaveProperty("sourceKind");
    expect(
      validPromptResult(
        {
          id: prompt.id,
          version: 1,
          language: "tw",
          cue: "Cue",
          sourceKind: "book",
          sourceCitation: "Citation",
          sourceLocator: "https://example.com/source",
        },
        prompt,
      ),
    ).toBe(true);
    expect(validPromptResult({ id: "wrong" }, prompt)).toBe(false);
  });
  it("deep-validates cohort and competition projections", () => {
    expect(validCohort(cohort)).toBe(true);
    expect(validCohort({ ...cohort, enrolled: 5 })).toBe(false);
    expect(validCohort({ ...cohort, enrolled: 3 })).toBe(false);
    expect(validCohort({ ...cohort, status: "started" })).toBe(false);
    expect(
      validCohort({ ...cohort, status: "started", competitionId: "comp-1" }),
    ).toBe(true);
    expect(validCompetition(competition)).toBe(true);
    expect(
      validCompetition({
        ...competition,
        matches: [{ ...competition.matches[0], slot: 0 }],
      }),
    ).toBe(true);
    expect(
      validCompetition({
        ...competition,
        reviews: [{ ...competition.reviews[0], openedAt: "bad" }],
      }),
    ).toBe(false);
    const pending = {
      kind: "review" as const,
      cohortId: "cohort-1",
      competitionId: "comp-1",
      reviewId: "review-1",
      decision: "no_action" as const,
      expectedRevision: 2,
      appeal: false,
      key: "review-key",
    };
    expect(
      validReviewResult(
        {
          ...competition,
          revision: 3,
          reviews: [
            {
              ...competition.reviews[0],
              status: "resolved",
              decision: "no_action",
              resolvedAt: "2026-08-22T13:00:00Z",
            },
          ],
        },
        pending,
      ),
    ).toBe(true);
    expect(
      validReviewResult(
        {
          ...competition,
          revision: 3,
          reviews: [
            {
              ...competition.reviews[0],
              status: "open",
              decision: "no_action",
            },
          ],
        },
        pending,
      ),
    ).toBe(false);
  });
  it("changes stable command terms when exact revision or decision changes", () => {
    const base = {
      kind: "review" as const,
      cohortId: "c",
      competitionId: "x",
      reviewId: "r",
      decision: "no_action" as const,
      expectedRevision: 1,
      appeal: false,
    };
    expect(tournamentKey(base)).toBe(tournamentKey({ ...base }));
    expect(tournamentKey(base)).not.toBe(
      tournamentKey({ ...base, expectedRevision: 2 }),
    );
    expect(tournamentKey(base)).not.toBe(
      tournamentKey({ ...base, decision: "rules_action" }),
    );
  });
});
