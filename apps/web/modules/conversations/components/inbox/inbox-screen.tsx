"use client";

import { useState } from "react";
import { ExpiredConnectionBanner } from "@/modules/settings";
import { secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useChannelConnections } from "../../hooks/use-channel-connections";
import { useInboxConversations } from "../../hooks/use-inbox-conversations";
import { useTenantMembers } from "../../hooks/use-tenant-members";
import { ConversationList } from "./conversation-list";
import { InboundSimulator } from "./inbound-simulator";
import { ThreadPanel } from "./thread-panel";

/** The desk: conversation list, selected conversation and the inbound simulator. */
export function InboxScreen() {
  const { tenant, isOwner } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [refresh, setRefresh] = useState(0);
  const [detailRefresh, setDetailRefresh] = useState(0);
  const [selected, setSelected] = useState("");
  const list = useInboxConversations(tenantID, refresh);
  const { members, error: membersError } = useTenantMembers(tenantID);
  const connections = useChannelConnections(tenantID, refresh);
  const reloadList = () => setRefresh((v) => v + 1);
  if (!tenant) return null;
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("desk.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("desk.subtitle")}</p>
        </div>
        <button type="button" className={secondaryClass} disabled={list.loading} onClick={reloadList}>
          {t("desk.refresh")}
        </button>
      </div>
      <p role="note" className="rounded-lg border border-border bg-surface p-3 text-sm text-muted">
        {t("desk.notice")}
      </p>
      <ExpiredConnectionBanner tenant={tenant.id} connections={connections ?? []} owner={isOwner} />
      <div className="flex flex-wrap items-start gap-4">
        <ConversationList
          tenantID={tenant.id}
          owner={isOwner}
          members={members}
          list={list}
          error={list.error || membersError}
          selected={selected}
          onSelect={setSelected}
        />
        {selected ? (
          <ThreadPanel
            key={`${tenant.id}:${selected}`}
            tenantID={tenant.id}
            threadID={selected}
            owner={isOwner}
            members={members}
            connections={connections}
            externalRefresh={detailRefresh}
            onChanged={reloadList}
          />
        ) : (
          <div className="flex-[999_1_420px] rounded-xl border border-dashed border-border p-8 text-center text-muted">
            {t("desk.selectPrompt")}
          </div>
        )}
      </div>
      <InboundSimulator
        tenantID={tenant.id}
        onSaved={(threadID) => {
          list.setFilter("all");
          setSelected(threadID);
          reloadList();
          setDetailRefresh((v) => v + 1);
        }}
      />
    </div>
  );
}
