import { sdk } from "@/lib/sdk";
import { readSseFrames } from "./sse-frames";
import type { MessagingMessage } from "./messaging";

/** What the server pushes: stored messages whole, delivery state changes, and list/resync triggers. */
export type StreamEvent =
  | { type: "ready" }
  | { type: "message"; threadID: string; message: MessagingMessage }
  | {
      type: "delivery";
      threadID: string;
      messageID: string;
      status: MessagingMessage["delivery_status"];
      reason: string;
    }
  | { type: "changed"; threadID: string }
  | { type: "resync" };

/**
 * Opens the tenant conversation stream with the SDK's fetch transport (Bearer auth and token refresh included)
 * and resolves when the server ends it. The database stays authoritative: callers dedupe by message id and
 * refetch on `resync` or after a reconnect.
 */
export async function streamConversations(
  tenant: string,
  signal: AbortSignal,
  onEvent: (event: StreamEvent) => void,
) {
  const response = await sdk.client.get<ReadableStream<Uint8Array>>(
    `/api/v1/tenants/${tenant}/conversations/stream`,
    { adapter: "fetch", responseType: "stream", signal },
  );
  if (!String(response.headers["content-type"]).includes("text/event-stream")) {
    throw new Error("Unexpected conversation stream response");
  }
  await readSseFrames(response.data, ({ event, data }) => {
    const body = data ? (JSON.parse(data) as Record<string, unknown>) : {};
    if (event === "ready") onEvent({ type: "ready" });
    else if (event === "resync") onEvent({ type: "resync" });
    else if (event === "message" && body.message)
      onEvent({
        type: "message",
        threadID: String(body.thread_id),
        message: body.message as MessagingMessage,
      });
    else if (event === "delivery")
      onEvent({
        type: "delivery",
        threadID: String(body.thread_id),
        messageID: String(body.message_id),
        status: body.delivery_status as MessagingMessage["delivery_status"],
        reason: String(body.delivery_error_reason ?? ""),
      });
    else if (event === "changed" && body.thread_id)
      onEvent({ type: "changed", threadID: String(body.thread_id) });
  });
}
