"use client";

import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";

const severities = [
  [
    "SEV-1",
    "15 minutes · 24/7",
    "Physical safety risk, Tier-A abuse in progress, or C4 data breach",
  ],
  [
    "SEV-2",
    "1 hour",
    "Live product-law violation, privileged misuse, or safety-provider compromise",
  ],
  ["SEV-3", "4 hours", "Degraded core journey without data exposure"],
  ["SEV-4", "Next business day", "Operational anomaly without member impact"],
] as const;

const responseFlow = [
  [
    "Detect",
    "Use retained safety queues, Sentinel signals, golden-path monitors, provider budgets, reports, or staff escalation.",
  ],
  [
    "Declare",
    "A human on-call owner declares severity and opens the external incident channel. SEV-1/2 requires safety-lead and DPO escalation.",
  ],
  [
    "Contain",
    "Use the runtime control desk to bound Sow, Fires, AI, Payments, or Gate. Disable affected adapters before widening impact.",
  ],
  [
    "Preserve",
    "Establish legal hold through the authorized privacy process before cleanup. Never rotate or flush affected evidence.",
  ],
  [
    "Mitigate",
    "Restore the golden path and verify it with synthetic monitoring before removing containment.",
  ],
  [
    "Communicate",
    "State what happened and what members should do. Use no-blame, no-pressure language.",
  ],
  [
    "Review",
    "Complete a blameless SEV-1/2 review within five business days and retain actions in the execution ledger.",
  ],
] as const;

const clock = [
  [
    "T+0",
    "Declare incident; page the appointed DPO and safety lead; begin containment.",
  ],
  [
    "T+2h",
    "Draft breach scope: data classes, estimated affected population, and systems.",
  ],
  [
    "T+24h",
    "Complete preservation and inform the CERT-GH liaison where applicable.",
  ],
  [
    "T+48h",
    "Record member-impact assessment and notification decision with reasons.",
  ],
  [
    "T+72h",
    "Human legal owner files with the Data Protection Commission and CERT-GH, or files an interim report if scope remains incomplete.",
  ],
] as const;

export function IncidentRunbook() {
  return (
    <main className="verification-shell incident-shell incident-redesign">
      <header className="verification-header incident-hero">
        <Box className="incident-hero-copy">
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Box className="incident-hero-kicker">
            <AdminIcon name="incidents" aria-hidden="true" />
            <Typography className="section-kicker">Incident command</Typography>
          </Box>
          <Typography component="h1">Move with clarity.</Typography>
          <Typography>
            Protect people, preserve evidence and coordinate the human response
            from one operational baseline.
          </Typography>
        </Box>
        <Box className="incident-hero-meta">
          <Box>
            <span>Runbook</span>
            <strong>v0</strong>
          </Box>
          <Box>
            <span>Prepared</span>
            <strong>27 Jul 2026</strong>
          </Box>
          <Box className="incident-baseline">
            <span>State</span>
            <strong>Pre-P0 baseline</strong>
          </Box>
        </Box>
        <AdminCardWatermark watermark="safety" />
      </header>

      <section
        className="incident-priority"
        aria-labelledby="incident-priority-title"
      >
        <span className="incident-priority-mark">
          <AdminIcon name="safety" aria-hidden="true" />
        </span>
        <Box className="incident-priority-copy">
          <Typography className="section-kicker">First principle</Typography>
          <Typography component="h2" id="incident-priority-title">
            People before platform.
          </Typography>
          <Typography>
            Member physical safety comes first, followed by dignity and privacy,
            platform integrity, then growth. Reasonable suspicion of a
            personal-data breach starts the clock—certainty is not required.
          </Typography>
        </Box>
      </section>

      <Box
        className="incident-severity-grid"
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: "1fr",
          mb: 3,
        }}
      >
        {severities.map(([severity, response, definition]) => (
          <AdminCard
            key={severity}
            variant="warning"
            watermark="safety"
            className={`incident-severity-card incident-severity-card--${severity.toLowerCase()}`}
          >
            <Box className="incident-severity-code">
              <span>Severity</span>
              <strong>{severity.slice(-1)}</strong>
            </Box>
            <Box className="incident-severity-copy">
              <Typography component="h2">{severity}</Typography>
              <Typography>{definition}</Typography>
            </Box>
            <Box className="incident-severity-time">
              <UtilityIcon name="clock" aria-hidden="true" />
              <span>Respond within</span>
              <strong>{response}</strong>
            </Box>
          </AdminCard>
        ))}
      </Box>

      <Box className="incident-runbook-grid">
        <AdminCard
          className="incident-flow-panel"
          variant="timeline"
          watermark="queue"
        >
          <Box className="verification-panel-heading">
            <Box>
              <Typography className="section-kicker">
                Seven movements
              </Typography>
              <Typography component="h2">Ordered response</Typography>
            </Box>
            <Box className="incident-human-owned">
              <AdminIcon name="operators" aria-hidden="true" />
              <span>Human-owned</span>
            </Box>
          </Box>
          <Stack spacing={1.25}>
            {responseFlow.map(([title, detail], index) => (
              <Box className="incident-flow-step" key={title}>
                <span className="incident-step-number">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <Box>
                  <Typography component="h3">{title}</Typography>
                  <Typography>{detail}</Typography>
                </Box>
              </Box>
            ))}
          </Stack>
        </AdminCard>

        <AdminCard
          className="incident-clock-panel"
          variant="timeline"
          watermark="clock"
        >
          <Box className="verification-panel-heading">
            <Box>
              <Typography className="section-kicker">
                Regulatory clock
              </Typography>
              <Typography component="h2">72 hours</Typography>
            </Box>
            <span className="incident-clock-trigger">
              Starts at reasonable suspicion
            </span>
          </Box>
          <Stack spacing={1.25}>
            {clock.map(([time, action]) => (
              <Box className="incident-clock-step" key={time}>
                <Typography component="strong">{time}</Typography>
                <Typography>{action}</Typography>
              </Box>
            ))}
          </Stack>
        </AdminCard>
      </Box>

      <AdminCard
        className="incident-packet incident-control-panel"
        variant="policy"
        watermark="evidence"
      >
        <Box>
          <Box className="incident-control-icon">
            <AdminIcon name="controls" aria-hidden="true" />
          </Box>
          <Typography className="section-kicker">
            Available control surfaces
          </Typography>
          <Typography component="h2">
            Move through existing authorities.
          </Typography>
          <Typography>
            Containment, safety review, and care coordination have dedicated
            authenticated workflows. Evidence remains least-exposure and
            purpose-audited.
          </Typography>
        </Box>
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
          <Button href="/controls" variant="contained">
            Runtime controls
          </Button>
          <Button href="/safety" variant="outlined">
            Safety queue
          </Button>
          <Button href="/care" variant="outlined">
            Care queue
          </Button>
        </Stack>
      </AdminCard>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: "1fr",
          mt: 3,
        }}
      >
        <Alert className="incident-boundary-note" severity="info">
          Evidence access belongs in the assigned safety case and requires a
          declared triage, appeal, or legal purpose plus fresh MFA. This runbook
          does not expose evidence.
        </Alert>
        <Alert className="incident-boundary-note" severity="warning">
          Pager dispatch, legal holds, regulatory packet generation/submission,
          external incident channels, and incident closure remain external or
          uncomposed authorities. Record them in the approved systems of record;
          this page never simulates completion.
        </Alert>
      </Box>
    </main>
  );
}
