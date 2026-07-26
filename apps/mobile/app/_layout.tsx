import {
  Outfit_400Regular,
  Outfit_500Medium,
  Outfit_600SemiBold,
  Outfit_700Bold,
  Outfit_800ExtraBold,
  useFonts,
} from "@expo-google-fonts/outfit";
import { Stack } from "expo-router";
import { View } from "react-native";
import { MobileThemeProvider } from "../../../packages/ui-mobile/src";

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
      <Stack screenOptions={{ headerShown: false }} />
    </MobileThemeProvider>
  );
}
