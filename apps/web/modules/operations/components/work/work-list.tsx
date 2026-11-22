"use client";

import type { TenantMember } from "@hexta/sdk";
import { EmptyState, Pager, Tabs } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import type { useWorkList } from "../../hooks/use-work-list";
import { workViews } from "../../utils";
import { WorkListItem } from "./work-list-item";

interface Props {
  list: ReturnType<typeof useWorkList>;
  members: TenantMember[];
  selectedID: string | undefined;
  busy: boolean;
  onOpen: (id: string) => void;
}

/** Tabbed list: my work, awaiting my acceptance, all. */
export function WorkList({ list, members, selectedID, busy, onOpen }: Props) {
  const t = useT();
  const person = useMemberName(members);
  const { view, rows, loading } = list;
  const tabs = workViews.map((value) => ({ value, label: t(`work.tabs.${value}`) }));
  return (
    <section className="min-w-0 flex-[1_1_320px] space-y-4 rounded-xl border border-border bg-surface p-4 xl:max-w-[460px]">
      <Tabs label={t("work.tabsLabel")} idPrefix="work" items={tabs} value={view} onChange={list.setView} />
      <div id="work-panel" role="tabpanel" aria-labelledby={`work-tab-${view}`} className="space-y-2">
        {loading ? (
          <p role="status">{t("work.loading")}</p>
        ) : !rows.length ? (
          <EmptyState title={t(`work.empty.${view}`)} />
        ) : (
          rows.map((row) => (
            <WorkListItem
              key={row.id}
              item={row}
              person={person}
              selected={selectedID === row.id}
              busy={busy}
              onOpen={() => onOpen(row.id)}
            />
          ))
        )}
      </div>
      <Pager page={list.page} total={list.total} onChange={list.setPage} disabled={busy || loading} />
    </section>
  );
}
