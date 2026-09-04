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
  DialogContentText,
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { SegmentedOtpInput } from "@obiara/ui-web";
import { useCallback, useEffect, useRef, useState } from "react";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon } from "../../admin-icons";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";
import {
  markets,
  validPack,
  validPackResult,
  type Market,
  type Pack,
  type PendingPack,
} from "../../content-model";
import { adminFetch } from "../../lib/admin-fetch";
const labels: Record<Market, string> = {
    gh_en: "Ghana · English",
    gh_tw: "Ghana · Twi",
    gh_pidgin: "Ghana · Pidgin",
    gh_ga: "Ghana · Ga",
  },
  caps = ["sow", "fires", "ai", "payments", "gate"];
export function GovernanceDesk() {
  const [packs, setPacks] = useState<Pack[]>([]),
    [loading, setLoading] = useState(true),
    [loadError, setLoadError] = useState(""),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [draftOpen, setDraftOpen] = useState(false),
    [market, setMarket] = useState<Market>("gh_tw"),
    [ref, setRef] = useState(""),
    [features, setFeatures] = useState<Record<string, boolean>>({
      sow: true,
      fires: true,
      ai: false,
      payments: true,
      gate: true,
    }),
    [pending, setPending] = useState<PendingPack | null>(null),
    [mfa, setMfa] = useState(false),
    [otp, setOtp] = useState(""),
    [busy, setBusy] = useState(false);
  const mounted = useRef(false),
    loadGen = useRef(0),
    actionGen = useRef(0),
    stepGen = useRef(0),
    abort = useRef<AbortController | null>(null);
  const load = useCallback(async () => {
    const gen = ++loadGen.current;
    abort.current?.abort();
    const c = new AbortController();
    abort.current = c;
    setLoading(true);
    setLoadError("");
    try {
      const r = await adminFetch("/api/governance", {
          cache: "no-store",
          signal: c.signal,
        }),
        b: unknown = await r.json().catch(() => null);
      if (
        !r.ok ||
        !b ||
        typeof b !== "object" ||
        !("packs" in b) ||
        !Array.isArray(b.packs) ||
        !b.packs.every(validPack)
      )
        throw new Error(
          b &&
            typeof b === "object" &&
            "message" in b &&
            typeof b.message === "string"
            ? b.message
            : "Market-pack governance could not be loaded.",
        );
      if (mounted.current && gen === loadGen.current) setPacks(b.packs);
    } catch (e) {
      if (!c.signal.aborted && mounted.current && gen === loadGen.current)
        setLoadError(
          e instanceof Error
            ? e.message
            : "Market-pack governance could not be loaded.",
        );
    } finally {
      if (mounted.current && gen === loadGen.current) setLoading(false);
    }
  }, []);
  useEffect(() => {
    mounted.current = true;
    const t = setTimeout(() => void load(), 0);
    const loads = loadGen,
      actions = actionGen,
      steps = stepGen;
    return () => {
      clearTimeout(t);
      mounted.current = false;
      loads.current++;
      actions.current++;
      steps.current++;
      abort.current?.abort();
    };
  }, [load]);
  async function execute(snapshot: PendingPack) {
    const gen = ++actionGen.current;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const r = await adminFetch("/api/governance", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(
            snapshot.action === "draft"
              ? snapshot
              : {
                  action: snapshot.action,
                  packId: snapshot.pack.packId,
                },
          ),
        }),
        b: unknown = await r.json().catch(() => null);
      if (!mounted.current || gen !== actionGen.current) return;
      if (needsStepUp(r.status, errorCode(b))) {
        setMfa(true);
        return;
      }
      if (!r.ok || !validPackResult(b, snapshot))
        throw new Error(
          b &&
            typeof b === "object" &&
            "message" in b &&
            typeof b.message === "string"
            ? b.message
            : "The governance action could not be completed.",
        );
      setPending(null);
      setMfa(false);
      setOtp("");
      setError("");
      if (snapshot.action === "draft") setRef("");
      setNotice(
        snapshot.action === "draft"
          ? "Draft retained with an immutable audit record."
          : snapshot.action === "publish"
            ? "Pack published by a distinct operator."
            : "Pack retired; its history remains intact.",
      );
      await load();
    } catch (e) {
      if (mounted.current && gen === actionGen.current) {
        setError(
          e instanceof Error
            ? e.message
            : "The governance action could not be completed.",
        );
      }
    } finally {
      if (mounted.current && gen === actionGen.current) setBusy(false);
    }
  }
  async function step(action: "start" | "complete") {
    const gen = ++stepGen.current;
    setBusy(true);
    try {
      const r = await adminFetch("/api/step-up", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(
            action === "start" ? { action } : { action, code: otp },
          ),
        }),
        b = await r.json().catch(() => null);
      if (!mounted.current || gen !== stepGen.current) return;
      if (!r.ok) throw new Error(b?.message ?? "MFA failed.");
      if (action === "complete" && pending) {
        const exact = pending;
        setMfa(false);
        setOtp("");
        await execute(exact);
      } else setNotice("A step-up code was sent.");
    } catch (e) {
      if (mounted.current && gen === stepGen.current)
        setError(e instanceof Error ? e.message : "MFA failed.");
    } finally {
      if (mounted.current && gen === stepGen.current) setBusy(false);
    }
  }
  const counts = {
    draft: packs.filter((pack) => pack.status === "draft").length,
    published: packs.filter((pack) => pack.status === "published").length,
    retired: packs.filter((pack) => pack.status === "retired").length,
  };
  return (
    <Box className="governance-redesign">
      <Box component="header" className="governance-hero">
        <AdminCardWatermark watermark="evidence" />
        <Box className="governance-hero-copy">
          <Box className="governance-kicker">
            <AdminIcon name="governance" aria-hidden="true" />
            <Typography className="section-kicker">
              LANGUAGE AUTHORITY · LIVE
            </Typography>
          </Box>
          <Typography component="h1">Every word carries authority.</Typography>
          <Typography className="governance-hero-intro">
            Govern the vocabulary and capabilities each market receives. Every
            release is versioned, independently approved, and permanently
            traceable.
          </Typography>
        </Box>
        <Box className="governance-hero-register" aria-label="Authority rules">
          <div>
            <span>Release rule</span>
            <strong>Two operators</strong>
            <Typography>Proposer and publisher stay distinct</Typography>
          </div>
          <div>
            <span>Source of truth</span>
            <strong>Terminology registry</strong>
            <Typography>Language assets remain separately reviewed</Typography>
          </div>
          <div>
            <span>History</span>
            <strong>Immutable</strong>
            <Typography>Retired packs remain on the record</Typography>
          </div>
        </Box>
      </Box>

      <Box component="section" className="governance-boundary">
        <Box className="governance-boundary-icon">
          <AdminIcon name="governance" aria-hidden="true" />
        </Box>
        <Box>
          <Typography className="section-kicker">CONTROL BOUNDARY</Typography>
          <Typography component="h2">
            Configuration, not translation.
          </Typography>
          <Typography>
            This desk references reviewed language assets. It never uploads,
            rewrites, or invents market terminology.
          </Typography>
        </Box>
        <span className="governance-boundary-state">SEPARATION ENFORCED</span>
      </Box>

      {loadError ? (
        <AdminCard
          className="governance-state-card"
          variant="warning"
          watermark="evidence"
          showWatermark={false}
        >
          <EmptyState
            icon="!"
            title="Governance unavailable"
            description={loadError}
            variant="warning"
            action={<Button onClick={() => void load()}>Retry</Button>}
          />
        </AdminCard>
      ) : null}
      {notice ? <Alert severity="success">{notice}</Alert> : null}
      {!loading && !loadError ? (
        <Box className="governance-register">
          <Box className="governance-register-heading">
            <Box>
              <Typography className="section-kicker">
                MARKET REGISTER
              </Typography>
              <Typography component="h2">Language pack ledger</Typography>
            </Box>
            <Button onClick={() => setDraftOpen(true)} variant="contained">
              Draft market pack
            </Button>
          </Box>
          <Box className="governance-metrics">
            {(["draft", "published", "retired"] as const).map((status) => (
              <Box
                key={status}
                className={`governance-metric governance-metric--${status}`}
              >
                <span>{status}</span>
                <strong>{counts[status].toString().padStart(2, "0")}</strong>
                <small>
                  {status === "draft"
                    ? "awaiting an independent decision"
                    : status === "published"
                      ? "active market authorities"
                      : "preserved historical records"}
                </small>
              </Box>
            ))}
          </Box>
        </Box>
      ) : null}
      {loading ? (
        <AdminCard
          className="governance-state-card"
          variant="panel"
          watermark="evidence"
          showWatermark={false}
        >
          <AdminSkeleton variant="card-list" rows={4} />
        </AdminCard>
      ) : !loadError && packs.length === 0 ? (
        <AdminCard
          className="governance-empty"
          variant="panel"
          watermark="evidence"
          showWatermark={false}
        >
          <EmptyState
            icon="✓"
            title="No market packs"
            description="No market packs have been drafted in this environment."
          />
        </AdminCard>
      ) : !loadError ? (
        <Stack className="governance-pack-list" spacing={1.5}>
          {packs.map((pack) => (
            <AdminCard
              key={pack.packId}
              component="article"
              variant="row"
              watermark="evidence"
              className={`governance-pack governance-pack--${pack.status}`}
            >
              <Stack spacing={0}>
                <Box className="governance-pack-topline">
                  <span>{labels[pack.market]}</span>
                  <Chip label={pack.status} />
                </Box>
                <Box className="governance-pack-body">
                  <Box className="governance-pack-identity">
                    <Typography component="h3">
                      {labels[pack.market].split(" · ")[1]}
                    </Typography>
                    <Typography>{pack.packId}</Typography>
                  </Box>
                  <Box className="governance-pack-fact">
                    <span>Version</span>
                    <strong>v{pack.version}</strong>
                  </Box>
                  <Box className="governance-pack-fact governance-pack-fact--reference">
                    <span>Terminology reference</span>
                    <strong>{pack.terminologyRef}</strong>
                  </Box>
                  <Box className="governance-pack-fact">
                    <span>Capabilities</span>
                    <strong>
                      {Object.entries(pack.features)
                        .filter(([, v]) => v)
                        .map(([k]) => k)
                        .join(" · ") || "None enabled"}
                    </strong>
                  </Box>
                </Box>
                {pack.status === "draft" ? (
                  <Button
                    className="governance-pack-action"
                    disabled={busy || Boolean(pack.proposedByMe)}
                    onClick={() => setPending({ action: "publish", pack })}
                  >
                    {pack.proposedByMe
                      ? "Second operator required"
                      : "Review publication"}
                  </Button>
                ) : pack.status === "published" ? (
                  <Button
                    className="governance-pack-action"
                    color="warning"
                    disabled={busy}
                    onClick={() => setPending({ action: "retire", pack })}
                  >
                    Review retirement
                  </Button>
                ) : null}
              </Stack>
            </AdminCard>
          ))}
        </Stack>
      ) : null}
      {!loading && !loadError && packs.length === 0 ? (
        <Button
          className="governance-empty-action"
          onClick={() => setDraftOpen(true)}
          variant="contained"
        >
          Draft market pack
        </Button>
      ) : null}
      <Dialog
        open={draftOpen}
        aria-labelledby="draft-pack-title"
        aria-describedby="draft-pack-description"
        onClose={() => {
          if (!busy) setDraftOpen(false);
        }}
      >
        <DialogTitle id="draft-pack-title">Draft a market pack</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <DialogContentText id="draft-pack-description">
              Prepare exact bounded market configuration for immutable review.
            </DialogContentText>
            <TextField
              select
              label="Market"
              value={market}
              onChange={(e) => setMarket(e.target.value as Market)}
            >
              {markets.map((m) => (
                <MenuItem key={m} value={m}>
                  {labels[m]}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="Terminology registry reference"
              helperText="Identifies separately reviewed language assets; this desk does not upload or invent translations."
              value={ref}
              onChange={(e) => setRef(e.target.value.slice(0, 128))}
            />
            <Box component="fieldset" sx={{ border: 0, p: 0, m: 0 }}>
              <Typography component="legend">Capabilities</Typography>
              <Stack>
                {caps.map((c) => (
                  <FormControlLabel
                    key={c}
                    control={
                      <Checkbox
                        checked={features[c] ?? false}
                        onChange={(e) =>
                          setFeatures({ ...features, [c]: e.target.checked })
                        }
                      />
                    }
                    label={c}
                  />
                ))}
              </Stack>
            </Box>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setDraftOpen(false);
              setError("");
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={busy || ref.trim().length < 3}
            onClick={() => {
              const snapshot: PendingPack = {
                action: "draft",
                market,
                terminologyRef: ref.trim(),
                features: { ...features },
              };
              setPending(snapshot);
              setDraftOpen(false);
              setError("");
            }}
            variant="contained"
          >
            Create audited draft
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        open={Boolean(pending) && !mfa}
        aria-labelledby="pack-confirm-title"
        onClose={() => {
          if (!busy) {
            setPending(null);
            setError("");
            setOtp("");
          }
        }}
        aria-describedby="pack-confirm"
      >
        <DialogTitle id="pack-confirm-title">
          Confirm exact governance transition
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="pack-confirm">
            Review the immutable pack identity, version and target status.
          </DialogContentText>
          {error ? <Alert severity="error">{error}</Alert> : null}
          {pending?.action === "draft" ? (
            <>
              <Typography>
                <strong>Action:</strong> draft
              </Typography>
              <Typography>
                <strong>Market:</strong> {labels[pending.market]}
              </Typography>
              <Typography>
                <strong>Terminology:</strong> {pending.terminologyRef}
              </Typography>
              <Typography>
                <strong>Capabilities:</strong>{" "}
                {Object.entries(pending.features)
                  .filter(([, enabled]) => enabled)
                  .map(([name]) => name)
                  .join(", ") || "None"}
              </Typography>
            </>
          ) : pending ? (
            <>
              <Typography>
                <strong>Pack:</strong> {pending.pack.packId}
              </Typography>
              <Typography>
                <strong>Version:</strong> {pending.pack.version}
              </Typography>
              <Typography>
                <strong>Action:</strong> {pending.action}
              </Typography>
              <Typography>
                <strong>Market:</strong> {labels[pending.pack.market]}
              </Typography>
              <Typography>
                <strong>Terminology:</strong> {pending.pack.terminologyRef}
              </Typography>
              <Typography>
                <strong>Capabilities:</strong>{" "}
                {Object.entries(pending.pack.features)
                  .filter(([, enabled]) => enabled)
                  .map(([name]) => name)
                  .join(", ") || "None"}
              </Typography>
              <Alert severity="warning">
                Publication requires a different operator from the proposer.
                Retirement preserves immutable history.
              </Alert>
            </>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setPending(null);
              setError("");
              setOtp("");
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={busy || !pending}
            onClick={() => (pending ? void execute(pending) : undefined)}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        open={mfa}
        aria-labelledby="governance-mfa-title"
        aria-describedby="governance-mfa-description"
        onClose={() => {
          if (!busy) {
            setMfa(false);
            setPending(null);
            setOtp("");
            setError("");
          }
        }}
      >
        <DialogTitle id="governance-mfa-title">Fresh MFA required</DialogTitle>
        <DialogContent>
          <DialogContentText id="governance-mfa-description">
            Verify fresh authority to retry the exact confirmed governance
            command.
          </DialogContentText>
          {error ? (
            <Alert severity="error" role="alert">
              {error}
            </Alert>
          ) : null}
          <SegmentedOtpInput
            label="Six-digit code"
            value={otp}
            onChange={setOtp}
            disabled={busy}
          />
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setMfa(false);
              setPending(null);
              setOtp("");
              setError("");
            }}
          >
            Cancel
          </Button>
          <Button disabled={busy} onClick={() => void step("start")}>
            Send code
          </Button>
          <Button
            disabled={busy || otp.length !== 6}
            onClick={() => void step("complete")}
          >
            Verify and retry
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
