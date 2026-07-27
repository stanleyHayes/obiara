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

const palette = {
  plum: "#3A0E2E",
  ink: "#26101F",
  cream: "#FFF3E6",
  paper: "#FFFDFC",
  gold: "#FF9F1C",
  pink: "#FF4D6D",
  green: "#12876B",
  muted: "#765F70",
  line: "rgba(58, 14, 46, 0.11)",
};

const account = {
  memberRef: "member···92K",
  displayName: "Ama Serwaa",
  verification: "Ghana Card · verified",
  tier: "Tier 2 · sowing",
  host: "Active",
  joined: "Mar 2026",
};

const displayNameLimit = 80;
const introductionLimit = 280;
const unsafePattern =
  /([A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,})|(https?:\/\/|www\.)|(\+?\d[\d ()-]{7,}\d)/i;

const visibilityOptions = [
  { value: "private", label: "Only me" },
  { value: "circles", label: "My circles" },
  { value: "community", label: "Community" },
] as const;

type Visibility = (typeof visibilityOptions)[number]["value"];

export function ProfileSettingsScreen() {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(account.displayName);
  const [introduction, setIntroduction] = useState(
    "Teacher in Accra. I cook groundnut soup for people I like.",
  );
  const [nameVisibility, setNameVisibility] = useState<Visibility>("circles");
  const [introVisibility, setIntroVisibility] = useState<Visibility>("private");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  function save() {
    const name = displayName.trim();
    if (name === "") {
      setError("A display name is required. It is how your circle knows you.");
      setSaved(false);
      return;
    }
    if ([...name].length > displayNameLimit) {
      setError(`Display name must be ${displayNameLimit} characters or fewer.`);
      setSaved(false);
      return;
    }
    if ([...introduction.trim()].length > introductionLimit) {
      setError(
        `Introduction must be ${introductionLimit} characters or fewer.`,
      );
      setSaved(false);
      return;
    }
    if (unsafePattern.test(name) || unsafePattern.test(introduction)) {
      setError(
        "Profile fields cannot carry contact details or links. Obiara connects people itself.",
      );
      setSaved(false);
      return;
    }
    setError(null);
    setSaved(true);
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
        <Text style={styles.eyebrow}>YOUR PROFILE</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Be known on your own terms.
        </Text>

        <View style={styles.card}>
          <View style={styles.identity}>
            <View style={styles.monogram}>
              <Text style={styles.monogramText}>AS</Text>
            </View>
            <View>
              <Text style={styles.identityName}>
                {displayName || account.displayName}
              </Text>
              <Text style={styles.identityMeta}>
                {account.verification} · host
              </Text>
            </View>
          </View>
          {[
            ["Member reference", account.memberRef],
            ["Tier", account.tier],
            ["Host capability", account.host],
            ["Member since", account.joined],
          ].map(([label, value]) => (
            <View key={label} style={styles.fact}>
              <Text style={styles.factLabel}>{label}</Text>
              <Text style={styles.factValue}>{value}</Text>
            </View>
          ))}
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>EDIT PROFILE</Text>
          <Text style={styles.label}>
            Display name · {[...displayName].length}/{displayNameLimit}
          </Text>
          <TextInput
            onChangeText={(value) => {
              setDisplayName(value);
              setSaved(false);
            }}
            style={styles.input}
            value={displayName}
          />
          <Text style={styles.label}>Visible to</Text>
          <View style={styles.choiceRow}>
            {visibilityOptions.map((option) => (
              <Pressable
                key={option.value}
                onPress={() => setNameVisibility(option.value)}
                style={[
                  styles.choice,
                  nameVisibility === option.value && styles.choiceActive,
                ]}
              >
                <Text
                  style={[
                    styles.choiceText,
                    nameVisibility === option.value && styles.choiceTextActive,
                  ]}
                >
                  {option.label}
                </Text>
              </Pressable>
            ))}
          </View>
          <Text style={styles.label}>
            Introduction · {[...introduction].length}/{introductionLimit}
          </Text>
          <TextInput
            multiline
            onChangeText={(value) => {
              setIntroduction(value);
              setSaved(false);
            }}
            style={[styles.input, styles.inputMultiline]}
            value={introduction}
          />
          <Text style={styles.label}>Visible to</Text>
          <View style={styles.choiceRow}>
            {visibilityOptions.map((option) => (
              <Pressable
                key={option.value}
                onPress={() => setIntroVisibility(option.value)}
                style={[
                  styles.choice,
                  introVisibility === option.value && styles.choiceActive,
                ]}
              >
                <Text
                  style={[
                    styles.choiceText,
                    introVisibility === option.value && styles.choiceTextActive,
                  ]}
                >
                  {option.label}
                </Text>
              </Pressable>
            ))}
          </View>
          <Text style={styles.note}>
            Profile fields never carry phone numbers, emails or links. Choosing
            Community records a consent receipt for that field.
          </Text>
          {error ? <Text style={styles.error}>{error}</Text> : null}
          {saved ? (
            <Text style={styles.saved}>
              Profile saved. Your circle sees the change on their next view.
            </Text>
          ) : null}
          <Pressable onPress={save} style={styles.saveButton}>
            <Text style={styles.saveButtonText}>Save changes</Text>
          </Pressable>
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>MORE SETTINGS</Text>
          {[
            [
              "Notifications",
              "Quiet hours, caps and channels",
              "/fie/settings/notifications",
            ],
            [
              "Membership",
              "Terms, receipts and renewal",
              "/fie/settings/membership",
            ],
            [
              "Suban",
              "Your character marks and history",
              "/fie/settings/suban",
            ],
          ].map(([label, detail, href]) => (
            <Pressable
              key={label}
              onPress={() => router.push(href as Href)}
              style={({ pressed }) => [
                styles.linkRow,
                pressed && styles.pressed,
              ]}
            >
              <View>
                <Text style={styles.linkLabel}>{label}</Text>
                <Text style={styles.linkDetail}>{detail}</Text>
              </View>
              <Text aria-hidden style={styles.linkChevron}>
                ›
              </Text>
            </Pressable>
          ))}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: palette.cream, flex: 1 },
  content: { padding: 20, paddingBottom: 40 },
  back: { alignSelf: "flex-start", minHeight: 44, justifyContent: "center" },
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
    marginBottom: 18,
    marginTop: 8,
  },
  card: {
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 14,
    borderWidth: 1,
    marginBottom: 14,
    padding: 18,
  },
  identity: {
    alignItems: "center",
    flexDirection: "row",
    gap: 12,
    marginBottom: 12,
  },
  monogram: {
    alignItems: "center",
    backgroundColor: palette.plum,
    borderRadius: 26,
    height: 52,
    justifyContent: "center",
    width: 52,
  },
  monogramText: {
    color: palette.cream,
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 18,
  },
  identityName: {
    color: palette.ink,
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 20,
  },
  identityMeta: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 13,
    marginTop: 2,
  },
  fact: {
    borderTopColor: palette.line,
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 10,
  },
  factLabel: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 13,
  },
  factValue: { color: palette.ink, fontFamily: "Outfit_700Bold", fontSize: 13 },
  label: {
    color: palette.muted,
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
    marginBottom: 6,
    marginTop: 12,
  },
  input: {
    backgroundColor: "#FFFFFF",
    borderColor: palette.line,
    borderRadius: 10,
    borderWidth: 1,
    color: palette.ink,
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
    minHeight: 46,
    paddingHorizontal: 12,
  },
  inputMultiline: { minHeight: 90, paddingTop: 10, textAlignVertical: "top" },
  choiceRow: { flexDirection: "row", gap: 8 },
  choice: {
    borderColor: palette.line,
    borderRadius: 999,
    borderWidth: 1,
    paddingHorizontal: 14,
    paddingVertical: 8,
  },
  choiceActive: { backgroundColor: palette.plum, borderColor: palette.plum },
  choiceText: {
    color: palette.plum,
    fontFamily: "Outfit_600SemiBold",
    fontSize: 12,
  },
  choiceTextActive: { color: palette.cream },
  note: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
    lineHeight: 18,
    marginTop: 14,
  },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    marginTop: 12,
  },
  saved: {
    color: palette.green,
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    marginTop: 12,
  },
  saveButton: {
    alignItems: "center",
    backgroundColor: palette.plum,
    borderRadius: 10,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 48,
  },
  saveButtonText: {
    color: palette.cream,
    fontFamily: "Outfit_700Bold",
    fontSize: 15,
  },
  linkRow: {
    alignItems: "center",
    borderTopColor: palette.line,
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 12,
  },
  linkLabel: { color: palette.ink, fontFamily: "Outfit_700Bold", fontSize: 15 },
  linkDetail: {
    color: palette.muted,
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
    marginTop: 2,
  },
  linkChevron: { color: palette.plum, fontSize: 20 },
  pressed: { opacity: 0.7 },
});
