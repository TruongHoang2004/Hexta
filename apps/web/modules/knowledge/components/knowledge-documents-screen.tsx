"use client";

import { useState } from "react";
import { Button } from "@hexta/ui";
import { ConfirmDialog, ConflictBanner, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useKnowledgeDocuments } from "../hooks/use-knowledge-documents";
import { useKnowledgeEditor } from "../hooks/use-knowledge-editor";
import { KnowledgeDocumentForm } from "./knowledge-document-form";
import { KnowledgeDocumentList } from "./knowledge-document-list";

/**
 * Settings › Documents: the business FAQ and policy texts the bot may answer from. Every member reads them; only
 * the owner creates, edits and deletes. Answers built from them always wait for staff review.
 */
export function KnowledgeDocumentsScreen() {
  const { tenant } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [revision, setRevision] = useState(0);
  const list = useKnowledgeDocuments(tenantID, revision);
  const editor = useKnowledgeEditor(tenantID, () => setRevision((value) => value + 1));
  const loaded = !list.loading || list.rows.length > 0;

  return (
    <div className="max-w-3xl space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{t("knowledge.title")}</h2>
          <p className="mt-1 text-sm text-muted">{t("knowledge.intro")}</p>
          <p className="mt-1 text-sm text-muted">{t("knowledge.review")}</p>
        </div>
        {list.canManage && !editor.editing && (
          <Button type="button" disabled={editor.busy} onClick={() => editor.open("new")}>
            {t("knowledge.add")}
          </Button>
        )}
      </div>
      {editor.conflict ? (
        <ConflictBanner message={t("knowledge.conflict")} actionLabel={t("knowledge.reload")} onReload={editor.reload} />
      ) : (
        <Notice error={editor.error} success={editor.notice ? t(`knowledge.notice.${editor.notice}`) : ""} />
      )}
      {loaded && !list.canManage && !list.error && <p className="text-sm text-muted">{t("knowledge.ownerOnly")}</p>}
      {editor.editing && (
        <KnowledgeDocumentForm
          creating={editor.editing === "new"}
          form={editor.form}
          valid={editor.valid}
          busy={editor.busy}
          onChange={editor.change}
          onSave={editor.save}
          onCancel={editor.close}
        />
      )}
      <KnowledgeDocumentList list={list} busy={editor.busy} onEdit={editor.open} onDelete={editor.askDelete} />
      <ConfirmDialog
        open={editor.deleting !== null}
        title={t("knowledge.confirmDelete.title")}
        confirmLabel={t("knowledge.confirmDelete.action")}
        cancelLabel={t("common.cancel")}
        busy={editor.busy}
        onCancel={() => editor.askDelete(null)}
        onConfirm={editor.confirmDelete}
      >
        <p>{t("knowledge.confirmDelete.body", { title: editor.deleting?.title ?? "" })}</p>
      </ConfirmDialog>
    </div>
  );
}
