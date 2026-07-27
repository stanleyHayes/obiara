"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  LinearProgress,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import {
  initialLaunchState,
  launchBlocked,
  launchReducer,
} from "./launch-model";

export function LaunchReadinessDesk() {
  const [state, dispatch] = useReducer(launchReducer, initialLaunchState);
  const blocked = launchBlocked(state);
  return (
    <Box sx={{ bgcolor: "#f7efe3", color: "#2b151f", minHeight: "100vh", py: 4 }}>
      <Container maxWidth="lg">
        <Stack direction={{ xs: "column", md: "row" }} spacing={2} sx={{ alignItems: { md: "center" }, justifyContent: "space-between", mb: 5 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.4 }}>LAUNCH READINESS</Typography>
            <Typography component="h1" sx={{ fontSize: { xs: 44, md: 72 }, fontWeight: 800, letterSpacing: "-0.06em", lineHeight: .95, mt: 1 }}>
              People before opening day.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>{state.market} · {state.readinessRef}</Typography>
          </Box>
          <Link href="/"><Button variant="outlined">Back to command centre</Button></Link>
        </Stack>

        <Alert severity={blocked ? "error" : "success"} sx={{ borderRadius: 3, mb: 3 }}>
          <strong>{blocked ? "Launch remains blocked." : "People-readiness gates pass."}</strong>{" "}
          Cohort size without density, training or current licensing is not readiness.
        </Alert>

        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" } }}>
          {state.gates.map((gate) => (
            <Card key={gate.id} sx={{ borderRadius: 4, display: "flex", flexDirection: "column", minHeight: 290, p: 3 }}>
              <Stack direction="row" sx={{ alignItems: "flex-start", justifyContent: "space-between" }}>
                <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.1 }}>{gate.label.toUpperCase()}</Typography>
                <Chip color={gate.evidenceComplete && gate.passes ? "success" : "error"} label={gate.evidenceComplete ? (gate.passes ? "Ready" : "Below gate") : "Incomplete"} size="small" />
              </Stack>
              <Typography sx={{ fontSize: 44, fontWeight: 800, mt: 2 }}>{gate.numerator}/{gate.denominator}</Typography>
              <LinearProgress color={gate.evidenceComplete && gate.passes ? "success" : "error"} sx={{ borderRadius: 99, my: 1.5 }} value={(gate.numerator / gate.denominator) * 100} variant="determinate" />
              <Typography sx={{ color: "text.secondary", lineHeight: 1.5 }}>{gate.requirement}</Typography>
              <Typography sx={{ fontSize: 13, fontWeight: 700, mt: "auto", pt: 2 }}>{gate.expires}</Typography>
            </Card>
          ))}
        </Box>

        <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}>Launch-day coverage</Typography>
        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" } }}>
          {state.staffing.map((desk) => (
            <Card key={desk.desk} sx={{ borderRadius: 4, p: 3 }}>
              <Stack direction="row" sx={{ alignItems: "center", justifyContent: "space-between" }}>
                <Typography sx={{ fontWeight: 800 }}>{desk.desk}</Typography>
                <Chip color={desk.staffed >= desk.required ? "success" : "error"} label={desk.staffed >= desk.required ? "Covered" : "Gap"} size="small" />
              </Stack>
              <Typography sx={{ fontSize: 36, fontWeight: 800, mt: 1 }}>{desk.staffed}/{desk.required}</Typography>
              <Typography sx={{ color: "text.secondary" }}>{desk.window}</Typography>
            </Card>
          ))}
        </Box>

        <Card sx={{ borderRadius: 4, display: "grid", gap: 4, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" }, mt: 3, p: 3 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>LAUNCH CALENDAR</Typography>
            <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}>Dates do not overrule gates.</Typography>
            <Stack spacing={1.2} sx={{ mt: 2 }}>
              {state.milestones.map((milestone) => (
                <Stack direction="row" key={milestone.label} sx={{ alignItems: "center", justifyContent: "space-between" }}>
                  <Box><Typography sx={{ fontWeight: 800 }}>{milestone.label}</Typography><Typography sx={{ color: "text.secondary" }}>{milestone.date}</Typography></Box>
                  <Chip color={milestone.state === "ready" ? "success" : "error"} label={milestone.state} size="small" />
                </Stack>
              ))}
            </Stack>
          </Box>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>WAITLIST THROTTLE</Typography>
            {state.throttleState === "none" ? (
              <>
                <Alert severity="warning" sx={{ my: 1.5 }}>Low-density evidence is current. A proposal can slow new entry; it cannot message or remove anyone.</Alert>
                <TextField
                  fullWidth
                  label="Throttle reason"
                  multiline
                  onChange={(event) => dispatch({ type: "throttle-reason", value: event.target.value })}
                  rows={3}
                  value={state.throttleReason}
                />
                <Button
                  disabled={!state.lowDensityEvidence || state.throttleReason.trim().length < 12}
                  fullWidth
                  onClick={() => dispatch({ type: "prepare-throttle" })}
                  sx={{ mt: 1.5 }}
                  variant="contained"
                >
                  Prepare waitlist throttle proposal
                </Button>
              </>
            ) : (
              <Alert severity="info">
                <strong>{state.throttleRef}</strong><br />
                Proposal ready for separate approval. Waitlist and notification state remain unchanged.
              </Alert>
            )}
          </Box>
        </Card>

        <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}>
          Quality-gated campus attribution
        </Typography>
        <Alert severity="info" sx={{ borderRadius: 3, mb: 2 }}>
          Aggregate outcomes recognise programme quality—not individual recruitment volume. No ambassador identity, leaderboard or payout appears here.
        </Alert>
        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" } }}>
          {state.campusAttribution.map((campus) => {
            const passes = campus.evidenceComplete && campus.unresolvedSafety === 0 && campus.sustainedThirtyDay > 0;
            return (
              <Card key={campus.campus} sx={{ borderRadius: 4, p: 3 }}>
                <Stack direction="row" sx={{ alignItems: "center", justifyContent: "space-between" }}>
                  <Typography sx={{ fontWeight: 800 }}>{campus.campus}</Typography>
                  <Chip color={passes ? "success" : "error"} label={passes ? "Quality gate met" : "Blocked"} size="small" />
                </Stack>
                <Typography sx={{ fontSize: 38, fontWeight: 800, mt: 2 }}>
                  {campus.sustainedThirtyDay}/{campus.consentedIntroductions}
                </Typography>
                <Typography sx={{ color: "text.secondary" }}>Consented introductions sustained at 30 days</Typography>
                <Typography sx={{ fontWeight: 700, mt: 2 }}>
                  {campus.evidenceComplete ? `${campus.unresolvedSafety} unresolved safety` : "Evidence incomplete"}
                </Typography>
              </Card>
            );
          })}
        </Box>

        <Card sx={{ borderRadius: 4, display: "grid", gap: 4, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" }, mt: 3, p: 3 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>UAT COHORT</Typography>
            <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}>Consent, training, completion.</Typography>
            <Stack spacing={1.2} sx={{ mt: 2 }}>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}><Typography>Consented</Typography><Typography sx={{ fontWeight: 800 }}>{state.uat.consented}/{state.uat.invited}</Typography></Stack>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}><Typography>Trained</Typography><Typography sx={{ fontWeight: 800 }}>{state.uat.trained}/{state.uat.consented}</Typography></Stack>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}><Typography>Completed</Typography><Typography sx={{ fontWeight: 800 }}>{state.uat.completed}/{state.uat.trained}</Typography></Stack>
              <Alert severity="error">{state.uat.criticalFeedbackOpen} critical findings remain open.</Alert>
            </Stack>
          </Box>
          <Box>
            {state.triageState === "none" ? (
              <>
                <TextField
                  fullWidth
                  label="Feedback triage reason"
                  multiline
                  onChange={(event) => dispatch({ type: "triage-reason", value: event.target.value })}
                  rows={4}
                  value={state.triageReason}
                />
                <Button
                  disabled={state.uat.criticalFeedbackOpen === 0 || state.triageReason.trim().length < 12}
                  fullWidth
                  onClick={() => dispatch({ type: "prepare-triage" })}
                  sx={{ mt: 1.5 }}
                  variant="contained"
                >
                  Prepare human triage record
                </Button>
              </>
            ) : (
              <Alert severity="info"><strong>{state.triageRef}</strong><br />Triage prepared. Findings, ownership and launch state remain unchanged.</Alert>
            )}
          </Box>
        </Card>

        <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}>
          Hypercare command center
        </Typography>
        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" } }}>
          {state.hypercare.map((signal) => (
            <Card key={signal.signal} sx={{ borderRadius: 4, p: 3 }}>
              <Stack direction="row" sx={{ alignItems: "center", justifyContent: "space-between" }}>
                <Typography sx={{ fontWeight: 800 }}>{signal.signal}</Typography>
                <Chip color={signal.state === "healthy" ? "success" : "error"} label={signal.state} size="small" />
              </Stack>
              <Typography sx={{ fontSize: 34, fontWeight: 800, mt: 2 }}>{signal.current}</Typography>
              <Typography sx={{ color: "text.secondary" }}>Gate {signal.target}</Typography>
              <Typography sx={{ fontSize: 13, fontWeight: 700, mt: 2 }}>{signal.owner}</Typography>
            </Card>
          ))}
        </Box>
        <Alert severity="warning" sx={{ borderRadius: 3, mt: 2 }}>
          Daily review stays blocked while any signal is red. This desk cannot close incidents, replenish budgets or activate production.
        </Alert>

        <Card sx={{ borderRadius: 4, display: "grid", gap: 4, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" }, mt: 3, p: 3 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>HUMAN READINESS REVIEW</Typography>
            <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}>A note cannot turn red into green.</Typography>
            <Typography sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}>
              This view contains aggregates and opaque evidence status only—no family contact list, outreach controls or individual training records. Reviews annotate the snapshot and never certify, license or activate a market.
            </Typography>
          </Box>
          {state.reviewState === "none" ? (
            <Box>
              <TextField
                fullWidth
                label="Readiness review note"
                multiline
                onChange={(event) => dispatch({ type: "review-note", value: event.target.value })}
                rows={4}
                value={state.reviewNote}
              />
              <Button
                disabled={state.reviewNote.trim().length < 12}
                fullWidth
                onClick={() => dispatch({ type: "record-review" })}
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Record readiness review
              </Button>
            </Box>
          ) : (
            <Alert severity="info">
              <strong>{state.reviewRef}</strong><br />
              Review recorded. Host, matchmaker and cohort gates remain blocked and unchanged.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
