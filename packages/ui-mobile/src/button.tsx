import { ActivityIndicator, StyleSheet, Text, View } from "react-native";
import { Pressable, type PressableProps } from "./pressable";
import { useMobileTheme } from "./theme-provider";

export type ButtonVariant = "primary" | "secondary" | "danger";

export interface ButtonProps extends Omit<
  PressableProps,
  "accessibilityRole" | "children"
> {
  readonly label: string;
  readonly busy?: boolean;
  readonly leading?: React.ReactNode;
  readonly variant?: ButtonVariant;
}

export function Button({
  label,
  busy = false,
  disabled = false,
  leading,
  variant = "primary",
  ...props
}: ButtonProps) {
  const theme = useMobileTheme();
  const unavailable = disabled || busy;
  const palette =
    variant === "secondary"
      ? {
          background: theme.colors.surfaceRaised,
          border: theme.colors.border,
          foreground: theme.colors.text,
        }
      : variant === "danger"
        ? {
            background: theme.colors.danger,
            border: theme.colors.danger,
            foreground: theme.colors.actionText,
          }
        : {
            background: theme.colors.action,
            border: theme.colors.action,
            foreground: theme.colors.actionText,
          };

  return (
    <Pressable
      {...props}
      accessibilityRole="button"
      accessibilityState={{ busy, disabled: unavailable }}
      disabled={unavailable}
      style={[
        styles.button,
        {
          backgroundColor: unavailable
            ? theme.colors.disabledSurface
            : palette.background,
          borderColor: unavailable ? theme.colors.disabled : palette.border,
          borderRadius: theme.radii.pill,
          paddingHorizontal: theme.spacing.xl,
        },
        props.style,
      ]}
    >
      <View style={styles.content}>
        {busy ? (
          <ActivityIndicator
            accessibilityElementsHidden
            color={theme.colors.disabled}
            size="small"
          />
        ) : (
          leading
        )}
        <Text
          style={{
            color: unavailable ? theme.colors.disabled : palette.foreground,
            fontFamily: theme.typography.nativeFamilies.bold,
            fontSize: theme.typography.scale.label.fontSize,
            lineHeight: theme.typography.scale.label.lineHeight,
          }}
        >
          {label}
        </Text>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    alignItems: "center",
    borderWidth: 1,
    justifyContent: "center",
  },
  content: {
    alignItems: "center",
    flexDirection: "row",
    gap: 8,
    justifyContent: "center",
  },
});
