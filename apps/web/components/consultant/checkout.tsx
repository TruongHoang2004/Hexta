"use client";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/sales/workspace";
import {
  sales,
  type Product,
  type Order,
  type Customer,
} from "@/lib/services/sales";
import { checkoutKey, type ConsultDraft } from "@/lib/services/consultant";
import {
  resolveCheckoutVariant,
  checkoutProductName,
} from "@/lib/services/consultant-checkout";
import { useAuthStore } from "@/store/useAuthStore";
import { buttonClass, Notice, inputClass } from "@/components/sales/primitives";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";

/** A checkout precondition failure with a localized message. */
class CheckoutError extends Error {
  constructor(public readonly key: MessageKey) {
    super(key);
  }
}
const vnd = (value: string | number) =>
  new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(Number(value));

export function ConsultantCheckout() {
  const { tenant } = useWorkspace();
  const user = useAuthStore((state) => state.user);
  const t = useT();
  const baseErrorText = useErrorText();
  const [customers, setCustomers] = useState<Customer[]>([]),
    [customer, setCustomer] = useState("");
  const [draft, setDraft] = useState<ConsultDraft | null>(null),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true),
    [busy, setBusy] = useState(false),
    [order, setOrder] = useState<Order | null>(null),
    [confirmed, setConfirmed] = useState(false),
    [uncertain, setUncertain] = useState(false);
  const lock = useRef(false);
  useEffect(() => {
    if (!tenant) return;
    const controller = new AbortController();
    async function load() {
      try {
        const raw: unknown = JSON.parse(
          sessionStorage.getItem(checkoutKey(tenant!.id)) || "null",
        );
        if (
          !raw ||
          typeof raw !== "object" ||
          !("items" in raw) ||
          !Array.isArray(raw.items) ||
          !raw.items.length ||
          raw.items.length > 50
        )
          throw new CheckoutError("checkout.noDraft");
        const [products, customerRows] = await Promise.all([
          sales.all<Product>("products", tenant!.id, controller.signal),
          sales.all<Customer>("customers", tenant!.id, controller.signal),
        ]);
        if (!controller.signal.aborted)
          setCustomers(
            customerRows.filter((c) => c.full_name && (c.phone || c.email)),
          );
        const items: ConsultDraft["items"] = [];
        for (const item of raw.items) {
          if (
            !item ||
            typeof item !== "object" ||
            typeof item.product_id !== "string" ||
            !Number.isInteger(item.quantity) ||
            item.quantity < 1 ||
            item.quantity > 100
          )
            throw new CheckoutError("checkout.invalidItem");
          const resolved = resolveCheckoutVariant(
            products,
            item.product_id,
            item.variant_id,
          );
          const product = resolved?.product;
          const variant = resolved?.variant;
          if (!variant?.price_provided)
            throw new CheckoutError("checkout.unavailable");
          items.push({
            product_id: product!.id,
            variant_id: variant.id,
            product_name: checkoutProductName(
              product!.name,
              variant.variant_name,
            ),
            quantity: item.quantity,
            unit_price: String(variant.price),
          });
        }
        if (!controller.signal.aborted)
          setDraft({ draft_id: "review", items, total_amount: "" });
      } catch (err) {
        if (!controller.signal.aborted)
          setError(err instanceof CheckoutError ? t(err.key) : baseErrorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenant, t, baseErrorText]);
  async function confirm() {
    if (
      !tenant ||
      !draft ||
      !user?.id ||
      !customer ||
      uncertain ||
      lock.current
    )
      return;
    lock.current = true;
    setBusy(true);
    setError("");
    let creating = false;
    try {
      let current = order;
      if (!current) {
        creating = true;
        current = await sales.saveDraft(tenant.id, null, {
          version: 0,
          customer_id: customer,
          employee_id: user.id,
          requires_delivery: false,
          delivery_address: "",
          notes: t("checkout.notes"),
          items: draft.items.map((item) => ({
            variant_id: item.variant_id!,
            quantity: item.quantity,
            unit_price: item.unit_price,
          })),
        });
        setOrder(current);
        creating = false;
        return;
      }
      const result = await sales.confirm(
        tenant.id,
        current.id,
        current.version,
      );
      setOrder(result);
      setConfirmed(true);
      sessionStorage.removeItem(checkoutKey(tenant.id));
    } catch (err) {
      if (creating) {
        setUncertain(true);
        setError(t("checkout.uncertain"));
      } else {
        setError(baseErrorText(err));
      }
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }
  if (!tenant) return null;
  return (
    <div className="max-w-xl space-y-4">
      <h1 className="text-2xl font-semibold">{t("checkout.title")}</h1>
      <Notice error={error} />
      {loading ? (
        <p role="status">{t("checkout.checking")}</p>
      ) : confirmed ? (
        <p role="status">{t("checkout.confirmedOrder", { number: order?.order_number ?? "" })}</p>
      ) : (
        draft && (
          <>
            <p>
              {t("checkout.review")}
            </p>
            {draft.items.map((item, index) => (
              <p key={index}>
                {t("checkout.line", { quantity: item.quantity, name: item.product_name, price: vnd(item.unit_price) })}
              </p>
            ))}
            <label className="block">
              {t("checkout.customer")}
              <select
                className={inputClass}
                disabled={busy || !!order}
                value={customer}
                onChange={(e) => setCustomer(e.target.value)}
              >
                <option value="">{t("checkout.selectCustomer")}</option>
                {customers.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.full_name} · {c.phone || c.email}
                  </option>
                ))}
              </select>
            </label>
            {!customers.length && (
              <p>
                {t("checkout.noCustomers")}
              </p>
            )}
            {order && (
              <p>
                {t("checkout.serverTotal", { amount: vnd(order.total_amount) })}
              </p>
            )}
            <button
              disabled={busy || uncertain || !customer || !user?.id}
              className={buttonClass}
              onClick={() => void confirm()}
            >
              {busy
                ? t("checkout.processing")
                : order
                  ? t("checkout.confirm")
                  : t("checkout.reviewTotal")}
            </button>
          </>
        )
      )}
      <Link
        className="block text-primary underline"
        href={`/tenants/${tenant.id}/storefront`}
      >
        {t("checkout.back")}
      </Link>
      {(order || uncertain) && (
        <Link
          className="block text-primary underline"
          href={`/tenants/${tenant.id}/orders`}
        >
          {t("checkout.viewOrder")}
        </Link>
      )}
    </div>
  );
}
