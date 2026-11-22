"use client";

import { useT } from "@/lib/i18n/provider";
import type { WorkEvent } from "../../types";
import { eventActionLabel } from "../../utils";

interface Props {
  events: WorkEvent[] | undefined;
  person: (userID: string | null | undefined) => string;
}

/** Append-only audit trail of the work item. */
export function WorkHistory({ events, person }: Props) {
  const t = useT();
  return (
    <div className="space-y-3 border-t border-border pt-4">
      <h3 className="font-semibold">{t("work.history.title")}</h3>
      <p className="text-xs text-muted">{t("work.history.note")}</p>
      <ol className="space-y-3">
        {events?.map((event) => (
          <li key={event.id} className="space-y-1 break-all rounded-lg bg-background p-3 text-xs">
            <p className="font-medium">
              v{event.version} · {eventActionLabel(t, event.action)}
            </p>
            <p>
              {t("work.history.actor", {
                name: person(event.actor_id),
                from: event.from_step + 1,
                to: event.current_step + 1,
              })}
            </p>
            <p>
              {t("work.history.responsible", {
                from: person(event.from_responsible_id),
                to: person(event.responsible_id),
              })}
              {event.recipient_id ? t("work.history.nominated", { name: person(event.recipient_id) }) : ""}
            </p>
            <p className="text-muted">{new Date(event.created_at).toLocaleString()}</p>
          </li>
        ))}
      </ol>
    </div>
  );
}
