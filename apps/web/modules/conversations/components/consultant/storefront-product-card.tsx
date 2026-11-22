"use client";

import { buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Product } from "@/modules/sales";
import { formatVnd } from "../../utils";

/** A product on the storefront; only variants with a price can be bought. */
export function StorefrontProductCard({
  product,
  onBuy,
}: {
  product: Product;
  onBuy: (variant: Product["variants"][number]) => void;
}) {
  const t = useT();
  return (
    <article id={product.id} className="space-y-3 rounded-xl border border-border bg-surface p-4">
      <div className="flex h-32 items-center justify-center rounded bg-background text-muted">{t("storefront.noImage")}</div>
      <h2 className="font-semibold">{product.name}</h2>
      <p>{product.description}</p>
      {product.variants
        .filter((v) => v.price_provided)
        .map((v) => (
          <div className="space-y-2" key={v.id}>
            <p>{t("storefront.price", { variant: v.variant_name, price: formatVnd(v.price) })}</p>
            <button className={buttonClass} onClick={() => onBuy(v)}>
              {t("storefront.buy")}
            </button>
          </div>
        ))}
    </article>
  );
}
