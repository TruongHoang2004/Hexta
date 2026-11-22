"use client";

import { Field, SegmentedFilter, inputClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { HandoffMode, WorkflowStepInput } from "../../types";

interface Props {
  step: WorkflowStepInput;
  index: number;
  count: number;
  onChange: (patch: Partial<WorkflowStepInput>) => void;
  onMove: (direction: -1 | 1) => void;
  onRemove: () => void;
}

/** One step: name, handoff mode, reorder and remove. */
export function WorkflowStepRow({ step, index, count, onChange, onMove, onRemove }: Props) {
  const t = useT();
  const modes: { value: HandoffMode; label: string }[] = [
    { value: "direct", label: t("workflows.editor.modes.direct") },
    { value: "acceptance", label: t("workflows.editor.modes.acceptance") },
  ];
  return (
    <li className="flex flex-wrap items-end gap-3 rounded-lg border border-border bg-background p-3">
      <span
        aria-hidden="true"
        className="inline-flex h-8 w-8 flex-none items-center justify-center rounded-full bg-surface font-semibold"
      >
        {index + 1}
      </span>
      <div className="min-w-0 flex-[1_1_200px]">
        <Field title={t("workflows.editor.step", { index: index + 1 })}>
          <input
            className={inputClass}
            required
            maxLength={120}
            value={step.name}
            onChange={(e) => onChange({ name: e.target.value })}
          />
        </Field>
      </div>
      <SegmentedFilter
        label={t("workflows.editor.mode", { index: index + 1 })}
        options={modes}
        value={step.handoff_mode}
        onChange={(value) => onChange({ handoff_mode: value })}
      />
      <div className="flex gap-1">
        <button
          type="button"
          className={secondaryClass}
          aria-label={t("workflows.editor.moveUpLabel", { index: index + 1 })}
          disabled={index === 0}
          onClick={() => onMove(-1)}
        >
          {t("workflows.editor.moveUp")}
        </button>
        <button
          type="button"
          className={secondaryClass}
          aria-label={t("workflows.editor.moveDownLabel", { index: index + 1 })}
          disabled={index === count - 1}
          onClick={() => onMove(1)}
        >
          {t("workflows.editor.moveDown")}
        </button>
        <button
          type="button"
          className={secondaryClass}
          aria-label={t("workflows.editor.remove", { index: index + 1 })}
          disabled={count === 1}
          onClick={onRemove}
        >
          ×
        </button>
      </div>
      {index === count - 1 && <p className="w-full text-xs text-muted">{t("workflows.editor.lastStepNote")}</p>}
    </li>
  );
}
