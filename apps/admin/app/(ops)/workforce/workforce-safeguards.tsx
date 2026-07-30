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

const safeguards = [
  [
    "Category before content",
    "Assignment previews identify the exposure category only. Evidence remains closed until a reviewer explicitly accepts through an authorized queue.",
  ],
  [
    "Protected breaks",
    "A break must remove new assignments and cannot lower a performance score, because moderation exposure is not a productivity target.",
  ],
  [
    "No-penalty opt-out",
    "A reviewer may decline graphic or personally unsafe material without retaliation, ranking, or hidden performance flags.",
  ],
  [
    "Bounded exposure",
    "Shift duration and sensitive-evidence exposure need enforceable limits set with occupational-health review, not invented client counters.",
  ],
  [
    "Human support",
    "Supervisor and counselling support must be confidential and must not create a diagnosis or employment inference in product telemetry.",
  ],
  [
    "Rotation and review",
    "High-exposure categories require rotation and a welfare check; incidents involving graphic evidence trigger reviewer follow-up.",
  ],
] as const;

export function WorkforceSafeguards() {
  return (
    <main className="verification-shell workforce-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Moderation workforce safeguards
          </Typography>
          <Typography component="h1">
            The work must not consume the worker.
          </Typography>
          <Typography>
            Policy guardrails for evidence-facing teams. No shift, assignment,
            exposure, break, opt-out, support, HR, or counselling system is
            composed in this application.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip color="success" label="No productivity score" />
          <Chip label="No hidden surveillance" variant="outlined" />
        </Stack>
      </header>

      <Alert severity="warning" className="verification-alert">
        This page is guidance, not a staffing console. Use the approved
        workforce system and supervisor channel for protected breaks, assignment
        refusal, rotations, and support. Do not record health information in
        case notes.
      </Alert>

      <Box
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2,1fr)" },
        }}
      >
        {safeguards.map(([title, description], index) => (
          <Card key={title} variant="outlined" sx={{ p: 2.5 }}>
            <Stack direction="row" spacing={1.5}>
              <Typography
                sx={{ color: "primary.main", fontSize: 22, fontWeight: 900 }}
              >
                {String(index + 1).padStart(2, "0")}
              </Typography>
              <Box>
                <Typography
                  component="h2"
                  sx={{ fontSize: 20, fontWeight: 800 }}
                >
                  {title}
                </Typography>
                <Typography sx={{ color: "text.secondary", mt: 0.75 }}>
                  {description}
                </Typography>
              </Box>
            </Stack>
          </Card>
        ))}
      </Box>

      <Card className="workforce-preview" sx={{ mt: 3 }}>
        <Box>
          <Typography className="section-kicker">Evidence boundary</Typography>
          <Typography component="h2">
            Open content only inside an assigned case.
          </Typography>
          <Typography>
            The safety queue owns assignment and least-exposure access. A
            fresh-MFA, purpose-audited view is the only composed path to
            retained evidence.
          </Typography>
        </Box>
        <Button component={Link} href="/safety" variant="contained">
          Open safety queue
        </Button>
      </Card>

      <Alert severity="info" sx={{ mt: 3 }}>
        Before production staffing, the external workforce authority must prove
        category-only preview, enforceable exposure ceilings, protected breaks,
        no-penalty refusal, confidential support escalation, rotation, and
        anti-retaliation audit. Until then, this product does not claim those
        controls were applied.
      </Alert>
    </main>
  );
}
