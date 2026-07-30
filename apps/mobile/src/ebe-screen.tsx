import { type Href, useRouter } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

type Prompt = {
  id: string;
  version: number;
  language: string;
  cue: string;
  sourceKind: string;
  sourceCitation: string;
  sourceLocator?: string;
};
type Duel = {
  id: string;
  revision: number;
  complete: boolean;
  yourTurn: boolean;
  currentPrompt?: Prompt;
  turns: {
    number: number;
    prompt: Prompt;
    yours: boolean;
    yourAnswer?: string;
    yourAnswerCorrect?: boolean;
  }[];
};

export function EbeScreen({
  duelId,
  circleId,
}: Readonly<{ duelId: string; circleId: string }>) {
  const router = useRouter();
  const [duel, setDuel] = useState<Duel | null>(null);
  const [answer, setAnswer] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const command = useRef<string | null>(null);
  const load = useCallback(async () => {
    if (!duelId || !circleId) {
      setError("This duel needs its private circle reference.");
      setLoading(false);
      return;
    }
    try {
      setDuel(
        await apiRequest<Duel>(
          `/v1/circles/${encodeURIComponent(circleId)}/ebe/${encodeURIComponent(duelId)}`,
        ),
      );
      setError("");
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The duel could not be opened.",
      );
    } finally {
      setLoading(false);
    }
  }, [circleId, duelId]);
  useEffect(() => {
    void load();
    const timer = setInterval(() => void load(), 5000);
    return () => clearInterval(timer);
  }, [load]);
  async function submit() {
    if (!duel || !answer.trim()) return;
    command.current ??= `ebe-answer-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    setBusy(true);
    try {
      setDuel(
        await apiRequest<Duel>(
          `/v1/circles/${encodeURIComponent(circleId)}/ebe/${encodeURIComponent(duelId)}/answers`,
          {
            method: "POST",
            headers: { "Idempotency-Key": command.current },
            body: JSON.stringify({
              answer: answer.trim(),
              expectedRevision: duel.revision,
            }),
          },
        ),
      );
      setAnswer("");
      setError("");
      command.current = null;
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The answer could not be retained.",
      );
      await load();
    } finally {
      setBusy(false);
    }
  }
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
        <Text style={s.eyebrow}>ƐBƐ · SOURCED PROVERB CATALOG</Text>
        <Text accessibilityRole="header" style={s.title}>
          Reflect without being ranked.
        </Text>
        <Text style={s.body}>
          Turns alternate quietly. Your answer and accepted forms are never
          shown to the other player.
        </Text>
        {loading ? <ActivityIndicator color="#FF91A6" /> : null}
        {error ? <Text style={s.error}>{error}</Text> : null}
        {duel?.currentPrompt ? (
          <View style={s.card}>
            <Text style={s.cardEyebrow}>
              {duel.currentPrompt.language.toUpperCase()} · REVIEWED REVISION{" "}
              {duel.currentPrompt.version}
            </Text>
            <Text accessibilityRole="header" style={s.cardTitle}>
              {duel.currentPrompt.cue}
            </Text>
            <Text style={s.cardCopy}>
              Source: {duel.currentPrompt.sourceCitation}
            </Text>
            {duel.yourTurn ? (
              <>
                <TextInput
                  accessibilityLabel="Your private answer"
                  maxLength={280}
                  multiline
                  onChangeText={setAnswer}
                  placeholder="Write your reflection"
                  placeholderTextColor="#8B7780"
                  style={s.input}
                  value={answer}
                />
                <Pressable
                  disabled={busy || !answer.trim()}
                  onPress={() => void submit()}
                  style={[s.primary, (!answer.trim() || busy) && s.disabled]}
                >
                  <Text style={s.primaryText}>
                    {busy ? "Retaining…" : "Submit answer"}
                  </Text>
                </Pressable>
              </>
            ) : (
              <Text style={s.cardCopy}>
                Waiting for the other player. This page checks every five
                seconds.
              </Text>
            )}
          </View>
        ) : null}
        {duel ? (
          <View style={s.card}>
            <Text style={s.cardEyebrow}>
              {duel.complete ? "DUEL COMPLETE" : "PRIVATE HISTORY"}
            </Text>
            {duel.turns.map((turn) => (
              <View key={turn.number} style={s.turn}>
                <Text style={s.turnTitle}>
                  Turn {turn.number} ·{" "}
                  {turn.yours ? "You answered" : "Other player answered"}
                </Text>
                <Text style={s.cardCopy}>{turn.prompt.cue}</Text>
                <Text style={s.note}>
                  {turn.yours
                    ? `Your answer: ${turn.yourAnswer} · ${turn.yourAnswerCorrect ? "accepted form" : "another reflection"}`
                    : "The other answer remains private."}
                </Text>
              </View>
            ))}
          </View>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#20101A", flex: 1 },
  content: { gap: 18, padding: 24, paddingBottom: 56 },
  back: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", paddingVertical: 12 },
  eyebrow: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 46,
    letterSpacing: -2.5,
    lineHeight: 47,
  },
  body: {
    color: "rgba(255,243,230,.7)",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 26,
  },
  error: { color: "#FFD1D9", fontFamily: "Outfit_700Bold" },
  card: { backgroundColor: "#FFF0D9", borderRadius: 28, gap: 15, padding: 24 },
  cardEyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  cardTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    lineHeight: 35,
  },
  cardCopy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
  },
  input: {
    borderColor: "#B99FA9",
    borderRadius: 16,
    borderWidth: 1,
    color: "#28161F",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    minHeight: 120,
    padding: 16,
    textAlignVertical: "top",
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#9B315D",
    borderRadius: 999,
    justifyContent: "center",
    minHeight: 52,
  },
  primaryText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.45 },
  turn: {
    borderTopColor: "#D5C1C7",
    borderTopWidth: 1,
    gap: 6,
    paddingTop: 14,
  },
  turnTitle: { color: "#28161F", fontFamily: "Outfit_700Bold" },
  note: { color: "#8A365B", fontFamily: "Outfit_600SemiBold" },
});
