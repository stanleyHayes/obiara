export type OkyeameCapability = "resting" | "available";

export interface OkyeameBoundary {
  readonly label: string;
  readonly detail: string;
  readonly canStart: boolean;
}

export function okyeameBoundary(
  capability: OkyeameCapability,
): OkyeameBoundary {
  if (capability === "available") {
    return {
      label: "Capability available",
      detail:
        "Okyeame can explain a Fie feature, but cannot make decisions or speak as a person.",
      canStart: true,
    };
  }
  return {
    label: "Okyeame is resting",
    detail:
      "This guided help capability is not available right now. Your Fie access is unchanged.",
    canStart: false,
  };
}

export const okyeameLimits = [
  "Does not make relationship, safety or verification decisions",
  "Does not pretend to be a member, host or elder",
  "Does not reveal private rooms, paths or evidence",
] as const;
