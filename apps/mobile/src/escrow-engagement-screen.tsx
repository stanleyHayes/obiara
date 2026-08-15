import { type Href, useRouter } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { apiRequest } from "./api";

type Milestone = {
  id: string;
  grossPesewas: number;
  feePesewas: number;
  deliveryConfirmed: boolean;
  acceptanceConfirmed: boolean;
  settled: boolean;
  statementRef?: string;
};
type Escrow = {
  escrowId: string;
  engagementId: string;
  fundedPesewas: number;
  settledPesewas: number;
  milestones: Milestone[];
  disputed: boolean;
  escalationRef?: string;
};
type EscrowList = { items: Escrow[] };

const money = (value: number) => `GHS ${(value / 100).toFixed(2)}`;

export function EscrowEngagementScreen() {
  const router = useRouter();
  const [items, setItems] = useState<Escrow[] | null>(null);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await apiRequest<EscrowList>("/v1/escrows")).items);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "Protected engagements could not load.",
      );
    }
  }, []);

  useEffect(() => void load(), [load]);

  async function mutate(
    escrowId: string,
    action: "accept" | "dispute",
    milestoneId?: string,
  ) {
    const key = `${action}:${escrowId}:${milestoneId ?? ""}`;
    setBusy(key);
    setError("");
    const path =
      action === "accept"
        ? `/v1/escrows/${escrowId}/milestones/${milestoneId}/acceptance`
        : `/v1/escrows/${escrowId}/disputes`;
    try {
      await apiRequest<Escrow>(path, {
        method: "POST",
        headers: {
          "Idempotency-Key": `${action}-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        },
        body: "{}",
      });
      await load();
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The escrow action could not be completed.",
      );
    } finally {
      setBusy("");
    }
  }

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie/matchmakers" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Matchmakers</Text>
        </Pressable>
        <Text style={styles.eyebrow}>PROTECTED ENGAGEMENTS</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Evidence before release.
        </Text>
        <Text style={styles.copy}>
          You can confirm acceptance or freeze an engagement. Delivery evidence
          and settlement stay with separate, audited authorities.
        </Text>
        {!items && !error ? <ActivityIndicator color="#8E3159" /> : null}
        {error ? (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {error}
          </Text>
        ) : null}
        {items?.length === 0 ? (
          <View style={styles.card}>
            <Text style={styles.cardTitle}>No funded engagement yet.</Text>
            <Text style={styles.meta}>
              A booked consultation appears here only after provider-confirmed
              funding.
            </Text>
          </View>
        ) : null}
        {items?.map((escrow) => (
          <View key={escrow.escrowId} style={styles.card}>
            <View style={styles.row}>
              <View>
                <Text style={styles.label}>FUNDED</Text>
                <Text style={styles.amount}>{money(escrow.fundedPesewas)}</Text>
              </View>
              <View>
                <Text style={styles.label}>SETTLED</Text>
                <Text style={styles.amount}>
                  {money(escrow.settledPesewas)}
                </Text>
              </View>
            </View>
            {escrow.disputed ? (
              <Text style={styles.frozen}>
                Settlement frozen · Mpanyimfo review opened
              </Text>
            ) : null}
            {escrow.milestones.map((milestone) => {
              const command = `accept:${escrow.escrowId}:${milestone.id}`;
              return (
                <View key={milestone.id} style={styles.milestone}>
                  <Text style={styles.milestoneTitle}>
                    {milestone.id.replaceAll("-", " ")}
                  </Text>
                  <Text style={styles.meta}>
                    {money(milestone.grossPesewas)} · fee{" "}
                    {money(milestone.feePesewas)}
                  </Text>
                  <Text style={styles.status}>
                    Delivery{" "}
                    {milestone.deliveryConfirmed ? "confirmed" : "pending"} ·
                    Your acceptance{" "}
                    {milestone.acceptanceConfirmed ? "confirmed" : "pending"}
                  </Text>
                  <Pressable
                    disabled={
                      escrow.disputed ||
                      milestone.acceptanceConfirmed ||
                      milestone.settled ||
                      busy !== ""
                    }
                    onPress={() =>
                      void mutate(escrow.escrowId, "accept", milestone.id)
                    }
                    style={[
                      styles.primary,
                      (escrow.disputed ||
                        milestone.acceptanceConfirmed ||
                        milestone.settled) &&
                        styles.disabled,
                    ]}
                  >
                    <Text style={styles.primaryText}>
                      {busy === command
                        ? "Confirming…"
                        : milestone.acceptanceConfirmed
                          ? "Acceptance confirmed"
                          : "Confirm my acceptance"}
                    </Text>
                  </Pressable>
                </View>
              );
            })}
            <Pressable
              disabled={escrow.disputed || busy !== ""}
              onPress={() => void mutate(escrow.escrowId, "dispute")}
              style={[styles.dispute, escrow.disputed && styles.disabled]}
            >
              <Text style={styles.disputeText}>
                {escrow.disputed
                  ? "Settlement is frozen"
                  : "Open dispute and freeze settlement"}
              </Text>
            </Pressable>
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#F5EEE4", flex: 1 },
  content: { padding: 20, paddingBottom: 64 },
  back: {
    alignSelf: "flex-start",
    borderColor: "#9F8793",
    borderRadius: 999,
    borderWidth: 1,
    minHeight: 48,
    justifyContent: "center",
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
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 52,
    letterSpacing: -3.2,
    lineHeight: 48,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginBottom: 20,
    marginTop: 18,
  },
  error: {
    color: "#A33A4B",
    fontFamily: "Outfit_600SemiBold",
    marginVertical: 12,
  },
  card: {
    backgroundColor: "#FFFAF2",
    borderColor: "#D8C7BD",
    borderRadius: 26,
    borderWidth: 1,
    marginTop: 14,
    padding: 20,
  },
  cardTitle: { color: "#2B151F", fontFamily: "Outfit_700Bold", fontSize: 23 },
  row: { flexDirection: "row", gap: 38, justifyContent: "space-between" },
  label: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  amount: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 23,
    marginTop: 4,
  },
  frozen: {
    backgroundColor: "#F3D9DC",
    borderRadius: 12,
    color: "#7A2535",
    fontFamily: "Outfit_700Bold",
    marginTop: 18,
    padding: 12,
  },
  milestone: {
    borderTopColor: "#E5D8CF",
    borderTopWidth: 1,
    marginTop: 20,
    paddingTop: 18,
  },
  milestoneTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 20,
    textTransform: "capitalize",
  },
  meta: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 14,
    lineHeight: 21,
    marginTop: 5,
  },
  status: {
    color: "#69535D",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 13,
    lineHeight: 20,
    marginTop: 12,
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#2B151F",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 50,
    paddingHorizontal: 18,
  },
  primaryText: { color: "#FFF7ED", fontFamily: "Outfit_700Bold" },
  dispute: {
    alignItems: "center",
    borderColor: "#9A3D51",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 20,
    minHeight: 50,
    paddingHorizontal: 18,
  },
  disputeText: { color: "#8B2E43", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.45 },
});
