import {
  Button,
  Card,
  MobileStateView,
  useMobileTheme,
} from "@obiara/ui-mobile";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

export default function MobileGallery() {
  const theme = useMobileTheme();

  return (
    <SafeAreaView style={{ backgroundColor: theme.colors.canvas, flex: 1 }}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.header}>
          <Text
            style={[
              styles.eyebrow,
              {
                color: theme.colors.accent,
                fontFamily: theme.typography.nativeFamilies.bold,
              },
            ]}
          >
            MOBILE COMPONENT GALLERY
          </Text>
          <Text
            accessibilityRole="header"
            style={[
              styles.title,
              {
                color: theme.colors.text,
                fontFamily: theme.typography.nativeFamilies.bold,
              },
            ]}
          >
            Shared pieces, real states.
          </Text>
        </View>

        <Card style={styles.stack}>
          <Text style={{ color: theme.colors.text }}>Actions</Text>
          <Button label="Enter the courtyard" onPress={() => undefined} />
          <Button
            label="Save for later"
            onPress={() => undefined}
            variant="secondary"
          />
        </Card>

        <MobileStateView
          actionLabel="Try again"
          body="Your place is safe. Reconnect when you are ready."
          kind="offline"
          onAction={() => undefined}
          title="You are offline"
        />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  content: { gap: 20, padding: 20, paddingBottom: 48 },
  eyebrow: { fontSize: 12, letterSpacing: 1.4 },
  header: { gap: 10 },
  stack: { gap: 14 },
  title: { fontSize: 34, lineHeight: 40 },
});
