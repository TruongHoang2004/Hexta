"use client";

import { useEffect, useRef, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type Conversation } from "../services/messaging";
import type { DeskFilter } from "../types";

/**
 * One page of desk conversations for a filter, reloaded whenever `refresh` changes. A change of `live`
 * (a realtime hint) reloads quietly, without the loading state, and only on the first page so paging is stable.
 */
export function useInboxConversations(tenantID: string, refresh: number, live: number) {
  const errorText = useErrorText();
  const [filter, setFilterState] = useState<DeskFilter>("all");
  const [cursor, setCursor] = useState("");
  const [next, setNext] = useState("");
  const [rows, setRows] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const seenLive = useRef(live);
  useEffect(() => {
    if (!tenantID) return;
    const silent = seenLive.current !== live;
    seenLive.current = live;
    if (silent && cursor) return;
    const controller = new AbortController();
    async function load() {
      if (!silent) {
        setLoading(true);
        setError("");
      }
      try {
        const result = await messaging.list(tenantID, filter, cursor, controller.signal);
        if (!controller.signal.aborted) {
          setRows(result.items);
          setNext(result.next_cursor);
        }
      } catch (err) {
        if (!controller.signal.aborted && !silent) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted && !silent) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, filter, cursor, refresh, live, errorText]);
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
