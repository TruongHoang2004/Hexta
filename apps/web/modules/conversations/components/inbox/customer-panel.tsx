"use client";

import Link from "next/link";
import { StatusChip, buttonClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { MessagingDetail } from "../../services/messaging";
import { initialOf } from "../../utils";

/** Customer facts and shortcuts beside the conversation; `children` holds owner-only controls. */
export function CustomerPanel({
  tenantID,
  detail,
  children,
}: {
  tenantID: string;
  detail: MessagingDetail | null;
  children?: React.ReactNode;
}) {
  const t = useT();
  const source = detail?.conversation.source;
  const turn = detail?.conversation.turn;
  // A draft started here is linked to the open session; a closed session no longer accepts orders.
  const draftHref = `/tenants/${tenantID}/orders${turn?.lifecycle === "open" ? `?turn=${turn.id}` : ""}`;
  return (
    <aside
      aria-label={t("desk.customer.title")}
      className="flex min-w-0 flex-[1_1_240px] flex-col gap-3 self-start rounded-xl border border-border bg-surface p-4 shadow-sm"
    >
      <h2 className="font-semibold">{t("desk.customer.title")}</h2>
      {detail && (
        <dl className="space-y-1 text-sm">
          <dt className="sr-only">{t("desk.conversation")}</dt>
          <dd className="flex items-center gap-3">
            <span
              aria-hidden="true"
              className="grid size-10 shrink-0 place-items-center rounded-full bg-primary text-sm font-bold text-white"
            >
              {initialOf(detail.conversation.thread.display_name)}
            </span>
            <span className="font-medium">{detail.conversation.thread.display_name}</span>
          </dd>
          <dd>
            <StatusChip>{source ? t(`desk.sources.${source}`) : ""}</StatusChip>
          </dd>
        </dl>
      )}
      <Link href={draftHref} className={`${buttonClass} text-center`} aria-describedby="draft-hint">
        {t("desk.customer.createDraft")}
      </Link>
      <p id="draft-hint" className="text-xs text-muted">
        {t("desk.customer.draftHint")}
      </p>
      <Link href={`/tenants/${tenantID}/storefront`} className={`${secondaryClass} text-center`}>
        {t("desk.customer.askConsultant")}
      </Link>
      {children}
    </aside>
  );
}
