"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { streamConversations, type StreamEvent } from "../services/conversation-stream";

export type StreamStatus = "connecting" | "live" | "reconnecting";
export type StreamListener = (event: StreamEvent) => void;

const DEBOUNCE_MS = 300;
const MAX_BACKOFF_MS = 15000;

function wait(ms: number, signal: AbortSignal) {
  return new Promise<void>((resolve) => {
    const timer = setTimeout(resolve, ms);
    signal.addEventListener("abort", () => (clearTimeout(timer), resolve()), { once: true });
  });
}

/**
 * Keeps one server-sent event stream open for a tenant, reconnecting with backoff.
 * Every event goes to subscribers immediately (an open conversation appends a message the moment it arrives).
 * `onListChange` is the coalesced trigger for the conversation list: any event, reconnect or tab-visible
 * resync; `resync` is true when events may have been missed.
 */
export function useConversationStream(
  tenantID: string,
  onListChange: (resync: boolean) => void,
) {
  const [status, setStatus] = useState<StreamStatus>("connecting");
  const listeners = useRef(new Set<StreamListener>());
  const listChange = useRef(onListChange);
  useEffect(() => {
    listChange.current = onListChange;
  });
  const subscribe = useCallback((listener: StreamListener) => {
    listeners.current.add(listener);
    return () => void listeners.current.delete(listener);
  }, []);
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout> | undefined;
    let pendingResync = false;
    const queueList = (resync: boolean) => {
      pendingResync ||= resync;
      timer ??= setTimeout(() => {
        timer = undefined;
        const resyncNow = pendingResync;
        pendingResync = false;
        if (!controller.signal.aborted) listChange.current(resyncNow);
      }, DEBOUNCE_MS);
    };
    const emit = (event: StreamEvent) => listeners.current.forEach((listener) => listener(event));
    const onVisible = () => {
      if (document.visibilityState !== "visible") return;
      emit({ type: "resync" });
      queueList(true);
    };
    document.addEventListener("visibilitychange", onVisible);

    async function run() {
      let delay = 1000;
      let first = true;
      while (!controller.signal.aborted) {
        let graceful = false;
        try {
          await streamConversations(tenantID, controller.signal, (event) => {
            if (event.type === "ready") {
              delay = 1000;
              setStatus("live");
              // Anything committed while disconnected was never pushed, so every reconnect resyncs.
              if (!first) {
                emit({ type: "resync" });
                queueList(true);
              }
              first = false;
              return;
            }
            emit(event);
            queueList(event.type === "resync");
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
  return { status, subscribe };
}
