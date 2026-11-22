"use client";

import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderLine, Product } from "../../types";
import { variantOptions } from "../../utils";
import { OrderLineRow } from "./order-line-row";

interface Props {
  lines: OrderLine[];
  products: Product[];
  onAdd: () => void;
  onUpdate: (index: number, patch: Partial<OrderLine>) => void;
  onRemove: (index: number) => void;
}

/** Order items, capped at 100 lines like the API. */
export function OrderLinesField({ lines, products, onAdd, onUpdate, onRemove }: Props) {
  const t = useT();
  const variants = variantOptions(products);
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">{t("orders.editor.items")}</h3>
        <button
          type="button"
          className={secondaryClass}
          disabled={!variants.length || lines.length >= 100}
          onClick={onAdd}
        >
          {t("orders.editor.addItem")}
        </button>
      </div>
      {!lines.length && (
        <p className="rounded-lg border border-dashed border-border p-4 text-sm text-muted">
          {t("orders.editor.noItems")} {!variants.length && t("orders.editor.noProducts")}
        </p>
      )}
      {lines.map((line, index) => (
        <OrderLineRow
          key={line.key}
          index={index}
          line={line}
          variants={variants}
          onChange={(patch) => onUpdate(index, patch)}
          onRemove={() => onRemove(index)}
        />
      ))}
      <p className="text-xs text-muted">{t("orders.editor.priceHint")}</p>
    </div>
  );
}
