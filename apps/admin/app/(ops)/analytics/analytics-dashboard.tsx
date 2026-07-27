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
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import {
  analyticsReducer,
  initialAnalyticsState,
  releaseBlocked,
  type GateMetric,
} from "./analytics-model";

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
  const [state, dispatch] = useReducer(analyticsReducer, initialAnalyticsState);
  const blocked = releaseBlocked(state);
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
              {state.window} · {state.snapshotRef}
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
          {state.gates.map((metric) => (
            <Metric key={metric.id} metric={metric} />
          ))}
        </Box>

        <Typography
          component="h2"
          sx={{ fontSize: 30, fontWeight: 800, mb: 2, mt: 5 }}
        >
          Quarterly fairness, regret and safety
        </Typography>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0,1fr))" },
          }}
        >
          {state.fairness.map((metric) => (
            <Metric key={metric.id} metric={metric} />
          ))}
        </Box>

        <Card
          sx={{
            borderRadius: 1,
            display: "grid",
            gap: 3,
            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
            mt: 3,
            p: 3,
          }}
        >
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              OPERATOR INTERPRETATION
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 30, fontWeight: 800, mt: 1 }}
            >
              Notes annotate; they never change facts.
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}
            >
              No member-level rows, content, protected-trait proxies or
              micro-cohorts appear here. A review note cannot release, correct,
              rank or enforce anything.
            </Typography>
          </Box>
          {state.reviewState === "none" ? (
            <Box>
              <TextField
                fullWidth
                label="Bounded review note"
                multiline
                onChange={(event) =>
                  dispatch({ type: "review-note", value: event.target.value })
                }
                rows={4}
                value={state.reviewNote}
              />
              <Button
                disabled={state.reviewNote.trim().length < 12}
                fullWidth
                onClick={() => dispatch({ type: "record-review" })}
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Record interpretation
              </Button>
            </Box>
          ) : (
            <Alert severity="info">
              <strong>{state.reviewRef}</strong>
              <br />
              Review note recorded. Aggregate facts and the blocked release
              state remain unchanged.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
