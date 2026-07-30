import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const games = [
  ["Oware", "Available in an exact two-member circle", true],
  ["Ɛbɛ", "Available with approved catalog content", true],
  ["Anansesɛm", "Available in an exact two-member circle", true],
  ["Ampe", "Available in an exact two-member circle", true],
] as const;

export function GamesHallScreen() {
  const router = useRouter();
  const [cohortRef, setCohortRef] = useState("");
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            onPress={() => router.push("/fie" as Href)}
            style={styles.control}
          >
            <Text style={styles.controlText}>Fie</Text>
          </Pressable>
          <Pressable style={styles.control}>
            <Text style={styles.controlText}>Safety</Text>
          </Pressable>
        </View>
        <Text style={styles.eyebrow}>PLAY REVEALS MOMENTS—NOT WORTH</Text>
        <Text accessibilityRole="header" style={styles.title}>
          A hall for skill, wit and shared stories.
        </Text>
        <Text style={styles.body}>
          Games stay separate from matching visibility. No global popularity
          rank or pay-to-win path.
        </Text>
        <Text accessibilityRole="header" style={styles.sectionTitle}>
          Your games.
        </Text>
        <View style={styles.gameGrid}>
          {games.map(([name, type, available]) => (
            <Pressable
              accessibilityState={{ disabled: !available }}
              disabled={!available}
              key={name}
              onPress={() => router.push("/fie/adiwo" as Href)}
              style={[styles.gameCard, !available && styles.disabled]}
            >
              <Text style={styles.gameType}>{type.toUpperCase()}</Text>
              <Text style={styles.gameName}>{name}</Text>
              <Text style={styles.gameOpen}>
                {available ? "Choose a private circle →" : "Unavailable"}
              </Text>
            </Pressable>
          ))}
        </View>
        <View style={styles.tournament}>
          <Text style={styles.cardEyebrow}>SERVER-AUTHORITATIVE PLAY</Text>
          <Text accessibilityRole="header" style={styles.cardTitle}>
            No invented games or tournament seats.
          </Text>
          <Text style={styles.cardCopy}>
            Oware, Anansesɛm, Ampe and reviewed-catalog Ɛbɛ begin inside a
            retained two-member circle. Competition and conduct review stay
            absent until their persistence and authority adapters are composed.
          </Text>
          <TextInput
            accessibilityLabel="Private competition cohort reference"
            autoCapitalize="none"
            onChangeText={setCohortRef}
            placeholder="cohort_…"
            placeholderTextColor="#8B7780"
            style={styles.cohortInput}
            value={cohortRef}
          />
          <Pressable
            disabled={!cohortRef.trim()}
            onPress={() =>
              router.push(
                `/fie/games/competition/${encodeURIComponent(cohortRef.trim())}` as Href,
              )
            }
            style={[styles.cohortButton, !cohortRef.trim() && styles.disabled]}
          >
            <Text style={styles.cohortButtonText}>Open private cohort</Text>
          </Pressable>
        </View>
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
    paddingHorizontal: 18,
  },
  controlText: { color: "#28161F", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
    marginTop: 52,
  },
  title: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 52,
    letterSpacing: -3.3,
    lineHeight: 48,
    marginTop: 14,
  },
  body: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 22,
  },
  sectionTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 40,
    letterSpacing: -2.2,
    marginTop: 46,
  },
  gameGrid: { gap: 10, marginTop: 18 },
  gameCard: {
    backgroundColor: "#28161F",
    borderRadius: 20,
    minHeight: 170,
    padding: 20,
  },
  gameType: {
    color: "#FFB7C4",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  gameName: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 32,
    letterSpacing: -1.4,
    marginTop: 28,
  },
  gameOpen: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    marginTop: "auto",
  },
  tournament: {
    backgroundColor: "#FFF0D9",
    borderRadius: 26,
    marginTop: 38,
    padding: 22,
  },
  cohortInput: {
    backgroundColor: "#FFF8EE",
    borderColor: "#CDB9C3",
    borderRadius: 14,
    borderWidth: 1,
    color: "#28161F",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    marginTop: 18,
    minHeight: 52,
    paddingHorizontal: 16,
  },
  cohortButton: {
    alignItems: "center",
    backgroundColor: "#9B315D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 12,
    minHeight: 50,
  },
  cohortButtonText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  cardEyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  cardTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 40,
    letterSpacing: -2.1,
    lineHeight: 40,
    marginTop: 12,
  },
  cardCopy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 14,
  },
  facts: { flexDirection: "row", flexWrap: "wrap", gap: 6, marginTop: 20 },
  fact: {
    borderColor: "#CAB5C0",
    borderRadius: 999,
    borderWidth: 1,
    color: "#28161F",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 10,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#6D244F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 24,
    minHeight: 52,
  },
  primaryText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.65 },
  review: {
    backgroundColor: "#28161F",
    borderRadius: 26,
    marginTop: 28,
    padding: 22,
  },
  reviewEyebrow: {
    color: "#FFB7C4",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  reviewTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  reviewLabel: {
    color: "#FFB7C4",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
    marginTop: 22,
  },
  reviewCopy: {
    color: "rgba(255,243,230,.65)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 14,
  },
  reviewButton: {
    alignItems: "center",
    backgroundColor: "#FFAD3D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 20,
    minHeight: 52,
  },
  reviewButtonText: { color: "#28161F", fontFamily: "Outfit_700Bold" },
});
