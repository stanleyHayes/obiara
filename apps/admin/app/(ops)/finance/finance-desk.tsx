"use client";

import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";

type ReconciliationException = {
  factRef: string;
  providerRef: string;
  statementRef: string;
  currency: "GHS" | "USD";
  minor: number;
  exception: string;
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [settlementEscrow, setSettlementEscrow] = useState("");
  const [settlementMilestone, setSettlementMilestone] = useState("");
  const [settlementBusy, setSettlementBusy] = useState(false);
  const [settlementError, setSettlementError] = useState<string | null>(null);
  const [statement, setStatement] = useState<{
    statementRef: string;
    grossPesewas: number;
    feePesewas: number;
    netPesewas: number;
  } | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/finance", { cache: "no-store" });
      const body = (await response.json().catch(() => null)) as {
        exceptions?: ReconciliationException[];
        checkpoints?: Checkpoint[];
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "Reconciliation evidence could not be loaded.",
        );
      setExceptions(body?.exceptions ?? []);
      setCheckpoints(body?.checkpoints ?? []);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Reconciliation evidence could not be loaded.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(initialLoad);
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

  async function settleMilestone() {
    setSettlementBusy(true);
    setSettlementError(null);
    setStatement(null);
    try {
      const response = await fetch("/api/escrows", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `settle.${crypto.randomUUID()}`,
        },
        body: JSON.stringify({
          action: "settle",
          escrowId: settlementEscrow.trim(),
          milestoneId: settlementMilestone.trim(),
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        statementRef?: string;
        grossPesewas?: number;
        feePesewas?: number;
        netPesewas?: number;
        message?: string;
      } | null;
      if (!response.ok || !payload?.statementRef)
        throw new Error(
          payload?.message ?? "The milestone could not be settled.",
        );
      setStatement({
        statementRef: payload.statementRef,
        grossPesewas: payload.grossPesewas ?? 0,
        feePesewas: payload.feePesewas ?? 0,
        netPesewas: payload.netPesewas ?? 0,
      });
    } catch (cause) {
      setSettlementError(
        cause instanceof Error
          ? cause.message
          : "The milestone could not be settled.",
      );
    } finally {
      setSettlementBusy(false);
    }
  }

  return (
    <main className="verification-shell finance-shell">
      <header className="verification-header">
        <Box>
          <Link href="/" className="verification-back">
            Return to command centre
          </Link>
          <Typography className="section-kicker">
            Finance operations · live
          </Typography>
          <Typography component="h1">
            Compare evidence. Preserve the books.
          </Typography>
          <Typography>
            Provider facts and ledger proofs are append-only. This desk exposes
            bounded reconciliation outcomes and separately authorizes
            evidence-complete escrow settlement.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Chip
            color={exceptions.length ? "warning" : "success"}
            label={`${exceptions.length} exceptions`}
          />
          <Chip label="No balance editing" variant="outlined" />
        </Stack>
      </header>

      {error ? (
        <Alert severity="error" className="verification-alert">
          {error}
        </Alert>
      ) : null}
      <Box
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: "repeat(3,minmax(0,1fr))",
          mb: 3,
        }}
      >
        {[
          ["Facts checked", totals.total],
          ["Reconciled", totals.reconciled],
          ["Exceptions", totals.excepted],
        ].map(([label, value]) => (
          <Card key={label} variant="outlined" sx={{ p: 2.25 }}>
            <Typography
              sx={{ color: "text.secondary", fontSize: 12, fontWeight: 800 }}
            >
              {label}
            </Typography>
            <Typography sx={{ fontSize: 30, fontWeight: 800 }}>
              {value}
            </Typography>
          </Card>
        ))}
      </Box>

      <Card sx={{ mb: 3, p: 3 }}>
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
        {settlementError ? (
          <Alert severity="warning" sx={{ mt: 2 }}>
            {settlementError}
          </Alert>
        ) : null}
        {statement ? (
          <Alert severity="success" sx={{ mt: 2 }}>
            Statement {statement.statementRef} · gross GHS{" "}
            {(statement.grossPesewas / 100).toFixed(2)}
            {" · "}fee GHS {(statement.feePesewas / 100).toFixed(2)}
            {" · "}net GHS {(statement.netPesewas / 100).toFixed(2)}
          </Alert>
        ) : null}
        <Stack
          direction={{ xs: "column", md: "row" }}
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
            onClick={() => void settleMilestone()}
            variant="contained"
          >
            {settlementBusy ? "Settling…" : "Settle atomically"}
          </Button>
        </Stack>
      </Card>

      <Box className="finance-grid">
        <Card className="finance-exceptions">
          <Box className="verification-panel-heading">
            <Box>
              <Typography className="section-kicker">
                Reconciliation exceptions
              </Typography>
              <Typography component="h2">
                Provider fact vs ledger proof
              </Typography>
            </Box>
            <Button onClick={() => void load()} variant="outlined">
              Refresh
            </Button>
          </Box>
          {loading ? (
            <Stack sx={{ alignItems: "center", py: 8 }}>
              <CircularProgress size={28} />
            </Stack>
          ) : exceptions.length === 0 ? (
            <Alert severity="success">
              No retained reconciliation exceptions are present.
            </Alert>
          ) : (
            <Stack spacing={1}>
              {exceptions.map((item) => (
                <Card
                  key={`${item.factRef}-${item.recordedAt}`}
                  variant="outlined"
                  sx={{ p: 2 }}
                >
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={1}
                    sx={{ justifyContent: "space-between" }}
                  >
                    <Box>
                      <Typography
                        sx={{ fontFamily: "monospace", fontWeight: 800 }}
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
                </Card>
              ))}
            </Stack>
          )}
        </Card>

        <Card className="finance-resolution">
          <Typography className="section-kicker">Daily checkpoints</Typography>
          <Typography component="h2">Immutable run evidence</Typography>
          <Typography sx={{ color: "text.secondary", mb: 2 }}>
            Checkpoints summarize completed comparisons. They cannot resolve
            exceptions or modify either source.
          </Typography>
          <Stack spacing={1}>
            {checkpoints.length === 0 ? (
              <Alert severity="info">
                No daily reconciliation run has completed yet.
              </Alert>
            ) : (
              checkpoints.map((item) => (
                <Card key={item.day} variant="outlined" sx={{ p: 1.75 }}>
                  <Typography sx={{ fontWeight: 800 }}>{item.day}</Typography>
                  <Typography sx={{ color: "text.secondary", fontSize: 13 }}>
                    {item.reconciled} reconciled · {item.excepted} exceptions ·{" "}
                    {item.total} total
                  </Typography>
                </Card>
              ))
            )}
          </Stack>
        </Card>
      </Box>

      <Alert severity="info" sx={{ mt: 3 }}>
        Pricing publication and redacted export preparation are unavailable here
        because no server-authoritative approval or export service is composed.
        The previous local controls were removed; this desk will not claim a
        financial action that was never persisted.
      </Alert>
    </main>
  );
}
