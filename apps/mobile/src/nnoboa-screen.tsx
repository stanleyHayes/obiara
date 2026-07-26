import {
  initialNnoboaState,
  nnoboaReducer,
} from "@obiara/nnoboa-policy";
import { type Href, useRouter } from "expo-router";
import { useReducer } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export function NnoboaScreen() {
  const router = useRouter();
  const [state, dispatch] = useReducer(nnoboaReducer, initialNnoboaState);

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
          You choose who may suggest.
        </Text>
        <Text style={styles.copy}>
          Name up to three people you trust. You still hold the final yes or no,
          and no private courtship content crosses.
        </Text>

        <View style={styles.panel}>
          <Text style={styles.eyebrow}>YOUR NOMINATORS</Text>
          <Text style={styles.panelTitle}>
            {state.nominators.length} of 3 places used
          </Text>
          {state.nominators.map((nominator) => (
            <View key={nominator.id} style={styles.person}>
              <View style={styles.avatar}>
                <Text style={styles.avatarText}>{nominator.label[0]}</Text>
              </View>
              <View style={styles.personCopy}>
                <Text style={styles.personName}>{nominator.label}</Text>
                <Text style={styles.personMeta}>
                  {nominator.channel === "whatsapp"
                    ? "OTP-gated WhatsApp"
                    : "In-app"}
                </Text>
              </View>
              <Pressable
                accessibilityRole="button"
                onPress={() =>
                  dispatch({ type: "remove-nominator", id: nominator.id })
                }
              >
                <Text style={styles.remove}>Remove</Text>
              </Pressable>
            </View>
          ))}
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ disabled: state.nominators.length >= 3 }}
            disabled={state.nominators.length >= 3}
            onPress={() =>
              dispatch({
                type: "add-nominator",
                nominator: {
                  id: "nom-ama",
                  label: "Ama K.",
                  channel: "app",
                },
              })
            }
            style={[
              styles.primary,
              state.nominators.length >= 3 && styles.disabled,
            ]}
          >
            <Text style={styles.primaryText}>Add a trusted nominator</Text>
          </Pressable>
        </View>

        <View style={styles.panel}>
          <Text style={styles.eyebrow}>A NOMINATION IS WAITING</Text>
          <Text style={styles.panelTitle}>{state.candidate?.reference}</Text>
          {state.memberDecision === "vetoed" ? (
            <View style={styles.result}>
              <Text style={styles.resultTitle}>Your no is enough.</Text>
              <Text style={styles.resultCopy}>
                Closed privately. No reason or negative mark is attached.
              </Text>
            </View>
          ) : state.memberDecision === "accepted" ? (
            <View style={styles.result}>
              <Text style={styles.resultTitle}>Mutual permission confirmed.</Text>
              <Text style={styles.resultCopy}>
                This does not open a room, spend a seed or send a message.
              </Text>
            </View>
          ) : (
            <>
              <View style={styles.fact}>
                <Text style={styles.factLabel}>Age band</Text>
                <Text style={styles.factValue}>{state.candidate?.ageBand}</Text>
              </View>
              <View style={styles.fact}>
                <Text style={styles.factLabel}>City</Text>
                <Text style={styles.factValue}>{state.candidate?.city}</Text>
              </View>
              <Text style={styles.consentCopy}>
                Name, contact and profile stay hidden until the nominee
                explicitly consents.
              </Text>
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ selected: state.nomineeConsented }}
                onPress={() =>
                  dispatch({
                    type: "nominee-consent",
                    value: !state.nomineeConsented,
                  })
                }
                style={styles.secondary}
              >
                <Text style={styles.secondaryText}>
                  {state.nomineeConsented
                    ? "Nominee consent confirmed"
                    : "Preview nominee consent"}
                </Text>
              </Pressable>
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ disabled: !state.nomineeConsented }}
                disabled={!state.nomineeConsented}
                onPress={() => dispatch({ type: "member-accept" })}
                style={[
                  styles.primary,
                  !state.nomineeConsented && styles.disabled,
                ]}
              >
                <Text style={styles.primaryText}>Review introduction</Text>
              </Pressable>
              <Pressable
                accessibilityRole="button"
                onPress={() => dispatch({ type: "member-veto" })}
                style={styles.veto}
              >
                <Text style={styles.vetoText}>Decline privately</Text>
              </Pressable>
            </>
          )}
        </View>

        <View style={styles.boundary}>
          <Text style={styles.boundaryEyebrow}>THE BOUNDARY</Text>
          <Text style={styles.boundaryTitle}>
            Aunties introduce. They never enter the courtship.
          </Text>
          <Text style={styles.boundaryCopy}>
            No doorway answers, room content, private voice, messages or
            decision reason is shown.
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
    marginTop: 42,
  },
  title: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 54,
    letterSpacing: -3.4,
    lineHeight: 49,
    marginTop: 14,
  },
  copy: {
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 27,
    marginBottom: 28,
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
  panelTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.5,
    marginBottom: 18,
    marginTop: 8,
  },
  person: {
    alignItems: "center",
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    gap: 10,
    minHeight: 68,
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
  personCopy: { flex: 1 },
  personName: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  personMeta: { color: "#745F68", fontFamily: "Outfit_400Regular" },
  remove: { color: "#8E3159", fontFamily: "Outfit_700Bold" },
  primary: {
    alignItems: "center",
    backgroundColor: "#38172C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 52,
    paddingHorizontal: 18,
  },
  primaryText: { color: "#FFF5E9", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.4 },
  fact: {
    borderTopColor: "#E4D6CC",
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 12,
  },
  factLabel: { color: "#745F68", fontFamily: "Outfit_400Regular" },
  factValue: { color: "#2B151F", fontFamily: "Outfit_700Bold" },
  consentCopy: {
    backgroundColor: "#F2E6DB",
    borderRadius: 16,
    color: "#69535D",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 10,
    padding: 16,
  },
  secondary: {
    alignItems: "center",
    borderColor: "#38172C",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 52,
  },
  secondaryText: { color: "#38172C", fontFamily: "Outfit_700Bold" },
  veto: { alignItems: "center", minHeight: 48, justifyContent: "center" },
  vetoText: { color: "#8E3159", fontFamily: "Outfit_700Bold" },
  result: {
    backgroundColor: "#38172C",
    borderRadius: 18,
    padding: 20,
  },
  resultTitle: {
    color: "#FFF5E9",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 28,
  },
  resultCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 10,
  },
  boundary: {
    backgroundColor: "#D98A42",
    borderRadius: 24,
    marginTop: 18,
    padding: 24,
  },
  boundaryEyebrow: {
    color: "#2B151F",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
  },
  boundaryTitle: {
    color: "#2B151F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.8,
    lineHeight: 34,
    marginTop: 10,
  },
  boundaryCopy: {
    color: "#4E2A21",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 12,
  },
});
