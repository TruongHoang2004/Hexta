"use client";

import { useEffect, useRef, useState } from "react";
import { streamConversations } from "../services/conversation-stream";

export type StreamStatus = "connecting" | "live" | "reconnecting";

export interface StreamChange {
  /** Thread ids named by hints since the last flush. */
  threads: Set<string>;
  /** True when hints may have been missed (reconnect, resync, tab shown again): refetch everything. */
  resync: boolean;
}

const DEBOUNCE_MS = 300;
const MAX_BACKOFF_MS = 15000;

function wait(ms: number, signal: AbortSignal) {
  return new Promise<void>((resolve) => {
    const timer = setTimeout(resolve, ms);
    signal.addEventListener("abort", () => (clearTimeout(timer), resolve()), { once: true });
  });
}

/**
 * Keeps one server-sent event stream open for a tenant, reconnecting with backoff, and reports coalesced
 * changes. Streams are only a trigger; the caller refetches the data it shows.
 */
export function useConversationStream(tenantID: string, onChange: (change: StreamChange) => void) {
  const [status, setStatus] = useState<StreamStatus>("connecting");
  const handler = useRef(onChange);
  useEffect(() => {
    handler.current = onChange;
  });
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    let pending: StreamChange = { threads: new Set(), resync: false };
    let timer: ReturnType<typeof setTimeout> | undefined;
    const flush = () => {
      timer = undefined;
      const change = pending;
      pending = { threads: new Set(), resync: false };
      if (!controller.signal.aborted) handler.current(change);
    };
    const queue = (threadID: string | null) => {
      if (threadID) pending.threads.add(threadID);
      else pending.resync = true;
      timer ??= setTimeout(flush, DEBOUNCE_MS);
    };
    const onVisible = () => {
      if (document.visibilityState === "visible") queue(null);
    };
    document.addEventListener("visibilitychange", onVisible);

    async function run() {
      let delay = 1000;
      let first = true;
      while (!controller.signal.aborted) {
        let graceful = false;
        try {
          await streamConversations(tenantID, controller.signal, (message) => {
            if (message.type === "ready") {
              delay = 1000;
              setStatus("live");
              // Anything committed while disconnected produced no hint, so every reconnect resyncs.
              if (!first) queue(null);
              first = false;
            } else if (message.type === "changed") queue(message.hint.thread_id);
            else queue(null);
          });
          // The server ends streams periodically so authorization is rechecked; reconnect straight away.
          graceful = true;
        } catch {
          // Network loss, auth refresh in progress or server shutdown: retry below.
        }
        if (controller.signal.aborted) break;
        if (graceful) {
          await wait(250, controller.signal);
          continue;
        }
        setStatus("reconnecting");
        await wait(delay, controller.signal);
        delay = Math.min(delay * 2, MAX_BACKOFF_MS);
      }
    }
    void run();
    return () => {
      controller.abort();
      document.removeEventListener("visibilitychange", onVisible);
      if (timer) clearTimeout(timer);
    };
  }, [tenantID]);
  return status;
}
