"use client";

import { useEffect, useState } from "react";
import { listAllConnections, type ChannelConnection } from "@/modules/settings";

/**
 * Channel connections for the expired banner and composer lock; null while unknown.
 * A failure here must not hide the inbox, so errors are ignored.
 */
export function useChannelConnections(tenantID: string, refresh: number) {
  const [connections, setConnections] = useState<ChannelConnection[] | null>(null);
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    listAllConnections(tenantID, controller.signal)
      .then((rows) => {
        if (!controller.signal.aborted) setConnections(rows);
      })
      .catch(() => undefined);
    return () => controller.abort();
  }, [tenantID, refresh]);
  return connections;
}
