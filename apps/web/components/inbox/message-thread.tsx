import type { Conversation, Preview } from "./types";
interface MessageThreadProps {
  conversation: Conversation;
  previews: Preview[];
}
export function MessageThread({
  conversation: selected,
  previews,
}: MessageThreadProps) {
  return (
    <div
      className="flex max-h-[420px] min-h-[300px] flex-1 flex-col gap-4 overflow-y-auto bg-background/40 p-5"
      role="log"
      aria-label={`Messages with ${selected.name}`}
    >
      <p className="text-center text-[10px] tracking-wide text-muted">
        SAMPLE CONVERSATION
      </p>
      {selected.id === "demo-1" && (
        <>
          <div className="max-w-[85%] rounded-xl border border-border bg-surface p-3 text-sm">
            Hi! I’m looking at the everyday tote in your new collection.
          </div>
          <div className="ml-auto max-w-[85%] rounded-xl bg-violet-100 p-3 text-sm text-violet-900">
            Hello Alex! Happy to help. Which size are you interested in?
            <p className="mt-2 text-right text-[10px]">Sample reply · 10:40</p>
          </div>
        </>
      )}
      <div className="max-w-[85%] rounded-xl border border-border bg-surface p-3 text-sm">
        {selected.message}
        <p className="mt-2 text-right text-[10px] text-muted">
          {selected.time}
        </p>
      </div>
      {previews.map((entry, index) => (
        <div
          key={index}
          className={`ml-auto max-w-[85%] whitespace-pre-wrap break-words rounded-xl p-3 text-sm ${entry.mode === "note" ? "bg-amber-100 text-amber-950" : "bg-violet-100 text-violet-900"}`}
        >
          {entry.text}
          <p className="mt-2 text-right text-[10px]">
            {entry.mode === "note" ? "Internal note" : "Reply"} · Local preview
          </p>
        </div>
      ))}
    </div>
  );
}
