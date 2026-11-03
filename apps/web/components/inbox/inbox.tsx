"use client";

import { MessageSquare } from "lucide-react";
import { useWorkspace } from "@/components/sales/workspace";
import { ConversationList } from "./conversation-list";
import { ConversationPanel } from "./conversation-panel";
import { CustomerDetails } from "./customer-details";
import { InboxFilters } from "./inbox-filters";
import { InboxHeader } from "./inbox-header";
import { useInbox } from "./use-inbox";

export function Inbox() {
  const { tenant } = useWorkspace();
  const inbox = useInbox();
  return (
    <section className="space-y-4">
      <InboxHeader tenantName={tenant?.name} />
      <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-sm">
        <div className="grid min-h-[680px] xl:grid-cols-[165px_260px_minmax(0,1fr)] 2xl:grid-cols-[175px_280px_minmax(0,1fr)_230px]">
          <InboxFilters
            queue={inbox.queue}
            channel={inbox.channel}
            onQueueChange={inbox.setQueue}
            onChannelChange={inbox.setChannel}
          />
          <ConversationList
            conversations={inbox.visible}
            selected={inbox.selected}
            query={inbox.query}
            onQueryChange={inbox.setQuery}
            onSelect={inbox.setSelectedId}
          />
          {inbox.selected ? (
            <ConversationPanel
              conversation={inbox.selected}
              previews={inbox.previews[inbox.selected.id] ?? []}
              mode={inbox.mode}
              draft={inbox.draft}
              onBack={() => inbox.setSelectedId(null)}
              onModeChange={inbox.setMode}
              onDraftChange={inbox.setDraft}
              onAddPreview={inbox.addPreview}
            />
          ) : (
            <div className="hidden items-center justify-center p-8 text-center text-muted xl:flex">
              <div>
                <MessageSquare className="mx-auto mb-3 size-9" />
                <p>Select a conversation to view messages.</p>
              </div>
            </div>
          )}
          <aside className="hidden border-l border-border p-5 2xl:block">
            <h2 className="text-sm font-semibold">Customer details</h2>
            {inbox.selected ? (
              <CustomerDetails conversation={inbox.selected} />
            ) : (
              <p className="mt-4 text-xs text-muted">Select a conversation.</p>
            )}
          </aside>
        </div>
      </div>
    </section>
  );
}
