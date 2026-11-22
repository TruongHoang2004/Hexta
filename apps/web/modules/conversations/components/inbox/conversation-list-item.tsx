"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import type { Conversation } from "../../services/messaging";
import { turnState } from "../../utils";

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
        <span className="flex items-center justify-between gap-2">
          <span className="truncate font-medium">{row.thread.display_name}</span>
          <span className="text-xs text-muted">{t(`desk.sources.${row.source}`)}</span>
        </span>
        <span className="flex flex-wrap items-center gap-1.5">
          <StatusChip tone={state.tone}>{state.label}</StatusChip>
          {row.turn?.needs_human && row.turn.handling_state !== "WAITING_STAFF" && (
            <StatusChip tone="warning">{t("desk.needsHuman")}</StatusChip>
          )}
        </span>
        <span className="text-xs text-muted">
          {assigneeName || t("desk.unassigned")}
          {row.turn && ` · ${new Date(row.turn.last_activity_at).toLocaleString()}`}
        </span>
      </button>
    </li>
  );
}
