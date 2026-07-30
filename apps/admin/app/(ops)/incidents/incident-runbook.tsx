"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";

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
    <main className="verification-shell incident-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Incident response · runbook v0
          </Typography>
          <Typography component="h1">
            Protect people. Preserve evidence. Coordinate humans.
          </Typography>
          <Typography>
            Operational guidance from the checked-in pre-P0 baseline. This page
            does not declare incidents, page staff, place holds, contact
            regulators, or close records.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip color="warning" label="Pre-P0 baseline" />
          <Chip label="Prepared 27 Jul 2026" variant="outlined" />
        </Stack>
      </header>

      <Alert severity="error" className="verification-alert">
        Priority in every conflict: member physical safety, member dignity and
        privacy, platform integrity, then growth. A confirmed or reasonably
        suspected personal-data breach starts the reporting clock; certainty is
        not required.
      </Alert>

      <Box
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2,1fr)" },
          mb: 3,
        }}
      >
        {severities.map(([severity, response, definition]) => (
          <Card key={severity} variant="outlined" sx={{ p: 2.5 }}>
            <Stack
              direction="row"
              spacing={1}
              sx={{ alignItems: "center", justifyContent: "space-between" }}
            >
              <Typography component="h2" sx={{ fontSize: 24, fontWeight: 800 }}>
                {severity}
              </Typography>
              <Chip
                label={response}
                color={
                  severity === "SEV-1"
                    ? "error"
                    : severity === "SEV-2"
                      ? "warning"
                      : "default"
                }
                size="small"
              />
            </Stack>
            <Typography sx={{ color: "text.secondary", mt: 1 }}>
              {definition}
            </Typography>
          </Card>
        ))}
      </Box>

      <Box className="incident-runbook-grid">
        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Ordered response</Typography>
            <Chip label="Human-owned" variant="outlined" />
          </Box>
          <Stack spacing={1.25}>
            {responseFlow.map(([title, detail], index) => (
              <Box
                key={title}
                sx={{
                  borderBottom: "1px solid",
                  borderColor: "divider",
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: "32px minmax(0,1fr)",
                  pb: 1.25,
                }}
              >
                <Typography sx={{ color: "primary.main", fontWeight: 900 }}>
                  {index + 1}
                </Typography>
                <Box>
                  <Typography sx={{ fontWeight: 800 }}>{title}</Typography>
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {detail}
                  </Typography>
                </Box>
              </Box>
            ))}
          </Stack>
        </Card>

        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">72-hour clock</Typography>
            <Chip color="error" label="From reasonable suspicion" />
          </Box>
          <Stack spacing={1.25}>
            {clock.map(([time, action]) => (
              <Box key={time} sx={{ bgcolor: "action.hover", p: 1.5 }}>
                <Typography sx={{ color: "#8e3159", fontWeight: 900 }}>
                  {time}
                </Typography>
                <Typography sx={{ fontSize: 13 }}>{action}</Typography>
              </Box>
            ))}
          </Stack>
        </Card>
      </Box>

      <Card className="incident-packet">
        <Box>
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
          <Button component={Link} href="/controls" variant="contained">
            Runtime controls
          </Button>
          <Button component={Link} href="/safety" variant="outlined">
            Safety queue
          </Button>
          <Button component={Link} href="/care" variant="outlined">
            Care queue
          </Button>
        </Stack>
      </Card>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2,1fr)" },
          mt: 3,
        }}
      >
        <Alert severity="info">
          Evidence access belongs in the assigned safety case and requires a
          declared triage, appeal, or legal purpose plus fresh MFA. This runbook
          does not expose evidence.
        </Alert>
        <Alert severity="warning">
          Pager dispatch, legal holds, regulatory packet generation/submission,
          external incident channels, and incident closure remain external or
          uncomposed authorities. Record them in the approved systems of record;
          this page never simulates completion.
        </Alert>
      </Box>
    </main>
  );
}
