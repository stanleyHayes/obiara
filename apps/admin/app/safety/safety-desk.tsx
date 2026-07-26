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
  FormControlLabel,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import {
  initialSafetyDeskState,
  safetyDeskReducer,
  type SafetyCase,
} from "./safety-model";

function SafetyQueueItem({
  item,
  selected,
  onSelect,
}: Readonly<{
  item: SafetyCase;
  selected: boolean;
  onSelect: () => void;
}>) {
  return (
    <Button
      aria-pressed={selected}
      className="safety-case"
      onClick={onSelect}
    >
      <Box>
        <Stack direction="row" spacing={1}>
          <Typography component="strong">{item.id}</Typography>
          <Chip
            color={item.tier === "A" ? "error" : "warning"}
            label={`Tier ${item.tier}`}
            size="small"
          />
        </Stack>
        <Typography>{item.category}</Typography>
        <Typography className="safety-reference">
          {item.subjectRef} / {item.age}
        </Typography>
      </Box>
      <span aria-hidden="true">›</span>
    </Button>
  );
}

export function SafetyDesk() {
  const [state, dispatch] = useReducer(
    safetyDeskReducer,
    initialSafetyDeskState,
  );
  const selected = state.cases.find((item) => item.id === state.selectedId);

  return (
    <main className="verification-shell safety-desk-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Trust and safety desk
          </Typography>
          <Typography component="h1">
            See enough to act, never everything.
          </Typography>
          <Typography>
            Reporter identity and raw content stay hidden from the queue.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`${state.cases.length} open`} color="warning" />
          <Chip label="Purpose logged" color="success" />
        </Stack>
      </header>

      <Box className="verification-grid">
        <Card className="verification-list">
          <Box className="verification-panel-heading">
            <Typography component="h2">Priority queue</Typography>
            <Typography>Oldest SLA first</Typography>
          </Box>
          <Box aria-label="Trust and safety cases">
            {state.cases.map((item) => (
              <SafetyQueueItem
                item={item}
                key={item.id}
                onSelect={() => dispatch({ type: "select", caseId: item.id })}
                selected={item.id === state.selectedId}
              />
            ))}
          </Box>
        </Card>

        <Card className="verification-review">
          {selected ? (
            <>
              <Box className="verification-panel-heading">
                <Box>
                  <Typography className="section-kicker">
                    Case {selected.id}
                  </Typography>
                  <Typography component="h2">
                    Review redacted evidence
                  </Typography>
                </Box>
                <Chip
                  label={
                    selected.holdStatus === "active"
                      ? "Legal hold active"
                      : selected.holdStatus === "pending"
                        ? "Hold requested"
                        : "No legal hold"
                  }
                  color={
                    selected.holdStatus === "active" ? "success" : "default"
                  }
                />
              </Box>

              <Box className="verification-facts">
                <Box>
                  <Typography>Private subject</Typography>
                  <Typography component="strong">
                    {selected.subjectRef}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Category</Typography>
                  <Typography component="strong">
                    {selected.category}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Queue age</Typography>
                  <Typography component="strong">{selected.age}</Typography>
                </Box>
              </Box>

              <Alert severity="info">
                Opening evidence creates an operator audit event. Enter a
                bounded case purpose first.
              </Alert>

              <TextField
                fullWidth
                helperText="At least 12 characters. Do not enter member content."
                label="Evidence access purpose"
                onChange={(event) =>
                  dispatch({ type: "purpose", value: event.target.value })
                }
                value={state.accessPurpose}
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={state.accessAcknowledged}
                    onChange={(event) =>
                      dispatch({
                        type: "acknowledge",
                        checked: event.target.checked,
                      })
                    }
                  />
                }
                label="I understand this access is audited."
              />

              <Box className="verification-actions">
                <Button
                  disabled={
                    state.accessPurpose.trim().length < 12 ||
                    !state.accessAcknowledged
                  }
                  onClick={() => dispatch({ type: "open-evidence" })}
                  variant="contained"
                >
                  Open redacted evidence
                </Button>
                <Button
                  disabled={selected.holdStatus !== "none"}
                  onClick={() => dispatch({ type: "request-hold" })}
                  variant="outlined"
                >
                  Request legal hold
                </Button>
              </Box>
            </>
          ) : null}
        </Card>
      </Box>

      <Dialog
        aria-labelledby="safety-evidence-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "close-evidence" })}
        open={state.evidenceOpen}
      >
        <DialogTitle id="safety-evidence-title">
          Redacted case evidence
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <Alert severity="warning">
              Access purpose: {state.accessPurpose}
            </Alert>
            {selected?.evidence.map((item) => (
              <Box className="evidence-row" key={item.label}>
                <Typography>{item.label}</Typography>
                <Typography component="strong">
                  {item.redacted ? `[${item.value}]` : item.value}
                </Typography>
              </Box>
            ))}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "close-evidence" })}>
            Close evidence
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        aria-labelledby="legal-hold-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-hold" })}
        open={state.holdPending}
      >
        <DialogTitle id="legal-hold-title">Request legal hold</DialogTitle>
        <DialogContent>
          <Typography>
            This creates a reversible hold request for an authorized custodian.
            It does not expose evidence, decide the case or delete anything.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-hold" })}>
            Go back
          </Button>
          <Button
            onClick={() => dispatch({ type: "confirm-hold" })}
            variant="contained"
          >
            Confirm request
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
