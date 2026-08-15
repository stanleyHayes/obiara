"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { type GateMetric } from "./analytics-model";

type FunnelReport = {
  windowDays: number;
  podsHeardRate: number;
  seedToSproutRate: number;
  sproutToRoomRate: number;
  fireAttendeeCount: number;
  fireAttendanceRate: number;
  regretCount: number;
  regretTrend: string;
  computedAt: string;
};

function percentMetric(
  id: string,
  label: string,
  rate: number,
  threshold: number,
): GateMetric {
  const percent = Math.round(rate * 1000) / 10;
  return {
    id,
    label,
    numerator: Math.round(rate * 1000),
    denominator: 1000,
    threshold: `≥ ${threshold}%`,
    value: `${percent}%`,
    complete: true,
    passes: percent >= threshold,
  };
}

function Metric({ metric }: Readonly<{ metric: GateMetric }>) {
  const progress =
    metric.denominator > 0
      ? Math.min(100, (metric.numerator / metric.denominator) * 100)
      : 0;
  return (
    <Card variant="outlined" sx={{ borderRadius: 1, p: 2.5 }}>
      <Stack
        direction="row"
        sx={{ alignItems: "flex-start", justifyContent: "space-between" }}
      >
        <Box>
          <Typography
            sx={{ color: "text.secondary", fontSize: 13, fontWeight: 700 }}
          >
            {metric.label}
          </Typography>
          <Typography
            component="strong"
            sx={{ display: "block", fontSize: 30, fontWeight: 800 }}
          >
            {metric.value}
          </Typography>
        </Box>
        <Chip
          color={metric.complete && metric.passes ? "success" : "error"}
          label={
            metric.complete && metric.passes
              ? "Pass"
              : metric.complete
                ? "Fail"
                : "Incomplete"
          }
          size="small"
        />
      </Stack>
      <LinearProgress
        color={metric.complete && metric.passes ? "success" : "error"}
        sx={{ borderRadius: 99, my: 1.5 }}
        value={progress}
        variant="determinate"
      />
      <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
        {metric.numerator} numerator / {metric.denominator} denominator ·{" "}
        {metric.threshold}
      </Typography>
    </Card>
  );
}

export function AnalyticsDashboard() {
  const [report, setReport] = useState<FunnelReport | null>(null);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void fetch("/api/analytics?days=30", {
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        const payload = (await response.json()) as FunnelReport & {
          message?: string;
        };
        if (!response.ok) throw new Error(payload.message);
        setReport(payload);
      })
      .catch((error: unknown) => {
        if ((error as Error).name !== "AbortError") {
          setLoadError(
            error instanceof Error
              ? error.message
              : "Funnel evidence could not load.",
          );
        }
      });
    return () => controller.abort();
  }, []);

  const liveGates = useMemo<readonly GateMetric[]>(() => {
    if (!report) return [];
    return [
      percentMetric("pods", "Pods heard", report.podsHeardRate, 65),
      percentMetric("seed", "Seed to sprout", report.seedToSproutRate, 25),
      percentMetric("room", "Sprout to room", report.sproutToRoomRate, 35),
      percentMetric("fire", "Weekly fire", report.fireAttendanceRate, 40),
    ];
  }, [report]);
  const blocked = !report || liveGates.some((metric) => !metric.passes);
  return (
    <Box
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="xl">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "center" },
            justifyContent: "space-between",
            mb: 5,
          }}
        >
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.4,
              }}
            >
              RELEASE EVIDENCE
            </Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 44, md: 72 },
                fontWeight: 800,
                letterSpacing: "-0.06em",
                lineHeight: 0.95,
                mt: 1,
              }}
            >
              Evidence before expansion.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              {report
                ? `${report.windowDays}-day live window · computed ${new Date(report.computedAt).toLocaleString()}`
                : "Loading live aggregate evidence…"}
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        <Alert
          severity={blocked ? "error" : "success"}
          sx={{ borderRadius: 1, mb: 3 }}
        >
          <strong>
            {blocked ? "Release is blocked." : "Evidence gates pass."}
          </strong>{" "}
          Missing evidence and failed thresholds cannot be overridden from this
          dashboard.
        </Alert>
        {loadError ? (
          <Alert severity="warning" sx={{ mb: 3 }}>
            {loadError}
          </Alert>
        ) : null}

        <Typography
          component="h2"
          sx={{ fontSize: 30, fontWeight: 800, mb: 2 }}
        >
          P0 phase-exit gates
        </Typography>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0,1fr))" },
          }}
        >
          {liveGates.map((metric) => (
            <Metric key={metric.id} metric={metric} />
          ))}
        </Box>

        <Typography
          component="h2"
          sx={{ fontSize: 30, fontWeight: 800, mb: 2, mt: 5 }}
        >
          Regret and unavailable release evidence
        </Typography>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0,1fr))" },
          }}
        >
          {report ? (
            <Card variant="outlined" sx={{ borderRadius: 1, p: 2.5 }}>
              <Typography
                sx={{ color: "text.secondary", fontSize: 13, fontWeight: 700 }}
              >
                Regret reports
              </Typography>
              <Typography
                component="strong"
                sx={{ display: "block", fontSize: 30, fontWeight: 800 }}
              >
                {report.regretCount}
              </Typography>
              <Chip
                color={report.regretTrend === "down" ? "success" : "warning"}
                label={`Trend ${report.regretTrend}`}
                size="small"
              />
            </Card>
          ) : null}
          <Alert severity="warning">
            D30 retention, fairness drift and unresolved safety-tier evidence
            are not composed into this runtime report. They remain release
            blockers; this desk does not invent substitute values.
          </Alert>
        </Box>
      </Container>
    </Box>
  );
}
