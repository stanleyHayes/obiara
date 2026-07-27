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

import { docketReducer, initialDocketState } from "./docket-model";

export function MpanyimfoDocket() {
  const [state, dispatch] = useReducer(docketReducer, initialDocketState);
  const activeVotes = state.seats.filter(
    (seat) => !seat.recused && seat.vote,
  ).length;

  return (
    <main className="verification-shell mpanyimfo-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">Mpanyimfo docket</Typography>
          <Typography component="h1">
            A ruling needs more than one voice.
          </Typography>
          <Typography>
            Redacted records, conflict recusal, quorum and separate appeals
            protect the review process.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`Docket ${state.caseId}`} />
          <Chip
            color={state.status === "deliberating" ? "warning" : "success"}
            label={state.status}
          />
        </Stack>
      </header>

      {state.appealReference ? (
        <Alert severity="success" className="verification-alert">
          Appeal docket {state.appealReference} opened for a different human
          panel. The original ruling remains intact.
        </Alert>
      ) : null}

      <Card className="mpanyimfo-record">
        <Box className="verification-panel-heading">
          <Box>
            <Typography className="section-kicker">Redacted record</Typography>
            <Typography component="h2">
              Proportionality review for SAFE-8Q2M
            </Typography>
          </Box>
          <Chip label="Reporter hidden" />
        </Box>
        <Box className="verification-facts">
          <Box>
            <Typography>Original action</Typography>
            <Typography component="strong">
              Temporary messaging restriction
            </Typography>
          </Box>
          <Box>
            <Typography>Evidence scope</Typography>
            <Typography component="strong">3 bounded references</Typography>
          </Box>
          <Box>
            <Typography>Appeal window</Typography>
            <Typography component="strong">Open for 7 days</Typography>
          </Box>
        </Box>
        <Alert severity="info">
          No raw messages, reporter identity, counsel content or device graph is
          visible in this docket.
        </Alert>
      </Card>

      <Box className="mpanyimfo-grid">
        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Panel and recusal</Typography>
            <Chip label={`${activeVotes} votes recorded`} />
          </Box>
          <Stack spacing={1.5}>
            {state.seats.map((seat) => (
              <Box className="elder-seat" key={seat.id}>
                <Box>
                  <Typography component="strong">{seat.label}</Typography>
                  <Typography>
                    {seat.recused
                      ? "Recused from this docket"
                      : seat.vote
                        ? `Vote: ${seat.vote}`
                        : "No vote recorded"}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1}>
                  <Button
                    disabled={state.status !== "deliberating"}
                    onClick={() =>
                      dispatch({ type: "toggle-recusal", elderId: seat.id })
                    }
                    variant="text"
                  >
                    {seat.recused ? "Restore seat" : "Recuse"}
                  </Button>
                  <Button
                    disabled={seat.recused || state.status !== "deliberating"}
                    onClick={() =>
                      dispatch({
                        type: "vote",
                        elderId: seat.id,
                        vote: "uphold",
                      })
                    }
                    variant="outlined"
                  >
                    Uphold
                  </Button>
                  <Button
                    disabled={seat.recused || state.status !== "deliberating"}
                    onClick={() =>
                      dispatch({
                        type: "vote",
                        elderId: seat.id,
                        vote: "revise",
                      })
                    }
                    variant="outlined"
                  >
                    Revise
                  </Button>
                </Stack>
              </Box>
            ))}
          </Stack>
        </Card>

        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Reasoned ruling</Typography>
            <Chip label="Two matching votes required" />
          </Box>
          {state.status === "deliberating" ? (
            <Stack spacing={2}>
              <TextField
                helperText="At least 20 characters. Explain proportionality and process."
                label="Panel reason"
                multiline
                onChange={(event) =>
                  dispatch({
                    type: "ruling-reason",
                    value: event.target.value,
                  })
                }
                rows={5}
                value={state.rulingReason}
              />
              <Button
                disabled={state.rulingReason.trim().length < 20}
                onClick={() => dispatch({ type: "confirm-ruling" })}
                variant="contained"
              >
                Record quorum ruling
              </Button>
            </Stack>
          ) : (
            <Stack spacing={2}>
              <Alert severity="success">
                Ruling recorded: {state.ruling}. The reason and panel votes are
                immutable in this view.
              </Alert>
              <Typography>{state.rulingReason}</Typography>
              <Button
                disabled={state.status === "appealed"}
                onClick={() => dispatch({ type: "request-appeal" })}
                variant="outlined"
              >
                Open separate appeal
              </Button>
            </Stack>
          )}
        </Card>
      </Box>

      <Dialog
        aria-labelledby="appeal-docket-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-appeal" })}
        open={state.appealPending}
      >
        <DialogTitle id="appeal-docket-title">
          Open a separate appeal docket
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="info">
              The appeal goes to a different human panel. It does not erase or
              overwrite the original ruling.
            </Alert>
            <TextField
              helperText="At least 20 characters."
              label="Appeal grounds"
              multiline
              onChange={(event) =>
                dispatch({ type: "appeal-reason", value: event.target.value })
              }
              rows={4}
              value={state.appealReason}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-appeal" })}>
            Go back
          </Button>
          <Button
            disabled={state.appealReason.trim().length < 20}
            onClick={() => dispatch({ type: "confirm-appeal" })}
            variant="contained"
          >
            Confirm appeal docket
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
