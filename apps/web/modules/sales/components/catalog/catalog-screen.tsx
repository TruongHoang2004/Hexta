"use client";

import { Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import { useCatalog } from "../../hooks/use-catalog";
import type { CatalogKind } from "../../types";
import { CatalogForm } from "./catalog-form";
import { CatalogList } from "./catalog-list";

/** Customers or products: paged list beside the add/edit form. */
export function CatalogScreen({ kind }: { kind: CatalogKind }) {
  const { tenant } = useWorkspace();
  const t = useT();
  const catalog = useCatalog(kind, tenant?.id ?? "");
  const k = (key: string) => t(`catalog.${kind}.${key}` as MessageKey);
  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold">{k("title")}</h1>
        <p className="mt-1 text-sm text-muted">{k("subtitle")}</p>
      </div>
      <Notice error={catalog.error} success={catalog.success} />
      <div className="grid items-start gap-5 xl:grid-cols-[1fr_320px]">
        <CatalogList kind={kind} catalog={catalog} />
        <CatalogForm kind={kind} catalog={catalog} />
      </div>
    </div>
  );
}
