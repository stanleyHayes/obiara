"use client";

import { ObiaraDateField } from "@obiara/ui-web";
import { useRef, useState } from "react";

import "./styles.css";

const acceptedTypes = ["image/jpeg", "image/png", "image/webp"] as const;
const maxImageBytes = 4 * 1024 * 1024;

type Side = "front" | "back";

interface Capture {
  readonly mediaType: string;
  readonly base64: string;
  readonly preview: string;
  readonly name: string;
}

/**
 * Reads a chosen file into the shape the API takes.
 *
 * The preview is a data URL held only in this tab: the picture a member is
 * about to send should be the picture they can see, and nothing is uploaded
 * until they press the button.
 */
function readCapture(file: File): Promise<Capture> {
  return new Promise((resolve, reject) => {
    if (!acceptedTypes.includes(file.type as (typeof acceptedTypes)[number])) {
      reject(new Error("Use a JPEG, PNG or WebP photograph."));
      return;
    }
    if (file.size > maxImageBytes) {
      reject(new Error("Each side must be under 4MB."));
      return;
    }
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("That image could not be read."));
    reader.onload = () => {
      const value = typeof reader.result === "string" ? reader.result : "";
      const comma = value.indexOf(",");
      if (comma < 0) {
        reject(new Error("That image could not be read."));
        return;
      }
      resolve({
        mediaType: file.type,
        base64: value.slice(comma + 1),
        preview: value,
        name: file.name,
      });
    };
    reader.readAsDataURL(file);
  });
}

export function VerificationSettings() {
  const [captures, setCaptures] = useState<Partial<Record<Side, Capture>>>({});
  const [cardNumber, setCardNumber] = useState("");
  const [dateOfBirth, setDateOfBirth] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [caseId, setCaseId] = useState<string | null>(null);
  const inputs = {
    front: useRef<HTMLInputElement>(null),
    back: useRef<HTMLInputElement>(null),
  };

  // A member is at least eighteen, so a date after today is always a mistake.
  const today = new Date().toISOString().slice(0, 10);
  const ready =
    Boolean(captures.front) &&
    Boolean(captures.back) &&
    cardNumber.trim().length >= 8 &&
    /^\d{4}-\d{2}-\d{2}$/.test(dateOfBirth);

  async function choose(side: Side, file: File | undefined) {
    if (!file) return;
    setError(null);
    try {
      const capture = await readCapture(file);
      setCaptures((current) => ({ ...current, [side]: capture }));
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "That image could not be read.",
      );
    }
  }

  async function submit() {
    if (!captures.front || !captures.back) return;
    setSubmitting(true);
    setError(null);
    try {
      const response = await fetch("/api/verification/ghana-card/documents", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          cardNumber: cardNumber.trim(),
          dateOfBirth,
          frontMediaType: captures.front.mediaType,
          frontBase64: captures.front.base64,
          backMediaType: captures.back.mediaType,
          backBase64: captures.back.base64,
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        caseId?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.caseId) {
        throw new Error(
          payload?.message ||
            "We could not send your card for review. Please try again.",
        );
      }
      // Cleared once the images are safely with the reviewer. Holding a card
      // photograph in a tab for longer than it takes to send is a copy nobody
      // asked for.
      setCaptures({});
      setCardNumber("");
      setDateOfBirth("");
      setCaseId(payload.caseId);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "We could not send your card for review. Please try again.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  if (caseId) {
    return (
      <main className="verification-shell">
        <section className="verification-card" role="status">
          <p className="fie-kicker">With a reviewer</p>
          <h1>Your card is being checked.</h1>
          <p>
            A trained reviewer looks at both sides — never an automatic
            approval. Your badge appears once they are done. You can keep using
            Obiara in the meantime.
          </p>
          <div className="verification-note">
            Reference {caseId.slice(0, 12)}…
          </div>
        </section>
      </main>
    );
  }

  return (
    <main className="verification-shell">
      <section className="verification-card">
        <p className="fie-kicker">Verification</p>
        <h1>Add your Ghana Card.</h1>
        <p>
          Both sides, photographed clearly. They are encrypted the moment they
          reach us and are only ever opened by a trained reviewer. Verification
          earns a badge on your profile — it is not needed to use Obiara.
        </p>

        <label className="verification-field">
          <span>Ghana Card number</span>
          <input
            autoComplete="off"
            onChange={(event) => setCardNumber(event.target.value)}
            placeholder="GHA-000000000-0"
            value={cardNumber}
          />
        </label>

        <ObiaraDateField
          label="Date of birth"
          max={today}
          onChange={setDateOfBirth}
          value={dateOfBirth}
        />

        <div className="verification-sides">
          {(["front", "back"] as const).map((side) => (
            <div className="verification-side" key={side}>
              <strong>{side === "front" ? "Front" : "Back"}</strong>
              {captures[side] ? (
                /* A data URL held in this tab, never a remote asset — there is
                   nothing for next/image to optimise, and routing a card
                   photograph through an image proxy is the opposite of what
                   this page promises. */
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  alt={`${side} of your card`}
                  src={captures[side].preview}
                />
              ) : (
                <div className="verification-placeholder" aria-hidden="true">
                  ▧
                </div>
              )}
              <input
                accept={acceptedTypes.join(",")}
                className="verification-file"
                onChange={(event) => choose(side, event.target.files?.[0])}
                ref={inputs[side]}
                type="file"
              />
              <button
                className="verification-choose"
                onClick={() => inputs[side].current?.click()}
                type="button"
              >
                {captures[side] ? "Replace photo" : "Choose photo"}
              </button>
            </div>
          ))}
        </div>

        <button
          className="verification-submit"
          disabled={!ready || submitting}
          onClick={submit}
          type="button"
        >
          {submitting ? "Sending securely…" : "Send for review"}
        </button>
        {error ? (
          <p className="verification-error" role="alert">
            {error}
          </p>
        ) : null}
      </section>
    </main>
  );
}
