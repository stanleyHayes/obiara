import { useMemo, useState } from "react";
import {
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { StatusBar } from "expo-status-bar";

import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-onlight_transparent.png";
import { buildConnectionCopy, type ConnectionMode } from "../src/connection";

const palette = {
  plum: "#3A0E2E",
  ink: "#26101F",
  cream: "#FFF3E6",
  paper: "#FFFDFC",
  gold: "#FF9F1C",
  pink: "#FF4D6D",
  green: "#12876B",
  muted: "#765F70",
  line: "rgba(58, 14, 46, 0.11)",
  white: "#FFFFFF",
};

const zones = [
  {
    name: "Abɔnten",
    gloss: "the street",
    count: "3 live",
    symbol: "✦",
    tone: palette.gold,
  },
  {
    name: "Adiwo",
    gloss: "the courtyard",
    count: "4 circles",
    symbol: "◌",
    tone: palette.green,
  },
  {
    name: "Ɛpono ano",
    gloss: "the doorway",
    count: "2 waiting",
    symbol: "⌂",
    tone: palette.pink,
  },
  {
    name: "Dan mu",
    gloss: "the inner room",
    count: "1 turn",
    symbol: "●",
    tone: palette.plum,
  },
] as const;

function ActionButton({
  label,
  onPress,
  variant = "plum",
}: Readonly<{
  label: string;
  onPress?: () => void;
  variant?: "plum" | "cream";
}>) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [
        styles.actionButton,
        variant === "cream"
          ? styles.actionButtonCream
          : styles.actionButtonPlum,
        pressed && styles.pressed,
      ]}
    >
      <Text
        style={[
          styles.actionLabel,
          variant === "cream"
            ? styles.actionLabelPlum
            : styles.actionLabelCream,
        ]}
      >
        {label}
      </Text>
      <Text
        aria-hidden
        style={[
          styles.actionArrow,
          variant === "cream"
            ? styles.actionLabelPlum
            : styles.actionLabelCream,
        ]}
      >
        ↗
      </Text>
    </Pressable>
  );
}

function ZoneCard({ zone }: Readonly<{ zone: (typeof zones)[number] }>) {
  return (
    <Pressable
      accessibilityHint={`Enter ${zone.gloss}`}
      accessibilityLabel={`${zone.name}, ${zone.count}`}
      accessibilityRole="button"
      style={({ pressed }) => [styles.zoneCard, pressed && styles.pressed]}
    >
      <View style={[styles.zoneSymbol, { backgroundColor: zone.tone }]}>
        <Text aria-hidden style={styles.zoneSymbolText}>
          {zone.symbol}
        </Text>
      </View>
      <View style={styles.zoneCopy}>
        <Text style={styles.zoneName}>{zone.name}</Text>
        <Text style={styles.zoneGloss}>{zone.gloss}</Text>
      </View>
      <View style={styles.zoneCount}>
        <Text style={styles.zoneCountText}>{zone.count}</Text>
      </View>
    </Pressable>
  );
}

