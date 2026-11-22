"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type MessagingDetail } from "../services/messaging";
import type { HandoffReason, ThreadCommand } from "../types";

/**
 * Loads one conversation with its newest messages, pages older ones and runs turn commands
 * (claim, assign, handoff, return to bot). Reloads when `externalRefresh` changes.
 */
export function useThread(args: {
  tenantID: string;
  threadID: string;
  externalRefresh: number;
  onChanged: () => void;
}) {
  const { tenantID, threadID, externalRefresh, onChanged } = args;
  const errorText = useErrorText();
  const [detail, setDetail] = useState<MessagingDetail | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const result = await messaging.detail(tenantID, threadID, "", controller.signal);
        if (!controller.signal.aborted) setDetail(result);
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, threadID, refresh, externalRefresh, errorText]);

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
