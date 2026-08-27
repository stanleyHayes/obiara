"use client";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { FormEvent, useEffect, useRef, useState } from "react";
import { AdminCard } from "../../admin-card";
import {
  promptErrors,
  validPromptResult,
  type PromptInput,
} from "../../content-model";
const initial: PromptInput = {
  id: "",
  version: 1,
  language: "",
  cue: "",
  acceptedAnswers: [],
  source: { kind: "book", citation: "" },
};
export function GameContentDesk() {
  const [form, setForm] = useState(initial),
    [answers, setAnswers] = useState(""),
    [errors, setErrors] = useState<Record<string, string>>({}),
    [pending, setPending] = useState<PromptInput | null>(null),
    [busy, setBusy] = useState(false),
    [confirmError, setConfirmError] = useState(""),
    [message, setMessage] = useState(""),
    [success, setSuccess] = useState(false);
  const mounted = useRef(true),
    generation = useRef(0),
    formRef = useRef<HTMLFormElement | null>(null);
  useEffect(() => {
    mounted.current = true;
    const actions = generation;
    return () => {
      mounted.current = false;
      actions.current++;
    };
  }, []);
  function snapshot(): PromptInput {
    return {
      ...form,
      id: form.id.trim(),
      language: form.language.trim(),
      cue: form.cue.trim(),
      acceptedAnswers: answers
        .split("\n")
        .map((x) => x.trim())
        .filter(Boolean),
      source: {
        ...form.source,
        citation: form.source.citation.trim(),
        locator: form.source.locator?.trim() || undefined,
      },
    };
  }
  function review(e: FormEvent) {
    e.preventDefault();
    formRef.current = e.currentTarget as HTMLFormElement;
    const value = snapshot(),
      next = promptErrors(value);
    setErrors(next);
    const first = Object.keys(next)[0];
    if (first) {
      formRef.current?.querySelector<HTMLElement>(`[name="${first}"]`)?.focus();
      return;
    }
    setPending(value);
    setConfirmError("");
    setMessage("");
  }
  async function submit() {
    if (!pending) return;
    const value = pending,
      id = ++generation.current;
    setBusy(true);
    setConfirmError("");
    try {
      const response = await fetch("/api/ebe-prompts", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(value),
        }),
        payload: unknown = await response.json().catch(() => null);
      if (!mounted.current || id !== generation.current) return;
      if (!response.ok || !validPromptResult(payload, value))
        throw new Error(
          payload &&
            typeof payload === "object" &&
            "message" in payload &&
            typeof payload.message === "string"
            ? payload.message
            : "The prompt could not be approved.",
        );
      setSuccess(true);
      setMessage(
        `${value.id} revision ${value.version} is now available to new private duels.`,
      );
      setPending(null);
      setConfirmError("");
      setForm(initial);
      setAnswers("");
      formRef.current?.reset();
    } catch (e) {
      if (mounted.current && id === generation.current) {
        setConfirmError(
          e instanceof Error ? e.message : "The prompt could not be approved.",
        );
      }
    } finally {
      if (mounted.current && id === generation.current) setBusy(false);
    }
  }
  const field = (name: string, fallback?: string) => ({
    name,
    error: Boolean(errors[name]),
    helperText: errors[name] ?? fallback,
  });
  return (
    <Stack spacing={3}>
      <Box>
        <Typography className="section-kicker">
          REVIEWED GAME CONTENT
        </Typography>
        <Typography component="h1">Approve sourced Ɛbɛ prompts.</Typography>
        <Typography color="text.secondary">
          Accepted forms remain server-private. Approval creates an immutable
          version tied to this authenticated operator and supplied source.
        </Typography>
      </Box>
      {message ? (
        <Alert severity={success ? "success" : "error"} role="alert">
          {message}
        </Alert>
      ) : null}
      <AdminCard
        component="form"
        onSubmit={review}
        variant="form"
        watermark="evidence"
      >
        <Stack spacing={2.5}>
          <TextField
            label="Stable prompt ID"
            value={form.id}
            onChange={(e) => setForm({ ...form, id: e.target.value })}
            {...field("id")}
          />
          <TextField
            label="Version"
            type="number"
            value={form.version}
            onChange={(e) =>
              setForm({ ...form, version: Number(e.target.value) })
            }
            slotProps={{ htmlInput: { min: 1, step: 1, inputMode: "numeric" } }}
            {...field("version")}
          />
          <TextField
            label="Language"
            value={form.language}
            onChange={(e) => setForm({ ...form, language: e.target.value })}
            {...field("language", "BCP 47, for example tw or en-GH")}
          />
          <TextField
            label="Prompt or proverb cue"
            multiline
            rows={4}
            value={form.cue}
            onChange={(e) => setForm({ ...form, cue: e.target.value })}
            {...field("cue", "The reviewed prompt shown to both players")}
          />
          <TextField
            label="Accepted answer forms"
            multiline
            rows={5}
            value={answers}
            onChange={(e) => setAnswers(e.target.value)}
            {...field(
              "acceptedAnswers",
              "One accepted normalized form per line. Never projected to members.",
            )}
          />
          <TextField
            select
            label="Source type"
            value={form.source.kind}
            onChange={(e) =>
              setForm({
                ...form,
                source: {
                  ...form.source,
                  kind: e.target.value as PromptInput["source"]["kind"],
                },
              })
            }
          >
            <MenuItem value="book">Book</MenuItem>
            <MenuItem value="oral_archive">Oral archive</MenuItem>
            <MenuItem value="institutional_archive">
              Institutional archive
            </MenuItem>
          </TextField>
          <TextField
            label="Full source citation"
            value={form.source.citation}
            onChange={(e) =>
              setForm({
                ...form,
                source: { ...form.source, citation: e.target.value },
              })
            }
            {...field("citation")}
          />
          <TextField
            label="Source locator"
            type="url"
            value={form.source.locator ?? ""}
            onChange={(e) =>
              setForm({
                ...form,
                source: { ...form.source, locator: e.target.value },
              })
            }
            {...field("locator", "Optional HTTPS archive or catalogue record")}
          />
          <Alert severity="warning">
            Approval is final for this ID and version. Corrections require a
            higher version; existing duels keep their reviewed snapshot.
          </Alert>
          <Button
            type="submit"
            variant="contained"
            disabled={busy}
            aria-busy={busy}
          >
            Review immutable prompt
          </Button>
        </Stack>
      </AdminCard>
      <Dialog
        open={Boolean(pending)}
        onClose={() => {
          if (!busy) {
            setPending(null);
            setConfirmError("");
          }
        }}
        aria-labelledby="prompt-confirm-title"
        aria-describedby="prompt-confirm-description"
      >
        <DialogTitle id="prompt-confirm-title">
          Confirm prompt approval
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="prompt-confirm-description">
            Verify the immutable public prompt and private answer count.
          </DialogContentText>
          {confirmError ? (
            <Alert severity="error" role="alert">
              {confirmError}
            </Alert>
          ) : null}
          {pending ? (
            <Stack spacing={1}>
              <Typography>
                <strong>ID/version:</strong> {pending.id} · v{pending.version}
              </Typography>
              <Typography>
                <strong>Language:</strong> {pending.language}
              </Typography>
              <Typography>
                <strong>Cue:</strong> {pending.cue}
              </Typography>
              <Typography>
                <strong>Accepted answer count:</strong>{" "}
                {pending.acceptedAnswers.length}
              </Typography>
              <Typography>
                <strong>Source:</strong> {pending.source.kind} ·{" "}
                {pending.source.citation}
              </Typography>
              <Typography>
                <strong>Source locator:</strong>{" "}
                {pending.source.locator ?? "None"}
              </Typography>
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setPending(null);
              setConfirmError("");
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={busy}
            aria-busy={busy}
            onClick={() => void submit()}
            variant="contained"
          >
            Approve immutable prompt
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
