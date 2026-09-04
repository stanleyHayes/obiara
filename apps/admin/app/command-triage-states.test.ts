import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

const routeSources = [
  "./(ops)/page.tsx",
  "./(ops)/verification/verification-queue.tsx",
  "./(ops)/safety/safety-desk.tsx",
  "./(ops)/care/care-queue.tsx",
].map((path) => readFileSync(new URL(path, import.meta.url), "utf8"));
const dashboard = routeSources[0];
const verification = routeSources[1];
const safety = routeSources[2];
const care = routeSources[3];
const styles = readFileSync(new URL("./styles.css", import.meta.url), "utf8");

describe("Command centre and triage designed states", () => {
  it("uses the shared shaped skeleton instead of generic progress UI", () => {
    for (const source of routeSources) {
      expect(source).toContain("AdminSkeleton");
      expect(source).not.toContain("CircularProgress");
      expect(source).not.toMatch(/<Skeleton\b/);
      expect(source).not.toContain("Loading the queue…");
    }
  });

  it("uses the shared animated empty state on every route", () => {
    for (const source of routeSources) {
      expect(source).toContain("EmptyState");
    }
    expect(routeSources.join("\n")).not.toContain(
      '<Alert severity="info">Select a queued case to begin.</Alert>',
    );
    expect(routeSources.join("\n")).not.toContain(
      '<Alert severity="success">No care cases are waiting.</Alert>',
    );
  });

  it("deep-links queue rows to dedicated detail routes", () => {
    expect(dashboard).toMatch(
      /buildCasePath\(\s*"verification",\s*item\.caseId/,
    );
    expect(dashboard).toMatch(/buildCasePath\(\s*"safety",\s*item\.caseId/);
    expect(verification).toMatch(
      /buildCasePath\(\s*"verification",\s*item\.caseId/,
    );
    expect(safety).toMatch(/buildCasePath\(\s*"safety",\s*item\.caseId/);
    expect(care).toMatch(/buildCasePath\(\s*"care",\s*item\.caseId/);
    for (const source of [verification, safety, care]) {
      expect(source).toContain("const detailMode = Boolean(caseId)");
      expect(source).not.toContain('window.location.search).get("case")');
    }
  });

  it("uses segmented OTP controls and stable busy semantics", () => {
    for (const source of [verification, safety, care]) {
      expect(source).toContain("SegmentedOtpInput");
      expect(source).toContain("aria-busy={busy}");
      expect(source).not.toContain('autoComplete="one-time-code"');
    }
  });

  it("keeps sensitive failures inside the active dialog live region", () => {
    for (const source of [verification, safety, care]) {
      expect(source).toContain("dialogError");
      expect(source).toContain('aria-live="assertive"');
      expect(source).toContain('setDialogError("")');
    }
  });

  it("guards overlapping loads and redirects terminal workflows", () => {
    for (const source of [verification, safety, care]) {
      expect(source).toContain("loadGeneration");
      expect(source).toContain("AbortController");
      expect(source).toContain("controller.abort()");
    }
    expect(verification).toMatch(
      /router\.replace\(\s*terminalQueuePath\(\s*"verification",\s*"decision-recorded"/,
    );
    expect(care).toMatch(
      /router\.replace\(\s*terminalQueuePath\(\s*"care",\s*"case-resolved"/,
    );
  });

  it("re-arms mounted state under the Strict Mode setup-cleanup-setup probe", () => {
    for (const source of [verification, safety, care]) {
      expect(source).toMatch(
        /useEffect\(\(\) => \{\s*mounted\.current = true;\s*return \(\) => \{/s,
      );
      expect(source).toContain("mounted.current = false");
      expect(source).toContain("actionGeneration.current += 1");
      expect(source).toContain("loadGeneration.current += 1");
    }
    for (const source of [safety, care]) {
      expect(source).toContain(
        "if (!mounted.current || generation !== loadGeneration.current) return",
      );
      expect(source).toMatch(
        /if \(mounted\.current && generation === loadGeneration\.current\)\s*setLoading\(false\)/,
      );
    }
  });

  it("keeps queue and detail work on separate route modes", () => {
    expect(styles).toMatch(
      /\.verification-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(styles).toMatch(/\.verification-list\s*\{[^}]*position:\s*static;/s);
    for (const source of [verification, safety, care]) {
      expect(source).toContain("{!detailMode ? (");
      expect(source).toContain("{detailMode ? (");
      expect(source).toContain("case not found");
    }
  });

  it("provides a real dashboard-to-verification search path", () => {
    expect(dashboard).toContain("/verification?search=1");
    expect(verification).toContain('label="Search verification cases"');
    expect(verification).toContain("filteredCases");
  });

  it("keeps dashboard work full-width and KPI signals responsive", () => {
    expect(styles).toMatch(
      /\.work-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(dashboard).toMatch(
      /if \(\s*!controller\.signal\.aborted &&\s*\(error as Error\)\.name !== "AbortError"/,
    );
    expect(styles).toMatch(
      /\.command-center \.metrics-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1\.35fr\) repeat\(2, minmax\(0, 1fr\)\);/s,
    );
    expect(styles).toMatch(
      /@media \(max-width:\s*560px\)[\s\S]*?\.command-center \.metrics-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s,
    );
    expect(dashboard).toContain('trusted={verifications.state === "ready"}');
  });

  it("never presents failed triage loads as empty or not found", () => {
    for (const source of [verification, safety, care]) {
      expect(source).toContain("loadError");
      expect(source).toContain("!loadError");
      expect(source).toMatch(/showWatermark=\{!loading && !loadError/);
      expect(source).toMatch(/loadError \? \(\s*<EmptyState/);
    }
  });
});
