import { type Href, useRouter } from "expo-router";
import { useEffect, useState } from "react";
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

type RoomEntry = {
  id: string;
  kind: "voice" | "event" | "notice";
  contentRef?: string;
  assetId?: string;
  durationMs?: number;
  startsAt?: string;
  createdAt: string;
  expiresAt: string;
};

export function RoomScreen({ roomId }: Readonly<{ roomId: string }>) {
  const router = useRouter();
  const [entries, setEntries] = useState<RoomEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [startingGame, setStartingGame] = useState(false);
  const [startingStory, setStartingStory] = useState(false);
  const [startingAmpe, setStartingAmpe] = useState(false);
  const [startingEbe, setStartingEbe] = useState(false);

  async function startOware() {
    setStartingGame(true);
    setError("");
    try {
      const game = await apiRequest<{ id: string }>(
        `/v1/circles/${encodeURIComponent(roomId)}/oware`,
        {
          method: "POST",
          headers: {
            "Idempotency-Key": `oware-create-${Date.now()}-${Math.random().toString(36).slice(2)}`,
          },
          body: JSON.stringify({}),
        },
      );
      router.push({
        pathname: "/fie/games/oware/[gameId]",
        params: { gameId: game.id, circleId: roomId },
      } as Href);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "A private Oware game could not be started.",
      );
    } finally {
      setStartingGame(false);
    }
  }

  async function startStory() {
    setStartingStory(true);
    setError("");
    try {
      const story = await apiRequest<{ id: string }>(
        `/v1/circles/${encodeURIComponent(roomId)}/stories`,
        {
          method: "POST",
          headers: {
            "Idempotency-Key": `story-create-${Date.now()}-${Math.random().toString(36).slice(2)}`,
          },
          body: JSON.stringify({ titleCode: "shared-story" }),
        },
      );
      router.push({
        pathname: "/fie/games/anansesem/[storyId]",
        params: { storyId: story.id, circleId: roomId },
      } as Href);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "A private story could not be started.",
      );
    } finally {
      setStartingStory(false);
    }
  }

  async function startAmpe() {
    setStartingAmpe(true);
    setError("");
    try {
      const round = await apiRequest<{ id: string }>(
        `/v1/circles/${encodeURIComponent(roomId)}/ampe`,
        {
          method: "POST",
          headers: {
            "Idempotency-Key": `ampe-create-${Date.now()}-${Math.random().toString(36).slice(2)}`,
          },
          body: JSON.stringify({}),
        },
      );
      router.push({
        pathname: "/fie/games/ampe/[roundId]",
        params: { roundId: round.id, circleId: roomId },
      } as Href);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "A private Ampe round could not be started.",
      );
    } finally {
      setStartingAmpe(false);
    }
  }

  async function startEbe() {
    setStartingEbe(true);
    setError("");
    try {
      const duel = await apiRequest<{ id: string }>(
        `/v1/circles/${encodeURIComponent(roomId)}/ebe`,
        {
          method: "POST",
          headers: {
            "Idempotency-Key": `ebe-create-${Date.now()}-${Math.random().toString(36).slice(2)}`,
          },
          body: JSON.stringify({}),
        },
      );
      router.push({
        pathname: "/fie/games/ebe/[duelId]",
        params: { duelId: duel.id, circleId: roomId },
      } as Href);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "A reviewed Ɛbɛ duel could not be started.",
      );
    } finally {
      setStartingEbe(false);
    }
  }

  useEffect(() => {
    let active = true;
    void apiRequest<{ items: RoomEntry[] }>(
      `/v1/circles/${encodeURIComponent(roomId)}/room?limit=50`,
    )
      .then((result) => {
        if (active) setEntries(result.items);
      })
      .catch((reason: unknown) => {
        if (active)
          setError(
            reason instanceof Error
              ? reason.message
              : "The room could not be opened.",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [roomId]);

  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <View style={s.topbar}>
          <Pressable
            onPress={() => router.push("/fie/adiwo" as Href)}
            style={s.back}
          >
            <Text style={s.backText}>Adiwo</Text>
          </Pressable>
          <Text style={s.private}>MEMBERS ONLY</Text>
        </View>
        <Text style={s.eyebrow}>RETAINED CIRCLE ROOM</Text>
        <Text accessibilityRole="header" style={s.title}>
          A quiet record, without invented activity.
        </Text>
        <Text style={s.lede}>
          Only persisted voice, event and notice references appear. Author
          identifiers remain privacy-keyed.
        </Text>
        <View style={s.reference}>
          <Text style={s.referenceLabel}>CIRCLE REFERENCE</Text>
          <Text selectable style={s.referenceValue}>
            {roomId}
          </Text>
        </View>
        {loading ? (
          <ActivityIndicator color="#8E3159" style={s.loader} />
        ) : null}
        {error ? (
          <View style={s.empty}>
            <Text style={s.emptyTitle}>Room unavailable</Text>
            <Text style={s.copy}>{error}</Text>
            <Text style={s.copy}>
              Private rooms remain indistinguishable from missing ones.
            </Text>
          </View>
        ) : null}
        {!loading && !error && entries.length === 0 ? (
          <View style={s.empty}>
            <Text style={s.emptyTitle}>This room is quiet.</Text>
            <Text style={s.copy}>
              Media upload and live-room providers are not simulated. Entries
              appear only after an authorized durable write.
            </Text>
          </View>
        ) : null}
        {entries.map((entry) => (
          <View key={entry.id} style={s.card}>
            <Text style={s.eyebrow}>{entry.kind}</Text>
            <Text style={s.cardTitle}>
              {entry.kind === "voice"
                ? `Private audio · ${Math.ceil((entry.durationMs ?? 0) / 1000)} seconds`
                : entry.kind === "event"
                  ? `Scheduled ${entry.startsAt ? new Date(entry.startsAt).toLocaleString() : "without a visible start"}`
                  : "Circle notice"}
            </Text>
            <Text style={s.copy}>
              Created {new Date(entry.createdAt).toLocaleString()}
            </Text>
            <Text style={s.copy}>
              Retained until {new Date(entry.expiresAt).toLocaleString()}
            </Text>
            <Text selectable style={s.itemRef}>
              {(entry.assetId || entry.contentRef || entry.id).slice(0, 24)}
            </Text>
          </View>
        ))}
        {!loading && !error ? (
          <View style={s.game}>
            <Text style={s.eyebrow}>PRIVATE PLAY</Text>
            <Text style={s.gameTitle}>Start one retained Oware board.</Text>
            <Text style={s.copy}>
              The server derives both players and accepts this only when the
              circle has exactly two active members.
            </Text>
            <Pressable
              disabled={startingGame}
              onPress={() => void startOware()}
              style={[s.gameButton, startingGame && s.disabled]}
            >
              <Text style={s.gameButtonText}>
                {startingGame ? "Preparing board…" : "Start private Oware"}
              </Text>
            </Pressable>
            <Pressable
              disabled={startingStory}
              onPress={() => void startStory()}
              style={[s.gameButton, startingStory && s.disabled]}
            >
              <Text style={s.gameButtonText}>
                {startingStory ? "Opening story…" : "Start private Anansesɛm"}
              </Text>
            </Pressable>
            <Pressable
              disabled={startingAmpe}
              onPress={() => void startAmpe()}
              style={[s.gameButton, startingAmpe && s.disabled]}
            >
              <Text style={s.gameButtonText}>
                {startingAmpe ? "Opening round…" : "Start private Ampe"}
              </Text>
            </Pressable>
            <Pressable
              disabled={startingEbe}
              onPress={() => void startEbe()}
              style={[s.gameButton, startingEbe && s.disabled]}
            >
              <Text style={s.gameButtonText}>
                {startingEbe ? "Opening duel…" : "Start reviewed Ɛbɛ"}
              </Text>
            </Pressable>
          </View>
        ) : null}
        <View style={s.boundary}>
          <Text style={s.boundaryTitle}>Truthful by design</Text>
          <Text style={s.boundaryCopy}>
            No fake transcript, caller, presence, timer, read receipt or turn
            state is shown.
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { padding: 20, paddingBottom: 50 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  back: { justifyContent: "center", minHeight: 44 },
  backText: { color: "#3A0E2E", fontFamily: "Outfit_700Bold" },
  private: {
    color: "#27755F",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1,
  },
  eyebrow: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
    marginTop: 30,
    textTransform: "uppercase",
  },
  title: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 44,
    letterSpacing: -2.7,
    lineHeight: 42,
    marginTop: 12,
  },
  lede: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginTop: 18,
  },
  reference: {
    borderColor: "rgba(58,14,46,.14)",
    borderRadius: 16,
    borderWidth: 1,
    marginVertical: 24,
    padding: 16,
  },
  referenceLabel: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 9,
    letterSpacing: 1,
  },
  referenceValue: {
    color: "#26101F",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
    marginTop: 7,
  },
  loader: { marginVertical: 40 },
  empty: {
    backgroundColor: "#FFFDFC",
    borderRadius: 20,
    marginBottom: 12,
    padding: 22,
  },
  emptyTitle: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 24,
  },
  copy: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    lineHeight: 20,
    marginTop: 7,
  },
  card: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,.11)",
    borderRadius: 18,
    borderWidth: 1,
    marginBottom: 10,
    padding: 20,
  },
  cardTitle: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 20,
    lineHeight: 25,
    marginTop: 9,
  },
  itemRef: {
    color: "#8E3159",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 11,
    marginTop: 12,
  },
  boundary: {
    backgroundColor: "#26101F",
    borderRadius: 20,
    marginTop: 12,
    padding: 22,
  },
  boundaryTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 20,
  },
  boundaryCopy: {
    color: "rgba(255,243,230,.68)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 21,
    marginTop: 8,
  },
  game: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,.11)",
    borderRadius: 20,
    borderWidth: 1,
    marginTop: 12,
    padding: 22,
  },
  gameTitle: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 25,
    lineHeight: 29,
    marginTop: 8,
  },
  gameButton: {
    alignItems: "center",
    backgroundColor: "#3A0E2E",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 17,
    minHeight: 50,
  },
  gameButtonText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.55 },
});
