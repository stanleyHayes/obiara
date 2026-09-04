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
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SegmentedOtpInput } from "@obiara/ui-web";
import { AdminCard, AdminCardWatermark } from "../../admin-card";
import { AdminIcon, UtilityIcon } from "../../admin-icons";
import {
  settlementTermsKey,
  validFinanceOverview,
  validSettlementFor,
  type FinanceExceptionCode,
  type PendingSettlement,
} from "../../commercial-model";
import { AdminSkeleton } from "../../loading-skeleton";
import { EmptyState } from "../../empty-state";
import { adminFetch } from "../../lib/admin-fetch";

type ReconciliationException = {
  factRef: string;
  providerRef: string;
  statementRef: string;
  currency: "GHS" | "USD";
  minor: number;
  exception: FinanceExceptionCode;
  occurredAt: string;
  recordedAt: string;
};
type Checkpoint = {
  day: string;
  total: number;
  reconciled: number;
  excepted: number;
  completedAt: string;
};

const exceptionLabels: Record<string, string> = {
  ledger_missing: "Ledger entry missing",
  reference_mismatch: "Reference mismatch",
  currency_mismatch: "Currency mismatch",
  amount_mismatch: "Amount mismatch",
  ledger_unbalanced: "Ledger is unbalanced",
};

