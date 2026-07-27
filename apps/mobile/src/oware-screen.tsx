import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

function sow(pits: readonly number[], origin: number) {
  const next = [...pits];
  let seeds = next[origin];
  let cursor = origin;
  next[origin] = 0;
  while (seeds > 0) {
    cursor = (cursor + 1) % next.length;
    if (cursor === origin) continue;
    next[cursor] += 1;
    seeds -= 1;
  }
  return next;
}

export function OwareScreen() {
  const router = useRouter();
  const [pits, setPits] = useState(() => Array.from({ length: 12 }, () => 4));
  const [selected, setSelected] = useState<number | null>(null);
  const [sent, setSent] = useState(false);
  const confirm = () => {
    if (selected === null || sent) return;
    setPits((current) => sow(current, selected));
    setSelected(null);
    setSent(true);
  };
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            onPress={() =>
              router.push("/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa" as Href)
            }
            style={styles.control}
          >
            <Text style={styles.controlText}>Private room</Text>
          </Pressable>
          <Pressable style={styles.control}>
            <Text style={styles.controlText}>Safety</Text>
          </Pressable>
        </View>
        <Text style={styles.eyebrow}>OWARE · ASYNC PLAY</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Take your time. Read the board.
        </Text>
        <Text style={styles.body}>
          A thoughtful game for two. Skill stays here—it never changes who sees
          you or how you are matched.
        </Text>
        <View style={styles.status}>
          <Text style={styles.statusLabel}>
            {sent ? "AMA’S TURN" : "YOUR TURN"}
          </Text>
          <Text style={styles.statusCopy}>
            18h 42m remains · no streak pressure
          </Text>
        </View>

        <View style={styles.table}>
          <View style={styles.scoreRow}>
            <View style={styles.score}>
              <Text style={styles.scoreLabel}>AMA CAPTURED</Text>
              <Text style={styles.scoreValue}>0</Text>
            </View>
            <View style={styles.score}>
              <Text style={styles.scoreLabel}>YOU CAPTURED</Text>
              <Text style={styles.scoreValue}>0</Text>
            </View>
          </View>
          <Text style={styles.tableEyebrow}>ABAPA RULES · 48 SEEDS</Text>
          <Text accessibilityRole="header" style={styles.tableTitle}>
            Choose one house, then confirm.
          </Text>
          <Text style={styles.tableCopy}>
            Your six houses are nearest you. The server verifies feeding,
            capture and grand-slam rules.
          </Text>
          <View accessibilityLabel="Oware board" style={styles.board}>
            <View accessibilityLabel="Ama's houses" style={styles.pitRow}>
              {pits
                .slice(6)
                .reverse()
                .map((seeds, index) => (
                  <View
                    key={`ama-${index}`}
                    style={[styles.pit, styles.amaPit]}
                  >
                    <Text style={styles.pitSeeds}>{seeds}</Text>
                  </View>
                ))}
            </View>
            <View accessibilityLabel="Your houses" style={styles.pitRow}>
              {pits.slice(0, 6).map((seeds, pit) => (
                <Pressable
                  accessibilityLabel={`House ${pit + 1}, ${seeds} seeds`}
                  accessibilityRole="button"
                  accessibilityState={{
                    disabled: sent,
                    selected: selected === pit,
                  }}
                  disabled={sent}
                  key={`you-${pit}`}
                  onPress={() => setSelected(pit)}
                  style={[styles.pit, selected === pit && styles.selectedPit]}
                >
                  <Text style={styles.pitSeeds}>{seeds}</Text>
                </Pressable>
              ))}
            </View>
          </View>
          <Text accessibilityLiveRegion="polite" style={styles.selection}>
            {sent
              ? "Move sent. The board is with Ama."
              : selected === null
                ? "No house selected."
                : `House ${selected + 1} selected.`}
          </Text>
          <Text style={styles.notation}>
            Notation: 18. 3C · awaiting next move
          </Text>
          <Pressable
            disabled={selected === null}
            onPress={confirm}
            style={[
              styles.confirm,
              selected === null && styles.confirmDisabled,
            ]}
          >
            <Text style={styles.confirmText}>Confirm one move</Text>
          </Pressable>
        </View>
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
  topbar: { flexDirection: "row", justifyContent: "space-between" },
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
    fontSize: 56,
    letterSpacing: -3.5,
    lineHeight: 51,
    marginTop: 14,
  },
  body: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
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
  amaPit: { backgroundColor: "#DBA45C" },
  selectedPit: { borderColor: "#FFF3E6", borderWidth: 4 },
  pitSeeds: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 17,
  },
  selection: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    marginTop: 24,
  },
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
