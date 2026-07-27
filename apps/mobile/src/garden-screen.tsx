import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const seeds = [
  {
    person: "Ama",
    state: "Sprouted",
    note: "A mutual doorway can begin",
    tone: "#0F7059",
  },
  {
    person: "Kwesi",
    state: "Heard",
    note: "Listened to; no reply promised",
    tone: "#9A4B00",
  },
  {
    person: "Nana",
    state: "Delivered",
    note: "Available to hear privately",
    tone: "#315F8A",
  },
  {
    person: "Efua",
    state: "Returned to earth",
    note: "Closed without a public signal",
    tone: "#866D80",
  },
] as const;

export function GardenScreen() {
  const router = useRouter();
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            accessibilityRole="button"
            onPress={() => router.push("/fie" as Href)}
            style={styles.back}
          >
            <Text style={styles.backText}>Fie</Text>
          </Pressable>
          <View style={styles.allowance}>
            <Text style={styles.allowanceCount}>4</Text>
            <Text style={styles.allowanceCopy}>seeds remain</Text>
          </View>
        </View>
        <Text style={styles.eyebrow}>DAWN SUMMARY</Text>
        <Text accessibilityRole="header" style={styles.title}>
          One doorway is ready when you are.
        </Text>
        <Text style={styles.intro}>
          A once-a-day view. No streaks, read receipts or pressure to act.
        </Text>
        <View style={styles.list}>
          {seeds.map((seed) => (
            <View key={seed.person} style={styles.card}>
              <View style={styles.stateRow}>
                <View style={[styles.dot, { backgroundColor: seed.tone }]} />
                <Text style={styles.state}>{seed.state}</Text>
              </View>
              <Text style={styles.person}>{seed.person}</Text>
              <Text style={styles.note}>{seed.note}</Text>
            </View>
          ))}
        </View>
        <View style={styles.privacy}>
          <Text style={styles.privacyTitle}>Quiet by design</Text>
          <Text style={styles.privacyCopy}>
            Expired seeds reveal nothing to others. A closed seed is never a
            public signal.
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F8F2EE", flex: 1 },
  content: { padding: 20, paddingBottom: 48 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  back: {
    alignItems: "center",
    backgroundColor: "#FFFFFF",
    borderRadius: 999,
    justifyContent: "center",
    minHeight: 48,
    minWidth: 64,
  },
  backText: { color: "#2A1022", fontFamily: "Outfit_700Bold" },
  allowance: {
    alignItems: "center",
    backgroundColor: "#0F7059",
    borderRadius: 18,
    paddingHorizontal: 18,
    paddingVertical: 10,
  },
  allowanceCount: {
    color: "#FFFFFF",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 24,
  },
  allowanceCopy: {
    color: "#EEFBF7",
    fontFamily: "Outfit_500Medium",
    fontSize: 10,
  },
  eyebrow: {
    color: "#9A3651",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
    marginTop: 42,
  },
  title: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 48,
    letterSpacing: -3,
    lineHeight: 44,
    marginTop: 14,
  },
  intro: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 22,
  },
  list: { gap: 10, marginTop: 32 },
  card: {
    backgroundColor: "#FFFFFF",
    borderRadius: 20,
    minHeight: 178,
    padding: 22,
  },
  stateRow: { alignItems: "center", flexDirection: "row", gap: 9 },
  dot: { borderRadius: 5, height: 10, width: 10 },
  state: { color: "#4C3446", fontFamily: "Outfit_600SemiBold", fontSize: 13 },
  person: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.5,
    marginTop: 24,
  },
  note: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 14,
    lineHeight: 21,
    marginTop: 6,
  },
  privacy: {
    backgroundColor: "#2A1022",
    borderRadius: 22,
    marginTop: 12,
    padding: 24,
  },
  privacyTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 20,
  },
  privacyCopy: {
    color: "rgba(255,243,230,0.68)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 8,
  },
});
