import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const capabilities = [
  {
    id: "okyeame-help",
    title: "Okyeame guided help",
    version: "policy-1.0",
    status: "Available within limits",
    detail: "Member-invoked help only. Prompt retention is off.",
  },
  {
    id: "resonance-explanation",
    title: "Introduction explanations",
    version: "rules-1.0",
    status: "Rules only",
    detail: "Only features enabled by both members may appear.",
  },
  {
    id: "matching-ranker",
    title: "Learned matching ranker",
    version: "not-released",
    status: "Not released",
    detail: "Offline fairness and human approval gates remain closed.",
  },
] as const;

export function AiAccountabilityScreen() {
  const [appealReference, setAppealReference] = useState<string | null>(null);

  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.intro}>
          <Text style={styles.kicker}>AI ACCOUNTABILITY</Text>
          <Text accessibilityRole="header" style={styles.title}>
            See what is on, limited or paused.
          </Text>
          <Text style={styles.body}>
            These are capability boundaries, not scores or promises of perfect
            safety.
          </Text>
        </View>

        <View style={styles.cards}>
          {capabilities.map((capability) => (
            <View key={capability.id} style={styles.card}>
              <Text style={styles.version}>{capability.version}</Text>
              <Text accessibilityRole="header" style={styles.cardTitle}>
                {capability.title}
              </Text>
              <Text style={styles.status}>{capability.status}</Text>
              <Text style={styles.detail}>{capability.detail}</Text>
              <Pressable
                accessibilityRole="button"
                disabled={appealReference !== null}
                onPress={() =>
                  setAppealReference(`appeal-${capability.id}`)
                }
                style={({ pressed }) => [
                  styles.button,
                  appealReference !== null && styles.buttonDisabled,
                  pressed && styles.pressed,
                ]}
              >
                <Text style={styles.buttonText}>Ask for human review</Text>
              </Pressable>
            </View>
          ))}
        </View>

        <View accessibilityLiveRegion="polite" style={styles.appeal}>
          <Text style={styles.kicker}>APPEAL PATH</Text>
          <Text accessibilityRole="header" style={styles.appealTitle}>
            {appealReference
              ? "Your request is with a person."
              : "A person reviews every appeal."}
          </Text>
          <Text style={styles.appealBody}>
            {appealReference
              ? `Reference ${appealReference}. No model decides the outcome.`
              : "Choose a capability to request review. No model decides the appeal."}
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { backgroundColor: "#F8F2EE", flex: 1 },
  content: { padding: 18, paddingBottom: 48 },
  intro: {
    backgroundColor: "#2A1022",
    borderRadius: 26,
    padding: 30,
  },
  kicker: {
    color: "#D04A69",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 50,
    letterSpacing: -2.8,
    lineHeight: 47,
    marginTop: 46,
  },
  body: {
    color: "rgba(255, 243, 230, 0.72)",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 34,
  },
  cards: { gap: 12, marginTop: 22 },
  card: {
    backgroundColor: "#FFFFFF",
    borderColor: "rgba(58, 14, 46, 0.12)",
    borderRadius: 18,
    borderWidth: 2,
    padding: 22,
  },
  version: {
    color: "#BD3153",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
  },
  cardTitle: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 28,
    letterSpacing: -1.2,
    lineHeight: 30,
    marginTop: 10,
  },
  status: {
    alignSelf: "flex-start",
    backgroundColor: "#F9E7EC",
    borderRadius: 999,
    color: "#711D38",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    marginTop: 18,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  detail: {
    color: "#634F5C",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 23,
    marginTop: 18,
  },
  button: {
    alignItems: "center",
    backgroundColor: "#2A1022",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 22,
    minHeight: 48,
    paddingHorizontal: 18,
  },
  buttonDisabled: { opacity: 0.48 },
  buttonText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
  },
  pressed: { opacity: 0.72 },
  appeal: {
    backgroundColor: "#FFEAD0",
    borderRadius: 18,
    marginTop: 22,
    padding: 24,
  },
  appealTitle: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.3,
    lineHeight: 32,
    marginTop: 14,
  },
  appealBody: {
    color: "#634F5C",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 23,
    marginTop: 16,
  },
});
