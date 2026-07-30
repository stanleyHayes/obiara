import { type Href, useRouter } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

type ConsentBoard = {
  purposes: { matching_personalization: boolean };
};

export function IntroductionExplanationScreen({
  introId,
}: Readonly<{ introId: string }>) {
  const router = useRouter();
  const [enabled, setEnabled] = useState<boolean | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    void apiRequest<ConsentBoard>("/v1/consent")
      .then((board) => {
        if (active) setEnabled(board.purposes.matching_personalization);
      })
      .catch((loadError: unknown) => {
        if (active) {
          setError(
            loadError instanceof Error
              ? loadError.message
              : "Your explanation controls could not be loaded.",
          );
        }
      });
    return () => {
      active = false;
    };
  }, []);

  async function allowPersonalization() {
    setBusy(true);
    setError(null);
    try {
      const result = await apiRequest<{
        purpose: string;
        enabled: boolean;
      }>("/v1/consent/purposes/matching_personalization", {
        method: "PUT",
        body: JSON.stringify({ enabled: true }),
      });
      setEnabled(result.enabled);
    } catch (saveError) {
      setError(
        saveError instanceof Error
          ? saveError.message
          : "Your consent choice could not be saved.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <View style={s.topbar}>
          <Pressable
            onPress={() => router.push("/fie/garden" as Href)}
            style={s.control}
          >
            <Text style={s.controlText}>Garden</Text>
          </Pressable>
          <Text style={s.reference}>{introId.slice(0, 8)}</Text>
        </View>

        <Text style={s.eyebrow}>WHY THIS INTRODUCTION</Text>
        <Text accessibilityRole="header" style={s.title}>
          No invented reasons. No hidden score.
        </Text>
        <Text style={s.body}>
          Obiara will show an explanation only when a retained introduction
          record supports it. No identity, match reason, or decision is
          manufactured here.
        </Text>

        <View style={s.status}>
          <Text style={s.statusEyebrow}>EXPLANATION STATUS</Text>
          <Text accessibilityRole="header" style={s.statusTitle}>
            Waiting for verified introduction data.
          </Text>
          <Text style={s.statusCopy}>
            Candidate identity, reciprocal preferences, trust paths, and voice
            comparisons are not inferred by this screen.
          </Text>
        </View>

        <Text style={s.sectionEyebrow}>YOUR REAL CONTROL</Text>
        <Text accessibilityRole="header" style={s.sectionTitle}>
          Introduction personalization
        </Text>
        <Text style={s.sectionCopy}>
          This purpose-bound choice comes from your consent record. It does not
          create an introduction or imply that one exists.
        </Text>

        <View style={s.card}>
          <Text style={s.cardTitle}>
            Use preferences for private introductions
          </Text>
          <Text style={s.cardCopy}>
            Optional and one-way under the current consent policy.
          </Text>
          <Pressable
            accessibilityRole="switch"
            accessibilityState={{
              checked: enabled ?? false,
              disabled: enabled !== false || busy,
            }}
            disabled={enabled !== false || busy}
            onPress={() => void allowPersonalization()}
            style={[
              s.button,
              enabled && s.buttonOn,
              (enabled !== false || busy) && s.disabled,
            ]}
          >
            <Text style={[s.buttonText, enabled && s.buttonTextOn]}>
              {busy
                ? "Saving…"
                : enabled === null
                  ? "Loading…"
                  : enabled
                    ? "Allowed"
                    : "Allow"}
            </Text>
          </Pressable>
        </View>

        <Pressable
          onPress={() => router.push("/fie/settings/consent" as Href)}
          style={s.secondary}
        >
          <Text style={s.secondaryText}>Open consent switchboard</Text>
        </Pressable>
        {error ? <Text style={s.error}>{error}</Text> : null}
        <Text style={s.footer}>
          No urgency, read receipt, public activity signal, or fabricated
          compatibility claim.
        </Text>
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
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
  reference: {
    color: "#705C67",
    fontFamily: "Outfit_700Bold",
    fontSize: 12,
    letterSpacing: 1,
  },
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
    fontSize: 49,
    letterSpacing: -3,
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
  status: {
    backgroundColor: "#28161F",
    borderRadius: 26,
    marginTop: 38,
    padding: 22,
  },
  statusEyebrow: {
    color: "#FFB7C4",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  statusTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.8,
    lineHeight: 36,
    marginTop: 12,
  },
  statusCopy: {
    color: "#E8D8DF",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 23,
    marginTop: 16,
  },
  sectionEyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
    marginTop: 40,
  },
  sectionTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 36,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 8,
  },
  sectionCopy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 23,
    marginTop: 12,
  },
  card: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,.12)",
    borderRadius: 18,
    borderWidth: 1,
    marginTop: 22,
    padding: 18,
  },
  cardTitle: {
    color: "#28161F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 19,
  },
  cardCopy: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 14,
    lineHeight: 21,
    marginTop: 7,
  },
  button: {
    alignItems: "center",
    borderColor: "#3A0E2E",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 48,
  },
  buttonOn: { backgroundColor: "#3A0E2E" },
  buttonText: {
    color: "#3A0E2E",
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
  },
  buttonTextOn: { color: "#FFF3E6" },
  disabled: { opacity: 0.6 },
  secondary: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: 52,
  },
  secondaryText: {
    color: "#3A0E2E",
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
    textDecorationLine: "underline",
  },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    marginTop: 8,
  },
  footer: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
    lineHeight: 18,
    marginTop: 28,
    textAlign: "center",
  },
});
