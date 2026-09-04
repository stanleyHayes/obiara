import { Box, Button, Chip, Typography } from "@mui/material";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon } from "../../admin-icons";

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
    <Box className="launch-redesign">
      <Box component="header" className="launch-hero">
        <AdminCardWatermark watermark="safety" />
        <Box className="launch-hero-copy">
          <Button className="launch-back" href="/">
            ← Command centre
          </Button>
          <Box className="launch-kicker">
            <AdminIcon name="launch" aria-hidden="true" />
            <Typography className="section-kicker">
              LAUNCH READINESS · STATIC INVENTORY
            </Typography>
          </Box>
          <Typography component="h1">
            Ready to review. Not cleared to launch.
          </Typography>
          <Typography className="launch-hero-intro">
            Repository handoffs are prepared for staging review. Production
            remains blocked until every named authority supplies current
            evidence.
          </Typography>
        </Box>
        <Box className="launch-verdict" aria-label="Current launch verdict">
          <span>CURRENT VERDICT</span>
          <strong>NO GO</strong>
          <Typography>External authority remains outstanding</Typography>
          <div>
            <i /> Production activation unavailable
          </div>
        </Box>
      </Box>

      <Box component="section" className="launch-boundary">
        <Box className="launch-boundary-icon">
          <AdminIcon name="launch" aria-hidden="true" />
        </Box>
        <Box>
          <Typography className="section-kicker">AUTHORITY BOUNDARY</Typography>
          <Typography component="h2">
            This desk cannot launch production.
          </Typography>
          <Typography>
            Green repository checks cannot approve legal posture, purchase
            providers, create credentials, recruit people, or activate
            infrastructure.
          </Typography>
        </Box>
        <span>NO LAUNCH ACTION</span>
      </Box>

      <Box className="launch-overview">
        <AdminCard
          variant="policy"
          watermark="evidence"
          className="launch-evidence-packet"
        >
          <Box className="launch-section-heading">
            <Box>
              <Typography className="section-kicker">
                ENGINEERING PACKET
              </Typography>
              <Typography component="h2">
                Evidence ready for staging.
              </Typography>
            </Box>
            <Chip label="Repository-controlled" variant="outlined" />
          </Box>
          <Box component="ul" className="launch-evidence-list">
            {repositoryEvidence.map((item, index) => (
              <Box component="li" key={item}>
                <span>{(index + 1).toString().padStart(2, "0")}</span>
                <Typography>{item}</Typography>
                <strong>READY</strong>
              </Box>
            ))}
          </Box>
          <Typography className="launch-evidence-note">
            Exact-candidate evidence must remain current. This is staging
            evidence, not production approval.
          </Typography>
        </AdminCard>
        <Box component="aside" className="launch-tally">
          <AdminCardWatermark watermark="evidence" />
          <Typography className="section-kicker">CLOSURE TALLY</Typography>
          <strong>{externalGates.length.toString().padStart(2, "0")}</strong>
          <Typography>external gates still require named authority</Typography>
          <div>
            <span>Repository packet</span>
            <b>READY</b>
          </div>
          <div>
            <span>Production authority</span>
            <b>BLOCKED</b>
          </div>
        </Box>
      </Box>

      <AdminCard
        variant="warning"
        watermark="evidence"
        className="launch-gate-ledger"
      >
        <Box className="launch-ledger-heading">
          <Box>
            <Typography className="section-kicker">EXTERNAL HANDOFF</Typography>
            <Typography component="h2">
              Named authority must close each gate.
            </Typography>
          </Box>
          <Chip color="error" label="Production blocked" />
        </Box>
        <Box
          component="ol"
          className="launch-gate-list"
          sx={{ gridTemplateColumns: "1fr" }}
        >
          {externalGates.map(([gate, owner], index) => (
            <Box component="li" key={gate}>
              <span>{(index + 1).toString().padStart(2, "0")}</span>
              <Typography>{gate}</Typography>
              <div>
                <small>ACCOUNTABLE AUTHORITY</small>
                <strong>{owner}</strong>
              </div>
              <i>OPEN</i>
            </Box>
          ))}
        </Box>
      </AdminCard>

      <AdminCard
        variant="policy"
        watermark="safety"
        className="launch-integrity"
      >
        <Box className="launch-integrity-icon">
          <AdminIcon name="governance" aria-hidden="true" />
        </Box>
        <Box>
          <Typography className="section-kicker">DECISION INTEGRITY</Typography>
          <Typography component="h2">
            Evidence may inform a decision. It never performs deployment.
          </Typography>
          <Typography>
            Synthetic, stale, wrong-environment, wrong-kind, duplicate,
            self-approved, or dependency-bypassing evidence must remain blocked.
            Even valid production-authorization evidence returns a decision
            only.
          </Typography>
        </Box>
        <Button href="/controls" variant="outlined">
          Review runtime controls
        </Button>
      </AdminCard>
    </Box>
  );
}
