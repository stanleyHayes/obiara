import type { ComponentType } from "react";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { MobileThemeProvider } from "@obiara/ui-mobile";

export function GestureScreen({
  eyebrow,
  title,
  description,
  Gesture,
}: Readonly<{
  eyebrow: string;
  title: string;
  description: string;
  Gesture: ComponentType;
}>) {
  return (
    <MobileThemeProvider colorMode="light">
      <SafeAreaView edges={["bottom"]} style={styles.safeArea}>
        <ScrollView contentContainerStyle={styles.content}>
          <View style={styles.header}>
            <Text style={styles.eyebrow}>{eyebrow}</Text>
            <Text accessibilityRole="header" style={styles.title}>
              {title}
            </Text>
            <Text style={styles.description}>{description}</Text>
          </View>
          <Gesture />
        </ScrollView>
      </SafeAreaView>
    </MobileThemeProvider>
  );
}

const styles = StyleSheet.create({
  content: { gap: 24, padding: 20, paddingBottom: 48 },
  description: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 17,
    lineHeight: 26,
  },
  eyebrow: {
    color: "#C54267",
    fontFamily: "Outfit_700Bold",
    fontSize: 12,
    letterSpacing: 1.2,
  },
  header: { gap: 10 },
  safeArea: { backgroundColor: "#FFF8F1", flex: 1 },
  title: {
    color: "#3A0E2E",
    fontFamily: "Outfit_700Bold",
    fontSize: 36,
    lineHeight: 41,
  },
});
