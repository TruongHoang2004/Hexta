"use client";

import { buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import type { WorkCommand, WorkCommandFields, WorkItem } from "../../types";

interface Props {
  item: WorkItem;
  busy: boolean;
  onHandoff: () => void;
  onCommand: (action: WorkCommand, fields?: WorkCommandFields) => Promise<boolean>;
  onRefresh: () => void;
}

/** Actions the current user may take; the server re-checks every command. */
export function WorkActions({ item, busy, onHandoff, onCommand, onRefresh }: Props) {
  const { isOwner } = useWorkspace();
  const userID = useAuthStore((state) => state.user?.id);
  const t = useT();
  const canProgress = item.responsible_id === userID && item.state === "active";
  const finalStep = item.current_step === item.workflow.steps.length - 1;
  return (
    <div className="flex flex-wrap gap-2">
      {canProgress && !finalStep && (
        <button type="button" className={buttonClass} disabled={busy} onClick={onHandoff}>
          {t("work.actions.handoff")}
        </button>
      )}
      {item.state === "waiting" && item.recipient_id === userID && (
        <button type="button" className={buttonClass} disabled={busy} onClick={() => void onCommand("accept")}>
          {t("work.actions.accept")}
        </button>
      )}
      {canProgress && finalStep && (
        <button type="button" className={buttonClass} disabled={busy} onClick={() => void onCommand("complete")}>
          {t("work.actions.complete")}
        </button>
      )}
      {isOwner && item.state === "waiting" && (
        <button
          type="button"
          className={secondaryClass}
          disabled={busy}
          onClick={() => void onCommand("assign", { responsible_id: item.responsible_id })}
        >
          {t("work.actions.cancelWait")}
        </button>
      )}
      <button type="button" className={secondaryClass} disabled={busy} onClick={onRefresh}>
        {t("work.actions.reload")}
      </button>
    </div>
  );
}
