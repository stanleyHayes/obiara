// The Voice of Introduction: three prompts, recorded one at a time.
//
// The Build Pack (S-06) asks for three prompts, a 120-second meter and a
// re-record per prompt. The meter is a bound, not a target — a member who
// answers in twenty seconds has finished, and the only thing 120 seconds
// decides is when recording stops on its own.

export const maxPromptSeconds = 120;

/** The three questions, in the order they are asked. */
export const voicePrompts = [
  {
    id: "arrival",
    question: "What brought you here?",
    hint: "Say it the way you would to someone sitting beside you.",
  },
  {
    id: "ordinary",
    question: "What does an ordinary good day look like for you?",
    hint: "Not the highlights — the ordinary one.",
  },
  {
    id: "welcome",
    question: "What would make someone feel welcome in your company?",
    hint: "This is the one people listen to twice.",
  },
] as const;

export type PromptID = (typeof voicePrompts)[number]["id"];

export type PromptStage =
  "idle" | "recording" | "recorded" | "uploading" | "saved" | "failed";

export interface PromptState {
  readonly stage: PromptStage;
  readonly seconds: number;
  /** Set once the API has a recording for this prompt. */
  readonly introductionId: string | null;
  readonly error: string | null;
}

export interface VoiceState {
  readonly prompts: Readonly<Record<PromptID, PromptState>>;
  /** Only one prompt records at a time; two microphones is not a thing. */
  readonly active: PromptID | null;
}

const blankPrompt: PromptState = {
  stage: "idle",
  seconds: 0,
  introductionId: null,
  error: null,
};

export const initialVoiceState: VoiceState = {
  prompts: {
    arrival: blankPrompt,
    ordinary: blankPrompt,
    welcome: blankPrompt,
  },
  active: null,
};

export type VoiceAction =
  | { readonly type: "start"; readonly prompt: PromptID }
  | { readonly type: "tick"; readonly prompt: PromptID }
  | { readonly type: "stop"; readonly prompt: PromptID }
  | { readonly type: "uploading"; readonly prompt: PromptID }
  | {
      readonly type: "saved";
      readonly prompt: PromptID;
      readonly introductionId: string;
    }
  | {
      readonly type: "failed";
      readonly prompt: PromptID;
      readonly message: string;
    }
  | { readonly type: "rerecord"; readonly prompt: PromptID }
  | {
      readonly type: "hydrate";
      readonly saved: Partial<Record<PromptID, string>>;
    };

function set(
  state: VoiceState,
  prompt: PromptID,
  next: Partial<PromptState>,
): VoiceState {
  return {
    ...state,
    prompts: {
      ...state.prompts,
      [prompt]: { ...state.prompts[prompt], ...next },
    },
  };
}

export function voiceReducer(
  state: VoiceState,
  action: VoiceAction,
): VoiceState {
  switch (action.type) {
    case "start":
      // Refused while another prompt is live rather than silently switching:
      // the member would lose the take they were part-way through.
      if (state.active !== null) return state;
      return {
        ...set(state, action.prompt, {
          stage: "recording",
          seconds: 0,
          error: null,
        }),
        active: action.prompt,
      };
    case "tick": {
      const prompt = state.prompts[action.prompt];
      if (prompt.stage !== "recording") return state;
      const seconds = prompt.seconds + 1;
      // The meter stops the take at the bound. Letting it run past would
      // record audio the API will refuse, which costs the member the answer.
      return seconds >= maxPromptSeconds
        ? {
            ...set(state, action.prompt, { stage: "recorded", seconds }),
            active: null,
          }
        : set(state, action.prompt, { seconds });
    }
    case "stop":
      if (state.prompts[action.prompt].stage !== "recording") return state;
      return {
        ...set(state, action.prompt, { stage: "recorded" }),
        active: null,
      };
    case "uploading":
      return set(state, action.prompt, { stage: "uploading", error: null });
    case "saved":
      return set(state, action.prompt, {
        stage: "saved",
        introductionId: action.introductionId,
        error: null,
      });
    case "failed":
      // Back to "recorded", not "idle": the take is still in the browser and
      // the member should be able to retry the upload without saying it again.
      return {
        ...set(state, action.prompt, {
          stage: "recorded",
          error: action.message,
        }),
        active: null,
      };
    case "rerecord":
      if (state.active !== null) return state;
      return set(state, action.prompt, blankPrompt);
    case "hydrate": {
      let next = state;
      for (const prompt of voicePrompts) {
        const introductionId = action.saved[prompt.id];
        if (introductionId) {
          next = set(next, prompt.id, { stage: "saved", introductionId });
        }
      }
      return next;
    }
  }
}

export function completedCount(state: VoiceState): number {
  return voicePrompts.filter(
    (prompt) => state.prompts[prompt.id].stage === "saved",
  ).length;
}

/** Formats the meter as m:ss, counting up toward the bound. */
export function formatMeter(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `${minutes}:${rest.toString().padStart(2, "0")}`;
}
