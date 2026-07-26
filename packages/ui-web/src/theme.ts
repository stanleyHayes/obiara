import { createTheme } from "@mui/material/styles";
import {
  obiaraColors,
  obiaraRadii,
  obiaraShadows,
  obiaraTypography,
} from "@obiara/design-tokens";

export const obiaraTheme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: obiaraColors.deepPlum,
      contrastText: obiaraColors.blushCream,
    },
    secondary: {
      main: obiaraColors.marigold,
      contrastText: obiaraColors.ink,
    },
    success: {
      main: obiaraColors.palmGreen,
    },
    error: {
      main: obiaraColors.hibiscus,
    },
    background: {
      default: obiaraColors.blushCream,
      paper: obiaraColors.paper,
    },
    text: {
      primary: obiaraColors.ink,
      secondary: obiaraColors.softPlum,
    },
  },
  typography: {
    fontFamily: obiaraTypography.fontFamily,
    h1: {
      fontWeight: obiaraTypography.weights.extrabold,
      letterSpacing: "-0.045em",
      lineHeight: 0.98,
    },
    h2: {
      fontWeight: obiaraTypography.weights.bold,
      letterSpacing: "-0.035em",
      lineHeight: 1.04,
    },
    h3: {
      fontWeight: obiaraTypography.weights.bold,
      letterSpacing: "-0.025em",
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
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          minHeight: 48,
          borderRadius: obiaraRadii.pill,
          paddingInline: 22,
          boxShadow: "none",
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: obiaraRadii.large,
          boxShadow: obiaraShadows.soft,
          backgroundImage: "none",
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
  },
});
