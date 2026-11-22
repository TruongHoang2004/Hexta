"use client";

import { useState } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { workflows } from "../services/workflows";
import type { WorkCommand, WorkCommandFields, WorkInput, WorkItem } from "../types";

/**
 * The selected work item and every mutation on it (open, create, commands). Shares one
 * error/conflict/success channel with the list; `onChanged` reloads list and lookups.
 */
export function useWorkItem(tenantID: string, onChanged: () => void) {
  const t = useT();
  const errorText = useErrorText();
  const [selected, setSelected] = useState<WorkItem | null>(null);
  const [creatorKey, setCreatorKey] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [success, setSuccess] = useState("");

  /** Runs one mutation; returns whether it succeeded. */
  async function run(action: () => Promise<void>): Promise<boolean> {
    if (!tenantID || busy) return false;
    setBusy(true);
    setError("");
    setConflict(false);
    setSuccess("");
    try {
      await action();
      return true;
    } catch (err) {
      setConflict(isConflict(err));
      setError(errorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  }

  const open = (id: string) =>
    run(async () => {
      setSelected(await workflows.item(tenantID, id));
    });

  const create = (input: WorkInput, key: string) =>
    run(async () => {
      setSelected(await workflows.create(tenantID, input, key));
      setCreatorKey((v) => v + 1);
      onChanged();
      setSuccess(t("work.created"));
    });

  function command(action: WorkCommand, fields?: WorkCommandFields) {
    const item = selected;
    if (!item) return Promise.resolve(false);
    return run(async () => {
      setSelected(await workflows.command(tenantID, item, action, fields));
      onChanged();
      setSuccess(t("work.saved"));
    });
  }

  return { selected, creatorKey, busy, error, setError, conflict, success, open, create, command };
}
