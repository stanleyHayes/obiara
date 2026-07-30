"use client";

import Link from "next/link";
import { useCallback, useEffect, useReducer, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

type Milestone = {
  id: string;
  grossPesewas: number;
  feePesewas: number;
  deliveryConfirmed: boolean;
  acceptanceConfirmed: boolean;
  settled: boolean;
  statementRef?: string;
};

type Escrow = {
  escrowId: string;
  engagementId: string;
  fundedPesewas: number;
  settledPesewas: number;
  termsId: string;
  termsVersion: number;
  milestones: Milestone[];
  disputed: boolean;
  escalationRef?: string;
  revision: number;
};

function formatGhs(pesewas: number) {
  return new Intl.NumberFormat("en-GH", {
    style: "currency",
    currency: "GHS",
  }).format(pesewas / 100);
}

export function EscrowEngagement() {
  const [escrows, setEscrows] = useReducer(
    (_: Escrow[], next: Escrow[]) => next,
    [],
  );
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  const load = useCallback(async () => {
    const response = await fetch("/api/escrows", { cache: "no-store" });
    const payload = (await response.json().catch(() => null)) as {
      items?: Escrow[];
      message?: string;
    } | null;
    if (!response.ok || !payload?.items) {
      setError(
        payload?.message ?? "Protected engagements could not be loaded.",
      );
      return;
    }
    setEscrows(payload.items);
    setSelectedId((current) =>
      current && payload.items?.some((item) => item.escrowId === current)
        ? current
        : (payload.items?.[0]?.escrowId ?? null),
    );
    setError("");
  }, []);

  useEffect(() => {
    void Promise.resolve().then(load);
  }, [load]);

  const selected = escrows.find((escrow) => escrow.escrowId === selectedId);

  async function mutate(body: object, success: string) {
    setBusy(true);
    setError("");
    const response = await fetch("/api/escrows", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": `escrow.${crypto.randomUUID()}`,
      },
      body: JSON.stringify(body),
    });
    const payload = (await response.json().catch(() => null)) as {
      message?: string;
    } | null;
    if (!response.ok)
      setError(payload?.message ?? "The escrow action could not be completed.");
    else {
      setMessage(success);
      await load();
    }
    setBusy(false);
  }

  return (
    <main className="fie-shell escrow-shell">
      <CompoundRail contextLabel="Engagement finance" />
      <section className="fie-main escrow-main">
        <header className="escrow-topbar">
          <Link href="/fie/matchmakers">Back to matchmakers</Link>
          <span>Member-owned · provider-funded · evidence-bound</span>
        </header>
        <section className="escrow-hero">
          <p className="fie-kicker">Protected engagement</p>
          <h1>Money moves only after shared evidence.</h1>
          <p>
            Delivery belongs to the matchmaker, acceptance belongs to you, and
            settlement belongs to finance. No client can impersonate another
            authority.
          </p>
        </section>
        {error ? (
          <p className="booking-note" role="alert">
            {error}
          </p>
        ) : null}
        {message ? (
          <p className="settlement-ready" role="status">
            {message}
          </p>
        ) : null}
        {escrows.length > 1 ? (
          <div className="milestone-tabs">
            {escrows.map((escrow) => (
              <button
                aria-pressed={escrow.escrowId === selectedId}
                key={escrow.escrowId}
                onClick={() => setSelectedId(escrow.escrowId)}
                type="button"
              >
                <span>{escrow.termsId}</span>
                <strong>{formatGhs(escrow.fundedPesewas)}</strong>
              </button>
            ))}
          </div>
        ) : null}
        {!selected ? (
          <section className="escrow-workspace">
            <article className="escrow-milestones">
              <p className="fie-kicker">No funded escrow</p>
              <h2>
                Your booked terms remain separate until a payment provider
                confirms funding.
              </h2>
              <p>
                No local preview can create money, evidence, settlement or a
                dispute.
              </p>
            </article>
          </section>
        ) : (
          <>
            <section className="escrow-ledger" aria-label="Engagement totals">
              <div>
                <span>Funded</span>
                <strong>{formatGhs(selected.fundedPesewas)}</strong>
              </div>
              <div>
                <span>Settled</span>
                <strong>{formatGhs(selected.settledPesewas)}</strong>
              </div>
              <div>
                <span>Terms</span>
                <strong>
                  {selected.termsId} · v{selected.termsVersion}
                </strong>
              </div>
              <div>
                <span>Status</span>
                <strong>
                  {selected.disputed ? "Frozen for review" : "Protected"}
                </strong>
              </div>
            </section>
            <section className="escrow-workspace">
              <article className="escrow-milestones">
                <p className="fie-kicker">Named milestones</p>
                <h2>Confirm only what you received.</h2>
                {selected.milestones.map((milestone) => (
                  <div className="evidence-card" key={milestone.id}>
                    <h3>{milestone.id}</h3>
                    <p>
                      {formatGhs(milestone.grossPesewas)} gross ·{" "}
                      {formatGhs(milestone.feePesewas)} disclosed fee
                    </p>
                    <div className="evidence-actions">
                      <span>
                        {milestone.deliveryConfirmed
                          ? "Matchmaker delivery recorded"
                          : "Waiting for matchmaker delivery"}
                      </span>
                      <button
                        disabled={
                          busy ||
                          selected.disputed ||
                          milestone.acceptanceConfirmed ||
                          !milestone.deliveryConfirmed
                        }
                        onClick={() =>
                          void mutate(
                            {
                              action: "accept",
                              escrowId: selected.escrowId,
                              milestoneId: milestone.id,
                            },
                            "Your acceptance evidence was recorded. It does not settle funds by itself.",
                          )
                        }
                        type="button"
                      >
                        {milestone.acceptanceConfirmed
                          ? "You accepted"
                          : "Confirm delivery received"}
                      </button>
                    </div>
                    {milestone.settled ? (
                      <p className="settlement-ready">
                        Settled · {milestone.statementRef}
                      </p>
                    ) : null}
                  </div>
                ))}
              </article>
              <article className="escrow-dispute">
                <p className="fie-kicker">Dispute protection</p>
                <h2>
                  {selected.disputed
                    ? "Settlement is frozen."
                    : "Something does not match?"}
                </h2>
                {selected.disputed ? (
                  <div className="dispute-status">
                    <strong>
                      Funds remain protected while people review the terms.
                    </strong>
                    <p>
                      {selected.escalationRef} · awaiting a separate Mpanyimfo
                      panel
                    </p>
                  </div>
                ) : (
                  <>
                    <p>
                      Opening a dispute permanently freezes new evidence and
                      settlement while preserving the complete record.
                    </p>
                    <button
                      disabled={busy}
                      onClick={() =>
                        void mutate(
                          { action: "dispute", escrowId: selected.escrowId },
                          "The escrow is frozen and a separate review reference was created.",
                        )
                      }
                      type="button"
                    >
                      Open dispute and freeze settlement
                    </button>
                  </>
                )}
              </article>
            </section>
          </>
        )}
      </section>
      <CompoundBottomNavigation contextLabel="Engagement finance" />
    </main>
  );
}
