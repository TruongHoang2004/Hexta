"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import type { Conversation } from "../../services/messaging";
import { initialOf, turnState } from "../../utils";

/** A row in the conversation list: name, channel, state, assignee and last activity. */
export function ConversationListItem({
  row,
  active,
  assigneeName,
  onSelect,
}: {
  row: Conversation;
  active: boolean;
  assigneeName: string;
  onSelect: () => void;
}) {
  const t = useT();
  const userID = useAuthStore((state) => state.user?.id);
  const state = turnState(t, row.turn, userID);
  return (
    <li>
      <button
        type="button"
        aria-current={active ? "true" : undefined}
        className={`flex w-full flex-col gap-1 px-3 py-3 text-left hover:bg-background ${active ? "bg-primary/10 shadow-[inset_3px_0_0_var(--primary)]" : ""}`}
        onClick={onSelect}
      >
        <span className="flex items-start gap-3">
          <span
            aria-hidden="true"
            className="grid size-9 shrink-0 place-items-center rounded-full bg-primary/90 text-sm font-bold text-white"
          >
            {initialOf(row.thread.display_name)}
          </span>
          <span className="flex min-w-0 flex-1 flex-col gap-1">
            <span className="flex items-center justify-between gap-2">
              <span className="truncate font-semibold">{row.thread.display_name}</span>
              <span className="shrink-0 text-xs text-muted">
                {row.turn &&
                  new Date(row.turn.last_activity_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
              </span>
            </span>
            <span className="flex flex-wrap items-center gap-1.5">
              <StatusChip>{t(`desk.sources.${row.source}`)}</StatusChip>
              <StatusChip tone={state.tone}>{state.label}</StatusChip>
              {row.turn?.needs_human && row.turn.lifecycle === "open" && (
                <StatusChip tone="warning">{t("desk.needsHuman")}</StatusChip>
              )}
              {row.turn?.review_pending && row.turn.lifecycle === "open" && (
                <StatusChip tone="warning">{t("desk.reviewPending")}</StatusChip>
              )}
            </span>
            <span className="truncate text-xs text-muted">{assigneeName || t("desk.unassigned")}</span>
          </span>
        </span>
      </button>
    </li>
  );
}
