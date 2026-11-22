"use client";

import { EmptyState, Pager, SegmentedFilter } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { useOrderList } from "../../hooks/use-order-list";
import type { Customer, Order } from "../../types";
import { orderListViews } from "../../utils";
import { OrderListItem } from "./order-list-item";

interface Props {
  list: ReturnType<typeof useOrderList>;
  customers: Customer[];
  selectedID: string | undefined;
  busy: boolean;
  onOpen: (order: Order) => void;
}

export function OrderList({ list, customers, selectedID, busy, onOpen }: Props) {
  const t = useT();
  const { rows, loading, view } = list;
  const filterOptions = orderListViews.map((value) => ({ value, label: t(`orders.filters.${value}`) }));
  return (
    <section className="min-w-0 flex-[1_1_300px] space-y-4 rounded-xl border border-border bg-surface p-4 xl:max-w-[420px]">
      <h2 className="font-semibold">{t("orders.listTitle")}</h2>
      <SegmentedFilter label={t("orders.filterLabel")} options={filterOptions} value={view} onChange={list.setView} />
      {loading ? (
        <p role="status" className="py-8 text-sm text-muted">
          {t("orders.loading")}
        </p>
      ) : !rows.length ? (
        view === "all" ? (
          <EmptyState title={t("orders.emptyTitle")} description={t("orders.emptyBody")} />
        ) : (
          <EmptyState title={t("orders.emptyFiltered")} />
        )
      ) : (
        <ul className="-mx-2 divide-y divide-border">
          {rows.map((row) => (
            <OrderListItem
              key={row.id}
              order={row}
              customer={customers.find((c) => c.id === row.customer_id)}
              selected={selectedID === row.id}
              busy={busy}
              onOpen={() => onOpen(row)}
            />
          ))}
        </ul>
      )}
      <Pager page={list.page} total={list.total} onChange={list.setPage} disabled={loading || busy} />
    </section>
  );
}
