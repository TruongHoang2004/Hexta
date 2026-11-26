"use client";

import { useEffect, useReducer, useState } from "react";
import { SDKApiError } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { overview } from "../services/overview";
import type { Overview, OverviewFilter } from "../types";
import { periodRange } from "../utils";

interface State {
  data: Overview | null;
  /** When `data` arrived; the reference time for "waiting" durations. */
  loadedAt: Date | null;
  loading: boolean;
  error: string;
  /** The API refused the caller (not the tenant owner, or the workspace is not active). */
  denied: boolean;
}

type Action =
  | { type: "start" }
  | { type: "done"; data: Overview; at: Date }
  | { type: "failed"; error: string; denied: boolean };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "start":
      return { ...state, loading: true, error: "" };
    case "done":
      return { data: action.data, loadedAt: action.at, loading: false, error: "", denied: false };
    case "failed":
      return action.denied
        ? { data: null, loadedAt: null, loading: false, error: "", denied: true }
        : { ...state, loading: false, error: action.error };
  }
}

/**
 * Loads the owner overview for the filter. Skipped when `enabled` is false (the UI already knows the
 * caller is not the owner); the API still authorizes every request. Older responses never replace newer ones.
 */
export function useOverview(tenantID: string, filter: OverviewFilter, enabled: boolean) {
  const errorText = useErrorText();
  const [revision, setRevision] = useState(0);
  const [state, dispatch] = useReducer(reducer, { data: null, loadedAt: null, loading: enabled, error: "", denied: false });

  useEffect(() => {
    if (!tenantID || !enabled) return;
    const abort = new AbortController();
    dispatch({ type: "start" });
    const query = { ...periodRange(filter.period, new Date()), ...(filter.employeeID ? { employee_id: filter.employeeID } : {}) };
    overview
      .read(tenantID, query, abort.signal)
      .then((data) => {
        if (!abort.signal.aborted) dispatch({ type: "done", data, at: new Date() });
      })
      .catch((err: unknown) => {
        if (abort.signal.aborted) return;
        const denied = err instanceof SDKApiError && err.status === 403;
        dispatch({ type: "failed", error: errorText(err), denied });
      });
    return () => abort.abort();
  }, [tenantID, filter.period, filter.employeeID, enabled, revision, errorText]);

  return { ...state, reload: () => setRevision((value) => value + 1) };
}
