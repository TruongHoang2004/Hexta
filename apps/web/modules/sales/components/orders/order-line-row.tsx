"use client";

import { Field, inputClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OrderLine } from "../../types";
import type { VariantOption } from "../../utils";

interface Props {
  index: number;
  line: OrderLine;
  variants: VariantOption[];
  onChange: (patch: Partial<OrderLine>) => void;
  onRemove: () => void;
}

/** One item: product variant (prefills its catalog price), quantity and unit price. */
export function OrderLineRow({ index, line, variants, onChange, onRemove }: Props) {
  const t = useT();
  const missingPrice = line.unit_price.trim() === "";
  return (
    <div className={`space-y-3 rounded-lg border bg-background p-3 ${missingPrice ? "border-error/60" : "border-border"}`}>
      <Field title={t("orders.editor.product", { index: index + 1 })}>
        <select
          className={inputClass}
          required
          value={line.variant_id}
          onChange={(e) => {
            const variant = variants.find((row) => row.id === e.target.value);
            onChange({
              variant_id: e.target.value,
              unit_price: variant?.price_provided ? String(variant.price) : "",
            });
          }}
        >
          <option value="">{t("orders.editor.selectProduct")}</option>
          {line.variant_id && !variants.some((row) => row.id === line.variant_id) && (
            <option value={line.variant_id}>{t("orders.editor.unavailableProduct", { id: line.variant_id })}</option>
          )}
          {variants.map((row) => (
            <option key={row.id} value={row.id}>
              {row.label}
            </option>
          ))}
        </select>
      </Field>
      <div className="grid grid-cols-[1fr_1fr_auto] items-end gap-3">
        <Field title={t("orders.editor.quantity")}>
          <input
            className={inputClass}
            type="number"
            min="1"
            max="1000000"
            step="1"
            required
            value={line.quantity}
            onChange={(e) => onChange({ quantity: e.target.value })}
          />
        </Field>
        <Field title={t("orders.editor.unitPrice")}>
          <input
            className={inputClass}
            type="number"
            min="0"
            max="9999999999999.99"
            step="0.01"
            placeholder={t("orders.editor.priceNotSet")}
            value={line.unit_price}
            onChange={(e) => onChange({ unit_price: e.target.value })}
          />
        </Field>
        <button
          type="button"
          aria-label={t("orders.editor.removeItem", { index: index + 1 })}
          className={secondaryClass}
          onClick={onRemove}
        >
          {t("orders.editor.remove")}
        </button>
      </div>
    </div>
  );
}
