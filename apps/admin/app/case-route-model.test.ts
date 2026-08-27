import { describe, expect, it } from "vitest";
import {
  buildCasePath,
  queueNoticeText,
  sanitizeQueueReturn,
  terminalQueuePath,
} from "./case-route-model";

describe("case route context model", () => {
  it.each(["case%value", "case%2Fvalue", "case with spaces"])(
    "encodes edge id %s exactly once",
    (caseId) => {
      expect(
        decodeURIComponent(
          buildCasePath("verification", caseId).split("/").at(-1)!,
        ),
      ).toBe(caseId);
    },
  );

  it("preserves only return URLs inside the originating queue", () => {
    expect(sanitizeQueueReturn("verification", "/verification?q=manual")).toBe(
      "/verification?q=manual",
    );
    expect(sanitizeQueueReturn("verification", "/safety")).toBe(
      "/verification",
    );
    expect(sanitizeQueueReturn("care", "https://attacker.example")).toBe(
      "/care",
    );
  });

  it.each([
    "/verification-evil",
    "/verification/../safety",
    "/verification/%2e%2e/safety",
    "/verification%2Fsafety",
    "/verification\\@evil.example",
    "//evil.example/verification",
    "https://user:pass@admin.obiara.invalid/verification",
    "https://evil.example/verification",
  ])("rejects confused or external return target %s", (candidate) => {
    expect(sanitizeQueueReturn("verification", candidate)).toBe(
      "/verification",
    );
  });

  it("whitelists and safely serializes queue search context", () => {
    expect(
      sanitizeQueueReturn(
        "verification",
        "/verification?q=two+words+%25+%26&notice=evil&other=x",
      ),
    ).toBe("/verification?q=two+words+%25+%26");
  });

  it("merges only a trusted terminal notice code into sanitized context", () => {
    expect(
      terminalQueuePath(
        "verification",
        "decision-recorded",
        "/verification?q=manual&notice=evil",
      ),
    ).toBe("/verification?q=manual&notice=decision-recorded");
    expect(queueNoticeText("decision-recorded")).toBe(
      "The audited verification decision was recorded.",
    );
    expect(queueNoticeText("attacker text")).toBeNull();
  });
});
