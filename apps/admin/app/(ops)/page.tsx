"use client";

import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { Suspense, useEffect, useState } from "react";

import { HandoverButton } from "./handover-button";
import { TourDialog } from "./tour-dialog";

type VerificationCase = {
  caseId: string;
  reasonCode: "provider_outage" | "provider_uncertain" | "manual_review";
  status?: "queued_manual" | "approved" | "rejected";
  subjectRef: string;
  submittedAt: string;
};

type SafetyCase = {
  caseId: string;
  queue: "triage" | "care";
  slaDueAt: string;
  status: "queued" | "in_review" | "resolved";
  tier: "A" | "B" | "C" | "D";
};

type CareCase = {
  caseId: string;
  status: "open" | "engaged" | "resolved";
};

type LoadResult<T> =
  | { state: "loading" }
  | { state: "error"; message: string }
  | { state: "ready"; value: T };

const loading: LoadResult<never> = { state: "loading" };

async function fetchCases<T>(url: string, fallback: string) {
  const response = await fetch(url, { cache: "no-store" });
  const body = (await response.json().catch(() => null)) as {
    cases?: T[];
    message?: string;
  } | null;
  if (!response.ok || !body?.cases) {
    throw new Error(body?.message ?? fallback);
  }
  return body.cases;
}

function useCases<T>(url: string, fallback: string): LoadResult<T[]> {
  const [result, setResult] = useState<LoadResult<T[]>>(loading);
  useEffect(() => {
    const controller = new AbortController();
    void fetchCases<T>(url, fallback)
      .then((cases) => {
        if (!controller.signal.aborted) {
          setResult({ state: "ready", value: cases });
        }
      })
      .catch((error: unknown) => {
        if ((error as Error).name !== "AbortError") {
          setResult({
            state: "error",
            message: error instanceof Error ? error.message : fallback,
          });
        }
      });
    // The underlying fetch is not aborted on purpose: fetchCases has no
    // signal support; the aborted flag above prevents stale writes.
    return () => controller.abort();
  }, [url, fallback]);
  return result;
}

const reasonLabels: Record<VerificationCase["reasonCode"], string> = {
  provider_outage: "Provider outage fallback",
  provider_uncertain: "Liveness uncertainty",
  manual_review: "Manual review",
};

