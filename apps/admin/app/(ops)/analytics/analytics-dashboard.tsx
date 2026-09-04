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
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";
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
    <AdminCard
      className={`analytics-gate ${passes ? "is-pass" : "is-fail"}`}
      component="article"
      variant="metric"
      watermark="analytics"
    >
      <Stack direction="row" sx={{ justifyContent: "space-between", gap: 1 }}>
        <Box className="analytics-gate-value">
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
        className="analytics-gate-progress"
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
    <Box component="main" className="analytics-redesign">
      <Stack
        component="header"
        className="analytics-hero"
        direction={{ xs: "column", md: "row" }}
        sx={{ justifyContent: "space-between", gap: 2 }}
      >
        <Box className="analytics-hero-copy">
          <Button component={Link} href="/" className="analytics-back">
            Return to command centre
          </Button>
          <Box className="analytics-kicker">
            <AdminIcon name="analytics" aria-hidden="true" />
            <Typography className="section-kicker">
              Analytics · release observatory
            </Typography>
          </Box>
          <Typography component="h1">
            Read the signal. Respect what is missing.
          </Typography>
          <Typography className="analytics-hero-intro">
            A privacy-bounded view of the live P0 funnel. Observed movement
            informs release review; it never clears unresolved safety, fairness
            or retention evidence.
          </Typography>
        </Box>
        <Box
          className="analytics-hero-register"
          aria-label="Analytics report status"
        >
          {loading ? (
            <AdminSkeleton variant="metric" rows={1} />
          ) : report ? (
            <Box>
              <span>Observation window</span>
              <strong>{report.windowDays} days</strong>
              <Typography color="text.secondary">
                <time dateTime={report.computedAt}>
                  Computed {new Date(report.computedAt).toLocaleString()}
                </time>
              </Typography>
            </Box>
          ) : null}
          <Box>
            <span>Overall release</span>
            <strong>Blocked</strong>
          </Box>
          <Box>
            <span>Override</span>
            <strong>Unavailable</strong>
          </Box>
        </Box>
        <AdminCardWatermark watermark="analytics" />
      </Stack>
      <section className="analytics-boundary" aria-label="Release boundary">
        <span className="analytics-boundary-icon">
          <UtilityIcon name="security" aria-hidden="true" />
        </span>
        <Box>
          <Typography className="section-kicker">
            Non-overridable boundary
          </Typography>
          <Typography component="h2">Release remains blocked.</Typography>
          <Typography>
            Uncomposed retention, fairness, safety, and external readiness
            evidence cannot be overridden here.
          </Typography>
        </Box>
        <span className="analytics-boundary-state">Evidence incomplete</span>
      </section>
      {loading ? (
        <AdminCard
          className="analytics-state-card"
          variant="panel"
          watermark="analytics"
          showWatermark={false}
        >
          <AdminSkeleton variant="card-list" rows={5} />
        </AdminCard>
      ) : error ? (
        <AdminCard
          variant="warning"
          watermark="analytics"
          className="analytics-state-card"
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
          <Alert
            className="analytics-observed-state"
            severity={observedPass ? "success" : "warning"}
          >
            Observed P0 funnel thresholds{" "}
            {observedPass ? "pass" : "do not all pass"}. This does not clear the
            overall release.
          </Alert>
          <section
            className="analytics-gates"
            aria-labelledby="analytics-gates-title"
          >
            <Box className="analytics-section-heading">
              <Box>
                <Typography className="section-kicker">
                  P0 gate runway
                </Typography>
                <Typography id="analytics-gates-title" component="h2">
                  Four observed thresholds
                </Typography>
              </Box>
              <Typography>
                Each rate is compared directly with its configured release
                threshold. Passing this runway does not clear the overall
                release.
              </Typography>
            </Box>
            <Box className="analytics-gate-grid">
              {metrics.map((metric) => (
                <Metric key={metric.id} metric={metric} />
              ))}
            </Box>
          </section>
          <Box className="analytics-secondary-signals">
            <AdminCard
              component="article"
              variant="metric"
              watermark="analytics"
              className="analytics-signal analytics-signal--attendance"
            >
              <span className="analytics-signal-icon">
                <AdminIcon name="community" aria-hidden="true" />
              </span>
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
            <AdminCard
              className="analytics-signal analytics-signal--regret"
              component="article"
              variant="metric"
              watermark="care"
            >
              <span className="analytics-signal-icon">
                <AdminIcon name="care" aria-hidden="true" />
              </span>
              <Typography color="text.secondary">Regret reports</Typography>
              <Typography component="h3" sx={{ fontSize: 30, fontWeight: 800 }}>
                {report.regretCount}
              </Typography>
              <Chip
                color={report.regretTrend === "down" ? "success" : "warning"}
                label={`Trend ${report.regretTrend}`}
              />
            </AdminCard>
          </Box>
        </>
      ) : null}
      <AdminCard
        className="analytics-unavailable"
        variant="warning"
        watermark="evidence"
      >
        <Typography component="h2">Unavailable release evidence</Typography>
        <Typography>
          D30 retention, fairness drift and unresolved safety-tier evidence are
          not composed into this runtime report. They remain release blockers;
          this desk does not invent substitute values.
        </Typography>
      </AdminCard>
    </Box>
  );
}
