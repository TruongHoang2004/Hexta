"use client";

import Link from "next/link";
import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { OverviewStuckItem } from "../types";
import { waitedFor } from "../utils";

interface Props {
  item: OverviewStuckItem;
  tenantID: string;
  person: (userID: string | null | undefined) => string;
  now: Date;
}

/** One waiting/stuck work item with a link that opens it on the Work page. */
export function StuckWorkRow({ item, tenantID, person, now }: Props) {
  const t = useT();
  const waited = waitedFor(item.waiting_since, now);
  const number = item.order_number || item.order_id.slice(-6);
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="px-3 py-2">
        <span className="block font-medium">{t("overview.stuck.order", { number })}</span>
        <span className="block text-xs text-muted">{item.workflow_name}</span>
      </td>
      <td className="px-3 py-2">
        {item.step_ordinal + 1}. {item.step_name}
      </td>
      <td className="break-all px-3 py-2">
        {item.responsible_id ? person(item.responsible_id) : t("overview.stuck.unassigned")}
        {item.state === "waiting" && item.recipient_id && (
          <span className="block text-xs text-muted">{t("overview.stuck.waitingFor", { name: person(item.recipient_id) })}</span>
        )}
      </td>
      <td className="whitespace-nowrap px-3 py-2 tabular-nums">
        <time dateTime={item.waiting_since}>{t(`overview.stuck.waited.${waited.unit}`, { value: waited.value })}</time>
      </td>
      <td className="px-3 py-2">
        <StatusChip tone={item.reason === "unassigned" ? "warning" : "neutral"}>{t(`overview.stuck.reasons.${item.reason}`)}</StatusChip>
      </td>
      <td className="px-3 py-2 text-right">
        <Link
          href={`/tenants/${tenantID}/work?item=${encodeURIComponent(item.work_item_id)}`}
          aria-label={t("overview.stuck.openLabel", { number })}
          className="text-primary underline"
        >
          {t("overview.stuck.open")}
        </Link>
      </td>
    </tr>
  );
}
