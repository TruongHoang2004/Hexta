"use client";

import Image from "next/image";
import { Package } from "lucide-react";
import { buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConsultProduct } from "../../services/consultant";
import { formatVnd } from "../../utils";

/** A product the consultant recommended, with details and buy-now actions. */
export function RecommendedProductCard({
  product,
  onDetails,
  onBuy,
}: {
  product: ConsultProduct;
  onDetails: () => void;
  onBuy: () => void;
}) {
  const t = useT();
  return (
    <div className="w-56 shrink-0 snap-start space-y-2 rounded-lg border border-border p-3">
      {product.image_url ? (
        <Image
          src={product.image_url}
          alt={product.product_name}
          width={200}
          height={96}
          unoptimized
          className="h-24 w-full rounded object-contain"
        />
      ) : (
        <div className="flex h-24 items-center justify-center rounded bg-background" aria-label={t("consultant.noImage")}>
          <Package className="h-10 w-10 text-muted" />
        </div>
      )}
      <p className="font-semibold">{product.product_name}</p>
      <p>{formatVnd(product.price)}</p>
      <button className={secondaryClass} onClick={onDetails}>
        {t("consultant.details")}
      </button>
      <button className={buttonClass} onClick={onBuy}>
        {t("consultant.buy")}
      </button>
    </div>
  );
}
