import type { Channel, Queue } from "./types";
import { conversations, queues } from "./sample-data";
import { matchesQueue } from "./filters";

interface InboxFiltersProps {
  queue: Queue;
  channel: Channel | "all";
  onQueueChange: (queue: Queue) => void;
  onChannelChange: (channel: Channel | "all") => void;
}
export function InboxFilters({
  queue,
  channel,
  onQueueChange,
  onChannelChange,
}: InboxFiltersProps) {
  return (
    <aside className="border-b border-border bg-background/50 p-3 xl:border-r xl:border-b-0">
      <h2 className="mb-3 px-2 text-xs font-semibold tracking-wide text-muted">
        WORKSPACE INBOX
      </h2>
      <div className="flex flex-wrap gap-1 xl:flex-col">
        {queues.map((item) => (
          <button
            key={item.id}
            onClick={() => onQueueChange(item.id)}
            aria-pressed={queue === item.id}
            className={`flex items-center justify-between gap-2 rounded-lg px-2 py-2.5 text-left text-xs ${queue === item.id ? "bg-violet-100 font-medium text-violet-800" : "text-muted hover:bg-background"}`}
          >
            <span>{item.label}</span>
            <span>
              {conversations.filter((c) => matchesQueue(c, item.id)).length}
            </span>
          </button>
        ))}
      </div>
      <label
        className="mt-5 block px-2 text-xs font-semibold text-muted"
        htmlFor="inbox-channel"
      >
        CHANNEL
      </label>
      <select
        id="inbox-channel"
        value={channel}
        onChange={(event) =>
          onChannelChange(event.target.value as Channel | "all")
        }
        className="mt-2 w-full rounded-lg border border-border bg-surface p-2 text-xs"
      >
        <option value="all">All channels</option>
        {(["Facebook", "Telegram", "Web chat"] as const).map((c) => (
          <option key={c}>{c}</option>
        ))}
      </select>
    </aside>
  );
}
