"use client";

import { useEffect, useRef, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { converse, type ConsultEvent } from "../services/consultant";
import type { ConsultTurn } from "../types";

/** Merges one streamed event into the assistant turn being written. */
function applyEvent(turn: ConsultTurn, event: ConsultEvent): ConsultTurn {
  switch (event.type) {
    case "message":
      return { ...turn, content: turn.content + (event.text || ""), needsHuman: event.needs_human };
    case "products":
      return { ...turn, products: event.products };
    case "draft_order":
      return { ...turn, draft: event.draft };
    case "quick_replies":
      return { ...turn, replies: event.quick_replies };
    default:
      return turn;
  }
}

/** Streaming chat with the AI consultant. One request at a time; `stop` aborts it. */
export function useConsultantChat(tenantID: string) {
  const errorText = useErrorText();
  const [turns, setTurns] = useState<ConsultTurn[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const pending = useRef<AbortController | null>(null);
  const session = useRef("");
  useEffect(() => () => pending.current?.abort(), []);

  async function send(text: string) {
    if (pending.current || !text.trim()) return;
    const controller = new AbortController();
    pending.current = controller;
    if (!session.current) session.current = crypto.randomUUID();
    const history = turns.filter((turn) => turn.content).map(({ role, content }) => ({ role, content }));
    setTurns((previous) => [...previous, { role: "user", content: text }, { role: "assistant", content: "" }]);
    setBusy(true);
    setError("");
    try {
      await converse(tenantID, session.current, text, history, controller.signal, (event) => {
        if (controller.signal.aborted) return;
        setTurns((previous) =>
          previous.map((turn, index) => (index === previous.length - 1 ? applyEvent(turn, event) : turn)),
        );
      });
    } catch (err) {
      if (!controller.signal.aborted) setError(errorText(err));
    } finally {
      if (pending.current === controller) {
        pending.current = null;
        setBusy(false);
      }
    }
  }

  const stop = () => pending.current?.abort();

  return { turns, busy, error, setError, send, stop };
}
