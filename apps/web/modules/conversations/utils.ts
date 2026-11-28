import type { ChipTone } from "@/components/primitives";
import type { Translate } from "@/lib/i18n/translate";
import type { ChannelConnection, ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail, MessagingSession, MessagingTurn } from "./services/messaging";
import { checkoutKey, type ConsultDraft } from "./services/consultant";
import { needsHumanReasons, sessionOutcomes } from "./types";

export function isOneOf<T extends string>(values: readonly T[], value: string): value is T {
  return (values as readonly string[]).includes(value);
}

export const formatVnd = (value: string | number) =>
  new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(Number(value));

/** Label and tone for a turn's handling state, from the viewer's perspective. */
export function turnState(
  t: Translate,
  turn: MessagingTurn | null | undefined,
  userID: string | undefined,
): { label: string; tone: ChipTone } {
  if (turn?.lifecycle === "closed" || turn?.handling_state === "CLOSED") {
    return { label: t("desk.states.CLOSED"), tone: "neutral" };
  }
  switch (turn?.handling_state) {
    case "BOT_ACTIVE":
      return { label: t("desk.states.BOT_ACTIVE"), tone: "neutral" };
    case "WAITING_STAFF":
      return { label: t("desk.states.WAITING_STAFF"), tone: "warning" };
    case "STAFF_ACTIVE":
      return turn.assignee_id && turn.assignee_id === userID
        ? { label: t("desk.states.STAFF_ACTIVE_ME"), tone: "solid" }
        : { label: t("desk.states.STAFF_ACTIVE"), tone: "neutral" };
    default:
      return { label: t("desk.states.UNKNOWN"), tone: "warning" };
  }
}

/** Localized needs-human reason; a flag raised without a handoff reason reads as a new customer message. */
export function needsHumanReason(t: Translate, turn: MessagingTurn): string {
  const code = turn.needs_human_reason ?? "";
  return isOneOf(needsHumanReasons, code) ? t(`desk.reasons.${code}`) : t("desk.needsHumanPanel.defaultReason");
}

/** Real channel of a conversation, or null for simulated ones. */
export function threadPlatform(detail: MessagingDetail | null): ChannelPlatform | null {
  const source = detail?.conversation.source;
  return source === "facebook" || source === "telegram" ? source : null;
}

export interface ThreadAccess {
  assigned: boolean;
  canReply: boolean;
  canClaim: boolean;
  canHandoff: boolean;
  canReturn: boolean;
  /** Only the assigned sale can resolve needs-human; the API enforces it. */
  canResolve: boolean;
  /** Why the composer is disabled, or "" when the viewer can reply. */
  readOnlyReason: string;
}

/**
 * What the viewer may do in a thread. UI gating only: the API authorizes every command and send.
 * `connections` is null while unknown, which never locks the composer.
 */
export function threadAccess(args: {
  t: Translate;
  detail: MessagingDetail | null;
  connections: ChannelConnection[] | null;
  userID: string | undefined;
  owner: boolean;
  channelName: string;
  assigneeName: string;
  /** The tenant's reply mode is off: the API refuses a return to the bot, so the action is hidden. */
  botOff?: boolean;
}): ThreadAccess {
  const { t, detail, connections, userID, owner, channelName, assigneeName, botOff = false } = args;
  const turn = detail?.conversation.turn;
  const platform = threadPlatform(detail);
  const assigned = !!userID && turn?.assignee_id === userID;
  // A known connection that is not active (paused, auth_expired, or disconnected and so absent) locks the composer.
  const connectionID = detail?.conversation.thread.connection_id;
  const connection = connections?.find((c) => c.id === connectionID);
  const locked = !!platform && !!connectionID && connections !== null && connection?.status !== "active";
  const lockReason = !connection
    ? t("desk.readOnly.disconnected", { channel: channelName })
    : connection.status === "paused"
      ? t("desk.readOnly.paused", { channel: channelName })
      : platform === "telegram"
        ? t("desk.readOnly.expiredTelegram")
        : t("desk.readOnly.expiredFacebook");
  const open = turn?.lifecycle === "open";
  const canReply = !locked && open && assigned && turn?.control === "sales";
  const readOnlyReason = locked
    ? lockReason
    : canReply || !turn
      ? ""
      : !open
        ? t("desk.readOnly.closed")
        : turn.assignee_id && !assigned
          ? t("desk.readOnly.otherStaff", { name: assigneeName })
          : turn.handling_state === "BOT_ACTIVE"
            ? t("desk.readOnly.bot")
            : t("desk.readOnly.waiting");
  return {
    assigned,
    canReply,
    readOnlyReason,
    canClaim: open && (!turn.assignee_id || assigned),
    canHandoff: open && turn.handling_state !== "WAITING_STAFF" && (assigned || owner || turn.handling_state === "BOT_ACTIVE"),
    canReturn: open && !botOff && (assigned || owner) && turn.control === "sales",
    canResolve: open && assigned && turn.handling_state === "STAFF_ACTIVE" && turn.needs_human,
  };
}

