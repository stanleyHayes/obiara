"use client";

import { Alert, Box, Button, Typography } from "@mui/material";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";

const safeguards = [
  {
    code: "PREVIEW",
    title: "Category before content",
    detail:
      "Assignment previews identify the exposure category only. Evidence remains closed until a reviewer explicitly accepts through an authorized queue.",
    icon: "verification",
  },
  {
    code: "RECOVERY",
    title: "Protected breaks",
    detail:
      "A break must remove new assignments and cannot lower a performance score, because moderation exposure is not a productivity target.",
    icon: "clock",
  },
  {
    code: "REFUSAL",
    title: "No-penalty opt-out",
    detail:
      "A reviewer may decline graphic or personally unsafe material without retaliation, ranking, or hidden performance flags.",
    icon: "safety",
  },
  {
    code: "CEILING",
    title: "Bounded exposure",
    detail:
      "Shift duration and sensitive-evidence exposure need enforceable limits set with occupational-health review, not invented client counters.",
    icon: "controls",
  },
  {
    code: "SUPPORT",
    title: "Human support",
    detail:
      "Supervisor and counselling support must be confidential and must not create a diagnosis or employment inference in product telemetry.",
    icon: "care",
  },
  {
    code: "ROTATION",
    title: "Rotation and review",
    detail:
      "High-exposure categories require rotation and a welfare check; incidents involving graphic evidence trigger reviewer follow-up.",
    icon: "replay",
  },
] as const;

export function WorkforceSafeguards() {
  return (
    <main className="workforce-redesign">
      <Box className="workforce-shell">
        <header className="workforce-hero">
          <Box className="workforce-hero-copy">
            <Button href="/" className="workforce-back">
              Return to command centre
            </Button>
            <Box className="workforce-kicker">
              <AdminIcon name="workforce" aria-hidden="true" />
              <Typography className="section-kicker">
                Workforce · exposure safeguards
              </Typography>
            </Box>
            <Typography component="h1">
              Protect the person doing the work.
            </Typography>
            <Typography className="workforce-hero-intro">
              Policy guardrails for evidence-facing teams. The product must
              limit exposure, protect refusal and keep care separate from
              performance.
            </Typography>
          </Box>
          <Box
            className="workforce-hero-markers"
            aria-label="Workforce principles"
          >
            <Box>
              <span>01</span>
              <strong>No productivity score</strong>
            </Box>
            <Box>
              <span>02</span>
              <strong>No hidden surveillance</strong>
            </Box>
            <Box>
              <span>03</span>
              <strong>No health inference</strong>
            </Box>
          </Box>
          <AdminCardWatermark watermark="care" />
        </header>

        <section
          className="workforce-guidance"
          aria-labelledby="guidance-title"
        >
          <span className="workforce-guidance-icon">
            <UtilityIcon name="security" aria-hidden="true" />
          </span>
          <Box>
            <Typography className="section-kicker">
              Operating boundary
            </Typography>
            <Typography id="guidance-title" component="h2">
              This page is guidance, not a staffing console.
            </Typography>
            <Typography>
              Use the approved workforce system and supervisor channel for
              protected breaks, assignment refusal, rotations, and support. Do
              not record health information in case notes.
            </Typography>
          </Box>
          <span className="workforce-guidance-state">Policy only</span>
        </section>

        <section
          className="workforce-protections"
          aria-labelledby="protections-title"
        >
          <Box className="workforce-section-heading">
            <Box>
              <Typography className="section-kicker">
                Six worker protections
              </Typography>
              <Typography id="protections-title" component="h2">
                Exposure has a boundary. Refusal has no penalty.
              </Typography>
            </Box>
            <Typography>
              Each protection belongs in the external workforce authority before
              production staffing can begin.
            </Typography>
          </Box>
          <Box className="workforce-protection-list">
            {safeguards.map((safeguard, index) => (
              <article className="workforce-protection" key={safeguard.title}>
                <span className="workforce-protection-number">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <span className="workforce-protection-icon">
                  {safeguard.icon === "replay" || safeguard.icon === "clock" ? (
                    <UtilityIcon name={safeguard.icon} aria-hidden="true" />
                  ) : (
                    <AdminIcon name={safeguard.icon} aria-hidden="true" />
                  )}
                </span>
                <Box>
                  <span className="workforce-protection-code">
                    {safeguard.code}
                  </span>
                  <Typography component="h3">{safeguard.title}</Typography>
                  <Typography>{safeguard.detail}</Typography>
                </Box>
              </article>
            ))}
          </Box>
        </section>

        <AdminCard
          variant="warning"
          watermark="safety"
          className="workforce-exposure-gate"
        >
          <Box className="workforce-gate-copy">
            <Typography className="section-kicker">
              Evidence exposure gate
            </Typography>
            <Typography component="h2">
              Category first. Consent next. Content last.
            </Typography>
            <Typography>
              The safety queue owns assignment and least-exposure access. A
              fresh-MFA, purpose-audited view is the only composed path to
              retained evidence.
            </Typography>
          </Box>
          <Box
            className="workforce-gate-flow"
            aria-label="Required exposure sequence"
          >
            <Box>
              <span>01</span>
              <strong>Category preview</strong>
            </Box>
            <i aria-hidden="true" />
            <Box>
              <span>02</span>
              <strong>Reviewer acceptance</strong>
            </Box>
            <i aria-hidden="true" />
            <Box>
              <span>03</span>
              <strong>Assigned evidence</strong>
            </Box>
          </Box>
        </AdminCard>

        <AdminCard
          variant="policy"
          watermark="evidence"
          className="workforce-preview"
        >
          <Box className="workforce-preview-icon">
            <AdminIcon name="safety" aria-hidden="true" />
          </Box>
          <Box>
            <Typography className="section-kicker">
              Composed access path
            </Typography>
            <Typography component="h2">
              Open content only inside an assigned case.
            </Typography>
            <Typography>
              No shift, assignment, exposure, break, opt-out, support, HR, or
              counselling system is composed in this application.
            </Typography>
          </Box>
          <Button href="/safety" variant="contained">
            Open safety queue
          </Button>
        </AdminCard>

        <Alert severity="info" className="workforce-production-note">
          Before production staffing, the external workforce authority must
          prove category-only preview, enforceable exposure ceilings, protected
          breaks, no-penalty refusal, confidential support escalation, rotation,
          and anti-retaliation audit. Until then, this product does not claim
          those controls were applied.
        </Alert>
      </Box>
    </main>
  );
}
