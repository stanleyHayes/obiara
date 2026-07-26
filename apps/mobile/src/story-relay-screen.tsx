import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const passages = [
  ["AMA", "At dusk, Ananse found a calabash humming beside the old silk-cotton tree."],
  ["YOU", "He leaned close, but the song moved into the path beneath his feet."],
  ["AMA", "Every step gave him one memory and borrowed another from the moon."],
  ["YOU", "So he stopped walking and asked the path what it wanted in return."],
] as const;

export function StoryRelayScreen() {
  const router = useRouter();
  const [draft, setDraft] = useState("");
  const [sent, setSent] = useState(false);
  const [publishConsent, setPublishConsent] = useState(false);
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable
            onPress={() => router.push("/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa" as Href)}
            style={styles.control}
          >
            <Text style={styles.controlText}>Private room</Text>
          </Pressable>
          <Pressable style={styles.control}><Text style={styles.controlText}>Safety</Text></Pressable>
        </View>
        <Text style={styles.eyebrow}>ONE LINE, THEN THE OTHER</Text>
        <Text accessibilityRole="header" style={styles.title}>Build a story without racing its ending.</Text>
        <Text style={styles.body}>
          Publishing is a separate choice both writers make after the latest contribution.
        </Text>
        <View style={styles.paper}>
          <Text style={styles.meta}>DRAFT · PRIVATE TO TWO · {sent ? "5" : "4"} PASSAGES</Text>
          <Text accessibilityRole="header" style={styles.storyTitle}>The path that remembered</Text>
          {passages.map(([who, text], index) => (
            <View key={text} style={styles.passage}>
              <Text style={styles.who}>{String(index + 1).padStart(2, "0")} · {who}</Text>
              <Text style={styles.passageText}>{text}</Text>
            </View>
          ))}
          {sent ? (
            <View style={styles.passage}>
              <Text style={styles.who}>05 · YOU</Text>
              <Text style={styles.passageText}>The path answered with a drumbeat.</Text>
            </View>
          ) : null}
          <View style={styles.composer}>
            <Text style={styles.composerLabel}>{sent ? "THE RELAY IS WITH AMA" : "THE RELAY IS WITH YOU"}</Text>
            <Text style={styles.composerTitle}>{sent ? "Your words are resting." : "Add one passage."}</Text>
            <TextInput
              accessibilityLabel="Your next passage"
              editable={!sent}
              maxLength={280}
              multiline
              onChangeText={setDraft}
              placeholder="Let the next moment unfold…"
              placeholderTextColor="#846E79"
              style={styles.input}
              value={draft}
            />
            <Text style={styles.count}>{draft.length}/280</Text>
            <Pressable
              disabled={draft.trim().length < 3 || sent}
              onPress={() => {
                setSent(true);
                setDraft("");
                setPublishConsent(false);
              }}
              style={[styles.primary, (draft.trim().length < 3 || sent) && styles.disabled]}
            >
              <Text style={styles.primaryText}>Add one passage</Text>
            </Pressable>
          </View>
        </View>
        <View style={styles.publish}>
          <Text style={styles.eyebrow}>SEPARATE PUBLISHING CONSENT</Text>
          <Text accessibilityRole="header" style={styles.publishTitle}>Private unless both say otherwise.</Text>
          <Text style={styles.publishCopy}>
            A public edition removes room references and private authorship. Either person can withdraw before publication.
          </Text>
          <View style={styles.consentRow}><Text style={styles.consentName}>Ama</Text><Text style={styles.consentValue}>Consents</Text></View>
          <View style={styles.consentRow}><Text style={styles.consentName}>You</Text><Text style={styles.consentValue}>{publishConsent ? "Consent" : "Private"}</Text></View>
          <Pressable
            accessibilityState={{ checked: publishConsent }}
            accessibilityRole="checkbox"
            onPress={() => setPublishConsent((value) => !value)}
            style={styles.consentButton}
          >
            <Text style={styles.consentButtonText}>
              {publishConsent ? "Withdraw my consent" : "Consent to a redacted edition"}
            </Text>
          </Pressable>
          <Text accessibilityLiveRegion="polite" style={styles.consentStatus}>
            {publishConsent ? "Both consent. A redacted preview can be prepared." : "This draft remains private."}
          </Text>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe:{backgroundColor:"#F5EAD8",flex:1},content:{padding:20,paddingBottom:56},
  topbar:{flexDirection:"row",justifyContent:"space-between"},control:{alignItems:"center",borderColor:"#927982",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:48,paddingHorizontal:17},controlText:{color:"#291720",fontFamily:"Outfit_700Bold"},
  eyebrow:{color:"#9B315D",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.2,marginTop:50},
  title:{color:"#291720",fontFamily:"Outfit_800ExtraBold",fontSize:54,letterSpacing:-3.4,lineHeight:49,marginTop:14},
  body:{color:"#705C67",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:25,marginTop:22},
  paper:{backgroundColor:"#FFFAF1",borderColor:"#DDCEC2",borderRadius:26,borderWidth:1,marginTop:36,padding:22},
  meta:{color:"#705C67",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},
  storyTitle:{color:"#291720",fontFamily:"Outfit_700Bold",fontSize:40,letterSpacing:-2,lineHeight:42,marginVertical:36},
  passage:{borderTopColor:"#DFD2C8",borderTopWidth:1,paddingVertical:22},who:{color:"#9B315D",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.2},
  passageText:{color:"#291720",fontFamily:"Outfit_400Regular",fontSize:19,lineHeight:29,marginTop:10},
  composer:{backgroundColor:"#291720",borderRadius:20,marginTop:28,padding:20},composerLabel:{color:"#FF91A6",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},
  composerTitle:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:30,letterSpacing:-1.4,marginTop:8},
  input:{backgroundColor:"#FFFAF1",borderRadius:14,color:"#291720",fontFamily:"Outfit_400Regular",marginTop:18,minHeight:120,padding:14,textAlignVertical:"top"},
  count:{color:"rgba(255,243,230,.6)",fontFamily:"Outfit_400Regular",fontSize:11,marginTop:6,textAlign:"right"},
  primary:{alignItems:"center",backgroundColor:"#FFAD3D",borderRadius:999,justifyContent:"center",marginTop:16,minHeight:52},primaryText:{color:"#291720",fontFamily:"Outfit_700Bold"},disabled:{opacity:.4},
  publish:{backgroundColor:"#291720",borderRadius:26,marginTop:38,padding:22},publishTitle:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:38,letterSpacing:-2,lineHeight:38,marginTop:12},
  publishCopy:{color:"rgba(255,243,230,.65)",fontFamily:"Outfit_400Regular",lineHeight:23,marginBottom:18,marginTop:14},
  consentRow:{borderBottomColor:"rgba(255,243,230,.18)",borderBottomWidth:1,flexDirection:"row",justifyContent:"space-between",paddingVertical:14},
  consentName:{color:"#FFF3E6",fontFamily:"Outfit_600SemiBold"},consentValue:{color:"#FFB7C4",fontFamily:"Outfit_700Bold"},
  consentButton:{alignItems:"center",backgroundColor:"#FFAD3D",borderRadius:999,justifyContent:"center",marginTop:22,minHeight:52,paddingHorizontal:12},
  consentButtonText:{color:"#291720",fontFamily:"Outfit_700Bold"},consentStatus:{color:"rgba(255,243,230,.65)",fontFamily:"Outfit_400Regular",lineHeight:22,marginTop:16},
});
