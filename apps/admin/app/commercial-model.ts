export type LicenseBody = {
  matchmakerId?: string;
  licenseId: string;
  jurisdiction: string;
  expectedVersion: number;
  validFrom: string;
  validUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  displayName: string;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};
export type PendingLicense = { kind: "license"; body: LicenseBody };
export type PendingFund = {
  kind: "fund";
  body: { action: "fund"; engagementId: string; fundingRef: string };
  key: string;
};
export type PendingDelivery = {
  kind: "delivery";
  body: { action: "delivery"; escrowId: string; milestoneId: string };
  key: string;
};
export type PendingCommercial = PendingLicense | PendingFund | PendingDelivery;
export type PendingSettlement = {
  escrowId: string;
  milestoneId: string;
  key: string;
};
export type MatchmakerProfile = {
  matchmakerId: string;
  displayName: string;
  licenseId: string;
  jurisdiction: string;
  licenseVersion: number;
  licenseValidUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};
export const financeExceptionCodes = [
  "ledger_missing",
  "reference_mismatch",
  "currency_mismatch",
  "amount_mismatch",
  "ledger_unbalanced",
] as const;
export type FinanceExceptionCode = (typeof financeExceptionCodes)[number];
export type MatchmakerFormInput = {
  displayName: string;
  licenseId: string;
  jurisdiction: string;
  validFrom: string;
  validUntil: string;
  minimumFeeGhs: string;
  maximumFeeGhs: string;
  languages: string;
  specialties: string;
  completedEngagements: string;
  rating: string;
};
export type MatchmakerField = keyof MatchmakerFormInput;

export function normalizedUniqueList(value: string) {
  const seen = new Set<string>();
  return value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => {
      const key = item.toLocaleLowerCase();
      if (!item || seen.has(key)) return false;
      seen.add(key);
      return true;
    });
}
export function matchmakerFieldErrors(
  form: MatchmakerFormInput,
): Partial<Record<MatchmakerField, string>> {
  const errors: Partial<Record<MatchmakerField, string>> = {};
  const name = form.displayName.trim(),
    start = Date.parse(form.validFrom),
    end = Date.parse(form.validUntil),
    min = Number(form.minimumFeeGhs),
    max = Number(form.maximumFeeGhs),
    completed = Number(form.completedEngagements),
    rating = Number(form.rating);
  if (name.length < 2 || name.length > 80)
    errors.displayName = "Use 2–80 characters.";
  if (!form.licenseId.trim()) errors.licenseId = "Enter the licence reference.";
  if (!form.jurisdiction.trim())
    errors.jurisdiction = "Enter the jurisdiction.";
  if (!form.validFrom || !Number.isFinite(start))
    errors.validFrom = "Enter a valid start date and time.";
  if (
    !form.validUntil ||
    !Number.isFinite(end) ||
    (Number.isFinite(start) && end <= start)
  )
    errors.validUntil = "Enter an end later than the start.";
  if (
    !form.minimumFeeGhs.trim() ||
    !Number.isFinite(min) ||
    !Number.isSafeInteger(Math.round(min * 100)) ||
    Math.round(min * 100) < 1
  )
    errors.minimumFeeGhs = "Enter at least GHS 0.01.";
  if (
    !form.maximumFeeGhs.trim() ||
    !Number.isFinite(max) ||
    !Number.isSafeInteger(Math.round(max * 100)) ||
    Math.round(max * 100) < Math.round(min * 100)
  )
    errors.maximumFeeGhs = "Enter a maximum at least equal to the minimum.";
  for (const field of ["languages", "specialties"] as const) {
    const raw = form[field]
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      unique = normalizedUniqueList(form[field]);
    if (unique.length < 1 || unique.length > 8 || raw.length !== unique.length)
      errors[field] = "Enter 1–8 unique comma-separated values.";
  }
  if (
    !form.completedEngagements.trim() ||
    !Number.isSafeInteger(completed) ||
    completed < 0
  )
    errors.completedEngagements = "Enter a non-negative whole number.";
  if (
    !form.rating.trim() ||
    !Number.isFinite(rating) ||
    rating < 0 ||
    rating > 5
  )
    errors.rating = "Enter a rating from 0 to 5.";
  return errors;
}
export function settlementTermsKey(escrowId: string, milestoneId: string) {
  return `${escrowId.trim()}\u0000${milestoneId.trim()}`;
}

