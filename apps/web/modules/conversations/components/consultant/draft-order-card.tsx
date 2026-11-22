"use client";

import { buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConsultDraft } from "../../services/consultant";
import { formatVnd } from "../../utils";

/** The consultant's suggested order. Prices are re-checked at checkout. */
export function DraftOrderCard({ draft, onConfirm }: { draft: ConsultDraft; onConfirm: () => void }) {
  const t = useT();
  return (
    <div className="space-y-2 rounded-lg border border-primary p-3">
      <h3 className="font-semibold">{t("consultant.draftTitle")}</h3>
      {draft.items.map((item, index) => (
        <p key={index}>
          {item.quantity} × {item.product_name} · {formatVnd(item.unit_price)}
        </p>
      ))}
      <p className="font-semibold">{t("consultant.total", { amount: formatVnd(draft.total_amount) })}</p>
      <p className="text-xs text-muted">{t("consultant.draftNote")}</p>
      <button className={buttonClass} onClick={onConfirm}>
        {t("consultant.confirm")}
      </button>
    </div>
  );
}
