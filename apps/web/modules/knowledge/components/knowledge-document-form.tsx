"use client";

import { Button, Input, Textarea } from "@hexta/ui";
import { Field } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { knowledgeLimits, type KnowledgeDocumentInput } from "../types";

interface Props {
  creating: boolean;
  form: KnowledgeDocumentInput;
  valid: boolean;
  busy: boolean;
  onChange: (patch: Partial<KnowledgeDocumentInput>) => void;
  onSave: () => void;
  onCancel: () => void;
}

/** Owner form for a new or edited document: a title and the plain-text body the bot may quote from. */
export function KnowledgeDocumentForm({ creating, form, valid, busy, onChange, onSave, onCancel }: Props) {
  const t = useT();
  return (
    <form
      aria-labelledby="knowledge-form"
      className="space-y-4 rounded-xl border border-border bg-surface p-5"
      onSubmit={(event) => {
        event.preventDefault();
        if (valid) onSave();
      }}
    >
      <h3 id="knowledge-form" className="font-semibold">
        {creating ? t("knowledge.form.createTitle") : t("knowledge.form.editTitle")}
      </h3>
      <Field title={t("knowledge.form.title")}>
        <Input
          value={form.title}
          maxLength={knowledgeLimits.title}
          disabled={busy}
          required
          onChange={(event) => onChange({ title: event.target.value })}
        />
      </Field>
      <Field title={t("knowledge.form.body")}>
        <Textarea
          className="min-h-48"
          value={form.body}
          maxLength={knowledgeLimits.body}
          disabled={busy}
          required
          aria-describedby="knowledge-body-hint"
          onChange={(event) => onChange({ body: event.target.value })}
        />
        <span id="knowledge-body-hint" className="block text-xs font-normal text-muted">
          {t("knowledge.form.bodyHint", { count: form.body.length, max: knowledgeLimits.body })}
        </span>
      </Field>
      <div className="flex flex-wrap gap-2">
        <Button type="submit" disabled={busy || !valid}>
          {busy ? t("knowledge.form.saving") : t("common.save")}
        </Button>
        <Button type="button" variant="outline" disabled={busy} onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </div>
    </form>
  );
}
