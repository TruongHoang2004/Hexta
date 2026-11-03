import { Inbox as InboxIcon } from "lucide-react";

export function InboxHeader({ tenantName }: { tenantName?: string }) {
  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-semibold">
            <InboxIcon className="size-6 text-violet-500" />
            Inbox
          </h1>
          <p className="mt-1 text-sm text-muted">
            Conversations for {tenantName}
          </p>
        </div>
        <span className="rounded-full bg-violet-100 px-3 py-1 text-xs font-medium text-violet-700">
          Design preview
        </span>
      </div>
      <p className="rounded-lg border border-border bg-surface px-4 py-3 text-xs text-muted">
        Sample conversations and channels. Replies and notes stay in this
        preview until you leave the page; no messages are sent.
      </p>
    </>
  );
}
