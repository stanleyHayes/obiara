import { type Href, useRouter } from "expo-router";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

interface VisibleEvent {
  id: string;
  kind: string;
  effect: string;
  sourceCategory: string;
  occurredAt: string;
}

interface Explanation {
  marks: string[];
  events: VisibleEvent[];
}

const adverseKinds = new Set([
  "ghost_pattern",
  "harassment_finding",
  "fraud_finding",
  "vouch_stake_loss",
]);

function humanize(value: string) {
  return value
    .replaceAll("_", " ")
    .replace(/^./, (letter) => letter.toUpperCase());
}

export function SubanExplanationScreen() {
  const router = useRouter();
  const [record, setRecord] = useState<Explanation | null>(null);
  const [selectedID, setSelectedID] = useState("");
  const [appealReason, setAppealReason] = useState("");
  const [appealRef, setAppealRef] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const commandID = useRef<string | null>(null);
  const selected = useMemo(
    () => record?.events.find((event) => event.id === selectedID) ?? null,
    [record, selectedID],
  );

  useEffect(() => {
    let active = true;
    void apiRequest<Explanation>("/v1/suban/explanation")
      .then((value) => {
        if (active) {
          setRecord(value);
          setSelectedID(value.events[0]?.id ?? "");
        }
      })
      .catch((loadError: unknown) => {
        if (active)
          setMessage(
            loadError instanceof Error
              ? loadError.message
              : "Your Suban record could not be loaded.",
          );
      });
    return () => {
      active = false;
    };
  }, []);

  async function appeal() {
    if (
      !selected ||
      !adverseKinds.has(selected.kind) ||
      appealReason.trim().length < 12
    )
      return;
    commandID.current ??= `suban-${Date.now()}`;
    setBusy(true);
    setMessage("");
    try {
      const result = await apiRequest<{ appealId: string }>(
        "/v1/suban/appeals",
        {
          method: "POST",
          headers: { "Idempotency-Key": commandID.current },
          body: JSON.stringify({
            eventId: selected.id,
            reason: "event_inaccurate",
          }),
        },
      );
      setAppealRef(result.appealId);
      commandID.current = null;
    } catch (appealError) {
      setMessage(
        appealError instanceof Error
          ? appealError.message
          : "The appeal could not be filed.",
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
        <Text style={styles.eyebrow}>YOUR SUBAN RECORD</Text>
        <Text accessibilityRole="header" style={styles.title}>
          See what shaped every mark.
        </Text>
        <Text style={styles.copy}>
          Suban is a reviewable event record, not a secret score or a permanent
          verdict. It never improves matching rank.
        </Text>
        <View style={styles.mark}>
          <Text style={styles.markLabel}>
            CURRENT MARKS · {record?.marks.length ? "VISIBLE" : "BUILDING"}
          </Text>
          <Text style={styles.markTitle}>
            {record?.marks.length
              ? record.marks.map(humanize).join(" · ")
              : "No visible mark"}
          </Text>
          <Text style={styles.markCopy}>
            Marks are recomputed from the append-only record and never expose a
            hidden score.
          </Text>
        </View>
        <View style={styles.card}>
          <Text style={styles.eyebrowCard}>VISIBLE EVENT HISTORY</Text>
          <Text style={styles.cardTitle}>Nothing contributing is hidden.</Text>
          {(record?.events ?? []).map((event) => (
            <Pressable
              key={event.id}
              onPress={() => {
                setSelectedID(event.id);
                setAppealReason("");
                setAppealRef("");
                commandID.current = null;
              }}
              style={[
                styles.event,
                event.id === selectedID && styles.eventActive,
              ]}
            >
              <Text style={styles.eventDate}>
                {new Date(event.occurredAt).toLocaleDateString("en-GH")} ·{" "}
                {event.sourceCategory.replaceAll("_", " ")}
              </Text>
              <Text style={styles.eventTitle}>{humanize(event.kind)}</Text>
            </Pressable>
          ))}
          {selected ? (
            <View style={styles.detail}>
              <Text style={styles.detailLabel}>BOUNDED EXPLANATION</Text>
              <Text style={styles.detailTitle}>{humanize(selected.kind)}</Text>
              <Text style={styles.detailCopy}>{humanize(selected.effect)}</Text>
            </View>
          ) : null}
        </View>
        <View style={styles.card}>
          <Text style={styles.eyebrowCard}>HUMAN APPEAL</Text>
          <Text style={styles.cardTitle}>
            The original record stays intact.
          </Text>
          {!appealRef && selected && adverseKinds.has(selected.kind) ? (
            <>
              <TextInput
                accessibilityLabel="Why should this be reviewed?"
                multiline
                onChangeText={(value) => {
                  setAppealReason(value.slice(0, 240));
                  commandID.current = null;
                }}
                placeholder="Share context without names or private messages"
                placeholderTextColor="#8A747D"
                style={styles.input}
                value={appealReason}
              />
              <Pressable
                disabled={busy || appealReason.trim().length < 12}
                onPress={() => void appeal()}
                style={[
                  styles.primary,
                  (busy || appealReason.trim().length < 12) && styles.disabled,
                ]}
              >
                <Text style={styles.primaryText}>
                  Submit appeal for human review
                </Text>
              </Pressable>
            </>
          ) : appealRef ? (
            <View style={styles.status}>
              <Text style={styles.statusRef}>{appealRef}</Text>
              <Text style={styles.statusTitle}>
                Awaiting a separate human panel
              </Text>
              <Text style={styles.statusCopy}>
                Your original record remains visible and unchanged.
              </Text>
            </View>
          ) : (
            <Text style={styles.detailCopy}>
              Only adverse reviewed events can be appealed from this record.
            </Text>
          )}
          {message ? (
            <Text accessibilityLiveRegion="polite" style={styles.error}>
              {message}
            </Text>
          ) : null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE3", flex: 1 },
  content: { padding: 20, paddingBottom: 60 },
  back: {
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
    marginTop: 38,
  },
  eyebrowCard: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 52,
    letterSpacing: -3.3,
    lineHeight: 48,
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
  mark: { backgroundColor: "#38172C", borderRadius: 24, padding: 22 },
  markLabel: {
    color: "#D9BECB",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.1,
  },
  markTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.7,
    marginTop: 10,
  },
  markCopy: {
    color: "#E9D8DF",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 10,
  },
  card: {
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 24,
    borderWidth: 1,
    marginTop: 12,
    padding: 20,
  },
  cardTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.5,
    lineHeight: 32,
    marginBottom: 16,
    marginTop: 8,
  },
  event: {
    backgroundColor: "#F3E8DD",
    borderColor: "transparent",
    borderRadius: 16,
    borderWidth: 2,
    marginBottom: 8,
    padding: 14,
  },
  eventActive: { borderColor: "#8E3159" },
  eventDate: {
    color: "#745F68",
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
  },
  eventTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 16,
    marginTop: 5,
  },
  detail: {
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    marginTop: 12,
    paddingTop: 16,
  },
  detailLabel: { color: "#8E3159", fontFamily: "Outfit_700Bold", fontSize: 11 },
  detailTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 20,
    marginTop: 6,
  },
  detailCopy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 6,
  },
  weight: { color: "#2B151F", fontFamily: "Outfit_700Bold", marginTop: 10 },
  input: {
    borderColor: "#BDA9B1",
    borderRadius: 14,
    borderWidth: 1,
    color: "#2B151F",
    fontFamily: "Outfit_400Regular",
    minHeight: 110,
    padding: 14,
    textAlignVertical: "top",
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 12,
    minHeight: 52,
    paddingHorizontal: 12,
  },
  primaryText: {
    color: "#FFF5E9",
    fontFamily: "Outfit_700Bold",
    textAlign: "center",
  },
  disabled: { opacity: 0.38 },
  error: { color: "#9D2948", fontFamily: "Outfit_600SemiBold", marginTop: 14 },
  status: { backgroundColor: "#DCEFE6", borderRadius: 16, padding: 16 },
  statusRef: { color: "#1C654F", fontFamily: "Outfit_400Regular" },
  statusTitle: {
    color: "#1C654F",
    fontFamily: "Outfit_700Bold",
    fontSize: 17,
    marginTop: 6,
  },
  statusCopy: {
    color: "#1C654F",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 6,
  },
});
