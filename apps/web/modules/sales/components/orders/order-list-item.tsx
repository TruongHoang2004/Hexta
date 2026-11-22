"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMoney } from "../../hooks/use-money";
import type { Customer, Order } from "../../types";
import { statusChip } from "../../utils";

interface Props {
  order: Order;
  customer: Customer | undefined;
  selected: boolean;
  busy: boolean;
  onOpen: () => void;
}

export function OrderListItem({ order, customer, selected, busy, onOpen }: Props) {
  const t = useT();
  const money = useMoney();
  const chip = statusChip(t, order.status);
  const incomplete = order.items.some((item) => !item.price_provided);
  return (
    <li>
      <button
        type="button"
        disabled={busy}
        aria-current={selected ? "true" : undefined}
        onClick={onOpen}
        className={`w-full rounded-lg p-3 text-left hover:bg-background disabled:opacity-50 ${selected ? "bg-primary/10" : ""}`}
      >
        <div className="flex items-center justify-between gap-3">
          <span className="truncate text-sm font-medium">
            {customer?.full_name ||
              (order.customer_id ? t("orders.unavailableCustomer") : t("orders.noCustomer"))}
          </span>
          <StatusChip tone={chip.tone}>{chip.label}</StatusChip>
        </div>
        <p className="mt-1 truncate font-mono text-xs text-muted">{order.order_number}</p>
        <p className="mt-2 text-sm">
          {t("orders.itemsCount", { count: order.items.length })} · {t("orders.orderValue")}{" "}
          {money(order.total_amount)}
          {incomplete && <span className="text-muted"> ({t("orders.incomplete")})</span>}
        </p>
      </button>
    </li>
  );
}
