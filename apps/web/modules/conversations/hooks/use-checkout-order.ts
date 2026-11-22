"use client";

import { useRef, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { sales, type Order } from "@/modules/sales";
import { useAuthStore } from "@/store/useAuthStore";
import { checkoutKey, type ConsultDraft } from "../services/consultant";

/**
 * Two-step checkout: the first call saves a draft order so the server prices it, the second confirms it.
 * If saving fails without an answer the order may exist, so further attempts are blocked (`uncertain`).
 */
export function useCheckoutOrder(tenantID: string, draft: ConsultDraft | null) {
  const t = useT();
  const errorText = useErrorText();
  const userID = useAuthStore((state) => state.user?.id);
  const [customer, setCustomer] = useState("");
  const [order, setOrder] = useState<Order | null>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const lock = useRef(false);

  async function advance() {
    if (!tenantID || !draft || !userID || !customer || uncertain || lock.current) return;
    lock.current = true;
    setBusy(true);
    setError("");
    let creating = false;
    try {
      if (!order) {
        creating = true;
        const saved = await sales.saveDraft(tenantID, null, {
          version: 0,
          customer_id: customer,
          employee_id: userID,
          requires_delivery: false,
          delivery_address: "",
          notes: t("checkout.notes"),
          items: draft.items.map((item) => ({
            variant_id: item.variant_id!,
            quantity: item.quantity,
            unit_price: item.unit_price,
          })),
        });
        setOrder(saved);
        return;
      }
      setOrder(await sales.confirm(tenantID, order.id, order.version));
      setConfirmed(true);
      sessionStorage.removeItem(checkoutKey(tenantID));
    } catch (err) {
      if (creating) {
        setUncertain(true);
        setError(t("checkout.uncertain"));
      } else {
        setError(errorText(err));
      }
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }

  return {
    customer,
    setCustomer,
    order,
    confirmed,
    uncertain,
    busy,
    error,
    canAdvance: !busy && !uncertain && !!customer && !!userID,
    advance,
  };
}
