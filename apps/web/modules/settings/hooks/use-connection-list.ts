"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";

/**
 * Shared list + mutation state for one channel platform. `busy` is "" when idle, "connect" while
 * connecting, or the ID of the row being changed. Every mutation reloads the list afterwards,
 * because a failed connector sync is still saved on the row.
 */
export function useConnectionList<T>(tenantID: string, load: (tenant: string, signal: AbortSignal) => Promise<T[]>) {
  const errorText = useErrorText();
  const [rows, setRows] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function run() {
      setLoading(true);
      try {
        const result = await load(tenantID, controller.signal);
        if (!controller.signal.aborted) setRows(result);
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void run();
    return () => controller.abort();
  }, [tenantID, revision, errorText, load]);

  /** Runs one mutation under `busyKey`; resolves to whether it succeeded. */
  async function mutate(busyKey: string, action: () => Promise<string>): Promise<boolean> {
    setBusy(busyKey);
    setError("");
    setSuccess("");
    try {
      setSuccess(await action());
      return true;
    } catch (err) {
      setError(errorText(err));
      return false;
    } finally {
      setBusy("");
      setRevision((v) => v + 1);
    }
  }

  function refresh() {
    setError("");
    setRevision((v) => v + 1);
  }

  return { rows, loading, busy, setBusy, error, setError, success, mutate, refresh };
}
