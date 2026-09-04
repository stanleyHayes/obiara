"use client";

import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";

const dimensions = [
  [
    "Harassment response",
    "Evidence that reports receive timely, proportionate human handling.",
  ],
  [
    "Coercion resistance",
    "Evidence that product and operations controls resist financial or relationship coercion.",
  ],
  [
    "Privacy control",
    "Evidence that members can understand and exercise purpose-bound privacy choices.",
  ],
  [
    "Reporting access",
    "Evidence that reporting remains reachable across supported journeys.",
  ],
  [
    "Care access",
    "Evidence that care remains separate from enforcement and offers approved resources.",
  ],
] as const;

const gaps = [
  "Cohort below reviewed minimum",
  "Response rate below reviewed minimum",
  "Configured dimension missing",
  "Evidence count below dimension minimum",
  "Women-reviewer coverage incomplete",
  "Observed gap acknowledgement incomplete",
] as const;

export function MpanyimfoDocket() {
  return (
    <main className="verification-shell mpanyimfo-shell mpanyimfo-redesign">
      <header className="verification-header mpanyimfo-hero">
        <Box className="mpanyimfo-hero-copy">
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Box className="mpanyimfo-hero-kicker">
            <AdminIcon name="governance" aria-hidden="true" />
            <Typography className="section-kicker">
              Mpanyimfo · evidence council
            </Typography>
          </Box>
          <Typography component="h1">
            Review with weight. Decide with limits.
          </Typography>
          <Typography>
            Women-led review of redacted cohort evidence against a current,
            versioned definition. Never a member score or case docket.
          </Typography>
        </Box>
        <Box className="mpanyimfo-hero-status">
          <Box>
            <span>Evidence level</span>
            <strong>Cohort only</strong>
          </Box>
          <Box>
            <span>Decision type</span>
            <strong>Neutral readiness</strong>
          </Box>
        </Box>
        <AdminCardWatermark watermark="evidence" />
      </header>

      <section
        className="mpanyimfo-outcome-boundary"
        aria-label="Implemented outcomes"
      >
        <span className="mpanyimfo-boundary-icon">
          <AdminIcon name="verification" aria-hidden="true" />
        </span>
        <Box>
          <Typography className="section-kicker">
            Two bounded outcomes
          </Typography>
          <Typography component="h2">
            <strong>evidence incomplete</strong>
            <i />
            or<strong>ready for release review</strong>
          </Typography>
          <Typography>
            Neither outcome releases a market, blocks a member, changes an
            account, scores a person or decides an appeal.
          </Typography>
        </Box>
      </section>

      <AdminCard
        variant="policy"
        watermark="evidence"
        className="mpanyimfo-record mpanyimfo-policy"
      >
        <Box className="verification-panel-heading">
          <Box>
            <Typography className="section-kicker">
              Current policy shape
            </Typography>
            <Typography component="h2">
              Representative evidence needs complete coverage.
            </Typography>
          </Box>
          <span className="mpanyimfo-version">
            <UtilityIcon name="security" aria-hidden="true" />
            Version-pinned
          </span>
        </Box>
        <Typography className="mpanyimfo-policy-copy">
          Every assessment revalidates the current reviewed definition, exact
          aggregate version, reviewer authority, configured dimensions, minimum
          cohort and response thresholds, and acknowledgement of every observed
          gap before recording a result.
        </Typography>
      </AdminCard>

      <Box className="mpanyimfo-evidence-grid">
        <AdminCard
          className="mpanyimfo-dimensions"
          variant="panel"
          watermark="evidence"
        >
          <Typography className="section-kicker">
            Reviewed dimensions
          </Typography>
          <Typography component="h2">What the aggregate may cover</Typography>
          <Stack spacing={1.25}>
            {dimensions.map(([title, description], index) => (
              <Box className="mpanyimfo-dimension" key={title}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                <Box>
                  <Typography component="h3">{title}</Typography>
                  <Typography>{description}</Typography>
                </Box>
              </Box>
            ))}
          </Stack>
        </AdminCard>
        <AdminCard
          className="mpanyimfo-gaps"
          variant="warning"
          watermark="safety"
        >
          <Typography className="section-kicker">
            Fail-closed gap codes
          </Typography>
          <Typography component="h2">
            Why evidence remains incomplete
          </Typography>
          <Stack spacing={1}>
            {gaps.map((gap, index) => (
              <Box className="mpanyimfo-gap" key={gap}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                <Typography>{gap}</Typography>
              </Box>
            ))}
          </Stack>
        </AdminCard>
      </Box>

      <Box className="mpanyimfo-grid mpanyimfo-boundaries">
        <Alert className="mpanyimfo-representable" severity="success">
          <strong>Representable:</strong> bounded aggregate counts, versioned
          definitions, complete dimension coverage, opaque cohort/reviewer keys,
          reviewed gaps, and a neutral readiness assessment.
        </Alert>
        <Alert className="mpanyimfo-never" severity="warning">
          <strong>Never representable:</strong> raw content, identity, member
          scores, subgroup microdata, hidden ranks, vendor/model decisions,
          automatic enforcement, release authority, or adjudication.
        </Alert>
      </Box>

      <AdminCard
        variant="warning"
        watermark="safety"
        className="mpanyimfo-record mpanyimfo-authority"
      >
        <Typography className="section-kicker">
          Uncomposed panel authority
        </Typography>
        <Typography component="h2">
          Dockets and appeals need their own system of record.
        </Typography>
        <Typography className="mpanyimfo-authority-copy">
          No persisted docket, conflict declaration, panel-seat authority, vote,
          ruling, member appeal, separate appeal panel, or immutable ruling
          record is composed. Those capabilities remain unavailable instead of
          being simulated in this browser.
        </Typography>
        <Button href="/safety" variant="outlined">
          Open the real safety queue
        </Button>
      </AdminCard>
    </main>
  );
}
