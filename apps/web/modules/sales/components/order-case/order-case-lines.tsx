"use client";

import { inputClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Order } from "../../types";

interface Props {
  order: Order;
  quantities: Record<string, string>;
  /** Original quantity an input exceeds, or 0 for an invalid number, keyed by order item ID. */
  errors: Record<string, number>;
  readonly: boolean;
  itemLabel: (variantID: string, itemID: string) => string;
  onChange: (itemID: string, quantity: string) => void;
}

/** Original order lines with the quantity returned for each; never above the original quantity. */
export function OrderCaseLines({ order, quantities, errors, readonly, itemLabel, onChange }: Props) {
  const t = useT();
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">{t("orderCase.lines.title")}</h4>
      <table className="w-full text-sm">
        <thead className="text-left text-xs text-muted">
          <tr>
            <th scope="col" className="py-1 font-medium">
              {t("orderCase.lines.product")}
            </th>
            <th scope="col" className="py-1 font-medium">
              {t("orderCase.lines.original")}
            </th>
            <th scope="col" className="py-1 font-medium">
              {t("orderCase.lines.returned")}
            </th>
          </tr>
        </thead>
        <tbody>
          {order.items.map((item) => {
            const label = itemLabel(item.variant_id, item.id);
            const error = errors[item.id];
            const errorID = `case-line-error-${item.id}`;
            return (
              <tr key={item.id} className="border-t border-border align-top">
                <td className="py-2 pr-3">{label}</td>
                <td className="py-2 pr-3">{item.quantity}</td>
                <td className="py-2">
                  <input
                    className={`${inputClass} w-28 ${error !== undefined ? "border-error" : ""}`}
                    type="number"
                    min="0"
                    max={item.quantity}
                    step="1"
                    inputMode="numeric"
                    aria-label={t("orderCase.lines.returnedFor", { product: label })}
                    aria-invalid={error !== undefined}
                    aria-describedby={error !== undefined ? errorID : undefined}
                    disabled={readonly}
                    placeholder="0"
                    value={quantities[item.id] ?? ""}
                    onChange={(event) => onChange(item.id, event.target.value)}
                  />
                  {error !== undefined && (
                    <p id={errorID} role="alert" className="mt-1 text-xs text-error">
                      {error > 0 ? t("orderCase.lines.over", { max: error }) : t("orderCase.lines.invalid")}
                    </p>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
