import { sdk } from "@/lib/sdk";
import type { ChannelConnection, ChannelPlatform, ConnectionAction, OAuthPages } from "../types";

// Meta redirects here; the path must match the API's allowed OAuth callback.
export const facebookCallbackPath = "/channels/facebook/callback";
const tenantBase = (tenant: string) => `/api/v1/tenants/${tenant}`;

/** Every connection of the workspace, whatever its type, is listed and managed through one resource. */
export const connections = {
  async list(tenant: string, signal?: AbortSignal, platform?: ChannelPlatform): Promise<ChannelConnection[]> {
    const result = await sdk.client.get<{ data: { connections: ChannelConnection[] } }>(
      `${tenantBase(tenant)}/connections`,
      { signal, params: platform ? { platform } : undefined },
    );
    return result.data.data.connections;
  },
  // The Bot Token is sent once for rotate_token and never returned; callers clear it from state before calling.
  async update(tenant: string, id: string, action: ConnectionAction, botToken?: string): Promise<ChannelConnection> {
    const result = await sdk.client.patch<{ data: ChannelConnection }>(
      `${tenantBase(tenant)}/connections/${id}`,
      action === "rotate_token" ? { action, bot_token: botToken } : { action },
    );
    return result.data.data;
  },
};

/** Creating a connection differs per type: Facebook uses Meta OAuth and Page selection. */
export const facebook = {
  async start(tenant: string): Promise<string> {
    const result = await sdk.client.post<{ data: { auth_url: string } }>(
      `${tenantBase(tenant)}/channels/facebook/oauth/start`,
      { redirect_uri: `${window.location.origin}${facebookCallbackPath}` },
    );
    return result.data.data.auth_url;
  },
  async complete(params: { code: string; state: string; error: string }): Promise<OAuthPages> {
    const result = await sdk.client.get<{ data: OAuthPages }>("/api/v1/channels/facebook/oauth/callback", { params });
    return result.data.data;
  },
  async connectPages(tenant: string, session: string, pageIDs: string[]): Promise<ChannelConnection[]> {
    const result = await sdk.client.post<{ data: { connections: ChannelConnection[] } }>(
      `${tenantBase(tenant)}/channels/facebook/connections`,
      { oauth_session_id: session, page_ids: pageIDs },
    );
    return result.data.data.connections;
  },
};

/** Telegram connects with a @BotFather Bot Token, sent once and never returned. */
export const telegram = {
  async connect(tenant: string, botToken: string): Promise<ChannelConnection> {
    const result = await sdk.client.post<{ data: ChannelConnection }>(
      `${tenantBase(tenant)}/channels/telegram/connections`,
      { bot_token: botToken },
    );
    return result.data.data;
  },
};

// Lists every channel connection for banners and composer locks; null when the list cannot be loaded.
export async function listAllConnections(tenant: string, signal?: AbortSignal): Promise<ChannelConnection[] | null> {
  try {
    return await connections.list(tenant, signal);
  } catch {
    return null;
  }
}
