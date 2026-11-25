"use client";

import { EmptyState, Pager } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { useKnowledgeDocuments } from "../hooks/use-knowledge-documents";
import type { KnowledgeDocument } from "../types";
import { KnowledgeDocumentItem } from "./knowledge-document-item";

interface Props {
  list: ReturnType<typeof useKnowledgeDocuments>;
  busy: boolean;
  onEdit: (document: KnowledgeDocument) => void;
  onDelete: (document: KnowledgeDocument) => void;
}

/** The tenant's documents with loading, error and empty states. */
export function KnowledgeDocumentList({ list, busy, onEdit, onDelete }: Props) {
  const t = useT();
  return (
    <section aria-labelledby="knowledge-list" className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <h3 id="knowledge-list" className="font-semibold">
        {t("knowledge.list")}
      </h3>
      {list.error && (
        <p role="alert" className="text-sm text-error">
          {list.error}
        </p>
      )}
      {list.loading && !list.rows.length ? (
        <p role="status">{t("knowledge.loading")}</p>
      ) : !list.rows.length ? (
        !list.error && (
          <EmptyState
            title={t("knowledge.emptyTitle")}
            description={list.canManage ? t("knowledge.emptyOwner") : t("knowledge.emptyMember")}
          />
        )
      ) : (
        <ul className="space-y-3">
          {list.rows.map((row) => (
            <KnowledgeDocumentItem
              key={row.id}
              document={row}
              canManage={list.canManage}
              busy={busy}
              onEdit={() => onEdit(row)}
              onDelete={() => onDelete(row)}
            />
          ))}
        </ul>
      )}
      <Pager page={list.page} total={list.total} onChange={list.setPage} disabled={list.loading} />
    </section>
  );
}
