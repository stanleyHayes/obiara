"use client";

import { FormEvent, useState } from "react";
import { ObiaraCheckbox } from "@obiara/ui-web";

type FormState = "idle" | "submitting" | "success" | "error";

export function WaitlistForm() {
  const [state, setState] = useState<FormState>("idle");
  const [message, setMessage] = useState("");
  const [consent, setConsent] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    setState("submitting");
    setMessage("");
    const form = new FormData(formElement);
    try {
      const response = await fetch("/api/waitlist", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: form.get("name"),
          email: form.get("email"),
          consent: form.get("consent") === "on",
        }),
      });
      const body = (await response.json().catch(() => null)) as {
        alreadyJoined?: boolean;
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          body?.message ?? "We could not save your place. Please try again.",
        );
      setState("success");
      setMessage(
        body?.alreadyJoined
          ? "You’re already on the list. We’ll email you when Obiara opens."
          : "Your place is saved. We’ll email you when Obiara opens.",
      );
      formElement.reset();
    } catch (cause) {
      setState("error");
      setMessage(
        cause instanceof Error
          ? cause.message
          : "We could not save your place. Please try again.",
      );
    }
  }

  return (
    <form className="waitlist-form reveal" onSubmit={submit}>
      <div className="waitlist-fields">
        <label>
          <span>Your name</span>
          <input
            autoComplete="name"
            maxLength={100}
            name="name"
            placeholder="Ama Mensah"
            required
          />
        </label>
        <label>
          <span>Email address</span>
          <input
            autoComplete="email"
            inputMode="email"
            name="email"
            placeholder="ama@example.com"
            required
            type="email"
          />
        </label>
      </div>
      <div className="consent-field">
        <ObiaraCheckbox
          checked={consent}
          label="Email me once Obiara is available. No newsletters or unrelated messages."
          name="consent"
          onChange={setConsent}
          required
        />
      </div>
      <button disabled={state === "submitting"} type="submit">
        {state === "submitting" ? "Saving your place…" : "Join the waitlist"}
        <span aria-hidden="true">↗</span>
      </button>
      <p aria-live="polite" className={`form-message ${state}`} role="status">
        {message}
      </p>
    </form>
  );
}
