export const obiaraColors = {
  marigold: "#FF9F1C",
  hibiscus: "#FF4D6D",
  deepPlum: "#3A0E2E",
  blushCream: "#FFF3E6",
  palmGreen: "#12A67C",
  ink: "#26101F",
  softPlum: "#6D435F",
  paper: "#FFFDFC",
  white: "#FFFFFF",
} as const;

export const obiaraTypography = {
  fontFamily: "Outfit Variable, Outfit, sans-serif",
  weights: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
} as const;

export const obiaraRadii = {
  small: 12,
  medium: 20,
  large: 30,
  pill: 999,
} as const;

export const obiaraShadows = {
  soft: "0 18px 48px rgba(58, 14, 46, 0.10)",
  lifted: "0 24px 70px rgba(58, 14, 46, 0.16)",
} as const;

export const fontFamily = obiaraTypography.fontFamily;
