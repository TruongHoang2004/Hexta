import { ChevronLeft } from "lucide-react";
import { Avatar } from "./avatar";
import { CustomerDetails } from "./customer-details";
import { MessageComposer } from "./message-composer";
import { MessageThread } from "./message-thread";
import type { Conversation, Mode, Preview } from "./types";
interface ConversationPanelProps {
  conversation: Conversation;
  previews: Preview[];
  mode: Mode;
  draft: string;
  onBack: () => void;
  onModeChange: (mode: Mode) => void;
  onDraftChange: (draft: string) => void;
  onAddPreview: () => void;
}
export function ConversationPanel({
  conversation: selected,
  previews,
  mode,
  draft,
  onBack,
  onModeChange,
  onDraftChange,
  onAddPreview,
}: ConversationPanelProps) {
  return (
    <div className="flex min-w-0 flex-col">
      <header className="flex items-center justify-between gap-3 border-b border-border p-4">
        <div className="flex min-w-0 items-center gap-3">
          <button
            aria-label="Back to conversations"
            onClick={() => onBack()}
            className="xl:hidden"
          >
            <ChevronLeft className="size-5" />
          </button>
          <Avatar initials={selected.initials} />
          <div>
            <h2 className="text-sm font-semibold">{selected.name}</h2>
            <p className="mt-1 text-xs text-muted">{selected.channel} · Demo</p>
          </div>
        </div>
        <span className="rounded-md bg-background px-2 py-1 text-xs">
          {selected.closed ? "Closed" : "Open"}
        </span>
      </header>
      <div className="flex flex-wrap justify-between gap-2 bg-violet-50 px-4 py-3 text-[11px] text-violet-700 dark:bg-violet-950/30 dark:text-violet-300">
        <span>Assignee: {selected.assignee ?? "Unassigned"}</span>
        <span>{selected.assignee ? "Sales control" : "System control"}</span>
      </div>
      <MessageThread conversation={selected} previews={previews} />
      <MessageComposer
        closed={selected.closed}
        mode={mode}
        draft={draft}
        onModeChange={onModeChange}
        onDraftChange={onDraftChange}
        onAddPreview={onAddPreview}
      />
      <details className="border-t border-border p-4 2xl:hidden">
        <summary className="cursor-pointer text-sm font-medium">
          Customer details
        </summary>
        <CustomerDetails conversation={selected} />
      </details>
    </div>
  );
}
