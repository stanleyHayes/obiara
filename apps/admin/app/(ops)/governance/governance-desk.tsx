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
import { AdminCard } from "../../admin-card";
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
      const r = await fetch("/api/governance", {
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
      const r = await fetch("/api/governance", {
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
      const r = await fetch("/api/step-up", {
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
  return (
    <Stack spacing={3}>
      <Box>
        <Typography className="section-kicker">
          MARKET GOVERNANCE · LIVE
        </Typography>
        <Typography component="h1">
          Configuration needs a second set of eyes.
        </Typography>
        <Typography color="text.secondary">
          Draft bounded market configuration, publish only through a distinct
          operator, and retire without erasing history.
        </Typography>
      </Box>
      {loadError ? (
        <AdminCard variant="warning" watermark="evidence" showWatermark={false}>
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
      <Button
        onClick={() => setDraftOpen(true)}
        variant="contained"
        sx={{ alignSelf: "flex-start" }}
      >
        Draft market pack
      </Button>
      {!loading && !loadError ? (
        <Stack spacing={1.25}>
          {(["draft", "published", "retired"] as const).map((status) => (
            <AdminCard key={status} variant="metric" watermark="analytics">
              <Typography color="text.secondary">
                {status.toUpperCase()}
              </Typography>
              <Typography sx={{ fontSize: 30, fontWeight: 800 }}>
                {packs.filter((pack) => pack.status === status).length}
              </Typography>
            </AdminCard>
          ))}
        </Stack>
      ) : null}
      {loading ? (
        <AdminCard variant="panel" watermark="evidence" showWatermark={false}>
          <AdminSkeleton variant="card-list" rows={4} />
        </AdminCard>
      ) : !loadError && packs.length === 0 ? (
        <AdminCard variant="panel" watermark="evidence" showWatermark={false}>
          <EmptyState
            icon="✓"
            title="No market packs"
            description="No market packs have been drafted in this environment."
          />
        </AdminCard>
      ) : !loadError ? (
        <Stack spacing={1.5}>
          {packs.map((pack) => (
            <AdminCard
              key={pack.packId}
              component="article"
              variant="row"
              watermark="evidence"
            >
              <Stack spacing={1}>
                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  sx={{ justifyContent: "space-between", gap: 1 }}
                >
                  <Box>
                    <Typography
                      component="h2"
                      sx={{ fontFamily: "monospace", overflowWrap: "anywhere" }}
                    >
                      {pack.packId}
                    </Typography>
                    <Typography>
                      {labels[pack.market]} · v{pack.version}
                    </Typography>
                  </Box>
                  <Chip label={pack.status} />
                </Stack>
                <Typography color="text.secondary">
                  {pack.terminologyRef}
                </Typography>
                <Typography>
                  {Object.entries(pack.features)
                    .filter(([, v]) => v)
                    .map(([k]) => k)
                    .join(" · ") || "No capabilities enabled"}
                </Typography>
                {pack.status === "draft" ? (
                  <Button
                    disabled={busy || Boolean(pack.proposedByMe)}
                    onClick={() => setPending({ action: "publish", pack })}
                  >
                    {pack.proposedByMe
                      ? "Second operator required"
                      : "Review publication"}
                  </Button>
                ) : pack.status === "published" ? (
                  <Button
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
    </Stack>
  );
}
