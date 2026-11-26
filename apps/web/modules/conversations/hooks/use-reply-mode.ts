"use client";

import { useEffect, useState } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { replyMode, type ReplyMode, type ReplySettings } from "../services/reply-mode";

/** Loads the tenant's reply mode and, for the owner, saves a change with its optimistic version. */
export function useReplyMode(tenantID: string) {
  const errorText = useErrorText();
  const [settings, setSettings] = useState<ReplySettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [saved, setSaved] = useState<ReplyMode | null>(null);
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const next = await replyMode.get(tenantID, abort.signal);
        if (abort.signal.aborted) return;
        setSettings(next);
        setError("");
        setConflict(false);
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, revision, errorText]);

  async function save(mode: ReplyMode): Promise<boolean> {
    if (!settings || saving) return false;
    setSaving(true);
    setError("");
    setSaved(null);
    try {
      setSettings(await replyMode.update(tenantID, mode, settings.version));
      setSaved(mode);
      return true;
    } catch (err) {
      setConflict(isConflict(err));
      if (!isConflict(err)) setError(errorText(err));
      return false;
    } finally {
      setSaving(false);
    }
  }

  return {
    settings,
    loading,
    saving,
    error,
    conflict,
    saved,
    save,
    reload: () => {
      setSaved(null);
      setRevision((v) => v + 1);
    },
  };
}
