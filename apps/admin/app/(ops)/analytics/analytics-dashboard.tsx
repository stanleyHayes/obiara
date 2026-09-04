"use client";
import {
  Alert,
  Box,
  Button,
  Chip,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard } from "../../admin-card";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";
import {
  gates,
  percent,
  validFunnelReport,
  type FunnelReport,
  type GateMetric,
} from "./analytics-model";
import { adminFetch } from "../../lib/admin-fetch";
function Metric({ metric }: { metric: GateMetric }) {
  // Use the same rounding as percent() so pass/fail agrees with displayed value
  const percentValue = Math.round(metric.rate * 1000) / 10,
    passes = percentValue >= metric.threshold;
  return (
    <AdminCard component="article" variant="metric" watermark="analytics">
      <Stack direction="row" sx={{ justifyContent: "space-between", gap: 1 }}>
        <Box>
          <Typography color="text.secondary">{metric.label}</Typography>
          <Typography component="h3" sx={{ fontSize: 30, fontWeight: 800 }}>
            {percent(metric.rate)}
          </Typography>
        </Box>
        <Chip
          color={passes ? "success" : "error"}
          label={passes ? "Pass" : "Fail"}
        />
      </Stack>
      <LinearProgress
        aria-label={`${metric.label} rate`}
        aria-valuetext={`${percent(metric.rate)}, threshold at least ${metric.threshold}%`}
        value={Math.min(100, percentValue)}
        variant="determinate"
      />
      <Typography color="text.secondary">
        Threshold ≥ {metric.threshold}% ·{" "}
        {passes ? "observed threshold passes" : "observed threshold fails"}
      </Typography>
    </AdminCard>
  );
}
export function AnalyticsDashboard() {
  const [report, setReport] = useState<FunnelReport | null>(null),
    [loading, setLoading] = useState(true),
    [error, setError] = useState("");
  const mounted = useRef(false),
    generation = useRef(0),
    controller = useRef<AbortController | null>(null);
  const load = useCallback(async () => {
    const gen = ++generation.current;
    controller.current?.abort();
    const c = new AbortController();
    controller.current = c;
    setLoading(true);
    setError("");
    setReport(null);
    try {
      const response = await adminFetch("/api/analytics?days=30", {
          cache: "no-store",
          signal: c.signal,
        }),
        payload: unknown = await response.json().catch(() => null);
      if (!response.ok || !validFunnelReport(payload))
        throw new Error(
          payload &&
            typeof payload === "object" &&
            "message" in payload &&
            typeof payload.message === "string"
            ? payload.message
            : "Funnel evidence could not load.",
        );
      if (mounted.current && gen === generation.current) setReport(payload);
    } catch (e) {
      if (!c.signal.aborted && mounted.current && gen === generation.current)
        setError(
          e instanceof Error ? e.message : "Funnel evidence could not load.",
        );
    } finally {
      if (mounted.current && gen === generation.current) setLoading(false);
    }
  }, []);
  useEffect(() => {
    mounted.current = true;
    const lifecycle = generation;
    const t = setTimeout(() => void load(), 0);
    return () => {
      clearTimeout(t);
      mounted.current = false;
      lifecycle.current++;
      controller.current?.abort();
    };
  }, [load]);
  const metrics = useMemo(() => (report ? gates(report) : []), [report]),
    observedPass =
      Boolean(report) && metrics.every((m) => m.rate * 100 >= m.threshold);
  return (
    <Stack spacing={3}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        sx={{ justifyContent: "space-between", gap: 2 }}
      >
        <Box>
          <Typography className="section-kicker">RELEASE EVIDENCE</Typography>
          <Typography component="h1">Evidence before expansion.</Typography>
          {loading ? (
            <AdminSkeleton variant="metric" rows={1} />
          ) : report ? (
            <Typography color="text.secondary">
              <time dateTime={report.computedAt}>
                {report.windowDays}-day live window · computed{" "}
                {new Date(report.computedAt).toLocaleString()}
              </time>
            </Typography>
          ) : null}
        </Box>
        <Button component={Link} href="/" variant="outlined">
          Back to command centre
        </Button>
      </Stack>
      <Alert severity="error">
        <strong>Release remains blocked.</strong> Uncomposed retention,
        fairness, safety, and external readiness evidence cannot be overridden
        here.
      </Alert>
      {loading ? (
        <AdminCard variant="panel" watermark="analytics" showWatermark={false}>
          <AdminSkeleton variant="card-list" rows={5} />
        </AdminCard>
      ) : error ? (
        <AdminCard
          variant="warning"
          watermark="analytics"
          showWatermark={false}
        >
          <EmptyState
            icon="!"
            title="Funnel evidence unavailable"
            description={error}
            variant="warning"
            action={<Button onClick={() => void load()}>Retry</Button>}
          />
        </AdminCard>
      ) : report ? (
        <>
          <Alert severity={observedPass ? "success" : "warning"}>
            Observed P0 funnel thresholds{" "}
            {observedPass ? "pass" : "do not all pass"}. This does not clear the
            overall release.
          </Alert>
          <Stack spacing={1.5}>
            {metrics.map((metric) => (
              <Metric key={metric.id} metric={metric} />
            ))}
            <AdminCard
              component="article"
              variant="metric"
              watermark="analytics"
            >
              <Typography color="text.secondary">
                Weekly fire attendees
              </Typography>
              <Typography component="h3" sx={{ fontSize: 30, fontWeight: 800 }}>
                {report.fireAttendeeCount}
              </Typography>
              <Typography color="text.secondary">
                Exact distinct attendee count returned for the weekly window.
              </Typography>
            </AdminCard>
            <AdminCard component="article" variant="metric" watermark="care">
              <Typography color="text.secondary">Regret reports</Typography>
              <Typography component="h3" sx={{ fontSize: 30, fontWeight: 800 }}>
                {report.regretCount}
              </Typography>
              <Chip
                color={report.regretTrend === "down" ? "success" : "warning"}
                label={`Trend ${report.regretTrend}`}
              />
            </AdminCard>
          </Stack>
        </>
      ) : null}
      <AdminCard variant="warning" watermark="evidence">
        <Typography component="h2">Unavailable release evidence</Typography>
        <Typography>
          D30 retention, fairness drift and unresolved safety-tier evidence are
          not composed into this runtime report. They remain release blockers;
          this desk does not invent substitute values.
        </Typography>
      </AdminCard>
    </Stack>
  );
}
