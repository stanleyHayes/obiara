import {type Href,useRouter} from "expo-router";
import {useState} from "react";
import {Pressable,ScrollView,StyleSheet,Text,View} from "react-native";
import {SafeAreaView} from "react-native-safe-area-context";

type Feature="shared"|"trust"|"voice";
export function IntroductionExplanationScreen(){
  const router=useRouter();
  const [enabled,setEnabled]=useState<Record<Feature,boolean>>({shared:true,trust:true,voice:false});
  const [details,setDetails]=useState(false);
  const reasons=[enabled.shared?"You both chose family-minded partnership.":null,enabled.trust?"A private trust path is available to each of you.":null,enabled.voice?"You both consented to compare selected voice reflections.":null].filter(Boolean) as string[];
  const features:readonly [Feature,string,string][]=[["shared","Shared intentions","Choices you made in profile preferences."],["trust","Private trust context","Only the existence of a permitted path, never its people or shape."],["voice","Selected voice reflections","Off by default; only reflections both people choose."]];
  return <SafeAreaView style={s.safe}><ScrollView contentContainerStyle={s.content}>
    <View style={s.topbar}><Pressable onPress={()=>router.push("/fie/garden" as Href)} style={s.control}><Text style={s.controlText}>Garden</Text></Pressable><Pressable style={s.control}><Text style={s.controlText}>Safety</Text></Pressable></View>
    <Text style={s.eyebrow}>WHY THIS INTRODUCTION</Text><Text accessibilityRole="header" style={s.title}>Grounded reasons. Your controls.</Text><Text style={s.body}>Obiara explains permitted signals. It cannot promise compatibility, destiny or an outcome.</Text>
    <View style={s.person}><Text style={s.personMeta}>PRIVATE · NO PUBLIC MATCH SCORE</Text><Text style={s.personName}>Ama K.</Text></View>
    <View style={s.reasons}><Text style={s.darkEyebrow}>WHAT IS ACTIVE NOW</Text><Text accessibilityRole="header" style={s.reasonsTitle}>{reasons.length?"Why you may have something to explore.":"No explanation features are active."}</Text>
      {reasons.length?reasons.map((reason,index)=><View key={reason} style={s.reason}><Text style={s.reasonNumber}>0{index+1}</Text><Text style={s.reasonText}>{reason}</Text></View>):<View style={s.reason}><Text style={s.reasonText}>This introduction can rest. Re-enable a feature only if you want it considered.</Text></View>}
    </View>
    <Text accessibilityRole="header" style={s.sectionTitle}>Nothing hidden in the recipe.</Text><Text style={s.sectionCopy}>Turn a feature off to remove it from future explanations. Withdrawing consent does not lower visibility.</Text>
    {features.map(([id,label,copy])=><View key={id} style={s.feature}><View style={s.featureCopy}><Text style={s.featureTitle}>{label}</Text><Text style={s.featureBody}>{copy}</Text></View><Pressable accessibilityRole="switch" accessibilityState={{checked:enabled[id]}} onPress={()=>setEnabled(current=>({...current,[id]:!current[id]}))} style={[s.toggle,enabled[id]&&s.toggleOn]}><Text style={[s.toggleText,enabled[id]&&s.toggleTextOn]}>{enabled[id]?"On":"Off"}</Text></Pressable></View>)}
    <Pressable onPress={()=>setDetails(value=>!value)} style={s.detailsButton}><Text style={s.detailsText}>{details?"Hide system details":"Show system details"}</Text></Pressable>
    {details?<View style={s.details}><Text style={s.featureTitle}>Rules first · AI wording only</Text><Text style={s.featureBody}>Reciprocal preferences and a privacy-scoped trust-path rule selected the candidate. AI may phrase the explanation; it did not choose or rank the person.</Text></View>:null}
    <Pressable style={s.primary}><Text style={s.primaryText}>Open the introduction gently</Text></Pressable><Text style={s.footer}>No urgency, read receipt or public activity signal.</Text>
  </ScrollView></SafeAreaView>
}
const s=StyleSheet.create({
  safe:{backgroundColor:"#F7EFE2",flex:1},content:{padding:20,paddingBottom:56},topbar:{flexDirection:"row",justifyContent:"space-between"},control:{alignItems:"center",borderColor:"#8F7885",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:48,paddingHorizontal:18},controlText:{color:"#28161F",fontFamily:"Outfit_700Bold"},
  eyebrow:{color:"#9B315D",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.2,marginTop:52},title:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:53,letterSpacing:-3.4,lineHeight:49,marginTop:14},body:{color:"#705C67",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:25,marginTop:22},
  person:{borderColor:"#C5A15C",borderRadius:18,borderWidth:1,marginTop:28,padding:18},personMeta:{color:"#705C67",fontFamily:"Outfit_700Bold",fontSize:9,letterSpacing:1},personName:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:28,marginTop:12},
  reasons:{backgroundColor:"#28161F",borderRadius:26,marginTop:38,padding:22},darkEyebrow:{color:"#FFB7C4",fontFamily:"Outfit_700Bold",fontSize:10,letterSpacing:1.1},reasonsTitle:{color:"#FFF3E6",fontFamily:"Outfit_800ExtraBold",fontSize:38,letterSpacing:-2,lineHeight:38,marginBottom:20,marginTop:12},
  reason:{backgroundColor:"rgba(255,243,230,.08)",borderColor:"rgba(255,243,230,.14)",borderRadius:16,borderWidth:1,marginTop:8,padding:18},reasonNumber:{color:"#FFB7C4",fontFamily:"Outfit_700Bold",fontSize:10},reasonText:{color:"#FFF3E6",fontFamily:"Outfit_400Regular",fontSize:16,lineHeight:24,marginTop:10},
  sectionTitle:{color:"#28161F",fontFamily:"Outfit_800ExtraBold",fontSize:40,letterSpacing:-2.2,lineHeight:40,marginTop:44},sectionCopy:{color:"#705C67",fontFamily:"Outfit_400Regular",lineHeight:23,marginBottom:16,marginTop:12},
  feature:{alignItems:"center",borderColor:"#D1BDC7",borderRadius:17,borderWidth:1,flexDirection:"row",gap:12,justifyContent:"space-between",marginTop:9,padding:16},featureCopy:{flex:1},featureTitle:{color:"#28161F",fontFamily:"Outfit_700Bold",fontSize:16},featureBody:{color:"#705C67",fontFamily:"Outfit_400Regular",lineHeight:21,marginTop:5},
  toggle:{alignItems:"center",borderColor:"#6D244F",borderRadius:999,borderWidth:1,justifyContent:"center",minHeight:46,minWidth:58},toggleOn:{backgroundColor:"#6D244F"},toggleText:{color:"#6D244F",fontFamily:"Outfit_700Bold"},toggleTextOn:{color:"#FFF3E6"},
  detailsButton:{alignItems:"center",borderColor:"#6D244F",borderRadius:999,borderWidth:1,justifyContent:"center",marginTop:16,minHeight:50},detailsText:{color:"#6D244F",fontFamily:"Outfit_700Bold"},details:{backgroundColor:"#FFF0D9",borderRadius:17,marginTop:12,padding:18},
  primary:{alignItems:"center",backgroundColor:"#6D244F",borderRadius:999,justifyContent:"center",marginTop:30,minHeight:52},primaryText:{color:"#FFF3E6",fontFamily:"Outfit_700Bold"},footer:{color:"#705C67",fontFamily:"Outfit_400Regular",marginTop:14,textAlign:"center"},
});
