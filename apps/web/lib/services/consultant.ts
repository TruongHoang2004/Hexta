import { readConsultEvents } from "./consultant-sse";
import { sdk } from "@/lib/sdk";

export interface ConsultProduct {
  product_id: string;
  product_name: string;
  variant_id: string;
  price: string;
  image_url?: string;
}
export interface ConsultDraft {
  draft_id: string;
  total_amount: string;
  items: {
    product_id: string;
    variant_id?: string;
    product_name: string;
    quantity: number;
    unit_price: string;
  }[];
}
export interface ConsultEvent {
  type: string;
  text?: string;
  error?: string;
  products?: ConsultProduct[];
  draft?: ConsultDraft;
  quick_replies?: string[];
  needs_human?: boolean;
}
export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

// POST SSE uses the configured SDK's fetch transport, including its auth/refresh interceptors.
export async function converse(
  tenant: string,
  session: string,
  prompt: string,
  history: ChatMessage[],
  signal: AbortSignal,
  onEvent: (event: ConsultEvent) => void,
) {
  const response = await sdk.client.post<ReadableStream<Uint8Array>>(
    "/api/v1/ai/converse",
    {
      prompt,
      conversation_id: session,
      history: history.slice(-40).map((message) => ({
        ...message,
        content: message.content.slice(0, 4000),
      })),
    },
    {
      headers: { "X-Tenant-ID": tenant },
      adapter: "fetch",
      responseType: "stream",
      signal,
    },
  );
  if (!String(response.headers["content-type"]).includes("text/event-stream"))
    throw new Error("Unexpected consultation response");
  await readConsultEvents(response.data, onEvent);
}

export const checkoutKey = (tenant: string) =>
  `hexta:consultant-checkout:${tenant}`;
