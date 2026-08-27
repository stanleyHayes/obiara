export const markets = ["gh_en", "gh_tw", "gh_pidgin", "gh_ga"] as const;
export type Market = (typeof markets)[number];
export type Pack = {
  packId: string;
  market: Market;
  terminologyRef: string;
  features: Record<string, boolean>;
  status: "draft" | "published" | "retired";
  version: number;
  createdAt: string;
  publishedAt?: string;
  proposedByMe?: boolean;
  approvedByMe?: boolean;
};
export type PendingPack =
  | {
      action: "draft";
      market: Market;
      terminologyRef: string;
      features: Record<string, boolean>;
    }
  | { action: "publish" | "retire"; pack: Pack };
export type PromptInput = {
  id: string;
  version: number;
  language: string;
  cue: string;
  acceptedAnswers: string[];
  source: {
    kind: "book" | "oral_archive" | "institutional_archive";
    citation: string;
    locator?: string;
  };
};
export type Cohort = {
  id: string;
  capacity: 4 | 8 | 16;
  enrolled: number;
  joined: boolean;
  status: "open" | "locked" | "started";
  competitionId?: string;
  revision: number;
};
export type Review = {
  id: string;
  matchId: string;
  status: "open" | "resolved" | "appealed" | "final";
  decision: "none" | "no_action" | "rules_action";
  openedAt: string;
  resolvedAt?: string;
  yours: boolean;
};
export type Competition = {
  id: string;
  revision: number;
  status: "active" | "completed";
  ladder: { label: string; played: number; wins: number; you: boolean }[];
  matches: {
    id: string;
    round: number;
    slot: number;
    firstLabel: string;
    secondLabel: string;
    winnerLabel?: string;
    resultRecorded: boolean;
    youArePlaying: boolean;
  }[];
  reviews: Review[];
};
export type PendingTournament =
  | { kind: "create"; capacity: 4 | 8 | 16; key: string }
  | { kind: "start"; cohortId: string; expectedRevision: number; key: string }
  | {
      kind: "review";
      cohortId: string;
      competitionId: string;
      reviewId: string;
      decision: "no_action" | "rules_action";
      expectedRevision: number;
      appeal: boolean;
      key: string;
    };
const record = (v: unknown): v is Record<string, unknown> =>
  Boolean(v) && typeof v === "object";
const text = (v: unknown) => typeof v === "string" && v.length > 0;
const integer = (v: unknown, min = 0) =>
  typeof v === "number" && Number.isSafeInteger(v) && v >= min;
const instant = (v: unknown) =>
  typeof v === "string" && Number.isFinite(Date.parse(v));
