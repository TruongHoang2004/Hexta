"use client";

import { useEffect, useState } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { scripts } from "../services/scripts";
import type { Script } from "../types";

type Command = { kind: "activate"; versionID: string } | { kind: "deactivate" } | { kind: "archive" };

/**
 * One script with its history, plus the lifecycle commands. Every command sends the script version it was shown;
 * a stale version comes back as a conflict that asks the user to reload.
 */
export function useScriptDetail(tenantID: string, scriptID: string, onChanged: (script: Script) => void) {
  const errorText = useErrorText();
  const [script, setScript] = useState<Script | null>(null);
  const [reload, setReload] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    if (!tenantID || !scriptID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const result = await scripts.get(tenantID, scriptID, abort.signal);
        if (!abort.signal.aborted) {
          setScript(result);
          setError("");
          setConflict(false);
        }
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, scriptID, reload, errorText]);

  /** Applies a server result (from a command or the draft editor) as the current state. */
  function accept(next: Script) {
    setScript(next);
    setConflict(false);
    setError("");
    onChanged(next);
  }

  async function run(command: Command): Promise<boolean> {
    if (!script || busy) return false;
    setBusy(true);
    setError("");
    try {
      const next =
        command.kind === "activate"
          ? await scripts.activate(tenantID, script, command.versionID)
          : await scripts.transition(tenantID, script, command.kind);
      accept(next);
      return true;
    } catch (err) {
      if (isConflict(err)) setConflict(true);
      setError(errorText(err));
      onChanged(script);
      return false;
    } finally {
      setBusy(false);
    }
  }

  return {
    script,
    loading,
    busy,
    error,
    conflict,
    accept,
    reload: () => setReload((v) => v + 1),
    activate: (versionID: string) => run({ kind: "activate", versionID }),
    deactivate: () => run({ kind: "deactivate" }),
    archive: () => run({ kind: "archive" }),
  };
}
