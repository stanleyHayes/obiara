import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

type Stage = "ready" | "choosing" | "locked" | "revealed" | "reconnecting";
type Gesture = "together" | "apart";

export function AmpeScreen() {
  const router = useRouter();
  const [stage, setStage] = useState<Stage>("ready");
  const [choice, setChoice] = useState<Gesture | null>(null);
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable onPress={() => router.push("/fie/dan-mu/rooms/room_7Qp9kL2xV4mN8zTa" as Href)} style={styles.control}><Text style={styles.controlText}>Private room</Text></Pressable>
          <Pressable style={styles.control}><Text style={styles.controlText}>Safety</Text></Pressable>
        </View>
        <Text style={styles.eyebrow}>NO CAMERA · NO BODY INFERENCE</Text>
        <Text accessibilityRole="header" style={styles.title}>Meet in the same beat.</Text>
        <Text style={styles.body}>Choose privately. Both gestures reveal together; a weak connection pauses without a forfeit.</Text>
        <View style={styles.stage}>
          <Text style={styles.meta}>ROUND 03 · LOW-DATA PULSE</Text>
          <View style={styles.players}>
            <View style={styles.player}><View style={styles.avatar}><Text style={styles.avatarText}>A</Text></View><Text style={styles.playerName}>Ama</Text><Text style={styles.ready}>Ready</Text></View>
            <View style={styles.player}><View style={styles.avatar}><Text style={styles.avatarText}>Y</Text></View><Text style={styles.playerName}>You</Text><Text style={styles.ready}>{stage === "ready" ? "Not ready" : "Ready"}</Text></View>
          </View>
          <Text accessibilityRole="header" style={styles.stageTitle}>
            {stage === "ready" ? "Join the next beat." : stage === "choosing" ? "Choose in private." : stage === "locked" ? "Your choice is held." : stage === "reconnecting" ? "Holding the round." : "Together—then reveal."}
          </Text>
          {stage === "ready" ? <Pressable onPress={() => setStage("choosing")} style={styles.primary}><Text style={styles.primaryText}>I’m ready</Text></Pressable> : null}
          {stage === "choosing" ? (
            <>
              <View accessibilityLabel="Private gesture choice" style={styles.choices}>
                {(["together","apart"] as const).map((gesture) => (
                  <Pressable accessibilityRole="radio" accessibilityState={{checked:choice===gesture}} key={gesture} onPress={() => setChoice(gesture)} style={[styles.choice,choice===gesture&&styles.choiceSelected]}>
                    <Text style={styles.choiceTitle}>{gesture === "together" ? "Together" : "Apart"}</Text>
                    <Text style={styles.choiceCopy}>{gesture === "together" ? "Feet meet" : "Feet open"}</Text>
                  </Pressable>
                ))}
              </View>
              <Pressable disabled={!choice} onPress={() => setStage("locked")} style={[styles.primary,!choice&&styles.disabled]}><Text style={styles.primaryText}>Lock my gesture</Text></Pressable>
            </>
          ) : null}
          {stage === "locked" ? <View style={styles.notice}><Text style={styles.noticeText}>Encrypted and hidden from Ama.</Text><Pressable onPress={() => setStage("revealed")} style={styles.noticeButton}><Text style={styles.noticeButtonText}>Reveal together</Text></Pressable></View> : null}
          {stage === "reconnecting" ? <View style={styles.notice}><Text style={styles.noticeText}>No forfeit. Your choice remains hidden.</Text><Pressable onPress={() => setStage(choice ? "locked" : "choosing")} style={styles.noticeButton}><Text style={styles.noticeButtonText}>Reconnect safely</Text></Pressable></View> : null}
          {stage === "revealed" ? <View accessibilityLiveRegion="polite" style={styles.reveal}><Text style={styles.revealLabel}>AMA · TOGETHER</Text><Text style={styles.revealLabel}>YOU · {choice?.toUpperCase()}</Text><Text style={styles.revealCopy}>No public score or profile signal was created.</Text></View> : null}
          <Pressable disabled={stage==="revealed"} onPress={() => setStage("reconnecting")} style={styles.secondary}><Text style={styles.secondaryText}>Test weak connection</Text></Pressable>
          <Pressable style={styles.secondary}><Text style={styles.secondaryText}>Leave round</Text></Pressable>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles=StyleSheet.create({
  safe:{backgroundColor:"#190D17",flex:1},content:{padding:20,paddingBottom:56},topbar:{flexDirection:"row",justifyContent:"space-between"},
  control:{alignItems:"center",borderColor:"rgba(255,243,230,.35)",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:48,paddingHorizontal:17},controlText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},
  eyebrow:{color:"#FF91A6",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.2,marginTop:52},title:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:58,letterSpacing:-3.7,lineHeight:52,marginTop:14},
  body:{color:"rgba(255,243,230,.64)",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:25,marginTop:22},stage:{backgroundColor:"#FFF0D9",borderRadius:28,marginTop:38,padding:22},
  meta:{color:"#705C67",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},players:{flexDirection:"row",gap:24,justifyContent:"center",marginVertical:36},player:{alignItems:"center"},
  avatar:{alignItems:"center",backgroundColor:"#6D244F",borderRadius:32,height:64,justifyContent:"center",width:64},avatarText:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:22},playerName:{color:"#291720",fontFamily:"Outfit_700Bold",marginTop:8},ready:{color:"#168565",fontFamily:"Outfit_600SemiBold",fontSize:11,marginTop:3},
  stageTitle:{color:"#291720",fontFamily:"Outfit_800ExtraBold",fontSize:42,letterSpacing:-2.4,lineHeight:42,textAlign:"center"},choices:{flexDirection:"row",gap:8,marginTop:28},
  choice:{borderColor:"#CBB5C0",borderRadius:18,borderWidth:1,flex:1,minHeight:130,padding:18},choiceSelected:{borderColor:"#6D244F",borderWidth:4},choiceTitle:{color:"#291720",fontFamily:"Outfit_700Bold",fontSize:21},choiceCopy:{color:"#705C67",fontFamily:"Outfit_400Regular",marginTop:8},
  primary:{alignItems:"center",backgroundColor:"#6D244F",borderRadius:999,justifyContent:"center",marginTop:24,minHeight:52},primaryText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},disabled:{opacity:.4},
  notice:{backgroundColor:"#291720",borderRadius:18,marginTop:26,padding:18},noticeText:{color:"#FFF3E6",fontFamily:"Outfit_400Regular"},noticeButton:{alignItems:"center",backgroundColor:"#FFAD3D",borderRadius:999,justifyContent:"center",marginTop:14,minHeight:48},noticeButtonText:{color:"#291720",fontFamily:"Outfit_700Bold"},
  reveal:{backgroundColor:"#291720",borderRadius:18,gap:10,marginTop:26,padding:20},revealLabel:{color:"#FFF3E6",fontFamily:"Outfit_700Bold",fontSize:18},revealCopy:{color:"rgba(255,243,230,.65)",fontFamily:"Outfit_400Regular",marginTop:8},
  secondary:{alignItems:"center",borderColor:"#806876",borderRadius:999,borderWidth:1,justifyContent:"center",marginTop:10,minHeight:50},secondaryText:{color:"#291720",fontFamily:"Outfit_700Bold"},
});
