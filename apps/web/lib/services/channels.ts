import { sdk } from "@/lib/sdk";

export type ConnectionStatus =
  | "active"
  | "paused"
  | "disconnected"
  | "auth_expired";
export type ChannelPlatform = "facebook" | "telegram";
// Bot Tokens and Page tokens never appear in these values; the API returns identities and status only.
export interface ChannelConnection {
  id: string;
  tenant_id: string;
  platform: ChannelPlatform;
  account_id: string;
  account_name: string;
  // Telegram @bot_username; absent for Facebook.
  account_username?: string;
  version: number;
  status: ConnectionStatus;
  last_error_code?: string;
  last_error_message?: string;
  updated_at: string;
}
export type FacebookConnection = ChannelConnection & { platform: "facebook" };
export type TelegramConnection = ChannelConnection & { platform: "telegram" };
export const platformLabel: Record<ChannelPlatform, string> = {
  facebook: "Facebook",
  telegram: "Telegram",
};
export interface SelectablePage {
  page_id: string;
  page_name: string;
  availability:
    | "available"
    | "already_connected_this_tenant"
    | "owned_by_other_tenant";
}
export interface OAuthPages {
  tenant_id: string;
  oauth_session_id: string;
  expires_at: string;
  pages: SelectablePage[];
}
export type ConnectionAction = "pause" | "resume" | "disconnect";

// Meta redirects here; the path must match the API's allowed OAuth callback.
export const facebookCallbackPath = "/channels/facebook/callback";
const base = (tenant: string) =>
  `/api/v1/tenants/${tenant}/channels/facebook`;

export const channels = {
  async startFacebook(tenant: string): Promise<string> {
    const result = await sdk.client.post<{ data: { auth_url: string } }>(
      `${base(tenant)}/oauth/start`,
      { redirect_uri: `${window.location.origin}${facebookCallbackPath}` },
    );
    return result.data.data.auth_url;
  },
  async completeFacebook(params: {
    code: string;
    state: string;
    error: string;
  }): Promise<OAuthPages> {
    const result = await sdk.client.get<{ data: OAuthPages }>(
      "/api/v1/channels/facebook/oauth/callback",
      { params },
    );
    return result.data.data;
  },
  async list(
    tenant: string,
    signal?: AbortSignal,
  ): Promise<FacebookConnection[]> {
    const result = await sdk.client.get<{
      data: { connections: FacebookConnection[] };
    }>(`${base(tenant)}/connections`, { signal });
    return result.data.data.connections;
  },
  async connect(
    tenant: string,
    session: string,
    pageIDs: string[],
  ): Promise<FacebookConnection[]> {
    const result = await sdk.client.post<{
      data: { connections: FacebookConnection[] };
    }>(`${base(tenant)}/connections`, {
      oauth_session_id: session,
      page_ids: pageIDs,
    });
    return result.data.data.connections;
  },
  async update(
    tenant: string,
    id: string,
    action: ConnectionAction,
  ): Promise<FacebookConnection> {
    const result = await sdk.client.patch<{ data: FacebookConnection }>(
      `${base(tenant)}/connections/${id}`,
      { action },
    );
    return result.data.data;
  },
};

export type TelegramAction = ConnectionAction | "rotate_token";
const telegramBase = (tenant: string) =>
  `/api/v1/tenants/${tenant}/channels/telegram/connections`;

// The Bot Token is sent once over the API and never returned; callers must clear it from state after submit.
export const telegram = {
  async list(
    tenant: string,
    signal?: AbortSignal,
  ): Promise<TelegramConnection[]> {
    const result = await sdk.client.get<{
      data: { connections: TelegramConnection[] };
    }>(telegramBase(tenant), { signal });
    return result.data.data.connections;
  },
  async connect(tenant: string, botToken: string): Promise<TelegramConnection> {
    const result = await sdk.client.post<{ data: TelegramConnection }>(
      telegramBase(tenant),
      { bot_token: botToken },
    );
    return result.data.data;
  },
  async update(
    tenant: string,
    id: string,
    action: TelegramAction,
    botToken?: string,
  ): Promise<TelegramConnection> {
    const result = await sdk.client.patch<{ data: TelegramConnection }>(
      `${telegramBase(tenant)}/${id}`,
      action === "rotate_token" ? { action, bot_token: botToken } : { action },
    );
    return result.data.data;
  },
};

// Lists every channel connection for banners and composer locks. A platform that fails to load is omitted.
export async function listAllConnections(
  tenant: string,
  signal?: AbortSignal,
): Promise<ChannelConnection[] | null> {
  const results = await Promise.allSettled([
    channels.list(tenant, signal),
    telegram.list(tenant, signal),
  ]);
  if (results.every((r) => r.status === "rejected")) return null;
  return results.flatMap((r): ChannelConnection[] =>
    r.status === "fulfilled" ? r.value : [],
  );
}

// Delivery failure text. The platform name is inserted where the reason depends on the provider.
export function deliveryReasonText(
  reason: string,
  platform: ChannelPlatform,
): string {
  const name = platformLabel[platform];
  const account = platform === "telegram" ? "bot" : "Page";
  const text: Record<string, string> = {
    window_24h:
      "Not sent: Facebook only allows replies within 24 hours of the customer's last message.",
    auth_expired:
      platform === "telegram"
        ? "Not sent: Telegram rejected the Bot Token. An owner must rotate the token."
        : "Not sent: the Page authorization expired. An owner must reconnect it.",
    rate_limit: `Not sent: ${name} rate limit reached. Try again later.`,
    invalid_recipient: "Not sent: this customer can no longer be reached.",
    bot_blocked:
      "Not sent: the customer blocked the bot or deleted their Telegram account.",
    chat_not_found: "Not sent: Telegram could not find this chat.",
    transport: `Not sent: ${name} could not be reached.`,
    expired: "Not sent: the send request expired before delivery.",
    connection_inactive: `Not sent: the ${account} connection changed before delivery.`,
    connection_paused: `Cancelled: the ${account} connection was paused.`,
    connection_disconnected: `Cancelled: the ${account} was disconnected.`,
  };
  return text[reason] ?? `Not sent to ${name}.`;
}
