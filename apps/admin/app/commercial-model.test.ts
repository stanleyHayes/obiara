import { describe, expect, it } from "vitest";
import {
  financeExceptionCodes,
  matchmakerFieldErrors,
  normalizedUniqueList,
  settlementTermsKey,
  validFinanceOverview,
  validEscrowWriteResult,
  validLicenseResult,
  validMatchmakerProfile,
  validSettlementFor,
  validSettlementResponse,
  type MatchmakerFormInput,
  type PendingSettlement,
} from "./commercial-model";

const form: MatchmakerFormInput = {
  displayName: "Trusted matchmaker",
  licenseId: "lic-1",
  jurisdiction: "GH",
  validFrom: "2026-08-22T12:00",
  validUntil: "2027-08-22T12:00",
  minimumFeeGhs: "0.01",
  maximumFeeGhs: "100",
  languages: "English, Twi",
  specialties: "Marriage, Coaching",
  completedEngagements: "0",
  rating: "0",
};
const expected: PendingSettlement = {
  escrowId: "esc-1",
  milestoneId: "mile-1",
  key: "settle-1",
};
const settlement = {
  statementRef: "statement-1",
  grossPesewas: 100,
  feePesewas: 10,
  netPesewas: 90,
  settledAt: "2026-08-22T12:00:00Z",
  escrow: {
    escrowId: "esc-1",
    engagementId: "eng-1",
    fundedPesewas: 100,
    settledPesewas: 100,
    termsId: "terms-1",
    termsVersion: 1,
    disputed: false,
    revision: 2,
    milestones: [
      {
        id: "mile-1",
        grossPesewas: 100,
        feePesewas: 10,
        deliveryConfirmed: true,
        acceptanceConfirmed: true,
        settled: true,
        statementRef: "statement-1",
      },
    ],
  },
};
const overview = {
  exceptionCodes: [...financeExceptionCodes],
  exceptions: [
    {
      factRef: "fact-1",
      providerRef: "provider-1",
      statementRef: "statement-1",
      currency: "GHS",
      minor: 1,
      exception: "ledger_missing",
      occurredAt: "2026-08-22T12:00:00Z",
      recordedAt: "2026-08-22T12:01:00Z",
    },
  ],
  checkpoints: [
    {
      day: "2026-08-22",
      total: 3,
      reconciled: 2,
      excepted: 1,
      completedAt: "2026-08-22T12:01:00Z",
    },
  ],
};

