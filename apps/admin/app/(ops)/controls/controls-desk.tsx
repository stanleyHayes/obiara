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
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon } from "../../admin-icons";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";
import { adminFetch } from "../../lib/admin-fetch";

type Capability = "sow" | "fires" | "ai" | "payments" | "gate";
type Environment = "staging" | "production";
type ControlAction = "enable" | "disable" | "kill" | "unkill";
type Reason = "staged_rollout" | "incident" | "maintenance";

interface Proposal {
  proposalId: string;
  capability: Capability;
  environment: Environment;
  market: "GH";
  action: ControlAction;
  reason: Reason;
  status: "proposed" | "approved" | "applied" | "expired";
  version: number;
  expiresAt: string;
  proposedByMe: boolean;
  approvedByMe: boolean;
}

const labels: Record<Capability, string> = {
  sow: "Seed sowing",
  fires: "Community Fires",
  ai: "AI assistance",
  payments: "Payments and escrow",
  gate: "Doorway and private gate",
};
const actionVerbs: Record<ControlAction, string> = {
  enable: "enabled",
  disable: "disabled",
  kill: "killed",
  unkill: "removed from runtime kill",
};

export function ControlsDesk() {
  const [proposals, setProposals] = useState<Proposal[]>([]);
  const [capability, setCapability] = useState<Capability>("ai");
  const [environment, setEnvironment] = useState<Environment>("staging");
  const [controlAction, setControlAction] = useState<ControlAction>("kill");
  const [reason, setReason] = useState<Reason>("incident");
  const [loading, setLoading] = useState(true);
  const [loaded, setLoaded] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState("");
  const [pending, setPending] = useState<null | {
    action: "propose" | "approve" | "apply";
    proposalId?: string;
  }>(null);
  const [stepUpCode, setStepUpCode] = useState("");
  const mounted = useRef(false);
  const loadGeneration = useRef(0);
  const actionGeneration = useRef(0);
  const stepUpGeneration = useRef(0);
  const controllerRef = useRef<AbortController | null>(null);
  const commandIdRef = useRef(`control:${crypto.randomUUID()}`);
  const [proposalOpen, setProposalOpen] = useState(false);
  const [dialogError, setDialogError] = useState("");
  const [confirmProposal, setConfirmProposal] = useState<Proposal | null>(null);
  const [confirmAction, setConfirmAction] = useState<
    "approve" | "apply" | null
  >(null);

  async function load() {
    const generation = ++loadGeneration.current;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setLoading(true);
    setLoaded(false);
    setLoadError("");
    try {
      const response = await adminFetch("/api/controls", {
        signal: controller.signal,
      });
      const payload = (await response.json()) as {
        proposals?: Proposal[];
        message?: string;
      };
      if (!response.ok || !Array.isArray(payload.proposals))
        throw new Error(
          payload.message || "Runtime-control proposals could not be loaded.",
        );
      if (mounted.current && generation === loadGeneration.current) {
        setProposals(payload.proposals);
        setLoaded(true);
      }
    } catch (error) {
      if (
        controller.signal.aborted ||
        !mounted.current ||
        generation !== loadGeneration.current
      )
        return;
      setProposalOpen(false);
      setConfirmProposal(null);
      setConfirmAction(null);
      setLoadError(
        error instanceof Error
          ? error.message
          : "Runtime-control proposals could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }

  useEffect(() => {
    mounted.current = true;
    const timer = window.setTimeout(() => void load(), 0);
    return () => {
      window.clearTimeout(timer);
      mounted.current = false;
      loadGeneration.current += 1;
      actionGeneration.current += 1;
      stepUpGeneration.current += 1;
      controllerRef.current?.abort();
    };
  }, []);

  useEffect(() => {
    commandIdRef.current = `control:${crypto.randomUUID()}`;
  }, [capability, environment, controlAction, reason]);

  async function mutate(
    action: "propose" | "approve" | "apply",
    proposalId?: string,
  ) {
    const generation = ++actionGeneration.current;
    setBusy(true);
    setMessage("");
    setDialogError("");
    try {
      const response = await adminFetch("/api/controls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "propose"
            ? {
                action,
                commandId: commandIdRef.current,
                capability,
                environment,
                controlAction,
                reason,
              }
            : { action, proposalId },
        ),
      });
      const payload = (await response.json().catch(() => null)) as
        (Proposal & { message?: string }) | null;
      if (!response.ok || !payload?.proposalId) {
        if (
          needsStepUp(response.status, errorCode(payload)) &&
          mounted.current &&
          generation === actionGeneration.current
        ) {
          setPending({ action, proposalId });
          setDialogError(
            payload?.message ||
              "Fresh MFA is required before this retained action can continue.",
          );
        }
        throw new Error(
          payload?.message || `The proposal could not be ${action}d.`,
        );
      }
      if (!mounted.current || generation !== actionGeneration.current) return;
      setProposals((current) => {
        const retained = current.filter(
          (item) => item.proposalId !== payload.proposalId,
        );
        return payload.status === "expired" ? retained : [payload, ...retained];
      });
      setSuccess(
        action === "propose"
          ? "Proposal retained. A distinct stepped-up administrator must approve it."
          : action === "approve"
            ? "Proposal approved. This approver may now apply the exact retained terms."
            : "Runtime change applied. It will fail closed automatically at expiry.",
      );
      setPending(null);
      setStepUpCode("");
      setDialogError("");
      setConfirmProposal(null);
      setConfirmAction(null);
      if (action === "propose") {
        commandIdRef.current = `control:${crypto.randomUUID()}`;
        setProposalOpen(false);
      }
    } catch (error) {
      if (!mounted.current || generation !== actionGeneration.current) return;
      const nextMessage =
        error instanceof Error
          ? error.message
          : "The runtime-control action failed.";
      if (pending || proposalOpen || confirmProposal)
        setDialogError(nextMessage);
      setMessage(nextMessage);
    } finally {
      if (mounted.current && generation === actionGeneration.current)
        setBusy(false);
    }
  }

  async function stepUp(action: "start" | "complete") {
    const generation = ++stepUpGeneration.current;
    let retrying = false;
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
      if (!mounted.current || generation !== stepUpGeneration.current) return;
      if (!response.ok)
        throw new Error(
          payload?.message || "The MFA step-up could not be completed.",
        );
      if (action === "start") {
        setSuccess("A fresh step-up code was sent to your admin email.");
      } else if (pending) {
        const next = pending;
        setPending(null);
        setStepUpCode("");
        setDialogError("");
        retrying = true;
        await mutate(next.action, next.proposalId);
      }
    } catch (error) {
      if (!mounted.current || generation !== stepUpGeneration.current) return;
      const nextMessage =
        error instanceof Error
          ? error.message
          : "The MFA step-up could not be completed.";
      setDialogError(nextMessage);
      setMessage(nextMessage);
    } finally {
      if (
        !retrying &&
        mounted.current &&
        generation === stepUpGeneration.current
      )
        setBusy(false);
    }
  }

  return (
    <Box className="controls-redesign">
      <Box component="header" className="controls-hero">
        <AdminCardWatermark watermark="analytics" />
        <Box className="controls-hero-copy">
          <Button className="controls-back" component={Link} href="/">
            ← Command centre
          </Button>
          <Box className="controls-kicker">
            <AdminIcon name="controls" aria-hidden="true" />
            <Typography className="section-kicker">
              RUNTIME AUTHORITY · LIVE
            </Typography>
          </Box>
          <Typography component="h1">
            Change less. Know when it ends.
          </Typography>
          <Typography className="controls-hero-intro">
            Durable two-person controls for one capability, one environment, and
            one market—with automatic fail-closed expiry.
          </Typography>
        </Box>
        <Box
          className="controls-hero-register"
          aria-label="Runtime control guarantees"
        >
          <div>
            <span>Approval</span>
            <strong>Two operators</strong>
            <Typography>Proposal and approval remain distinct</Typography>
          </div>
          <div>
            <span>Maximum lifetime</span>
            <strong>02 hours</strong>
            <Typography>Every command expires automatically</Typography>
          </div>
          <div>
            <span>Expiry posture</span>
            <strong>Disabled + killed</strong>
            <Typography>The platform fails closed</Typography>
          </div>
        </Box>
      </Box>

      <Box component="section" className="controls-boundary">
        <Box className="controls-boundary-icon">
          <AdminIcon name="controls" aria-hidden="true" />
        </Box>
        <Box>
          <Typography className="section-kicker">CHANGE BOUNDARY</Typography>
          <Typography component="h2">
            Exact terms, retained before action.
          </Typography>
          <Typography>
            No broad switches. Every command fixes the capability, environment,
            market, reason, and expiry before another administrator reviews it.
          </Typography>
        </Box>
        <span className="controls-boundary-state">FAIL-CLOSED</span>
      </Box>

      {message ? <Alert severity="error">{message}</Alert> : null}
      {success ? (
        <Alert severity="success" onClose={() => setSuccess("")}>
          {success}
        </Alert>
      ) : null}

      <AdminCard
        variant="form"
        watermark="evidence"
        className="controls-proposal-launcher"
      >
        <Box className="controls-launcher-copy">
          <Typography className="section-kicker">
            NEW RETAINED COMMAND
          </Typography>
          <Typography component="h2">Propose exact terms</Typography>
          <Typography>
            Every proposal expires within two hours. Expiry publishes disabled
            plus killed, regardless of the requested action.
          </Typography>
        </Box>
        <Box className="controls-launcher-sequence">
          <div>
            <span>01</span>Propose
          </div>
          <i />
          <div>
            <span>02</span>Approve
          </div>
          <i />
          <div>
            <span>03</span>Apply
          </div>
        </Box>
        <Button
          disabled={!loaded || Boolean(loadError)}
          onClick={() => setProposalOpen(true)}
          variant="contained"
        >
          Create proposal
        </Button>
      </AdminCard>

      <Dialog
        fullWidth
        maxWidth="sm"
        open={proposalOpen}
        onClose={() => {
          if (!busy) {
            setProposalOpen(false);
            setDialogError("");
          }
        }}
        aria-labelledby="control-proposal-title"
      >
        <DialogTitle id="control-proposal-title">
          Propose exact terms
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <FormControl>
              <InputLabel>Capability</InputLabel>
              <Select
                label="Capability"
                value={capability}
                onChange={(event) =>
                  setCapability(event.target.value as Capability)
                }
              >
                {Object.entries(labels).map(([key, label]) => (
                  <MenuItem key={key} value={key}>
                    {label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl>
              <InputLabel>Environment</InputLabel>
              <Select
                label="Environment"
                value={environment}
                onChange={(event) =>
                  setEnvironment(event.target.value as Environment)
                }
              >
                <MenuItem value="staging">Staging</MenuItem>
                <MenuItem value="production">Production</MenuItem>
              </Select>
            </FormControl>
            <FormControl>
              <InputLabel>Action</InputLabel>
              <Select
                label="Action"
                value={controlAction}
                onChange={(event) =>
                  setControlAction(event.target.value as ControlAction)
                }
              >
                <MenuItem value="enable">Enable</MenuItem>
                <MenuItem value="disable">Disable</MenuItem>
                <MenuItem value="kill">Kill immediately</MenuItem>
                <MenuItem value="unkill">Remove runtime kill</MenuItem>
              </Select>
            </FormControl>
            <FormControl>
              <InputLabel>Reason</InputLabel>
              <Select
                label="Reason"
                value={reason}
                onChange={(event) => setReason(event.target.value as Reason)}
              >
                <MenuItem value="staged_rollout">Staged rollout</MenuItem>
                <MenuItem value="incident">Incident</MenuItem>
                <MenuItem value="maintenance">Maintenance</MenuItem>
              </Select>
            </FormControl>
            <Alert
              severity={
                controlAction === "kill" || controlAction === "disable"
                  ? "warning"
                  : "info"
              }
            >
              {labels[capability]} will be {actionVerbs[controlAction]} in{" "}
              {environment} for Ghana. The retained command expires within two
              hours and then fails closed.
            </Alert>
            {dialogError ? (
              <Alert severity="error" role="alert" aria-live="assertive">
                {dialogError}
              </Alert>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setProposalOpen(false);
              setDialogError("");
            }}
          >
            Cancel
          </Button>
          <Button
            aria-busy={busy}
            disabled={busy}
            onClick={() => void mutate("propose")}
            variant="contained"
          >
            Submit retained proposal
          </Button>
        </DialogActions>
      </Dialog>

      <Stack className="controls-proposal-register" spacing={2}>
        <Box className="controls-register-heading">
          <Box>
            <Typography className="section-kicker">
              RETAINED REGISTER
            </Typography>
            <Typography component="h2">Active proposals</Typography>
          </Box>
          <Button disabled={loading} onClick={() => void load()}>
            Refresh
          </Button>
        </Box>
        {loading ? (
          <AdminCard variant="panel" watermark="queue" showWatermark={false}>
            <AdminSkeleton
              variant="card-list"
              rows={2}
              label="Loading runtime-control proposals"
            />
          </AdminCard>
        ) : loadError ? (
          <AdminCard variant="warning" watermark="queue" showWatermark={false}>
            <EmptyState
              icon="!"
              title="Controls unavailable"
              description={loadError}
              variant="warning"
              action={
                <Button onClick={() => void load()} variant="outlined">
                  Retry
                </Button>
              }
            />
          </AdminCard>
        ) : loaded && proposals.length ? (
          proposals.map((proposal) => (
            <AdminCard
              key={proposal.proposalId}
              variant="row"
              watermark="queue"
              className={`controls-proposal controls-proposal--${proposal.status}`}
            >
              <Box className="controls-proposal-layout">
                <Box className="controls-proposal-copy">
                  <Stack
                    className="controls-proposal-chips"
                    direction="row"
                    spacing={1}
                  >
                    <Chip
                      label={proposal.status}
                      color={
                        proposal.status === "applied" ? "success" : "warning"
                      }
                      size="small"
                    />
                    <Chip
                      label={`${proposal.environment} · ${proposal.market}`}
                      size="small"
                    />
                  </Stack>
                  <Typography component="h3">
                    {labels[proposal.capability]} · {proposal.action}
                  </Typography>
                  <Typography>
                    {proposal.reason.replaceAll("_", " ")} · expires{" "}
                    {new Date(proposal.expiresAt).toLocaleString()}
                  </Typography>
                </Box>
                <Stack
                  className="controls-proposal-actions"
                  direction="row"
                  spacing={1}
                >
                  {proposal.status === "proposed" ? (
                    <Button
                      disabled={busy}
                      onClick={() => {
                        setDialogError("");
                        setConfirmProposal(proposal);
                        setConfirmAction("approve");
                      }}
                      variant="outlined"
                    >
                      Approve as second admin
                    </Button>
                  ) : null}
                  {proposal.status === "approved" ? (
                    <Button
                      disabled={busy || !proposal.approvedByMe}
                      onClick={() => {
                        setDialogError("");
                        setConfirmProposal(proposal);
                        setConfirmAction("apply");
                      }}
                      variant="contained"
                    >
                      Apply retained terms
                    </Button>
                  ) : null}
                  {proposal.status === "applied" ? (
                    <Chip
                      label="Live until fail-closed expiry"
                      color="success"
                    />
                  ) : null}
                </Stack>
              </Box>
            </AdminCard>
          ))
        ) : loaded ? (
          <AdminCard variant="panel" watermark="queue" showWatermark={false}>
            <EmptyState
              icon="✓"
              title="No active proposals"
              description="There are no retained runtime-control proposals awaiting review or application."
            />
          </AdminCard>
        ) : null}
      </Stack>
      <Dialog
        fullWidth
        maxWidth="sm"
        open={Boolean(confirmProposal && confirmAction)}
        onClose={() => {
          if (!busy) {
            setConfirmProposal(null);
            setConfirmAction(null);
            setDialogError("");
          }
        }}
        aria-labelledby="control-action-confirm-title"
      >
        <DialogTitle id="control-action-confirm-title">
          {confirmAction === "approve"
            ? "Approve retained terms?"
            : "Apply retained terms?"}
        </DialogTitle>
        <DialogContent>
          {confirmProposal ? (
            <Stack spacing={1.5} sx={{ pt: 1 }}>
              <Typography>
                <strong>{labels[confirmProposal.capability]}</strong> ·{" "}
                {actionVerbs[confirmProposal.action]}
              </Typography>
              <Typography color="text.secondary">
                {confirmProposal.environment} · {confirmProposal.market} ·{" "}
                {confirmProposal.reason.replaceAll("_", " ")} · expires{" "}
                {new Date(confirmProposal.expiresAt).toLocaleString()}
              </Typography>
              <Alert severity={confirmAction === "apply" ? "warning" : "info"}>
                {confirmAction === "approve"
                  ? "A distinct stepped-up administrator must approve the exact retained proposal."
                  : "This applies the exact retained terms until fail-closed expiry."}
              </Alert>
              {dialogError ? (
                <Alert severity="error" role="alert" aria-live="assertive">
                  {dialogError}
                </Alert>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setConfirmProposal(null);
              setConfirmAction(null);
              setDialogError("");
            }}
          >
            Cancel
          </Button>
          <Button
            aria-busy={busy}
            disabled={busy || !confirmProposal || !confirmAction}
            onClick={() =>
              confirmProposal && confirmAction
                ? void mutate(confirmAction, confirmProposal.proposalId)
                : undefined
            }
            variant="contained"
          >
            Confirm {confirmAction}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="xs"
        onClose={() => {
          if (!busy) {
            setPending(null);
            setStepUpCode("");
            setDialogError("");
          }
        }}
        open={Boolean(pending)}
      >
        <DialogTitle>Fresh MFA required</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="info">
              No control state changes until fresh step-up succeeds.
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
          <Button
            disabled={busy}
            onClick={() => {
              setPending(null);
              setStepUpCode("");
              setDialogError("");
            }}
          >
            Cancel
          </Button>
          <Button disabled={busy} onClick={() => void stepUp("start")}>
            Send code
          </Button>
          <Button
            disabled={busy || stepUpCode.trim().length < 6}
            onClick={() => void stepUp("complete")}
            variant="contained"
          >
            Verify and continue
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
