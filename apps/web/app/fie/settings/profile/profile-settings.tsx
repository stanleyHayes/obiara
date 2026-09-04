"use client";

import Link from "next/link";
import { useEffect, useReducer, useRef, useState } from "react";

import {
  CompoundBottomNavigation,
  CompoundRail,
} from "../../compound-navigation";
import {
  displayNameLimit,
  initialProfileSettingsState,
  introductionLimit,
  profileSettingsReducer,
  validateProfileForm,
  type FieldVisibility,
} from "./profile-model";
import { ObiaraSelect } from "@obiara/ui-web";

const visibilityOptions: ReadonlyArray<{
  value: FieldVisibility;
  label: string;
}> = [
  { value: "private", label: "Only me" },
  { value: "circles", label: "My circles" },
  { value: "community", label: "Community" },
];

export function ProfileSettings() {
  const [state, dispatch] = useReducer(
    profileSettingsReducer,
    initialProfileSettingsState,
  );
  const [copied, setCopied] = useState(false);
  const [revision, setRevision] = useState(0);
  const [updatedAt, setUpdatedAt] = useState("");
  const [requestState, setRequestState] = useState<
    "loading" | "ready" | "saving" | "error"
  >("loading");
  const [requestError, setRequestError] = useState("");
  const [initialized, setInitialized] = useState(false);
  const [doorwayQuestion, setDoorwayQuestion] = useState("");
  const [doorwayState, setDoorwayState] = useState<
    "loading" | "ready" | "saving" | "saved" | "error"
  >("loading");
  const [doorwayMessage, setDoorwayMessage] = useState("");
  const commandID = useRef<string | null>(null);
  const { account } = state;

  useEffect(() => {
    let active = true;
    void fetch("/api/profile")
      .then(async (response) => {
        const payload = (await response.json()) as {
          profile?: {
            memberId: string;
            displayName: string | null;
            introduction: string | null;
            displayNameVisibility: FieldVisibility;
            introductionVisibility: FieldVisibility;
            revision: number;
            updatedAt: string;
          } | null;
          message?: string;
        };
        if (!response.ok)
          throw new Error(
            payload.message || "Your profile could not be loaded.",
          );
        if (active && payload.profile) {
          dispatch({
            type: "hydrate",
            memberRef: payload.profile.memberId,
            displayName: payload.profile.displayName || "",
            introduction: payload.profile.introduction || "",
            nameVisibility: payload.profile.displayNameVisibility,
            introVisibility: payload.profile.introductionVisibility,
          });
          setRevision(payload.profile.revision);
          setUpdatedAt(payload.profile.updatedAt);
        } else if (active) {
          dispatch({
            type: "hydrate",
            memberRef: "Not created",
            displayName: "",
            introduction: "",
            nameVisibility: "private",
            introVisibility: "private",
          });
        }
        if (active) {
          setInitialized(true);
          setRequestState("ready");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setRequestError(
            error instanceof Error
              ? error.message
              : "Your profile could not be loaded.",
          );
          setInitialized(true);
          setRequestState("error");
        }
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    void fetch("/api/doorway-question")
      .then(async (response) => {
        const payload = (await response.json()) as {
          question?: { text: string } | null;
          message?: string;
        };
        if (!response.ok) {
          throw new Error(
            payload.message || "Your doorway question could not be loaded.",
          );
        }
        if (active) {
          setDoorwayQuestion(payload.question?.text ?? "");
          setDoorwayState("ready");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setDoorwayMessage(
            error instanceof Error
              ? error.message
              : "Your doorway question could not be loaded.",
          );
          setDoorwayState("error");
        }
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    commandID.current = null;
  }, [
    state.displayName,
    state.introduction,
    state.nameVisibility,
    state.introVisibility,
  ]);

  async function saveProfile() {
    const validationError = validateProfileForm(state);
    if (validationError) {
      setRequestError(validationError);
      setRequestState("error");
      return;
    }
    commandID.current ??= `profile-${crypto.randomUUID()}`;
    setRequestState("saving");
    setRequestError("");
    try {
      const response = await fetch("/api/profile", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": commandID.current,
        },
        body: JSON.stringify({
          displayName: state.displayName,
          introduction: state.introduction,
          displayNameVisibility: state.nameVisibility,
          introductionVisibility: state.introVisibility,
          expectedRevision: revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as {
        memberId?: string;
        displayName?: string | null;
        introduction?: string | null;
        displayNameVisibility?: FieldVisibility;
        introductionVisibility?: FieldVisibility;
        revision?: number;
        updatedAt?: string;
        message?: string;
      } | null;
      if (
        !response.ok ||
        !payload?.memberId ||
        typeof payload.revision !== "number"
      ) {
        throw new Error(payload?.message || "Your profile could not be saved.");
      }
      dispatch({
        type: "hydrate",
        memberRef: payload.memberId,
        displayName: payload.displayName || "",
        introduction: payload.introduction || "",
        nameVisibility: payload.displayNameVisibility || "private",
        introVisibility: payload.introductionVisibility || "private",
        saved: true,
      });
      setRevision(payload.revision);
      setUpdatedAt(payload.updatedAt || "");
      commandID.current = null;
      setRequestState("ready");
    } catch (error) {
      setRequestError(
        error instanceof Error
          ? error.message
          : "Your profile could not be saved.",
      );
      setRequestState("error");
    }
  }

  async function saveDoorwayQuestion() {
    const text = doorwayQuestion.trim();
    if (!text || [...text].length > 60) {
      setDoorwayMessage("Use 1–60 characters for your doorway question.");
      setDoorwayState("error");
      return;
    }
    setDoorwayState("saving");
    setDoorwayMessage("");
    try {
      const response = await fetch("/api/doorway-question", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text, custom: true }),
      });
      const payload = (await response.json().catch(() => null)) as {
        text?: string;
        message?: string;
      } | null;
      if (!response.ok || !payload?.text) {
        throw new Error(
          payload?.message || "Your doorway question could not be saved.",
        );
      }
      setDoorwayQuestion(payload.text);
      setDoorwayState("saved");
    } catch (error) {
      setDoorwayMessage(
        error instanceof Error
          ? error.message
          : "Your doorway question could not be saved.",
      );
      setDoorwayState("error");
    }
  }

  const initials = account.displayName
    .split(" ")
    .map((part) => part[0] ?? "")
    .join("")
    .slice(0, 2)
    .toUpperCase();

  function copyRef() {
    navigator.clipboard
      ?.writeText(account.memberRef)
      .then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      })
      .catch(() => {
        // Clipboard unavailable (insecure context); stay silent.
      });
  }

  if (
    !initialized ||
    (requestState === "error" && revision === 0 && account.memberRef === "")
  ) {
    return (
      <main className="fie-shell profile-shell">
        <CompoundRail contextLabel="Profile" />
        <section className="fie-main profile-main">
          <header className="profile-topbar">
            <Link href="/fie">Back to Fie</Link>
          </header>
          <section className="profile-hero" role="status">
            <p className="fie-kicker">Your profile</p>
            <h1>{requestError || "Loading your profile…"}</h1>
          </section>
        </section>
        <CompoundBottomNavigation contextLabel="Profile" />
      </main>
    );
  }

  return (
    <main className="fie-shell profile-shell">
      <CompoundRail contextLabel="Profile" />
      <section className="fie-main profile-main">
        <header className="profile-topbar">
          <Link href="/fie">Back to Fie</Link>
          <span>Your words stay yours · visibility per field</span>
        </header>

        <section className="profile-hero">
          <p className="fie-kicker">Your profile</p>
          <h1>Be known on your own terms.</h1>
          <p>
            Your name and introduction are yours. Choose who sees each field —
            community visibility always records your consent.
          </p>
        </section>

        <section className="profile-overview" aria-label="Account overview">
          <div className="profile-identity">
            <span className="profile-monogram" aria-hidden="true">
              {initials}
            </span>
            <div>
              <h2>{account.displayName}</h2>
              <p>{revision ? `Revision ${revision}` : "Create your profile"}</p>
            </div>
          </div>
          <dl className="profile-tiles">
            <div>
              <dt>Member reference</dt>
              <dd>
                <code>{account.memberRef}</code>
                <button type="button" onClick={copyRef}>
                  {copied ? "Copied" : "Copy"}
                </button>
              </dd>
            </div>
            <div>
              <dt>Last updated</dt>
              <dd>
                {updatedAt
                  ? new Date(updatedAt).toLocaleString("en-GH")
                  : "Not saved yet"}
              </dd>
            </div>
          </dl>
        </section>

        <section className="profile-edit" aria-label="Edit profile">
          <header>
            <p className="fie-kicker">Edit profile</p>
            <h2>Only what you choose to share.</h2>
          </header>

          <form
            onSubmit={(event) => {
              event.preventDefault();
              void saveProfile();
            }}
          >
            <div className="profile-field-row">
              <label htmlFor="display-name">
                Display name
                <span className="profile-count">
                  {[...state.displayName].length}/{displayNameLimit}
                </span>
              </label>
              <input
                id="display-name"
                onChange={(event) =>
                  dispatch({ type: "display-name", value: event.target.value })
                }
                required
                value={state.displayName}
              />
              <ObiaraSelect
                label="Visible to"
                onChange={(value) =>
                  dispatch({
                    type: "name-visibility",
                    value: value as FieldVisibility,
                  })
                }
                options={visibilityOptions}
                value={state.nameVisibility}
              />
            </div>

            <div className="profile-field-row">
              <label htmlFor="introduction">
                Introduction
                <span className="profile-count">
                  {[...state.introduction].length}/{introductionLimit}
                </span>
              </label>
              <textarea
                id="introduction"
                onChange={(event) =>
                  dispatch({ type: "introduction", value: event.target.value })
                }
                rows={4}
                value={state.introduction}
              />
              <ObiaraSelect
                label="Visible to"
                onChange={(value) =>
                  dispatch({
                    type: "intro-visibility",
                    value: value as FieldVisibility,
                  })
                }
                options={visibilityOptions}
                value={state.introVisibility}
              />
            </div>

            <p className="profile-note">
              Profile fields never carry phone numbers, emails or links — Obiara
              connects people itself. Choosing Community records a consent
              receipt for that field.
            </p>

            {requestError ? (
              <p className="profile-error" role="alert">
                {requestError}
              </p>
            ) : null}
            {state.saved && requestState === "ready" ? (
              <p className="profile-saved" role="status">
                Profile saved. Your circle sees the change on their next view.
              </p>
            ) : null}

            <button
              className="profile-save"
              disabled={requestState === "loading" || requestState === "saving"}
              type="submit"
            >
              {requestState === "saving" ? "Saving securely" : "Save changes"}
            </button>
          </form>
        </section>

        <section
          className="profile-edit"
          aria-labelledby="doorway-question-title"
        >
          <header>
            <p className="fie-kicker">Doorway question</p>
            <h2 id="doorway-question-title">Choose what a seed must answer.</h2>
          </header>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void saveDoorwayQuestion();
            }}
          >
            <div className="profile-field-row">
              <label htmlFor="doorway-question">
                Your question
                <span className="profile-count">
                  {[...doorwayQuestion].length}/60
                </span>
              </label>
              <input
                disabled={doorwayState === "loading"}
                id="doorway-question"
                maxLength={60}
                onChange={(event) => {
                  setDoorwayQuestion(event.target.value);
                  setDoorwayState("ready");
                  setDoorwayMessage("");
                }}
                placeholder="What does care look like on an ordinary day?"
                required
                value={doorwayQuestion}
              />
            </div>
            <p className="profile-note">
              This is the one bounded prompt a person answers before sowing a
              seed. Contact details and links are rejected.
            </p>
            {doorwayMessage ? (
              <p className="profile-error" role="alert">
                {doorwayMessage}
              </p>
            ) : null}
            {doorwayState === "saved" ? (
              <p className="profile-saved" role="status">
                Doorway question saved.
              </p>
            ) : null}
            <button
              className="profile-save"
              disabled={doorwayState === "loading" || doorwayState === "saving"}
              type="submit"
            >
              {doorwayState === "saving"
                ? "Saving question"
                : "Save doorway question"}
            </button>
          </form>
        </section>

        <nav className="profile-settings-links" aria-label="More settings">
          <Link href="/fie/settings/verification">
            <strong>Verification</strong>
            <span>Add your Ghana Card to earn a verified badge</span>
          </Link>
          <Link href="/fie/settings/notifications">
            <strong>Notifications</strong>
            <span>Quiet hours, caps and channels</span>
          </Link>
          <Link href="/fie/settings/membership">
            <strong>Membership</strong>
            <span>Terms, receipts and renewal</span>
          </Link>
          <Link href="/fie/settings/suban">
            <strong>Suban</strong>
            <span>Your character marks and history</span>
          </Link>
          <Link href="/fie/settings/privacy">
            <strong>Privacy requests</strong>
            <span>Export, deletion and request status</span>
          </Link>
          <Link href="/fie/settings/consent">
            <strong>Consent controls</strong>
            <span>Purpose-bound processing choices</span>
          </Link>
        </nav>
      </section>
      <CompoundBottomNavigation contextLabel="Profile" />
    </main>
  );
}
