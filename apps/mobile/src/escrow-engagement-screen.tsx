import {
  canPreviewSettlement,
  escrowReducer,
  formatGhs,
  initialEscrowState,
} from "@obiara/escrow-engagement";
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

export function EscrowEngagementScreen() {
  const router = useRouter();
  const [state, dispatch] = useReducer(escrowReducer, initialEscrowState);
  const selected = state.milestones.find(
    (item) => item.id === state.selectedMilestone,
  )!;
  const frozen = state.disputeState !== "none";

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie/matchmakers" as Href)}
          style={styles.back}
        >
          <Text style={styles.backText}>Matchmakers</Text>
        </Pressable>
        <Text style={styles.eyebrow}>PROTECTED ENGAGEMENT</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Money moves only after shared evidence.
        </Text>
        <Text style={styles.copy}>
          Terms are locked. This preview never releases money or contacts a
          provider.
        </Text>

        <View style={styles.totals}>
          {[
            ["Funded", formatGhs(state.fundedPesewas)],
            ["Platform fee", formatGhs(state.platformFeePesewas)],
            ["Payout", formatGhs(state.payoutPesewas)],
            ["Statement", state.payoutStatementRef],
          ].map(([label, value]) => (
            <View key={label} style={styles.total}>
              <Text style={styles.totalLabel}>{label}</Text>
              <Text style={styles.totalValue}>{value}</Text>
            </View>
          ))}
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrowCard}>NAMED MILESTONES</Text>
          <Text style={styles.cardTitle}>Confirm what was delivered.</Text>
          {state.milestones.map((item) => (
            <Pressable
              key={item.id}
              onPress={() => dispatch({ type: "select", id: item.id })}
              style={[
                styles.choice,
                item.id === state.selectedMilestone && styles.choiceActive,
              ]}
            >
              <Text style={styles.choiceTitle}>{item.name}</Text>
              <Text style={styles.choiceAmount}>
                {formatGhs(item.amountPesewas)}
              </Text>
            </Pressable>
          ))}
          <Text style={styles.sectionTitle}>{selected.name}</Text>
          <Pressable
            disabled={selected.memberConfirmed || frozen}
            onPress={() => dispatch({ type: "confirm-member" })}
            style={[
              styles.primary,
              (selected.memberConfirmed || frozen) && styles.disabled,
            ]}
          >
            <Text style={styles.primaryText}>
              {selected.memberConfirmed
                ? "Member confirmed"
                : "Confirm as member"}
            </Text>
          </Pressable>
          <Pressable
            disabled={selected.matchmakerConfirmed || frozen}
            onPress={() => dispatch({ type: "confirm-matchmaker" })}
            style={[
              styles.secondary,
              (selected.matchmakerConfirmed || frozen) && styles.disabled,
            ]}
          >
            <Text style={styles.secondaryText}>
              {selected.matchmakerConfirmed
                ? "Matchmaker confirmed"
                : "Preview matchmaker confirmation"}
            </Text>
          </Pressable>
          <Pressable
            disabled={!canPreviewSettlement(state)}
            onPress={() => dispatch({ type: "preview-settlement" })}
            style={[
              styles.primary,
              !canPreviewSettlement(state) && styles.disabled,
            ]}
          >
            <Text style={styles.primaryText}>Preview settlement</Text>
          </Pressable>
          {state.settlementPreview ? (
            <Text style={styles.notice}>
              Eligible for backend review. No money has moved.
            </Text>
          ) : null}
        </View>

        <View style={styles.card}>
          <Text style={styles.eyebrowCard}>DISPUTE PROTECTION</Text>
          <Text style={styles.cardTitle}>
            {frozen ? "Settlement is frozen." : "Something does not match?"}
          </Text>
          {state.disputeState === "none" ? (
            <>
              <TextInput
                accessibilityLabel="Dispute reason"
                multiline
                onChangeText={(value) =>
                  dispatch({ type: "dispute-reason", value })
                }
                placeholder="Describe what differs from the agreed milestone"
                placeholderTextColor="#8A747D"
                style={styles.input}
                value={state.disputeReason}
              />
              <Pressable
                disabled={state.disputeReason.trim().length < 12}
                onPress={() => dispatch({ type: "open-dispute" })}
                style={[
                  styles.primary,
                  state.disputeReason.trim().length < 12 && styles.disabled,
                ]}
              >
                <Text style={styles.primaryText}>
                  Open dispute and freeze settlement
                </Text>
              </Pressable>
            </>
          ) : state.disputeState === "open" ? (
            <Pressable
              onPress={() => dispatch({ type: "escalate-dispute" })}
              style={styles.primary}
            >
              <Text style={styles.primaryText}>
                Escalate to Mpanyimfo review
              </Text>
            </Pressable>
          ) : (
            <Text style={styles.notice}>
              {state.escalationRef} · awaiting a separate panel
            </Text>
          )}
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
  eyebrowCard: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 50,
    letterSpacing: -3.2,
    lineHeight: 47,
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
  totals: { backgroundColor: "#38172C", borderRadius: 24, overflow: "hidden" },
  total: { borderBottomColor: "#68415A", borderBottomWidth: 1, padding: 18 },
  totalLabel: {
    color: "#D9BECB",
    fontFamily: "Outfit_400Regular",
    fontSize: 12,
  },
  totalValue: {
    color: "#FFF5E9",
    fontFamily: "Outfit_700Bold",
    fontSize: 22,
    marginTop: 5,
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
  choice: {
    backgroundColor: "#F3E8DD",
    borderColor: "transparent",
    borderRadius: 16,
    borderWidth: 2,
    marginBottom: 8,
    padding: 14,
  },
  choiceActive: { borderColor: "#8E3159" },
  choiceTitle: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  choiceAmount: {
    color: "#745F68",
    fontFamily: "Outfit_400Regular",
    marginTop: 4,
  },
  sectionTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 19,
    marginTop: 15,
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
  secondary: {
    alignItems: "center",
    borderColor: "#8E3159",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 12,
    minHeight: 52,
    paddingHorizontal: 12,
  },
  secondaryText: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    textAlign: "center",
  },
  disabled: { opacity: 0.38 },
  notice: {
    backgroundColor: "#DCEFE6",
    borderRadius: 16,
    color: "#1C654F",
    fontFamily: "Outfit_700Bold",
    lineHeight: 22,
    marginTop: 14,
    padding: 16,
  },
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
});
