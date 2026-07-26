import { createTheme, type Theme } from "@mui/material/styles";
import {
  obiaraAccessibility,
  obiaraElevation,
  obiaraMotion,
  obiaraRadii,
  obiaraSemanticColors,
  obiaraSpacing,
  obiaraTypography,
} from "@obiara/design-tokens";

export interface ObiaraThemePreferences {
  mode?: "light" | "dark";
  highContrast?: boolean;
  reducedMotion?: boolean;
}

function paletteFor(preferences: ObiaraThemePreferences) {
  if (preferences.highContrast) {
    const colors = obiaraSemanticColors.highContrast;
    return {
      mode: "light" as const,
      primary: { main: colors.action, contrastText: colors.actionText },
      secondary: { main: colors.focus, contrastText: colors.surface },
      error: { main: colors.danger },
      success: { main: colors.positive },
      background: { default: colors.canvas, paper: colors.surface },
      text: { primary: colors.text, secondary: colors.text },
      divider: colors.border,
    };
  }

  const mode = preferences.mode ?? "light";
  const colors = obiaraSemanticColors[mode];
  return {
    mode,
    primary: { main: colors.action, contrastText: colors.actionText },
    secondary: { main: colors.accent, contrastText: colors.accentText },
    success: {
      main: colors.positive,
      contrastText: mode === "light" ? colors.surface : colors.canvas,
    },
    warning: {
      main: colors.warning,
      contrastText: mode === "light" ? colors.surface : colors.canvas,
    },
    error: {
      main: colors.danger,
      contrastText: mode === "light" ? colors.surface : colors.canvas,
    },
    info: {
      main: colors.info,
      contrastText: mode === "light" ? colors.surface : colors.canvas,
    },
    background: { default: colors.canvas, paper: colors.surface },
    text: { primary: colors.text, secondary: colors.textMuted },
    divider: colors.border,
    action: {
      disabled: colors.disabled,
      disabledBackground: colors.disabledSurface,
    },
  };
}

export function createObiaraTheme(
  preferences: ObiaraThemePreferences = {},
): Theme {
  const mode = preferences.mode ?? "light";
  const semantic = preferences.highContrast
    ? obiaraSemanticColors.highContrast
    : obiaraSemanticColors[mode];
  const focusRing = `${obiaraAccessibility.focusRingWidth}px solid ${semantic.focus}`;
  const transitionDuration = preferences.reducedMotion
    ? obiaraMotion.durationMs.instant
    : obiaraMotion.durationMs.standard;

  return createTheme({
    palette: paletteFor(preferences),
    typography: {
      fontFamily: obiaraTypography.fontFamily,
      h1: {
        fontSize: obiaraTypography.scale.display.fontSize,
        fontWeight: obiaraTypography.scale.display.fontWeight,
        letterSpacing: "-0.045em",
        lineHeight:
          obiaraTypography.scale.display.lineHeight /
          obiaraTypography.scale.display.fontSize,
      },
      h2: {
        fontSize: obiaraTypography.scale.headline.fontSize,
        fontWeight: obiaraTypography.scale.headline.fontWeight,
        letterSpacing: "-0.035em",
        lineHeight:
          obiaraTypography.scale.headline.lineHeight /
          obiaraTypography.scale.headline.fontSize,
      },
      h3: {
        fontSize: obiaraTypography.scale.titleLarge.fontSize,
        fontWeight: obiaraTypography.scale.titleLarge.fontWeight,
        letterSpacing: "-0.025em",
      },
      body1: {
        fontSize: obiaraTypography.scale.body.fontSize,
        lineHeight:
          obiaraTypography.scale.body.lineHeight /
          obiaraTypography.scale.body.fontSize,
      },
      body2: {
        fontSize: obiaraTypography.scale.bodySmall.fontSize,
        lineHeight:
          obiaraTypography.scale.bodySmall.lineHeight /
          obiaraTypography.scale.bodySmall.fontSize,
      },
      button: {
        fontWeight: obiaraTypography.weights.bold,
        textTransform: "none",
        letterSpacing: "-0.01em",
      },
    },
    shape: {
      borderRadius: obiaraRadii.medium,
    },
    spacing: obiaraSpacing.sm,
    transitions: {
      duration: {
        shortest: transitionDuration,
        shorter: transitionDuration,
        short: transitionDuration,
        standard: transitionDuration,
        complex: preferences.reducedMotion
          ? 0
          : obiaraMotion.durationMs.deliberate,
        enteringScreen: transitionDuration,
        leavingScreen: transitionDuration,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          ":root": {
            colorScheme: preferences.highContrast ? "light" : mode,
          },
          "*:focus-visible": {
            outline: focusRing,
            outlineOffset: 3,
          },
          "@media (prefers-reduced-motion: reduce)": {
            "*, *::before, *::after": {
              animationDuration: "0.01ms !important",
              animationIterationCount: "1 !important",
              scrollBehavior: "auto !important",
              transitionDuration: "0.01ms !important",
            },
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            minHeight: obiaraAccessibility.minimumTouchTarget,
            borderRadius: obiaraRadii.pill,
            paddingInline: obiaraSpacing.xl,
            transitionDuration: `${transitionDuration}ms`,
            "&:focus-visible": {
              outline: focusRing,
              outlineOffset: 3,
            },
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            minWidth: obiaraAccessibility.minimumTouchTarget,
            minHeight: obiaraAccessibility.minimumTouchTarget,
            "&:focus-visible": {
              outline: focusRing,
              outlineOffset: 3,
            },
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            borderRadius: obiaraRadii.large,
            boxShadow: preferences.highContrast
              ? "none"
              : obiaraElevation.web.soft,
            backgroundImage: "none",
            border: preferences.highContrast
              ? `2px solid ${semantic.border}`
              : undefined,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            minHeight: 32,
            borderRadius: obiaraRadii.pill,
            fontWeight: obiaraTypography.weights.semibold,
          },
        },
      },
      MuiInputBase: {
        styleOverrides: {
          root: {
            minHeight: obiaraAccessibility.minimumTouchTarget,
          },
        },
      },
    },
  });
}

export const obiaraTheme = createObiaraTheme();
