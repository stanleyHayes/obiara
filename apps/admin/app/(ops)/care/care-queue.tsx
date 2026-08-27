"use client";

import { errorCode, needsStepUp } from "../../lib/step-up";

import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Stack,
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
  terminalQueuePath,
} from "../../case-route-model";
import { AdminCard, AdminCardWatermark } from "../../admin-card";

type ScriptKey =
  | "helpline_directory_gh"
  | "counselor_referral"
  | "support_content"
  | "closure_quietening";

interface CareCase {
  caseId: string;
  subjectRef: string;
  signal:
    | "distress_report"
    | "self_harm_indication"
    | "victim_report"
    | "okyeame_escalation"
    | "closure";
  status: "open" | "engaged" | "resolved";
  scripts: ScriptKey[];
  createdAt: string;
  version: number;
}

const resources: ReadonlyArray<{
  key: ScriptKey;
  label: string;
  detail: string;
}> = [
  {
    key: "helpline_directory_gh",
    label: "Ghana helpline directory",
    detail: "Record that the reviewed local directory was shared.",
  },
  {
    key: "counselor_referral",
    label: "Counsellor referral",
    detail: "Record that the approved referral path was offered.",
  },
  {
    key: "support_content",
    label: "Support content",
    detail: "Record that reviewed, non-diagnostic support material was used.",
  },
  {
    key: "closure_quietening",
    label: "Closure quietening",
    detail: "Record the closure-specific quietening resource.",
  },
];

const signalLabels: Record<CareCase["signal"], string> = {
  distress_report: "Distress report",
  self_harm_indication: "Self-harm indication",
  victim_report: "Victim support",
  okyeame_escalation: "Okyeame escalation",
  closure: "Closure support",
};

function ageLabel(value: string) {
  const minutes = Math.max(
    0,
    Math.floor((Date.now() - new Date(value).getTime()) / 60_000),
  );
  if (minutes < 60) return `${minutes}m waiting`;
  return `${Math.floor(minutes / 60)}h waiting`;
}

