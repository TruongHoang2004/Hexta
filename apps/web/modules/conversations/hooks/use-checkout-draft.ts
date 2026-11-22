"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import { sales, type Customer, type Product } from "@/modules/sales";
import { checkoutKey, type ConsultDraft } from "../services/consultant";
import { checkoutProductName, resolveCheckoutVariant } from "../services/consultant-checkout";

/** A checkout precondition failure with a localized message. */
class CheckoutError extends Error {
  constructor(public readonly key: MessageKey) {
    super(key);
  }
}

/** The raw item list of a handed-over draft; checked before any request is made. */
function draftItems(raw: unknown): unknown[] {
  if (
    !raw ||
    typeof raw !== "object" ||
    !("items" in raw) ||
    !Array.isArray(raw.items) ||
    !raw.items.length ||
    raw.items.length > 50
  )
    throw new CheckoutError("checkout.noDraft");
  return raw.items;
}

/** Re-prices every item from current products; never trusts prices stored in the browser. */
function priceDraft(rawItems: unknown[], products: Product[]): ConsultDraft {
  const items: ConsultDraft["items"] = [];
  for (const item of rawItems as ({ product_id?: unknown; variant_id?: string; quantity?: unknown } | null)[]) {
    if (
      !item ||
      typeof item !== "object" ||
      typeof item.product_id !== "string" ||
      typeof item.quantity !== "number" ||
      !Number.isInteger(item.quantity) ||
      item.quantity < 1 ||
      item.quantity > 100
    )
      throw new CheckoutError("checkout.invalidItem");
    const resolved = resolveCheckoutVariant(products, item.product_id, item.variant_id);
    const variant = resolved?.variant;
    if (!resolved || !variant?.price_provided) throw new CheckoutError("checkout.unavailable");
    items.push({
      product_id: resolved.product.id,
      variant_id: variant.id,
      product_name: checkoutProductName(resolved.product.name, variant.variant_name),
      quantity: item.quantity,
      unit_price: String(variant.price),
    });
  }
  return { draft_id: "review", items, total_amount: "" };
}

/** Loads the draft handed over by the storefront or consultant, plus customers that can receive an order. */
export function useCheckoutDraft(tenantID: string) {
  const t = useT();
  const errorText = useErrorText();
  const [draft, setDraft] = useState<ConsultDraft | null>(null);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    async function load() {
      try {
        const raw: unknown = JSON.parse(sessionStorage.getItem(checkoutKey(tenantID)) || "null");
        const rawItems = draftItems(raw);
        const [products, customerRows] = await Promise.all([
          sales.all<Product>("products", tenantID, controller.signal),
          sales.all<Customer>("customers", tenantID, controller.signal),
        ]);
        if (controller.signal.aborted) return;
        setCustomers(customerRows.filter((c) => c.full_name && (c.phone || c.email)));
        setDraft(priceDraft(rawItems, products));
      } catch (err) {
        if (!controller.signal.aborted) setError(err instanceof CheckoutError ? t(err.key) : errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenantID, t, errorText]);
  return { draft, customers, loading, error };
}
