import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const read = (path: string) =>
  readFileSync(new URL(path, import.meta.url), "utf8");
const members = read("./(ops)/members/members-desk.tsx");
const waitlist = read("./(ops)/waitlist/waitlist-desk.tsx");
const community = read("./(ops)/community/community-desk.tsx");
const mpanyimfo = read("./(ops)/mpanyimfo/mpanyimfo-docket.tsx");
const workforce = read("./(ops)/workforce/workforce-safeguards.tsx");
const styles = read("./styles.css");

describe("people and community redesign batch", () => {
  it("keeps API-backed lists skeleton-first and Strict Mode safe", () => {
    for (const source of [members, waitlist]) {
      expect(source).toContain("AdminSkeleton");
      expect(source).toContain("EmptyState");
      expect(source).not.toContain("CircularProgress");
      expect(source).toContain("mounted.current = true");
      expect(source).toContain("loadGeneration.current += 1");
      expect(source).toContain(
        "!mounted.current || generation !== loadGeneration.current",
      );
      expect(source).toContain("requestController.current?.abort()");
      expect(source).toContain("signal: controller.signal");
      expect(source).toContain('name === "AbortError"');
      expect(source).toContain("setLoaded(true)");
    }
    expect(waitlist).not.toContain("Refreshing…");
    expect(waitlist).toContain('overflowWrap: "anywhere"');
  });

  it("moves bounded member detail into an accessible dialog", () => {
    expect(members).toContain("<Dialog");
    expect(members).toMatch(/setSelectedRef\(\(current\) =>\s*next\.some/);
    expect(members).toContain('aria-haspopup="dialog"');
    expect(members).toContain('aria-controls="member-detail-dialog"');
    expect(members).not.toContain("aria-pressed");
    expect(members).toMatch(/loaded && !error \? \(\s*<Chip/);
    expect(members).toContain("setLoaded(false);\n      setSelectedRef(null);");
    expect(members).toContain(
      "open={loaded && !loading && !error && Boolean(selected)}",
    );
    expect(members).toMatch(
      /id="member-detail-description"\s*className="visually-hidden"/,
    );
    expect(members).not.toContain(
      '<DialogContent id="member-detail-description">',
    );
    expect(members).not.toContain("CircularProgress");
  });

  it("keeps static evidence surfaces vertical and watermark-backed", () => {
    for (const source of [community, mpanyimfo, workforce]) {
      expect(source).toContain("AdminCard");
      expect(source).not.toContain("CircularProgress");
    }
    expect(community).not.toContain('md: "1fr 1fr"');
    expect(mpanyimfo).not.toContain('md: "repeat(2,1fr)"');
    expect(workforce).not.toContain('md: "repeat(2,1fr)"');
    expect(styles).toMatch(
      /\.mpanyimfo-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/s,
    );
    expect(styles).not.toContain(
      "grid-template-columns: minmax(0, 1.05fr) minmax(320px, 0.95fr)",
    );
    const workforceBlocks =
      styles.match(/\.workforce-preview\s*\{[^}]*\}/gs) ?? [];
    expect(workforceBlocks.join("\n")).not.toContain("minmax(220px, 0.35fr)");
  });

  it("does not present failed loads as zero, empty, or stale content", () => {
    expect(members).toContain(
      'error || !loaded ? "Unavailable" : counts[status]',
    );
    expect(members).toContain("error || !loaded ? null : members.length === 0");
    expect(waitlist).toContain(
      'error || !loaded ? "Unavailable" : entries.length',
    );
    expect(waitlist).toContain(
      "!loading && loaded && !error && entries.length === 0",
    );
    expect(waitlist).toMatch(
      /!loading && loaded && !error\s*\?\s*entries\.map/,
    );
    expect(members).not.toMatch(/<Typography[^>]*>\s*<AdminSkeleton/s);
    expect(waitlist).not.toMatch(/<Typography[^>]*>\s*<AdminSkeleton/s);
  });

  it("preserves critical fail-closed copy and ordered evidence", () => {
    expect(community).toContain("Five checks. One bounded proposal.");
    expect(community).toContain("ready for human review only");
    expect(mpanyimfo).toContain("evidence incomplete");
    expect(mpanyimfo).toContain("ready for release review");
    expect(mpanyimfo).toContain("No persisted docket");
    expect(workforce).toContain(
      "This page is guidance, not a staffing console.",
    );
    expect(workforce).toContain("No-penalty opt-out");
  });
});
