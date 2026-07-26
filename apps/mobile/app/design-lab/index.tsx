import { Link } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const flows = [
  {
    href: "/design-lab/hold",
    label: "Hold",
    detail: "Pause without accidental consequence",
  },
  {
    href: "/design-lab/sow",
    label: "Sow",
    detail: "Stage before deliberate release",
  },
  { href: "/design-lab/stone", label: "Stone", detail: "Close a turn slowly" },
  {
    href: "/design-lab/gather",
    label: "Gather",
    detail: "Shape a circle accessibly",
  },
] as const;

export default function DesignLabIndex() {
  return (
    <SafeAreaView edges={["bottom"]} style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.header}>
          <Text style={styles.eyebrow}>MOBILE DESIGN LAB</Text>
          <Text accessibilityRole="header" style={styles.title}>
            Gestures with a way out.
          </Text>
          <Text style={styles.copy}>
            Four one-handed prototypes. Every gesture has a labelled, visible
            alternative.
          </Text>
        </View>
        <View style={styles.list}>
          {flows.map((flow, index) => (
            <Link
              accessibilityHint={`Opens the ${flow.label} interaction prototype`}
              accessibilityLabel={`${flow.label}. ${flow.detail}`}
              accessibilityRole="button"
              asChild
              href={flow.href}
              key={flow.href}
            >
              <Pressable
                style={({ pressed }) => [
                  styles.card,
                  pressed && styles.cardPressed,
                ]}
              >
                <Text aria-hidden style={styles.number}>
                  {String(index + 1).padStart(2, "0")}
                </Text>
                <View style={styles.cardCopy}>
                  <Text style={styles.cardTitle}>{flow.label}</Text>
                  <Text style={styles.cardDetail}>{flow.detail}</Text>
                </View>
                <Text aria-hidden style={styles.arrow}>
                  →
                </Text>
              </Pressable>
            </Link>
          ))}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  arrow: { color: "#3A0E2E", fontSize: 24 },
  card: {
    alignItems: "center",
    backgroundColor: "#FFFFFF",
    borderColor: "rgba(58,14,46,0.12)",
    borderRadius: 20,
    borderWidth: 1,
    flexDirection: "row",
    gap: 14,
    minHeight: 88,
    padding: 18,
  },
  cardCopy: { flex: 1, gap: 4 },
  cardPressed: { opacity: 0.76 },
  cardDetail: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 15,
  },
  cardTitle: { color: "#3A0E2E", fontFamily: "Outfit_700Bold", fontSize: 21 },
  content: { gap: 28, padding: 20, paddingBottom: 48 },
  copy: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 26,
  },
  eyebrow: {
    color: "#C54267",
    fontFamily: "Outfit_700Bold",
    fontSize: 12,
    letterSpacing: 1.2,
  },
  header: { gap: 12 },
  list: { gap: 12 },
  number: { color: "#C54267", fontFamily: "Outfit_700Bold", fontSize: 12 },
  safeArea: { backgroundColor: "#FFF8F1", flex: 1 },
  title: {
    color: "#3A0E2E",
    fontFamily: "Outfit_700Bold",
    fontSize: 38,
    lineHeight: 43,
  },
});
