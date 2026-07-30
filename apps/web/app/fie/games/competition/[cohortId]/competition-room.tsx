"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import { SafetySheet } from "../../../safety-sheet";

type Cohort = {
  id: string;
  capacity: number;
  enrolled: number;
  joined: boolean;
  status: "open" | "locked" | "started";
  competitionId?: string;
  revision: number;
};
type Competition = {
  id: string;
  status: "active" | "completed";
  revision: number;
  matches: {
    id: string;
    round: number;
    firstLabel: string;
    secondLabel: string;
    winnerLabel?: string;
    resultRecorded: boolean;
    youArePlaying: boolean;
  }[];
  ladder: { label: string; played: number; wins: number; you: boolean }[];
  reviews: {
    id: string;
    matchId: string;
    status: "open" | "resolved" | "appealed" | "final";
    decision: "none" | "no_action" | "rules_action";
    yours: boolean;
  }[];
};
type TournamentGame = {
  id: string;
  houses: number[];
  captured: number[];
  turn: "south" | "north";
  yourPlayer: "south" | "north";
  yourTurn: boolean;
  status: "active" | "completed" | "expired";
  winner: number;
  revision: number;
  moveDeadline: string;
};

export function CompetitionRoom({ cohortId }: Readonly<{ cohortId: string }>) {
  const [cohort, setCohort] = useState<Cohort | null>(null);
  const [competition, setCompetition] = useState<Competition | null>(null);
  const [game, setGame] = useState<{
    matchId: string;
    value: TournamentGame;
  } | null>(null);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const command = useRef<string | null>(null);
  const load = useCallback(async () => {
    try {
      const response = await fetch(
        `/api/competition?cohortId=${encodeURIComponent(cohortId)}`,
        { cache: "no-store" },
      );
      const payload = (await response.json().catch(() => null)) as
        (Cohort & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(payload?.message || "The cohort could not be opened.");
      setCohort(payload);
      if (payload.competitionId) {
        const bracketResponse = await fetch(
          `/api/competition?cohortId=${encodeURIComponent(cohortId)}&competitionId=${encodeURIComponent(payload.competitionId)}`,
          { cache: "no-store" },
        );
        const bracket = (await bracketResponse.json().catch(() => null)) as
          (Competition & { message?: string }) | null;
        if (!bracketResponse.ok || !bracket?.id)
          throw new Error(
            bracket?.message || "The bracket could not be opened.",
          );
        setCompetition(bracket);
      } else setCompetition(null);
      setMessage("");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The cohort could not be opened.",
      );
    }
  }, [cohortId]);
  useEffect(() => {
    const initial = window.setTimeout(() => void load(), 0);
    const timer = window.setInterval(() => void load(), 5000);
    return () => {
      window.clearTimeout(initial);
      window.clearInterval(timer);
    };
  }, [load]);
  async function mutate(action: "join" | "leave") {
    if (!cohort) return;
    command.current ??= `competition-${action}-${crypto.randomUUID()}`;
    setBusy(true);
    try {
      const response = await fetch("/api/competition", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": command.current,
        },
        body: JSON.stringify({
          action,
          cohortId,
          expectedRevision: cohort.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Cohort & { message?: string }) | null;
      if (!response.ok || !payload?.id)
        throw new Error(payload?.message || "The cohort action failed.");
      setCohort(payload);
      setMessage("");
      command.current = null;
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : "The cohort action failed.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }
  async function gameAction(
    action: "launch" | "move" | "finalize",
    matchId: string,
    pit?: number,
  ) {
    if (!competition) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/competition-oware", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `competition-oware-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({
          action,
          cohortId,
          competitionId: competition.id,
          matchId,
          gameId: game?.matchId === matchId ? game.value.id : undefined,
          pit,
          expectedRevision: game?.value.revision,
          expectedCompetitionRevision: competition.revision,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (TournamentGame & Competition & { message?: string }) | null;
      if (!response.ok || !payload)
        throw new Error(
          payload?.message || "The tournament board action failed.",
        );
      if (action === "finalize") {
        if (!payload.matches)
          throw new Error("The advanced bracket response was incomplete.");
        setCompetition(payload);
        setGame(null);
      } else {
        if (!payload.id || !payload.houses)
          throw new Error("The board response was incomplete.");
        setGame({ matchId, value: payload });
      }
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The tournament board action failed.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }
  async function reviewAction(
    action: "open" | "appeal",
    input: { matchId?: string; reviewId?: string; evidenceRef?: string },
  ) {
    if (!competition) return;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/competition-review", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": `competition-review-${crypto.randomUUID()}`,
        },
        body: JSON.stringify({
          action,
          cohortId,
          competitionId: competition.id,
          expectedRevision: competition.revision,
          ...input,
        }),
      });
      const payload = (await response.json().catch(() => null)) as
        (Competition & { message?: string }) | null;
      if (!response.ok || !payload?.matches)
        throw new Error(
          payload?.message || "The neutral review action failed.",
        );
      setCompetition(payload);
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The neutral review action failed.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="games-hall">
      <header>
        <Link href="/fie/games">← Games hall</Link>
        <strong>Private competition</strong>
        <SafetySheet
          context="this competition"
          contextRef={cohortId}
          surface="game"
        />
      </header>
      <section className="games-hero">
        <p className="fie-kicker">Invitation reference only</p>
        <h1>A bracket without a member grid.</h1>
        <p>
          Enrollment is explicit. Standings use private entrant labels, and game
          performance never changes matching visibility.
        </p>
      </section>
      {message ? <p role="alert">{message}</p> : null}
      {cohort ? (
        <section className="tournament">
          <p className="fie-kicker">
            {cohort.status} · revision {cohort.revision}
          </p>
          <h2>
            {cohort.enrolled} of {cohort.capacity} opted in.
          </h2>
          <p>
            {cohort.status === "open"
              ? "You may join or withdraw until the final seat locks the cohort."
              : cohort.status === "locked"
                ? "The cohort is full and waiting for operations to start the bracket."
                : "The private bracket is active."}
          </p>
          {cohort.status === "open" ? (
            <button
              disabled={busy}
              onClick={() => void mutate(cohort.joined ? "leave" : "join")}
              type="button"
            >
              {busy
                ? "Updating…"
                : cohort.joined
                  ? "Withdraw before lock"
                  : "Join this private cohort"}
            </button>
          ) : null}
        </section>
      ) : null}
      {competition ? (
        <>
          <section className="games-list" aria-label="Private bracket">
            {competition.matches.map((match) => (
              <article key={match.id}>
                <span>
                  Round {match.round} · {match.id}
                </span>
                <h3>
                  {match.firstLabel} vs {match.secondLabel}
                </h3>
                <strong>
                  {match.resultRecorded
                    ? `Winner: ${match.winnerLabel}`
                    : match.youArePlaying
                      ? "Your match · server board ready"
                      : "Result pending"}
                </strong>
                {match.youArePlaying && !match.resultRecorded ? (
                  <button
                    disabled={busy}
                    onClick={() => void gameAction("launch", match.id)}
                    type="button"
                  >
                    {game?.matchId === match.id
                      ? "Refresh tournament board"
                      : "Open tournament Oware"}
                  </button>
                ) : null}
                {game?.matchId === match.id ? (
                  <div className="competition-board">
                    <p>
                      {game.value.status === "active"
                        ? game.value.yourTurn
                          ? "Your move. Choose a non-empty house."
                          : "The other entrant is considering their move."
                        : game.value.winner < 0
                          ? "This board ended level. Open a rematch."
                          : game.value.winner ===
                              (game.value.yourPlayer === "north" ? 1 : 0)
                            ? "You won this board."
                            : "The other entrant won this board."}
                    </p>
                    <div aria-label="Tournament Oware board">
                      {game.value.houses.map((seeds, pit) => {
                        const yours =
                          game.value.yourPlayer === "north"
                            ? pit >= 6
                            : pit < 6;
                        return (
                          <button
                            aria-label={`House ${pit + 1}, ${seeds} seeds`}
                            disabled={
                              busy ||
                              game.value.status !== "active" ||
                              !game.value.yourTurn ||
                              !yours ||
                              seeds === 0
                            }
                            key={pit}
                            onClick={() =>
                              void gameAction("move", match.id, pit)
                            }
                            type="button"
                          >
                            {seeds}
                            <small>{yours ? " yours" : " other"}</small>
                          </button>
                        );
                      })}
                    </div>
                    {game.value.status === "completed" &&
                    game.value.winner >= 0 ? (
                      <button
                        disabled={busy}
                        onClick={() => void gameAction("finalize", match.id)}
                        type="button"
                      >
                        Verify board and advance winner
                      </button>
                    ) : null}
                    {game.value.status === "expired" ? (
                      <button
                        disabled={busy}
                        onClick={() =>
                          void reviewAction("open", {
                            matchId: match.id,
                            evidenceRef: game.value.id,
                          })
                        }
                        type="button"
                      >
                        Request neutral review of expiry
                      </button>
                    ) : null}
                  </div>
                ) : null}
              </article>
            ))}
          </section>
          {competition.reviews.length ? (
            <section className="tournament">
              <p className="fie-kicker">Neutral review record</p>
              <h2>Evidence first. Human decision.</h2>
              {competition.reviews.map((review) => (
                <article key={review.id}>
                  <strong>
                    {review.matchId} · {review.status}
                  </strong>
                  <p>
                    {review.decision === "none"
                      ? "No decision has been recorded."
                      : review.decision === "no_action"
                        ? "Human review found no rules action."
                        : "Human review recorded a rules action."}
                  </p>
                  {review.yours && review.status === "resolved" ? (
                    <button
                      disabled={busy}
                      onClick={() =>
                        void reviewAction("appeal", { reviewId: review.id })
                      }
                      type="button"
                    >
                      Appeal this decision once
                    </button>
                  ) : null}
                </article>
              ))}
            </section>
          ) : null}
          <section className="tournament">
            <p className="fie-kicker">Private ladder</p>
            <h2>No public rank.</h2>
            {competition.ladder.map((entry) => (
              <p key={entry.label}>
                <strong>{entry.label}</strong> · {entry.played} played ·{" "}
                {entry.wins} won
              </p>
            ))}
            <small>
              Every ladder result comes from a completed server-owned board.
              Expiry review accepts only the bound server session—never an
              accusation, score, reason, or free text.
            </small>
          </section>
        </>
      ) : null}
    </main>
  );
}
