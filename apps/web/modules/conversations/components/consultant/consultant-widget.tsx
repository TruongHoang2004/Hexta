"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { MessageCircle, X } from "lucide-react";
import { Button, Input } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";
import { useConsultantChat } from "../../hooks/use-consultant-chat";
import type { ConsultDraft } from "../../services/consultant";
import { saveCheckoutDraft, singleItemDraft } from "../../utils";
import { ConsultantMessage } from "./consultant-message";

/** Floating AI consultant chat on the storefront. */
export function ConsultantWidget({ tenantID }: { tenantID: string }) {
  const router = useRouter();
  const t = useT();
  const chat = useConsultantChat(tenantID);
  const [open, setOpen] = useState(false);
  const [input, setInput] = useState("");
  const end = useRef<HTMLDivElement>(null);
  const toggle = useRef<HTMLButtonElement>(null);
  const field = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (open) field.current?.focus();
  }, [open]);
  useEffect(() => {
    end.current?.scrollIntoView({ block: "nearest" });
  }, [chat.turns, chat.busy]);

  function close() {
    chat.stop();
    setOpen(false);
    toggle.current?.focus();
  }
  function send(text: string) {
    if (chat.busy || !text.trim()) return;
    setInput("");
    void chat.send(text);
  }
  function checkout(draft: ConsultDraft) {
    try {
      saveCheckoutDraft(tenantID, draft);
      router.push(`/tenants/${tenantID}/checkout`);
    } catch {
      chat.setError(t("consultant.checkoutUnavailable"));
    }
  }

  return (
    <>
      <Button
        ref={toggle}
        className="fixed bottom-5 right-5 z-40 rounded-full shadow-lg"
        aria-label={t("consultant.open")}
        aria-expanded={open}
        aria-controls="ai-consultant"
        onClick={() => (open ? close() : setOpen(true))}
      >
        <MessageCircle className="mr-2 h-5 w-5" />
        {t("consultant.launcher")}
      </Button>
      {open && (
        <section
          id="ai-consultant"
          role="dialog"
          aria-label={t("consultant.title")}
          onKeyDown={(e) => {
            if (e.key === "Escape") close();
          }}
          className="fixed bottom-20 right-3 z-50 flex max-h-[75dvh] w-[calc(100%-1.5rem)] flex-col rounded-xl border border-border bg-background text-foreground shadow-xl sm:right-5 sm:w-96"
        >
          <header className="flex items-center justify-between border-b border-border p-4">
            <h2 className="font-semibold">{t("consultant.title")}</h2>
            <Button variant="ghost" size="icon" aria-label={t("consultant.close")} onClick={close}>
              <X className="h-4 w-4" />
            </Button>
          </header>
          <div
            className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4"
            role="log"
            aria-label="Conversation"
            aria-live="polite"
          >
            {!chat.turns.length && <p className="text-sm text-muted">{t("consultant.intro")}</p>}
            {chat.turns.map((turn, i) => (
              <ConsultantMessage
                key={i}
                turn={turn}
                latest={i === chat.turns.length - 1}
                busy={chat.busy}
                onReply={send}
                onProductDetails={(p) =>
                  router.push(`/tenants/${tenantID}/storefront#${encodeURIComponent(p.product_id)}`)
                }
                onBuyProduct={(p) => checkout(singleItemDraft(p))}
                onCheckout={checkout}
              />
            ))}
            {chat.busy && <p role="status">{t("consultant.preparing")}</p>}
            <div ref={end} />
          </div>
          {chat.error && (
            <p role="alert" className="px-4 text-sm text-error">
              {chat.error}
            </p>
          )}
          <form
            className="flex gap-2 border-t border-border p-3"
            onSubmit={(e) => {
              e.preventDefault();
              send(input);
            }}
          >
            <Input
              ref={field}
              aria-label={t("consultant.message")}
              maxLength={4000}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder={t("consultant.placeholder")}
            />
            <Button disabled={chat.busy || !input.trim()} type="submit">
              {t("consultant.send")}
            </Button>
            {chat.busy && (
              <Button variant="ghost" type="button" onClick={chat.stop}>
                {t("consultant.stop")}
              </Button>
            )}
          </form>
        </section>
      )}
    </>
  );
}
