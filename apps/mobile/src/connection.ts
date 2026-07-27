export type ConnectionMode = "online" | "constrained";

export interface ConnectionCopy {
  accessibilityLabel: string;
  body: string;
  queueBody: string;
  queueTitle: string;
  title: string;
}

export function buildConnectionCopy(
  mode: ConnectionMode,
  queued: boolean,
): ConnectionCopy {
  if (mode === "constrained") {
    return {
      accessibilityLabel: "Data saver is on. Switch to full connection mode.",
      body: "Voice stays light. Photos wait for Wi-Fi.",
      queueBody: queued
        ? "Saved safely on this device. It will send when your connection returns."
        : "Try the offline state without sending anything.",
      queueTitle: queued
        ? "Your voice reply is queued"
        : "Replies can wait with you",
      title: "Data saver · 3G ready",
    };
  }

  return {
    accessibilityLabel: "Full connection mode is on. Switch to data saver.",
    body: "Everything is ready. Tap to preview data saver.",
    queueBody: queued
      ? "The connection is back. Your queued reply is ready to send."
      : "If the network drops, replies stay safely on this device.",
    queueTitle: queued ? "Ready to send your reply" : "Connection is strong",
    title: "You’re online",
  };
}
