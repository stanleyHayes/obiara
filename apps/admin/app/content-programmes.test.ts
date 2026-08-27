import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";
const root = path.resolve(import.meta.dirname),
  read = (p: string) => fs.readFileSync(path.join(root, p), "utf8");
describe("content programme routes", () => {
  it("keeps governance list vertical and response-bound", () => {
    const s = read("(ops)/governance/governance-desk.tsx");
    expect(s).toContain("b.packs.every(validPack)");
    expect(s).toContain("validPackResult(b, snapshot)");
    expect(s).toContain('pending?.action === "draft"');
    expect(s).toContain("<SegmentedOtpInput");
    expect(s).not.toContain("CircularProgress");
    expect(s).not.toContain("<Card");
  });
  it("keeps prompt approval immutable and answer-private", () => {
    const s = read("(ops)/game-content/game-content-desk.tsx");
    expect(s).toContain("validPromptResult(payload, value)");
    expect(s).toContain("Accepted answer count:");
    expect(s).toContain("Source locator:");
    expect(s).not.toContain("Approving…");
    expect(s).toContain("formRef.current = e.currentTarget");
  });
  it("uses dedicated exact tournament routes and stable keys", () => {
    const s = read("(ops)/tournaments/tournament-desk.tsx"),
      detail = read("(ops)/tournaments/[cohortId]/page.tsx"),
      competition = read(
        "(ops)/tournaments/[cohortId]/competitions/[competitionId]/page.tsx",
      );
    expect(s).toContain("validCompetition");
    expect(s).toContain("keyFor(terms)");
    expect(s).toContain("aria-valuetext");
    expect(detail).not.toContain("decodeURIComponent");
    expect(detail).toContain("key={cohortId}");
    expect(competition).not.toContain("decodeURIComponent");
    expect(competition).toContain("key={`${cohortId}");
    expect(s).not.toContain("<Card");
  });
});
