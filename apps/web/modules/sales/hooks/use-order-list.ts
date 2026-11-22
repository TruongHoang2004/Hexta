"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { sales } from "../services/sales";
import type { Order, OrderListView } from "../types";
import { viewFilters } from "../utils";

/** Paged, filtered order list. Reloads when `revision` changes. */
export function useOrderList(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [view, setViewState] = useState<OrderListView>("all");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<Order[]>([]);
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
        const data = await sales.orders(tenantID, page, viewFilters[view], abort.signal);
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
  }, [tenantID, page, view, revision, errorText]);

  function setView(value: OrderListView) {
    setViewState(value);
    setPage(1);
  }

  return { view, setView, page, setPage, rows, total, loading, error };
}
