import {
  initialMembershipState,
  membershipReducer,
} from "@obiara/membership-settings";
import { type Href, useRouter } from "expo-router";
import { useReducer } from "react";
import {
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export function MembershipSettingsScreen() {
  const router = useRouter();
  const [state, dispatch] = useReducer(
    membershipReducer,
    initialMembershipState,
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
        <Text style={styles.eyebrow}>YOUR MEMBERSHIP</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Know exactly what is paid for.
        </Text>
        <Text style={styles.copy}>
          Membership never buys seeds, visibility, rank, a match or access to
          another person.
        </Text>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>CURRENT PASS</Text>
          <Text style={styles.cardTitle}>{state.passName}</Text>
          {[
            ["Paid through", state.paidThrough],
            ["Automatic renewal", state.renewsAutomatically ? "On" : "Off"],
            ["Receipt", state.receiptRef],
            ["Grace", state.graceEnds ?? "Not currently in grace"],
          ].map(([label, value]) => (
            <View key={label} style={styles.fact}>
              <Text style={styles.factLabel}>{label}</Text>
              <Text style={styles.factValue}>{value}</Text>
            </View>
          ))}
          {state.status === "cancelled" ? (
            <View style={styles.notice}>
              <Text style={styles.noticeTitle}>
                Cancellation recorded without penalty.
              </Text>
              <Text style={styles.noticeCopy}>
                Access remains through {state.paidThrough}. Standing and safety
                remain unchanged.
              </Text>
            </View>
          ) : (
            <Pressable
              onPress={() => dispatch({ type: "request-cancellation" })}
              style={styles.secondary}
            >
              <Text style={styles.secondaryText}>Review cancellation</Text>
            </Pressable>
          )}
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrow}>RECEIPT AND REFUND</Text>
          <Text style={styles.cardTitle}>
            {state.refundState === "provider_confirmed"
              ? "Provider confirmed the refund."
              : state.refundState === "pending"
                ? "Refund review is pending."
                : "Need a payment reviewed?"}
          </Text>
          <Text style={styles.copySmall}>
            A request becomes complete only after provider confirmation.
          </Text>
          {state.refundState === "none" ? (
            <>
              <TextInput
                accessibilityLabel="Reason for refund review"
                multiline
                onChangeText={(value) =>
                  dispatch({ type: "refund-reason", value })
                }
                placeholder="Describe the issue without payment details"
                placeholderTextColor="#8A747D"
                style={styles.input}
                value={state.refundReason}
              />
              <Pressable
                accessibilityState={{
                  disabled: state.refundReason.trim().length < 12,
                }}
                disabled={state.refundReason.trim().length < 12}
                onPress={() => dispatch({ type: "request-refund" })}
                style={[
                  styles.primary,
                  state.refundReason.trim().length < 12 && styles.disabled,
                ]}
              >
                <Text style={styles.primaryText}>Request refund review</Text>
              </Pressable>
            </>
          ) : (
            <View style={styles.notice}>
              <Text style={styles.factLabel}>{state.refundRef}</Text>
              <Text style={styles.noticeTitle}>
                {state.refundState === "pending"
                  ? "Awaiting provider confirmation"
                  : "Provider confirmation recorded"}
              </Text>
              {state.refundState === "pending" ? (
                <Pressable
                  onPress={() =>
                    dispatch({ type: "provider-confirm-refund" })
                  }
                  style={styles.primary}
                >
                  <Text style={styles.primaryText}>
                    Preview provider confirmation
                  </Text>
                </Pressable>
              ) : null}
            </View>
          )}
        </View>

        <View style={styles.law}>
          <Text style={styles.lawEyebrow}>WHAT NEVER CHANGES</Text>
          <Text style={styles.lawTitle}>
            No purchase improves romantic standing.
          </Text>
          <Text style={styles.lawCopy}>
            Cancellation has no penalty, grace never hides expiry, and raw
            payment data never appears here.
          </Text>
        </View>
      </ScrollView>

      <Modal
        animationType="fade"
        onRequestClose={() => dispatch({ type: "keep-membership" })}
        transparent
        visible={state.cancellationPending}
      >
        <View style={styles.modalBackdrop}>
          <View style={styles.modalCard}>
            <Text style={styles.eyebrow}>BEFORE YOU CANCEL</Text>
            <Text style={styles.cardTitle}>Your paid time remains yours.</Text>
            <Text style={styles.copySmall}>
              Cancellation stops future renewal only. Access continues through{" "}
              {state.paidThrough}.
            </Text>
            <Pressable
              onPress={() => dispatch({ type: "confirm-cancellation" })}
              style={styles.primary}
            >
              <Text style={styles.primaryText}>Confirm cancellation</Text>
            </Pressable>
            <Pressable
              onPress={() => dispatch({ type: "keep-membership" })}
              style={styles.secondary}
            >
              <Text style={styles.secondaryText}>Keep membership</Text>
            </Pressable>
          </View>
        </View>
      </Modal>
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
  copySmall: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
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
  fact: {
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 12,
  },
  factLabel: { color: "#745F68", fontFamily: "Outfit_400Regular" },
  factValue: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    maxWidth: "60%",
    textAlign: "right",
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 52,
  },
  primaryText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
  secondary: {
    alignItems: "center",
    borderColor: "#8E3159",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 12,
    minHeight: 52,
  },
  secondaryText: { color: "#8E3159", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.4 },
  input: {
    borderColor: "#BDA9B1",
    borderRadius: 14,
    borderWidth: 1,
    color: "#2B151F",
    fontFamily: "Outfit_400Regular",
    minHeight: 100,
    marginTop: 16,
    padding: 14,
    textAlignVertical: "top",
  },
  notice: {
    backgroundColor: "#EEE1D7",
    borderRadius: 16,
    marginTop: 14,
    padding: 16,
  },
  noticeTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 17,
  },
  noticeCopy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 8,
  },
  law: {
    backgroundColor: "#D98A42",
    borderRadius: 24,
    marginTop: 12,
    padding: 22,
  },
  lawEyebrow: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  lawTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.8,
    lineHeight: 34,
    marginTop: 10,
  },
  lawCopy: {
    color: "#4E2A21",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 12,
  },
  modalBackdrop: {
    alignItems: "center",
    backgroundColor: "rgba(32,12,24,0.72)",
    flex: 1,
    justifyContent: "center",
    padding: 20,
  },
  modalCard: {
    backgroundColor: "#FFFAF2",
    borderRadius: 24,
    padding: 22,
    width: "100%",
  },
});
