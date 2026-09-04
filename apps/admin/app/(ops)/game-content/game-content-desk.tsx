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
  TextField,
  Typography,
} from "@mui/material";
import { FormEvent, useEffect, useRef, useState } from "react";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon } from "../../admin-icons";
import {
  promptErrors,
  validPromptResult,
  type PromptInput,
} from "../../content-model";
import { adminFetch } from "../../lib/admin-fetch";
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
      const response = await adminFetch("/api/ebe-prompts", {
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
    <Box className="game-content-redesign">
      <Box component="header" className="game-content-hero">
        <AdminCardWatermark watermark="evidence" />
        <Box className="game-content-hero-copy">
          <Box className="game-content-kicker">
            <AdminIcon name="games" aria-hidden="true" />
            <Typography className="section-kicker">
              REVIEWED GAME CONTENT
            </Typography>
          </Box>
          <Typography component="h1">
            Preserve the cue. Prove the source.
          </Typography>
          <Typography className="game-content-hero-intro">
            Approve sourced Ɛbɛ prompts for private play. Each decision binds a
            public cue to a reviewed source and a server-private answer set.
          </Typography>
        </Box>
        <Box
          className="game-content-hero-register"
          aria-label="Review doctrine"
        >
          <div>
            <span>Public surface</span>
            <strong>Prompt cue only</strong>
            <Typography>Players never receive accepted forms</Typography>
          </div>
          <div>
            <span>Revision model</span>
            <strong>Append, never overwrite</strong>
            <Typography>Corrections require a higher version</Typography>
          </div>
          <div>
            <span>Authority</span>
            <strong>Authenticated reviewer</strong>
            <Typography>Every approval retains operator provenance</Typography>
          </div>
        </Box>
      </Box>

      <Box component="section" className="game-content-boundary">
        <Box className="game-content-boundary-icon">
          <AdminIcon name="games" aria-hidden="true" />
        </Box>
        <Box>
          <Typography className="section-kicker">REVIEW BOUNDARY</Typography>
          <Typography component="h2">
            The answer key stays backstage.
          </Typography>
          <Typography>
            Reviewers may verify normalized answers here. Members only receive
            the approved cue, while evaluation remains server-side.
          </Typography>
        </Box>
        <span className="game-content-boundary-state">PRIVATE BY DESIGN</span>
      </Box>

      {message ? (
        <Alert severity={success ? "success" : "error"} role="alert">
          {message}
        </Alert>
      ) : null}
      <Box className="game-content-workspace">
        <Box component="aside" className="game-content-review-rail">
          <Typography className="section-kicker">APPROVAL SEQUENCE</Typography>
          <Typography component="h2">One record, three proofs.</Typography>
          <ol>
            <li>
              <span>01</span>
              <div>
                <strong>Identity</strong>
                <Typography>
                  Fix the stable ID, version, and language.
                </Typography>
              </div>
            </li>
            <li>
              <span>02</span>
              <div>
                <strong>Substance</strong>
                <Typography>
                  Review the cue and normalized private answers.
                </Typography>
              </div>
            </li>
            <li>
              <span>03</span>
              <div>
                <strong>Provenance</strong>
                <Typography>
                  Attach a source another operator can verify.
                </Typography>
              </div>
            </li>
          </ol>
          <Box className="game-content-rail-note">
            <AdminCardWatermark watermark="evidence" />
            <span>Decision effect</span>
            <strong>Immediate for new private duels</strong>
            <Typography>
              Existing duels retain their original reviewed snapshot.
            </Typography>
          </Box>
        </Box>
        <AdminCard
          component="form"
          onSubmit={review}
          variant="form"
          watermark="evidence"
          className="game-content-form"
        >
          <Box className="game-content-form-heading">
            <Box>
              <Typography className="section-kicker">NEW REVIEW</Typography>
              <Typography component="h2">Build the approval record</Typography>
            </Box>
            <span>ALL REQUIRED EXCEPT LOCATOR</span>
          </Box>
          <section className="game-content-form-section">
            <Box className="game-content-section-index">
              <span>01</span>
              <div>
                <strong>Record identity</strong>
                <Typography>
                  The durable coordinates for this revision.
                </Typography>
              </div>
            </Box>
            <Box className="game-content-field-grid game-content-field-grid--identity">
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
                slotProps={{
                  htmlInput: { min: 1, step: 1, inputMode: "numeric" },
                }}
                {...field("version")}
              />
              <TextField
                label="Language"
                value={form.language}
                onChange={(e) => setForm({ ...form, language: e.target.value })}
                {...field("language", "BCP 47, for example tw or en-GH")}
              />
            </Box>
          </section>
          <section className="game-content-form-section">
            <Box className="game-content-section-index">
              <span>02</span>
              <div>
                <strong>Prompt substance</strong>
                <Typography>
                  What members see and what only the server knows.
                </Typography>
              </div>
            </Box>
            <Box className="game-content-field-grid">
              <TextField
                label="Prompt or proverb cue"
                multiline
                rows={4}
                value={form.cue}
                onChange={(e) => setForm({ ...form, cue: e.target.value })}
                {...field("cue", "The reviewed prompt shown to both players")}
              />
              <TextField
                className="game-content-private-field"
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
            </Box>
          </section>
          <section className="game-content-form-section">
            <Box className="game-content-section-index">
              <span>03</span>
              <div>
                <strong>Source provenance</strong>
                <Typography>The evidence supporting this prompt.</Typography>
              </div>
            </Box>
            <Box className="game-content-field-grid game-content-field-grid--source">
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
                {...field(
                  "locator",
                  "Optional HTTPS archive or catalogue record",
                )}
              />
            </Box>
          </section>
          <Box className="game-content-finality">
            <AdminIcon name="governance" aria-hidden="true" />
            <Typography>
              Approval is final for this ID and version. Corrections require a
              higher version; existing duels keep their reviewed snapshot.
            </Typography>
          </Box>
          <Button
            className="game-content-submit"
            type="submit"
            variant="contained"
            disabled={busy}
            aria-busy={busy}
          >
            Review immutable prompt
          </Button>
        </AdminCard>
      </Box>
      <Dialog
        className="game-content-confirm-dialog"
        open={Boolean(pending)}
        fullWidth
        maxWidth="md"
        onClose={() => {
          if (!busy) {
            setPending(null);
            setConfirmError("");
          }
        }}
        aria-labelledby="prompt-confirm-title"
        aria-describedby="prompt-confirm-description"
      >
        <DialogTitle
          id="prompt-confirm-title"
          className="game-content-confirm-title"
        >
          <span className="game-content-confirm-icon">
            <AdminIcon name="games" aria-hidden="true" />
          </span>
          <span>
            <small>IMMUTABLE REVIEW</small>Confirm prompt approval
          </span>
        </DialogTitle>
        <DialogContent>
          <DialogContentText
            id="prompt-confirm-description"
            className="game-content-confirm-intro"
          >
            Verify the immutable public prompt and private answer count.
          </DialogContentText>
          {confirmError ? (
            <Alert severity="error" role="alert">
              {confirmError}
            </Alert>
          ) : null}
          {pending ? (
            <Box className="game-content-confirm-docket">
              <Box className="game-content-confirm-meta">
                <div>
                  <span>ID/version:</span>
                  <strong>
                    {pending.id} · v{pending.version}
                  </strong>
                </div>
                <div>
                  <span>Language:</span>
                  <strong>{pending.language}</strong>
                </div>
              </Box>
              <Box className="game-content-confirm-cue">
                <span>PUBLIC PROMPT CUE</span>
                <Typography>“{pending.cue}”</Typography>
              </Box>
              <Box className="game-content-confirm-proof-grid">
                <Box className="game-content-confirm-private">
                  <span>SERVER-PRIVATE MATERIAL</span>
                  <strong>
                    {pending.acceptedAnswers.length.toString().padStart(2, "0")}
                  </strong>
                  <Typography>
                    <strong>Accepted answer count:</strong>
                  </Typography>
                  <small>
                    Answers are counted here and never projected to members.
                  </small>
                </Box>
                <Box className="game-content-confirm-source">
                  <span>PROVENANCE</span>
                  <Typography>
                    <strong>Source:</strong> {pending.source.kind} ·{" "}
                    {pending.source.citation}
                  </Typography>
                  <Typography>
                    <strong>Source locator:</strong>{" "}
                    {pending.source.locator ?? "None"}
                  </Typography>
                </Box>
              </Box>
              <Box className="game-content-confirm-warning">
                <AdminIcon name="governance" aria-hidden="true" />
                <Typography>
                  This exact revision becomes available to new private duels. It
                  cannot be edited after approval.
                </Typography>
              </Box>
            </Box>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            className="game-content-confirm-approve"
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
    </Box>
  );
}
