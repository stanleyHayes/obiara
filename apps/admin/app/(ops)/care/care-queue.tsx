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
import { AdminIcon, UtilityIcon } from "../../admin-icons";
import { adminFetch } from "../../lib/admin-fetch";

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

function CareQueueItem({ item }: Readonly<{ item: CareCase }>) {
  return (
    <Button
      className="care-case"
      href={buildCasePath("care", item.caseId, "/care")}
    >
      <span className="care-case-mark" aria-hidden="true">
        <AdminIcon name="care" />
      </span>
      <Box className="admin-watermarked-row care-case-copy">
        <AdminCardWatermark watermark="care" />
        <Typography className="care-case-id">{item.caseId}</Typography>
        <Typography component="strong" className="care-case-title">
          {signalLabels[item.signal]}
        </Typography>
        <Typography className="care-case-reference">
          {item.subjectRef}
        </Typography>
      </Box>
      <Box className="care-case-wait">
        <UtilityIcon name="clock" aria-hidden="true" />
        <span>{ageLabel(item.createdAt)}</span>
        <small>
          {item.status === "engaged" ? "Care in progress" : "Needs response"}
        </small>
      </Box>
      <span className="care-case-open" aria-hidden="true">
        <UtilityIcon name="arrow-right" />
      </span>
    </Button>
  );
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
      const response = await adminFetch("/api/care");
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
    void adminFetch("/api/care", { signal: controller.signal })
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
      const response = await adminFetch("/api/care", {
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
      const response = await adminFetch("/api/step-up", {
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
    <main
      className={`verification-shell care-shell care-redesign ${detailMode ? "is-detail" : "is-queue"}`}
      aria-busy={busy}
    >
      <header className="verification-header care-hero">
        <Box className="care-hero-copy">
          <Link href={detailMode ? "/care" : "/"} className="verification-back">
            {detailMode ? "Back to care queue" : "Return to command centre"}
          </Link>
          <Box className="care-hero-kicker">
            <AdminIcon name="care" aria-hidden="true" />
            <Typography className="section-kicker">Care response</Typography>
          </Box>
          <Typography component="h1">Hold space. Act with care.</Typography>
          <Typography>
            A calm, non-punitive workspace for responding to people who may need
            support—without diagnosis or enforcement.
          </Typography>
        </Box>
        {loading ? (
          <AdminSkeleton
            variant="identity"
            label="Loading care queue status"
            className="triage-status-skeleton"
          />
        ) : loadError ? null : (
          <Box className="care-hero-status">
            <Box>
              <strong>{cases.length}</strong>
              <span>people waiting</span>
            </Box>
            <Box>
              <strong>
                {cases.filter((item) => item.status === "engaged").length}
              </strong>
              <span>responses active</span>
            </Box>
            <Box className="care-principle">
              <AdminIcon name="verification" aria-hidden="true" />
              <span>non-punitive</span>
            </Box>
          </Box>
        )}
        <AdminCardWatermark watermark="care" />
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
            className="verification-list care-queue-panel"
          >
            <Box className="verification-panel-heading">
              <Box>
                <Typography className="section-kicker">
                  Response order
                </Typography>
                <Typography component="h2">People waiting</Typography>
                <Typography>
                  Oldest signals come first. Every response stays human-led.
                </Typography>
              </Box>
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
                  <CareQueueItem item={item} key={item.caseId} />
                ))
              ) : (
                <EmptyState
                  icon={<AdminIcon name="care" aria-hidden="true" />}
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
            className="verification-review care-review-panel"
          >
            {selected ? (
              <Stack spacing={3} className="care-review-stack">
                <Box className="verification-panel-heading care-review-heading">
                  <Box>
                    <Box className="care-detail-overline">
                      <AdminIcon name="care" aria-hidden="true" />
                      <Typography>
                        {selected.status === "engaged"
                          ? "Response in progress"
                          : "Awaiting first response"}
                      </Typography>
                    </Box>
                    <Typography component="h2">
                      {signalLabels[selected.signal]}
                    </Typography>
                    <Typography className="care-detail-id">
                      {selected.caseId}
                    </Typography>
                  </Box>
                  <Chip
                    label={selected.status}
                    color={
                      selected.status === "engaged" ? "success" : "warning"
                    }
                  />
                </Box>
                <Box className="verification-facts care-facts">
                  <Box>
                    <Typography>Private subject</Typography>
                    <Typography component="strong">
                      {selected.subjectRef}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography>Waiting since</Typography>
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
                <Box className="care-action-stage">
                  {selected.status === "open" ? (
                    <>
                      <Alert severity="info">
                        Engaging acknowledges that trained staff have begun this
                        care case. It does not contact the member or create an
                        enforcement record.
                      </Alert>
                      <Button
                        className="care-primary-action"
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
                        Record only resources actually used. This closes the
                        care case; it does not claim that any message was
                        delivered.
                      </Alert>
                      <Box className="care-resource-intro">
                        <Typography className="section-kicker">
                          Resource record
                        </Typography>
                        <Typography component="h3">
                          What support was used?
                        </Typography>
                      </Box>
                      <Stack className="care-resource-list">
                        {resources.map((resource) => (
                          <FormControlLabel
                            className="care-resource-option"
                            key={resource.key}
                            control={
                              <Checkbox
                                checked={scripts.includes(resource.key)}
                                onChange={() => toggleScript(resource.key)}
                              />
                            }
                            label={
                              <Box>
                                <span
                                  className="care-resource-icon"
                                  aria-hidden="true"
                                >
                                  <AdminIcon name="care" />
                                </span>
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
                        className="care-primary-action"
                        disabled={busy || scripts.length === 0}
                        onClick={() => void mutate("resolve")}
                        variant="contained"
                      >
                        Resolve with selected resources
                      </Button>
                    </>
                  )}
                </Box>
              </Stack>
            ) : loading ? (
              <AdminSkeleton
                variant="form"
                label="Loading care case workspace"
              />
            ) : loadError ? (
              <EmptyState
                icon={<AdminIcon name="incidents" aria-hidden="true" />}
                title="Care case unavailable"
                description={loadError}
                variant="warning"
                action={<Button onClick={() => void loadQueue()}>Retry</Button>}
              />
            ) : (
              <EmptyState
                icon={<AdminIcon name="care" aria-hidden="true" />}
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
