import {
  Outfit_400Regular,
  Outfit_500Medium,
  Outfit_600SemiBold,
  Outfit_700Bold,
  Outfit_800ExtraBold,
  useFonts,
} from "@expo-google-fonts/outfit";
import { Stack } from "expo-router";
import { Platform, View } from "react-native";
import { MobileThemeProvider } from "@obiara/ui-mobile";

// On web the app presents as a phone-width frame centered on a soft
// canvas; on device it fills the screen.
const frameStyle =
  Platform.OS === "web"
    ? { flex: 1, maxWidth: 520, width: "100%" as const }
    : { flex: 1 };

const canvasStyle =
  Platform.OS === "web"
    ? { alignItems: "center" as const, backgroundColor: "#E9DED6", flex: 1 }
    : { flex: 1 };

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    Outfit_400Regular,
    Outfit_500Medium,
    Outfit_600SemiBold,
    Outfit_700Bold,
    Outfit_800ExtraBold,
  });

  if (!fontsLoaded) {
    return <View style={{ backgroundColor: "#FFF3E6", flex: 1 }} />;
  }

  return (
    <MobileThemeProvider>
      <View style={canvasStyle}>
        <View style={frameStyle}>
          <Stack screenOptions={{ headerShown: false }} />
        </View>
      </View>
    </MobileThemeProvider>
  );
}