export default function Home() {
  const [connectionMode, setConnectionMode] =
    useState<ConnectionMode>("constrained");
  const [queued, setQueued] = useState(false);
  const connection = useMemo(
    () => buildConnectionCopy(connectionMode, queued),
    [connectionMode, queued],
  );

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.topbar}>
          <View style={styles.brand}>
            <Image
              accessibilityIgnoresInvertColors
              resizeMode="contain"
              source={brandMark}
              style={styles.brandMark}
            />
            <Text style={styles.brandName}>obiara</Text>
          </View>
          <View style={styles.topbarActions}>
            <Pressable
              accessibilityLabel="Switch language"
              accessibilityRole="button"
              style={({ pressed }) => [
                styles.languageButton,
                pressed && styles.pressed,
              ]}
            >
              <Text style={styles.languageText}>EN · TWI</Text>
            </Pressable>
            <Pressable
              accessibilityLabel="Open Ama's profile"
              accessibilityRole="button"
              style={({ pressed }) => [
                styles.avatar,
                pressed && styles.pressed,
              ]}
            >
              <Text style={styles.avatarText}>A</Text>
            </Pressable>
          </View>
        </View>

        <Pressable
          accessibilityHint="Toggles the network simulation for this feasibility prototype"
          accessibilityLabel={connection.accessibilityLabel}
          accessibilityRole="switch"
          accessibilityState={{ checked: connectionMode === "constrained" }}
          onPress={() =>
            setConnectionMode((current) =>
              current === "constrained" ? "online" : "constrained",
            )
          }
          style={({ pressed }) => [
            styles.networkBanner,
            pressed && styles.pressed,
          ]}
        >
          <View style={styles.networkPulse} />
          <View style={styles.networkCopy}>
            <Text style={styles.networkTitle}>{connection.title}</Text>
            <Text style={styles.networkBody}>{connection.body}</Text>
          </View>
          <Text aria-hidden style={styles.chevron}>
            ›
          </Text>
        </Pressable>

        <View style={styles.hero}>
          <View style={styles.dayPill}>
            <View style={styles.dayDot} />
            <Text style={styles.dayText}>SUNDAY · YOUR QUIET MORNING</Text>
          </View>
          <Text style={styles.heroTitle}>
            Akwaaba, Ama.{"\n"}
            <Text style={styles.heroOutline}>Your fie is awake.</Text>
          </Text>
          <Text style={styles.heroBody}>
            Two voices are waiting at your doorway. The drum is with you in one
            room, and tonight’s Legon fire still has a seat.
          </Text>
        </View>

        <View style={styles.ritualCard}>
          <View style={styles.ritualTop}>
            <View style={styles.ritualHeading}>
              <Text style={styles.eyebrow}>YOUR DAWN RITUAL</Text>
              <Text style={styles.ritualTitle}>
                A small look at what is growing.
              </Text>
            </View>
            <View style={styles.sun}>
              <Text aria-hidden style={styles.sunText}>
                ☼
              </Text>
            </View>
          </View>
          <View style={styles.metrics}>
            <View style={styles.metric}>
              <Text style={styles.metricValue}>2</Text>
              <Text style={styles.metricLabel}>pods waiting</Text>
            </View>
            <View style={[styles.metric, styles.metricBorder]}>
              <Text style={styles.metricValue}>1</Text>
              <Text style={styles.metricLabel}>drum turn</Text>
            </View>
            <View style={[styles.metric, styles.metricBorder]}>
              <Text style={styles.metricValue}>7</Text>
              <Text style={styles.metricLabel}>seeds this week</Text>
            </View>
          </View>
          <ActionButton label="Visit my garden" />
        </View>

        <View style={styles.sectionHeading}>
          <View>
            <Text style={styles.sectionTitle}>Walk through your compound</Text>
            <Text style={styles.sectionBody}>
              Every place has a purpose. Enter with intention.
            </Text>
          </View>
          <View style={styles.mapBadge}>
            <Text aria-hidden style={styles.mapIcon}>
              ⌘
            </Text>
          </View>
        </View>

        <View style={styles.zoneList}>
          {zones.map((zone) => (
            <ZoneCard key={zone.name} zone={zone} />
          ))}
        </View>

        <View style={styles.fireCard}>
          <View style={styles.fireGlow} />
          <View style={styles.fireMeta}>
            <View style={styles.livePill}>
              <View style={styles.liveDot} />
              <Text style={styles.liveText}>TONIGHT · 8:00 PM</Text>
            </View>
            <Text style={styles.seatText}>18 seats left</Text>
          </View>
          <Text style={styles.fireEyebrow}>GYAASE FIRE · LEGON COURTYARD</Text>
          <Text style={styles.fireTitle}>
            Come and sit.{"\n"}The fire is catching.
          </Text>
          <Text style={styles.fireBody}>
            Voice games, an Oware table and one ember to carry home. Audio
            automatically becomes listen-only when the network is too weak.
          </Text>
          <ActionButton label="Keep my seat" variant="cream" />
        </View>

        <View style={styles.queueCard}>
          <View style={styles.queueIcon}>
            <Text aria-hidden style={styles.queueIconText}>
              {queued ? "✓" : "↥"}
            </Text>
          </View>
          <View style={styles.queueCopy}>
            <Text style={styles.queueTitle}>{connection.queueTitle}</Text>
            <Text style={styles.queueBody}>{connection.queueBody}</Text>
          </View>
          <Pressable
            accessibilityRole="button"
            onPress={() => setQueued((current) => !current)}
            style={({ pressed }) => [
              styles.queueButton,
              pressed && styles.pressed,
            ]}
          >
            <Text style={styles.queueButtonText}>
              {queued ? "Undo" : "Try it"}
            </Text>
          </Pressable>
        </View>

        <View style={styles.bottomNav} accessibilityRole="tablist">
          {[
            ["⌂", "Fie", true],
            ["◌", "Circles", false],
            ["●", "Rooms", false],
            ["♙", "Me", false],
          ].map(([symbol, label, active]) => (
            <Pressable
              accessibilityRole="tab"
              accessibilityState={{ selected: Boolean(active) }}
              key={String(label)}
              style={styles.navItem}
            >
              <Text
                aria-hidden
                style={[
                  styles.navSymbol,
                  active ? styles.navSymbolActive : undefined,
                ]}
              >
                {symbol}
              </Text>
              <Text
                style={[
                  styles.navLabel,
                  active ? styles.navLabelActive : undefined,
                ]}
              >
                {label}
              </Text>
            </Pressable>
          ))}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const text = {
  regular: "Outfit_400Regular",
  medium: "Outfit_500Medium",
  semibold: "Outfit_600SemiBold",
  bold: "Outfit_700Bold",
  extraBold: "Outfit_800ExtraBold",
};

