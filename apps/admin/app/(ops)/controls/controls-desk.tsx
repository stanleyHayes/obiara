"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import { controlsReducer, initialControlsState } from "./controls-model";

export function ControlsDesk() {
  const [state, dispatch] = useReducer(controlsReducer, initialControlsState);
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
              RUNTIME CONTROLS
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
              Small scope. Clear expiry.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              {state.proposalRef} · immutable proposal preview
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        <Alert
          severity={state.state === "apply_ready" ? "success" : "warning"}
          sx={{ borderRadius: 1, mb: 3 }}
        >
          <strong>
            {state.state === "apply_ready"
              ? "Apply-ready; nothing activated."
              : "No runtime state has changed."}
          </strong>{" "}
          A separate audited control plane applies approved proposals and
          verifies rollback.
        </Alert>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(4,minmax(0,1fr))" },
          }}
        >
          {[
            ["Capability", state.capability],
            ["Action", "Kill switch · disabled"],
            ["Scope", `${state.environment} · ${state.market}`],
            ["Expiry", `${state.expiresInHours} hours · fail closed`],
          ].map(([label, value]) => (
            <Card key={label} sx={{ borderRadius: 1, p: 2.5 }}>
              <Typography
                sx={{ color: "text.secondary", fontSize: 13, fontWeight: 700 }}
              >
                {label}
              </Typography>
              <Typography
                sx={{
                  fontSize: 20,
                  fontWeight: 800,
                  mt: 1,
                  overflowWrap: "anywhere",
                }}
              >
                {value}
              </Typography>
            </Card>
          ))}
        </Box>

        <Card
          sx={{
            borderRadius: 1,
            display: "grid",
            gap: 4,
            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
            mt: 3,
            p: 3,
          }}
        >
          <Box>
            <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
              <Chip color="success" label="MFA step-up current" size="small" />
              <Chip label="Two approvals" size="small" />
            </Stack>
            <Typography
              component="h2"
              sx={{ fontSize: 34, fontWeight: 800, mt: 2 }}
            >
              Disabling is the safe direction.
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}
            >
              This proposal cannot target a person, widen itself to production,
              remain forever, deploy code or claim an automatic rollback. Its
              exact scope and expiry are fixed.
            </Typography>
          </Box>

          {state.state === "draft" ? (
            <Box>
              <TextField
                fullWidth
                label="Operational reason"
                multiline
                onChange={(event) =>
                  dispatch({ type: "reason", value: event.target.value })
                }
                rows={4}
                value={state.reason}
              />
              <Button
                disabled={state.reason.trim().length < 12}
                fullWidth
                onClick={() => dispatch({ type: "first-approve" })}
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Record stepped-up proposal
              </Button>
            </Box>
          ) : state.state === "first_approved" ? (
            <Box>
              <Alert severity="info" sx={{ mb: 2 }}>
                Proposed by {state.proposer}
              </Alert>
              <FormControl fullWidth>
                <InputLabel id="control-approver-label">
                  Distinct second approver
                </InputLabel>
                <Select
                  label="Distinct second approver"
                  labelId="control-approver-label"
                  onChange={(event) =>
                    dispatch({
                      type: "second-approver",
                      actor: event.target.value,
                    })
                  }
                  value={state.secondApprover}
                >
                  <MenuItem value="operator•••A1">Adwoa · proposer</MenuItem>
                  <MenuItem value="operator•••C9">
                    Efua · release controller
                  </MenuItem>
                </Select>
              </FormControl>
              <Button
                disabled={
                  !state.secondApprover ||
                  state.secondApprover === state.proposer
                }
                fullWidth
                onClick={() => dispatch({ type: "confirm-second" })}
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Confirm second approval
              </Button>
            </Box>
          ) : (
            <Alert severity="success">
              <strong>Kill-switch proposal is apply-ready.</strong>
              <br />
              Staging/GH only, two-hour expiry, disabled fail-closed posture. No
              runtime or member state changed.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
