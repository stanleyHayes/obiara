"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";

interface Membership {
  passId: string;
  passName: string;
  status: "active" | "grace" | "expired" | "refund_pending" | "refunded";
  paidThrough: string;
  graceUntil: string;
  renewsAutomatically: boolean;
  receiptRef: string;
  refundRequestRef?: string;
  revision: number;
}

export function MembershipSettings() {
  const [membership, setMembership] = useState<Membership | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [action, setAction] = useState<"cancel" | "refund" | null>(null);
  const [confirmingCancel, setConfirmingCancel] = useState(false);
  const [message, setMessage] = useState("");
  const command = useRef<string | null>(null);

  useEffect(() => {
    let active = true;
    void fetch("/api/membership")
      .then(async (response) => {
        const payload = (await response.json()) as {
          membership?: Membership | null;
          message?: string;
        };
        if (!response.ok)
          throw new Error(payload.message || "Membership could not be loaded.");
        if (active) {
          setMembership(payload.membership ?? null);
          setLoaded(true);
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setMessage(
            error instanceof Error
              ? error.message
              : "Membership could not be loaded.",
          );
          setLoaded(true);
        }
      });
    return () => {
      active = false;
    };
  }, []);

  async function mutate(nextAction: "cancel" | "refund") {
    command.current ??= `membership-${nextAction}-${crypto.randomUUID()}`;
    setAction(nextAction);
    setMessage("");
    try {
      const response = await fetch("/api/membership", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": command.current,
        },
        body: JSON.stringify({ action: nextAction }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Membership & { message?: string }) | null;
      if (!response.ok || !payload?.passId) {
        throw new Error(
          payload?.message || "The membership action could not be completed.",
        );
      }
      setMembership(payload);
      setConfirmingCancel(false);
      command.current = null;
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The membership action could not be completed.",
      );
    } finally {
      setAction(null);
    }
  }

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

        {!loaded ? <p role="status">Loading membership…</p> : null}
        {message ? (
          <p className="profile-error" role="alert">
            {message}
          </p>
        ) : null}

        {loaded && !membership ? (
          <section className="membership-grid">
            <article className="membership-pass">
              <p className="fie-kicker">No current pass</p>
              <h2>You do not have an active paid membership.</h2>
              <p>
                Nothing will renew or be charged from this account. Purchase
                options appear only after the approved payment provider is
                available.
              </p>
            </article>
          </section>
        ) : null}

        {membership ? (
          <section className="membership-grid">
            <article className="membership-pass">
              <header>
                <div>
                  <p className="fie-kicker">Current pass</p>
                  <h2>{membership.passName.replaceAll("_", " ")}</h2>
                </div>
                <span className={`is-${membership.status}`}>
                  {membership.status.replaceAll("_", " ")}
                </span>
              </header>
              <dl>
                <div>
                  <dt>Paid through</dt>
                  <dd>
                    {new Date(membership.paidThrough).toLocaleDateString(
                      "en-GH",
                    )}
                  </dd>
                </div>
                <div>
                  <dt>Automatic renewal</dt>
                  <dd>{membership.renewsAutomatically ? "On" : "Off"}</dd>
                </div>
                <div>
                  <dt>Receipt</dt>
                  <dd>{membership.receiptRef.slice(0, 12)}…</dd>
                </div>
                <div>
                  <dt>Grace period</dt>
                  <dd>
                    Until{" "}
                    {new Date(membership.graceUntil).toLocaleDateString(
                      "en-GH",
                    )}
                  </dd>
                </div>
              </dl>
              {membership.renewsAutomatically ? (
                <button
                  className="membership-secondary"
                  onClick={() => setConfirmingCancel(true)}
                  type="button"
                >
                  Review cancellation
                </button>
              ) : (
                <div className="membership-confirmation">
                  <strong>Renewal is cancelled without penalty.</strong>
                  <p>
                    Your purchased access remains available through the
                    paid-through date.
                  </p>
                </div>
              )}
            </article>

            <article className="membership-refund">
              <p className="fie-kicker">Receipt and refund</p>
              <h2>
                {membership.status === "refunded"
                  ? "Provider confirmed the refund."
                  : membership.status === "refund_pending"
                    ? "Refund review is pending."
                    : "Need the cancelled payment reviewed?"}
              </h2>
              <p>
                A request is not a refund promise. Obiara marks it complete only
                after the payment provider confirms the return.
              </p>
              {!membership.renewsAutomatically &&
              membership.status !== "refund_pending" &&
              membership.status !== "refunded" ? (
                <button
                  disabled={action === "refund"}
                  onClick={() => void mutate("refund")}
                  type="button"
                >
                  {action === "refund"
                    ? "Requesting review"
                    : "Request refund review"}
                </button>
              ) : null}
              {membership.refundRequestRef ? (
                <div className="refund-status">
                  <span>{membership.refundRequestRef.slice(0, 12)}…</span>
                  <strong>
                    {membership.status === "refunded"
                      ? "Provider confirmation recorded"
                      : "Awaiting provider confirmation"}
                  </strong>
                </div>
              ) : null}
            </article>
          </section>
        ) : null}

        <aside className="membership-law">
          <p className="fie-kicker">What never changes</p>
          <h2>No purchase can improve your romantic standing.</h2>
          <p>
            Cancelling does not punish you. Grace never hides an expiry.
            Receipts and refund references stay opaque.
          </p>
        </aside>
      </section>

      {confirmingCancel && membership ? (
        <div
          className="membership-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="cancel-title"
        >
          <div>
            <p className="fie-kicker">Before you cancel</p>
            <h2 id="cancel-title">Your paid time remains yours.</h2>
            <p>
              Cancellation stops future renewal only. Access continues through{" "}
              {new Date(membership.paidThrough).toLocaleDateString("en-GH")}.
            </p>
            <div>
              <button
                className="membership-secondary"
                onClick={() => setConfirmingCancel(false)}
                type="button"
              >
                Keep membership
              </button>
              <button
                disabled={action === "cancel"}
                onClick={() => void mutate("cancel")}
                type="button"
              >
                {action === "cancel"
                  ? "Cancelling renewal"
                  : "Confirm cancellation"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
      <CompoundBottomNavigation contextLabel="Membership" />
    </main>
  );
}
