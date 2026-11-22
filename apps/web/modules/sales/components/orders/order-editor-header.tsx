"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Order } from "../../types";
import { statusChip } from "../../utils";

export function OrderEditorHeader({ order }: { order: Order | null }) {
  const t = useT();
  const chip = order ? statusChip(t, order.status) : null;
  return (
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 className="text-lg font-semibold">{order ? order.order_number : t("orders.editor.newTitle")}</h2>
        <p className="mt-1 text-xs text-muted">
          {order ? t("orders.editor.version", { version: order.version }) : t("orders.editor.newHint")}
        </p>
      </div>
      {chip && <StatusChip tone={chip.tone}>{chip.label}</StatusChip>}
    </div>
  );
}
