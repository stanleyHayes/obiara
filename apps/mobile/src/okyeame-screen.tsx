import { decideOkyeame, type OkyeameCapability } from "@obiara/okyeame-policy";
import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const requestOptions: readonly {
  readonly capability: OkyeameCapability;
  readonly label: string;
}[] = [
  { capability: "feature_help", label: "Explain a Fie feature" },
  { capability: "navigation_help", label: "Help me find an area" },
  { capability: "wording_help", label: "Help revise my words" },
  { capability: "matchmaking_decision", label: "Choose a match for me" },
  { capability: "autonomous_romance", label: "Message someone for me" },
  { capability: "counsel_disclosure", label: "Show counsel notes" },
];

export function OkyeameScreen() {
  const router = useRouter();
  const [selected, setSelected] = useState<OkyeameCapability>("feature_help");
  const decision = decideOkyeame({
    capability: selected,
    memberInvoked: true,
  });

  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.intro}>
          <Text style={styles.disclosure}>AI-GUIDED HELP, NOT A PERSON</Text>
          <Text accessibilityRole="header" style={styles.title}>
            Help should know its place.
          </Text>
          <Text style={styles.body}>
            You choose when Okyeame responds. It cannot decide, act or remember
            for you.
          </Text>
        </View>

        <View style={styles.section}>
          <Text accessibilityRole="header" style={styles.sectionTitle}>
            Check the boundary first.
          </Text>
          <Text style={styles.sectionBody}>
            Pick an example to see what guided help may answer.
          </Text>
          <View accessibilityLabel="Choose a request" style={styles.requests}>
            {requestOptions.map((request) => {
              const isSelected = selected === request.capability;
              return (
                <Pressable
                  accessibilityRole="button"
                  accessibilityState={{ selected: isSelected }}
                  key={request.capability}
                  onPress={() => setSelected(request.capability)}
                  style={({ pressed }) => [
                    styles.request,
                    isSelected && styles.requestSelected,
                    pressed && styles.pressed,
                  ]}
                >
                  <Text style={styles.requestText}>{request.label}</Text>
                </Pressable>
              );
            })}
          </View>

          <View
            accessibilityLiveRegion="polite"
            style={[
              styles.decision,
              decision.allowed
                ? styles.decisionAllowed
                : styles.decisionRefused,
            ]}
          >
            <Text style={styles.decisionLabel}>AI-GUIDED HELP</Text>
            <Text style={styles.decisionTitle}>{decision.heading}</Text>
            <Text style={styles.decisionBody}>{decision.message}</Text>
            <View style={styles.decisionFooter}>
              <Text style={styles.decisionMeta}>
                {decision.allowed ? "Within whitelist" : "Request refused"}
              </Text>
              <Text style={styles.decisionMeta}>Prompt not retained</Text>
            </View>
          </View>
          <Pressable
            accessibilityRole="link"
            onPress={() => router.push("/fie/okyeame-accountability" as Href)}
            style={({ pressed }) => [
              styles.accountabilityLink,
              pressed && styles.pressed,
            ]}
          >
            <Text style={styles.accountabilityLinkText}>
              View AI accountability
            </Text>
          </Pressable>
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
    minHeight: 500,
    padding: 30,
  },
  disclosure: {
    color: "#FFB44F",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 54,
    letterSpacing: -3,
    lineHeight: 50,
    marginTop: 54,
  },
  body: {
    color: "rgba(255, 243, 230, 0.72)",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginTop: "auto",
  },
  section: { marginTop: 38 },
  sectionTitle: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.5,
    lineHeight: 36,
  },
  sectionBody: {
    color: "#6C5965",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
  },
  requests: { gap: 8, marginTop: 22 },
  request: {
    backgroundColor: "#FFFFFF",
    borderColor: "rgba(58, 14, 46, 0.12)",
    borderRadius: 14,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 56,
    paddingHorizontal: 16,
  },
  requestSelected: {
    backgroundColor: "#F9E7EC",
    borderColor: "#BD3153",
  },
  requestText: {
    color: "#2A1022",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 15,
  },
  pressed: { opacity: 0.72 },
  decision: {
    backgroundColor: "#2A1022",
    borderRadius: 20,
    borderWidth: 2,
    marginTop: 16,
    minHeight: 280,
    padding: 24,
  },
  decisionAllowed: { borderColor: "#4CAE91" },
  decisionRefused: { borderColor: "#FFB44F" },
  decisionLabel: {
    color: "#FFB44F",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1,
  },
  decisionTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 28,
    letterSpacing: -1,
    lineHeight: 31,
    marginTop: 30,
  },
  decisionBody: {
    color: "rgba(255, 243, 230, 0.72)",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 14,
  },
  decisionFooter: {
    borderTopColor: "rgba(255, 243, 230, 0.16)",
    borderTopWidth: 1,
    gap: 7,
    marginTop: "auto",
    paddingTop: 16,
  },
  decisionMeta: {
    color: "rgba(255, 243, 230, 0.82)",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
  },
  accountabilityLink: {
    alignItems: "center",
    borderColor: "#2A1022",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 18,
    minHeight: 48,
  },
  accountabilityLinkText: {
    color: "#2A1022",
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
  },
});
