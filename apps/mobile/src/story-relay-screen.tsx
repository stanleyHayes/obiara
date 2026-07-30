import { type Href, useRouter } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { apiRequest } from "./api";

type Passage = {
  id: string;
  ordinal: number;
  content: string;
  yours: boolean;
  createdAt: string;
  editedAt: string;
};

type Story = {
  id: string;
  titleCode: string;
  passages: Passage[];
  yourTurn: boolean;
  yourGrant: boolean;
  otherGrant: boolean;
  bothGranted: boolean;
  editions: Array<{ version: number; publishedAt: string }>;
  revision: number;
};

type Action = "add" | "edit" | "grant" | "publish";

export function StoryRelayScreen({
  storyId,
  circleId,
}: Readonly<{ storyId: string; circleId: string }>) {
  const router = useRouter();
  const [story, setStory] = useState<Story | null>(null);
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState<string | null>(null);
  const [busy, setBusy] = useState<Action | "load" | null>("load");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (!circleId || !storyId) {
      setError("This story needs its private circle reference.");
      setBusy(null);
      return;
    }
    try {
      const result = await apiRequest<Story>(
        `/v1/circles/${encodeURIComponent(circleId)}/stories/${encodeURIComponent(storyId)}`,
      );
      setStory(result);
      setError("");
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : "The private story could not be opened.",
      );
    } finally {
      setBusy(null);
    }
  }, [circleId, storyId]);

  useEffect(() => {
    let active = true;
    void Promise.resolve().then(() => {
      if (active) void load();
    });
    return () => {
      active = false;
    };
  }, [load]);

  async function mutate(action: Action) {
    if (!story) return;
    setBusy(action);
    setError("");
    const suffix =
      action === "add"
        ? "/passages"
        : action === "edit"
          ? `/passages/${encodeURIComponent(editing ?? "")}`
          : action === "grant"
            ? "/publication-grants"
            : "/publish";
    try {
      const result = await apiRequest<Story>(
        `/v1/circles/${encodeURIComponent(circleId)}/stories/${encodeURIComponent(storyId)}${suffix}`,
        {
          method: action === "edit" ? "PUT" : "POST",
          headers: {
            "Idempotency-Key": `story-${action}-${Date.now()}-${Math.random().toString(36).slice(2)}`,
          },
          body: JSON.stringify({
            ...(action === "add" || action === "edit"
              ? { content: draft.trim() }
              : {}),
            expectedRevision: story.revision,
          }),
        },
      );
      setStory(result);
      setDraft("");
      setEditing(null);
    } catch (actionError) {
      setError(
        actionError instanceof Error
          ? actionError.message
          : "The story action could not be retained.",
      );
      await load();
    } finally {
      setBusy(null);
    }
  }

  function beginEdit(passage: Passage) {
    setEditing(passage.id);
    setDraft(passage.content);
  }

  const roomHref = circleId
    ? (`/fie/dan-mu/rooms/${circleId}` as Href)
    : ("/fie/adiwo" as Href);
  const canAdd =
    !!story && story.yourTurn && story.passages.length < 40 && !editing;

  return (
    <SafeAreaView style={s.safe}>
      <ScrollView contentContainerStyle={s.content}>
        <View style={s.topbar}>
          <Pressable onPress={() => router.push(roomHref)} style={s.control}>
            <Text style={s.controlText}>Private room</Text>
          </Pressable>
          <Text style={s.reference}>{storyId.slice(0, 8)}</Text>
        </View>
        <Text style={s.eyebrow}>ONE PASSAGE, THEN THE OTHER</Text>
        <Text accessibilityRole="header" style={s.title}>
          Build a story without inventing its history.
        </Text>
        <Text style={s.body}>
          Every passage is retained and alternation is verified by the server.
          Publishing requires two grants bound to the current draft.
        </Text>

        {busy === "load" ? (
          <ActivityIndicator color="#9B315D" style={s.loader} />
        ) : null}
        {error ? <Text style={s.error}>{error}</Text> : null}
        {story ? (
          <>
            <View style={s.paper}>
              <Text style={s.meta}>
                DRAFT · PRIVATE TO TWO · {story.passages.length} PASSAGES ·
                REVISION {story.revision}
              </Text>
              <Text accessibilityRole="header" style={s.storyTitle}>
                {story.titleCode.replaceAll("-", " ")}
              </Text>
              {story.passages.length === 0 ? (
                <View style={s.passage}>
                  <Text style={s.who}>THE FIRST PAGE IS OPEN</Text>
                  <Text style={s.passageText}>
                    No sample prose is inserted. The first retained passage
                    begins here.
                  </Text>
                </View>
              ) : null}
              {story.passages.map((passage) => (
                <View key={passage.id} style={s.passage}>
                  <Text style={s.who}>
                    {String(passage.ordinal + 1).padStart(2, "0")} ·{" "}
                    {passage.yours ? "YOU" : "OTHER AUTHOR"}
                  </Text>
                  <Text style={s.passageText}>{passage.content}</Text>
                  {passage.yours ? (
                    <Pressable
                      disabled={busy !== null}
                      onPress={() => beginEdit(passage)}
                      style={s.editButton}
                    >
                      <Text style={s.editText}>Revise this passage</Text>
                    </Pressable>
                  ) : null}
                </View>
              ))}
              <View style={s.composer}>
                <Text style={s.composerLabel}>
                  {editing
                    ? "AUTHOR-OWNED REVISION"
                    : story.yourTurn
                      ? "THE RELAY IS WITH YOU"
                      : "THE RELAY IS WITH THE OTHER AUTHOR"}
                </Text>
                <Text style={s.composerTitle}>
                  {editing
                    ? "Revise your passage."
                    : story.yourTurn
                      ? "Add one passage."
                      : "Your words are resting."}
                </Text>
                <TextInput
                  accessibilityLabel={
                    editing ? "Revised passage" : "Your next passage"
                  }
                  editable={busy === null && (!!editing || canAdd)}
                  maxLength={280}
                  multiline
                  onChangeText={setDraft}
                  placeholder="Let the next moment unfold…"
                  placeholderTextColor="#846E79"
                  style={s.input}
                  value={draft}
                />
                <Text style={s.count}>{draft.length}/280</Text>
                <Pressable
                  disabled={
                    busy !== null ||
                    draft.trim().length === 0 ||
                    (!editing && !canAdd)
                  }
                  onPress={() => void mutate(editing ? "edit" : "add")}
                  style={[
                    s.primary,
                    (busy !== null ||
                      draft.trim().length === 0 ||
                      (!editing && !canAdd)) &&
                      s.disabled,
                  ]}
                >
                  <Text style={s.primaryText}>
                    {busy === "add" || busy === "edit"
                      ? "Retaining…"
                      : editing
                        ? "Retain revision"
                        : "Add one passage"}
                  </Text>
                </Pressable>
                {editing ? (
                  <Pressable
                    disabled={busy !== null}
                    onPress={() => {
                      setEditing(null);
                      setDraft("");
                    }}
                    style={s.cancel}
                  >
                    <Text style={s.cancelText}>Cancel revision</Text>
                  </Pressable>
                ) : null}
              </View>
            </View>

            <View style={s.publish}>
              <Text style={s.publishEyebrow}>
                FINGERPRINT-BOUND PUBLICATION
              </Text>
              <Text accessibilityRole="header" style={s.publishTitle}>
                Private unless both grant this draft.
              </Text>
              <Text style={s.publishCopy}>
                New writing clears prior grants. Published editions contain no
                room reference or private authorship.
              </Text>
              <View style={s.consentRow}>
                <Text style={s.consentName}>Other author</Text>
                <Text style={s.consentValue}>
                  {story.otherGrant ? "Granted" : "Private"}
                </Text>
              </View>
              <View style={s.consentRow}>
                <Text style={s.consentName}>You</Text>
                <Text style={s.consentValue}>
                  {story.yourGrant ? "Granted" : "Private"}
                </Text>
              </View>
              <Pressable
                disabled={
                  busy !== null ||
                  story.yourGrant ||
                  story.passages.length === 0
                }
                onPress={() => void mutate("grant")}
                style={[
                  s.consentButton,
                  (busy !== null ||
                    story.yourGrant ||
                    story.passages.length === 0) &&
                    s.disabled,
                ]}
              >
                <Text style={s.consentButtonText}>
                  {busy === "grant"
                    ? "Retaining grant…"
                    : story.yourGrant
                      ? "Current draft granted"
                      : "Grant publication for this draft"}
                </Text>
              </Pressable>
              <Pressable
                disabled={busy !== null || !story.bothGranted}
                onPress={() => void mutate("publish")}
                style={[
                  s.consentButton,
                  (busy !== null || !story.bothGranted) && s.disabled,
                ]}
              >
                <Text style={s.consentButtonText}>
                  {busy === "publish"
                    ? "Publishing…"
                    : "Publish current edition"}
                </Text>
              </Pressable>
              <Text accessibilityLiveRegion="polite" style={s.consentStatus}>
                {story.bothGranted
                  ? "Both current-draft grants are present."
                  : "This draft remains private."}
              </Text>
              {story.editions.length ? (
                <Text style={s.consentStatus}>
                  {story.editions.length} retained public{" "}
                  {story.editions.length === 1 ? "edition" : "editions"}.
                </Text>
              ) : null}
            </View>
          </>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  safe: { backgroundColor: "#F5EAD8", flex: 1 },
  content: { padding: 20, paddingBottom: 56 },
  topbar: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  control: {
    alignItems: "center",
    borderColor: "#927982",
    borderRadius: 999,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 48,
    paddingHorizontal: 17,
  },
  controlText: { color: "#291720", fontFamily: "Outfit_700Bold" },
  reference: {
    color: "#705C67",
    fontFamily: "Outfit_700Bold",
    fontSize: 11,
    letterSpacing: 1,
  },
  eyebrow: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
    marginTop: 50,
  },
  title: {
    color: "#291720",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 50,
    letterSpacing: -3.1,
    lineHeight: 47,
    marginTop: 14,
  },
  body: {
    color: "#705C67",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 25,
    marginTop: 22,
  },
  loader: { marginTop: 30 },
  error: {
    color: "#8E1F3C",
    fontFamily: "Outfit_600SemiBold",
    lineHeight: 21,
    marginTop: 22,
  },
  paper: {
    backgroundColor: "#FFFAF1",
    borderColor: "#DDCEC2",
    borderRadius: 26,
    borderWidth: 1,
    marginTop: 36,
    padding: 22,
  },
  meta: {
    color: "#705C67",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
    lineHeight: 16,
  },
  storyTitle: {
    color: "#291720",
    fontFamily: "Outfit_700Bold",
    fontSize: 40,
    letterSpacing: -2,
    lineHeight: 42,
    marginVertical: 36,
    textTransform: "capitalize",
  },
  passage: {
    borderTopColor: "#DFD2C8",
    borderTopWidth: 1,
    paddingVertical: 22,
  },
  who: {
    color: "#9B315D",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.2,
  },
  passageText: {
    color: "#291720",
    fontFamily: "Outfit_400Regular",
    fontSize: 19,
    lineHeight: 29,
    marginTop: 10,
  },
  editButton: {
    alignSelf: "flex-start",
    justifyContent: "center",
    marginTop: 12,
    minHeight: 44,
  },
  editText: {
    color: "#7C294E",
    fontFamily: "Outfit_700Bold",
    textDecorationLine: "underline",
  },
  composer: {
    backgroundColor: "#291720",
    borderRadius: 20,
    marginTop: 28,
    padding: 20,
  },
  composerLabel: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  composerTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 30,
    letterSpacing: -1.4,
    marginTop: 8,
  },
  input: {
    backgroundColor: "#FFFAF1",
    borderRadius: 14,
    color: "#291720",
    fontFamily: "Outfit_400Regular",
    marginTop: 18,
    minHeight: 120,
    padding: 14,
    textAlignVertical: "top",
  },
  count: {
    color: "rgba(255,243,230,.6)",
    fontFamily: "Outfit_400Regular",
    fontSize: 11,
    marginTop: 6,
    textAlign: "right",
  },
  primary: {
    alignItems: "center",
    backgroundColor: "#FFAD3D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 52,
  },
  primaryText: { color: "#291720", fontFamily: "Outfit_700Bold" },
  cancel: { alignItems: "center", justifyContent: "center", minHeight: 48 },
  cancelText: {
    color: "#FFF3E6",
    fontFamily: "Outfit_700Bold",
    textDecorationLine: "underline",
  },
  disabled: { opacity: 0.4 },
  publish: {
    backgroundColor: "#291720",
    borderRadius: 26,
    marginTop: 38,
    padding: 22,
  },
  publishEyebrow: {
    color: "#FF91A6",
    fontFamily: "Outfit_700Bold",
    fontSize: 10,
    letterSpacing: 1.1,
  },
  publishTitle: {
    color: "#FFF3E6",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 38,
    letterSpacing: -2,
    lineHeight: 38,
    marginTop: 12,
  },
  publishCopy: {
    color: "rgba(255,243,230,.65)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 23,
    marginBottom: 18,
    marginTop: 14,
  },
  consentRow: {
    borderBottomColor: "rgba(255,243,230,.18)",
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 14,
  },
  consentName: { color: "#FFF3E6", fontFamily: "Outfit_600SemiBold" },
  consentValue: { color: "#FFB7C4", fontFamily: "Outfit_700Bold" },
  consentButton: {
    alignItems: "center",
    backgroundColor: "#FFAD3D",
    borderRadius: 999,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 52,
    paddingHorizontal: 12,
  },
  consentButtonText: {
    color: "#291720",
    fontFamily: "Outfit_700Bold",
    textAlign: "center",
  },
  consentStatus: {
    color: "rgba(255,243,230,.65)",
    fontFamily: "Outfit_400Regular",
    lineHeight: 22,
    marginTop: 16,
  },
});
