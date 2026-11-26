"use client";

import { useT } from "@/lib/i18n/provider";
import type { Overview } from "../types";
import { PendingCaseRow } from "./pending-case-row";
import { StuckWorkRow } from "./stuck-work-row";

interface Props {
  stuck: Overview["stuck"];
  /** Cancellation/return cases waiting for the owner, listed after the work rows. */
  pending: Overview["pending_cases"];
  tenantID: string;
  person: (userID: string | null | undefined) => string;
  /** Reference time for "waiting" durations, fixed per loaded response. */
  now: Date;
}

/** Open work that is unassigned or has not moved for longer than the threshold, then cases awaiting the owner. */
export function StuckWorkTable({ stuck, pending, tenantID, person, now }: Props) {
  const t = useT();
  return (
    <section aria-labelledby="overview-stuck" className="min-w-0 space-y-3 rounded-xl border border-border bg-surface p-4">
      <div>
        <h2 id="overview-stuck" className="font-semibold">
          {t("overview.stuck.title")}
        </h2>
        <p className="mt-1 text-sm text-muted">{t("overview.stuck.note", { hours: stuck.threshold_hours })}</p>
      </div>
      {stuck.items.length === 0 && pending.items.length === 0 ? (
        <p className="text-sm text-muted">{t("overview.stuck.empty")}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] border-collapse text-left text-sm">
            <thead>
              <tr className="border-b border-border text-xs text-muted">
                <th scope="col" className="px-3 py-2 font-medium">{t("overview.stuck.columns.item")}</th>
                <th scope="col" className="px-3 py-2 font-medium">{t("overview.stuck.columns.step")}</th>
                <th scope="col" className="px-3 py-2 font-medium">{t("overview.stuck.columns.responsible")}</th>
                <th scope="col" className="px-3 py-2 font-medium">{t("overview.stuck.columns.waited")}</th>
                <th scope="col" className="px-3 py-2 font-medium">{t("overview.stuck.columns.reason")}</th>
                <th scope="col" className="px-3 py-2 text-right font-medium">
                  <span className="sr-only">{t("overview.stuck.columns.action")}</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {stuck.items.map((item) => (
                <StuckWorkRow key={item.work_item_id} item={item} tenantID={tenantID} person={person} now={now} />
              ))}
              {pending.items.map((item) => (
                <PendingCaseRow key={item.case_id} item={item} tenantID={tenantID} person={person} now={now} />
              ))}
            </tbody>
          </table>
        </div>
      )}
      {stuck.total > stuck.items.length && (
        <p className="text-xs text-muted">{t("overview.stuck.partial", { shown: stuck.items.length, total: stuck.total })}</p>
      )}
      {pending.total > pending.items.length && (
        <p className="text-xs text-muted">{t("overview.stuck.partialCases", { shown: pending.items.length, total: pending.total })}</p>
      )}
    </section>
  );
}
