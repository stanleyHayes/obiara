"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Skeleton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useEffect, useState } from "react";

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

export function CareQueue() {
  const [cases, setCases] = useState<CareCase[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [scripts, setScripts] = useState<ScriptKey[]>([]);
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
          : "The care queue could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    void fetch("/api/care")
      .then(async (response) => {
        const payload = (await response.json()) as {
          cases?: CareCase[];
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "The care queue could not be loaded.",
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
              : "The care queue could not be loaded.",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function mutate(action: "engage" | "resolve") {
    if (!selected) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/care", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, caseId: selected.caseId, scripts }),
      });
      const payload = (await response.json().catch(() => null)) as
        (CareCase & { message?: string }) | null;
      if (!response.ok || !payload?.caseId) {
        if (action === "resolve" && response.status === 403)
          setStepUpOpen(true);
        throw new Error(
          payload?.message ||
            `The care case could not be ${action === "engage" ? "engaged" : "resolved"}.`,
        );
      }
      setCases((current) =>
        action === "resolve"
          ? current.filter((item) => item.caseId !== payload.caseId)
          : current.map((item) =>
              item.caseId === payload.caseId ? payload : item,
            ),
      );
      setScripts([]);
      setSuccess(
        action === "engage"
          ? `${payload.caseId} is now engaged by the care desk.`
          : `${payload.caseId} was resolved with approved resource keys.`,
      );
      if (action === "resolve") setSelectedID("");
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : "The care action failed.",
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
        await mutate("resolve");
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

  function toggleScript(key: ScriptKey) {
    setScripts((current) =>
      current.includes(key)
        ? current.filter((item) => item !== key)
        : [...current, key],
    );
  }

  return (
    <main className="verification-shell care-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
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
        <Stack direction="row" spacing={1}>
          <Chip label={`${cases.length} active`} color="warning" />
          <Chip label="Care is non-punitive" color="success" />
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
              <Stack spacing={1.5} sx={{ p: 2 }}>
                <Skeleton height={76} />
                <Skeleton height={76} />
              </Stack>
            ) : cases.length ? (
              cases.map((item) => (
                <Button
                  aria-pressed={item.caseId === selectedID}
                  className="care-case"
                  key={item.caseId}
                  onClick={() => {
                    setSelectedID(item.caseId);
                    setScripts([]);
                  }}
                >
                  <Box>
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
              <Alert severity="success">No care cases are waiting.</Alert>
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
                    {signalLabels[selected.signal]}
                  </Typography>
                </Box>
                <Chip
                  label={selected.status}
                  color={selected.status === "engaged" ? "success" : "warning"}
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
                  <Typography component="strong">{selected.version}</Typography>
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
          ) : (
            <Alert severity="info">Select an active care case to begin.</Alert>
          )}
        </Card>
      </Box>

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
              Resolution remains unchanged until this session completes a fresh
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
            Verify and resolve
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
