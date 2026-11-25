"use client";

import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import { useMoney } from "@/modules/sales";
import type { Overview } from "../types";
import { StatTile } from "./stat-tile";

/** Statuses shown as tiles; other non-empty buckets are listed under the tiles. */
const tileStatuses = ["draft", "confirmed", "cancelled", "returned"];

const statusLabels: Record<string, MessageKey> = {
  draft: "orders.status.draft",
  confirmed: "orders.status.confirmed",
  processing: "orders.status.processing",
  fulfilled: "orders.status.fulfilled",
  cancelled: "orders.status.cancelled",
  returned: "orders.status.returned",
  other: "orders.status.other",
};

/** Order counts by status plus order value (explicitly not revenue). */
export function OrderStatTiles({ orders }: { orders: Overview["orders"] }) {
  const t = useT();
  const money = useMoney();
  // A status added to the API later still shows, under its raw name, until it gets a label.
  const label = (status: string) => (statusLabels[status] ? t(statusLabels[status]) : status);
  const count = (status: string) => orders.by_status.find((bucket) => bucket.status === status)?.count ?? 0;
  const more = orders.by_status.filter((bucket) => !tileStatuses.includes(bucket.status) && bucket.count > 0);
  return (
    <section aria-labelledby="overview-orders" className="space-y-3">
      <h2 id="overview-orders" className="text-sm font-medium text-muted">
        {t("overview.orders.label")} · {t("overview.orders.total", { count: orders.total })}
      </h2>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        {tileStatuses.map((status) => (
          <StatTile key={status} label={label(status)} value={String(count(status))} />
        ))}
        <StatTile label={t("overview.orders.value")} value={money(orders.order_value)} note={t("overview.orders.valueNote")} />
      </div>
      {more.length > 0 && (
        <p className="text-sm text-muted">
          {t("overview.orders.more", { list: more.map((bucket) => `${label(bucket.status)} ${bucket.count}`).join(" · ") })}
        </p>
      )}
    </section>
  );
}
