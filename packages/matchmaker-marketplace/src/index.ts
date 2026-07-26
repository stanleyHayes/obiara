export type MatchmakerService = "consultation" | "curated" | "family_liaison";

export interface MatchmakerProfile {
  readonly id: string;
  readonly name: string;
  readonly licensed: true;
  readonly licenseRef: string;
  readonly languages: readonly string[];
  readonly specialties: readonly string[];
  readonly consultationFeeGhs: number;
  readonly rating: number;
  readonly completedEngagements: number;
}

export interface MarketplaceState {
  readonly profiles: readonly MatchmakerProfile[];
  readonly language: string;
  readonly selectedId: string | null;
  readonly service: MatchmakerService;
  readonly bookingConfirmed: boolean;
  readonly yourProposalConsent: boolean;
  readonly candidateProposalConsent: boolean;
}

export type MarketplaceAction =
  | { readonly type: "language"; readonly value: string }
  | { readonly type: "select"; readonly id: string }
  | { readonly type: "service"; readonly value: MatchmakerService }
  | { readonly type: "confirm-booking" }
  | { readonly type: "your-consent"; readonly value: boolean }
  | { readonly type: "candidate-consent"; readonly value: boolean };

export const initialMarketplaceState: MarketplaceState = {
  profiles: [
    {
      id: "agyina-esi",
      name: "Esi Agyapong",
      licensed: true,
      licenseRef: "AGY-204",
      languages: ["Twi", "English"],
      specialties: ["Family liaison", "Second marriages"],
      consultationFeeGhs: 120,
      rating: 4.9,
      completedEngagements: 38,
    },
    {
      id: "agyina-kwame",
      name: "Kwame Ofori",
      licensed: true,
      licenseRef: "AGY-118",
      languages: ["Ga", "English"],
      specialties: ["Diaspora families", "Introductions"],
      consultationFeeGhs: 180,
      rating: 4.8,
      completedEngagements: 51,
    },
  ],
  language: "All",
  selectedId: null,
  service: "consultation",
  bookingConfirmed: false,
  yourProposalConsent: false,
  candidateProposalConsent: false,
};

export function marketplaceReducer(
  state: MarketplaceState,
  action: MarketplaceAction,
): MarketplaceState {
  if (action.type === "language") {
    return { ...state, language: action.value, selectedId: null };
  }
  if (action.type === "select") {
    return state.profiles.some(
      (profile) => profile.id === action.id && profile.licensed,
    )
      ? { ...state, selectedId: action.id, bookingConfirmed: false }
      : state;
  }
  if (action.type === "service") {
    return { ...state, service: action.value, bookingConfirmed: false };
  }
  if (action.type === "confirm-booking") {
    return state.selectedId && state.service === "consultation"
      ? { ...state, bookingConfirmed: true }
      : state;
  }
  if (action.type === "your-consent") {
    return { ...state, yourProposalConsent: action.value };
  }
  if (action.type === "candidate-consent") {
    return { ...state, candidateProposalConsent: action.value };
  }
  return state;
}

export function canExposeCuratedProposal(state: MarketplaceState) {
  return state.yourProposalConsent && state.candidateProposalConsent;
}

export function visibleProfiles(state: MarketplaceState) {
  return state.language === "All"
    ? state.profiles
    : state.profiles.filter((profile) =>
        profile.languages.includes(state.language),
      );
}
