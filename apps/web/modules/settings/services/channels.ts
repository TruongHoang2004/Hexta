import { sdk } from "@/lib/sdk";
import type {
  ChannelConnection,
  ConnectionAction,
  FacebookConnection,
  OAuthPages,
  TelegramAction,
  TelegramConnection,
} from "../types";

// Meta redirects here; the path must match the API's allowed OAuth callback.
export const facebookCallbackPath = "/channels/facebook/callback";
const base = (tenant: string) => `/api/v1/tenants/${tenant}/channels/facebook`;

export const channels = {
  async startFacebook(tenant: string): Promise<string> {
    const result = await sdk.client.post<{ data: { auth_url: string } }>(`${base(tenant)}/oauth/start`, {
      redirect_uri: `${window.location.origin}${facebookCallbackPath}`,
    });
    return result.data.data.auth_url;
  },
  async completeFacebook(params: { code: string; state: string; error: string }): Promise<OAuthPages> {
    const result = await sdk.client.get<{ data: OAuthPages }>("/api/v1/channels/facebook/oauth/callback", { params });
    return result.data.data;
  },
  async list(tenant: string, signal?: AbortSignal): Promise<FacebookConnection[]> {
    const result = await sdk.client.get<{ data: { connections: FacebookConnection[] } }>(
      `${base(tenant)}/connections`,
      { signal },
    );
    return result.data.data.connections;
  },
  async connect(tenant: string, session: string, pageIDs: string[]): Promise<FacebookConnection[]> {
    const result = await sdk.client.post<{ data: { connections: FacebookConnection[] } }>(
      `${base(tenant)}/connections`,
      { oauth_session_id: session, page_ids: pageIDs },
    );
    return result.data.data.connections;
  },
  async update(tenant: string, id: string, action: ConnectionAction): Promise<FacebookConnection> {
    const result = await sdk.client.patch<{ data: FacebookConnection }>(`${base(tenant)}/connections/${id}`, {
      action,
    });
    return result.data.data;
  },
};

const telegramBase = (tenant: string) => `/api/v1/tenants/${tenant}/channels/telegram/connections`;

// The Bot Token is sent once over the API and never returned; callers must clear it from state after submit.
export const telegram = {
  async list(tenant: string, signal?: AbortSignal): Promise<TelegramConnection[]> {
    const result = await sdk.client.get<{ data: { connections: TelegramConnection[] } }>(telegramBase(tenant), {
      signal,
    });
    return result.data.data.connections;
  },
  async connect(tenant: string, botToken: string): Promise<TelegramConnection> {
    const result = await sdk.client.post<{ data: TelegramConnection }>(telegramBase(tenant), { bot_token: botToken });
    return result.data.data;
  },
  async update(tenant: string, id: string, action: TelegramAction, botToken?: string): Promise<TelegramConnection> {
    const result = await sdk.client.patch<{ data: TelegramConnection }>(
      `${telegramBase(tenant)}/${id}`,
      action === "rotate_token" ? { action, bot_token: botToken } : { action },
    );
    return result.data.data;
  },
};

// Lists every channel connection for banners and composer locks. A platform that fails to load is omitted.
export async function listAllConnections(tenant: string, signal?: AbortSignal): Promise<ChannelConnection[] | null> {
  const results = await Promise.allSettled([channels.list(tenant, signal), telegram.list(tenant, signal)]);
  if (results.every((r) => r.status === "rejected")) return null;
  return results.flatMap((r): ChannelConnection[] => (r.status === "fulfilled" ? r.value : []));
}
