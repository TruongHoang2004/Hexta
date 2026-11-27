import type { ChipTone } from "@/components/primitives";
import type { Translate } from "@/lib/i18n/translate";
import type { ChannelConnection, ChannelPlatform, ConnectionStatus } from "./types";

export const platformLabel: Record<ChannelPlatform, string> = {
  facebook: "Facebook",
  telegram: "Telegram",
};

/** Connection types in the order the "add connection" picker and filters show them. */
export const channelPlatforms: ChannelPlatform[] = ["facebook", "telegram"];

export const connectionTone: Record<ConnectionStatus, ChipTone> = {
  active: "success",
  paused: "warning",
  auth_expired: "danger",
  disconnected: "neutral",
};

// Mirrors the API check: numeric bot ID, colon, secret from @BotFather.
export const telegramTokenPattern = /^[0-9]{1,20}:[A-Za-z0-9_-]{30,80}$/;

/** Display name of any connection: @handle when the platform has one, else account name, else ID. */
export const connectionName = (row: ChannelConnection) =>
  row.account_username ? `@${row.account_username}` : row.account_name || row.account_id;

const roles = ["owner", "admin", "member"] as const;
const inviteStatuses = ["pending", "accepted", "declined", "revoked", "expired"] as const;
type InviteStatus = (typeof inviteStatuses)[number];

const inviteTone: Record<InviteStatus, ChipTone> = {
  pending: "warning",
  accepted: "success",
  declined: "neutral",
  revoked: "neutral",
  expired: "neutral",
};

export function roleLabel(t: Translate, role: string) {
  return (roles as readonly string[]).includes(role) ? t(`members.roles.${role as (typeof roles)[number]}`) : role;
}

export function inviteChip(t: Translate, status: string): { label: string; tone: ChipTone } {
  const known = (inviteStatuses as readonly string[]).includes(status) ? (status as InviteStatus) : null;
  return known
    ? { label: t(`members.status.${known}`), tone: inviteTone[known] }
    : { label: t("members.status.other"), tone: "neutral" };
}
