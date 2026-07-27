"use client";

import {
  initialMembershipState,
  membershipReducer,
} from "@obiara/membership-settings";
import Link from "next/link";
import { useReducer } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../../compound-navigation";

export function MembershipSettings() {
  const [state, dispatch] = useReducer(
    membershipReducer,
    initialMembershipState,
  );

  return (
    <main className="fie-shell membership-shell">
      <CompoundRail contextLabel="Membership" />
      <section className="fie-main membership-main">
        <header className="membership-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Terms stay visible · no silent renewal</span>
        </header>

        <section className="membership-hero">
          <p className="fie-kicker">Your membership</p>
          <h1>Know exactly what is paid for.</h1>
          <p>
            Membership supports guidance and tools. It never buys seeds,
            visibility, rank, a match or access to another person.
          </p>
        </section>

        <section className="membership-grid">
          <article className="membership-pass">
            <header>
              <div>
                <p className="fie-kicker">Current pass</p>
                <h2>{state.passName}</h2>
              </div>
              <span className={`is-${state.status}`}>{state.status}</span>
            </header>
            <dl>
              <div>
                <dt>Paid through</dt>
                <dd>{state.paidThrough}</dd>
              </div>
              <div>
                <dt>Automatic renewal</dt>
                <dd>{state.renewsAutomatically ? "On" : "Off"}</dd>
              </div>
              <div>
                <dt>Receipt</dt>
                <dd>{state.receiptRef}</dd>
              </div>
              <div>
                <dt>Grace period</dt>
                <dd>
                  {state.graceEnds
                    ? `Until ${state.graceEnds}`
                    : "Not currently in grace"}
                </dd>
              </div>
            </dl>
            {state.status === "cancelled" ? (
              <div className="membership-confirmation">
                <strong>Cancellation recorded without penalty.</strong>
                <p>
                  Your purchased access remains available through{" "}
                  {state.paidThrough}. No standing, safety or matching signal
                  changes.
                </p>
              </div>
            ) : (
              <button
                className="membership-secondary"
                onClick={() => dispatch({ type: "request-cancellation" })}
                type="button"
              >
                Review cancellation
              </button>
            )}
          </article>

          <article className="membership-refund">
            <p className="fie-kicker">Receipt and refund</p>
            <h2>
              {state.refundState === "provider_confirmed"
                ? "Provider confirmed the refund."
                : state.refundState === "pending"
                  ? "Refund review is pending."
                  : "Need a payment reviewed?"}
            </h2>
            <p>
              A request is not a refund promise. Obiara marks it complete only
              after the payment provider confirms the return.
            </p>
            {state.refundState === "none" ? (
              <>
                <label>
                  <span>Reason for review</span>
                  <textarea
                    onChange={(event) =>
                      dispatch({
                        type: "refund-reason",
                        value: event.target.value,
                      })
                    }
                    placeholder="Describe the payment issue without card or phone details"
                    rows={4}
                    value={state.refundReason}
                  />
                </label>
                <button
                  disabled={state.refundReason.trim().length < 12}
                  onClick={() => dispatch({ type: "request-refund" })}
                  type="button"
                >
                  Request refund review
                </button>
              </>
            ) : (
              <div className="refund-status">
                <span>{state.refundRef}</span>
                <strong>
                  {state.refundState === "pending"
                    ? "Awaiting provider confirmation"
                    : "Provider confirmation recorded"}
                </strong>
                {state.refundState === "pending" ? (
                  <button
                    onClick={() =>
                      dispatch({ type: "provider-confirm-refund" })
                    }
                    type="button"
                  >
                    Preview provider confirmation
                  </button>
                ) : null}
              </div>
            )}
          </article>
        </section>

        <aside className="membership-law">
          <p className="fie-kicker">What never changes</p>
          <h2>No purchase can improve your romantic standing.</h2>
          <p>
            Cancelling does not punish you. Grace never hides an expiry.
            Receipts and refund references stay opaque, and no raw payment data
            appears here.
          </p>
        </aside>
      </section>

      {state.cancellationPending ? (
        <div className="membership-modal" role="dialog" aria-modal="true" aria-labelledby="cancel-title">
          <div>
            <p className="fie-kicker">Before you cancel</p>
            <h2 id="cancel-title">Your paid time remains yours.</h2>
            <p>
              Cancellation stops future renewal only. Access continues through{" "}
              {state.paidThrough}, with no standing or safety penalty.
            </p>
            <div>
              <button
                className="membership-secondary"
                onClick={() => dispatch({ type: "keep-membership" })}
                type="button"
              >
                Keep membership
              </button>
              <button
                onClick={() => dispatch({ type: "confirm-cancellation" })}
                type="button"
              >
                Confirm cancellation
              </button>
            </div>
          </div>
        </div>
      ) : null}
      <CompoundBottomNavigation contextLabel="Membership" />
    </main>
  );
}
