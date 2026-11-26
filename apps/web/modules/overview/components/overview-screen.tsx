"use client";

import { useState } from "react";
import { EmptyState, Notice, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { useOverview } from "../hooks/use-overview";
import { useOverviewMembers } from "../hooks/use-overview-members";
import type { OverviewFilter } from "../types";
import { OverviewContent } from "./overview-content";
import { OverviewFilters } from "./overview-filters";

/** S-06 owner overview: read-only counts of orders and work with the waiting/stuck list. */
export function OverviewScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [filter, setFilter] = useState<OverviewFilter>({ period: "7d", employeeID: "" });
  const view = useOverview(tenantID, filter, isOwner);
  const lookups = useOverviewMembers(tenantID, isOwner);
  const person = useMemberName(lookups.members);
  const denied = !isOwner || view.denied;
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("overview.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("overview.subtitle")}</p>
        </div>
        {!denied && (
          <button type="button" className={secondaryClass} disabled={view.loading} onClick={view.reload}>
            {t("overview.refresh")}
          </button>
        )}
      </div>
      {denied ? (
        <EmptyState title={t("overview.denied.title")} description={t("overview.denied.body")} />
      ) : (
        <>
          <OverviewFilters filter={filter} members={lookups.members} person={person} onChange={setFilter} />
          <p className="text-xs text-muted">{t("overview.scopeNote")}</p>
          <Notice error={view.error || lookups.error} />
          {view.loading && <p role="status">{t("overview.loading")}</p>}
          {view.data && view.loadedAt && (
            <div aria-busy={view.loading} className={view.loading ? "opacity-60" : undefined}>
              <OverviewContent data={view.data} loadedAt={view.loadedAt} tenantID={tenantID} person={person} />
            </div>
          )}
        </>
      )}
    </div>
  );
}
