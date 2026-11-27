"use client";

import { Button } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";
import type { MessagingTurn } from "../../services/messaging";
import { needsHumanReason } from "../../utils";

/** Needs-human flag beside the conversation: why, what it blocks, and the assigned sale's resolve action. */
export function NeedsHumanPanel({
  turn,
  canResolve,
  disabled,
  onResolve,
}: {
  turn: MessagingTurn;
  canResolve: boolean;
  disabled: boolean;
  onResolve: () => void;
}) {
  const t = useT();
  return (
    <section
      aria-labelledby="needs-human-title"
      className="flex flex-col gap-2 rounded-lg border border-warning bg-warning/10 p-3 text-sm"
    >
      <h3 id="needs-human-title" className="font-semibold">
        {t("desk.needsHumanPanel.title")}
      </h3>
      <p className="text-muted">
        {t("desk.needsHumanPanel.reason", { reason: needsHumanReason(t, turn) })} {t("desk.needsHumanPanel.blocks")}
      </p>
      {canResolve ? (
        <>
          <Button type="button" variant="outline" disabled={disabled} onClick={onResolve} aria-describedby="resolve-hint">
            {t("desk.needsHumanPanel.resolve")}
          </Button>
          <p id="resolve-hint" className="text-xs text-muted">
            {t("desk.needsHumanPanel.resolveHint")}
          </p>
        </>
      ) : (
        <p className="text-xs text-muted">{t("desk.needsHumanPanel.assignedOnly")}</p>
      )}
    </section>
  );
}
