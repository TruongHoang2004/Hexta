"use client";

import { StatusChip, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { WorkflowDefinition } from "../../types";

interface Props {
  definition: WorkflowDefinition;
  selected: boolean;
  /** Omitted for members, who can only view. */
  onEdit?: () => void;
}

/** One workflow with its latest version's steps. */
export function WorkflowCard({ definition, selected, onEdit }: Props) {
  const t = useT();
  return (
    <article className={`space-y-2 rounded-lg border p-3 ${selected ? "border-foreground" : "border-border"}`}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-medium">
          {definition.name} <StatusChip>{t("workflows.versionLabel", { number: definition.version })}</StatusChip>
        </h3>
        {onEdit && (
          <button type="button" className={secondaryClass} onClick={onEdit}>
            {t("workflows.edit")}
          </button>
        )}
      </div>
      <ol className="list-decimal pl-5 text-sm text-muted">
        {definition.versions[0]?.steps.map((step, i, all) => (
          <li key={step.id}>
            {step.name} ·{" "}
            {i < all.length - 1 ? t(`workflows.modeShort.${step.handoff_mode}`) : t("workflows.finalStep")}
          </li>
        ))}
      </ol>
    </article>
  );
}
