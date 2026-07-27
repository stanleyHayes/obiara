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
  decisionGateSummary,
  initialLaunchState,
  launchBlocked,
  launchReducer,
} from "./launch-model";

export function LaunchReadinessDesk() {
  const [state, dispatch] = useReducer(launchReducer, initialLaunchState);
  const blocked = launchBlocked(state);
  const decisionSummary = decisionGateSummary(state);
  const selectedHandoff = state.decisionGates.find(
    (gate) => gate.id === state.selectedHandoffId,
  );
  const authorityLabels = {
    repository: "Repository",
    founder_legal: "Founder / legal",
    provider_procurement: "Provider / procurement",
    credential_store: "Credentials / store",
    cohort_operations: "Cohort / operations",
    production_action: "Production action",
  } as const;
  return (
    <Box
      sx={{ bgcolor: "#f7efe3", color: "#2b151f", minHeight: "100vh", py: 4 }}
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
              People before opening day.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              {state.market} · {state.readinessRef}
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        <Alert
          severity={blocked ? "error" : "success"}
          sx={{ borderRadius: 3, mb: 3 }}
        >
          <strong>
            {blocked
              ? "Launch remains blocked."
              : "People-readiness gates pass."}
          </strong>{" "}
          Cohort size without density, training or current licensing is not
          readiness.
        </Alert>

        <Card
          sx={{
            bgcolor: "#2b151f",
            borderRadius: 4,
            color: "#fff8f0",
            overflow: "hidden",
            p: { xs: 2.5, md: 4 },
          }}
        >
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={3}
            sx={{
              alignItems: { md: "flex-end" },
              justifyContent: "space-between",
            }}
          >
            <Box sx={{ maxWidth: 680 }}>
              <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
                <Box aria-hidden sx={{ fontSize: 17, lineHeight: 1 }}>
                  ◆
                </Box>
                <Typography
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1.25 }}
                >
                  PRODUCTION DECISION CONTROL
                </Typography>
              </Stack>
              <Typography
                component="h2"
                sx={{
                  fontSize: { xs: 34, md: 48 },
                  fontWeight: 800,
                  letterSpacing: "-0.045em",
                  lineHeight: 1,
                  mt: 1.5,
                }}
              >
                Evidence can inform a decision. It cannot make one.
              </Typography>
              <Typography sx={{ color: "#d9c8cf", lineHeight: 1.6, mt: 1.5 }}>
                Candidate {state.candidateSha} · snapshot {state.generatedAt}.
                Repository proof is separated from legal authority, provider
                diligence, credentials, cohort readiness and production action.
              </Typography>
            </Box>
            <Box
              sx={{
                border: "1px solid rgba(255,255,255,.18)",
                borderRadius: 3,
                minWidth: { md: 280 },
                p: 2,
              }}
            >
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography sx={{ color: "#d9c8cf" }}>Verified</Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {decisionSummary.verified}
                </Typography>
              </Stack>
              <Stack
                direction="row"
                sx={{ justifyContent: "space-between", mt: 0.75 }}
              >
                <Typography sx={{ color: "#d9c8cf" }}>
                  Awaiting authority
                </Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {decisionSummary.awaiting_external}
                </Typography>
              </Stack>
              <Stack
                direction="row"
                sx={{ justifyContent: "space-between", mt: 0.75 }}
              >
                <Typography sx={{ color: "#d9c8cf" }}>Blocked</Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {decisionSummary.blocked}
                </Typography>
              </Stack>
            </Box>
          </Stack>

          <Box
            sx={{
              display: "grid",
              gap: 1.25,
              gridTemplateColumns: { xs: "1fr", md: "repeat(2,minmax(0,1fr))" },
              mt: 3,
            }}
          >
            {state.decisionGates.map((gate) => {
              const statusLabel =
                gate.state === "verified"
                  ? "Verified"
                  : gate.state === "awaiting_external"
                    ? "Awaiting authority"
                    : "Blocked";
              return (
                <Box
                  key={gate.id}
                  sx={{
                    bgcolor:
                      gate.state === "verified"
                        ? "#173d32"
                        : "rgba(255,255,255,.065)",
                    border: "1px solid",
                    borderColor:
                      gate.state === "verified"
                        ? "#3d8068"
                        : "rgba(255,255,255,.13)",
                    borderRadius: 3,
                    p: 2,
                  }}
                >
                  <Stack
                    direction="row"
                    spacing={1.25}
                    sx={{ alignItems: "flex-start" }}
                  >
                    <Box
                      aria-hidden
                      sx={{
                        color:
                          gate.state === "verified" ? "#8ee0bc" : "#f3a3b5",
                        fontSize: 18,
                        lineHeight: 1.2,
                      }}
                    >
                      {gate.state === "verified"
                        ? "●"
                        : gate.state === "awaiting_external"
                          ? "◌"
                          : "◆"}
                    </Box>
                    <Box sx={{ minWidth: 0, width: "100%" }}>
                      <Stack
                        direction={{ xs: "column", sm: "row" }}
                        spacing={0.75}
                        sx={{
                          alignItems: { sm: "center" },
                          justifyContent: "space-between",
                        }}
                      >
                        <Typography sx={{ fontWeight: 800 }}>
                          {gate.label}
                        </Typography>
                        <Chip
                          label={statusLabel}
                          size="small"
                          sx={{
                            alignSelf: { xs: "flex-start", sm: "auto" },
                            bgcolor:
                              gate.state === "verified" ? "#d5f5e6" : "#f8dce3",
                            color:
                              gate.state === "verified" ? "#173d32" : "#671d36",
                            fontWeight: 800,
                          }}
                        />
                      </Stack>
                      <Typography
                        sx={{ color: "#d9c8cf", fontSize: 13, mt: 0.75 }}
                      >
                        {authorityLabels[gate.authority]} · {gate.owner}
                      </Typography>
                      <Typography sx={{ fontWeight: 700, mt: 1.25 }}>
                        {gate.evidence}
                      </Typography>
                      <Typography
                        sx={{ color: "#d9c8cf", fontSize: 13, mt: 0.5 }}
                      >
                        {gate.freshness}
                      </Typography>
                      <Stack
                        direction="row"
                        spacing={0.75}
                        sx={{ alignItems: "center", mt: 1.25 }}
                      >
                        <Box aria-hidden sx={{ fontSize: 14 }}>
                          ↗
                        </Box>
                        <Typography sx={{ color: "#fff8f0", fontSize: 13 }}>
                          Depends on {gate.dependency}
                        </Typography>
                      </Stack>
                    </Box>
                  </Stack>
                </Box>
              );
            })}
          </Box>
          <Alert
            icon={
              <Box aria-hidden sx={{ fontSize: 20 }}>
                ◆
              </Box>
            }
            severity="warning"
            sx={{ bgcolor: "#fff1cc", borderRadius: 2.5, mt: 2.5 }}
          >
            This desk has no launch action. Production remains absent until
            every prerequisite is independently evidenced and the named human
            authorities record a go/no-go decision.
          </Alert>
        </Card>

        <Card sx={{ borderRadius: 4, mt: 3, overflow: "hidden" }}>
          <Box
            sx={{
              borderBottom: "1px solid",
              borderColor: "divider",
              p: { xs: 2.5, md: 3.5 },
            }}
          >
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              EXTERNAL EVIDENCE HANDOFF
            </Typography>
            <Typography
              component="h2"
              sx={{
                fontSize: { xs: 32, md: 42 },
                fontWeight: 800,
                letterSpacing: "-0.04em",
                mt: 1,
              }}
            >
              Prepare the packet. Preserve the authority.
            </Typography>
            <Typography
              sx={{
                color: "text.secondary",
                lineHeight: 1.6,
                maxWidth: 760,
                mt: 1,
              }}
            >
              Choose one blocked gate to see exactly what the named authority
              must supply. A prepared handoff is an opaque coordination record
              only—it cannot upload evidence, approve a decision or change
              readiness.
            </Typography>
          </Box>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: {
                xs: "1fr",
                lg: "minmax(0,1.15fr) minmax(340px,.85fr)",
              },
            }}
          >
            <Box
              sx={{
                borderBottom: { xs: "1px solid", lg: 0 },
                borderColor: "divider",
                borderRight: { lg: "1px solid" },
                p: { xs: 2, md: 3 },
              }}
            >
              <Stack spacing={1}>
                {state.decisionGates
                  .filter((gate) => gate.authority !== "repository")
                  .map((gate) => (
                    <Button
                      aria-pressed={state.selectedHandoffId === gate.id}
                      disabled={state.preparedHandoffRef !== null}
                      key={gate.id}
                      onClick={() =>
                        dispatch({ type: "select-handoff", gateId: gate.id })
                      }
                      sx={{
                        alignItems: "flex-start",
                        bgcolor:
                          state.selectedHandoffId === gate.id
                            ? "#f3e1e7"
                            : "transparent",
                        border: "1px solid",
                        borderColor:
                          state.selectedHandoffId === gate.id
                            ? "#8e3159"
                            : "divider",
                        borderRadius: 2.5,
                        color: "#2b151f",
                        display: "flex",
                        flexDirection: "column",
                        minHeight: 72,
                        p: 1.5,
                        textAlign: "left",
                        textTransform: "none",
                        width: "100%",
                      }}
                    >
                      <Stack
                        direction="row"
                        sx={{
                          alignItems: "center",
                          justifyContent: "space-between",
                          width: "100%",
                        }}
                      >
                        <Typography sx={{ fontWeight: 800 }}>
                          {gate.label}
                        </Typography>
                        <Typography aria-hidden sx={{ color: "#8e3159" }}>
                          →
                        </Typography>
                      </Stack>
                      <Typography
                        sx={{ color: "text.secondary", fontSize: 13, mt: 0.4 }}
                      >
                        {gate.owner}
                      </Typography>
                    </Button>
                  ))}
              </Stack>
            </Box>
            <Box sx={{ bgcolor: "#fffaf4", p: { xs: 2.5, md: 3.5 } }}>
              {state.preparedHandoffRef ? (
                <Alert severity="info">
                  <strong>{state.preparedHandoffRef}</strong>
                  <br />
                  Coordination record prepared. Evidence, authority and gate
                  state are unchanged.
                </Alert>
              ) : selectedHandoff ? (
                <Stack spacing={2}>
                  <Box>
                    <Typography sx={{ fontSize: 24, fontWeight: 800 }}>
                      {selectedHandoff.label}
                    </Typography>
                    <Typography sx={{ color: "text.secondary", mt: 0.5 }}>
                      {authorityLabels[selectedHandoff.authority]} ·{" "}
                      {selectedHandoff.owner}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography
                      sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1 }}
                    >
                      REQUIRED EVIDENCE
                    </Typography>
                    <Stack
                      component="ul"
                      spacing={0.75}
                      sx={{ m: 0, mt: 1, pl: 2.5 }}
                    >
                      {selectedHandoff.requiredEvidence.map((requirement) => (
                        <Typography
                          component="li"
                          key={requirement}
                          sx={{ lineHeight: 1.5 }}
                        >
                          {requirement}
                        </Typography>
                      ))}
                    </Stack>
                  </Box>
                  <Alert severity="warning">
                    {selectedHandoff.externalAct}
                  </Alert>
                  <TextField
                    fullWidth
                    label="Handoff coordination note"
                    multiline
                    onChange={(event) =>
                      dispatch({
                        type: "handoff-note",
                        value: event.target.value,
                      })
                    }
                    rows={3}
                    value={state.handoffNote}
                  />
                  <Button
                    disabled={state.handoffNote.trim().length < 12}
                    onClick={() => dispatch({ type: "prepare-handoff" })}
                    variant="contained"
                  >
                    Prepare opaque handoff record
                  </Button>
                </Stack>
              ) : (
                <Box sx={{ py: { xs: 2, md: 8 }, textAlign: "center" }}>
                  <Typography sx={{ fontSize: 28, fontWeight: 800 }}>
                    Select a blocked gate
                  </Typography>
                  <Typography sx={{ color: "text.secondary", mt: 1 }}>
                    Requirements and the irreducible external act will appear
                    here.
                  </Typography>
                </Box>
              )}
            </Box>
          </Box>
        </Card>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" },
          }}
        >
          {state.gates.map((gate) => (
            <Card
              key={gate.id}
              sx={{
                borderRadius: 4,
                display: "flex",
                flexDirection: "column",
                minHeight: 290,
                p: 3,
              }}
            >
              <Stack
                direction="row"
                sx={{
                  alignItems: "flex-start",
                  justifyContent: "space-between",
                }}
              >
                <Typography
                  sx={{
                    color: "#8e3159",
                    fontSize: 12,
                    fontWeight: 800,
                    letterSpacing: 1.1,
                  }}
                >
                  {gate.label.toUpperCase()}
                </Typography>
                <Chip
                  color={
                    gate.evidenceComplete && gate.passes ? "success" : "error"
                  }
                  label={
                    gate.evidenceComplete
                      ? gate.passes
                        ? "Ready"
                        : "Below gate"
                      : "Incomplete"
                  }
                  size="small"
                />
              </Stack>
              <Typography sx={{ fontSize: 44, fontWeight: 800, mt: 2 }}>
                {gate.numerator}/{gate.denominator}
              </Typography>
              <LinearProgress
                color={
                  gate.evidenceComplete && gate.passes ? "success" : "error"
                }
                sx={{ borderRadius: 99, my: 1.5 }}
                value={(gate.numerator / gate.denominator) * 100}
                variant="determinate"
              />
              <Typography sx={{ color: "text.secondary", lineHeight: 1.5 }}>
                {gate.requirement}
              </Typography>
              <Typography
                sx={{ fontSize: 13, fontWeight: 700, mt: "auto", pt: 2 }}
              >
                {gate.expires}
              </Typography>
            </Card>
          ))}
        </Box>

        <Typography
          component="h2"
          sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}
        >
          Launch-day coverage
        </Typography>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" },
          }}
        >
          {state.staffing.map((desk) => (
            <Card key={desk.desk} sx={{ borderRadius: 4, p: 3 }}>
              <Stack
                direction="row"
                sx={{ alignItems: "center", justifyContent: "space-between" }}
              >
                <Typography sx={{ fontWeight: 800 }}>{desk.desk}</Typography>
                <Chip
                  color={desk.staffed >= desk.required ? "success" : "error"}
                  label={desk.staffed >= desk.required ? "Covered" : "Gap"}
                  size="small"
                />
              </Stack>
              <Typography sx={{ fontSize: 36, fontWeight: 800, mt: 1 }}>
                {desk.staffed}/{desk.required}
              </Typography>
              <Typography sx={{ color: "text.secondary" }}>
                {desk.window}
              </Typography>
            </Card>
          ))}
        </Box>

        <Card
          sx={{
            borderRadius: 4,
            display: "grid",
            gap: 4,
            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
            mt: 3,
            p: 3,
          }}
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
              LAUNCH CALENDAR
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}
            >
              Dates do not overrule gates.
            </Typography>
            <Stack spacing={1.2} sx={{ mt: 2 }}>
              {state.milestones.map((milestone) => (
                <Stack
                  direction="row"
                  key={milestone.label}
                  sx={{ alignItems: "center", justifyContent: "space-between" }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 800 }}>
                      {milestone.label}
                    </Typography>
                    <Typography sx={{ color: "text.secondary" }}>
                      {milestone.date}
                    </Typography>
                  </Box>
                  <Chip
                    color={milestone.state === "ready" ? "success" : "error"}
                    label={milestone.state}
                    size="small"
                  />
                </Stack>
              ))}
            </Stack>
          </Box>
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              WAITLIST THROTTLE
            </Typography>
            {state.throttleState === "none" ? (
              <>
                <Alert severity="warning" sx={{ my: 1.5 }}>
                  Low-density evidence is current. A proposal can slow new
                  entry; it cannot message or remove anyone.
                </Alert>
                <TextField
                  fullWidth
                  label="Throttle reason"
                  multiline
                  onChange={(event) =>
                    dispatch({
                      type: "throttle-reason",
                      value: event.target.value,
                    })
                  }
                  rows={3}
                  value={state.throttleReason}
                />
                <Button
                  disabled={
                    !state.lowDensityEvidence ||
                    state.throttleReason.trim().length < 12
                  }
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
                <strong>{state.throttleRef}</strong>
                <br />
                Proposal ready for separate approval. Waitlist and notification
                state remain unchanged.
              </Alert>
            )}
          </Box>
        </Card>

        <Typography
          component="h2"
          sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}
        >
          Quality-gated campus attribution
        </Typography>
        <Alert severity="info" sx={{ borderRadius: 3, mb: 2 }}>
          Aggregate outcomes recognise programme quality—not individual
          recruitment volume. No ambassador identity, leaderboard or payout
          appears here.
        </Alert>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" },
          }}
        >
          {state.campusAttribution.map((campus) => {
            const passes =
              campus.evidenceComplete &&
              campus.unresolvedSafety === 0 &&
              campus.sustainedThirtyDay > 0;
            return (
              <Card key={campus.campus} sx={{ borderRadius: 4, p: 3 }}>
                <Stack
                  direction="row"
                  sx={{ alignItems: "center", justifyContent: "space-between" }}
                >
                  <Typography sx={{ fontWeight: 800 }}>
                    {campus.campus}
                  </Typography>
                  <Chip
                    color={passes ? "success" : "error"}
                    label={passes ? "Quality gate met" : "Blocked"}
                    size="small"
                  />
                </Stack>
                <Typography sx={{ fontSize: 38, fontWeight: 800, mt: 2 }}>
                  {campus.sustainedThirtyDay}/{campus.consentedIntroductions}
                </Typography>
                <Typography sx={{ color: "text.secondary" }}>
                  Consented introductions sustained at 30 days
                </Typography>
                <Typography sx={{ fontWeight: 700, mt: 2 }}>
                  {campus.evidenceComplete
                    ? `${campus.unresolvedSafety} unresolved safety`
                    : "Evidence incomplete"}
                </Typography>
              </Card>
            );
          })}
        </Box>

        <Card
          sx={{
            borderRadius: 4,
            display: "grid",
            gap: 4,
            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
            mt: 3,
            p: 3,
          }}
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
              UAT COHORT
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}
            >
              Consent, training, completion.
            </Typography>
            <Stack spacing={1.2} sx={{ mt: 2 }}>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography>Consented</Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {state.uat.consented}/{state.uat.invited}
                </Typography>
              </Stack>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography>Trained</Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {state.uat.trained}/{state.uat.consented}
                </Typography>
              </Stack>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography>Completed</Typography>
                <Typography sx={{ fontWeight: 800 }}>
                  {state.uat.completed}/{state.uat.trained}
                </Typography>
              </Stack>
              <Alert severity="error">
                {state.uat.criticalFeedbackOpen} critical findings remain open.
              </Alert>
            </Stack>
          </Box>
          <Box>
            {state.triageState === "none" ? (
              <>
                <TextField
                  fullWidth
                  label="Feedback triage reason"
                  multiline
                  onChange={(event) =>
                    dispatch({
                      type: "triage-reason",
                      value: event.target.value,
                    })
                  }
                  rows={4}
                  value={state.triageReason}
                />
                <Button
                  disabled={
                    state.uat.criticalFeedbackOpen === 0 ||
                    state.triageReason.trim().length < 12
                  }
                  fullWidth
                  onClick={() => dispatch({ type: "prepare-triage" })}
                  sx={{ mt: 1.5 }}
                  variant="contained"
                >
                  Prepare human triage record
                </Button>
              </>
            ) : (
              <Alert severity="info">
                <strong>{state.triageRef}</strong>
                <br />
                Triage prepared. Findings, ownership and launch state remain
                unchanged.
              </Alert>
            )}
          </Box>
        </Card>

        <Typography
          component="h2"
          sx={{ fontSize: 34, fontWeight: 800, mb: 2, mt: 5 }}
        >
          Hypercare command center
        </Typography>
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" },
          }}
        >
          {state.hypercare.map((signal) => (
            <Card key={signal.signal} sx={{ borderRadius: 4, p: 3 }}>
              <Stack
                direction="row"
                sx={{ alignItems: "center", justifyContent: "space-between" }}
              >
                <Typography sx={{ fontWeight: 800 }}>
                  {signal.signal}
                </Typography>
                <Chip
                  color={signal.state === "healthy" ? "success" : "error"}
                  label={signal.state}
                  size="small"
                />
              </Stack>
              <Typography sx={{ fontSize: 34, fontWeight: 800, mt: 2 }}>
                {signal.current}
              </Typography>
              <Typography sx={{ color: "text.secondary" }}>
                Gate {signal.target}
              </Typography>
              <Typography sx={{ fontSize: 13, fontWeight: 700, mt: 2 }}>
                {signal.owner}
              </Typography>
            </Card>
          ))}
        </Box>
        <Alert severity="warning" sx={{ borderRadius: 3, mt: 2 }}>
          Daily review stays blocked while any signal is red. This desk cannot
          close incidents, replenish budgets or activate production.
        </Alert>

        <Card
          sx={{
            borderRadius: 4,
            display: "grid",
            gap: 4,
            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
            mt: 3,
            p: 3,
          }}
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
              HUMAN READINESS REVIEW
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}
            >
              A note cannot turn red into green.
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}
            >
              This view contains aggregates and opaque evidence status only—no
              family contact list, outreach controls or individual training
              records. Reviews annotate the snapshot and never certify, license
              or activate a market.
            </Typography>
          </Box>
          {state.reviewState === "none" ? (
            <Box>
              <TextField
                fullWidth
                label="Readiness review note"
                multiline
                onChange={(event) =>
                  dispatch({ type: "review-note", value: event.target.value })
                }
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
              <strong>{state.reviewRef}</strong>
              <br />
              Review recorded. Host, matchmaker and cohort gates remain blocked
              and unchanged.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
