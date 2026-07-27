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

import {
  checksPass,
  governanceReducer,
  initialGovernanceState,
} from "./governance-model";

export function GovernanceDesk() {
  const [state, dispatch] = useReducer(
    governanceReducer,
    initialGovernanceState,
  );
  const ready = checksPass(state);
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
              LANGUAGE GOVERNANCE
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
              Meaning needs human custody.
            </Typography>
            <Typography sx={{ color: "#69535d", mt: 2 }}>
              {state.proposalRef} · {state.locale} · version {state.version}
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        <Alert
          severity={state.publishState === "publish_ready" ? "success" : "info"}
          sx={{ borderRadius: 1, mb: 3 }}
        >
          <strong>
            {state.publishState === "publish_ready"
              ? "Publish-ready preview."
              : "Current registry remains unchanged."}
          </strong>{" "}
          This desk records review readiness only; it cannot deploy or activate
          a market.
        </Alert>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(3,minmax(0,1fr))" },
          }}
        >
          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
              Key parity
            </Typography>
            <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
              {state.translatedKeys}/{state.sourceKeys}
            </Typography>
            <Chip
              color={
                state.translatedKeys === state.sourceKeys ? "success" : "error"
              }
              label={
                state.translatedKeys === state.sourceKeys
                  ? "Complete"
                  : "Missing keys"
              }
            />
          </Card>
          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
              Placeholder validation
            </Typography>
            <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
              {state.placeholdersValid ? "Valid" : "Drift"}
            </Typography>
            <Typography sx={{ color: "text.secondary" }}>
              Names and counts preserve typed parameters.
            </Typography>
          </Card>
          <Card sx={{ borderRadius: 1, p: 3 }}>
            <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
              Terminology review
            </Typography>
            <Typography sx={{ fontSize: 36, fontWeight: 800 }}>
              {state.terminologyReviewed ? "Reviewed" : "Pending"}
            </Typography>
            <Typography sx={{ color: "text.secondary" }}>
              Cultural meaning is not inferred by a machine.
            </Typography>
          </Card>
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
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.2,
              }}
            >
              REVIEW ACKNOWLEDGEMENT
            </Typography>
            <Typography
              component="h2"
              sx={{ fontSize: 32, fontWeight: 800, mt: 1 }}
            >
              Two people protect one meaning.
            </Typography>
            <Typography
              sx={{ color: "text.secondary", lineHeight: 1.6, mt: 1 }}
            >
              English fallback remains a runtime safety net—not approval
              evidence. Reviewers confirm parity, placeholders and culturally
              accurate product terms before a version becomes publish-ready.
            </Typography>
          </Box>
          {state.publishState === "draft" ? (
            <Box>
              <TextField
                fullWidth
                label="Human review note"
                multiline
                onChange={(event) =>
                  dispatch({ type: "review-note", value: event.target.value })
                }
                rows={4}
                value={state.humanReviewNote}
              />
              <Button
                disabled={!ready || state.humanReviewNote.trim().length < 12}
                fullWidth
                onClick={() =>
                  dispatch({ type: "first-approve", actor: "operator•••A1" })
                }
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Record first approval
              </Button>
            </Box>
          ) : state.publishState === "first_approved" ? (
            <Box>
              <Alert severity="info" sx={{ mb: 2 }}>
                First approval · {state.primaryApprover}
              </Alert>
              <FormControl fullWidth>
                <InputLabel id="second-approver-label">
                  Distinct second approver
                </InputLabel>
                <Select
                  label="Distinct second approver"
                  labelId="second-approver-label"
                  onChange={(event) =>
                    dispatch({
                      type: "second-approver",
                      actor: event.target.value,
                    })
                  }
                  value={state.secondApprover}
                >
                  <MenuItem value="operator•••A1">
                    Adwoa · same operator
                  </MenuItem>
                  <MenuItem value="operator•••B8">
                    Kofi · language governance
                  </MenuItem>
                </Select>
              </FormControl>
              <Button
                disabled={
                  !state.secondApprover ||
                  state.secondApprover === state.primaryApprover
                }
                fullWidth
                onClick={() => dispatch({ type: "confirm-second-approval" })}
                sx={{ mt: 1.5 }}
                variant="contained"
              >
                Confirm second approval
              </Button>
            </Box>
          ) : (
            <Alert severity="success">
              <strong>Version {state.version} is publish-ready.</strong>
              <br />
              The immutable proposal is ready for a separate deployment
              workflow. No registry or market was changed here.
            </Alert>
          )}
        </Card>
      </Container>
    </Box>
  );
}
