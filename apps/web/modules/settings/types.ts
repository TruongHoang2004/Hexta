export type ConnectionStatus = "active" | "paused" | "disconnected" | "auth_expired";
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
  // Owner actions allowed in the current state, decided by the API per connection type.
  actions: ConnectionAction[];
  last_error_code?: string;
  last_error_message?: string;
  updated_at: string;
}
export interface SelectablePage {
  page_id: string;
  page_name: string;
  availability: "available" | "already_connected_this_tenant" | "owned_by_other_tenant";
}
export interface OAuthPages {
  tenant_id: string;
  oauth_session_id: string;
  expires_at: string;
  pages: SelectablePage[];
}
export type ConnectionAction = "pause" | "resume" | "disconnect" | "rotate_token";
/** Actions confirmed before they run; rotate_token has its own form instead. */
export type ConfirmedAction = Exclude<ConnectionAction, "rotate_token">;
