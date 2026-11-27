"use client";

import { useT } from "@/lib/i18n/provider";

const steps = ["script", "ai", "clarify", "handoff"] as const;

/** How the bot routes each customer message, in order. */
export function ReplyModeFlow() {
  const t = useT();
  return (
    <section aria-labelledby="reply-mode-flow" className="space-y-2 rounded-xl border border-border bg-surface p-4">
      <h3 id="reply-mode-flow" className="font-semibold">
        {t("replyMode.flow.title")}
      </h3>
      <ol className="list-decimal space-y-1 pl-5 text-sm">
        {steps.map((step) => (
          <li key={step}>{t(`replyMode.flow.${step}`)}</li>
        ))}
      </ol>
      <p className="text-xs text-muted">{t("replyMode.flow.note")}</p>
    </section>
  );
}
