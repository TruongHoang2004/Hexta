"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { workflows } from "../services/workflows";
import type { WorkItem, WorkView } from "../types";

/** Paged work items for one tab. Reloads on `revision`; load errors go to `onError`. */
export function useWorkList(tenantID: string, revision: number, onError: (message: string) => void) {
  const errorText = useErrorText();
  const [view, setViewState] = useState<WorkView>("mine");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<WorkItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const data = await workflows.items(tenantID, page, abort.signal, view);
        if (!abort.signal.aborted) {
          setRows(data.items);
          setTotal(data.total);
        }
      } catch (err) {
        if (!abort.signal.aborted) onError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, page, view, revision, errorText, onError]);

  function setView(value: WorkView) {
    setViewState(value);
    setPage(1);
  }

  return { view, setView, page, setPage, rows, total, loading };
}
