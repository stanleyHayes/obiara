import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const answers = [
  "One person cannot make every decision alone.",
  "A journey should always begin before dawn.",
  "Silence is the only sign of wisdom.",
] as const;

export function EbeScreen() {
  const router = useRouter();
  const [selected, setSelected] = useState<string | null>(null);
  const [stage, setStage] = useState<"answering" | "waiting" | "revealed">(
    "answering",
  );
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
        <Text style={styles.eyebrow}>ƐBƐ · REVIEWED PROVERB PACK</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Listen for the wisdom between the words.
        </Text>
        <Text style={styles.body}>
          No timer, public score or matching advantage. Both answers unfold
          together.
        </Text>
        <View style={styles.card}>
          <View style={styles.tags}>
            <Text style={styles.tag}>Twi · Greater Accra</Text>
            <Text style={styles.tag}>Reviewed · revision 04</Text>
          </View>
          <Text style={styles.cardEyebrow}>ROUND ONE OF THREE</Text>
          <Text accessibilityRole="header" style={styles.proverb}>
            “Tikoro nko agyina.”
          </Text>
          <Text style={styles.prompt}>
            Which reflection sits closest to this proverb?
          </Text>
          {answers.map((answer) => (
            <Pressable
              accessibilityRole="radio"
              accessibilityState={{
                checked: selected === answer,
                disabled: stage !== "answering",
              }}
              disabled={stage !== "answering"}
              key={answer}
              onPress={() => setSelected(answer)}
              style={[
                styles.answer,
                selected === answer && styles.answerSelected,
              ]}
            >
              <Text style={styles.answerText}>{answer}</Text>
            </Pressable>
          ))}
          <Text accessibilityLiveRegion="polite" style={styles.status}>
            {stage === "answering"
              ? selected
                ? "Your reflection is ready."
                : "Nothing selected yet."
              : stage === "waiting"
                ? "Your answer is folded. Ama’s is ready."
                : "Both reflections are open."}
          </Text>
          {stage === "answering" ? (
            <Pressable
              disabled={!selected}
              onPress={() => setStage("waiting")}
              style={[styles.primary, !selected && styles.disabled]}
            >
              <Text style={styles.primaryText}>Fold my answer</Text>
            </Pressable>
          ) : null}
          {stage === "waiting" ? (
            <Pressable
              onPress={() => setStage("revealed")}
              style={styles.primary}
            >
              <Text style={styles.primaryText}>Reveal together</Text>
            </Pressable>
          ) : null}
          {stage === "revealed" ? (
            <View style={styles.reveal}>
              <Text style={styles.revealLabel}>YOU AND AMA CHOSE</Text>
              <Text style={styles.revealAnswer}>{answers[0]}</Text>
              <Text style={styles.context}>
                Reviewed context: shared counsel can see beyond one person’s
                view. This is a learning note, not a measure of character.
              </Text>
            </View>
          ) : null}
        </View>
        <Text style={styles.privacy}>
          Reviewed cultural context stays versioned and attributable.
        </Text>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#20101A", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  topbar: { flexDirection: "row", justifyContent: "space-between" },
  control: {
    alignItems: "center",
    borderColor: "rgba(255,243,230,.35)",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 17,
  },
  controlText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
    marginTop: 52,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 54,
    letterSpacing: -3.5,
    lineHeight: 50,
    marginTop: 14,
  },
  body: {
    color: "rgba(255,243,230,.64)",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 22,
  },
  card: {
    backgroundColor: "#FFF0D9",
    borderRadius: 28,
    marginTop: 38,
    padding: 22,
  },
  tags: { flexDirection: "row", flexWrap: "wrap", gap: 6 },
  tag: {
    borderColor: "#CDB9C3",
    borderRadius: 999,
    borderWidth: 1,
    color: "#705C67",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 10,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  cardEyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
    marginTop: 40,
  },
  proverb: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 44,
    letterSpacing: -2.5,
    lineHeight: 44,
    marginTop: 12,
  },
  prompt: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginBottom: 18,
    marginTop: 16,
  },
  answer: {
    borderColor: "#CDB9C3",
    borderRadius: 15,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 8,
    minHeight: 58,
    padding: 14,
  },
  answerSelected: { backgroundColor: "#F1DCE6", borderColor: "#6D244F" },
  answerText: { color: "#2A1022", fontFamily: "Outfit_600SemiBold" },
  status: {
    color: "#2A1022",
    fontFamily: "Outfit_700Bold",
    marginTop: 26,
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#6D244F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 18,
    minHeight: 52,
  },
  primaryText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.4 },
  reveal: {
    backgroundColor: "#2A1022",
    borderRadius: 18,
    marginTop: 22,
    padding: 20,
  },
  revealLabel: {
    color: "#FFB7C4",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  revealAnswer: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 20,
    lineHeight: 28,
    marginTop: 12,
  },
  context: {
    color: "rgba(255,243,230,.68)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 18,
  },
  privacy: {
    color: "rgba(255,243,230,.6)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 24,
  },
});