export function CareQueue({ caseId }: Readonly<{ caseId?: string }>) {
  const router = useRouter();
  const [cases, setCases] = useState<CareCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [scripts, setScripts] = useState<ScriptKey[]>([]);
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
      const response = await fetch("/api/care");
      const payload = (await response.json()) as {
        cases?: CareCase[];
        message?: string;
      };
      if (!response.ok)
        throw new Error(
          payload.message || "The care queue could not be loaded.",
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
          : "The care queue could not be loaded.",
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
    const params = new URLSearchParams(window.location.search);
    const notice = queueNoticeText(params.get("notice"));
    if (notice) {
      window.queueMicrotask(() => setSuccess(notice));
      // Consume the notice once and strip it from the URL.
      params.delete("notice");
      window.history.replaceState(
        null,
        "",
        `${window.location.pathname}${params.size ? `?${params}` : ""}`,
      );
    }
    void fetch("/api/care", { signal: controller.signal })
      .then(async (response) => {
        const payload = (await response.json()) as {
          cases?: CareCase[];
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "The care queue could not be loaded.",
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
              : "The care queue could not be loaded.",
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

  async function mutate(action: "engage" | "resolve") {
    if (!selected) return;
    const actionRequest = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
    try {
      const response = await fetch("/api/care", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, caseId: selected.caseId, scripts }),
      });
      const payload = (await response.json().catch(() => null)) as
        (CareCase & { message?: string }) | null;
      if (!mounted.current || actionRequest !== actionGeneration.current)
        return;
      if (!response.ok || !payload?.caseId) {
        if (
          action === "resolve" &&
          needsStepUp(response.status, errorCode(payload))
        )
          setStepUpOpen(true);
        throw new Error(
          payload?.message ||
            `The care case could not be ${action === "engage" ? "engaged" : "resolved"}.`,
        );
      }
      if (action === "resolve") {
        // On detail route resolving—don't mutate state, let the page reload handle it.
        // This prevents the "case not found" empty state from flashing before navigation.
      } else {
        // On detail route engaging—update the case in place.
        setCases((current) =>
          current.map((item) =>
            item.caseId === payload.caseId ? payload : item,
          ),
        );
      }
      setScripts([]);
      setSuccess(
        action === "engage"
          ? `${payload.caseId} is now engaged by the care desk.`
          : `${payload.caseId} was resolved with approved resource keys.`,
      );
      if (action === "resolve") setSelectedID("");
      if (action === "resolve") {
        router.replace(terminalQueuePath("care", "case-resolved"));
      }
    } catch (error) {
      if (!mounted.current || actionRequest !== actionGeneration.current)
        return;
      const text =
        error instanceof Error ? error.message : "The care action failed.";
      setMessage(text);
      if (stepUpOpen || action === "resolve") setDialogError(text);
    } finally {
      if (mounted.current && actionRequest === actionGeneration.current)
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
        await mutate("resolve");
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

  function toggleScript(key: ScriptKey) {
    setScripts((current) =>
      current.includes(key)
        ? current.filter((item) => item !== key)
        : [...current, key],
    );
  }

  return (
    <main className="verification-shell care-shell" aria-busy={busy}>
      <header className="verification-header">
        <Box>
          <Link href={detailMode ? "/care" : "/"} className="verification-back">
            {detailMode ? "Back to care queue" : "Return to command centre"}
          </Link>
          <Typography className="section-kicker">Care queue</Typography>
          <Typography component="h1">
            Resources first. People always.
          </Typography>
          <Typography>
            Persisted care signals stay separate from enforcement, diagnosis and
            message delivery.
          </Typography>
        </Box>
        {loading ? (
          <AdminSkeleton
            variant="identity"
            label="Loading care queue status"
            className="triage-status-skeleton"
          />
        ) : loadError ? null : (
          <Stack direction="row" spacing={1}>
            <Chip label={`${cases.length} active`} color="warning" />
            <Chip label="Care is non-punitive" color="success" />
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
              <Typography component="h2">Oldest first</Typography>
              <Button
                disabled={loading}
                onClick={() => void loadQueue()}
                size="small"
              >
                Refresh
              </Button>
            </Box>
            <Box aria-label="Care cases">
              {loading ? (
                <AdminSkeleton
                  variant="card-list"
                  rows={4}
                  label="Loading care queue"
                />
              ) : loadError ? null : cases.length ? (
                cases.map((item) => (
                  <Button
                    className="care-case"
                    href={buildCasePath("care", item.caseId, "/care")}
                    key={item.caseId}
                  >
                    <Box className="admin-watermarked-row">
                      <AdminCardWatermark watermark="care" />
                      <Typography component="strong">{item.caseId}</Typography>
                      <Typography>{signalLabels[item.signal]}</Typography>
                      <Typography className="safety-reference">
                        {item.subjectRef} · {ageLabel(item.createdAt)}
                      </Typography>
                    </Box>
                    <Chip
                      label={item.status}
                      color={item.status === "engaged" ? "success" : "warning"}
                      size="small"
                    />
                  </Button>
                ))
              ) : (
                <EmptyState
                  icon="♡"
                  title="Care queue is clear"
                  description="No care follow-ups are waiting. New support signals will appear here oldest first."
                  variant="success"
                />
              )}
            </Box>
          </AdminCard>
        ) : null}
        {detailMode ? (
          <AdminCard
            variant="detail"
            watermark="care"
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
                      {signalLabels[selected.signal]}
                    </Typography>
                  </Box>
                  <Chip
                    label={selected.status}
                    color={
                      selected.status === "engaged" ? "success" : "warning"
                    }
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
                    <Typography>Created</Typography>
                    <Typography component="strong">
                      {new Date(selected.createdAt).toLocaleString()}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography>Revision</Typography>
                    <Typography component="strong">
                      {selected.version}
                    </Typography>
                  </Box>
                </Box>
                {selected.status === "open" ? (
                  <>
                    <Alert severity="info">
                      Engaging acknowledges that trained staff have begun this
                      care case. It does not contact the member or create an
                      enforcement record.
                    </Alert>
                    <Button
                      disabled={busy}
                      onClick={() => void mutate("engage")}
                      variant="contained"
                    >
                      Engage care case
                    </Button>
                  </>
                ) : (
                  <>
                    <Alert severity="info">
                      Record only resources actually used. This closes the care
                      case; it does not claim that any message was delivered.
                    </Alert>
                    <Stack>
                      {resources.map((resource) => (
                        <FormControlLabel
                          key={resource.key}
                          control={
                            <Checkbox
                              checked={scripts.includes(resource.key)}
                              onChange={() => toggleScript(resource.key)}
                            />
                          }
                          label={
                            <Box>
                              <Typography component="strong">
                                {resource.label}
                              </Typography>
                              <Typography>{resource.detail}</Typography>
                            </Box>
                          }
                        />
                      ))}
                    </Stack>
                    <Button
                      disabled={busy || scripts.length === 0}
                      onClick={() => void mutate("resolve")}
                      variant="contained"
                    >
                      Resolve with selected resources
                    </Button>
                  </>
                )}
              </Stack>
            ) : loading ? (
              <AdminSkeleton
                variant="form"
                label="Loading care case workspace"
              />
            ) : loadError ? (
              <EmptyState
                icon="!"
                title="Care case unavailable"
                description={loadError}
                variant="warning"
                action={<Button onClick={() => void loadQueue()}>Retry</Button>}
              />
            ) : (
              <EmptyState
                icon="♡"
                title="Care case not found"
                description="This case is no longer active, or the link is invalid."
                variant="warning"
                action={<Button href="/care">Back to queue</Button>}
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
                  Resolution remains unchanged until this session completes a
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
                Verify and resolve
              </Button>
            </DialogActions>
          </Dialog>
        </>
      ) : null}
    </main>
  );
}
