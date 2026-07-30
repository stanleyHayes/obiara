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
    <main className="verification-shell mpanyimfo-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Mpanyimfo · evidence boundary
          </Typography>
          <Typography component="h1">
            Women-led review without invented authority.
          </Typography>
          <Typography>
            The implemented service evaluates redacted cohort evidence against a
            current, versioned definition and substantive women-reviewer
            approval. It is not a case docket.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip color="success" label="Cohort-level only" />
          <Chip label="Neutral outcomes" variant="outlined" />
        </Stack>
      </header>

      <Alert severity="info" className="verification-alert">
        The only implemented outcomes are <strong>evidence incomplete</strong>{" "}
        and
        <strong> ready for release review</strong>. Neither outcome releases a
        market, blocks a member, changes an account, scores a person, or decides
        an appeal.
      </Alert>

      <Card className="mpanyimfo-record">
        <Box className="verification-panel-heading">
          <Box>
            <Typography className="section-kicker">
              Current policy shape
            </Typography>
            <Typography component="h2">
              Representative evidence needs complete coverage.
            </Typography>
          </Box>
          <Chip label="Version-pinned" />
        </Box>
        <Typography sx={{ color: "text.secondary", maxWidth: 820 }}>
          Every assessment revalidates the current reviewed definition, exact
          aggregate version, reviewer authority, configured dimensions, minimum
          cohort and response thresholds, and acknowledgement of every observed
          gap before recording a result.
        </Typography>
      </Card>

      <Box
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2,1fr)" },
          mt: 3,
        }}
      >
        <Card sx={{ p: 3 }}>
          <Typography className="section-kicker">
            Reviewed dimensions
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
          >
            What the aggregate may cover
          </Typography>
          <Stack spacing={1.25}>
            {dimensions.map(([title, description]) => (
              <Box
                key={title}
                sx={{
                  borderBottom: "1px solid",
                  borderColor: "divider",
                  pb: 1.25,
                }}
              >
                <Typography sx={{ fontWeight: 800 }}>{title}</Typography>
                <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                  {description}
                </Typography>
              </Box>
            ))}
          </Stack>
        </Card>
        <Card sx={{ p: 3 }}>
          <Typography className="section-kicker">
            Fail-closed gap codes
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 24, fontWeight: 800, mb: 2 }}
          >
            Why evidence remains incomplete
          </Typography>
          <Stack spacing={1}>
            {gaps.map((gap) => (
              <Box
                key={gap}
                sx={{ alignItems: "center", display: "flex", gap: 1.25 }}
              >
                <Chip color="warning" label="gap" size="small" />
                <Typography>{gap}</Typography>
              </Box>
            ))}
          </Stack>
        </Card>
      </Box>

      <Box className="mpanyimfo-grid" sx={{ mt: 3 }}>
        <Alert severity="success">
          <strong>Representable:</strong> bounded aggregate counts, versioned
          definitions, complete dimension coverage, opaque cohort/reviewer keys,
          reviewed gaps, and a neutral readiness assessment.
        </Alert>
        <Alert severity="warning">
          <strong>Never representable:</strong> raw content, identity, member
          scores, subgroup microdata, hidden ranks, vendor/model decisions,
          automatic enforcement, release authority, or adjudication.
        </Alert>
      </Box>

      <Card className="mpanyimfo-record" sx={{ mt: 3 }}>
        <Typography className="section-kicker">
          Uncomposed panel authority
        </Typography>
        <Typography component="h2" sx={{ fontSize: 24, fontWeight: 800 }}>
          Dockets and appeals need their own system of record.
        </Typography>
        <Typography sx={{ color: "text.secondary", my: 1.5 }}>
          No persisted docket, conflict declaration, panel-seat authority, vote,
          ruling, member appeal, separate appeal panel, or immutable ruling
          record is composed. Those capabilities remain unavailable instead of
          being simulated in this browser.
        </Typography>
        <Button component={Link} href="/safety" variant="outlined">
          Open the real safety queue
        </Button>
      </Card>
    </main>
  );
}
