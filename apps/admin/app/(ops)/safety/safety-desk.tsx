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
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { SegmentedOtpInput } from "@obiara/ui-web";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";
import { buildCasePath } from "../../case-route-model";
import { AdminCard, AdminCardWatermark } from "../../admin-card";

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

function SafetyQueueItem({ item }: Readonly<{ item: SafetyCase }>) {
  return (
    <Button
      className="safety-case"
      href={buildCasePath("safety", item.caseId, "/safety")}
    >
      <Box className="admin-watermarked-row">
        <AdminCardWatermark watermark="safety" />
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

export function SafetyDesk({ caseId }: Readonly<{ caseId?: string }>) {
  const [cases, setCases] = useState<SafetyCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [purpose, setPurpose] = useState<EvidencePurpose>("triage");
  const [evidence, setEvidence] = useState<SafetyEvidence | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [loadError, setLoadError] = useState("");
  const [success, setSuccess] = useState("");
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [dialogError, setDialogError] = useState("");
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

  async function loadQueue() {
    const generation = ++loadGeneration.current;
    setLoading(true);
    window.queueMicrotask(() => setLoadError(""));
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
      if (!mounted.current || generation !== loadGeneration.current) return;
      setCases(next);
      setSelectedID(
        caseId && next.some((item) => item.caseId === caseId) ? caseId : "",
      );
    } catch (error) {
      if (!mounted.current || generation !== loadGeneration.current) return;
      setLoadError(
        error instanceof Error
          ? error.message
          : "The safety queue could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    const generation = ++loadGeneration.current;
    const controller = new AbortController();
    window.queueMicrotask(() => setLoadError(""));
    void fetch("/api/safety", { signal: controller.signal })
      .then(async (response) => {
        const payload = (await response.json()) as {
          cases?: SafetyCase[];
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "The safety queue could not be loaded.",
          );
        if (active && generation === loadGeneration.current) {
          const next = payload.cases ?? [];
          setCases(next);
          setSelectedID(
            caseId && next.some((item) => item.caseId === caseId) ? caseId : "",
          );
        }
      })
      .catch((error: unknown) => {
        if (active && generation === loadGeneration.current)
          setLoadError(
            error instanceof Error
              ? error.message
              : "The safety queue could not be loaded.",
          );
      })
      .finally(() => {
        if (active && generation === loadGeneration.current) setLoading(false);
      });
    return () => {
      active = false;
      controller.abort();
    };
  }, [caseId]);

  async function assignCase() {
    if (!selected) return;
    const action = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
    try {
      const response = await fetch("/api/safety", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "assign", caseId: selected.caseId }),
      });
      const payload = (await response.json().catch(() => null)) as
        (SafetyCase & { message?: string }) | null;
      if (!mounted.current || action !== actionGeneration.current) return;
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
      if (!mounted.current || action !== actionGeneration.current) return;
      setMessage(
        error instanceof Error
          ? error.message
          : "The case could not be assigned.",
      );
    } finally {
      if (mounted.current && action === actionGeneration.current)
        setBusy(false);
    }
  }

  async function requestEvidence() {
    if (!selected?.assignedToMe) return;
    const action = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
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
      if (!mounted.current || action !== actionGeneration.current) return;
      if (!response.ok || !payload?.caseId) {
        if (needsStepUp(response.status, errorCode(payload)))
          setStepUpOpen(true);
        throw new Error(
          payload?.message || "Redacted evidence could not be opened.",
        );
      }
      setEvidence(payload);
      setSuccess(`Audited ${purpose} access recorded for ${payload.caseId}.`);
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

  async function stepUp(action: "start" | "complete") {
    const actionRequest = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
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
      if (!mounted.current || actionRequest !== actionGeneration.current)
        return;
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
      if (!mounted.current || actionRequest !== actionGeneration.current)
        return;
      const text =
        error instanceof Error
          ? error.message
          : "The MFA step-up could not be completed.";
      setMessage(text);
      setDialogError(text);
    } finally {
      if (mounted.current && actionRequest === actionGeneration.current)
        setBusy(false);
    }
  }

  return (
    <main className="verification-shell safety-desk-shell" aria-busy={busy}>
      <header className="verification-header">
        <Box>
          <Link
            href={detailMode ? "/safety" : "/"}
            className="verification-back"
          >
            {detailMode ? "Back to safety queue" : "Return to command centre"}
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
        {loading ? (
          <AdminSkeleton
            variant="form"
            label="Loading trust and safety desk header"
            className="triage-header-skeleton"
          />
        ) : loadError ? null : (
          <Stack direction="row" spacing={1}>
            <Chip label={`${cases.length} queued`} color="warning" />
            <Chip label="Evidence access audited" color="success" />
          </Stack>
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
            showWatermark={!loading && !loadError && cases.length > 0}
            className="verification-list"
          >
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
                <AdminSkeleton
                  variant="card-list"
                  rows={4}
                  label="Loading trust and safety queue"
                />
              ) : loadError ? null : cases.length ? (
                cases.map((item) => (
                  <SafetyQueueItem item={item} key={item.caseId} />
                ))
              ) : (
                <EmptyState
                  icon="✓"
                  title="Triage queue is clear"
                  description="No trust and safety cases are waiting. New priority work will appear here in SLA order."
                  variant="success"
                />
              )}
            </Box>
          </AdminCard>
        ) : null}
        {detailMode ? (
          <AdminCard
            variant="detail"
            watermark="safety"
            showWatermark={!loading && !loadError && Boolean(selected)}
            className="verification-review"
          >
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
            ) : loading ? (
              <AdminSkeleton
                variant="form"
                label="Loading safety case workspace"
              />
            ) : loadError ? (
              <EmptyState
                icon="!"
                title="Safety case unavailable"
                description={loadError}
                variant="warning"
                action={<Button onClick={() => void loadQueue()}>Retry</Button>}
              />
            ) : (
              <EmptyState
                icon="⌁"
                title="Safety case not found"
                description="This case is no longer in the active queue, or the link is invalid."
                variant="warning"
                action={<Button href="/safety">Back to queue</Button>}
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
              setEvidence(null);
              setDialogError("");
            }}
            open={Boolean(evidence)}
          >
            <DialogTitle>Redacted case evidence</DialogTitle>
            <DialogContent>
              {evidence ? (
                <Stack spacing={2}>
                  <Alert severity="warning">
                    Purpose: {purpose}. Access has been logged.
                  </Alert>
                  {dialogError ? (
                    <Alert severity="error" role="alert" aria-live="assertive">
                      {dialogError}
                    </Alert>
                  ) : null}
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
                      <Typography component="strong">
                        {evidence.surface}
                      </Typography>
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
            className="admin-form-dialog"
            fullWidth
            maxWidth="xs"
            onClose={() => {
              setStepUpOpen(false);
              setDialogError("");
            }}
            open={stepUpOpen}
          >
            <DialogTitle>Fresh MFA required</DialogTitle>
            <DialogContent>
              <Stack spacing={2} sx={{ pt: 1 }}>
                <Alert severity="info">
                  The evidence remains sealed until this session completes a
                  fresh step-up.
                </Alert>
                {dialogError ? (
                  <Alert severity="error" role="alert" aria-live="assertive">
                    {dialogError}
                  </Alert>
                ) : null}
                <SegmentedOtpInput
                  label="Step-up code"
                  onChange={setStepUpCode}
                  value={stepUpCode}
                  disabled={busy}
                  required
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
        </>
      ) : null}
    </main>
  );
}
