"use client";

import {
  canPreviewSettlement,
  escrowReducer,
  formatGhs,
  initialEscrowState,
} from "@obiara/escrow-engagement";
import Link from "next/link";
import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../../compound-navigation";

export function EscrowEngagement() {
  const [state, dispatch] = useReducer(escrowReducer, initialEscrowState);
  const selected = state.milestones.find(
    (milestone) => milestone.id === state.selectedMilestone,
  )!;
  const frozen = state.disputeState !== "none";

  return (
    <main className="fie-shell escrow-shell">
      <CompoundRail contextLabel="Engagement finance" />
      <section className="fie-main escrow-main">
        <header className="escrow-topbar">
          <Link href="/fie/matchmakers">Back to matchmakers</Link>
          <span>{state.engagementRef} · terms locked</span>
        </header>
        <section className="escrow-hero">
          <p className="fie-kicker">Protected engagement</p>
          <h1>Money moves only after shared evidence.</h1>
          <p>
            Funding, fees and milestones are fixed to the agreed engagement.
            This preview never releases money or contacts a payment provider.
          </p>
        </section>

        <section className="escrow-ledger" aria-label="Engagement totals">
          <div><span>Funded</span><strong>{formatGhs(state.fundedPesewas)}</strong></div>
          <div><span>Platform fee</span><strong>{formatGhs(state.platformFeePesewas)}</strong></div>
          <div><span>Matchmaker payout</span><strong>{formatGhs(state.payoutPesewas)}</strong></div>
          <div><span>Payout statement</span><strong>{state.payoutStatementRef}</strong></div>
        </section>

        <section className="escrow-workspace">
          <article className="escrow-milestones">
            <p className="fie-kicker">Named milestones</p>
            <h2>Confirm what was delivered.</h2>
            <div className="milestone-tabs">
              {state.milestones.map((milestone) => (
                <button
                  aria-pressed={milestone.id === state.selectedMilestone}
                  key={milestone.id}
                  onClick={() => dispatch({ type: "select", id: milestone.id })}
                  type="button"
                >
                  <span>{milestone.name}</span>
                  <strong>{formatGhs(milestone.amountPesewas)}</strong>
                </button>
              ))}
            </div>
            <div className="evidence-card">
              <h3>{selected.name}</h3>
              <p>Both sides confirm delivery independently. Neither confirmation can release funds alone.</p>
              <div className="evidence-actions">
                <button
                  disabled={selected.memberConfirmed || frozen}
                  onClick={() => dispatch({ type: "confirm-member" })}
                  type="button"
                >
                  {selected.memberConfirmed ? "Member confirmed" : "Confirm as member"}
                </button>
                <button
                  disabled={selected.matchmakerConfirmed || frozen}
                  onClick={() => dispatch({ type: "confirm-matchmaker" })}
                  type="button"
                >
                  {selected.matchmakerConfirmed ? "Matchmaker confirmed" : "Preview matchmaker confirmation"}
                </button>
              </div>
              <button
                className="settlement-button"
                disabled={!canPreviewSettlement(state)}
                onClick={() => dispatch({ type: "preview-settlement" })}
                type="button"
              >
                Preview settlement
              </button>
              {state.settlementPreview ? (
                <p className="settlement-ready" role="status">
                  Settlement is eligible for backend review. No money has moved.
                </p>
              ) : null}
            </div>
          </article>

          <article className="escrow-dispute">
            <p className="fie-kicker">Dispute protection</p>
            <h2>{frozen ? "Settlement is frozen." : "Something does not match?"}</h2>
            {state.disputeState === "none" ? (
              <>
                <p>Explain the delivery issue without phone, card or private conversation details.</p>
                <textarea
                  aria-label="Dispute reason"
                  onChange={(event) => dispatch({ type: "dispute-reason", value: event.target.value })}
                  placeholder="Describe what differs from the agreed milestone"
                  rows={5}
                  value={state.disputeReason}
                />
                <button
                  disabled={state.disputeReason.trim().length < 12}
                  onClick={() => dispatch({ type: "open-dispute" })}
                  type="button"
                >
                  Open dispute and freeze settlement
                </button>
              </>
            ) : (
              <div className="dispute-status">
                <strong>Funds remain protected while people review the terms.</strong>
                <p>This does not delete the engagement, evidence or payout statement.</p>
                {state.disputeState === "open" ? (
                  <button onClick={() => dispatch({ type: "escalate-dispute" })} type="button">
                    Escalate to Mpanyimfo review
                  </button>
                ) : (
                  <span>{state.escalationRef} · awaiting a separate panel</span>
                )}
              </div>
            )}
          </article>
        </section>
      </section>
      <CompoundBottomNavigation contextLabel="Engagement finance" />
    </main>
  );
}
