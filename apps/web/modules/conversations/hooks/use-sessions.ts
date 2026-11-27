"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type MessagingSession, type MessagingTurn, type SessionInput } from "../services/messaging";

/**
 * A conversation's sessions (newest first) with paging, manual close and outcome edits. Reloads when `refresh`
 * changes, e.g. after the thread itself reloaded.
 */
export function useSessions(args: { tenantID: string; threadID: string; refresh: string; onChanged: () => void }) {
  const { tenantID, threadID, refresh, onChanged } = args;
  const errorText = useErrorText();
  const [items, setItems] = useState<MessagingSession[]>([]);
  const [nextCursor, setNextCursor] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  // Loading the list and running commands report separately, so a background reload never hides a command error.
  const [loadError, setLoadError] = useState("");
  const [reloads, setReloads] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    messaging
      .sessions(tenantID, threadID, "", controller.signal)
      .then((page) => {
        if (controller.signal.aborted) return;
        setLoadError("");
        setItems(page.items);
        setNextCursor(page.next_cursor);
      })
      .catch((err: unknown) => {
        if (!controller.signal.aborted) setLoadError(errorText(err));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [tenantID, threadID, refresh, reloads, errorText]);

  async function loadMore() {
    if (!nextCursor || busy) return;
    setBusy(true);
    setError("");
    try {
      const page = await messaging.sessions(tenantID, threadID, nextCursor);
      setItems((current) => [...current, ...page.items.filter((row) => !current.some((old) => old.turn.id === row.turn.id))]);
      setNextCursor(page.next_cursor);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  /** Runs a session command; true when it succeeded. Errors stay visible in `error`. */
  async function run(action: () => Promise<MessagingTurn>) {
    if (busy) return false;
    setBusy(true);
    setError("");
    try {
      await action();
      setReloads((v) => v + 1);
      onChanged();
      return true;
    } catch (err) {
      setError(errorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  }
  const close = (turn: MessagingTurn, input: SessionInput) => run(() => messaging.closeSession(tenantID, turn, input));
  const updateOutcome = (turn: MessagingTurn, input: SessionInput) =>
    run(() => messaging.updateOutcome(tenantID, turn, input));

  return { items, nextCursor, loading, busy, error: error || loadError, commandError: error, loadMore, close, updateOutcome };
}
