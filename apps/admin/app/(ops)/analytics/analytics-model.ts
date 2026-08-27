export type FunnelReport = {
  windowDays: number;
  podsHeardRate: number;
  seedToSproutRate: number;
  sproutToRoomRate: number;
  fireAttendeeCount: number;
  fireAttendanceRate: number;
  regretCount: number;
  regretTrend: "up" | "down" | "flat";
  computedAt: string;
};
export type GateMetric = {
  id: string;
  label: string;
  rate: number;
  threshold: number;
};
export function validFunnelReport(
  value: unknown,
  now = Date.now(),
  maxAgeMs = 48 * 60 * 60 * 1000,
): value is FunnelReport {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>,
    rates = [
      v.podsHeardRate,
      v.seedToSproutRate,
      v.sproutToRoomRate,
      v.fireAttendanceRate,
    ],
    counts = [v.fireAttendeeCount, v.regretCount],
    computedAt =
      typeof v.computedAt === "string" ? Date.parse(v.computedAt) : Number.NaN;
  return (
    v.windowDays === 30 &&
    rates.every(
      (x) => typeof x === "number" && Number.isFinite(x) && x >= 0 && x <= 1,
    ) &&
    counts.every(
      (x) => typeof x === "number" && Number.isSafeInteger(x) && x >= 0,
    ) &&
    ["up", "down", "flat"].includes(String(v.regretTrend)) &&
    Number.isFinite(computedAt) &&
    computedAt <= now + 5 * 60 * 1000 &&
    computedAt >= now - maxAgeMs
  );
}
export function gates(report: FunnelReport): GateMetric[] {
  return [
    {
      id: "pods",
      label: "Pods heard",
      rate: report.podsHeardRate,
      threshold: 65,
    },
    {
      id: "seed",
      label: "Seed to sprout",
      rate: report.seedToSproutRate,
      threshold: 25,
    },
    {
      id: "room",
      label: "Sprout to room",
      rate: report.sproutToRoomRate,
      threshold: 35,
    },
    {
      id: "fire",
      label: "Weekly fire",
      rate: report.fireAttendanceRate,
      threshold: 40,
    },
  ];
}
export const percent = (rate: number) => `${Math.round(rate * 1000) / 10}%`;
