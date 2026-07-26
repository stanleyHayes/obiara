import {
  initialNotificationSettings,
  notificationSettingsReducer,
  type NotificationCategory,
  type NotificationChannel,
} from "@obiara/notification-settings";
import { type Href, useRouter } from "expo-router";
import { useReducer } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const categories: ReadonlyArray<[NotificationCategory, string]> = [
  ["courtship", "Courtship"],
  ["community", "Community"],
  ["games", "Games"],
  ["rituals", "Rituals"],
];
const channels: ReadonlyArray<[NotificationChannel, string]> = [
  ["push", "Push"],
  ["in_app", "In-app"],
  ["sms", "SMS"],
  ["whatsapp", "WhatsApp"],
];

export function NotificationSettingsScreen() {
  const router = useRouter();
  const [state, dispatch] = useReducer(
    notificationSettingsReducer,
    initialNotificationSettings,
  );

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Fie</Text>
        </Pressable>
        <Text style={styles.eyebrow}>ATTENTION BELONGS TO YOU</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Choose what may knock.
        </Text>
        <Text style={styles.copy}>
          The server enforces your quiet hours and one daily limit across every
          ordinary channel.
        </Text>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>CATEGORIES</Text>
          <Text style={styles.cardTitle}>What deserves a reminder?</Text>
          {categories.map(([id, label]) => {
            const selected = state.enabledCategories.includes(id);
            return (
              <Pressable
                accessibilityRole="switch"
                accessibilityState={{ checked: selected }}
                key={id}
                onPress={() =>
                  dispatch({ type: "toggle-category", value: id })
                }
                style={styles.row}
              >
                <Text style={styles.rowLabel}>{label}</Text>
                <Text style={[styles.state, selected && styles.stateOn]}>
                  {selected ? "On" : "Off"}
                </Text>
              </Pressable>
            );
          })}
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>CHANNELS</Text>
          <Text style={styles.cardTitle}>Where may they arrive?</Text>
          <View style={styles.channels}>
            {channels.map(([id, label]) => {
              const selected = state.enabledChannels.includes(id);
              return (
                <Pressable
                  accessibilityRole="button"
                  accessibilityState={{ selected }}
                  key={id}
                  onPress={() =>
                    dispatch({ type: "toggle-channel", value: id })
                  }
                  style={[styles.channel, selected && styles.channelSelected]}
                >
                  <Text
                    style={[
                      styles.channelText,
                      selected && styles.channelTextSelected,
                    ]}
                  >
                    {label}
                  </Text>
                </Pressable>
              );
            })}
          </View>
        </View>

        <View style={styles.rules}>
          <Text style={styles.rulesEyebrow}>QUIET HOURS · ACCRA TIME</Text>
          <Text style={styles.rulesTitle}>Rest without losing anything.</Text>
          <View style={styles.times}>
            <View style={styles.timeField}>
              <Text style={styles.timeLabel}>From</Text>
              <TextInput
                accessibilityLabel="Quiet hours start"
                onChangeText={(value) =>
                  dispatch({ type: "quiet-start", value })
                }
                style={styles.timeInput}
                value={state.quietStart}
              />
            </View>
            <View style={styles.timeField}>
              <Text style={styles.timeLabel}>Until</Text>
              <TextInput
                accessibilityLabel="Quiet hours end"
                onChangeText={(value) =>
                  dispatch({ type: "quiet-end", value })
                }
                style={styles.timeInput}
                value={state.quietEnd}
              />
            </View>
          </View>
          <View style={styles.cap}>
            <Text style={styles.capNumber}>{state.dailyCap}</Text>
            <Text style={styles.capCopy}>
              ordinary notifications per day, shared across every channel
            </Text>
          </View>
          <View style={styles.critical}>
            <Text style={styles.criticalTitle}>
              Safety and OTP service messages stay available.
            </Text>
            <Text style={styles.criticalCopy}>
              They bypass ordinary preferences only for safety or account
              access—never for view signals, jealousy, popularity or fake
              urgency.
            </Text>
          </View>
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
    fontSize: 56,
    letterSpacing: -3.6,
    lineHeight: 50,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginBottom: 26,
    marginTop: 20,
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
    fontSize: 28,
    letterSpacing: -1.4,
    marginBottom: 14,
    marginTop: 8,
  },
  row: {
    alignItems: "center",
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 58,
  },
  rowLabel: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  state: { color: "#745F68", fontFamily: "Outfit_700Bold" },
  stateOn: { color: "#27755F" },
  channels: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  channel: {
    borderColor: "#A98F9A",
    borderRadius: 999,
    borderWidth: 1,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  channelSelected: { backgroundColor: "#38172C" },
  channelText: { color: "#38172C", fontFamily: "Outfit_700Bold" },
  channelTextSelected: { color: "#FFF5E9" },
  rules: {
    backgroundColor: "#38172C",
    borderRadius: 24,
    marginTop: 12,
    padding: 22,
  },
  rulesEyebrow: {
    color: "#FF9AB0",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  rulesTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.8,
    lineHeight: 34,
    marginTop: 10,
  },
  times: { flexDirection: "row", gap: 10, marginTop: 20 },
  timeField: { flex: 1 },
  timeLabel: {
    color: "#D8C4CE",
    fontFamily: "Outfit_600SemiBold",
    marginBottom: 6,
  },
  timeInput: {
    backgroundColor: "#FFF5E9",
    borderRadius: 12,
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    minHeight: 50,
    paddingHorizontal: 12,
  },
  cap: {
    alignItems: "center",
    borderTopColor: "rgba(255,245,233,0.18)",
    borderTopWidth: 1,
    flexDirection: "row",
    gap: 16,
    marginTop: 22,
    paddingTop: 18,
  },
  capNumber: {
    color: "#FFB34F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 64,
    letterSpacing: -4,
  },
  capCopy: {
    color: "#FFF5E9",
    flex: 1,
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 21,
  },
  critical: {
    borderTopColor: "rgba(255,245,233,0.18)",
    borderTopWidth: 1,
    marginTop: 14,
    paddingTop: 18,
  },
  criticalTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_700Bold",
    fontSize: 17,
  },
  criticalCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 8,
  },
});
