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
  black: "#000000",
} as const;

export const obiaraSemanticColors = {
  light: {
    canvas: obiaraColors.blushCream,
    surface: obiaraColors.paper,
    surfaceRaised: obiaraColors.white,
    text: obiaraColors.ink,
    textMuted: obiaraColors.softPlum,
    border: "#D8C4D0",
    focus: "#8A3D00",
    action: obiaraColors.deepPlum,
    actionText: obiaraColors.white,
    accent: obiaraColors.marigold,
    accentText: obiaraColors.ink,
    positive: "#08795A",
    positiveSurface: "#E3F7F0",
    warning: "#8A3D00",
    warningSurface: "#FFF0D6",
    danger: "#B4233F",
    dangerSurface: "#FFE7EC",
    info: "#285A9B",
    infoSurface: "#E9F1FF",
    disabled: "#8A7582",
    disabledSurface: "#EEE6EA",
    scrim: "rgba(38, 16, 31, 0.64)",
  },
  dark: {
    canvas: "#180B14",
    surface: obiaraColors.ink,
    surfaceRaised: "#45203A",
    text: obiaraColors.white,
    textMuted: "#E2CCD9",
    border: "#765469",
    focus: "#FFC66E",
    action: obiaraColors.marigold,
    actionText: obiaraColors.ink,
    accent: obiaraColors.hibiscus,
    accentText: obiaraColors.ink,
    positive: "#55D9B1",
    positiveSurface: "#123F34",
    warning: "#FFC66E",
    warningSurface: "#593414",
    danger: "#FF8FA5",
    dangerSurface: "#5E1E2E",
    info: "#9CC5FF",
    infoSurface: "#173A69",
    disabled: "#BBA5B2",
    disabledSurface: "#503C48",
    scrim: "rgba(0, 0, 0, 0.72)",
  },
  highContrast: {
    canvas: obiaraColors.white,
    surface: obiaraColors.white,
    text: obiaraColors.black,
    border: obiaraColors.black,
    focus: "#005FCC",
    action: obiaraColors.black,
    actionText: obiaraColors.white,
    danger: "#9B001B",
    positive: "#005A3E",
  },
} as const;

export const obiaraTypography = {
  fontFamily: "Outfit Variable, Outfit, sans-serif",
  nativeFamilies: {
    regular: "Outfit_400Regular",
    medium: "Outfit_500Medium",
    semibold: "Outfit_600SemiBold",
    bold: "Outfit_700Bold",
    extrabold: "Outfit_800ExtraBold",
  },
  weights: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
  scale: {
    display: { fontSize: 56, lineHeight: 62, fontWeight: 800 },
    headline: { fontSize: 40, lineHeight: 46, fontWeight: 800 },
    titleLarge: { fontSize: 30, lineHeight: 36, fontWeight: 700 },
    title: { fontSize: 24, lineHeight: 30, fontWeight: 700 },
    subtitle: { fontSize: 20, lineHeight: 26, fontWeight: 600 },
    bodyLarge: { fontSize: 18, lineHeight: 28, fontWeight: 400 },
    body: { fontSize: 16, lineHeight: 24, fontWeight: 400 },
    bodySmall: { fontSize: 14, lineHeight: 20, fontWeight: 400 },
    label: { fontSize: 14, lineHeight: 20, fontWeight: 600 },
    caption: { fontSize: 12, lineHeight: 18, fontWeight: 500 },
  },
} as const;

export const obiaraSpacing = {
  none: 0,
  xxs: 2,
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  "2xl": 32,
  "3xl": 48,
  "4xl": 64,
  "5xl": 96,
} as const;

export const obiaraRadii = {
  none: 0,
  small: 12,
  medium: 20,
  large: 30,
  xlarge: 40,
  pill: 999,
} as const;

export const obiaraElevation = {
  native: {
    flat: 0,
    raised: 2,
    overlay: 6,
    modal: 12,
  },
  web: {
    flat: "none",
    soft: "0 18px 48px rgba(58, 14, 46, 0.10)",
    lifted: "0 24px 70px rgba(58, 14, 46, 0.16)",
    overlay: "0 30px 90px rgba(38, 16, 31, 0.24)",
    focus: "0 0 0 3px rgba(255, 159, 28, 0.52)",
  },
} as const;

