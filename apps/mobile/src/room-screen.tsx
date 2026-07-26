import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import {
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const messages = [
  ["Ama", "Home feels like people making room for one another."],
  ["You", "Mine sounds like highlife and too many voices in one kitchen."],
  ["Ama", "What is one tradition you would keep?"],
] as const;
const themes = [
  ["01", "What home carries", "REVEALED"],
  ["02", "How care feels", "READY"],
  ["03", "What we protect", "RESTING"],
  ["04", "What we might build", "RESTING"],
] as const;

export function RoomScreen() {
  const router = useRouter();
  const [safetyOpen, setSafetyOpen] = useState(false);
  const [safetyStep, setSafetyStep] = useState<
    "menu" | "report" | "reported" | "blocked"
  >("menu");
  const [category, setCategory] = useState<string | null>(null);
  const [callState, setCallState] = useState<
    "incoming" | "active" | "declined" | "ended"
  >("incoming");
  const [callCaptions, setCallCaptions] = useState(true);
  const openSafety = () => {
    setSafetyStep("menu");
    setCategory(null);
    setSafetyOpen(true);
  };
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            onPress={() => router.push("/fie/dan-mu" as Href)}
            style={styles.control}
          >
            <Text style={styles.controlText}>Dan mu</Text>
          </Pressable>
          <Pressable
            accessibilityRole="button"
            onPress={openSafety}
            style={styles.control}
          >
            <Text style={styles.controlText}>Safety</Text>
          </Pressable>
        </View>
        <Text style={styles.eyebrow}>GUIDED ROOM · THEME ONE</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Make room for honesty.
        </Text>
        <Text style={styles.body}>
          Strict alternation, no read receipts, streaks or public activity.
        </Text>
        <View style={styles.status}>
          <Text style={styles.statusLabel}>YOUR TURN</Text>
          <Text style={styles.statusCopy}>
            Nothing needs an immediate response.
          </Text>
        </View>
        <View style={styles.callSection}>
          <Text style={styles.eyebrow}>PRIVATE CALL · OBIARA NAMES ONLY</Text>
          <Text accessibilityRole="header" style={styles.callTitle}>
            {callState === "incoming"
              ? "Ama would like to talk."
              : callState === "active"
                ? "You’re together now."
                : callState === "declined"
                  ? "The invitation rested."
                  : "The call has ended."}
          </Text>
          <Text style={styles.callCopy}>
            {callState === "incoming"
              ? "This audio call starts only if you accept. Your phone number, email and contacts are never shared."
              : callState === "active"
                ? "Connected inside this room. Captions, Safety and leave stay within reach."
                : "No contact details were shared. You can keep talking here."}
          </Text>
          <View style={styles.callPanel}>
            {callState === "incoming" ? (
              <>
                <View style={styles.callIdentity}>
                  <View style={styles.callAvatar}>
                    <Text style={styles.callAvatarText}>A</Text>
                  </View>
                  <View>
                    <Text style={styles.callName}>Ama</Text>
                    <Text style={styles.callMeta}>
                      Audio invitation · no phone number
                    </Text>
                  </View>
                </View>
                <Pressable
                  accessibilityRole="button"
                  onPress={() => setCallState("active")}
                  style={styles.callPrimary}
                >
                  <Text style={styles.callPrimaryText}>Accept audio call</Text>
                </Pressable>
                <Pressable
                  accessibilityRole="button"
                  onPress={() => setCallState("declined")}
                  style={styles.callSecondary}
                >
                  <Text style={styles.callSecondaryText}>Not now</Text>
                </Pressable>
              </>
            ) : null}
            {callState === "active" ? (
              <>
                <View style={styles.callIdentity}>
                  <View style={styles.callLive} />
                  <View>
                    <Text style={styles.callName}>Private audio · 00:24</Text>
                    <Text style={styles.callMeta}>
                      Ama sees only your Obiara name
                    </Text>
                  </View>
                </View>
                <Text accessibilityLiveRegion="polite" style={styles.caption}>
                  {callCaptions
                    ? "Ama: I wanted to hear how your week has been."
                    : "Captions are off."}
                </Text>
                <Pressable
                  onPress={() => setCallCaptions((visible) => !visible)}
                  style={styles.callSecondary}
                >
                  <Text style={styles.callSecondaryText}>
                    {callCaptions ? "Hide captions" : "Show captions"}
                  </Text>
                </Pressable>
                <Pressable onPress={openSafety} style={styles.callSecondary}>
                  <Text style={styles.callSecondaryText}>Safety</Text>
                </Pressable>
                <Pressable
                  onPress={() => setCallState("ended")}
                  style={styles.callPrimary}
                >
                  <Text style={styles.callPrimaryText}>Leave call</Text>
                </Pressable>
              </>
            ) : null}
            {callState === "declined" || callState === "ended" ? (
              <Text style={styles.callRest}>
                {callState === "declined"
                  ? "Ama sees only that you were not available."
                  : "Disconnected safely. Nothing was recorded."}
              </Text>
            ) : null}
          </View>
        </View>
        <View style={styles.timeline}>
          <Text
            accessibilityElementsHidden
            importantForAccessibility="no-hide-descendants"
            style={styles.watermark}
          >
            OBIARA · PRIVATE ROOM
          </Text>
          {messages.map(([who, message]) => (
            <View
              key={message}
              style={[styles.message, who === "You" && styles.mine]}
            >
              <Text style={[styles.who, who === "You" && styles.mineText]}>
                {who}
              </Text>
              <Text
                style={[styles.messageText, who === "You" && styles.mineText]}
              >
                {message}
              </Text>
              <Text style={[styles.meta, who === "You" && styles.mineMeta]}>
                Voice · private transcript
              </Text>
            </View>
          ))}
        </View>
        <View style={styles.themes}>
          <Text style={styles.safetyEyebrow}>A GUIDED ARC, NEVER A RACE</Text>
          <Text accessibilityRole="header" style={styles.themesTitle}>
            Four ways to listen deeper.
          </Text>
          <Text style={styles.themesCopy}>
            Each reflection stays folded until you both answer. Themes open in
            order after a shared reveal.
          </Text>
          {themes.map(([number, title, state]) => (
            <View
              key={number}
              style={[styles.themeCard, state === "READY" && styles.themeReady]}
            >
              <View style={styles.themeMeta}>
                <Text
                  style={[
                    styles.themeNumber,
                    state === "READY" && styles.themeReadyText,
                  ]}
                >
                  THEME {number}
                </Text>
                <Text
                  style={[
                    styles.themeNumber,
                    state === "READY" && styles.themeReadyText,
                  ]}
                >
                  {state}
                </Text>
              </View>
              <Text
                style={[
                  styles.themeTitle,
                  state === "READY" && styles.themeReadyText,
                ]}
              >
                {title}
              </Text>
              <Text
                style={[
                  styles.themeDescription,
                  state === "READY" && styles.themeReadyDescription,
                ]}
              >
                {state === "REVEALED"
                  ? "Both reflections are visible."
                  : state === "READY"
                    ? "Ready whenever you both want to continue."
                    : `Opens after theme ${Number(number) - 1} is revealed.`}
              </Text>
            </View>
          ))}
        </View>
        <View style={styles.composer}>
          <Text style={styles.composerTitle}>Speak when it feels right.</Text>
          <Text style={styles.composerCopy}>
            A voice draft is saved before one deliberate send.
          </Text>
          <Pressable style={styles.record}>
            <Text style={styles.recordText}>Record voice reply</Text>
          </Pressable>
          <Pressable style={styles.pause}>
            <Text style={styles.pauseText}>Pause this room</Text>
          </Pressable>
        </View>
        <Modal
          animationType="slide"
          onRequestClose={() => setSafetyOpen(false)}
          transparent
          visible={safetyOpen}
        >
          <View style={styles.modalBackdrop}>
            <View accessibilityViewIsModal style={styles.safetySheet}>
              {safetyStep === "menu" ? (
                <>
                  <Text style={styles.safetyEyebrow}>SAFETY STAYS CLOSE</Text>
                  <Text accessibilityRole="header" style={styles.safetyTitle}>
                    You control your boundary.
                  </Text>
                  <Text style={styles.safetyCopy}>
                    Blocking ends contact immediately. Reports go to care review
                    without notifying the other person.
                  </Text>
                  <Pressable
                    onPress={() => setSafetyStep("blocked")}
                    style={styles.dangerAction}
                  >
                    <Text style={styles.actionText}>Block and leave</Text>
                  </Pressable>
                  <Pressable
                    onPress={() => setSafetyStep("report")}
                    style={styles.sheetAction}
                  >
                    <Text style={styles.actionText}>Report concern</Text>
                  </Pressable>
                </>
              ) : null}
              {safetyStep === "report" ? (
                <>
                  <Text style={styles.safetyEyebrow}>PRIVATE CARE REPORT</Text>
                  <Text accessibilityRole="header" style={styles.safetyTitle}>
                    What should care review?
                  </Text>
                  {["Harassment", "Identity concern", "Threat", "Other"].map(
                    (item) => (
                      <Pressable
                        accessibilityRole="radio"
                        accessibilityState={{ checked: category === item }}
                        key={item}
                        onPress={() => setCategory(item)}
                        style={[
                          styles.category,
                          category === item && styles.categorySelected,
                        ]}
                      >
                        <Text style={styles.categoryText}>{item}</Text>
                      </Pressable>
                    ),
                  )}
                  <Pressable
                    disabled={!category}
                    onPress={() => setSafetyStep("reported")}
                    style={[styles.sheetAction, !category && styles.disabled]}
                  >
                    <Text style={styles.actionText}>Send to care review</Text>
                  </Pressable>
                </>
              ) : null}
              {safetyStep === "reported" ? (
                <>
                  <Text style={styles.safetyEyebrow}>REPORT RECEIVED</Text>
                  <Text accessibilityRole="header" style={styles.safetyTitle}>
                    Care review has it.
                  </Text>
                  <Text style={styles.safetyCopy}>
                    Your category and protected evidence stay private from the
                    other person.
                  </Text>
                </>
              ) : null}
              {safetyStep === "blocked" ? (
                <>
                  <Text style={styles.safetyEyebrow}>CONTACT BLOCKED</Text>
                  <Text accessibilityRole="header" style={styles.safetyTitle}>
                    This room is closed.
                  </Text>
                  <Text style={styles.safetyCopy}>
                    No new messages can reach you here.
                  </Text>
                </>
              ) : null}
              <Pressable
                onPress={() => setSafetyOpen(false)}
                style={styles.closeSheet}
              >
                <Text style={styles.closeText}>Close safety sheet</Text>
              </Pressable>
            </View>
          </View>
        </Modal>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#1D0B18", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  topbar: { flexDirection: "row", justifyContent: "space-between" },
  control: {
    alignItems: "center",
    borderColor: "rgba(255,243,230,.3)",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 18,
  },
  controlText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#FF849B",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
    marginTop: 54,
  },
  title: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 58,
    letterSpacing: -3.8,
    lineHeight: 52,
    marginTop: 16,
  },
  body: {
    color: "rgba(255,243,230,.65)",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginTop: 24,
  },
  status: {
    borderColor: "rgba(255,173,61,.4)",
    borderRadius: 18,
    borderWidth: 1,
    marginTop: 30,
    padding: 18,
  },
  statusLabel: { color: "#FFB44F", fontFamily: "Outfit_700Bold", fontSize: 12 },
  statusCopy: {
    color: "rgba(255,243,230,.6)",
    fontFamily: "Outfit_400Regular",
    marginTop: 6,
  },
  timeline: { gap: 10, marginTop: 32 },
  callSection: {
    borderTopColor: "rgba(255,243,230,.12)",
    borderTopWidth: 1,
    marginTop: 42,
    paddingTop: 36,
  },
  callTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  callCopy: {
    color: "rgba(255,243,230,.68)",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginTop: 14,
  },
  callPanel: {
    backgroundColor: "#FFF0D9",
    borderRadius: 24,
    marginTop: 24,
    padding: 22,
  },
  callIdentity: { alignItems: "center", flexDirection: "row", gap: 14 },
  callAvatar: {
    alignItems: "center",
    backgroundColor: "#6D244F",
    borderRadius: 26,
    height: 52,
    justifyContent: "center",
    width: 52,
  },
  callAvatarText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 20,
  },
  callLive: {
    backgroundColor: "#168565",
    borderColor: "#CCEADD",
    borderRadius: 26,
    borderWidth: 12,
    height: 52,
    width: 52,
  },
  callName: {
    color: "#2A1022",
    fontFamily: "Outfit_700Bold",
    fontSize: 17,
  },
  callMeta: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    marginTop: 3,
  },
  caption: {
    backgroundColor: "#2A1022",
    borderRadius: 16,
    color: "#FFF3E6",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 24,
    padding: 16,
  },
  callPrimary: {
    alignItems: "center",
    backgroundColor: "#6D244F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 10,
    minHeight: 52,
  },
  callPrimaryText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  callSecondary: {
    alignItems: "center",
    borderColor: "#6D244F",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 10,
    minHeight: 52,
  },
  callSecondaryText: { color: "#2A1022", fontFamily: "Outfit_700Bold" },
  callRest: {
    color: "#624B59",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
  },
  watermark: {
    color: "rgba(255,243,230,.38)",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 2,
    marginBottom: 4,
  },
  themes: {
    borderTopColor: "rgba(255,243,230,.12)",
    borderTopWidth: 1,
    marginTop: 42,
    paddingTop: 36,
  },
  themesTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  themesCopy: {
    color: "rgba(255,243,230,.68)",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 18,
    marginTop: 14,
  },
  themeCard: {
    borderColor: "rgba(255,243,230,.18)",
    borderRadius: 18,
    borderWidth: 1,
    marginTop: 10,
    minHeight: 170,
    padding: 20,
  },
  themeReady: { backgroundColor: "#FFF0D9" },
  themeMeta: { flexDirection: "row", justifyContent: "space-between" },
  themeNumber: {
    color: "rgba(255,243,230,.72)",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.4,
  },
  themeTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    fontSize: 24,
    letterSpacing: -0.8,
    marginTop: 30,
  },
  themeDescription: {
    color: "rgba(255,243,230,.62)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 8,
  },
  themeReadyText: { color: "#2A1022" },
  themeReadyDescription: { color: "#624B59" },
  message: {
    alignSelf: "flex-start",
    backgroundColor: "rgba(255,243,230,.09)",
    borderRadius: 18,
    maxWidth: "88%",
    padding: 20,
  },
  mine: { alignSelf: "flex-end", backgroundColor: "#FFF0D9" },
  who: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  messageText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginTop: 18,
  },
  meta: {
    color: "rgba(255,243,230,.5)",
    fontFamily: "Outfit_400Regular",
    fontSize: 11,
    marginTop: 18,
  },
  mineText: { color: "#2A1022" },
  mineMeta: { color: "#765F70" },
  composer: {
    borderTopColor: "rgba(255,243,230,.12)",
    borderTopWidth: 1,
    marginTop: 42,
    paddingTop: 36,
  },
  composerTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.6,
  },
  composerCopy: {
    color: "rgba(255,243,230,.6)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 10,
  },
  record: {
    alignItems: "center",
    backgroundColor: "#FFAD3D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 26,
    minHeight: 52,
  },
  recordText: { color: "#2A1022", fontFamily: "Outfit_700Bold" },
  pause: {
    alignItems: "center",
    borderColor: "rgba(255,243,230,.3)",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 10,
    minHeight: 52,
  },
  pauseText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  modalBackdrop: {
    backgroundColor: "rgba(15,5,12,.82)",
    flex: 1,
    justifyContent: "flex-end",
    padding: 12,
  },
  safetySheet: {
    backgroundColor: "#FFF3E6",
    borderRadius: 28,
    padding: 24,
    paddingBottom: 32,
  },
  safetyEyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
  },
  safetyTitle: {
    color: "#2A1022",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  safetyCopy: {
    color: "#624B59",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 20,
    marginTop: 14,
  },
  dangerAction: {
    alignItems: "center",
    backgroundColor: "#7A183E",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 10,
    minHeight: 52,
  },
  sheetAction: {
    alignItems: "center",
    backgroundColor: "#2A1022",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 10,
    minHeight: 52,
  },
  actionText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold" },
  category: {
    borderColor: "#D7C7CF",
    borderRadius: 14,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 8,
    minHeight: 50,
    paddingHorizontal: 16,
  },
  categorySelected: { backgroundColor: "#F1DCE6", borderColor: "#9B315D" },
  categoryText: { color: "#2A1022", fontFamily: "Outfit_600SemiBold" },
  disabled: { opacity: 0.4 },
  closeSheet: {
    alignItems: "center",
    justifyContent: "center",
    marginTop: 14,
    minHeight: 48,
  },
  closeText: { color: "#2A1022", fontFamily: "Outfit_700Bold" },
});
