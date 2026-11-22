import { sdk } from "@/lib/sdk";
import { readSseFrames } from "./sse-frames";

export interface ConversationHint {
  thread_id: string;
  reason: string;
}

export type StreamMessage =
  | { type: "ready" }
  | { type: "changed"; hint: ConversationHint }
  | { type: "resync" };

/**
 * Opens the tenant conversation stream with the SDK's fetch transport (Bearer auth and token refresh included)
 * and resolves when the server ends it. Hints carry no content: callers refetch through the read endpoints.
 */
export async function streamConversations(
  tenant: string,
  signal: AbortSignal,
  onMessage: (message: StreamMessage) => void,
) {
  const response = await sdk.client.get<ReadableStream<Uint8Array>>(
    `/api/v1/tenants/${tenant}/conversations/stream`,
    { adapter: "fetch", responseType: "stream", signal },
  );
  if (!String(response.headers["content-type"]).includes("text/event-stream")) {
    throw new Error("Unexpected conversation stream response");
  }
  await readSseFrames(response.data, ({ event, data }) => {
    if (event === "ready") onMessage({ type: "ready" });
    else if (event === "resync") onMessage({ type: "resync" });
    else if (event === "changed") {
      const hint = JSON.parse(data) as ConversationHint;
      if (hint.thread_id) onMessage({ type: "changed", hint });
    }
  });
}
