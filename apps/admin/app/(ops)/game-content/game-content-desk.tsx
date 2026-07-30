"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { FormEvent, useState } from "react";

export function GameContentDesk() {
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const answers = String(form.get("acceptedAnswers") ?? "")
      .split("\n")
      .map((value) => value.trim())
      .filter(Boolean);
    setBusy(true);
    setMessage("");
    setSuccess(false);
    try {
      const response = await fetch("/api/ebe-prompts", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: String(form.get("id") ?? "").trim(),
          version: Number(form.get("version")),
          language: String(form.get("language") ?? "").trim(),
          cue: String(form.get("cue") ?? "").trim(),
          acceptedAnswers: answers,
          source: {
            kind: String(form.get("sourceKind")),
            citation: String(form.get("citation") ?? "").trim(),
            locator: String(form.get("locator") ?? "").trim() || undefined,
          },
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        id?: string;
        version?: number;
        message?: string;
      } | null;
      if (!response.ok || !payload?.id)
        throw new Error(
          payload?.message || "The prompt could not be approved.",
        );
      setSuccess(true);
      setMessage(
        `${payload.id} revision ${payload.version} is now available to new private duels.`,
      );
      event.currentTarget.reset();
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The prompt could not be approved.",
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
          REVIEWED GAME CONTENT
        </Typography>
        <Typography
          component="h1"
          sx={{
            fontSize: { xs: 36, md: 52 },
            fontWeight: 800,
            letterSpacing: -2,
          }}
        >
          Approve sourced Ɛbɛ prompts.
        </Typography>
        <Typography sx={{ color: "text.secondary", maxWidth: 760 }}>
          Accepted forms remain server-private. Approval creates an immutable
          version tied to this authenticated operator and the supplied source.
        </Typography>
      </Box>
      {message ? (
        <Alert severity={success ? "success" : "error"}>{message}</Alert>
      ) : null}
      <Card
        component="form"
        onSubmit={submit}
        sx={{ maxWidth: 900, p: { xs: 2.5, md: 4 } }}
      >
        <Stack spacing={2.5}>
          <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
            <TextField fullWidth label="Stable prompt ID" name="id" required />
            <TextField
              fullWidth
              label="Version"
              name="version"
              required
              slotProps={{ htmlInput: { min: 1 } }}
              type="number"
            />
            <TextField
              fullWidth
              helperText="BCP 47, for example tw or en-GH"
              label="Language"
              name="language"
              required
            />
          </Stack>
          <TextField
            helperText="The reviewed prompt shown to both players"
            label="Prompt or proverb cue"
            multiline
            name="cue"
            required
            rows={4}
          />
          <TextField
            helperText="One accepted normalized form per line. Never projected to members."
            label="Accepted answer forms"
            multiline
            name="acceptedAnswers"
            required
            rows={5}
          />
          <TextField
            defaultValue="book"
            label="Source type"
            name="sourceKind"
            required
            select
          >
            <MenuItem value="book">Book</MenuItem>
            <MenuItem value="oral_archive">Oral archive</MenuItem>
            <MenuItem value="institutional_archive">
              Institutional archive
            </MenuItem>
          </TextField>
          <TextField label="Full source citation" name="citation" required />
          <TextField
            helperText="Optional HTTPS archive or catalogue record"
            label="Source locator"
            name="locator"
            type="url"
          />
          <Alert severity="warning">
            Approval is final for this ID and version. Corrections require a
            higher version; existing duels keep their original reviewed
            snapshot.
          </Alert>
          <Button
            disabled={busy}
            size="large"
            type="submit"
            variant="contained"
          >
            {busy ? "Approving…" : "Approve immutable prompt"}
          </Button>
        </Stack>
      </Card>
    </Stack>
  );
}
