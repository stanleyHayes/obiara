import { type Href, useRouter } from "expo-router";
import { useEffect, useRef, useState } from "react";
import {
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { apiRequest } from "./api";

interface Membership {
  passId: string;
  passName: string;
  status: "active" | "grace" | "expired" | "refund_pending" | "refunded";
  paidThrough: string;
  graceUntil: string;
  renewsAutomatically: boolean;
  receiptRef: string;
  refundRequestRef?: string;
}

export function MembershipSettingsScreen() {
  const router = useRouter();
  const [membership, setMembership] = useState<Membership | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [cancellationPending, setCancellationPending] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const commandID = useRef<string | null>(null);

  useEffect(() => {
    let active = true;
    void apiRequest<Membership>("/v1/membership")
      .then((value) => {
        if (active) setMembership(value);
      })
      .catch((loadError: unknown) => {
        if (active) {
          const text =
            loadError instanceof Error
              ? loadError.message
              : "Membership could not be loaded.";
          if (text.toLowerCase().includes("no membership pass"))
            setMembership(null);
          else setMessage(text);
        }
      })
      .finally(() => {
        if (active) setLoaded(true);
      });
    return () => {
      active = false;
    };
  }, []);

  async function mutate(action: "cancel" | "refund") {
    commandID.current ??= `membership-${action}-${Date.now()}`;
    setBusy(true);
    setMessage("");
    try {
      const value = await apiRequest<Membership>(
        action === "cancel"
          ? "/v1/membership/cancel"
          : "/v1/membership/refunds",
        { method: "POST", headers: { "Idempotency-Key": commandID.current } },
      );
      setMembership(value);
      setCancellationPending(false);
      commandID.current = null;
    } catch (actionError) {
      setMessage(
        actionError instanceof Error
          ? actionError.message
          : "The membership action could not be completed.",
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
        <Text style={styles.eyebrow}>YOUR MEMBERSHIP</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Know exactly what is paid for.
        </Text>
        <Text style={styles.copy}>
          Membership never buys seeds, visibility, rank, a match or access to
          another person.
        </Text>

        {!loaded ? (
          <Text style={styles.status}>Loading membership…</Text>
        ) : null}
        {message ? (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {message}
          </Text>
        ) : null}

        {loaded && !membership ? (
          <View style={styles.card}>
            <Text style={styles.eyebrowCard}>NO CURRENT PASS</Text>
            <Text style={styles.cardTitle}>
              Nothing is renewing or being charged.
            </Text>
            <Text style={styles.copySmall}>
              Purchase options appear only after an approved payment provider is
              available.
            </Text>
          </View>
        ) : null}

        {membership ? (
          <>
            <View style={styles.card}>
              <Text style={styles.eyebrowCard}>
                CURRENT PASS ·{" "}
                {membership.status.replaceAll("_", " ").toUpperCase()}
              </Text>
              <Text style={styles.cardTitle}>
                {membership.passName.replaceAll("_", " ")}
              </Text>
              {[
                [
                  "Paid through",
                  new Date(membership.paidThrough).toLocaleDateString("en-GH"),
                ],
                [
                  "Automatic renewal",
                  membership.renewsAutomatically ? "On" : "Off",
                ],
                ["Receipt", `${membership.receiptRef.slice(0, 12)}…`],
                [
                  "Grace",
                  new Date(membership.graceUntil).toLocaleDateString("en-GH"),
                ],
              ].map(([label, value]) => (
                <View key={label} style={styles.fact}>
                  <Text style={styles.factLabel}>{label}</Text>
                  <Text style={styles.factValue}>{value}</Text>
                </View>
              ))}
              {membership.renewsAutomatically ? (
                <Pressable
                  onPress={() => setCancellationPending(true)}
                  style={styles.secondary}
                >
                  <Text style={styles.secondaryText}>Review cancellation</Text>
                </Pressable>
              ) : (
                <View style={styles.notice}>
                  <Text style={styles.noticeTitle}>
                    Renewal is cancelled without penalty.
                  </Text>
                  <Text style={styles.copySmall}>
                    Purchased access remains through the paid-through date.
                  </Text>
                </View>
              )}
            </View>

            <View style={styles.card}>
              <Text style={styles.eyebrowCard}>RECEIPT AND REFUND</Text>
              <Text style={styles.cardTitle}>
                {membership.status === "refunded"
                  ? "Provider confirmed the refund."
                  : membership.status === "refund_pending"
                    ? "Refund review is pending."
                    : "Need the cancelled payment reviewed?"}
              </Text>
              <Text style={styles.copySmall}>
                A request is not a refund promise. Completion appears only after
                provider confirmation.
              </Text>
              {!membership.renewsAutomatically &&
              membership.status !== "refund_pending" &&
              membership.status !== "refunded" ? (
                <Pressable
                  disabled={busy}
                  onPress={() => void mutate("refund")}
                  style={[styles.primary, busy && styles.disabled]}
                >
                  <Text style={styles.primaryText}>
                    {busy ? "Requesting review…" : "Request refund review"}
                  </Text>
                </Pressable>
              ) : null}
              {membership.refundRequestRef ? (
                <View style={styles.notice}>
                  <Text style={styles.factLabel}>
                    {membership.refundRequestRef.slice(0, 12)}…
                  </Text>
                  <Text style={styles.noticeTitle}>
                    {membership.status === "refunded"
                      ? "Provider confirmation recorded"
                      : "Awaiting provider confirmation"}
                  </Text>
                </View>
              ) : null}
            </View>
          </>
        ) : null}

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
        onRequestClose={() => setCancellationPending(false)}
        transparent
        visible={cancellationPending}
      >
        <View style={styles.modalBackdrop}>
          <View style={styles.modalCard}>
            <Text style={styles.eyebrowCard}>BEFORE YOU CANCEL</Text>
            <Text style={styles.cardTitle}>Your paid time remains yours.</Text>
            <Text style={styles.copySmall}>
              Cancellation stops future renewal only. Access continues through{" "}
              {membership
                ? new Date(membership.paidThrough).toLocaleDateString("en-GH")
                : "the paid-through date"}
              .
            </Text>
            <Pressable
              disabled={busy}
              onPress={() => void mutate("cancel")}
              style={[styles.primary, busy && styles.disabled]}
            >
              <Text style={styles.primaryText}>
                {busy ? "Cancelling renewal…" : "Confirm cancellation"}
              </Text>
            </Pressable>
            <Pressable
              disabled={busy}
              onPress={() => setCancellationPending(false)}
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
    fontSize: 48,
    letterSpacing: -3,
    lineHeight: 46,
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
    marginTop: 6,
  },
  status: { color: "#69535D", fontFamily: "Outfit_600SemiBold" },
  error: {
    color: "#9D2948",
    fontFamily: "Outfit_600SemiBold",
    marginBottom: 12,
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
  notice: {
    backgroundColor: "#EEE1D7",
    borderRadius: 16,
    marginTop: 14,
    padding: 16,
  },
  noticeTitle: { color: "#2B151F", fontFamily: "Outfit_700Bold", fontSize: 17 },
  law: {
    backgroundColor: "#38172C",
    borderRadius: 24,
    marginTop: 12,
    padding: 22,
  },
  lawEyebrow: {
    color: "#FF9AB0",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  lawTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.5,
    lineHeight: 32,
    marginTop: 10,
  },
  lawCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 10,
  },
  modalBackdrop: {
    alignItems: "center",
    backgroundColor: "rgba(29,11,23,0.72)",
    flex: 1,
    justifyContent: "center",
    padding: 20,
  },
  modalCard: {
    backgroundColor: "#FFFAF2",
    borderRadius: 24,
    maxWidth: 420,
    padding: 22,
    width: "100%",
  },
});
