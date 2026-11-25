"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ProposalCitation } from "../../services/messaging";

/** Short fallback for a removed source: the part after "kind:" cut to 8 characters. */
function shortID(sourceID: string): string {
  const id = sourceID.includes(":") ? sourceID.slice(sourceID.indexOf(":") + 1) : sourceID;
  return id.length > 8 ? `${id.slice(0, 8)}…` : id;
}

/** The knowledge records an AI-written answer relies on, so staff can check them before sending. */
export function ProposalCitations({ citations }: { citations: ProposalCitation[] }) {
  const t = useT();
  if (!citations.length) return null;
  return (
    <div className="flex flex-wrap items-center gap-1.5 text-xs">
      <span className="text-muted">{t("desk.proposal.citations")}</span>
      <ul className="flex flex-wrap gap-1.5" aria-label={t("desk.proposal.citations")}>
        {citations.map((citation) => (
          <li key={citation.source_id}>
            <StatusChip>
              {t(`desk.proposal.citationKinds.${citation.kind}`)}: {citation.label || shortID(citation.source_id)}
            </StatusChip>
          </li>
        ))}
      </ul>
    </div>
  );
}
