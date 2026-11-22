"use client";

import { useRouter } from "next/navigation";
import { Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useStorefrontProducts } from "../../hooks/use-storefront-products";
import { saveCheckoutDraft, singleItemDraft } from "../../utils";
import { ConsultantWidget } from "./consultant-widget";
import { StorefrontProductCard } from "./storefront-product-card";

/** Customer-facing storefront simulation with the AI consultant. */
export function StorefrontScreen() {
  const { tenant } = useWorkspace();
  const router = useRouter();
  const t = useT();
  const { products, loading, error, setError } = useStorefrontProducts(tenant?.id ?? "");
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
          {products.map((product) => (
            <StorefrontProductCard
              key={product.id}
              product={product}
              onBuy={(variant) => {
                try {
                  saveCheckoutDraft(
                    tenant.id,
                    singleItemDraft({
                      product_id: product.id,
                      variant_id: variant.id,
                      product_name: product.name,
                      price: String(variant.price),
                    }),
                  );
                  router.push(`/tenants/${tenant.id}/checkout`);
                } catch {
                  setError(t("storefront.checkoutUnavailable"));
                }
              }}
            />
          ))}
        </div>
      )}
      <ConsultantWidget key={tenant.id} tenantID={tenant.id} />
    </div>
  );
}
