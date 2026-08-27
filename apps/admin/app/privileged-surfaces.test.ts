import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";

const root = path.resolve(import.meta.dirname);
const source = (relative: string) =>
  fs.readFileSync(path.join(root, relative), "utf8");

// Source assertions match intent, not line breaks: prettier is free to reflow
// a long condition across lines and that must not read as a regression.
const flat = (source: string) => source.replace(/\s+/g, " ");

describe("privileged admin surfaces", () => {
  it("keeps account reads abortable and sign-out failure truthful", () => {
    const account = source("(ops)/account/account-settings.tsx");
    expect(account).toContain("controllerRef.current?.abort()");
    expect(account).toMatch(/if \(!response\.ok\)\s*throw new Error/);
    expect(account).toContain("interface is not translated yet");
    expect(account).toContain('disabled={item !== "English"}');
    expect(account).toContain("coming soon");
    expect(account).toContain('() => "English" as const');
    expect(account).not.toContain("CircularProgress");
  });

  it("retains one control command id across ambiguous retries and segments MFA", () => {
    const controls = source("(ops)/controls/controls-desk.tsx");
    expect(controls).toContain("commandId: commandIdRef.current");
    expect(controls).toContain(
      "[capability, environment, controlAction, reason]",
    );
    expect(controls).toContain("<SegmentedOtpInput");
    expect(controls).toContain("Loading runtime-control proposals");
    expect(controls).toContain("loaded && proposals.length");
    expect(controls).toContain("Controls unavailable");
    expect(controls).toContain("Approve retained terms?");
    expect(controls).toContain("Apply retained terms?");
    expect(controls).toContain("stepUpGeneration.current");
    expect(controls).toContain("actionVerbs[controlAction]");
    expect(
      controls.match(/<MenuItem value="kill">Kill immediately<\/MenuItem>/g),
    ).toHaveLength(1);
    expect(flat(controls)).toContain("if ( !retrying && mounted.current");
    expect(controls).toContain("commandIdRef.current");
    expect(controls.match(/Controls unavailable/g)).toHaveLength(1);
    expect(controls).not.toContain("<Skeleton");
  });

  it("separates operator queue and exact keyed detail without decoding params", () => {
    const desk = source("(ops)/operators/operators-desk.tsx");
    const page = source("(ops)/operators/[principalId]/page.tsx");
    expect(desk).toContain(
      "href={`/operators/${encodeURIComponent(operator.id)}",
    );
    expect(desk).toMatch(/principalId \? \(\s*<AdminCard\s*variant="detail"/);
    expect(page).toContain("key={principalId}");
    expect(page).not.toContain("decodeURIComponent");
    expect(desk).toContain("<SegmentedOtpInput");
    expect(desk).not.toContain("onDelete=");
    expect(desk).toContain("Confirm audited change");
    expect(desk).toContain("Approve this admin-role change?");
    expect(desk).toContain('aria-live="assertive"');
    expect(desk).toContain("directoryReady && state.operators.length");
    expect(desk).toContain("roleChangesLoaded && roleChanges.length === 0");
    expect(desk).toContain("Suspend this operator?");
    expect(desk).toContain("Reactivate this operator?");
    expect(desk).toMatch(/gridTemplateColumns:\s*"1fr"/);
    expect(desk).toContain("stepUpGeneration.current");
    expect(desk).toContain('aria-label="Exact audited terms"');
    expect(desk).toContain("pendingConfirmation.terms.reason");
    expect(desk).toContain("pendingConfirmation.terms.roles.join");
    expect(desk).toContain("pendingConfirmation.terms.status");
    // A role change must travel as a delta, never as a frozen array. The
    // confirm dialog can sit open indefinitely, and the roles PATCH is a full
    // replace with no version token, so replaying a snapshot silently reverts
    // whatever another administrator granted in the meantime. Assert the
    // mechanism, not the comment above it.
    expect(desk).toMatch(
      /operation:\s*\{\s*type: held \? "remove" : "add",\s*role,\s*\}/,
    );
    expect(desk).toContain('"operation" in body');
    expect(desk).toMatch(
      /terms:\s*\{\s*target: selected\.email,\s*reason: state\.actionReason,/,
    );
    expect(desk).toMatch(
      /terms:\s*\{\s*target: target\?\.email \?\? change\.targetId,\s*proposer: change\.proposerId,\s*reason: change\.reason,\s*roles: \[\.\.\.change\.roles\]/,
    );
  });
});
