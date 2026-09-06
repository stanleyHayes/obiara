/**
 * The house front: what is resting for you, and what you can tell about it
 * without opening it.
 *
 * A pod is closed until it is opened. The server deliberately does not say
 * who left one, so nothing here tries to infer it — the only things a member
 * knows before opening are that something is there and when it will close.
 */

export interface RestingPod {
  readonly podId: string;
  readonly closesAt: string;
  readonly opened: boolean;
}

export type PodStage = "closed" | "opening" | "open" | "failed";

export interface PodState {
  readonly stage: PodStage;
  /** The short-lived grant, held only for as long as it plays. */
  readonly playbackUrl: string | null;
  readonly error: string | null;
}

export const initialPod: PodState = {
  stage: "closed",
  playbackUrl: null,
  error: null,
};

/**
 * How long a pod has left, in words a person uses.
 *
 * Closing is the only urgency a member is given, so it is worth saying
 * plainly rather than as a timestamp they have to do arithmetic on.
 */
export function closesIn(closesAt: string, now: Date): string {
  const closing = Date.parse(closesAt);
  if (!Number.isFinite(closing)) return "";
  const hours = Math.floor((closing - now.getTime()) / 3_600_000);
  if (hours < 0) return "Closed";
  if (hours < 1) return "Closes within the hour";
  if (hours < 24) return `Closes in ${hours} ${hours === 1 ? "hour" : "hours"}`;
  const days = Math.floor(hours / 24);
  return `Closes in ${days} ${days === 1 ? "day" : "days"}`;
}

/**
 * The line under the house front, which is the whole state of somebody's
 * doorway in one sentence.
 *
 * Zero says so plainly. An empty house front is not a failure and should not
 * read like one — it is the ordinary state of most days.
 */
export function houseFrontSummary(pods: readonly RestingPod[]): string {
  if (pods.length === 0) {
    return "Nothing is resting here yet.";
  }
  const unopened = pods.filter((pod) => !pod.opened).length;
  if (unopened === 0) {
    return pods.length === 1
      ? "One pod, already heard."
      : `${pods.length} pods, all heard.`;
  }
  if (unopened === pods.length) {
    return unopened === 1 ? "One pod, unopened." : `${unopened} pods, unopened.`;
  }
  return `${pods.length} pods · ${unopened} still unopened.`;
}

/** Sorts soonest to close first, which is the order the server sends. */
export function soonestFirst(pods: readonly RestingPod[]): RestingPod[] {
  return [...pods].sort((a, b) => Date.parse(a.closesAt) - Date.parse(b.closesAt));
}
