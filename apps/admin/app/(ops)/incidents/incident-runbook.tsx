"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import { incidentReducer, initialIncidentState } from "./incident-model";

export function IncidentRunbook() {
  const [state, dispatch] = useReducer(incidentReducer, initialIncidentState);
  const mandatoryComplete = state.steps
    .filter((step) => step.mandatory)
    .every((step) => step.complete);

  return (
    <main className="verification-shell incident-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">Incident response</Typography>
          <Typography component="h1">
            One runbook. Two accountable roles.
          </Typography>
          <Typography>
            Mandatory checkpoints stay ordered and regulator packets stay
            redacted.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip color="error" label={state.severity} />
          <Chip label={state.runbookVersion} />
          <Chip label={state.status} />
        </Stack>
      </header>

      {state.packetReference ? (
        <Alert severity="success" className="verification-alert">
          Redacted packet {state.packetReference} is ready for authorized human
          submission. Nothing was sent automatically.
        </Alert>
      ) : null}

      <Box className="incident-runbook-grid">
        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Declared roles</Typography>
            <Chip label="Distinct people required" />
          </Box>
          <Stack spacing={2}>
            <TextField
              disabled={state.status !== "active"}
              label="Incident commander"
              onChange={(event) =>
                dispatch({
                  type: "assign-commander",
                  value: event.target.value,
                })
              }
              value={state.commander}
            />
            <TextField
              disabled={state.status !== "active"}
              error={
                Boolean(state.commander) &&
                state.commander.trim() === state.recorder.trim()
              }
              helperText="Must be different from the commander."
              label="Incident recorder"
              onChange={(event) =>
                dispatch({
                  type: "assign-recorder",
                  value: event.target.value,
                })
              }
              value={state.recorder}
            />
            <Alert severity="info">
              P1 regulatory notification clock: assess and record the applicable
              deadline immediately. External notice remains a human legal
              decision.
            </Alert>
          </Stack>
        </Card>

        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Ordered checkpoints</Typography>
            <Chip
              color={mandatoryComplete ? "success" : "warning"}
              label={mandatoryComplete ? "Mandatory complete" : "In progress"}
            />
          </Box>
          <Stack spacing={1.5}>
            {state.steps.map((step, index) => (
              <Button
                aria-pressed={step.complete}
                className="runbook-step"
                disabled={step.complete || state.status !== "active"}
                key={step.id}
                onClick={() =>
                  dispatch({ type: "complete-step", stepId: step.id })
                }
              >
                <Checkbox checked={step.complete} tabIndex={-1} />
                <Box>
                  <Typography component="strong">
                    {index + 1}. {step.label}
                  </Typography>
                  <Typography>
                    {step.mandatory ? "Mandatory" : "Recommended"}
                  </Typography>
                </Box>
              </Button>
            ))}
          </Stack>
        </Card>
      </Box>

      <Card className="incident-packet">
        <Box>
          <Typography className="section-kicker">Regulatory packet</Typography>
          <Typography component="h2">
            Redacted facts, explicit human submission.
          </Typography>
          <Typography>
            Includes incident reference, severity, runbook version and
            checkpoint status only. No raw member data or evidence body.
          </Typography>
        </Box>
        <Stack spacing={1}>
          <Button
            disabled={!mandatoryComplete || state.status !== "active"}
            onClick={() => dispatch({ type: "prepare-packet" })}
            variant="contained"
          >
            Review regulator packet
          </Button>
          <Button
            disabled={state.status !== "packet_ready"}
            onClick={() => dispatch({ type: "prepare-close" })}
            variant="outlined"
          >
            Prepare incident close
          </Button>
        </Stack>
      </Card>

      <Dialog
        aria-labelledby="packet-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-packet" })}
        open={state.packetPending}
      >
        <DialogTitle id="packet-title">Confirm redacted packet</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <Alert severity="warning">
              Confirmation marks a packet ready. It does not notify a regulator
              or change severity automatically.
            </Alert>
            <Typography>Incident: {state.id}</Typography>
            <Typography>Severity: {state.severity}</Typography>
            <Typography>Runbook: {state.runbookVersion}</Typography>
            <Typography>Member data: excluded</Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-packet" })}>
            Go back
          </Button>
          <Button
            onClick={() => dispatch({ type: "confirm-packet" })}
            variant="contained"
          >
            Mark packet ready
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        aria-labelledby="close-incident-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-close" })}
        open={state.closePending}
      >
        <DialogTitle id="close-incident-title">
          Confirm incident close
        </DialogTitle>
        <DialogContent>
          <Typography>
            Commander {state.commander} and recorder {state.recorder} remain
            accountable. Closing does not delete evidence or release legal
            holds.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-close" })}>
            Go back
          </Button>
          <Button
            onClick={() => dispatch({ type: "confirm-close" })}
            variant="contained"
          >
            Confirm close
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