function stringList(value: unknown): value is string[] {
  return (
    Array.isArray(value) &&
    value.length >= 1 &&
    value.length <= 8 &&
    value.every(
      (item) =>
        typeof item === "string" && item.trim() === item && item.length > 0,
    ) &&
    new Set(value.map((item) => item.toLocaleLowerCase())).size === value.length
  );
}
export function validMatchmakerProfile(
  value: unknown,
): value is MatchmakerProfile {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return (
    ["matchmakerId", "licenseId", "jurisdiction"].every(
      (key) =>
        typeof item[key] === "string" && String(item[key]).trim().length > 0,
    ) &&
    typeof item.displayName === "string" &&
    item.displayName.trim() === item.displayName &&
    item.displayName.length >= 2 &&
    item.displayName.length <= 80 &&
    typeof item.licenseVersion === "number" &&
    Number.isSafeInteger(item.licenseVersion) &&
    item.licenseVersion >= 1 &&
    typeof item.licenseValidUntil === "string" &&
    Number.isFinite(Date.parse(item.licenseValidUntil)) &&
    typeof item.minimumFeePesewas === "number" &&
    Number.isSafeInteger(item.minimumFeePesewas) &&
    item.minimumFeePesewas >= 1 &&
    typeof item.maximumFeePesewas === "number" &&
    Number.isSafeInteger(item.maximumFeePesewas) &&
    item.maximumFeePesewas >= item.minimumFeePesewas &&
    stringList(item.languages) &&
    stringList(item.specialties) &&
    typeof item.completedEngagements === "number" &&
    Number.isSafeInteger(item.completedEngagements) &&
    item.completedEngagements >= 0 &&
    typeof item.ratingBasisPoints === "number" &&
    Number.isSafeInteger(item.ratingBasisPoints) &&
    item.ratingBasisPoints >= 0 &&
    item.ratingBasisPoints <= 500
  );
}
export function validLicenseResult(value: unknown, pending: PendingLicense) {
  if (!validMatchmakerProfile(value)) return false;
  const body = pending.body;
  return (
    (!body.matchmakerId || value.matchmakerId === body.matchmakerId) &&
    value.licenseId === body.licenseId &&
    value.jurisdiction === body.jurisdiction &&
    value.licenseVersion === body.expectedVersion + 1 &&
    Date.parse(value.licenseValidUntil) === Date.parse(body.validUntil) &&
    value.displayName === body.displayName &&
    value.minimumFeePesewas === body.minimumFeePesewas &&
    value.maximumFeePesewas === body.maximumFeePesewas &&
    JSON.stringify(value.languages) === JSON.stringify(body.languages) &&
    JSON.stringify(value.specialties) === JSON.stringify(body.specialties) &&
    value.completedEngagements === body.completedEngagements &&
    value.ratingBasisPoints === body.ratingBasisPoints
  );
}

