import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export function FireRoomScreen() {
  const router = useRouter();
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.top}>
          <Pressable
            onPress={() => router.push("/fie/abonten" as Href)}
            style={styles.control}
          >
            <Text style={styles.controlText}>Abɔnten</Text>
          </Pressable>
          <Pressable style={styles.control}>
            <Text style={styles.controlText}>Safety</Text>
          </Pressable>
        </View>
        <Text style={styles.eyebrow}>LIVE · COMMUNITY FIRE</Text>
        <Text accessibilityRole="header" style={styles.title}>
          Stories we inherited.
        </Text>
        <Text style={styles.copy}>
          Nana Esi is speaking. No phone numbers, follower counts or public
          attendance trail.
        </Text>
        <View style={styles.host}>
          <View style={styles.portrait}>
            <Text style={styles.initials}>NE</Text>
          </View>
          <Text style={styles.hostName}>Nana Esi · host</Text>
        </View>
        <View style={styles.signal}>
          <Text style={styles.signalLabel}>CONNECTION MODE</Text>
          <Text style={styles.signalTitle}>Audio only</Text>
          <Text style={styles.signalCopy}>
            Using less data. You are still in the room.
          </Text>
          <Pressable style={styles.mode}>
            <Text style={styles.modeText}>Use captions only</Text>
          </Pressable>
        </View>
        <View style={styles.caption}>
          <Text style={styles.captionLabel}>LIVE CAPTIONS</Text>
          <Text style={styles.captionText}>
            “The name was a map back to the people waiting for you.”
          </Text>
        </View>
        <Pressable style={styles.leave}>
          <Text style={styles.leaveText}>Leave fire</Text>
        </Pressable>
      </ScrollView>
    </SafeAreaView>
  );
}
const styles = StyleSheet.create({
  safe: { backgroundColor: "#0C1017", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  top: { flexDirection: "row", justifyContent: "space-between" },
  control: {
    alignItems: "center",
    borderColor: "#596170",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 18,
  },
  controlText: { color: "#F7EFE2", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#FF9D87",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.3,
    marginTop: 50,
  },
  title: {
    color: "#F7EFE2",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 57,
    letterSpacing: -3.7,
    lineHeight: 51,
    marginTop: 14,
  },
  copy: {
    color: "#AEB3BD",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 20,
  },
  host: {
    backgroundColor: "#3C1730",
    borderRadius: 24,
    marginTop: 30,
    padding: 16,
  },
  portrait: {
    alignItems: "center",
    backgroundColor: "#E9866F",
    borderRadius: 18,
    justifyContent: "center",
    minHeight: 240,
  },
  initials: { color: "#301522", fontFamily: "Outfit_900Black", fontSize: 64 },
  hostName: { color: "#F7EFE2", fontFamily: "Outfit_700Bold", marginTop: 14 },
  signal: {
    backgroundColor: "#151B24",
    borderColor: "#303947",
    borderRadius: 22,
    borderWidth: 1,
    marginTop: 10,
    padding: 22,
  },
  signalLabel: {
    color: "#9FA9B8",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  signalTitle: {
    color: "#F7EFE2",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 34,
    letterSpacing: -1.5,
    marginTop: 24,
  },
  signalCopy: {
    color: "#BAC1CC",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 8,
  },
  mode: {
    alignItems: "center",
    borderColor: "#687182",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 22,
    minHeight: 52,
  },
  modeText: { color: "#F7EFE2", fontFamily: "Outfit_700Bold" },
  caption: {
    backgroundColor: "#F4C06A",
    borderRadius: 20,
    marginTop: 10,
    padding: 20,
  },
  captionLabel: {
    color: "#261622",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  captionText: {
    color: "#261622",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 17,
    lineHeight: 26,
    marginTop: 12,
  },
  leave: {
    alignItems: "center",
    borderColor: "#FF927F",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 24,
    minHeight: 52,
  },
  leaveText: { color: "#FFB0A2", fontFamily: "Outfit_700Bold" },
});