const styles = StyleSheet.create({
  safeArea: { backgroundColor: palette.cream, flex: 1 },
  scrollContent: { paddingBottom: 22 },
  topbar: {
    alignItems: "center",
    borderBottomColor: palette.line,
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    marginHorizontal: 20,
    minHeight: 70,
  },
  brand: { alignItems: "center", flexDirection: "row", gap: 7 },
  brandMark: { height: 34, width: 34 },
  brandName: {
    color: palette.plum,
    fontFamily: text.extraBold,
    fontSize: 23,
    letterSpacing: -1.1,
  },
  topbarActions: { alignItems: "center", flexDirection: "row", gap: 8 },
  languageButton: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 10,
  },
  languageText: {
    color: palette.plum,
    fontFamily: text.bold,
    fontSize: 11,
    letterSpacing: 0.5,
  },
  avatar: {
    alignItems: "center",
    backgroundColor: palette.plum,
    borderColor: palette.gold,
    borderRadius: 23,
    borderWidth: 3,
    height: 46,
    justifyContent: "center",
    width: 46,
  },
  avatarText: { color: palette.cream, fontFamily: text.bold, fontSize: 16 },
  networkBanner: {
    alignItems: "center",
    backgroundColor: "rgba(255, 255, 255, 0.58)",
    borderColor: palette.line,
    borderRadius: 16,
    borderWidth: 1,
    flexDirection: "row",
    gap: 10,
    marginHorizontal: 20,
    marginTop: 16,
    minHeight: 58,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  networkPulse: {
    backgroundColor: palette.green,
    borderRadius: 5,
    height: 10,
    width: 10,
  },
  networkCopy: { flex: 1 },
  networkTitle: { color: palette.plum, fontFamily: text.bold, fontSize: 13 },
  networkBody: {
    color: palette.muted,
    fontFamily: text.regular,
    fontSize: 11,
    marginTop: 1,
  },
  chevron: { color: palette.plum, fontFamily: text.medium, fontSize: 25 },
  hero: { paddingHorizontal: 20, paddingBottom: 30, paddingTop: 38 },
  dayPill: {
    alignItems: "center",
    alignSelf: "flex-start",
    backgroundColor: "rgba(255,255,255,0.65)",
    borderColor: palette.line,
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: "row",
    gap: 7,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  dayDot: {
    backgroundColor: palette.gold,
    borderRadius: 4,
    height: 7,
    width: 7,
  },
  dayText: {
    color: palette.plum,
    fontFamily: text.bold,
    fontSize: 9,
    letterSpacing: 0.9,
  },
  heroTitle: {
    color: palette.plum,
    fontFamily: text.extraBold,
    fontSize: 43,
    letterSpacing: -2.1,
    lineHeight: 45,
    marginTop: 19,
  },
  heroOutline: { color: palette.pink },
  heroBody: {
    color: palette.muted,
    fontFamily: text.regular,
    fontSize: 16,
    lineHeight: 24,
    marginTop: 17,
    maxWidth: 420,
  },
  ritualCard: {
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 28,
    borderWidth: 1,
    boxShadow: "0 12px 22px rgba(58, 14, 46, 0.08)",
    marginHorizontal: 20,
    padding: 21,
  },
  ritualTop: { flexDirection: "row", gap: 14, justifyContent: "space-between" },
  ritualHeading: { flex: 1 },
  eyebrow: {
    color: palette.pink,
    fontFamily: text.bold,
    fontSize: 10,
    letterSpacing: 1.2,
  },
  ritualTitle: {
    color: palette.plum,
    fontFamily: text.bold,
    fontSize: 24,
    letterSpacing: -0.7,
    lineHeight: 28,
    marginTop: 7,
  },
  sun: {
    alignItems: "center",
    backgroundColor: palette.gold,
    borderRadius: 24,
    height: 48,
    justifyContent: "center",
    width: 48,
  },
  sunText: { color: palette.plum, fontSize: 24 },
  metrics: {
    borderBottomColor: palette.line,
    borderBottomWidth: 1,
    borderTopColor: palette.line,
    borderTopWidth: 1,
    flexDirection: "row",
    marginVertical: 21,
    paddingVertical: 17,
  },
  metric: { flex: 1, paddingHorizontal: 8 },
  metricBorder: { borderLeftColor: palette.line, borderLeftWidth: 1 },
  metricValue: {
    color: palette.plum,
    fontFamily: text.extraBold,
    fontSize: 25,
    lineHeight: 28,
  },
  metricLabel: { color: palette.muted, fontFamily: text.regular, fontSize: 11 },
  actionButton: {
    alignItems: "center",
    borderRadius: 15,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 52,
    paddingHorizontal: 18,
  },
  actionButtonPlum: { backgroundColor: palette.plum },
  actionButtonCream: { backgroundColor: palette.cream },
  actionLabel: { fontFamily: text.bold, fontSize: 14 },
  actionLabelCream: { color: palette.cream },
  actionLabelPlum: { color: palette.plum },
  actionArrow: { fontFamily: text.medium, fontSize: 17 },
  sectionHeading: {
    alignItems: "flex-end",
    flexDirection: "row",
    justifyContent: "space-between",
    paddingBottom: 16,
    paddingHorizontal: 20,
    paddingTop: 40,
  },
  sectionTitle: {
    color: palette.plum,
    fontFamily: text.bold,
    fontSize: 24,
    letterSpacing: -0.7,
  },
  sectionBody: {
    color: palette.muted,
    fontFamily: text.regular,
    fontSize: 13,
    marginTop: 4,
  },
  mapBadge: {
    alignItems: "center",
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 18,
    borderWidth: 1,
    height: 40,
    justifyContent: "center",
    width: 40,
  },
  mapIcon: { color: palette.plum, fontFamily: text.bold, fontSize: 16 },
  zoneList: { gap: 10, paddingHorizontal: 20 },
  zoneCard: {
    alignItems: "center",
    backgroundColor: "rgba(255,255,255,0.62)",
    borderColor: palette.line,
    borderRadius: 20,
    borderWidth: 1,
    flexDirection: "row",
    minHeight: 76,
    padding: 12,
  },
  zoneSymbol: {
    alignItems: "center",
    borderRadius: 16,
    height: 50,
    justifyContent: "center",
    width: 50,
  },
  zoneSymbolText: { color: palette.white, fontFamily: text.bold, fontSize: 20 },
  zoneCopy: { flex: 1, paddingHorizontal: 13 },
  zoneName: { color: palette.plum, fontFamily: text.bold, fontSize: 17 },
  zoneGloss: { color: palette.muted, fontFamily: text.regular, fontSize: 12 },
  zoneCount: {
    backgroundColor: palette.cream,
    borderRadius: 999,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  zoneCountText: {
    color: palette.plum,
    fontFamily: text.semibold,
    fontSize: 11,
  },
  fireCard: {
    backgroundColor: palette.plum,
    borderRadius: 30,
    marginHorizontal: 20,
    marginTop: 36,
    overflow: "hidden",
    padding: 24,
    position: "relative",
  },
  fireGlow: {
    backgroundColor: "rgba(255,159,28,0.22)",
    borderRadius: 100,
    height: 170,
    position: "absolute",
    right: -50,
    top: -60,
    width: 170,
  },
  fireMeta: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  livePill: {
    alignItems: "center",
    backgroundColor: palette.pink,
    borderRadius: 999,
    flexDirection: "row",
    gap: 7,
    paddingHorizontal: 11,
    paddingVertical: 8,
  },
  liveDot: {
    backgroundColor: palette.white,
    borderRadius: 3,
    height: 6,
    width: 6,
  },
  liveText: {
    color: palette.white,
    fontFamily: text.bold,
    fontSize: 9,
    letterSpacing: 0.6,
  },
  seatText: { color: palette.cream, fontFamily: text.medium, fontSize: 11 },
  fireEyebrow: {
    color: palette.gold,
    fontFamily: text.bold,
    fontSize: 10,
    letterSpacing: 1,
    marginTop: 32,
  },
  fireTitle: {
    color: palette.cream,
    fontFamily: text.extraBold,
    fontSize: 32,
    letterSpacing: -1.2,
    lineHeight: 35,
    marginTop: 9,
  },
  fireBody: {
    color: "rgba(255,243,230,0.76)",
    fontFamily: text.regular,
    fontSize: 14,
    lineHeight: 21,
    marginBottom: 21,
    marginTop: 13,
  },
  queueCard: {
    alignItems: "center",
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 20,
    borderWidth: 1,
    flexDirection: "row",
    gap: 12,
    marginHorizontal: 20,
    marginTop: 12,
    padding: 14,
  },
  queueIcon: {
    alignItems: "center",
    backgroundColor: "rgba(18,135,107,0.12)",
    borderRadius: 15,
    height: 46,
    justifyContent: "center",
    width: 46,
  },
  queueIconText: { color: palette.green, fontFamily: text.bold, fontSize: 19 },
  queueCopy: { flex: 1 },
  queueTitle: { color: palette.plum, fontFamily: text.bold, fontSize: 14 },
  queueBody: {
    color: palette.muted,
    fontFamily: text.regular,
    fontSize: 11,
    lineHeight: 16,
    marginTop: 2,
  },
  queueButton: {
    alignItems: "center",
    borderColor: palette.line,
    borderRadius: 13,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    minWidth: 62,
    paddingHorizontal: 10,
  },
  queueButtonText: { color: palette.plum, fontFamily: text.bold, fontSize: 12 },
  bottomNav: {
    backgroundColor: palette.paper,
    borderColor: palette.line,
    borderRadius: 24,
    borderWidth: 1,
    flexDirection: "row",
    marginHorizontal: 20,
    marginTop: 22,
    paddingHorizontal: 8,
    paddingVertical: 8,
  },
  navItem: {
    alignItems: "center",
    flex: 1,
    justifyContent: "center",
    minHeight: 54,
  },
  navSymbol: { color: palette.muted, fontFamily: text.medium, fontSize: 18 },
  navSymbolActive: { color: palette.pink },
  navLabel: {
    color: palette.muted,
    fontFamily: text.medium,
    fontSize: 10,
    marginTop: 2,
  },
  navLabelActive: { color: palette.plum, fontFamily: text.bold },
  pressed: { opacity: 0.68, transform: [{ scale: 0.985 }] },
});
