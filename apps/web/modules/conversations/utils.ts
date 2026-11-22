import type { ChipTone } from "@/components/primitives";
import type { Translate } from "@/lib/i18n/translate";
import type { ChannelConnection, ChannelPlatform } from "@/modules/settings";
import type { MessagingDetail, MessagingTurn } from "./services/messaging";
import { checkoutKey, type ConsultDraft } from "./services/consultant";

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
}): ThreadAccess {
  const { t, detail, connections, userID, owner, channelName, assigneeName } = args;
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
  const canReply = !locked && assigned && turn?.control === "sales";
  const readOnlyReason = locked
    ? lockReason
    : canReply || !turn
      ? ""
      : turn.assignee_id && !assigned
        ? t("desk.readOnly.otherStaff", { name: assigneeName })
        : turn.handling_state === "BOT_ACTIVE"
          ? t("desk.readOnly.bot")
          : t("desk.readOnly.waiting");
  const open = turn?.lifecycle === "open";
  return {
    assigned,
    canReply,
    readOnlyReason,
    canClaim: open && (!turn.assignee_id || assigned),
    canHandoff: open && turn.handling_state !== "WAITING_STAFF" && (assigned || owner || turn.handling_state === "BOT_ACTIVE"),
    canReturn: open && (assigned || owner) && turn.control === "sales",
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
