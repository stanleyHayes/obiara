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
