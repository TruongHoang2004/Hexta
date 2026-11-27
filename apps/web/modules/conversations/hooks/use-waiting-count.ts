"use client";

import { useEffect, useState } from "react";
import { messaging } from "../services/messaging";

/** How often the desk re-reads the waiting-for-staff count while the tab is visible. */
const POLL_MS = 30_000;

/**
 * Tenant-wide count of conversations waiting for staff, polled while the page is visible and reloaded whenever
 * `refresh` changes. `null` while unknown; a failed poll keeps the last known count.
 */
export function useWaitingCount(tenantID: string, refresh: number) {
  const [count, setCount] = useState<number | null>(null);
  useEffect(() => {
    if (!tenantID) return;
    let controller: AbortController | null = null;
    async function load() {
      if (document.visibilityState === "hidden") return;
      controller?.abort();
      const current = new AbortController();
      controller = current;
      try {
        const summary = await messaging.summary(tenantID, current.signal);
        if (!current.signal.aborted) setCount(summary.waiting_staff_count);
      } catch {
        // Polling is best effort; the list and its errors stay authoritative.
      }
    }
    void load();
    const timer = window.setInterval(() => void load(), POLL_MS);
    const onVisible = () => {
      if (document.visibilityState === "visible") void load();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", onVisible);
      controller?.abort();
    };
  }, [tenantID, refresh]);
  return count;
}
