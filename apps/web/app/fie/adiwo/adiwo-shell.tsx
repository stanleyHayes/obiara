"use client";

import { useCallback, useEffect, useState } from "react";

import { CompoundBottomNavigation, CompoundRail } from "../compound-navigation";
import { ObiaraSelect } from "@obiara/ui-web";

type CircleType =
  "community" | "campus" | "professional" | "interest" | "support";
interface Circle {
  id: string;
  type: CircleType;
  visibility: "private" | "discoverable";
  membership:
    "none" | "requested" | "member" | "host" | "owner" | "expelled" | "left";
  memberCount: number;
  revision: number;
  updatedAt: string;
  members?: { id: string; state: "requested" | "member" | "host" | "owner" }[];
}

const typeLabels: Record<CircleType, string> = {
  community: "Community",
  campus: "Campus",
  professional: "Professional",
  interest: "Interest",
  support: "Support",
};

export function AdiwoShell() {
  const [view, setView] = useState<"mine" | "discover">("mine");
  const [circles, setCircles] = useState<Circle[]>([]);
  const [busy, setBusy] = useState<string | null>("load");
  const [message, setMessage] = useState("");
  const [creating, setCreating] = useState(false);
  const [type, setType] = useState<CircleType>("community");

  const load = useCallback(async () => {
    setBusy("load");
    setMessage("");
    try {
      const response = await fetch(`/api/circles?view=${view}`);
      const payload = (await response.json()) as {
        items?: Circle[];
        message?: string;
      };
      if (!response.ok || !payload.items)
        throw new Error(payload.message || "Circles could not be loaded.");
      setCircles(payload.items);
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : "Circles could not be loaded.",
      );
    } finally {
      setBusy(null);
    }
  }, [view]);

  useEffect(() => {
    let active = true;
    void fetch(`/api/circles?view=${view}`)
      .then(async (response) => {
        const payload = (await response.json()) as {
          items?: Circle[];
          message?: string;
        };
        if (!response.ok || !payload.items)
          throw new Error(payload.message || "Circles could not be loaded.");
        if (active) setCircles(payload.items);
      })
      .catch((error: unknown) => {
        if (active)
          setMessage(
            error instanceof Error
              ? error.message
              : "Circles could not be loaded.",
          );
      })
      .finally(() => {
        if (active) setBusy(null);
      });
    return () => {
      active = false;
    };
  }, [view]);

  async function act(body: Record<string, unknown>, key: string) {
    setBusy(key);
    setMessage("");
    try {
      const response = await fetch("/api/circles", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `circle-${crypto.randomUUID()}`,
        },
        body: JSON.stringify(body),
      });
      const payload = (await response.json()) as Circle & { message?: string };
      if (!response.ok || !payload.id)
        throw new Error(
          payload.message || "The circle action could not be completed.",
        );
      await load();
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The circle action could not be completed.",
      );
      setBusy(null);
    }
  }

  return (
    <main className="fie-shell adiwo-shell">
      <CompoundRail current="adiwo" />
      <section className="fie-main adiwo-main">
        <header className="adiwo-topbar">
          <div>
            <p className="fie-kicker">Adiwo · the courtyard</p>
            <h1>Belonging is deliberate.</h1>
            <p>
              Private circles reveal nothing before entry. Discoverable circles
              show only their type, reference and aggregate size until a host
              approves a request.
            </p>
          </div>
          <div className="adiwo-count">
            <strong>{circles.length}</strong>
            <span>{view === "mine" ? "Your circles" : "Available now"}</span>
          </div>
        </header>

        <section className="adiwo-circles" aria-labelledby="circles-title">
          <header>
            <div>
              <p className="fie-kicker">Your places</p>
              <h2 id="circles-title">Circles in the courtyard</h2>
            </div>
            <div aria-label="Choose circle view" className="adiwo-switch">
              <button
                aria-pressed={view === "mine"}
                onClick={() => {
                  setBusy("load");
                  setView("mine");
                }}
                type="button"
              >
                My circles
              </button>
              <button
                aria-pressed={view === "discover"}
                onClick={() => {
                  setBusy("load");
                  setView("discover");
                }}
                type="button"
              >
                Find a circle
              </button>
            </div>
          </header>

          {view === "mine" ? (
            <div className="adiwo-request">
              <div>
                <strong>Open a private courtyard</strong>
                <p>
                  New circles begin private. You can make one discoverable after
                  reviewing the boundary.
                </p>
              </div>
              <button
                onClick={() => setCreating((value) => !value)}
                type="button"
              >
                {creating ? "Close" : "Create circle"}
              </button>
            </div>
          ) : null}
          {creating ? (
            <div className="adiwo-request">
              <ObiaraSelect
                label="Circle type"
                onChange={(value) => setType(value as CircleType)}
                options={Object.entries(typeLabels).map(([value, label]) => ({
                  value,
                  label,
                }))}
                value={type}
              />
              <button
                disabled={busy !== null}
                onClick={() =>
                  void act(
                    {
                      action: "create",
                      id: `circle_${crypto.randomUUID()}`,
                      type,
                    },
                    "create",
                  )
                }
                type="button"
              >
                Create private circle
              </button>
            </div>
          ) : null}

          {message ? (
            <p className="profile-error" role="alert">
              {message}
            </p>
          ) : null}
          {busy === "load" ? <p role="status">Opening the courtyard…</p> : null}
          {busy !== "load" && circles.length === 0 ? (
            <div className="adiwo-now">
              <div>
                <p className="fie-kicker">Quiet for now</p>
                <h2>No circles in this view.</h2>
                <p>
                  {view === "mine"
                    ? "Create a private circle or browse discoverable courtyards."
                    : "No hosts have opened a discoverable courtyard yet."}
                </p>
              </div>
            </div>
          ) : null}

          <div className="adiwo-grid" aria-live="polite">
            {circles.map((circle) => (
              <article className="adiwo-card" key={circle.id}>
                <div className="adiwo-mark" aria-hidden="true">
                  {typeLabels[circle.type].slice(0, 2).toUpperCase()}
                </div>
                <p className="fie-kicker">{typeLabels[circle.type]} circle</p>
                <h3>{circle.id.slice(0, 18)}</h3>
                <p>
                  {circle.memberCount} active{" "}
                  {circle.memberCount === 1 ? "member" : "members"} ·{" "}
                  {circle.visibility}
                </p>
                <small>Your state: {circle.membership}</small>
                {circle.membership === "none" ? (
                  <button
                    disabled={busy !== null}
                    onClick={() =>
                      void act(
                        {
                          action: "request",
                          id: circle.id,
                          expectedRevision: circle.revision,
                        },
                        circle.id,
                      )
                    }
                    type="button"
                  >
                    {busy === circle.id
                      ? "Sending request…"
                      : "Request to join"}
                  </button>
                ) : null}
                {circle.membership === "owner" ? (
                  <button
                    disabled={busy !== null}
                    onClick={() =>
                      void act(
                        {
                          action: "visibility",
                          id: circle.id,
                          expectedRevision: circle.revision,
                          visibility:
                            circle.visibility === "private"
                              ? "discoverable"
                              : "private",
                        },
                        circle.id,
                      )
                    }
                    type="button"
                  >
                    {circle.visibility === "private"
                      ? "Allow discovery"
                      : "Make private"}
                  </button>
                ) : null}
                {circle.membership === "member" ||
                circle.membership === "host" ? (
                  <button
                    disabled={busy !== null}
                    onClick={() =>
                      void act(
                        {
                          action: "leave",
                          id: circle.id,
                          expectedRevision: circle.revision,
                        },
                        circle.id,
                      )
                    }
                    type="button"
                  >
                    Leave circle
                  </button>
                ) : null}
                {circle.membership === "owner"
                  ? circle.members?.map((member) =>
                      member.state === "owner" ? null : (
                        <div className="adiwo-request" key={member.id}>
                          <div>
                            <strong>{member.id.slice(0, 18)}</strong>
                            <p>{member.state}</p>
                          </div>
                          {member.state === "requested" ? (
                            <button
                              disabled={busy !== null}
                              onClick={() =>
                                void act(
                                  {
                                    action: "approve",
                                    id: circle.id,
                                    memberId: member.id,
                                    expectedRevision: circle.revision,
                                  },
                                  member.id,
                                )
                              }
                              type="button"
                            >
                              Approve
                            </button>
                          ) : member.state === "member" ? (
                            <button
                              disabled={busy !== null}
                              onClick={() =>
                                void act(
                                  {
                                    action: "promote",
                                    id: circle.id,
                                    memberId: member.id,
                                    expectedRevision: circle.revision,
                                  },
                                  member.id,
                                )
                              }
                              type="button"
                            >
                              Make host
                            </button>
                          ) : (
                            <button
                              disabled={busy !== null}
                              onClick={() =>
                                void act(
                                  {
                                    action: "expel",
                                    id: circle.id,
                                    memberId: member.id,
                                    expectedRevision: circle.revision,
                                  },
                                  member.id,
                                )
                              }
                              type="button"
                            >
                              Remove
                            </button>
                          )}
                        </div>
                      ),
                    )
                  : null}
              </article>
            ))}
          </div>
        </section>
        <CompoundBottomNavigation current="adiwo" />
      </section>
    </main>
  );
}
