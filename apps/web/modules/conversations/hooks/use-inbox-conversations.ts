"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type Conversation } from "../services/messaging";
import type { DeskFilter } from "../types";

/** One page of desk conversations for a filter, reloaded whenever `refresh` changes. */
export function useInboxConversations(tenantID: string, refresh: number) {
  const errorText = useErrorText();
  const [filter, setFilterState] = useState<DeskFilter>("all");
  const [cursor, setCursor] = useState("");
  const [next, setNext] = useState("");
  const [rows, setRows] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const result = await messaging.list(tenantID, filter, cursor, controller.signal);
        if (!controller.signal.aborted) {
          setRows(result.items);
          setNext(result.next_cursor);
        }
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, filter, cursor, refresh, errorText]);
  /** Changing the filter always returns to the first page. */
  function setFilter(value: DeskFilter) {
    setFilterState(value);
    setCursor("");
  }
  return {
    rows,
    loading,
    error,
    filter,
    setFilter,
    hasPrevious: !!cursor,
    hasNext: !!next,
    firstPage: () => setCursor(""),
    nextPage: () => setCursor(next),
  };
}