describe("commercial command model", () => {
  it("keys exact normalized settlement terms", () => {
    expect(settlementTermsKey("esc-1", "mile-1")).toBe(
      settlementTermsKey(" esc-1 ", "mile-1"),
    );
    expect(settlementTermsKey("esc-1", "mile-1")).not.toBe(
      settlementTermsKey("esc-1", "mile-2"),
    );
  });

  it("normalizes lists and enforces every matchmaker boundary", () => {
    expect(normalizedUniqueList(" English, twi, ENGLISH ")).toEqual([
      "English",
      "twi",
    ]);
    expect(matchmakerFieldErrors(form)).toEqual({});
    expect(
      matchmakerFieldErrors({
        ...form,
        displayName: "x",
        minimumFeeGhs: "0",
        maximumFeeGhs: "-1",
        languages: "English, english",
        specialties: "",
        completedEngagements: "1.5",
        rating: "5.1",
        validUntil: form.validFrom,
      }),
    ).toMatchObject({
      displayName: expect.any(String),
      minimumFeeGhs: expect.any(String),
      maximumFeeGhs: expect.any(String),
      languages: expect.any(String),
      specialties: expect.any(String),
      completedEngagements: expect.any(String),
      rating: expect.any(String),
      validUntil: expect.any(String),
    });
    expect(
      matchmakerFieldErrors({
        ...form,
        languages: Array.from({ length: 9 }, (_, i) => `L${i}`).join(","),
      }).languages,
    ).toBeTruthy();
  });

  it("validates register profiles and exact licence write identity and version", () => {
    const pending = {
      kind: "license" as const,
      body: {
        matchmakerId: "match-1",
        licenseId: "lic-1",
        jurisdiction: "GH",
        expectedVersion: 2,
        validFrom: "2026-08-22T12:00:00.000Z",
        validUntil: "2027-08-22T12:00:00.000Z",
        minimumFeePesewas: 1,
        maximumFeePesewas: 10_000,
        displayName: "Trusted matchmaker",
        languages: ["English"],
        specialties: ["Marriage"],
        completedEngagements: 0,
        ratingBasisPoints: 0,
      },
    };
    const profile = {
      matchmakerId: "match-1",
      displayName: "Trusted matchmaker",
      licenseId: "lic-1",
      jurisdiction: "GH",
      licenseVersion: 3,
      licenseValidUntil: "2027-08-22T12:00:00.000Z",
      minimumFeePesewas: 1,
      maximumFeePesewas: 10_000,
      languages: ["English"],
      specialties: ["Marriage"],
      completedEngagements: 0,
      ratingBasisPoints: 0,
    };
    expect(validMatchmakerProfile(profile)).toBe(true);
    expect(validMatchmakerProfile({ ...profile, languages: [] })).toBe(false);
    expect(
      validMatchmakerProfile({ ...profile, licenseValidUntil: "bad" }),
    ).toBe(false);
    expect(validLicenseResult(profile, pending)).toBe(true);
    expect(
      validLicenseResult(
        { ...profile, licenseValidUntil: "2027-08-22T12:00:00Z" },
        pending,
      ),
    ).toBe(true);
    expect(
      validLicenseResult(
        { ...profile, licenseValidUntil: "2027-08-22T12:00:01Z" },
        pending,
      ),
    ).toBe(false);
    expect(validLicenseResult(null, pending)).toBe(false);
    expect(validLicenseResult({}, pending)).toBe(false);
    expect(
      validLicenseResult({ ...profile, matchmakerId: "wrong" }, pending),
    ).toBe(false);
    expect(validLicenseResult({ ...profile, licenseVersion: 2 }, pending)).toBe(
      false,
    );
  });

  it("validates exact funded and delivery escrow results", () => {
    const escrow = settlement.escrow;
    const fund = {
      kind: "fund" as const,
      key: "fund-1",
      body: {
        action: "fund" as const,
        engagementId: "eng-1",
        fundingRef: "a".repeat(64),
      },
    };
    const delivery = {
      kind: "delivery" as const,
      key: "delivery-1",
      body: {
        action: "delivery" as const,
        escrowId: "esc-1",
        milestoneId: "mile-1",
      },
    };
    expect(validEscrowWriteResult(escrow, fund)).toBe(true);
    expect(
      validEscrowWriteResult({ ...escrow, engagementId: "wrong" }, fund),
    ).toBe(false);
    expect(validEscrowWriteResult(escrow, delivery)).toBe(true);
    expect(
      validEscrowWriteResult({ ...escrow, escrowId: "wrong" }, delivery),
    ).toBe(false);
    expect(
      validEscrowWriteResult(
        { ...escrow, milestones: [{ ...escrow.milestones[0], id: "wrong" }] },
        delivery,
      ),
    ).toBe(false);
    expect(
      validEscrowWriteResult(
        {
          ...escrow,
          milestones: [{ ...escrow.milestones[0], deliveryConfirmed: false }],
        },
        delivery,
      ),
    ).toBe(false);
    expect(validEscrowWriteResult(null, delivery)).toBe(false);
    expect(validEscrowWriteResult({}, delivery)).toBe(false);
  });

  it("validates settlement arithmetic, identity, and milestone state", () => {
    expect(validSettlementResponse(settlement)).toBe(true);
    expect(validSettlementResponse({ ...settlement, netPesewas: 89 })).toBe(
      false,
    );
    expect(validSettlementFor(settlement, expected)).toBe(true);
    expect(
      validSettlementFor(
        {
          ...settlement,
          escrow: { ...settlement.escrow, escrowId: "other" },
        },
        expected,
      ),
    ).toBe(false);
    expect(
      validSettlementFor(
        {
          ...settlement,
          escrow: {
            ...settlement.escrow,
            milestones: [
              { ...settlement.escrow.milestones[0], settled: false },
            ],
          },
        },
        expected,
      ),
    ).toBe(false);
  });

  it("accepts only the exact bounded reconciliation contract", () => {
    expect(validFinanceOverview(overview)).toBe(true);
    expect(
      validFinanceOverview({ ...overview, exceptionCodes: ["ledger_missing"] }),
    ).toBe(false);
    expect(
      validFinanceOverview({
        ...overview,
        exceptions: [{ ...overview.exceptions[0], exception: "invented" }],
      }),
    ).toBe(false);
    expect(
      validFinanceOverview({
        ...overview,
        exceptions: [{ ...overview.exceptions[0], minor: 0 }],
      }),
    ).toBe(false);
    expect(
      validFinanceOverview({
        ...overview,
        checkpoints: [{ ...overview.checkpoints[0], day: "2026-99-99" }],
      }),
    ).toBe(false);
    // reconciled + excepted BOUNDS total, it does not equal it: a fact the
    // service scored "not_due" is counted in neither part. Demanding equality
    // rejected the entire overview the moment one payment was not yet due.
    expect(
      validFinanceOverview({
        ...overview,
        checkpoints: [{ ...overview.checkpoints[0], total: 4 }],
      }),
    ).toBe(true);
    // Parts exceeding the total remains impossible and must still be refused.
    expect(
      validFinanceOverview({
        ...overview,
        checkpoints: [{ ...overview.checkpoints[0], total: 2 }],
      }),
    ).toBe(false);
    expect(
      validFinanceOverview({
        ...overview,
        exceptions: Array.from({ length: 51 }, () => overview.exceptions[0]),
      }),
    ).toBe(false);
    expect(
      validFinanceOverview({
        ...overview,
        checkpoints: Array.from({ length: 15 }, () => overview.checkpoints[0]),
      }),
    ).toBe(false);
  });
});
