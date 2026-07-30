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

type Choice = "together" | "apart";
type Round = {
  id: string;
  sequence: number;
  you: { ready: boolean; connected: boolean; locked: boolean };
  other: { ready: boolean; connected: boolean; locked: boolean };
  paused: boolean;
  ownChoice?: Choice;
  yourReveal?: Choice;
  otherReveal?: Choice;
  complete: boolean;
};

export function AmpeScreen({
  roundId,
  circleId,
}: Readonly<{ roundId: string; circleId: string }>) {
  const router = useRouter();
  const [round, setRound] = useState<Round | null>(null);
  const [choice, setChoice] = useState<Choice | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const command = useRef<string | null>(null);

  const load = useCallback(async () => {
    if (!circleId || !roundId) {
      setError("This round needs its private circle reference.");
      setLoading(false);
      return;
    }
    try {
      setRound(
        await apiRequest<Round>(
          `/v1/circles/${encodeURIComponent(circleId)}/ampe/${encodeURIComponent(roundId)}`,
        ),
      );
      setError("");
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The round could not be opened.",
      );
    } finally {
      setLoading(false);
    }
  }, [circleId, roundId]);

  useEffect(() => {
    void load();
    const timer = setInterval(() => void load(), 3000);
    return () => clearInterval(timer);
  }, [load]);

  async function act(action: "ready" | "lock") {
    if (!round || (action === "lock" && !choice)) return;
    command.current ??= `ampe-${action}-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    setBusy(true);
    try {
      setRound(
        await apiRequest<Round>(
          `/v1/circles/${encodeURIComponent(circleId)}/ampe/${encodeURIComponent(roundId)}/commands`,
          {
            method: "POST",
            headers: { "Idempotency-Key": command.current },
            body: JSON.stringify({
              action,
              choice: action === "lock" ? choice : undefined,
              expectedSequence: round.sequence,
            }),
          },
        ),
      );
      setError("");
      command.current = null;
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The command could not be accepted.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }

  const canLock =
    round?.you.ready && round.other.ready && !round.you.locked && !round.paused;
  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <Pressable
          onPress={() =>
            router.push(
              (circleId
                ? `/fie/dan-mu/rooms/${circleId}`
                : "/fie/games") as Href,
            )
          }
        >
          <Text style={s.back}>Private room</Text>
        </Pressable>
        <Text style={s.eyebrow}>NO CAMERA · NO BODY INFERENCE</Text>
        <Text accessibilityRole="header" style={s.title}>
          Meet in the same beat.
        </Text>
        <Text style={s.body}>
          Manual choices remain private until both players lock. No score or
          profile signal is created.
        </Text>
        {loading ? <ActivityIndicator color="#FF91A6" /> : null}
        {error ? <Text style={s.error}>{error}</Text> : null}
        {round ? (
          <View style={s.stage}>
            <Text style={s.eyebrow}>SEQUENCE {round.sequence}</Text>
            <View style={s.statusRow}>
              <Text style={s.status}>
                OTHER ·{" "}
                {!round.other.connected
                  ? "RECONNECTING"
                  : round.other.locked
                    ? "LOCKED"
                    : round.other.ready
                      ? "READY"
                      : "NOT READY"}
              </Text>
              <Text style={s.status}>
                YOU ·{" "}
                {!round.you.connected
                  ? "RECONNECTING"
                  : round.you.locked
                    ? "LOCKED"
                    : round.you.ready
                      ? "READY"
                      : "NOT READY"}
              </Text>
            </View>
            <Text accessibilityRole="header" style={s.stageTitle}>
              {round.complete
                ? "Both choices revealed."
                : round.paused
                  ? "Round paused."
                  : !round.you.ready
                    ? "Join this beat."
                    : !round.other.ready
                      ? "Waiting quietly."
                      : round.you.locked
                        ? "Your choice is held."
                        : "Choose in private."}
            </Text>
            {!round.you.ready ? (
              <Pressable
                disabled={busy}
                onPress={() => void act("ready")}
                style={s.primary}
              >
                <Text style={s.primaryText}>I’m ready</Text>
              </Pressable>
            ) : null}
            {canLock ? (
              <>
                <View style={s.choices}>
                  {(["together", "apart"] as const).map((gesture) => (
                    <Pressable
                      accessibilityRole="radio"
                      accessibilityState={{ checked: choice === gesture }}
                      key={gesture}
                      onPress={() => setChoice(gesture)}
                      style={[s.choice, choice === gesture && s.selected]}
                    >
                      <Text style={s.choiceText}>
                        {gesture === "together" ? "Together" : "Apart"}
                      </Text>
                    </Pressable>
                  ))}
                </View>
                <Pressable
                  disabled={busy || !choice}
                  onPress={() => void act("lock")}
                  style={[s.primary, !choice && s.disabled]}
                >
                  <Text style={s.primaryText}>Lock my gesture</Text>
                </Pressable>
              </>
            ) : null}
            {round.you.locked && !round.complete ? (
              <Text style={s.body}>
                Your choice is hidden until the other player locks.
              </Text>
            ) : null}
            {round.paused ? (
              <Text style={s.body}>
                The server paused after a missing heartbeat. Reconnecting never
                forfeits or reveals either choice.
              </Text>
            ) : null}
            {round.complete ? (
              <View accessibilityLiveRegion="polite" style={s.reveal}>
                <Text style={s.revealText}>
                  OTHER · {round.otherReveal?.toUpperCase()}
                </Text>
                <Text style={s.revealText}>
                  YOU · {round.yourReveal?.toUpperCase()}
                </Text>
              </View>
            ) : null}
          </View>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#190D17", flex: 1 },
  content: { gap: 18, padding: 24, paddingBottom: 56 },
  back: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", paddingVertical: 12 },
  eyebrow: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 12,
    letterSpacing: 1.4,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Fraunces_700Bold",
    fontSize: 40,
    lineHeight: 44,
  },
  body: {
    color: "rgba(255,243,230,.76)",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 26,
  },
  error: { color: "#FFD1D9", fontFamily: "Outfit_700Bold" },
  stage: {
    backgroundColor: "#28131F",
    borderColor: "rgba(255,243,230,.15)",
    borderRadius: 24,
    borderWidth: 1,
    gap: 18,
    padding: 24,
  },
  statusRow: { gap: 6 },
  status: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", fontSize: 12 },
  stageTitle: {
    color: "#FFF3E6",
    fontFamily: "Fraunces_700Bold",
    fontSize: 28,
    lineHeight: 34,
  },
  choices: { flexDirection: "row", gap: 12 },
  choice: {
    borderColor: "rgba(255,243,230,.3)",
    borderRadius: 18,
    borderWidth: 1,
    flex: 1,
    padding: 18,
  },
  selected: { backgroundColor: "#4B1E37", borderColor: "#FF91A6" },
  choiceText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    textAlign: "center",
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#FF6F91",
    borderRadius: 999,
    justifyContent: "center",
    minHeight: 52,
    paddingHorizontal: 20,
  },
  primaryText: { color: "#190D17", fontFamily: "Outfit_700Bold", fontSize: 16 },
  disabled: { opacity: 0.45 },
  reveal: { backgroundColor: "#FFF3E6", borderRadius: 18, gap: 8, padding: 18 },
  revealText: { color: "#190D17", fontFamily: "Outfit_700Bold", fontSize: 18 },
});
