export {
  createObiaraClient,
  type ObiaraClient,
  type ObiaraClientOptions,
} from "./client";
export type { components, operations, paths } from "./generated/schema";

export type ApiErrorEnvelope =
  import("./generated/schema").components["schemas"]["ErrorEnvelope"];
export type ApiFieldError =
  import("./generated/schema").components["schemas"]["FieldError"];
