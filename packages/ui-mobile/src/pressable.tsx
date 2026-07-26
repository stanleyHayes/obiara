import { useState, type ReactNode } from "react";
import {
  Pressable as NativePressable,
  type PressableProps as NativePressableProps,
  type StyleProp,
  type ViewStyle,
} from "react-native";
import { useMobileTheme, useOptionalHaptic } from "./theme-provider";
import type { MobileHaptic } from "./theme";

export interface PressableProps extends Omit<
  NativePressableProps,
  "children" | "style"
> {
  readonly children: ReactNode;
  readonly haptic?: MobileHaptic;
  readonly style?: StyleProp<ViewStyle>;
}

export function Pressable({
  children,
  disabled,
  haptic = "none",
  onFocus,
  onBlur,
  onPress,
  style,
  ...props
}: PressableProps) {
  const theme = useMobileTheme();
  const triggerHaptic = useOptionalHaptic();
  const [focused, setFocused] = useState(false);

  return (
    <NativePressable
      {...props}
      disabled={disabled}
      onBlur={(event) => {
        setFocused(false);
        onBlur?.(event);
      }}
      onFocus={(event) => {
        setFocused(true);
        onFocus?.(event);
      }}
      onPress={(event) => {
        if (disabled) return;
        triggerHaptic(haptic);
        onPress?.(event);
      }}
      style={({ pressed }) => [
        {
          minHeight: theme.accessibility.minimumTouchTarget,
          minWidth: theme.accessibility.minimumTouchTarget,
        },
        style,
        {
          opacity: disabled ? 0.56 : pressed ? 0.82 : 1,
          transform:
            pressed && !theme.motion.reduceMotion
              ? [{ scale: 0.98 }]
              : undefined,
          ...(focused
            ? {
                borderColor: theme.colors.focus,
                borderWidth: theme.accessibility.focusRingWidth,
              }
            : {}),
        },
      ]}
    >
      {children}
    </NativePressable>
  );
}
