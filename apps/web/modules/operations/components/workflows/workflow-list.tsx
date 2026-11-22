"use client";

import { EmptyState, Pager } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { useWorkflowList } from "../../hooks/use-workflow-list";
import type { WorkflowDefinition } from "../../types";
import { WorkflowCard } from "./workflow-card";

interface Props {
  list: ReturnType<typeof useWorkflowList>;
  selectedID: string | undefined;
  canEdit: boolean;
  onEdit: (row: WorkflowDefinition) => void;
}

export function WorkflowList({ list, selectedID, canEdit, onEdit }: Props) {
  const t = useT();
  return (
    <section className="min-w-0 flex-[1_1_320px] space-y-4 rounded-xl border border-border bg-surface p-5">
      <h2 className="font-semibold">{t("workflows.list")}</h2>
      {list.loading ? (
        <p role="status">{t("workflows.loading")}</p>
      ) : !list.rows.length ? (
        <EmptyState title={t("workflows.emptyTitle")} description={t("workflows.emptyBody")} />
      ) : (
        list.rows.map((row) => (
          <WorkflowCard
            key={row.id}
            definition={row}
            selected={selectedID === row.id}
            onEdit={canEdit ? () => onEdit(row) : undefined}
          />
        ))
      )}
      <Pager page={list.page} total={list.total} onChange={list.setPage} disabled={list.loading} />
    </section>
  );
}
