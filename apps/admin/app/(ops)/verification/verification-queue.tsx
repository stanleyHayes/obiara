"use client";

import { errorCode, needsStepUp } from "../../lib/step-up";

import {
  Alert,
  Box,
  Button,
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
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { SegmentedOtpInput } from "@obiara/ui-web";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";
import {
  buildCasePath,
  queueNoticeText,
  sanitizeQueueReturn,
  terminalQueuePath,
} from "../../case-route-model";
import { AdminCard, AdminCardWatermark } from "../../admin-card";

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

export function VerificationQueue({ caseId }: Readonly<{ caseId?: string }>) {
  const router = useRouter();
  const [cases, setCases] = useState<VerificationCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [loadError, setLoadError] = useState("");
  const [success, setSuccess] = useState("");
  const [pendingOutcome, setPendingOutcome] = useState<Outcome | null>(null);
  const [decisionReason, setDecisionReason] = useState("");
  const [evidenceReason, setEvidenceReason] = useState("");
  const [evidence, setEvidence] = useState<Evidence | null>(null);
  const [evidenceOpen, setEvidenceOpen] = useState(false);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [query, setQuery] = useState("");
  const [dialogError, setDialogError] = useState("");
  const [returnHref, setReturnHref] = useState("/verification");
  const searchInput = useRef<HTMLInputElement>(null);
  const commandID = useRef<string | null>(null);
  const loadGeneration = useRef(0);
  const mounted = useRef(true);
  const actionGeneration = useRef(0);
  const selected = cases.find((item) => item.caseId === selectedID);
  const detailMode = Boolean(caseId);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      actionGeneration.current += 1;
      loadGeneration.current += 1;
    };
  }, []);

  useEffect(() => {
    let active = true;
    const generation = ++loadGeneration.current;
    const controller = new AbortController();
    const params = new URLSearchParams(window.location.search);
    window.queueMicrotask(() => {
      setLoadError("");
      if (detailMode) {
        const requestedReturn = params.get("return");
        setReturnHref(sanitizeQueueReturn("verification", requestedReturn));
      } else setQuery(params.get("q") ?? "");
      const notice = queueNoticeText(params.get("notice"));
      if (notice) {
        setSuccess(notice);
        // Consume the notice once and strip it from the URL.
        params.delete("notice");
        window.history.replaceState(
          null,
          "",
          `${window.location.pathname}${params.size ? `?${params}` : ""}`,
        );
      }
    });
    void fetch("/api/verifications", { signal: controller.signal })
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
        if (active && generation === loadGeneration.current) {
          const next = payload.cases ?? [];
          setCases(next);
          setSelectedID(
            caseId && next.some((item) => item.caseId === caseId) ? caseId : "",
          );
          if (
            !detailMode &&
            new URLSearchParams(window.location.search).get("search") === "1"
          ) {
            window.requestAnimationFrame(() => searchInput.current?.focus());
          }
        }
      })
      .catch((error: unknown) => {
        if (active && generation === loadGeneration.current) {
          setLoadError(
            error instanceof Error
              ? error.message
              : "The verification queue could not be loaded.",
          );
        }
      })
      .finally(() => {
        if (active && generation === loadGeneration.current) setLoading(false);
      });
    return () => {
      active = false;
      controller.abort();
    };
  }, [caseId, detailMode]);

  const normalizedQuery = query.trim().toLowerCase();
  const filteredCases = normalizedQuery
    ? cases.filter((item) =>
        [item.caseId, item.subjectRef, reasonLabels[item.reasonCode]].some(
          (value) => value.toLowerCase().includes(normalizedQuery),
        ),
      )
    : cases;

  function updateQuery(value: string) {
    setQuery(value);
    const params = new URLSearchParams(window.location.search);
    if (value.trim()) params.set("q", value);
    else params.delete("q");
    params.delete("search");
    params.delete("notice");
    window.history.replaceState(
      null,
      "",
      `${window.location.pathname}${params.size ? `?${params}` : ""}`,
    );
  }

  async function requestEvidence() {
    if (!selected || evidenceReason.trim().length < 8) return;
    const action = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
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
      if (!mounted.current || action !== actionGeneration.current) return;
      if (!response.ok || !payload?.caseId) {
        const error = new Error(
          payload?.message || "Redacted evidence could not be opened.",
        );
        if (needsStepUp(response.status, errorCode(payload)))
          setStepUpOpen(true);
        throw error;
      }
      setEvidence(payload);
    } catch (error) {
      if (!mounted.current || action !== actionGeneration.current) return;
      const text =
        error instanceof Error
          ? error.message
          : "Redacted evidence could not be opened.";
      setMessage(text);
      setDialogError(text);
    } finally {
      if (mounted.current && action === actionGeneration.current)
        setBusy(false);
    }
  }

  async function startStepUp() {
    const action = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
    try {
      const response = await fetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "start" }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!mounted.current || action !== actionGeneration.current) return;
      if (!response.ok)
        throw new Error(
          payload?.message || "The step-up code could not be sent.",
        );
      setSuccess("A fresh step-up code was sent to your admin email.");
    } catch (error) {
      if (!mounted.current || action !== actionGeneration.current) return;
      const text =
        error instanceof Error
          ? error.message
          : "The step-up code could not be sent.";
      setMessage(text);
      setDialogError(text);
    } finally {
      if (mounted.current && action === actionGeneration.current)
        setBusy(false);
    }
  }

  async function completeStepUp() {
    const action = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
    try {
      const response = await fetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "complete", code: stepUpCode }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!mounted.current || action !== actionGeneration.current) return;
      if (!response.ok)
        throw new Error(
          payload?.message || "The step-up code could not be verified.",
        );
      setStepUpOpen(false);
      setStepUpCode("");
      setSuccess("Sensitive evidence access is unlocked for this session.");
      await requestEvidence();
    } catch (error) {
      if (!mounted.current || action !== actionGeneration.current) return;
      const text =
        error instanceof Error
          ? error.message
          : "The step-up code could not be verified.";
      setMessage(text);
      setDialogError(text);
    } finally {
      if (mounted.current && action === actionGeneration.current)
        setBusy(false);
    }
  }

  async function decide() {
    if (!selected || !pendingOutcome || decisionReason.trim().length < 8)
      return;
    const action = ++actionGeneration.current;
    commandID.current ??= `verification-${crypto.randomUUID()}`;
    setBusy(true);
    setMessage("");
    setDialogError("");
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
      if (!mounted.current || action !== actionGeneration.current) return;
      if (!response.ok)
        throw new Error(
          payload?.message || "The decision could not be recorded.",
        );
      setSuccess(`${selected.caseId} was recorded as ${pendingOutcome}.`);
      setPendingOutcome(null);
      setDecisionReason("");
      commandID.current = null;
      router.replace(
        terminalQueuePath("verification", "decision-recorded", returnHref),
      );
    } catch (error) {
      if (!mounted.current || action !== actionGeneration.current) return;
      const text =
        error instanceof Error
          ? error.message
          : "The decision could not be recorded.";
      setMessage(text);
      setDialogError(text);
    } finally {
      if (mounted.current && action === actionGeneration.current)
        setBusy(false);
    }
  }

  return (
    <main className="verification-shell" aria-busy={busy}>
      <header className="verification-header">
        <Box>
          <Link
            href={detailMode ? returnHref : "/"}
            className="verification-back"
          >
            {detailMode
              ? "Back to verification queue"
              : "Return to command centre"}
          </Link>
          <Typography className="section-kicker">Verification desk</Typography>
          <Typography component="h1">
            Human review, with less exposed.
          </Typography>
          <Typography>
            Provider uncertainty comes here. Approval never happens silently.
          </Typography>
        </Box>
        {loading ? (
          <AdminSkeleton
            variant="identity"
            label="Loading verification queue status"
            className="triage-status-skeleton"
          />
        ) : loadError ? null : (
          <Chip
            label={`${cases.length} waiting`}
            color={cases.length ? "warning" : "success"}
          />
        )}
      </header>

      {loadError ? (
        <Alert severity="error" className="verification-alert">
          {loadError}
        </Alert>
      ) : null}
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
        {!detailMode ? (
          <AdminCard
            variant="panel"
            watermark="queue"
            showWatermark={!loading && !loadError && filteredCases.length > 0}
            className="verification-list"
          >
            <Box className="verification-panel-heading">
              <Typography component="h2">Waiting cases</Typography>
              <Typography>Oldest first</Typography>
            </Box>
            {!loading ? (
              <TextField
                className="verification-search"
                fullWidth
                inputRef={searchInput}
                label="Search verification cases"
                onChange={(event) => updateQuery(event.target.value)}
                placeholder="Case ID, private reference or reason"
                size="small"
                value={query}
              />
            ) : null}
            <Box aria-label="Manual verification queue">
              {loading ? (
                <AdminSkeleton
                  variant="card-list"
                  rows={4}
                  label="Loading manual verification queue"
                />
              ) : null}
              {!loading && !loadError && cases.length === 0 ? (
                <EmptyState
                  icon="✓"
                  title="Verification queue is clear"
                  description="No uncertain verification cases are waiting. New cases will appear here oldest first."
                  variant="success"
                />
              ) : null}
              {!loading &&
              !loadError &&
              cases.length > 0 &&
              filteredCases.length === 0 ? (
                <EmptyState
                  icon="⌕"
                  title="No matching cases"
                  description="Try another case ID, private reference or review reason."
                  variant="search"
                  action={
                    <Button onClick={() => updateQuery("")}>
                      Clear search
                    </Button>
                  }
                />
              ) : null}
              {!loadError
                ? filteredCases.map((item) => (
                    <Button
                      className="verification-case"
                      href={buildCasePath(
                        "verification",
                        item.caseId,
                        query
                          ? `/verification?${new URLSearchParams({ q: query })}`
                          : "/verification",
                      )}
                      key={item.caseId}
                    >
                      <Box className="admin-watermarked-row">
                        <AdminCardWatermark watermark="verification" />
                        <Typography component="strong">
                          {item.caseId}
                        </Typography>
                        <Typography>{reasonLabels[item.reasonCode]}</Typography>
                        <Typography className="verification-reference">
                          {item.subjectRef}
                        </Typography>
                      </Box>
                      <span aria-hidden="true">›</span>
                    </Button>
                  ))
                : null}
            </Box>
          </AdminCard>
        ) : null}
        {detailMode ? (
          <AdminCard
            variant="detail"
            watermark="evidence"
            showWatermark={!loading && !loadError && Boolean(selected)}
            className="verification-review"
          >
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
                  Full card numbers, raw media and contact details are not
                  shown. Opening evidence creates an operator audit event.
                </Alert>
                <Button
                  variant="outlined"
                  onClick={() => setEvidenceOpen(true)}
                >
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
            ) : loading ? (
              <AdminSkeleton
                variant="form"
                label="Loading verification review workspace"
              />
            ) : loadError ? (
              <EmptyState
                icon="!"
                title="Verification case unavailable"
                description={loadError}
                variant="warning"
                action={
                  <Button onClick={() => window.location.reload()}>
                    Retry
                  </Button>
                }
              />
            ) : (
              <EmptyState
                icon="⌁"
                title="Verification case not found"
                description="This case is no longer in the active manual-review queue, or the link is invalid."
                variant="warning"
                action={<Button href="/verification">Back to queue</Button>}
              />
            )}
          </AdminCard>
        ) : null}
      </Box>

      {detailMode ? (
        <>
          <Dialog
            className="admin-form-dialog"
            fullWidth
            maxWidth="sm"
            onClose={() => {
              setEvidenceOpen(false);
              setDialogError("");
            }}
            open={evidenceOpen}
          >
            <DialogTitle>Redacted evidence</DialogTitle>
            <DialogContent>
              <Stack spacing={2} sx={{ pt: 1 }}>
                <Alert severity="warning">
                  Access requires recent MFA and is recorded against this
                  operator session.
                </Alert>
                {dialogError ? (
                  <Alert severity="error" role="alert" aria-live="assertive">
                    {dialogError}
                  </Alert>
                ) : null}
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
            className="admin-form-dialog"
            fullWidth
            maxWidth="sm"
            onClose={() => {
              setPendingOutcome(null);
              setDialogError("");
            }}
            open={pendingOutcome !== null}
          >
            <DialogTitle>
              Confirm {pendingOutcome === "approve" ? "approval" : "rejection"}
            </DialogTitle>
            <DialogContent>
              {dialogError ? (
                <Alert
                  severity="error"
                  role="alert"
                  aria-live="assertive"
                  sx={{ mb: 2 }}
                >
                  {dialogError}
                </Alert>
              ) : null}
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
              <Button
                onClick={() => {
                  setPendingOutcome(null);
                  setDialogError("");
                }}
              >
                Go back
              </Button>
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
            className="admin-form-dialog"
            fullWidth
            maxWidth="xs"
            onClose={() => {
              setStepUpOpen(false);
              setDialogError("");
            }}
            open={stepUpOpen}
          >
            <DialogTitle>Confirm this sensitive access</DialogTitle>
            <DialogContent>
              <Stack spacing={2} sx={{ pt: 1 }}>
                <Typography>
                  Request a fresh code, then enter it to unlock redacted
                  evidence for this short-lived session.
                </Typography>
                {dialogError ? (
                  <Alert severity="error" role="alert" aria-live="assertive">
                    {dialogError}
                  </Alert>
                ) : null}
                <Button
                  disabled={busy}
                  onClick={() => void startStepUp()}
                  variant="outlined"
                >
                  Send step-up code
                </Button>
                <SegmentedOtpInput
                  label="Six-digit code"
                  onChange={setStepUpCode}
                  value={stepUpCode}
                  disabled={busy}
                  required
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
        </>
      ) : null}
    </main>
  );
}
