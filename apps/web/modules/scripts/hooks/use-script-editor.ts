"use client";

import { useState } from "react";
import { useRetryKey } from "@/modules/operations";
import { isConflict, useErrorText } from "@/lib/errors";
import { scripts } from "../services/scripts";
import type { Script, ScriptDraftInput } from "../types";
import { editorSource, formatTriggers, parseTriggers } from "../utils";

export interface ScriptForm {
  name: string;
  triggers: string;
  answer: string;
}

function formFrom(script: Script | null): ScriptForm {
  const source = script ? editorSource(script) : undefined;
  return { name: script?.name ?? "", triggers: formatTriggers(source?.triggers ?? []), answer: source?.answer ?? "" };
}

/**
 * Draft form for a new script (create) or an existing one (save draft). Saving an active script opens a new draft
 * version on the server; published content is never edited. Remount with a new key to reset after a save.
 */
export function useScriptEditor(tenantID: string, script: Script | null, onSaved: (script: Script) => void) {
  const errorText = useErrorText();
  const retryKey = useRetryKey();
  const [initial] = useState(() => formFrom(script));
  const [form, setForm] = useState<ScriptForm>(initial);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);

  const input: ScriptDraftInput = { name: form.name.trim(), triggers: parseTriggers(form.triggers), answer: form.answer.trim() };
  const dirty = form.name !== initial.name || form.triggers !== initial.triggers || form.answer !== initial.answer;
  const valid = input.name !== "" && input.answer !== "" && input.triggers.length > 0 && input.triggers.length <= 20;

  const update = (patch: Partial<ScriptForm>) => setForm((current) => ({ ...current, ...patch }));

  async function save() {
    if (!tenantID || busy || !valid) return;
    setBusy(true);
    setError("");
    try {
      const saved = script ? await scripts.saveDraft(tenantID, script, input) : await scripts.create(tenantID, input, retryKey(input));
      onSaved(saved);
    } catch (err) {
      setConflict(isConflict(err) && script !== null);
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  return { form, update, input, dirty, valid, busy, error, conflict, save };
}
