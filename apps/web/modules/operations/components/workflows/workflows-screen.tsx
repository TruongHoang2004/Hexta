"use client";

import { useState } from "react";
import { Notice, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useWorkflowList } from "../../hooks/use-workflow-list";
import type { WorkflowDefinition } from "../../types";
import { WorkflowEditor } from "./workflow-editor";
import { WorkflowList } from "./workflow-list";

/** Workflow definitions; only owners can publish new versions. */
export function WorkflowsScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const [revision, setRevision] = useState(0);
  const [selected, setSelected] = useState<WorkflowDefinition | null>(null);
  const [success, setSuccess] = useState("");
  const list = useWorkflowList(tenant?.id ?? "", revision);
  const select = (row: WorkflowDefinition | null) => {
    setSelected(row);
    setSuccess("");
  };
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("workflows.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("workflows.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className={secondaryClass}
            onClick={() => {
              setSelected(null);
              setRevision((v) => v + 1);
            }}
          >
            {t("workflows.refresh")}
          </button>
          {isOwner && (
            <button type="button" className={buttonClass} onClick={() => select(null)}>
              {t("workflows.new")}
            </button>
          )}
        </div>
      </div>
      <Notice error={list.error} success={success} />
      {!isOwner && <p className="text-sm text-muted">{t("workflows.ownerOnly")}</p>}
      <div className="flex flex-wrap items-start gap-5">
        <WorkflowList list={list} selectedID={selected?.id} canEdit={isOwner} onEdit={select} />
        {isOwner && (
          <div className="min-w-0 flex-[999_1_480px]">
            <WorkflowEditor
              key={selected ? `${selected.id}:${selected.version}` : `new:${revision}`}
              definition={selected}
              onSaved={(row) => {
                setSelected(row);
                setRevision((v) => v + 1);
                setSuccess(t("workflows.published", { name: row.name, version: row.version }));
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
