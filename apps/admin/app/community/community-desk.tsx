"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  Chip,
  Container,
  FormControlLabel,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import {
  communityReducer,
  initialCommunityState,
  selectedHostEligible,
} from "./community-model";

export function CommunityDesk() {
  const [state, dispatch] = useReducer(communityReducer, initialCommunityState);
  const eligible = selectedHostEligible(state);
  return (
    <Box sx={{ bgcolor: "#f7efe3", color: "#2b151f", minHeight: "100vh", py: 4 }}>
      <Container maxWidth="lg">
        <Stack direction={{ xs: "column", md: "row" }} spacing={2} sx={{ alignItems: { md: "center" }, justifyContent: "space-between", mb: 5 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.4 }}>COMMUNITY OPERATIONS</Typography>
            <Typography component="h1" sx={{ fontSize: { xs: 44, md: 72 }, fontWeight: 800, letterSpacing: "-0.06em", lineHeight: .95, mt: 1 }}>
              Keep the circle held.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>{state.circleLabel} · {state.circleRef}</Typography>
          </Box>
          <Link href="/"><Button variant="outlined">Back to command centre</Button></Link>
        </Stack>

        <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" } }}>
          <Card sx={{ borderRadius: 4, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>Circle density</Typography>
            <Typography sx={{ fontSize: 38, fontWeight: 800 }}>{state.activeMembers}/{state.capacity}</Typography>
            <Typography sx={{ color: "text.secondary" }}>Active seats · no member list exposed</Typography>
          </Card>
          <Card sx={{ borderRadius: 4, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>Scheduled fire</Typography>
            <Typography sx={{ fontSize: 24, fontWeight: 800, mt: 1 }}>{state.fireStarts}</Typography>
            <Typography sx={{ color: "text.secondary" }}>{state.fireRef} · immutable reference</Typography>
          </Card>
          <Card sx={{ borderRadius: 4, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>Capacity posture</Typography>
            <Typography sx={{ fontSize: 30, fontWeight: 800, mt: 1 }}>6 seats remain</Typography>
            <Chip color="success" label="Within capacity" size="small" />
          </Card>
        </Box>

        <Card sx={{ borderRadius: 4, mt: 3, p: 3 }}>
          <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>HOST ELIGIBILITY</Typography>
          <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}>Verification and certification stay current.</Typography>
          <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" }, mt: 2 }}>
            {state.hostCandidates.map((host) => (
              <Button
                key={host.ref}
                onClick={() => dispatch({ type: "select-host", ref: host.ref })}
                sx={{ alignItems: "flex-start", borderRadius: 3, justifyContent: "space-between", p: 2, textAlign: "left" }}
                variant={state.selectedHostRef === host.ref ? "contained" : "outlined"}
              >
                <Box>
                  <Typography sx={{ fontWeight: 800 }}>{host.label}</Typography>
                  <Typography sx={{ opacity: .8 }}>{host.ref} · {host.certificationEnds}</Typography>
                </Box>
                <Chip color={host.verified && host.certified ? "success" : "warning"} label={host.verified && host.certified ? "Eligible" : "Not eligible"} size="small" />
              </Button>
            ))}
          </Box>
        </Card>

        <Card sx={{ borderRadius: 4, display: "grid", gap: 4, gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" }, mt: 3, p: 3 }}>
          <Box>
            <Typography sx={{ color: "#8e3159", fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}>BOUNDED ACTION PROPOSAL</Typography>
            <Typography component="h2" sx={{ fontSize: 34, fontWeight: 800, mt: 1 }}>No silent host or fire change.</Typography>
            <Typography sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}>
              The preview shows a generic participant notice without names, contact details or community content. Preparing it does not assign a host, cancel a fire or send a notification.
            </Typography>
          </Box>
          {state.proposalState === "draft" ? (
            <Box>
              {!eligible ? <Alert severity="error" sx={{ mb: 2 }}>Selected host is not currently certified. Proposal is blocked.</Alert> : null}
              <TextField
                fullWidth
                label="Operational reason"
                multiline
                onChange={(event) => dispatch({ type: "reason", value: event.target.value })}
                rows={3}
                value={state.actionReason}
              />
              <Alert severity="info" sx={{ mt: 1.5 }}>
                Notice preview: “Your scheduled fire has an operations update. The time remains {state.fireStarts}. Open Fie for the reviewed host status.”
              </Alert>
              <FormControlLabel
                control={<Checkbox checked={state.noticePreviewConfirmed} onChange={() => dispatch({ type: "confirm-notice-preview" })} />}
                label="I reviewed the participant notice preview"
              />
              <Button
                disabled={!eligible || state.actionReason.trim().length < 12 || !state.noticePreviewConfirmed}
                fullWidth
                onClick={() => dispatch({ type: "prepare-proposal" })}
                variant="contained"
              >
                Prepare operations proposal
              </Button>
            </Box>
          ) : (
            <Alert severity="success">
              <strong>{state.proposalRef} · proposal ready</strong><br />
              Eligible host, reason and notice preview recorded. Circle, fire and notification state remain unchanged.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
