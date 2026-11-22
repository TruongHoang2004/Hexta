"use client";

import { useRef, useState } from "react";
import { SDKApiError } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import {
  messaging,
  type ChannelSendInput,
  type MessagingTurn,
  type SendInput,
} from "../services/messaging";

/**
 * Sends a staff reply. Retries after an uncertain failure reuse the same request key or
 * client_message_id, so the API replays instead of duplicating the message.
 */
export function useReplySender(args: { tenantID: string; realChannel: boolean; onSent: () => void }) {
  const { tenantID, realChannel, onSent } = args;
  const errorText = useErrorText();
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const pending = useRef<SendInput | null>(null);
  const pendingChannel = useRef<ChannelSendInput | null>(null);

  function succeeded() {
    setUncertain(false);
    setBody("");
    onSent();
  }

  async function sendSimulated(turn: MessagingTurn) {
    if (!pending.current) {
      pending.current = {
        expected_version: turn.version,
        context_revision: turn.context_revision,
        body: body.trim(),
        request_key: crypto.randomUUID(),
      };
    }
    try {
      await messaging.send(tenantID, turn.id, pending.current);
      pending.current = null;
      succeeded();
    } catch (err) {
      const rejected =
        err instanceof SDKApiError &&
        (err.status === 400 || err.status === 409 || (!uncertain && err.status >= 400 && err.status < 500));
      if (rejected) pending.current = null;
      setUncertain(!rejected);
      setError(errorText(err));
    }
  }

  async function sendChannel(turn: MessagingTurn) {
    if (!pendingChannel.current) {
      pendingChannel.current = {
        client_message_id: crypto.randomUUID(),
        text: body.trim(),
        expected_version: turn.version,
        context_revision: turn.context_revision,
      };
    }
    try {
      await messaging.sendChannel(tenantID, turn.id, pendingChannel.current);
      pendingChannel.current = null;
      succeeded();
    } catch (err) {
      const rejected = err instanceof SDKApiError && err.status >= 400 && err.status < 500;
      if (rejected) pendingChannel.current = null;
      setUncertain(!rejected);
      setError(errorText(err));
    }
  }

  async function send(turn: MessagingTurn) {
    if (busy || !body.trim()) return;
    setBusy(true);
    setError("");
    try {
      await (realChannel ? sendChannel(turn) : sendSimulated(turn));
    } finally {
      setBusy(false);
    }
  }

  return { body, setBody, busy, error, uncertain, send };
}
