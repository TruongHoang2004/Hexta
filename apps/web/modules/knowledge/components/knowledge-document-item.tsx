"use client";

import { Button } from "@hexta/ui";
import { useLocale, useT } from "@/lib/i18n/provider";
import type { KnowledgeDocument } from "../types";
import { KnowledgeSyncChip } from "./knowledge-sync-chip";

interface Props {
  document: KnowledgeDocument;
  /** Edit and delete are shown only to the owner; the API authorizes them again. */
  canManage: boolean;
  busy: boolean;
  onEdit: () => void;
  onDelete: () => void;
}

/** One document: title, index state, last change and a short preview of the body. */
export function KnowledgeDocumentItem({ document, canManage, busy, onEdit, onDelete }: Props) {
  const t = useT();
  const locale = useLocale();
  const headingID = `knowledge-${document.id}`;
  return (
    <li aria-labelledby={headingID} className="space-y-2 rounded-lg border border-border bg-background p-4">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0 space-y-1">
          <h3 id={headingID} className="break-words font-semibold">
            {document.title}
          </h3>
          <p className="text-xs text-muted">
            {t("knowledge.updatedAt", { date: new Date(document.updated_at).toLocaleString(locale) })}
          </p>
        </div>
        <KnowledgeSyncChip state={document.sync_state} error={document.sync_error} />
      </div>
      <p className="line-clamp-3 whitespace-pre-line break-words text-sm">{document.body}</p>
      {canManage && (
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" size="sm" disabled={busy} onClick={onEdit}>
            {t("knowledge.edit")}
          </Button>
          <Button type="button" variant="outline" size="sm" disabled={busy} onClick={onDelete}>
            {t("knowledge.delete")}
          </Button>
        </div>
      )}
    </li>
  );
}
