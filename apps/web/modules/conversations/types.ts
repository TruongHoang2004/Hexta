import type { InboxFilter } from "./services/messaging";
import type { ChatMessage, ConsultDraft, ConsultProduct } from "./services/consultant";

/** Reasons staff can give when handing a conversation to the waiting queue. */
export const handoffReasons = [
  "customer_requested_staff",
  "insufficient_information",
  "ai_error",
  "business_review",
  "staff_handoff",
] as const;
export type HandoffReason = (typeof handoffReasons)[number];

/** Every needs-human reason with a localized label: staff reasons plus bot_off (a customer wrote while the bot was off). */
export const needsHumanReasons = [...handoffReasons, "bot_off"] as const;

/** Delivery failure codes that have a localized explanation. */
export const deliveryReasons = [
  "window_24h",
  "auth_expired",
  "rate_limit",
  "invalid_recipient",
  "bot_blocked",
  "chat_not_found",
  "transport",
  "expired",
  "connection_inactive",
  "connection_paused",
  "connection_disconnected",
  "cancelled",
] as const;
export type DeliveryReason = (typeof deliveryReasons)[number];

/** Attachment kinds that have a localized name. */
export const mediaKinds = ["image", "video", "audio", "file", "sticker"] as const;
export type MediaKind = (typeof mediaKinds)[number];

/** Inbox list filters shown on the desk. */
export type DeskFilter = Extract<InboxFilter, "bot" | "waiting" | "mine" | "all">;

export type ThreadCommand = "claim" | "assign" | "handoff" | "return-to-bot" | "resolve";

/** One message in the consultant chat, with the structured parts streamed alongside it. */
export type ConsultTurn = ChatMessage & {
  products?: ConsultProduct[];
  draft?: ConsultDraft;
  replies?: string[];
  needsHuman?: boolean;
};
