"use client";

import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMoney } from "../../hooks/use-money";
import type { Order } from "../../types";

interface Props {
  order: Order;
  dirty: boolean;
  busy: boolean;
  onReload: () => void;
}

/** Server-computed value of the saved order, with a reload that discards local edits. */
export function OrderSavedSummary({ order, dirty, busy, onReload }: Props) {
  const t = useT();
  const money = useMoney();
  return (
    <div className="space-y-2 border-t border-border pt-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-sm">{t("orders.editor.savedValue")}</span>
        <strong>{money(order.total_amount)}</strong>
      </div>
      <p className="text-xs text-muted">
        {order.items.some((item) => !item.price_provided)
          ? t("orders.editor.valueIncomplete")
          : t("orders.editor.valueNote")}
      </p>
      {dirty && order.status === "draft" && (
        <p role="status" className="text-sm">
          {t("orders.editor.unsaved")}
        </p>
      )}
      <button type="button" className={secondaryClass} disabled={busy} onClick={onReload}>
        {t("orders.editor.reload")}
      </button>
    </div>
  );
}
