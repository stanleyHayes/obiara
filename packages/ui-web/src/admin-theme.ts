import { createTheme, type Theme } from "@mui/material/styles";
import {
  obiaraAccessibility,
  obiaraElevation,
  obiaraRadii,
  obiaraSemanticColors,
  obiaraSpacing,
  obiaraTypography,
} from "@obiara/design-tokens";

export interface ObiaraAdminThemePreferences {
  highContrast?: boolean;
  reducedMotion?: boolean;
}

export const obiaraAdminStatusColors = {
  healthy: obiaraSemanticColors.light.positive,
  attention: obiaraSemanticColors.light.warning,
  critical: obiaraSemanticColors.light.danger,
  informational: obiaraSemanticColors.light.info,
} as const;

export function createObiaraAdminTheme(
  preferences: ObiaraAdminThemePreferences = {},
): Theme {
  const colors = preferences.highContrast
    ? obiaraSemanticColors.highContrast
    : obiaraSemanticColors.light;
  const textSecondary = preferences.highContrast
    ? obiaraSemanticColors.highContrast.text
    : obiaraSemanticColors.light.textMuted;
  const focusRing = `${obiaraAccessibility.focusRingWidth}px solid ${colors.focus}`;
  const transitionDuration = preferences.reducedMotion ? 0 : 120;

  return createTheme({
    palette: {
      mode: "light",
      primary: {
        main: colors.action,
        contrastText: colors.actionText,
      },
      secondary: {
        main: colors.focus,
        contrastText: colors.surface,
      },
      success: { main: obiaraAdminStatusColors.healthy },
      warning: { main: obiaraAdminStatusColors.attention },
      error: { main: obiaraAdminStatusColors.critical },
      info: { main: obiaraAdminStatusColors.informational },
      background: {
        default: preferences.highContrast ? colors.canvas : "#F8F4F2",
        paper: colors.surface,
      },
      text: {
        primary: colors.text,
        secondary: textSecondary,
      },
      divider: colors.border,
    },
    typography: {
      fontFamily: obiaraTypography.fontFamily,
      h1: {
        fontSize: obiaraTypography.scale.headline.fontSize,
        fontWeight: obiaraTypography.weights.extrabold,
        lineHeight:
          obiaraTypography.scale.headline.lineHeight /
          obiaraTypography.scale.headline.fontSize,
      },
      h2: {
        fontSize: obiaraTypography.scale.title.fontSize,
        fontWeight: obiaraTypography.weights.bold,
      },
      body1: {
        fontSize: obiaraTypography.scale.bodySmall.fontSize,
        lineHeight:
          obiaraTypography.scale.bodySmall.lineHeight /
          obiaraTypography.scale.bodySmall.fontSize,
      },
      button: {
        fontSize: obiaraTypography.scale.label.fontSize,
        fontWeight: obiaraTypography.weights.bold,
        textTransform: "none",
      },
    },
    shape: { borderRadius: obiaraRadii.small },
    spacing: obiaraSpacing.sm,
    transitions: {
      duration: {
        shortest: transitionDuration,
        shorter: transitionDuration,
        short: transitionDuration,
        standard: transitionDuration,
        complex: preferences.reducedMotion ? 0 : 220,
        enteringScreen: transitionDuration,
        leavingScreen: transitionDuration,
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          ":root": { colorScheme: "light" },
          "*:focus-visible": {
            outline: focusRing,
            outlineOffset: 2,
          },
          "@media (prefers-reduced-motion: reduce)": {
            "*, *::before, *::after": {
              animationDuration: "0.01ms !important",
              animationIterationCount: "1 !important",
              transitionDuration: "0.01ms !important",
            },
          },
        },
      },
      MuiButton: {
        defaultProps: { disableElevation: true },
        styleOverrides: {
          root: {
            minHeight: obiaraAccessibility.minimumTouchTarget,
            borderRadius: obiaraRadii.small,
            paddingInline: obiaraSpacing.lg,
            "&:focus-visible": {
              outline: focusRing,
              outlineOffset: 2,
            },
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            minHeight: obiaraAccessibility.minimumTouchTarget,
            minWidth: obiaraAccessibility.minimumTouchTarget,
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            borderRadius: obiaraRadii.small,
            backgroundImage: "none",
            border: `1px solid ${colors.border}`,
            boxShadow: preferences.highContrast
              ? "none"
              : obiaraElevation.web.soft,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            minHeight: 32,
            borderRadius: obiaraRadii.small,
            fontWeight: obiaraTypography.weights.bold,
          },
        },
      },
      MuiTableCell: {
        styleOverrides: {
          head: {
            color: textSecondary,
            fontSize: obiaraTypography.scale.caption.fontSize,
            fontWeight: obiaraTypography.weights.bold,
            letterSpacing: "0.06em",
            textTransform: "uppercase",
          },
          body: {
            fontSize: obiaraTypography.scale.bodySmall.fontSize,
          },
        },
      },
    },
  });
}

export const obiaraAdminTheme = createObiaraAdminTheme();
