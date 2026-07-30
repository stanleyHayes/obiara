import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const capabilities = [
  [
    "Guided help",
    "Member-invoked navigation and wording help only. No private-source access or decision authority.",
  ],
  [
    "Introduction explanations",
    "Only retained, mutually permitted reasons may appear. No hidden score or learned ranker is composed.",
  ],
  [
    "Human review",
    "Okyeame appeal intake is not composed. The authenticated Suban route owns Suban appeals.",
  ],
] as const;

export function AiAccountabilityScreen() {
  const router = useRouter();
  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.intro}>
          <Text style={styles.kicker}>AI ACCOUNTABILITY</Text>
          <Text accessibilityRole="header" style={styles.title}>
            Capability without invented certainty.
          </Text>
          <Text style={styles.body}>
            These are enforced product boundaries, not live model evaluations,
            certifications or compatibility scores.
          </Text>
        </View>
        <View style={styles.cards}>
          {capabilities.map(([title, detail]) => (
            <View key={title} style={styles.card}>
              <Text style={styles.version}>RUNTIME BOUNDARY</Text>
              <Text accessibilityRole="header" style={styles.cardTitle}>
                {title}
              </Text>
              <Text style={styles.status}>Fail closed</Text>
              <Text style={styles.detail}>{detail}</Text>
            </View>
          ))}
        </View>
        <View style={styles.appeal}>
          <Text style={styles.kicker}>AVAILABLE REVIEW PATH</Text>
          <Text accessibilityRole="header" style={styles.appealTitle}>
            Suban decisions have a real appeal route.
          </Text>
          <Text style={styles.appealBody}>
            No local reference is generated here. Open the authenticated Suban
            explanation to submit a retained appeal.
          </Text>
          <Pressable
            onPress={() => router.push("/fie/settings/suban" as Href)}
            style={styles.button}
          >
            <Text style={styles.buttonText}>Open Suban explanation</Text>
          </Pressable>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { backgroundColor: "#F8F2EE", flex: 1 },
  content: { padding: 18, paddingBottom: 48 },
  intro: { backgroundColor: "#2A1022", borderRadius: 26, padding: 30 },
  kicker: {
    color: "#D04A69",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 48,
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
  version: { color: "#BD3153", fontFamily: "Outfit_700Bold", fontSize: 11 },
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
  button: {
    alignItems: "center",
    backgroundColor: "#2A1022",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 22,
    minHeight: 48,
    paddingHorizontal: 18,
  },
  buttonText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", fontSize: 14 },
});
