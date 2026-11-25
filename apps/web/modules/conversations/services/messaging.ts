import { sdk } from "@/lib/sdk";
import type { SessionOutcome } from "../types";

export interface MessagingTurn {
  id: string;
  thread_id: string;
  lifecycle: "open" | "closed";
  control: "system" | "sales";
  handling_state?: "BOT_ACTIVE" | "WAITING_STAFF" | "STAFF_ACTIVE" | "UNKNOWN";
  assignee_id: string | null;
  needs_human: boolean;
  /** Handoff reason code behind needs_human; absent when none was given. */
  needs_human_reason?: string;
  context_revision: number;
  control_generation: number;
  version: number;
  last_activity_at: string;
  /** Set once the turn (session) closed: after 24 hours of inactivity or manually. */
  closed_at?: string;
  /** How a closed session ended. */
  close_cause?: "expired" | "closed_manually";
  /** Member who closed the session manually. */
  closed_by?: string;
  /** Session outcome; an expired session has none until the owner or its last handler records one. */
  outcome?: SessionOutcome;
  note?: string;
  /** Last staff member assigned to the session; kept after return to the bot. */
  last_handler_id?: string;
  /**
   * The bot handles the turn and its proposal waits for staff review: the conversation counts as waiting for staff
   * (and is listed under the waiting filter) and does not expire until the proposal is sent, discarded or superseded.
   */
  review_pending?: boolean;
}
export interface Conversation {
  thread: {
    id: string;
    display_name: string;
    external_key: string;
    // Bound channel connection of a real conversation; absent for simulator threads.
    connection_id?: string;
    created_at: string;
    updated_at: string;
  };
  turn: MessagingTurn | null;
  source: "simulator" | "facebook" | "telegram";
}
export interface MessagingMessage {
  id: string;
  thread_id: string;
  turn_id: string;
  direction: "inbound" | "outbound";
  actor_id: string;
  body: string;
  // channel: a real provider reply whose progress is in delivery_status.
  status: "simulated" | "received" | "channel";
  created_at: string;
  content_type: "text" | "unsupported_media";
  placeholder_meta: { media_kinds: string[] } | null;
  delivery_status: "received" | "sending" | "sent" | "failed" | "unknown";
  delivery_error_reason: string;
}
/** A knowledge record an AI-written answer relies on; label is empty when the source was removed. */
export interface ProposalCitation {
  source_id: string;
  kind: "product" | "document" | "script" | "fact";
  label?: string;
}
/** A bot reply proposal. `stale` means conversation control changed after it was created: it can never be sent. */
export interface ReplyProposal {
  id: string;
  turn_id: string;
  /** routing: the bot is still preparing it (no text yet); pending: waiting for a sale. */
  status: "routing" | "pending" | "sent" | "auto_sent" | "discarded" | "superseded" | "handed_off" | "no_candidate";
  route?: "script" | "answer" | "clarify" | "handoff";
  source?: "script" | "ai";
  script_version_id?: string;
  script_name?: string;
  script_version?: number;
  ai_run_id?: string;
  text?: string;
  reason?: string;
  control_generation: number;
  stale: boolean;
  requested_by?: string;
  /** Knowledge sources of an AI-written answer (route answer, source ai), which always waits for staff review. */
  citations?: ProposalCitation[];
  created_at: string;
}
export interface MessagingDetail {
  conversation: Conversation;
  messages: MessagingMessage[];
  next_cursor: string;
  /** Live bot proposals of the latest turn (at most one). */
  proposals?: ReplyProposal[];
}
/** One conversation session (a consultation turn) with its message count and orders created from Desk. */
export interface MessagingSession {
  turn: MessagingTurn;
  opened_at: string;
  message_count: number;
  /** Where the session starts in the thread. */
  first_message_id: string;
  orders: { id: string; order_number: string; status: string }[];
}
export interface MessagingSessionPage {
  items: MessagingSession[];
  next_cursor: string;
}
export interface SessionInput {
  outcome: SessionOutcome;
  note: string;
}
export interface MessagingPage {
  items: Conversation[];
  next_cursor: string;
}
export type InboxFilter = "all" | "unassigned" | "mine" | "needs_human" | "waiting" | "bot";
/** Tenant-wide Desk notification state. */
export interface MessagingSummary {
  waiting_staff_count: number;
}
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
export interface ChannelSendInput {
  client_message_id: string;
  text: string;
  expected_version: number;
  context_revision: number;
}
/** Actor recorded on messages the bot sent itself (auto_send). */
export const botActorID = "system:bot";
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
  async summary(tenant: string, signal?: AbortSignal): Promise<MessagingSummary> {
    const result = await sdk.client.get<{ data: MessagingSummary }>(
      `${base(tenant)}/conversations/summary`,
      { signal },
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
    action: "claim" | "assign" | "handoff" | "return-to-bot" | "resolve",
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
  /** Sessions of a conversation, newest first. */
  async sessions(tenant: string, thread: string, cursor = "", signal?: AbortSignal): Promise<MessagingSessionPage> {
    const result = await sdk.client.get<{ data: MessagingSessionPage }>(
      `${base(tenant)}/conversations/${thread}/sessions`,
      { params: { cursor, limit: 20 }, signal },
    );
    return result.data.data;
  },
  /** Closes the open session now (assigned sale under staff control, or the owner). */
  async closeSession(tenant: string, turn: MessagingTurn, input: SessionInput): Promise<MessagingTurn> {
    const result = await sdk.client.post<{ data: MessagingTurn }>(`${base(tenant)}/turns/${turn.id}/close`, {
      expected_version: turn.version,
      ...input,
    });
    return result.data.data;
  },
  /** Records or edits a closed session's outcome and note (owner or the session's last handler). */
  async updateOutcome(tenant: string, turn: MessagingTurn, input: SessionInput): Promise<MessagingTurn> {
    const result = await sdk.client.post<{ data: MessagingTurn }>(`${base(tenant)}/turns/${turn.id}/outcome`, {
      expected_version: turn.version,
      ...input,
    });
    return result.data.data;
  },
  // Real Facebook reply; retry with the same client_message_id to replay instead of duplicating.
  async sendChannel(
    tenant: string,
    turn: string,
    input: ChannelSendInput,
  ): Promise<MessagingMessage> {
    const result = await sdk.client.post<{ data: MessagingMessage }>(
      `${base(tenant)}/turns/${turn}/messages`,
      input,
    );
    return result.data.data;
  },
  /** Sends a proposal as-is (no text) or edited; retry with the same client_message_id to replay. */
  async sendProposal(
    tenant: string,
    turn: string,
    proposal: string,
    input: { client_message_id: string; text?: string },
  ): Promise<MessagingMessage> {
    const result = await sdk.client.post<{ data: MessagingMessage }>(
      `${base(tenant)}/turns/${turn}/proposals/${proposal}/send`,
      input,
    );
    return result.data.data;
  },
  async discardProposal(tenant: string, turn: string, proposal: string): Promise<ReplyProposal> {
    const result = await sdk.client.post<{ data: ReplyProposal }>(
      `${base(tenant)}/turns/${turn}/proposals/${proposal}/discard`,
    );
    return result.data.data;
  },
  /** Asks the bot for a new proposal at the current conversation control. */
  async regenerateProposal(tenant: string, turn: string): Promise<ReplyProposal> {
    const result = await sdk.client.post<{ data: ReplyProposal }>(
      `${base(tenant)}/turns/${turn}/regenerate-proposal`,
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
