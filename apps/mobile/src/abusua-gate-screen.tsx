import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const choices = [
  ["The first thread", true],
  ["Theme one reflection", false],
  ["Care promises", false],
] as const;
export function AbusuaGateScreen() {
  const router = useRouter();
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie/dan-mu" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Dan mu</Text>
        </Pressable>
        <Text style={styles.eyebrow}>ABUSUA GATE</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Open one careful window.
        </Text>
        <Text style={styles.copy}>
          Invite one trusted person to review only what you both choose. Either
          of you can close the gate.
        </Text>
        <Text style={styles.step}>01 · CHOOSE THE MATERIAL</Text>
        {choices.map(([title, selected]) => (
          <View
            key={title}
            style={[styles.choice, selected && styles.selected]}
          >
            <Text style={[styles.choiceMark, selected && styles.selectedText]}>
              {selected ? "✓" : "+"}
            </Text>
            <Text style={[styles.choiceText, selected && styles.selectedText]}>
              {title}
            </Text>
          </View>
        ))}
        <Text style={styles.step}>02 · BOTH HANDS ON THE LATCH</Text>
        <View style={styles.consent}>
          <Text style={styles.consentName}>You</Text>
          <Text style={styles.ready}>Consent given</Text>
        </View>
        <View style={styles.consent}>
          <Text style={styles.consentName}>Ama</Text>
          <Text style={styles.waiting}>Waiting privately</Text>
        </View>
        <View style={styles.passage}>
          <Text style={styles.passageEyebrow}>REVIEWER PASSAGE</Text>
          <Text style={styles.passageTitle}>A link is not enough.</Text>
          <Text style={styles.passageCopy}>
            Both consents create a 24-hour invite and a separately delivered
            10-minute one-time code. Every view is watermarked.
          </Text>
          <Pressable
            accessibilityState={{ disabled: true }}
            disabled
            style={styles.disabled}
          >
            <Text style={styles.disabledText}>Waiting for both consents</Text>
          </Pressable>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
const styles = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { padding: 20, paddingBottom: 60 },
  back: {
    alignItems: "center",
    borderColor: "#9F8793",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 18,
    alignSelf: "flex-start",
  },
  backText: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
    marginTop: 54,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 58,
    letterSpacing: -3.8,
    lineHeight: 52,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginTop: 24,
  },
  step: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
    marginBottom: 10,
    marginTop: 42,
  },
  choice: {
    alignItems: "center",
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 18,
    borderWidth: 1,
    flexDirection: "row",
    gap: 14,
    marginTop: 8,
    minHeight: 70,
    padding: 16,
  },
  selected: { backgroundColor: "#38172C", borderColor: "#38172C" },
  choiceMark: { color: "#2B151F", fontFamily: "Outfit_700Bold", fontSize: 20 },
  choiceText: { color: "#2B151F", fontFamily: "Outfit_700Bold", fontSize: 16 },
  selectedText: { color: "#FFF5E9" },
  consent: {
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 18,
    borderWidth: 1,
    gap: 8,
    marginTop: 8,
    padding: 18,
  },
  consentName: { color: "#2B151F", fontFamily: "Outfit_700Bold", fontSize: 18 },
  ready: { color: "#27755F", fontFamily: "Outfit_700Bold" },
  waiting: { color: "#69535D", fontFamily: "Outfit_600SemiBold" },
  passage: {
    backgroundColor: "#38172C",
    borderRadius: 24,
    marginTop: 42,
    padding: 24,
  },
  passageEyebrow: {
    color: "#FF9AB0",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  passageTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  passageCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 24,
    marginTop: 16,
  },
  disabled: {
    alignItems: "center",
    backgroundColor: "#FFB34F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 24,
    minHeight: 52,
    opacity: 0.45,
  },
  disabledText: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
});
