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
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Skeleton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useEffect, useState } from "react";

type EvidencePurpose = "triage" | "appeal" | "legal";

interface SafetyCase {
  caseId: string;
  subjectRef: string;
  tier: "A" | "B" | "C" | "D";
  queue: "triage" | "care";
  status: "queued" | "in_review" | "resolved";
  slaDueAt: string;
  assigned: boolean;
  assignedToMe: boolean;
  version: number;
}

interface SafetyEvidence {
  caseId: string;
  subjectRef: string;
  tier: string;
  category: string;
  surface: string;
  contextRef?: string;
  description?: string;
}

function deadlineLabel(value: string) {
  const due = new Date(value);
  const difference = due.getTime() - Date.now();
  const hours = Math.ceil(Math.abs(difference) / 3_600_000);
  return difference < 0 ? `${hours}h overdue` : `${hours}h remaining`;
}

function SafetyQueueItem({
  item,
  selected,
  onSelect,
}: Readonly<{ item: SafetyCase; selected: boolean; onSelect: () => void }>) {
  return (
    <Button aria-pressed={selected} className="safety-case" onClick={onSelect}>
      <Box>
        <Stack direction="row" spacing={1}>
          <Typography component="strong">{item.caseId}</Typography>
          <Chip
            color={item.tier === "A" ? "error" : "warning"}
            label={`Tier ${item.tier}`}
            size="small"
          />
        </Stack>
        <Typography>
          {item.queue === "care" ? "Care response" : "Safety triage"}
        </Typography>
        <Typography className="safety-reference">
          {item.subjectRef} · {deadlineLabel(item.slaDueAt)}
        </Typography>
      </Box>
      <span aria-hidden="true">›</span>
    </Button>
  );
}

