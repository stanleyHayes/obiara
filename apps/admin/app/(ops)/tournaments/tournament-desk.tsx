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
  DialogContentText,
  DialogTitle,
  LinearProgress,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { SegmentedOtpInput } from "@obiara/ui-web";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon } from "../../admin-icons";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";
import {
  validCohort,
  validCompetition,
  validReviewResult,
  type Cohort,
  type Competition,
  type PendingTournament,
} from "../../content-model";
import { adminFetch } from "../../lib/admin-fetch";
type Mode = "landing" | "cohort" | "competition";
const command = (kind: string) => `${kind}.${crypto.randomUUID()}`;
export function TournamentDesk({
  mode = "landing",
  cohortId,
  competitionId,
}: {
  mode?: Mode;
  cohortId?: string;
  competitionId?: string;
}) {
  const router = useRouter(),
    [reference, setReference] = useState(""),
    [createOpen, setCreateOpen] = useState(false),
    [capacity, setCapacity] = useState<4 | 8 | 16>(4),
    [cohort, setCohort] = useState<Cohort | null>(null),
    [competition, setCompetition] = useState<Competition | null>(null),
    [loading, setLoading] = useState(mode !== "landing"),
    [loadError, setLoadError] = useState(""),
    [loadStatus, setLoadStatus] = useState(0),
    [actionError, setActionError] = useState(""),
    [notice, setNotice] = useState(""),
    [pending, setPending] = useState<PendingTournament | null>(null),
    [mfa, setMfa] = useState(false),
    [otp, setOtp] = useState(""),
    [busy, setBusy] = useState(false);
  const mounted = useRef(false),
    loadGen = useRef(0),
    actionGen = useRef(0),
    stepGen = useRef(0),
    abort = useRef<AbortController | null>(null),
    keys = useRef(new Map<string, string>());
  const keyFor = (terms: string) => {
    if (!keys.current.has(terms))
      keys.current.set(terms, command("tournament"));
    return keys.current.get(terms)!;
  };
  const load = useCallback(async () => {
    if (mode === "landing" || !cohortId) return;
    const gen = ++loadGen.current;
    abort.current?.abort();
    const c = new AbortController();
    abort.current = c;
    setLoading(true);
    setLoadError("");
    setLoadStatus(0);
    let responseStatus = 0;
    try {
      const query = `cohortId=${encodeURIComponent(cohortId)}${competitionId ? `&competitionId=${encodeURIComponent(competitionId)}` : ""}`,
        r = await fetch(`/api/game-cohorts?${query}`, {
          signal: c.signal,
          cache: "no-store",
        }),
        b: unknown = await r.json().catch(() => null);
      responseStatus = r.status;
      if (!r.ok || (competitionId ? !validCompetition(b) : !validCohort(b)))
        throw new Error(
          b &&
            typeof b === "object" &&
            "message" in b &&
            typeof b.message === "string"
            ? b.message
            : "The requested tournament record is unavailable.",
        );
      if (!mounted.current || gen !== loadGen.current) return;
      if (competitionId) {
        if ((b as Competition).id !== competitionId)
          throw new Error("The exact competition was not returned.");
        setCompetition(b as Competition);
      } else {
        if ((b as Cohort).id !== cohortId)
          throw new Error("The exact cohort was not returned.");
        setCohort(b as Cohort);
      }
    } catch (e) {
      if (!c.signal.aborted && mounted.current && gen === loadGen.current) {
        setLoadStatus(responseStatus);
        setLoadError(
          e instanceof Error
            ? e.message
            : "The requested tournament record is unavailable.",
        );
      }
    } finally {
      if (mounted.current && gen === loadGen.current) setLoading(false);
    }
  }, [mode, cohortId, competitionId]);
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
  function payload(p: PendingTournament) {
    if (p.kind === "create") return { action: "create", capacity: p.capacity };
    if (p.kind === "start")
      return {
        action: "start",
        cohortId: p.cohortId,
        expectedRevision: p.expectedRevision,
      };
    return {
      action: p.appeal ? "resolve-appeal" : "resolve-review",
      cohortId: p.cohortId,
      competitionId: p.competitionId,
      reviewId: p.reviewId,
      decision: p.decision,
      expectedRevision: p.expectedRevision,
    };
  }
  async function execute(p: PendingTournament) {
    const gen = ++actionGen.current;
    setBusy(true);
    setActionError("");
    try {
      const r = await adminFetch("/api/game-cohorts", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": p.key,
          },
          body: JSON.stringify(payload(p)),
        }),
        b: unknown = await r.json().catch(() => null);
      if (!mounted.current || gen !== actionGen.current) return;
      if (needsStepUp(r.status, errorCode(b)) && p.kind === "review") {
        setMfa(true);
        return;
      }
      if (!r.ok) {
        // On revision conflict: drop stale pending and cached key, refresh data.
        // Operator can retry with current revision; new command gets fresh idempotency key.
        if (
          r.status === 409 &&
          b &&
          typeof b === "object" &&
          "code" in b &&
          b.code === "competition_conflict"
        ) {
          setPending(null);
          // Reconstruct terms to delete the cached idempotency key, ensuring
          // next attempt generates a fresh key with the current revision.
          if (p.kind === "create") {
            keys.current.delete(
              JSON.stringify({ kind: "create", capacity: p.capacity }),
            );
          } else if (p.kind === "start") {
            keys.current.delete(
              JSON.stringify({
                kind: "start",
                cohortId: p.cohortId,
                expectedRevision: p.expectedRevision,
              }),
            );
          } else {
            keys.current.delete(
              JSON.stringify({
                kind: "review",
                cohortId: p.cohortId,
                competitionId: p.competitionId,
                reviewId: p.reviewId,
                decision: p.decision,
                expectedRevision: p.expectedRevision,
                appeal: p.appeal,
              }),
            );
          }
          await load();
          throw new Error(
            "Revision conflict: the bracket has changed. Data reloaded—review and retry.",
          );
        }
        throw new Error(
          b &&
            typeof b === "object" &&
            "message" in b &&
            typeof b.message === "string"
            ? b.message
            : "The tournament action failed.",
        );
      }
      if (p.kind === "create") {
        if (!validCohort(b) || b.capacity !== p.capacity)
          throw new Error("The cohort response was invalid.");
        keys.current.delete(
          JSON.stringify({ kind: "create", capacity: p.capacity }),
        );
        setPending(null);
        router.push(`/tournaments/${encodeURIComponent(b.id)}`);
        return;
      }
      if (p.kind === "start") {
        if (
          !b ||
          typeof b !== "object" ||
          !("cohort" in b) ||
          !("competition" in b) ||
          !validCohort(b.cohort) ||
          !validCompetition(b.competition) ||
          b.cohort.id !== p.cohortId ||
          b.cohort.status !== "started" ||
          b.cohort.revision <= p.expectedRevision ||
          b.cohort.competitionId !== b.competition.id
        )
          throw new Error("The bracket response was invalid.");
        setCohort(b.cohort);
        setNotice("Bracket started.");
        keys.current.delete(
          JSON.stringify({
            kind: "start",
            cohortId: p.cohortId,
            expectedRevision: p.expectedRevision,
          }),
        );
      } else {
        if (!validReviewResult(b, p))
          throw new Error("The review result was invalid.");
        setCompetition(b as Competition);
        setNotice("Human review decision retained.");
      }
      setPending(null);
      setMfa(false);
      setOtp("");
      setActionError("");
    } catch (e) {
      if (mounted.current && gen === actionGen.current)
        setActionError(
          e instanceof Error ? e.message : "The tournament action failed.",
        );
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
        setActionError(e instanceof Error ? e.message : "MFA failed.");
    } finally {
      if (mounted.current && gen === stepGen.current) setBusy(false);
    }
  }
  function create() {
    const terms = JSON.stringify({ kind: "create", capacity });
    setPending({ kind: "create", capacity, key: keyFor(terms) });
    setCreateOpen(false);
  }
  function resume(e: FormEvent) {
    e.preventDefault();
    if (reference.trim())
      router.push(`/tournaments/${encodeURIComponent(reference.trim())}`);
  }
  if (mode === "landing")
    return (
      <Box className="tournament-redesign">
        <Heading />
        {notice ? <Alert severity="success">{notice}</Alert> : null}
        <Box className="tournament-command-deck">
          <Box className="tournament-create-command">
            <AdminCardWatermark watermark="identity" />
            <Typography className="section-kicker">
              NEW INVITATION FIELD
            </Typography>
            <Typography component="h2">Open a private bracket.</Typography>
            <Typography>
              Set the seat count now. The cohort locks itself only when every
              invited place is claimed.
            </Typography>
            <Button variant="contained" onClick={() => setCreateOpen(true)}>
              Create private cohort
            </Button>
          </Box>
          <AdminCard
            component="form"
            onSubmit={resume}
            variant="form"
            watermark="evidence"
            className="tournament-resume-command"
          >
            <Box className="tournament-command-icon">
              <AdminIcon name="tournaments" aria-hidden="true" />
            </Box>
            <Typography className="section-kicker">EXACT LOOKUP</Typography>
            <Typography component="h2">Resume control</Typography>
            <Typography className="tournament-command-copy">
              Return directly to a known private cohort. No directory or public
              discovery is exposed.
            </Typography>
            <TextField
              label="Private cohort reference"
              value={reference}
              onChange={(e) => setReference(e.target.value)}
              required
            />
            <Button type="submit" disabled={!reference.trim()}>
              Open exact cohort
            </Button>
          </AdminCard>
        </Box>
        <Dialog
          open={createOpen}
          aria-labelledby="cohort-create-title"
          aria-describedby="cohort-create-description"
          onClose={() => {
            if (!busy) setCreateOpen(false);
          }}
        >
          <DialogTitle id="cohort-create-title">
            Create a private cohort
          </DialogTitle>
          <DialogContent>
            <DialogContentText id="cohort-create-description">
              Seats are opt-in. The cohort locks automatically when full;
              performance never changes member matching or trust.
            </DialogContentText>
            <TextField
              select
              label="Entrant capacity"
              value={capacity}
              onChange={(e) =>
                setCapacity(Number(e.target.value) as 4 | 8 | 16)
              }
              sx={{ mt: 1 }}
            >
              {[4, 8, 16].map((x) => (
                <MenuItem key={x} value={x}>
                  {x} entrants
                </MenuItem>
              ))}
            </TextField>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button disabled={busy} onClick={create}>
              Review creation
            </Button>
          </DialogActions>
        </Dialog>
        <Confirm
          pending={pending}
          busy={busy}
          error={actionError}
          close={() => {
            setPending(null);
            setActionError("");
            setOtp("");
          }}
          run={execute}
        />
      </Box>
    );
  return (
    <Box className="tournament-redesign">
      <Heading />
      {notice ? <Alert severity="success">{notice}</Alert> : null}
      {loading ? (
        <AdminCard variant="detail" watermark="identity" showWatermark={false}>
          <AdminSkeleton variant="form" />
        </AdminCard>
      ) : loadError ? (
        <AdminCard variant="warning" watermark="identity" showWatermark={false}>
          <EmptyState
            icon="!"
            title={
              loadStatus === 404
                ? "Tournament record not found"
                : "Tournament record unavailable"
            }
            description={loadError}
            variant="warning"
            action={
              loadStatus === 404 ? (
                <Button component={Link} href="/tournaments">
                  Return to tournaments
                </Button>
              ) : (
                <Button onClick={() => void load()}>Retry</Button>
              )
            }
          />
        </AdminCard>
      ) : mode === "cohort" && cohort ? (
        <CohortView
          cohort={cohort}
          busy={busy}
          review={() => {
            const terms = JSON.stringify({
              kind: "start",
              cohortId: cohort.id,
              expectedRevision: cohort.revision,
            });
            setPending({
              kind: "start",
              cohortId: cohort.id,
              expectedRevision: cohort.revision,
              key: keyFor(terms),
            });
          }}
        />
      ) : mode === "competition" && competition ? (
        <CompetitionView
          competition={competition}
          busy={busy}
          review={(id, decision, appeal) => {
            const terms = JSON.stringify({
              kind: "review",
              cohortId,
              competitionId: competition.id,
              reviewId: id,
              decision,
              expectedRevision: competition.revision,
              appeal,
            });
            setPending({
              kind: "review",
              cohortId: cohortId!,
              competitionId: competition.id,
              reviewId: id,
              decision,
              expectedRevision: competition.revision,
              appeal,
              key: keyFor(terms),
            });
          }}
        />
      ) : null}
      <Confirm
        pending={mfa ? null : pending}
        busy={busy}
        error={actionError}
        close={() => {
          setPending(null);
          setActionError("");
          setOtp("");
        }}
        run={execute}
      />
      <Dialog
        open={mfa}
        aria-labelledby="tournament-mfa-title"
        aria-describedby="tournament-mfa-description"
        onClose={() => {
          if (!busy) {
            setMfa(false);
            setPending(null);
            setOtp("");
            setActionError("");
          }
        }}
      >
        <DialogTitle id="tournament-mfa-title">Fresh MFA required</DialogTitle>
        <DialogContent>
          <DialogContentText id="tournament-mfa-description">
            Verify fresh authority to retry this exact review decision.
          </DialogContentText>
          {actionError ? (
            <Alert severity="error" role="alert">
              {actionError}
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
              setActionError("");
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
function Heading() {
  return (
    <Box component="header" className="tournament-hero">
      <AdminCardWatermark watermark="identity" />
      <Box className="tournament-hero-copy">
        <Box className="tournament-kicker">
          <AdminIcon name="tournaments" aria-hidden="true" />
          <Typography className="section-kicker">
            PRIVATE COMPETITION CONTROL
          </Typography>
        </Box>
        <Typography component="h1">Competition without the crowd.</Typography>
        <Typography className="tournament-hero-intro">
          Run invitation-only tournaments with exact cohort access, locked
          brackets, privacy-safe ladders, and accountable human review.
        </Typography>
      </Box>
      <Box
        className="tournament-hero-register"
        aria-label="Tournament privacy controls"
      >
        <div>
          <span>Discovery</span>
          <strong>None</strong>
          <Typography>Exact references only</Typography>
        </div>
        <div>
          <span>Matching effect</span>
          <strong>None</strong>
          <Typography>Performance stays isolated</Typography>
        </div>
        <div>
          <span>Disputes</span>
          <strong>Human reviewed</strong>
          <Typography>Fresh authority for decisions</Typography>
        </div>
      </Box>
    </Box>
  );
}
function CohortView({
  cohort,
  busy,
  review,
}: {
  cohort: Cohort;
  busy: boolean;
  review: () => void;
}) {
  return (
    <AdminCard
      component="article"
      variant="detail"
      watermark="identity"
      className="tournament-cohort-dossier"
    >
      <Box className="tournament-dossier-topline">
        <Typography className="section-kicker">COHORT DOSSIER</Typography>
        <Chip label={cohort.status} />
      </Box>
      <Box className="tournament-dossier-heading">
        <Box>
          <Typography component="h2">Private field</Typography>
          <Typography>{cohort.id}</Typography>
        </Box>
        <div>
          <span>Revision</span>
          <strong>{cohort.revision.toString().padStart(2, "0")}</strong>
        </div>
      </Box>
      <Box className="tournament-seat-control">
        <Box
          role="progressbar"
          aria-label="Seats claimed"
          aria-valuemin={0}
          aria-valuemax={cohort.capacity}
          aria-valuenow={cohort.enrolled}
          aria-valuetext={`${cohort.enrolled} of ${cohort.capacity} seats claimed`}
        >
          <Box className="tournament-seat-label">
            <span>SEAT CLAIM</span>
            <strong>
              {cohort.enrolled} / {cohort.capacity}
            </strong>
            <Typography>seats claimed</Typography>
          </Box>
          <LinearProgress
            value={(cohort.enrolled / cohort.capacity) * 100}
            variant="determinate"
          />
        </Box>
      </Box>
      <Box className="tournament-dossier-actions">
        {cohort.competitionId ? (
          <Button
            component={Link}
            href={`/tournaments/${encodeURIComponent(cohort.id)}/competitions/${encodeURIComponent(cohort.competitionId)}`}
          >
            Open competition
          </Button>
        ) : null}
        <Button
          disabled={
            busy ||
            cohort.status !== "locked" ||
            cohort.enrolled !== cohort.capacity
          }
          onClick={review}
        >
          {cohort.status === "locked"
            ? "Review bracket start"
            : cohort.status === "open"
              ? "Waiting for every seat"
              : "Bracket already started"}
        </Button>
      </Box>
    </AdminCard>
  );
}
function CompetitionView({
  competition,
  busy,
  review,
}: {
  competition: Competition;
  busy: boolean;
  review: (id: string, d: "no_action" | "rules_action", a: boolean) => void;
}) {
  return (
    <Stack className="tournament-competition" spacing={2}>
      <AdminCard
        variant="detail"
        watermark="evidence"
        className="tournament-competition-header"
      >
        <Box>
          <Typography className="section-kicker">ACTIVE BRACKET</Typography>
          <Typography component="h2">Competition {competition.id}</Typography>
          <Typography>
            {competition.status} · revision {competition.revision}
          </Typography>
        </Box>
        <Box className="tournament-competition-counts">
          <div>
            <strong>
              {competition.matches.length.toString().padStart(2, "0")}
            </strong>
            <span>matches</span>
          </div>
          <div>
            <strong>
              {competition.ladder.length.toString().padStart(2, "0")}
            </strong>
            <span>ladder entries</span>
          </div>
        </Box>
      </AdminCard>
      <Box className="tournament-competition-grid">
        <AdminCard
          variant="panel"
          watermark="analytics"
          className="tournament-ladder"
        >
          <Stack spacing={1}>
            <Typography component="h3">Privacy-safe ladder</Typography>
            {competition.ladder.map((entry, index) => (
              <Box
                className={entry.you ? "is-you" : ""}
                component="article"
                key={`${entry.label}-${index}`}
                sx={{ py: 1, borderBottom: 1, borderColor: "divider" }}
              >
                <Typography sx={{ fontWeight: 800 }}>
                  {entry.label}
                  {entry.you ? " · you" : ""}
                </Typography>
                <Typography color="text.secondary">
                  {entry.wins} wins · {entry.played} played
                </Typography>
              </Box>
            ))}
          </Stack>
        </AdminCard>
        <AdminCard
          variant="panel"
          watermark="clock"
          className="tournament-matches"
        >
          <Stack spacing={1}>
            <Typography component="h3">Bracket matches</Typography>
            {competition.matches.map((match) => (
              <Box
                component="article"
                key={match.id}
                sx={{ py: 1, borderBottom: 1, borderColor: "divider" }}
              >
                <Typography sx={{ fontWeight: 800 }}>
                  Round {match.round} · slot {match.slot}
                </Typography>
                <Typography>
                  {match.firstLabel} vs {match.secondLabel}
                </Typography>
                <Typography color="text.secondary">
                  {match.resultRecorded
                    ? `Winner: ${match.winnerLabel ?? "Recorded"}`
                    : "Result pending"}
                  {match.youArePlaying ? " · your match" : ""}
                </Typography>
              </Box>
            ))}
          </Stack>
        </AdminCard>
      </Box>
      {competition.reviews.length === 0 ? (
        <AdminCard
          variant="panel"
          watermark="evidence"
          showWatermark={false}
          className="tournament-review-empty"
        >
          <EmptyState
            icon="✓"
            title="No neutral reviews"
            description="No review requires an operator decision."
          />
        </AdminCard>
      ) : (
        competition.reviews.map((r) => (
          <AdminCard
            component="article"
            key={r.id}
            variant="row"
            watermark="evidence"
            className="tournament-review-record"
          >
            <Stack spacing={1}>
              <Typography component="h3">{r.matchId}</Typography>
              <Typography>
                {r.status} · {r.decision.replaceAll("_", " ")} · opened{" "}
                {new Date(r.openedAt).toLocaleString()}
              </Typography>
              {r.status === "open" || r.status === "appealed" ? (
                <Stack direction="column" spacing={1}>
                  <Button
                    disabled={busy}
                    onClick={() =>
                      review(r.id, "no_action", r.status === "appealed")
                    }
                  >
                    Review no action
                  </Button>
                  <Button
                    disabled={busy}
                    color="warning"
                    onClick={() =>
                      review(r.id, "rules_action", r.status === "appealed")
                    }
                  >
                    Review rules action
                  </Button>
                </Stack>
              ) : null}
            </Stack>
          </AdminCard>
        ))
      )}
    </Stack>
  );
}
function Confirm({
  pending,
  busy,
  error,
  close,
  run,
}: {
  pending: PendingTournament | null;
  busy: boolean;
  error: string;
  close: () => void;
  run: (p: PendingTournament) => Promise<void>;
}) {
  return (
    <Dialog
      open={Boolean(pending)}
      onClose={() => {
        if (!busy) close();
      }}
      aria-describedby="tournament-confirm"
      aria-labelledby="tournament-confirm-title"
    >
      <DialogTitle id="tournament-confirm-title">
        Confirm exact tournament command
      </DialogTitle>
      <DialogContent>
        <DialogContentText id="tournament-confirm">
          Review the immutable identifiers, revision, and decision.
        </DialogContentText>
        {error ? (
          <Alert severity="error" role="alert">
            {error}
          </Alert>
        ) : null}
        {pending ? (
          <Typography sx={{ overflowWrap: "anywhere" }}>
            {JSON.stringify(payloadForDisplay(pending))}
          </Typography>
        ) : null}
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          Cancel
        </Button>
        <Button
          disabled={busy || !pending}
          onClick={() => (pending ? void run(pending) : undefined)}
        >
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  );
}
function payloadForDisplay(p: PendingTournament) {
  if (p.kind === "create") return { kind: p.kind, capacity: p.capacity };
  if (p.kind === "start")
    return {
      kind: p.kind,
      cohortId: p.cohortId,
      expectedRevision: p.expectedRevision,
    };
  return {
    kind: p.kind,
    cohortId: p.cohortId,
    competitionId: p.competitionId,
    reviewId: p.reviewId,
    decision: p.decision,
    expectedRevision: p.expectedRevision,
    appeal: p.appeal,
  };
}
