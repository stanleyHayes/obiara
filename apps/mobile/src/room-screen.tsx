import { type Href, useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const messages = [
  ["Ama", "Home feels like people making room for one another."],
  ["You", "Mine sounds like highlife and too many voices in one kitchen."],
  ["Ama", "What is one tradition you would keep?"],
] as const;

export function RoomScreen() {
  const router = useRouter();
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable onPress={() => router.push("/fie/dan-mu" as Href)} style={styles.control}><Text style={styles.controlText}>Dan mu</Text></Pressable>
          <Pressable style={styles.control}><Text style={styles.controlText}>Safety</Text></Pressable>
        </View>
        <Text style={styles.eyebrow}>GUIDED ROOM · THEME ONE</Text>
        <Text accessibilityRole="header" style={styles.title}>Make room for honesty.</Text>
        <Text style={styles.body}>Strict alternation, no read receipts, streaks or public activity.</Text>
        <View style={styles.status}><Text style={styles.statusLabel}>YOUR TURN</Text><Text style={styles.statusCopy}>Nothing needs an immediate response.</Text></View>
        <View style={styles.timeline}>
          {messages.map(([who, message]) => (
            <View key={message} style={[styles.message, who === "You" && styles.mine]}>
              <Text style={[styles.who, who === "You" && styles.mineText]}>{who}</Text>
              <Text style={[styles.messageText, who === "You" && styles.mineText]}>{message}</Text>
              <Text style={[styles.meta, who === "You" && styles.mineMeta]}>Voice · private transcript</Text>
            </View>
          ))}
        </View>
        <View style={styles.composer}>
          <Text style={styles.composerTitle}>Speak when it feels right.</Text>
          <Text style={styles.composerCopy}>A voice draft is saved before one deliberate send.</Text>
          <Pressable style={styles.record}><Text style={styles.recordText}>Record voice reply</Text></Pressable>
          <Pressable style={styles.pause}><Text style={styles.pauseText}>Pause this room</Text></Pressable>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe:{backgroundColor:"#1D0B18",flex:1},content:{padding:20,paddingBottom:56},topbar:{flexDirection:"row",justifyContent:"space-between"},control:{alignItems:"center",borderColor:"rgba(255,243,230,.3)",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:48,paddingHorizontal:18},controlText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},eyebrow:{color:"#FF849B",fontFamily:"Outfit_700Bold",fontSize:11,letterSpacing:1.2,marginTop:54},title:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:58,letterSpacing:-3.8,lineHeight:52,marginTop:16},body:{color:"rgba(255,243,230,.65)",fontFamily:"Outfit_400Regular",fontSize:17,lineHeight:27,marginTop:24},status:{borderColor:"rgba(255,173,61,.4)",borderRadius:18,borderWidth:1,marginTop:30,padding:18},statusLabel:{color:"#FFB44F",fontFamily:"Outfit_700Bold",fontSize:12},statusCopy:{color:"rgba(255,243,230,.6)",fontFamily:"Outfit_400Regular",marginTop:6},timeline:{gap:10,marginTop:32},message:{alignSelf:"flex-start",backgroundColor:"rgba(255,243,230,.09)",borderRadius:18,maxWidth:"88%",padding:20},mine:{alignSelf:"flex-end",backgroundColor:"#FFF0D9"},who:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},messageText:{color:"#FFF3E6",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:24,marginTop:18},meta:{color:"rgba(255,243,230,.5)",fontFamily:"Outfit_400Regular",fontSize:11,marginTop:18},mineText:{color:"#2A1022"},mineMeta:{color:"#765F70"},composer:{borderTopColor:"rgba(255,243,230,.12)",borderTopWidth:1,marginTop:42,paddingTop:36},composerTitle:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:34,letterSpacing:-1.6},composerCopy:{color:"rgba(255,243,230,.6)",fontFamily:"Outfit_400Regular",lineHeight:22,marginTop:10},record:{alignItems:"center",backgroundColor:"#FFAD3D",borderRadius:999,justifyContent:"center",marginTop:26,minHeight:52},recordText:{color:"#2A1022",fontFamily:"Outfit_700Bold"},pause:{alignItems:"center",borderColor:"rgba(255,243,230,.3)",borderRadius:999,borderWidth:1,justifyContent:"center",marginTop:10,minHeight:52},pauseText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"}
});