export function SafetyDesk() {
  const [cases, setCases] = useState<SafetyCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [purpose, setPurpose] = useState<EvidencePurpose>("triage");
  const [evidence, setEvidence] = useState<SafetyEvidence | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState("");
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const selected = cases.find((item) => item.caseId === selectedID);

  async function loadQueue() {
    setLoading(true);
    setMessage("");
    try {
      const response = await fetch("/api/safety");
      const payload = (await response.json()) as {
        cases?: SafetyCase[];
        message?: string;
      };
      if (!response.ok)
        throw new Error(
          payload.message || "The safety queue could not be loaded.",
        );
      const next = payload.cases ?? [];
      setCases(next);
      setSelectedID((current) =>
        next.some((item) => item.caseId === current)
          ? current
          : (next[0]?.caseId ?? ""),
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The safety queue could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    void fetch("/api/safety")
      .then(async (response) => {
        const payload = (await response.json()) as {
          cases?: SafetyCase[];
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "The safety queue could not be loaded.",
          );
        if (active) {
          const next = payload.cases ?? [];
          setCases(next);
          setSelectedID(next[0]?.caseId ?? "");
        }
      })
      .catch((error: unknown) => {
        if (active)
          setMessage(
            error instanceof Error
              ? error.message
              : "The safety queue could not be loaded.",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function assignCase() {
    if (!selected) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/safety", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "assign", caseId: selected.caseId }),
      });
      const payload = (await response.json().catch(() => null)) as
        (SafetyCase & { message?: string }) | null;
      if (!response.ok || !payload?.caseId)
        throw new Error(payload?.message || "The case could not be assigned.");
      setCases((current) =>
        current.map((item) =>
          item.caseId === payload.caseId ? payload : item,
        ),
      );
      setSuccess(
        `${payload.caseId} is assigned to you. Evidence remains sealed until fresh MFA.`,
      );
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The case could not be assigned.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function requestEvidence() {
    if (!selected?.assignedToMe) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/safety", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          action: "evidence",
          caseId: selected.caseId,
          purpose,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (SafetyEvidence & { message?: string }) | null;
      if (!response.ok || !payload?.caseId) {
        if (response.status === 403) setStepUpOpen(true);
        throw new Error(
          payload?.message || "Redacted evidence could not be opened.",
        );
      }
      setEvidence(payload);
      setSuccess(`Audited ${purpose} access recorded for ${payload.caseId}.`);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Redacted evidence could not be opened.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function stepUp(action: "start" | "complete") {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "start" ? { action } : { action, code: stepUpCode },
        ),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          payload?.message || "The MFA step-up could not be completed.",
        );
      if (action === "start") {
        setSuccess("A fresh step-up code was sent to your admin email.");
      } else {
        setStepUpOpen(false);
        setStepUpCode("");
        setSuccess("Fresh MFA verified. Opening the assigned evidence now.");
        await requestEvidence();
      }
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The MFA step-up could not be completed.",
      );
    } finally {
      setBusy(false);
    }
  }

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
            Real queued cases, privacy-keyed subjects and purpose-bound evidence
            access.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip label={`${cases.length} queued`} color="warning" />
          <Chip label="Evidence access audited" color="success" />
        </Stack>
      </header>

      {message ? (
        <Alert severity="error" className="verification-alert">
          {message}
        </Alert>
      ) : null}
      {success ? (
        <Alert
          severity="success"
          className="verification-alert"
          onClose={() => setSuccess("")}
        >
          {success}
        </Alert>
      ) : null}

      <Box className="verification-grid">
        <Card className="verification-list">
          <Box className="verification-panel-heading">
            <Typography component="h2">Priority queue</Typography>
            <Button
              disabled={loading}
              onClick={() => void loadQueue()}
              size="small"
            >
              Refresh
            </Button>
          </Box>
          <Box aria-label="Trust and safety cases">
            {loading ? (
              <Stack spacing={1.5} sx={{ p: 2 }}>
                <Skeleton height={76} />
                <Skeleton height={76} />
                <Skeleton height={76} />
              </Stack>
            ) : cases.length ? (
              cases.map((item) => (
                <SafetyQueueItem
                  item={item}
                  key={item.caseId}
                  onSelect={() => setSelectedID(item.caseId)}
                  selected={item.caseId === selectedID}
                />
              ))
            ) : (
              <Alert severity="success">No triage cases are waiting.</Alert>
            )}
          </Box>
        </Card>

        <Card className="verification-review">
          {selected ? (
            <Stack spacing={3}>
              <Box className="verification-panel-heading">
                <Box>
                  <Typography className="section-kicker">
                    Case {selected.caseId}
                  </Typography>
                  <Typography component="h2">
                    Controlled evidence review
                  </Typography>
                </Box>
                <Chip
                  label={
                    selected.assignedToMe
                      ? "Assigned to you"
                      : selected.assigned
                        ? "Assigned"
                        : "Unassigned"
                  }
                  color={selected.assignedToMe ? "success" : "default"}
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
                  <Typography>Severity</Typography>
                  <Typography component="strong">
                    Tier {selected.tier}
                  </Typography>
                </Box>
                <Box>
                  <Typography>SLA</Typography>
                  <Typography component="strong">
                    {deadlineLabel(selected.slaDueAt)}
                  </Typography>
                </Box>
              </Box>
              {!selected.assigned ? (
                <Alert severity="info">
                  Claim the case before requesting its least-exposure evidence
                  bundle.
                </Alert>
              ) : !selected.assignedToMe ? (
                <Alert severity="warning">
                  This case belongs to another agent. Evidence access is
                  blocked.
                </Alert>
              ) : (
                <Alert severity="info">
                  Choose the exact review purpose. Opening evidence requires
                  fresh MFA and creates an immutable access record.
                </Alert>
              )}
              <Box className="verification-actions">
                {!selected.assigned ? (
                  <Button
                    disabled={busy}
                    onClick={() => void assignCase()}
                    variant="contained"
                  >
                    Assign to me
                  </Button>
                ) : null}
                <FormControl sx={{ minWidth: 180 }}>
                  <InputLabel id="safety-purpose-label">
                    Access purpose
                  </InputLabel>
                  <Select
                    label="Access purpose"
                    labelId="safety-purpose-label"
                    onChange={(event) =>
                      setPurpose(event.target.value as EvidencePurpose)
                    }
                    value={purpose}
                  >
                    <MenuItem value="triage">Triage</MenuItem>
                    <MenuItem value="appeal">Appeal</MenuItem>
                    <MenuItem value="legal">Legal</MenuItem>
                  </Select>
                </FormControl>
                <Button
                  disabled={busy || !selected.assignedToMe}
                  onClick={() => void requestEvidence()}
                  variant="outlined"
                >
                  Open redacted evidence
                </Button>
              </Box>
            </Stack>
          ) : (
            <Alert severity="info">Select a queued case to begin.</Alert>
          )}
        </Card>
      </Box>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setEvidence(null)}
        open={Boolean(evidence)}
      >
        <DialogTitle>Redacted case evidence</DialogTitle>
        <DialogContent>
          {evidence ? (
            <Stack spacing={2}>
              <Alert severity="warning">
                Purpose: {purpose}. Access has been logged.
              </Alert>
              <Box className="verification-facts">
                <Box>
                  <Typography>Subject</Typography>
                  <Typography component="strong">
                    {evidence.subjectRef}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Category</Typography>
                  <Typography component="strong">
                    {evidence.category}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Surface</Typography>
                  <Typography component="strong">{evidence.surface}</Typography>
                </Box>
              </Box>
              {evidence.contextRef ? (
                <Typography>Context: {evidence.contextRef}</Typography>
              ) : null}
              <Alert severity="info">
                {evidence.description ||
                  "No free-text description was supplied."}
              </Alert>
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEvidence(null)}>Close evidence</Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="xs"
        onClose={() => setStepUpOpen(false)}
        open={stepUpOpen}
      >
        <DialogTitle>Fresh MFA required</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="info">
              The evidence remains sealed until this session completes a fresh
              step-up.
            </Alert>
            <TextField
              autoComplete="one-time-code"
              label="Step-up code"
              onChange={(event) => setStepUpCode(event.target.value)}
              value={stepUpCode}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={() => void stepUp("start")}>
            Send code
          </Button>
          <Button
            disabled={busy || stepUpCode.trim().length < 6}
            onClick={() => void stepUp("complete")}
            variant="contained"
          >
            Verify and open
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
