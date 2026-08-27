export type CaseQueue = "verification" | "safety" | "care";
export type QueueNoticeCode = "decision-recorded" | "case-resolved";

const fixedOrigin = "https://admin.obiara.invalid";
const noticeText: Record<QueueNoticeCode, string> = {
  "decision-recorded": "The audited verification decision was recorded.",
  "case-resolved": "The care case was resolved with approved resources.",
};

export function sanitizeQueueReturn(
  queue: CaseQueue,
  candidate?: string | null,
) {
  const pathname = `/${queue}`;
  if (
    !candidate ||
    candidate.startsWith("//") ||
    candidate.includes("\\") ||
    /%(?:2e|2f|5c)/i.test(candidate)
  )
    return pathname;
  try {
    const parsed = new URL(candidate, fixedOrigin);
    if (
      parsed.origin !== fixedOrigin ||
      parsed.username ||
      parsed.password ||
      parsed.pathname !== pathname
    )
      return pathname;
    const output = new URLSearchParams();
    const q = parsed.searchParams.get("q");
    if (q) output.set("q", q);
    if (parsed.searchParams.get("search") === "1") output.set("search", "1");
    return `${pathname}${output.size ? `?${output}` : ""}`;
  } catch {
    return pathname;
  }
}

export function buildCasePath(
  queue: CaseQueue,
  caseId: string,
  returnHref?: string,
) {
  const path = `/${queue}/${encodeURIComponent(caseId)}`;
  const safeReturn = returnHref ? sanitizeQueueReturn(queue, returnHref) : null;
  return safeReturn
    ? `${path}?${new URLSearchParams({ return: safeReturn })}`
    : path;
}

export function terminalQueuePath(
  queue: "verification" | "care",
  code: QueueNoticeCode,
  returnHref?: string,
) {
  const parsed = new URL(sanitizeQueueReturn(queue, returnHref), fixedOrigin);
  parsed.searchParams.set("notice", code);
  return `${parsed.pathname}?${parsed.searchParams}`;
}

export function queueNoticeText(code?: string | null) {
  return code && Object.hasOwn(noticeText, code)
    ? noticeText[code as QueueNoticeCode]
    : null;
}
