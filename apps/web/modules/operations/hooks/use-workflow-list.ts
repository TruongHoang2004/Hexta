"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { workflows } from "../services/workflows";
import type { WorkflowDefinition } from "../types";

/** Paged workflow definitions. Reloads on `revision`. */
export function useWorkflowList(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<WorkflowDefinition[]>([]);
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
        const data = await workflows.definitions(tenantID, page, abort.signal);
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
  }, [tenantID, page, revision, errorText]);

  return { page, setPage, rows, total, loading, error };
}
