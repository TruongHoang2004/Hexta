import { sdk } from "@/lib/sdk";

export type ConnectionStatus =
  | "active"
  | "paused"
  | "disconnected"
  | "auth_expired";
export interface FacebookConnection {
  id: string;
  tenant_id: string;
  platform: "facebook";
  account_id: string;
  account_name: string;
  version: number;
  status: ConnectionStatus;
  last_error_code?: string;
  last_error_message?: string;
  updated_at: string;
}
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

export const deliveryReasonText: Record<string, string> = {
  window_24h:
    "Not sent: Facebook only allows replies within 24 hours of the customer's last message.",
  auth_expired:
    "Not sent: the Page authorization expired. An owner must reconnect it.",
  rate_limit: "Not sent: Facebook rate limit reached. Try again later.",
  invalid_recipient: "Not sent: this customer can no longer be reached.",
  transport: "Not sent: Facebook could not be reached.",
  expired: "Not sent: the send request expired before delivery.",
  connection_inactive: "Not sent: the Page connection changed before delivery.",
  connection_paused: "Cancelled: the Page connection was paused.",
  connection_disconnected: "Cancelled: the Page was disconnected.",
};
