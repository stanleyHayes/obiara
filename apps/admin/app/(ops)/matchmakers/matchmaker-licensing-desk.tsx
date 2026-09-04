"use client";

import { errorCode, needsStepUp } from "../../lib/step-up";

import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { SegmentedOtpInput } from "@obiara/ui-web";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AdminCard } from "../../admin-card";
import {
  matchmakerFieldErrors,
  normalizedUniqueList,
  validEscrowWriteResult,
  validLicenseResult,
  validMatchmakerProfile,
  type LicenseBody,
  type MatchmakerField,
  type MatchmakerProfile,
  type PendingCommercial,
} from "../../commercial-model";
import { EmptyState } from "../../empty-state";
import { AdminSkeleton } from "../../loading-skeleton";
import { adminFetch } from "../../lib/admin-fetch";

export type { MatchmakerProfile } from "../../commercial-model";
type Mode = "list" | "form" | "escrow";
const blank = {
  matchmakerId: "",
  licenseId: "",
  jurisdiction: "ghana",
  expectedVersion: 0,
  validFrom: "",
  validUntil: "",
  minimumFeeGhs: "",
  maximumFeeGhs: "",
  displayName: "",
  languages: "",
  specialties: "",
  completedEngagements: "",
  rating: "",
};
type Form = typeof blank;
const split = normalizedUniqueList;
const newCommand = (kind: "fund" | "delivery") =>
  `${kind}.${crypto.randomUUID()}`;
function fromProfile(profile: MatchmakerProfile): Form {
  return {
    matchmakerId: profile.matchmakerId,
    licenseId: profile.licenseId,
    jurisdiction: profile.jurisdiction,
    expectedVersion: profile.licenseVersion,
    validFrom: "",
    validUntil: "",
    minimumFeeGhs: String(profile.minimumFeePesewas / 100),
    maximumFeeGhs: String(profile.maximumFeePesewas / 100),
    displayName: profile.displayName,
    languages: profile.languages.join(", "),
    specialties: profile.specialties.join(", "),
    completedEngagements: String(profile.completedEngagements),
    rating: String(profile.ratingBasisPoints / 100),
  };
}

export function validateMatchmakerForm(form: Form) {
  return Object.values(matchmakerFieldErrors(form))[0] ?? "";
}

