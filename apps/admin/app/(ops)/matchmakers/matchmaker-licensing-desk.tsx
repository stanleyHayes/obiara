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
  DialogContentText,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useReducer, useState } from "react";

type Profile = {
  matchmakerId: string;
  displayName: string;
  licenseId: string;
  jurisdiction: string;
  licenseVersion: number;
  licenseValidUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};

const initialForm = {
  matchmakerId: "",
  licenseId: "",
  jurisdiction: "ghana",
  expectedVersion: 0,
  validFrom: "",
  validUntil: "",
  minimumFeeGhs: "80",
  maximumFeeGhs: "250",
  displayName: "",
  languages: "Twi, English",
  specialties: "Consultation",
  completedEngagements: "0",
  rating: "0",
};

export function MatchmakerLicensingDesk() {
  const [profiles, setProfiles] = useReducer(
    (_: Profile[], next: Profile[]) => next,
    [],
  );
  const [form, setForm] = useState(initialForm);
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [stepUpOpen, setStepUpOpen] = useState(false);
  const [stepUpCode, setStepUpCode] = useState("");
  const [engagementID, setEngagementID] = useState("");
  const [fundingRef, setFundingRef] = useState("");
  const [escrowID, setEscrowID] = useState("");
  const [milestoneID, setMilestoneID] = useState("");
  const [openedAt] = useState(() => Date.now());

  const load = useCallback(async () => {
    const response = await fetch("/api/matchmakers", { cache: "no-store" });
    const payload = (await response.json().catch(() => null)) as {
      items?: Profile[];
      message?: string;
    } | null;
    if (!response.ok || !payload?.items) {
      setError(
        payload?.message ?? "The licensing register could not be loaded.",
      );
      return;
    }
    setProfiles(payload.items);
    setError(null);
  }, []);

  useEffect(() => {
    void Promise.resolve().then(load);
  }, [load]);

  function edit(profile?: Profile) {
    if (!profile) {
      setForm(initialForm);
    } else {
      setForm({
        matchmakerId: profile.matchmakerId,
        licenseId: profile.licenseId,
        jurisdiction: profile.jurisdiction,
        expectedVersion: profile.licenseVersion,
        validFrom: new Date().toISOString().slice(0, 16),
        validUntil: "",
        minimumFeeGhs: String(profile.minimumFeePesewas / 100),
        maximumFeeGhs: String(profile.maximumFeePesewas / 100),
        displayName: profile.displayName,
        languages: profile.languages.join(", "),
        specialties: profile.specialties.join(", "),
        completedEngagements: String(profile.completedEngagements),
        rating: String(profile.ratingBasisPoints / 100),
      });
    }
    setOpen(true);
  }

  async function save() {
    setBusy(true);
    const response = await fetch("/api/matchmakers", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        matchmakerId: form.matchmakerId || undefined,
        licenseId: form.licenseId.trim(),
        jurisdiction: form.jurisdiction.trim(),
        expectedVersion: form.expectedVersion,
        validFrom: new Date(form.validFrom).toISOString(),
        validUntil: new Date(form.validUntil).toISOString(),
        minimumFeePesewas: Math.round(Number(form.minimumFeeGhs) * 100),
        maximumFeePesewas: Math.round(Number(form.maximumFeeGhs) * 100),
        displayName: form.displayName.trim(),
        languages: form.languages
          .split(",")
          .map((value) => value.trim())
          .filter(Boolean),
        specialties: form.specialties
          .split(",")
          .map((value) => value.trim())
          .filter(Boolean),
        completedEngagements: Number(form.completedEngagements),
        ratingBasisPoints: Math.round(Number(form.rating) * 100),
      }),
    });
    const payload = (await response.json().catch(() => null)) as {
      message?: string;
    } | null;
    if (!response.ok) {
      if (response.status === 403) setStepUpOpen(true);
      setError(payload?.message ?? "The licence could not be retained.");
    } else {
      setOpen(false);
      setMessage(
        form.matchmakerId
          ? "The next licence version is active."
          : "The matchmaker licence was created.",
      );
      await load();
    }
    setBusy(false);
  }

  async function stepUp(action: "start" | "complete") {
    setBusy(true);
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
    if (!response.ok) {
      setError(payload?.message ?? "The MFA step-up failed.");
    } else if (action === "complete") {
      setStepUpOpen(false);
      setStepUpCode("");
      setMessage("MFA step-up is current. Submit the licence again.");
    } else {
      setMessage("A step-up code was sent to your admin email.");
    }
    setBusy(false);
  }

  async function escrowAction(action: "fund" | "delivery") {
    setBusy(true);
    setError(null);
    const response = await fetch("/api/escrows", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": `${action}.${crypto.randomUUID()}`,
      },
      body: JSON.stringify(
        action === "fund"
          ? {
              action,
              engagementId: engagementID.trim(),
              fundingRef: fundingRef.trim(),
            }
          : {
              action,
              escrowId: escrowID.trim(),
              milestoneId: milestoneID.trim(),
            },
      ),
    });
    const payload = (await response.json().catch(() => null)) as {
      escrowId?: string;
      message?: string;
    } | null;
    if (!response.ok) {
      if (response.status === 403) setStepUpOpen(true);
      setError(payload?.message ?? "The escrow action could not be retained.");
    } else {
      if (action === "fund" && payload?.escrowId) setEscrowID(payload.escrowId);
      setMessage(
        action === "fund"
          ? "Provider-confirmed funding was retained and audited."
          : "Delivery evidence was retained under operations authority.",
      );
    }
    setBusy(false);
  }

  return (
    <Box sx={{ bgcolor: "background.default", minHeight: "100vh", py: 4 }}>
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{ justifyContent: "space-between", mb: 4 }}
        >
          <Box>
            <Typography
              sx={{
                color: "#8e3159",
                fontSize: 12,
                fontWeight: 800,
                letterSpacing: 1.3,
              }}
            >
              AGYINA LICENSING
            </Typography>
            <Typography
              component="h1"
              sx={{
                fontSize: { xs: 40, md: 62 },
                fontWeight: 800,
                letterSpacing: "-0.05em",
                lineHeight: 0.98,
              }}
            >
              Licence the guide. Never the shortcut.
            </Typography>
            <Typography sx={{ color: "text.secondary", mt: 2, maxWidth: 720 }}>
              Only current, versioned records reach the member marketplace.
              Every change requires MFA and is atomically audited.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
            <Button onClick={() => edit()} variant="contained">
              Add matchmaker
            </Button>
            <Link href="/">
              <Button variant="outlined">Command centre</Button>
            </Link>
          </Stack>
        </Stack>
        {error ? (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {error}
          </Alert>
        ) : null}
        {message ? (
          <Alert severity="success" sx={{ mb: 2 }}>
            {message}
          </Alert>
        ) : null}
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: { xs: "1fr", md: "repeat(2,minmax(0,1fr))" },
          }}
        >
          {profiles.map((profile) => {
            const expiry = new Date(profile.licenseValidUntil).getTime();
            const current = expiry > openedAt;
            return (
              <Card key={profile.matchmakerId} sx={{ borderRadius: 1, p: 3 }}>
                <Stack
                  direction="row"
                  sx={{
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                  }}
                >
                  <Box>
                    <Typography sx={{ fontSize: 24, fontWeight: 800 }}>
                      {profile.displayName}
                    </Typography>
                    <Typography sx={{ color: "text.secondary" }}>
                      {profile.licenseId} · v{profile.licenseVersion}
                    </Typography>
                  </Box>
                  <Chip
                    color={current ? "success" : "default"}
                    label={current ? "Current" : "Expired"}
                  />
                </Stack>
                <Typography sx={{ mt: 2 }}>
                  {profile.specialties.join(" · ")}
                </Typography>
                <Typography sx={{ color: "text.secondary", mt: 0.5 }}>
                  {profile.languages.join(" / ")} · expires{" "}
                  {new Date(profile.licenseValidUntil).toLocaleDateString()}
                </Typography>
                <Button
                  onClick={() => edit(profile)}
                  sx={{ mt: 2 }}
                  variant="outlined"
                >
                  Issue next version
                </Button>
              </Card>
            );
          })}
        </Box>
        <Card sx={{ borderRadius: 1, mt: 3, p: 3 }}>
          <Typography
            sx={{
              color: "#8e3159",
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: 1.3,
            }}
          >
            ESCROW CONTROL
          </Typography>
          <Typography
            component="h2"
            sx={{ fontSize: 30, fontWeight: 800, mt: 0.5 }}
          >
            Two authorities. One immutable trail.
          </Typography>
          <Typography sx={{ color: "text.secondary", mt: 1, maxWidth: 760 }}>
            Funding is accepted only after provider confirmation and inherits
            the booked member, amount and milestones. Delivery evidence is
            separate from member acceptance. Both actions require fresh MFA and
            write an audit record atomically.
          </Typography>
          <Box
            sx={{
              display: "grid",
              gap: 2,
              gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
              mt: 3,
            }}
          >
            <Stack spacing={2}>
              <Typography sx={{ fontWeight: 800 }}>
                Retain confirmed funding
              </Typography>
              <TextField
                label="Booked engagement ID"
                value={engagementID}
                onChange={(event) => setEngagementID(event.target.value)}
              />
              <TextField
                helperText="Opaque 64-character provider reference"
                label="Provider funding reference"
                value={fundingRef}
                onChange={(event) =>
                  setFundingRef(event.target.value.toLowerCase())
                }
              />
              <Button
                disabled={
                  busy ||
                  !engagementID.trim() ||
                  !/^[a-f0-9]{64}$/.test(fundingRef.trim())
                }
                onClick={() => void escrowAction("fund")}
                variant="contained"
              >
                Retain funding
              </Button>
            </Stack>
            <Stack spacing={2}>
              <Typography sx={{ fontWeight: 800 }}>
                Confirm provider delivery
              </Typography>
              <TextField
                label="Escrow ID"
                value={escrowID}
                onChange={(event) => setEscrowID(event.target.value)}
              />
              <TextField
                label="Milestone ID"
                value={milestoneID}
                onChange={(event) => setMilestoneID(event.target.value)}
              />
              <Button
                disabled={busy || !escrowID.trim() || !milestoneID.trim()}
                onClick={() => void escrowAction("delivery")}
                variant="outlined"
              >
                Record delivery evidence
              </Button>
            </Stack>
          </Box>
        </Card>
      </Container>

      <Dialog
        fullWidth
        maxWidth="sm"
        open={open}
        onClose={() => setOpen(false)}
      >
        <DialogTitle>
          {form.matchmakerId
            ? "Issue next licence version"
            : "License a matchmaker"}
        </DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            Use public professional information only. Do not enter phone, email,
            member identity or private review text.
          </DialogContentText>
          <Stack spacing={2}>
            <TextField
              label="Display name"
              value={form.displayName}
              onChange={(event) =>
                setForm({ ...form, displayName: event.target.value })
              }
            />
            <TextField
              label="Licence reference"
              value={form.licenseId}
              onChange={(event) =>
                setForm({ ...form, licenseId: event.target.value })
              }
            />
            <TextField
              label="Jurisdiction"
              value={form.jurisdiction}
              onChange={(event) =>
                setForm({ ...form, jurisdiction: event.target.value })
              }
            />
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <TextField
                fullWidth
                label="Valid from"
                type="datetime-local"
                slotProps={{ inputLabel: { shrink: true } }}
                value={form.validFrom}
                onChange={(event) =>
                  setForm({ ...form, validFrom: event.target.value })
                }
              />
              <TextField
                fullWidth
                label="Valid until"
                type="datetime-local"
                slotProps={{ inputLabel: { shrink: true } }}
                value={form.validUntil}
                onChange={(event) =>
                  setForm({ ...form, validUntil: event.target.value })
                }
              />
            </Stack>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <TextField
                fullWidth
                label="Minimum fee (GHS)"
                value={form.minimumFeeGhs}
                onChange={(event) =>
                  setForm({ ...form, minimumFeeGhs: event.target.value })
                }
              />
              <TextField
                fullWidth
                label="Maximum fee (GHS)"
                value={form.maximumFeeGhs}
                onChange={(event) =>
                  setForm({ ...form, maximumFeeGhs: event.target.value })
                }
              />
            </Stack>
            <TextField
              label="Languages, comma separated"
              value={form.languages}
              onChange={(event) =>
                setForm({ ...form, languages: event.target.value })
              }
            />
            <TextField
              label="Specialties, comma separated"
              value={form.specialties}
              onChange={(event) =>
                setForm({ ...form, specialties: event.target.value })
              }
            />
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <TextField
                fullWidth
                label="Completed engagements"
                value={form.completedEngagements}
                onChange={(event) =>
                  setForm({ ...form, completedEngagements: event.target.value })
                }
              />
              <TextField
                fullWidth
                label="Completed-only rating (0–5)"
                value={form.rating}
                onChange={(event) =>
                  setForm({ ...form, rating: event.target.value })
                }
              />
            </Stack>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            disabled={busy || !form.validFrom || !form.validUntil}
            onClick={() => void save()}
            variant="contained"
          >
            Retain licence
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="xs"
        open={stepUpOpen}
        onClose={() => setStepUpOpen(false)}
      >
        <DialogTitle>Confirm licensing authority</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            A fresh MFA step-up is required before licensing changes.
          </DialogContentText>
          <Button
            disabled={busy}
            onClick={() => void stepUp("start")}
            variant="outlined"
          >
            Send step-up code
          </Button>
          <TextField
            fullWidth
            label="Six-digit code"
            sx={{ mt: 2 }}
            value={stepUpCode}
            onChange={(event) =>
              setStepUpCode(event.target.value.replace(/\D/g, "").slice(0, 6))
            }
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setStepUpOpen(false)}>Cancel</Button>
          <Button
            disabled={busy || stepUpCode.length !== 6}
            onClick={() => void stepUp("complete")}
            variant="contained"
          >
            Verify code
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
