import { Stack } from "expo-router";

export default function DesignLabLayout() {
  return (
    <Stack
      screenOptions={{
        headerBackTitle: "Lab",
        headerTitleStyle: { fontFamily: "Outfit_600SemiBold" },
      }}
    />
  );
}
