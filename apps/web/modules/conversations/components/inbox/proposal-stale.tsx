"use client";

import { Button } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";

/** A stale proposal: control changed after the bot created it, so it can never be sent. */
export function ProposalStale({
  text,
  canRegenerate,
  canDiscard,
  busy,
  onRegenerate,
  onDiscard,
}: {
  text: string;
  canRegenerate: boolean;
  canDiscard: boolean;
  busy: boolean;
  onRegenerate: () => void;
  onDiscard: () => void;
}) {
  const t = useT();
  return (
    <div className="space-y-2">
      <p role="status" className="rounded-lg border border-warning bg-warning/15 p-2 text-sm">
        <strong>{t("desk.proposal.staleTitle")}</strong> — {t("desk.proposal.staleBody")}
      </p>
      {text && <p className="whitespace-pre-wrap text-sm text-muted line-through">{text}</p>}
      <div className="flex flex-wrap justify-end gap-2">
        {canDiscard && (
          <Button type="button" variant="outline" disabled={busy} onClick={onDiscard}>
            {t("desk.proposal.discard")}
          </Button>
        )}
        {canRegenerate ? (
          <Button type="button" variant="outline" disabled={busy} onClick={onRegenerate}>
            {t("desk.proposal.regenerate")}
          </Button>
        ) : (
          <p className="text-xs text-muted">{t("desk.proposal.regenerateAssignedOnly")}</p>
        )}
      </div>
    </div>
  );
}
