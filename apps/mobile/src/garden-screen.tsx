import { apiRequest } from "./api";
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

type GardenSummary = {
  asOf: string;
  movingQuietly: number;
  sprouts: number;
  message: string;
};

export function GardenScreen() {
  const router = useRouter();
  const [summary, setSummary] = useState<GardenSummary | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    void apiRequest<GardenSummary>("/v1/garden")
      .then(setSummary)
      .catch((reason: unknown) =>
        setError(
          reason instanceof Error
            ? reason.message
            : "Your garden could not load.",
        ),
      );
  }, []);

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Fie</Text>
        </Pressable>
        <Text style={styles.eyebrow}>DAWN SUMMARY</Text>
        <Text accessibilityRole="header" style={styles.title}>
          {summary?.message ??
            (error
              ? "Your garden is resting."
              : "Listening for quiet movement.")}
        </Text>
        <Text style={styles.intro}>
          A once-a-day aggregate. No names, streaks, read receipts or pressure
          to act.
        </Text>
        {!summary && !error ? (
          <ActivityIndicator color="#9A3651" style={styles.loader} />
        ) : null}
        {error ? (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {error}
          </Text>
        ) : null}
        {summary ? (
          <View style={styles.grid}>
            <View style={styles.card}>
              <Text style={styles.count}>{summary.movingQuietly}</Text>
              <Text style={styles.label}>moving quietly</Text>
              <Text style={styles.note}>
                Queued, delivered or heard—without a read receipt.
              </Text>
            </View>
            <View style={styles.card}>
              <Text style={styles.count}>{summary.sprouts}</Text>
              <Text style={styles.label}>doorways ready</Text>
              <Text style={styles.note}>
                Mutual readiness, never public popularity.
              </Text>
            </View>
          </View>
        ) : null}
        <View style={styles.privacy}>
          <Text style={styles.privacyTitle}>Quiet by design</Text>
          <Text style={styles.privacyCopy}>
            Expired and declined seeds reveal nothing to others. This surface
            intentionally does not identify recipients or expose individual
            playback state.
          </Text>
          {summary ? (
            <Text style={styles.asOf}>
              Updated {new Date(summary.asOf).toLocaleString()}
            </Text>
          ) : null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F8F2EE", flex: 1 },
  content: { padding: 20, paddingBottom: 48 },
  back: {
    alignItems: "center",
    alignSelf: "flex-start",
    backgroundColor: "#FFFFFF",
    borderRadius: 999,
    justifyContent: "center",
    minHeight: 48,
    minWidth: 64,
  },
  backText: { color: "#2A1022", fontFamily: "Outfit_700Bold" },
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
  loader: { marginVertical: 50 },
  error: {
    color: "#9A3651",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 22,
    marginTop: 24,
  },
  grid: { gap: 10, marginTop: 32 },
  card: {
    backgroundColor: "#FFFFFF",
    borderRadius: 22,
    minHeight: 180,
    padding: 24,
  },
  count: {
    color: "#0F7059",
    fontFamily: "Outfit_900Black",
    fontSize: 58,
    letterSpacing: -3,
  },
  label: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 24,
  },
  note: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    lineHeight: 21,
    marginTop: 9,
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
  asOf: {
    color: "#FFB44F",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
    marginTop: 18,
  },
});
