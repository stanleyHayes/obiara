import { type Href, useRouter } from "expo-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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

type Player = "south" | "north";
type OwareGame = {
  id: string;
  houses: number[];
  captured: number[];
  turn: Player;
  yourPlayer: Player;
  yourTurn: boolean;
  status: "active" | "completed" | "expired";
  winner: number;
  revision: number;
  moveDeadline: string;
  serverTime: string;
};

export function OwareScreen({
  gameId,
  circleId,
}: Readonly<{ gameId: string; circleId: string }>) {
  const router = useRouter();
  const [game, setGame] = useState<OwareGame | null>(null);
  const [selected, setSelected] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [moving, setMoving] = useState(false);
  const [error, setError] = useState("");
  const command = useRef<string | null>(null);

  const load = useCallback(async () => {
    if (!circleId || !gameId) {
      setError("This game needs its private circle reference.");
      setLoading(false);
      return;
    }
    try {
      const result = await apiRequest<OwareGame>(
        `/v1/circles/${encodeURIComponent(circleId)}/oware/${encodeURIComponent(gameId)}`,
      );
      setGame(result);
      setError("");
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : "The private game could not be opened.",
      );
    } finally {
      setLoading(false);
    }
  }, [circleId, gameId]);

  useEffect(() => {
    let active = true;
    void Promise.resolve().then(() => {
      if (active) void load();
    });
    return () => {
      active = false;
    };
  }, [load]);

  const yourIndex = game?.yourPlayer === "north" ? 1 : 0;
  const opponentIndex = yourIndex === 0 ? 1 : 0;
  const yourPits = useMemo(
    () =>
      game?.yourPlayer === "north" ? [11, 10, 9, 8, 7, 6] : [0, 1, 2, 3, 4, 5],
    [game?.yourPlayer],
  );
  const opponentPits = useMemo(
    () =>
      game?.yourPlayer === "north" ? [0, 1, 2, 3, 4, 5] : [11, 10, 9, 8, 7, 6],
    [game?.yourPlayer],
  );

  async function confirm() {
    if (!game || selected === null) return;
    command.current ??= `oware-move-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    setMoving(true);
    setError("");
    try {
      const result = await apiRequest<OwareGame>(
        `/v1/circles/${encodeURIComponent(circleId)}/oware/${encodeURIComponent(gameId)}/moves`,
        {
          method: "POST",
          headers: { "Idempotency-Key": command.current },
          body: JSON.stringify({
            pit: selected,
            expectedRevision: game.revision,
          }),
        },
      );
      setGame(result);
      setSelected(null);
      command.current = null;
    } catch (moveError) {
      setError(
        moveError instanceof Error
          ? moveError.message
          : "The move could not be accepted.",
      );
      command.current = null;
      await load();
    } finally {
      setMoving(false);
    }
  }

  const roomHref = circleId
    ? (`/fie/dan-mu/rooms/${circleId}` as Href)
    : ("/fie/adiwo" as Href);

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            onPress={() => router.push(roomHref)}
            style={styles.control}
          >
            <Text style={styles.controlText}>Private room</Text>
          </Pressable>
          <Text style={styles.reference}>{gameId.slice(0, 8)}</Text>
        </View>
        <Text style={styles.eyebrow}>OWARE · RETAINED ASYNC PLAY</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Take your time. Read the real board.
        </Text>
        <Text style={styles.body}>
          Every move is revision-checked and persisted. Player identities stay
          privacy-keyed, and skill never changes matching or visibility.
        </Text>

        {loading ? (
          <ActivityIndicator color="#9B315D" style={styles.loader} />
        ) : null}
        {error ? <Text style={styles.error}>{error}</Text> : null}
        {game ? (
          <>
            <View style={styles.status}>
              <Text style={styles.statusLabel}>
                {game.status !== "active"
                  ? game.status.toUpperCase()
                  : game.yourTurn
                    ? "YOUR MOVE"
                    : "OTHER PLAYER’S MOVE"}
              </Text>
              <Text style={styles.statusCopy}>
                Revision {game.revision} · window ends{" "}
                {new Date(game.moveDeadline).toLocaleString()}
              </Text>
            </View>

            <View style={styles.table}>
              <View style={styles.scoreRow}>
                <View style={styles.score}>
                  <Text style={styles.scoreLabel}>OTHER PLAYER</Text>
                  <Text style={styles.scoreValue}>
                    {game.captured[opponentIndex]}
                  </Text>
                </View>
                <View style={styles.score}>
                  <Text style={styles.scoreLabel}>YOU CAPTURED</Text>
                  <Text style={styles.scoreValue}>
                    {game.captured[yourIndex]}
                  </Text>
                </View>
              </View>
              <Text style={styles.tableEyebrow}>
                ABAPA RULES · SERVER VERIFIED
              </Text>
              <Text accessibilityRole="header" style={styles.tableTitle}>
                {game.yourTurn && game.status === "active"
                  ? "Choose one house, then confirm."
                  : game.status === "active"
                    ? "The board is waiting quietly."
                    : "This board is closed."}
              </Text>
              <Text style={styles.tableCopy}>
                The API verifies turn, feeding, capture, grand-slam, deadline
                and revision rules. This client never sows locally.
              </Text>
              <View accessibilityLabel="Oware board" style={styles.board}>
                <View
                  accessibilityLabel="Other player's houses"
                  style={styles.pitRow}
                >
                  {opponentPits.map((pit) => (
                    <View
                      key={`opponent-${pit}`}
                      style={[styles.pit, styles.opponentPit]}
                    >
                      <Text style={styles.pitSeeds}>{game.houses[pit]}</Text>
                    </View>
                  ))}
                </View>
                <View accessibilityLabel="Your houses" style={styles.pitRow}>
                  {yourPits.map((pit, index) => {
                    const disabled =
                      moving ||
                      !game.yourTurn ||
                      game.status !== "active" ||
                      game.houses[pit] === 0;
                    return (
                      <Pressable
                        accessibilityLabel={`House ${index + 1}, ${game.houses[pit]} seeds`}
                        accessibilityRole="button"
                        accessibilityState={{
                          disabled,
                          selected: selected === pit,
                        }}
                        disabled={disabled}
                        key={`you-${pit}`}
                        onPress={() => {
                          setSelected(pit);
                          command.current = null;
                        }}
                        style={[
                          styles.pit,
                          selected === pit && styles.selectedPit,
                          disabled && styles.pitDisabled,
                        ]}
                      >
                        <Text style={styles.pitSeeds}>{game.houses[pit]}</Text>
                      </Pressable>
                    );
                  })}
                </View>
              </View>
              <Text accessibilityLiveRegion="polite" style={styles.selection}>
                {game.status !== "active"
                  ? "No more moves can be submitted."
                  : !game.yourTurn
                    ? "Your move has been retained."
                    : selected === null
                      ? "No house selected."
                      : `House ${yourPits.indexOf(selected) + 1} selected.`}
              </Text>
              <Text style={styles.notation}>
                Current revision: {game.revision}
              </Text>
              <Pressable
                disabled={selected === null || moving || !game.yourTurn}
                onPress={() => void confirm()}
                style={[
                  styles.confirm,
                  (selected === null || moving || !game.yourTurn) &&
                    styles.confirmDisabled,
                ]}
              >
                <Text style={styles.confirmText}>
                  {moving ? "Verifying move…" : "Confirm one move"}
                </Text>
              </Pressable>
            </View>
          </>
        ) : null}
        <Text style={styles.privacy}>
          Game outcome never influences matching visibility or trust paths.
        </Text>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  control: {
    alignItems: "center",
    borderColor: "#8F7885",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 17,
  },
  controlText: { color: "#28161F", fontFamily: "Outfit_700Bold" },
  reference: {
    color: "#705C67",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1,
  },
  eyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
    marginTop: 52,
  },
  title: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 52,
    letterSpacing: -3.3,
    lineHeight: 49,
    marginTop: 14,
  },
  body: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 22,
  },
  loader: { marginTop: 30 },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 21,
    marginTop: 22,
  },
  status: {
    borderColor: "#C5A15C",
    borderRadius: 18,
    borderWidth: 1,
    marginTop: 28,
    padding: 18,
  },
  statusLabel: { color: "#6D244F", fontFamily: "Outfit_700Bold" },
  statusCopy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    marginTop: 5,
  },
  table: {
    backgroundColor: "#381829",
    borderRadius: 28,
    marginTop: 38,
    padding: 20,
  },
  scoreRow: { flexDirection: "row", gap: 8, justifyContent: "flex-end" },
  score: {
    borderColor: "rgba(255,243,230,.22)",
    borderRadius: 14,
    borderWidth: 1,
    padding: 12,
  },
  scoreLabel: {
    color: "rgba(255,243,230,.65)",
    fontFamily: "Outfit_700Bold",
    fontSize: 9,
  },
  scoreValue: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 25,
    marginTop: 5,
  },
  tableEyebrow: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
    marginTop: 36,
  },
  tableTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 37,
    letterSpacing: -2,
    lineHeight: 37,
    marginTop: 12,
  },
  tableCopy: {
    color: "rgba(255,243,230,.66)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 14,
  },
  board: {
    backgroundColor: "#874819",
    borderRadius: 24,
    gap: 8,
    marginHorizontal: -8,
    marginTop: 28,
    padding: 10,
  },
  pitRow: { flexDirection: "row", gap: 4 },
  pit: {
    alignItems: "center",
    aspectRatio: 1,
    backgroundColor: "#F7C97F",
    borderColor: "#6E3817",
    borderRadius: 999,
    borderWidth: 2,
    flex: 1,
    justifyContent: "center",
  },
  opponentPit: { backgroundColor: "#DBA45C" },
  selectedPit: { borderColor: "#FFF3E6", borderWidth: 4 },
  pitDisabled: { opacity: 0.55 },
  pitSeeds: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 17,
  },
  selection: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", marginTop: 24 },
  notation: {
    color: "rgba(255,243,230,.6)",
    fontFamily: "Outfit_400Regular",
    marginTop: 5,
  },
  confirm: {
    alignItems: "center",
    backgroundColor: "#FFAD3D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 22,
    minHeight: 52,
  },
  confirmDisabled: { opacity: 0.4 },
  confirmText: { color: "#28161F", fontFamily: "Outfit_700Bold" },
  privacy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 24,
  },
});
