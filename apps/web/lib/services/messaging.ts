import { sdk } from "@/lib/sdk";

export interface MessagingTurn {
  id: string;
  thread_id: string;
  lifecycle: "open" | "closed";
  control: "system" | "sales";
  handling_state?: "BOT_ACTIVE" | "WAITING_STAFF" | "STAFF_ACTIVE" | "UNKNOWN";
  assignee_id: string | null;
  needs_human: boolean;
  context_revision: number;
  control_generation: number;
  version: number;
  last_activity_at: string;
}
export interface Conversation {
  thread: {
    id: string;
    display_name: string;
    external_key: string;
    created_at: string;
    updated_at: string;
  };
  turn: MessagingTurn | null;
  source: "simulator";
}
export interface MessagingMessage {
  id: string;
  thread_id: string;
  turn_id: string;
  direction: "inbound" | "outbound";
  actor_id: string;
  body: string;
  status: "simulated";
  created_at: string;
}
export interface MessagingDetail {
  conversation: Conversation;
  messages: MessagingMessage[];
  next_cursor: string;
}
export interface MessagingPage {
  items: Conversation[];
  next_cursor: string;
}
export type InboxFilter = "all" | "unassigned" | "mine" | "needs_human";
export interface InboundInput {
  external_key: string;
  display_name: string;
  body: string;
  request_key: string;
}
export interface SendInput {
  expected_version: number;
  context_revision: number;
  body: string;
  request_key: string;
}
const base = (tenant: string) => `/api/v1/tenants/${tenant}`;
export const messaging = {
  async list(
    tenant: string,
    filter: InboxFilter,
    cursor = "",
    signal?: AbortSignal,
  ): Promise<MessagingPage> {
    const result = await sdk.client.get<{ data: MessagingPage }>(
      `${base(tenant)}/conversations`,
      { params: { filter, cursor, limit: 20 }, signal },
    );
    return result.data.data;
  },
  async detail(
    tenant: string,
    thread: string,
    cursor = "",
    signal?: AbortSignal,
  ): Promise<MessagingDetail> {
    const result = await sdk.client.get<{ data: MessagingDetail }>(
      `${base(tenant)}/conversations/${thread}/messages`,
      { params: { cursor, limit: 50 }, signal },
    );
    return result.data.data;
  },
  async inbound(
    tenant: string,
    input: InboundInput,
  ): Promise<MessagingMessage> {
    const result = await sdk.client.post<{ data: MessagingMessage }>(
      `${base(tenant)}/simulator/messages`,
      input,
    );
    return result.data.data;
  },
  async command(
    tenant: string,
    turn: MessagingTurn,
    action: "claim" | "assign" | "handoff" | "return-to-bot",
    assignee?: string | null,
    reason?: string,
  ): Promise<MessagingTurn> {
    const result = await sdk.client.post<{ data: MessagingTurn }>(
      `${base(tenant)}/turns/${turn.id}/${action}`,
      {
        expected_version: turn.version,
        ...(action === "assign" ? { assignee_id: assignee ?? null } : {}),
        ...(action === "handoff" ? { reason } : {}),
      },
    );
    return result.data.data;
  },
  async send(
    tenant: string,
    turn: string,
    input: SendInput,
  ): Promise<MessagingMessage> {
    const result = await sdk.client.post<{ data: MessagingMessage }>(
      `${base(tenant)}/turns/${turn}/send-simulation`,
      input,
    );
    return result.data.data;
  },
};