export function FinanceDesk() {
  const [exceptions, setExceptions] = useState<ReconciliationException[]>([]);
  const [checkpoints, setCheckpoints] = useState<Checkpoint[]>([]);
  const [exceptionCodes, setExceptionCodes] = useState<FinanceExceptionCode[]>(
    [],
  );
  const [loading, setLoading] = useState(true);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [settlementEscrow, setSettlementEscrow] = useState("");
  const [settlementMilestone, setSettlementMilestone] = useState("");
  const [settlementBusy, setSettlementBusy] = useState(false);
  const [settlementError, setSettlementError] = useState<string | null>(null);
  const [settlementNotice, setSettlementNotice] = useState("");
  const [pending, setPending] = useState<PendingSettlement | null>(null),
    [mfaOpen, setMfaOpen] = useState(false),
    [otp, setOtp] = useState("");
  const mounted = useRef(false),
    loadGeneration = useRef(0),
    actionGeneration = useRef(0),
    stepUpGeneration = useRef(0),
    controllerRef = useRef<AbortController | null>(null),
    settlementKey = useRef(`settle.${crypto.randomUUID()}`),
    settlementTerms = useRef("");
  const [statement, setStatement] = useState<{
    statementRef: string;
    grossPesewas: number;
    feePesewas: number;
    netPesewas: number;
    settledAt: string;
    escrow: { escrowId: string };
  } | null>(null);

  const load = useCallback(async () => {
    const generation = ++loadGeneration.current;
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    setLoading(true);
    setLoaded(false);
    setError(null);
    try {
      const response = await adminFetch("/api/finance", {
        cache: "no-store",
        signal: controller.signal,
      });
      const body = (await response.json().catch(() => null)) as {
        exceptions?: ReconciliationException[];
        checkpoints?: Checkpoint[];
        exceptionCodes?: FinanceExceptionCode[];
        message?: string;
      } | null;
      if (!response.ok || !validFinanceOverview(body))
        throw new Error(
          body?.message ?? "Reconciliation evidence could not be loaded.",
        );
      if (!mounted.current || generation !== loadGeneration.current) return;
      setExceptions(body.exceptions);
      setCheckpoints(body.checkpoints);
      setExceptionCodes(body.exceptionCodes);
      setLoaded(true);
    } catch (cause) {
      if (
        controller.signal.aborted ||
        !mounted.current ||
        generation !== loadGeneration.current
      )
        return;
      setError(
        cause instanceof Error
          ? cause.message
          : "Reconciliation evidence could not be loaded.",
      );
    } finally {
      if (mounted.current && generation === loadGeneration.current)
        setLoading(false);
    }
  }, []);

  useEffect(() => {
    mounted.current = true;
    const initialLoad = window.setTimeout(() => void load(), 0);
    const loads = loadGeneration,
      actions = actionGeneration,
      stepUps = stepUpGeneration;
    return () => {
      window.clearTimeout(initialLoad);
      mounted.current = false;
      loads.current++;
      actions.current++;
      stepUps.current++;
      controllerRef.current?.abort();
    };
  }, [load]);

  const totals = useMemo(
    () =>
      checkpoints.reduce(
        (sum, item) => ({
          total: sum.total + item.total,
          reconciled: sum.reconciled + item.reconciled,
          excepted: sum.excepted + item.excepted,
        }),
        { total: 0, reconciled: 0, excepted: 0 },
      ),
    [checkpoints],
  );

  // Sole owner of the settlement idempotency key. Minting it anywhere else
  // — a keystroke handler, for instance — breaks the binding between one key
  // and one set of terms, so a retry after an ambiguous timeout would reach
  // the API as a fresh command instead of a replay and could settle twice.
  function settlementSnapshot(): PendingSettlement {
    const terms = settlementTermsKey(settlementEscrow, settlementMilestone);
    if (settlementTerms.current !== terms) {
      settlementTerms.current = terms;
      settlementKey.current = `settle.${crypto.randomUUID()}`;
    }
    return {
      escrowId: settlementEscrow.trim(),
      milestoneId: settlementMilestone.trim(),
      key: settlementKey.current,
    };
  }

  async function settleMilestone(snapshot: PendingSettlement) {
    const generation = ++actionGeneration.current;
    setSettlementBusy(true);
    setSettlementError(null);
    setStatement(null);
    try {
      const response = await adminFetch("/api/escrows", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": snapshot.key,
        },
        body: JSON.stringify({
          action: "settle",
          escrowId: snapshot.escrowId,
          milestoneId: snapshot.milestoneId,
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        statementRef?: string;
        grossPesewas?: number;
        feePesewas?: number;
        netPesewas?: number;
        settledAt?: string;
        escrow?: { escrowId?: string };
        message?: string;
      } | null;
      const responseMessage = payload?.message;
      if (!mounted.current || generation !== actionGeneration.current) return;
      if (needsStepUp(response.status, errorCode(payload))) {
        setMfaOpen(true);
        return;
      }
      if (
        !response.ok ||
        !validSettlementFor(payload, snapshot) ||
        payload.feePesewas + payload.netPesewas !== payload.grossPesewas
      )
        throw new Error(
          responseMessage ?? "The milestone could not be settled.",
        );
      setStatement({
        statementRef: payload.statementRef,
        grossPesewas: payload.grossPesewas!,
        feePesewas: payload.feePesewas!,
        netPesewas: payload.netPesewas!,
        settledAt: payload.settledAt,
        escrow: payload.escrow,
      });
      settlementKey.current = `settle.${crypto.randomUUID()}`;
      settlementTerms.current = "";
      setPending(null);
      setMfaOpen(false);
      setOtp("");
    } catch (cause) {
      if (!mounted.current || generation !== actionGeneration.current) return;
      setSettlementError(
        cause instanceof Error
          ? cause.message
          : "The milestone could not be settled.",
      );
    } finally {
      if (mounted.current && generation === actionGeneration.current)
        setSettlementBusy(false);
    }
  }
  async function stepUp(action: "start" | "complete") {
    const generation = ++stepUpGeneration.current;
    let retry = false;
    setSettlementBusy(true);
    setSettlementError(null);
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
      // When start succeeds, show feedback that code was sent
      if (action === "start") {
        setSettlementNotice("A fresh code was sent to your admin email.");
      } else if (action === "complete" && pending) {
        const snapshot = pending;
        retry = true;
        setMfaOpen(false);
        setOtp("");
        setSettlementNotice("");
        await settleMilestone(snapshot);
      }
    } catch (error) {
      if (mounted.current && generation === stepUpGeneration.current)
        setSettlementError(
          error instanceof Error ? error.message : "The MFA step-up failed.",
        );
    } finally {
      if (!retry && mounted.current && generation === stepUpGeneration.current)
        setSettlementBusy(false);
    }
  }

  return (
    <main className="verification-shell finance-shell finance-redesign">
      <header className="verification-header finance-hero">
        <Box className="finance-hero-copy">
          <Link href="/" className="verification-back finance-back">
            Return to command centre
          </Link>
          <Box className="finance-kicker">
            <AdminIcon name="finance" aria-hidden="true" />
            <Typography className="section-kicker">
              Finance operations · live
            </Typography>
          </Box>
          <Typography component="h1">
            Money moves only when evidence agrees.
          </Typography>
          <Typography>
            Provider facts and ledger proofs are append-only. This desk exposes
            bounded reconciliation outcomes and separately authorizes
            evidence-complete escrow settlement.
          </Typography>
        </Box>
        <Box className="finance-hero-register" aria-label="Finance desk status">
          <Box>
            <span>Exceptions</span>
            <strong>{loaded && !error ? exceptions.length : "—"}</strong>
          </Box>
          <Box>
            <span>Ledger policy</span>
            <strong>Append only</strong>
          </Box>
          <Box>
            <span>Balance editing</span>
            <strong>Unavailable</strong>
          </Box>
        </Box>
        <AdminCardWatermark watermark="analytics" />
      </header>

      <section className="finance-boundary" aria-label="Settlement boundary">
        <span className="finance-boundary-icon">
          <UtilityIcon name="security" aria-hidden="true" />
        </span>
        <Box>
          <Typography className="section-kicker">Release boundary</Typography>
          <Typography component="h2">
            Provider fact, ledger proof and human authority must agree.
          </Typography>
          <Typography>
            Reconciliation can expose a bounded exception. Settlement requires
            complete delivery and member-acceptance evidence, no open dispute
            and fresh finance MFA.
          </Typography>
        </Box>
        <span className="finance-boundary-state">Fail closed</span>
      </section>

      {error ? (
        <AdminCard
          variant="warning"
          watermark="analytics"
          showWatermark={false}
          sx={{ mb: 3 }}
        >
          <EmptyState
            icon="!"
            title="Finance evidence unavailable"
            description={error}
            variant="warning"
            action={<Button onClick={() => void load()}>Retry</Button>}
          />
        </AdminCard>
      ) : null}
      {!error ? (
        <>
          <Box
            className="finance-pulse"
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: "1fr",
              mb: 3,
            }}
          >
            {[
              ["Facts checked", totals.total],
              ["Reconciled", totals.reconciled],
              ["Exceptions", totals.excepted],
            ].map(([label, value]) => (
              <AdminCard
                className={`finance-pulse-card finance-pulse-card--${String(label).toLowerCase().replace(" ", "-")}`}
                key={label}
                variant="metric"
                watermark="analytics"
                showWatermark={loaded}
                sx={{ p: 2.25 }}
              >
                <Typography
                  sx={{
                    color: "text.secondary",
                    fontSize: 12,
                    fontWeight: 800,
                  }}
                >
                  {label}
                </Typography>
                {loading ? (
                  <AdminSkeleton variant="metric" />
                ) : (
                  <Typography sx={{ fontSize: 30, fontWeight: 800 }}>
                    {loaded && !error ? value : "Unavailable"}
                  </Typography>
                )}
              </AdminCard>
            ))}
          </Box>
        </>
      ) : null}

      <AdminCard
        className="finance-settlement"
        variant="form"
        watermark="evidence"
        sx={{ mb: 3, p: 3 }}
      >
        <span className="finance-settlement-icon">
          <AdminIcon name="finance" aria-hidden="true" />
        </span>
        <Typography className="section-kicker">
          Evidence-gated escrow settlement
        </Typography>
        <Typography component="h2" sx={{ fontSize: 28, fontWeight: 800 }}>
          Commit release and journal together.
        </Typography>
        <Typography sx={{ color: "text.secondary", mt: 1, maxWidth: 780 }}>
          Finance scope and fresh MFA are required. Settlement fails closed
          unless distinct delivery and member-acceptance evidence exist and no
          dispute is open.
        </Typography>
        {settlementNotice ? (
          <Alert severity="success" sx={{ mt: 2 }}>
            {settlementNotice}
          </Alert>
        ) : null}
        {settlementError ? (
          <Alert severity="warning" sx={{ mt: 2 }}>
            {settlementError}
          </Alert>
        ) : null}
        {statement ? (
          <Alert severity="success" sx={{ mt: 2 }}>
            Escrow {statement.escrow.escrowId} · statement{" "}
            {statement.statementRef}
            {" · "}gross GHS {(statement.grossPesewas / 100).toFixed(2)}
            {" · "}fee GHS {(statement.feePesewas / 100).toFixed(2)}
            {" · "}net GHS {(statement.netPesewas / 100).toFixed(2)}
            {" · settled "}
            {new Date(statement.settledAt).toLocaleString()}
          </Alert>
        ) : null}
        <Stack
          className="finance-settlement-form"
          direction="column"
          spacing={2}
          sx={{ mt: 3 }}
        >
          <TextField
            fullWidth
            label="Escrow ID"
            value={settlementEscrow}
            onChange={(event) => setSettlementEscrow(event.target.value)}
          />
          <TextField
            fullWidth
            label="Evidence-complete milestone ID"
            value={settlementMilestone}
            onChange={(event) => setSettlementMilestone(event.target.value)}
          />
          <Button
            disabled={
              settlementBusy ||
              !settlementEscrow.trim() ||
              !settlementMilestone.trim()
            }
            onClick={() => {
              setSettlementError(null);
              setPending(settlementSnapshot());
            }}
            variant="contained"
          >
            Settle atomically
          </Button>
        </Stack>
      </AdminCard>

      {!error ? (
        <>
          <Box
            className="finance-grid"
            sx={{
              display: "grid",
              gridTemplateColumns: "1fr !important",
              gap: 2,
            }}
          >
            <AdminCard
              variant="panel"
              watermark="evidence"
              showWatermark={!loading && loaded && exceptions.length > 0}
              className="finance-exceptions"
            >
              <Box className="verification-panel-heading">
                <Box>
                  <Typography className="section-kicker">
                    Reconciliation exceptions
                  </Typography>
                  <Typography component="h2">
                    Provider fact vs ledger proof
                  </Typography>
                </Box>
                <Button
                  startIcon={<UtilityIcon name="replay" aria-hidden="true" />}
                  onClick={() => void load()}
                  variant="outlined"
                >
                  Refresh
                </Button>
              </Box>
              {loading ? (
                <AdminSkeleton variant="card-list" rows={3} />
              ) : error ? null : loaded && exceptions.length === 0 ? (
                <EmptyState
                  icon="✓"
                  title="No reconciliation exceptions"
                  description="No retained reconciliation exceptions are present."
                />
              ) : (
                <Stack spacing={1}>
                  {exceptions.map((item) => (
                    <Box
                      component="article"
                      className="finance-exception-record"
                      key={`${item.factRef}-${item.recordedAt}`}
                      sx={{ p: 2, borderBottom: 1, borderColor: "divider" }}
                    >
                      <Stack
                        direction={{ xs: "column", sm: "row" }}
                        spacing={1}
                        sx={{ justifyContent: "space-between" }}
                      >
                        <Box>
                          <Typography
                            sx={{ fontFamily: "Geist Mono", fontWeight: 800 }}
                          >
                            {item.factRef}
                          </Typography>
                          <Typography
                            sx={{ color: "text.secondary", fontSize: 12 }}
                          >
                            {item.providerRef} · {item.statementRef}
                          </Typography>
                        </Box>
                        <Box sx={{ textAlign: { sm: "right" } }}>
                          <Typography sx={{ fontWeight: 800 }}>
                            {item.currency} {(item.minor / 100).toFixed(2)}
                          </Typography>
                          <Chip
                            color="warning"
                            label={
                              exceptionLabels[item.exception] ?? item.exception
                            }
                            size="small"
                          />
                        </Box>
                      </Stack>
                    </Box>
                  ))}
                </Stack>
              )}
            </AdminCard>

            <AdminCard
              variant="timeline"
              watermark="clock"
              showWatermark={!loading && loaded && checkpoints.length > 0}
              className="finance-resolution"
            >
              <Typography className="section-kicker">
                Daily checkpoints
              </Typography>
              <Typography component="h2">Immutable run evidence</Typography>
              <Typography sx={{ color: "text.secondary", mb: 2 }}>
                Checkpoints summarize completed comparisons. They cannot resolve
                exceptions or modify either source.
              </Typography>
              <Stack spacing={1}>
                {loading ? (
                  <AdminSkeleton variant="card-list" rows={2} />
                ) : error ? null : loaded && checkpoints.length === 0 ? (
                  <EmptyState
                    icon="✓"
                    title="No completed checkpoints"
                    description="No daily reconciliation run has completed yet."
                  />
                ) : (
                  checkpoints.map((item) => (
                    <Box
                      component="article"
                      className="finance-checkpoint-record"
                      key={item.day}
                      sx={{ p: 1.75, borderBottom: 1, borderColor: "divider" }}
                    >
                      <Typography sx={{ fontWeight: 800 }}>
                        {item.day}
                      </Typography>
                      <Typography
                        sx={{ color: "text.secondary", fontSize: 13 }}
                      >
                        {item.reconciled} reconciled · {item.excepted}{" "}
                        exceptions · {item.total} total
                      </Typography>
                    </Box>
                  ))
                )}
              </Stack>
            </AdminCard>
          </Box>

          <Alert className="finance-limit-note" severity="info" sx={{ mt: 3 }}>
            Pricing publication and redacted export preparation are unavailable
            here because no server-authoritative approval or export service is
            composed. The previous local controls were removed; this desk will
            not claim a financial action that was never persisted.
          </Alert>
          {exceptionCodes.length ? (
            <Typography sx={{ mt: 1.5, color: "text.secondary", fontSize: 12 }}>
              Recognized exception codes: {exceptionCodes.join(", ")}
            </Typography>
          ) : null}
        </>
      ) : null}
      <Dialog
        aria-labelledby="finance-confirm-title"
        aria-describedby="finance-confirm-description"
        open={Boolean(pending) && !mfaOpen}
        onClose={() => {
          if (!settlementBusy) {
            setPending(null);
            setSettlementError(null);
          }
        }}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle id="finance-confirm-title">
          Confirm atomic settlement
        </DialogTitle>
        <DialogContent>
          <Stack spacing={1.5}>
            <DialogContentText id="finance-confirm-description">
              Review the immutable escrow and milestone identifiers before MFA.
            </DialogContentText>
            {pending ? (
              <>
                <Typography>
                  <strong>Escrow:</strong> {pending.escrowId}
                </Typography>
                <Typography>
                  <strong>Evidence-complete milestone:</strong>{" "}
                  {pending.milestoneId}
                </Typography>
              </>
            ) : null}
            <Alert severity="warning">
              Finance scope and fresh MFA are required. Delivery and
              member-acceptance evidence must remain complete and no dispute may
              be open.
            </Alert>
            {settlementError ? (
              <Alert severity="error" role="alert" aria-live="assertive">
                {settlementError}
              </Alert>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={settlementBusy}
            onClick={() => {
              setPending(null);
              setSettlementError(null);
            }}
          >
            Cancel
          </Button>
          <Button
            aria-busy={settlementBusy}
            disabled={settlementBusy || !pending}
            onClick={() =>
              pending ? void settleMilestone(pending) : undefined
            }
            variant="contained"
          >
            Confirm settlement
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        aria-labelledby="finance-mfa-title"
        aria-describedby="finance-mfa-description"
        open={mfaOpen}
        onClose={() => {
          if (!settlementBusy) {
            setMfaOpen(false);
            setPending(null);
            setOtp("");
            setSettlementError(null);
          }
        }}
      >
        <DialogTitle id="finance-mfa-title">Fresh MFA required</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            <DialogContentText id="finance-mfa-description">
              Enter the six-digit code to retry this exact settlement request.
            </DialogContentText>
            {settlementError ? (
              <Alert severity="error" role="alert" aria-live="assertive">
                {settlementError}
              </Alert>
            ) : null}
            <SegmentedOtpInput
              label="Six-digit code"
              value={otp}
              onChange={setOtp}
              disabled={settlementBusy}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={settlementBusy}
            onClick={() => {
              setMfaOpen(false);
              setPending(null);
              setOtp("");
              setSettlementError(null);
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={settlementBusy}
            onClick={() => void stepUp("start")}
          >
            Send code
          </Button>
          <Button
            disabled={settlementBusy || otp.length !== 6}
            onClick={() => void stepUp("complete")}
          >
            Verify and retry
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
