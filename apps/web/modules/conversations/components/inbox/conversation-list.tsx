"use client";

import Link from "next/link";
import type { TenantMember } from "@hexta/sdk";
import { EmptyState, SegmentedFilter, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import type { useInboxConversations } from "../../hooks/use-inbox-conversations";
import type { DeskFilter } from "../../types";
import { ConversationListItem } from "./conversation-list-item";

/** Filterable, paged list of conversations on the left of the desk. */
export function ConversationList({
  tenantID,
  owner,
  members,
  list,
  error,
  selected,
  onSelect,
}: {
  tenantID: string;
  owner: boolean;
  members: TenantMember[];
  list: ReturnType<typeof useInboxConversations>;
  error: string;
  selected: string;
  onSelect: (threadID: string) => void;
}) {
  const t = useT();
  const memberName = useMemberName(members);
  const filters: { value: DeskFilter; label: string }[] = [
    { value: "bot", label: t("desk.filters.bot") },
    { value: "needs_human", label: t("desk.filters.needs_human") },
    { value: "mine", label: t("desk.filters.mine") },
    { value: "all", label: t("desk.filters.all") },
  ];
  return (
    <section
      className="flex min-w-0 flex-[1_1_280px] flex-col gap-3 rounded-xl border border-border bg-surface p-3 lg:max-w-[340px]"
      aria-label={t("desk.listLabel")}
    >
      <SegmentedFilter label={t("desk.filterLabel")} options={filters} value={list.filter} onChange={list.setFilter} />
      {error && (
        <p role="alert" className="text-sm text-error">
          {error}
        </p>
      )}
      {list.loading ? (
        <p role="status">{t("desk.loadingList")}</p>
      ) : list.rows.length ? (
        <ul className="-mx-3 divide-y divide-border">
          {list.rows.map((row) => (
            <ConversationListItem
              key={row.thread.id}
              row={row}
              active={selected === row.thread.id}
              assigneeName={row.turn?.assignee_id ? memberName(row.turn.assignee_id) : ""}
              onSelect={() => onSelect(row.thread.id)}
            />
          ))}
        </ul>
      ) : (
        <EmptyState
          title={t("desk.emptyTitle")}
          description={t("desk.emptyBody")}
          action={
            owner ? (
              <Link href={`/tenants/${tenantID}/settings/channels`} className={secondaryClass}>
                {t("desk.connectChannels")}
              </Link>
            ) : undefined
          }
        />
      )}
      <div className="flex gap-2">
        {list.hasPrevious && (
          <button type="button" className={secondaryClass} onClick={list.firstPage}>
            {t("desk.firstPage")}
          </button>
        )}
        {list.hasNext && (
          <button type="button" className={secondaryClass} onClick={list.nextPage}>
            {t("desk.nextPage")}
          </button>
        )}
      </div>
    </section>
  );
}
