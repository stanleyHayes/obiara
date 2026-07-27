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
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import {
  careQueueReducer,
  careScripts,
  initialCareQueueState,
} from "./care-model";

export function CareQueue() {
  const [state, dispatch] = useReducer(careQueueReducer, initialCareQueueState);
  const selected = state.cases.find((item) => item.id === state.selectedId);
  const script = careScripts.find((item) => item.id === state.selectedScriptId);

  return (
    <main className="verification-shell care-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">Care queue</Typography>
          <Typography component="h1">
            Resources first. People always.
          </Typography>
          <Typography>
            This desk offers reviewed support resources without diagnosis,
            pressure or punitive action.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`${state.cases.length} waiting`} color="warning" />
          <Chip label="Human contact only" color="success" />
        </Stack>
      </header>

      {state.lastSent ? (
        <Alert severity="success" className="verification-alert">
          Approved resource message prepared for {state.lastSent.caseId}.
          Delivery remains with the member-contact service.
        </Alert>
      ) : null}

      <Box className="verification-grid">
        <Card className="verification-list">
          <Box className="verification-panel-heading">
            <Typography component="h2">Care requests</Typography>
            <Typography>Least exposure</Typography>
          </Box>
          <Box aria-label="Care cases">
            {state.cases.map((item) => (
              <Button
                aria-pressed={item.id === state.selectedId}
                className="care-case"
                key={item.id}
                onClick={() => dispatch({ type: "select", caseId: item.id })}
              >
                <Box>
                  <Typography component="strong">{item.id}</Typography>
                  <Typography>
                    {item.reason === "requested_support"
                      ? "Member requested support"
                      : "Safety follow-up"}
                  </Typography>
                  <Typography className="safety-reference">
                    {item.memberRef} / {item.age}
                  </Typography>
                </Box>
                <Chip
                  label={
                    item.contactPreference === "none"
                      ? "Do not contact"
                      : item.contactPreference
                  }
                  size="small"
                />
              </Button>
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
                    Choose an approved resource script
                  </Typography>
                </Box>
                <Chip
                  color={
                    selected.contactPreference === "none"
                      ? "default"
                      : "success"
                  }
                  label={
                    selected.contactPreference === "none"
                      ? "No contact consent"
                      : `Contact: ${selected.contactPreference}`
                  }
                />
              </Box>

              <Alert
                severity={
                  selected.contactPreference === "none" ? "warning" : "info"
                }
              >
                {selected.contactPreference === "none"
                  ? "The member has not permitted contact. Scripts can be reviewed but not prepared for sending."
                  : "Use only the approved script. Do not add diagnoses, promises or pressure."}
              </Alert>

              <Box className="care-scripts">
                {careScripts.map((item) => (
                  <Button
                    aria-pressed={item.id === state.selectedScriptId}
                    className="care-script"
                    key={item.id}
                    onClick={() =>
                      dispatch({ type: "choose-script", scriptId: item.id })
                    }
                  >
                    <Box>
                      <Typography className="section-kicker">
                        {item.version} / approved
                      </Typography>
                      <Typography component="h3">{item.title}</Typography>
                      <Typography>{item.body}</Typography>
                      <Typography component="strong">
                        Resource: {item.resource}
                      </Typography>
                    </Box>
                  </Button>
                ))}
              </Box>

              <Box className="verification-actions">
                <Button
                  disabled={!script || selected.contactPreference === "none"}
                  onClick={() => dispatch({ type: "prepare-send" })}
                  variant="contained"
                >
                  Review contact
                </Button>
              </Box>
            </>
          ) : null}
        </Card>
      </Box>

      <Dialog
        aria-labelledby="care-send-title"
        fullWidth
        maxWidth="sm"
        onClose={() => dispatch({ type: "cancel-send" })}
        open={state.sendPending}
      >
        <DialogTitle id="care-send-title">Confirm human contact</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <Alert severity="info">
              This prepares one approved message for the member&apos;s chosen
              channel. It does not diagnose, schedule repeated contact or change
              a safety case.
            </Alert>
            <Typography component="strong">{script?.title}</Typography>
            <Typography>{script?.body}</Typography>
            <Typography>Resource: {script?.resource}</Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: "cancel-send" })}>
            Go back
          </Button>
          <Button
            onClick={() => dispatch({ type: "confirm-send" })}
            variant="contained"
          >
            Confirm approved message
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
