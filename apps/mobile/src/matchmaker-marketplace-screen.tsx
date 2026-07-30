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

import { apiRequest } from "./api";

type Profile = {
  matchmakerId: string;
  displayName: string;
  licenseId: string;
  licenseValidUntil: string;
  minimumFeePesewas: number;
  maximumFeePesewas: number;
  languages: string[];
  specialties: string[];
  completedEngagements: number;
  ratingBasisPoints: number;
};

type ProfileList = { items: Profile[] };
type Engagement = { engagementId: string };

function formatGhs(pesewas: number) {
  return `GHS ${(pesewas / 100).toFixed(2)}`;
}

export function MatchmakerMarketplaceScreen() {
  const router = useRouter();
  const [profiles, setProfiles] = useState<Profile[] | null>(null);
  const [selected, setSelected] = useState<Profile | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    void apiRequest<ProfileList>("/v1/matchmakers")
      .then((result) => {
        setProfiles(result.items);
        setSelected(result.items[0] ?? null);
      })
      .catch((reason: unknown) =>
        setError(
          reason instanceof Error
            ? reason.message
            : "Licensed matchmakers could not load.",
        ),
      );
  }, []);

  async function book() {
    if (!selected) return;
    setBusy(true);
    setError("");
    try {
      await apiRequest<Engagement>("/v1/matchmaker-engagements", {
        method: "POST",
        headers: {
          "Idempotency-Key": `book.${Date.now()}-${Math.random().toString(36).slice(2)}`,
        },
        body: JSON.stringify({
          matchmakerId: selected.matchmakerId,
          termsId: "consultation.v1",
          termsVersion: 1,
          milestones: [
            {
              id: "consultation",
              feePesewas: selected.minimumFeePesewas,
              dueAfterDays: 0,
            },
          ],
        }),
      });
      setMessage(
        "Consultation booked with immutable terms. No candidate was exposed and no payment moved.",
      );
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The consultation could not be booked.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Fie</Text>
        </Pressable>
        <Text style={styles.eyebrow}>AGYINA · CURRENT LICENCES</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Find a guide, not a shortcut.
        </Text>
        <Text style={styles.copy}>
          Every profile comes from the licensing register. Matchmakers cannot
          sell seeds, rank, visibility or access to another member.
        </Text>
        {!profiles && !error ? <ActivityIndicator color="#8E3159" /> : null}
        {error ? (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {error}
          </Text>
        ) : null}
        {message ? (
          <Text accessibilityLiveRegion="polite" style={styles.success}>
            {message}
          </Text>
        ) : null}
        {profiles?.length === 0 ? (
          <View style={styles.profile}>
            <Text style={styles.profileName}>
              No current matchmaker is available.
            </Text>
            <Text style={styles.meta}>
              Expired, future and incomplete licences fail closed.
            </Text>
          </View>
        ) : null}
        {profiles?.map((profile) => (
          <View key={profile.matchmakerId} style={styles.profile}>
            <Text style={styles.license}>LICENSED · {profile.licenseId}</Text>
            <Text style={styles.profileName}>{profile.displayName}</Text>
            <Text style={styles.meta}>{profile.specialties.join(" · ")}</Text>
            <Text style={styles.meta}>{profile.languages.join(" / ")}</Text>
            <View style={styles.feeRow}>
              <Text style={styles.fee}>
                {formatGhs(profile.minimumFeePesewas)}–
                {formatGhs(profile.maximumFeePesewas)}
              </Text>
              <Text style={styles.rating}>
                {(profile.ratingBasisPoints / 100).toFixed(2)} ·{" "}
                {profile.completedEngagements} completed
              </Text>
            </View>
            <Pressable
              onPress={() => setSelected(profile)}
              style={styles.primary}
            >
              <Text style={styles.primaryText}>Review consultation</Text>
            </Pressable>
          </View>
        ))}
        <View style={styles.booking}>
          <Text style={styles.bookingEyebrow}>
            IMMUTABLE CONSULTATION TERMS
          </Text>
          <Text style={styles.bookingTitle}>
            {selected?.displayName ?? "Choose a licensed matchmaker"}
          </Text>
          {selected ? (
            <>
              <Text style={styles.disclosure}>
                One consultation at {formatGhs(selected.minimumFeePesewas)}.
                Licence current through{" "}
                {new Date(selected.licenseValidUntil).toLocaleDateString()}.
              </Text>
              <Pressable
                disabled={busy}
                onPress={() => void book()}
                style={[styles.action, busy && styles.disabled]}
              >
                <Text style={styles.actionText}>
                  {busy ? "Booking…" : "Book consultation"}
                </Text>
              </Pressable>
              <Text style={styles.disclosure}>
                Booking does not contact a payment provider or expose a
                candidate. Candidate consent remains a separate private action.
              </Text>
            </>
          ) : (
            <Text style={styles.disclosure}>
              The marketplace stays closed without a current licence.
            </Text>
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F5EEE4", flex: 1 },
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
    marginTop: 38,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 54,
    letterSpacing: -3.5,
    lineHeight: 49,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginBottom: 24,
    marginTop: 20,
  },
  error: {
    color: "#A33A4B",
    fontFamily: "Outfit_600SemiBold",
    marginVertical: 12,
  },
  success: {
    color: "#27755F",
    fontFamily: "Outfit_600SemiBold",
    marginVertical: 12,
  },
  profile: {
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 24,
    borderWidth: 1,
    marginTop: 10,
    padding: 20,
  },
  license: {
    color: "#27755F",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1,
  },
  profileName: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.5,
    marginTop: 8,
  },
  meta: { color: "#69535D", fontFamily: "Outfit_400Regular", marginTop: 5 },
  feeRow: {
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    gap: 10,
    marginTop: 16,
    paddingTop: 14,
  },
  fee: { color: "#2B151F", fontFamily: "Outfit_800ExtraBold" },
  rating: { color: "#69535D", fontFamily: "Outfit_600SemiBold" },
  primary: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 52,
  },
  primaryText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
  booking: {
    backgroundColor: "#38172C",
    borderRadius: 24,
    marginTop: 14,
    padding: 22,
  },
  bookingEyebrow: {
    color: "#FF9AB0",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  bookingTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.8,
    lineHeight: 34,
    marginBottom: 16,
    marginTop: 10,
  },
  action: {
    alignItems: "center",
    backgroundColor: "#FFB34F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 52,
  },
  actionText: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.42 },
  disclosure: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 21,
    marginTop: 16,
  },
});
