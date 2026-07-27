"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useReducer } from "react";

import { EmptyState } from "../../empty-state";
import {
  initialWorkforceState,
  workforceReducer,
  type ExposureCategory,
} from "./workforce-model";

const categoryLabels: Readonly<Record<ExposureCategory, string>> = {
  financial_coercion: "Financial coercion",
  harassment: "Harassment",
  sexual_safety: "Sexual safety",
};

export function WorkforceSafeguards() {
  const [state, dispatch] = useReducer(workforceReducer, initialWorkforceState);

  return (
    <main className="verification-shell workforce-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Workforce safeguards
          </Typography>
          <Typography component="h1">
            The work must not consume the worker.
          </Typography>
          <Typography>
            Preview exposure, take protected breaks, opt out without penalty,
            and ask a supervisor for support.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`${state.shiftMinutes} min on shift`} />
          <Chip
            color={
              state.exposureCount >= state.maxExposure ? "warning" : "success"
            }
            label={`${state.exposureCount}/${state.maxExposure} exposures`}
          />
        </Stack>
      </header>

      {state.supportRequested ? (
        <Alert severity="success" className="verification-alert">
          Supervisor support requested. No diagnosis, performance flag or
          retaliation marker was created.
        </Alert>
      ) : null}
      {state.optedOut ? (
        <Alert severity="info" className="verification-alert">
          You opted out of this assignment without penalty. No new evidence is
          shown.
        </Alert>
      ) : null}

      <Box className="workforce-grid">
        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Exposure boundary</Typography>
            <Chip label="No productivity score" />
          </Box>
          <Typography>
            Exposure limits protect wellbeing. They are not performance targets.
          </Typography>
          <LinearProgress
            aria-label="Current exposure limit"
            sx={{ my: 3 }}
            value={(state.exposureCount / state.maxExposure) * 100}
            variant="determinate"
          />
          <Stack spacing={1.2}>
            {(Object.keys(categoryLabels) as readonly ExposureCategory[]).map(
              (category) => (
                <Button
                  disabled={
                    state.breakActive ||
                    state.optedOut ||
                    state.exposureCount >= state.maxExposure
                  }
                  key={category}
                  onClick={() => dispatch({ type: "preview", category })}
                  variant="outlined"
                >
                  Preview {categoryLabels[category]} assignment
                </Button>
              ),
            )}
          </Stack>
        </Card>

        <Card>
          <Box className="verification-panel-heading">
            <Typography component="h2">Protected controls</Typography>
            <Chip
              color={state.breakActive ? "success" : "default"}
              label={state.breakActive ? "Break active" : "Available"}
            />
          </Box>
          <Stack spacing={1.4}>
            <Button
              onClick={() =>
                dispatch({
                  type: state.breakActive ? "end-break" : "start-break",
                })
              }
              variant="contained"
            >
              {state.breakActive
                ? "End protected break"
                : "Start protected break"}
            </Button>
            <Button
              disabled={state.optedOut}
              onClick={() => dispatch({ type: "opt-out" })}
              variant="outlined"
            >
              Opt out without penalty
            </Button>
            <Button
              disabled={state.supportRequested}
              onClick={() => dispatch({ type: "request-support" })}
              variant="outlined"
            >
              Request supervisor support
            </Button>
          </Stack>
          <Alert severity="info" sx={{ mt: 3 }}>
            These controls do not create a health diagnosis, HR score or hidden
            surveillance event.
          </Alert>
        </Card>
      </Box>

      <Card className="workforce-preview">
        {state.selectedCategory ? (
          <>
            <Box>
              <Typography className="section-kicker">
                Category preview only
              </Typography>
              <Typography component="h2">
                {categoryLabels[state.selectedCategory]}
              </Typography>
              <Typography>
                No evidence is visible until you explicitly accept this
                assignment.
              </Typography>
            </Box>
            <Stack spacing={1}>
              <Button
                disabled={state.acceptedAssignment}
                onClick={() => dispatch({ type: "accept" })}
                variant="contained"
              >
                Accept assignment
              </Button>
              <Button
                disabled={!state.acceptedAssignment}
                onClick={() => dispatch({ type: "complete" })}
                variant="outlined"
              >
                Record bounded exposure
              </Button>
            </Stack>
          </>
        ) : (
          <EmptyState
            icon="◌"
            title="No assignment selected"
            description="Choose a category preview only when you are ready."
          />
        )}
      </Card>
    </main>
  );
}
