import { useRouter } from "expo-router";
import Constants from "expo-constants";
import { useState } from "react";
import {
  Linking,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

interface PrivacyRequest {
  requestId: string;
  kind: "export" | "deletion";
  status: string;
  dueAt: string;
}

const palette = {
  cream: "#FFF3E6",
  ink: "#26101F",
  line: "rgba(58, 14, 46, 0.11)",
  muted: "#765F70",
  paper: "#FFFDFC",
  pink: "#FF4D6D",
  plum: "#3A0E2E",
};

export function PrivacySettingsScreen() {
  const router = useRouter();
  const [record, setRecord] = useState<PrivacyRequest | null>(null);
  const [reference, setReference] = useState("");
  const [busy, setBusy] = useState<"export" | "deletion" | "status" | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const privacyPolicyUrl = Constants.expoConfig?.extra?.privacyPolicyUrl;
  const supportUrl = Constants.expoConfig?.extra?.supportUrl;

  async function open(kind: "export" | "deletion") {
    setBusy(kind);
    setError(null);
    try {
      const result = await apiRequest<PrivacyRequest>(
        kind === "export" ? "/v1/privacy/exports" : "/v1/privacy/deletions",
        { method: "POST", body: "{}" },
      );
      setRecord(result);
      setReference(result.requestId);
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "The request could not be opened.",
      );
    } finally {
      setBusy(null);
    }
  }

  async function check() {
    const id = reference.trim();
    if (!id) return;
    setBusy("status");
    setError(null);
    try {
      setRecord(
        await apiRequest<PrivacyRequest>(
          `/v1/privacy/requests/${encodeURIComponent(id)}`,
        ),
      );
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "The request could not be loaded.",
      );
    } finally {
      setBusy(null);
    }
  }

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable onPress={() => router.back()} style={styles.back}>
          <Text style={styles.backText}>Profile</Text>
        </Pressable>
        <Text style={styles.eyebrow}>YOUR DATA RIGHTS</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Take your record, or ask us to close it.
        </Text>
        <Text style={styles.lede}>
          Exports are prepared within 72 hours. Deletion is completed within 30
          days unless a lawful preservation hold applies.
        </Text>

        <View style={styles.card}>
          <Text style={styles.cardTitle}>Portable archive</Text>
          <Text style={styles.copy}>
            Request a machine-readable copy delivered through your verified
            account channel.
          </Text>
          <Pressable
            disabled={busy !== null}
            onPress={() => void open("export")}
            style={[styles.button, busy !== null && styles.disabled]}
          >
            <Text style={styles.buttonText}>
              {busy === "export" ? "Opening request…" : "Request my export"}
            </Text>
          </Pressable>
        </View>

        <View style={styles.card}>
          <Text style={styles.cardTitle}>Policy and help</Text>
          <Text style={styles.copy}>
            Read how Obiara handles personal data or contact support without
            sharing a password or one-time code.
          </Text>
          {typeof privacyPolicyUrl === "string" ? (
            <Pressable
              accessibilityRole="link"
              onPress={() => void Linking.openURL(privacyPolicyUrl)}
              style={styles.textLink}
            >
              <Text style={styles.textLinkLabel}>
                Read the privacy policy ↗
              </Text>
            </Pressable>
          ) : null}
          {typeof supportUrl === "string" ? (
            <Pressable
              accessibilityRole="link"
              onPress={() => void Linking.openURL(supportUrl)}
              style={styles.textLink}
            >
              <Text style={styles.textLinkLabel}>Open member support ↗</Text>
            </Pressable>
          ) : null}
        </View>

        <View style={styles.card}>
          <Text style={styles.cardTitle}>Close the account</Text>
          <Text style={styles.copy}>
            Request deletion and cryptographic erasure. A legal or safety hold
            may delay eligible records, and the status will remain visible.
          </Text>
          <Pressable
            disabled={busy !== null}
            onPress={() => void open("deletion")}
            style={[
              styles.button,
              styles.danger,
              busy !== null && styles.disabled,
            ]}
          >
            <Text style={styles.buttonText}>
              {busy === "deletion"
                ? "Opening request…"
                : "Request account deletion"}
            </Text>
          </Pressable>
        </View>

        <View style={styles.card}>
          <Text style={styles.cardTitle}>Track a request</Text>
          <TextInput
            accessibilityLabel="Privacy request reference"
            autoCapitalize="none"
            onChangeText={setReference}
            placeholder="Opaque request reference"
            placeholderTextColor={palette.muted}
            style={styles.input}
            value={reference}
          />
          <Pressable
            disabled={busy !== null || !reference.trim()}
            onPress={() => void check()}
            style={[
              styles.button,
              (busy !== null || !reference.trim()) && styles.disabled,
            ]}
          >
            <Text style={styles.buttonText}>
              {busy === "status" ? "Checking…" : "Check status"}
            </Text>
          </Pressable>
          {error ? <Text style={styles.error}>{error}</Text> : null}
          {record ? (
            <View accessibilityLiveRegion="polite" style={styles.result}>
              <Text style={styles.resultTitle}>
                {record.kind === "export" ? "Export" : "Deletion"} ·{" "}
                {record.status.replaceAll("_", " ")}
              </Text>
              <Text style={styles.copy}>
                Due by {new Date(record.dueAt).toLocaleString("en-GH")}
              </Text>
              <Text selectable style={styles.reference}>
                {record.requestId}
              </Text>
            </View>
          ) : null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: palette.cream, flex: 1 },
  content: { padding: 20, paddingBottom: 40 },
  back: { alignSelf: "flex-start", justifyContent: "center", minHeight: 44 },
  backText: { color: palette.plum, fontFamily: "Outfit_700Bold", fontSize: 15 },
  eyebrow: {
    color: palette.pink,
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.4,
  },
  title: {
    color: palette.ink,
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.6,
    lineHeight: 38,
    marginTop: 8,
  },
  lede: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    lineHeight: 22,
    marginBottom: 20,
    marginTop: 10,
  },
  card: {
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 14,
    borderWidth: 1,
    marginBottom: 14,
    padding: 18,
  },
  cardTitle: {
    color: palette.ink,
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 19,
  },
  copy: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 13,
    lineHeight: 19,
    marginTop: 7,
  },
  button: {
    alignItems: "center",
    backgroundColor: palette.plum,
    borderRadius: 10,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 48,
    paddingHorizontal: 14,
  },
  danger: { backgroundColor: "#8E1F3C" },
  disabled: { opacity: 0.55 },
  buttonText: {
    color: palette.cream,
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
    textAlign: "center",
  },
  input: {
    backgroundColor: "#FFFFFF",
    borderColor: palette.line,
    borderRadius: 10,
    borderWidth: 1,
    color: palette.ink,
    fontFamily: "Outfit_400Regular",
    fontSize: 14,
    marginTop: 12,
    minHeight: 48,
    paddingHorizontal: 12,
  },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    marginTop: 12,
  },
  result: {
    borderTopColor: palette.line,
    borderTopWidth: 1,
    marginTop: 16,
    paddingTop: 14,
  },
  resultTitle: {
    color: palette.ink,
    fontFamily: "Outfit_700Bold",
    fontSize: 15,
    textTransform: "capitalize",
  },
  reference: {
    color: palette.plum,
    fontFamily: "Outfit_600SemiBold",
    fontSize: 11,
    marginTop: 8,
  },
  textLink: { alignSelf: "flex-start", marginTop: 14, minHeight: 32 },
  textLinkLabel: {
    color: palette.pink,
    fontFamily: "Outfit_700Bold",
    fontSize: 14,
  },
});
