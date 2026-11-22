"use client";

import { StatusChip, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { CatalogRow } from "../../types";
import { isCustomer } from "../../utils";

interface Props {
  row: CatalogRow;
  editing: boolean;
  busy: boolean;
  onEdit: () => void;
  onArchive: () => void;
}

export function CatalogListItem({ row, editing, busy, onEdit, onArchive }: Props) {
  const t = useT();
  const variant = isCustomer(row) ? undefined : row.variants[0];
  return (
    <li
      className={`flex flex-wrap items-start justify-between gap-3 py-4 ${editing ? "rounded-lg bg-primary/5 px-2" : ""}`}
    >
      <div className="min-w-0">
        <p className="font-medium">{isCustomer(row) ? row.full_name : row.name}</p>
        {isCustomer(row) ? (
          <p className="mt-1 break-all text-sm text-muted">
            {[row.phone, row.email].filter(Boolean).join(" · ") || t("catalog.customers.noContact")}
          </p>
        ) : (
          <>
            <p className="mt-1 text-sm text-muted">
              {variant?.sku} · {variant?.variant_name}
            </p>
            <p className="mt-1 text-sm">
              {!variant?.price_provided ? (
                <StatusChip tone="danger">{t("catalog.products.priceNotSet")}</StatusChip>
              ) : Number(variant.price) === 0 ? (
                t("catalog.products.free")
              ) : (
                t("catalog.products.perUnit", { price: String(variant.price) })
              )}
            </p>
          </>
        )}
      </div>
      <div className="flex gap-2">
        <button type="button" className={secondaryClass} disabled={busy} onClick={onEdit}>
          {t("catalog.edit")}
        </button>
        <button type="button" className={secondaryClass} disabled={busy} onClick={onArchive}>
          {t("catalog.archive")}
        </button>
      </div>
    </li>
  );
}
