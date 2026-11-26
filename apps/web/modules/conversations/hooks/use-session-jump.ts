"use client";

import { useEffect, useEffectEvent, useRef, useState } from "react";
import type { MessagingSession } from "../services/messaging";

/** Session the history scrolls to; `key` changes on every jump request. */
export type SessionFocus = { turnID: string; key: number };

/**
 * Jumps the thread history to a session's first message. A jump requested while the thread is loading or busy is
 * queued (latest request wins) and starts once the current load finishes, so a click is never silently dropped.
 * `clearFocus` is called once the history has scrolled, so later loads do not scroll back.
 */
export function useSessionJump(thread: {
  busy: boolean;
  loading: boolean;
  loadUntil: (messageID: string) => Promise<boolean>;
}) {
  const [focus, setFocus] = useState<SessionFocus | null>(null);
  const queued = useRef<MessagingSession | null>(null);

  function start(session: MessagingSession) {
    void thread.loadUntil(session.first_message_id).then((loaded) => {
      if (loaded) setFocus({ turnID: session.turn.id, key: Date.now() });
    });
  }

  function jump(session: MessagingSession) {
    if (thread.busy || thread.loading) {
      queued.current = session;
      return;
    }
    queued.current = null;
    start(session);
  }

  const startQueued = useEffectEvent(() => {
    const session = queued.current;
    if (!session || thread.busy || thread.loading) return;
    queued.current = null;
    start(session);
  });
  useEffect(() => {
    if (!thread.busy && !thread.loading) startQueued();
  }, [thread.busy, thread.loading]);

  return { focus, jump, clearFocus: () => setFocus(null) };
}
