"use client";

import { useState } from "react";
import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { MessagingSession, SessionInput } from "../../services/messaging";
import { sessionAccess } from "../../utils";
import { SessionListItem } from "./session-list-item";
import { SessionOutcomeDialog } from "./session-outcome-dialog";

/**
 * The conversation's sessions ("Phiên"), newest first. Clicking one jumps the thread to its first message; the
 * current session is highlighted. Only the open session can be closed; closed sessions are read-only except for
 * their outcome, which the owner or the session's last handler may edit.
 */
export function SessionPanel({
  tenantID,
  sessions,
  currentTurnID,
  userID,
  owner,
  disabled,
  memberName,
  onJump,
}: {
  tenantID: string;
  sessions: {
    items: MessagingSession[];
    nextCursor: string;
    loading: boolean;
    busy: boolean;
    error: string;
    commandError: string;
    loadMore: () => Promise<void>;
    close: (turn: MessagingSession["turn"], input: SessionInput) => Promise<boolean>;
    updateOutcome: (turn: MessagingSession["turn"], input: SessionInput) => Promise<boolean>;
  };
  currentTurnID: string;
  userID: string | undefined;
  owner: boolean;
  /** Thread commands are running or a send is uncertain. */
  disabled: boolean;
  memberName: (id: string | null | undefined) => string;
  onJump: (session: MessagingSession) => void;
}) {
  const t = useT();
  const [dialog, setDialog] = useState<{ mode: "close" | "outcome"; session: MessagingSession } | null>(null);
  const busy = sessions.busy || disabled;

  async function confirm(input: SessionInput) {
    if (!dialog) return;
    const action = dialog.mode === "close" ? sessions.close : sessions.updateOutcome;
    if (await action(dialog.session.turn, input)) setDialog(null);
  }

  return (
    <section aria-labelledby="session-panel-title" className="space-y-2 border-t border-border pt-3">
      <h3 id="session-panel-title" className="font-semibold">
        {t("desk.sessions.title")}
      </h3>
      <p className="text-xs text-muted">{t("desk.sessions.help")}</p>
      {sessions.error && (
        <p role="alert" className="text-sm text-error">
          {sessions.error}
        </p>
      )}
      {sessions.loading ? (
        <p role="status" className="text-sm">
          {t("desk.sessions.loading")}
        </p>
      ) : sessions.items.length === 0 ? (
        <p className="text-sm text-muted">{t("desk.sessions.empty")}</p>
      ) : (
        <ol className="max-h-[420px] space-y-2 overflow-y-auto" aria-label={t("desk.sessions.title")}>
          {sessions.items.map((session) => {
            const access = sessionAccess(session.turn, userID, owner);
            const handler = session.turn.assignee_id ?? session.turn.last_handler_id ?? null;
            return (
              <SessionListItem
                key={session.turn.id}
                tenantID={tenantID}
                session={session}
                current={session.turn.id === currentTurnID}
                handlerName={handler ? memberName(handler) : ""}
                canClose={access.canClose}
                canEditOutcome={access.canEditOutcome}
                disabled={busy}
                onJump={() => onJump(session)}
                onClose={() => setDialog({ mode: "close", session })}
                onEditOutcome={() => setDialog({ mode: "outcome", session })}
              />
            );
          })}
        </ol>
      )}
      {sessions.nextCursor && (
        <button type="button" className={`${secondaryClass} w-full`} disabled={busy} onClick={() => void sessions.loadMore()}>
          {t("desk.sessions.loadMore")}
        </button>
      )}
      {dialog && (
        <SessionOutcomeDialog
          key={`${dialog.mode}:${dialog.session.turn.id}:${dialog.session.turn.version}`}
          open
          mode={dialog.mode}
          initial={dialog.mode === "outcome" ? { outcome: dialog.session.turn.outcome, note: dialog.session.turn.note } : undefined}
          busy={sessions.busy}
          error={sessions.commandError}
          onCancel={() => setDialog(null)}
          onConfirm={(input) => void confirm(input)}
        />
      )}
    </section>
  );
}
