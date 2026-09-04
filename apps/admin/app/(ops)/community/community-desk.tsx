import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";

const evidence = [
  {
    code: "CIRCLE",
    title: "Circle and fire state",
    detail:
      "Current density, capacity, schedule and source versions must come from the community authorities.",
    icon: "circles",
  },
  {
    code: "HOST",
    title: "Host eligibility",
    detail:
      "Verification, training, Suban vetting and certification must all be current and versioned.",
    icon: "community",
  },
  {
    code: "NOTICE",
    title: "Participant notice",
    detail:
      "The approved template, locale, audience and content digest must be resolved before preview.",
    icon: "waitlist",
  },
  {
    code: "REASON",
    title: "Bounded reason",
    detail:
      "Only host unavailable, certification expired, safety capacity or schedule conflict is accepted.",
    icon: "controls",
  },
  {
    code: "HUMAN",
    title: "Human acknowledgement",
    detail:
      "The exact notice is acknowledged only after every source version is revalidated.",
    icon: "verification",
  },
] as const;

export function CommunityDesk() {
  return (
    <main className="community-redesign">
      <Box className="community-shell">
        <header className="community-hero">
          <Box className="community-hero-copy">
            <Button href="/" className="community-back">
              Return to command centre
            </Button>
            <Box className="community-kicker">
              <AdminIcon name="circles" aria-hidden="true" />
              <Typography className="section-kicker">
                Circles &amp; hosts · community desk
              </Typography>
            </Box>
            <Typography component="h1">
              Hold the circle before you move it.
            </Typography>
            <Typography className="community-hero-intro">
              One place to understand the host, gathering, capacity and notice
              evidence that must agree before an operator can propose change.
            </Typography>
          </Box>
          <Box className="community-hero-register" aria-label="Desk status">
            <Box>
              <span>Operating mode</span>
              <strong>Evidence only</strong>
            </Box>
            <Box>
              <span>Proposal authority</span>
              <strong>Not composed</strong>
            </Box>
            <Box>
              <span>Decision owner</span>
              <strong>Human operator</strong>
            </Box>
          </Box>
          <AdminCardWatermark watermark="identity" />
        </header>

        <section
          className="community-boundary"
          aria-labelledby="boundary-title"
        >
          <span className="community-boundary-icon">
            <UtilityIcon name="security" aria-hidden="true" />
          </span>
          <Box>
            <Typography className="section-kicker">Current boundary</Typography>
            <Typography id="boundary-title" component="h2">
              Observe now. Intervene only when the authorities are connected.
            </Typography>
            <Typography>
              Operations proposals are unavailable. Host certification and
              participant-notice authorities are not yet composed into the API
              runtime.
            </Typography>
          </Box>
          <span className="community-boundary-state">Fail closed</span>
        </section>

        <section className="community-relay" aria-labelledby="relay-title">
          <Box className="community-section-heading">
            <Box>
              <Typography className="section-kicker">
                Required evidence relay
              </Typography>
              <Typography id="relay-title" component="h2">
                Five checks. One bounded proposal.
              </Typography>
            </Box>
            <Typography>
              Every handoff stays traceable. A missing or stale source stops the
              chain before preview.
            </Typography>
          </Box>
          <ol className="community-evidence-list">
            {evidence.map((item, index) => (
              <li className="community-evidence-step" key={item.title}>
                <span className="community-step-number">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <span className="community-step-icon">
                  <AdminIcon name={item.icon} aria-hidden="true" />
                </span>
                <Box className="community-step-copy">
                  <span>{item.code}</span>
                  <Typography component="h3">{item.title}</Typography>
                  <Typography>{item.detail}</Typography>
                </Box>
                <span className="community-step-state">Required</span>
              </li>
            ))}
          </ol>
        </section>

        <Box className="community-decision-grid">
          <AdminCard
            variant="policy"
            watermark="evidence"
            className="community-readiness"
          >
            <Typography className="section-kicker">Readiness means</Typography>
            <Typography component="h2">
              A review packet—not an action.
            </Typography>
            <Typography>
              A successful proposal is ready for human review only. It does not
              assign a host, cancel a fire, change a circle or send a
              notification.
            </Typography>
            <Box
              className="community-not-actions"
              aria-label="Excluded actions"
            >
              <span>No host assignment</span>
              <span>No circle change</span>
              <span>No participant notice</span>
            </Box>
          </AdminCard>
          <AdminCard
            variant="panel"
            watermark="queue"
            className="community-handoff"
          >
            <Box className="community-handoff-icon">
              <AdminIcon name="operations" aria-hidden="true" />
            </Box>
            <Typography className="section-kicker">Operator handoff</Typography>
            <Typography component="h2">
              Use the authority that already exists.
            </Typography>
            <Typography>
              Govern feature behavior through runtime controls. Send active harm
              or safeguarding concerns to the safety desk.
            </Typography>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={1.25}>
              <Button href="/controls" variant="contained">
                Open runtime controls
              </Button>
              <Button href="/safety" variant="outlined">
                Open safety desk
              </Button>
            </Stack>
          </AdminCard>
        </Box>
        <Alert className="community-footnote" severity="info">
          This desk represents evidence readiness only; it never simulates a
          host decision or community intervention in the browser.
        </Alert>
      </Box>
    </main>
  );
}