function validEscrow(value: unknown): value is Record<string, unknown> & {
  escrowId: string;
  engagementId: string;
  milestones: Record<string, unknown>[];
} {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return (
    typeof item.escrowId === "string" &&
    item.escrowId.length > 0 &&
    typeof item.engagementId === "string" &&
    item.engagementId.length > 0 &&
    typeof item.termsId === "string" &&
    item.termsId.length > 0 &&
    typeof item.disputed === "boolean" &&
    [item.fundedPesewas, item.termsVersion, item.revision].every(
      (number) =>
        typeof number === "number" &&
        Number.isSafeInteger(number) &&
        number >= 1,
    ) &&
    typeof item.settledPesewas === "number" &&
    Number.isSafeInteger(item.settledPesewas) &&
    item.settledPesewas >= 0 &&
    item.settledPesewas <= Number(item.fundedPesewas) &&
    Array.isArray(item.milestones) &&
    item.milestones.length > 0 &&
    item.milestones.every((raw) => {
      if (!raw || typeof raw !== "object") return false;
      const milestone = raw as Record<string, unknown>;
      return (
        typeof milestone.id === "string" &&
        milestone.id.length > 0 &&
        typeof milestone.grossPesewas === "number" &&
        Number.isSafeInteger(milestone.grossPesewas) &&
        milestone.grossPesewas >= 1 &&
        typeof milestone.feePesewas === "number" &&
        Number.isSafeInteger(milestone.feePesewas) &&
        milestone.feePesewas >= 0 &&
        milestone.feePesewas <= milestone.grossPesewas &&
        typeof milestone.deliveryConfirmed === "boolean" &&
        typeof milestone.acceptanceConfirmed === "boolean" &&
        typeof milestone.settled === "boolean" &&
        (milestone.statementRef === undefined ||
          (typeof milestone.statementRef === "string" &&
            milestone.statementRef.length > 0))
      );
    })
  );
}
export function validEscrowWriteResult(
  value: unknown,
  pending: PendingFund | PendingDelivery,
) {
  if (!validEscrow(value)) return false;
  if (pending.kind === "fund")
    return value.engagementId === pending.body.engagementId;
  if (value.escrowId !== pending.body.escrowId) return false;
  const milestone = value.milestones.find(
    (entry) => entry?.id === pending.body.milestoneId,
  );
  return Boolean(milestone && milestone.deliveryConfirmed === true);
}

export function validSettlementResponse(value: unknown): value is {
  statementRef: string;
  grossPesewas: number;
  feePesewas: number;
  netPesewas: number;
  settledAt: string;
  escrow: Record<string, unknown>;
  message?: string;
} {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return (
    typeof item.statementRef === "string" &&
    item.statementRef.length > 0 &&
    typeof item.settledAt === "string" &&
    Number.isFinite(Date.parse(item.settledAt)) &&
    Boolean(item.escrow) &&
    typeof item.escrow === "object" &&
    typeof item.grossPesewas === "number" &&
    Number.isSafeInteger(item.grossPesewas) &&
    typeof item.feePesewas === "number" &&
    Number.isSafeInteger(item.feePesewas) &&
    item.feePesewas >= 0 &&
    typeof item.netPesewas === "number" &&
    Number.isSafeInteger(item.netPesewas) &&
    item.netPesewas >= 0 &&
    item.grossPesewas > 0 &&
    item.feePesewas + item.netPesewas === item.grossPesewas
  );
}
export function validSettlementFor(
  value: unknown,
  expected: PendingSettlement,
): value is {
  statementRef: string;
  grossPesewas: number;
  feePesewas: number;
  netPesewas: number;
  settledAt: string;
  escrow: {
    escrowId: string;
    engagementId: string;
    fundedPesewas: number;
    settledPesewas: number;
    termsId: string;
    termsVersion: number;
    milestones: Array<{
      id: string;
      grossPesewas: number;
      feePesewas: number;
      deliveryConfirmed: boolean;
      acceptanceConfirmed: boolean;
      settled: boolean;
      statementRef?: string;
    }>;
    disputed: boolean;
    revision: number;
  };
} {
  if (!validSettlementResponse(value)) return false;
  const item = value.escrow;
  if (
    item.escrowId !== expected.escrowId ||
    typeof item.engagementId !== "string" ||
    !item.engagementId ||
    typeof item.termsId !== "string" ||
    !item.termsId ||
    item.disputed !== false
  )
    return false;
  for (const key of ["fundedPesewas", "termsVersion", "revision"]) {
    const amount = item[key];
    if (
      typeof amount !== "number" ||
      !Number.isSafeInteger(amount) ||
      amount < 1
    )
      return false;
  }
  if (
    typeof item.settledPesewas !== "number" ||
    !Number.isSafeInteger(item.settledPesewas) ||
    item.settledPesewas < 0 ||
    !Array.isArray(item.milestones)
  )
    return false;
  const milestone = item.milestones.find(
    (entry) =>
      entry &&
      typeof entry === "object" &&
      (entry as Record<string, unknown>).id === expected.milestoneId,
  ) as Record<string, unknown> | undefined;
  if (
    !milestone ||
    milestone.deliveryConfirmed !== true ||
    milestone.acceptanceConfirmed !== true ||
    milestone.settled !== true ||
    milestone.statementRef !== value.statementRef
  )
    return false;
  return (
    typeof milestone.grossPesewas === "number" &&
    Number.isSafeInteger(milestone.grossPesewas) &&
    milestone.grossPesewas >= 1 &&
    typeof milestone.feePesewas === "number" &&
    Number.isSafeInteger(milestone.feePesewas) &&
    milestone.feePesewas >= 0 &&
    milestone.grossPesewas === value.grossPesewas &&
    milestone.feePesewas === value.feePesewas
  );
}

