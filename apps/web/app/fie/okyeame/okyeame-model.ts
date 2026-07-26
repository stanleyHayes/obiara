import {
  decideOkyeame,
  type OkyeameCapability as OkyeameRequestCapability,
} from "@obiara/okyeame-policy";

export type OkyeameAvailability = "resting" | "available";

export interface OkyeameBoundary {
  readonly label: string;
  readonly detail: string;
  readonly canStart: boolean;
}

export function okyeameBoundary(
  capability: OkyeameAvailability,
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
  "No relationship, safety or verification decisions",
  "No impersonation or autonomous conversations",
  "No counsel, private evidence or hidden memory",
] as const;

export const okyeameRequests: readonly {
  readonly capability: OkyeameRequestCapability;
  readonly label: string;
}[] = [
  { capability: "feature_help", label: "Explain a Fie feature" },
  { capability: "navigation_help", label: "Help me find an area" },
  { capability: "wording_help", label: "Help revise my words" },
  { capability: "matchmaking_decision", label: "Choose a match for me" },
  { capability: "autonomous_romance", label: "Message someone for me" },
  { capability: "counsel_disclosure", label: "Show counsel notes" },
];

export function previewOkyeameRequest(capability: OkyeameRequestCapability) {
  return decideOkyeame({ capability, memberInvoked: true });
}
