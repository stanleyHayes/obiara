import {
  obiaraAccessibility,
  obiaraElevation,
  obiaraFeedback,
  obiaraMotion,
  obiaraRadii,
  obiaraSemanticColors,
  obiaraSpacing,
  obiaraTypography,
} from "@obiara/design-tokens";

export type MobileColorMode = "light" | "dark";
export type MobileHaptic =
  (typeof obiaraFeedback.haptic)[keyof typeof obiaraFeedback.haptic];

export interface MobileMotion {
  readonly reduceMotion: boolean;
  readonly quick: number;
  readonly standard: number;
  readonly deliberate: number;
  readonly distance: number;
}

export function createMobileMotion(reduceMotion: boolean): MobileMotion {
  if (reduceMotion) {
    return {
      reduceMotion: true,
      quick: obiaraMotion.reduced.durationMs,
      standard: obiaraMotion.reduced.durationMs,
      deliberate: obiaraMotion.reduced.durationMs,
      distance: obiaraMotion.reduced.distance,
    };
  }

  return {
    reduceMotion: false,
    quick: obiaraMotion.durationMs.quick,
    standard: obiaraMotion.durationMs.standard,
    deliberate: obiaraMotion.durationMs.deliberate,
    distance: obiaraSpacing.md,
  };
}

export function createMobileTheme(mode: MobileColorMode, reduceMotion = false) {
  return {
    mode,
    colors: obiaraSemanticColors[mode],
    typography: obiaraTypography,
    spacing: obiaraSpacing,
    radii: obiaraRadii,
    elevation: obiaraElevation.native,
    accessibility: obiaraAccessibility,
    motion: createMobileMotion(reduceMotion),
    haptics: obiaraFeedback.haptic,
  } as const;
}

export type MobileTheme = ReturnType<typeof createMobileTheme>;
