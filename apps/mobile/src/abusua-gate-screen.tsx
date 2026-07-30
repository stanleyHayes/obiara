import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const requirements = [
  "A retained private-room record",
  "A bounded material selection",
  "Current consent from both room members",
  "A separately delivered reviewer authority",
  "Expiry, revocation and audited access",
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
          A gate opens only on retained consent.
        </Text>
        <Text style={styles.copy}>
          Reviewer access is not composed in the mobile runtime. This screen
          does not invent a room, material, person, consent or one-time code.
        </Text>
        <View style={styles.passage}>
          <Text style={styles.passageEyebrow}>REQUIRED AUTHORITY CHAIN</Text>
          <Text style={styles.passageTitle}>A link is never enough.</Text>
          {requirements.map((requirement, index) => (
            <Text key={requirement} style={styles.passageCopy}>
              {index + 1}. {requirement}
            </Text>
          ))}
          <View accessibilityState={{ disabled: true }} style={styles.disabled}>
            <Text style={styles.disabledText}>Reviewer access unavailable</Text>
          </View>
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
    alignSelf: "flex-start",
    borderColor: "#9F8793",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 18,
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
    fontSize: 54,
    letterSpacing: -3.5,
    lineHeight: 51,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginTop: 24,
  },
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
