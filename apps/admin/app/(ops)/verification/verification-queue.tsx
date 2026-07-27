"use client";

import {
  Alert,
  Box,
  Button,
  Card,
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

import {
  initialReviewState,
  reviewReducer,
  type VerificationCase,
} from "./review-model";

const reasonLabels: Record<VerificationCase["reason"], string> = {
  provider_uncertain: "Provider response was uncertain",
  provider_outage: "Provider was unavailable",
  known_name: "Known-name review requested",
};

function QueueItem({
  item,
  selected,
  onSelect,
}: Readonly<{
  item: VerificationCase;
  selected: boolean;
  onSelect: () => void;
}>) {
  return (
    <Button
      aria-pressed={selected}
      className="verification-case"
      disabled={item.status === "decided"}
      onClick={onSelect}
    >
      <Box>
        <Stack direction="row" spacing={1}>
          <Typography component="strong">{item.id}</Typography>
          <Chip
            color={item.tier === "urgent" ? "warning" : "default"}
            label={item.tier === "urgent" ? "Urgent" : "Standard"}
            size="small"
          />
        </Stack>
        <Typography>{reasonLabels[item.reason]}</Typography>
        <Typography className="verification-reference">
          {item.subjectRef}
        </Typography>
      </Box>
      <span aria-hidden="true">›</span>
    </Button>
  );
}

export function VerificationQueue() {
  const [state, dispatch] = useReducer(reviewReducer, initialReviewState);
  const selected = state.cases.find((item) => item.id === state.selectedId);
  const queuedCount = state.cases.filter(
    (item) => item.status === "queued",
  ).length;

  return (
    <main className="verification-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">Verification desk</Typography>
          <Typography component="h1">
            Human review, with less exposed.
          </Typography>
          <Typography>
            Provider uncertainty comes here. Approval never happens silently.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`${queuedCount} waiting`} color="warning" />
          <Chip label="SLA healthy" color="success" />
        </Stack>
      </header>

      {state.lastDecision ? (
        <Alert severity="success" className="verification-alert">
          {state.lastDecision.caseId} recorded as {state.lastDecision.outcome}.
          The audit event is ready to persist.
        </Alert>
      ) : null}

      <Box className="verification-grid">
        <Card className="verification-list">
          <Box className="verification-panel-heading">
            <Typography component="h2">Waiting cases</Typography>
            <Typography>Oldest first</Typography>
          </Box>
          <Box aria-label="Manual verification queue">
            {state.cases.map((item) => (
              <QueueItem
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
                  <Typography component="h2">Review bounded proof</Typography>
                </Box>
                <Chip label="Not yet decided" />
              </Box>

              <Box className="verification-facts">
                <Box>
                  <Typography>Private reference</Typography>
                  <Typography component="strong">
                    {selected.subjectRef}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Queue reason</Typography>
                  <Typography component="strong">
                    {reasonLabels[selected.reason]}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Submitted</Typography>
                  <Typography component="strong">
                    26 Jul, 17:{selected.id === "IDV-2841" ? "54" : "41"} GMT
                  </Typography>
                </Box>
              </Box>

              <Alert severity="info">
                Full card numbers, raw media and contact details are not shown.
                Opening evidence creates an operator audit event.
              </Alert>

              <Button
                variant="outlined"
                onClick={() => dispatch({ type: "open-evidence" })}
              >
                Open redacted evidence
              </Button>

              <Box className="verification-actions">
                <Button
                  variant="contained"
                  color="success"
                  onClick={() =>
                    dispatch({ type: "propose", outcome: "approve" })
                  }
                >
                  Propose approval
                </Button>
                <Button
                  variant="outlined"
                  color="error"
                  onClick={() =>
                    dispatch({ type: "propose", outcome: "reject" })
                  }
                >
                  Propose rejection
                </Button>
              </Box>
            </>
          ) : (
            <Box className="verification-empty" role="status">
              <Typography component="h2">The queue is clear.</Typography>
              <Typography>
                New uncertain cases will appear here without exposing raw
                identity data.
              </Typography>
            </Box>
          )}
        </Card>
      </Box>

      <Dialog
        aria-labelledby="evidence-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "close-evidence" })}
        open={state.evidenceOpened}
      >
        <DialogTitle id="evidence-title">Redacted evidence</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <Alert severity="warning">
              Access is recorded against the operator session.
            </Alert>
            <Box className="evidence-row">
              <Typography>Document match</Typography>
              <Typography component="strong">
                Partial, confidence 0.71
              </Typography>
            </Box>
            <Box className="evidence-row">
              <Typography>Provider proof</Typography>
              <Typography component="strong">Reference ending 72CA</Typography>
            </Box>
            <Box className="evidence-row">
              <Typography>Raw media</Typography>
              <Typography component="strong">Not retained</Typography>
            </Box>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "close-evidence" })}>
            Close evidence
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        aria-labelledby="decision-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-decision" })}
        open={state.pendingOutcome !== null}
      >
        <DialogTitle id="decision-title">
          Confirm{" "}
          {state.pendingOutcome === "approve" ? "approval" : "rejection"}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography>
              Add a clear operator reason. This decision affects account access
              and will be audited.
            </Typography>
            <TextField
              autoFocus
              helperText="At least 8 characters, no raw identity data"
              label="Decision reason"
              multiline
              onChange={(event) =>
                dispatch({ type: "set-reason", reason: event.target.value })
              }
              rows={3}
              value={state.decisionReason}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-decision" })}>
            Go back
          </Button>
          <Button
            disabled={state.decisionReason.trim().length < 8}
            onClick={() => dispatch({ type: "confirm-decision" })}
            variant="contained"
          >
            Record decision
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
