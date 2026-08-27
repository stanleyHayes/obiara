import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import { AdminCard } from "../../admin-card";

const evidence = [
  {
    title: "Circle and fire state",
    detail:
      "Current density, capacity, schedule and source versions must come from the community authorities.",
  },
  {
    title: "Host eligibility",
    detail:
      "Verification, training, Suban vetting and certification must all be current and versioned.",
  },
  {
    title: "Participant notice",
    detail:
      "The approved template, locale, audience and content digest must be resolved before preview.",
  },
  {
    title: "Bounded reason",
    detail:
      "Only host unavailable, certification expired, safety capacity or schedule conflict is accepted.",
  },
  {
    title: "Human acknowledgement",
    detail:
      "The exact notice is acknowledged only after every source version is revalidated.",
  },
] as const;

export function CommunityDesk() {
  return (
    <Box
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="lg">
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
              COMMUNITY OPERATIONS
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
              Evidence before intervention.
            </Typography>
            <Typography sx={{ color: "text.secondary", maxWidth: 720, mt: 2 }}>
              Community changes stay closed until the runtime can prove the
              circle, fire, host and participant-notice evidence together.
            </Typography>
          </Box>
          <Button href="/" variant="outlined">
            Back to command centre
          </Button>
        </Stack>

        <Alert severity="warning" sx={{ mb: 3 }}>
          Operations proposals are unavailable. Host certification and
          participant-notice authorities are not yet composed into the API
          runtime.
        </Alert>

        <AdminCard
          variant="policy"
          watermark="evidence"
          sx={{ borderRadius: 1, p: { xs: 2.5, md: 4 } }}
        >
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={2}
            sx={{ justifyContent: "space-between" }}
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
                REQUIRED EVIDENCE CHAIN
              </Typography>
              <Typography
                component="h2"
                sx={{ fontSize: { xs: 30, md: 42 }, fontWeight: 800, mt: 1 }}
              >
                Five checks. One bounded proposal.
              </Typography>
            </Box>
            <Chip
              color="warning"
              label="Fail closed"
              sx={{ alignSelf: "flex-start" }}
            />
          </Stack>

          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: "1fr",
              mt: 3,
            }}
          >
            {evidence.map((item, index) => (
              <Box
                key={item.title}
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 1,
                  p: 2.5,
                }}
              >
                <Typography
                  sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800 }}
                >
                  {String(index + 1).padStart(2, "0")}
                </Typography>
                <Typography sx={{ fontSize: 20, fontWeight: 800, mt: 0.5 }}>
                  {item.title}
                </Typography>
                <Typography
                  sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}
                >
                  {item.detail}
                </Typography>
              </Box>
            ))}
          </Box>
        </AdminCard>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: "1fr",
            mt: 3,
          }}
        >
          <AdminCard
            variant="policy"
            watermark="evidence"
            sx={{ borderRadius: 1, p: 3 }}
          >
            <Typography sx={{ fontSize: 24, fontWeight: 800 }}>
              What readiness means
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.7, mt: 1 }}
            >
              A successful proposal is ready for human review only. It does not
              assign a host, cancel a fire, change a circle or send a
              notification.
            </Typography>
          </AdminCard>
          <AdminCard
            variant="panel"
            watermark="queue"
            sx={{ borderRadius: 1, p: 3 }}
          >
            <Typography sx={{ fontSize: 24, fontWeight: 800 }}>
              Available controls
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.7, mt: 1 }}
            >
              Use runtime controls for governed feature changes, or the safety
              queue for an active safety case.
            </Typography>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              spacing={1.5}
              sx={{ mt: 2 }}
            >
              <Button href="/controls" variant="contained">
                Runtime controls
              </Button>
              <Button href="/safety" variant="outlined">
                Safety queue
              </Button>
            </Stack>
          </AdminCard>
        </Box>
      </Container>
    </Box>
  );
}
