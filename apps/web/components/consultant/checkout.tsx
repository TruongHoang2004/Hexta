"use client";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/sales/workspace";
import {
  sales,
  errorText,
  type Product,
  type Order,
  type Customer,
} from "@/lib/services/sales";
import { checkoutKey, type ConsultDraft } from "@/lib/services/consultant";
import { useAuthStore } from "@/store/useAuthStore";
import { buttonClass, Notice, inputClass } from "@/components/sales/primitives";

export function ConsultantCheckout() {
  const { tenant } = useWorkspace();
  const user = useAuthStore((state) => state.user);
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
          throw new Error("No valid order draft. Please return to the store.");
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
            throw new Error("Invalid draft item");
          const variantID =
            typeof item.variant_id === "string"
              ? item.variant_id
              : item.product_id;
          const product = products.find((p) =>
            p.variants.some((v) => v.id === variantID),
          );
          const variant = product?.variants.find((v) => v.id === variantID);
          if (!variant?.price_provided)
            throw new Error(
              "A product is no longer available. Please return to the store.",
            );
          items.push({
            product_id: product!.id,
            variant_id: variant.id,
            product_name: product!.name,
            quantity: item.quantity,
            unit_price: String(variant.price),
          });
        }
        if (!controller.signal.aborted)
          setDraft({ draft_id: "review", items, total_amount: "" });
      } catch (err) {
        if (!controller.signal.aborted) setError(errorText(err));
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => controller.abort();
  }, [tenant]);
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
          notes: "Shopping consultant checkout",
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
        setError(
          "Unable to verify draft creation. Check Orders before trying again.",
        );
      } else {
        setError(errorText(err));
      }
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }
  if (!tenant) return null;
  return (
    <div className="max-w-xl space-y-4">
      <h1 className="text-2xl font-semibold">Checkout</h1>
      <Notice error={error} />
      {loading ? (
        <p role="status">Checking current products and prices…</p>
      ) : confirmed ? (
        <p role="status">Order {order?.order_number} confirmed.</p>
      ) : (
        draft && (
          <>
            <p>
              Review these items. Prices were refreshed from the catalog. The
              server calculates totals and checks order eligibility when you
              confirm.
            </p>
            {draft.items.map((item, index) => (
              <p key={index}>
                {item.quantity} × {item.product_name} ·{" "}
                {Number(item.unit_price).toLocaleString("vi-VN")} VND each
              </p>
            ))}
            <label className="block">
              Customer
              <select
                className={inputClass}
                disabled={busy || !!order}
                value={customer}
                onChange={(e) => setCustomer(e.target.value)}
              >
                <option value="">Select a customer</option>
                {customers.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.full_name} · {c.phone || c.email}
                  </option>
                ))}
              </select>
            </label>
            {!customers.length && (
              <p>
                Add a customer with contact details in Customers before
                confirming.
              </p>
            )}
            {order && (
              <p>
                Server total:{" "}
                {Number(order.total_amount).toLocaleString("vi-VN")} VND
              </p>
            )}
            <button
              disabled={busy || uncertain || !customer || !user?.id}
              className={buttonClass}
              onClick={() => void confirm()}
            >
              {busy
                ? "Processing…"
                : order
                  ? "Confirm purchase"
                  : "Review server total"}
            </button>
          </>
        )
      )}
      <Link
        className="block text-primary underline"
        href={`/tenants/${tenant.id}/storefront`}
      >
        Back to store
      </Link>
      {(order || uncertain) && (
        <Link
          className="block text-primary underline"
          href={`/tenants/${tenant.id}/orders`}
        >
          View saved order
        </Link>
      )}
    </div>
  );
}
