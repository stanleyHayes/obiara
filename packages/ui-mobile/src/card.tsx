import { StyleSheet, View, type ViewProps } from "react-native";
import { useMobileTheme } from "./theme-provider";

export interface CardProps extends ViewProps {
  readonly elevated?: boolean;
}

export function Card({ elevated = false, style, ...props }: CardProps) {
  const theme = useMobileTheme();

  return (
    <View
      {...props}
      style={[
        styles.base,
        {
          backgroundColor: theme.colors.surface,
          borderColor: theme.colors.border,
          borderRadius: theme.radii.medium,
          elevation: elevated ? theme.elevation.raised : theme.elevation.flat,
          padding: theme.spacing.lg,
        },
        style,
      ]}
    />
  );
}

const styles = StyleSheet.create({
  base: {
    borderWidth: 1,
  },
});
