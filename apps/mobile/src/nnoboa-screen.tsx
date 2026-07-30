import { apiRequest } from "./api";
import { type Href, useRouter } from "expo-router";
import { useEffect, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

type Relationship = "aunt" | "uncle" | "mother" | "father" | "elder";

type Nomination = {
  id: string;
  kinName: string;
  relationship: Relationship;
  status: "pending" | "consented" | "declined" | "expired";
  createdAt: string;
};

type NominationList = { nominations: Nomination[] };

const relationships: readonly Relationship[] = [
  "aunt",
  "uncle",
  "mother",
  "father",
  "elder",
];

export function NnoboaScreen() {
  const router = useRouter();
  const [nominations, setNominations] = useState<Nomination[]>([]);
  const [kinName, setKinName] = useState("");
  const [kinPhone, setKinPhone] = useState("");
  const [relationship, setRelationship] = useState<Relationship>("aunt");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState("");

  async function load(clearNotice = true) {
    setLoading(true);
    if (clearNotice) setNotice("");
    try {
      const result = await apiRequest<NominationList>("/v1/nominations");
      setNominations(result.nominations);
    } catch (error) {
      setNotice(
        error instanceof Error ? error.message : "Nominations could not load.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function nominate() {
    if (!kinName.trim() || !/^\+[1-9]\d{7,14}$/.test(kinPhone.trim())) {
      setNotice("Add a name and an international phone number such as +233…");
      return;
    }
    setSaving(true);
    setNotice("");
    try {
      await apiRequest<Nomination>("/v1/nominations", {
        method: "POST",
        body: JSON.stringify({
          kinName: kinName.trim(),
          kinPhone: kinPhone.trim(),
          relationship,
        }),
      });
      setKinName("");
      setKinPhone("");
      setNotice("Invitation sent privately. They have 30 days to respond.");
      await load(false);
    } catch (error) {
      setNotice(
        error instanceof Error ? error.message : "The invitation was not sent.",
      );
    } finally {
      setSaving(false);
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

        <Text style={styles.eyebrow}>NNOBOA · TRUSTED HANDS</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Invite someone you trust.
        </Text>
        <Text style={styles.copy}>
          Choose a trusted elder or family member. They receive a private
          invitation to support you—not access to your courtship.
        </Text>

        <View style={styles.panel}>
          <Text style={styles.panelLabel}>NEW INVITATION</Text>
          <TextInput
            accessibilityLabel="Trusted person name"
            onChangeText={setKinName}
            placeholder="Their name"
            placeholderTextColor="#8A747D"
            style={styles.input}
            value={kinName}
          />
          <TextInput
            accessibilityLabel="Trusted person phone"
            keyboardType="phone-pad"
            onChangeText={setKinPhone}
            placeholder="+233…"
            placeholderTextColor="#8A747D"
            style={styles.input}
            value={kinPhone}
          />
          <View style={styles.relationships}>
            {relationships.map((item) => (
              <Pressable
                accessibilityRole="radio"
                accessibilityState={{ selected: relationship === item }}
                key={item}
                onPress={() => setRelationship(item)}
                style={[
                  styles.relationship,
                  relationship === item && styles.relationshipSelected,
                ]}
              >
                <Text
                  style={[
                    styles.relationshipText,
                    relationship === item && styles.relationshipTextSelected,
                  ]}
                >
                  {item}
                </Text>
              </Pressable>
            ))}
          </View>
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ disabled: saving }}
            disabled={saving}
            onPress={() => void nominate()}
            style={[styles.primary, saving && styles.disabled]}
          >
            {saving ? (
              <ActivityIndicator color="#FFF5E9" />
            ) : (
              <Text style={styles.primaryText}>Send private invitation</Text>
            )}
          </Pressable>
          {notice ? <Text style={styles.notice}>{notice}</Text> : null}
        </View>

        <View style={styles.panel}>
          <View style={styles.listHeading}>
            <Text style={styles.panelLabel}>INVITATIONS</Text>
            <Pressable onPress={() => void load()}>
              <Text style={styles.refresh}>Refresh</Text>
            </Pressable>
          </View>
          {loading ? (
            <ActivityIndicator color="#8E3159" style={styles.loader} />
          ) : nominations.length === 0 ? (
            <Text style={styles.empty}>
              No invitations yet. Start with one person whose judgment feels
              calm and kind.
            </Text>
          ) : (
            nominations.map((nomination) => (
              <View key={nomination.id} style={styles.nomination}>
                <View style={styles.avatar}>
                  <Text style={styles.avatarText}>
                    {nomination.kinName.slice(0, 1).toUpperCase()}
                  </Text>
                </View>
                <View style={styles.nominationCopy}>
                  <Text style={styles.nominationName}>
                    {nomination.kinName}
                  </Text>
                  <Text style={styles.nominationMeta}>
                    {nomination.relationship} ·{" "}
                    {new Date(nomination.createdAt).toLocaleDateString()}
                  </Text>
                </View>
                <Text style={styles.status}>{nomination.status}</Text>
              </View>
            ))
          )}
        </View>

        <View style={styles.boundary}>
          <Text style={styles.boundaryEyebrow}>THE BOUNDARY</Text>
          <Text style={styles.boundaryTitle}>
            Trusted hands never enter the room.
          </Text>
          <Text style={styles.boundaryCopy}>
            Doorway answers, voice, messages, profile details and decisions
            remain private. Declining carries no consequence.
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE3", flex: 1 },
  content: { padding: 20, paddingBottom: 60 },
  back: {
    alignItems: "center",
    alignSelf: "flex-start",
    borderColor: "#9F8793",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 18,
  },
  backText: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
    marginTop: 42,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 52,
    letterSpacing: -3.2,
    lineHeight: 49,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginBottom: 20,
    marginTop: 20,
  },
  panel: {
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 24,
    borderWidth: 1,
    marginTop: 12,
    padding: 20,
  },
  panelLabel: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
  },
  input: {
    backgroundColor: "#F7EFE3",
    borderColor: "#D8C7BD",
    borderRadius: 14,
    borderWidth: 1,
    color: "#2B151F",
    fontFamily: "Outfit_500Medium",
    fontSize: 16,
    marginTop: 12,
    minHeight: 52,
    paddingHorizontal: 16,
  },
  relationships: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 8,
    marginTop: 12,
  },
  relationship: {
    borderColor: "#BBA6AE",
    borderRadius: 999,
    borderWidth: 1,
    paddingHorizontal: 13,
    paddingVertical: 9,
  },
  relationshipSelected: { backgroundColor: "#8E3159", borderColor: "#8E3159" },
  relationshipText: {
    color: "#69535D",
    fontFamily: "Outfit_600SemiBold",
    textTransform: "capitalize",
  },
  relationshipTextSelected: { color: "#FFF5E9" },
  primary: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 52,
    paddingHorizontal: 18,
  },
  primaryText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.45 },
  notice: {
    color: "#69535D",
    fontFamily: "Outfit_500Medium",
    lineHeight: 21,
    marginTop: 12,
  },
  listHeading: { flexDirection: "row", justifyContent: "space-between" },
  refresh: { color: "#8E3159", fontFamily: "Outfit_700Bold" },
  loader: { marginVertical: 28 },
  empty: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 18,
  },
  nomination: {
    alignItems: "center",
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    gap: 10,
    minHeight: 72,
  },
  avatar: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    height: 42,
    justifyContent: "center",
    width: 42,
  },
  avatarText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
  nominationCopy: { flex: 1 },
  nominationName: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  nominationMeta: {
    color: "#745F68",
    fontFamily: "Outfit_400Regular",
    textTransform: "capitalize",
  },
  status: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 12,
    textTransform: "uppercase",
  },
  boundary: {
    backgroundColor: "#38172C",
    borderRadius: 24,
    marginTop: 24,
    padding: 22,
  },
  boundaryEyebrow: {
    color: "#DFA267",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
  },
  boundaryTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 28,
    lineHeight: 32,
    marginTop: 9,
  },
  boundaryCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 10,
  },
});
