"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ReplyProposal } from "../../services/messaging";
import type { ReplyMode } from "../../services/reply-mode";
import type { ProposalAccess } from "../../utils";
import { ProposalEditor } from "./proposal-editor";
import { ProposalStale } from "./proposal-stale";

/** S-07 proposal card: the bot's reply waiting for a sale, its source, and the routing or stale states. */
export function ProposalCard({
  proposal,
  mode,
  access,
  assigneeName,
  busy,
  error,
  uncertain,
  onSend,
  onDiscard,
  onRegenerate,
}: {
  proposal: ReplyProposal;
  mode: ReplyMode | null;
  access: ProposalAccess;
  assigneeName: string;
  busy: boolean;
  error: string;
  uncertain: boolean;
  onSend: (text?: string) => void;
  onDiscard: () => void;
  onRegenerate: () => void;
}) {
  const t = useT();
  const source =
    proposal.source === "script"
      ? proposal.script_name
        ? t("desk.proposal.sourceScript", { name: proposal.script_name, version: proposal.script_version ?? 0 })
        : t("desk.proposal.sourceScriptUnknown")
      : proposal.route === "clarify"
        ? t("desk.proposal.sourceClarify")
        : t("desk.proposal.sourceAI");
  const sendHint = !access.canSend
    ? assigneeName
      ? t("desk.proposal.assignedOnly", { name: assigneeName })
      : t("desk.proposal.claimFirst")
    : access.claimsOnSend
      ? t("desk.proposal.claimsOnSend")
      : "";
  return (
    <section
      aria-labelledby={`proposal-${proposal.id}`}
      className="space-y-2 rounded-xl border border-primary/40 bg-primary/5 p-3"
    >
      <header className="flex flex-wrap items-center justify-between gap-2 text-sm">
        <h3 id={`proposal-${proposal.id}`} className="flex flex-wrap items-center gap-2 font-semibold">
          <StatusChip tone={proposal.stale ? "warning" : "neutral"}>
            {proposal.stale ? t("desk.proposal.staleChip") : t("desk.proposal.chip")}
          </StatusChip>
          {proposal.status !== "routing" && <span className="font-normal text-muted">{t("desk.proposal.source", { source })}</span>}
        </h3>
        {mode && <span className="text-xs text-muted">{t("desk.proposal.mode", { mode: t(`replyMode.modes.${mode}.name`) })}</span>}
      </header>
      {error && (
        <p role="alert" className="text-sm text-error">
          {error}
        </p>
      )}
      {uncertain && (
        <p role="status" className="rounded-lg border border-warning bg-warning/15 p-2 text-sm">
          {t("desk.proposal.uncertain")}
        </p>
      )}
      {proposal.status === "routing" ? (
        <p role="status" className="text-sm text-muted">
          {t("desk.proposal.routing")}
        </p>
      ) : proposal.stale ? (
        <ProposalStale
          text={proposal.text ?? ""}
          canRegenerate={access.canRegenerate}
          canDiscard={access.canDiscard}
          busy={busy}
          onRegenerate={onRegenerate}
          onDiscard={onDiscard}
        />
      ) : (
        <ProposalEditor
          key={proposal.id}
          proposalID={proposal.id}
          text={proposal.text ?? ""}
          canSend={access.canSend}
          canDiscard={access.canDiscard}
          sendHint={sendHint}
          busy={busy}
          uncertain={uncertain}
          onSend={onSend}
          onDiscard={onDiscard}
        />
      )}
    </section>
  );
}