export function MatchmakerLicensingDesk({
  mode = "list",
  matchmakerId,
}: {
  mode?: Mode;
  matchmakerId?: string;
}) {
  const router = useRouter(),
    needsRegister = mode === "list" || Boolean(matchmakerId);
  const [items, setItems] = useState<MatchmakerProfile[]>([]),
    [loaded, setLoaded] = useState(!needsRegister),
    [loading, setLoading] = useState(needsRegister),
    [loadError, setLoadError] = useState("");
  const [form, setForm] = useState<Form>(blank),
    [busy, setBusy] = useState(false),
    [notice, setNotice] = useState(""),
    [localError, setLocalError] = useState("");
  const [engagementId, setEngagementId] = useState(""),
    [fundingRef, setFundingRef] = useState(""),
    [escrowId, setEscrowId] = useState(""),
    [milestoneId, setMilestoneId] = useState("");
  const [pending, setPending] = useState<PendingCommercial | null>(null),
    [mfaOpen, setMfaOpen] = useState(false),
    [otp, setOtp] = useState("");
  const [now, setNow] = useState(() => Date.now());
  const mounted = useRef(false),
    loadGeneration = useRef(0),
    actionGeneration = useRef(0),
    stepUpGeneration = useRef(0),
    abortRef = useRef<AbortController | null>(null),
    fundKey = useRef(newCommand("fund")),
    deliveryKey = useRef(newCommand("delivery"));
  const fieldRefs = useRef<
    Partial<Record<MatchmakerField, HTMLInputElement | null>>
  >({});
  const load = useCallback(async () => {
    if (!needsRegister) return;
    const generation = ++loadGeneration.current;
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setLoading(true);
    setLoaded(false);
    setLoadError("");
    try {
      const response = await adminFetch("/api/matchmakers", {
        cache: "no-store",
        signal: controller.signal,
      });
      const body = (await response.json().catch(() => null)) as {
        items?: MatchmakerProfile[];
        message?: string;
      } | null;
      if (
        !response.ok ||
        !Array.isArray(body?.items) ||
        !body.items.every(validMatchmakerProfile)
      )
        throw new Error(
          body?.message ?? "The licensing register could not be loaded.",
        );
      if (mounted.current && generation === loadGeneration.current) {
        setItems(body.items);
        setLoaded(true);
      }
    } catch (error) {
      if (
        !controller.signal.aborted &&
        mounted.current &&
        generation === loadGeneration.current
      )
        setLoadError(
          error instanceof Error
            ? error.message
            : "The licensing register could not be loaded.",
        );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }, [needsRegister]);
  useEffect(() => {
    mounted.current = true;
    const timer = window.setTimeout(() => void load(), 0);
    const loads = loadGeneration,
      actions = actionGeneration,
      stepUps = stepUpGeneration;
    return () => {
      window.clearTimeout(timer);
      mounted.current = false;
      loads.current++;
      actions.current++;
      stepUps.current++;
      abortRef.current?.abort();
    };
  }, [load]);
  const profile = items.find((item) => item.matchmakerId === matchmakerId);
  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (mode === "form" && loaded)
        setForm(profile ? fromProfile(profile) : blank);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loaded, mode, profile]);
  const fieldErrors = useMemo(() => matchmakerFieldErrors(form), [form]);
  useEffect(() => {
    if (mode !== "list" || !items.length) return;
    const future = items
      .map((item) => Date.parse(item.licenseValidUntil))
      .filter((time) => Number.isFinite(time) && time > now)
      .sort((a, b) => a - b)[0];
    if (!future) return;
    const timer = window.setTimeout(
      () => setNow(Date.now()),
      Math.min(future - now + 100, 2_147_000_000),
    );
    return () => window.clearTimeout(timer);
  }, [items, mode, now]);
  function licenseSnapshot(): LicenseBody {
    return {
      matchmakerId: form.matchmakerId || undefined,
      licenseId: form.licenseId.trim(),
      jurisdiction: form.jurisdiction.trim(),
      expectedVersion: form.matchmakerId ? form.expectedVersion : 0,
      validFrom: new Date(form.validFrom).toISOString(),
      validUntil: new Date(form.validUntil).toISOString(),
      minimumFeePesewas: Math.round(Number(form.minimumFeeGhs) * 100),
      maximumFeePesewas: Math.round(Number(form.maximumFeeGhs) * 100),
      displayName: form.displayName.trim(),
      languages: split(form.languages),
      specialties: split(form.specialties),
      completedEngagements: Number(form.completedEngagements),
      ratingBasisPoints: Math.round(Number(form.rating) * 100),
    };
  }

  async function execute(snapshot: PendingCommercial) {
    const generation = ++actionGeneration.current;
    setBusy(true);
    setLocalError("");
    try {
      const response = await fetch(
        snapshot.kind === "license" ? "/api/matchmakers" : "/api/escrows",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...("key" in snapshot ? { "Idempotency-Key": snapshot.key } : {}),
          },
          body: JSON.stringify(snapshot.body),
        },
      );
      const payload = (await response.json().catch(() => null)) as unknown;
      const responseMessage =
        payload &&
        typeof payload === "object" &&
        "message" in payload &&
        typeof payload.message === "string"
          ? payload.message
          : undefined;
      if (!mounted.current || generation !== actionGeneration.current) return;
      if (!response.ok) {
        if (needsStepUp(response.status, errorCode(payload))) setMfaOpen(true);
        throw new Error(responseMessage ?? "The retained action failed.");
      }
      if (
        snapshot.kind === "license"
          ? !validLicenseResult(payload, snapshot)
          : !validEscrowWriteResult(payload, snapshot)
      )
        throw new Error(
          "The server returned an invalid retained-action result.",
        );
      setPending(null);
      setMfaOpen(false);
      setOtp("");
      if (snapshot.kind === "license") {
        router.replace("/matchmakers");
        router.refresh();
        return;
      }
      if (snapshot.kind === "fund") {
        setEscrowId((payload as { escrowId: string }).escrowId);
        deliveryKey.current = newCommand("delivery");
        setEngagementId("");
        setFundingRef("");
        fundKey.current = newCommand("fund");
        setNotice("Provider-confirmed funding was retained and audited.");
      } else {
        setMilestoneId("");
        deliveryKey.current = newCommand("delivery");
        setNotice("Delivery evidence was retained under operations authority.");
      }
    } catch (error) {
      if (mounted.current && generation === actionGeneration.current)
        setLocalError(
          error instanceof Error
            ? error.message
            : "The retained action failed.",
        );
    } finally {
      if (mounted.current && generation === actionGeneration.current)
        setBusy(false);
    }
  }
  async function stepUp(action: "start" | "complete") {
    const generation = ++stepUpGeneration.current;
    let retrying = false;
    setBusy(true);
    setLocalError("");
    try {
      const response = await adminFetch("/api/step-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "start" ? { action } : { action, code: otp },
        ),
      });
      const body = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!mounted.current || generation !== stepUpGeneration.current) return;
      if (!response.ok)
        throw new Error(body?.message ?? "The MFA step-up failed.");
      if (action === "complete" && pending) {
        const snapshot = pending;
        setMfaOpen(false);
        setOtp("");
        retrying = true;
        await execute(snapshot);
      } else setNotice("A step-up code was sent to your admin email.");
    } catch (error) {
      if (mounted.current && generation === stepUpGeneration.current)
        setLocalError(
          error instanceof Error ? error.message : "The MFA step-up failed.",
        );
    } finally {
      if (
        !retrying &&
        mounted.current &&
        generation === stepUpGeneration.current
      )
        setBusy(false);
    }
  }

  const fields = [
    ["displayName", "Display name"],
    ["licenseId", "Licence reference"],
    ["jurisdiction", "Jurisdiction"],
    ["validFrom", "Valid from"],
    ["validUntil", "Valid until"],
    ["minimumFeeGhs", "Minimum fee (GHS)"],
    ["maximumFeeGhs", "Maximum fee (GHS)"],
    ["languages", "Languages, comma separated"],
    ["specialties", "Specialties, comma separated"],
    ["completedEngagements", "Completed engagements"],
    ["rating", "Completed-only rating (0–5)"],
  ] as const;
  return (
    <Box sx={{ minHeight: "100vh", py: 4 }}>
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{ justifyContent: "space-between", mb: 4 }}
        >
          <Box>
            <Typography className="section-kicker">AGYINA LICENSING</Typography>
            <Typography
              component="h1"
              sx={{ fontSize: { xs: 40, md: 62 }, fontWeight: 800 }}
            >
              Licence the guide. Never the shortcut.
            </Typography>
            <Typography color="text.secondary">
              Only non-expired, versioned records reach the member marketplace.
              Every change requires MFA and is atomically audited.
            </Typography>
          </Box>
          <Stack direction="row" sx={{ gap: 1, flexWrap: "wrap" }}>
            <Button component={Link} href="/matchmakers/new">
              Add matchmaker
            </Button>
            <Button component={Link} href="/matchmakers/escrow">
              Escrow workflows
            </Button>
            <Button component={Link} href="/">
              Command centre
            </Button>
          </Stack>
        </Stack>
        {notice ? (
          <Alert severity="success" sx={{ mb: 2 }}>
            {notice}
          </Alert>
        ) : null}
        {loading ? (
          <AdminCard variant="panel" watermark="identity" showWatermark={false}>
            <AdminSkeleton
              variant={mode === "form" ? "form" : "card-list"}
              rows={4}
            />
          </AdminCard>
        ) : loadError ? (
          <AdminCard
            variant="warning"
            watermark="identity"
            showWatermark={false}
          >
            <EmptyState
              icon="!"
              title="Licensing unavailable"
              description={loadError}
              variant="warning"
              action={<Button onClick={() => void load()}>Retry</Button>}
            />
          </AdminCard>
        ) : null}
        {!loading && !loadError && mode === "list" && loaded ? (
          items.length ? (
            <Stack spacing={1.5}>
              {items.map((item) => (
                <AdminCard
                  key={item.matchmakerId}
                  variant="row"
                  watermark="identity"
                >
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    sx={{ justifyContent: "space-between", gap: 2 }}
                  >
                    <Box>
                      <Typography
                        component="h2"
                        sx={{ fontSize: 24, fontWeight: 800 }}
                      >
                        {item.displayName}
                      </Typography>
                      <Typography color="text.secondary">
                        {item.licenseId} · v{item.licenseVersion} ·{" "}
                        {item.languages.join(" / ")}
                      </Typography>
                    </Box>
                    <Stack direction="row" sx={{ gap: 1, flexWrap: "wrap" }}>
                      <Chip
                        label={
                          Date.parse(item.licenseValidUntil) > now
                            ? "Non-expired"
                            : "Expired"
                        }
                      />
                      <Button
                        component={Link}
                        href={`/matchmakers/${encodeURIComponent(item.matchmakerId)}`}
                      >
                        Issue next version
                      </Button>
                    </Stack>
                  </Stack>
                </AdminCard>
              ))}
            </Stack>
          ) : (
            <AdminCard
              variant="panel"
              watermark="identity"
              showWatermark={false}
            >
              <EmptyState
                icon="✓"
                title="No licensed matchmakers"
                description="No versioned licensing records are currently retained."
              />
            </AdminCard>
          )
        ) : null}
        {!loading && !loadError && mode === "form" && loaded ? (
          !matchmakerId || profile ? (
            <AdminCard variant="form" watermark="identity">
              <Stack spacing={2}>
                <Typography component="h2">
                  {matchmakerId
                    ? "Issue next licence version"
                    : "License a matchmaker"}
                </Typography>
                <Alert severity="info">
                  Use public professional information only. Do not enter phone,
                  email, member identity or private review text.
                </Alert>
                {fields.map(([key, label]) => (
                  <TextField
                    key={key}
                    label={label}
                    type={
                      key.startsWith("valid")
                        ? "datetime-local"
                        : [
                              "minimumFeeGhs",
                              "maximumFeeGhs",
                              "completedEngagements",
                              "rating",
                            ].includes(key)
                          ? "number"
                          : "text"
                    }
                    inputRef={(element) => {
                      fieldRefs.current[key] = element;
                    }}
                    error={Boolean(fieldErrors[key])}
                    helperText={fieldErrors[key]}
                    slotProps={{
                      ...(key.startsWith("valid")
                        ? { inputLabel: { shrink: true } }
                        : {}),
                      htmlInput: {
                        inputMode: [
                          "minimumFeeGhs",
                          "maximumFeeGhs",
                          "rating",
                        ].includes(key)
                          ? "decimal"
                          : key === "completedEngagements"
                            ? "numeric"
                            : undefined,
                        min: ["minimumFeeGhs", "maximumFeeGhs"].includes(key)
                          ? 0.01
                          : key === "completedEngagements" || key === "rating"
                            ? 0
                            : undefined,
                        max: key === "rating" ? 5 : undefined,
                        step: key === "completedEngagements" ? 1 : "any",
                      },
                    }}
                    value={form[key]}
                    onChange={(event) =>
                      setForm({ ...form, [key]: event.target.value })
                    }
                  />
                ))}
                <Button
                  disabled={busy}
                  onClick={() => {
                    const first = fields.find(([key]) => fieldErrors[key]);
                    if (first) {
                      fieldRefs.current[first[0]]?.focus();
                      return;
                    }
                    setPending({ kind: "license", body: licenseSnapshot() });
                  }}
                >
                  Review licence
                </Button>
              </Stack>
            </AdminCard>
          ) : (
            <EmptyState
              icon="!"
              title="Matchmaker not found"
              description="The exact requested licensing record is unavailable."
            />
          )
        ) : null}
        {mode === "escrow" ? (
          <Stack spacing={2}>
            {(["fund", "delivery"] as const).map((kind) => (
              <AdminCard key={kind} variant="form" watermark="evidence">
                <Stack spacing={2}>
                  <Typography component="h2">
                    {kind === "fund"
                      ? "Retain confirmed funding"
                      : "Confirm provider delivery"}
                  </Typography>
                  {kind === "fund" ? (
                    <>
                      <TextField
                        label="Booked engagement ID"
                        value={engagementId}
                        onChange={(event) => {
                          setEngagementId(event.target.value);
                          fundKey.current = newCommand("fund");
                        }}
                      />
                      <TextField
                        helperText="Opaque 64-character provider reference"
                        label="Provider funding reference"
                        value={fundingRef}
                        onChange={(event) => {
                          setFundingRef(event.target.value.toLowerCase());
                          fundKey.current = newCommand("fund");
                        }}
                      />
                      <Button
                        disabled={
                          busy ||
                          !engagementId.trim() ||
                          !/^[a-f0-9]{64}$/.test(fundingRef.trim())
                        }
                        onClick={() =>
                          setPending({
                            kind: "fund",
                            key: fundKey.current,
                            body: {
                              action: "fund",
                              engagementId: engagementId.trim(),
                              fundingRef: fundingRef.trim(),
                            },
                          })
                        }
                      >
                        Review funding
                      </Button>
                    </>
                  ) : (
                    <>
                      <TextField
                        label="Escrow ID"
                        value={escrowId}
                        onChange={(event) => {
                          setEscrowId(event.target.value);
                          deliveryKey.current = newCommand("delivery");
                        }}
                      />
                      <TextField
                        label="Milestone ID"
                        value={milestoneId}
                        onChange={(event) => {
                          setMilestoneId(event.target.value);
                          deliveryKey.current = newCommand("delivery");
                        }}
                      />
                      <Button
                        disabled={
                          busy || !escrowId.trim() || !milestoneId.trim()
                        }
                        onClick={() =>
                          setPending({
                            kind: "delivery",
                            key: deliveryKey.current,
                            body: {
                              action: "delivery",
                              escrowId: escrowId.trim(),
                              milestoneId: milestoneId.trim(),
                            },
                          })
                        }
                      >
                        Review delivery evidence
                      </Button>
                    </>
                  )}
                </Stack>
              </AdminCard>
            ))}
          </Stack>
        ) : null}
      </Container>
      <Dialog
        aria-labelledby="commercial-confirm-title"
        aria-describedby="commercial-confirm-description"
        open={Boolean(pending) && !mfaOpen}
        onClose={() => {
          if (!busy) {
            setPending(null);
            setLocalError("");
          }
        }}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle id="commercial-confirm-title">
          Confirm exact retained terms
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="commercial-confirm-description">
            Review the exact server-bound terms. They remain unchanged through
            any MFA retry.
          </DialogContentText>
          {pending ? (
            <Stack spacing={1.25} sx={{ pt: 1 }}>
              <Typography>
                <strong>Action:</strong> {pending.kind}
              </Typography>
              {pending.kind === "license" ? (
                <>
                  <Typography>
                    <strong>Matchmaker:</strong>{" "}
                    {pending.body.matchmakerId ??
                      "New profile (server assigns the ID)"}
                  </Typography>
                  <Typography>
                    <strong>Licence:</strong> {pending.body.licenseId} ·
                    expected v{pending.body.expectedVersion}
                  </Typography>
                  <Typography>
                    <strong>Display name:</strong> {pending.body.displayName}
                  </Typography>
                  <Typography>
                    <strong>Jurisdiction:</strong> {pending.body.jurisdiction}
                  </Typography>
                  <Typography>
                    <strong>Languages:</strong>{" "}
                    {pending.body.languages.join(", ")}
                  </Typography>
                  <Typography>
                    <strong>Specialties:</strong>{" "}
                    {pending.body.specialties.join(", ")}
                  </Typography>
                  <Typography>
                    <strong>Completed engagements:</strong>{" "}
                    {pending.body.completedEngagements}
                  </Typography>
                  <Typography>
                    <strong>Completed-only rating:</strong>{" "}
                    {(pending.body.ratingBasisPoints / 100).toFixed(2)} / 5
                  </Typography>
                  <Typography>
                    <strong>Validity:</strong> {pending.body.validFrom} to{" "}
                    {pending.body.validUntil}
                  </Typography>
                  <Typography>
                    <strong>Fees:</strong> GHS{" "}
                    {(pending.body.minimumFeePesewas / 100).toFixed(2)}–
                    {(pending.body.maximumFeePesewas / 100).toFixed(2)}
                  </Typography>
                </>
              ) : pending.kind === "fund" ? (
                <>
                  <Typography>
                    <strong>Engagement:</strong> {pending.body.engagementId}
                  </Typography>
                  <Typography sx={{ overflowWrap: "anywhere" }}>
                    <strong>Provider reference:</strong>{" "}
                    {pending.body.fundingRef}
                  </Typography>
                </>
              ) : (
                <>
                  <Typography>
                    <strong>Escrow:</strong> {pending.body.escrowId}
                  </Typography>
                  <Typography>
                    <strong>Milestone:</strong> {pending.body.milestoneId}
                  </Typography>
                </>
              )}
              {localError ? (
                <Alert severity="error" role="alert" aria-live="assertive">
                  {localError}
                </Alert>
              ) : null}
            </Stack>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setPending(null);
              setLocalError("");
            }}
          >
            Cancel
          </Button>
          <Button
            aria-busy={busy}
            disabled={busy || !pending}
            onClick={() => (pending ? void execute(pending) : undefined)}
            variant="contained"
          >
            Confirm retained action
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        aria-labelledby="commercial-mfa-title"
        aria-describedby="commercial-mfa-description"
        open={mfaOpen}
        onClose={() => {
          if (!busy) {
            setMfaOpen(false);
            setPending(null);
            setOtp("");
            setLocalError("");
          }
        }}
        fullWidth
        maxWidth="xs"
      >
        <DialogTitle id="commercial-mfa-title">Fresh MFA required</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <DialogContentText id="commercial-mfa-description">
              Verify fresh commercial authority to retry the exact confirmed
              terms.
            </DialogContentText>
            {localError ? (
              <Alert severity="error" role="alert" aria-live="assertive">
                {localError}
              </Alert>
            ) : null}
            <SegmentedOtpInput
              label="Six-digit code"
              value={otp}
              onChange={setOtp}
              disabled={busy}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={busy}
            onClick={() => {
              setMfaOpen(false);
              setPending(null);
              setOtp("");
              setLocalError("");
            }}
          >
            Cancel
          </Button>
          <Button disabled={busy} onClick={() => void stepUp("start")}>
            Send code
          </Button>
          <Button
            disabled={busy || otp.length !== 6}
            onClick={() => void stepUp("complete")}
          >
            Verify and retry
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
