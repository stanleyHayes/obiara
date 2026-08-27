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

const repositoryEvidence = [
  "Exact candidate and full engineering checks",
  "Security policy and synthetic DAST",
  "Backup and restore orchestration",
  "Rollback and hypercare contract",
] as const;

const externalGates = [
  ["Residency and DPIA", "Founder and DPO/legal"],
  ["Production topology", "Architecture owner"],
  ["Atlas production service", "Procurement and data platform"],
  ["Object storage, CDN and keys", "Procurement and security"],
  ["Live media and egress", "Procurement and realtime owner"],
  ["Transactional communications", "DPO/legal and communications owner"],
  ["Ghana device and network evidence", "QA and platform"],
  ["Production credentials", "Credential custodians"],
  ["Mobile store release", "Mobile release owner and reviewer"],
  ["Real UAT cohort", "UAT lead and safety reviewer"],
  ["Circles, hosts and operational cover", "Launch operations lead"],
  ["Founder go/no-go", "Founder and distinct release reviewer"],
  ["Production activation", "Platform and release authorities"],
] as const;

export function LaunchReadinessDesk() {
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
              LAUNCH READINESS
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
              Repository handoff. Production not cleared.
            </Typography>
            <Typography sx={{ color: "text.secondary", maxWidth: 760, mt: 2 }}>
              This static inventory records repository handoffs prepared for
              staging review. It is not live verification or production
              approval; named authorities must provide current evidence.
            </Typography>
          </Box>
          <Button href="/" variant="outlined">
            Back to command centre
          </Button>
        </Stack>

        <Alert severity="error" sx={{ mb: 3 }}>
          There is no launch action here. Green repository checks cannot approve
          legal posture, purchase providers, create credentials, recruit people
          or activate production.
        </Alert>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: "1fr",
          }}
        >
          <AdminCard
            variant="policy"
            watermark="evidence"
            sx={{
              p: { xs: 2.5, md: 3.5 },
            }}
          >
            <Chip label="Repository-controlled" variant="outlined" />
            <Typography
              component="h2"
              sx={{ fontSize: 34, fontWeight: 800, lineHeight: 1.05, mt: 2 }}
            >
              Evidence ready for staging.
            </Typography>
            <Stack
              component="ul"
              spacing={1.25}
              sx={{ mt: 3, p: 0, listStyle: "none" }}
            >
              {repositoryEvidence.map((item) => (
                <Box
                  component="li"
                  key={item}
                  sx={{
                    borderTop: "1px solid",
                    borderColor: "divider",
                    pt: 1.25,
                  }}
                >
                  <Typography sx={{ fontWeight: 700 }}>✓ {item}</Typography>
                </Box>
              ))}
            </Stack>
            <Typography
              sx={{
                color: "text.secondary",
                fontSize: 13,
                lineHeight: 1.6,
                mt: 3,
              }}
            >
              Exact-candidate evidence must remain current. This is staging
              evidence, not production approval.
            </Typography>
          </AdminCard>

          <AdminCard
            variant="warning"
            watermark="evidence"
            sx={{ p: { xs: 2.5, md: 3.5 } }}
          >
            <Stack
              direction={{ xs: "column", sm: "row" }}
              spacing={1}
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
                  EXTERNAL HANDOFF
                </Typography>
                <Typography
                  component="h2"
                  sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}
                >
                  Named authority must close each gate.
                </Typography>
              </Box>
              <Chip
                color="error"
                label="Production blocked"
                sx={{ alignSelf: "flex-start" }}
              />
            </Stack>
            <Box
              component="ul"
              sx={{ display: "grid", gap: 1, mt: 2.5, p: 0, listStyle: "none" }}
            >
              {externalGates.map(([gate, owner]) => (
                <Stack
                  component="li"
                  key={gate}
                  direction={{ xs: "column", sm: "row" }}
                  spacing={0.75}
                  sx={{
                    borderTop: "1px solid",
                    borderColor: "divider",
                    justifyContent: "space-between",
                    py: 1.25,
                  }}
                >
                  <Typography sx={{ fontWeight: 800 }}>{gate}</Typography>
                  <Typography
                    sx={{ color: "text.secondary", textAlign: { sm: "right" } }}
                  >
                    {owner}
                  </Typography>
                </Stack>
              ))}
            </Box>
          </AdminCard>
        </Box>

        <AdminCard variant="policy" watermark="safety" sx={{ mt: 3, p: 3 }}>
          <Typography sx={{ fontSize: 24, fontWeight: 800 }}>
            Decision integrity
          </Typography>
          <Typography sx={{ color: "text.secondary", lineHeight: 1.7, mt: 1 }}>
            Synthetic, stale, wrong-environment, wrong-kind, duplicate,
            self-approved or dependency-bypassing evidence must remain blocked.
            Even valid production-authorization evidence returns a decision
            only; it never deploys or mutates infrastructure.
          </Typography>
          <Button href="/controls" sx={{ mt: 2 }} variant="outlined">
            Review runtime controls
          </Button>
        </AdminCard>
      </Container>
    </Box>
  );
}