export function validFinanceOverview(value: unknown): value is {
  exceptions: Array<Record<string, unknown>>;
  checkpoints: Array<Record<string, unknown>>;
  exceptionCodes: FinanceExceptionCode[];
} {
  if (!value || typeof value !== "object") return false;
  const body = value as Record<string, unknown>;
  if (
    !Array.isArray(body.exceptions) ||
    !Array.isArray(body.checkpoints) ||
    !Array.isArray(body.exceptionCodes) ||
    body.exceptions.length > 50 ||
    body.checkpoints.length > 14
  )
    return false;
  const allowed = new Set<string>(financeExceptionCodes);
  if (
    body.exceptionCodes.length !== financeExceptionCodes.length ||
    new Set(body.exceptionCodes).size !== body.exceptionCodes.length ||
    !body.exceptionCodes.every(
      (code) => typeof code === "string" && allowed.has(code),
    )
  )
    return false;
  const exceptionsValid = body.exceptions.every((raw) => {
    if (!raw || typeof raw !== "object") return false;
    const item = raw as Record<string, unknown>;
    return (
      [
        "factRef",
        "providerRef",
        "statementRef",
        "currency",
        "exception",
        "occurredAt",
        "recordedAt",
      ].every((key) => typeof item[key] === "string") &&
      ["factRef", "providerRef", "statementRef"].every(
        (key) => String(item[key]).length > 0,
      ) &&
      ["GHS", "USD"].includes(String(item.currency)) &&
      allowed.has(String(item.exception)) &&
      Number.isSafeInteger(item.minor) &&
      Number(item.minor) >= 1 &&
      ["occurredAt", "recordedAt"].every((key) =>
        Number.isFinite(Date.parse(String(item[key]))),
      )
    );
  });
  const checkpointsValid = body.checkpoints.every((raw) => {
    if (!raw || typeof raw !== "object") return false;
    const item = raw as Record<string, unknown>;
    const counts = [item.total, item.reconciled, item.excepted];
    return (
      typeof item.day === "string" &&
      /^\d{4}-\d{2}-\d{2}$/.test(item.day) &&
      Number.isFinite(Date.parse(`${item.day}T00:00:00Z`)) &&
      typeof item.completedAt === "string" &&
      Number.isFinite(Date.parse(item.completedAt)) &&
      counts.every(
        (count) =>
          typeof count === "number" &&
          Number.isSafeInteger(count) &&
          count >= 0,
      ) &&
      // The service counts three outcomes but only two of them into these
      // fields: a fact that is "not_due" lands in neither, so the totals
      // bound the parts rather than equalling them. Demanding equality
      // rejected the whole overview the moment one payment was not yet
      // due (reconciliation/application/service.go counts len(facts)).
      Number(item.reconciled) + Number(item.excepted) <= Number(item.total)
    );
  });
  return exceptionsValid && checkpointsValid;
}
