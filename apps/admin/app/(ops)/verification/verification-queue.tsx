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
import { useEffect, useRef, useState } from "react";

type Outcome = "approve" | "reject";

interface VerificationCase {
  caseId: string;
  subjectRef: string;
  reasonCode: "provider_uncertain" | "provider_outage" | "manual_review";
  submittedAt: string;
  status?: string;
  version: number;
}

interface Evidence {
  caseId: string;
  maskedCard: string;
  ageBand: string;
  providerStatus: string;
}

const reasonLabels: Record<VerificationCase["reasonCode"], string> = {
  provider_uncertain: "Provider response was uncertain",
  provider_outage: "Provider was unavailable",
  manual_review: "Manual review requested",
};

export function VerificationQueue() {
  const [cases, setCases] = useState<VerificationCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState("");
  const [pendingOutcome, setPendingOutcome] = useState<Outcome | null>(null);
  const [decisionReason, setDecisionReason] = useState("");
  const [evidenceReason, setEvidenceReason] = useState("");
  const [evidence, setEvidence] = useState<Evidence | null>(null);
  const [evidenceOpen, setEvidenceOpen] = useState(false);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [busy, setBusy] = useState(false);
  const commandID = useRef<string | null>(null);
  const selected = cases.find((item) => item.caseId === selectedID);

  async function loadQueue() {
    setLoading(true);
    setMessage("");
    try {
      const response = await fetch("/api/verifications");
      const payload = (await response.json()) as {
        cases?: VerificationCase[];
        message?: string;
      };
      if (!response.ok)
        throw new Error(
          payload.message || "The verification queue could not be loaded.",
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
          : "The verification queue could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    void fetch("/api/verifications")
      .then(async (response) => {
        const payload = (await response.json()) as {
          cases?: VerificationCase[];
          message?: string;
        };
        if (!response.ok) {
          throw new Error(
            payload.message || "The verification queue could not be loaded.",
          );
        }
        if (active) {
          const next = payload.cases ?? [];
          setCases(next);
          setSelectedID(next[0]?.caseId ?? "");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setMessage(
            error instanceof Error
              ? error.message
              : "The verification queue could not be loaded.",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function requestEvidence() {
    if (!selected || evidenceReason.trim().length < 8) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/verifications", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          action: "evidence",
          caseId: selected.caseId,
          purpose: "verification_review",
          reason: evidenceReason,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Evidence & { message?: string }) | null;
      if (!response.ok || !payload?.caseId) {
        const error = new Error(
          payload?.message || "Redacted evidence could not be opened.",
        );
        if (response.status === 403) setStepUpOpen(true);
        throw error;
      }
      setEvidence(payload);
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

  async function startStepUp() {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "start" }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          payload?.message || "The step-up code could not be sent.",
        );
      setSuccess("A fresh step-up code was sent to your admin email.");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The step-up code could not be sent.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function completeStepUp() {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "complete", code: stepUpCode }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          payload?.message || "The step-up code could not be verified.",
        );
      setStepUpOpen(false);
      setStepUpCode("");
      setSuccess("Sensitive evidence access is unlocked for this session.");
      await requestEvidence();
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The step-up code could not be verified.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function decide() {
    if (!selected || !pendingOutcome || decisionReason.trim().length < 8)
      return;
    commandID.current ??= `verification-${crypto.randomUUID()}`;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/verifications", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": commandID.current,
        },
        body: JSON.stringify({
          action: "decision",
          caseId: selected.caseId,
          outcome: pendingOutcome,
          reason: decisionReason,
          expectedVersion: selected.version,
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          payload?.message || "The decision could not be recorded.",
        );
      setSuccess(`${selected.caseId} was recorded as ${pendingOutcome}.`);
      setPendingOutcome(null);
      setDecisionReason("");
      commandID.current = null;
      await loadQueue();
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The decision could not be recorded.",
      );
    } finally {
      setBusy(false);
    }
  }

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
        <Chip
          label={loading ? "Loading queue" : `${cases.length} waiting`}
          color={cases.length ? "warning" : "success"}
        />
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
            <Typography component="h2">Waiting cases</Typography>
            <Typography>Oldest first</Typography>
          </Box>
          <Box aria-label="Manual verification queue">
            {cases.map((item) => (
              <Button
                aria-pressed={item.caseId === selectedID}
                className="verification-case"
                key={item.caseId}
                onClick={() => {
                  setSelectedID(item.caseId);
                  setEvidence(null);
                  setEvidenceOpen(false);
                }}
              >
                <Box>
                  <Typography component="strong">{item.caseId}</Typography>
                  <Typography>{reasonLabels[item.reasonCode]}</Typography>
                  <Typography className="verification-reference">
                    {item.subjectRef}
                  </Typography>
                </Box>
                <span aria-hidden="true">›</span>
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
                    Case {selected.caseId}
                  </Typography>
                  <Typography component="h2">Review bounded proof</Typography>
                </Box>
                <Chip label={selected.status || "queued"} />
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
                    {reasonLabels[selected.reasonCode]}
                  </Typography>
                </Box>
                <Box>
                  <Typography>Submitted</Typography>
                  <Typography component="strong">
                    {new Date(selected.submittedAt).toLocaleString("en-GH")}
                  </Typography>
                </Box>
              </Box>
              <Alert severity="info">
                Full card numbers, raw media and contact details are not shown.
                Opening evidence creates an operator audit event.
              </Alert>
              <Button variant="outlined" onClick={() => setEvidenceOpen(true)}>
                Open redacted evidence
              </Button>
              <Box className="verification-actions">
                <Button
                  variant="contained"
                  color="success"
                  onClick={() => setPendingOutcome("approve")}
                >
                  Propose approval
                </Button>
                <Button
                  variant="outlined"
                  color="error"
                  onClick={() => setPendingOutcome("reject")}
                >
                  Propose rejection
                </Button>
              </Box>
            </>
          ) : (
            <Box className="verification-empty" role="status">
              <Typography component="h2">
                {loading ? "Loading the queue…" : "The queue is clear."}
              </Typography>
              <Typography>
                New uncertain cases appear here without exposing raw identity
                data.
              </Typography>
            </Box>
          )}
        </Card>
      </Box>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setEvidenceOpen(false)}
        open={evidenceOpen}
      >
        <DialogTitle>Redacted evidence</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="warning">
              Access requires recent MFA and is recorded against this operator
              session.
            </Alert>
            {!evidence ? (
              <TextField
                autoFocus
                helperText="At least 8 characters; do not include raw identity data"
                label="Reason for opening evidence"
                multiline
                onChange={(event) =>
                  setEvidenceReason(event.target.value.slice(0, 240))
                }
                rows={3}
                value={evidenceReason}
              />
            ) : (
              <>
                <Box className="evidence-row">
                  <Typography>Ghana Card</Typography>
                  <Typography component="strong">
                    {evidence.maskedCard}
                  </Typography>
                </Box>
                <Box className="evidence-row">
                  <Typography>Age band</Typography>
                  <Typography component="strong">
                    {evidence.ageBand.replaceAll("_", " ")}
                  </Typography>
                </Box>
                <Box className="evidence-row">
                  <Typography>Provider status</Typography>
                  <Typography component="strong">
                    {evidence.providerStatus.replaceAll("_", " ")}
                  </Typography>
                </Box>
                <Box className="evidence-row">
                  <Typography>Raw media</Typography>
                  <Typography component="strong">Not retained</Typography>
                </Box>
              </>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEvidenceOpen(false)}>Close</Button>
          {!evidence ? (
            <Button
              disabled={busy || evidenceReason.trim().length < 8}
              onClick={() => void requestEvidence()}
              variant="contained"
            >
              Open and audit
            </Button>
          ) : null}
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setPendingOutcome(null)}
        open={pendingOutcome !== null}
      >
        <DialogTitle>
          Confirm {pendingOutcome === "approve" ? "approval" : "rejection"}
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            helperText="At least 8 characters, no raw identity data"
            label="Decision reason"
            multiline
            onChange={(event) => {
              setDecisionReason(event.target.value.slice(0, 240));
              commandID.current = null;
            }}
            rows={3}
            sx={{ mt: 1 }}
            value={decisionReason}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPendingOutcome(null)}>Go back</Button>
          <Button
            disabled={busy || decisionReason.trim().length < 8}
            onClick={() => void decide()}
            variant="contained"
          >
            Record audited decision
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="xs"
        onClose={() => setStepUpOpen(false)}
        open={stepUpOpen}
      >
        <DialogTitle>Confirm this sensitive access</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography>
              Request a fresh code, then enter it to unlock redacted evidence
              for this short-lived session.
            </Typography>
            <Button
              disabled={busy}
              onClick={() => void startStepUp()}
              variant="outlined"
            >
              Send step-up code
            </Button>
            <TextField
              label="Six-digit code"
              onChange={(event) =>
                setStepUpCode(event.target.value.replace(/\D/g, "").slice(0, 6))
              }
              slotProps={{ htmlInput: { inputMode: "numeric", maxLength: 6 } }}
              value={stepUpCode}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setStepUpOpen(false)}>Cancel</Button>
          <Button
            disabled={busy || stepUpCode.length !== 6}
            onClick={() => void completeStepUp()}
            variant="contained"
          >
            Verify code
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
