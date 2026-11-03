import { Search } from "lucide-react";
import { Avatar } from "./avatar";
import type { Conversation } from "./types";
interface ConversationListProps {
  conversations: Conversation[];
  selected: Conversation | null;
  query: string;
  onQueryChange: (query: string) => void;
  onSelect: (id: string) => void;
}
export function ConversationList({
  conversations: visible,
  selected,
  query,
  onQueryChange,
  onSelect,
}: ConversationListProps) {
  return (
    <div
      className={`${selected ? "hidden xl:block" : "block"} border-r border-border`}
    >
      <div className="space-y-4 p-4">
        <div className="flex justify-between">
          <h2 className="font-semibold">Conversations</h2>
          <span className="text-xs text-muted">{visible.length}</span>
        </div>
        <label className="flex items-center gap-2 rounded-lg border border-border px-3">
          <Search className="size-4 text-muted" />
          <input
            aria-label="Search conversations"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search name or message…"
            className="min-w-0 w-full bg-transparent py-2.5 text-xs outline-none"
          />
        </label>
        <p className="text-xs text-muted">Newest first</p>
      </div>
      {visible.map((item) => (
        <button
          key={item.id}
          onClick={() => onSelect(item.id)}
          aria-pressed={selected?.id === item.id}
          className={`flex w-full gap-3 border-t border-border p-4 text-left ${selected?.id === item.id ? "border-l-2 border-l-violet-500 bg-violet-50 dark:bg-violet-950/30" : "hover:bg-background"}`}
        >
          <Avatar initials={item.initials} />
          <span className="min-w-0 flex-1">
            <span className="flex justify-between gap-2">
              <strong className="text-xs">{item.name}</strong>
              <span className="text-[10px] text-muted">{item.time}</span>
            </span>
            <span className="mt-2 block truncate text-xs text-muted">
              {item.message}
            </span>
            <span className="mt-2 flex items-center justify-between text-[10px] text-muted">
              <span>
                {item.channel}
                {item.needsHuman ? " · Attention" : ""}
              </span>
              {item.unread > 0 && (
                <span className="rounded bg-violet-100 px-1.5 text-violet-700">
                  {item.unread}
                </span>
              )}
            </span>
          </span>
        </button>
      ))}
      {visible.length === 0 && (
        <p role="status" className="p-5 text-sm text-muted">
          No conversations match these filters.
        </p>
      )}
    </div>
  );
}
