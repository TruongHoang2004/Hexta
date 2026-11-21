"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useWorkspace } from "@/components/sales/workspace";
import { sales, type Product } from "@/lib/services/sales";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { checkoutKey } from "@/lib/services/consultant";
import { AIConsultantWidget } from "./widget";
import { buttonClass, Notice } from "@/components/sales/primitives";
export function Storefront() {
  const { tenant } = useWorkspace();
  const router = useRouter();
  const t = useT();
  const errorText = useErrorText();
  const [products, setProducts] = useState<Product[]>([]),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true);
  useEffect(() => {
    if (!tenant) return;
    const controller = new AbortController();
    sales
      .all<Product>("products", tenant.id, controller.signal)
      .then(setProducts)
      .catch((err) => {
        if (!controller.signal.aborted) setError(errorText(err));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [tenant, errorText]);
  if (!tenant) return null;
  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-semibold">{t("storefront.title", { name: tenant.name })}</h1>
      <Notice error={error} />
      {loading ? (
        <p role="status">{t("storefront.loading")}</p>
      ) : !products.length ? (
        <p>{t("storefront.empty")}</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {products.map((p) => (
            <article
              id={p.id}
              key={p.id}
              className="space-y-3 rounded-xl border border-border bg-surface p-4"
            >
              <div className="flex h-32 items-center justify-center rounded bg-background text-muted">
                {t("storefront.noImage")}
              </div>
              <h2 className="font-semibold">{p.name}</h2>
              <p>{p.description}</p>
              {p.variants
                .filter((v) => v.price_provided)
                .map((v) => (
                  <div className="space-y-2" key={v.id}>
                    <p>
                      {t("storefront.price", {
                        variant: v.variant_name,
                        price: new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(Number(v.price)),
                      })}
                    </p>
                    <button
                      className={buttonClass}
                      onClick={() => {
                        try {
                          sessionStorage.setItem(
                            checkoutKey(tenant.id),
                            JSON.stringify({
                              draft_id: crypto.randomUUID(),
                              items: [
                                {
                                  product_id: p.id,
                                  variant_id: v.id,
                                  product_name: p.name,
                                  quantity: 1,
                                  unit_price: String(v.price),
                                },
                              ],
                              total_amount: String(v.price),
                            }),
                          );
                          router.push(`/tenants/${tenant.id}/checkout`);
                        } catch {
                          setError(t("storefront.checkoutUnavailable"));
                        }
                      }}
                    >
                      {t("storefront.buy")}
                    </button>
                  </div>
                ))}
            </article>
          ))}
        </div>
      )}
      <AIConsultantWidget key={tenant.id} tenantID={tenant.id} />
    </div>
  );
}
