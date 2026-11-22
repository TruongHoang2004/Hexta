"use client";

import { useEffect, useRef, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type MessagingDetail } from "../services/messaging";
import type { HandoffReason, ThreadCommand } from "../types";

/** Replaces the newest page with fresh data while keeping older pages the user already loaded. */
function mergeNewest(current: MessagingDetail, fresh: MessagingDetail): MessagingDetail {
  const older = current.messages.filter((row) => !fresh.messages.some((next) => next.id === row.id));
  return {
    ...fresh,
    messages: [...fresh.messages, ...older],
    next_cursor: older.length ? current.next_cursor : fresh.next_cursor,
  };
}

/**
 * Loads one conversation with its newest messages, pages older ones and runs turn commands
 * (claim, assign, handoff, return to bot). Reloads when `externalRefresh` changes. A change of `live` (a realtime
 * hint for this thread) reloads quietly and keeps older pages already loaded.
 */
export function useThread(args: {
  tenantID: string;
  threadID: string;
  externalRefresh: number;
  live: number;
  onChanged: () => void;
}) {
  const { tenantID, threadID, externalRefresh, live, onChanged } = args;
  const errorText = useErrorText();
  const [detail, setDetail] = useState<MessagingDetail | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const seenLive = useRef(live);
  useEffect(() => {
    const silent = seenLive.current !== live;
    seenLive.current = live;
    const controller = new AbortController();
    async function load() {
      if (!silent) {
        setLoading(true);
        setError("");
      }
      try {
        const result = await messaging.detail(tenantID, threadID, "", controller.signal);
        if (!controller.signal.aborted) setDetail((current) => (silent && current ? mergeNewest(current, result) : result));
      } catch (err) {
        if (!controller.signal.aborted && !silent) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted && !silent) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, threadID, refresh, externalRefresh, live, errorText]);

  const reload = () => setRefresh((v) => v + 1);

  async function command(action: ThreadCommand, assignee: string | null, reason: HandoffReason) {
    const turn = detail?.conversation.turn;
    if (!turn || busy) return false;
    setBusy(true);
    setError("");
    try {
      await messaging.command(tenantID, turn, action, assignee, reason);
      reload();
      onChanged();
      return true;
    } catch (err) {
      setError(errorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  }

  async function loadOlder() {
    if (!detail?.next_cursor || busy) return;
    setBusy(true);
    setError("");
    try {
      const result = await messaging.detail(tenantID, threadID, detail.next_cursor);
      setDetail((current) =>
        current
          ? {
              ...current,
              messages: [
                ...current.messages,
                ...result.messages.filter((row) => !current.messages.some((old) => old.id === row.id)),
              ],
              next_cursor: result.next_cursor,
            }
          : current,
      );
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  return { detail, loading, busy, error, reload, command, loadOlder };
}
