"use client";

import { useState } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { sales } from "../services/sales";
import type { Order, OrderLine } from "../types";
import { draftInput, linesFrom, type DraftFields } from "../utils";

export interface OrderEditorOptions {
  tenantID: string;
  order: Order | null;
  busy: boolean;
  setBusy: (busy: boolean) => void;
  dirty: boolean;
  setDirty: (dirty: boolean) => void;
  onSaved: (order: Order) => void;
  /** Links a new draft to this open conversation session (from the Desk); unused for existing orders. */
  sessionTurnID?: string;
}

/** Draft form fields plus save, claim, confirm and reload; `busy`/`dirty` are owned by the screen. */
export function useOrderEditor({ tenantID, order, busy, setBusy, dirty, setDirty, onSaved, sessionTurnID }: OrderEditorOptions) {
  const t = useT();
  const errorText = useErrorText();
  const [fields, setFields] = useState<DraftFields>(() => ({
    customer: order?.customer_id || "",
    employee: order?.employee_id || "",
    delivery: order?.requires_delivery || false,
    address: order?.delivery_address || "",
    notes: order?.notes || "",
    lines: linesFrom(order),
  }));
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);

  function update(patch: Partial<DraftFields>) {
    setDirty(true);
    setFields((current) => ({ ...current, ...patch }));
  }
  const setLines = (lines: OrderLine[]) => update({ lines });
  const updateLine = (index: number, patch: Partial<OrderLine>) =>
    setLines(fields.lines.map((line, i) => (i === index ? { ...line, ...patch } : line)));
  const addLine = () =>
    setLines([...fields.lines, { key: crypto.randomUUID(), variant_id: "", quantity: "1", unit_price: "" }]);
  const removeLine = (index: number) => setLines(fields.lines.filter((_, i) => i !== index));

  async function run(action: () => Promise<Order>) {
    setBusy(true);
    setError("");
    setConflict(false);
    try {
      onSaved(await action());
    } catch (err) {
      setConflict(isConflict(err));
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  /** Save the draft with the given assignee (claim passes the current user). */
  function persist(employeeID: string) {
    if (!tenantID || busy) return;
    const input = draftInput(order, fields, employeeID);
    if (!order && sessionTurnID) input.messaging_turn_id = sessionTurnID;
    return run(() => sales.saveDraft(tenantID, order?.id || null, input));
  }
  function confirm() {
    if (!tenantID || !order || dirty || busy) return;
    return run(() => sales.confirm(tenantID, order.id, order.version));
  }
  function reload(askFirst: boolean) {
    if (!tenantID || !order || (askFirst && dirty && !window.confirm(t("orders.discardReload")))) return;
    return run(() => sales.order(tenantID, order.id));
  }

  return { fields, update, updateLine, addLine, removeLine, error, conflict, persist, confirm, reload };
}
