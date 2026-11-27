"use client";

import { EmptyState, Pager, SegmentedFilter } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { useScriptList } from "../../hooks/use-script-list";
import type { ScriptStatusFilter } from "../../types";
import { ScriptListItem } from "./script-list-item";

const filters: ScriptStatusFilter[] = ["all", "active", "inactive", "archived"];

interface Props {
  list: ReturnType<typeof useScriptList>;
  selectedID: string | undefined;
  onSelect: (id: string) => void;
}

export function ScriptList({ list, selectedID, onSelect }: Props) {
  const t = useT();
  return (
    <section className="min-w-0 flex-[1_1_320px] space-y-4 rounded-xl border border-border bg-surface p-5">
      <h2 className="font-semibold">{t("scripts.list")}</h2>
      <SegmentedFilter
        label={t("scripts.filterLabel")}
        value={list.status}
        onChange={list.setStatus}
        options={filters.map((value) => ({ value, label: t(`scripts.filters.${value}`) }))}
      />
      <div className="grid grid-cols-[1fr_auto_auto] gap-3 px-3 text-xs font-medium text-muted" aria-hidden="true">
        <span>{t("scripts.columns.name")}</span>
        <span>{t("scripts.columns.active")}</span>
        <span>{t("scripts.columns.status")}</span>
      </div>
      {list.loading ? (
        <p role="status">{t("scripts.loading")}</p>
      ) : !list.rows.length ? (
        <EmptyState title={t("scripts.emptyTitle")} description={t("scripts.emptyBody")} />
      ) : (
        <ul className="space-y-2">
          {list.rows.map((row) => (
            <ScriptListItem key={row.id} script={row} selected={row.id === selectedID} onSelect={() => onSelect(row.id)} />
          ))}
        </ul>
      )}
      <Pager page={list.page} total={list.total} onChange={list.setPage} disabled={list.loading} />
    </section>
  );
}
