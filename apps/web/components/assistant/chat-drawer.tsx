"use client";

import React, { useState, useRef, useEffect } from "react";
import {
  Bot,
  Sparkles,
  X,
  Minimize2,
  Send,
  Loader2,
  Trash2,
  HelpCircle,
  MessageSquare,
  Wrench,
} from "lucide-react";
import { Button } from "@hexta/ui";
import { useAIAgent } from "@/hooks/use-ai-agent";
import { InteractiveDraftCard } from "./interactive-draft-card";

interface ChatDrawerProps {
  tenantId?: string;
}

export function ChatDrawer({ tenantId = "tenant-default" }: ChatDrawerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const {
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
  } = useAIAgent({ tenantId });

  // Auto-scroll to bottom on message updates
  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, isStreaming, isOpen, activeToolCall]);

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    const query = inputValue.trim();
    if (!query || isStreaming) return;

    setInputValue("");
    sendMessage(query, tenantId);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleQuickPrompt = (prompt: string) => {
    sendMessage(prompt, tenantId);
  };

  const renderToolName = (tool: string) => {
    switch (tool) {
      case "search_products":
        return "Đang tra cứu danh mục sản phẩm...";
      case "check_stock_availability":
        return "Đang kiểm tra số lượng tồn kho...";
      case "propose_order_draft":
        return "Đang tạo bản nháp đơn hàng...";
      case "lookup_customer":
        return "Đang tìm kiếm thông tin khách hàng...";
      default:
        return `Đang thực thi công cụ: ${tool}`;
    }
  };

  return (
    <div className="fixed bottom-6 right-6 z-50">
      {/* Floating Trigger Button */}
      {!isOpen && (
        <button
          type="button"
          onClick={() => setIsOpen(true)}
          className="group relative flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-2xl transition-all duration-300 hover:scale-110 active:scale-95"
          aria-label="Open AI Assistant"
        >
          <div className="absolute -inset-0.5 rounded-full bg-gradient-to-r from-primary to-blue-400 opacity-70 blur group-hover:opacity-100 transition duration-300" />
          <div className="relative flex h-full w-full items-center justify-center rounded-full bg-primary">
            <Sparkles className="h-6 w-6 text-white transition-transform group-hover:rotate-12" />
          </div>
          <span className="absolute -top-1 -right-1 flex h-4 w-4">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
            <span className="relative inline-flex h-4 w-4 rounded-full bg-emerald-500" />
          </span>
        </button>
      )}

      {/* Floating Chat Drawer Window */}
      {isOpen && (
        <div className="flex h-[620px] w-[420px] max-h-[85vh] max-w-[calc(100vw-2rem)] flex-col rounded-3xl border border-border/80 bg-background/95 backdrop-blur-xl shadow-2xl overflow-hidden transition-all duration-200">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-border/60 bg-muted/30 px-5 py-3.5">
            <div className="flex items-center gap-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
                <Bot className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-sm font-semibold text-foreground flex items-center gap-1.5">
                  Hexta AI Assistant
                </h3>
                <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                  <span className="h-2 w-2 rounded-full bg-emerald-500" />
                  <span>Sẵn sàng hỗ trợ bán hàng</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={clearChat}
                className="rounded-lg p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground"
                title="Xóa đoạn chat"
              >
                <Trash2 className="h-4 w-4" />
              </button>
              <button
                type="button"
                onClick={() => setIsOpen(false)}
                className="rounded-lg p-1.5 text-muted-foreground transition hover:bg-muted hover:text-foreground"
                title="Thu nhỏ"
              >
                <Minimize2 className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Conversation Thread */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.length === 0 && (
              <div className="flex flex-col items-center justify-center h-full text-center p-6 space-y-4 text-muted-foreground">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                  <Sparkles className="h-6 w-6" />
                </div>
                <div>
                  <h4 className="font-semibold text-foreground text-sm">
                    Trợ Lý AI Thông Minh Hexta
                  </h4>
                  <p className="text-xs text-muted-foreground mt-1">
                    Hỗ trợ tạo đơn hàng nhanh, tra cứu sản phẩm và kiểm tra tồn kho bằng ngôn ngữ tự nhiên.
                  </p>
                </div>

                <div className="w-full space-y-2 pt-2 text-left">
                  <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                    Gợi ý câu hỏi:
                  </p>
                  <button
                    type="button"
                    onClick={() =>
                      handleQuickPrompt(
                        "Anh Minh (0912345678) lấy 2 hộp sữa bắp non và 1 bánh mì bơ tỏi"
                      )
                    }
                    className="w-full rounded-xl border border-border/70 bg-card p-2.5 text-xs text-foreground transition hover:border-primary hover:bg-primary/5 text-left"
                  >
                    📦 "Anh Minh (0912345678) lấy 2 hộp sữa bắp non..."
                  </button>
                  <button
                    type="button"
                    onClick={() => handleQuickPrompt("Kiểm tra tồn kho sữa bắp non")}
                    className="w-full rounded-xl border border-border/70 bg-card p-2.5 text-xs text-foreground transition hover:border-primary hover:bg-primary/5 text-left"
                  >
                    🔍 "Kiểm tra tồn kho sữa bắp non"
                  </button>
                </div>
              </div>
            )}

            {messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex flex-col ${
                  msg.role === "user" ? "items-end" : "items-start"
                }`}
              >
                {/* Text Message Bubble */}
                {msg.content && (
                  <div
                    className={`max-w-[85%] rounded-2xl px-4 py-2.5 text-xs leading-relaxed ${
                      msg.role === "user"
                        ? "bg-primary text-primary-foreground rounded-br-xs shadow-sm"
                        : "bg-muted/70 text-foreground rounded-bl-xs border border-border/40"
                    }`}
                  >
                    <p className="whitespace-pre-wrap">{msg.content}</p>
                  </div>
                )}

                {/* Inline Draft Card if attached */}
                {msg.draft && (
                  <div className="w-full">
                    <InteractiveDraftCard
                      draft={msg.draft}
                      onUpdateItem={(idx, qty, variant) =>
                        updateDraftItem(msg.draft!.draft_id, idx, qty, variant)
                      }
                      onConfirm={async (draft) => {
                        await confirmDraft(draft);
                      }}
                      onDiscard={(draftId) => discardDraft(draftId)}
                      isConfirming={isConfirming}
                    />
                  </div>
                )}
              </div>
            ))}

            {/* Active Tool Call Indicator */}
            {activeToolCall && (
              <div className="flex items-center gap-2 rounded-xl bg-primary/5 border border-primary/20 px-3 py-2 text-xs text-primary animate-pulse">
                <Wrench className="h-3.5 w-3.5 shrink-0" />
                <span>{renderToolName(activeToolCall)}</span>
              </div>
            )}

            {/* Streaming Typing Indicator */}
            {isStreaming && !activeToolCall && (
              <div className="flex items-center gap-1.5 text-muted-foreground px-2 py-1">
                <span className="h-1.5 w-1.5 rounded-full bg-primary animate-bounce" />
                <span className="h-1.5 w-1.5 rounded-full bg-primary animate-bounce [animation-delay:0.2s]" />
                <span className="h-1.5 w-1.5 rounded-full bg-primary animate-bounce [animation-delay:0.4s]" />
              </div>
            )}

            {error && (
              <div className="rounded-xl border border-destructive/20 bg-destructive/10 p-3 text-xs text-destructive">
                {error}
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {/* Input Bar */}
          <form
            onSubmit={handleSubmit}
            className="border-t border-border/60 bg-background/80 p-3"
          >
            <div className="relative flex items-center">
              <textarea
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Nhập yêu cầu tạo đơn hoặc câu hỏi..."
                rows={1}
                disabled={isStreaming}
                className="w-full resize-none rounded-2xl border border-border bg-card px-4 py-2.5 pr-12 text-xs text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:opacity-50"
              />
              <button
                type="submit"
                disabled={!inputValue.trim() || isStreaming}
                className="absolute right-2 flex h-8 w-8 items-center justify-center rounded-xl bg-primary text-primary-foreground transition hover:opacity-90 disabled:opacity-30"
              >
                {isStreaming ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Send className="h-3.5 w-3.5" />
                )}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
