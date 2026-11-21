"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import ReactMarkdown from "react-markdown";
import { MessageCircle, X, Package } from "lucide-react";
import { Button, Input } from "@hexta/ui";
import {
  converse,
  checkoutKey,
  type ChatMessage,
  type ConsultDraft,
  type ConsultProduct,
} from "@/lib/services/consultant";
import { buttonClass, secondaryClass } from "@/components/sales/primitives";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";

const money = (value: string) =>
  new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(
    Number(value),
  );
type Turn = ChatMessage & {
  products?: ConsultProduct[];
  draft?: ConsultDraft;
  replies?: string[];
  needsHuman?: boolean;
};
export function AIConsultantWidget({ tenantID }: { tenantID: string }) {
  const router = useRouter();
  const t = useT();
  const errorText = useErrorText();
  const [open, setOpen] = useState(false),
    [input, setInput] = useState(""),
    [turns, setTurns] = useState<Turn[]>([]),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const pending = useRef<AbortController | null>(null),
    session = useRef(""),
    end = useRef<HTMLDivElement>(null),
    toggle = useRef<HTMLButtonElement>(null),
    field = useRef<HTMLInputElement>(null);
  useEffect(() => () => pending.current?.abort(), []);
  useEffect(() => {
    if (open) field.current?.focus();
  }, [open]);
  useEffect(() => {
    end.current?.scrollIntoView({ block: "nearest" });
  }, [turns, busy]);
  function close() {
    pending.current?.abort();
    setOpen(false);
    toggle.current?.focus();
  }
  async function send(text: string) {
    if (pending.current || !text.trim()) return;
    const controller = new AbortController();
    pending.current = controller;
    if (!session.current) session.current = crypto.randomUUID();
    const history = turns
      .filter((turn) => turn.content)
      .map(({ role, content }) => ({ role, content }));
    setTurns((previous) => [
      ...previous,
      { role: "user", content: text },
      { role: "assistant", content: "" },
    ]);
    setInput("");
    setBusy(true);
    setError("");
    try {
      await converse(
        tenantID,
        session.current,
        text,
        history,
        controller.signal,
        (event) => {
          if (controller.signal.aborted) return;
          setTurns((previous) =>
            previous.map((turn, index) =>
              index === previous.length - 1
                ? {
                    ...turn,
                    ...(event.type === "message"
                      ? {
                          content: turn.content + (event.text || ""),
                          needsHuman: event.needs_human,
                        }
                      : {}),
                    ...(event.type === "products"
                      ? { products: event.products }
                      : {}),
                    ...(event.type === "draft_order"
                      ? { draft: event.draft }
                      : {}),
                    ...(event.type === "quick_replies"
                      ? { replies: event.quick_replies }
                      : {}),
                  }
                : turn,
            ),
          );
        },
      );
    } catch (err) {
      if (!controller.signal.aborted)
        setError(errorText(err));
    } finally {
      if (pending.current === controller) {
        pending.current = null;
        setBusy(false);
      }
    }
  }
  function checkout(draft: ConsultDraft) {
    try {
      sessionStorage.setItem(checkoutKey(tenantID), JSON.stringify(draft));
      router.push(`/tenants/${tenantID}/checkout`);
    } catch {
      setError(t("consultant.checkoutUnavailable"));
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
            <Button
              variant="ghost"
              size="icon"
              aria-label={t("consultant.close")}
              onClick={close}
            >
              <X className="h-4 w-4" />
            </Button>
          </header>
          <div
            className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4"
            role="log"
            aria-label="Conversation"
            aria-live="polite"
          >
            {!turns.length && (
              <p className="text-sm text-muted">
                {t("consultant.intro")}
              </p>
            )}
            {turns.map((turn, i) => (
              <article
                key={i}
                className={`space-y-3 rounded-lg p-3 text-sm ${turn.role === "user" ? "bg-primary/10" : "bg-surface"}`}
              >
                <p className="text-xs font-semibold">
                  {turn.role === "user" ? t("consultant.you") : t("consultant.assistant")}
                </p>
                <div className="space-y-2 break-words [&_ul]:list-inside [&_ul]:list-disc [&_strong]:font-bold">
                  <ReactMarkdown
                    skipHtml
                    components={{
                      a: ({ children }) => <span>{children}</span>,
                      img: () => null,
                    }}
                  >
                    {turn.content}
                  </ReactMarkdown>
                </div>
                {turn.needsHuman && (
                  <p role="status">
                    {t("consultant.needsHuman")}
                  </p>
                )}
                {!!turn.products?.length && (
                  <div
                    className="flex snap-x gap-3 overflow-x-auto pb-2"
                    aria-label={t("consultant.recommended")}
                  >
                    {turn.products.map((p) => (
                      <div
                        key={p.variant_id || p.product_id}
                        className="w-56 shrink-0 snap-start space-y-2 rounded-lg border border-border p-3"
                      >
                        {p.image_url ? (
                          <Image
                            src={p.image_url}
                            alt={p.product_name}
                            width={200}
                            height={96}
                            unoptimized
                            className="h-24 w-full rounded object-contain"
                          />
                        ) : (
                          <div
                            className="flex h-24 items-center justify-center rounded bg-background"
                            aria-label={t("consultant.noImage")}
                          >
                            <Package className="h-10 w-10 text-muted" />
                          </div>
                        )}
                        <p className="font-semibold">{p.product_name}</p>
                        <p>{money(p.price)}</p>
                        <button
                          className={secondaryClass}
                          onClick={() =>
                            router.push(
                              `/tenants/${tenantID}/storefront#${encodeURIComponent(p.product_id)}`,
                            )
                          }
                        >
                          {t("consultant.details")}
                        </button>
                        <button
                          className={buttonClass}
                          onClick={() =>
                            checkout({
                              draft_id: crypto.randomUUID(),
                              total_amount: p.price,
                              items: [
                                {
                                  product_id: p.product_id,
                                  variant_id: p.variant_id,
                                  product_name: p.product_name,
                                  quantity: 1,
                                  unit_price: p.price,
                                },
                              ],
                            })
                          }
                        >
                          {t("consultant.buy")}
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                {turn.draft && (
                  <div className="space-y-2 rounded-lg border border-primary p-3">
                    <h3 className="font-semibold">{t("consultant.draftTitle")}</h3>
                    {turn.draft.items.map((item, index) => (
                      <p key={index}>
                        {item.quantity} × {item.product_name} ·{" "}
                        {money(item.unit_price)}
                      </p>
                    ))}
                    <p className="font-semibold">
                      {t("consultant.total", { amount: money(turn.draft.total_amount) })}
                    </p>
                    <p className="text-xs text-muted">
                      {t("consultant.draftNote")}
                    </p>
                    <button
                      className={buttonClass}
                      onClick={() => checkout(turn.draft!)}
                    >
                      {t("consultant.confirm")}
                    </button>
                  </div>
                )}
                {i === turns.length - 1 && (
                  <div className="flex flex-wrap gap-2">
                    {turn.replies?.map((reply) => (
                      <button
                        disabled={busy}
                        key={reply}
                        className={secondaryClass}
                        onClick={() => void send(reply)}
                      >
                        {reply}
                      </button>
                    ))}
                  </div>
                )}
              </article>
            ))}
            {busy && <p role="status">{t("consultant.preparing")}</p>}
            <div ref={end} />
          </div>
          {error && (
            <p role="alert" className="px-4 text-sm text-error">
              {error}
            </p>
          )}
          <form
            className="flex gap-2 border-t border-border p-3"
            onSubmit={(e) => {
              e.preventDefault();
              void send(input);
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
            <Button disabled={busy || !input.trim()} type="submit">
              {t("consultant.send")}
            </Button>
            {busy && (
              <Button
                variant="ghost"
                type="button"
                onClick={() => pending.current?.abort()}
              >
                {t("consultant.stop")}
              </Button>
            )}
          </form>
        </section>
      )}
    </>
  );
}
