export type OkyeameCapability =
  | "feature_help"
  | "navigation_help"
  | "wording_help"
  | "matchmaking_decision"
  | "autonomous_romance"
  | "impersonation"
  | "counsel_disclosure"
  | "private_evidence"
  | "hidden_memory";

export interface OkyeameRequest {
  readonly capability: OkyeameCapability;
  readonly memberInvoked: boolean;
}

export interface OkyeameDecision {
  readonly allowed: boolean;
  readonly capability: OkyeameCapability;
  readonly disclosure: "AI_GUIDED_HELP";
  readonly heading: string;
  readonly message: string;
  readonly retainsPrompt: false;
}

const allowedCopy: Readonly<
  Record<
    "feature_help" | "navigation_help" | "wording_help",
    { readonly heading: string; readonly message: string }
  >
> = {
  feature_help: {
    heading: "Feature help is available",
    message:
      "I can explain how a Fie feature works. You remain in control of every action.",
  },
  navigation_help: {
    heading: "Navigation help is available",
    message:
      "I can point you to a Fie area. I cannot open private spaces or act for you.",
  },
  wording_help: {
    heading: "Wording help is available",
    message:
      "I can help you revise words you provide. I will not send or save them for you.",
  },
};

const refusalCopy: Readonly<
  Record<
    Exclude<
      OkyeameCapability,
      "feature_help" | "navigation_help" | "wording_help"
    >,
    string
  >
> = {
  matchmaking_decision:
    "I cannot choose, rank or recommend a person for you.",
  autonomous_romance:
    "I cannot start, continue or manage a romantic conversation for you.",
  impersonation: "I cannot speak as a member, host, elder or reviewer.",
  counsel_disclosure:
    "I cannot access or reveal counsel conversations. Counsel stays isolated from matching and guided help.",
  private_evidence:
    "I cannot access private rooms, trust paths, reports or safety evidence.",
  hidden_memory:
    "I do not build a hidden memory from this exchange. Start again when you want more help.",
};

export const okyeameAllowedCapabilities = [
  "feature_help",
  "navigation_help",
  "wording_help",
] as const;

export const okyeameRefusedCapabilities = [
  "matchmaking_decision",
  "autonomous_romance",
  "impersonation",
  "counsel_disclosure",
  "private_evidence",
  "hidden_memory",
] as const;

export function decideOkyeame(request: OkyeameRequest): OkyeameDecision {
  const base = {
    capability: request.capability,
    disclosure: "AI_GUIDED_HELP" as const,
    retainsPrompt: false as const,
  };

  if (!request.memberInvoked) {
    return {
      ...base,
      allowed: false,
      heading: "You must ask first",
      message:
        "Okyeame only responds after you request help. It never starts a conversation or acts in the background.",
    };
  }

  if (
    request.capability === "feature_help" ||
    request.capability === "navigation_help" ||
    request.capability === "wording_help"
  ) {
    return { ...base, allowed: true, ...allowedCopy[request.capability] };
  }

  return {
    ...base,
    allowed: false,
    heading: "That request is outside Okyeame",
    message: refusalCopy[request.capability],
  };
}
