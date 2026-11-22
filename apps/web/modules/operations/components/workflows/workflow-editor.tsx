"use client";

import { Field, Notice, buttonClass, inputClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useWorkflowEditor } from "../../hooks/use-workflow-editor";
import type { WorkflowDefinition } from "../../types";
import { WorkflowStepRow } from "./workflow-step-row";

interface Props {
  definition: WorkflowDefinition | null;
  onSaved: (row: WorkflowDefinition) => void;
}

/** Create a workflow or publish its next version (up to 30 steps). */
export function WorkflowEditor({ definition, onSaved }: Props) {
  const { tenant } = useWorkspace();
  const t = useT();
  const editor = useWorkflowEditor(tenant?.id ?? "", definition, onSaved);
  const { steps, busy } = editor;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <h2 className="text-lg font-semibold">
        {definition ? t("workflows.editor.editTitle", { version: definition.version }) : t("workflows.editor.newTitle")}
      </h2>
      <p className="rounded-lg bg-background p-3 text-sm">{t("workflows.editor.body")}</p>
      <Notice error={editor.error} />
      <form
        onSubmit={(event) => {
          event.preventDefault();
          void editor.publish();
        }}
        className="space-y-4"
      >
        <fieldset disabled={busy} className="space-y-4">
          <Field title={t("workflows.editor.name")}>
            <input
              className={inputClass}
              required
              maxLength={120}
              value={editor.name}
              onChange={(e) => editor.setName(e.target.value)}
            />
          </Field>
          <ol className="space-y-3">
            {steps.map((step, index) => (
              <WorkflowStepRow
                key={index}
                step={step}
                index={index}
                count={steps.length}
                onChange={(patch) => editor.update(index, patch)}
                onMove={(direction) => editor.move(index, direction)}
                onRemove={() => editor.remove(index)}
              />
            ))}
          </ol>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <button type="button" className={secondaryClass} disabled={steps.length >= 30} onClick={editor.add}>
              {t("workflows.editor.add")}
            </button>
            <span className="text-sm text-muted">{t("workflows.editor.count", { count: steps.length })}</span>
          </div>
          <button className={buttonClass} type="submit">
            {busy
              ? t("workflows.editor.publishing")
              : definition
                ? t("workflows.editor.publishNew", { version: definition.version + 1 })
                : t("workflows.editor.publish")}
          </button>
        </fieldset>
      </form>
      <div className="space-y-1 border-t border-border pt-4 text-sm">
        <h3 className="font-semibold">{t("workflows.editor.explainTitle")}</h3>
        <p>{t("workflows.editor.explainDirect")}</p>
        <p>{t("workflows.editor.explainAcceptance")}</p>
      </div>
    </section>
  );
}