export function validPack(v: unknown): v is Pack {
  if (!record(v)) return false;
  return (
    text(v.packId) &&
    markets.includes(v.market as Market) &&
    text(v.terminologyRef) &&
    record(v.features) &&
    Object.values(v.features).every((x) => typeof x === "boolean") &&
    ["draft", "published", "retired"].includes(String(v.status)) &&
    integer(v.version, 1) &&
    instant(v.createdAt) &&
    (v.publishedAt === undefined || instant(v.publishedAt)) &&
    (v.proposedByMe === undefined || typeof v.proposedByMe === "boolean") &&
    (v.approvedByMe === undefined || typeof v.approvedByMe === "boolean")
  );
}
export function validPackResult(v: unknown, p: PendingPack) {
  if (!validPack(v)) return false;
  if (p.action === "draft")
    return (
      v.status === "draft" &&
      v.version === 1 &&
      v.market === p.market &&
      v.terminologyRef === p.terminologyRef &&
      Object.keys(v.features).length === Object.keys(p.features).length &&
      Object.entries(p.features).every(
        ([key, enabled]) => v.features[key] === enabled,
      )
    );
  return (
    v.packId === p.pack.packId &&
    v.version === p.pack.version + 1 &&
    v.status === (p.action === "publish" ? "published" : "retired") &&
    v.market === p.pack.market &&
    v.terminologyRef === p.pack.terminologyRef &&
    Date.parse(v.createdAt) === Date.parse(p.pack.createdAt) &&
    Object.keys(v.features).length === Object.keys(p.pack.features).length &&
    Object.entries(p.pack.features).every(
      ([key, enabled]) => v.features[key] === enabled,
    ) &&
    (p.action !== "retire" ||
      !p.pack.publishedAt ||
      Date.parse(v.publishedAt ?? "") === Date.parse(p.pack.publishedAt))
  );
}
export function promptErrors(p: PromptInput) {
  const e: Record<string, string> = {};
  if (!/^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/.test(p.id))
    e.id = "Use a stable ID up to 128 safe characters.";
  if (!integer(p.version, 1)) e.version = "Use a whole version of 1 or more.";
  if (!/^[a-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$/.test(p.language))
    e.language = "Enter a valid BCP 47 language tag.";
  if (p.cue.length < 1 || [...p.cue].length > 500)
    e.cue = "Enter 1–500 characters.";
  if (
    !["book", "oral_archive", "institutional_archive"].includes(p.source.kind)
  )
    e.sourceKind = "Choose an approved source kind.";
  const normalized = p.acceptedAnswers.map((answer) =>
    answer.normalize("NFKC").toLocaleLowerCase().trim().replace(/\s+/g, " "),
  );
  if (
    p.acceptedAnswers.length < 1 ||
    p.acceptedAnswers.length > 20 ||
    p.acceptedAnswers.some((x) => !x || [...x].length > 280) ||
    new Set(normalized).size !== p.acceptedAnswers.length
  )
    e.acceptedAnswers = "Enter 1–20 unique answers, one per line.";
  if (p.source.citation.length < 1 || [...p.source.citation].length > 500)
    e.citation = "Enter a complete citation.";
  if (p.source.locator && p.source.locator.length <= 1000) {
    try {
      if (new URL(p.source.locator).protocol !== "https:") throw 0;
    } catch {
      e.locator = "Use a valid HTTPS locator.";
    }
  } else if (p.source.locator)
    e.locator = "Use an HTTPS locator up to 1000 characters.";
  return e;
}
export function validPromptResult(v: unknown, p: PromptInput) {
  return (
    record(v) &&
    v.id === p.id &&
    v.version === p.version &&
    v.language === p.language &&
    v.cue === p.cue &&
    ["book", "oral_archive", "institutional_archive"].includes(
      String(v.sourceKind),
    ) &&
    v.sourceKind === p.source.kind &&
    v.sourceCitation === p.source.citation &&
    ((!p.source.locator && v.sourceLocator === undefined) ||
      v.sourceLocator === p.source.locator)
  );
}
export function validCohort(v: unknown): v is Cohort {
  if (!record(v)) return false;
  return (
    text(v.id) &&
    [4, 8, 16].includes(v.capacity as number) &&
    integer(v.enrolled) &&
    Number(v.enrolled) <= Number(v.capacity) &&
    typeof v.joined === "boolean" &&
    ["open", "locked", "started"].includes(String(v.status)) &&
    (v.competitionId === undefined || text(v.competitionId)) &&
    (v.status === "started") === text(v.competitionId) &&
    (v.status !== "locked" || v.enrolled === v.capacity) &&
    (v.status !== "open" || Number(v.enrolled) < Number(v.capacity)) &&
    integer(v.revision, 1)
  );
}
export function validCompetition(v: unknown): v is Competition {
  if (
    !record(v) ||
    !text(v.id) ||
    !integer(v.revision, 1) ||
    !["active", "completed"].includes(String(v.status)) ||
    !Array.isArray(v.ladder) ||
    !Array.isArray(v.matches) ||
    !Array.isArray(v.reviews)
  )
    return false;
  return (
    v.ladder.every(
      (x) =>
        record(x) &&
        text(x.label) &&
        integer(x.played) &&
        integer(x.wins) &&
        Number(x.wins) <= Number(x.played) &&
        typeof x.you === "boolean",
    ) &&
    v.matches.every(
      (x) =>
        record(x) &&
        text(x.id) &&
        integer(x.round, 1) &&
        integer(x.slot, 0) &&
        text(x.firstLabel) &&
        text(x.secondLabel) &&
        (x.winnerLabel === undefined || text(x.winnerLabel)) &&
        typeof x.resultRecorded === "boolean" &&
        typeof x.youArePlaying === "boolean",
    ) &&
    v.reviews.every(
      (x) =>
        record(x) &&
        text(x.id) &&
        text(x.matchId) &&
        ["open", "resolved", "appealed", "final"].includes(String(x.status)) &&
        ["none", "no_action", "rules_action"].includes(String(x.decision)) &&
        instant(x.openedAt) &&
        (x.resolvedAt === undefined || instant(x.resolvedAt)) &&
        typeof x.yours === "boolean",
    )
  );
}
export function validReviewResult(
  value: unknown,
  pending: Extract<PendingTournament, { kind: "review" }>,
) {
  if (
    !validCompetition(value) ||
    value.id !== pending.competitionId ||
    value.revision <= pending.expectedRevision
  )
    return false;
  const review = value.reviews.find((item) => item.id === pending.reviewId);
  return Boolean(
    review &&
    review.decision === pending.decision &&
    review.status === (pending.appeal ? "final" : "resolved") &&
    review.resolvedAt &&
    instant(review.resolvedAt),
  );
}
type TournamentTerms = PendingTournament extends infer T
  ? T extends { key: string }
    ? Omit<T, "key">
    : never
  : never;
export const tournamentKey = (p: TournamentTerms) => JSON.stringify(p);
