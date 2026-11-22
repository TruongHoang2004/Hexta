"use client";

import Link from "next/link";
import { Notice, Picker, buttonClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useCheckoutDraft } from "../../hooks/use-checkout-draft";
import { useCheckoutOrder } from "../../hooks/use-checkout-order";
import { formatVnd } from "../../utils";

/** Review a handed-over draft, pick a customer, let the server price it, then confirm. */
export function CheckoutScreen() {
  const { tenant } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const { draft, customers, loading, error: loadError } = useCheckoutDraft(tenantID);
  const checkout = useCheckoutOrder(tenantID, draft);
  if (!tenant) return null;
  const { order } = checkout;
  return (
    <div className="max-w-xl space-y-4">
      <h1 className="text-2xl font-semibold">{t("checkout.title")}</h1>
      <Notice error={checkout.error || loadError} />
      {loading ? (
        <p role="status">{t("checkout.checking")}</p>
      ) : checkout.confirmed ? (
        <p role="status">{t("checkout.confirmedOrder", { number: order?.order_number ?? "" })}</p>
      ) : (
        draft && (
          <>
            <p>{t("checkout.review")}</p>
            {draft.items.map((item, index) => (
              <p key={index}>
                {t("checkout.line", { quantity: item.quantity, name: item.product_name, price: formatVnd(item.unit_price) })}
              </p>
            ))}
            <label className="block space-y-1.5">
              {t("checkout.customer")}
              <Picker
                disabled={checkout.busy || !!order}
                value={checkout.customer}
                onChange={checkout.setCustomer}
                placeholder={t("checkout.selectCustomer")}
                options={customers.map((c) => ({ value: c.id, label: c.full_name, description: c.phone || c.email }))}
              />
            </label>
            {!customers.length && <p>{t("checkout.noCustomers")}</p>}
            {order && <p>{t("checkout.serverTotal", { amount: formatVnd(order.total_amount) })}</p>}
            <button disabled={!checkout.canAdvance} className={buttonClass} onClick={() => void checkout.advance()}>
              {checkout.busy ? t("checkout.processing") : order ? t("checkout.confirm") : t("checkout.reviewTotal")}
            </button>
          </>
        )
      )}
      <Link className="block text-primary underline" href={`/tenants/${tenant.id}/storefront`}>
        {t("checkout.back")}
      </Link>
      {(order || checkout.uncertain) && (
        <Link className="block text-primary underline" href={`/tenants/${tenant.id}/orders`}>
          {t("checkout.viewOrder")}
        </Link>
      )}
    </div>
  );
}
