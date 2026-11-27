"use client";

import Link from "next/link";
import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OverviewPendingCase } from "../types";
import { waitedFor } from "../utils";

interface Props {
  item: OverviewPendingCase;
  tenantID: string;
  person: (userID: string | null | undefined) => string;
  now: Date;
}

/** One cancellation/return case waiting for the owner, with a link that opens its order on the Orders page. */
export function PendingCaseRow({ item, tenantID, person, now }: Props) {
  const t = useT();
  const waited = waitedFor(item.waiting_since, now);
  const number = item.order_number || item.order_id.slice(-6);
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="px-3 py-2">
        <span className="block font-medium">{t(`overview.stuck.case.${item.case_type}`, { number })}</span>
      </td>
      <td className="px-3 py-2">{t("overview.stuck.awaitingOwner")}</td>
      <td className="break-all px-3 py-2">{person(item.handler_id)}</td>
      <td className="whitespace-nowrap px-3 py-2 tabular-nums">
        <time dateTime={item.waiting_since}>{t(`overview.stuck.waited.${waited.unit}`, { value: waited.value })}</time>
      </td>
      <td className="px-3 py-2">
        <StatusChip tone="warning">{t("overview.stuck.reasons.awaiting_owner")}</StatusChip>
      </td>
      <td className="px-3 py-2 text-right">
        <Link
          href={`/tenants/${tenantID}/orders?order=${encodeURIComponent(item.order_id)}`}
          aria-label={t("overview.stuck.openCaseLabel", { number })}
          className="text-primary underline"
        >
          {t("overview.stuck.open")}
        </Link>
      </td>
    </tr>
  );
}
