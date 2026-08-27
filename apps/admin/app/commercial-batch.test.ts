import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";
const root = path.resolve(import.meta.dirname),
  read = (p: string) => fs.readFileSync(path.join(root, p), "utf8");
describe("commercial routes", () => {
  it("separates exact matchmaker licensing and escrow authorities", () => {
    const desk = read("(ops)/matchmakers/matchmaker-licensing-desk.tsx"),
      detail = read("(ops)/matchmakers/[matchmakerId]/page.tsx"),
      create = read("(ops)/matchmakers/new/page.tsx"),
      escrow = read("(ops)/matchmakers/escrow/page.tsx");
    expect(create).toContain('mode="form"');
    expect(escrow).toContain('mode="escrow"');
    expect(desk).toContain("expectedVersion: form.matchmakerId");
    expect(desk).toContain('? "Non-expired"');
    expect(desk).toContain("fundKey.current");
    expect(desk).toContain("deliveryKey.current");
    expect(desk).toContain("body.items.every(validMatchmakerProfile)");
    expect(desk).toContain("validLicenseResult(payload, snapshot)");
    expect(desk).toContain("validEscrowWriteResult(payload, snapshot)");
    expect(desk).toContain("<SegmentedOtpInput");
    expect(detail).not.toContain("decodeURIComponent");
  });
  it("keeps finance evidence truthful and settlement retry safe", () => {
    const desk = read("(ops)/finance/finance-desk.tsx");
    expect(desk).toContain("validFinanceOverview(body)");
    expect(desk).toContain("loaded && exceptions.length === 0");
    expect(desk).toContain('"Idempotency-Key": snapshot.key');
    expect(desk).toContain("validSettlementFor(payload, snapshot)");
    expect(desk).toContain("settlementSnapshot()");
    expect(desk).toContain("Boolean(pending) && !mfaOpen");
    expect(desk).toContain("Escrow {statement.escrow.escrowId}");
    expect(desk).toContain("!error ? (");
    expect(desk).toContain("exceptions.length > 0");
    expect(desk).toContain("checkpoints.length > 0");
    expect(desk).toContain('component="article"');
    expect(desk).not.toContain("<Card");
    expect(desk).toContain("<SegmentedOtpInput");
    expect(desk).not.toContain("CircularProgress");
    expect(desk).not.toContain("Settling…");
    expect(desk.indexOf('variant="form"')).toBeLessThan(
      desk.lastIndexOf("!error ? ("),
    );
  });
});