// Kept as a compatibility alias while theme adapters migrate to elevation.
export const obiaraShadows = {
  soft: obiaraElevation.web.soft,
  lifted: obiaraElevation.web.lifted,
} as const;

export const obiaraMotion = {
  durationMs: {
    instant: 0,
    quick: 120,
    standard: 220,
    deliberate: 360,
    ceremonial: 520,
  },
  easing: {
    standard: "cubic-bezier(0.2, 0, 0, 1)",
    enter: "cubic-bezier(0, 0, 0, 1)",
    exit: "cubic-bezier(0.3, 0, 1, 1)",
  },
  reduced: {
    durationMs: 0,
    distance: 0,
  },
} as const;

export const obiaraFeedback = {
  haptic: {
    none: "none",
    selection: "selection",
    light: "impact-light",
    deliberate: "impact-medium",
    warning: "notification-warning",
  },
  sound: {
    none: "none",
    confirmation: "confirmation-soft",
    warning: "warning-soft",
  },
} as const;

export const obiaraBreakpoints = {
  compact: 0,
  mobile: 390,
  tablet: 768,
  desktop: 1200,
  wide: 1536,
} as const;

export const obiaraLayers = {
  base: 0,
  navigation: 100,
  sticky: 200,
  overlay: 400,
  modal: 500,
  toast: 600,
} as const;

export const obiaraAccessibility = {
  minimumTouchTarget: 48,
  minimumTextContrast: 4.5,
  minimumLargeTextContrast: 3,
  focusRingWidth: 3,
} as const;

export const obiaraZones = {
  fie: { accent: obiaraColors.marigold, surface: "#FFF0D6" },
  sow: { accent: obiaraColors.palmGreen, surface: "#E3F7F0" },
  fires: { accent: obiaraColors.hibiscus, surface: "#FFE7EC" },
  gate: { accent: obiaraColors.deepPlum, surface: "#EEE6EA" },
  care: { accent: "#285A9B", surface: "#E9F1FF" },
} as const;

export const obiaraKente = {
  stripeWidth: 4,
  gap: 4,
  colors: [
    obiaraColors.marigold,
    obiaraColors.hibiscus,
    obiaraColors.palmGreen,
    obiaraColors.deepPlum,
  ],
  usage: {
    maximumStripes: 4,
    allowed: ["ceremony", "milestone", "section-accent"],
    prohibited: ["body-background", "paragraph-decoration", "form-field"],
  },
} as const;

export const obiaraAssets = {
  sourceRoot: "Obiara_Handover_Package/3_Brand/assets",
  logos: {
    markOnLight: "logo/svg/mark-color-onlight.svg",
    markOnDark: "logo/svg/mark-color-ondark.svg",
    horizontalOnLight: "logo/svg/lockup-h-color-onlight.svg",
    horizontalOnDark: "logo/svg/lockup-h-color-ondark.svg",
    stackedOnLight: "logo/svg/lockup-s-color-onlight.svg",
    stackedOnDark: "logo/svg/lockup-s-color-ondark.svg",
    monochromeInk: "logo/svg/mark-mono-ink.svg",
    monochromeCream: "logo/svg/mark-mono-cream.svg",
  },
  icons: {
    favicon48: "icons/favicon-48.png",
    favicon192: "icons/favicon-192.png",
    favicon512: "icons/favicon-512.png",
    appGold: "icons/app-icon-gold-1024.png",
    appInk: "icons/app-icon-ink-1024.png",
  },
  rules: {
    useSuppliedMarkOnly: true,
    allowPlaceholderMonogram: false,
    minimumClearspaceInMarkWidths: 0.5,
  },
} as const;

export type ObiaraColorMode = keyof typeof obiaraSemanticColors;
export type ObiaraZone = keyof typeof obiaraZones;
export type ObiaraSpacing = keyof typeof obiaraSpacing;

export const fontFamily = obiaraTypography.fontFamily;
