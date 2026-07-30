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

type CircleType =
  "community" | "campus" | "professional" | "interest" | "support";
type Circle = {
  id: string;
  type: CircleType;
  visibility: "private" | "discoverable";
  membership: "none" | "requested" | "member" | "host" | "owner";
  memberCount: number;
  revision: number;
  members?: { id: string; state: "requested" | "member" | "host" | "owner" }[];
};
const types: CircleType[] = [
  "community",
  "campus",
  "professional",
  "interest",
  "support",
];

export function AdiwoScreen() {
  const router = useRouter();
  const [view, setView] = useState<"mine" | "discover">("mine");
  const [circles, setCircles] = useState<Circle[]>([]);
  const [type, setType] = useState<CircleType>("community");
  const [busy, setBusy] = useState("load");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setBusy("load");
    setError("");
    try {
      const result = await apiRequest<{ items: Circle[] }>(
        `/v1/circles?view=${view}&limit=50`,
      );
      setCircles(result.items);
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The courtyard could not be opened.",
      );
    } finally {
      setBusy("");
    }
  }, [view]);

  useEffect(() => {
    void load();
  }, [load]);

  async function create() {
    const stamp = Date.now();
    setBusy("create");
    setError("");
    try {
      await apiRequest<Circle>("/v1/circles", {
        method: "POST",
        headers: { "Idempotency-Key": `circle-create-${stamp}` },
        body: JSON.stringify({ id: `circle_${stamp}`, type }),
      });
      await load();
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The circle could not be created.",
      );
      setBusy("");
    }
  }

  async function act(
    circle: Circle,
    action: "request" | "leave" | "visibility",
  ) {
    setBusy(circle.id);
    setError("");
    const path =
      action === "request"
        ? `/v1/circles/${circle.id}/requests`
        : action === "leave"
          ? `/v1/circles/${circle.id}/leave`
          : `/v1/circles/${circle.id}/visibility`;
    try {
      await apiRequest<Circle>(path, {
        method: action === "visibility" ? "PUT" : "POST",
        headers: { "Idempotency-Key": `circle-${action}-${Date.now()}` },
        body: JSON.stringify({
          expectedRevision: circle.revision,
          ...(action === "visibility"
            ? {
                visibility:
                  circle.visibility === "private" ? "discoverable" : "private",
              }
            : {}),
        }),
      });
      await load();
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The circle action could not be completed.",
      );
      setBusy("");
    }
  }

  async function manage(
    circle: Circle,
    memberId: string,
    action: "approve" | "promote" | "expel",
  ) {
    setBusy(memberId);
    setError("");
    try {
      await apiRequest<Circle>(
        `/v1/circles/${circle.id}/members/${memberId}/${action}`,
        {
          method: "POST",
          headers: { "Idempotency-Key": `circle-${action}-${Date.now()}` },
          body: JSON.stringify({ expectedRevision: circle.revision }),
        },
      );
      await load();
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : "The membership action could not be completed.",
      );
      setBusy("");
    }
  }

  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <Pressable onPress={() => router.push("/fie" as Href)} style={s.back}>
          <Text style={s.backText}>Fie</Text>
        </Pressable>
        <Text style={s.eyebrow}>ADIWO · THE COURTYARD</Text>
        <Text accessibilityRole="header" style={s.title}>
          Belonging is deliberate.
        </Text>
        <Text style={s.lede}>
          Private circles reveal nothing before entry. Discovery shows only
          bounded circle facts.
        </Text>
        <View style={s.tabs}>
          {(["mine", "discover"] as const).map((item) => (
            <Pressable
              accessibilityState={{ selected: view === item }}
              key={item}
              onPress={() => setView(item)}
              style={[s.tab, view === item && s.tabOn]}
            >
              <Text style={[s.tabText, view === item && s.tabTextOn]}>
                {item === "mine" ? "My circles" : "Discover"}
              </Text>
            </Pressable>
          ))}
        </View>
        {view === "mine" ? (
          <View style={s.create}>
            <Text style={s.cardTitle}>Open a private courtyard</Text>
            <ScrollView
              horizontal
              showsHorizontalScrollIndicator={false}
              style={s.typeRail}
            >
              {types.map((item) => (
                <Pressable
                  key={item}
                  onPress={() => setType(item)}
                  style={[s.type, type === item && s.typeOn]}
                >
                  <Text style={[s.typeText, type === item && s.typeTextOn]}>
                    {item}
                  </Text>
                </Pressable>
              ))}
            </ScrollView>
            <Pressable
              disabled={Boolean(busy)}
              onPress={() => void create()}
              style={[s.primary, Boolean(busy) && s.disabled]}
            >
              <Text style={s.primaryText}>
                {busy === "create" ? "Creating…" : "Create private circle"}
              </Text>
            </Pressable>
          </View>
        ) : null}
        {busy === "load" ? (
          <ActivityIndicator color="#8E3159" style={s.loader} />
        ) : null}
        {error ? (
          <Text accessibilityLiveRegion="polite" style={s.error}>
            {error}
          </Text>
        ) : null}
        {!busy && !error && circles.length === 0 ? (
          <Text style={s.empty}>No circles in this view yet.</Text>
        ) : null}
        {circles.map((circle) => (
          <View key={circle.id} style={s.card}>
            <Text style={s.control}>
              {circle.type} · {circle.visibility}
            </Text>
            <Text selectable style={s.cardTitle}>
              {circle.id.slice(0, 24)}
            </Text>
            <Text style={s.copy}>
              {circle.memberCount} active{" "}
              {circle.memberCount === 1 ? "member" : "members"} · your state:{" "}
              {circle.membership}
            </Text>
            {circle.membership === "none" ? (
              <Pressable
                disabled={Boolean(busy)}
                onPress={() => void act(circle, "request")}
                style={s.primary}
              >
                <Text style={s.primaryText}>Request to join</Text>
              </Pressable>
            ) : null}
            {circle.membership === "owner" ? (
              <Pressable
                disabled={Boolean(busy)}
                onPress={() => void act(circle, "visibility")}
                style={s.primary}
              >
                <Text style={s.primaryText}>
                  {circle.visibility === "private"
                    ? "Allow discovery"
                    : "Make private"}
                </Text>
              </Pressable>
            ) : null}
            {circle.membership === "member" || circle.membership === "host" ? (
              <Pressable
                disabled={Boolean(busy)}
                onPress={() => void act(circle, "leave")}
                style={s.secondary}
              >
                <Text style={s.secondaryText}>Leave circle</Text>
              </Pressable>
            ) : null}
            {circle.membership === "owner"
              ? circle.members?.map((member) =>
                  member.state === "owner" ? null : (
                    <View key={member.id} style={s.member}>
                      <View>
                        <Text style={s.memberId}>{member.id.slice(0, 18)}</Text>
                        <Text style={s.copy}>{member.state}</Text>
                      </View>
                      <Pressable
                        disabled={Boolean(busy)}
                        onPress={() =>
                          void manage(
                            circle,
                            member.id,
                            member.state === "requested"
                              ? "approve"
                              : member.state === "member"
                                ? "promote"
                                : "expel",
                          )
                        }
                        style={s.memberAction}
                      >
                        <Text style={s.memberActionText}>
                          {member.state === "requested"
                            ? "Approve"
                            : member.state === "member"
                              ? "Make host"
                              : "Remove"}
                        </Text>
                      </Pressable>
                    </View>
                  ),
                )
              : null}
          </View>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#F7EFE2", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  back: { alignSelf: "flex-start", justifyContent: "center", minHeight: 44 },
  backText: { color: "#3A0E2E", fontFamily: "Outfit_700Bold" },
  eyebrow: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1.2,
    marginTop: 22,
  },
  title: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 46,
    letterSpacing: -2.8,
    lineHeight: 43,
    marginTop: 12,
  },
  lede: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginTop: 18,
  },
  tabs: { flexDirection: "row", gap: 8, marginVertical: 24 },
  tab: {
    borderColor: "rgba(58,14,46,.18)",
    borderRadius: 999,
    borderWidth: 1,
    paddingHorizontal: 18,
    paddingVertical: 11,
  },
  tabOn: { backgroundColor: "#3A0E2E" },
  tabText: { color: "#3A0E2E", fontFamily: "Outfit_700Bold" },
  tabTextOn: { color: "#FFF3E6" },
  create: {
    backgroundColor: "#26101F",
    borderRadius: 20,
    marginBottom: 12,
    padding: 20,
  },
  typeRail: { marginTop: 14 },
  type: {
    borderColor: "rgba(255,243,230,.22)",
    borderRadius: 999,
    borderWidth: 1,
    marginRight: 7,
    paddingHorizontal: 13,
    paddingVertical: 9,
  },
  typeOn: { backgroundColor: "#FFF3E6" },
  typeText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_600SemiBold",
    textTransform: "capitalize",
  },
  typeTextOn: { color: "#26101F" },
  card: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,.11)",
    borderRadius: 18,
    borderWidth: 1,
    marginBottom: 10,
    padding: 20,
  },
  control: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1,
    textTransform: "uppercase",
  },
  cardTitle: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 20,
    marginTop: 6,
  },
  copy: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    lineHeight: 20,
    marginTop: 7,
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#FF9F1C",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 15,
    minHeight: 48,
  },
  primaryText: { color: "#26101F", fontFamily: "Outfit_700Bold" },
  secondary: {
    alignItems: "center",
    borderColor: "#8E3159",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    marginTop: 15,
    minHeight: 48,
  },
  secondaryText: { color: "#8E3159", fontFamily: "Outfit_700Bold" },
  disabled: { opacity: 0.55 },
  loader: { marginVertical: 34 },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 20,
    marginVertical: 15,
  },
  empty: {
    color: "#765F70",
    fontFamily: "Outfit_600SemiBold",
    paddingVertical: 30,
    textAlign: "center",
  },
  member: {
    alignItems: "center",
    borderTopColor: "rgba(58,14,46,.11)",
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    marginTop: 14,
    paddingTop: 12,
  },
  memberId: { color: "#26101F", fontFamily: "Outfit_700Bold", fontSize: 12 },
  memberAction: {
    borderColor: "#8E3159",
    borderRadius: 999,
    borderWidth: 1,
    paddingHorizontal: 12,
    paddingVertical: 9,
  },
  memberActionText: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
  },
});
