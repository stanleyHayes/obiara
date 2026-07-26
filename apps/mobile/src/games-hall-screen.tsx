import { type Href, useRouter } from "expo-router";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

const games = [
  ["Oware", "Async strategy", "/fie/games/oware/game_4Nq8mK2xP7vR5tZa"],
  ["Ɛbɛ", "Reviewed proverb duel", "/fie/games/ebe/duel_8Km2qP4vN7xR5tZa"],
  ["Anansesɛm", "Private story relay", "/fie/games/anansesem/story_8Km2qP4vN7xR5tZa"],
  ["Ampe", "Low-data live pulse", "/fie/games/ampe/round_8Km2qP4vN7xR5tZa"],
] as const;

export function GamesHallScreen() {
  const router = useRouter();
  const [joined, setJoined] = useState(false);
  const [review, setReview] = useState<"clear"|"review"|"appealed">("clear");
  return (
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.topbar}>
          <Pressable onPress={() => router.push("/fie" as Href)} style={styles.control}><Text style={styles.controlText}>Fie</Text></Pressable>
          <Pressable style={styles.control}><Text style={styles.controlText}>Safety</Text></Pressable>
        </View>
        <Text style={styles.eyebrow}>PLAY REVEALS MOMENTS—NOT WORTH</Text>
        <Text accessibilityRole="header" style={styles.title}>A hall for skill, wit and shared stories.</Text>
        <Text style={styles.body}>Games stay separate from matching visibility. No global popularity rank or pay-to-win path.</Text>
        <Text accessibilityRole="header" style={styles.sectionTitle}>Your games.</Text>
        <View style={styles.gameGrid}>
          {games.map(([name,type,href]) => (
            <Pressable key={name} onPress={() => router.push(href as Href)} style={styles.gameCard}>
              <Text style={styles.gameType}>{type.toUpperCase()}</Text>
              <Text style={styles.gameName}>{name}</Text>
              <Text style={styles.gameOpen}>Open privately →</Text>
            </Pressable>
          ))}
        </View>
        <View style={styles.tournament}>
          <Text style={styles.cardEyebrow}>OPT-IN COHORT · 24 SEATS</Text>
          <Text accessibilityRole="header" style={styles.cardTitle}>Sunday Oware table.</Text>
          <Text style={styles.cardCopy}>Three calm rounds over one week. Standing stays inside this joined cohort and never affects matching.</Text>
          <View style={styles.facts}><Text style={styles.fact}>Sunday · 4 PM</Text><Text style={styles.fact}>18 of 24 seats</Text><Text style={styles.fact}>No entry fee</Text></View>
          <Pressable disabled={joined} onPress={() => setJoined(true)} style={[styles.primary,joined&&styles.disabled]}><Text style={styles.primaryText}>{joined?"Seat held privately":"Join this cohort"}</Text></Pressable>
        </View>
        <View style={styles.review}>
          <Text style={styles.reviewEyebrow}>FAIR-PLAY REVIEW</Text>
          <Text accessibilityRole="header" style={styles.reviewTitle}>Evidence before action.</Text>
          {review==="clear"?<><Text style={styles.reviewCopy}>Unusual play creates a private human review, never an automatic accusation.</Text><Pressable onPress={() => setReview("review")} style={styles.reviewButton}><Text style={styles.reviewButtonText}>Preview review path</Text></Pressable></>:null}
          {review==="review"?<><Text style={styles.reviewLabel}>REVIEW FP-84Q</Text><Text style={styles.reviewCopy}>Timing pattern needs human review. No public label or automatic penalty.</Text><Pressable onPress={() => setReview("appealed")} style={styles.reviewButton}><Text style={styles.reviewButtonText}>Submit private appeal</Text></Pressable></>:null}
          {review==="appealed"?<><Text style={styles.reviewLabel}>APPEAL RECEIVED</Text><Text style={styles.reviewCopy}>A different reviewer will look again. Your cohort sees no accusation.</Text></>:null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles=StyleSheet.create({
  safe:{backgroundColor:"#F7EFE2",flex:1},content:{padding:20,paddingBottom:56},topbar:{flexDirection:"row",justifyContent:"space-between"},
  control:{alignItems:"center",borderColor:"#8F7885",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:48,paddingHorizontal:18},controlText:{color:"#28161F",fontFamily:"Outfit_700Bold"},
  eyebrow:{color:"#9B315D",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.2,marginTop:52},title:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:52,letterSpacing:-3.3,lineHeight:48,marginTop:14},
  body:{color:"#705C67",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:25,marginTop:22},sectionTitle:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:40,letterSpacing:-2.2,marginTop:46},
  gameGrid:{gap:10,marginTop:18},gameCard:{backgroundColor:"#28161F",borderRadius:20,minHeight:170,padding:20},gameType:{color:"#FFB7C4",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},
  gameName:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:32,letterSpacing:-1.4,marginTop:28},gameOpen:{color:"#FFF3E6",fontFamily:"Outfit_700Bold",marginTop:"auto"},
  tournament:{backgroundColor:"#FFF0D9",borderRadius:26,marginTop:38,padding:22},cardEyebrow:{color:"#9B315D",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},
  cardTitle:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:40,letterSpacing:-2.1,lineHeight:40,marginTop:12},cardCopy:{color:"#705C67",fontFamily:"Outfit_400Regular",lineHeight:23,marginTop:14},
  facts:{flexDirection:"row",flexWrap:"wrap",gap:6,marginTop:20},fact:{borderColor:"#CAB5C0",borderRadius:999,borderWidth:1,color:"#28161F",fontFamily:"Outfit_600SemiBold",fontSize:10,paddingHorizontal:10,paddingVertical:7},
  primary:{alignItems:"center",backgroundColor:"#6D244F",borderRadius:999,justifyContent:"center",marginTop:24,minHeight:52},primaryText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},disabled:{opacity:.65},
  review:{backgroundColor:"#28161F",borderRadius:26,marginTop:28,padding:22},reviewEyebrow:{color:"#FFB7C4",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},
  reviewTitle:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:38,letterSpacing:-2,lineHeight:38,marginTop:12},reviewLabel:{color:"#FFB7C4",fontFamily:"Outfit_700Bold",fontSize:11,letterSpacing:1.1,marginTop:22},
  reviewCopy:{color:"rgba(255,243,230,.65)",fontFamily:"Outfit_400Regular",lineHeight:23,marginTop:14},reviewButton:{alignItems:"center",backgroundColor:"#FFAD3D",borderRadius:999,justifyContent:"center",marginTop:20,minHeight:52},
  reviewButtonText:{color:"#28161F",fontFamily:"Outfit_700Bold"},
});
