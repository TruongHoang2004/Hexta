"use client";

import { useState, useCallback, useRef } from "react";
import {
  aiAgentService,
  ChatMessage,
  OrderDraft,
} from "@/lib/services/ai-agent-service";

export interface UseAIAgentOptions {
  tenantId?: string;
  initialMessages?: ChatMessage[];
}

export function useAIAgent(options: UseAIAgentOptions = {}) {
  const { tenantId = "tenant-default", initialMessages = [] } = options;

  const [messages, setMessages] = useState<ChatMessage[]>(initialMessages);
  const [isStreaming, setIsStreaming] = useState(false);
  const [activeToolCall, setActiveToolCall] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isConfirming, setIsConfirming] = useState(false);

  const abortControllerRef = useRef<AbortController | null>(null);

  /**
   * Dispatches a user query to the AI Agent over SSE.
   */
  const sendMessage = useCallback(
    async (prompt: string, overrideTenantId?: string) => {
      const trimmed = prompt.trim();
      if (!trimmed || isStreaming) return;

      setError(null);
      const userMessageId = `msg_user_${Date.now()}`;
      const assistantMessageId = `msg_asst_${Date.now()}`;

      const userMsg: ChatMessage = {
        id: userMessageId,
        role: "user",
        content: trimmed,
        status: "done",
        timestamp: Date.now(),
      };

      const assistantMsg: ChatMessage = {
        id: assistantMessageId,
        role: "assistant",
        content: "",
        status: "streaming",
        timestamp: Date.now(),
      };

      setMessages((prev) => [...prev, userMsg, assistantMsg]);
      setIsStreaming(true);
      setActiveToolCall(null);

      // Create new abort controller for this turn
      abortControllerRef.current?.abort();
      const controller = new AbortController();
      abortControllerRef.current = controller;

      // Extract past 5 turns for context
      const history = messages.slice(-10).map((m) => ({
        role: m.role,
        content: m.content,
      }));

      try {
        await aiAgentService.converseStream({
          prompt: trimmed,
          history,
          tenantId: overrideTenantId || tenantId,
          signal: controller.signal,
          onToken: (token) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === assistantMessageId
                  ? { ...msg, content: msg.content + token }
                  : msg
              )
            );
          },
          onToolCall: (toolName) => {
            setActiveToolCall(toolName);
          },
          onDraftProposed: (draft) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === assistantMessageId
                  ? { ...msg, draft: { ...draft, status: "proposed" } }
                  : msg
              )
            );
            setActiveToolCall(null);
          },
          onError: (errMsg) => {
            setError(errMsg);
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === assistantMessageId
                  ? { ...msg, status: "error" }
                  : msg
              )
            );
            setIsStreaming(false);
            setActiveToolCall(null);
          },
          onDone: () => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === assistantMessageId
                  ? { ...msg, status: "done" }
                  : msg
              )
            );
            setIsStreaming(false);
            setActiveToolCall(null);
          },
        });
      } catch (err: unknown) {
        if (err instanceof Error && err.name === "AbortError") {
          return;
        }
        const msg = err instanceof Error ? err.message : "Đã xảy ra lỗi khi trao đổi với AI.";
        setError(msg);
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantMessageId ? { ...m, status: "error" } : m
          )
        );
      } finally {
        setIsStreaming(false);
        setActiveToolCall(null);
      }
    },
    [isStreaming, messages, tenantId]
  );

  /**
   * Inline item adjustment before confirmation: adjusts quantity and recalculated totals.
   */
  const updateDraftItem = useCallback(
    (draftId: string, itemIndex: number, newQuantity: number, newVariant?: string) => {
      if (newQuantity < 1) return;

      setMessages((prev) =>
        prev.map((msg) => {
          if (!msg.draft || msg.draft.draft_id !== draftId) return msg;

          const updatedItems = msg.draft.items.map((item, idx) => {
            if (idx !== itemIndex) return item;
            const qty = Math.max(1, newQuantity);
            const subtotal = item.unit_price * qty;
            return {
              ...item,
              quantity: qty,
              variant: newVariant !== undefined ? newVariant : item.variant,
              subtotal,
            };
          });

          const totalAmount = updatedItems.reduce(
            (sum, item) => sum + item.subtotal,
            0
          );

          return {
            ...msg,
            draft: {
              ...msg.draft,
              items: updatedItems,
              total_amount: totalAmount,
            },
          };
        })
      );
    },
    []
  );

  /**
   * Human-in-the-Loop Confirmation: dispatches POST /api/v1/orders with draft_id.
   */
  const confirmDraft = useCallback(
    async (draft: OrderDraft, overrideTenantId?: string) => {
      setIsConfirming(true);
      setError(null);

      try {
        const orderResult = await aiAgentService.confirmDraftOrder({
          tenantId: overrideTenantId || draft.tenant_id || tenantId,
          draft,
        });

        // Update draft state to confirmed in message history
        setMessages((prev) =>
          prev.map((msg) => {
            if (msg.draft?.draft_id === draft.draft_id) {
              return {
                ...msg,
                draft: {
                  ...msg.draft,
                  status: "confirmed",
                  order_id: orderResult.id,
                  order_number: orderResult.order_number,
                },
              };
            }
            return msg;
          })
        );

        return orderResult;
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : "Xác nhận đơn hàng thất bại";
        setError(msg);
        throw err;
      } finally {
        setIsConfirming(false);
      }
    },
    [tenantId]
  );

  /**
   * Discards an unconfirmed draft.
   */
  const discardDraft = useCallback((draftId: string) => {
    setMessages((prev) =>
      prev.map((msg) => {
        if (msg.draft?.draft_id === draftId) {
          return {
            ...msg,
            draft: {
              ...msg.draft,
              status: "discarded",
            },
          };
        }
        return msg;
      })
    );
  }, []);

  /**
   * Clears conversation history.
   */
  const clearChat = useCallback(() => {
    abortControllerRef.current?.abort();
    setMessages([]);
    setIsStreaming(false);
    setActiveToolCall(null);
    setError(null);
  }, []);

  return {
    messages,
    isStreaming,
    activeToolCall,
    error,
    isConfirming,
    sendMessage,
    updateDraftItem,
    confirmDraft,
    discardDraft,
    clearChat,
  };
}
