"use client";

import { useT } from "@/lib/i18n/provider";
import type { OrderCaseEvent } from "../../types";

interface Props {
  events: OrderCaseEvent[];
  person: (id: string | null | undefined) => string;
}

/** Append-only case history, newest first. */
export function OrderCaseTimeline({ events, person }: Props) {
  const t = useT();
  return (
    <div className="space-y-3 border-t border-border pt-4">
      <h4 className="text-sm font-semibold">{t("orderCase.history.title")}</h4>
      <ol className="space-y-2">
        {[...events].reverse().map((event) => (
          <li key={event.id} className="space-y-1 break-words rounded-lg bg-background p-3 text-xs">
            <p>
              <span className="text-muted">{new Date(event.created_at).toLocaleString()}</span> ·{" "}
              <strong>{person(event.actor_id)}</strong> · {t(`orderCase.history.${event.action}`)}
            </p>
            {event.note && <p className="whitespace-pre-wrap">{event.note}</p>}
            {event.lines.length > 0 && (
              <p className="text-muted">{t("orderCase.history.lines", { count: event.lines.length })}</p>
            )}
          </li>
        ))}
      </ol>
    </div>
  );
}
