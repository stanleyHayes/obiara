import { type Href, useRouter } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

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

export function CompetitionScreen({
  cohortId,
}: Readonly<{ cohortId: string }>) {
  const router = useRouter();
  const [cohort, setCohort] = useState<Cohort | null>(null);
  const [competition, setCompetition] = useState<Competition | null>(null);
  const [game, setGame] = useState<{
    matchId: string;
    value: TournamentGame;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const command = useRef<string | null>(null);
  const load = useCallback(async () => {
    if (!cohortId) {
      setError("A private cohort reference is required.");
      setLoading(false);
      return;
    }
    try {
      const value = await apiRequest<Cohort>(
        `/v1/game-cohorts/${encodeURIComponent(cohortId)}`,
      );
      setCohort(value);
      setCompetition(
        value.competitionId
          ? await apiRequest<Competition>(
              `/v1/game-cohorts/${encodeURIComponent(cohortId)}/competitions/${encodeURIComponent(value.competitionId)}`,
            )
          : null,
      );
      setError("");
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The private cohort could not be opened.",
      );
    } finally {
      setLoading(false);
    }
  }, [cohortId]);
  useEffect(() => {
    void load();
    const timer = setInterval(() => void load(), 5000);
    return () => clearInterval(timer);
  }, [load]);
  async function mutate(action: "join" | "leave") {
    if (!cohort) return;
    command.current ??= `competition-${action}-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    setBusy(true);
    try {
      setCohort(
        await apiRequest<Cohort>(
          `/v1/game-cohorts/${encodeURIComponent(cohortId)}/${action}`,
          {
            method: "POST",
            headers: { "Idempotency-Key": command.current },
            body: JSON.stringify({ expectedRevision: cohort.revision }),
          },
        ),
      );
      command.current = null;
      setError("");
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : "The cohort action failed.",
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
    const root = `/v1/game-cohorts/${encodeURIComponent(cohortId)}/competitions/${encodeURIComponent(competition.id)}/matches/${encodeURIComponent(matchId)}/oware`;
    const gameId = game?.matchId === matchId ? game.value.id : "";
    const path =
      action === "launch"
        ? root
        : action === "move"
          ? `${root}/${encodeURIComponent(gameId)}/moves`
          : `${root}/${encodeURIComponent(gameId)}/finalize`;
    setBusy(true);
    setError("");
    try {
      const headers = {
        "Idempotency-Key": `competition-oware-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      };
      if (action === "finalize") {
        const value = await apiRequest<Competition>(path, {
          method: "POST",
          headers,
          body: JSON.stringify({
            expectedCompetitionRevision: competition.revision,
          }),
        });
        setCompetition(value);
        setGame(null);
      } else {
        const value = await apiRequest<TournamentGame>(path, {
          method: "POST",
          headers,
          body: JSON.stringify(
            action === "move"
              ? { pit, expectedRevision: game?.value.revision }
              : {},
          ),
        });
        setGame({ matchId, value });
      }
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
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
    const path =
      action === "open"
        ? `/v1/game-cohorts/${encodeURIComponent(cohortId)}/competitions/${encodeURIComponent(competition.id)}/matches/${encodeURIComponent(input.matchId ?? "")}/reviews`
        : `/v1/game-cohorts/${encodeURIComponent(cohortId)}/competitions/${encodeURIComponent(competition.id)}/reviews/${encodeURIComponent(input.reviewId ?? "")}/appeal`;
    setBusy(true);
    setError("");
    try {
      const value = await apiRequest<Competition>(path, {
        method: "POST",
        headers: {
          "Idempotency-Key": `competition-review-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        },
        body: JSON.stringify(
          action === "open"
            ? {
                evidenceRef: input.evidenceRef,
                expectedRevision: competition.revision,
              }
            : { expectedRevision: competition.revision },
        ),
      });
      setCompetition(value);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The neutral review action failed.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }
  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <Pressable onPress={() => router.push("/fie/games" as Href)}>
          <Text style={s.back}>Games hall</Text>
        </Pressable>
        <Text style={s.eyebrow}>INVITATION REFERENCE ONLY</Text>
        <Text accessibilityRole="header" style={s.title}>
          A bracket without a member grid.
        </Text>
        <Text style={s.body}>
          Enrollment is explicit. Private entrant labels never affect matching
          visibility.
        </Text>
        {loading ? <ActivityIndicator color="#9B315D" /> : null}
        {error ? <Text style={s.error}>{error}</Text> : null}
        {cohort ? (
          <View style={s.card}>
            <Text style={s.eyebrow}>
              {cohort.status.toUpperCase()} · REVISION {cohort.revision}
            </Text>
            <Text style={s.cardTitle}>
              {cohort.enrolled} of {cohort.capacity} opted in.
            </Text>
            <Text style={s.copy}>
              {cohort.status === "open"
                ? "Join or withdraw before the final seat locks enrollment."
                : cohort.status === "locked"
                  ? "Full cohort. Waiting for operations to start."
                  : "The private bracket is active."}
            </Text>
            {cohort.status === "open" ? (
              <Pressable
                disabled={busy}
                onPress={() => void mutate(cohort.joined ? "leave" : "join")}
                style={s.primary}
              >
                <Text style={s.primaryText}>
                  {busy
                    ? "Updating…"
                    : cohort.joined
                      ? "Withdraw before lock"
                      : "Join private cohort"}
                </Text>
              </Pressable>
            ) : null}
          </View>
        ) : null}
        {competition ? (
          <>
            <Text accessibilityRole="header" style={s.sectionTitle}>
              Private bracket
            </Text>
            {competition.matches.map((match) => (
              <View key={match.id} style={s.card}>
                <Text style={s.eyebrow}>
                  ROUND {match.round} · {match.id}
                </Text>
                <Text style={s.cardTitle}>
                  {match.firstLabel} vs {match.secondLabel}
                </Text>
                <Text style={s.copy}>
                  {match.resultRecorded
                    ? `Winner: ${match.winnerLabel}`
                    : match.youArePlaying
                      ? "Your match · server board ready"
                      : "Result pending"}
                </Text>
                {match.youArePlaying && !match.resultRecorded ? (
                  <Pressable
                    disabled={busy}
                    onPress={() => void gameAction("launch", match.id)}
                    style={s.primary}
                  >
                    <Text style={s.primaryText}>
                      {game?.matchId === match.id
                        ? "Refresh tournament board"
                        : "Open tournament Oware"}
                    </Text>
                  </Pressable>
                ) : null}
                {game?.matchId === match.id ? (
                  <View style={s.board}>
                    <Text style={s.copy}>
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
                    </Text>
                    <View
                      accessibilityLabel="Tournament Oware board"
                      style={s.pits}
                    >
                      {game.value.houses.map((seeds, pit) => {
                        const yours =
                          game.value.yourPlayer === "north"
                            ? pit >= 6
                            : pit < 6;
                        return (
                          <Pressable
                            accessibilityLabel={`House ${pit + 1}, ${seeds} seeds`}
                            disabled={
                              busy ||
                              game.value.status !== "active" ||
                              !game.value.yourTurn ||
                              !yours ||
                              seeds === 0
                            }
                            key={pit}
                            onPress={() =>
                              void gameAction("move", match.id, pit)
                            }
                            style={[
                              s.pit,
                              (!yours || !game.value.yourTurn || seeds === 0) &&
                                s.pitDisabled,
                            ]}
                          >
                            <Text style={s.pitCount}>{seeds}</Text>
                            <Text style={s.pitLabel}>
                              {yours ? "yours" : "other"}
                            </Text>
                          </Pressable>
                        );
                      })}
                    </View>
                    {game.value.status === "completed" &&
                    game.value.winner >= 0 ? (
                      <Pressable
                        disabled={busy}
                        onPress={() => void gameAction("finalize", match.id)}
                        style={s.primary}
                      >
                        <Text style={s.primaryText}>
                          Verify board and advance winner
                        </Text>
                      </Pressable>
                    ) : null}
                    {game.value.status === "expired" ? (
                      <Pressable
                        disabled={busy}
                        onPress={() =>
                          void reviewAction("open", {
                            matchId: match.id,
                            evidenceRef: game.value.id,
                          })
                        }
                        style={s.primary}
                      >
                        <Text style={s.primaryText}>
                          Request neutral review of expiry
                        </Text>
                      </Pressable>
                    ) : null}
                  </View>
                ) : null}
              </View>
            ))}
            {competition.reviews.length ? (
              <View style={s.card}>
                <Text style={s.eyebrow}>NEUTRAL REVIEW RECORD</Text>
                <Text style={s.cardTitle}>Evidence first. Human decision.</Text>
                {competition.reviews.map((review) => (
                  <View key={review.id} style={s.review}>
                    <Text style={s.copy}>
                      {review.matchId} · {review.status}
                    </Text>
                    <Text style={s.note}>
                      {review.decision === "none"
                        ? "No decision has been recorded."
                        : review.decision === "no_action"
                          ? "Human review found no rules action."
                          : "Human review recorded a rules action."}
                    </Text>
                    {review.yours && review.status === "resolved" ? (
                      <Pressable
                        disabled={busy}
                        onPress={() =>
                          void reviewAction("appeal", { reviewId: review.id })
                        }
                        style={s.primary}
                      >
                        <Text style={s.primaryText}>
                          Appeal this decision once
                        </Text>
                      </Pressable>
                    ) : null}
                  </View>
                ))}
              </View>
            ) : null}
            <View style={s.card}>
              <Text style={s.eyebrow}>PRIVATE LADDER</Text>
              {competition.ladder.map((entry) => (
                <Text key={entry.label} style={s.copy}>
                  {entry.label} · {entry.played} played · {entry.wins} won
                </Text>
              ))}
              <Text style={s.note}>
                Every result comes from a completed server board. Expiry review
                accepts only that bound session—never an accusation, reason, or
                free text.
              </Text>
            </View>
          </>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { gap: 18, padding: 22, paddingBottom: 56 },
  back: { color: "#3A0E2E", fontFamily: "Outfit_700Bold", paddingVertical: 10 },
  eyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  title: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 46,
    letterSpacing: -2.7,
    lineHeight: 46,
  },
  body: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 26,
  },
  error: { color: "#A12C4D", fontFamily: "Outfit_700Bold" },
  card: {
    backgroundColor: "#FFF8EE",
    borderColor: "#DCCBD1",
    borderRadius: 22,
    borderWidth: 1,
    gap: 12,
    padding: 20,
  },
  cardTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 27,
    lineHeight: 32,
  },
  copy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#9B315D",
    borderRadius: 999,
    justifyContent: "center",
    minHeight: 50,
  },
  primaryText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  sectionTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    marginTop: 12,
  },
  note: {
    color: "#8A365B",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 22,
    marginTop: 8,
  },
  board: { gap: 12, marginTop: 6 },
  pits: { flexDirection: "row", flexWrap: "wrap", gap: 7 },
  pit: {
    alignItems: "center",
    backgroundColor: "#F1D9BC",
    borderColor: "#B98B69",
    borderRadius: 16,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 58,
    width: "30%",
  },
  pitDisabled: { opacity: 0.48 },
  pitCount: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 20,
  },
  pitLabel: { color: "#705C67", fontFamily: "Outfit_500Medium", fontSize: 10 },
  review: {
    borderTopColor: "#DCCBD1",
    borderTopWidth: 1,
    gap: 10,
    paddingTop: 12,
  },
});
