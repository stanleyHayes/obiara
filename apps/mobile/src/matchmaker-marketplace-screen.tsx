import {
  canExposeCuratedProposal,
  initialMarketplaceState,
  marketplaceReducer,
} from "@obiara/matchmaker-marketplace";
import { type Href, useRouter } from "expo-router";
import { useReducer } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export function MatchmakerMarketplaceScreen() {
  const router = useRouter();
  const [state, dispatch] = useReducer(
    marketplaceReducer,
    initialMarketplaceState,
  );
  const selected = state.profiles.find((item) => item.id === state.selectedId);

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Fie</Text>
        </Pressable>
        <Text style={styles.eyebrow}>AGYINA · LICENSED GUIDANCE</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Find a guide, not a shortcut.
        </Text>
        <Text style={styles.copy}>
          Matchmakers cannot sell seeds, visibility, rank or access to another
          member.
        </Text>

        {state.profiles.map((profile) => (
          <View key={profile.id} style={styles.profile}>
            <Text style={styles.license}>LICENSED · {profile.licenseRef}</Text>
            <Text style={styles.profileName}>{profile.name}</Text>
            <Text style={styles.meta}>{profile.specialties.join(" · ")}</Text>
            <Text style={styles.meta}>{profile.languages.join(" / ")}</Text>
            <View style={styles.feeRow}>
              <Text style={styles.fee}>GHS {profile.consultationFeeGhs}</Text>
              <Text style={styles.rating}>
                {profile.rating} · {profile.completedEngagements} completed
              </Text>
            </View>
            <Pressable
              onPress={() => dispatch({ type: "select", id: profile.id })}
              style={styles.primary}
            >
              <Text style={styles.primaryText}>Review services</Text>
            </Pressable>
          </View>
        ))}

        <View style={styles.booking}>
          <Text style={styles.bookingEyebrow}>ENGAGEMENT PREVIEW</Text>
          <Text style={styles.bookingTitle}>
            {selected?.name ?? "Choose a licensed matchmaker"}
          </Text>
          {selected ? (
            <>
              <Pressable
                accessibilityState={{
                  selected: state.service === "consultation",
                }}
                onPress={() =>
                  dispatch({ type: "service", value: "consultation" })
                }
                style={styles.service}
              >
                <Text style={styles.serviceText}>Consultation</Text>
                <Text style={styles.serviceText}>GHS 80–250</Text>
              </Pressable>
              <Pressable
                accessibilityState={{ selected: state.service === "curated" }}
                onPress={() => dispatch({ type: "service", value: "curated" })}
                style={styles.service}
              >
                <Text style={styles.serviceText}>Three curated proposals</Text>
                <Text style={styles.serviceText}>GHS 250–600</Text>
              </Pressable>
              {state.service === "consultation" ? (
                <Pressable
                  onPress={() => dispatch({ type: "confirm-booking" })}
                  style={styles.action}
                >
                  <Text style={styles.actionText}>
                    {state.bookingConfirmed
                      ? "Consultation intent recorded"
                      : "Confirm consultation intent"}
                  </Text>
                </Pressable>
              ) : (
                <>
                  <Text style={styles.consentCopy}>
                    Candidate exposure needs two current consents.
                  </Text>
                  <Pressable
                    accessibilityRole="checkbox"
                    accessibilityState={{ checked: state.yourProposalConsent }}
                    onPress={() =>
                      dispatch({
                        type: "your-consent",
                        value: !state.yourProposalConsent,
                      })
                    }
                    style={styles.consent}
                  >
                    <Text style={styles.consentText}>
                      {state.yourProposalConsent ? "✓" : "○"} I consent to a
                      bounded proposal
                    </Text>
                  </Pressable>
                  <Pressable
                    accessibilityRole="checkbox"
                    accessibilityState={{
                      checked: state.candidateProposalConsent,
                    }}
                    onPress={() =>
                      dispatch({
                        type: "candidate-consent",
                        value: !state.candidateProposalConsent,
                      })
                    }
                    style={styles.consent}
                  >
                    <Text style={styles.consentText}>
                      {state.candidateProposalConsent ? "✓" : "○"} Preview
                      candidate consent
                    </Text>
                  </Pressable>
                  <Pressable
                    accessibilityState={{
                      disabled: !canExposeCuratedProposal(state),
                    }}
                    disabled={!canExposeCuratedProposal(state)}
                    style={[
                      styles.action,
                      !canExposeCuratedProposal(state) && styles.disabled,
                    ]}
                  >
                    <Text style={styles.actionText}>
                      Review consented proposal
                    </Text>
                  </Pressable>
                </>
              )}
              <Text style={styles.disclosure}>
                No charge or booking is created here. Fees, milestones and
                dispute terms appear before payment.
              </Text>
            </>
          ) : (
            <Text style={styles.disclosure}>
              Ratings appear only after completed engagements.
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
  meta: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    marginTop: 5,
  },
  feeRow: {
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
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
  service: {
    borderColor: "rgba(255,245,233,0.35)",
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    marginTop: 8,
    padding: 14,
  },
  serviceText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
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
  consentCopy: {
    color: "#FFF5E9",
    fontFamily: "Outfit_700Bold",
    marginTop: 16,
  },
  consent: { paddingVertical: 12 },
  consentText: { color: "#FFF5E9", fontFamily: "Outfit_600SemiBold" },
  disclosure: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 21,
    marginTop: 16,
  },
});
