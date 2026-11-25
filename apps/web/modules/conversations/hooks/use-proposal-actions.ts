"use client";

import { useEffect, useRef, useState } from "react";
import { SDKApiError } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { messaging, type ReplyProposal } from "../services/messaging";

/** How often the Desk re-reads a conversation while the bot is still preparing a proposal, and for how long. */
const routingPollMs = 3000;
const routingPollLimit = 20;

/**
 * Sends (as-is or edited), discards and regenerates the live bot proposal of a turn. A retry after an uncertain
 * send reuses the same client_message_id, so the API replays instead of sending twice. While the proposal is
 * still routing, the conversation is re-read on a short bounded poll.
 */
export function useProposalActions(args: {
  tenantID: string;
  turnID: string;
  proposal: ReplyProposal | null;
  /** Called after a command changed the conversation. */
  onChanged: () => void;
  /** Re-reads only this conversation while the proposal is still routing. */
  onPoll: () => void;
}) {
  const { tenantID, turnID, proposal, onChanged, onPoll } = args;
  const errorText = useErrorText();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [uncertain, setUncertain] = useState(false);
  const pending = useRef<{ proposal: string; client_message_id: string; text?: string } | null>(null);
  const onPollRef = useRef(onPoll);
  useEffect(() => {
    onPollRef.current = onPoll;
  });

  const routingID = proposal?.status === "routing" ? proposal.id : "";
  useEffect(() => {
    if (!routingID) return;
    let polls = 0;
    const timer = setInterval(() => {
      polls += 1;
      if (polls > routingPollLimit) clearInterval(timer);
      else onPollRef.current();
    }, routingPollMs);
    return () => clearInterval(timer);
  }, [routingID]);

  async function run(action: () => Promise<unknown>) {
    if (busy) return false;
    setBusy(true);
    setError("");
    try {
      await action();
      onChanged();
      return true;
    } catch (err) {
      setError(errorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  }

  async function send(text?: string) {
    if (!proposal || busy) return false;
    const input = text?.trim() && text.trim() !== proposal.text ? text.trim() : undefined;
    if (!pending.current || pending.current.proposal !== proposal.id || pending.current.text !== input) {
      pending.current = { proposal: proposal.id, client_message_id: crypto.randomUUID(), text: input };
    }
    const request = pending.current;
    setBusy(true);
    setError("");
    try {
      await messaging.sendProposal(tenantID, turnID, proposal.id, {
        client_message_id: request.client_message_id,
        ...(request.text ? { text: request.text } : {}),
      });
      pending.current = null;
      setUncertain(false);
      onChanged();
      return true;
    } catch (err) {
      const rejected = err instanceof SDKApiError && err.status >= 400 && err.status < 500;
      if (rejected) pending.current = null;
      setUncertain(!rejected);
      setError(errorText(err));
      if (rejected) onChanged();
      return false;
    } finally {
      setBusy(false);
    }
  }

  return {
    busy,
    error,
    uncertain,
    send,
    discard: () => (proposal ? run(() => messaging.discardProposal(tenantID, turnID, proposal.id)) : Promise.resolve(false)),
    regenerate: () => run(() => messaging.regenerateProposal(tenantID, turnID)),
  };
}
