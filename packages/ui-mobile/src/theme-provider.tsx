import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { AccessibilityInfo, useColorScheme } from "react-native";
import {
  createMobileTheme,
  type MobileColorMode,
  type MobileHaptic,
  type MobileTheme,
} from "./theme";

export interface HapticAdapter {
  trigger(feedback: Exclude<MobileHaptic, "none">): void | Promise<void>;
}

interface MobileThemeContextValue {
  readonly theme: MobileTheme;
  readonly haptics?: HapticAdapter;
}

const MobileThemeContext = createContext<MobileThemeContextValue | undefined>(
  undefined,
);

export interface MobileThemeProviderProps {
  readonly children: ReactNode;
  readonly colorMode?: MobileColorMode;
  readonly haptics?: HapticAdapter;
}

export function MobileThemeProvider({
  children,
  colorMode,
  haptics,
}: MobileThemeProviderProps) {
  const systemMode = useColorScheme();
  const [reduceMotion, setReduceMotion] = useState(false);

  useEffect(() => {
    let mounted = true;

    void AccessibilityInfo.isReduceMotionEnabled().then((enabled) => {
      if (mounted) setReduceMotion(enabled);
    });
    const subscription = AccessibilityInfo.addEventListener(
      "reduceMotionChanged",
      setReduceMotion,
    );

    return () => {
      mounted = false;
      subscription.remove();
    };
  }, []);

  const mode = colorMode ?? (systemMode === "dark" ? "dark" : "light");
  const value = useMemo(
    () => ({ theme: createMobileTheme(mode, reduceMotion), haptics }),
    [haptics, mode, reduceMotion],
  );

  return (
    <MobileThemeContext.Provider value={value}>
      {children}
    </MobileThemeContext.Provider>
  );
}

export function useMobileTheme(): MobileTheme {
  const context = useContext(MobileThemeContext);
  if (!context) {
    throw new Error("useMobileTheme must be used within MobileThemeProvider");
  }
  return context.theme;
}

export function useOptionalHaptic() {
  const context = useContext(MobileThemeContext);

  return (feedback: MobileHaptic) => {
    if (feedback === "none" || !context?.haptics) return;

    try {
      void Promise.resolve(context.haptics.trigger(feedback)).catch(() => {
        // Haptics are progressive enhancement and must not break an action.
      });
    } catch {
      // Native adapters can fail synchronously when hardware support is absent.
    }
  };
}
