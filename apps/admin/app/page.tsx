import {
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Container,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import Image from "next/image";
import Link from "next/link";
import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";

const navigation = [
  ["⌂", "Command centre", "active"],
  ["◇", "Verification", "18"],
  ["◉", "Trust & safety", "7"],
  ["◎", "Care queue", "2"],
  ["♢", "Mpanyimfo", ""],
  ["!", "Incidents", ""],
  ["+", "Workforce", ""],
  ["◌", "Circles & hosts", ""],
  ["✦", "Fires", ""],
  ["¤", "Finance", ""],
  ["▦", "Analytics", ""],
  ["Aa", "Language governance", ""],
  ["⏻", "Runtime controls", ""],
  ["≡", "Audit trail", ""],
];

const queue = [
  {
    name: "Ama Boateng",
    case: "IDV-2841",
    type: "Ghana Card fallback",
    wait: "8 min",
    priority: "Ready",
    tone: "green",
  },
  {
    name: "Kwame Asante",
    case: "IDV-2838",
    type: "Liveness uncertainty",
    wait: "14 min",
    priority: "Review",
    tone: "gold",
  },
  {
    name: "Esi Mensah",
    case: "IDV-2834",
    type: "Known-as name check",
    wait: "21 min",
    priority: "Review",
    tone: "gold",
  },
];

const incidents = [
  ["TS-448", "Payment language intercepted", "Doorway", "6 min", "Tier A"],
  ["TS-447", "Report after decline", "Member", "18 min", "Tier B"],
  ["CARE-92", "Distress phrase referred", "Room", "3 min", "Care"],
];

function MetricCard({
  label,
  value,
  note,
  accent,
}: Readonly<{ label: string; value: string; note: string; accent: string }>) {
  return (
    <Card className="metric-card" sx={{ "--metric-accent": accent }}>
      <Typography className="metric-label">{label}</Typography>
      <Typography className="metric-value">{value}</Typography>
      <Typography className="metric-note">{note}</Typography>
    </Card>
  );
}

export default function AdminHome() {
  return (
    <Box className="admin-page">
      <Box component="aside" className="admin-rail">
        <Box className="rail-brand">
          <Image src={brandMark} alt="" className="rail-mark" priority />
          <Box>
            <Typography className="rail-wordmark">obiara</Typography>
            <Typography className="rail-kicker">operations</Typography>
          </Box>
        </Box>

        <Box component="nav" aria-label="Admin navigation" className="rail-nav">
          {navigation.map(([icon, label, badge]) => (
            <Button
              key={label}
              className={`rail-link ${badge === "active" ? "is-active" : ""}`}
              aria-current={badge === "active" ? "page" : undefined}
              href={
                label === "Circles & hosts"
                  ? "/community"
                  : label === "Care queue"
                  ? "/care"
                  : label === "Mpanyimfo"
                    ? "/mpanyimfo"
                    : label === "Incidents"
                      ? "/incidents"
                    : label === "Workforce"
                      ? "/workforce"
                    : label === "Finance"
                      ? "/finance"
                      : label === "Analytics"
                        ? "/analytics"
                        : label === "Language governance"
                          ? "/governance"
                          : label === "Runtime controls"
                            ? "/controls"
                      : undefined
              }
            >
              <span aria-hidden="true">{icon}</span>
              <span>{label}</span>
              {badge && badge !== "active" ? <strong>{badge}</strong> : null}
            </Button>
          ))}
        </Box>

        <Card className="safety-card">
          <Box className="safety-pulse" />
          <Typography className="safety-label">Safety desk</Typography>
          <Typography className="safety-value">All critical queues covered</Typography>
          <Typography>Last handover · 11:42 GMT</Typography>
        </Card>

        <Box className="rail-user">
          <Avatar className="operator-avatar">AE</Avatar>
          <Box>
            <Typography sx={{ fontWeight: 700 }}>Adwoa E.</Typography>
            <Typography>T&amp;S lead · Accra</Typography>
          </Box>
          <span aria-hidden="true">⋮</span>
        </Box>
      </Box>

      <Box component="main" className="admin-main">
        <Container maxWidth={false} className="admin-shell">
          <Box component="header" className="admin-header">
            <Box>
              <Typography className="date-line">Sunday, 26 July · Accra</Typography>
              <Typography component="h1">Good morning, Adwoa.</Typography>
              <Typography>Here is what needs a human pair of eyes.</Typography>
            </Box>
            <Stack direction="row" spacing={1.25} sx={{ alignItems: "center" }}>
              <Button className="search-button">⌕ Search cases</Button>
              <Button variant="contained" className="handover-button">
                Start handover
              </Button>
              <Avatar className="header-avatar">AE</Avatar>
            </Stack>
          </Box>

          <Box className="status-banner">
            <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
              <span className="status-dot" />
              <Box>
                <Typography sx={{ fontWeight: 800 }}>Platform posture is steady</Typography>
                <Typography>Golden path passed 2 minutes ago · no Tier-A SLA breach</Typography>
              </Box>
            </Stack>
            <Button>View system health ↗</Button>
          </Box>

          <Box className="metrics-grid">
            <MetricCard label="Waiting for verification" value="18" note="p90 · 9 min" accent="#FF9F1C" />
            <MetricCard label="Open safety cases" value="7" note="2 high priority" accent="#FF4D6D" />
            <MetricCard label="Care follow-ups" value="2" note="both acknowledged" accent="#12A67C" />
            <MetricCard label="Live fires tonight" value="6" note="742 expected seats" accent="#3A0E2E" />
          </Box>

          <Box className="work-grid">
            <Card className="queue-panel">
              <Box className="panel-heading">
                <Box>
                  <Typography className="section-kicker">Verification desk</Typography>
                  <Typography component="h2">People waiting for review</Typography>
                </Box>
                <Button>Open full queue ↗</Button>
              </Box>

              <Box className="queue-list">
                {queue.map((item) => (
                  <Box className="queue-row" key={item.case}>
                    <Avatar className="queue-avatar">{item.name.split(" ").map((part) => part[0]).join("")}</Avatar>
                    <Box className="queue-person">
                      <Typography sx={{ fontWeight: 800 }}>{item.name}</Typography>
                      <Typography>{item.case} · {item.type}</Typography>
                    </Box>
                    <Typography className="wait-time">{item.wait}</Typography>
                    <Chip className={`tone-${item.tone}`} label={item.priority} />
                    <Button className="review-button">Review</Button>
                  </Box>
                ))}
              </Box>
            </Card>

            <Card className="sla-panel">
              <Box className="panel-heading compact">
                <Box>
                  <Typography className="section-kicker">Today’s response</Typography>
                  <Typography component="h2">SLA pulse</Typography>
                </Box>
                <Chip className="healthy-chip" label="Healthy" />
              </Box>
              <Box className="sla-score">
                <Typography component="strong">94%</Typography>
                <Typography>within target</Typography>
              </Box>
              <Box className="sla-lines">
                <Box>
                  <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                    <Typography>Tier A · under 2h</Typography><strong>100%</strong>
                  </Stack>
                  <LinearProgress variant="determinate" value={100} />
                </Box>
                <Box>
                  <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                    <Typography>Tier B · under 24h</Typography><strong>92%</strong>
                  </Stack>
                  <LinearProgress variant="determinate" value={92} />
                </Box>
                <Box>
                  <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                    <Typography>Verification · under 10m</Typography><strong>89%</strong>
                  </Stack>
                  <LinearProgress variant="determinate" value={89} />
                </Box>
              </Box>
              <Button className="plain-action">Review SLA detail</Button>
            </Card>
          </Box>

          <Card className="incident-panel">
            <Box className="panel-heading">
              <Box>
                <Typography className="section-kicker">Trust &amp; safety</Typography>
                <Typography component="h2">Recent signals</Typography>
              </Box>
              <Link href="/safety">
                <Button>Open safety desk</Button>
              </Link>
            </Box>
            <Box className="incident-table" role="table" aria-label="Recent trust and safety signals">
              <Box className="incident-head" role="row">
                <span>Case</span><span>Signal</span><span>Surface</span><span>Age</span><span>Route</span><span />
              </Box>
              {incidents.map(([id, signal, surface, age, route]) => (
                <Box className="incident-row" role="row" key={id}>
                  <strong>{id}</strong>
                  <span>{signal}</span>
                  <span>{surface}</span>
                  <span>{age}</span>
                  <Chip label={route} className={route === "Care" ? "tone-green" : "tone-pink"} />
                  <Button aria-label={`Open ${id}`}>→</Button>
                </Box>
              ))}
            </Box>
          </Card>
        </Container>
      </Box>
    </Box>
  );
}
