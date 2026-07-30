"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Divider,
  LinearProgress,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { FormEvent, useState } from "react";

type Cohort = {
  id: string;
  capacity: 4 | 8 | 16;
  enrolled: number;
  joined: boolean;
  status: "open" | "locked" | "started";
  competitionId?: string;
  revision: number;
};
type Competition = {
  id: string;
  revision: number;
  status: "active" | "completed";
  reviews: {
    id: string;
    matchId: string;
    status: "open" | "resolved" | "appealed" | "final";
    decision: "none" | "no_action" | "rules_action";
  }[];
};

function statusColour(status: Cohort["status"]) {
  if (status === "started") return "success" as const;
  if (status === "locked") return "warning" as const;
  return "info" as const;
}

export function TournamentDesk() {
  const [cohort, setCohort] = useState<Cohort | null>(null);
  const [competition, setCompetition] = useState<Competition | null>(null);
  const [reference, setReference] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function readResponse(response: Response) {
    const payload = (await response.json().catch(() => null)) as
      (Cohort & { message?: string }) | null;
    if (!response.ok)
      throw new Error(
        payload?.message ||
          "The tournament desk could not complete that action.",
      );
    return payload;
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const capacity = Number(
      new FormData(event.currentTarget).get("capacity"),
    ) as 4 | 8 | 16;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/game-cohorts", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        body: JSON.stringify({ action: "create", capacity }),
      });
      const value = await readResponse(response);
      if (!value?.id) throw new Error("The cohort response was incomplete.");
      setCohort(value);
      setCompetition(null);
      setReference(value.id);
      setMessage(
        "Private cohort created. Share only its invitation reference.",
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The cohort could not be created.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function load(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const id = reference.trim();
    if (!id) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch(
        `/api/game-cohorts?cohortId=${encodeURIComponent(id)}`,
      );
      const value = await readResponse(response);
      if (!value?.id) throw new Error("The cohort response was incomplete.");
      setCohort(value);
      if (value.competitionId)
        await loadCompetition(value.id, value.competitionId);
      else setCompetition(null);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The cohort could not be opened.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function start() {
    if (!cohort) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/game-cohorts", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        body: JSON.stringify({
          action: "start",
          cohortId: cohort.id,
          expectedRevision: cohort.revision,
        }),
      });
      const value = (await response.json().catch(() => null)) as {
        cohort?: Cohort;
        message?: string;
      } | null;
      if (!response.ok || !value?.cohort)
        throw new Error(value?.message || "The bracket could not be started.");
      setCohort(value.cohort);
      if (value.cohort.competitionId)
        await loadCompetition(value.cohort.id, value.cohort.competitionId);
      setMessage(
        "Bracket started. Entrants can now open the privacy-safe ladder.",
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The bracket could not be started.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function loadCompetition(cohortId: string, competitionId: string) {
    const response = await fetch(
      `/api/game-cohorts?cohortId=${encodeURIComponent(cohortId)}&competitionId=${encodeURIComponent(competitionId)}`,
    );
    const payload = (await response.json().catch(() => null)) as
      (Competition & { message?: string }) | null;
    if (!response.ok || !payload?.id)
      throw new Error(
        payload?.message || "The competition review desk could not be opened.",
      );
    setCompetition(payload);
  }

  async function resolve(
    reviewId: string,
    decision: "no_action" | "rules_action",
    appeal: boolean,
  ) {
    if (!cohort || !competition) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/game-cohorts", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        body: JSON.stringify({
          action: appeal ? "resolve-appeal" : "resolve-review",
          cohortId: cohort.id,
          competitionId: competition.id,
          reviewId,
          decision,
          expectedRevision: competition.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Competition & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(
          payload?.message || "The review decision could not be retained.",
        );
      setCompetition(payload);
      setMessage(
        "Human review decision retained with its immutable evidence reference.",
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The review decision could not be retained.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <Stack spacing={3}>
      <Box>
        <Typography
          sx={{
            color: "text.secondary",
            fontSize: 12,
            fontWeight: 800,
            letterSpacing: 1.4,
          }}
        >
          PRIVATE COMPETITION CONTROL
        </Typography>
        <Typography
          component="h1"
          sx={{
            fontSize: { xs: 36, md: 52 },
            fontWeight: 800,
            letterSpacing: -2,
          }}
        >
          Run invitation-only tournaments.
        </Typography>
        <Typography sx={{ color: "text.secondary", maxWidth: 760 }}>
          Create a fixed cohort, share its private reference, then start the
          bracket after every seat is claimed. No discovery, popularity ranking,
          or member directory is exposed.
        </Typography>
      </Box>
      {message ? (
        <Alert
          severity={
            message.includes("created") || message.includes("started")
              ? "success"
              : "info"
          }
        >
          {message}
        </Alert>
      ) : null}
      <Stack
        direction={{ xs: "column", lg: "row" }}
        spacing={3}
        sx={{ alignItems: "stretch" }}
      >
        <Card
          component="form"
          onSubmit={create}
          sx={{ flex: 1, p: { xs: 2.5, md: 4 } }}
        >
          <Stack spacing={2.5}>
            <Typography component="h2" variant="h5" sx={{ fontWeight: 800 }}>
              Create a cohort
            </Typography>
            <TextField
              defaultValue={4}
              label="Entrant capacity"
              name="capacity"
              select
            >
              {[4, 8, 16].map((capacity) => (
                <MenuItem key={capacity} value={capacity}>
                  {capacity} entrants
                </MenuItem>
              ))}
            </TextField>
            <Alert severity="info">
              Seats are opt-in. The cohort locks automatically when full;
              performance never changes member matching or trust.
            </Alert>
            <Button
              disabled={busy}
              size="large"
              type="submit"
              variant="contained"
            >
              {busy ? "Working…" : "Create private cohort"}
            </Button>
          </Stack>
        </Card>
        <Card
          component="form"
          onSubmit={load}
          sx={{ flex: 1, p: { xs: 2.5, md: 4 } }}
        >
          <Stack spacing={2.5}>
            <Typography component="h2" variant="h5" sx={{ fontWeight: 800 }}>
              Resume control
            </Typography>
            <TextField
              label="Private cohort reference"
              onChange={(event) => setReference(event.target.value)}
              required
              value={reference}
            />
            <Button
              disabled={busy || !reference.trim()}
              size="large"
              type="submit"
              variant="outlined"
            >
              Open cohort
            </Button>
          </Stack>
        </Card>
      </Stack>
      {cohort ? (
        <Card sx={{ p: { xs: 2.5, md: 4 } }}>
          <Stack spacing={2.5}>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              sx={{ justifyContent: "space-between", gap: 2 }}
            >
              <Box>
                <Typography
                  color="text.secondary"
                  sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}
                >
                  INVITATION REFERENCE
                </Typography>
                <Typography
                  component="h2"
                  sx={{
                    fontFamily: "monospace",
                    fontSize: { xs: 22, md: 30 },
                    overflowWrap: "anywhere",
                  }}
                >
                  {cohort.id}
                </Typography>
              </Box>
              <Chip
                color={statusColour(cohort.status)}
                label={cohort.status.toUpperCase()}
              />
            </Stack>
            <Divider />
            <Box>
              <Stack direction="row" sx={{ justifyContent: "space-between" }}>
                <Typography sx={{ fontWeight: 750 }}>Seats claimed</Typography>
                <Typography>
                  {cohort.enrolled} / {cohort.capacity}
                </Typography>
              </Stack>
              <LinearProgress
                sx={{ height: 10, borderRadius: 5, mt: 1 }}
                value={(cohort.enrolled / cohort.capacity) * 100}
                variant="determinate"
              />
            </Box>
            {cohort.competitionId ? (
              <Alert severity="success">
                Competition reference: {cohort.competitionId}
              </Alert>
            ) : null}
            <Button
              disabled={busy || cohort.status !== "locked"}
              onClick={start}
              size="large"
              variant="contained"
            >
              {cohort.status === "open"
                ? "Waiting for every seat"
                : cohort.status === "started"
                  ? "Bracket already started"
                  : "Start bracket"}
            </Button>
          </Stack>
        </Card>
      ) : null}
      {competition?.reviews.length ? (
        <Card sx={{ p: { xs: 2.5, md: 4 } }}>
          <Stack spacing={2.5}>
            <Box>
              <Typography
                color="text.secondary"
                sx={{ fontSize: 12, fontWeight: 800, letterSpacing: 1.2 }}
              >
                NEUTRAL FAIR-PLAY REVIEW
              </Typography>
              <Typography component="h2" variant="h5" sx={{ fontWeight: 800 }}>
                Decide from server evidence, never accusation.
              </Typography>
            </Box>
            {competition.reviews.map((review) => (
              <Card key={review.id} variant="outlined" sx={{ p: 2.5 }}>
                <Stack spacing={2}>
                  <Typography sx={{ fontWeight: 750 }}>
                    {review.matchId} · {review.status}
                  </Typography>
                  <Typography color="text.secondary">
                    Current decision: {review.decision.replaceAll("_", " ")}
                  </Typography>
                  {review.status === "open" || review.status === "appealed" ? (
                    <Stack
                      direction={{ xs: "column", sm: "row" }}
                      spacing={1.5}
                    >
                      <Button
                        disabled={busy}
                        onClick={() =>
                          void resolve(
                            review.id,
                            "no_action",
                            review.status === "appealed",
                          )
                        }
                        variant="outlined"
                      >
                        No action
                      </Button>
                      <Button
                        color="warning"
                        disabled={busy}
                        onClick={() =>
                          void resolve(
                            review.id,
                            "rules_action",
                            review.status === "appealed",
                          )
                        }
                        variant="contained"
                      >
                        Rules action
                      </Button>
                    </Stack>
                  ) : null}
                </Stack>
              </Card>
            ))}
          </Stack>
        </Card>
      ) : null}
    </Stack>
  );
}
