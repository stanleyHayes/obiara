"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  Container,
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

export function ControlsDesk() {
  const [proposals, setProposals] = useState<Proposal[]>([]);
  const [capability, setCapability] = useState<Capability>("ai");
  const [environment, setEnvironment] = useState<Environment>("staging");
  const [controlAction, setControlAction] = useState<ControlAction>("kill");
  const [reason, setReason] = useState<Reason>("incident");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState("");
  const [pending, setPending] = useState<null | {
    action: "propose" | "approve" | "apply";
    proposalId?: string;
  }>(null);
  const [stepUpCode, setStepUpCode] = useState("");

  async function load() {
    setLoading(true);
    setMessage("");
    try {
      const response = await fetch("/api/controls");
      const payload = (await response.json()) as {
        proposals?: Proposal[];
        message?: string;
      };
      if (!response.ok)
        throw new Error(
          payload.message || "Runtime-control proposals could not be loaded.",
        );
      setProposals(payload.proposals ?? []);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Runtime-control proposals could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    void fetch("/api/controls")
      .then(async (response) => {
        const payload = (await response.json()) as {
          proposals?: Proposal[];
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "Runtime-control proposals could not be loaded.",
          );
        if (active) setProposals(payload.proposals ?? []);
      })
      .catch((error: unknown) => {
        if (active)
          setMessage(
            error instanceof Error
              ? error.message
              : "Runtime-control proposals could not be loaded.",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function mutate(
    action: "propose" | "approve" | "apply",
    proposalId?: string,
  ) {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/controls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "propose"
            ? {
                action,
                commandId: `control:${crypto.randomUUID()}`,
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
        if (response.status === 403) setPending({ action, proposalId });
        throw new Error(
          payload?.message || `The proposal could not be ${action}d.`,
        );
      }
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
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The runtime-control action failed.",
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
      } else if (pending) {
        const next = pending;
        setPending(null);
        setStepUpCode("");
        await mutate(next.action, next.proposalId);
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
    <Box
      sx={{
        bgcolor: "background.default",
        color: "text.primary",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{
            alignItems: { md: "center" },
            justifyContent: "space-between",
            mb: 5,
          }}
        >
          <Box>
            <Typography className="section-kicker">Runtime controls</Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 44, md: 72 },
                fontWeight: 800,
                letterSpacing: "-0.06em",
                lineHeight: 0.95,
                mt: 1,
              }}
            >
              Small scope. Clear expiry.
            </Typography>
            <Typography sx={{ color: "text.secondary", mt: 2 }}>
              Durable two-person controls for one capability, environment and
              market.
            </Typography>
          </Box>
          <Link href="/">
            <Button variant="outlined">Back to command centre</Button>
          </Link>
        </Stack>

        {message ? (
          <Alert severity="error" sx={{ mb: 2 }}>
            {message}
          </Alert>
        ) : null}
        {success ? (
          <Alert
            severity="success"
            onClose={() => setSuccess("")}
            sx={{ mb: 2 }}
          >
            {success}
          </Alert>
        ) : null}

        <Card sx={{ borderRadius: 1, p: 3 }}>
          <Typography component="h2" sx={{ fontSize: 30, fontWeight: 800 }}>
            Propose exact terms
          </Typography>
          <Typography sx={{ color: "text.secondary", mb: 3 }}>
            Every proposal expires within two hours. Expiry publishes disabled
            plus killed, regardless of the requested action.
          </Typography>
          <Box
            sx={{
              display: "grid",
              gap: 2,
              gridTemplateColumns: { xs: "1fr", md: "repeat(4, 1fr)" },
            }}
          >
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
          </Box>
          <Button
            disabled={busy}
            onClick={() => void mutate("propose")}
            sx={{ mt: 2 }}
            variant="contained"
          >
            Create stepped-up proposal
          </Button>
        </Card>

        <Stack spacing={2} sx={{ mt: 3 }}>
          <Stack
            direction="row"
            sx={{ alignItems: "center", justifyContent: "space-between" }}
          >
            <Typography component="h2" sx={{ fontSize: 30, fontWeight: 800 }}>
              Active proposals
            </Typography>
            <Button disabled={loading} onClick={() => void load()}>
              Refresh
            </Button>
          </Stack>
          {loading ? (
            <>
              <Skeleton height={130} />
              <Skeleton height={130} />
            </>
          ) : proposals.length ? (
            proposals.map((proposal) => (
              <Card key={proposal.proposalId} sx={{ borderRadius: 1, p: 2.5 }}>
                <Stack
                  direction={{ xs: "column", md: "row" }}
                  spacing={2}
                  sx={{
                    alignItems: { md: "center" },
                    justifyContent: "space-between",
                  }}
                >
                  <Box>
                    <Stack direction="row" spacing={1}>
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
                    <Typography
                      component="h3"
                      sx={{ fontSize: 24, fontWeight: 800, mt: 1 }}
                    >
                      {labels[proposal.capability]} · {proposal.action}
                    </Typography>
                    <Typography sx={{ color: "text.secondary" }}>
                      {proposal.reason.replaceAll("_", " ")} · expires{" "}
                      {new Date(proposal.expiresAt).toLocaleString()}
                    </Typography>
                  </Box>
                  <Stack direction="row" spacing={1}>
                    {proposal.status === "proposed" ? (
                      <Button
                        disabled={busy}
                        onClick={() =>
                          void mutate("approve", proposal.proposalId)
                        }
                        variant="outlined"
                      >
                        Approve as second admin
                      </Button>
                    ) : null}
                    {proposal.status === "approved" ? (
                      <Button
                        disabled={busy || !proposal.approvedByMe}
                        onClick={() =>
                          void mutate("apply", proposal.proposalId)
                        }
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
                </Stack>
              </Card>
            ))
          ) : (
            <Alert severity="info">No active runtime-control proposals.</Alert>
          )}
        </Stack>
      </Container>

      <Dialog
        fullWidth
        maxWidth="xs"
        onClose={() => setPending(null)}
        open={Boolean(pending)}
      >
        <DialogTitle>Fresh MFA required</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Alert severity="info">
              No control state changes until fresh step-up succeeds.
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
            Verify and continue
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
