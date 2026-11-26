"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { ConflictBanner, Notice, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useWorkItem } from "../../hooks/use-work-item";
import { useWorkList } from "../../hooks/use-work-list";
import { useWorkLookups } from "../../hooks/use-work-lookups";
import { WorkCreator } from "./work-creator";
import { WorkDeepLink } from "./work-deep-link";
import { WorkDetail } from "./work-detail";
import { WorkList } from "./work-list";

/** Work items: owners start work from confirmed orders; everyone moves their own work along. */
export function WorkScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [revision, setRevision] = useState(0);
  const reload = () => setRevision((v) => v + 1);
  const work = useWorkItem(tenantID, reload);
  const list = useWorkList(tenantID, revision, work.setError);
  const lookups = useWorkLookups(tenantID, revision);
  const { selected, busy } = work;
  return (
    <div className="space-y-5">
      <Suspense fallback={null}>
        <WorkDeepLink onOpen={(id) => void work.open(id)} />
      </Suspense>
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("work.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("work.subtitle")}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link className={secondaryClass} href={`/tenants/${tenant?.id}/workflows`}>
            {t("work.manageWorkflows")}
          </Link>
          <button
            type="button"
            className={secondaryClass}
            disabled={busy || list.loading}
            onClick={() => {
              work.setError("");
              reload();
              if (selected) void work.open(selected.id);
            }}
          >
            {t("work.refresh")}
          </button>
        </div>
      </div>
      {work.conflict && selected ? (
        <ConflictBanner
          message={work.error}
          actionLabel={t("work.actions.reload")}
          onReload={() => void work.open(selected.id)}
        />
      ) : (
        <Notice error={work.error || lookups.error} success={work.success} />
      )}
      {lookups.loading ? (
        <p role="status">{t("work.loadingLookups")}</p>
      ) : (
        !lookups.error &&
        isOwner && (
          <WorkCreator
            key={work.creatorKey}
            definitions={lookups.definitions}
            orders={lookups.orders}
            members={lookups.members}
            busy={busy}
            onCreate={work.create}
          />
        )
      )}
      <div className="flex flex-wrap items-start gap-5">
        <WorkList
          list={list}
          members={lookups.members}
          selectedID={selected?.id}
          busy={busy}
          onOpen={(id) => void work.open(id)}
        />
        <div className="min-w-0 flex-[999_1_480px]">
          {selected ? (
            <WorkDetail
              key={`${selected.id}:${selected.version}`}
              item={selected}
              members={lookups.members}
              busy={busy}
              onCommand={work.command}
              onRefresh={() => void work.open(selected.id)}
            />
          ) : (
            <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted">
              {t("work.selectPrompt")}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
