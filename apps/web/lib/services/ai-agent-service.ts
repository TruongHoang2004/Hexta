import { sdk } from "@/lib/sdk";

export interface DraftItem {
  product_id?: string;
  product_name: string;
  variant?: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
}

export interface OrderDraft {
  draft_id: string;
  tenant_id: string;
  customer_name?: string;
  customer_phone?: string;
  shipping_address?: string;
  items: DraftItem[];
  total_amount: number;
  payment_method?: string;
  notes?: string;
  status: "proposed" | "confirmed" | "discarded";
  order_id?: string;
  order_number?: string;
  created_at?: string;
  expires_at?: string;
}

export interface SSEEvent {
  type: "token" | "tool_call" | "tool_result" | "draft_proposed" | "done" | "error";
  text?: string;
  tool?: string;
  draft?: OrderDraft;
  error?: string;
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  draft?: OrderDraft;
  status?: "sending" | "streaming" | "done" | "error";
  timestamp: number;
}

export interface ConverseStreamOptions {
  prompt: string;
  history?: Array<{ role: string; content: string }>;
  tenantId?: string;
  onToken?: (text: string) => void;
  onToolCall?: (tool: string) => void;
  onDraftProposed?: (draft: OrderDraft) => void;
  onError?: (err: string) => void;
  onDone?: () => void;
  signal?: AbortSignal;
}

export interface ConfirmDraftParams {
  tenantId: string;
  draft: OrderDraft;
  notes?: string;
}

export interface OrderCreationResult {
  id: string;
  order_number: string;
  status: string;
  total_amount: number;
}

class AIAgentService {
  private getApiUrl(): string {
    return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  }

  private async getAuthToken(): Promise<string | null> {
    return await sdk.identity.getAccessToken();
  }

  /**
   * Converse with AI Agent over Server-Sent Events (SSE).
   * Implements resilient streaming with silent token refresh on 401.
   */
  async converseStream(options: ConverseStreamOptions): Promise<void> {
    const {
      prompt,
      history = [],
      tenantId = "tenant-default",
      onToken,
      onToolCall,
      onDraftProposed,
      onError,
      onDone,
      signal,
    } = options;

    const executeStreamRequest = async (token: string | null): Promise<Response> => {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        Accept: "text/event-stream",
        "X-Tenant-ID": tenantId,
      };

      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const body = JSON.stringify({
        prompt,
        history: history.map((h) => ({
          role: h.role,
          content: h.content,
        })),
      });

      return await fetch(`${this.getApiUrl()}/api/v1/ai/agent/converse`, {
        method: "POST",
        headers,
        body,
        signal,
      });
    };

    let token = await this.getAuthToken();
    let response = await executeStreamRequest(token);

    // If 401 Unauthorized, perform silent token renewal and retry once
    if (response.status === 401) {
      try {
        const refreshResult = await sdk.identity.refreshToken();
        token = refreshResult.token || null;
        response = await executeStreamRequest(token);
      } catch (refreshErr) {
        const errMsg = "Authentication expired. Please sign in again.";
        onError?.(errMsg);
        throw new Error(errMsg);
      }
    }

    if (!response.ok) {
      let errDetail = `Server returned HTTP ${response.status}`;
      try {
        const errJson = await response.json();
        if (errJson?.error?.message) {
          errDetail = errJson.error.message;
        }
      } catch {
        // use default status message
      }
      onError?.(errDetail);
      throw new Error(errDetail);
    }

    if (!response.body) {
      onError?.("Response stream body is empty");
      return;
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    try {
      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const events = buffer.split("\n\n");
        buffer = events.pop() || "";

        for (const evtBlock of events) {
          const lines = evtBlock.split("\n");
          for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed.startsWith("data:")) {
              const dataStr = trimmed.replace(/^data:\s*/, "");
              if (!dataStr) continue;

              try {
                const parsed: SSEEvent = JSON.parse(dataStr);
                switch (parsed.type) {
                  case "token":
                    if (parsed.text) onToken?.(parsed.text);
                    break;
                  case "tool_call":
                    if (parsed.tool) onToolCall?.(parsed.tool);
                    break;
                  case "draft_proposed":
                    if (parsed.draft) onDraftProposed?.(parsed.draft);
                    break;
                  case "error":
                    if (parsed.error) onError?.(parsed.error);
                    break;
                  case "done":
                    onDone?.();
                    break;
                }
              } catch (parseErr) {
                // Ignore SSE malformed line parse errors
              }
            }
          }
        }
      }

      onDone?.();
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "AbortError") {
        return;
      }
      const msg = err instanceof Error ? err.message : String(err);
      onError?.(msg);
      throw err;
    }
  }

  /**
   * Retrieves an existing order draft from Redis cache.
   */
  async getDraft(tenantId: string, draftId: string): Promise<OrderDraft> {
    const token = await this.getAuthToken();
    const headers: Record<string, string> = {
      "X-Tenant-ID": tenantId,
    };
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${this.getApiUrl()}/api/v1/ai/agent/drafts/${draftId}`, {
      method: "GET",
      headers,
    });

    if (!res.ok) {
      throw new Error(`Failed to fetch draft: HTTP ${res.status}`);
    }

    const json = await res.json();
    return json.data;
  }

  /**
   * Human-in-the-Loop action: Converts draft into active confirmed order in OMS.
   * Dispatches POST /api/v1/orders with draft_id and auto_confirm: true.
   */
  async confirmDraftOrder(params: ConfirmDraftParams): Promise<OrderCreationResult> {
    const { tenantId, draft, notes } = params;

    const executeConfirm = async (token: string | null): Promise<Response> => {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        "X-Tenant-ID": tenantId,
      };
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const orderPayload = {
        tenant_id: tenantId,
        items: draft.items.map((item) => ({
          variant_id: item.product_id || item.variant || "var-default",
          quantity: item.quantity,
          unit_price: item.unit_price,
        })),
        source: "ai_draft",
        draft_id: draft.draft_id,
        notes: notes || draft.notes || "Confirmed from AI assistant conversational draft",
        auto_confirm: true,
      };

      return await fetch(`${this.getApiUrl()}/api/v1/orders`, {
        method: "POST",
        headers,
        body: JSON.stringify(orderPayload),
      });
    };

    let token = await this.getAuthToken();
    let res = await executeConfirm(token);

    if (res.status === 401) {
      const refreshResult = await sdk.identity.refreshToken();
      token = refreshResult.token || null;
      res = await executeConfirm(token);
    }

    if (!res.ok) {
      let detail = `Failed to confirm order (HTTP ${res.status})`;
      try {
        const body = await res.json();
        if (body?.error?.message) {
          detail = body.error.message;
        }
      } catch {
        // default message
      }
      throw new Error(detail);
    }

    const json = await res.json();
    return json.data;
  }
}

export const aiAgentService = new AIAgentService();
