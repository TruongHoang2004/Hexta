"use client";

import Link from "next/link";
import { EmptyState, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { Overview } from "../types";
import { isEmptySelection } from "../utils";
import { OrderStatTiles } from "./order-stat-tiles";
import { StuckWorkTable } from "./stuck-work-table";
import { WorkByStepCard } from "./work-by-step-card";

interface Props {
  data: Overview;
  loadedAt: Date;
  tenantID: string;
  person: (userID: string | null | undefined) => string;
}

/** Loaded overview: empty-workspace or empty-period states, otherwise tiles, work by step and stuck work. */
export function OverviewContent({ data, loadedAt, tenantID, person }: Props) {
  const t = useT();
  if (!data.has_activity) {
    return (
      <EmptyState
        title={t("overview.empty.tenantTitle")}
        description={t("overview.empty.tenantBody")}
        action={
          <Link className={secondaryClass} href={`/tenants/${tenantID}/orders`}>
            {t("overview.empty.goOrders")}
          </Link>
        }
      />
    );
  }
  if (isEmptySelection(data)) {
    return <EmptyState title={t("overview.empty.periodTitle")} description={t("overview.empty.periodBody")} />;
  }
  return (
    <div className="space-y-5">
      <OrderStatTiles orders={data.orders} exceptions={data.exceptions} />
      <div className="grid gap-5 lg:grid-cols-3">
        <WorkByStepCard work={data.work} />
        <div className="min-w-0 lg:col-span-2">
          <StuckWorkTable stuck={data.stuck} pending={data.pending_cases} tenantID={tenantID} person={person} now={loadedAt} />
        </div>
      </div>
    </div>
  );
}
