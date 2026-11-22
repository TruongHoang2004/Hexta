"use client";

import { EmptyState, Pager, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import type { useCatalog } from "../../hooks/use-catalog";
import type { CatalogKind } from "../../types";
import { CatalogListItem } from "./catalog-list-item";

export function CatalogList({ kind, catalog }: { kind: CatalogKind; catalog: ReturnType<typeof useCatalog> }) {
  const t = useT();
  const k = (key: string) => t(`catalog.${kind}.${key}` as MessageKey);
  const { rows, loading, busy } = catalog;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold">{k("list")}</h2>
        <button type="button" className={secondaryClass} onClick={catalog.refresh} disabled={busy || loading}>
          {t("catalog.refresh")}
        </button>
      </div>
      {loading ? (
        <p role="status" className="py-10 text-sm text-muted">
          {t("catalog.loading")}
        </p>
      ) : !rows.length ? (
        <EmptyState title={k("emptyTitle")} description={k("emptyBody")} />
      ) : (
        <ul className="divide-y divide-border">
          {rows.map((row) => (
            <CatalogListItem
              key={row.id}
              row={row}
              editing={catalog.editing === row.id}
              busy={busy}
              onEdit={() => catalog.edit(row)}
              onArchive={() => void catalog.archive(row)}
            />
          ))}
        </ul>
      )}
      <Pager page={catalog.page} total={catalog.total} disabled={loading || busy} onChange={catalog.setPage} />
    </section>
  );
}