function waitLabel(submittedAt: string): string {
  const minutes = Math.max(
    0,
    Math.round((Date.now() - new Date(submittedAt).getTime()) / 60000),
  );
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes} min`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

function greeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning.";
  if (hour < 17) return "Good afternoon.";
  return "Good evening.";
}

function MetricCard({
  label,
  value,
  note,
  accent,
  href,
}: Readonly<{
  label: string;
  value: string;
  note: string;
  accent: string;
  href: string;
}>) {
  return (
    <Link href={href} style={{ textDecoration: "none" }}>
      <Card className="metric-card" sx={{ "--metric-accent": accent }}>
        <Typography className="metric-label">{label}</Typography>
        <Typography className="metric-value">{value}</Typography>
        <Typography className="metric-note">{note}</Typography>
      </Card>
    </Link>
  );
}

function metricValue<T>(result: LoadResult<T[]>): string {
  if (result.state === "loading") return "…";
  if (result.state === "error") return "unavailable";
  return String(result.value.length);
}

export default function AdminHome() {
  const [account, setAccount] = useState<{ email: string } | null>(null);
  const verifications = useCases<VerificationCase>(
    "/api/verifications",
    "The verification queue could not be loaded.",
  );
  const safety = useCases<SafetyCase>(
    "/api/safety",
    "The safety queue could not be loaded.",
  );
  const care = useCases<CareCase>(
    "/api/care",
    "The care queue could not be loaded.",
  );

  useEffect(() => {
    const controller = new AbortController();
    void fetch("/api/account", { cache: "no-store", signal: controller.signal })
      .then(async (response) => {
        const body = (await response.json().catch(() => null)) as {
          email?: string;
        } | null;
        if (response.ok && body?.email) setAccount({ email: body.email });
      })
      .catch(() => undefined);
    return () => controller.abort();
  }, []);

  const loadErrors = [verifications, safety, care].filter(
    (result): result is { state: "error"; message: string } =>
      result.state === "error",
  );

  const openSafety =
    safety.state === "ready"
      ? safety.value.filter((item) => item.status !== "resolved")
      : [];
  const openCare =
    care.state === "ready"
      ? care.value.filter((item) => item.status !== "resolved")
      : [];
  const queuedVerifications =
    verifications.state === "ready"
      ? verifications.value.filter(
          (item) => !item.status || item.status === "queued_manual",
        )
      : [];
  const oldestSubmitted = queuedVerifications.reduce<string | null>(
    (oldest, item) =>
      !oldest || item.submittedAt < oldest ? item.submittedAt : oldest,
    null,
  );
  const initials = account?.email
    ? account.email
        .split("@")[0]
        .split(/[._-]+/)
        .filter(Boolean)
        .slice(0, 2)
        .map((part) => part[0])
        .join("")
        .toUpperCase() || "?"
    : "–";

  return (
    <Box component="main">
      <Suspense>
        <TourDialog />
      </Suspense>
      <Container maxWidth={false} className="admin-shell">
        <Box component="header" className="admin-header">
          <Box>
            <Typography className="date-line">
              {new Date().toLocaleDateString("en-GH", {
                weekday: "long",
                day: "numeric",
                month: "long",
              })}
            </Typography>
            <Typography component="h1">{greeting()}</Typography>
            <Typography>Here is what needs a human pair of eyes.</Typography>
          </Box>
          <Stack direction="row" spacing={1.25} sx={{ alignItems: "center" }}>
            <Button className="search-button" href="/verification">
              ⌕ Search cases
            </Button>
            <HandoverButton />
            <Avatar className="header-avatar">{initials}</Avatar>
          </Stack>
        </Box>

        {loadErrors.length > 0 ? (
          <Alert severity="warning" sx={{ mb: 3 }}>
            {loadErrors.map((error) => error.message).join(" ")}
          </Alert>
        ) : null}

        <Box className="metrics-grid">
          <MetricCard
            label="Waiting for verification"
            value={
              verifications.state === "ready"
                ? String(queuedVerifications.length)
                : metricValue(verifications)
            }
            note={
              verifications.state === "error"
                ? "queue unavailable — fail closed"
                : oldestSubmitted
                  ? `oldest waiting ${waitLabel(oldestSubmitted)}`
                  : "live queue"
            }
            accent="#FF9F1C"
            href="/verification"
          />
          <MetricCard
            label="Open safety cases"
            value={
              safety.state === "ready"
                ? String(openSafety.length)
                : metricValue(safety)
            }
            note={
              safety.state === "error"
                ? "queue unavailable — fail closed"
                : `${openSafety.filter((item) => item.tier === "A").length} tier A`
            }
            accent="#FF4D6D"
            href="/safety"
          />
          <MetricCard
            label="Care follow-ups"
            value={
              care.state === "ready"
                ? String(openCare.length)
                : metricValue(care)
            }
            note={
              care.state === "error"
                ? "queue unavailable — fail closed"
                : `${openCare.filter((item) => item.status === "open").length} awaiting engagement`
            }
            accent="#12A67C"
            href="/care"
          />
        </Box>

        <Box className="work-grid">
          <Card className="queue-panel">
            <Box className="panel-heading">
              <Box>
                <Typography className="section-kicker">
                  Verification desk
                </Typography>
                <Typography component="h2">
                  People waiting for review
                </Typography>
              </Box>
              <Button href="/verification">Open full queue ↗</Button>
            </Box>

            <Box className="queue-list">
              {verifications.state === "ready" &&
              queuedVerifications.length === 0 ? (
                <Typography sx={{ color: "text.secondary", p: 2 }}>
                  Nobody is waiting for review right now.
                </Typography>
              ) : null}
              {queuedVerifications.slice(0, 5).map((item) => (
                <Box className="queue-row" key={item.caseId}>
                  <Box className="queue-person">
                    <Typography sx={{ fontWeight: 800 }}>
                      {item.subjectRef}
                    </Typography>
                    <Typography>
                      {item.caseId} · {reasonLabels[item.reasonCode]}
                    </Typography>
                  </Box>
                  <Typography className="wait-time">
                    {waitLabel(item.submittedAt)}
                  </Typography>
                  <Chip className="tone-gold" label="Queued" />
                  <Button className="review-button" href="/verification">
                    Review
                  </Button>
                </Box>
              ))}
            </Box>
          </Card>

          <Card className="sla-panel">
            <Box className="panel-heading compact">
              <Box>
                <Typography className="section-kicker">
                  Today’s response
                </Typography>
                <Typography component="h2">SLA pulse</Typography>
              </Box>
            </Box>
            <Alert severity="info" sx={{ mt: 1 }}>
              Response-time evidence is not composed into this command centre.
              This page does not invent substitute percentages.
            </Alert>
            <Button className="plain-action" href="/analytics">
              Open analytics desk
            </Button>
          </Card>
        </Box>

        <Card className="incident-panel">
          <Box className="panel-heading">
            <Box>
              <Typography className="section-kicker">
                Trust &amp; safety
              </Typography>
              <Typography component="h2">Open cases</Typography>
            </Box>
            <Link href="/safety">
              <Button>Open safety desk</Button>
            </Link>
          </Box>
          <Box
            className="incident-table"
            role="table"
            aria-label="Open trust and safety cases"
          >
            <Box className="incident-head" role="row">
              <span>Case</span>
              <span>Tier</span>
              <span>Queue</span>
              <span>Status</span>
              <span>SLA due</span>
              <span />
            </Box>
            {safety.state === "ready" && openSafety.length === 0 ? (
              <Typography sx={{ color: "text.secondary", p: 2 }}>
                No open safety cases.
              </Typography>
            ) : null}
            {openSafety.slice(0, 5).map((item) => (
              <Box className="incident-row" role="row" key={item.caseId}>
                <strong>{item.caseId}</strong>
                <span>Tier {item.tier}</span>
                <span>{item.queue}</span>
                <span>{item.status.replace("_", " ")}</span>
                <Chip
                  label={new Date(item.slaDueAt).toLocaleString("en-GH", {
                    hour: "2-digit",
                    minute: "2-digit",
                    day: "numeric",
                    month: "short",
                  })}
                  className="tone-pink"
                />
                <Button aria-label={`Open ${item.caseId}`} href="/safety">
                  →
                </Button>
              </Box>
            ))}
          </Box>
        </Card>
      </Container>
    </Box>
  );
}
