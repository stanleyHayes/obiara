import { useRouter } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

type Purpose =
  | "identity_safety"
  | "matching_personalization"
  | "scam_arc_monitoring"
  | "play_portraits"
  | "product_analytics"
  | "profile_visibility";
const rows: readonly [
  Purpose,
  string,
  string,
  "required" | "opt-in" | "opt-out" | "toggle",
][] = [
  [
    "identity_safety",
    "Identity and safety",
    "Required to secure the community and respond to harm.",
    "required",
  ],
  [
    "matching_personalization",
    "Introduction personalization",
    "Allows preferences to shape private introductions.",
    "opt-in",
  ],
  [
    "scam_arc_monitoring",
    "Scam-pattern monitoring",
    "Looks for bounded risk patterns without exposing trust paths.",
    "opt-out",
  ],
  [
    "play_portraits",
    "Play portraits",
    "Allows consented play to inform a private self-portrait.",
    "opt-in",
  ],
  [
    "product_analytics",
    "Product analytics",
    "Uses purpose-limited events to improve reliability.",
    "opt-out",
  ],
  [
    "profile_visibility",
    "Profile visibility",
    "Records when profile fields may be shown beyond you.",
    "toggle",
  ],
];

export function ConsentSettingsScreen() {
  const router = useRouter();
  const [purposes, setPurposes] = useState<Partial<Record<Purpose, boolean>>>(
    {},
  );
  const [busy, setBusy] = useState<Purpose | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    void apiRequest<{ purposes: Record<Purpose, boolean> }>("/v1/consent")
      .then((board) => {
        if (active) setPurposes(board.purposes);
      })
      .catch((loadError: unknown) => {
        if (active)
          setError(
            loadError instanceof Error
              ? loadError.message
              : "Your consent choices could not be loaded.",
          );
      });
    return () => {
      active = false;
    };
  }, []);

  async function change(purpose: Purpose, enabled: boolean) {
    setBusy(purpose);
    setError(null);
    try {
      const result = await apiRequest<{ purpose: Purpose; enabled: boolean }>(
        `/v1/consent/purposes/${purpose}`,
        { method: "PUT", body: JSON.stringify({ enabled }) },
      );
      setPurposes((current) => ({
        ...current,
        [result.purpose]: result.enabled,
      }));
    } catch (saveError) {
      setError(
        saveError instanceof Error
          ? saveError.message
          : "Your consent choice could not be saved.",
      );
    } finally {
      setBusy(null);
    }
  }

  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <Pressable onPress={() => router.back()} style={s.back}>
          <Text style={s.backText}>Profile</Text>
        </Pressable>
        <Text style={s.eyebrow}>CONSENT SWITCHBOARD</Text>
        <Text accessibilityRole="header" style={s.title}>
          See exactly what you have allowed.
        </Text>
        <Text style={s.lede}>
          Each purpose is separate. Optional choices do not change your standing
          or visibility.
        </Text>
        {rows.map(([purpose, label, detail, control]) => {
          const enabled = purposes[purpose];
          const locked =
            enabled === undefined ||
            control === "required" ||
            (control === "opt-in" && enabled) ||
            (control === "opt-out" && !enabled);
          return (
            <View key={purpose} style={s.card}>
              <Text style={s.control}>{control.replace("-", " ")}</Text>
              <Text style={s.cardTitle}>{label}</Text>
              <Text style={s.copy}>{detail}</Text>
              <Pressable
                accessibilityRole="switch"
                accessibilityState={{
                  checked: enabled ?? false,
                  disabled: locked,
                }}
                disabled={locked || busy !== null}
                onPress={() => void change(purpose, !enabled)}
                style={[
                  s.button,
                  enabled && s.buttonOn,
                  (locked || busy !== null) && s.disabled,
                ]}
              >
                <Text style={[s.buttonText, enabled && s.buttonTextOn]}>
                  {busy === purpose
                    ? "Saving…"
                    : enabled === undefined
                      ? "Loading…"
                      : enabled
                        ? "Allowed"
                        : "Not allowed"}
                </Text>
              </Pressable>
            </View>
          );
        })}
        {error ? <Text style={s.error}>{error}</Text> : null}
        <Text style={s.footnote}>
          Required processing cannot be disabled. One-way choices cannot
          manufacture renewed consent later.
        </Text>
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#FFF3E6", flex: 1 },
  content: { padding: 20, paddingBottom: 48 },
  back: { alignSelf: "flex-start", justifyContent: "center", minHeight: 44 },
  backText: { color: "#3A0E2E", fontFamily: "Outfit_700Bold", fontSize: 15 },
  eyebrow: {
    color: "#FF4D6D",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.4,
  },
  title: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.6,
    lineHeight: 38,
    marginTop: 8,
  },
  lede: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 22,
    marginBottom: 20,
    marginTop: 10,
  },
  card: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,.11)",
    borderRadius: 14,
    borderWidth: 1,
    marginBottom: 10,
    padding: 17,
  },
  control: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1,
    textTransform: "uppercase",
  },
  cardTitle: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 18,
    marginTop: 6,
  },
  copy: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 13,
    lineHeight: 19,
    marginTop: 5,
  },
  button: {
    alignItems: "center",
    borderColor: "#3A0E2E",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 13,
    minHeight: 44,
  },
  buttonOn: { backgroundColor: "#3A0E2E" },
  buttonText: { color: "#3A0E2E", fontFamily: "Outfit_700Bold", fontSize: 13 },
  buttonTextOn: { color: "#FFF3E6" },
  disabled: { opacity: 0.55 },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    marginVertical: 10,
  },
  footnote: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
    lineHeight: 18,
    marginTop: 8,
  },
});