export interface ProposalAccess {
  /** The viewer may send the proposal (UI gating only; the API authorizes the send). */
  canSend: boolean;
  /** Sending from bot handling claims the conversation for the viewer. */
  claimsOnSend: boolean;
  canDiscard: boolean;
  canRegenerate: boolean;
}

/** What the viewer may do with a bot proposal, mirroring the API rules. A new proposal is refused while the bot is off. */
export function proposalAccess(
  turn: MessagingTurn | null | undefined,
  userID: string | undefined,
  owner: boolean,
  botOff = false,
): ProposalAccess {
  const open = turn?.lifecycle === "open";
  const assigned = !!userID && turn?.assignee_id === userID;
  const bot = open && turn?.handling_state === "BOT_ACTIVE";
  const staffMine = open && turn?.handling_state === "STAFF_ACTIVE" && assigned;
  return {
    canSend: bot || staffMine,
    claimsOnSend: bot,
    canDiscard: open && (owner || assigned || !turn?.assignee_id),
    canRegenerate: !botOff && (bot || staffMine),
  };
}

/** Hands a draft to the checkout page. Throws when session storage is unavailable. */
export function saveCheckoutDraft(tenantID: string, draft: ConsultDraft) {
  sessionStorage.setItem(checkoutKey(tenantID), JSON.stringify(draft));
}

/** A single-item draft for buying one variant directly. */
export function singleItemDraft(item: {
  product_id: string;
  variant_id: string;
  product_name: string;
  price: string;
}): ConsultDraft {
  return {
    draft_id: crypto.randomUUID(),
    total_amount: item.price,
    items: [
      {
        product_id: item.product_id,
        variant_id: item.variant_id,
        product_name: item.product_name,
        quantity: 1,
        unit_price: item.price,
      },
    ],
  };
}

/** What the viewer may do with a session. UI gating only: the API authorizes close and outcome edits. */
export function sessionAccess(turn: MessagingTurn, userID: string | undefined, owner: boolean) {
  const open = turn.lifecycle === "open";
  const assignedStaff = !!userID && turn.assignee_id === userID && turn.handling_state === "STAFF_ACTIVE";
  return {
    canClose: open && (owner || assignedStaff),
    canEditOutcome: !open && (owner || (!!userID && turn.last_handler_id === userID)),
  };
}

/** "start – end" of a session (end is "now" while it is open). */
export function sessionRange(t: Translate, session: Pick<MessagingSession, "opened_at" | "turn">, locale?: string): string {
  const format = (value: string) => new Date(value).toLocaleString(locale);
  const start = session.opened_at ? format(session.opened_at) : "";
  return session.turn.closed_at
    ? t("desk.sessions.range", { start, end: format(session.turn.closed_at) })
    : t("desk.sessions.ongoing", { start });
}

/** Localized outcome of a session, or the "no outcome" text. */
export function sessionOutcome(t: Translate, turn: MessagingTurn): string {
  const code = turn.outcome ?? "";
  return isOneOf(sessionOutcomes, code) ? t(`desk.sessions.outcomes.${code}`) : t("desk.sessions.noOutcome");
}

/** Localized close cause of a closed session. */
export function sessionCause(t: Translate, turn: MessagingTurn): string {
  return turn.close_cause === "closed_manually" ? t("desk.sessions.cause.closed_manually") : t("desk.sessions.cause.expired");
}

/** Shows only the edges of a channel customer key; the full value is for the system, not the screen. */
export function maskExternalKey(key: string): string {
  const colon = key.indexOf(":");
  const prefix = colon >= 0 ? key.slice(0, colon + 1) : "";
  const value = colon >= 0 ? key.slice(colon + 1) : key;
  if (value.length <= 8) return `${prefix}${"•".repeat(value.length)}`;
  return `${prefix}${value.slice(0, 4)}••••••${value.slice(-2)}`;
}

/** First letter of a display name for avatars. */
export const initialOf = (name: string) => (name.trim().charAt(0) || "?").toUpperCase();
