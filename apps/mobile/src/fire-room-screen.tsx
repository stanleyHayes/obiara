import { apiRequest } from "./api";
import { type Href, useLocalSearchParams, useRouter } from "expo-router";
import { useEffect, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

type Fire = {
  fireId: string;
  title: string;
  startsAt: string;
  capacity: number;
  goingCount: number;
  status: string;
};

export function FireRoomScreen() {
  const router = useRouter();
  const { fireId } = useLocalSearchParams<{ fireId: string }>();
  const [fire, setFire] = useState<Fire | null>(null);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [notice, setNotice] = useState("");

  useEffect(() => {
    void apiRequest<{ fires: Fire[] }>("/v1/fires")
      .then((result) => {
        const selected =
          result.fires.find((item) => item.fireId === fireId) ?? null;
        setFire(selected);
        if (!selected) {
          setNotice(
            "This fire is not in the retained upcoming schedule. It may have closed or the reference may be invalid.",
          );
        }
      })
      .catch((error: unknown) =>
        setNotice(
          error instanceof Error ? error.message : "The fire could not load.",
        ),
      )
      .finally(() => setLoading(false));
  }, [fireId]);

  async function reserve() {
    if (!fire) return;
    setJoining(true);
    setNotice("");
    try {
      const rsvp = await apiRequest<{ status: string; position?: number }>(
        `/v1/fires/${encodeURIComponent(fire.fireId)}/rsvps`,
        { method: "POST", body: "{}" },
      );
      setNotice(
        rsvp.status === "waitlisted"
          ? `You are waitlisted${rsvp.position ? ` at position ${rsvp.position}` : ""}.`
          : "Your place is held.",
      );
      setFire((current) =>
        current && rsvp.status !== "waitlisted"
          ? {
              ...current,
              goingCount: Math.min(current.capacity, current.goingCount + 1),
            }
          : current,
      );
    } catch (error) {
      setNotice(
        error instanceof Error
          ? error.message
          : "Your place could not be held.",
      );
    } finally {
      setJoining(false);
    }
  }

  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <Pressable
          onPress={() => router.push("/fie/abonten" as Href)}
          style={styles.control}
        >
          <Text style={styles.controlText}>Abɔnten</Text>
        </Pressable>
        <Text style={styles.eyebrow}>COMMUNITY FIRE</Text>
        {loading ? (
          <ActivityIndicator color="#FF9D87" style={styles.loader} />
        ) : fire ? (
          <>
            <Text accessibilityRole="header" style={styles.title}>
              {fire.title}
            </Text>
            <Text style={styles.copy}>
              A bounded community gathering. Attendance is private and contact
              details are never exchanged.
            </Text>
            <View style={styles.panel}>
              <Text style={styles.panelLabel}>BEGINS</Text>
              <Text style={styles.panelValue}>
                {new Date(fire.startsAt).toLocaleString()}
              </Text>
              <View style={styles.rule} />
              <Text style={styles.panelLabel}>PLACES</Text>
              <Text style={styles.panelValue}>
                {fire.goingCount} of {fire.capacity} held
              </Text>
              <Text style={styles.status}>{fire.status}</Text>
            </View>
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ disabled: joining }}
              disabled={joining}
              onPress={() => void reserve()}
              style={[styles.reserve, joining && styles.disabled]}
            >
              {joining ? (
                <ActivityIndicator color="#301522" />
              ) : (
                <Text style={styles.reserveText}>Hold my place</Text>
              )}
            </Pressable>
          </>
        ) : null}
        {notice ? (
          <Text accessibilityLiveRegion="polite" style={styles.notice}>
            {notice}
          </Text>
        ) : null}
        <View style={styles.boundary}>
          <Text style={styles.boundaryTitle}>Quiet attendance</Text>
          <Text style={styles.boundaryCopy}>
            No public attendance trail, follower count or contact exchange. Full
            verification is required before a place can be held.
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: "#0C1017", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  control: {
    alignItems: "center",
    alignSelf: "flex-start",
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
  loader: { marginVertical: 80 },
  title: {
    color: "#F7EFE2",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 55,
    letterSpacing: -3.5,
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
  panel: {
    backgroundColor: "#151B24",
    borderColor: "#303947",
    borderRadius: 22,
    borderWidth: 1,
    marginTop: 30,
    padding: 22,
  },
  panelLabel: {
    color: "#9FA9B8",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  panelValue: {
    color: "#F7EFE2",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 27,
    marginTop: 8,
  },
  rule: { backgroundColor: "#303947", height: 1, marginVertical: 20 },
  status: {
    color: "#FF9D87",
    fontFamily: "Outfit_700Bold",
    marginTop: 18,
    textTransform: "uppercase",
  },
  reserve: {
    alignItems: "center",
    backgroundColor: "#F4C06A",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 14,
    minHeight: 54,
  },
  reserveText: { color: "#301522", fontFamily: "Outfit_800ExtraBold" },
  disabled: { opacity: 0.5 },
  notice: {
    color: "#FFB0A2",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 22,
    marginTop: 16,
  },
  boundary: {
    backgroundColor: "#3C1730",
    borderRadius: 22,
    marginTop: 22,
    padding: 22,
  },
  boundaryTitle: {
    color: "#F7EFE2",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 26,
  },
  boundaryCopy: {
    color: "#D8C4CE",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginTop: 9,
  },
});
