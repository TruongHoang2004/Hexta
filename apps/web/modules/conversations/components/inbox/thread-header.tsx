"use client";

import { StatusChip, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import type { MessagingDetail } from "../../services/messaging";
import type { ThreadAccess } from "../../utils";
import { turnState } from "../../utils";

/** Conversation title, handling state and the claim / handoff / return / refresh actions. */
export function ThreadHeader({
  detail,
  access,
  disabled,
  loading,
  onClaim,
  onHandoff,
  onReturnToBot,
  onRefresh,
}: {
  detail: MessagingDetail | null;
  access: ThreadAccess;
  /** Busy, or a send is uncertain: commands could act on stale state. */
  disabled: boolean;
  loading: boolean;
  onClaim: () => void;
  onHandoff: () => void;
  onReturnToBot: () => void;
  onRefresh: () => void;
}) {
  const t = useT();
  const userID = useAuthStore((state) => state.user?.id);
  const turn = detail?.conversation.turn;
  const state = turnState(t, turn, userID);
  const source = detail?.conversation.source;
  return (
    <header className="flex flex-wrap items-center justify-between gap-3 border-b border-border p-4">
      <div className="min-w-0">
        <h2 className="truncate text-lg font-semibold">
          {detail?.conversation.thread.display_name ?? t("desk.conversation")}
        </h2>
        {turn && (
          <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted">
            <StatusChip tone={state.tone}>{state.label}</StatusChip>
            {turn.needs_human && <StatusChip tone="warning">{t("desk.needsHuman")}</StatusChip>}
            <span>{source ? t(`desk.sources.${source}`) : ""}</span>
          </div>
        )}
      </div>
      <div className="flex flex-wrap gap-2">
        {access.canClaim && (
          <button type="button" disabled={disabled} className={buttonClass} onClick={onClaim}>
            {access.assigned ? t("desk.actions.takeControl") : t("desk.actions.claim")}
          </button>
        )}
        {access.canHandoff && (
          <button type="button" className={secondaryClass} disabled={disabled} onClick={onHandoff}>
            {t("desk.actions.handoff")}
          </button>
        )}
        {access.canReturn && (
          <button type="button" className={secondaryClass} disabled={disabled} onClick={onReturnToBot}>
            {t("desk.actions.returnToBot")}
          </button>
        )}
        <button type="button" disabled={loading} className={secondaryClass} onClick={onRefresh}>
          {t("desk.refresh")}
        </button>
      </div>
    </header>
  );
}
