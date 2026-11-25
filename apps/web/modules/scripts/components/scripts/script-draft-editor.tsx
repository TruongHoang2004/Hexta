"use client";

import { ConflictBanner, Notice, StatusChip, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useScriptEditor } from "../../hooks/use-script-editor";
import type { Script } from "../../types";
import { draftNumber, openDraft } from "../../utils";
import { ScriptContentFields } from "./script-content-fields";

interface Props {
  /** null creates a new script. */
  script: Script | null;
  canActivate: boolean;
  busy: boolean;
  onSaved: (script: Script) => void;
  onActivate: (versionID: string) => void;
  onReload: () => void;
}

/**
 * Draft card (S-15). Saving an existing script writes its open draft, or opens the next version when the
 * active content must stay unchanged. Activation applies only to a saved draft.
 */
export function ScriptDraftEditor({ script, canActivate, busy, onSaved, onActivate, onReload }: Props) {
  const t = useT();
  const { tenant } = useWorkspace();
  const editor = useScriptEditor(tenant?.id ?? "", script, onSaved);
  const draft = script ? openDraft(script) : undefined;
  const number = script ? draftNumber(script) : 1;
  const locked = editor.busy || busy;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="font-semibold">
          {script ? t("scripts.editor.draftTitle", { number }) : t("scripts.editor.newTitle")}
        </h2>
        {(editor.dirty || (script && !draft)) && <StatusChip tone="warning">{t("scripts.editor.unsaved")}</StatusChip>}
      </div>
      {!script && <p className="text-sm text-muted">{t("scripts.editor.newBody")}</p>}
      {editor.conflict ? (
        <ConflictBanner message={t("scripts.detail.conflict")} actionLabel={t("scripts.detail.reload")} onReload={onReload} />
      ) : (
        <Notice error={editor.error} />
      )}
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          void editor.save();
        }}
      >
        <ScriptContentFields form={editor.form} onChange={editor.update} disabled={locked} />
        <div className="flex flex-wrap items-center gap-2">
          <button type="submit" className={script ? secondaryClass : buttonClass} disabled={locked || !editor.valid || (!!script && !editor.dirty)}>
            {editor.busy ? t("scripts.editor.saving") : script ? t("scripts.editor.save") : t("scripts.editor.createAction")}
          </button>
          {script && canActivate && draft && (
            <button
              type="button"
              className={buttonClass}
              disabled={locked || editor.dirty}
              onClick={() => onActivate(draft.id)}
            >
              {t("scripts.editor.activate", { number: draft.number })}
            </button>
          )}
          {script && canActivate && draft && editor.dirty && (
            <span className="text-xs text-muted">{t("scripts.editor.activateHint")}</span>
          )}
        </div>
      </form>
    </section>
  );
}
