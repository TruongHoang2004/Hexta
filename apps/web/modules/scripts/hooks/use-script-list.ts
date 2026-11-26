"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { scripts } from "../services/scripts";
import type { ScriptStatusFilter, ScriptSummary } from "../types";

/** Paged scripts for one status filter. Reloads on `revision`; a filter change returns to page 1. */
export function useScriptList(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [status, setStatusState] = useState<ScriptStatusFilter>("all");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<ScriptSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const data = await scripts.list(tenantID, status, page, abort.signal);
        if (!abort.signal.aborted) {
          setRows(data.items);
          setTotal(data.total);
        }
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, status, page, revision, errorText]);

  const setStatus = (next: ScriptStatusFilter) => {
    setStatusState(next);
    setPage(1);
  };
  return { status, setStatus, page, setPage, rows, total, loading, error };
}
