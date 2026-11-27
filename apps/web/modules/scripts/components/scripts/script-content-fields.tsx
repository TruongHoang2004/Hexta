"use client";

import { Textarea } from "@hexta/ui";
import { Field, inputClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ScriptForm } from "../../hooks/use-script-editor";
import { scriptLimits } from "../../types";

interface Props {
  form: ScriptForm;
  onChange: (patch: Partial<ScriptForm>) => void;
  disabled: boolean;
}

/** Name, trigger phrases and answer of a script draft. */
export function ScriptContentFields({ form, onChange, disabled }: Props) {
  const t = useT();
  return (
    <div className="space-y-4">
      <Field title={t("scripts.editor.name")}>
        <input
          className={inputClass}
          value={form.name}
          maxLength={scriptLimits.name}
          disabled={disabled}
          required
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </Field>
      <Field title={t("scripts.editor.triggers")}>
        <input
          className={inputClass}
          value={form.triggers}
          disabled={disabled}
          required
          aria-describedby="script-triggers-hint"
          onChange={(e) => onChange({ triggers: e.target.value })}
        />
        <span id="script-triggers-hint" className="block text-xs font-normal text-muted">
          {t("scripts.editor.triggersHint")}
        </span>
      </Field>
      <Field title={t("scripts.editor.answer")}>
        <Textarea
          className="min-h-32"
          value={form.answer}
          maxLength={scriptLimits.answer}
          disabled={disabled}
          required
          aria-describedby="script-answer-hint"
          onChange={(e) => onChange({ answer: e.target.value })}
        />
        <span id="script-answer-hint" className="block text-xs font-normal text-muted">
          {t("scripts.editor.answerHint", { count: form.answer.length, max: scriptLimits.answer })}
        </span>
      </Field>
    </div>
  );
}
