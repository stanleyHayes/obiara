import { fieRoutes, type FieRouteId } from "@obiara/fie-routing";
import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const copy: Record<
  FieRouteId,
  { eyebrow: string; title: string; body: string }
> = {
  home: {
    eyebrow: "YOUR COMPOUND",
    title: "Akwaaba home.",
    body: "Four places, each with its own pace, privacy and purpose.",
  },
  welcome: {
    eyebrow: "FIRST WALK",
    title: "Find your place in Fie.",
    body: "This walk is skippable and never changes your account standing.",
  },
  abonten: {
    eyebrow: "THE PUBLIC STREET",
    title: "Step outside. Stay yourself.",
    body: "Community fires, learning and notices, with no romantic initiation.",
  },
  adiwo: {
    eyebrow: "THE COURTYARD",
    title: "Familiar people. Shared purpose.",
    body: "Your circles and host-led moments, protected by membership.",
  },
  "epono-ano": {
    eyebrow: "THE DOORWAY",
    title: "Pause before you open.",
    body: "Tier 1 introductions are bounded, voice-first and deliberate.",
  },
  "dan-mu": {
    eyebrow: "THE INNER ROOM",
    title: "Private means private.",
    body: "Tier 2 mutual rooms have no public activity or popularity signal.",
  },
  garden: {
    eyebrow: "YOUR SEED GARDEN",
    title: "Sow with intention.",
    body: "Listen first, speak in your own voice, and spend only after server confirmation.",
  },
  okyeame: {
    eyebrow: "A CAPABILITY, NOT A PERSON",
    title: "Help should know its place.",
    body: "Okyeame is resting. Your access to Fie is unchanged.",
  },
};

const primaryRouteIds: readonly FieRouteId[] = [
  "home",
  "abonten",
  "adiwo",
  "epono-ano",
  "dan-mu",
];

export function FieScreen({ routeId }: Readonly<{ routeId: FieRouteId }>) {
  const router = useRouter();
  const current = fieRoutes.find((route) => route.id === routeId)!;
  const content = copy[routeId];

  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.topbar}>
          <Text style={styles.wordmark}>obiara</Text>
          <View style={styles.connection}>
            <View style={styles.connectionDot} />
            <Text style={styles.connectionText}>Connection saver</Text>
          </View>
        </View>
        <View style={styles.hero}>
          <Text style={styles.eyebrow}>{content.eyebrow}</Text>
          <Text accessibilityRole="header" style={styles.title}>
            {content.title}
          </Text>
          <Text style={styles.body}>{content.body}</Text>
          <View style={styles.boundary}>
            <Text style={styles.boundaryLabel}>{current.label}</Text>
            <Text style={styles.boundaryCopy}>
              {current.gloss} · Tier {current.minimumTier}
            </Text>
          </View>
        </View>
      </ScrollView>
      <View accessibilityRole="tablist" style={styles.navigation}>
        {primaryRouteIds.map((id) => {
          const route = fieRoutes.find((candidate) => candidate.id === id)!;
          const selected = route.id === routeId;
          return (
            <Pressable
              accessibilityLabel={`${route.label}, ${route.gloss}`}
              accessibilityRole="tab"
              accessibilityState={{ selected }}
              key={route.id}
              onPress={() => router.push(route.expoPath as Href)}
              style={({ pressed }) => [
                styles.navigationItem,
                selected && styles.navigationItemSelected,
                pressed && styles.pressed,
              ]}
            >
              <Text
                numberOfLines={1}
                style={[
                  styles.navigationLabel,
                  selected && styles.navigationLabelSelected,
                ]}
              >
                {route.label}
              </Text>
              <Text style={styles.navigationGloss}>{route.gloss}</Text>
            </Pressable>
          );
        })}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { backgroundColor: "#F8F2EE", flex: 1 },
  content: { padding: 18, paddingBottom: 112 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 64,
  },
  wordmark: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 22,
    letterSpacing: -1.2,
  },
  connection: {
    alignItems: "center",
    backgroundColor: "#FFFFFF",
    borderRadius: 999,
    flexDirection: "row",
    gap: 8,
    minHeight: 48,
    paddingHorizontal: 14,
  },
  connectionDot: {
    backgroundColor: "#FF9F1C",
    borderRadius: 4,
    height: 8,
    width: 8,
  },
  connectionText: {
    color: "#2A1022",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
  },
  hero: {
    backgroundColor: "#2A1022",
    borderRadius: 26,
    marginTop: 20,
    minHeight: 620,
    padding: 34,
  },
  eyebrow: {
    color: "#FF6481",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 58,
    letterSpacing: -3.4,
    lineHeight: 52,
    marginTop: 48,
  },
  body: {
    color: "rgba(255, 243, 230, 0.68)",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginTop: 28,
  },
  boundary: {
    borderColor: "rgba(255, 243, 230, 0.18)",
    borderRadius: 18,
    borderWidth: 1,
    marginTop: "auto",
    padding: 18,
  },
  boundaryLabel: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 18,
  },
  boundaryCopy: {
    color: "rgba(255, 243, 230, 0.55)",
    fontFamily: "Outfit_400Regular",
    marginTop: 5,
  },
  navigation: {
    backgroundColor: "#200B19",
    bottom: 0,
    flexDirection: "row",
    left: 0,
    paddingBottom: 8,
    paddingHorizontal: 8,
    paddingTop: 8,
    position: "absolute",
    right: 0,
  },
  navigationItem: {
    alignItems: "center",
    borderRadius: 12,
    flex: 1,
    justifyContent: "center",
    minHeight: 58,
    paddingHorizontal: 3,
  },
  navigationItemSelected: { backgroundColor: "rgba(255, 243, 230, 0.1)" },
  navigationLabel: {
    color: "rgba(255, 243, 230, 0.62)",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 10,
  },
  navigationLabelSelected: { color: "#FFF3E6" },
  navigationGloss: {
    color: "#FFB44F",
    fontFamily: "Outfit_400Regular",
    fontSize: 8,
    marginTop: 3,
  },
  pressed: { opacity: 0.72 },
});
