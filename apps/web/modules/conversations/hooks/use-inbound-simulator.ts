"use client";

import { useRef, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { messaging, type InboundInput } from "../services/messaging";

/**
 * Saves a simulated customer message. An unchanged retry reuses its request key so the
 * API does not store it twice.
 */
export function useInboundSimulator(tenantID: string, onSaved: (threadID: string) => void) {
  const errorText = useErrorText();
  const [external, setExternal] = useState("");
  const [name, setName] = useState("");
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const pending = useRef<InboundInput | null>(null);
  const values = { external_key: external.trim(), display_name: name.trim(), body: body.trim() };
  const ready = !!values.external_key && !!values.display_name && !!values.body;

  async function submit() {
    if (busy || !ready) return;
    const changed =
      !pending.current ||
      Object.entries(values).some(([key, value]) => pending.current?.[key as keyof InboundInput] !== value);
    if (changed) pending.current = { ...values, request_key: crypto.randomUUID() };
    setBusy(true);
    setError("");
    try {
      const message = await messaging.inbound(tenantID, pending.current!);
      pending.current = null;
      setBody("");
      onSaved(message.thread_id);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  return { external, setExternal, name, setName, body, setBody, busy, error, ready, submit };
}
